package sunset

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

const (
	spiSetDesktopWallpaper = 0x0014
	spifUpdateIniFile     = 0x01
	spifSendChange        = 0x02
)

// SetWindowsTheme toggles between light and dark mode on Windows.
// isLight = true sets Light mode, isLight = false sets Dark mode.
func SetWindowsTheme(isLight bool) error {
	val := uint32(0)
	if isLight {
		val = 1
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if err := key.SetDWordValue("AppsUseLightTheme", val); err != nil {
		return fmt.Errorf("failed to set AppsUseLightTheme: %w", err)
	}

	if err := key.SetDWordValue("SystemUsesLightTheme", val); err != nil {
		return fmt.Errorf("failed to set SystemUsesLightTheme: %w", err)
	}

	return nil
}

// SetWallpaper sets the desktop wallpaper on Windows.
func SetWallpaper(path string) error {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	ret, _, err := systemParametersInfo.Call(
		uintptr(spiSetDesktopWallpaper),
		0,
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(spifUpdateIniFile|spifSendChange),
	)

	if ret == 0 {
		return fmt.Errorf("SystemParametersInfoW failed: %w", err)
	}

	return nil
}
