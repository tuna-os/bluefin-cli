# Troubleshooting

## Start with `doctor`

```bash
bluefin-cli doctor
```

This runs six checks ([`cmd/doctor.go`](../cmd/doctor.go)) and prints `✓`
(pass), `!` (warning — not fatal), or `✗` (failure) for each:

| Check | What it verifies | Typical fix |
|---|---|---|
| Package manager | `brew` on PATH (or `coldbrew`/`apk` on Alpine/musl systems) | See [Homebrew not found](#homebrew-is-installed-but-not-found) below |
| Shell integration | The shell experience is enabled for your current shell | `bluefin-cli shell <shell> on` |
| Managed tools | `eza`, `fzf`, `starship` are installed | Install via the TUI's Shell Experience menu |
| GitHub reachable | `https://api.github.com` responds to a HEAD request | Check network/proxy/firewall; required for updates and bundle downloads |
| Self-update viable | The install directory is writable (direct installs only; skipped for package-manager installs) | Fix directory permissions, or install via a package manager instead |
| Version | Compares your version against the latest GitHub release | `bluefin-cli update`, or your package manager's upgrade command |

Two useful flags:

```bash
bluefin-cli doctor --fix     # applies safe automatic fixes (installs missing tools)
bluefin-cli doctor --bench   # measures shell startup cost vs. a bare shell (bash/zsh/fish only)
```

`--fix` covers two cases: it sets up `coldbrew` on Alpine/musl systems, and
it installs `eza`, `fzf`, and `starship` when any of them are absent. It does
not fix shell-integration or network issues — those need the manual steps
above.

## Common issues

### Homebrew is installed but not found

`doctor` checks common install locations
(`/opt/homebrew/bin/brew`, `/usr/local/bin/brew`,
`/home/linuxbrew/.linuxbrew/bin/brew`) and tells you to run:

```bash
eval "$(/opt/homebrew/bin/brew shellenv)"   # adjust path to match your system
```

Add that line to your shell's rc file to make it permanent, or re-run
`bluefin-cli shell <shell> on` to let bluefin-cli manage it.

### No package manager on Alpine / postmarketOS

Homebrew's Ruby bootstrap does not run on musl libc, so `bluefin-cli` uses
[`coldbrew`](https://github.com/coldbrewcli/coldbrew) instead on Alpine-based
systems. If `doctor` reports coldbrew is "not set up yet", run:

```bash
bluefin-cli doctor --fix
```

This installs coldbrew rootless and sandboxed. Until it's set up, package
installs fall back to `sudo apk add` one package at a time.

### Shell integration isn't picking up changes

- **Nushell** cannot evaluate a string, so `bluefin-cli shell nu on` writes
  the init script to `~/.config/nushell/bluefin-cli.nu` and `config.nu` reads
  it from there. If you change your tool configuration, re-run
  `bluefin-cli shell nu on` to regenerate that file. Do not edit it
  directly, because a re-run writes over your changes.
- **ash** (busybox, used on Alpine) has no rc-file convention. bluefin-cli
  writes `~/.ashrc` and exports `ENV` from `~/.profile` so ash picks it up.
  atuin, starship, and carapace have no ash target, so bluefin-cli skips
  them there.
- **PowerShell**: by default, PowerShell scopes a function inside another
  function to the parent. The generated script marks each user-facing
  function `global:`. If the shell acts in a strange way after a manual
  edit, check that each function keeps the `global:` prefix.

### Windows: `install.ps1` fails with an execution-policy error

By default, the execution policy of PowerShell blocks a script that you
downloaded. Run
PowerShell as yourself (not elevated) and either bypass for the single
command:

```powershell
powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.ps1 | iex"
```

or set a less restrictive policy for your user account (see
[Microsoft's `Set-ExecutionPolicy` documentation](https://learn.microsoft.com/powershell/module/microsoft.powershell.security/set-executionpolicy)
for the tradeoffs before you change this system-wide).

### Wallpaper or theme install fails on Windows/WSL

Wallpaper and Windows-theme registration use `LOCALAPPDATA` and WSL-specific
detection (`WSL_DISTRO_NAME`, `WSL_INTEROP`, and `/proc` markers — see
[`internal/env/env.go`](../internal/env/env.go)). Detection can be wrong —
for example, in a container whose `/proc` content resembles WSL. In that
case, installs that depend on the Windows filesystem fail or silently no-op.
Run `bluefin-cli doctor` first. A failed package-manager or network check
often explains a downstream wallpaper or theme error. The wallpaper code
itself is usually not at fault.

### Bundle install fails to download

Curated bundles ship embedded in the binary and are used first. If a bundle
isn't embedded (or you're on an older build), bluefin-cli falls back to a
download from `bundles.base_url` (see
[Configuration](configuration.md#config-file)). If that fails:

- Check `bluefin-cli doctor`'s "GitHub reachable" result — the default base
  URL is a `raw.githubusercontent.com` URL and needs the same network access.
- If you're behind a mirror or air-gapped network, point
  `BLUEFIN_BUNDLES_BASE_URL` and `BLUEFIN_BUNDLES_DEFAULT_PATH` at your own
  mirror (see [Configuration](configuration.md#environment-variables)).

### `go test ./...` passes locally but CI's containerized suite fails (or vice versa)

`go test ./...` and `just test` are not the same suite. `just test` builds a
Podman container and runs `go test ./test/...` inside it — a narrower,
integration-focused path that assumes a Linux container environment. If one
passes and the other does not, the cause is usually a host-specific
assumption. A common case: a tool on your PATH that is not in the container,
or vice versa. Run both before you open a PR; see the
[contributor guide](../CONTRIBUTING.md).

## Still stuck?

Search [existing issues](https://github.com/tuna-os/bluefin-cli/issues) —
many platform quirks (especially Windows/WSL edge cases) are already tracked
there with workarounds. If you don't find a match, open a new issue with your
`bluefin-cli doctor` output and the command that failed.
