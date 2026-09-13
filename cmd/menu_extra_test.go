//go:build extra

package cmd

import (
	"slices"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

// TestExtraMenu_SunsetTracksPlatform pins bluefin-cli-plus's half of the
// variant contract (tuna-os/bluefin-cli#264): sunset switching is offered on
// Windows and WSL only, and only ever by the plus binary.
func TestExtraMenu_SunsetTracksPlatform(t *testing.T) {
	got := slices.Contains(menuValues(extraMenuItems()), "sunset")
	want := env.IsWSL() || env.IsWindows()
	if got != want {
		t.Errorf("plus menu offers sunset = %v, want %v (WSL=%v Windows=%v)",
			got, want, env.IsWSL(), env.IsWindows())
	}
}

// TestExtraMenu_IsASupersetOfStandard the plus binary must never register
// fewer entries than the standard one.
func TestExtraMenu_IsASupersetOfStandard(t *testing.T) {
	values := menuValues(extraMenuItems())
	for _, shared := range []string{"wallpapers", "fonts", "starship"} {
		if !slices.Contains(values, shared) {
			t.Errorf("plus menu is missing shared entry %q: %v", shared, values)
		}
	}
}
