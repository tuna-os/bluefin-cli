# Brewfile Advanced Usage Guide

Bluefin CLI provides a unified Brewfile workflow across Linux, macOS, and Windows. A single Brewfile can describe tools across multiple operating systems and package managers.

This guide explains how `bluefin-cli brewfile` works, how to manage cross-platform declarations, debug install issues, and structure portable setups.

---

## 1. Brewfile Syntax Reference

Bluefin CLI recognizes five entry kinds in a unified Brewfile:

```ruby
# Homebrew formulas and casks (macOS, Linux)
tap "homebrew/cask-fonts"
brew "ripgrep"
cask "ghostty"

# Windows package managers
winget "Microsoft.VisualStudioCode"
scoop "neovim"
choco "git"
```

### Supported Entry Kinds

| Kind | Target OS | Backend Command | Notes |
|---|---|---|---|
| `tap` | macOS, Linux | `brew bundle` | Registers a Homebrew repository/tap. Preserved on list and remove operations. |
| `brew` | macOS, Linux | `brew bundle` / `brew install` | Standard Homebrew CLI formula. |
| `cask` | macOS, Linux | `brew bundle` / `brew install --cask` | GUI applications and fonts managed via Homebrew. |
| `winget` | Windows | `winget install --exact --accept-package-agreements --accept-source-agreements --id <id>` | Official Windows Package Manager ID. |
| `scoop` | Windows | `cmd /c scoop install <id>` | Scoop package manager shim. |
| `choco` | Windows | `choco install -y <id>` | Chocolatey package manager. |
| `flatpak` | Linux | Embedded bundle syntax | Skipped by `brewfile add/remove/list`, parsed by internal installers. |

Lines starting with `#` are comments. Blank lines and comments are preserved when editing Brewfiles.

---

## 2. Cross-Platform Brewfiles

In mixed environments, engineers often want a single repository dotfile or shared setup file that works regardless of whether the machine runs macOS, Ubuntu, or Windows.

### How Bluefin CLI Executes a Unified Brewfile

When you run `bluefin-cli brewfile install --file ~/Brewfile`:

1. **Unix (`brew` and `cask`)**:
   - If `brew` or `cask` lines are present, `bluefin-cli` executes `brew bundle install --file=<path>`.
   - On Windows, or systems where `brew` is absent, `bluefin-cli` prints:
     ```text
     skipping brew/cask entries: homebrew not found (https://brew.sh)
     ```
     and cleanly continues to Windows packages without aborting.

2. **Windows (`winget`, `scoop`, `choco`)**:
   - Each package entry is inspected.
   - If the corresponding package manager binary is installed on `PATH`, `bluefin-cli` invokes the appropriate install command.
   - If the manager is missing (for instance, on Linux or macOS), `bluefin-cli` prints:
     ```text
     skipping winget "Microsoft.VisualStudioCode": winget not installed
     ```
     and continues processing remaining entries.

Because missing package managers are skipped gracefully rather than crashing the process, you can safely commit both Unix and Windows package lines into the exact same Brewfile.

---

## 3. Platform-Specific Sections

You can organize a unified Brewfile by platform using comment headers:

```ruby
# ==============================================================================
# Shared / Unix CLI Tools (macOS & Linux via Homebrew)
# ==============================================================================
tap "homebrew/bundle"
brew "bat"
brew "eza"
brew "fzf"
brew "ripgrep"
brew "starship"
brew "zoxide"

# ==============================================================================
# macOS Desktop Applications
# ==============================================================================
cask "ghostty"
cask "font-fira-code-nerd-font"

# ==============================================================================
# Windows Essentials (winget, scoop, choco)
# ==============================================================================
winget "Microsoft.VisualStudioCode"
winget "Git.Git"
scoop "7zip"
choco "curl"
```

When run on macOS or Linux, Homebrew installs the formulas and casks while Windows manager lines are skipped. When run on Windows, Homebrew lines are skipped while `winget`, `scoop`, or `choco` install the Windows tools.

---

## 4. Version Pinning

Homebrew and Windows package managers handle version pinning differently:

### Homebrew (`brew`)
Standard `brew` in Homebrew does not support arbitrary version strings on arbitrary formulas in a Brewfile. To pin versions in Homebrew:
- **Versioned Formulas**: Use versioned formula names provided by taps or core (e.g., `brew "python@3.12"`, `brew "node@20"`, `brew "openjdk@17"`).
- **Homebrew Pin**: Run `brew pin <formula>` after installation to prevent `brew upgrade` from updating that formula.
- **Tapped Commit / Extract**: If you need an older formula version, extract it into a custom tap using `brew extract <formula> <user/tap>`.

### Windows Managers
- **winget**: Supports pinning via the CLI command `winget pin add --id <id>`. In a Brewfile, declare the package ID.
- **choco**: If version-specific installation is needed with Chocolatey, packages can be pinned using `choco pin add -n=<id>`.

---

## 5. Repository and Tap Management

### Adding Homebrew Taps
To ensure formulas from third-party taps resolve correctly in Unix, specify `tap` lines before the formulas that rely on them:

```ruby
tap "ublue-os/tap"
tap "homebrew/cask-fonts"
cask "font-symbols-only-nerd-font"
```

### Third-Party Windows Sources
- **winget**: Ensure custom sources are registered prior to installation (`winget source add ...`). By default, `bluefin-cli` queries and installs from standard winget agreements (`--accept-source-agreements`).
- **scoop**: Add buckets before installing bucket-specific packages (`scoop bucket add extras`).
- **choco**: Add sources using `choco source add -n=<name> -s=<url>`.

---

## 6. Debugging Failed Installs

When `bluefin-cli brewfile install` runs, each manager handles errors distinctly:

1. **Which manager ran?**
   - Homebrew bundles run under `brew bundle install`. Homebrew prints its standard status and exit code.
   - Windows managers print their progress line before execution:
     ```text
     winget: installing Microsoft.VisualStudioCode
     ```
2. **Identifying failure causes**:
   - **Missing manager on PATH**: If a line is skipped unexpectedly, verify the manager is on `PATH`:
     ```bash
     command -v brew    # Linux / macOS
     where.exe winget   # Windows PowerShell / CMD
     ```
   - **Checksum or Hash Mismatch**: Usually indicates an outdated local formula index or mirror issue. Update your manager:
     ```bash
     brew update        # Homebrew
     winget source update  # WinGet
     ```
   - **Conflicting Existing Install**: If an application is already installed outside the manager, Homebrew Cask or winget might refuse to overwrite. Run the manager directly to view detailed output:
     ```bash
     brew install --cask <app>
     winget install --id <id> --exact
     ```
   - **Per-Package Windows Recovery**: Windows managers do not abort the entire batch when a single package fails. `bluefin-cli` reports:
     ```text
       winget <id> failed: exit status ...
     ```
     and continues with the next package.

---

## 7. Portable Configurations and Team Sharing

When creating Brewfiles for an organization or team repository:

1. **Avoid machine-specific paths**:
   Keep Brewfiles located at standard positions (`~/Brewfile` or checked into the root of a repository dotfiles project).
2. **Use `bluefin-cli brewfile dump` to bootstrap**:
   Generate an initial file from an existing setup:
   ```bash
   bluefin-cli brewfile dump --file ./Team.Brewfile
   ```
   On Windows, `dump` exports installed `winget`, `scoop`, and `choco` packages into the unified format. On macOS/Linux, it invokes `brew bundle dump`.
3. **Curate and prune**:
   Remove transient dependencies or machine-specific software from the dumped file before sharing with teammates.
4. **Custom Bundles Directory**:
   Place shared team bundles into:
   ```bash
   ~/.config/bluefin-cli/bundles/<team-name>.Brewfile
   ```
   `bluefin-cli` automatically exposes these bundles in `bluefin-cli install --list` and within the interactive TUI menu under **Install Apps**.

---

## 8. Integration with Profiles

Bluefin CLI provides two complementary configuration mechanisms:

- **Profiles (`bluefin-cli profile`)**: Captures shell preferences, enabled rc integrations (`bash`, `zsh`, `fish`, `powershell`), tool toggles (`eza`, `starship`, `zoxide`, `bat`, `atuin`), and theme flavors (`catppuccin` mocha/latte/etc.).
- **Brewfiles (`bluefin-cli brewfile`)**: Captures the exact list of software packages and CLI tools installed on the system.

### Unified Setup Workflow

To completely configure a fresh machine or clone a workstation:

1. **Export on Source Machine**:
   ```bash
   # Export shell tools and theme
   bluefin-cli profile export > my-profile.json

   # Dump installed packages
   bluefin-cli brewfile dump --file ~/MyBrewfile
   ```

2. **Replay on Target Machine**:
   ```bash
   # Apply shell tools, integrations, and theme
   bluefin-cli profile import my-profile.json

   # Install all packages
   bluefin-cli brewfile install --file ~/MyBrewfile
   ```

3. **Check for Configuration Drift**:
   ```bash
   bluefin-cli profile diff my-profile.json
   ```

---

## 9. Advanced Examples

### Example 1: Cross-Platform Developer Workstation

A complete Brewfile for developers working across Linux, macOS, and Windows:

```ruby
# Taps
tap "homebrew/bundle"

# Core Command-line Tools (Unix)
brew "bat"
brew "eza"
brew "fzf"
brew "jq"
brew "ripgrep"
brew "starship"
brew "zoxide"

# macOS GUI Applications
cask "ghostty"
cask "visual-studio-code"

# Windows Packages
winget "Microsoft.VisualStudioCode"
winget "Ghostty.Ghostty"
winget "Git.Git"
winget "Starship.Starship"
```

### Example 2: Web Development Environment

```ruby
# Node & Web Runtimes
brew "node@22"
brew "pnpm"
brew "bun"
brew "deno"

# Databases & Utilities (Unix)
brew "postgresql@16"
brew "redis"
brew "httpie"

# Windows Web Tools
winget "OpenJS.NodeJS.LTS"
winget "Schniz.fnm"
winget "PostgreSQL.PostgreSQL"
```

### Example 3: Data Science & ML Environment

```ruby
# Python & Scientific Tools
brew "python@3.12"
brew "uv"
brew "ruff"
brew "duckdb"

# Data Science Applications (macOS & Windows)
cask "rstudio"
winget "Posit.RStudio"
winget "Astral.uv"
```

To install any of these files, run:

```bash
bluefin-cli brewfile install --file /path/to/example.Brewfile
```
