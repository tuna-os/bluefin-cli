package shell

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// SuccessStyle returns the lipgloss style for success messages.
func SuccessStyle() lipgloss.Style {
	return successStyle
}

// ErrorStyle returns the lipgloss style for error messages.
func ErrorStyle() lipgloss.Style {
	return errorStyle
}

// InfoStyle returns the lipgloss style for info messages.
func InfoStyle() lipgloss.Style {
	return infoStyle
}

// FormatToggleResult formats the human-readable result of a shell toggle operation.
func FormatToggleResult(shell string, enabled bool, changed bool) string {
	if enabled {
		if changed {
			return successStyle.Render(fmt.Sprintf("✓ Enabled shell experience for %s", shell))
		}
		return infoStyle.Render(fmt.Sprintf("%s is already enabled for %s", shell, shell))
	}

	if changed {
		return successStyle.Render(fmt.Sprintf("✓ Disabled shell experience for %s", shell))
	}
	return infoStyle.Render(fmt.Sprintf("%s is already disabled for %s", shell, shell))
}
