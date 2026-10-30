package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/install"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

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
