// Package render formats shell configuration and toggle results for terminal output.
package render

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

// ToggleStatus formats the outcome of enabling or disabling shell integration.
func ToggleStatus(res shell.ToggleResult) string {
	if res.Changed {
		if res.Enabled {
			return tui.SuccessStyle.Render(fmt.Sprintf("✓ Enabled shell experience for %s", res.Shell))
		}
		return tui.SuccessStyle.Render(fmt.Sprintf("✓ Disabled shell experience for %s", res.Shell))
	}
	if res.Enabled {
		return tui.InfoStyle.Render(fmt.Sprintf("%s is already enabled for %s", res.Shell, res.Shell))
	}
	return tui.InfoStyle.Render(fmt.Sprintf("%s is already disabled for %s", res.Shell, res.Shell))
}

// UninstallSuccess formats the uninstallation completion message.
func UninstallSuccess() string {
	return tui.SuccessStyle.Render("✓ Bluefin shell setup uninstalled")
}
