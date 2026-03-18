//go:build !extra

package cmd

import (
	"charm.land/huh/v2"
)

func addExtraMenuOptions(opts []huh.Option[string]) []huh.Option[string] {
	wallpapersLabel := "🖼  Wallpapers ❯"
	fontsLabel := "🔤 Fonts"
	starshipLabel := "🚀 Starship Theme ❯"

	opts = append(opts, huh.NewOption(wallpapersLabel, "wallpapers"))
	opts = append(opts, huh.NewOption(fontsLabel, "fonts"))
	opts = append(opts, huh.NewOption(starshipLabel, "starship"))
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
	}
	return false, nil
}
