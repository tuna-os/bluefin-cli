package env

import "sync/atomic"

// Whether another part of the process currently owns the terminal's input.
//
// The TUI's RunnerScreen executes a task while the Bubble Tea program is still
// running. The task can *print* -- the runner captures stdout and stderr into
// its log view -- but it cannot *read*: Bubble Tea still holds the keyboard, so
// an interactive prompt opened from inside a runner renders nowhere the user
// can answer it and blocks forever (tuna-os/bluefin-cli#273).
//
// A counter rather than a bool, so that a nested hold does not release the
// terminal early when the inner one returns.
var terminalHolds atomic.Int32

// TerminalOwned reports whether something else holds the terminal's input.
// Code that would otherwise open an interactive prompt must take a
// non-interactive path -- return an actionable error, or use a flag that tells
// the subprocess not to ask -- rather than prompting into a terminal that
// cannot deliver the keystrokes.
func TerminalOwned() bool { return terminalHolds.Load() > 0 }

// HoldTerminal marks the terminal as owned and returns the release function.
// Callers release with defer:
//
//	defer env.HoldTerminal()()
func HoldTerminal() func() {
	terminalHolds.Add(1)
	return func() { terminalHolds.Add(-1) }
}
