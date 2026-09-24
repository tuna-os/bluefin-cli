package cmd

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var motdCmd = &cobra.Command{
	Use:   "motd",
	Short: "Manage Message of the Day",
	Long:  `Configure and display the Message of the Day (MOTD) with system info and tips.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchFlow(app.Push(motdMenuScreen()))
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

		targets := []string{target}
		if target == "all" || target == "" {
			targets = []string{"bash", "zsh", "fish"}
		}
		for _, sh := range targets {
			cfg, err := shell.LoadConfig(sh)
			if err != nil {
				cfg = shell.DefaultConfig(sh)
			}
			cfg.SetEnabled("Motd", enable)
			if err := shell.SaveConfig(cfg); err != nil {
				return fmt.Errorf("saving %s config: %w", sh, err)
			}
		}
		state := "disabled"
		if enable {
			state = "enabled"
		}
		fmt.Printf("MOTD %s for %s.\n", state, strings.Join(targets, ", "))
		return nil
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
	Short: "Configure the settings of the MOTD",
	Long:  `Configure the theme and settings of the MOTD in a form.`,
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
