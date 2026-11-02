package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

type shellInfo struct {
	shellStatus  map[string]bool
	shellsToShow []string
	defaultShell string
	currentShell string
}

func getShellInfo() shellInfo {
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

	return shellInfo{
		shellStatus:  shellStatus,
		shellsToShow: shellsToShow,
		defaultShell: defaultShell,
		currentShell: currentShell,
	}
}

func reportShellExperience(info shellInfo) string {
	var out string
	out += labelStyle.Render("Shell Experience:") + "\n"

	if len(info.shellsToShow) == 0 {
		out += "  (no compatible shells found)\n"
	}

	for _, s := range info.shellsToShow {
		status := "disabled"
		if info.shellStatus[s] {
			status = "enabled"
		}

		markers := ""
		isDefault := s == info.defaultShell
		isCurrent := s == info.currentShell

		if isDefault && isCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default, current)")
		} else if isDefault {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default)")
		} else if isCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(" ● (current)")
		}

		out += report(info.shellStatus[s], s, status, markers)
	}
	return out
}

func reportManagedTools(currentShell string) string {
	var out string
	out += labelStyle.Render("Managed Tools:") + "\n"
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
		if deps[tool.Binary] {
			status = "installed"
		}

		out += report(deps[tool.Binary], tool.Name, status)
	}
	return out
}
