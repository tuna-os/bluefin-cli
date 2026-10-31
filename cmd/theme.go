package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

var themeCmd = &cobra.Command{
	Use:   "theme [flavor]",
	Short: "Show or set the Catppuccin flavor (auto, latte, frappe, macchiato, mocha)",
	Long: `Show or set the UI color flavor.

"auto" (the default) picks Latte on light terminals and Mocha on dark ones;
naming a flavor pins it regardless of the terminal background.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: theme.Flavors(),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Printf("Flavor: %s (options: %s)\n",
				config.GetUIFlavor(), strings.Join(theme.Flavors(), ", "))
			return nil
		}
		flavor := strings.ToLower(args[0])
		if !slices.Contains(theme.Flavors(), flavor) {
			return fmt.Errorf("unknown flavor %q (options: %s)", flavor, strings.Join(theme.Flavors(), ", "))
		}
		config.SetUIFlavor(flavor)
		if err := config.Save(); err != nil {
			return err
		}
		fmt.Printf("Flavor set to %s.\n", flavor)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(themeCmd)
}

// applyThemeFlavor loads the configured flavor into the theme package; called
// from the root PersistentPreRun so every command renders with it.
func applyThemeFlavor() {
	flavor := strings.ToLower(config.GetUIFlavor())
	if flavor == "" || flavor == "auto" {
		return
	}
	theme.Flavor = flavor
	theme.DefaultTheme = theme.Resolve(theme.DefaultTheme.IsDark)
}
