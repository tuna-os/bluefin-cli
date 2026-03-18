//go:build !extra

package cmd

import (
	"github.com/spf13/cobra"
)

func maybeHandleWindowsThemePostInstall(cmd *cobra.Command, casks []string) error {
	return nil
}

func maybeHandlePostFontInstall() error {
	return nil
}

func RunSunsetSetupFlow() error {
	return nil
}
