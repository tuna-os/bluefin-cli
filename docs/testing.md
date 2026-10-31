# TUI tests and test layout

This guide tells you how to test the TUI on your computer. It also shows
the test layers and how to add a test.

## Test layers

The TUI uses [Bubble Tea v2](https://github.com/charmbracelet/bubbletea)
(`charm.land/bubbletea/v2`). The code is in `internal/tui/app`. The tests
for the TUI have four layers:

| Layer | Location | What it checks |
|---|---|---|
| Model tests | `internal/tui/app/*_test.go` | Navigation state, the screen stack, the cursor, the filter, and the frame |
| Menu wire tests | `cmd/menu_test.go` | Each menu item and bundle category goes to an action |
| Smoke test | `scripts/tui-smoke.sh` | The real binary in a `tmux` pane: screens, filter, palette, help, and quit |
| State test | `scripts/tui-state.sh` | Menu actions write the correct changes to files in a temporary `$HOME` |

## Run the tests

### Model and menu tests

These tests do not open a terminal:

```bash
go test ./internal/tui/...
go test ./cmd -run 'Test(MainMenu|ExtraMenu|BundleCategories)'
```

### Smoke test

You must install `tmux`. The `just` recipe builds the binary with
`-tags extra`, runs the script, and then removes the binary:

```bash
just tui-smoke
```

To run the same steps yourself:

```bash
go build -tags extra -o tmp/bfc-smoke .
./scripts/tui-smoke.sh tmp/bfc-smoke
rm -f tmp/bfc-smoke
```

If `tmux` is not available, the script shows `SKIP: tmux not available`
and stops with exit code 0.

### State test

```bash
go build -tags extra -o tmp/bfc-state .
./scripts/tui-state.sh tmp/bfc-state
rm -f tmp/bfc-state
```

### All checks

To do the CI checks on your computer, use this recipe:

```bash
just verify
```

## How the tests work

### Model tests

The model tests use the `drive` helper in `internal/tui/app/app_test.go`:

```go
func drive(root Screen, msgs ...tea.Msg) Model
```

The helper makes a new model with `root` as the first screen. It then
sends each message to `Update`. It runs each command that `Update` returns
at once, and sends the result back to `Update`. The real program runs
commands in goroutines, so the order of messages can change. The helper
keeps the order the same for each test run.

### Smoke test

The script starts a `tmux` session with a size of 100 x 28. It sends keys
with `tmux send-keys` and reads the screen with `tmux capture-pane -p`.
It then uses `grep -q` to find each expected pattern. If a check fails, the
script shows the first 20 lines of the screen.

The `SLEEP` variable sets the time in seconds after each key. The default
is 0.5. On a slow computer, set a larger value:

```bash
SLEEP=1.0 ./scripts/tui-smoke.sh tmp/bfc-smoke
```

### State test

The script makes a temporary directory and sets `HOME` to it. It moves
through the menus with the `/` filter, because the cursor keys can go to
the wrong row on a slow computer. For actions that run in a runner screen,
it uses `wait_runner` to wait until the task stops. Then it examines the
files in the temporary home.

## Add a test

### New screen

1. Add a model test in `internal/tui/app/`. Use `drive` to send keys.
2. Use `stripAnsi` to remove the ANSI codes from the frame before you
   compare text.
3. Make sure that `Escape` or `PopMsg` goes back to the previous screen.

### New menu item

1. If you add an item to `mainMenuItems()`, make sure that
   `mainMenuSelect()` gives an action for its value.
2. Run `go test ./cmd`.

### New main flow

Add keys and checks to `scripts/tui-smoke.sh`. For example:

```bash
keys /; keys m y - f e a t u r e; keys Enter; cap
assert "my feature renders" "› My Feature"
keys Escape; cap
assert "returns to parent" "› Home"
```

Then run `just tui-smoke`.

### New change to files

Add steps to `scripts/tui-state.sh`. Use the `/` filter to go to the menu
item. Then examine the files in `$H`, the temporary home.

## Tests in CI

The CI workflow is `.github/workflows/ci.yml`:

- **Run Tests (standard)** and **Run Tests (extra)** run
  `go test -race ./...` with each build variant.
- **TUI Smoke (tmux)** installs `tmux`, builds the binary with
  `-tags extra`, and runs `scripts/tui-smoke.sh` and
  `scripts/tui-state.sh`.
