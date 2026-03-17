package sunset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	currentTime = time.Now
	getHomeDir  = os.UserHomeDir
)

// GetMonthlyWallpaper returns the path to the wallpaper for the current month and state.
// theme: "Bluefin", "Aurora", "Bazzite"
// isDay: true for day, false for night
func GetMonthlyWallpaper(theme string, isDay bool) (string, error) {
	if theme == "" {
		return "", nil // No theme selected, can't find monthly wallpaper
	}

	home, err := getHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// BluefinCLI wallpapers are stored in ~/Pictures/BluefinCLI/<theme>
	wallpaperDir := filepath.Join(home, "Pictures", "BluefinCLI", strings.ToLower(theme))

	if _, err := os.Stat(wallpaperDir); os.IsNotExist(err) {
		return "", fmt.Errorf("wallpaper directory not found: %s", wallpaperDir)
	}

	now := currentTime()
	monthPrefix := fmt.Sprintf("%02d", int(now.Month()))
	suffix := "day"
	if !isDay {
		suffix = "night"
	}

	// Look for MM-day.jpg, MM-night.jpg (or png, etc)
	files, err := os.ReadDir(wallpaperDir)
	if err != nil {
		return "", fmt.Errorf("failed to read wallpaper directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := strings.ToLower(file.Name())
		if strings.HasPrefix(name, monthPrefix) && strings.Contains(name, suffix) {
			return filepath.Join(wallpaperDir, file.Name()), nil
		}
	}

	return "", fmt.Errorf("no monthly wallpaper found for %s in %s", suffix, wallpaperDir)
}
