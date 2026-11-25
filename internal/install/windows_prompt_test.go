package install

import (
	"strings"
	"sync"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

// With the terminal owned by a TUI runner, a missing winget must produce an
// actionable error and never open a prompt (tuna-os/bluefin-cli#532).
func TestEnsureWindowsManagersBootstrap_DoesNotPromptWhenTerminalOwned(t *testing.T) {
	origConfirm := windowsPromptConfirm
	windowsPromptConfirm = func(title, description string) (bool, error) {
		t.Error("windowsPromptConfirm called while the terminal is owned")
		return false, nil
	}
	origManagers, origErr := windowsBootstrapManagers, windowsBootstrapErr
	windowsBootstrapManagersOnce = sync.Once{}
	windowsBootstrapManagers, windowsBootstrapErr = nil, nil
	t.Cleanup(func() {
		windowsPromptConfirm = origConfirm
		windowsBootstrapManagersOnce = sync.Once{}
		windowsBootstrapManagers, windowsBootstrapErr = origManagers, origErr
	})
	defer env.HoldTerminal()()

	_, err := ensureWindowsManagersBootstrap(func() []string { return nil })
	if err == nil {
		t.Fatal("expected an error when winget is missing and the terminal is owned")
	}
	if !strings.Contains(err.Error(), "winget is not installed") || !strings.Contains(err.Error(), "apps.microsoft.com/detail/9NBLGGH4NNS1") {
		t.Errorf("error does not say how to install winget: %v", err)
	}
}
