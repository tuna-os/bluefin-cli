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
}

// GetConfigDir returns the primary configuration directory.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bluefin-cli"), nil
}
