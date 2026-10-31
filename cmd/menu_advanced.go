package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

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
