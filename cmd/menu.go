package cmd

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/update"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the Bluefin menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(mainMenuScreen(), checkForUpdate)
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}

// checkForUpdate runs in the background when the menu starts and surfaces a
// toast when a newer release exists. Failures are silent — an update nudge
// is never worth an error message — and the GitHub API is only asked once
// per day.
func checkForUpdate() tea.Msg {
	if last := config.GetUpdateLastCheck(); time.Since(time.Unix(last, 0)) < 24*time.Hour {
		return nil
	}
	config.SetUpdateLastCheck(time.Now().Unix())
	_ = config.Save()
	rel, err := update.Latest()
	if err != nil || !update.IsNewer(version, rel.TagName) {
		return nil
	}
	hint := "bluefin-cli update"
	if h := update.Detect().UpdateHint(); h != "" {
		hint = h
	}
	return app.ToastMsg{Text: fmt.Sprintf("⬆ %s available — run: %s", rel.TagName, hint)}
}

// launchFlow opens the interactive shell and immediately runs flow — used
// by commands like `bluefin-cli fonts` so every interactive path shares the
// native TUI (esc leads back to the Home menu).
func launchFlow(flow tea.Cmd) error {
	return app.Run(mainMenuScreen(), flow, checkForUpdate)
}

func mainMenuScreen() app.Screen {
	router := defaultMenuRouter()
	return app.NewMenu("Home", nil, mainMenuItems, router.Route)
}

func mainMenuItems() []app.MenuItem {
	return BuildMainMenuModel().Items
}

func mainMenuSelect(it app.MenuItem) tea.Cmd {
	return defaultMenuRouter().Route(it)
}
