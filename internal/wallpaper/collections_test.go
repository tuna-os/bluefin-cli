package wallpaper

import (
	"path/filepath"
	"testing"
)

func TestCollectionForCask(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: "bluefin-wallpaper", want: "Bluefin", ok: true},
		{input: "ublue-os/tap/bluefin-wallpapers", want: "Bluefin", ok: true},
		{input: "aurora-dynamic-wallpaper", want: "Aurora", ok: true},
		{input: "Bazzite-Wallpapers", want: "Bazzite", ok: true},
		{input: "some-other-wallpaper", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := CollectionForCask(tt.input)
		if got.Name != tt.want || ok != tt.ok {
			t.Errorf("CollectionForCask(%q) = (%q, %v), want (%q, %v)", tt.input, got.Name, ok, tt.want, tt.ok)
		}
	}
}

// Every collection's cask must resolve back to that collection, or the
// Windows installer files it under the raw cask name, where sunset never
// looks.
func TestCollectionsCaskRoundTrip(t *testing.T) {
	for _, c := range Collections {
		got, ok := CollectionForCask(c.Cask)
		if !ok || got != c {
			t.Errorf("CollectionForCask(%q) = (%+v, %v), want (%+v, true)", c.Cask, got, ok, c)
		}
	}
}

// The on-disk layout is shared with collections installed by earlier
// releases; changing it strands them.
func TestCollectionDirLayout(t *testing.T) {
	home := filepath.Join("home", "user")
	want := filepath.Join(home, "Pictures", "BluefinCLI", "bluefin")
	for _, theme := range []string{"bluefin", "Bluefin", "BLUEFIN"} {
		if got := CollectionDir(home, theme); got != want {
			t.Errorf("CollectionDir(%q, %q) = %q, want %q", home, theme, got, want)
		}
	}
	if got, want := CollectionRoot(home), filepath.Join(home, "Pictures", "BluefinCLI"); got != want {
		t.Errorf("CollectionRoot(%q) = %q, want %q", home, got, want)
	}
}
