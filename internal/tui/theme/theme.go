package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the application
type Theme struct {
	PrimaryBorder   lipgloss.AdaptiveColor
	SecondaryBorder lipgloss.AdaptiveColor
	FaintBorder     lipgloss.AdaptiveColor

	PrimaryText   lipgloss.AdaptiveColor
	SecondaryText lipgloss.AdaptiveColor
	FaintText     lipgloss.AdaptiveColor
	InvertedText  lipgloss.AdaptiveColor

	SuccessText lipgloss.AdaptiveColor
	WarningText lipgloss.AdaptiveColor
	ErrorText   lipgloss.AdaptiveColor
	InfoText    lipgloss.AdaptiveColor

	SelectedBackground lipgloss.AdaptiveColor
}

// DefaultTheme is the standard Catppuccin-like theme
var DefaultTheme = Theme{
	// Borders
	PrimaryBorder:   lipgloss.AdaptiveColor{Light: "39", Dark: "39"},   // Blue
	SecondaryBorder: lipgloss.AdaptiveColor{Light: "241", Dark: "241"}, // Gray
	FaintBorder:     lipgloss.AdaptiveColor{Light: "238", Dark: "238"}, // Dark Gray

	// Text
	PrimaryText:   lipgloss.AdaptiveColor{Light: "15", Dark: "15"},   // White
	SecondaryText: lipgloss.AdaptiveColor{Light: "247", Dark: "247"}, // Gray
	FaintText:     lipgloss.AdaptiveColor{Light: "241", Dark: "241"}, // Dark Gray
	InvertedText:  lipgloss.AdaptiveColor{Light: "0", Dark: "0"},     // Black

	// Status
	SuccessText: lipgloss.AdaptiveColor{Light: "10", Dark: "10"}, // Green
	WarningText: lipgloss.AdaptiveColor{Light: "11", Dark: "11"}, // Yellow
	ErrorText:   lipgloss.AdaptiveColor{Light: "9", Dark: "9"},   // Red
	InfoText:    lipgloss.AdaptiveColor{Light: "12", Dark: "12"}, // Cyan

	// Backgrounds
	SelectedBackground: lipgloss.AdaptiveColor{Light: "237", Dark: "237"}, // Dark Gray
}
