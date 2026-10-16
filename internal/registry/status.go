package registry

import (
	"context"
)

// StatusEntry represents one line of status output.
type StatusEntry struct {
	Key   string // Label (e.g., "Shell", "Homebrew")
	Value string // Value (e.g., "zsh", "installed")
	Error bool   // True if this represents an error condition
}

// StatusSection groups related status entries.
type StatusSection struct {
	Title   string        // Section header (e.g., "System", "Configuration")
	Entries []StatusEntry // Status lines in this section
}

// StatusProvider is implemented by subsystems that report status.
// Each subsystem calls registry.RegisterStatusProvider during init().
type StatusProvider interface {
	// ReportStatus returns status information for this subsystem.
	// Called by the status aggregator.
	ReportStatus(ctx context.Context) StatusSection
}

// RegisterStatusProvider registers a status provider.
func (cr *CommandRegistry) RegisterStatusProvider(name string, provider StatusProvider) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.statusProviders == nil {
		cr.statusProviders = make(map[string]StatusProvider)
	}
	cr.statusProviders[name] = provider
}

// StatusProviders returns all registered status providers.
func (cr *CommandRegistry) StatusProviders() map[string]StatusProvider {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	if cr.statusProviders == nil {
		return make(map[string]StatusProvider)
	}

	providers := make(map[string]StatusProvider, len(cr.statusProviders))
	for name, provider := range cr.statusProviders {
		providers[name] = provider
	}
	return providers
}
