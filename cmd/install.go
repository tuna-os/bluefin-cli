package cmd

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install [bundle]",
	Short:   "Install tool bundles",
	Long: `Install predefined bundles or custom Brewfiles.

Available bundles:
  ai               - AI tools (Goose, Codex, Gemini, Ramalama, etc.)
  cli              - CLI essentials (gh, chezmoi, etc.)
  cncf             - Cloud Native Computing Foundation tools.
  experimental-ide - Experimental IDE tools.
  ide              - IDE tools: VS Code, JetBrains Toolbox, etc.
  k8s              - Kubernetes tools: kubectl, k9s, kubectx, etc.
  
Or provide a path to a local Brewfile.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runBundlesMenu()
		}

		return install.Bundle(args[0])
	},
}

var installListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available bundles",
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
	rootCmd.AddCommand(installWallpapersCmd)
	rootCmd.AddCommand(installWallpapersCleanupCmd)

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

		return runWallpapersMenu()
	},
}

var installWallpapersCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean wallpaper sync artifacts",
	Long:  "Remove Bluefin CLI wallpaper sync artifacts. In WSL this removes generated Windows themes, copied wallpaper folders, helper scripts, scheduled tasks, and state. Use --all to also uninstall known wallpaper casks and remove local wallpaper folders.",
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
	Label    string
	ID       string
	LinuxOnly bool
}

var bundleCategories = []bundleCategory{
	{Label: "🤖 AI Tools", ID: "ai"},
	{Label: "💻 CLI Essentials", ID: "cli"},
	{Label: "🌐 CNCF Tools", ID: "cncf"},
	{Label: "🧪 Experimental IDE", ID: "experimental-ide"},
	{Label: "📝 IDE Tools", ID: "ide"},
	{Label: "🎡 Kubernetes Tools", ID: "k8s"},
	{Label: "🐧 Full GNOME Desktop", ID: "full-desktop", LinuxOnly: true},
}

func runBundlesMenu() error {
	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Main Menu > Install Apps")

		opts := make([]huh.Option[string], 0)
		for _, cat := range bundleCategories {
			if cat.LinuxOnly && !(install.IsLinux() && install.IsGnome()) {
				continue
			}
			opts = append(opts, huh.NewOption(cat.Label+" ❯", cat.ID))
		}

		var category string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a category").
					Options(opts...).
					Value(&category),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				return nil
			}
			return fmt.Errorf("form error: %w", err)
		}

		if err := runPackageMenu(category); err != nil {
			if err == huh.ErrUserAborted {
				continue
			}
			fmt.Println(tui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)))
			tui.Pause()
		}
	}
}

// runPackageMenu shows a per-category multi-select with installed packages pre-checked,
// then diffs and applies installs/uninstalls with confirmation. Works on both Unix (brew)
// and Windows (winget).
func runPackageMenu(bundleName string) error {
	tui.ClearScreen()

	var categoryLabel string
	for _, cat := range bundleCategories {
		if cat.ID == bundleName {
			categoryLabel = cat.Label
			break
		}
	}
	tui.RenderHeader("Bluefin CLI", "Install Apps > "+categoryLabel)

	fmt.Println(tui.InfoStyle.Render("Loading packages..."))
	pkgs, err := install.GetBundlePackages(bundleName)
	if err != nil {
		return fmt.Errorf("could not load bundle: %w", err)
	}

	pkgs = install.MarkInstalled(pkgs)

	// Pre-populate selection with currently installed packages
	preSelected := make([]string, 0)
	for _, p := range pkgs {
		if p.Installed {
			preSelected = append(preSelected, p.ID)
		}
	}

	opts := make([]huh.Option[string], 0, len(pkgs))
	for _, p := range pkgs {
		label := p.Name
		if p.Installed {
			label += " ✓"
		}
		opts = append(opts, huh.NewOption(label, p.ID))
	}

	selected := make([]string, len(preSelected))
	copy(selected, preSelected)

	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Install Apps > "+categoryLabel)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select packages (✓ = installed)").
				Description("Space toggles. Enter confirms.").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		return err
	}

	// Diff: what changed vs pre-installed state
	preSet := make(map[string]bool, len(preSelected))
	for _, id := range preSelected {
		preSet[id] = true
	}
	newSet := make(map[string]bool, len(selected))
	for _, id := range selected {
		newSet[id] = true
	}

	var toInstall, toRemove []install.Package
	for _, p := range pkgs {
		wasInstalled := preSet[p.ID]
		isSelected := newSet[p.ID]
		switch {
		case isSelected && !wasInstalled:
			toInstall = append(toInstall, p)
		case !isSelected && wasInstalled:
			toRemove = append(toRemove, p)
		}
	}

	if len(toInstall) == 0 && len(toRemove) == 0 {
		fmt.Println(tui.InfoStyle.Render("No changes selected."))
		tui.Pause()
		return nil
	}

	// Confirmation
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Install Apps > "+categoryLabel+" > Confirm")

	if len(toInstall) > 0 {
		fmt.Println(tui.SuccessStyle.Render("Will install:"))
		for _, p := range toInstall {
			fmt.Printf("  + %s\n", p.Name)
		}
	}
	if len(toRemove) > 0 {
		fmt.Println(tui.ErrorStyle.Render("Will uninstall:"))
		for _, p := range toRemove {
			fmt.Printf("  - %s\n", p.Name)
		}
	}
	fmt.Println()

	var confirmed bool
	confirm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply these changes?").
				Value(&confirmed),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())

	if err := confirm.Run(); err != nil || !confirmed {
		return nil
	}

	// Execute
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Install Apps > "+categoryLabel+" > Installing")

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

	fmt.Println()
	fmt.Println(tui.SuccessStyle.Render("✓ Done!"))
	tui.Pause()
	return nil
}

func runWallpapersMenu() error {
	tui.ClearScreen()
	tui.RenderHeader("Bluefin CLI", "Main Menu > Wallpapers")
	casks, err := install.GetWallpaperCasks()
	if err != nil {
		return fmt.Errorf("failed to discover wallpaper casks: %w", err)
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks found in ublue-os/tap")
	}

	opts := make([]huh.Option[string], 0, len(casks))
	for _, c := range casks {
		opts = append(opts, huh.NewOption(c, c))
	}

	var selected []string
	wallpaperSelect := huh.NewMultiSelect[string]().
		Title("Select wallpapers to install").
		Description("Space toggles selections. Enter confirms. If none selected, Enter installs the highlighted item.").
		Options(opts...).
		Value(&selected)

	form := huh.NewForm(
		huh.NewGroup(
			wallpaperSelect,
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	if len(selected) == 0 {
		hovered, ok := wallpaperSelect.Hovered()
		if !ok || strings.TrimSpace(hovered) == "" {
			return fmt.Errorf("no wallpapers selected")
		}
		selected = []string{hovered}
	}

	if err := install.InstallWallpaperCasks(selected); err != nil {
		return err
	}

	return maybeHandleWindowsThemePostInstall(nil, selected)
}

func supportsWindowsThemePostInstall() bool {
	return env.IsWSL() || env.IsWindows()
}
