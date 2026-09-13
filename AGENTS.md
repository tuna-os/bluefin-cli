# AGENTS.md

## Project overview

Bluefin CLI is a Go command-line application for shell configuration, package
installation, and desktop customization. Cobra provides the command tree. The
interactive interface is a persistent Bubble Tea v2 application.

The repository produces two binaries from the same source:

- `bluefin-cli`, the standard build;
- `bluefin-cli-plus`, built with `-tags extra` to include wallpapers, fonts,
  sunset automation, and the complete interactive menu.

## Architecture

- `cmd/` defines Cobra commands and assembles TUI destinations. Keep argument
  parsing and command wiring here; place reusable behavior under `internal/`.
- `internal/shell/` manages the shell experience for Bash, Zsh, Fish, ash,
  Nushell, and PowerShell. The user-facing command is wired in `cmd/shell.go`.
  `shells.go` holds the registry of managed shells: one entry per shell with
  its rc path, init line and syntax flavor. Add a shell there rather than in
  the functions that read it. Installation backends are separate from
  rendering: `installers.go` holds the shared entry points and the Homebrew
  path, `windows_tools.go` the winget/PowerShell path, `install_alpine.go` the
  coldbrew/apk path, and `shell.go` keeps only enablement, init rendering and
  status. Do not name a new file with a `_windows`/`_linux`/`_darwin` suffix
  unless you mean the GOOS build constraint that comes with it.
- `internal/install/` handles packages, bundles, and wallpaper collections.
  Brewfiles and wallpaper metadata are embedded from
  `internal/install/resources/`; update them with `just update-resources`, which
  also rewrites `resources/PROVENANCE.json`. That manifest pins each embedded
  file to an upstream commit and a digest, and `provenance_test.go` fails if the
  tree and the manifest disagree -- so do not hand-edit an embedded resource.
- `internal/tui/app/` implements the persistent screen stack, shared header and
  footer, command palette, and runners for external or streaming operations.
  Menus and actions are registered from `cmd/menu.go` and related `cmd/menu_*`
  files.
- `internal/config/`, `internal/profile/`, and `internal/update/` own persisted
  configuration, portable profiles, and checksum-verified self-update.
- `docs/commands/` holds the generated command reference. Regenerate it with
  `just gen-docs` after changing commands or flags.

## Development workflow

The module needs Go 1.26.0 or later -- the `go` directive in `go.mod` is the
source of truth for this number. CI now runs Go 1.27.

```bash
just build                    # build standard and plus binaries
go test ./...                 # run the local test suite
go test -tags extra -race ./... # exercise the CI build tag with race checks
just test                     # run the containerized integration suite
just gen-docs                 # regenerate docs/commands
```

Use `just --list` for the complete recipe list. Podman is required by the
container-based recipes.

Three shell suites sit under `scripts/`, and they check different things:
`tui-smoke.sh` asserts what appears on screen, `tui-state.sh` asserts what
lands in config files, and `shell-experience.sh` starts each supported shell
and asserts the experience the init script produces -- the aliases, the
prompt, PATH, and that startup stays silent. Run the last one with
`scripts/shell-experience.sh <binary> [shell ...]`; it skips any shell that is
not installed, so it is useful locally with only bash.

## Change guidelines

- Put shared behavior in the relevant `internal/` package, and keep the
  functions for each Cobra command small.
- Test both the standard and `extra` build-tag paths after a change to a
  conditional feature. Stubs for the standard build live in `cmd/extra_stubs.go`.
- Add TUI destinations as `app.Screen` implementations or registered
  `app.Action` values. Use the existing runner/external-process bridge for
  commands that must temporarily own the terminal.
- Do not edit the generated command pages by hand. Change the Cobra definition and
  run `just gen-docs`. CI fails if a regeneration makes a diff.
- Update the embedded package data with `just update-resources`. Do not add a
  runtime download for a resource that must ship in the binary.
