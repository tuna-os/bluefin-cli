package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Init initializes the configuration system using Viper.
func Init() error {
	viper.SetEnvPrefix("BLUEFIN")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Standard config paths
	viper.AddConfigPath(filepath.Join(home, ".config", "bluefin-cli"))

	// Homebrew prefix path
	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" {
		viper.AddConfigPath(filepath.Join(prefix, "etc", "bluefin-cli"))
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("bundles.base_url", "https://raw.githubusercontent.com/projectbluefin/common/main/system_files")
	viper.SetDefault("bundles.default_path", "shared/usr/share/ublue-os/homebrew")
	viper.SetDefault("theme", "catppuccin")
	viper.SetDefault("ui.flavor", "auto")
	viper.SetDefault("ui.dark_mode", true)
}

// Save persists the current viper configuration to disk.
// If no config file has been loaded yet, it creates one in the primary config dir.
func Save() error {
	if viper.ConfigFileUsed() != "" {
		return viper.WriteConfig()
	}
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	return viper.WriteConfigAs(filepath.Join(configDir, "config.yaml"))
}

// GetConfigDir returns the primary configuration directory.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bluefin-cli"), nil
}

// Reset wraps viper.Reset for tests.
func Reset() {
	viper.Reset()
}

// GetTheme returns the configured theme name.
func GetTheme() string {
	return viper.GetString("theme")
}

// SetTheme sets the configured theme name.
func SetTheme(theme string) {
	viper.Set("theme", theme)
}

// GetUIFlavor returns the configured UI flavor (e.g. "auto", "latte", "frappe", "macchiato", "mocha").
func GetUIFlavor() string {
	return viper.GetString("ui.flavor")
}

// SetUIFlavor sets the configured UI flavor.
func SetUIFlavor(flavor string) {
	viper.Set("ui.flavor", flavor)
}

// GetUIDarkMode returns whether UI dark mode is enabled.
func GetUIDarkMode() bool {
	return viper.GetBool("ui.dark_mode")
}

// SetUIDarkMode sets the UI dark mode preference.
func SetUIDarkMode(enabled bool) {
	viper.Set("ui.dark_mode", enabled)
}

// GetBundlesBaseURL returns the base URL for downloading bundles.
func GetBundlesBaseURL() string {
	return viper.GetString("bundles.base_url")
}

// SetBundlesBaseURL sets the base URL for downloading bundles.
func SetBundlesBaseURL(url string) {
	viper.Set("bundles.base_url", url)
}

// GetBundlesDefaultPath returns the default path under the base URL for bundles.
func GetBundlesDefaultPath() string {
	return viper.GetString("bundles.default_path")
}

// SetBundlesDefaultPath sets the default path under the base URL for bundles.
func SetBundlesDefaultPath(path string) {
	viper.Set("bundles.default_path", path)
}

// GetGameHighScore returns the saved dino game high score.
func GetGameHighScore() int {
	return viper.GetInt("game.high_score")
}

// SetGameHighScore updates the saved dino game high score.
func SetGameHighScore(score int) {
	viper.Set("game.high_score", score)
}

// GetUpdateLastCheck returns the timestamp (unix seconds) of the last update check.
func GetUpdateLastCheck() int64 {
	return viper.GetInt64("update.last_check")
}

// SetUpdateLastCheck records the timestamp (unix seconds) of an update check.
func SetUpdateLastCheck(timestamp int64) {
	viper.Set("update.last_check", timestamp)
}

// GetProfileGistID returns the GitHub Gist ID used for profile sync.
func GetProfileGistID() string {
	return viper.GetString("profile.gist_id")
}

// SetProfileGistID sets the GitHub Gist ID used for profile sync.
func SetProfileGistID(id string) {
	viper.Set("profile.gist_id", id)
}
