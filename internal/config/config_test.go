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
	viper.Reset()
	t.Cleanup(viper.Reset)
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

// setTestHome points the process home at dir on every platform:
// os.UserHomeDir consults HOME on unix and USERPROFILE on Windows.
func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestGetConfigDir(t *testing.T) {
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

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
	viper.Reset()
	t.Cleanup(viper.Reset)
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

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
	viper.Reset()
	t.Cleanup(viper.Reset)
	// When config dir doesn't exist, Save() should create it
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

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
	viper.Reset()
	t.Cleanup(viper.Reset)
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

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
	setTestHome(t, tmpHome)

	if err := Init(); err != nil {
		t.Fatalf("Init() should not error when no config file exists: %v", err)
	}

	// Default theme should still be set
	if got := viper.GetString("theme"); got != "catppuccin" {
		t.Errorf("theme with no config file = %s, want catppuccin", got)
	}
}

func TestGetConfigDir_NoHome(t *testing.T) {
	// Unsetting the home env should cause GetConfigDir to error
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := GetConfigDir()
	if err == nil {
		t.Error("GetConfigDir() should error when HOME is not set")
	}
}

// ── Typed accessors tests ───────────────────────────────────────────────────

func TestTypedAccessors(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Default values
	if got := GetTheme(); got != "catppuccin" {
		t.Errorf("GetTheme() = %q, want catppuccin", got)
	}
	if got := GetUIFlavor(); got != "auto" {
		t.Errorf("GetUIFlavor() = %q, want auto", got)
	}
	if got := GetUIDarkMode(); got != true {
		t.Errorf("GetUIDarkMode() = %v, want true", got)
	}
	if got := GetBundlesBaseURL(); got != "https://raw.githubusercontent.com/projectbluefin/common/main/system_files" {
		t.Errorf("GetBundlesBaseURL() = %q", got)
	}
	if got := GetBundlesDefaultPath(); got != "shared/usr/share/ublue-os/homebrew" {
		t.Errorf("GetBundlesDefaultPath() = %q", got)
	}
	if got := GetGameHighScore(); got != 0 {
		t.Errorf("GetGameHighScore() = %d, want 0", got)
	}
	if got := GetUpdateLastCheck(); got != 0 {
		t.Errorf("GetUpdateLastCheck() = %d, want 0", got)
	}
	if got := GetProfileGistID(); got != "" {
		t.Errorf("GetProfileGistID() = %q, want empty", got)
	}

	// Setters and Getters roundtrip
	SetTheme("nord")
	if got := GetTheme(); got != "nord" {
		t.Errorf("after SetTheme: GetTheme() = %q, want nord", got)
	}

	SetUIFlavor("mocha")
	if got := GetUIFlavor(); got != "mocha" {
		t.Errorf("after SetUIFlavor: GetUIFlavor() = %q, want mocha", got)
	}

	SetUIDarkMode(false)
	if got := GetUIDarkMode(); got != false {
		t.Errorf("after SetUIDarkMode: GetUIDarkMode() = %v, want false", got)
	}

	SetBundlesBaseURL("https://example.com/bundles")
	if got := GetBundlesBaseURL(); got != "https://example.com/bundles" {
		t.Errorf("after SetBundlesBaseURL: GetBundlesBaseURL() = %q", got)
	}

	SetBundlesDefaultPath("custom/path")
	if got := GetBundlesDefaultPath(); got != "custom/path" {
		t.Errorf("after SetBundlesDefaultPath: GetBundlesDefaultPath() = %q", got)
	}

	SetGameHighScore(42)
	if got := GetGameHighScore(); got != 42 {
		t.Errorf("after SetGameHighScore: GetGameHighScore() = %d, want 42", got)
	}

	SetUpdateLastCheck(1234567890)
	if got := GetUpdateLastCheck(); got != 1234567890 {
		t.Errorf("after SetUpdateLastCheck: GetUpdateLastCheck() = %d, want 1234567890", got)
	}

	SetProfileGistID("abc123gist")
	if got := GetProfileGistID(); got != "abc123gist" {
		t.Errorf("after SetProfileGistID: GetProfileGistID() = %q, want abc123gist", got)
	}
}
