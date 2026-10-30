# macOS Setup Guide

This guide covers setup instructions, Homebrew integration, terminal emulator
configuration, wallpaper automation, and Apple Silicon notes for `bluefin-cli`
on macOS.

## Installation

### Installation options

- **Homebrew tap** (recommended):
  ```bash
  brew install tuna-os/homebrew-tap/bluefin-cli
  ```
- **Direct script install**:
  ```bash
  curl -fsSL https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.sh | sh
  ```
- **Direct download**: Pre-built tarballs for Darwin (arm64 and x86_64) are available on GitHub releases.

### Architecture notes (Apple Silicon vs. Intel)

- **Apple Silicon (M1/M2/M3/M4, arm64)**:
  - Default Homebrew prefix: `/opt/homebrew`
  - Ensure Homebrew is on PATH by adding the shell evaluation to `~/.zprofile` or `~/.bash_profile`:
    ```bash
    eval "$(/opt/homebrew/bin/brew shellenv)"
    ```
- **Intel (x86_64)**:
  - Default Homebrew prefix: `/usr/local`
  - Shell evaluation:
    ```bash
    eval "$(/usr/local/bin/brew shellenv)"
    ```

## Configuration directory resolution

`bluefin-cli` resolves the configuration directory in the following order:

1. `~/.config/bluefin-cli/` if that directory exists.
2. `$HOMEBREW_PREFIX/etc/bluefin-cli/` if `$HOMEBREW_PREFIX` is set.
3. `~/.config/bluefin-cli/` (created on first run).

On macOS systems using Homebrew, setting `$HOMEBREW_PREFIX` permits system-wide or
prefix-wide configuration, while user overrides stay in `~/.config/bluefin-cli/`.

## Terminal setup

`bluefin-cli` configures fonts and integrations for macOS terminal emulators:

### Terminal.app

- Default system terminal on macOS.
- Font management writes settings to `com.apple.Terminal` via `defaults write`
  for `Font`, `NSFixedPitchFont`, and `NSFixedPitchFontSize`.
- Changes apply after restarting Terminal.app.

### iTerm2

- Detected when `~/Library/Application Support/iTerm2` is present.
- Configures dynamic profiles in `DynamicProfiles/bluefin-cli-font.plist` without
  mutating the main preferences file.

### Ghostty and WezTerm

- Ghostty installs via Homebrew cask:
  ```bash
  brew install --cask ghostty
  ```
- You can pin Ghostty or WezTerm to the macOS Dock:
  ```bash
  bluefin-cli terminal dock
  ```
  This command uses `dockutil` when installed, or updates `com.apple.dock`
  `persistent-apps` and restarts the Dock service.

## Wallpaper integration

`bluefin-cli` can configure desktop wallpapers natively on macOS:
- Preferred backend: `wallpaper` CLI from Homebrew (`brew install wallpaper`).
  This drives `NSWorkspace` directly and requires no extra prompt.
- Fallback backend: `osascript` AppleScript system event calls. This triggers
  a one-time macOS Automation privacy authorization prompt.
- Wallpaper candidate locations searched by `bluefin-cli install wallpapers` and picker:
  - `/opt/homebrew/Caskroom/*wallpaper*`
  - `/usr/local/Caskroom/*wallpaper*`
  - `~/Pictures/Wallpapers`
  - `~/.local/share/backgrounds`

## Shell setup

macOS defaults to Zsh. Enable shell tools and prompt integration with:
```bash
bluefin-cli shell zsh on
```
Also supports Bash, Fish, and Nushell. Ensure Homebrew-installed binaries
reside on your PATH before sourcing shell configuration.

## Verification and troubleshooting

Verify your installation with:
```bash
bluefin-cli doctor
```

If Homebrew tools are not found, confirm that `eval "$($(which brew) shellenv)"` is
present in your shell startup file. See [Troubleshooting](troubleshooting.md) for
more details.
