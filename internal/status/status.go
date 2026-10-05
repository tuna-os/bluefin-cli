package status

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Underline(true)
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

// Render builds the status report for the given width using the StatusProvider registry.
func Render(width int) (string, error) {
	reg := NewRegistry()

	// Register all status providers in order.
	reg.Register(NewShellStatusProvider())
	reg.Register(NewMotdStatusProvider())
	reg.Register(NewInstallStatusProvider())
	reg.Register(NewSunsetStatusProvider())

	// Render all providers with their individual width awareness.
	content, err := reg.RenderAll(width)
	if err != nil {
		return "", err
	}

	// Add header with version info.
	head := titleStyle.Render("Bluefin CLI Status") + "  " +
		labelStyle.Render("v"+AppVersion+" · flavor "+viper.GetString("ui.flavor"))
	return head + "\n\n" + content, nil
}
