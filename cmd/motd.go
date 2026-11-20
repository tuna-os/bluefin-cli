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
	Short: "Set up daily shell messages",
	Long:  `Set up and display the daily shell message with system info and tips.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchFlow(app.Push(motdMenuScreen()))
	},
}

var motdToggleCmd = &cobra.Command{
	Use:   "toggle [shell|all] [on|off]",
	Short: "Set MOTD display for shells",
	Long: `Enable or disable MOTD display on shell startup for bash, zsh, fish, or all shells.

Where uwelcome (the terminal welcome tool of Bluefin) is installed,
this turns uwelcome on or off for every shell. Without on|off, it flips
the current state of uwelcome.`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if motd.Uwelcome() {
			enable := !motd.UwelcomeEnabled()
			if len(args) > 1 {
				enable = args[1] == "on"
			}
			if err := motd.SetUwelcome(enable); err != nil {
				return err
			}
			state := "off"
			if enable {
				state = "on"
			}
			fmt.Printf("uwelcome greeting turned %s.\n", state)
			return nil
		}

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
	Long: `Display the Message of the Day with system information and a random tip.

Where uwelcome (the terminal welcome tool of Bluefin) is installed,
this shows the welcome of uwelcome. Otherwise the message comes from the first
template that exists. The order is: template-file in motd.json,
/etc/bluefin-cli/motd.md, /usr/share/bluefin-cli/motd.md, then the built-in
one. Tips come the same way from tips-directory, /etc/bluefin-cli/tips and
/usr/share/bluefin-cli/tips.

With --if-enabled it prints nothing when the MOTD is off
('bluefin-cli motd toggle all off', or uwelcome's own switch), for login hooks
that greet every terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ifEnabled, _ := cmd.Flags().GetBool("if-enabled"); ifEnabled {
			if motd.Uwelcome() {
				if !motd.UwelcomeEnabled() {
					return nil
				}
			} else if cfg, err := shell.LoadConfig(currentShellName()); err == nil && !cfg.IsEnabled("Motd") {
				return nil
			}
		}
		return motd.Show()
	},
}

var motdConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Set MOTD options",
	Long:  `Set the MOTD theme and options from the menu.`,
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
	motdShowCmd.Flags().Bool("if-enabled", false, "Show nothing when the MOTD is off")
	rootCmd.AddCommand(motdCmd)
	motdCmd.AddCommand(motdToggleCmd)
	motdCmd.AddCommand(motdShowCmd)
	motdCmd.AddCommand(motdConfigCmd)
}
