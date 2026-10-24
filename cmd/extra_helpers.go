//go:build extra

package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/platform"
	"github.com/tuna-os/bluefin-cli/internal/terminal"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

func supportsWindowsThemePostInstall() bool {
	return platform.GetCapabilities().SupportsSunset
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
	font := terminal.DetectNerdFont()
	if font == "" {
		fmt.Println(tui.InfoStyle.Render("No Nerd Fonts detected after install. They may need a logout/restart to become visible."))
		return nil
	}
	fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("✓ Detected Nerd Font: %s", font)))

	terminals := terminal.DetectTerminals()
	if len(terminals) == 0 {
		fmt.Println(tui.InfoStyle.Render("No supported terminals detected. Installed fonts are ready to use — set them in your terminal's preferences."))
		return nil
	}

	fmt.Println(tui.InfoStyle.Render(fmt.Sprintf("Found %d terminal(s) to configure…", len(terminals))))
	for _, t := range terminals {
		fmt.Printf("  Configuring %s…\n", t.Name)
		if err := terminal.SetTerminalFont(t, font); err != nil {
			fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("  ✗ %s: %v", t.Name, err)))
		} else {
			fmt.Println(tui.SuccessStyle.Render(fmt.Sprintf("  ✓ %s configured", t.Name)))
		}
	}
	fmt.Println(tui.SuccessStyle.Render("✓ Font setup complete. Restart your terminals to see the new font."))
	return nil
}
