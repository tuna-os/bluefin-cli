package shell

import (
	"os"
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
