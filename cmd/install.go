package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var installCmd = &cobra.Command{
	Use:   "install [bundle]",
	Short: "Install tool bundles",
	Long: `Install predefined bundles or custom Brewfiles.

Available bundles:
  ai:               AI tools (Goose, Codex, Gemini, Ramalama, etc.)
  cli:              CLI essentials (gh, chezmoi, etc.)
  cncf:             Cloud Native Computing Foundation tools.
  experimental-ide: Experimental IDE tools.
  ide:              IDE tools: VS Code, JetBrains Toolbox, etc.
  k8s:              Kubernetes tools: kubectl, k9s, kubectx, etc.
  
Or provide a path to a local Brewfile.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return launchFlow(app.Push(bundlesMenuScreen()))
		}

		return install.Bundle(args[0])
	},
}

var installListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the available bundles",
	Long:  `Show all available bundles with descriptions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		install.ListBundles()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	installCmd.AddCommand(installListCmd)
	installCmd.AddCommand(installWallpapersCmd)
	installWallpapersCmd.AddCommand(installWallpapersCleanupCmd)

	installWallpapersCmd.Flags().Bool("non-interactive", false, "Skip prompts and use flag values")
	installWallpapersCmd.Flags().Bool("yes", false, "Non-interactive shortcut: run sunset setup after install")
	installWallpapersCleanupCmd.Flags().Bool("all", false, "Also uninstall known wallpaper casks and remove local wallpaper folders")
}

var installWallpapersCmd = &cobra.Command{
	Use:   "wallpapers [cask...]",
	Short: "Install wallpaper casks from ublue-os/tap",
	Long:  "Install wallpapers published as Homebrew casks from the ublue-os/tap tap.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := install.InstallWallpaperCasks(args); err != nil {
				return err
			}
			return maybeHandleWindowsThemePostInstall(cmd, args)
		}

		return launchFlow(wallpapersFlow())
	},
}

var installWallpapersCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up the files of the wallpaper sync",
	Long:  "Remove the files that the wallpaper sync of Bluefin CLI made. In WSL this removes generated Windows themes, copied wallpaper folders, helper scripts, scheduled tasks, and state. Use --all to also uninstall the known casks of wallpapers and remove local wallpaper folders.",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if err := install.CleanupWallpapers(all); err != nil {
			return err
		}

		if all {
			fmt.Println(tui.SuccessStyle.Render("✓ Wallpaper cleanup complete (including installed casks/local wallpaper folders)."))
		} else {
			fmt.Println(tui.SuccessStyle.Render("✓ Wallpaper sync cleanup complete."))
		}

		return nil
	},
}

type bundleCategory struct {
	Label     string
	ID        string
	Desc      string
	LinuxOnly bool
}

var bundleCategories = []bundleCategory{
	{Label: "🤖 AI Tools", ID: "ai", Desc: "Coding agents & LLM tools"},
	{Label: "💻 CLI Essentials", ID: "cli", Desc: "Everyday terminal essentials"},
	{Label: "🌐 CNCF Tools", ID: "cncf", Desc: "Cloud-native toolchain"},
	{Label: "🧪 Experimental IDE", ID: "experimental-ide", Desc: "Bleeding-edge editors"},
	{Label: "📝 IDE Tools", ID: "ide", Desc: "Editors & IDEs"},
	{Label: "🎡 Kubernetes Tools", ID: "k8s", Desc: "Kubernetes workflow"},
	{Label: "🐧 Full GNOME Desktop", ID: "full-desktop", Desc: "Complete desktop environment", LinuxOnly: true},
}

// runPackageMenu shows a per-category multi-select with installed packages pre-checked,
// then diffs and applies installs/uninstalls with confirmation. Works on both Unix (brew)
// and Windows (winget).
// applyPackageChanges installs and removes packages with the platform's
// package manager, printing errors as it goes (used by both the legacy
// flow and the native menu flow via the legacy bridge).
func applyPackageChanges(toInstall, toRemove []install.Package) {
	if env.IsWindows() {
		if len(toInstall) > 0 {
			var winPkgs []install.WindowsPackage
			for _, p := range toInstall {
				winPkgs = append(winPkgs, install.WindowsPackage{ID: p.ID, Name: p.Name})
			}
			if err := install.InstallWindowsPackages(winPkgs); err != nil {
				fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Install error: %v", err)))
			}
		}
		if len(toRemove) > 0 {
			ids := make([]string, 0, len(toRemove))
			for _, p := range toRemove {
				ids = append(ids, p.ID)
			}
			if err := install.UninstallWingetPackages(ids); err != nil {
				fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Uninstall error: %v", err)))
			}
		}
	} else {
		if len(toInstall) > 0 {
			if err := install.InstallBrewPackages(toInstall); err != nil {
				fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Install error: %v", err)))
			}
		}
		if len(toRemove) > 0 {
			ids := make([]string, 0, len(toRemove))
			for _, p := range toRemove {
				ids = append(ids, p.Name)
			}
			if err := install.UninstallBrewPackages(ids); err != nil {
				fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Uninstall error: %v", err)))
			}
		}
	}

}
