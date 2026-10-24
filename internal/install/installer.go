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

// MockInstaller is a test/mock implementation of Installer.
type MockInstaller struct {
	InstallBundleFunc     func(nameOrPath ...string) error
	InstallWallpapersFunc func(casks []string) error
	CleanupWallpapersFunc func(all bool) error
}

func (m *MockInstaller) InstallBundle(nameOrPath ...string) error {
	if m.InstallBundleFunc != nil {
		return m.InstallBundleFunc(nameOrPath...)
	}
	return nil
}

func (m *MockInstaller) InstallWallpapers(casks []string) error {
	if m.InstallWallpapersFunc != nil {
		return m.InstallWallpapersFunc(casks)
	}
	return nil
}

func (m *MockInstaller) CleanupWallpapers(all bool) error {
	if m.CleanupWallpapersFunc != nil {
		return m.CleanupWallpapersFunc(all)
	}
	return nil
}
