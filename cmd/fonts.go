package cmd

import (
	"fmt"

	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var fontsCmd = &cobra.Command{
	Use:     "fonts",
	GroupID: "extra",
	Short:   "Automatically install recommended development fonts",
	Long: `Automatically download and install a curated set of development fonts.
This includes:
- Fira Code
- JetBrains Mono
- Cascadia Code
- Hack
- Ubuntu Mono`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(tui.InfoStyle.Render("Installing recommended development fonts..."))
		// Reuse the fonts bundle logic but make it a top-level easy command
		return install.Bundle("fonts")
	},
}

func init() {
	rootCmd.AddCommand(fontsCmd)
}
