//go:build extra

package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

func supportsWindowsThemePostInstall() bool {
	return env.IsWSL() || env.IsWindows()
}

func maybeHandleWindowsThemePostInstall(cmd *cobra.Command, casks []string) error {
	if !supportsWindowsThemePostInstall() {
		return nil
	}

	nonInteractive := false
	yes := false
	if cmd != nil {
		nonInteractive, _ = cmd.Flags().GetBool("non-interactive")
		yes, _ = cmd.Flags().GetBool("yes")
	}

	if yes {
		return runSunsetSetup()
	}

	if nonInteractive {
		return nil
	}

	return maybePromptForSunsetSetup()
}

func maybePromptForSunsetSetup() error {
	var startSetup bool
	confirm := huh.NewConfirm().
		Title("Would you like to configure solar-based theme and wallpaper switching now?").
		Description("This uses the new 'sunset' feature to automatically manage your desktop experience.").
		Value(&startSetup).
		WithTheme(tui.AppTheme).
		WithKeyMap(tui.ConfirmKeyMap())

	if err := confirm.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return err
	}

	if startSetup {
		return runSunsetSetup()
	}

	return nil
}

func runSunsetSetup() error {
	return RunSunsetSetupFlow()
}

func maybeHandlePostFontInstall() error {
	// Placeholder for future automated font setting logic (e.g. configuring Windows Terminal or GNOME Console)
	fmt.Println(tui.SuccessStyle.Render("✓ Recommended fonts downloaded. (Extra: Automated terminal configuration coming soon!)"))
	return nil
}
