package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigData(t *testing.T) {
	// Setup temp home
	tmpHome := t.TempDir()
	if err := os.Setenv("HOMEBREW_PREFIX", tmpHome); err != nil {
		t.Fatalf("Failed to set mock HOMEBREW_PREFIX: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("HOMEBREW_PREFIX")
	}()

	// Test Default Config
	cfg := DefaultConfig("bash")
	if !cfg.IsEnabled("Eza") {
		t.Error("Default config should have Eza enabled")
	}

	// Test Save and Load
	cfg.SetEnabled("Eza", false)
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := LoadConfig("bash")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.IsEnabled("Eza") {
		t.Error("Expected Eza to be disabled after save")
	}
	if !loadedCfg.IsEnabled("Starship") {
		t.Error("Expected Starship to be enabled (unchanged)")
	}
}

func TestMOTDSetting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOMEBREW_PREFIX", "")
	if !MOTDEnabled("bash") {
		t.Fatal("MOTD should default to enabled")
	}
	cfg := DefaultConfig("bash")
	cfg.SetEnabled("Eza", false)
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	for _, want := range []bool{false, true} {
		got, err := ToggleMOTD("bash")
		if err != nil || got != want || MOTDEnabled("bash") != want {
			t.Fatalf("ToggleMOTD = %v, %v; want %v", got, err, want)
		}
	}
	loaded, err := LoadConfig("bash")
	if err != nil || loaded.IsEnabled("Eza") {
		t.Fatalf("toggle changed another setting: %v, %v", loaded, err)
	}

	path := filepath.Join(home, ".config", "bluefin-cli", "shell.json")
	if err := os.WriteFile(path, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}
	if !MOTDEnabled("bash") {
		t.Fatal("invalid config should use defaults")
	}
	if got, err := ToggleMOTD("bash"); err != nil || got {
		t.Fatalf("toggle should use default on invalid config: %v, %v", got, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := ToggleMOTD("bash"); err == nil {
		t.Fatal("saving to a directory must fail")
	}
}

func TestDefaultConfigPerShell(t *testing.T) {
	// bash should not have Gsudo enabled
	bashCfg := DefaultConfig("bash")
	if bashCfg.IsEnabled("Gsudo") {
		t.Error("Default bash config should have Gsudo disabled")
	}

	// powershell SHOULD have Gsudo enabled
	pwshCfg := DefaultConfig("powershell")
	if !pwshCfg.IsEnabled("Gsudo") {
		t.Error("Default powershell config should have Gsudo enabled")
	}

	// bash should have Ugrep enabled
	if !bashCfg.IsEnabled("Ugrep") {
		t.Error("Default bash config should have Ugrep enabled")
	}

	// powershell should NOT have Ugrep enabled
	if pwshCfg.IsEnabled("Ugrep") {
		t.Error("Default powershell config should have Ugrep disabled")
	}
}
