package shellmenu

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// Starship presents the preset picker; apply runs in a runner after selection.
func Starship(apply func(string) error) tea.Cmd {
	items := []app.MenuItem{
		{Icon: "◆", Label: "Nerd Font Symbols", Value: "nerd-font-symbols", Desc: "❯ ~/project  main ·  v1.22 — icons for every tool"},
		{Icon: "🌃", Label: "Tokyo Night", Value: "tokyo-night", Desc: "Neon segments on a midnight palette"},
		{Icon: "🌈", Label: "Gruvbox Rainbow", Value: "gruvbox-rainbow", Desc: "Warm retro powerline blocks"},
		{Icon: "🐈", Label: "Catppuccin Powerline", Value: "catppuccin-powerline", Desc: "Pastel powerline matching the app theme"},
		{Icon: "🚀", Label: "Jetpack", Value: "jetpack", Desc: "Dense, information-rich two-liner"},
		{Icon: "🎨", Label: "Pastel Powerline", Value: "pastel-powerline", Desc: "Soft colors, hard arrows"},
		{Icon: "✨", Label: "Pure Preset", Value: "pure-preset", Desc: "Minimal two-line prompt, zero noise"},
		{Icon: "🔤", Label: "Plain Text Symbols", Value: "plain-text-symbols", Desc: "ASCII-safe — no special font needed"},
		{Icon: "🔡", Label: "No Nerd Font", Value: "no-nerd-font", Desc: "Unicode-only, works in any terminal"},
		{Icon: "⏱", Label: "No Runtime Versions", Value: "no-runtime-versions", Desc: "Hides language version clutter"},
		{Icon: "▫️", Label: "No Empty Icons", Value: "no-empty-icons", Desc: "Icons only when they mean something"},
	}
	menu := app.NewMenu("Starship Theme", items, nil, func(it app.MenuItem) tea.Cmd {
		pick := it.Value
		return app.Push(app.NewRunner("Applying "+it.Label, func() error {
			return apply(pick)
		}))
	})
	return app.Push(menu)
}
