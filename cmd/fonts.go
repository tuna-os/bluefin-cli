package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var fontsCmd = &cobra.Command{
	Use:   "fonts",
	Short: "Automatically install recommended development fonts",
	Long: `Automatically download and install a curated set of development fonts.
This includes:
- Fira Code
- JetBrains Mono
- Cascadia Code
- Hack
- Ubuntu Mono`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFontsMenu()
	},
}

func runFontsMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Fonts")

	fmt.Println("Bluefin recommends a set of modern development fonts:")
	fmt.Println("- Fira Code")
	fmt.Println("- JetBrains Mono")
	fmt.Println("- Cascadia Code")
	fmt.Println("- Hack")
	fmt.Println("- Ubuntu Mono")
	fmt.Println()

	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install recommended development fonts?").
				Value(&confirm),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return err
	}

	if !confirm {
		return nil
	}

	fmt.Println(tui.InfoStyle.Render("Installing recommended development fonts..."))
	if err := install.Bundle("fonts"); err != nil {
		return err
	}
	err := maybeHandlePostFontInstall()
	tui.Pause()
	return err
}

func init() {
	rootCmd.AddCommand(fontsCmd)
}
