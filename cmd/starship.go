package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/starship"
)

var starshipCmd = &cobra.Command{
	Use:   "starship",
	Short: "Use Starship themes",
	Long:  `Install Starship themes and set their options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchFlow(starshipFlow())
	},
}

var starshipThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Select and apply a Starship theme",
	Long:  `Choose a Starship preset from the menu.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return starship.ApplyTheme(args[0])
		}
		return launchFlow(starshipFlow())
	},
}

var starshipInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Starship prompt",
	Long:  `Download and install the Starship prompt if not already installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return starship.Install()
	},
}

func init() {
	rootCmd.AddCommand(starshipCmd)
	starshipCmd.AddCommand(starshipThemeCmd)
	starshipCmd.AddCommand(starshipInstallCmd)
}
