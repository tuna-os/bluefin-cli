# ADR 0001: One persistent TUI shell, native screens only

**Status**: accepted (2026-07)

## Context
The menu used separate huh `form.Run()` calls. Each call cleared and took over
the terminal. Each flow rebuilt the navigation state, headers, and themes.
Drill-down and back behavior differed between flows.

## Decision
A long-lived program in bubbletea v2 (`internal/tui/app`) hosts a stack of
`Screen`s (k9s-style push/pop). Every interactive flow renders inside it through
four wrappers: MenuScreen (lists), FormScreen (huh v2 forms), TextScreen
(read-only, scrollable), and RunnerScreen (tasks that print). The runner captures
output through an os.Stdout pipe; the renderer keeps its own fd.
Interactive CLI commands (`bluefin-cli fonts`…) start the same shell at the
right screen. This gives the project a single UI codepath. `RunExternal`
(`tea.Exec`) remains only for external programs.

## Consequences
All screens use the same chrome, theme, and keybindings. Models make flows
unit-testable without a pty. Anything that prints or blocks MUST go through
RunnerScreen; synchronous capture deadlocks on glow terminal queries.
