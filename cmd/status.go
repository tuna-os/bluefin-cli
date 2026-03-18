package cmd

import (
	"github.com/hanthor/bluefin-cli/internal/status"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show configuration status",
	Long:  `Display the current configuration status for shell experience, MOTD, and installed tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return status.Show()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
