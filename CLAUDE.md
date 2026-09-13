# bluefin-cli — agent guide

Make your terminal experience awesome on any OS (Linux, macOS, Windows, WSL).
Go CLI + bubbletea v2 TUI. Read `CONTEXT.md` for the domain glossary and
`docs/adr/` for why things are the way they are.

## Build & test

```bash
go build ./...                    # vanilla variant
go build -tags extra ./...        # plus variant (wallpapers, fonts, sunset…)
go test -tags extra ./...         # cmd tests REQUIRE -tags extra
go build -tags extra -o bluefin-cli .   # integration tests exec ./bluefin-cli from repo root
scripts/tui-smoke.sh <binary>     # tmux end-to-end suite (15 assertions; runs in CI)
```

- Both build variants must compile before committing; CI gates on tests,
  tui-smoke, gofmt, and golangci-lint.
- Local quirk on this dev box: /tmp is noexec — use `GOTMPDIR=$PWD/tmp/gotmp go test …`
  and build scratch binaries into `tmp/` (gitignored).

## Landmines

- **Braille sprite rows must be 100% braille chars** (blanks are U+2800, not
  spaces): some fonts draw braille double-wide and mixed-width rows shear.
  Same class of bug: never put East-Asian-ambiguous glyphs (✓ · ❯ ↑↓←→)
  inside a bordered box — that's why huh forms use a left-bar style.
- **Anything that prints or blocks must run in `app.RunnerScreen`**, never
  synchronously in a `View`/`Update` (glow queries the terminal and will
  deadlock a synchronous capture).
- **But a runner must never prompt.** A runner captures stdout, so a task can
  print; Bubble Tea still holds the keyboard, so a `huh` prompt opened from
  inside one renders nowhere the user can answer and waits forever (#273).
  Ask on the menu thread before you push the runner, or push a `FormScreen`,
  which owns input. `env.TerminalOwned()` reports the runner case, so shared
  code can take a non-interactive path; note that a test or CI job has no TTY,
  where `huh` errors at once instead of hanging — passing there proves nothing.
- **Keep `go.mod`/`go.sum` tidy in every commit.** `go build`/`go test` ignore
  stale go.sum lines, so drift (typically a renovate bump that leaves the old
  version's hashes behind) passes CI, then the release job's `go mod tidy`
  rewrites the files and goreleaser aborts with "git is in a dirty state" —
  *after* the tag and GitHub release already exist. The `go.mod tidy` CI job
  now gates this; release-side it is `go mod tidy -diff` (non-mutating).
- **goreleaser repository `token`/`private_key` fields render env-only** —
  no template functions (`envOrDefault` crashes there); `skip_upload` gets
  the full template context.
- Config tests must `viper.Reset()` when overriding HOME, or `Save()` writes
  through the leaked config path into the user's real config.
- Publisher `ids` filters in .goreleaser.yaml match **archive** IDs, not
  build IDs.

## Sprites

Pixel art is generated, not hand-typed: `scripts/braillegen.py` (braille
header dino) and the PixelCanvas grids in `internal/tui/app/game.go`
(downsampled from the real Chrome sprite; see tmp workflow in the script).
