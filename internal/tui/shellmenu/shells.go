package shellmenu

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// Shells lets users change installed shells through the supplied operations.
func Shells(installedShells func() []string, checkStatus func() map[string]bool, toggle func(string, bool) error) app.Screen {
	var installed []string
	var selected []string
	initial := map[string]bool{}

	build := func() *huh.Form {
		installed = installedShells()
		status := checkStatus()
		selected = selected[:0]
		options := make([]huh.Option[string], 0, len(installed))
		for _, sh := range installed {
			initial[sh] = status[sh]
			if status[sh] {
				selected = append(selected, sh)
			}
			options = append(options, huh.NewOption(sh, sh))
		}
		return huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Manage other shells").
				Description("Selected = enabled. Space toggles, enter applies.").
				Options(options...).
				Value(&selected),
		)).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())
	}

	return app.NewForm("Other Shells", build, func(aborted bool) tea.Cmd {
		if aborted {
			return nil
		}
		changes := shellChanges(installed, selected, initial)
		if len(changes) == 0 {
			return app.Toast("No changes.", false)
		}
		return app.Push(app.NewRunner("Applying", func() error {
			return applyShellChanges(changes, toggle)
		}))
	})
}

type shellChange struct {
	name    string
	enabled bool
}

func shellChanges(installed, selected []string, initial map[string]bool) []shellChange {
	want := make(map[string]bool, len(selected))
	for _, sh := range selected {
		want[sh] = true
	}
	var changes []shellChange
	for _, sh := range installed {
		if want[sh] != initial[sh] {
			changes = append(changes, shellChange{sh, want[sh]})
		}
	}
	return changes
}

func applyShellChanges(changes []shellChange, toggle func(string, bool) error) error {
	for _, change := range changes {
		fmt.Printf("%s -> %v\n", change.name, change.enabled)
		if err := toggle(change.name, change.enabled); err != nil {
			return err
		}
	}
	return nil
}
