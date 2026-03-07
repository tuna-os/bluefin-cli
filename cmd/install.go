package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/install"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [bundle]",
	Short: "Install tool bundles",
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
	installCmd.AddCommand(installWallpapersCmd)
	installWallpapersCmd.AddCommand(installWallpapersCleanupCmd)

	installWallpapersCmd.Flags().Bool("non-interactive", false, "Skip prompts and use flag values")
	installWallpapersCmd.Flags().Bool("yes", false, "Non-interactive shortcut: apply theme + enable mode sync + enable 6 AM/6 PM switching")
	installWallpapersCmd.Flags().Bool("apply-theme", false, "Apply a Windows theme after registration (WSL only)")
	installWallpapersCmd.Flags().String("theme", "", "Theme name to apply in non-interactive mode (Bluefin, Aurora, Bazzite)")
	installWallpapersCmd.Flags().Bool("enable-mode-sync", false, "Enable day/night wallpaper sync task in non-interactive mode")
	installWallpapersCmd.Flags().Bool("enable-auto-dark-light", false, "Enable 6 AM/6 PM light/dark switching tasks (requires --enable-mode-sync)")
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

func maybeHandleWindowsThemePostInstall(cmd *cobra.Command, casks []string) error {
	if cmd == nil {
		return maybePromptForWindowsTheme(casks)
	}

	if !env.IsWSL() {
		return nil
	}

	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	yes, _ := cmd.Flags().GetBool("yes")
	applyTheme, _ := cmd.Flags().GetBool("apply-theme")
	selectedTheme, _ := cmd.Flags().GetString("theme")
	enableModeSync, _ := cmd.Flags().GetBool("enable-mode-sync")
	enableAutoDarkLight, _ := cmd.Flags().GetBool("enable-auto-dark-light")

	if yes {
		nonInteractive = true
		applyTheme = true
		enableModeSync = true
		enableAutoDarkLight = true
	}

	if enableAutoDarkLight && !enableModeSync {
		return fmt.Errorf("--enable-auto-dark-light requires --enable-mode-sync")
	}

	flagsRequestedAutomation := cmd.Flags().Changed("enable-mode-sync") || cmd.Flags().Changed("enable-auto-dark-light")
	flagsRequestedThemeApply := cmd.Flags().Changed("apply-theme") || cmd.Flags().Changed("theme")
	if !nonInteractive && !flagsRequestedAutomation && !flagsRequestedThemeApply {
		return maybePromptForWindowsTheme(casks)
	}

	themes := install.ThemesFromWallpaperCasks(casks)
	if len(themes) == 0 {
		return nil
	}

	if applyTheme {
		if strings.TrimSpace(selectedTheme) == "" {
			selectedTheme = themes[0]
		}

		if !containsTheme(themes, selectedTheme) {
			return fmt.Errorf("theme %q not found in installed wallpaper casks (available: %s)", selectedTheme, strings.Join(themes, ", "))
		}

		if err := install.ApplyWindowsTheme(selectedTheme); err != nil {
			return err
		}

		if err := install.SetWindowsThemePreference(selectedTheme, false); err != nil {
			return err
		}

		fmt.Println(tui.SuccessStyle.Render("✓ Applied Windows theme: " + selectedTheme))
		fmt.Println(tui.InfoStyle.Render("Monthly wallpaper updates are enabled for supported themes."))
	}

	if enableModeSync {
		if err := install.ConfigureWindowsThemeAutomation(enableAutoDarkLight); err != nil {
			return err
		}

		if enableAutoDarkLight {
			fmt.Println(tui.SuccessStyle.Render("✓ Enabled theme mode sync + 6 AM/6 PM auto light/dark switching"))
		} else {
			fmt.Println(tui.SuccessStyle.Render("✓ Enabled theme mode sync task"))
		}
	}

	return nil
}

func containsTheme(themes []string, theme string) bool {
	for _, candidate := range themes {
		if strings.EqualFold(candidate, theme) {
			return true
		}
	}

	return false
}

func runBundlesMenu() error {
	selectedBundles := []string{}

	for {
		tui.ClearScreen()
		tui.RenderHeader("Bluefin CLI", "Main Menu > Install Apps")

		var selectedBundle string

		opts := []huh.Option[string]{
			huh.NewOption("🤖 AI Tools", "ai"),
			huh.NewOption("💻 CLI Essentials", "cli"),
			huh.NewOption("☁️  CNCF Tools", "cncf"),
			huh.NewOption("🧪 Experimental IDE", "experimental-ide"),
			huh.NewOption("🔤 Development Fonts", "fonts"),
			huh.NewOption("📝 IDE Tools", "ide"),
			huh.NewOption("☸️  Kubernetes Tools", "k8s"),
		}

		if env.IsWindows() {
			opts = []huh.Option[string]{
				huh.NewOption("AI Tools", "ai"),
				huh.NewOption("CLI Essentials", "cli"),
				huh.NewOption("CNCF Tools", "cncf"),
				huh.NewOption("Experimental IDE", "experimental-ide"),
				huh.NewOption("Development Fonts", "fonts"),
				huh.NewOption("IDE Tools", "ide"),
				huh.NewOption("Kubernetes Tools", "k8s"),
			}
		}

		if install.IsLinux() && install.IsGnome() {
			if env.IsWindows() {
				opts = append(opts, huh.NewOption("Full GNOME Desktop", "full-desktop"))
			} else {
				opts = append(opts, huh.NewOption("🖥️  Full GNOME Desktop", "full-desktop"))
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

			selectedPackages := make([]string, 0, len(packages))
			opts := make([]huh.Option[string], 0, len(packages))
			for _, pkg := range packages {
				selectedPackages = append(selectedPackages, pkg.ID)
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
						Title("Select packages to install (preselected from bundles)").
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

		fmt.Println(tui.InfoStyle.Render(fmt.Sprintf("🍺 Opening apps in bbrew...")))
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
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select wallpapers to install (space to select, enter to confirm)").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}
	if len(selected) == 0 {
		return fmt.Errorf("no wallpapers selected")
	}
	if err := install.InstallWallpaperCasks(selected); err != nil {
		return err
	}

	return maybePromptForWindowsTheme(selected)
}

func maybePromptForWindowsTheme(casks []string) error {
	if !env.IsWSL() {
		return nil
	}

	themes := install.ThemesFromWallpaperCasks(casks)
	if len(themes) == 0 {
		return nil
	}

	var applyNow bool
	confirm := huh.NewConfirm().
		Title("Set an installed Windows theme now and keep supported wallpapers updated monthly?").
		Description("If no, themes are only registered in Windows settings.").
		Value(&applyNow)

	if err := confirm.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return err
	}

	if !applyNow {
		return maybeConfigureWindowsThemeAutomation()
	}

	selectedTheme := themes[0]
	if len(themes) > 1 {
		themeOptions := make([]huh.Option[string], 0, len(themes))
		for _, theme := range themes {
			themeOptions = append(themeOptions, huh.NewOption(theme, theme))
		}

		picker := huh.NewSelect[string]().
			Title("Choose the Windows theme to apply").
			Options(themeOptions...).
			Value(&selectedTheme)

		if err := huh.NewForm(huh.NewGroup(picker)).WithTheme(tui.AppTheme).Run(); err != nil {
			if err == huh.ErrUserAborted {
				return nil
			}
			return err
		}
	}

	if err := install.ApplyWindowsTheme(selectedTheme); err != nil {
		return err
	}

	if err := install.SetWindowsThemePreference(selectedTheme, false); err != nil {
		return err
	}

	fmt.Println(tui.SuccessStyle.Render("✓ Applied Windows theme: " + selectedTheme))
	fmt.Println(tui.InfoStyle.Render("Monthly wallpaper updates are enabled for supported themes."))

	return maybeConfigureWindowsThemeAutomation()
}

func maybeConfigureWindowsThemeAutomation() error {
	var enableModeSync bool
	modeSyncConfirm := huh.NewConfirm().
		Title("Enable wallpaper day/night sync when Windows light/dark theme changes?").
		Description("If a day/night variant exists for the same wallpaper name, it will switch accordingly.").
		Value(&enableModeSync)

	if err := modeSyncConfirm.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return err
	}

	if !enableModeSync {
		return nil
	}

	var enableAutoDarkLight bool
	autoConfirm := huh.NewConfirm().
		Title("Enable automatic dark/light theme switching at 6:00 AM and 6:00 PM?").
		Description("Registers Windows scheduled tasks to set light mode at 6 AM and dark mode at 6 PM.").
		Value(&enableAutoDarkLight)

	if err := autoConfirm.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return err
	}

	if err := install.ConfigureWindowsThemeAutomation(enableAutoDarkLight); err != nil {
		return err
	}

	if enableAutoDarkLight {
		fmt.Println(tui.SuccessStyle.Render("✓ Enabled theme mode sync + 6 AM/6 PM auto light/dark switching"))
	} else {
		fmt.Println(tui.SuccessStyle.Render("✓ Enabled theme mode sync task"))
	}

	return nil
}
