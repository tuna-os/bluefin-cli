package app

import (
	"errors"
	"fmt"
	"sync"

	tea "charm.land/bubbletea/v2"
)

// Sentinel errors returned by Registry.Register.
var (
	ErrEmptyActionID   = errors.New("action ID must not be empty")
	ErrNilActionDo     = errors.New("action callback (Do) must not be nil")
	ErrDuplicateAction = errors.New("duplicate action ID")
)

// Action is a single invokable capability. Menus and the ctrl+p command
// palette both render from the same registry, so anything reachable through
// the menu tree is also one fuzzy search away.
type Action struct {
	ID        string
	Icon      string
	Label     string
	Section   string
	Available func() bool // nil means always available
	Do        func() tea.Cmd
}

// Registry manages a validated collection of palette actions.
type Registry struct {
	mu      sync.RWMutex
	actions []Action
	byID    map[string]int
}

// NewRegistry constructs an empty, isolated action registry.
func NewRegistry() *Registry {
	return &Registry{
		byID: make(map[string]int),
	}
}

// Register validates and registers one or more actions. It returns an error if
// any action has an empty ID, a nil Do callback, or an ID that duplicates an
// existing or previously registered action in this call. If an error is returned,
// no actions from this call are registered.
func (r *Registry) Register(actions ...Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.byID == nil {
		r.byID = make(map[string]int)
	}

	// Validate all actions first
	incoming := make(map[string]bool, len(actions))
	for _, a := range actions {
		if a.ID == "" {
			return ErrEmptyActionID
		}
		if a.Do == nil {
			return fmt.Errorf("%w: %q", ErrNilActionDo, a.ID)
		}
		if _, exists := r.byID[a.ID]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateAction, a.ID)
		}
		if incoming[a.ID] {
			return fmt.Errorf("%w: %q", ErrDuplicateAction, a.ID)
		}
		incoming[a.ID] = true
	}

	for _, a := range actions {
		r.byID[a.ID] = len(r.actions)
		r.actions = append(r.actions, a)
	}
	return nil
}

// Actions returns a snapshot copy of the currently available registered actions.
func (r *Registry) Actions() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Action, 0, len(r.actions))
	for _, a := range r.actions {
		if a.Available == nil || a.Available() {
			out = append(out, a)
		}
	}
	return out
}

// All returns a snapshot copy of all registered actions regardless of availability.
func (r *Registry) All() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Action, len(r.actions))
	copy(out, r.actions)
	return out
}

// Count returns the total number of registered actions.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.actions)
}

// Reset clears all registered actions.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions = nil
	r.byID = make(map[string]int)
}

// Snapshot returns a copy of all actions currently registered in the registry.
func (r *Registry) Snapshot() []Action {
	return r.All()
}

// Restore resets the registry to the given slice of actions after validating them.
func (r *Registry) Restore(actions []Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	newByID := make(map[string]int, len(actions))
	for _, a := range actions {
		if a.ID == "" {
			return ErrEmptyActionID
		}
		if a.Do == nil {
			return fmt.Errorf("%w: %q", ErrNilActionDo, a.ID)
		}
		if _, exists := newByID[a.ID]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateAction, a.ID)
		}
		newByID[a.ID] = len(newByID)
	}

	r.actions = make([]Action, len(actions))
	copy(r.actions, actions)
	r.byID = newByID
	return nil
}

// defaultRegistry is the package-level default registry.
var defaultRegistry = NewRegistry()

// DefaultRegistry returns the package-level default action registry.
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// SetDefaultRegistry swaps the package-level default registry, returning the previous one.
func SetDefaultRegistry(r *Registry) *Registry {
	if r == nil {
		r = NewRegistry()
	}
	prev := defaultRegistry
	defaultRegistry = r
	return prev
}

// ResetRegistry clears the default registry. Primarily for tests.
func ResetRegistry() {
	defaultRegistry.Reset()
}

// SnapshotRegistry returns a snapshot of all actions in the default registry.
func SnapshotRegistry() []Action {
	return defaultRegistry.Snapshot()
}

// RestoreRegistry restores the default registry from a previously taken snapshot.
func RestoreRegistry(actions []Action) error {
	return defaultRegistry.Restore(actions)
}

// Register adds actions to the default registry. Called from init-time wiring
// in cmd; the extra build variant registers additional actions.
// If an action is invalid (empty ID, nil Do, or duplicate ID), Register panics
// to surface programming errors immediately during startup/initialization.
func Register(actions ...Action) {
	if err := defaultRegistry.Register(actions...); err != nil {
		panic(fmt.Sprintf("app.Register failed: %v", err))
	}
}

// Actions returns the currently-available registered actions from the default registry.
func Actions() []Action {
	return defaultRegistry.Actions()
}

// newPalette builds the ctrl+p command palette from the given or default registry.
func newPalette() Screen {
	return newPaletteFromRegistry(defaultRegistry)
}

func newPaletteFromRegistry(reg *Registry) Screen {
	byID := map[string]Action{}
	items := func() []MenuItem {
		actions := reg.Actions()
		out := make([]MenuItem, 0, len(actions))
		for _, a := range actions {
			if _, exists := byID[a.ID]; exists {
				continue
			}
			byID[a.ID] = a
			out = append(out, MenuItem{Icon: a.Icon, Label: a.Label, Value: a.ID, Desc: a.Section})
		}
		return out
	}
	menu := NewMenu("Palette", nil, items, func(it MenuItem) tea.Cmd {
		a, ok := byID[it.Value]
		if !ok {
			return nil
		}
		return tea.Sequence(Pop(), a.Do())
	})
	menu.filtering = true
	return menu
}
