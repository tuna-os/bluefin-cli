package sunset

// ThemeOperator defines the operations for changing system theme and wallpaper.
type ThemeOperator interface {
	SetTheme(isLight bool) error
	SetWallpaper(path string) error
}
