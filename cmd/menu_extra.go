//go:build extra

package cmd

import (
	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

func addExtraMenuOptions(opts []huh.Option[string]) []huh.Option[string] {
	wallpapersLabel := "🖼  Wallpapers ❯"
	fontsLabel := "🔤 Fonts ❯"
	starshipLabel := "🚀 Starship Theme ❯"

	// Standard (Standard) section
	opts = append(opts, huh.NewOption(wallpapersLabel, "wallpapers"))
	opts = append(opts, huh.NewOption(fontsLabel, "fonts"))
	opts = append(opts, huh.NewOption(starshipLabel, "starship"))

	if env.IsWSL() || env.IsWindows() {
		sunsetLabel := "🌇 Sunset Switching ❯"
		opts = append(opts, huh.NewOption(sunsetLabel, "sunset"))
	}
	
	return opts
}

func handleExtraMenuChoice(choice string) (bool, error) {
	switch choice {
	case "wallpapers":
		return true, runWallpapersMenu()
	case "fonts":
		return true, runFontsMenu()
	case "starship":
		return true, runStarshipMenu()
	case "sunset":
		return true, RunSunsetSetupFlow()
	}
	return false, nil
}
