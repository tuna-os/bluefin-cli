// Package render formats package and bundle installation output for terminal display.
package render

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// BundlesText formats the list of available bundles with terminal styles.
func BundlesText(bundles map[string]install.BundleSpec) string {
	t := theme.DefaultTheme
	titleStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(t.Success).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(t.Info)

	var b strings.Builder
	b.WriteString(titleStyle.Render("📦 Available Bundles") + "\n\n")

	names := make([]string, 0, len(bundles))
	for name := range bundles {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Fprintf(&b, "  %s %s\n", nameStyle.Render(name+":"), bundles[name].Description)
	}

	b.WriteString("\n" + infoStyle.Render("Usage:") + "\n")
	b.WriteString("  bluefin-cli install <bundle-name>\n")
	b.WriteString("  bluefin-cli install /path/to/Brewfile\n")

	return b.String()
}

// WallpaperSuccess formats the wallpaper installation completion message.
func WallpaperSuccess() string {
	return tui.SuccessStyle.Render("✓ Wallpaper casks installed!")
}

// DarwinWallpaperInstructions formats the macOS-specific wallpaper instructions.
func DarwinWallpaperInstructions(home string) string {
	var b strings.Builder
	b.WriteString("\n" + tui.InfoStyle.Render("Wallpapers installed to: "+home+"/Library/Desktop Pictures") + "\n")
	b.WriteString(tui.InfoStyle.Render("To use: System Settings > Wallpaper > Add Folder") + "\n")
	return b.String()
}
