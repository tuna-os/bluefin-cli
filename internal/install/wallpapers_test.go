package install

import (
	"testing"
)

func TestNormalizeCaskName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "bluefin-wallpaper", want: "bluefin-wallpaper"},
		{input: "ublue-os/tap/bluefin-wallpaper", want: "bluefin-wallpaper"},
	}

	for _, tt := range tests {
		if got := normalizeCaskName(tt.input); got != tt.want {
			t.Fatalf("normalizeCaskName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDetectThemeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: "bluefin-wallpaper", want: "Bluefin", ok: true},
		{input: "aurora-dynamic-wallpaper", want: "Aurora", ok: true},
		{input: "bazzite-wallpaper", want: "Bazzite", ok: true},
		{input: "some-other-wallpaper", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := detectThemeName(tt.input)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("detectThemeName(%q) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

func TestThemesFromWallpaperCasks(t *testing.T) {
	casks := []string{
		"ublue-os/tap/bluefin-wallpapers",
		"aurora-wallpapers",
		"bazzite-wallpapers",
		"bluefin-wallpapers",
		"something-else",
	}

	themes := ThemesFromWallpaperCasks(casks)
	if len(themes) != 3 {
		t.Fatalf("expected 3 themes, got %d (%v)", len(themes), themes)
	}
}
