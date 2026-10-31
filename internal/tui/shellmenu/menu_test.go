package shellmenu

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

func selectKey(t *testing.T, s app.Screen, key rune) tea.Cmd {
	t.Helper()
	_, cmd := s.Update(tea.KeyPressMsg{Code: key, Text: string(key)})
	if cmd == nil {
		t.Fatal("selection returned no command")
	}
	return cmd
}

func TestMOTDMenuRefreshAndPreview(t *testing.T) {
	enabled := true
	shown := false
	s := MOTD(func() bool { return enabled }, func() (bool, error) { return false, nil }, func() error { shown = true; return nil })
	s.Init()
	if !strings.Contains(s.View(80, 20), "Disable MOTD") {
		t.Fatal("enabled MOTD should offer disable")
	}
	enabled = false
	s.(app.Reloader).Reload()
	if !strings.Contains(s.View(80, 20), "Enable MOTD") {
		t.Fatal("reload should reflect the injected setting")
	}
	msg, ok := selectKey(t, s, '2')().(app.PushMsg)
	if !ok || msg.Screen.Title() != "MOTD" {
		t.Fatalf("preview should push MOTD runner, got %#v", msg)
	}
	if shown {
		t.Fatal("preview must not run before the runner starts")
	}
	driveRunner(t, msg.Screen)
	if !shown {
		t.Fatal("runner did not call preview")
	}
}

func TestMOTDToggleMessages(t *testing.T) {
	for _, tc := range []struct {
		name  string
		on    bool
		err   error
		text  string
		isErr bool
	}{
		{"enable", true, nil, "MOTD enabled.", false},
		{"disable", false, nil, "MOTD disabled.", false},
		{"error", false, errors.New("save failed"), "Error: save failed", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := motdToggleCmd(func() (bool, error) { return tc.on, tc.err })().(app.ToastMsg)
			if got.Text != tc.text || got.IsErr != tc.isErr {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

// Drive only the runner's wait command, not its animation tick, until it
// finishes. No real shell, MOTD process, or Starship installation is involved.
func driveRunner(t *testing.T, s app.Screen) {
	t.Helper()
	batch, ok := s.Init()().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatal("runner did not initialize wait and tick commands")
	}
	cmd := batch[0]
	for s.(app.InputCapturer).CapturingInput() {
		_, cmd = s.Update(cmd())
		if cmd == nil && s.(app.InputCapturer).CapturingInput() {
			t.Fatal("runner stopped before completion")
		}
	}
}

func TestStarshipSelectionRunsInjectedApply(t *testing.T) {
	var picked string
	msg := Starship(func(preset string) error { picked = preset; return nil })().(app.PushMsg)
	msg.Screen.Init()
	if !strings.Contains(msg.Screen.View(80, 40), "Tokyo Night") {
		t.Fatal("preset menu missing Tokyo Night")
	}
	runner := selectKey(t, msg.Screen, '2')().(app.PushMsg).Screen
	if runner.Title() != "Applying Tokyo Night" || picked != "" {
		t.Fatalf("unexpected runner %q or early application %q", runner.Title(), picked)
	}
	driveRunner(t, runner)
	if picked != "tokyo-night" {
		t.Fatalf("applied %q", picked)
	}
}

func TestStarshipApplyErrorAppearsInRunner(t *testing.T) {
	msg := Starship(func(string) error { return errors.New("preset failed") })().(app.PushMsg)
	msg.Screen.Init()
	runner := selectKey(t, msg.Screen, '1')().(app.PushMsg).Screen
	driveRunner(t, runner)
	if !strings.Contains(runner.View(80, 20), "preset failed") {
		t.Fatal("runner did not display the apply error")
	}
}

func TestShellsFormUsesInjectedStatus(t *testing.T) {
	calls := 0
	s := Shells(func() []string { return []string{"bash", "fish"} }, func() map[string]bool {
		calls++
		return map[string]bool{"bash": true}
	}, func(string, bool) error { t.Error("unexpected toggle"); return nil })
	s.Init()
	if calls != 1 || s.Title() != "Other Shells" {
		t.Fatalf("form did not query status: %d calls, title %q", calls, s.Title())
	}
}

func TestShellChanges(t *testing.T) {
	installed := []string{"bash", "fish", "zsh"}
	initial := map[string]bool{"bash": true, "zsh": true}
	changes := shellChanges(installed, []string{"fish", "zsh"}, initial)
	if len(changes) != 2 || changes[0] != (shellChange{"bash", false}) || changes[1] != (shellChange{"fish", true}) {
		t.Fatalf("unexpected changes: %#v", changes)
	}
	if len(shellChanges(installed, []string{"bash", "zsh"}, initial)) != 0 {
		t.Fatal("unchanged selections should produce no changes")
	}
	var applied []shellChange
	fail := errors.New("toggle failed")
	err := applyShellChanges(changes, func(name string, enabled bool) error {
		applied = append(applied, shellChange{name, enabled})
		return fail
	})
	if !errors.Is(err, fail) || len(applied) != 1 || applied[0] != changes[0] {
		t.Fatalf("failure should stop applying: %v, %v", applied, err)
	}
	applied = nil
	err = applyShellChanges(changes, func(name string, enabled bool) error {
		applied = append(applied, shellChange{name, enabled})
		return nil
	})
	if err != nil || len(applied) != 2 || applied[1] != changes[1] {
		t.Fatalf("unexpected applied changes: %v, %v", applied, err)
	}
}
