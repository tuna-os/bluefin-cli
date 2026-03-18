package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/hanthor/bluefin-cli/internal/tui/theme"
)

var (
	// Current Theme
	CurrentTheme = theme.DefaultTheme

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.PrimaryBorder).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(CurrentTheme.FaintBorder)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(CurrentTheme.SecondaryText).
			PaddingLeft(1)

	SuccessStyle = lipgloss.NewStyle().Foreground(CurrentTheme.SuccessText).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(CurrentTheme.ErrorText).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(CurrentTheme.WarningText)
	InfoStyle    = lipgloss.NewStyle().Foreground(CurrentTheme.InfoText)

	PopupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.ErrorText).
			Padding(1, 2).
			Align(lipgloss.Center)
)

type appTheme struct{}

func (t appTheme) Theme(isDark bool) *huh.Styles {
	return huh.ThemeCatppuccin(isDark)
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

// ClearScreen clears the terminal screen
func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// RenderHeader renders a consistent header for menus
func RenderHeader(title string, subtitle string) {
	fmt.Println(TitleStyle.Render(title))
	if subtitle != "" {
		fmt.Println(SubtitleStyle.Render(subtitle))
	}
	fmt.Println()
}

// Pause waits for user input before continuing
func Pause() {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Faint(true).Render("Press Enter to continue..."))
	_, _ = fmt.Scanln()
}
