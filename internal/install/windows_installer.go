//go:build windows

package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

type WindowsInstaller struct{}

func init() {
	SetInstaller(&WindowsInstaller{})
}

func (i *WindowsInstaller) InstallBundle(nameOrPath string) error {
	return BundleWindows(nameOrPath)
}

func (i *WindowsInstaller) InstallWallpapers(casks []string) error {
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks selected")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	installRoot := filepath.Join(homeDir, "Pictures", wallpaperUserInstallRootName)
	if err := os.MkdirAll(installRoot, 0755); err != nil {
		return fmt.Errorf("failed to create wallpaper install directory: %w", err)
	}

	successChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	for _, c := range casks {
		normalized := normalizeCaskName(c)
		targetDir := ""

		done := make(chan struct{})
		go func() {
			defer close(done)
			archiveURL, err := fetchWallpaperArchiveURLFromTap(normalized)
			if err != nil {
				errChan <- err
				return
			}

			targetDirName := normalized
			if themeName, ok := detectThemeName(normalized); ok {
				targetDirName = strings.ToLower(themeName)
			}
			targetDir = filepath.Join(installRoot, targetDirName)

			if err := downloadAndExtractWallpaperArchive(archiveURL, targetDir); err != nil {
				errChan <- err
				return
			}
			successChan <- true
		}()

		// Throbber
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
				fmt.Printf("\r%s %s...", infoStyle.Foreground(lipgloss.Color("13")).Render(frames[i%len(frames)]), infoStyle.Render("Installing "+normalized))
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}

		select {
		case err := <-errChan:
			return fmt.Errorf("failed to install %s: %w", normalized, err)
		case <-successChan:
			fmt.Printf("\r\033[K") // Clear throbber line
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Installed %s wallpapers to %s", normalized, targetDir)))
		}
	}

	fmt.Println(infoStyle.Render("Wallpaper files downloaded without Homebrew."))
	fmt.Println(infoStyle.Render("You can apply them from Windows Settings > Personalization > Background."))
	return nil
}

func (i *WindowsInstaller) CleanupWallpapers(all bool) error {
	if !all {
		return nil
	}

	return removeKnownLinuxWallpaperDirs()
}
