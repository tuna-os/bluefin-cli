package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/status"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
	"github.com/tuna-os/bluefin-cli/internal/update"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Open the main menu of Bluefin",
	RunE: func(cmd *cobra.Command, args []string) error {
		registerPaletteActions()
		return app.Run(mainMenuScreen(), checkForUpdate)
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}

// checkForUpdate runs in the background when the menu starts and surfaces a
// toast when a newer release exists. Failures are silent — an update nudge
// is never worth an error message — and the GitHub API is only asked once
// per day.
func checkForUpdate() tea.Msg {
	if last := viper.GetInt64("update.last_check"); time.Since(time.Unix(last, 0)) < 24*time.Hour {
		return nil
	}
	viper.Set("update.last_check", time.Now().Unix())
	_ = config.Save()
	rel, err := update.Latest()
	if err != nil || !update.IsNewer(version, rel.TagName) {
		return nil
	}
	hint := "bluefin-cli update"
	if h := update.Detect().UpdateHint(); h != "" {
		hint = h
	}
	return app.ToastMsg{Text: fmt.Sprintf("⬆ %s available — run: %s", rel.TagName, hint)}
}

// launchFlow opens the interactive shell and immediately runs flow — used
// by commands like `bluefin-cli fonts` so every interactive path shares the
// native TUI (esc leads back to the Home menu).
func launchFlow(flow tea.Cmd) error {
	registerPaletteActions()
	return app.Run(mainMenuScreen(), flow, checkForUpdate)
}

func mainMenuScreen() app.Screen {
	return app.NewMenu("Home", nil, mainMenuItems, mainMenuSelect)
}

func mainMenuItems() []app.MenuItem {
	// Show exactly which shells have the experience enabled, not a vague
	// "Enabled" that may not match the user's current shell.
	var enabledShells []string
	for sh, enabled := range shell.CheckStatus() {
		if enabled {
			enabledShells = append(enabledShells, sh)
		}
	}
	sort.Strings(enabledShells)
	shellHint := "off"
	if len(enabledShells) > 0 {
		shellHint = strings.Join(enabledShells, " ") + " ✓"
	}

	items := []app.MenuItem{
		{Icon: "📊", Label: "Status", Value: "status", Desc: "Environment health at a glance"},
		{Icon: "🩺", Label: "Doctor", Value: "doctor", Desc: "Diagnose setup problems with fix hints"},
		{Icon: "🐚", Label: "Bluefin Shell", Value: "shell", Desc: "Aliases, prompt, and modern CLI tools", Hint: shellHint, Submenu: true},
		{Icon: "📦", Label: "Install Apps", Value: "bundles", Desc: "Curated tool bundles", Submenu: true},
		{Icon: "👻", Label: "Terminal Setup", Value: "terminal", Desc: "Ghostty install, Dock pin, themed config", Submenu: true},
	}
	items = append(items, extraMenuItems()...)
	items = append(items, app.MenuItem{Icon: "👋", Label: "Exit", Value: "exit", Desc: "Back to your shell"})
	return items
}

func mainMenuSelect(it app.MenuItem) tea.Cmd {
	switch it.Value {
	case "status":
		return app.Push(app.NewText("Status", status.Render))
	case "doctor":
		return doctorScreenCmd()
	case "shell":
		return app.Push(shellMenuScreen())
	case "bundles":
		return app.Push(bundlesMenuScreen())
	case "terminal":
		return app.Push(terminalMenuScreen())
	case "exit":
		return app.Pop()
	default:
		return extraMenuDo(it.Value)
	}
}

func currentShellName() string {
	name := filepath.Base(os.Getenv("SHELL"))
	if name == "" || name == "." {
		if env.IsWindows() {
			return "powershell"
		}
		return "bash"
	}
	return name
}

// runWithHomebrew runs task in a RunnerScreen, first settling the question of
// Homebrew if the task will need it.
//
// Tool installation needs brew, and the old code asked for it from inside the
// runner -- where Bubble Tea still holds the keyboard, so the confirm rendered
// nowhere and the runner span forever (tuna-os/bluefin-cli#273). The decision
// belongs here on the menu thread, where a form screen can own input; the
// install itself then runs inside the same runner as the task, so its output
// lands in one log.
//
// A failed or declined install is not fatal: the shell configuration still
// applies, and InstallTools reports the tools it had to skip.
func runWithHomebrew(title string, task func() error) tea.Cmd {
	if !homebrewConfirmNeeded() {
		return app.Push(app.NewRunner(title, task))
	}

	var install bool
	build := func() *huh.Form {
		install = false
		return huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Install Homebrew?").
				Description("Homebrew is missing, and bluefin-cli needs it to install the tools you enabled.\nAnswer no to configure the shell and skip the tools.").
				Value(&install),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.Push(app.NewForm("Homebrew", build, func(aborted bool) tea.Cmd {
		wanted := !aborted && install
		return app.Push(app.NewRunner(title, func() error {
			if wanted {
				if err := shell.InstallHomebrew(); err != nil {
					fmt.Println("Homebrew install failed, continuing without it:", err)
				}
			}
			return task()
		}))
	}))
}

// homebrewConfirmNeeded reports whether this machine would hit the missing-brew
// path. Windows installs through winget and Alpine through apk, so neither ever
// asks.
func homebrewConfirmNeeded() bool {
	if runtime.GOOS == "windows" || env.IsAlpine() {
		return false
	}
	return !shell.HomebrewAvailable()
}

func shellMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		current := currentShellName()
		toggle := fmt.Sprintf("Enable for current shell (%s)", current)
		if shell.CheckStatus()[current] {
			toggle = fmt.Sprintf("Disable for current shell (%s)", current)
		}
		return []app.MenuItem{
			{Icon: "🔄", Label: toggle, Value: "toggle_current", Desc: "One switch for the whole experience"},
			{Icon: "🔧", Label: "Configure Components", Value: "components", Desc: "Pick which tools load with your shell", Submenu: true},
			{Icon: "📰", Label: "MOTD Settings", Value: "motd", Desc: "Message of the day on new terminals", Submenu: true},
			{Icon: "🐚", Label: "Other Shells", Value: "shells", Desc: "Enable for bash, zsh, or fish", Submenu: true},
			{Icon: "🎨", Label: "Advanced", Value: "advanced", Desc: "Fine-grained shell integration", Submenu: true},
		}
	}
	return app.NewMenu("Shell", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "toggle_current":
			// Toggling can trigger brew installs (e.g. bash-preexec), which
			// print — so it must run in a RunnerScreen, never inline.
			current := currentShellName()
			enabled := shell.CheckStatus()[current]
			verb := "Enabling"
			if enabled {
				verb = "Disabling"
			}
			task := func() error { return shell.Toggle(current, !enabled) }
			if enabled {
				// Disabling only rewrites the rc file; it never installs.
				return app.Push(app.NewRunner(verb+" "+current, task))
			}
			return runWithHomebrew(verb+" "+current, task)
		case "components":
			return app.Push(componentsFormScreen())
		case "motd":
			return app.Push(motdMenuScreen())
		case "shells":
			return app.Push(shellsFormScreen())
		case "advanced":
			return app.Push(advancedMenuScreen())
		}
		return nil
	})
}

// componentsFormScreen hosts the shell-tools multiselect natively; the save
// is instant, and the brew-driven tool installation runs in a RunnerScreen.
func componentsFormScreen() app.Screen {
	current := currentShellName()
	var selected []string

	build := func() *huh.Form {
		cfg, err := shell.LoadConfig(current)
		if err != nil {
			cfg = shell.DefaultConfig(current)
		}
		tools := shell.ToolsForShell(current)
		selected = selected[:0]
		options := make([]huh.Option[string], 0, len(tools))
		for _, tool := range tools {
			if cfg.IsEnabled(tool.Name) {
				selected = append(selected, tool.Name)
			}
			options = append(options,
				huh.NewOption(fmt.Sprintf("%s (%s)", tool.Name, tool.Description), tool.Name))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select tools to enable").
				Description("Uncheck to disable specific tools").
				Options(options...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm("Components", build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		newCfg := shell.DefaultConfig(current)
		on := make(map[string]bool, len(selected))
		for _, s := range selected {
			on[s] = true
		}
		for _, tool := range shell.ToolsForShell(current) {
			newCfg.SetEnabled(tool.Name, on[tool.Name])
		}
		if err := shell.SaveConfig(newCfg); err != nil {
			return app.Toast("Error: "+err.Error(), true)
		}
		return tea.Sequence(
			runWithHomebrew("Installing tools", func() error {
				shell.InstallTools(current, newCfg)
				return nil
			}),
			app.Toast("Components saved.", false),
		)
	})
}

func availableBundleCategories() []bundleCategory {
	out := make([]bundleCategory, 0, len(bundleCategories))
	for _, cat := range bundleCategories {
		if cat.LinuxOnly && (!install.IsLinux() || !install.IsGnome()) {
			continue
		}
		out = append(out, cat)
	}
	return out
}

func bundlesMenuScreen() app.Screen {
	items := func() []app.MenuItem {
		cats := availableBundleCategories()
		out := make([]app.MenuItem, 0, len(cats))
		for _, cat := range cats {
			out = append(out, app.MenuItem{Label: cat.Label, Value: cat.ID, Desc: cat.Desc, Submenu: true})
		}
		return out
	}
	withCustom := func() []app.MenuItem {
		out := items()
		for _, name := range install.CustomBundles() {
			path, err := install.CustomBundlePath(name)
			if err != nil {
				continue
			}
			out = append(out, app.MenuItem{Icon: "🧰", Label: name, Value: "file:" + path,
				Desc: "Your bundle (~/.config/bluefin-cli/bundles)", Submenu: true})
		}
		// The `brew bundle` convention: the home Brewfile is a first-class
		// managed entity (create it from installed packages if absent).
		hb := install.HomeBrewfile()
		if _, err := os.Stat(hb); err == nil {
			out = append(out, app.MenuItem{Icon: "🏠", Label: "My Brewfile", Value: "mybrewfile",
				Desc: "Install, add, and remove entries in " + hb, Submenu: true})
		} else {
			out = append(out, app.MenuItem{Icon: "🏠", Label: "Create My Brewfile", Value: "dump",
				Desc: "Capture installed packages into " + hb})
		}
		return out
	}
	return app.NewMenu("Install Apps", nil, withCustom, func(it app.MenuItem) tea.Cmd {
		if path, ok := strings.CutPrefix(it.Value, "file:"); ok {
			return brewfileFlow(path, it.Label)
		}
		switch it.Value {
		case "mybrewfile":
			return app.Push(myBrewfileScreen())
		case "dump":
			return app.Push(app.NewRunner("Capturing installed packages", func() error {
				return install.DumpBrewfile(install.HomeBrewfile())
			}))
		}
		return packagesFlow(it.Value, it.Label)
	})
}

// myBrewfileScreen manages the home Brewfile: install all, per-package
// multiselect, add entries, remove entries, re-capture from installed.
func myBrewfileScreen() app.Screen {
	path := install.HomeBrewfile()
	items := func() []app.MenuItem {
		n := 0
		if pkgs, err := install.GetBrewfilePackages(path); err == nil {
			n = len(pkgs)
		}
		return []app.MenuItem{
			{Icon: "📥", Label: "Install everything", Value: "all", Desc: fmt.Sprintf("Apply all %d entries with one command", n)},
			{Icon: "🎛", Label: "Manage packages", Value: "manage", Desc: "Pick installs and removals individually", Submenu: true},
			{Icon: "➕", Label: "Add a package", Value: "add", Desc: "Append a formula or cask entry", Submenu: true},
			{Icon: "➖", Label: "Remove entries", Value: "remove", Desc: "Delete entries from the file", Submenu: true},
			{Icon: "📸", Label: "Re-capture from installed", Value: "dump", Desc: "Overwrite with what's installed now (brew bundle dump)"},
		}
	}
	return app.NewMenu("My Brewfile", nil, items, func(it app.MenuItem) tea.Cmd {
		switch it.Value {
		case "all":
			return app.Push(app.NewRunner("Installing Brewfile", func() error {
				return install.InstallBrewfileAll(path)
			}))
		case "manage":
			return brewfileFlow(path, "My Brewfile")
		case "add":
			return app.Push(brewfileAddScreen(path))
		case "remove":
			return app.Push(brewfileRemoveScreen(path))
		case "dump":
			return app.Push(app.NewRunner("Capturing installed packages", func() error {
				return install.DumpBrewfile(path)
			}))
		}
		return nil
	})
}

// brewfileAddScreen searches every package manager on this platform and
// adds the picked result to the Brewfile — type, search, choose.
func brewfileAddScreen(path string) app.Screen {
	var query string
	build := func() *huh.Form {
		query = ""
		return huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Search packages").
				Description("Searches brew formulae & casks (winget/scoop/choco on Windows)").
				Placeholder("e.g. ripgrep").
				Value(&query),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.NewForm("Add Package", build, func(aborted bool) tea.Cmd {
		q := strings.TrimSpace(query)
		if aborted || q == "" {
			return nil
		}
		return tea.Batch(
			app.Toast("Searching for "+q+"…", false),
			func() tea.Msg {
				results := install.SearchPackages(q)
				if len(results) == 0 {
					return app.ToastMsg{Text: "No packages found for " + q + ".", IsErr: true}
				}
				items := make([]app.MenuItem, 0, len(results))
				for _, p := range results {
					desc := p.Kind
					if p.Name != "" && p.Name != p.ID {
						desc = p.Name + " · " + p.Kind
					}
					items = append(items, app.MenuItem{Label: p.ID, Value: p.Kind + ":" + p.ID, Desc: desc})
				}
				menu := app.NewMenu("Results", items, nil, func(it app.MenuItem) tea.Cmd {
					kind, name, _ := strings.Cut(it.Value, ":")
					return tea.Sequence(func() tea.Msg {
						if err := install.AddToBrewfile(path, name, kind); err != nil {
							return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
						}
						return app.ToastMsg{Text: fmt.Sprintf("Added %s %q — Install everything to apply.", kind, name)}
					}, app.Pop())
				})
				return app.PushMsg{Screen: menu}
			},
		)
	})
}

func brewfileRemoveScreen(path string) app.Screen {
	var picked []string
	build := func() *huh.Form {
		picked = picked[:0]
		pkgs, _ := install.GetBrewfilePackages(path)
		opts := make([]huh.Option[string], 0, len(pkgs))
		for _, p := range pkgs {
			opts = append(opts, huh.NewOption(fmt.Sprintf("%s (%s)", p.ID, p.Kind), p.ID))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Remove from Brewfile").
				Description("Space marks entries for removal; enter confirms.").
				Options(opts...).
				Value(&picked),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}
	return app.NewForm("Remove Entries", build, func(aborted bool) tea.Cmd {
		if aborted || len(picked) == 0 {
			return nil
		}
		names := append([]string(nil), picked...)
		return func() tea.Msg {
			if err := install.RemoveFromBrewfile(path, names); err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			return app.ToastMsg{Text: fmt.Sprintf("Removed %d entr%s from the Brewfile.", len(names), map[bool]string{true: "y", false: "ies"}[len(names) == 1])}
		}
	})
}

// brewfileFlow opens any on-disk Brewfile as a managed multiselect —
// installed packages pre-checked, uncheck to uninstall, same diff+confirm
// flow as the curated bundles.
func brewfileFlow(path, label string) tea.Cmd {
	return tea.Batch(
		app.Toast("Reading "+label+"…", false),
		func() tea.Msg {
			pkgs, err := install.GetBrewfilePackages(path)
			if err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			if len(pkgs) == 0 {
				return app.ToastMsg{Text: label + " has no brew/cask entries.", IsErr: true}
			}
			pkgs = install.MarkInstalled(pkgs)
			return app.PushMsg{Screen: packagesFormScreen(label, pkgs)}
		},
	)
}

// packagesFlow loads a bundle in the background (the menu stays live), then
// pushes a native multiselect; confirmed changes run in a native RunnerScreen
// that captures brew output.
func packagesFlow(id, label string) tea.Cmd {
	return tea.Batch(
		app.Toast("Loading "+label+"…", false),
		func() tea.Msg {
			pkgs, err := install.GetBundlePackages(id)
			if err != nil {
				return app.ToastMsg{Text: "Error: " + err.Error(), IsErr: true}
			}
			pkgs = install.MarkInstalled(pkgs)
			return app.PushMsg{Screen: packagesFormScreen(label, pkgs)}
		},
	)
}

func packagesFormScreen(label string, pkgs []install.Package) app.Screen {
	var selected []string

	build := func() *huh.Form {
		selected = selected[:0]
		opts := make([]huh.Option[string], 0, len(pkgs))
		for _, p := range pkgs {
			if p.Installed {
				selected = append(selected, p.ID)
			}
			opts = append(opts, huh.NewOption(p.Name, p.ID))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select packages").
				Description("Pre-checked = already installed. Space toggles, enter confirms.").
				Options(opts...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm(label, build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		on := make(map[string]bool, len(selected))
		for _, id := range selected {
			on[id] = true
		}
		var toInstall, toRemove []install.Package
		for _, p := range pkgs {
			switch {
			case on[p.ID] && !p.Installed:
				toInstall = append(toInstall, p)
			case !on[p.ID] && p.Installed:
				toRemove = append(toRemove, p)
			}
		}
		if len(toInstall) == 0 && len(toRemove) == 0 {
			return app.Toast("No changes selected.", false)
		}
		return app.Push(confirmChangesScreen(toInstall, toRemove))
	})
}

func confirmChangesScreen(toInstall, toRemove []install.Package) app.Screen {
	var lines []string
	for _, p := range toInstall {
		lines = append(lines, "+ "+p.Name)
	}
	for _, p := range toRemove {
		lines = append(lines, "- "+p.Name)
	}
	confirmed := false

	build := func() *huh.Form {
		confirmed = false
		return huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Apply these changes?").
				Description(strings.Join(lines, "\n")).
				Value(&confirmed),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())
	}

	return app.NewForm("Confirm", build, func(aborted bool) tea.Cmd {
		if aborted || !confirmed {
			return nil
		}
		return app.Push(app.NewRunner("Applying changes", func() error {
			applyPackageChanges(toInstall, toRemove)
			return nil
		}))
	})
}

var paletteOnce sync.Once

// registerPaletteActions exposes every menu destination to the ctrl+p
// command palette.
func registerPaletteActions() {
	paletteOnce.Do(func() {
		app.Register(app.Action{
			ID: "status", Icon: "📊", Label: "Show Status", Section: "Home",
			Do: func() tea.Cmd { return mainMenuSelect(app.MenuItem{Value: "status"}) },
		})
		app.Register(app.Action{
			ID: "dino", Icon: "🦕", Label: "Dino Run", Section: "Fun",
			Do: func() tea.Cmd { return app.Push(gameScreen()) },
		})
		app.Register(app.Action{
			ID: "doctor", Icon: "🩺", Label: "Doctor", Section: "Home",
			Do: func() tea.Cmd { return doctorScreenCmd() },
		})
		app.Register(app.Action{
			ID: "terminal", Icon: "👻", Label: "Terminal Setup", Section: "Home",
			Do: func() tea.Cmd { return app.Push(terminalMenuScreen()) },
		})
		app.Register(app.Action{
			ID: "update", Icon: "⬆", Label: "Check for Updates", Section: "Home",
			Do: func() tea.Cmd {
				return app.Push(app.NewRunner("Update", func() error { return runUpdate(false) }))
			},
		})
		for _, cat := range availableBundleCategories() {
			id, label := cat.ID, cat.Label
			app.Register(app.Action{
				ID: "install-" + id, Label: label, Section: "Install",
				Do: func() tea.Cmd { return packagesFlow(id, label) },
			})
		}
		for _, it := range extraMenuItems() {
			it := it
			app.Register(app.Action{
				ID: it.Value, Icon: it.Icon, Label: it.Label, Section: "Customize",
				Do: func() tea.Cmd { return extraMenuDo(it.Value) },
			})
		}
	})
}
