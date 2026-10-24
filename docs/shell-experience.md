# Shell Experience Setup Guide

Bluefin CLI provides an enhanced shell setup. It sets up modern tools, aliases, and prompts across shells and platforms.

## 1. What is Shell Experience?

Bluefin CLI adds tools to your daily terminal workflow:

- **Modern Aliases**: Replaces legacy tools with modern choices:
  - `ls`, `ll`, `l.`, and `l1` call `eza` for directory trees, icons, and permissions.
  - `cat` calls `bat` for color syntax and clean output in the terminal.
  - `grep`, `egrep`, `fgrep`, and `xzgrep` call `ugrep` for fast queries.
- **Enhanced Prompt**: Renders prompts with `starship` in every supported shell.
- **Directory Traversal**: Adds fast directory jumps with `zoxide`.
- **Command History**: Manages history search and sync through `atuin`.
- **Core Replacements**: Adds Rust GNU tools (`uutils-coreutils`, `uutils-findutils`, `uutils-diffutils`) to `PATH`.
- **Startup Cache**: Stores the init script at `~/.cache/bluefin-cli/init-<shell>.sh` (or Windows app data). This cache avoids delays when a shell opens.
- **Clean Fallbacks**: Skips missing tools without error output.

## 2. Supported Shells

| Shell | Identifier / Alias | Configuration File | Generated Script / Cache | Platform Support |
|---|---|---|---|---|
| **Bash** | `bash` | `~/.bashrc` | `~/.cache/bluefin-cli/init-bash.sh` | Linux, macOS, WSL |
| **Zsh** | `zsh` | `~/.zshrc` | `~/.cache/bluefin-cli/init-zsh.sh` | Linux, macOS, WSL |
| **Fish** | `fish` | `~/.config/fish/config.fish` | `~/.cache/bluefin-cli/init-fish.sh` | Linux, macOS, WSL |
| **PowerShell** | `powershell`, `pwsh` | `$PROFILE` (`Documents\PowerShell\...`) | `%LOCALAPPDATA%\bluefin-cli\shell-cache\init.ps1` | Windows, Linux, macOS |
| **Nushell** | `nu`, `nushell` | `~/.config/nushell/config.nu` | `~/.config/nushell/bluefin-cli.nu` | Linux, macOS, WSL, Windows |
| **ash / dash / sh** | `ash`, `dash`, `sh` | `~/.ashrc` (via `~/.profile`) | `~/.cache/bluefin-cli/init-ash.sh` | Alpine Linux, postmarketOS, POSIX |

## 3. Setup by Shell

Enable the shell setup with one command for each shell.

### Bash

Enable the setup:
```bash
bluefin-cli shell bash on
```

- **Configuration File**: `~/.bashrc`
- **Init Line**:
  ```bash
  eval "$(bluefin-cli init bash)" # managed by bluefin-cli
  ```
- **Verification**:
  Open a new Bash terminal or reload your rc file:
  ```bash
  source ~/.bashrc
  type ll
  ```

### Zsh

Enable the setup:
```bash
bluefin-cli shell zsh on
```

- **Configuration File**: `~/.zshrc`
- **Init Line**:
  ```bash
  eval "$(bluefin-cli init zsh)" # managed by bluefin-cli
  ```
- **Completion Notice**: The init script runs `compinit -i` if `compdef` is missing.
- **Verification**:
  Open a new Zsh terminal or reload your rc file:
  ```bash
  source ~/.zshrc
  which ll
  ```

### Fish

Enable the setup:
```bash
bluefin-cli shell fish on
```

- **Configuration File**: `~/.config/fish/config.fish`
- **Init Line**:
  ```fish
  bluefin-cli init fish | source # managed by bluefin-cli
  ```
- **Verification**:
  Open a new Fish shell:
  ```fish
  type ll
  ```

### PowerShell (Windows / pwsh)

Enable the setup:
```powershell
bluefin-cli shell powershell on
```

- **Configuration File**:
  - PowerShell 7+: `Documents\PowerShell\Microsoft.PowerShell_profile.ps1`
  - Windows PowerShell 5.1: `Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1`
- **Init Line**: Appends profile lines that load the cached script and mark user functions `global:`.
- **Verification**:
  Open a new PowerShell window:
  ```powershell
  Get-Command ll
  ```

### Nushell

Enable the setup:
```bash
bluefin-cli shell nu on
```

- **Configuration File**: `~/.config/nushell/config.nu`
- **Generated Script**: `~/.config/nushell/bluefin-cli.nu`
- **Init Line**:
  ```nu
  source ~/.config/nushell/bluefin-cli.nu # managed by bluefin-cli
  ```
- **Important**: Nushell parses syntax at startup and cannot run string-based `eval`. Bluefin CLI writes a file directly to disk. Nushell reads helper scripts from the vendor path (`$nu.data-dir/vendor/autoload`) for starship, zoxide, and atuin.
- **Verification**:
  Open a new Nushell session:
  ```nu
  which ll
  ```

### ash / dash (Alpine Linux & postmarketOS)

Enable the setup:
```sh
bluefin-cli shell ash on
```

- **Configuration File**: `~/.ashrc`
- **Environment Line**: Bluefin CLI appends `export ENV="$HOME/.ashrc" # managed by bluefin-cli` to `~/.profile`.
- **Supported Features**: Busybox ash uses POSIX aliases and `zoxide` prompt hooks (`zoxide init posix --hook prompt`). Tools without POSIX support (such as `starship` or `atuin`) do not run under ash.
- **Verification**:
  ```sh
  . ~/.profile
  alias ll
  ```

## 4. Managed Tools Explained

Bluefin CLI manages these tools:

| Tool | Binary | Default | Package (Brew / WinGet / Apk) | Description |
|---|---|---|---|---|
| **Eza** | `eza` | Enabled | `eza` / `eza-community.eza` | Modern `ls` tool with colors, icons, and tree views. |
| **Bat** | `bat` | Enabled | `bat` / `sharkdp.bat` | Modern `cat` tool with syntax colors. |
| **Fzf** | `fzf` | Enabled | `fzf` / `junegunn.fzf` | Fast fuzzy finder for commands and files. |
| **Starship** | `starship` | Enabled | `starship` / `Starship.Starship` | Fast prompt for all major shells. |
| **Zoxide** | `zoxide` | Enabled | `zoxide` / `ajeetdsouza.zoxide` | Smart `cd` tool that learns your paths. |
| **Ugrep** | `ug` | Enabled | `ugrep` / `ugrep` | Fast grep tool with query modes. |
| **Atuin** | `atuin` | Disabled (Enabled on Zsh/Fish) | `atuin` / `atuin` | History search tool with sync. |
| **UutilsCoreutils** | `ucat` | Enabled | `uutils-coreutils` | Rust rewrite of GNU coreutils. |
| **UutilsFindutils** | `ufind` | Enabled | `uutils-findutils` | Rust rewrite of GNU findutils. |
| **UutilsDiffutils** | `udiffutils` | Enabled | `uutils-diffutils` | Rust rewrite of GNU diffutils. |
| **Carapace** | `carapace` | Disabled | `carapace` / `rsteube.carapace` | Multi-shell completion tool. |
| **Gsudo** | `gsudo` | Enabled (PowerShell only) | `gerardog.gsudo` | Sudo tool for Windows terminals. |
| **Glow** | `glow` | Enabled | `glow` / `charmbracelet.glow` | Markdown tool for message of the day (MOTD). |

## 5. Installing on Alpine and musl Systems

Alpine Linux and postmarketOS use `musl` libc instead of `glibc`. Homebrew does not run on musl systems.

On Alpine, Bluefin CLI uses `coldbrew` (rootless sandbox), or falls back to `sudo apk add`:

1. Run the diagnostic tool:
   ```sh
   bluefin-cli doctor
   ```
2. If `coldbrew` is not ready, run the automatic fix:
   ```sh
   bluefin-cli doctor --fix
   ```
   This command installs `coldbrew` with `sudo apk add coldbrew`.
3. When `bluefin-cli shell ash on` runs on Alpine, tool installs run via `coldbrew install <pkg>` and `coldbrew wrap <pkg>`.

## 6. Enabling and Disabling Components

You can control the whole shell or toggle single tools.

### Full Shell Control

Enable or disable tools for an entire shell:
```bash
bluefin-cli shell bash on
bluefin-cli shell bash off
```

To view and toggle all shells interactively, run:
```bash
bluefin-cli shell
```

### Single Component Control

Configure individual tools in the TUI:
```bash
bluefin-cli shell config
```
This opens a form to toggle `eza`, `bat`, `starship`, `zoxide`, `atuin`, `ugrep`, and `motd`.

The CLI saves settings in `~/.config/bluefin-cli/shell.json`.

You can also pass flags directly to `bluefin-cli init`:
```bash
bluefin-cli init bash --atuin=false --starship=true
```

## 7. Verification

Verify your setup with built-in commands:

1. **Doctor Diagnostics**:
   ```bash
   bluefin-cli doctor
   ```
   Confirm that the "Shell integration" check prints `✓`.

2. **Configuration Status**:
   ```bash
   bluefin-cli status
   ```
   Inspect active shells and tool toggles.

3. **Manual Check**:
   Confirm that your shell defines the aliases:
   ```bash
   alias ll
   alias cat
   ```

## 8. Performance and Benchmarks

Bluefin CLI uses a disk cache to keep shell startup fast:

- **Startup Cache**: The CLI stores its init output at `~/.cache/bluefin-cli/init-<shell>.sh`. The cache updates when the `bluefin-cli` binary or `shell.json` changes.
- **Benchmark Startup Time**:
  Measure startup cost against a bare shell:
  ```bash
  bluefin-cli doctor --bench
  ```
  This command measures latency in Bash, Zsh, and Fish.
- **Nushell Autoload Refresh**:
  In Nushell, set `BLUEFIN_SKIP_AUTOLOAD_REFRESH=1` to skip updates to vendor scripts on every shell launch.

## 9. Troubleshooting

### Init Script Not Loading

- **Check RC File**: Inspect your config file (such as `~/.bashrc` or `~/.zshrc`). Confirm the `# managed by bluefin-cli` line is present.
- **Login Shells**: On macOS, terminals open login shells by default. Ensure your profile sources `~/.bashrc` or `~/.zshrc`.
- **Alpine ash**: Check that `~/.profile` exports `ENV="$HOME/.ashrc"`. Ash needs `ENV` to find `~/.ashrc` in interactive shells.

### Aliases Not Available

- **Missing Binary**: Bluefin CLI checks `PATH` before it sets an alias. The `ll` alias only exists when `eza` is installed. Run `bluefin-cli doctor --fix` to install missing tools.
- **Nushell Scope**: Nushell aliases exist at parse time. If you edit `~/.config/bluefin-cli/shell.json`, run `bluefin-cli shell nu on` to rebuild the script.

### Slow Startup

- Check startup latency: `bluefin-cli doctor --bench`.
- Confirm your user account can write to `~/.cache/bluefin-cli/`.
- If Atuin sync slows the prompt, check settings in `~/.config/atuin/config.toml`.

## 10. Updates and Maintenance

When you update `bluefin-cli`:

1. Update the binary:
   ```bash
   bluefin-cli update
   ```
2. The startup cache reads the timestamps on the new binary. The cache refreshes on the next shell start.
3. For Nushell or new profile paths, run:
   ```bash
   bluefin-cli shell <shell> on
   ```
   This rebuilds all generated files.
