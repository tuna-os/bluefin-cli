// Package shellmenu provides native shell-related screens without depending on
// shell configuration, MOTD rendering, or Starship implementations.
package shellmenu

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// MOTD presents the current setting and delegates changes and previews to its
// caller. Toggle returns the new state, or an error if it could not be saved.
func MOTD(enabled func() bool, toggle func() (bool, error), show func() error) app.Screen {
	items := func() []app.MenuItem {
		label := "Enable MOTD"
		if enabled() {
			label = "Disable MOTD"
		}
		return []app.MenuItem{
			{Icon: "🔄", Label: label, Value: "toggle", Desc: "Message of the day on new terminals"},
			{Icon: "📰", Label: "Show MOTD", Value: "show", Desc: "Preview today's message"},
		}
	}
	return app.NewMenu("MOTD", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle":
			return tea.Sequence(motdToggleCmd(toggle), app.ReloadTop())
		case "show":
			return app.Push(app.NewRunner("MOTD", show))
		}
		return nil
	})
}

func motdToggleCmd(toggle func() (bool, error)) tea.Cmd {
	return func() tea.Msg {
		on, err := toggle()
		if err != nil {
			return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
		}
		if on {
			return app.ToastMsg{Text: "MOTD enabled."}
		}
		return app.ToastMsg{Text: "MOTD disabled."}
	}
}
