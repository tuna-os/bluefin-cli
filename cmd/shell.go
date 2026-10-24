package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/platform"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var shellCmd = &cobra.Command{
	Use:   "shell [shell] [on|off]",
	Short: "Turn shell tools on or off",
	Long: `Turn shell tools on or off. This includes modern aliases and tool initialization.

The Shell Experience provides:
  - Modern ls replacement with eza (ll, ls aliases)
  - bat for cat with syntax highlighting
  - ugrep for faster grep
  - Initialization for atuin, starship, and zoxide`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return launchFlow(app.Push(shellMenuScreen()))
		}

		// Args provided
		selectedShell := args[0]
		enable := true // default to on
		if len(args) > 1 {
			enable = args[1] == "on"
		}

		return shell.Toggle(selectedShell, enable)
	},
}

var shellConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Set options for each shell tool",
	Long:  `Set options for each shell tool from the menu.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchFlow(app.Push(componentsFormScreen()))
	},
}

func init() {
	describedTools := shell.Tools
	if platform.IsWindows() {
		describedTools = shell.ToolsForShell("powershell")
	}

	// Generate dynamic long description
	var sb strings.Builder
	sb.WriteString("Turn shell tools on or off. This includes modern aliases and tool initialization.\n\nThe Shell Experience provides:\n")
	for _, tool := range describedTools {
		fmt.Fprintf(&sb, "  - %s: %s\n", tool.Name, tool.Description)
	}
	shellCmd.Long = sb.String()

	rootCmd.AddCommand(shellCmd)
	shellCmd.AddCommand(shellConfigCmd)
}
