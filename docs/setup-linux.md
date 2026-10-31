# Linux Setup Guide

This guide shows how to install and set up `bluefin-cli` on Linux.

## Install

- Install script:

  ```bash
  curl -fsSL https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.sh | sh
  ```

- Native packages: get a `.deb` or `.rpm` file from the GitHub releases page.
- Homebrew on Linux:

  ```bash
  brew tap ublue-os/homebrew-experimental-tap
  brew install bluefin-cli
  ```

## Alpine Linux and musl

Homebrew does not run on musl libc. The tool finds Alpine from the
`/etc/alpine-release` file. On Alpine, the tool installs packages with
`coldbrew` if it is available. If not, it uses `apk`.

To set up `coldbrew`, run this command:

```bash
bluefin-cli doctor --fix
```

## Files

- Settings: `~/.config/bluefin-cli/config.yaml`
- Your bundles: `~/.config/bluefin-cli/bundles/*.Brewfile`
- Shell init cache: `~/.cache/bluefin-cli/init-<shell>.sh`

For the full rules about the settings folder, see
[Configuration](configuration.md).

## Shell

To turn on the shell tools, run this command:

```bash
bluefin-cli shell bash on
```

Use `zsh`, `fish`, or `nu` in place of `bash` for other shells.

## Wallpaper

On GNOME, `bluefin-cli wallpaper` uses `gsettings` to set the wallpaper. It
sets the `picture-uri` and `picture-uri-dark` keys.

The wallpaper picker looks for images in these folders:

- `~/.local/share/backgrounds`
- `~/Pictures/Wallpapers`
- `/usr/share/backgrounds`

## Check your setup

```bash
bluefin-cli doctor
```

For more help, see the [help page](troubleshooting.md).
