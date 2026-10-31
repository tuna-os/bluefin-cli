package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// MenuHandler maps a menu item selection to a Bubble Tea command.
type MenuHandler func(app.MenuItem) tea.Cmd

// MenuRouter maps menu selection values to their respective handlers.
type MenuRouter struct {
	routes map[string]MenuHandler
}

// NewMenuRouter creates an empty MenuRouter.
func NewMenuRouter() *MenuRouter {
	return &MenuRouter{
		routes: make(map[string]MenuHandler),
	}
}

// Register binds a menu item value to a handler function.
func (r *MenuRouter) Register(value string, handler MenuHandler) {
	r.routes[value] = handler
}

// Route dispatches the menu selection to the registered handler.
func (r *MenuRouter) Route(it app.MenuItem) tea.Cmd {
	if handler, ok := r.routes[it.Value]; ok {
		return handler(it)
	}
	return extraMenuDo(it.Value)
}

// defaultMenuRouter wires default handlers for the main menu items.
func defaultMenuRouter() *MenuRouter {
	r := NewMenuRouter()
	r.Register("status", func(_ app.MenuItem) tea.Cmd {
		return statusScreenCmd()
	})
	r.Register("doctor", func(_ app.MenuItem) tea.Cmd {
		return doctorScreenCmd()
	})
	r.Register("shell", func(_ app.MenuItem) tea.Cmd {
		return app.Push(shellMenuScreen())
	})
	r.Register("bundles", func(_ app.MenuItem) tea.Cmd {
		return app.Push(bundlesMenuScreen())
	})
	r.Register("terminal", func(_ app.MenuItem) tea.Cmd {
		return app.Push(terminalMenuScreen())
	})
	r.Register("exit", func(_ app.MenuItem) tea.Cmd {
		return app.Pop()
	})
	return r
}
