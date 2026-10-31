package status

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

const commandTimeout = 1500 * time.Millisecond

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Underline(true)
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// report formats a single status entry with standard icons and styling.
func report(enabled bool, name, status string, markers ...string) string {
	style := disabledStyle
	symbol := "✗"
	if enabled {
		style = enabledStyle
		symbol = "✓"
	}
	marker := ""
	if len(markers) > 0 {
		marker = markers[0]
	}
	return fmt.Sprintf("  %s %s: %s%s\n",
		style.Render(symbol),
		name,
		style.Render(status),
		marker)
}

// Show prints the status report to stdout, sized to the terminal.
func Show(version string) error {
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || width <= 0 {
		width = 80
	}
	out, err := Render(width, version)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

// Render builds the status report for the given width and version. Two columns need
// ~76 cells; narrower widths stack vertically so the columns can't overlap.
func Render(width int, version string) (string, error) {
	info := getShellInfo()

	// --- Left Column ---
	leftCol := reportShellExperience(info) + "\n" + reportMOTD(info.shellsToShow)

	// --- Right Column ---
	rightCol := reportManagedTools(info.currentShell) + "\n" + reportPackageManager() + "\n"
	if sunset := reportSunset(); sunset != "" {
		rightCol += sunset
	}

	var formatted string
	if width < 76 {
		formatted = leftCol + "\n" + rightCol
	} else {
		formatted = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(40).Render(leftCol),
			rightCol,
		)
	}

	head := titleStyle.Render("Bluefin CLI Status") + "  " +
		labelStyle.Render("v"+version+" · flavor "+viper.GetString("ui.flavor"))
	return head + "\n\n" + formatted, nil
}
