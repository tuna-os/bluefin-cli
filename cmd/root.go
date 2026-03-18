package cmd

import (
	"fmt"

	"github.com/hanthor/bluefin-cli/internal/countme"
	"github.com/spf13/cobra"
)

var (
	version = "0.0.3"
)

var rootCmd = &cobra.Command{
	Use:   "bluefin-cli",
	Short: "A powerful CLI tool for managing Homebrew and shell customization",
	Long: `Bluefin CLI brings the bluefin terminal experience to you.

Standard (Vanilla) Features:
- Homebrew & Tool Management
- Shell Environment Configuration
- System Status & MOTD

Extra Features:
- Automated Theme & Wallpaper Switching (Sunset)
- Automated Font Installation
- Monthly Wallpaper Themes`,
	Version: version,
	// Fire the countme ping in the background on every invocation.
	// This is a no-op if already counted this week, or if opted out.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip the ping when the user is managing countme itself, to avoid
		// a ping firing just before an explicit --disable.
		if cmd.Name() != "countme" {
			go countme.Count(version)
		}
		return nil
	},
	// If no subcommand is provided, open the interactive main menu by default.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Defer to the interactive menu
		if menuCmd != nil && menuCmd.RunE != nil {
			return menuCmd.RunE(menuCmd, nil)
		}
		// Fallback: show help if menu is not available for some reason
		return cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bluefin-cli version %s\n", version))

	rootCmd.AddGroup(&cobra.Group{
		ID:    "vanilla",
		Title: "Standard (Vanilla) Features:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "extra",
		Title: "Extra Features:",
	})
}
