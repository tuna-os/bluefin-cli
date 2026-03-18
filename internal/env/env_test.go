package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	// Mock HOME for consistent test
	tmpHome := t.TempDir()
	originalGetHomeDir := getHomeDir
	getHomeDir = func() (string, error) { return tmpHome, nil }
	defer func() { getHomeDir = originalGetHomeDir }()

	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	defer func() { _ = os.Unsetenv("HOME") }()

	// Mock HOMEBREW_PREFIX
	prefix := filepath.Join(tmpHome, "homebrew")
	if err := os.Setenv("HOMEBREW_PREFIX", prefix); err != nil {
		t.Fatalf("Failed to set HOMEBREW_PREFIX: %v", err)
	}
	defer func() { _ = os.Unsetenv("HOMEBREW_PREFIX") }()

	homeConfig := filepath.Join(tmpHome, ".config", "bluefin-cli")
	brewConfig := filepath.Join(prefix, "etc", "bluefin-cli")

	// 1. Test Default: No local config, HOMEBREW_PREFIX set
	// Should return Brew config
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	if dir != brewConfig {
		t.Errorf("Expected Brew config %s, got %s", brewConfig, dir)
	}

	// 2. Test Override: Local config exists
	// Create local config dir
	if err := os.MkdirAll(homeConfig, 0755); err != nil {
		t.Fatalf("Failed to create mock local config: %v", err)
	}

	dir, err = GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	// Should now prefer local config
	if dir != homeConfig {
		t.Errorf("Expected Local config %s (override), got %s", homeConfig, dir)
	}

	// 3. Test Fallback: No HOMEBREW_PREFIX
	if err := os.Unsetenv("HOMEBREW_PREFIX"); err != nil {
		t.Fatalf("Failed to unset HOMEBREW_PREFIX: %v", err)
	}
	dir, err = GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	if dir != homeConfig {
		t.Errorf("Expected Local config %s (fallback), got %s", homeConfig, dir)
	}
}

func TestIsWSL(t *testing.T) {
	originalGOOS := runtimeGOOS
	originalGetEnv := getEnv
	originalReadFile := readFile
	defer func() {
		runtimeGOOS = originalGOOS
		getEnv = originalGetEnv
		readFile = originalReadFile
	}()

	t.Run("returns false on non-linux", func(t *testing.T) {
		runtimeGOOS = "darwin"
		getEnv = func(key string) string { return "" }
		readFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }

		if IsWSL() {
			t.Fatal("expected IsWSL to be false on non-linux")
		}
	})

	t.Run("detects via wsl env vars", func(t *testing.T) {
		runtimeGOOS = "linux"
		getEnv = func(key string) string {
			switch key {
			case "WSL_DISTRO_NAME":
				return "Fedora"
			default:
				return ""
			}
		}
		readFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }

		if !IsWSL() {
			t.Fatal("expected IsWSL to be true with WSL env vars")
		}
	})

	t.Run("detects via proc markers", func(t *testing.T) {
		runtimeGOOS = "linux"
		getEnv = func(key string) string { return "" }
		readFile = func(name string) ([]byte, error) {
			if name == "/proc/sys/kernel/osrelease" {
				return []byte("5.15.153.1-microsoft-standard-WSL2"), nil
			}
			return nil, os.ErrNotExist
		}

		if !IsWSL() {
			t.Fatal("expected IsWSL to be true with microsoft proc marker")
		}
	})

	t.Run("fails closed when no markers", func(t *testing.T) {
		runtimeGOOS = "linux"
		getEnv = func(key string) string { return "" }
		readFile = func(name string) ([]byte, error) {
			return []byte("Linux version 6.8.0 Fedora"), nil
		}

		if IsWSL() {
			t.Fatal("expected IsWSL to be false without WSL markers")
		}
	})
}
