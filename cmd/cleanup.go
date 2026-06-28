package cmd

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:     "cleanup",
	Short:   "Uninstall Bluefin shell setup and managed tools (alias for uninstall)",
	Long: `Remove Bluefin shell initialization setup across all supported shells.
By default, this command also attempts to uninstall managed software and modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := shell.UninstallOptions{
			Shells:         []string{"powershell", "bash", "zsh", "fish"},
			RemoveSoftware: true,
			RemoveModules:  true,
			RemoveConfig:   false, // Keep config by default unless --all or similar is used
		}

		fmt.Println(tui.InfoStyle.Render("Starting full cleanup..."))

		if err := shell.UninstallSetup(opts); err != nil {
			return fmt.Errorf("cleanup completed with issues: %w", err)
		}

		fmt.Println(tui.SuccessStyle.Render("\n✓ Cleanup complete. Shell enhancements have been removed."))
		fmt.Println(tui.InfoStyle.Render("Note: You may need to restart your terminal for changes to take full effect."))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
