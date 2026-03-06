package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/motd"
	"github.com/hanthor/bluefin-cli/internal/shell"
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
	installedShells := shell.GetInstalledShells()

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

	if len(installedShells) == 0 {
		leftCol += "  (no compatible shells found)\n"
	}

	for _, s := range installedShells {
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
	for _, s := range installedShells {
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

	for _, tool := range shell.Tools {
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

	// WSL wallpaper/theme sync readiness (WSL-only)
	if env.IsWSL() {
		rightCol += labelStyle.Render("WSL Wallpaper Sync:") + "\n"
		rightCol += fmt.Sprintf("  %s Runtime: %s\n",
			enabledStyle.Render("✓"),
			enabledStyle.Render("WSL detected"))

		if _, err := exec.LookPath("powershell.exe"); err == nil {
			rightCol += fmt.Sprintf("  %s powershell.exe: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render("available"))
		} else {
			rightCol += fmt.Sprintf("  %s powershell.exe: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("not available"))
		}

		if _, err := exec.LookPath("wslpath"); err == nil {
			rightCol += fmt.Sprintf("  %s wslpath: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render("available"))
		} else {
			rightCol += fmt.Sprintf("  %s wslpath: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("not available"))
		}

		rightCol += fmt.Sprintf("  %s Startup Sync: %s\n",
			statusSymbol(checkStartupRunEntry()),
			statusText(checkStartupRunEntry(), "HKCU Run key", "missing"))

		rightCol += fmt.Sprintf("  %s Task: %s\n",
			statusSymbol(checkTaskExists("BluefinCLI-ThemeModeSync")),
			statusText(checkTaskExists("BluefinCLI-ThemeModeSync"), "BluefinCLI-ThemeModeSync", "missing"))

		rightCol += fmt.Sprintf("  %s Task: %s\n",
			statusSymbol(checkTaskExists("BluefinCLI-SetLightAt6AM")),
			statusText(checkTaskExists("BluefinCLI-SetLightAt6AM"), "BluefinCLI-SetLightAt6AM", "missing"))

		rightCol += fmt.Sprintf("  %s Task: %s\n",
			statusSymbol(checkTaskExists("BluefinCLI-SetDarkAt6PM")),
			statusText(checkTaskExists("BluefinCLI-SetDarkAt6PM"), "BluefinCLI-SetDarkAt6PM", "missing"))

		if mode, ok := windowsThemeMode(); ok {
			rightCol += fmt.Sprintf("  %s Windows Mode: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render(mode))
		}
	}

	// Combine columns with padding
	formatted := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(40).Render(leftCol),
		string(rightCol),
	)

	fmt.Println(formatted)

	return nil
}

func checkTaskExists(taskName string) bool {
	if _, err := exec.LookPath("schtasks.exe"); err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "schtasks.exe", "/Query", "/TN", taskName)
	return cmd.Run() == nil
}

func checkStartupRunEntry() bool {
	if _, err := exec.LookPath("powershell.exe"); err != nil {
		return false
	}

	script := `$runPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"; $name = "BluefinCLIThemeModeSync"; $value = (Get-ItemProperty -Path $runPath -Name $name -ErrorAction SilentlyContinue).$name; if($value){ exit 0 } ; exit 1`
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Run() == nil
}

func statusSymbol(ok bool) string {
	if ok {
		return enabledStyle.Render("✓")
	}

	return disabledStyle.Render("✗")
}

func statusText(ok bool, successText, failureText string) string {
	if ok {
		return enabledStyle.Render(successText)
	}

	return disabledStyle.Render(failureText)
}

func windowsThemeMode() (string, bool) {
	if _, err := exec.LookPath("powershell.exe"); err != nil {
		return "", false
	}

	script := `$path = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize"; $value = (Get-ItemProperty -Path $path -Name AppsUseLightTheme -ErrorAction SilentlyContinue).AppsUseLightTheme; if ($null -eq $value) { exit 1 }; if ($value -eq 0) { Write-Output "🌙 Dark" } else { Write-Output "🌞 Light" }`
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	mode := strings.TrimSpace(string(out))
	if mode == "" {
		return "", false
	}

	return mode, true
}
