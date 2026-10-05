package status

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// MotdStatusProvider renders message-of-the-day status per shell.
type MotdStatusProvider struct{}

func NewMotdStatusProvider() *MotdStatusProvider {
	return &MotdStatusProvider{}
}

func (p *MotdStatusProvider) Name() string {
	return "Message of the Day"
}

func (p *MotdStatusProvider) Render(width int) (string, error) {
	var output string

	output += labelStyle.Render("Message of the Day:") + "\n"
	motdStatus := motd.CheckStatus()
	installed := shell.GetInstalledShells()

	// Show MOTD status for installed shells only.
	for _, s := range installed {
		status := "disabled"
		style := disabledStyle
		symbol := "✗"

		if motdStatus[s] {
			status = "enabled"
			style = enabledStyle
			symbol = "✓"
		}

		output += fmt.Sprintf("  %s %s: %s\n",
			style.Render(symbol),
			s,
			style.Render(status))
	}

	return output, nil
}
