package cmd

import (
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// The standard binary and bluefin-cli-plus are built from mutually exclusive
// build tags (`!extra` and `extra`), and CI compiles both. Compilation alone
// never caught menu/command registration drifting between them
// (tuna-os/bluefin-cli#264), so the shared half of the contract is asserted
// here — this file carries no build tag and therefore runs in both variants.
// The variant-specific halves live in menu_vanilla_test.go and
// menu_extra_test.go.

// TestExtraMenuItems_SharedEntries pins the three entries both variants must
// register, in order. Plus may add to this list; it may not drop from it.
func TestExtraMenuItems_SharedEntries(t *testing.T) {
	items := extraMenuItems()
	want := []string{"wallpapers", "fonts", "starship"}

	if len(items) < len(want) {
		t.Fatalf("extraMenuItems() returned %d items, want at least %d: %+v", len(items), len(want), items)
	}
	for i, value := range want {
		if items[i].Value != value {
			t.Errorf("extraMenuItems()[%d].Value = %q, want %q", i, items[i].Value, value)
		}
	}
}

// TestExtraMenuItems_WellFormed guards against a half-filled entry reaching a
// menu: every registered item needs a label, a value and a description.
func TestExtraMenuItems_WellFormed(t *testing.T) {
	for i, item := range extraMenuItems() {
		if item.Label == "" || item.Value == "" || item.Desc == "" {
			t.Errorf("extraMenuItems()[%d] is incomplete: %+v", i, item)
		}
	}
}

// TestExtraMenuDo_HandlesEveryRegisteredItem is the registration contract that
// compilation cannot check: an item offered in the menu must be dispatchable.
// A value added to extraMenuItems without a matching extraMenuDo case fails
// here instead of silently doing nothing at runtime.
func TestExtraMenuDo_HandlesEveryRegisteredItem(t *testing.T) {
	for _, item := range extraMenuItems() {
		if extraMenuDo(item.Value) == nil {
			t.Errorf("extraMenuDo(%q) returned nil — menu item %q is registered but not dispatchable", item.Value, item.Label)
		}
	}
}

// TestExtraMenuDo_UnknownValue is the other half: an unregistered value must
// fall through to nil rather than dispatching something unrelated.
func TestExtraMenuDo_UnknownValue(t *testing.T) {
	if cmd := extraMenuDo("definitely-not-a-menu-entry"); cmd != nil {
		t.Error("extraMenuDo() should return nil for an unregistered value")
	}
}

// menuValues is a helper shared by the variant-specific contract tests.
func menuValues(items []app.MenuItem) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Value)
	}
	return values
}
