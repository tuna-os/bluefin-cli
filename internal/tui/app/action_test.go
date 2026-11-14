package app

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func dummyCmd() tea.Cmd {
	return func() tea.Msg { return nil }
}

func TestRegistry_Register_Valid(t *testing.T) {
	r := NewRegistry()
	err := r.Register(
		Action{ID: "act-1", Label: "Action 1", Do: dummyCmd},
		Action{ID: "act-2", Label: "Action 2", Do: dummyCmd},
	)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if count := r.Count(); count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	actions := r.Actions()
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].ID != "act-1" || actions[1].ID != "act-2" {
		t.Errorf("unexpected actions: %+v", actions)
	}
}

func TestRegistry_Register_EmptyID(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Action{ID: "", Label: "No ID", Do: dummyCmd})
	if !errors.Is(err, ErrEmptyActionID) {
		t.Errorf("expected ErrEmptyActionID, got %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 actions registered on error, got %d", r.Count())
	}
}

func TestRegistry_Register_NilDo(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Action{ID: "act-nil", Label: "Nil Callback", Do: nil})
	if !errors.Is(err, ErrNilActionDo) {
		t.Errorf("expected ErrNilActionDo, got %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 actions registered on error, got %d", r.Count())
	}
}

func TestRegistry_Register_DuplicateWithinCall(t *testing.T) {
	r := NewRegistry()
	err := r.Register(
		Action{ID: "act-dup", Label: "First", Do: dummyCmd},
		Action{ID: "act-dup", Label: "Second", Do: dummyCmd},
	)
	if !errors.Is(err, ErrDuplicateAction) {
		t.Errorf("expected ErrDuplicateAction, got %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 actions registered on error, got %d", r.Count())
	}
}

func TestRegistry_Register_DuplicateAcrossCalls(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(Action{ID: "act-1", Label: "First", Do: dummyCmd}); err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}

	err := r.Register(Action{ID: "act-1", Label: "Second", Do: dummyCmd})
	if !errors.Is(err, ErrDuplicateAction) {
		t.Errorf("expected ErrDuplicateAction, got %v", err)
	}
	if r.Count() != 1 {
		t.Errorf("expected 1 action registered, got %d", r.Count())
	}
}

func TestRegistry_Actions_AvailableFilter(t *testing.T) {
	r := NewRegistry()
	available := true
	err := r.Register(
		Action{ID: "act-always", Label: "Always", Do: dummyCmd},
		Action{
			ID:        "act-cond",
			Label:     "Conditional",
			Available: func() bool { return available },
			Do:        dummyCmd,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(r.Actions()) != 2 {
		t.Errorf("expected 2 actions when condition is true, got %d", len(r.Actions()))
	}
	if len(r.All()) != 2 {
		t.Errorf("expected 2 actions in All(), got %d", len(r.All()))
	}

	available = false
	actions := r.Actions()
	if len(actions) != 1 || actions[0].ID != "act-always" {
		t.Errorf("expected only 'act-always' when condition is false, got %+v", actions)
	}
	if len(r.All()) != 2 {
		t.Errorf("expected 2 actions in All() regardless of availability, got %d", len(r.All()))
	}
}

func TestRegistry_ResetAndSnapshotRestore(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(
		Action{ID: "a", Label: "A", Do: dummyCmd},
		Action{ID: "b", Label: "B", Do: dummyCmd},
	)

	snap := r.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected snapshot of length 2, got %d", len(snap))
	}

	r.Reset()
	if r.Count() != 0 {
		t.Errorf("expected 0 after Reset, got %d", r.Count())
	}

	if err := r.Restore(snap); err != nil {
		t.Fatalf("failed to restore: %v", err)
	}
	if r.Count() != 2 {
		t.Errorf("expected 2 after Restore, got %d", r.Count())
	}

	// Restoring invalid slice should fail and leave existing state untouched
	bad := []Action{
		{ID: "", Label: "bad", Do: dummyCmd},
	}
	if err := r.Restore(bad); err == nil {
		t.Error("expected error when restoring invalid action, got nil")
	}
	if r.Count() != 2 {
		t.Errorf("expected registry untouched after failed restore, got count %d", r.Count())
	}
}

func TestPackageRegister_PanicsOnMalformed(t *testing.T) {
	orig := SnapshotRegistry()
	ResetRegistry()
	t.Cleanup(func() { _ = RestoreRegistry(orig) })

	assertPanic := func(name string, f func()) {
		t.Helper()
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("%s: expected panic, got none", name)
			} else if !strings.Contains(r.(string), "app.Register failed") {
				t.Errorf("%s: unexpected panic message: %v", name, r)
			}
		}()
		f()
	}

	assertPanic("empty ID", func() {
		Register(Action{ID: "", Label: "Empty", Do: dummyCmd})
	})

	assertPanic("nil callback", func() {
		Register(Action{ID: "valid-id", Label: "Nil Do", Do: nil})
	})

	Register(Action{ID: "unique-id", Label: "Valid", Do: dummyCmd})
	assertPanic("duplicate ID", func() {
		Register(Action{ID: "unique-id", Label: "Duplicate", Do: dummyCmd})
	})
}

func TestPaletteCannotOverwriteEarlierAction(t *testing.T) {
	r := NewRegistry()
	firstExecuted := false
	secondExecuted := false

	// If a registry could have duplicates (e.g. bypass or multiple actions in custom palette constructor)
	// ensure palette construction / selection respects the first registered action.
	r.actions = []Action{
		{
			ID:    "test-cmd",
			Label: "Original Action",
			Do: func() tea.Cmd {
				firstExecuted = true
				return nil
			},
		},
		{
			ID:    "test-cmd",
			Label: "Shadowing Action",
			Do: func() tea.Cmd {
				secondExecuted = true
				return nil
			},
		},
	}
	r.byID = map[string]int{"test-cmd": 0}

	screen := newPaletteFromRegistry(r)
	screen.Init()
	menu, ok := screen.(*MenuScreen)
	if !ok {
		t.Fatalf("expected MenuScreen, got %T", screen)
	}

	items := menu.items
	if len(items) != 1 {
		t.Fatalf("expected exactly 1 item in palette to prevent duplicate display, got %d", len(items))
	}
	if items[0].Label != "Original Action" {
		t.Errorf("expected 'Original Action', got %q", items[0].Label)
	}

	// Trigger action select
	cmd := menu.onSelect(items[0])
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd from onSelect")
	}

	// Execute cmd batch or sequence
	// onSelect returns tea.Sequence(Pop(), a.Do())
	// Let's invoke the cmd directly
	msg := cmd()
	// Bubble Tea Sequence executes batch of cmds
	if seq, ok := msg.(tea.BatchMsg); ok {
		for _, c := range seq {
			if c != nil {
				c()
			}
		}
	}

	if !firstExecuted {
		t.Error("expected first action to execute, but it did not")
	}
	if secondExecuted {
		t.Error("second action was executed instead of first action")
	}
}
