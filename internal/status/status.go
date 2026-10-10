package status

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Underline(true)
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

// Show prints the status report to stdout, sized to the terminal.
func Show() error {
	width, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || width <= 0 {
		width = 80
	}
	out, err := Render(width)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

// AppVersion is stamped by cmd at startup so the report can show it.
var AppVersion = "dev"

// Render builds the status report for the given width using SystemCollector and Renderer.
func Render(width int) (string, error) {
	collector := NewSystemCollector()
	renderer := NewRenderer(width)
	report := collector.Collect()
	return renderer.Render(report), nil
}
