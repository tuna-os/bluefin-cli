package cmd

import (
	"fmt"
	"strings"

	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/spf13/cobra"
)

var (
	// Flags are now dynamic, stored in a map
	toolFlags = make(map[string]*bool)
)

var initCmd = &cobra.Command{
	Use:     "init [bash|zsh|fish|powershell|pwsh]",
	Short:   "Generate shell initialization script",
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
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell", "pwsh"},
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
