package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMOTDController(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	ctrl := NewMOTDController()

	// Default state
	enabled := ctrl.IsEnabled("bash")
	if !enabled {
		t.Fatalf("expected MOTD to default to enabled")
	}

	// Toggle to false
	res, err := ctrl.Toggle("bash")
	if err != nil {
		t.Fatalf("toggle failed: %v", err)
	}
	if !res.Success || res.Data != false {
		t.Fatalf("expected res.Data=false, got %+v", res)
	}
	if ctrl.IsEnabled("bash") {
		t.Fatalf("expected MOTD to be disabled after toggle")
	}

	// Toggle back to true
	res, err = ctrl.Toggle("bash")
	if err != nil {
		t.Fatalf("toggle failed: %v", err)
	}
	if !res.Success || res.Data != true {
		t.Fatalf("expected res.Data=true, got %+v", res)
	}
	if !ctrl.IsEnabled("bash") {
		t.Fatalf("expected MOTD to be enabled after toggle")
	}

	// SetEnabled
	res, err = ctrl.SetEnabled([]string{"bash", "zsh"}, false)
	if err != nil {
		t.Fatalf("SetEnabled failed: %v", err)
	}
	if !res.Success || res.Data != false {
		t.Fatalf("expected res.Data=false, got %+v", res)
	}
	if ctrl.IsEnabled("bash") || ctrl.IsEnabled("zsh") {
		t.Fatalf("expected bash and zsh to be disabled")
	}
}

func TestShellController(t *testing.T) {
	ctrl := NewShellController()
	shells := ctrl.GetInstalledShells()
	_ = shells

	status := ctrl.CheckStatus()
	if status == nil {
		t.Fatalf("expected status map, got nil")
	}

	want := map[string]bool{"dummy": true}
	initial := map[string]bool{"dummy": false}
	changed, err := ctrl.ApplyChanges(want, initial, []string{"dummy"})
	// dummy shell will error on toggle or be handled, but let's test empty diff
	if err == nil && len(changed) == 0 {
		t.Fatalf("expected either error or change")
	}

	// zero changes
	changed, err = ctrl.ApplyChanges(want, want, []string{"dummy"})
	if err != nil || len(changed) != 0 {
		t.Fatalf("expected no changes and no error")
	}
}

func TestWallpaperController(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	ctrl := NewWallpaperController()
	_ = ctrl.Supported()
	imgs := ctrl.ListImages()
	_ = imgs

	testImg := filepath.Join(tmpDir, "test.png")
	_ = os.WriteFile(testImg, []byte("fake"), 0644)

	// Set wallpaper might fail if wallpaper CLI is not present, but should return a result or error
	_, _ = ctrl.SetWallpaper(testImg)
}

func TestStarshipController(t *testing.T) {
	ctrl := NewStarshipController()
	if ctrl == nil {
		t.Fatalf("expected non-nil StarshipController")
	}
}
