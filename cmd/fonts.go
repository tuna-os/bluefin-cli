package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

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

	if runtime.GOOS == "windows" {
		return runFontsMenuWindows()
	}
	return runFontsMenuUnix()
}

// runFontsMenuWindows uses NerdFont structs with winget/scoop support.
func runFontsMenuWindows() error {
	opts := make([]huh.Option[int], 0, len(install.NerdFonts))
	for i, f := range install.NerdFonts {
		label := f.Name
		if install.IsFontInstalled(f) {
			label += " ✓"
		}
		opts = append(opts, huh.NewOption(label, i))
	}

	var selected []int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
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

	for _, idx := range selected {
		font := install.NerdFonts[idx]
		fmt.Println(tui.InfoStyle.Render("Installing " + font.Name + "..."))
		if err := install.InstallFont(font); err != nil {
			fmt.Println(tui.ErrorStyle.Render("Failed: " + err.Error()))
		} else {
			fmt.Println(tui.SuccessStyle.Render("✓ " + font.Name))
		}
	}

	err := maybeHandlePostFontInstall()
	tui.Pause()
	return err
}

// runFontsMenuUnix uses the brew-cask based font list for macOS/Linux.
func runFontsMenuUnix() error {
	type brewFont struct {
		Name string
		Cask string
	}
	availableFonts := []brewFont{
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
