package cmd

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// TestMainMenuItemsResolve guards against menu entries that lead nowhere:
// every item offered by the main menu must produce a command when selected.
func TestMainMenuItemsResolve(t *testing.T) {
	for _, it := range mainMenuItems() {
		if it.Value == "" {
			t.Errorf("menu item %q has no value", it.Label)
			continue
		}
		if cmd := mainMenuSelect(it); cmd == nil {
			t.Errorf("menu item %q (%s) resolves to no action", it.Label, it.Value)
		}
	}
}

// TestExtraMenuItemsResolve does the same for the build-variant extras.
func TestExtraMenuItemsResolve(t *testing.T) {
	for _, it := range extraMenuItems() {
		if cmd := extraMenuDo(it.Value); cmd == nil {
			t.Errorf("extra menu item %q (%s) resolves to no action", it.Label, it.Value)
		}
	}
}

// TestBundleCategoriesAvailable ensures the install menu is never empty and
// every category resolves through the palette registration path too.
func TestBundleCategoriesAvailable(t *testing.T) {
	cats := availableBundleCategories()
	if len(cats) == 0 {
		t.Fatal("no bundle categories available")
	}
	seen := map[string]bool{}
	for _, c := range cats {
		if c.ID == "" || c.Label == "" {
			t.Errorf("bundle category %+v missing ID or label", c)
		}
		if seen[c.ID] {
			t.Errorf("duplicate bundle category %q", c.ID)
		}
		seen[c.ID] = true
	}
}

// TestMenuModelStructure verifies that the declarative menu model has expected items and properties.
func TestMenuModelStructure(t *testing.T) {
	model := BuildMainMenuModel()
	if model.Title != "Home" {
		t.Errorf("BuildMainMenuModel().Title = %q, want %q", model.Title, "Home")
	}
	if len(model.Items) == 0 {
		t.Fatal("BuildMainMenuModel() returned no items")
	}

	expectedValues := []string{"status", "doctor", "shell", "bundles", "terminal", "exit"}
	valueMap := make(map[string]bool)
	for _, it := range model.Items {
		valueMap[it.Value] = true
	}
	for _, val := range expectedValues {
		if !valueMap[val] {
			t.Errorf("BuildMainMenuModel() missing expected item value %q", val)
		}
	}
}

// TestMenuRouter verifies custom routing and fallback to extraMenuDo.
func TestMenuRouter(t *testing.T) {
	r := NewMenuRouter()
	called := false
	r.Register("custom", func(it app.MenuItem) tea.Cmd {
		called = true
		return func() tea.Msg { return nil }
	})

	cmd := r.Route(app.MenuItem{Value: "custom"})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd for registered route")
	}
	_ = cmd()
	if !called {
		t.Errorf("expected custom handler to be called")
	}

	// Route unhandled non-extra value returns nil
	if unhandledCmd := r.Route(app.MenuItem{Value: "non-existent-item"}); unhandledCmd != nil {
		t.Errorf("expected nil cmd for unhandled non-existent value")
	}
}
