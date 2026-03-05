package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/shell"
)

var (
	// Flags are now dynamic, stored in a map
	toolFlags = make(map[string]*bool)
)

var initCmd = &cobra.Command{
	Use:   "init [bash|zsh|fish]",
	Short: "Generate shell initialization script",
	Long:  `Generate the shell initialization script for bluefin-cli.
Add the following to your shell configuration file:

Bash (~/.bashrc):
  eval "$(bluefin-cli init bash)"

Zsh (~/.zshrc):
  eval "$(bluefin-cli init zsh)"

Fish (~/.config/fish/config.fish):
  bluefin-cli init fish | source
`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = install.MaybeRollOverWindowsThemeOnInit()

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
		fmt.Println()

		// Add MOTD hook if enabled in config
		if config.IsEnabled("Motd") {
			switch shellName {
			case "bash", "zsh":
				// Only run MOTD if interactive
				fmt.Println(`# bluefin-cli motd hook
if [ -n "$PS1" ] && [ -t 1 ]; then
    bluefin-cli motd show
fi`)
			case "fish":
				fmt.Println(`# bluefin-cli motd hook
if status is-interactive
    bluefin-cli motd show
end`)
			default:
				return fmt.Errorf("unsupported shell: %s", shellName)
			}
		}

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
