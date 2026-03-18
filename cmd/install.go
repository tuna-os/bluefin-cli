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
  fonts            - Development fonts (Fira Code, JetBrains Mono, etc.)
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

func runBundlesMenu() error {
	var selectedBundles []string

	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Main Menu > Install Apps")

		var selectedBundle string

		opts := []huh.Option[string]{
			huh.NewOption("🤖 AI Tools", "ai"),
			huh.NewOption("💻 CLI Essentials", "cli"),
			huh.NewOption("☁️ CNCF Tools", "cncf"),
			huh.NewOption("🧪 Experimental IDE", "experimental-ide"),
			huh.NewOption("📝 IDE Tools", "ide"),
			huh.NewOption("☸️ Kubernetes Tools", "k8s"),
		}

		if env.IsWindows() {
			opts = []huh.Option[string]{
				huh.NewOption("AI Tools", "ai"),
				huh.NewOption("CLI Essentials", "cli"),
				huh.NewOption("CNCF Tools", "cncf"),
				huh.NewOption("Experimental IDE", "experimental-ide"),
				huh.NewOption("IDE Tools", "ide"),
				huh.NewOption("Kubernetes Tools", "k8s"),
			}
		}

		if install.IsLinux() && install.IsGnome() {
			if env.IsWindows() {
				opts = append(opts, huh.NewOption("Full GNOME Desktop", "full-desktop"))
			} else {
				opts = append(opts, huh.NewOption("🖥️ Full GNOME Desktop", "full-desktop"))
			}
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a bundle to install").
					Options(opts...).
					Value(&selectedBundle),
			),
		).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				return nil
			}
			return fmt.Errorf("form error: %w", err)
		}

		selectedBundles = []string{selectedBundle}

		if env.IsWindows() {
			packages, err := install.WindowsPackagesForBundles(selectedBundles)
			if err != nil {
				return err
			}
			if len(packages) == 0 {
				return fmt.Errorf("no Windows packages available for selected bundles")
			}

			selectedPackages := []string{}
			opts := make([]huh.Option[string], 0, len(packages))
			for _, pkg := range packages {
				label := pkg.Name
				if strings.TrimSpace(label) == "" {
					label = pkg.ID
				}
				desc := strings.TrimSpace(pkg.Description)
				if desc == "" {
					desc = pkg.ID
				}
				opts = append(opts, huh.NewOption(fmt.Sprintf("%s - %s", label, desc), pkg.ID))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Select packages to install").
						Description("Space toggles package selection. Enter to continue.").
						Options(opts...).
						Value(&selectedPackages),
				),
			).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

			if err := form.Run(); err != nil {
				if err == huh.ErrUserAborted {
					continue
				}
				return fmt.Errorf("form error: %w", err)
			}

			if len(selectedPackages) == 0 {
				return fmt.Errorf("no packages selected")
			}

			selectedSet := make(map[string]bool, len(selectedPackages))
			for _, id := range selectedPackages {
				selectedSet[id] = true
			}

			finalPackages := make([]install.WindowsPackage, 0, len(selectedPackages))
			for _, pkg := range packages {
				if selectedSet[pkg.ID] {
					finalPackages = append(finalPackages, pkg)
				}
			}

			return install.InstallWindowsPackages(finalPackages)
		}

		break
	}

	var brewfiles []string
	var cleanups []func()

	defer func() {
		for _, c := range cleanups {
			c()
		}
	}()

	for _, bundle := range selectedBundles {
		path, cleanup, err := install.GetBrewfile(bundle)
		if err != nil {
			return err
		}
		brewfiles = append(brewfiles, path)
		cleanups = append(cleanups, cleanup)
	}

	if len(brewfiles) > 0 {
		if err := install.EnsureBbrew(); err != nil {
			return err
		}

		var finalPath string
		if len(brewfiles) > 1 {
			mergedPath, cleanup, err := install.MergeBrewfiles(brewfiles)
			if err != nil {
				return err
			}
			cleanups = append(cleanups, cleanup)
			finalPath = mergedPath
			fmt.Println(tui.InfoStyle.Render("🍺 Merged Brewfiles into single view..."))
		} else {
			finalPath = brewfiles[0]
		}

		fmt.Println(tui.InfoStyle.Render("🍺 Opening apps in bbrew..."))
		if err := install.RunBbrew(finalPath); err != nil {
			return err
		}
	}

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
