package env

import "testing"

func TestTerminalOwned_DefaultsToFree(t *testing.T) {
	if TerminalOwned() {
		t.Fatal("the terminal should start unowned")
	}
}

func TestHoldTerminal_ReleasesOnlyAtTheOutermostHold(t *testing.T) {
	releaseOuter := HoldTerminal()
	if !TerminalOwned() {
		t.Fatal("HoldTerminal did not take the terminal")
	}

	releaseInner := HoldTerminal()
	releaseInner()
	if !TerminalOwned() {
		t.Error("a nested release freed the terminal while the outer hold was open")
	}

	releaseOuter()
	if TerminalOwned() {
		t.Error("the outermost release did not free the terminal")
	}
}
