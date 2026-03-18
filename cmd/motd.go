package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/motd"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var motdCmd = &cobra.Command{
	Use:     "motd",
	Short:   "Manage Message of the Day",
	Long:  `Configure and display the Message of the Day (MOTD) with system info and tips.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMotdMenu()
	},
}

var motdToggleCmd = &cobra.Command{
	Use:   "toggle [shell|all] [on|off]",
	Short: "Toggle MOTD for shells",
	Long:  `Enable or disable MOTD display on shell startup for bash, zsh, fish, or all shells.`,
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "all"
		enable := true

		if len(args) > 0 {
			target = args[0]
		}
		if len(args) > 1 {
			enable = args[1] == "on"
		}

		return motd.Toggle(target, enable)
	},
}

var motdShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the MOTD",
	Long:  `Display the Message of the Day with system information and a random tip.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return motd.Show()
	},
}

var motdConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure MOTD settings",
	Long:  `Interactively configure MOTD theme and settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return motd.SetTheme(args[0])
		}

		var selectedTheme string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose MOTD theme").
					Options(
						huh.NewOption("Slate (default)", "slate"),
						huh.NewOption("Dark", "dark"),
						huh.NewOption("Light", "light"),
						huh.NewOption("Dracula", "dracula"),
						huh.NewOption("Pink", "pink"),
					).
					Value(&selectedTheme),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				return nil
			}
			return fmt.Errorf("form error: %w", err)
		}

		return motd.SetTheme(selectedTheme)
	},
}

func init() {
	rootCmd.AddCommand(motdCmd)
	motdCmd.AddCommand(motdToggleCmd)
	motdCmd.AddCommand(motdShowCmd)
	motdCmd.AddCommand(motdConfigCmd)
}

func runMotdMenu() error {
	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Main Menu > Shell > MOTD")

		// Load config to check if MOTD is enabled
		// Determine shell (fallback to bash if unknown, as this affects defaults)
		currentShellPath := os.Getenv("SHELL")
		currentShell := filepath.Base(currentShellPath)
		if currentShell == "" || currentShell == "." {
			currentShell = "bash"
		}

		cfg, err := shell.LoadConfig(currentShell)
		if err != nil {
			cfg = shell.DefaultConfig(currentShell)
		}
		isEnabled := cfg.IsEnabled("Motd")

		// Build toggle label based on current state
		toggleLabel := "Enable MOTD"
		if isEnabled {
			toggleLabel = "Disable MOTD"
		}

		var action string
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("MOTD – What do you want to do?").
					Options(
						huh.NewOption(toggleLabel, "toggle_motd"),
						huh.NewOption("Show MOTD", "show"),
						huh.NewOption("Exit to Shell Menu", "exit"),
					).
					Value(&action),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap()).Run(); err != nil {
			return nil
		}

		switch action {
		case "toggle_motd":
			cfg.SetEnabled("Motd", !isEnabled)
			if err := shell.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			if !isEnabled {
				fmt.Println(tui.SuccessStyle.Render("✓ MOTD enabled"))
			} else {
				fmt.Println(tui.SuccessStyle.Render("✓ MOTD disabled"))
			}
			tui.Pause()
		case "show":
			if err := motd.Show(); err != nil {
				return err
			}
			tui.Pause()
		case "exit":
			return nil
		}
	}
}
