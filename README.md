# Bluefin CLI

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/tuna-os/bluefin-cli/blob/main/LICENSE)

A powerful, modern CLI tool for managing shell configuration and development environment customization. Built with beautiful TUIs using [Charm](https://charm.sh/) libraries.

## ✨ Features

- **🎨 Interactive Menu**: Default TUI experience for easy navigation
- **✨ Bling**: Toggle modern shell enhancements (eza, bat, ugrep, zoxide, atuin, starship)
- **📰 MOTD**: Beautiful Message of the Day with system info and random tips
- **📦 Bundle Installer**: Install curated tool bundles (ai, cli, fonts, k8s) from Universal Blue
- **�️ Wallpapers**: Install desktop wallpaper collections from ublue-os/tap
- **🎨 Starship Themes**: Browse and apply Starship prompt themes
- **📊 Status Command**: View configuration and installed tools at a glance
- **🩺 Doctor**: `bluefin-cli doctor` diagnoses setup problems with fix hints
- **🎨 Theme Flavors**: `bluefin-cli theme <flavor>` pins a Catppuccin flavor (latte, frappe, macchiato, mocha) or follows your terminal with `auto`
- **⬆ Self-Update**: `bluefin-cli update` for script installs — sha256-verified against the release checksums; package-manager installs are pointed at the right upgrade command
- **🏠 My Brewfile**: one file describes your machine's packages — `brew`/`cask` lines plus `winget`/`scoop`/`choco` on Windows. `bluefin-cli brewfile dump` captures what's installed, `add`/`remove` edit it, `install` applies everything; the TUI's Install Apps → My Brewfile does all of it interactively with per-package management. Extra recipes in `~/.config/bluefin-cli/bundles/*.Brewfile` appear alongside the curated bundles
- **📦 Profiles**: `bluefin-cli profile export > setup.json` on one machine, `profile import setup.json` on another — shells, tools, and theme replayed exactly
- **🦕 A fully native TUI**: persistent shell with breadcrumbs, fuzzy filtering (`/`), a `ctrl+p` command palette, and a dot-matrix dino running the header — plus a hidden surprise for those who find it

## 🚀 Installation

### One-liner (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.sh | sh
```

### One-liner (Windows PowerShell)

```powershell
irm https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.ps1 | iex
```

Then enable shell integration:

```powershell
bluefin-cli shell powershell on
```

### Homebrew (Linux / macOS)

The formula is published automatically by GoReleaser to the
[`tuna-os/homebrew-tap`](https://github.com/tuna-os/homebrew-tap) tap on every
release ([#15](https://github.com/tuna-os/bluefin-cli/issues/15)):

```bash
brew tap tuna-os/tap
brew install bluefin-cli
```

> **Status (2026-08-11):** the tuna-os tap currently ships only `corral-vm.rb` —
> the `bluefin-cli` formula has not been published by the release pipeline yet
> (see [#141](https://github.com/tuna-os/bluefin-cli/issues/141)). Until it
> appears, use `ublue-os/homebrew-experimental-tap` below.

It is also available from `ublue-os/homebrew-experimental-tap`, synced from the
source-build formula in [`contrib/homebrew/bluefin-cli.rb`](contrib/homebrew/bluefin-cli.rb):

```bash
brew tap ublue-os/homebrew-experimental-tap
brew install bluefin-cli
```

### Winget (Windows)

```powershell
winget install --id Hanthor.BluefinCLI --exact
```

### Chocolatey (Windows)

```powershell
choco install bluefin-cli
```

### Scoop (Windows)

```powershell
scoop bucket add tuna-os https://github.com/tuna-os/scoop-bucket
scoop install bluefin-cli
```

> **Status (2026-08-11):** `tuna-os/scoop-bucket` has no manifests yet — the
> manifest is published by the release pipeline and is currently pending
> ([#141](https://github.com/tuna-os/bluefin-cli/issues/141)).

### deb / rpm (Debian, Ubuntu, Fedora, openSUSE…)

Releases are configured to ship native packages (`nfpm`) — grab the one for
your distro from the
[latest release](https://github.com/tuna-os/bluefin-cli/releases/latest)
when it is present:

```bash
# Debian/Ubuntu
sudo dpkg -i bluefin-cli_<version>_linux_amd64.deb
# Fedora & friends
sudo rpm -i bluefin-cli_<version>_linux_amd64.rpm
```

### AUR (Arch)

```bash
yay -S bluefin-cli-bin
```

### Go Install

```bash
go install github.com/tuna-os/bluefin-cli@latest
```

### Build from Source (Any OS)

**Prerequisites:**
- Go 1.25 or later

```bash
git clone https://github.com/tuna-os/bluefin-cli.git
cd bluefin-cli
go build -o bluefin-cli .
```

On Windows, use `go build -o bluefin-cli.exe .`.

Maintainers: package publishing (Homebrew tap, Winget, Scoop) is automated by GoReleaser on release; `.github/workflows/winget.yml` is a manual fallback for re-submitting a Winget version.

Scoop publishing requires the `SCOOP_BUCKET_TOKEN` repository secret (a fine-grained PAT with write access to `tuna-os/scoop-bucket`), without which the Scoop manifest upload step is safely skipped during release workflows. See [docs/release-publishing.md](docs/release-publishing.md) for details.

Homebrew tap publishing requires the `HOMEBREW_TAP_TOKEN` repository secret (a fine-grained PAT with write access to `tuna-os/homebrew-tap`), without which the formula upload step is safely skipped during release workflows. See [docs/release-publishing.md](docs/release-publishing.md) for details.

Homebrew release process: on every tagged release GoReleaser publishes the
binary formula to `tuna-os/homebrew-tap` (requires the `HOMEBREW_TAP_TOKEN`
secret). External taps that build from source, such as
`ublue-os/homebrew-experimental-tap`, must be synced manually — bump `url` and
`sha256` in [`contrib/homebrew/bluefin-cli.rb`](contrib/homebrew/bluefin-cli.rb)
and open a PR in that tap.

## 📖 Usage

### Interactive Menu (Default)

Simply run the command to launch the interactive menu:

```bash
bluefin-cli
```

Or explicitly:

```bash
bluefin-cli menu
```

### Command Line Usage

#### Check Status

View your current configuration and installed tools:

```bash
bluefin-cli status
```

#### Diagnose Problems

```bash
bluefin-cli doctor
```

#### Update

```bash
bluefin-cli update          # self-update (script installs)
bluefin-cli update --check  # just check
```

## ✨ Shell Experience

Bluefin CLI includes a "Shell Experience" module (formerly "bling") that configures your shell with modern tools and aliases.

To enable the shell experience:

```bash
bluefin-cli shell bash on
# or

bluefin-cli shell zsh on
# or

bluefin-cli shell fish on
```

Or use the interactive menu: `bluefin-cli menu` -> "Shell Experience".

Features:
- **eza**: Modern replacement for `ls`
- **bat**: Syntax highlighting for `cat`
- **ugrep**: Faster grep
- **zoxide**: Smarter `cd`
- **atuin**: Shell history sync
- **starship**: Cross-shell prompt
- **uutils**: Rust rewrite of coreutilsl

#### MOTD - Message of the Day

Show the MOTD:

```bash
bluefin-cli motd show
```

Toggle MOTD for shells:

```bash
# Enable for all shells

bluefin-cli motd toggle all on

# Enable for specific shell

bluefin-cli motd toggle zsh on

# Disable MOTD

bluefin-cli motd toggle all off
```

#### Install Tool Bundles

Install curated Homebrew bundles:

```bash
# List available bundles

bluefin-cli install list

# Install specific bundle

bluefin-cli install ai       # AI tools
bluefin-cli install cli      # CLI essentials
bluefin-cli install fonts    # Development fonts
bluefin-cli install k8s      # Kubernetes tools

# Interactive mode

bluefin-cli install
```

#### Install Wallpapers

Install desktop wallpaper collections:

```bash
# Interactive selection

bluefin-cli install wallpapers

# Install specific wallpaper casks

bluefin-cli install wallpapers bluefin-wallpapers aurora-wallpapers bazzite-wallpapers

# Non-interactive test run: apply theme + enable all automation

bluefin-cli install wallpapers bluefin-wallpapers --yes

# Non-interactive with explicit controls

bluefin-cli install wallpapers bluefin-wallpapers --non-interactive --apply-theme --theme Bluefin --enable-mode-sync --enable-auto-dark-light --trigger-source polling

# Use startup-only mode sync (no minute polling task)

bluefin-cli install wallpapers bluefin-wallpapers --non-interactive --enable-mode-sync --trigger-source startup

# Auto Dark Mode integration mode (startup sync + external mode-change utility)

bluefin-cli install wallpapers bluefin-wallpapers --non-interactive --enable-mode-sync --trigger-source autodarkmode

# Cleanup Windows sync artifacts/state/tasks generated by wallpaper integration

bluefin-cli install wallpapers cleanup

# Full reset for testing: cleanup + uninstall known wallpaper casks + local wallpaper folders

bluefin-cli install wallpapers cleanup --all

```

Non-interactive wallpaper flags:

- `--non-interactive`: Skip prompts and use provided flags.
- `--yes`: Shortcut for `--non-interactive --apply-theme --enable-mode-sync --enable-auto-dark-light`.
- `--apply-theme`: Apply a Windows theme after registration (WSL only).
- `--theme <name>`: Theme to apply in non-interactive mode (`Bluefin`, `Aurora`, `Bazzite`).
- `--enable-mode-sync`: Enable day/night wallpaper sync task.
- `--enable-auto-dark-light`: Enable 6 AM/6 PM light/dark switching tasks (requires `--enable-mode-sync`).
- `--trigger-source <source>`: Mode-sync trigger source (`polling`, `startup`, `autodarkmode`).

`autodarkmode` notes:
- Bluefin CLI ensures `%LOCALAPPDATA%\\BluefinCLI\\set-light-mode.ps1` and `%LOCALAPPDATA%\\BluefinCLI\\set-dark-mode.ps1` exist.
- In Auto Dark Mode, point light/dark custom script hooks to those two scripts.

#### Starship Themes
you can change your prompy lookks
Browse and apply Starship preset themes:

```bash
bluefin-cli starship theme
```

Install Starship if not already present:

```bash
bluefin-cli starship install
```


## 🔧 What Gets Configured

### Bling Tools

The bling command configures these modern CLI tools:

- **eza**: Modern replacement for `ls` with icons and colors
- **bat**: `cat` clone with syntax highlighting
- **zoxide**: Smarter `cd` command that learns your habits
- **atuin**: Magical shell history with sync and search (optional)
- **starship**: Fast, customizable prompt for any shell
- **ugrep**: Ultra-fast grep alternative (optional)

### Shell Aliases

When bling is enabled in your shell:

```bash
ll      # eza -l --icons=auto --group-directories-first
ls      # eza
cat     # bat --style=plain --pager=never
grep    # ugrep (if installed)
```

## 📚 Documentation

- [Interactive Menu Structure](docs/menus.md): A visual guide to the application's menu hierarchy and options.
- [Available Tools](docs/tools.md): A comprehensive list of all tools and bundles available.
- [Winget Publishing Guide](docs/winget.md): Steps to publish and maintain the Windows package.

## 🏗️ Project Structure

```
bluefin-cli/
├── main.go              # Application entry point
├── cmd/                 # Cobra commands
│   ├── root.go         # Root command & menu default
│   ├── menu.go         # Interactive TUI menu
│   ├── bling.go        # Bling command
│   ├── motd.go         # MOTD command
│   ├── install.go      # Install bundles/wallpapers
│   ├── starship.go     # Starship theme management
│   └── status.go       # Status display
├── internal/            # Internal packages
│   ├── bling/          # Bling logic & embedded scripts
│   ├── motd/           # MOTD generation
│   ├── install/        # Bundle & wallpaper installation
│   ├── starship/       # Starship integration
│   └── status/         # Status checking
└── test/                # Integration tests
```

## 📚 Inspiration

This project consolidates and modernizes functionality from:

- **ublue-bling**: Shell aliases and tool initialization scripts
- **bluefin-cli (cask)**: Homebrew package management and MOTD
- **ujust recipes**: Task runner and development environment helpers

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Podman (for containerized testing)
- just (for running recipes)

### Building

```bash
just build
```

### Testing

```bash
# Run tests in container

just test

# Run tests locally

go test ./...
```

### Interactive Development

Launch shells with bling pre-configured:

```bash
just bash   # Test in bash
just zsh    # Test in zsh
just fish   # Test in fish
```

### Dependencies

This project uses:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Huh](https://charm.land/huh/v2) - Forms and prompts
- [Lipgloss](https://charm.land/lipgloss/v2) - Style definitions

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## 🙏 Acknowledgments

- [Universal Blue](https://universal-blue.org/) - For the original bluefin-cli and ublue-bling
- [Charm](https://charm.sh/) - For the amazing TUI libraries
- The Homebrew community

## 🔗 Related Projects

- [ublue-os/packages](https://github.com/ublue-os/packages) - Original package implementations
- [Starship](https://starship.rs/) - Cross-shell prompt
- [Homebrew](https://brew.sh/) - Package manager for macOS and Linux

---

Part of the [TunaOS](https://tunaos.org) ecosystem. [Docs](https://tunaos.org) · [Contributing](CONTRIBUTING.md)