package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme defines the color palette for the application
type Theme struct {
	PrimaryBorder   color.Color
	SecondaryBorder color.Color
	FaintBorder     color.Color

	PrimaryText   color.Color
	SecondaryText color.Color
	FaintText     color.Color
	InvertedText  color.Color

	SuccessText color.Color
	WarningText color.Color
	ErrorText   color.Color
	InfoText    color.Color

	SelectedBackground color.Color
}

// DefaultTheme is the standard Catppuccin-like theme
var DefaultTheme = Theme{
	// Borders
	PrimaryBorder:   lipgloss.Color("39"),  // Blue
	SecondaryBorder: lipgloss.Color("241"), // Gray
	FaintBorder:     lipgloss.Color("238"), // Dark Gray

	// Text
	PrimaryText:   lipgloss.Color("15"),  // White
	SecondaryText: lipgloss.Color("247"), // Gray
	FaintText:     lipgloss.Color("241"), // Dark Gray
	InvertedText:  lipgloss.Color("0"),   // Black

	// Status
	SuccessText: lipgloss.Color("10"), // Green
	WarningText: lipgloss.Color("11"), // Yellow
	ErrorText:   lipgloss.Color("9"),  // Red
	InfoText:    lipgloss.Color("12"), // Cyan

	// Backgrounds
	SelectedBackground: lipgloss.Color("237"), // Dark Gray
}
