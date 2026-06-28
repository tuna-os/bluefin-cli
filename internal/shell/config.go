package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

type Config map[string]bool

func (c Config) IsEnabled(toolName string) bool {
	key := strings.ToLower(toolName)
	if enabled, ok := c[key]; ok {
		return enabled
	}
	// Special case: MOTD defaults to true
	if key == "motd" {
		return true
	}
	for _, t := range Tools {
		if t.Name == toolName {
			return t.Default
		}
	}
	return false
}

func (c Config) SetEnabled(toolName string, enabled bool) {
	key := strings.ToLower(toolName)
	c[key] = enabled
}

func DefaultConfig(shell string) *Config {
	cfg := make(Config)

	for _, tool := range Tools {
		if !tool.SupportsShell(shell) {
			cfg[strings.ToLower(tool.Name)] = false
			continue
		}

		def := tool.Default
		if val, ok := tool.ShellDefaults[shell]; ok {
			def = val
		}
		cfg[strings.ToLower(tool.Name)] = def
	}

	// MOTD is enabled by default (managed separately from tools)
	cfg["motd"] = true

	return &cfg
}

func LoadConfig(shell string) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(shell), nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	dir, err := env.GetConfigDir()
	if err != nil {
		return "", err
	}

	shellConfig := filepath.Join(dir, "shell.json")

	return shellConfig, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
