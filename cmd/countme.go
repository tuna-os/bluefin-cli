package cmd

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/countme"
	"github.com/spf13/cobra"
)

var (
	countmeDisable bool
	countmeEnable  bool
	countmeStatus  bool
)

var countmeCmd = &cobra.Command{
	Use:     "countme",
	Short:   "Manage anonymous usage counting",
	Long: `bluefin-cli participates in Fedora's countme protocol to report
anonymous install counts alongside native Bluefin Linux installs.

Each week, a single GET request is sent to Fedora's mirror infrastructure
with a User-Agent that identifies the platform (mac, wsl, powershell).
No personal data, IP addresses, or machine identifiers are transmitted.
The aggregate data is publicly available at:
  https://data-analysis.fedoraproject.org/csv-reports/countme/totals.csv

Opt out at any time:
  bluefin-cli countme --disable
  # or permanently via environment:
  export BLUEFIN_DISABLE_COUNTME=1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case countmeDisable:
			if err := countme.Disable(); err != nil {
				return fmt.Errorf("failed to disable countme: %w", err)
			}
			fmt.Println("countme disabled. Set BLUEFIN_DISABLE_COUNTME=1 in your environment for a session-only opt-out.")
		case countmeEnable:
			if err := countme.Enable(); err != nil {
				return fmt.Errorf("failed to enable countme: %w", err)
			}
			fmt.Println("countme enabled.")
		default:
			fmt.Println(countme.StatusString(version))
		}
		return nil
	},
}

func init() {
	countmeCmd.Flags().BoolVar(&countmeDisable, "disable", false, "Persistently opt out of anonymous usage counting")
	countmeCmd.Flags().BoolVar(&countmeEnable, "enable", false, "Re-enable anonymous usage counting after opting out")
	countmeCmd.Flags().BoolVar(&countmeStatus, "status", false, "Show current countme configuration (default behaviour)")
	countmeCmd.MarkFlagsMutuallyExclusive("disable", "enable")

	rootCmd.AddCommand(countmeCmd)
}
