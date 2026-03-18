package cmd

import (
	"fmt"

	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/spf13/cobra"
)

var (
	uninstallRemoveSoftware bool
	uninstallRemoveModules  bool
	uninstallKeepConfig     bool
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	GroupID: "vanilla",
	Short:   "Uninstall Bluefin shell setup and managed tools",
	Long: `Remove Bluefin shell initialization setup across powershell, bash, zsh, and fish.

By default this command also attempts to uninstall managed software:
  - Windows: winget-managed shell tools
  - Linux/macOS/WSL: Homebrew-managed shell tools

Use flags to keep software/modules/config as needed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := shell.UninstallOptions{
			Shells:         []string{"powershell", "bash", "zsh", "fish"},
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
