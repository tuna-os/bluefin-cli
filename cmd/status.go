package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/status"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration status",
	Long:  `Display the current configuration status for shell experience, MOTD, and installed tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return status.Show(reportedVersion())
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
