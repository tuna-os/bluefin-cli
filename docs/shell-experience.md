# Shell experience guide

The **shell experience** is the set of aliases, tool initializers, and prompt
settings that Bluefin CLI adds to your shell. When you enable it, it does three things:

1. Adds one managed line to your shell's startup file. That line runs
   `bluefin-cli init <shell>`, which prints the init script for that shell.
2. Tries to install the enabled tools (Homebrew on Linux and macOS, winget on
   Windows, coldbrew or `apk` on Alpine).
3. Records which tools are enabled in `shell.json` in the
   [config directory](configuration.md).

Open a new terminal after you enable it. Run `bluefin-cli status` to see the
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
from `~/.profile`. atuin, starship, and carapace have no ash target, so
they are not available there. See [Alpine and musl](#alpine-and-musl).

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
the binary or `shell.json` changes. If PowerShell blocks scripts, read the
execution policy notes in [the problem guide](troubleshooting.md). Tools install
with winget. Bluefin CLI offers `gsudo` on PowerShell only. It does not offer
ugrep or the uutils packages there.

## Managed tools

`bluefin-cli shell --help` lists every tool. The init script enables a tool
only when its binary is on `PATH`.

| Tool | What it does | On by default |
|---|---|---|
| Eza | Modern replacement for `ls` | yes |
| Gsudo | `sudo` for Windows (PowerShell only) | yes |
| Fzf | Fuzzy finder | yes |
| Ugrep | Fast grep with interactive mode | when installed |
| Bat | `cat` with syntax highlighting | yes |
| Atuin | Shell history | no |
| Starship | Prompt | yes |
| Zoxide | Smarter `cd` | yes |
| UutilsCoreutils, UutilsFindutils, UutilsDiffutils | Rust rewrites of GNU coreutils, findutils, diffutils | yes |
| Carapace | Completion for many commands | no |
| Direnv | Per-directory environment variables (bash, zsh, fish) | when installed |
| Mise | Tool versions and environments (bash, zsh, fish); `MISE_<SHELL>_AUTO_ACTIVATE=0` opts a shell out | when installed |
| Glow | Markdown renderer for the MOTD | yes |

`shell <shell> on` does not install the "when installed" tools. The init
script adds hooks for them when their binary is on `PATH`. Atuin stays off
even when it is installed, because Bluefin removed its shell integration.
Turn it on with `bluefin-cli shell config`. On Linux, `open` is an
alias for `xdg-open` where the system has no `open` command of its own.
In bash, the init script keeps the bash-preexec `DEBUG` trap armed after other
hooks rewrite `PROMPT_COMMAND`. This lets preexec hooks, such as atuin's, run.

In fish, the aliases are abbreviations, as in Bluefin's `bling.fish`. An
abbreviation expands on the command line, so you see the full command before
it runs. To use plain aliases, set `BLING_USE_ABBR` to `0` before the init
line in `config.fish`.

These are the hooks and aliases of Universal Blue's `bling.sh` and
`bling.fish`. When you enable the shell experience, you do not need to load
those files. `bluefin-cli shell <shell> on` and `off` both remove the block
that `ublue-bling` wrote (`### bling.sh source start` to `end`). They also
remove each line that loads `/usr/share/ublue-os/bling/bling.sh` or
`bling.fish`.

The MOTD is also on by default. Use `bluefin-cli motd` to configure it
separately.

## Enable or disable individual tools

Run the config menu to switch tools one at a time:

```sh
bluefin-cli shell config
```

You can reach the same choices from **Bluefin Shell** > **Configure
Components** in `bluefin-cli menu`. Bluefin CLI saves your choices in `shell.json`. Re-run
`bluefin-cli shell <shell> on` afterwards so generated files, such as the
Nushell init file, pick up the change.

## Alpine and musl

Homebrew does not run on musl libc. On Alpine systems, Bluefin CLI uses
`coldbrew`. Until you set up coldbrew, it uses `sudo apk add` for one package
at a time. To set it up:

```sh
bluefin-cli doctor --fix
```

Then enable the shell with `bluefin-cli shell ash on`.

## Verify

```sh
bluefin-cli status   # enabled shells and installed tools
bluefin-cli doctor   # checks package manager, shell integration, and tools
```

`doctor` and `status` act on the shell you use now. If they cannot find it, they
use the `SHELL` environment variable. `doctor --fix` enables shell integration and
installs missing `eza`, `fzf`, and `starship`.

## Performance

```sh
bluefin-cli doctor --bench
```

This compares the startup time of an interactive shell with a bare shell (bash,
zsh, and fish). If startup is slow, turn off tools you do not use with
`bluefin-cli shell config`.

## Troubleshooting

- **Nothing changes in a new terminal.** Make sure the startup file in the
  table above has the managed line. Then re-run
  `bluefin-cli shell <shell> on`.
- **An alias is missing.** The tool behind it is probably not installed or is
  disabled. Run `bluefin-cli status`, then `bluefin-cli doctor --fix`.
- **Nushell or ash does not use your changes.** Read
  [the notes on stale shell integration](troubleshooting.md#shell-integration-isnt-picking-up-changes).
- **Homebrew not found.** Read [the problem guide](troubleshooting.md#homebrew-is-installed-but-not-found).

After you update Bluefin CLI, re-run `bluefin-cli shell <shell> on` if the init
script does not work as you expect. For Nushell, this regenerates the sourced
file.
