package cmd

import (
	"github.com/hanthor/bluefin-cli/internal/status"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	GroupID: "vanilla",
	Short:   "Verify environment health and configuration",
	Long:    `Check the status of Homebrew, shell configurations, and other dependencies to ensure everything is working correctly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return status.Check()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
