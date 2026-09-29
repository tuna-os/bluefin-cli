package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

var (
	uninstallRemoveSoftware bool
	uninstallRemoveModules  bool
	uninstallKeepConfig     bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Bluefin shell setup and its tools",
	Long: `Remove the Bluefin shell setup from powershell, bash, zsh, and fish.

By default this command also tries to uninstall the software it manages:
  - Windows: winget-managed shell tools
  - Linux/macOS/WSL: Homebrew-managed shell tools

Use flags to keep software/modules/config as needed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := shell.UninstallOptions{
			Shells:         append([]string{"powershell"}, shell.ManagedShells()...),
			RemoveSoftware: uninstallRemoveSoftware,
			RemoveModules:  uninstallRemoveModules,
			RemoveConfig:   !uninstallKeepConfig,
		}

		if err := shell.UninstallSetup(opts); err != nil {
			return fmt.Errorf("uninstall completed with issues: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVar(&uninstallRemoveSoftware, "software", true, "Uninstall managed software packages")
	uninstallCmd.Flags().BoolVar(&uninstallRemoveModules, "modules", true, "Uninstall PowerShell modules managed by Bluefin CLI")
	uninstallCmd.Flags().BoolVar(&uninstallKeepConfig, "keep-config", false, "Keep Bluefin shell preferences JSON")
}
