# macOS Setup Guide

This guide shows how to install and set up `bluefin-cli` on macOS.

## Install

- Homebrew:

  ```bash
  brew tap ublue-os/homebrew-experimental-tap
  brew install bluefin-cli
  ```

- Install script:

  ```bash
  curl -fsSL https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.sh | sh
  ```

- Archive: get the Darwin archive from the GitHub releases page.

Homebrew is in `/opt/homebrew` on Apple silicon and in `/usr/local` on Intel.
Put `brew shellenv` in your shell startup file:

```bash
eval "$(/opt/homebrew/bin/brew shellenv)"
```

## Terminal

The `bluefin-cli terminal setup` command sets up Ghostty. It also adds Ghostty
to the Dock. It uses `dockutil` if it is available.

When the tool sets a terminal font, it changes these terminals:

- Terminal.app: the tool writes the font to `com.apple.Terminal` with
  `defaults write`. Start Terminal.app again to see the change.
- iTerm2: the tool writes a dynamic profile to
  `DynamicProfiles/bluefin-cli-font.plist`.

## Wallpaper

The `bluefin-cli wallpaper` command uses the `wallpaper` CLI if it is
available. If not, it uses `osascript`. macOS can then ask for automation
permission one time.

The wallpaper picker looks for images in these folders:

- `/opt/homebrew/Caskroom` and `/usr/local/Caskroom` (wallpaper casks)
- `~/.local/share/backgrounds`
- `~/Pictures/Wallpapers`
- `/usr/share/backgrounds`

## Shell

macOS uses Zsh by default. To turn on the shell tools, run this command:

```bash
bluefin-cli shell zsh on
```

## Check your setup

```bash
bluefin-cli doctor
```

If the tool cannot find Homebrew programs, make sure that your shell startup
file runs `brew shellenv`. For more help, see
the [help page](troubleshooting.md).
