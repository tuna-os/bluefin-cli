package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/starship"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/tui/shellmenu"
)

// Compose the presentation screens with the real shell and prompt services.
// The same screens are used by both the menu and their standalone commands.
func motdMenuScreen() app.Screen {
	return shellmenu.MOTD(
		func() bool { return shell.MOTDEnabled(currentShellName()) },
		func() (bool, error) { return shell.ToggleMOTD(currentShellName()) },
		motd.Show,
	)
}

func shellsFormScreen() app.Screen {
	return shellmenu.Shells(shell.GetInstalledShells, shell.CheckStatus, shell.Toggle)
}

func starshipFlow() tea.Cmd {
	return shellmenu.Starship(starship.InstallAndApplyTheme)
}
