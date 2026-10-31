package cmd

import (
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

func wallpapersFlow() tea.Cmd {
	items := []app.MenuItem{
		{Icon: "📥", Label: "Install collections", Value: "install", Desc: "Curated wallpaper packs via Homebrew", Submenu: true},
	}
	if wallpaper.Supported() {
		items = append(items, app.MenuItem{Icon: "🖼", Label: "Set wallpaper", Value: "set", Desc: "Pick an image and apply it to the desktop", Submenu: true})
	}
	return app.Push(app.NewMenu("Wallpapers", items, nil, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "install":
			return installWallpapersFlow()
		case "set":
			return setWallpaperFlow()
		}
		return nil
	}))
}

func installWallpapersFlow() tea.Cmd {
	return tea.Batch(
		app.Toast("Loading wallpapers…", false),
		func() tea.Msg {
			casks, err := install.GetWallpaperCasks()
			if err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			if len(casks) == 0 {
				return app.ToastMsg{Text: "No wallpaper casks found in ublue-os/tap.", IsErr: true}
			}
			return app.PushMsg{Screen: wallpapersFormScreen(casks, install.InstalledCasks())}
		},
	)
}

// setWallpaperFlow lists installed images (collections, backgrounds dirs)
// and applies the chosen one natively.
func setWallpaperFlow() tea.Cmd {
	return tea.Batch(
		app.Toast("Scanning for images…", false),
		func() tea.Msg {
			imgs := wallpaper.List()
			if len(imgs) == 0 {
				return app.ToastMsg{Text: "No images found — install a collection first.", IsErr: true}
			}
			items := make([]app.MenuItem, 0, len(imgs))
			for _, p := range imgs {
				items = append(items, app.MenuItem{
					Label: filepath.Base(p),
					Value: p,
					Desc:  filepath.Dir(p),
				})
			}
			picker := app.NewWallpaperPicker("Set Wallpaper", items, func(it app.MenuItem) tea.Cmd {
				pick := it.Value
				return func() tea.Msg {
					if err := wallpaper.Set(pick); err != nil {
						return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
					}
					return app.ToastMsg{Text: "Wallpaper set: " + filepath.Base(pick)}
				}
			})
			return app.PushMsg{Screen: picker}
		},
	)
}

func wallpapersFormScreen(casks []string, installed map[string]bool) app.Screen {
	var selected []string
	build := func() *huh.Form {
		selected = selected[:0]
		opts := make([]huh.Option[string], 0, len(casks))
		for _, c := range casks {
			label := c
			if installed[c] {
				label += " (installed)"
			}
			opts = append(opts, huh.NewOption(label, c))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select wallpapers to install").
				Description("Space toggles, enter confirms.").
				Options(opts...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.NewForm("Wallpapers", build, func(aborted bool) tea.Cmd {
		if aborted || len(selected) == 0 {
			return nil
		}
		picks := append([]string(nil), selected...)
		run := app.NewRunner("Installing wallpapers", func() error {
			return install.InstallWallpaperCasks(picks)
		})
		if env.IsWSL() || env.IsWindows() {
			run = app.NewRunnerWithPost("Installing wallpapers", func() error {
				return install.InstallWallpaperCasks(picks)
			}, func() tea.Cmd {
				return app.Push(wallpaperSunsetPostInstallScreen())
			})
		}
		return app.Push(run)
	})
}

// wallpaperSunsetPostInstallScreen returns a confirm FormScreen that offers
// to run the sunset (solar theme/wallpaper switching) setup after wallpaper
// casks have been installed on Windows.
func wallpaperSunsetPostInstallScreen() app.Screen {
	var startSetup bool
	build := func() *huh.Form {
		startSetup = false
		return huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to configure solar-based theme and wallpaper switching now?").
				Description("This uses the new 'sunset' feature to automatically manage your desktop experience.").
				Value(&startSetup),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())
	}
	return app.NewForm("Sunset Setup", build, func(aborted bool) tea.Cmd {
		if aborted || !startSetup {
			return nil
		}
		return app.RunExternal(RunSunsetSetupFlow)
	})
}
