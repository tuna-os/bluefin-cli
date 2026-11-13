package wallpaper

import (
	"path/filepath"
	"strings"
)

// Collection is a wallpaper collection bluefin-cli can install and that
// sunset switching can draw monthly wallpapers from.
type Collection struct {
	// Name is the display name shown in menus ("Bluefin").
	Name string
	// ID is the lowercase identifier stored in the sunset config and used
	// as the collection's directory name ("bluefin").
	ID string
	// Cask is the Homebrew cask in the ublue-os tap ("bluefin-wallpapers").
	Cask string
}

// Collections is the catalogue of installable wallpaper collections, in
// menu order. Every installer, cleanup path and theme picker reads it, so a
// new collection is added here and nowhere else.
var Collections = []Collection{
	{Name: "Bluefin", ID: "bluefin", Cask: "bluefin-wallpapers"},
	{Name: "Aurora", ID: "aurora", Cask: "aurora-wallpapers"},
	{Name: "Bazzite", ID: "bazzite", Cask: "bazzite-wallpapers"},
}

// CollectionForCask returns the collection a cask name belongs to. The match
// is a case-insensitive substring match on the collection ID, so tap-qualified
// names and variant casks ("aurora-dynamic-wallpaper") resolve too.
func CollectionForCask(cask string) (Collection, bool) {
	name := strings.ToLower(cask)
	for _, c := range Collections {
		if strings.Contains(name, c.ID) {
			return c, true
		}
	}
	return Collection{}, false
}

// CollectionRoot is the directory bluefin-cli downloads wallpaper collections
// into on platforms without Homebrew casks (Windows).
func CollectionRoot(home string) string {
	return filepath.Join(home, "Pictures", "BluefinCLI")
}

// CollectionDir is the directory holding one collection's images under
// CollectionRoot. The installer writes it and sunset switching reads it, so
// both must call this rather than building the path themselves.
func CollectionDir(home, theme string) string {
	return filepath.Join(CollectionRoot(home), strings.ToLower(theme))
}
