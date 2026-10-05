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
	"github.com/charmbracelet/x/term"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

const commandTimeout = 1500 * time.Millisecond

var (
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// ShellStatusProvider renders shell experience status (installation, enablement, default, current shell).
type ShellStatusProvider struct{}

func NewShellStatusProvider() *ShellStatusProvider {
	return &ShellStatusProvider{}
}

func (p *ShellStatusProvider) Name() string {
	return "Shell Experience"
}

func (p *ShellStatusProvider) Render(width int) (string, error) {
	var output string

	output += labelStyle.Render("Shell Experience:") + "\n"
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
		output += "  (no compatible shells found)\n"
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

		output += fmt.Sprintf("  %s %s: %s%s\n",
			style.Render(symbol),
			s,
			style.Render(status),
			markers)
	}

	return output, nil
}
