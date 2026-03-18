package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/motd"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/sunset"
)

const commandTimeout = 1500 * time.Millisecond

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Underline(true)
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// Show displays the current configuration status
func Show() error {
	fmt.Println(titleStyle.Render("Bluefin CLI Status"))
	fmt.Println()

	// --- Left Column ---
	var leftCol string

	// Shell status
	leftCol += labelStyle.Render("Shell Experience:") + "\n"
	shellStatus := shell.CheckStatus()
	installed := shell.GetInstalledShells()

	// Ensure we show shells that are enabled even if not "installed" (in PATH)
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

	// Get default shell
	defaultShellPath := os.Getenv("SHELL")
	defaultShell := filepath.Base(defaultShellPath)

	// Get current shell (heuristic using parent process)
	// We use `ps -p $PPID -o comm=` to get the command name of the parent process
	var currentShell string
	if ppid := os.Getppid(); ppid > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=")
		if out, err := cmd.Output(); err == nil {
			comm := strings.TrimSpace(string(out))
			// Handle e.g. /bin/zsh or -zsh
			comm = strings.TrimPrefix(comm, "-")
			currentShell = filepath.Base(comm)
		}
	}

	if len(shellsToShow) == 0 {
		leftCol += "  (no compatible shells found)\n"
	}

	for _, s := range shellsToShow {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if shellStatus[s] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		markers := ""
		isDefault := s == defaultShell
		isCurrent := s == currentShell

		if isDefault && isCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default, current)")
		} else if isDefault {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default)")
		} else if isCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(" ● (current)")
		}

		leftCol += fmt.Sprintf("  %s %s: %s%s\n",
			style.Render(symbol),
			s,
			style.Render(status),
			markers)
	}
	leftCol += "\n"

	// MOTD status
	leftCol += labelStyle.Render("Message of the Day:") + "\n"
	motdStatus := motd.CheckStatus()
	for _, s := range shellsToShow {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if motdStatus[s] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		leftCol += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			s,
			style.Render(status))
	}

	// --- Right Column ---
	var rightCol string

	// Tool dependencies
	rightCol += labelStyle.Render("Managed Tools:") + "\n"
	deps := shell.CheckDependencies()

	var toolsToShow []shell.Tool
	if env.IsWindows() {
		toolsToShow = shell.ToolsForShell("powershell")
	} else {
		// Heuristic: if we can't determine current shell, at least filter for platform-aware tools
		// But ideally we use the detected currentShell if it's one of the supported ones.
		targetShell := currentShell
		if targetShell == "" {
			targetShell = "bash" // default fallback for filtering
		}
		toolsToShow = shell.ToolsForShell(targetShell)
	}

	for _, tool := range toolsToShow {
		status := "not installed"
		style := disabledStyle
		symbol := "✗"

		if deps[tool.Binary] {
			status = "installed"
			style = enabledStyle
			symbol = "✓"
		}

		rightCol += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			tool.Name,
			style.Render(status))
	}
	rightCol += "\n"

	// Homebrew status
	rightCol += labelStyle.Render("Package Manager:") + "\n"
	if env.IsWindows() {
		managers := install.AvailableWindowsManagers()
		if len(managers) == 0 {
			rightCol += fmt.Sprintf("  %s Windows PMs: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("none detected"))
			rightCol += "    Install winget, scoop, or chocolatey\n"
		} else {
			rightCol += fmt.Sprintf("  %s Windows PMs: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render(strings.Join(managers, ", ")))
		}
	} else {
		if _, err := exec.LookPath("brew"); err == nil {
			rightCol += fmt.Sprintf("  %s Homebrew: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render("installed"))

			ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
			defer cancel()
			if output, err := exec.CommandContext(ctx, "brew", "--version").Output(); err == nil {
				version := string(output)
				if len(version) > 0 {
					rightCol += fmt.Sprintf("    %s\n", version[:len(version)-1])
				}
			}
		} else {
			rightCol += fmt.Sprintf("  %s Homebrew: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("not installed"))
			rightCol += "    Install from: https://brew.sh\n"
		}
	}
	rightCol += "\n"

	// Sunset status
	rightCol += labelStyle.Render("Sunset Automation:") + "\n"
	if cfg, err := sunset.LoadConfig(); err == nil && cfg.Enabled {
		rightCol += fmt.Sprintf("  %s Status: %s\n",
			enabledStyle.Render("✓"),
			enabledStyle.Render("enabled"))
		rightCol += fmt.Sprintf("    Location: %.4f, %.4f\n", cfg.Latitude, cfg.Longitude)
		if cfg.WallpaperTheme != "" {
			rightCol += fmt.Sprintf("    Theme: %s\n", cfg.WallpaperTheme)
		}
	} else {
		rightCol += fmt.Sprintf("  %s Status: %s\n",
			disabledStyle.Render("✗"),
			disabledStyle.Render("disabled"))
		rightCol += "    Run 'bluefin-cli sunset setup' to enable\n"
	}

	// Combine columns with padding
	formatted := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(40).Render(leftCol),
		string(rightCol),
	)

	fmt.Println(formatted)

	return nil
}
