package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/registry"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// registerMenuActions registers all command palette actions into the registry.
// This function is called once at startup via registerPaletteActions().
func registerMenuActions() {
	// Home screen actions
	registry.Register(app.Action{
		ID: "status", Icon: "📊", Label: "Show Status", Section: "Home",
		Do: func() tea.Cmd { return mainMenuSelect(app.MenuItem{Value: "status"}) },
	})

	registry.Register(app.Action{
		ID: "doctor", Icon: "🩺", Label: "Doctor", Section: "Home",
		Do: func() tea.Cmd { return doctorScreenCmd() },
	})

	registry.Register(app.Action{
		ID: "terminal", Icon: "👻", Label: "Terminal Setup", Section: "Home",
		Do: func() tea.Cmd { return app.Push(terminalMenuScreen()) },
	})

	registry.Register(app.Action{
		ID: "update", Icon: "⬆", Label: "Check for Updates", Section: "Home",
		Do: func() tea.Cmd {
			return app.Push(app.NewRunner("Update", func() error { return runUpdate(false) }))
		},
	})

	// Fun actions
	registry.Register(app.Action{
		ID: "dino", Icon: "🦕", Label: "Dino Run", Section: "Fun",
		Do: func() tea.Cmd { return app.Push(gameScreen()) },
	})

	// Bundle install actions (dynamic)
	for _, cat := range availableBundleCategories() {
		id, label := cat.ID, cat.Label
		registry.Register(app.Action{
			ID: "install-" + id, Label: label, Section: "Install",
			Do: func() tea.Cmd { return packagesFlow(id, label) },
		})
	}

	// Extra menu actions (dynamic)
	for _, it := range extraMenuItems() {
		it := it
		registry.Register(app.Action{
			ID: it.Value, Icon: it.Icon, Label: it.Label, Section: "Customize",
			Do: func() tea.Cmd { return extraMenuDo(it.Value) },
		})
	}
}
