package cmd

import (
	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/status"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the interactive Bluefin main menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		for {
			tui.ClearScreen()
			tui.RenderHeader("Bluefin CLI", "Main Menu")

			shellStatus := shell.CheckStatus()
			hasShell := false
			for _, v := range shellStatus {
				if v {
					hasShell = true
					break
				}
			}

			var shellLabel string
			statusLabel := "📊 Status"
			installLabel := "📦 Install Apps ❯"
			wallpapersLabel := "🖼  Wallpapers ❯"
			starshipLabel := "🚀 Starship Theme ❯"

			if env.IsWindows() {
				statusLabel = "Status"
				installLabel = "Install Apps >"
				wallpapersLabel = "Wallpapers >"
				starshipLabel = "Starship Theme >"
			}

			if hasShell {
				shellLabel = "🐚 Bluefin Shell (Enabled)"
			} else {
				shellLabel = "🐚 Bluefin Shell (Disabled)"
			}

			if env.IsWindows() {
				shellLabel = "Bluefin Shell (Disabled)"
				if hasShell {
					shellLabel = "Bluefin Shell (Enabled)"
				}
				shellLabel += " >"
			} else {
				shellLabel += " ❯"
			}

			opts := []huh.Option[string]{
				huh.NewOption(statusLabel, "status"),
				huh.NewOption(shellLabel, "shell"),
				huh.NewOption(installLabel, "bundles"),
				huh.NewOption(wallpapersLabel, "wallpapers"),
				huh.NewOption(starshipLabel, "starship"),
			}
			opts = append(opts, huh.NewOption("Exit", "exit"))

			var choice string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose an action").
						Options(opts...).
						Value(&choice),
				),
			).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

			if err := form.Run(); err != nil {
				// ESC pressed on main menu - exit cleanly
				return nil
			}

			switch choice {
			case "status":
				if err := status.Show(); err != nil {
					return err
				}
				tui.Pause()
			case "shell":
				if err := runShellMenu(); err != nil {
					return err
				}

			case "bundles":
				if err := runBundlesMenu(); err != nil {
					return err
				}
			case "wallpapers":
				if err := runWallpapersMenu(); err != nil {
					return err
				}
			case "starship":
				if err := runStarshipMenu(); err != nil {
					return err
				}
			case "exit":
				return nil
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}
