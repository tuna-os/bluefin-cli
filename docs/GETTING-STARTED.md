# Getting Started with Bluefin CLI

This guide helps you set up and use Bluefin CLI for the first time. It covers your first steps after installation, shell setup, package management, profiles, customizations, and common troubleshooting.

---

## 1. After Installation: Your First Steps

Once you have installed `bluefin-cli`, check your installation and explore the interface.

### Verify Your Environment

Start by running diagnostics and checking your configuration status:

```bash
# Check configuration and installed tools
bluefin-cli status

# Diagnose setup issues and verify dependencies
bluefin-cli doctor
```

`bluefin-cli doctor` checks your package manager, shell integration, managed CLI tools, GitHub connectivity, self-update capability, and release version.

If `doctor` reports missing tools or unconfigured shell integration, you can apply automatic fixes:

```bash
bluefin-cli doctor --fix
```

### Launch the Interactive Menu

Bluefin CLI features a full terminal user interface (TUI) powered by Charm. Launch it with:

```bash
bluefin-cli menu
```

*(You can also run `bluefin-cli` without arguments to launch the menu).*

**Key Navigation Shortcuts in the Menu:**
- `↑` / `↓` or `j` / `k`: Move selection cursor
- `Enter` or `→` or `l`: Select item or enter submenu
- `Esc` or `←` or `h`: Go back (quits at top-level menu)
- `/`: Fuzzy filter items in the active menu
- `Ctrl+P`: Global command palette to jump directly to any screen
- `?`: Show help overlay
- `q` or `Ctrl+C`: Exit

### Standard vs. Plus Builds

Bluefin CLI comes in two binary builds:

- **`bluefin-cli` (Standard)**: The lightweight core binary. Provides shell experience management, curated bundle installation, Homebrew/My Brewfile support, Starship prompt management, font downloads, and profiles.
- **`bluefin-cli-plus` (Plus)**: Built with `-tags extra`. Includes everything in Standard plus desktop automation: solar sunset/sunrise theme switching (Windows/WSL), automated Windows theme generation, and wallpaper integration.

The interactive menu dynamically adjusts its visible options based on which binary variant you run. See [Vanilla vs. Extra Features](vanilla-vs-extra.md) for full technical details.

---

## 2. Setting Up Your Shell

The **Shell Experience** equips your shell with modern CLI tools, intelligent shell completion, syntax highlighting, and fast directory switching.

### Supported Shells

Bluefin CLI natively supports:
- **Bash** (`~/.bashrc`)
- **Zsh** (`~/.zshrc`)
- **Fish** (`~/.config/fish/config.fish`)
- **PowerShell / pwsh** (`$PROFILE`)
- **Nushell** (`~/.config/nushell/config.nu`)
- **ash** (BusyBox on Alpine Linux / postmarketOS via `~/.ashrc` and `~/.profile`)

### Enabling the Shell Experience

Enable integration for your active shell:

```bash
# Detect and enable for Bash
bluefin-cli shell bash on

# Enable for Zsh
bluefin-cli shell zsh on

# Enable for Fish
bluefin-cli shell fish on

# Enable for PowerShell / pwsh
bluefin-cli shell powershell on

# Enable for Nushell
bluefin-cli shell nu on
```

You can also toggle shells interactively from `bluefin-cli menu` -> **Bluefin Shell**.

After turning shell integration on, restart your shell or source your shell's initialization file.

### Managed Shell Tools

When the shell experience is enabled, Bluefin CLI provides integrations and aliases for modern CLI utilities:

| Tool | Purpose | Replacement / Benefit |
|------|---------|-----------------------|
| `eza` | Modern directory listing | Replaces `ls` with tree views, Git status, and colors |
| `bat` | File viewer with wings | Replaces `cat` with syntax highlighting and Git paging |
| `ugrep` | Fast file searching | Interactive query UI and regex file searching |
| `fzf` | Command-line fuzzy finder | Interactive command history and path navigation |
| `zoxide` | Smarter `cd` command | Tracks frequent directories (`z <dir>`) |
| `starship` | Cross-shell prompt | Fast prompt showing Git branch, language runtimes, status |
| `atuin` | Shell history | Encrypted, searchable, cross-machine shell history |
| `carapace` | Multi-shell completion | Rich tab completions across hundreds of CLI tools |
| `glow` | Terminal Markdown reader | Formatted Markdown rendering in the terminal |
| `uutils` | Modern core utilities | Rust-based drop-in replacements for standard utils |

Manage individual tool integrations via CLI:

```bash
bluefin-cli shell config
```

---

## 3. Managing Packages and Tools

Bluefin CLI makes it easy to install curated tool sets and manage machine packages declaratively.

### Curated Tool Bundles

Install pre-configured bundles with a single command:

```bash
# List all available bundles
bluefin-cli install --list

# Install curated developer essentials
bluefin-cli install cli

# Install AI tools
bluefin-cli install ai

# Install Kubernetes ecosystem tools
bluefin-cli install k8s

# Install CNCF tools
bluefin-cli install cncf

# Install developer coding fonts
bluefin-cli install fonts
```

You can also browse and install bundles visually from `bluefin-cli menu` -> **Install Apps**.

### Declarative Package Management: My Brewfile

Bluefin CLI provides a cross-platform declarative package workflow using Brewfile format:

- **Linux / macOS**: Uses `brew` (formulae) and `cask` (GUI apps).
- **Windows**: Uses native entries like `winget`, `scoop`, and `choco`.

#### Common Brewfile Workflows

1. **Dump installed packages to a Brewfile:**
   ```bash
   bluefin-cli brewfile dump
   ```
   This captures currently installed packages into `~/.config/bluefin-cli/Brewfile`.

2. **View your tracked packages:**
   ```bash
   bluefin-cli brewfile list
   ```

3. **Search for packages across package managers:**
   ```bash
   bluefin-cli brewfile search ripgrep
   ```

4. **Add and remove packages:**
   ```bash
   # Add a package and install it
   bluefin-cli brewfile add ripgrep

   # Remove a package entry
   bluefin-cli brewfile remove ripgrep
   ```

5. **Install everything defined in your Brewfile:**
   ```bash
   bluefin-cli brewfile install
   ```

### Custom Bundles

You can create reusable custom bundles by dropping `.Brewfile` files into `~/.config/bluefin-cli/bundles/`:

```bash
mkdir -p ~/.config/bluefin-cli/bundles
cat << 'EOF' > ~/.config/bluefin-cli/bundles/webdev.Brewfile
brew "nodejs"
brew "pnpm"
cask "httpie"
EOF

# Install your custom bundle
bluefin-cli install webdev
```

Custom bundles appear automatically in both CLI listing (`bluefin-cli install --list`) and the interactive menu.

---

## 4. Replicating Setups with Profiles

Profiles allow you to capture your machine's state (enabled shells, tool selection, and theme flavor) and replicate it across machines.

### Exporting and Importing

Export your setup to a portable JSON file:

```bash
# Export configuration to a file
bluefin-cli profile export > my-setup.json

# Preview what importing a profile would change on this machine
bluefin-cli profile diff my-setup.json

# Apply the profile to the current machine
bluefin-cli profile import my-setup.json
```

### Syncing Across Machines via GitHub Gist

If you have GitHub CLI (`gh`) configured, sync profiles securely across machines:

```bash
# Push profile to a private GitHub Gist
bluefin-cli profile push

# Pull and apply the profile on another machine using the Gist ID
bluefin-cli profile pull <gist-id>
```

---

## 5. Customization

Customize the terminal environment to match your workflow and aesthetic preferences.

### Theme Flavors

Bluefin CLI uses Catppuccin color schemes for the TUI and terminal tools:

```bash
# Set automatic mode (adapts to terminal dark/light background)
bluefin-cli theme auto

# Pin specific Catppuccin flavors
bluefin-cli theme latte      # Light
bluefin-cli theme frappe     # Medium dark
bluefin-cli theme macchiato  # Dark
bluefin-cli theme mocha      # High-contrast dark
```

### Prompt Themes (Starship)

Switch Starship prompt presets directly:

```bash
# List available starship presets
bluefin-cli starship list

# Apply a preset theme
bluefin-cli starship use pastel-powerline
```

Or browse themes interactively via `bluefin-cli menu` -> **Starship Theme**.

### Desktop Wallpapers

Download and apply curated wallpaper packs from `ublue-os/tap`:

```bash
# View current desktop wallpaper
bluefin-cli wallpaper

# Set desktop wallpaper (macOS and Linux GNOME)
bluefin-cli wallpaper /path/to/image.png
```

### Message of the Day (MOTD)

Display system metrics, weather, and tips on shell launch:

```bash
# Display the MOTD in your current terminal
bluefin-cli motd show

# Toggle MOTD on or off for your shell
bluefin-cli motd toggle bash on
bluefin-cli motd toggle all off

# Configure MOTD sections
bluefin-cli motd config
```

### Modern Terminal Setup

Set up Ghostty or WezTerm with Catppuccin themes and Nerd Fonts:

```bash
# Run non-interactive full terminal setup
bluefin-cli terminal setup
```

---

## 6. Platform-Specific Notes

### Linux

- **Package Managers**: Homebrew is preferred for user-space bundles. On Alpine Linux or musl-based distributions, `coldbrew` is used.
- **Desktop Environments**: Wallpaper settings and GNOME-specific bundles target GNOME environments via `gsettings`.
- **Packages**: Native `.deb` and `.rpm` packages are distributed with each GitHub release.

### macOS

- **Package Managers**: Homebrew (`brew`) is the primary backend. Ensure Homebrew is on your PATH (`eval "$(/opt/homebrew/bin/brew shellenv)"`).
- **Terminal & Wallpapers**: Native wallpaper switching uses AppleScript or the Homebrew `wallpaper` package. Ghostty setup supports native Dock pinning.

### Windows & WSL

- **Shell**: Use PowerShell 7 (`pwsh`) or Windows PowerShell.
- **Execution Policy**: If `install.ps1` fails, execute with execution bypass:
  ```powershell
  powershell -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.ps1 | iex"
  ```
- **Package Managers**: My Brewfile maps formulas to `winget`, `scoop`, or `choco` on Windows.
- **Sunset Automation**: In `bluefin-cli-plus`, solar sunset switching runs natively on Windows or bridges from WSL to change desktop themes at dusk and dawn.

---

## 7. Troubleshooting Common First-Time Issues

When encountering issues, run `bluefin-cli doctor` first. Here are solutions to frequent first-time hurdles:

### Shell Not Recognizing Aliases or Functions

1. Confirm shell integration is enabled:
   ```bash
   bluefin-cli status
   ```
2. Restart your terminal session, or reload your configuration:
   - Bash: `source ~/.bashrc`
   - Zsh: `source ~/.zshrc`
   - Fish: `source ~/.config/fish/config.fish`
   - Nushell: restart `nu` (or re-run `bluefin-cli shell nu on` if you updated tools)
   - PowerShell: `. $PROFILE`

### Tools Not Found in PATH

If tools like `eza` or `starship` are reported missing:
1. Run `bluefin-cli doctor --fix` to automatically install missing managed tools.
2. If using Homebrew on macOS/Linux, ensure Homebrew's `bin` directory is in your `$PATH`.

### Updating Bluefin CLI

Keep Bluefin CLI up to date:

```bash
# Check for updates
bluefin-cli update --check

# Apply self-update (for curl/irm script installations)
bluefin-cli update
```

For package manager installs, use your system package manager (e.g., `brew upgrade bluefin-cli` or `winget upgrade Hanthor.BluefinCLI`).

For more troubleshooting details, see [Troubleshooting](troubleshooting.md).
