# TUI Testing and Test Architecture

This guide explains how to test the Bluefin CLI Terminal User Interface (TUI) locally, outlines the test architecture across suites, and shows how to add new TUI tests.

## Architecture Overview

Bluefin CLI uses [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) (`charm.land/bubbletea/v2`) for its interactive interface (`internal/tui/app`). The application maintains a persistent navigation stack of screens, contextual footers, a breadcrumb header, and a fuzzy-search command palette.

Testing the TUI is organized into four complementary layers:

| Layer | Location | What It Validates | Mechanism |
|---|---|---|---|
| **Unit & Model Tests** | `internal/tui/app/*_test.go` | Navigation state, screen stack, cursor movement, filter logic, keymaps, component rendering | Synchronous message dispatch (`drive()` / `teatest`) |
| **Menu Wiring Tests** | `cmd/menu_test.go` | Completeness of menu trees and bundle categories; ensures all options resolve to commands | Static registry assertions |
| **End-to-End Smoke Tests** | `scripts/tui-smoke.sh` | Visual rendering, screen transitions, palette, help overlay, interactive exit | Runs real binary inside a headless `tmux` pane; captures terminal output |
| **End-to-End State Tests** | `scripts/tui-state.sh` | Desired-state persistence; verifies actions in the UI write expected filesystem/config changes | Drives binary via `tmux` in an isolated `$HOME` sandbox; asserts config files |

---

## Quick Start: Running Local Tests

You can run each test layer individually or all together through `just` recipes.

### 1. Model & Unit Tests

Run the Bubble Tea model tests without launching a terminal:

```bash
# Test internal/tui packages
go test -v ./internal/tui/...

# Test menu wiring
go test -v ./cmd -run 'Test(MainMenu|ExtraMenu|BundleCategories)'
```

### 2. End-to-End TUI Smoke Test (`tui-smoke.sh`)

Asserts visual rendering and navigation in `tmux`:

```bash
# Via just (builds with -tags extra to tmp/bfc-smoke, runs suite, cleans up)
just tui-smoke

# Or manually:
go build -tags extra -o tmp/bfc-smoke .
./scripts/tui-smoke.sh tmp/bfc-smoke
rm -f tmp/bfc-smoke
```

**Expected output:**
```
ok   main menu renders
ok   menu items present
ok   footer hints present
ok   dino is on screen
ok   drill-down breadcrumb
ok   categories listed
ok   back hint appears
ok   escape pops to home
ok   filter echoes query
ok   filter matches fuzzily
ok   palette opens
ok   dino game opens
ok   game exits to home
ok   help overlay opens
ok   q quits
TUI smoke: all checks passed
```

### 3. End-to-End Desired-State Test (`tui-state.sh`)

Verifies that selecting options in the TUI modifies configuration files as expected:

```bash
go build -tags extra -o tmp/bfc-state .
./scripts/tui-state.sh tmp/bfc-state
rm -f tmp/bfc-state
```

**Expected output:**
```
ok   missing-Homebrew confirm opened before the runner, not inside it
ok   TUI toggle wrote the init line to .bashrc
ok   menu label reflects the new state
ok   TUI disable removed the init line
ok   pre-existing rc content survived the round trip
ok   TUI MOTD toggle persisted to shell config
TUI state: all checks passed
```

### 4. Shell Experience Test (`shell-experience.sh`)

Verifies the shell init scripts produced by the CLI across installed shells (`bash`, `zsh`, `fish`, etc.):

```bash
just shell-experience
# Or specify binary and target shells:
./scripts/shell-experience.sh ./bluefin-cli bash
```

### 5. Full Verification Gauntlet

To run all checks matching CI:

```bash
just verify
```

---

## Test Suites in Detail

### Unit & Model Tests (`internal/tui/app/app_test.go`)

Model tests do not start an interactive terminal or execute external processes. Instead, they use a synchronous driver pattern:

```go
func drive(m tea.Model, msgs ...tea.Msg) tea.Model
```

- **Why:** Real Bubble Tea programs run commands in goroutines, introducing race conditions in deterministic tests. The synchronous helper evaluates returned `tea.Cmd` values immediately on the same thread.
- **Coverage:** Breadcrumb navigation, stack push/pop, cursor motion (`j`/`k`, arrow keys), fuzzy filtering (`/`), help toggle (`?`), palette display (`ctrl+p`), text scrolling, and canvas rendering.

### Menu Wiring Tests (`cmd/menu_test.go`)

Guards against "dead-end" items:
- `TestMainMenuItemsResolve`: Ensures every entry returned by `mainMenuItems()` produces a non-nil `tea.Cmd` when selected.
- `TestExtraMenuItemsResolve`: Ensures extra features in `-tags extra` resolve to valid actions.
- `TestBundleCategoriesAvailable`: Ensures package installation categories are defined with valid IDs and labels.

### Visual Smoke Testing (`scripts/tui-smoke.sh`)

Catches integration breakage that unit tests cannot detect (terminal initialization, escape sequence formatting, terminal sizing, and real key handling).

- **How it works:**
  1. Starts a headless `tmux` session at a fixed resolution (`100x28`).
  2. Waits for the application to paint its initial view (`Bluefin CLI › Home`).
  3. Sends keystrokes via `tmux send-keys`.
  4. Captures the terminal pane using `tmux capture-pane -p`.
  5. Evaluates regular expression patterns against captured output with `grep -q`.
  6. Automatically cleans up the tmux session via an `EXIT` trap.

### Desired-State Testing (`scripts/tui-state.sh`)

Complements smoke tests by verifying filesystem effects:
- Creates an isolated temporary directory set to `$HOME`.
- Navigates through menus using fuzzy filtering (`/`) to prevent cursor desynchronization under high CPU load.
- Triggers actions (such as enabling shell integrations or toggling MOTD).
- Waits for asynchronous runner screens to finish (`wait_runner`).
- Inspects target files (such as `~/.bashrc` and `~/.config/bluefin-cli/*.json`) to confirm changes persist and rollback cleanly.

---

## Adding New TUI Test Cases

When introducing a new screen, menu item, or interactive component:

### 1. Add Model Tests (`internal/tui/app/`)

- For new screen types (implementing `app.Screen`), add unit tests in `internal/tui/app/<feature>_test.go` or `screens_test.go`.
- Test `Init()`, `Update()`, and `View()` methods.
- Assert expected text appears in the output using `stripAnsi(view)` to strip ANSI escape sequences.
- Verify navigation contracts: `Escape` or `BackMsg` returns to the previous screen.

### 2. Verify Menu Wiring (`cmd/menu_test.go`)

- If adding an item to `mainMenuItems()` or `extraMenuItems()`, ensure `mainMenuSelect()` handles the item value.
- Run `go test ./cmd` to verify `TestMainMenuItemsResolve` passes.

### 3. Add End-to-End Assertions (`scripts/tui-smoke.sh`)

If the change affects core navigation flows or adds a primary submenu:
1. Open `scripts/tui-smoke.sh`.
2. Add keystrokes to reach your screen, for example:
   ```bash
   keys /; keys m y - f e a t u r e; keys Enter; sleep 0.5; cap
   assert "my feature renders" "› My Feature"
   keys Escape; cap
   assert "returns to parent" "› Home"
   ```
3. Test locally with `just tui-smoke`.

### 4. Add State Assertions (`scripts/tui-state.sh`)

If your UI modifies configuration or system state:
1. Open `scripts/tui-state.sh`.
2. Navigate to the option using fuzzy filter (`keys /; keys ...`).
3. If the action executes an asynchronous task, use `wait_runner <timeout-seconds>`.
4. Assert expected filesystem changes in `$H` (the sandbox home).
5. Trigger undo/disable and assert cleanup.

---

## CI Test Flow

In GitHub Actions (`.github/workflows/ci.yml`), TUI tests run as part of the automated pull request checks:

1. **`test` job matrix:**
   - Runs `go test ./...` across `standard` and `extra` build tags with race detection (`-race`).
   - Executes model and menu wiring tests.
2. **`tui-smoke` job:**
   - Installs `tmux`.
   - Builds the full binary with `-tags extra`.
   - Runs `./scripts/tui-smoke.sh`.
   - Runs `./scripts/tui-state.sh`.
3. **`container-test` / `just test`:**
   - Containerized integration test suite run in Podman/Docker, testing the compiled binaries in an isolated Linux environment.

---

## Debugging and Troubleshooting

### Inspecting Captured Terminal Output
If `tui-smoke.sh` fails an assertion, it prints the first 20 lines of the captured terminal buffer:
```
FAIL drill-down breadcrumb: expected /› Install Apps/ in capture:
  | Bluefin CLI › Home
  | ...
```
To view the full terminal capture during debugging, modify `scripts/tui-smoke.sh` temporarily to print `$DIR/cap.txt` or copy it out.

### Timing and Keystroke Desynchronization
- Under heavy load, screen updates may take longer than the default sleep intervals.
- `tui-smoke.sh` supports tuning keystroke delays via the `SLEEP` environment variable:
  ```bash
  SLEEP=1.0 ./scripts/tui-smoke.sh tmp/bfc-smoke
  ```
- In `tui-state.sh`, always navigate with the `/` filter instead of fixed `j`/`k` cursor movements to prevent landing on wrong rows if rendering lags.

### Missing `tmux`
If `tmux` is not installed, `tui-smoke.sh` exits cleanly with `SKIP: tmux not available`. For local testing on Linux or macOS, install tmux:
```bash
# Ubuntu/Debian
sudo apt-get install tmux

# macOS
brew install tmux
```
