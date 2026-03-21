package install

import (
	"embed"
)

//go:embed resources/brewfiles/*.Brewfile
var EmbeddedBrewfiles embed.FS

//go:embed resources/wallpaper-casks.json
var embeddedWallpaperCasks []byte
