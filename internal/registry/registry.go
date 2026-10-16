package registry

import (
	"sync"

	"github.com/charmbracelet/bubbletea"
)

// PaletteAction describes a command that can be invoked via the palette.
type PaletteAction struct {
	ID          string        // Unique identifier (e.g., "shell.switch")
	Title       string        // Display name for palette
	Description string        // Help text
	Handler     func() tea.Cmd // Command to run
	Tags        []string      // For filtering (e.g., "shell", "config")
}

// MenuItem represents an item in the main menu.
type MenuItem struct {
	Title      string        // Display text
	Selected   func() tea.Cmd // Action when selected
	Visible    func() bool    // Whether this item should be shown
}

// CommandRegistry manages discovery and registration of available commands.
// Subsystems register their palette actions and menu items once at init time.
type CommandRegistry struct {
	mu                sync.RWMutex
	actions           map[string]PaletteAction
	menuItems         []MenuItem
	statusProviders   map[string]StatusProvider
}

// Global registry instance
var defaultRegistry = &CommandRegistry{
	actions:         make(map[string]PaletteAction),
	menuItems:       []MenuItem{},
	statusProviders: make(map[string]StatusProvider),
}

// Default returns the global command registry.
func Default() *CommandRegistry {
	return defaultRegistry
}

// RegisterPaletteAction registers a palette action.
// Called during subsystem init().
func (cr *CommandRegistry) RegisterPaletteAction(action PaletteAction) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.actions[action.ID] = action
}

// RegisterMenuItem registers a menu item.
func (cr *CommandRegistry) RegisterMenuItem(item MenuItem) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.menuItems = append(cr.menuItems, item)
}

// RegisterStatusProvider registers a status provider.
func (cr *CommandRegistry) RegisterStatusProvider(name string, provider StatusProvider) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.statusProviders[name] = provider
}

// Actions returns all registered palette actions.
func (cr *CommandRegistry) Actions() []PaletteAction {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	actions := make([]PaletteAction, 0, len(cr.actions))
	for _, action := range cr.actions {
		actions = append(actions, action)
	}
	return actions
}

// MenuItems returns all registered menu items.
func (cr *CommandRegistry) MenuItems() []MenuItem {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	items := make([]MenuItem, len(cr.menuItems))
	copy(items, cr.menuItems)
	return items
}

// StatusProviders returns all registered status providers.
func (cr *CommandRegistry) StatusProviders() map[string]StatusProvider {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	providers := make(map[string]StatusProvider, len(cr.statusProviders))
	for name, provider := range cr.statusProviders {
		providers[name] = provider
	}
	return providers
}

// Clear clears all registrations. Used for testing.
func (cr *CommandRegistry) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.actions = make(map[string]PaletteAction)
	cr.menuItems = []MenuItem{}
	cr.statusProviders = make(map[string]StatusProvider)
}
