package commands

import (
	"path/filepath"

	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

// WallpaperController manages wallpaper discovery, installation, and desktop application.
type WallpaperController struct{}

// NewWallpaperController returns a new WallpaperController.
func NewWallpaperController() *WallpaperController {
	return &WallpaperController{}
}

// Supported returns true if setting wallpaper natively is supported on the current platform.
func (c *WallpaperController) Supported() bool {
	return wallpaper.Supported()
}

// ListImages returns available wallpaper image files discovered on the system.
func (c *WallpaperController) ListImages() []string {
	return wallpaper.List()
}

// SetWallpaper applies the given image path as the desktop wallpaper.
func (c *WallpaperController) SetWallpaper(path string) (Result, error) {
	if err := wallpaper.Set(path); err != nil {
		return Result{Success: false, Message: "Error: " + err.Error()}, err
	}
	return Result{
		Success: true,
		Message: "Wallpaper set: " + filepath.Base(path),
		Data:    path,
	}, nil
}

// GetWallpaperCasks returns the list of available wallpaper casks.
func (c *WallpaperController) GetWallpaperCasks() ([]string, error) {
	return install.GetWallpaperCasks()
}

// InstalledCasks returns a set of installed wallpaper casks.
func (c *WallpaperController) InstalledCasks() map[string]bool {
	return install.InstalledCasks()
}

// InstallCasks installs the specified wallpaper casks.
func (c *WallpaperController) InstallCasks(casks []string) error {
	return install.InstallWallpaperCasks(casks)
}

// CleanUp cleans up wallpaper files and sync state.
func (c *WallpaperController) CleanUp(all bool) error {
	return install.CleanupWallpapers(all)
}
