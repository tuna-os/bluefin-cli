//go:build extra

package cmd

import (
	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
)

func addExtraMenuOptions(opts []huh.Option[string]) []huh.Option[string] {
	opts = append(opts, huh.NewOption("🖼  Wallpapers ❯", "wallpapers"))
	opts = append(opts, huh.NewOption("🔤 Fonts ❯", "fonts"))
	opts = append(opts, huh.NewOption("🖥️  Configure Terminals ❯", "configure"))
	opts = append(opts, huh.NewOption("🚀 Starship Theme ❯", "starship"))

	if env.IsWSL() || env.IsWindows() {
		opts = append(opts, huh.NewOption("🌇 Sunset Switching ❯", "sunset"))
	}

	return opts
}

func handleExtraMenuChoice(choice string) (bool, error) {
	switch choice {
	case "wallpapers":
		return true, runWallpapersMenu()
	case "fonts":
		return true, runFontsMenu()
	case "configure":
		return true, runConfigureMenu()
	case "starship":
		return true, runStarshipMenu()
	case "sunset":
		return true, RunSunsetSetupFlow()
	}
	return false, nil
}
