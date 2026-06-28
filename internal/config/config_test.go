package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// ── Init tests ───────────────────────────────────────────────────────────────

func TestInit_SetsEnvPrefix(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() should not error in test environment: %v", err)
	}
	if viper.GetEnvPrefix() != "BLUEFIN" {
		t.Errorf("EnvPrefix = %s, want BLUEFIN", viper.GetEnvPrefix())
	}
}

func TestInit_SetsDefaults(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	tests := []struct {
		key  string
		want interface{}
	}{
		{"bundles.base_url", "https://raw.githubusercontent.com/projectbluefin/common/main/system_files"},
		{"bundles.default_path", "shared/usr/share/ublue-os/homebrew"},
		{"theme", "catppuccin"},
		{"ui.dark_mode", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := viper.Get(tt.key)
			if got != tt.want {
				t.Errorf("viper.Get(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestInit_ReadsEnvOverride(t *testing.T) {
	t.Setenv("BLUEFIN_THEME", "solarized")

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if got := viper.GetString("theme"); got != "solarized" {
		t.Errorf("theme after env override = %s, want solarized", got)
	}
}

func TestInit_EnvKeyReplace(t *testing.T) {
	// Verify that BLUEFIN_UI_DARK_MODE overrides ui.dark_mode
	t.Setenv("BLUEFIN_UI_DARK_MODE", "false")

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if got := viper.GetBool("ui.dark_mode"); got != false {
		t.Errorf("ui.dark_mode after env override = %v, want false", got)
	}
}

// ── GetConfigDir tests ───────────────────────────────────────────────────────

func TestGetConfigDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	expected := filepath.Join(tmpHome, ".config", "bluefin-cli")
	if dir != expected {
		t.Errorf("GetConfigDir() = %s, want %s", dir, expected)
	}
}

// ── Save tests ───────────────────────────────────────────────────────────────

func TestSave_CreatesDefaultConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	viper.Set("theme", "nord")

	if err := Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	configDir := filepath.Join(tmpHome, ".config", "bluefin-cli")
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file not created at %s: %v", configFile, err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "theme: nord") {
		t.Errorf("config should contain 'theme: nord', got:\n%s", content)
	}
}

func TestSave_CreatesConfigDir(t *testing.T) {
	// When config dir doesn't exist, Save() should create it
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	configDir := filepath.Join(tmpHome, ".config", "bluefin-cli")
	if _, err := os.Stat(configDir); err == nil {
		t.Fatal("config dir should not exist before Save()")
	}

	if err := Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	if _, err := os.Stat(configDir); err != nil {
		t.Errorf("config dir was not created by Save(): %v", err)
	}
}

func TestSave_PreservesExistingValues(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Save once with custom values
	viper.Set("theme", "catppuccin")
	viper.Set("bundles.default_path", "custom/path")
	if err := Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Re-init and check values are preserved
	viper.Reset()
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if got := viper.GetString("theme"); got != "catppuccin" {
		t.Errorf("theme after save/reload = %s, want catppuccin", got)
	}
	if got := viper.GetString("bundles.default_path"); got != "custom/path" {
		t.Errorf("bundles.default_path after save/reload = %s, want custom/path", got)
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────────

func TestInit_MissingConfigFile(t *testing.T) {
	// Init should not error when no config file exists (defaults should be used)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := Init(); err != nil {
		t.Fatalf("Init() should not error when no config file exists: %v", err)
	}

	// Default theme should still be set
	if got := viper.GetString("theme"); got != "catppuccin" {
		t.Errorf("theme with no config file = %s, want catppuccin", got)
	}
}

func TestGetConfigDir_NoHome(t *testing.T) {
	// Unsetting HOME should cause GetConfigDir to error
	t.Setenv("HOME", "")

	_, err := GetConfigDir()
	if err == nil {
		t.Error("GetConfigDir() should error when HOME is not set")
	}
}
