package status

// StatusProvider is the interface that status reporting subsystems must implement
// to participate in the unified status report. Each subsystem (shell, motd, sunset, install)
// should implement this interface instead of exposing ad-hoc CheckStatus/LoadConfig functions.
type StatusProvider interface {
	// Name returns the human-readable name of this status section (e.g. "Shell Experience", "Sunset Automation").
	Name() string

	// Render returns the formatted status section for the given terminal width.
	// Implementations should handle their own wrapping and formatting using lipgloss styles.
	Render(width int) (string, error)
}

// Registry holds all registered StatusProviders and renders them in order.
type Registry struct {
	providers []StatusProvider
}

// NewRegistry creates an empty status provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: []StatusProvider{},
	}
}

// Register adds a StatusProvider to the registry.
func (r *Registry) Register(provider StatusProvider) {
	r.providers = append(r.providers, provider)
}

// RenderAll renders all registered providers and returns the formatted output.
func (r *Registry) RenderAll(width int) (string, error) {
	var sections []string

	for _, provider := range r.providers {
		section, err := provider.Render(width)
		if err != nil {
			return "", err
		}
		sections = append(sections, section)
	}

	if len(sections) == 0 {
		return "", nil
	}

	// Join all sections with a blank line separator.
	// If width supports multi-column layout (width >= 76), this will be handled per-section.
	result := sections[0]
	for _, section := range sections[1:] {
		result += "\n\n" + section
	}

	return result, nil
}
