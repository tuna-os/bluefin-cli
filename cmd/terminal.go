package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/terminal"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var terminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Set up a great terminal: install Ghostty, pin it, theme it",
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchFlow(app.Push(terminalMenuScreen()))
	},
}

var terminalSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Do the full setup with no prompts: install + pin + config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := terminal.InstallGhostty(); err != nil {
			return err
		}
		if runtime.GOOS == "darwin" {
			if err := terminal.PinToDock("/Applications/Ghostty.app"); err != nil {
				return err
			}
		}
		if err := terminal.WriteGhosttyConfig(); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Terminal setup complete."))
		return nil
	},
}

func init() {
	terminalCmd.AddCommand(terminalSetupCmd)
	rootCmd.AddCommand(terminalCmd)
}

// macOSAutoAppearance reads whether macOS switches light/dark automatically.
func macOSAutoAppearance() bool {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyleSwitchesAutomatically").Output()
	return err == nil && strings.TrimSpace(string(out)) == "1"
}

func terminalMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		name := terminal.PreferredName()
		installLabel := "Install " + name
		installDesc := "The fast, modern GPU terminal"
		if terminal.GhosttyInstalled() {
			installLabel = name + " installed ✓"
			installDesc = "Reinstall / upgrade via the package manager"
		}
		out := []app.MenuItem{
			{Icon: "👻", Label: installLabel, Value: "install", Desc: installDesc},
			{Icon: "🎨", Label: "Apply cool " + name + " config", Value: "config",
				Desc: "Catppuccin auto light/dark, padding, Nerd Font"},
		}
		if runtime.GOOS == "darwin" {
			out = append(out, app.MenuItem{Icon: "📌", Label: "Pin Ghostty to Dock", Value: "pin",
				Desc: "One click from anywhere"})
			auto := "Auto Light/Dark: off"
			if macOSAutoAppearance() {
				auto = "Auto Light/Dark: on"
			}
			out = append(out, app.MenuItem{Icon: "🌗", Label: auto, Value: "appearance",
				Desc: "macOS switches appearance with the sun"})
		}
		out = append(out, app.MenuItem{Icon: "🚀", Label: "Do it all", Value: "all",
			Desc: "Install + pin + config in one go"})
		return out
	}

	return app.NewMenu("Terminal", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "install":
			return app.Push(app.NewRunner("Installing "+terminal.PreferredName(), terminal.InstallGhostty))
		case "pin":
			return app.Push(app.NewRunner("Pinning to Dock", func() error {
				return terminal.PinToDock("/Applications/Ghostty.app")
			}))
		case "config":
			return app.Push(app.NewRunner("Writing "+terminal.PreferredName()+" config", terminal.WriteGhosttyConfig))
		case "appearance":
			return tea.Sequence(func() tea.Msg {
				target := "1"
				verb := "enabled"
				if macOSAutoAppearance() {
					target = "0"
					verb = "disabled"
				}
				if out, err := exec.Command("defaults", "write", "-g",
					"AppleInterfaceStyleSwitchesAutomatically", "-bool", map[string]string{"1": "true", "0": "false"}[target]).CombinedOutput(); err != nil {
					return app.ToastMsg{Text: fmt.Sprintf("Error: %v: %s", err, out), IsErr: true}
				}
				return app.ToastMsg{Text: "Auto appearance " + verb + " (takes effect after next login)."}
			}, app.ReloadTop())
		case "all":
			return app.Push(app.NewRunner("Terminal setup", func() error {
				if err := terminal.InstallGhostty(); err != nil {
					return err
				}
				if runtime.GOOS == "darwin" {
					if err := terminal.PinToDock("/Applications/Ghostty.app"); err != nil {
						return err
					}
				}
				return terminal.WriteGhosttyConfig()
			}))
		}
		return nil
	})
}
