package cmd

import (
	"fmt"
	"os/exec"

	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

// availableFonts maps display name → brew cask name
var availableFonts = []struct {
	Name string
	Cask string
}{
	{"0xProto Nerd Font", "font-0xproto-nerd-font"},
	{"Cascadia Mono Nerd Font", "font-caskaydia-mono-nerd-font"},
	{"Comic Shanns Mono Nerd Font", "font-comic-shanns-mono-nerd-font"},
	{"Droid Sans Mono Nerd Font", "font-droid-sans-mono-nerd-font"},
	{"Fira Code Nerd Font", "font-fira-code-nerd-font"},
	{"Go Mono Nerd Font", "font-go-mono-nerd-font"},
	{"IBM Plex Mono Nerd Font", "font-blex-mono-nerd-font"},
	{"JetBrains Mono Nerd Font", "font-jetbrains-mono-nerd-font"},
	{"Source Code Pro", "font-source-code-pro"},
	{"Source Code Pro Nerd Font", "font-sauce-code-pro-nerd-font"},
	{"Ubuntu Nerd Font", "font-ubuntu-nerd-font"},
}

var fontsCmd = &cobra.Command{
	Use:   "fonts",
	Short: "Install individual development fonts",
	Long:  `Select and install individual development fonts from a curated list.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFontsMenu()
	},
}

func runFontsMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Fonts")

	opts := make([]huh.Option[string], 0, len(availableFonts))
	for _, f := range availableFonts {
		opts = append(opts, huh.NewOption(f.Name, f.Cask))
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select fonts to install").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		return nil
	}

	if len(selected) == 0 {
		fmt.Println(tui.InfoStyle.Render("No fonts selected."))
		tui.Pause()
		return nil
	}

	for _, cask := range selected {
		fmt.Println(tui.InfoStyle.Render("Installing " + cask + "..."))
		cmd := exec.Command("brew", "install", "--cask", cask)
		cmd.Stdout = nil
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Println(tui.ErrorStyle.Render("Failed: " + string(out)))
		} else {
			fmt.Println(tui.SuccessStyle.Render("✓ " + cask))
		}
	}

	err := maybeHandlePostFontInstall()
	tui.Pause()
	return err
}

func init() {
	rootCmd.AddCommand(fontsCmd)
}
