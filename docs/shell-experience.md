# Shell experience guide

The **shell experience** is the set of aliases, tool initializers, and prompt
settings that Bluefin CLI adds to your shell. Turning it on does three things:

1. Appends one managed line to your shell's startup file. That line runs
   `bluefin-cli init <shell>`, which prints the init script for that shell.
2. Tries to install the enabled tools (Homebrew on Linux and macOS, winget on
   Windows, coldbrew or `apk` on Alpine).
3. Records which tools are enabled in `shell.json` in the
   [config directory](configuration.md).

Open a new terminal after enabling it. Run `bluefin-cli status` to see the
result.

## Supported shells

| Shell | Command | Startup file written |
|---|---|---|
| Bash | `bluefin-cli shell bash on` | `~/.bashrc` |
| Zsh | `bluefin-cli shell zsh on` | `~/.zshrc` |
| Fish | `bluefin-cli shell fish on` | `~/.config/fish/config.fish` |
| ash (busybox; also `dash`, `sh`) | `bluefin-cli shell ash on` | `~/.ashrc`, plus `export ENV="$HOME/.ashrc"` in `~/.profile` |
| Nushell | `bluefin-cli shell nu on` (or `nushell`) | `~/.config/nushell/config.nu` sources `~/.config/nushell/bluefin-cli.nu` |
| PowerShell | `bluefin-cli shell powershell on` (or `pwsh`) | `Documents\PowerShell\Microsoft.PowerShell_profile.ps1` and `Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1` under your home directory |

Use `off` instead of `on` to remove the managed line. You can enable several
shells on one machine; run the command once per shell. With no arguments,
`bluefin-cli shell` and the **Bluefin Shell** menu act on your current shell.

## Setup by shell

### Bash and Zsh

```sh
bluefin-cli shell bash on   # or: bluefin-cli shell zsh on
```

The startup file gets `eval "$(bluefin-cli init bash)"` (or `zsh`), marked so
Bluefin CLI can find and remove it later.

### Fish

```sh
bluefin-cli shell fish on
```

The startup file gets `bluefin-cli init fish | source`. Bluefin CLI creates
`~/.config/fish` if it does not exist.

### ash (Alpine, postmarketOS)

```sh
bluefin-cli shell ash on
```

ash has no rc-file convention, so Bluefin CLI writes `~/.ashrc` and sets `ENV`
from `~/.profile`. atuin, starship, and carapace have no ash target and are
skipped there. See [Alpine and musl](#alpine-and-musl).

### Nushell

```sh
bluefin-cli shell nu on
```

Nushell cannot evaluate a string at startup, so Bluefin CLI renders the init
script to `~/.config/nushell/bluefin-cli.nu` and `config.nu` sources it. Re-run
the command to regenerate that file after you change tool settings. Do not
edit it by hand.

### PowerShell

```powershell
bluefin-cli shell powershell on
```

The profile line caches the output of `bluefin-cli init powershell` under
`%LOCALAPPDATA%\bluefin-cli\shell-cache\init.ps1` and rebuilds the cache when
the binary or `shell.json` changes. If PowerShell blocks scripts, see the
execution policy notes in [troubleshooting](troubleshooting.md). Tools install
with winget; `gsudo` is offered on PowerShell only, and ugrep and the uutils
packages are not offered there.

## Managed tools

`bluefin-cli shell --help` lists every tool. The init script enables a tool
only when its binary is on `PATH`.

| Tool | What it does | On by default |
|---|---|---|
| Eza | Modern replacement for `ls` | yes |
| Gsudo | `sudo` for Windows (PowerShell only) | yes |
| Fzf | Fuzzy finder | yes |
| Ugrep | Fast grep with interactive mode | yes |
| Bat | `cat` with syntax highlighting | yes |
| Atuin | Shell history | zsh and fish only |
| Starship | Prompt | yes |
| Zoxide | Smarter `cd` | yes |
| UutilsCoreutils, UutilsFindutils, UutilsDiffutils | Rust rewrites of GNU coreutils, findutils, diffutils | yes |
| Carapace | Completion for many commands | no |
| Glow | Markdown renderer for the MOTD | yes |

The MOTD is also on by default and is configured separately with
`bluefin-cli motd`.

## Enable or disable individual tools

Run the config menu to switch tools one at a time:

```sh
bluefin-cli shell config
```

You can reach the same choices from **Bluefin Shell** > **Configure
Components** in `bluefin-cli menu`. Choices are saved in `shell.json`. Re-run
`bluefin-cli shell <shell> on` afterwards so generated files, such as the
Nushell init file, pick up the change.

## Alpine and musl

Homebrew does not run on musl libc, so on Alpine-based systems Bluefin CLI uses
`coldbrew`, and falls back to `sudo apk add` one package at a time until
coldbrew is set up. To set it up:

```sh
bluefin-cli doctor --fix
```

Then enable the shell with `bluefin-cli shell ash on`.

## Verify

```sh
bluefin-cli status   # enabled shells and installed tools
bluefin-cli doctor   # checks package manager, shell integration, and tools
```

`doctor` and `status` act on the shell you are running, falling back to the
`SHELL` environment variable. `doctor --fix` enables shell integration and
installs missing `eza`, `fzf`, and `starship`.

## Performance

```sh
bluefin-cli doctor --bench
```

This compares interactive startup time against a bare shell (bash, zsh, and
fish). If startup is slow, turn off tools you do not use with
`bluefin-cli shell config`.

## Troubleshooting

- **Nothing changes in a new terminal.** Check that the startup file in the
  table above contains the managed line, then re-run
  `bluefin-cli shell <shell> on`.
- **An alias is missing.** The tool behind it is probably not installed or is
  disabled. Run `bluefin-cli status`, then `bluefin-cli doctor --fix`.
- **Changes not picked up in Nushell or ash.** See
  [Shell integration isn't picking up changes](troubleshooting.md#shell-integration-isnt-picking-up-changes).
- **Homebrew not found.** See [troubleshooting](troubleshooting.md#homebrew-is-installed-but-not-found).

After you update Bluefin CLI, re-run `bluefin-cli shell <shell> on` if the init
script behaves unexpectedly; for Nushell this regenerates the sourced file.
