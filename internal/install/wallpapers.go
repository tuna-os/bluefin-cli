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
)

const wallpapersTap = "ublue-os/tap"

const (
	wallpaperUserInstallRootName = "BluefinCLI"
)

var knownWallpaperCasks = []string{
	"bluefin-wallpapers",
	"aurora-wallpapers",
	"bazzite-wallpapers",
}

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

func postInstallWallpaperSetup(casks []string) {
	// Legacy function, no-op now.
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
	for _, cask := range knownWallpaperCasks {
		args = append(args, wallpapersTap+"/"+cask)
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
		windowsDir := filepath.Join(homeDir, "Pictures", wallpaperUserInstallRootName)
		if err := os.RemoveAll(windowsDir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", windowsDir, err)
		}
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	dirs := []string{
		filepath.Join(homeDir, ".local", "share", "backgrounds", "bluefin"),
		filepath.Join(homeDir, ".local", "share", "backgrounds", "aurora"),
		filepath.Join(homeDir, ".local", "share", "backgrounds", "bazzite"),
		filepath.Join(homeDir, ".local", "share", "wallpapers", "bluefin"),
		filepath.Join(homeDir, ".local", "share", "wallpapers", "aurora"),
		filepath.Join(homeDir, ".local", "share", "wallpapers", "bazzite"),
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

func detectThemeName(cask string) (string, bool) {
	name := strings.ToLower(cask)
	switch {
	case strings.Contains(name, "bluefin"):
		return "Bluefin", true
	case strings.Contains(name, "aurora"):
		return "Aurora", true
	case strings.Contains(name, "bazzite"):
		return "Bazzite", true
	default:
		return "", false
	}
}

func ThemesFromWallpaperCasks(casks []string) []string {
	seen := map[string]struct{}{}
	themes := make([]string, 0, len(casks))

	for _, cask := range casks {
		normalized := normalizeCaskName(cask)
		themeName, ok := detectThemeName(normalized)
		if !ok {
			continue
		}
		if _, exists := seen[themeName]; exists {
			continue
		}
		seen[themeName] = struct{}{}
		themes = append(themes, themeName)
	}

	sort.Strings(themes)
	return themes
}


