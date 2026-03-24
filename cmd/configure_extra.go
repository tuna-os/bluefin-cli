//go:build extra

package cmd

import (
	"fmt"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/terminal"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)


var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure terminal font and theme (plus only)",
	Long:  `Automatically set Nerd Font and Catppuccin theme in Windows Terminal, VS Code, Ghostty, and more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigureMenu()
	},
}

func runConfigureMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Configure Terminals")

	// Discover installed terminals first so we can show them in the description
	discovered := terminal.DiscoverTerminals()
	var discoveredNames []string
	for _, t := range discovered {
		discoveredNames = append(discoveredNames, string(t))
	}
	discoveredStr := "none detected"
	if len(discoveredNames) > 0 {
		discoveredStr = strings.Join(discoveredNames, ", ")
	}

	// Step 1: Font selection
	defaultFont := install.DefaultFont()
	selectedFontName := defaultFont.Name
	fontOpts := make([]huh.Option[string], 0, len(install.NerdFonts))
	for _, f := range install.NerdFonts {
		label := f.Name
		if install.IsFontInstalled(f) {
			label += " ✓"
		}
		fontOpts = append(fontOpts, huh.NewOption(label, f.Name))
	}

	fontForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a Nerd Font").
				Description("Detected terminals: "+discoveredStr).
				Options(fontOpts...).
				Value(&selectedFontName),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := fontForm.Run(); err != nil {
		return nil
	}
	if fontForm.State == huh.StateAborted {
		return nil
	}

	var selectedFont install.NerdFont
	for _, f := range install.NerdFonts {
		if f.Name == selectedFontName {
			selectedFont = f
			break
		}
	}

	// Step 2: Optional Catppuccin theme
	var applyTheme bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply a Catppuccin color theme too?").
				Value(&applyTheme),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())

	if err := confirmForm.Run(); err != nil {
		return nil
	}

	var selectedTheme string
	if applyTheme && confirmForm.State != huh.StateAborted {
		themeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a Catppuccin theme").
					Options(
						huh.NewOption("Frappe — balanced dark", "catppuccin-frappe"),
						huh.NewOption("Macchiato — darker", "catppuccin-macchiato"),
						huh.NewOption("Mocha — darkest", "catppuccin-mocha"),
						huh.NewOption("Latte — light", "catppuccin-latte"),
					).
					Value(&selectedTheme),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

		if err := themeForm.Run(); err != nil {
			return nil
		}
		if themeForm.State == huh.StateAborted {
			selectedTheme = ""
		}
	}

	// Install font if not already present
	if !install.IsFontInstalled(selectedFont) {
		fmt.Println(tui.InfoStyle.Render("Installing " + selectedFont.Name + "..."))
		if err := install.InstallFont(selectedFont); err != nil {
			fmt.Println(tui.ErrorStyle.Render("Font install failed: " + err.Error()))
			fmt.Println(tui.InfoStyle.Render("You can manually set the font face to: " + selectedFont.Face))
		} else {
			fmt.Println(tui.SuccessStyle.Render("✓ " + selectedFont.Name + " installed"))
		}
	}

	if len(discovered) == 0 {
		fmt.Println(tui.InfoStyle.Render("No supported terminals found for automatic configuration."))
		fmt.Println(tui.InfoStyle.Render("Manually set your terminal font face to: " + selectedFont.Face))
		tui.Pause()
		return nil
	}

	// Step 3: Target selection — all discovered or pick specific
	var applyMode string
	targetModeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Apply configuration to:").
				Options(
					huh.NewOption("All discovered apps ("+discoveredStr+")", "all"),
					huh.NewOption("Choose specific apps...", "pick"),
				).
				Value(&applyMode),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := targetModeForm.Run(); err != nil {
		return nil
	}
	if targetModeForm.State == huh.StateAborted {
		return nil
	}

	finalTargets := discovered
	if applyMode == "pick" {
		targetOpts := make([]huh.Option[string], 0, len(discovered))
		for _, t := range discovered {
			targetOpts = append(targetOpts, huh.NewOption(string(t), string(t)))
		}
		var chosen []string
		pickForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select apps to configure").
					Options(targetOpts...).
					Value(&chosen),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

		if err := pickForm.Run(); err != nil {
			return nil
		}
		finalTargets = finalTargets[:0]
		for _, c := range chosen {
			finalTargets = append(finalTargets, terminal.TerminalType(c))
		}
	}

	if len(finalTargets) == 0 {
		tui.Pause()
		return nil
	}

	if err := terminal.SetFontAndThemeToTargets(finalTargets, selectedFont, selectedTheme); err != nil {
		fmt.Println(tui.InfoStyle.Render("Some configuration errors: " + err.Error()))
	} else {
		fmt.Println(tui.SuccessStyle.Render("✓ Terminal configuration applied!"))
	}

	tui.Pause()
	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
