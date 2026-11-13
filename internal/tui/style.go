package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// CurrentTheme and the shared styles below are derived from the active theme.
// They are rebuilt by ApplyTheme, so every writer of the theme must go
// through it; assigning theme.DefaultTheme directly leaves them stale.
var (
	// Current Theme
	CurrentTheme theme.Theme

	// Styles
	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style
	SuccessStyle  lipgloss.Style
	ErrorStyle    lipgloss.Style
	WarningStyle  lipgloss.Style
	InfoStyle     lipgloss.Style
	PopupStyle    lipgloss.Style
)

func init() {
	buildStyles(theme.DefaultTheme)
}

// ApplyTheme makes t the active theme: it sets theme.DefaultTheme and
// CurrentTheme and rebuilds the shared styles from it.
//
// Not safe to call while other goroutines render with the shared styles;
// callers apply it at startup or from the UI loop when no runner is active.
func ApplyTheme(t theme.Theme) {
	theme.DefaultTheme = t
	buildStyles(t)
}

func buildStyles(t theme.Theme) {
	CurrentTheme = t

	TitleStyle = lipgloss.NewStyle().
		Foreground(t.PrimaryBorder).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(t.FaintBorder)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(t.SecondaryText).
		PaddingLeft(1)

	SuccessStyle = lipgloss.NewStyle().Foreground(t.SuccessText).Bold(true)
	ErrorStyle = lipgloss.NewStyle().Foreground(t.ErrorText).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(t.WarningText)
	InfoStyle = lipgloss.NewStyle().Foreground(t.InfoText)

	PopupStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.ErrorText).
		Padding(1, 2).
		Align(lipgloss.Center)
}

type appTheme struct{}

// Theme styles huh forms with the app's semantic tokens: an accent bar on
// the left edge only, accent-colored cursors, and check/highlight selection
// instead of [ ] checkboxes. isDark comes from huh's own terminal
// background detection.
//
// Deliberately no full box: the check/cursor glyphs (✓ · ❯) are
// East-Asian-ambiguous width, and terminals that draw them double-wide
// would push a right border out of alignment on some rows. A left-only bar
// has no right edge to break.
func (t appTheme) Theme(isDark bool) *huh.Styles {
	tok := theme.Resolve(isDark)
	s := huh.ThemeCatppuccin(isDark)

	s.Focused.Base = lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(tok.Accent)
	s.Focused.Card = s.Focused.Base
	s.Blurred.Base = s.Focused.Base.BorderForeground(tok.Surface)
	s.Blurred.Card = s.Blurred.Base

	s.Focused.Title = s.Focused.Title.Foreground(tok.Accent).Bold(true)
	s.Focused.Description = s.Focused.Description.Foreground(tok.TextFaint)

	// Select: accent cursor, muted rest.
	s.Focused.SelectSelector = lipgloss.NewStyle().SetString("❯ ").Foreground(tok.Accent).Bold(true)
	s.Focused.Option = s.Focused.Option.Foreground(tok.TextMuted)

	// MultiSelect: colored highlights instead of [ ] checkboxes.
	s.Focused.MultiSelectSelector = lipgloss.NewStyle().SetString("❯ ").Foreground(tok.Accent).Bold(true)
	s.Focused.SelectedPrefix = lipgloss.NewStyle().SetString("✓ ").Foreground(tok.Success).Bold(true)
	s.Focused.UnselectedPrefix = lipgloss.NewStyle().SetString("· ").Foreground(tok.TextFaint)
	s.Focused.SelectedOption = lipgloss.NewStyle().Foreground(tok.Success).Bold(true)
	s.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(tok.TextMuted)

	s.Focused.FocusedButton = s.Focused.FocusedButton.Background(tok.Accent).Foreground(tok.Bg).Bold(true)
	s.Focused.BlurredButton = s.Focused.BlurredButton.Background(tok.Surface).Foreground(tok.TextMuted)

	return s
}

var (
	// Theme
	AppTheme huh.Theme = appTheme{}
)

// MenuKeyMap returns a keymap that includes standard navigation + Left/Backspace for abort (Back)
func MenuKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()

	// Global Quit/Back
	km.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "left", "backspace"),
		key.WithHelp("esc / ←", "back"),
	)

	// Select
	km.Select.Submit = key.NewBinding(
		key.WithKeys("enter", "right"),
		key.WithHelp("enter / →", "select"),
	)

	// MultiSelect
	km.MultiSelect.Submit = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)

	return km
}

// ConfirmKeyMap returns a keymap for confirmation dialogs — only Enter/Y/N, no left/back.
func ConfirmKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()

	km.Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)

	km.Confirm.Submit = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)

	km.Confirm.Accept = key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "yes"),
	)

	km.Confirm.Reject = key.NewBinding(
		key.WithKeys("n", "N"),
		key.WithHelp("n", "no"),
	)

	return km
}
