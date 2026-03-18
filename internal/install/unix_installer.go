//go:build !windows

package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type UnixInstaller struct{}

func init() {
	SetInstaller(&UnixInstaller{})
}

func (i *UnixInstaller) InstallBundle(nameOrPath string) error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found. Please install Homebrew first: https://brew.sh")
	}

	brewfilePath, cleanup, err := GetBrewfile(nameOrPath)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := EnsureBbrew(); err != nil {
		return err
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("🍺 Opening %s in bbrew...", brewfilePath)))

	if err := RunBbrew(brewfilePath); err != nil {
		return fmt.Errorf("bbrew failed: %w", err)
	}

	return nil
}

func (i *UnixInstaller) InstallWallpapers(casks []string) error {
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

func (i *UnixInstaller) CleanupWallpapers(all bool) error {
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
