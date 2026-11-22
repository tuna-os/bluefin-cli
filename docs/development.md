# Developer guide: code organization

This guide maps the repository so you can find where a change belongs. For
build, test, and lint commands, see [`CONTRIBUTING.md`](../CONTRIBUTING.md). For
the test layers, see [`testing.md`](testing.md). For architecture rules that
reviewers enforce, see [AGENTS.md](../AGENTS.md).

## Top-level layout

| Path | What it holds |
|---|---|
| `main.go` | Entry point. Hands control to the `cmd` package. |
| `cmd/` | Cobra commands and the TUI destinations they open. |
| `internal/` | Reusable behavior, grouped by package. |
| `docs/` | User and developer docs. `docs/commands/` is generated. |
| `docs/adr/` | Architecture decision records. |
| `scripts/` | Shell suites for TUI smoke tests, state tests, and shell experience. |
| `test/` | Go integration tests. |
| `src/bluefin-cli/` | Dev container feature files. |
| `tools/release/` | Release tooling. |
| `justfile` | Task recipes. Run `just --list`. |

The repository builds two binaries from one source tree: `bluefin-cli`
(standard) and `bluefin-cli-plus` (`-tags extra`). See
[vanilla-vs-extra.md](vanilla-vs-extra.md).

## `cmd/`: command wiring

Each file registers a Cobra command on the root command in `cmd/root.go`
(for example `doctor.go`, `install.go`, `shell.go`, `profile.go`, `update.go`).
Keep argument parse and wiring code here. Put reusable behavior in `internal/`.

The `menu*.go`, `shell_menu.go`, and `install_menu.go` files assemble the
interactive menu:

- `menu.go` defines the `menu` command, starts the TUI with `app.Run`, and
  runs the background update check.
- `menu_model.go` describes the main menu as data (items, icons, descriptions).
- `menu_router.go` maps a selected menu value to a handler.
- `menu_native.go` holds in-shell flows: forms, text views, and runner screens
  that capture command output.
- `menu_vanilla.go`, `menu_extra.go`, and `menu_extra_actions.go` hold the
  parts that differ between builds.

### Build-tag split

Features that only ship in the plus build (wallpapers, fonts, sunset) sit
behind the `extra` build tag. Files marked `//go:build extra` hold the real
code. `cmd/extra_stubs.go` (`//go:build !extra`) supplies no-op stand-ins so
the standard build compiles. After you change a conditional feature, check both
builds. `cmd/variant_contract_test.go` and the `*_extra_test.go` files cover
this split.

## `internal/`: packages

| Package | Owns |
|---|---|
| `tui/app` | The persistent Bubble Tea program: screen stack, header and footer, command palette, and runners. See below. |
| `tui` and `tui/theme` | Shared styles and theme definitions for the TUI. |
| `shell` | Shell experience for Bash, Zsh, Fish, ash, Nushell, and PowerShell: enabling and disabling, rendering the init script, status, and uninstall. |
| `install` | Packages, bundles, Brewfiles, and wallpaper collections. Embeds Brewfiles and wallpaper metadata from `install/resources/`. |
| `pkgmanager` | Package manager selection on Alpine-family hosts (coldbrew, then apk). |
| `config` | Persisted configuration file. |
| `env` | Config directory lookup and terminal environment detection. |
| `profile` | Portable profiles: export, import, diff, and gist sync. |
| `update` | Checksum-verified self-update. |
| `doctor` | Diagnostic probes, `--fix` remediation, shell startup benchmark, and report rendering (`doctor/render`). |
| `status` | The status report. |
| `motd` | Message-of-the-day handling. |
| `starship` | Starship install and preset themes. |
| `terminal` | Terminal emulator and font setup. |
| `wallpaper` | Native per-platform wallpaper setting. |
| `sunset` | Sunset and sunrise theme and wallpaper switching. |
| `countme` | The Fedora countme protocol implementation. |

### The TUI (`internal/tui/app`)

- A `Screen` (`app.go`) is one level of the navigation stack: `Title`, `Init`,
  `Update`, `View`, and `KeyHints`.
- Screen types: `MenuScreen` (`menu.go`), `FormScreen` (`form.go`),
  `TextScreen` (`text.go`), and `RunnerScreen` (`runner.go`) for captured
  task output. `wallpaper_picker.go`, `game.go`, and `dino.go` hold special
  screens and art.
- An `Action` (`action.go`) is a registered capability. Menus and the
  `ctrl+p` command palette read the same registry, so each action is
  reachable from both. Modules call `app.Register` for their own actions.
- `exec.go` releases the terminal for an external interactive program. Use it
  only when a command must own the terminal.

### Shell integration (`internal/shell`)

- `shells.go` is the registry of managed shells: rc path, init line, and
  syntax flavor. Add a shell there.
- `resources/` holds the init scripts (`shell.sh`, `shell.fish`, `shell.nu`,
  `shell.ps1`). `nu_aliases.go` generates the Nushell aliases.
- Installation backends are separate from the shell init code: `installers.go`
  (shared entry points and Homebrew), `windows_tools.go` (winget),
  `install_alpine.go` (coldbrew/apk). `shell.go` keeps enablement, the init
  output, and status.

### Embedded resources (`internal/install/resources`)

Brewfiles and wallpaper metadata ship inside the binary. Update them with
`just update-resources`, which also rewrites `PROVENANCE.json`. Do not edit
them by hand.

## How a command reaches `internal/`

1. `main.go` calls into `cmd`, which runs the root Cobra command.
2. The subcommand's `RunE` in `cmd/<name>.go` parses flags and calls a
   function in the matching `internal/` package.
3. For the TUI, `bluefin-cli menu` calls `app.Run` with the main menu screen.
   Menu selections go through the router to screens or actions, which call
   the same `internal/` functions the CLI commands use.

## Where to start

- **Add a command.** Create `cmd/<name>.go`, register it with
  `rootCmd.AddCommand` in `init`, and put the logic in `internal/`. Then
  regenerate `docs/commands/` from the extra variant, as the
  [contributor guide](../CONTRIBUTING.md) shows. CI fails on a stale diff.
- **Add a TUI destination.** Make a type that satisfies `app.Screen`, or
  register an `app.Action`. Reuse `RunnerScreen` for tasks that show command
  output.
- **Fix or extend shell integration.** Start in `internal/shell`
  (`shells.go`, `resources/`), and see the scope rules in
  [AGENTS.md](../AGENTS.md).
- **Add a package or bundle.** Update the embedded resources with
  `just update-resources`. See [brewfile-guide.md](brewfile-guide.md).
- **Change config or profile behavior.** Use `internal/config` and
  `internal/profile`. See [configuration.md](configuration.md) and
  [profiles.md](profiles.md).

## Test layout

- Unit tests sit beside the code as `*_test.go` in each package.
- `cmd/*_test.go` covers command and menu wiring, including build-variant
  contracts.
- `test/` holds Go integration tests that the containerized suite
  (`just test`) runs.
- `scripts/` holds the shell suites: `tui-smoke.sh`, `tui-state.sh`, and
  `shell-experience.sh`.

See [`testing.md`](testing.md) for how to run each layer and add tests.
