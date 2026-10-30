package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

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
