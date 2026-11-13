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
