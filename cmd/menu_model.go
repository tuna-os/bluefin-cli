package cmd

import (
	"sort"
	"strings"

	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// MenuModel defines a data-driven menu structure with items, icons, and descriptions,
// decoupled from execution logic.
type MenuModel struct {
	Title string
	Items []app.MenuItem
}

// BuildMainMenuModel constructs the primary menu model for the Home screen.
func BuildMainMenuModel() MenuModel {
	var enabledShells []string
	for sh, enabled := range shell.CheckStatus() {
		if enabled {
			enabledShells = append(enabledShells, sh)
		}
	}
	sort.Strings(enabledShells)
	shellHint := "off"
	if len(enabledShells) > 0 {
		shellHint = strings.Join(enabledShells, " ") + " ✓"
	}

	items := []app.MenuItem{
		{Icon: "📊", Label: "Status", Value: "status", Desc: "Environment health at a glance"},
		{Icon: "🩺", Label: "Doctor", Value: "doctor", Desc: "Diagnose setup problems with fix hints"},
		{Icon: "🐚", Label: "Bluefin Shell", Value: "shell", Desc: "Aliases, prompt, and modern CLI tools", Hint: shellHint, Submenu: true},
		{Icon: "📦", Label: "Install Apps", Value: "bundles", Desc: "Curated tool bundles", Submenu: true},
		{Icon: "👻", Label: "Terminal Setup", Value: "terminal", Desc: "Ghostty install, Dock pin, themed config", Submenu: true},
	}
	items = append(items, extraMenuItems()...)
	items = append(items, app.MenuItem{Icon: "👋", Label: "Exit", Value: "exit", Desc: "Back to your shell"})

	return MenuModel{
		Title: "Home",
		Items: items,
	}
}
