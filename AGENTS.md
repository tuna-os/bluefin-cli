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
  the functions that read it.
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
- `docs/commands/` is generated command reference. Regenerate it with
  `just gen-docs` after changing commands or flags.

## Development workflow

The module requires Go 1.26.0 or later -- the `go` directive in `go.mod` is the
source of truth for this number. CI currently runs Go 1.27.

```bash
just build                    # build standard and plus binaries
go test ./...                 # run the local test suite
go test -tags extra -race ./... # exercise the CI build tag with race checks
just test                     # run the containerized integration suite
just gen-docs                 # regenerate docs/commands
```

Use `just --list` for the complete recipe list. Podman is required by the
container-based recipes.

## Change guidelines

- Put shared behavior in the relevant `internal/` package and keep Cobra
  command functions small.
- Test both the standard and `extra` build-tag paths when changing conditional
  features. Stubs for the standard build live in `cmd/extra_stubs.go`.
- Add TUI destinations as `app.Screen` implementations or registered
  `app.Action` values. Use the existing runner/external-process bridge for
  commands that must temporarily own the terminal.
- Do not edit generated command pages by hand. Change the Cobra definition and
  run `just gen-docs`. CI fails if regenerating produces a diff.
- Update embedded package data with `just update-resources`; do not add runtime
  downloads for resources that are intended to ship with the binary.
