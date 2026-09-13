package shell

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

// hideHomebrew makes findBrew report a machine with no brew: nothing on PATH,
// and none of the install prefixes present.
func hideHomebrew(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	missing := t.TempDir() + "/nowhere/bin/brew"
	original := brewFallbackPaths
	brewFallbackPaths = []string{missing}
	t.Cleanup(func() { brewFallbackPaths = original })
}

func TestHomebrewAvailable_ReportsAMachineWithoutIt(t *testing.T) {
	hideHomebrew(t)
	if HomebrewAvailable() {
		t.Error("HomebrewAvailable() is true with no brew on PATH and no install prefix")
	}
}

// TestEnsureHomebrew_DoesNotPromptWhenTerminalOwned is the regression test for
// tuna-os/bluefin-cli#273: the confirm this used to open from inside a TUI
// runner rendered nowhere the user could answer it, and the runner span
// forever. With the terminal held, ensureHomebrew must return an actionable
// error instead. The call runs in a goroutine so that a reintroduced prompt
// fails the test rather than hanging the suite.
func TestEnsureHomebrew_DoesNotPromptWhenTerminalOwned(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows installs through winget and never reaches the prompt")
	}
	hideHomebrew(t)
	defer env.HoldTerminal()()

	errs := make(chan error, 1)
	go func() { errs <- ensureHomebrew() }()

	select {
	case err := <-errs:
		if err == nil {
			t.Fatal("ensureHomebrew() succeeded with no brew installed")
		}
		if !strings.Contains(err.Error(), "homebrew is not installed") {
			t.Errorf("error does not say what is wrong: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("ensureHomebrew() blocked while the TUI held the terminal — it is prompting again")
	}
}
