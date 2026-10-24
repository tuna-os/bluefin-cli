# Linux Setup Guide

This guide covers setup instructions, package manager integration, configuration
resolution, and desktop environment integration for `bluefin-cli` on Linux.

## Installation

### Installation options

- **Direct script install** (recommended for standard glibc distributions):
  ```bash
  curl -fsSL https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.sh | sh
  ```
- **Native packages**: Download `.deb` or `.rpm` packages from GitHub releases.
- **Homebrew on Linux** (Linuxbrew):
  ```bash
  brew install tuna-os/homebrew-tap/bluefin-cli
  ```

### Distro package managers

- **glibc distributions** (Fedora, Ubuntu, Debian, Arch): Standard tools use
  Homebrew (Linuxbrew) when available. Ensure `brew shellenv` runs in your shell rc:
  ```bash
  eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
  ```
- **Alpine Linux and musl libc** (including postmarketOS):
  Homebrew does not run on musl libc. `bluefin-cli` detects Alpine Linux
  (`/etc/alpine-release`) and uses:
  1. `coldbrew` (rootless, sandboxed tool runner via bubblewrap) when available.
  2. `sudo apk add <package>` as a fallback.
  To prepare coldbrew on Alpine:
  ```bash
  sudo apk add coldbrew
  ```
  Running `bluefin-cli doctor --fix` automatically prepares coldbrew on Alpine systems.

## Configuration resolution

`bluefin-cli` resolves configuration in the following order:

1. `~/.config/bluefin-cli/` if that directory already exists.
2. `$HOMEBREW_PREFIX/etc/bluefin-cli/` if `$HOMEBREW_PREFIX` is set.
3. `~/.config/bluefin-cli/` (created on first run).

`bluefin-cli` complies with the XDG Base Directory specification:
- Configuration: `~/.config/bluefin-cli/config.yaml`
- Custom Brewfiles: `~/.config/bluefin-cli/bundles/*.Brewfile`
- Shell startup cache: `$XDG_CACHE_HOME/bluefin-cli/shell-cache` (or `~/.cache/bluefin-cli/shell-cache`)

## Shell integration

`bluefin-cli` manages shell configuration for Bash, Zsh, Fish, and Nushell on Linux.

Enable integration with:
```bash
bluefin-cli shell <shell-name> on
```

### musl libc shell notes

On musl systems such as Alpine Linux, default shells often include `ash` or `sh`.
Bluefin CLI provides full alias and init support for `ash`, `bash`, `zsh`, `fish`,
and `nushell`. When installing managed tools (such as `starship`, `eza`, or `fzf`),
tools install via `coldbrew` or `apk` without requiring glibc compatibility layers.

## Desktop integration

### GNOME

On GNOME desktop environments (`XDG_CURRENT_DESKTOP` contains `GNOME`), `bluefin-cli`
integrates with `gsettings` to control wallpaper and appearance:
- Wallpaper configuration updates `org.gnome.desktop.background picture-uri` and `picture-uri-dark`.
- Curated GNOME application bundles are documented in [GNOME Tools](gnome.md).

### Other desktop environments (KDE, XFCE, Sway, Hyprland)

On non-GNOME desktop environments:
- CLI tools, shell integrations, and terminal settings operate normally.
- Native wallpaper management via `gsettings` is skipped if `gsettings` is absent.
  Wallpapers installed by `bluefin-cli install wallpapers` are saved to standard
  locations (`~/.local/share/backgrounds/`, `~/Pictures/Wallpapers/`, or `/usr/share/backgrounds/`).

## Verification and troubleshooting

Run the built-in diagnostic tool to verify your configuration:
```bash
bluefin-cli doctor
```

To automatically resolve missing shell integrations or set up coldbrew on Alpine:
```bash
bluefin-cli doctor --fix
```

For more diagnostic guidance, see [Troubleshooting](troubleshooting.md).
