package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDiff(t *testing.T) {
	current := &Profile{
		Version:       1,
		Flavor:        "auto",
		Tools:         map[string]bool{"eza": true, "bat": false},
		EnabledShells: []string{"bash", "zsh"},
	}
	want := &Profile{
		Version:       1,
		Flavor:        "mocha",
		Tools:         map[string]bool{"eza": true, "bat": true},
		EnabledShells: []string{"bash", "fish"},
	}

	diff := Diff(current, want)
	joined := strings.Join(diff, "\n")
	for _, expect := range []string{
		"flavor: auto -> mocha",
		"tool bat: off -> on",
		"shell fish: disabled -> enabled",
		"shell zsh: enabled -> disabled",
	} {
		if !strings.Contains(joined, expect) {
			t.Errorf("diff missing %q:\n%s", expect, joined)
		}
	}
	if strings.Contains(joined, "eza") || strings.Contains(joined, "bash") {
		t.Errorf("diff contains unchanged entries:\n%s", joined)
	}

	if d := Diff(want, want); len(d) != 0 {
		t.Errorf("self-diff should be empty, got %v", d)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-profile.json")

	p := &Profile{
		Version:       1,
		Flavor:        "latte",
		Tools:         map[string]bool{"eza": true},
		EnabledShells: []string{"bash"},
	}

	if err := p.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != p.Version || loaded.Flavor != p.Flavor || !loaded.Tools["eza"] || len(loaded.EnabledShells) != 1 {
		t.Errorf("loaded profile mismatch: %+v", loaded)
	}
}

func TestLoadErrors(t *testing.T) {
	dir := t.TempDir()

	// Non-existent file
	if _, err := Load(filepath.Join(dir, "nonexistent.json")); err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}

	// Invalid JSON
	invalidPath := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}
	if _, err := Load(invalidPath); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	// Unsupported version
	badVerPath := filepath.Join(dir, "bad_version.json")
	if err := os.WriteFile(badVerPath, []byte(`{"version": 99}`), 0o644); err != nil {
		t.Fatalf("failed to write bad version file: %v", err)
	}
	if _, err := Load(badVerPath); err == nil || !strings.Contains(err.Error(), "unsupported profile version") {
		t.Errorf("expected unsupported profile version error, got: %v", err)
	}
}

func TestExportAndApply(t *testing.T) {
	dir := t.TempDir()
	// t.Setenv restores HOME automatically and fails the test if it is ever
	// run in parallel, which this must not be: HOME and viper are process
	// globals.
	t.Setenv("HOME", dir)
	// Apply() does viper.Set + config.Save(). Viper resolves its config file
	// path once, globally, so without this reset Save() writes through a path
	// left over from an earlier Init -- i.e. into the developer's real
	// ~/.config/bluefin-cli/config.yaml rather than the temp HOME above.
	viper.Reset()
	t.Cleanup(viper.Reset)

	prof, err := Export("bash")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if prof.Version != 1 {
		t.Errorf("expected Version=1, got %d", prof.Version)
	}

	// Test Save to "-"
	if err := prof.Save("-"); err != nil {
		t.Errorf("Save to stdout failed: %v", err)
	}

	prof.Flavor = "frappe"
	prof.Tools = map[string]bool{"eza": true}
	if err := prof.Apply(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
}
