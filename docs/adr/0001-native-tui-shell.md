# ADR 0001: One persistent TUI shell, native screens only

**Status**: accepted (2026-07)

## Context
The menus grew one at a time, as separate huh `form.Run()` calls. Each call
cleared the terminal and took control of it. Each flow built its own
navigation state, headers, and themes, and drill-down/back behavior was not
consistent.

## Decision
A single program (`internal/tui/app`, bubbletea v2) stays open for the whole
session. It hosts a stack of `Screen`s (k9s-style push/pop). Every
interactive flow renders inside it through one of four wrappers:

- MenuScreen (lists).
- FormScreen (huh v2 forms, embedded).
- TextScreen (read-only, scrollable).
- RunnerScreen (tasks that print). It captures the output: it swaps
  os.Stdout for a pipe, and the renderer keeps its own fd.

Interactive CLI commands (`bluefin-cli fonts`…) open the same shell at the
correct screen. Thus the UI has only one path through the code. `RunExternal`
(tea.Exec) stays only to start interactive programs that are truly external.

## Consequences
The chrome, themes, and keybindings are the same everywhere. Unit tests can
test flows as models (no pty). Anything that prints or blocks MUST go through
RunnerScreen — synchronous capture deadlocks (glow terminal queries).
