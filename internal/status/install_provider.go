package status

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// InstallStatusProvider renders managed tools and package manager status.
type InstallStatusProvider struct{}

func NewInstallStatusProvider() *InstallStatusProvider {
	return &InstallStatusProvider{}
}

func (p *InstallStatusProvider) Name() string {
	return "Installation Status"
}

func (p *InstallStatusProvider) Render(width int) (string, error) {
	var output string

	// Tool dependencies
	output += labelStyle.Render("Managed Tools:") + "\n"
	deps := shell.CheckDependencies()

	var toolsToShow []shell.Tool
	if env.IsWindows() {
		toolsToShow = shell.ToolsForShell("powershell")
	} else {
		// Default to bash for tool filtering if we can't determine current shell
		toolsToShow = shell.ToolsForShell("bash")
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

		output += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			tool.Name,
			style.Render(status))
	}
	output += "\n"

	// Homebrew status
	output += labelStyle.Render("Package Manager:") + "\n"
	if env.IsWindows() {
		managers := install.AvailableWindowsManagers()
		if len(managers) == 0 {
			output += fmt.Sprintf("  %s Windows PMs: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("none detected"))
			output += "    Install winget, scoop, or chocolatey\n"
		} else {
			output += fmt.Sprintf("  %s Windows PMs: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render(strings.Join(managers, ", ")))
		}
	} else {
		if _, err := exec.LookPath("brew"); err == nil {
			output += fmt.Sprintf("  %s Homebrew: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render("installed"))

			ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
			defer cancel()
			if out, err := exec.CommandContext(ctx, "brew", "--version").Output(); err == nil {
				version := string(out)
				if len(version) > 0 {
					output += fmt.Sprintf("    %s\n", version[:len(version)-1])
				}
			}
		} else {
			output += fmt.Sprintf("  %s Homebrew: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("not installed"))
			output += "    Install from: https://brew.sh\n"
		}
	}

	return output, nil
}
