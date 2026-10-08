package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/sunset"
)

const commandTimeout = 1500 * time.Millisecond

// ShellItem represents status of a single shell.
type ShellItem struct {
	Name      string
	Enabled   bool
	IsDefault bool
	IsCurrent bool
}

// ToolItem represents status of a managed tool dependency.
type ToolItem struct {
	Name      string
	Binary    string
	Installed bool
}

// PackageManagerStatus holds package manager diagnostics.
type PackageManagerStatus struct {
	IsWindows       bool
	WindowsManagers []string
	BrewInstalled   bool
	BrewVersion     string
}

// SunsetInfo holds status information for sunset automation.
type SunsetInfo struct {
	Enabled        bool
	Latitude       float64
	Longitude      float64
	WallpaperTheme string
}

// Report holds the collected state of all subsystems needed for status rendering.
type Report struct {
	AppVersion     string
	Flavor         string
	Shells         []ShellItem
	MOTD           map[string]bool
	Tools          []ToolItem
	PackageManager PackageManagerStatus
	Sunset         *SunsetInfo
}

// Collector gathers system state across subsystems without rendering.
type Collector interface {
	Collect() Report
}

// SystemCollector is the default collector gathering live system state.
type SystemCollector struct{}

// NewSystemCollector creates a default Collector.
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{}
}

// Collect gathers system status from shell, motd, tools, package manager, and sunset.
func (c *SystemCollector) Collect() Report {
	shellStatus := shell.CheckStatus()
	installed := shell.GetInstalledShells()

	shellsToShow := installed
	for s, enabled := range shellStatus {
		if enabled {
			found := false
			for _, is := range installed {
				if is == s {
					found = true
					break
				}
			}
			if !found {
				shellsToShow = append(shellsToShow, s)
			}
		}
	}

	defaultShellPath := os.Getenv("SHELL")
	defaultShell := filepath.Base(defaultShellPath)

	var currentShell string
	if ppid := os.Getppid(); ppid > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=")
		if out, err := cmd.Output(); err == nil {
			comm := strings.TrimSpace(string(out))
			comm = strings.TrimPrefix(comm, "-")
			currentShell = filepath.Base(comm)
		}
	}

	var shellItems []ShellItem
	for _, s := range shellsToShow {
		shellItems = append(shellItems, ShellItem{
			Name:      s,
			Enabled:   shellStatus[s],
			IsDefault: s == defaultShell,
			IsCurrent: s == currentShell,
		})
	}

	motdStatus := motd.CheckStatus()

	deps := shell.CheckDependencies()
	var toolsToShow []shell.Tool
	if env.IsWindows() {
		toolsToShow = shell.ToolsForShell("powershell")
	} else {
		targetShell := currentShell
		if targetShell == "" {
			targetShell = "bash"
		}
		toolsToShow = shell.ToolsForShell(targetShell)
	}

	var toolItems []ToolItem
	for _, tool := range toolsToShow {
		toolItems = append(toolItems, ToolItem{
			Name:      tool.Name,
			Binary:    tool.Binary,
			Installed: deps[tool.Binary],
		})
	}

	var pm PackageManagerStatus
	if env.IsWindows() {
		pm.IsWindows = true
		pm.WindowsManagers = install.AvailableWindowsManagers()
	} else {
		pm.IsWindows = false
		if _, err := exec.LookPath("brew"); err == nil {
			pm.BrewInstalled = true
			ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
			defer cancel()
			if output, err := exec.CommandContext(ctx, "brew", "--version").Output(); err == nil {
				version := string(output)
				if len(version) > 0 {
					pm.BrewVersion = version[:len(version)-1]
				}
			}
		}
	}

	var sunsetInfo *SunsetInfo
	if env.IsWindows() || env.IsWSL() {
		sunsetInfo = &SunsetInfo{}
		if cfg, err := sunset.LoadConfig(); err == nil && cfg.Enabled {
			sunsetInfo.Enabled = true
			sunsetInfo.Latitude = cfg.Latitude
			sunsetInfo.Longitude = cfg.Longitude
			sunsetInfo.WallpaperTheme = cfg.WallpaperTheme
		}
	}

	return Report{
		AppVersion:     AppVersion,
		Flavor:         viper.GetString("ui.flavor"),
		Shells:         shellItems,
		MOTD:           motdStatus,
		Tools:          toolItems,
		PackageManager: pm,
		Sunset:         sunsetInfo,
	}
}
