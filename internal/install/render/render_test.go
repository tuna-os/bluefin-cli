package render

import (
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/install"
)

func TestBundlesText(t *testing.T) {
	bundles := map[string]install.BundleSpec{
		"cli": {File: "cli.Brewfile", Description: "CLI essentials"},
		"ai":  {File: "ai.Brewfile", Description: "AI tools"},
	}

	out := BundlesText(bundles)
	if !strings.Contains(out, "📦 Available Bundles") {
		t.Fatalf("expected title in output: %s", out)
	}
	if !strings.Contains(out, "cli:") || !strings.Contains(out, "CLI essentials") {
		t.Fatalf("expected cli bundle in output: %s", out)
	}
	if !strings.Contains(out, "ai:") || !strings.Contains(out, "AI tools") {
		t.Fatalf("expected ai bundle in output: %s", out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("expected Usage in output: %s", out)
	}
}

func TestWallpaperSuccess(t *testing.T) {
	out := WallpaperSuccess()
	if !strings.Contains(out, "✓ Wallpaper casks installed!") {
		t.Fatalf("unexpected wallpaper success: %s", out)
	}
}

func TestDarwinWallpaperInstructions(t *testing.T) {
	out := DarwinWallpaperInstructions("/Users/test")
	if !strings.Contains(out, "/Users/test/Library/Desktop Pictures") {
		t.Fatalf("expected home path in instructions: %s", out)
	}
	if !strings.Contains(out, "System Settings > Wallpaper > Add Folder") {
		t.Fatalf("expected System Settings hint: %s", out)
	}
}
