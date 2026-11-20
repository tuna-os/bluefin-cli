package status

import (
	"github.com/tuna-os/bluefin-cli/internal/motd"
)

func reportMOTD(shells []string) string {
	var out string
	out += labelStyle.Render("Message of the Day:") + "\n"
	// uwelcome's one switch covers every shell.
	if motd.Uwelcome() {
		enabled := motd.UwelcomeEnabled()
		status := "disabled"
		if enabled {
			status = "enabled"
		}
		return out + report(enabled, "uwelcome", status)
	}
	motdStatus := motd.CheckStatus()
	for _, s := range shells {
		status := "disabled"
		if motdStatus[s] {
			status = "enabled"
		}

		out += report(motdStatus[s], s, status)
	}
	return out
}
