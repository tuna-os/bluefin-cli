package sunset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

// Config represents sunset/theme/wallpaper settings.
type Config struct {
	Enabled        bool    `json:"enabled"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	DayWallpaper   string  `json:"day_wallpaper"`
	NightWallpaper string  `json:"night_wallpaper"`
	WallpaperTheme string  `json:"wallpaper_theme"`
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

	config := &Config{}
	if err := config.LoadFrom(configPath); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves the sunset configuration to disk.
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	return config.SaveTo(configPath)
}

// SaveTo saves the configuration to a specific path.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// LoadFrom loads the configuration from a specific path.
func (c *Config) LoadFrom(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		*c = *DefaultConfig()
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := json.Unmarshal(content, c); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
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
