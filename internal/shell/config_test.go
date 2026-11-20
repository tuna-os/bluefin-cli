package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	// Ugrep and Atuin follow Bluefin's cli bundle, which installs neither:
	// ugrep is aliased once present, atuin only when switched on.
	for _, tool := range []string{"Ugrep", "Atuin"} {
		if bashCfg.IsEnabled(tool) {
			t.Errorf("Default bash config should have %s disabled", tool)
		}
	}

	// powershell should NOT have Ugrep enabled
	if pwshCfg.IsEnabled("Ugrep") {
		t.Error("Default powershell config should have Ugrep disabled")
	}
}

// Finding a tool's binary turns it on, except for opt-in tools: Bluefin
// dropped atuin's shell integration, so having atuin installed must not
// hook it.
func TestInitSyncSkipsOptInTools(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("HOMEBREW_PREFIX", home)

	bin := t.TempDir()
	for _, name := range []string{"atuin", "direnv", "ug"} {
<<<<<<< Updated upstream
		if runtime.GOOS == "windows" {
			name += ".exe" // LookPath only finds PATHEXT extensions on Windows
||||||| Stash base
=======
		// exec.LookPath on Windows only finds names with a PATHEXT suffix.
		if runtime.GOOS == "windows" {
			name += ".exe"
>>>>>>> Stashed changes
		}
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)

	cfg := DefaultConfig("bash")
	script, err := Init("bash", cfg)
	if err != nil {
		t.Fatal(err)
	}
	for tool, want := range map[string]bool{"Atuin": false, "Direnv": true, "Ugrep": true} {
		if got := cfg.IsEnabled(tool); got != want {
			t.Errorf("after sync %s enabled = %v, want %v", tool, got, want)
		}
	}
	if !strings.Contains(script, "export BLUEFIN_SHELL_ENABLE_ATUIN=0") {
		t.Error("init script turned atuin on although it is opt-in")
	}
}
