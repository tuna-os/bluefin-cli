//go:build !extra

package cmd

import (
	"slices"
	"testing"
)

// TestVanillaMenu_ExcludesExtraOnlyEntries pins the standard binary's half of
// the variant contract (tuna-os/bluefin-cli#264): sunset switching is a plus
// feature, and the standard menu must never offer it. Without this, moving an
// entry out of menu_extra.go into menu_vanilla.go compiles clean in both
// variants and ships a broken menu in the standard one, because the flows it
// dispatches to are stubbed out in extra_stubs.go.
func TestVanillaMenu_ExcludesExtraOnlyEntries(t *testing.T) {
	values := menuValues(extraMenuItems())
	for _, extraOnly := range []string{"sunset"} {
		if slices.Contains(values, extraOnly) {
			t.Errorf("standard menu offers plus-only entry %q: %v", extraOnly, values)
		}
	}
}

// TestVanillaMenu_IsExactlyTheSharedSet the standard variant registers the
// shared three entries and nothing else.
func TestVanillaMenu_IsExactlyTheSharedSet(t *testing.T) {
	got := menuValues(extraMenuItems())
	want := []string{"wallpapers", "fonts", "starship"}
	if !slices.Equal(got, want) {
		t.Errorf("standard extraMenuItems() = %v, want %v", got, want)
	}
}
