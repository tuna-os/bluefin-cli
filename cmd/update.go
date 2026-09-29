package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/update"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update bluefin-cli to the latest release",
	Long: `Check for and install the latest bluefin-cli release.

Binaries installed via a package manager (Homebrew, Winget, Scoop) are not
self-updated. For those, this command tells you the upgrade command to use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")
		return runUpdate(checkOnly)
	},
}

// runUpdate is shared by the CLI command and the menu's palette action
// (where it runs inside a RunnerScreen with output captured).
func runUpdate(checkOnly bool) error {
	{
		rel, err := update.Latest()
		if err != nil {
			return err
		}

		if !update.IsNewer(version, rel.TagName) {
			if version == "dev" {
				fmt.Println(tui.InfoStyle.Render(fmt.Sprintf("Running a dev build; latest release is %s. Not self-updating.", rel.TagName)))
				return nil
			}
			fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("✓ Already up to date (%s).", version)))
			return nil
		}

		fmt.Printf("Update available: %s → %s\n", version, rel.TagName)

		if hint := update.Detect().UpdateHint(); hint != "" {
			fmt.Println(tui.InfoStyle.Render("This binary is managed by a package manager. Update with:"))
			fmt.Printf("  %s\n", hint)
			return nil
		}

		if checkOnly {
			fmt.Println("Run 'bluefin-cli update' to install it.")
			return nil
		}

		if err := update.Apply(rel, os.Stdout); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("✓ Updated to %s. Restart bluefin-cli to use it.", rel.TagName)))
		return nil
	}
}

func init() {
	updateCmd.Flags().Bool("check", false, "Only check whether an update is available")
	rootCmd.AddCommand(updateCmd)
}
