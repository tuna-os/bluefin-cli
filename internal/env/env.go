package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	runtimeGOOS = runtime.GOOS
	getEnv      = os.Getenv
	readFile    = os.ReadFile
)

// GetConfigDir returns the directory where configuration files should be stored.
// It prefers ~/.config/bluefin-cli if it exists (override).
// Otherwise, it defaults to $HOMEBREW_PREFIX/etc/bluefin-cli if HOMEBREW_PREFIX is set.
// Finally, it falls back to ~/.config/bluefin-cli.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	homeConfig := filepath.Join(home, ".config", "bluefin-cli")

	// 1. Prefer user config if it explicitly exists
	if _, err := os.Stat(homeConfig); err == nil {
		return homeConfig, nil
	}

	// 2. Default to Homebrew prefix if set
	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" {
		return filepath.Join(prefix, "etc", "bluefin-cli"), nil
	}

	// 3. Fallback to user home config
	return homeConfig, nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() (string, error) {
	path, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory at %s: %w", path, err)
	}

	return path, nil
}

// IsWSL returns true when running under Windows Subsystem for Linux.
// The check is fail-closed: if we cannot confidently detect WSL, it returns false.
func IsWSL() bool {
	if runtimeGOOS != "linux" {
		return false
	}

	if getEnv("WSL_DISTRO_NAME") != "" || getEnv("WSL_INTEROP") != "" {
		return true
	}

	if hasMicrosoftMarker("/proc/sys/kernel/osrelease") {
		return true
	}

	if hasMicrosoftMarker("/proc/version") {
		return true
	}

	return false
}

func hasMicrosoftMarker(path string) bool {
	data, err := readFile(path)
	if err != nil {
		return false
	}

	content := strings.ToLower(string(data))
	return strings.Contains(content, "microsoft") || strings.Contains(content, "wsl")
}
