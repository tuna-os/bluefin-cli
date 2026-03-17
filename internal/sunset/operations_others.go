//go:build !windows

package sunset

import "fmt"

type stubThemeOperator struct{}

func (s *stubThemeOperator) SetTheme(isLight bool) error {
	return fmt.Errorf("theme switching is only supported on Windows")
}

func (s *stubThemeOperator) SetWallpaper(path string) error {
	return fmt.Errorf("wallpaper switching is only supported on Windows")
}

func NewThemeOperator() ThemeOperator {
	return &stubThemeOperator{}
}
