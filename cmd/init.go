package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

var (
	// Flags are now dynamic, stored in a map
	toolFlags = make(map[string]*bool)
)

var initCmd = &cobra.Command{
	Use:   "init [bash|zsh|fish|ash|nu|powershell|pwsh]",
	Short: "Generate shell initialization script",
	Long: `Generate the shell initialization script for bluefin-cli.
Add the following to your shell configuration file:

Bash (~/.bashrc):
  eval "$(bluefin-cli init bash)"

Zsh (~/.zshrc):
  eval "$(bluefin-cli init zsh)"

Fish (~/.config/fish/config.fish):
  bluefin-cli init fish | source

PowerShell ($PROFILE):
  Invoke-Expression (& bluefin-cli init powershell)

Ash (~/.ashrc, with ENV="$HOME/.ashrc" exported from ~/.profile):
  eval "$(bluefin-cli init ash)"

Nushell (~/.config/nushell/config.nu) — nushell cannot evaluate a string, so
the script is saved to a file and sourced from there:
  bluefin-cli init nu | save -f ~/.config/nushell/bluefin-cli.nu
  source ~/.config/nushell/bluefin-cli.nu

'bluefin-cli shell enable <shell>' wires any of these up for you.
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "ash", "nu", "nushell", "powershell", "pwsh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shellName := args[0]

		config, err := shell.LoadConfig(shellName)
		if err != nil {
			config = shell.DefaultConfig(shellName)
		}

		for _, tool := range shell.Tools {
			flagName := strings.ToLower(tool.Name)
			if cmd.Flags().Changed(flagName) {
				if val, ok := toolFlags[flagName]; ok {
					config.SetEnabled(tool.Name, *val)
				}
			}
		}

		// Generate bling/shell script
		script, err := shell.Init(shellName, config)
		if err != nil {
			return err
		}

		// Print the bling script
		fmt.Println(script)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	for _, tool := range shell.Tools {
		flagName := strings.ToLower(tool.Name)
		toolFlags[flagName] = initCmd.Flags().Bool(flagName, tool.Default, fmt.Sprintf("Enable %s", tool.Name))
	}

	// MOTD is managed separately from tools
	motdDefault := true
	toolFlags["motd"] = &motdDefault
	initCmd.Flags().BoolVar(toolFlags["motd"], "motd", true, "Enable MOTD")
}
