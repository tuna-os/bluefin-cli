//go:build extra

package cmd

import (
	"slices"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/platform"
)

// TestExtraMenu_SunsetTracksPlatform pins bluefin-cli-plus's half of the
// variant contract (tuna-os/bluefin-cli#264): sunset switching is offered on
// Windows and WSL only, and only ever by the plus binary.
func TestExtraMenu_SunsetTracksPlatform(t *testing.T) {
	got := slices.Contains(menuValues(extraMenuItems()), "sunset")
	want := platform.GetCapabilities().SupportsSunset
	if got != want {
		t.Errorf("plus menu offers sunset = %v, want %v (WSL=%v Windows=%v)",
			got, want, platform.IsWSL(), platform.IsWindows())
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
