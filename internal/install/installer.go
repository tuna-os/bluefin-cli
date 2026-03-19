package install

// Installer defines the interface for platform-specific installation operations.
type Installer interface {
	InstallBundle(nameOrPath ...string) error
	InstallWallpapers(casks []string) error
	CleanupWallpapers(all bool) error
}

var currentInstaller Installer

// SetInstaller sets the global installer instance.
func SetInstaller(i Installer) {
	currentInstaller = i
}

// GetInstaller returns the current global installer.
func GetInstaller() Installer {
	return currentInstaller
}
