package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hanthor/bluefin-cli/internal/env"
)

const wallpapersTap = "ublue-os/tap"

var knownWallpaperCasks = []string{
	"bluefin-wallpapers",
	"aurora-wallpapers",
	"bazzite-wallpapers",
}

var (
	isWSL                     = env.IsWSL
	syncWallpapersToWindowsWSL = syncWallpapersToWindows
)

func EnsureBrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install Homebrew first: https://brew.sh")
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
	if err := ensureTap(wallpapersTap); err != nil {
		return nil, err
	}

	cmd := exec.Command("brew", "--repository", wallpapersTap)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tap repository path: %w", err)
	}

	tapPath := strings.TrimSpace(string(out))
	casksDir := filepath.Join(tapPath, "Casks")

	entries, err := os.ReadDir(casksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read casks directory at %s: %w", casksDir, err)
	}

	var casks []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".rb") {
			caskName := strings.TrimSuffix(name, ".rb")
			if strings.Contains(strings.ToLower(caskName), "wallpaper") {
				casks = append(casks, caskName)
			}
		}
	}

	return casks, nil
}

func InstallWallpaperCasks(casks []string) error {
	if err := ensureTap(wallpapersTap); err != nil {
		return err
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks selected")
	}
	args := []string{"install", "--cask"}
	for _, c := range casks {
		if strings.Contains(c, "/") {
			args = append(args, c)
		} else {
			args = append(args, wallpapersTap+"/"+c)
		}
	}
	cmd := exec.Command("brew", args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_ENV_HINTS=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install wallpaper casks: %w", err)
	}
	fmt.Println(successStyle.Render("✓ Wallpaper casks installed!"))

	postInstallWallpaperSetup(casks)

	// macOS specific instructions
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		fmt.Println("\n" + infoStyle.Render("Wallpapers installed to: "+filepath.Join(home, "Library/Desktop Pictures")))
		fmt.Println(infoStyle.Render("To use: System Settings > Wallpaper > Add Folder"))
	}

	return nil
}

func postInstallWallpaperSetup(casks []string) {
	if !isWSL() {
		return
	}

	if err := syncWallpapersToWindowsWSL(casks); err != nil {
		fmt.Println(infoStyle.Render("WSL detected, but Windows wallpaper/theme sync could not be completed: " + err.Error()))
		fmt.Println(infoStyle.Render("Homebrew wallpaper installation succeeded; continuing without Windows sync."))
	}
}

func CleanupWallpapers(all bool) error {
	if isWSL() {
		if err := cleanupWindowsWallpaperSyncArtifacts(); err != nil {
			return err
		}
	}

	if !all {
		return nil
	}

	if err := uninstallKnownWallpaperCasks(); err != nil {
		return err
	}

	if err := removeKnownLinuxWallpaperDirs(); err != nil {
		return err
	}

	return nil
}

func uninstallKnownWallpaperCasks() error {
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
