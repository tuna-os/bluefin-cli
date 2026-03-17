package sunset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanthor/bluefin-cli/internal/env"
)

// Config represents sunset/theme/wallpaper settings.
type Config struct {
	Enabled        bool    `json:"enabled"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	DayWallpaper   string  `json:"day_wallpaper"`
	NightWallpaper string  `json:"night_wallpaper"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:   false,
		Latitude:  40.7128,  // NYC
		Longitude: -74.0060, // NYC
	}
}

// LoadConfig loads the sunset configuration from disk.
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sunset config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse sunset config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the sunset configuration to disk.
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create sunset config directory: %w", err)
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sunset config: %w", err)
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write sunset config: %w", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	dir, err := env.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "sunset.json"), nil
}
