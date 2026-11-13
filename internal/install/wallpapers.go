package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

const wallpapersTap = "ublue-os/tap"

func EnsureBrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found. Please install Homebrew first: https://brew.sh")
	}
	return nil
}

func ensureTap(tap string) error {
	if err := EnsureBrew(); err != nil {
		return err
	}
	cmd := exec.Command("brew", "tap", tap)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetWallpaperCasks() ([]string, error) {
	var casks []string
	if err := json.Unmarshal(embeddedWallpaperCasks, &casks); err != nil {
		return nil, fmt.Errorf("failed to read embedded wallpaper cask list: %w", err)
	}
	sort.Strings(casks)
	return casks, nil
}

func InstallWallpaperCasks(casks []string) error {
	return GetInstaller().InstallWallpapers(casks)
}

func CleanupWallpapers(all bool) error {
	return GetInstaller().CleanupWallpapers(all)
}

func uninstallKnownWallpaperCasks() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	if err := ensureTap(wallpapersTap); err != nil {
		return err
	}

	args := []string{"uninstall", "--cask"}
	for _, c := range wallpaper.Collections {
		args = append(args, wallpapersTap+"/"+c.Cask)
	}

	cmd := exec.Command("brew", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.ToLower(string(out))
		if strings.Contains(message, "is not installed") {
			return nil
		}
		return fmt.Errorf("failed to uninstall wallpaper casks: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func removeKnownLinuxWallpaperDirs() error {
	if runtime.GOOS == "windows" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home directory: %w", err)
		}
		windowsDir := wallpaper.CollectionRoot(homeDir)
		if err := os.RemoveAll(windowsDir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", windowsDir, err)
		}
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	var dirs []string
	for _, kind := range []string{"backgrounds", "wallpapers"} {
		for _, c := range wallpaper.Collections {
			dirs = append(dirs, filepath.Join(homeDir, ".local", "share", kind, c.ID))
		}
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
	}

	return nil
}

func normalizeCaskName(cask string) string {
	parts := strings.Split(cask, "/")
	return parts[len(parts)-1]
}
