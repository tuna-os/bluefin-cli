package status

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/sunset"
)

// SunsetStatusProvider renders sunset automation status (Windows/WSL only).
type SunsetStatusProvider struct{}

func NewSunsetStatusProvider() *SunsetStatusProvider {
	return &SunsetStatusProvider{}
}

func (p *SunsetStatusProvider) Name() string {
	return "Sunset Automation"
}

func (p *SunsetStatusProvider) Render(width int) (string, error) {
	// Only render on Windows or WSL
	if !env.IsWindows() && !env.IsWSL() {
		return "", nil
	}

	var output string

	output += labelStyle.Render("Sunset Automation:") + "\n"
	if cfg, err := sunset.LoadConfig(); err == nil && cfg.Enabled {
		output += fmt.Sprintf("  %s Status: %s\n",
			enabledStyle.Render("✓"),
			enabledStyle.Render("enabled"))
		output += fmt.Sprintf("    Location: %.4f, %.4f\n", cfg.Latitude, cfg.Longitude)
		if cfg.WallpaperTheme != "" {
			output += fmt.Sprintf("    Theme: %s\n", cfg.WallpaperTheme)
		}
	} else {
		output += fmt.Sprintf("  %s Status: %s\n",
			disabledStyle.Render("✗"),
			disabledStyle.Render("disabled"))
		output += "    Run 'bluefin-cli sunset setup' to enable\n"
	}

	return output, nil
}
