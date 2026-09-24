package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

// brewfileTarget resolves --file or the home Brewfile.
func brewfileTarget(cmd *cobra.Command) string {
	if f, _ := cmd.Flags().GetString("file"); f != "" {
		return f
	}
	return install.HomeBrewfile()
}

var brewfileCmd = &cobra.Command{
	Use:   "brewfile",
	Short: "Manage your machine's package file (brew/cask + winget/scoop/choco)",
	Long: `One file describes your machine's packages — Homebrew formulae and
casks on Linux/macOS, and winget/scoop/choco entries on Windows:

  brew "jq"
  cask "wezterm"
  winget "Microsoft.VisualStudioCode"
  scoop "ripgrep"

'dump' captures what's installed, 'add'/'remove' edit the file, 'install'
applies all of it, and 'list' shows the entries. The interactive menu
(Install Apps -> My Brewfile) offers the same plus per-package management.`,
}

var brewfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the entries in your Brewfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := brewfileTarget(cmd)
		pkgs, err := install.GetBrewfilePackages(path)
		if err != nil {
			return fmt.Errorf("no Brewfile at %s — create one with 'bluefin-cli brewfile dump'", path)
		}
		for _, p := range pkgs {
			fmt.Printf("%-7s %s\n", p.Kind, p.ID)
		}
		return nil
	},
}

var brewfileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a package entry (and optionally install it)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kind, _ := cmd.Flags().GetString("kind")
		return install.AddToBrewfile(brewfileTarget(cmd), args[0], kind)
	},
}

var brewfileRemoveCmd = &cobra.Command{
	Use:   "remove <name>...",
	Short: "Remove package entries from the file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return install.RemoveFromBrewfile(brewfileTarget(cmd), args)
	},
}

var brewfileSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search packages across this platform's managers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		results := install.SearchPackages(args[0])
		if len(results) == 0 {
			return fmt.Errorf("no packages found for %q", args[0])
		}
		for _, p := range results {
			line := fmt.Sprintf("%-7s %s", p.Kind, p.ID)
			if p.Name != "" && p.Name != p.ID {
				line += "  (" + p.Name + ")"
			}
			fmt.Println(line)
		}
		return nil
	},
}

var brewfileDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Write the installed packages into the file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return install.DumpBrewfile(brewfileTarget(cmd))
	},
}

var brewfileInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install everything in the file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := install.InstallBrewfileAll(brewfileTarget(cmd)); err != nil {
			return err
		}
		fmt.Println(tui.SuccessStyle.Render("✓ Brewfile applied."))
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{brewfileListCmd, brewfileAddCmd, brewfileRemoveCmd, brewfileDumpCmd, brewfileInstallCmd, brewfileSearchCmd} {
		c.Flags().String("file", "", "Brewfile path (default: ~/Brewfile)")
		brewfileCmd.AddCommand(c)
	}
	brewfileAddCmd.Flags().String("kind", "brew", "Entry kind: "+strings.Join(install.PackageKinds, ", "))
	rootCmd.AddCommand(brewfileCmd)
}
