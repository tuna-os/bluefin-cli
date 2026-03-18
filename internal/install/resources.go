package install

import (
	"embed"
)

//go:embed resources/brewfiles/*.Brewfile
var EmbeddedBrewfiles embed.FS
