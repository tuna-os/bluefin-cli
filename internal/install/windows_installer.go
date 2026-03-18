//go:build windows

package install

type WindowsInstaller struct{}

func init() {
	SetInstaller(&WindowsInstaller{})
}

func (i *WindowsInstaller) InstallBundle(nameOrPath string) error {
	return BundleWindows(nameOrPath)
}

func (i *WindowsInstaller) InstallWallpapers(casks []string) error {
	return installWallpaperCasksWindows(casks)
}

func (i *WindowsInstaller) CleanupWallpapers(all bool) error {
	if !all {
		return nil
	}

	return removeKnownLinuxWallpaperDirs() // This function handles Windows logic internally if GOOS is windows
}
