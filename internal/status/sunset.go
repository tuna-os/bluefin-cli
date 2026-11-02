package status

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/sunset"
)

func reportSunset() string {
	if !env.IsWindows() && !env.IsWSL() {
		return ""
	}

	var out string
	out += labelStyle.Render("Sunset Automation:") + "\n"
	if cfg, err := sunset.LoadConfig(); err == nil && cfg.Enabled {
		out += report(true, "Status", "enabled")
		out += fmt.Sprintf("    Location: %.4f, %.4f\n", cfg.Latitude, cfg.Longitude)
		if cfg.WallpaperTheme != "" {
			out += fmt.Sprintf("    Theme: %s\n", cfg.WallpaperTheme)
		}
	} else {
		out += report(false, "Status", "disabled")
		out += "    Run 'bluefin-cli sunset setup' to enable\n"
	}
	return out
}
