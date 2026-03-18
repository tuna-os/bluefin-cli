//go:build !extra

package cmd

import (
	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
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
		err := install.Bundle("fonts")
		if err == nil {
			err = maybeHandlePostFontInstall()
			tui.Pause()
		}
		return true, err
	case "starship":
		return true, runStarshipMenu()
	}
	return false, nil
}
