package registry

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbletea"
)

func TestRegistryPaletteActionRegistration(t *testing.T) {
	registry := &CommandRegistry{
		screens:   make(map[string]app.Screen),
		actions:   make(map[string]app.PaletteAction),
		menuItems: []app.MenuItem{},
	}

	action := PaletteAction{
		ID:          "test.action",
		Title:       "Test Action",
		Description: "A test action",
		Handler: func() tea.Cmd {
			return nil
		},
		Tags: []string{"test"},
	}

	registry.RegisterPaletteAction(action)

	actions := registry.Actions()
	if len(actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(actions))
	}
	if actions[0].ID != "test.action" {
		t.Fatal("Expected action ID to match")
	}
}

func TestRegistryStatusProviderRegistration(t *testing.T) {
	registry := &CommandRegistry{
		screens:         make(map[string]app.Screen),
		actions:         make(map[string]app.PaletteAction),
		statusProviders: make(map[string]StatusProvider),
	}

	provider := &mockStatusProvider{}
	registry.RegisterStatusProvider("test", provider)

	providers := registry.StatusProviders()
	if len(providers) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(providers))
	}
	if providers["test"] != provider {
		t.Fatal("Expected to retrieve registered provider")
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	registry := &CommandRegistry{
		screens:   make(map[string]app.Screen),
		actions:   make(map[string]app.PaletteAction),
		menuItems: []app.MenuItem{},
	}

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 10; i++ {
			action := PaletteAction{
				ID:      "action-test",
				Title:   "Action",
				Handler: func() tea.Cmd { return nil },
			}
			registry.RegisterPaletteAction(action)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 10; i++ {
			registry.Actions()
		}
		done <- true
	}()

	<-done
	<-done
}

// Mock implementations for testing

type mockStatusProvider struct{}

func (m *mockStatusProvider) ReportStatus(ctx context.Context) StatusSection {
	return StatusSection{
		Title: "Mock",
		Entries: []StatusEntry{
			{Key: "Status", Value: "OK"},
		},
	}
}
