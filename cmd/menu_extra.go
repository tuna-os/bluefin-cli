//go:build extra

package cmd

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/sunset"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

func extraMenuItems() []app.MenuItem {
	items := []app.MenuItem{
		{Icon: "🖼 ", Label: "Wallpapers", Value: "wallpapers", Desc: "Bluefin wallpaper gallery", Submenu: true},
		{Icon: "🔤", Label: "Fonts", Value: "fonts", Desc: "Nerd Fonts for your terminal", Submenu: true},
		{Icon: "🚀", Label: "Starship Theme", Value: "starship", Desc: "Prompt presets & customization", Submenu: true},
	}
	if env.IsWSL() || env.IsWindows() {
		items = append(items, app.MenuItem{Icon: "🌇", Label: "Sunset Switching", Value: "sunset", Desc: "Auto light/dark by time of day", Submenu: true})
	}
	return items
}

func extraMenuDo(value string) tea.Cmd {
	switch value {
	case "wallpapers":
		return wallpapersFlow()
	case "fonts":
		return fontsFlow()
	case "starship":
		return starshipFlow()
	case "sunset":
		return sunsetFlow()
	}
	return nil
}

// sunsetFlow configures solar theme switching natively. On WSL the work is
// delegated to the Windows CLI — an external interactive program, so that
// path releases the terminal (the one legitimate use of the exec bridge).
func sunsetFlow() tea.Cmd {
	if env.IsWSL() {
		return app.RunExternal(RunSunsetSetupFlow)
	}
	if !env.IsWindows() {
		return app.Toast("Sunset switching is Windows/WSL only.", true)
	}

	var city, themeChoice string
	build := func() *huh.Form {
		return huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Enter your city").
				Placeholder("e.g. New York, London, Tokyo").
				Value(&city).
				Validate(func(s string) error {
					if len(s) < 2 {
						return fmt.Errorf("please enter a valid city name")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Choose a wallpaper theme").
				Options(
					huh.NewOption("Bluefin", "bluefin"),
					huh.NewOption("Aurora", "aurora"),
					huh.NewOption("Bazzite", "bazzite"),
					huh.NewOption("None (Keep current)", ""),
				).
				Value(&themeChoice),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.Push(app.NewForm("Sunset", build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		return tea.Batch(
			app.Toast("Searching for "+city+"…", false),
			func() tea.Msg {
				res, err := sunset.GeocodeCity(city)
				if err != nil {
					return app.ToastMsg{Text: "Could not resolve city: " + err.Error(), IsErr: true}
				}
				confirmed := false
				desc := fmt.Sprintf("%s, %s, %s (Lat: %.4f, Long: %.4f)",
					res.Name, res.Admin1, res.Country, res.Latitude, res.Longitude)
				cbuild := func() *huh.Form {
					confirmed = false
					return huh.NewForm(huh.NewGroup(
						huh.NewConfirm().
							Title("Use these coordinates?").
							Description(desc).
							Value(&confirmed),
					)).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())
				}
				screen := app.NewForm("Confirm", cbuild, func(aborted bool) tea.Cmd {
					if aborted || !confirmed {
						return nil
					}
					return app.Push(app.NewRunner("Enabling sunset switching", func() error {
						cfg, err := sunset.LoadConfig()
						if err != nil {
							return err
						}
						cfg.Latitude, cfg.Longitude = res.Latitude, res.Longitude
						cfg.WallpaperTheme = themeChoice
						cfg.Enabled = true
						if err := sunset.SaveConfig(cfg); err != nil {
							return err
						}
						fmt.Println("Configuration updated and feature enabled!")
						return sunset.Apply(cfg, sunset.NewThemeOperator(), nil)
					}))
				})
				return app.PushMsg{Screen: screen}
			},
		)
	}))
}
