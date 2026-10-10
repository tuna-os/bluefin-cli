// Package registry manages command palette actions and screens.
package registry

import (
	"sync"

	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// ActionRegistration holds a palette action and its associated metadata.
type ActionRegistration struct {
	Action app.Action
}

// Registry manages palette action registration with lazy initialization.
type Registry struct {
	mu      sync.Mutex
	actions []ActionRegistration
	once    sync.Once
}

// New creates a new Registry.
func New() *Registry {
	return &Registry{}
}

// Register adds an action to the registry.
func (r *Registry) Register(action app.Action) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions = append(r.actions, ActionRegistration{Action: action})
}

// RegisterBatch adds multiple actions at once.
func (r *Registry) RegisterBatch(actions ...app.Action) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, action := range actions {
		r.actions = append(r.actions, ActionRegistration{Action: action})
	}
}

// Finalize registers all queued actions to the app.
// It ensures registration happens exactly once via sync.Once.
func (r *Registry) Finalize() {
	r.once.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, reg := range r.actions {
			app.Register(reg.Action)
		}
	})
}

// Global registry instance used by commands.
var global = New()

// Register adds an action to the global registry.
func Register(action app.Action) {
	global.Register(action)
}

// RegisterBatch adds multiple actions to the global registry.
func RegisterBatch(actions ...app.Action) {
	global.RegisterBatch(actions...)
}

// Finalize registers all queued actions to the app.
func Finalize() {
	global.Finalize()
}

// Clear resets the global registry (for testing).
func Clear() {
	global = New()
}
