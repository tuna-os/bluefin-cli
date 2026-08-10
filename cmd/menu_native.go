package cmd

// Native in-shell flows for everything that used to hand over the terminal:
// selection UIs are FormScreens, read-only views are TextScreens, and
// anything that prints (brew installs, toggles with output) runs in a
// RunnerScreen that captures stdout into a log view.

import (
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/starship"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

// --- Shell submenu -----------------------------------------------------

func motdMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		cfg, err := shell.LoadConfig(currentShellName())
		if err != nil {
			cfg = shell.DefaultConfig(currentShellName())
		}
		toggle := "Enable MOTD"
		if cfg.IsEnabled("Motd") {
			toggle = "Disable MOTD"
		}
		return []app.MenuItem{
			{Icon: "🔄", Label: toggle, Value: "toggle", Desc: "Message of the day on new terminals"},
			{Icon: "📰", Label: "Show MOTD", Value: "show", Desc: "Preview today's message"},
		}
	}
	return app.NewMenu("MOTD", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle":
			return tea.Sequence(func() tea.Msg {
				sh := currentShellName()
				cfg, err := shell.LoadConfig(sh)
				if err != nil {
					cfg = shell.DefaultConfig(sh)
				}
				enable := !cfg.IsEnabled("Motd")
				cfg.SetEnabled("Motd", enable)
				if err := shell.SaveConfig(cfg); err != nil {
					return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
				}
				if enable {
					return app.ToastMsg{Text: "MOTD enabled."}
				}
				return app.ToastMsg{Text: "MOTD disabled."}
			}, app.ReloadTop())
		case "show":
			return app.Push(app.NewRunner("MOTD", motd.Show))
		}
		return nil
	})
}

func shellsFormScreen() app.Screen {
	var installed []string
	var selected []string
	initial := map[string]bool{}

	build := func() *huh.Form {
		installed = shell.GetInstalledShells()
		status := shell.CheckStatus()
		selected = selected[:0]
		options := make([]huh.Option[string], 0, len(installed))
		for _, sh := range installed {
			initial[sh] = status[sh]
			if status[sh] {
				selected = append(selected, sh)
			}
			options = append(options, huh.NewOption(sh, sh))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Manage other shells").
				Description("Selected = enabled. Space toggles, enter applies.").
				Options(options...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm("Other Shells", build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		want := map[string]bool{}
		for _, sh := range selected {
			want[sh] = true
		}
		changed := 0
		for _, sh := range installed {
			if want[sh] != initial[sh] {
				changed++
			}
		}
		if changed == 0 {
			return app.Toast("No changes.", false)
		}
		shells := append([]string(nil), installed...)
		return app.Push(app.NewRunner("Applying", func() error {
			for _, sh := range shells {
				if want[sh] != initial[sh] {
					fmt.Printf("%s -> %v\n", sh, want[sh])
					if err := shell.Toggle(sh, want[sh]); err != nil {
						return err
					}
				}
			}
			return nil
		}))
	})
}

// --- Extras ------------------------------------------------------------

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

func fontsFlow() tea.Cmd {
	var selected []string
	build := func() *huh.Form {
		selected = selected[:0]
		opts := make([]huh.Option[string], 0, len(availableFonts))
		for _, f := range availableFonts {
			opts = append(opts, huh.NewOption(f.Name, f.Cask))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select fonts to install").
				Description("Space toggles, enter confirms.").
				Options(opts...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.Push(app.NewForm("Fonts", build, func(aborted bool) tea.Cmd {
		if aborted || len(selected) == 0 {
			return nil
		}
		picks := append([]string(nil), selected...)
		return app.Push(app.NewRunner("Installing fonts", func() error {
			installFontCasks(picks)
			return maybeHandlePostFontInstall()
		}))
	}))
}

func starshipFlow() tea.Cmd {
	items := []app.MenuItem{
		{Icon: "◆", Label: "Nerd Font Symbols", Value: "nerd-font-symbols", Desc: "❯ ~/project  main ·  v1.22 — icons for every tool"},
		{Icon: "🌃", Label: "Tokyo Night", Value: "tokyo-night", Desc: "Neon segments on a midnight palette"},
		{Icon: "🌈", Label: "Gruvbox Rainbow", Value: "gruvbox-rainbow", Desc: "Warm retro powerline blocks"},
		{Icon: "🐈", Label: "Catppuccin Powerline", Value: "catppuccin-powerline", Desc: "Pastel powerline matching the app theme"},
		{Icon: "🚀", Label: "Jetpack", Value: "jetpack", Desc: "Dense, information-rich two-liner"},
		{Icon: "🎨", Label: "Pastel Powerline", Value: "pastel-powerline", Desc: "Soft colors, hard arrows"},
		{Icon: "✨", Label: "Pure Preset", Value: "pure-preset", Desc: "Minimal two-line prompt, zero noise"},
		{Icon: "🔤", Label: "Plain Text Symbols", Value: "plain-text-symbols", Desc: "ASCII-safe — no special font needed"},
		{Icon: "🔡", Label: "No Nerd Font", Value: "no-nerd-font", Desc: "Unicode-only, works in any terminal"},
		{Icon: "⏱", Label: "No Runtime Versions", Value: "no-runtime-versions", Desc: "Hides language version clutter"},
		{Icon: "▫️", Label: "No Empty Icons", Value: "no-empty-icons", Desc: "Icons only when they mean something"},
	}
	menu := app.NewMenu("Starship Theme", items, nil, func(it app.MenuItem) tea.Cmd {
		pick := it.Value
		return app.Push(app.NewRunner("Applying "+it.Label, func() error {
			if err := starship.Install(); err != nil {
				return err
			}
			return starship.ApplyTheme(pick)
		}))
	})
	return app.Push(menu)
}

// --- Advanced ----------------------------------------------------------

func advancedMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		dark := "🌙 Dark Mode: On"
		if !viper.GetBool("ui.dark_mode") {
			dark = "☀️  Dark Mode: Off"
		}
		return []app.MenuItem{
			{Label: dark, Value: "toggle_dark", Desc: "Color hint for plain CLI output"},
			{Icon: "🎨", Label: "Flavor: " + viper.GetString("ui.flavor"), Value: "flavor",
				Desc: "Cycle Catppuccin flavor (auto follows the terminal)"},
		}
	}
	return app.NewMenu("Advanced", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle_dark":
			return tea.Sequence(func() tea.Msg {
				viper.Set("ui.dark_mode", !viper.GetBool("ui.dark_mode"))
				if err := config.Save(); err != nil {
					return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
				}
				return app.ToastMsg{Text: "Saved."}
			}, app.ReloadTop())
		case "flavor":
			return tea.Sequence(func() tea.Msg {
				flavors := theme.Flavors()
				cur := viper.GetString("ui.flavor")
				next := flavors[0]
				for i, f := range flavors {
					if f == cur {
						next = flavors[(i+1)%len(flavors)]
						break
					}
				}
				viper.Set("ui.flavor", next)
				if err := config.Save(); err != nil {
					return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
				}
				theme.Flavor = ""
				if next != "auto" {
					theme.Flavor = next
				}
				theme.DefaultTheme = theme.Resolve(theme.DefaultTheme.IsDark)
				return app.ToastMsg{Text: "Flavor: " + next}
			}, app.ReloadTop())
		}
		return nil
	})
}

// --- Doctor & game wiring ----------------------------------------------

// doctorScreenCmd runs the doctor checks in a RunnerScreen (they hit the
// network, so they must not block the render loop).
func doctorScreenCmd() tea.Cmd {
	return app.Push(app.NewRunner("Doctor", func() error {
		report, _ := doctorReport()
		fmt.Println(report)
		return nil
	}))
}

// gameScreen builds the dino runner with a config-persisted high score.
func gameScreen() app.Screen {
	return app.NewGame(viper.GetInt("game.high_score"), func(n int) {
		viper.Set("game.high_score", n)
		_ = config.Save()
	})
}
