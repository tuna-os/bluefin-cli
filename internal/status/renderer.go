package status

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// Renderer formats a status report into a styled terminal string.
type Renderer struct {
	width int
}

// NewRenderer creates a Renderer configured for the given terminal width.
func NewRenderer(width int) *Renderer {
	return &Renderer{width: width}
}

// Render formats the given Report into a styled string.
func (r *Renderer) Render(report Report) string {
	// --- Left Column ---
	var leftCol string

	// Shell status
	leftCol += labelStyle.Render("Shell Experience:") + "\n"
	if len(report.Shells) == 0 {
		leftCol += "  (no compatible shells found)\n"
	}

	for _, s := range report.Shells {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if s.Enabled {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		markers := ""
		if s.IsDefault && s.IsCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default, current)")
		} else if s.IsDefault {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" ★ (default)")
		} else if s.IsCurrent {
			markers = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(" ● (current)")
		}

		leftCol += fmt.Sprintf("  %s %s: %s%s\n",
			style.Render(symbol),
			s.Name,
			style.Render(status),
			markers)
	}
	leftCol += "\n"

	// MOTD status
	leftCol += labelStyle.Render("Message of the Day:") + "\n"
	for _, s := range report.Shells {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if report.MOTD[s.Name] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		leftCol += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			s.Name,
			style.Render(status))
	}

	// --- Right Column ---
	var rightCol string

	// Tool dependencies
	rightCol += labelStyle.Render("Managed Tools:") + "\n"
	for _, tool := range report.Tools {
		status := "not installed"
		style := disabledStyle
		symbol := "✗"

		if tool.Installed {
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

	// Package manager status
	rightCol += labelStyle.Render("Package Manager:") + "\n"
	if report.PackageManager.IsWindows {
		managers := report.PackageManager.WindowsManagers
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
		if report.PackageManager.BrewInstalled {
			rightCol += fmt.Sprintf("  %s Homebrew: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render("installed"))
			if report.PackageManager.BrewVersion != "" {
				rightCol += fmt.Sprintf("    %s\n", report.PackageManager.BrewVersion)
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
	if report.Sunset != nil {
		rightCol += labelStyle.Render("Sunset Automation:") + "\n"
		if report.Sunset.Enabled {
			rightCol += fmt.Sprintf("  %s Status: %s\n",
				enabledStyle.Render("✓"),
				enabledStyle.Render("enabled"))
			rightCol += fmt.Sprintf("    Location: %.4f, %.4f\n", report.Sunset.Latitude, report.Sunset.Longitude)
			if report.Sunset.WallpaperTheme != "" {
				rightCol += fmt.Sprintf("    Theme: %s\n", report.Sunset.WallpaperTheme)
			}
		} else {
			rightCol += fmt.Sprintf("  %s Status: %s\n",
				disabledStyle.Render("✗"),
				disabledStyle.Render("disabled"))
			rightCol += "    Run 'bluefin-cli sunset setup' to enable\n"
		}
	}

	var formatted string
	if r.width < 76 {
		formatted = leftCol + "\n" + rightCol
	} else {
		formatted = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(40).Render(leftCol),
			rightCol,
		)
	}

	flavor := report.Flavor
	if flavor == "" {
		flavor = "classic"
	}
	head := titleStyle.Render("Bluefin CLI Status") + "  " +
		labelStyle.Render("v"+report.AppVersion+" · flavor "+flavor)
	return head + "\n\n" + formatted
}
