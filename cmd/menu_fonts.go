package cmd

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

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
