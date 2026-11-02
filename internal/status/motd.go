package status

import (
	"github.com/tuna-os/bluefin-cli/internal/motd"
)

func reportMOTD(shells []string) string {
	var out string
	out += labelStyle.Render("Message of the Day:") + "\n"
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
