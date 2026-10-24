# Developer Guide: Code Organization

This guide explains the code organization, package structure, and architecture of `bluefin-cli`.

## Overview and Binary Variants

`bluefin-cli` is a Go CLI tool built with [Cobra](https://github.com/spf13/cobra) for commands and [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) for its interactive TUI.

The repository builds two binaries from one source tree:

1. **`bluefin-cli` (Standard)**: The core binary. It provides package installation, shell configuration, health diagnostics, profiles, and basic menus.
2. **`bluefin-cli-plus` (Plus)**: Built with `-tags extra`. It adds desktop wallpaper management, font installation, solar sunset automation, and interactive screens in the menu.

Build tags select extra features. Files with `//go:build extra` compile only into the plus binary. Stubs for the standard build live in `cmd/extra_stubs.go`.

---

## Directory Layout

```
bluefin-cli/
├── main.go                       # Application entry point
├── cmd/                          # Cobra commands, flag parsing, and TUI wiring
├── internal/                     # Domain packages and core business logic
│   ├── config/                   # Persistent user configuration
│   ├── countme/                  # Privacy-respecting weekly user telemetry
│   ├── doctor/                   # System diagnostic checks and automated fixes
│   ├── env/                      # Platform, OS, and environment detection
│   ├── install/                  # Package, bundle, and wallpaper installers
│   ├── motd/                     # Message of the Day generation and display
│   ├── pkgmanager/               # Package manager abstractions (Alpine, etc.)
│   ├── profile/                  # Portable system configuration profiles
│   ├── shell/                    # Shell integration and tool configuration
│   ├── starship/                 # Starship prompt configuration and presets
│   ├── status/                   # Configuration state and inspection views
│   ├── sunset/                   # Day/night solar automation (WSL/Windows)
│   ├── terminal/                 # Terminal emulator setup and font management
│   ├── tui/                      # Bubble Tea TUI shell and presentation styles
│   ├── update/                   # Checksum-verified self-update
│   └── wallpaper/                # Desktop wallpaper preview and setter logic
├── docs/                         # User and architecture documentation
│   └── commands/                 # Generated command reference docs
├── scripts/                      # Validation and smoke testing scripts
└── test/                         # End-to-end and integration test suites
```

---

## Command Routing (`cmd/`)

The application entry point is `main.go`. It initializes user configuration with `config.Init()` and delegates execution to `cmd.Execute()`.

The `cmd/` package holds Cobra commands. Keep argument flags and input checks in `cmd/`; keep business logic and domain execution in `internal/`.

Key files in `cmd/`:

- `root.go`: Defines the root Cobra command (`rootCmd`) and global persistent flags. When run without subcommands in an interactive terminal, it opens the main menu.
- `menu.go`: Wires the interactive menu hierarchy and screen navigation.
- `menu_native.go`, `menu_extra.go`, `menu_vanilla.go`: Build menu entries based on compile-time build tags.
- `shell.go`: Handles CLI commands to turn shell integrations on or off (`bluefin-cli shell <name> on|off`).
- `install.go`: Handles package and bundle installation commands (`bluefin-cli install <bundle>`).
- `doctor.go`: Handles diagnostic checks (`bluefin-cli doctor`) and automated repairs (`--fix`).
- `profile.go`: Commands to export, import, and sync profiles for user setup.
- `status.go`: Shows active status of shells, installed tools, and system state.
- `extra_stubs.go` and `extra_helpers.go`: Provide build-tag fallback logic for plus-only features when you compile the standard binary.

---

## Key `internal/` Packages

### `internal/shell/` — Shell Configuration and Integration

Manages shell integration across supported shells: Bash, Zsh, Fish, ash (BusyBox), Nushell, and PowerShell.

- `shells.go`: Defines the central registry of managed shells. Each shell specifies the path to its startup file, an initialization line, and syntax flavor. Always register a new shell in this registry.
- `shell.go`: Creates startup hook scripts, manages edits to shell files, and checks integration status.
- `tools.go`: Defines tools for the CLI (such as `eza`, `bat`, `ugrep`, `zoxide`, `atuin`, `starship`, and `carapace`).
- `installers.go`: Coordinates tool installations through the detected package managers on the system.
- `nu_aliases.go`: Generates parse-time aliases for Nushell.
- `powershell.go` and `windows_tools.go`: Manage scopes in PowerShell scripts and Winget tool installations.
- `install_alpine.go`: Handles package installs with `coldbrew` or `apk` on Alpine Linux systems.

### `internal/tui/` — Terminal User Interface

Built on Bubble Tea v2, Lipgloss v2, and Huh.

- `internal/tui/style.go`: Palettes of Catppuccin colors, border styles, and text styles.
- `internal/tui/theme/`: Theme flavor registry (Latte, Frappe, Macchiato, Mocha, and Auto detection).
- `internal/tui/app/`: Interactive application shell:
  - `app.go`: Coordinates the navigation screen stack, global key bindings, headers, and footers.
  - `action.go`: Command palette (`ctrl+p`) action registry. Each module registers its own palette items.
  - `menu.go`: Standard list menus.
  - `form.go`: Multi-select form screens (`app.FormScreen`).
  - `text.go`: Scrollable text viewers for read-only content (`app.TextScreen`).
  - `runner.go`: Command execution screen with an animated spinner and live output logging (`app.RunnerScreen`).
  - `exec.go`: Process bridge for commands that take terminal control (`app.RunExternal`).
  - `dino.go` and `game.go`: Hidden mini-game with a dot-matrix runner.

### `internal/install/` — Package and Bundle Management

Handles package installations from curated Homebrew bundles, Winget, and Alpine APKs.

- `install.go`: Core installer orchestration and bundle resolution.
- `packages.go`: Resolves package names and dependencies across package managers.
- `brewfile.go`: Parses, formats, and executes Homebrew `Brewfile` workflows.
- `resources/`: Embedded bundle definitions (`ai.Brewfile`, `cli.Brewfile`, `k8s.Brewfile`, etc.) and `PROVENANCE.json`.
- `wallpapers.go`: Manages wallpaper installations from casks and directory data.
- `windows.go` and `windows_installer.go`: Map Homebrew packages to IDs of Winget packages.

### `internal/config/` — Configuration

Manages settings for configuration in `~/.config/bluefin-cli/config.json`. Loads and saves user choices such as theme flavor, telemetry preference, and shell tool selections.

### `internal/profile/` — Profiles

Enables users to export their machine setup into a portable JSON profile, import profiles onto new systems, inspect diffs, and sync setups across workstations.

### `internal/terminal/` — Terminal Interaction and Fonts

Configures terminal emulators (such as Ghostty and WezTerm) and installs fonts (Nerd Fonts). Provides platform-specific font activation for macOS, Linux, and Windows.

### `internal/doctor/` — Diagnostic Engine

Runs health checks on the user environment (package manager presence, shell integration state, tool availability, network connectivity, and file permissions). Provides automated fixes for common configuration issues.

### `internal/status/` — System Status

Inspects environment variables, default shells, active tools, and package manager states to produce formatted CLI and TUI reports.

### `internal/sunset/` — Solar Automation (Plus variant)

Calculates solar sunrise and sunset times based on location coordinates. Automatically switches themes for the desktop, terminal color schemes, and wallpaper styles.

### `internal/update/` — Self-Update

Queries GitHub releases, downloads release assets, verifies SHA-256 checksums, and safely replaces the current binary on script-based installations.

---

## Common Contributor Tasks

### 1. Adding a New CLI Command

1. Create a new command file in `cmd/` (for example, `cmd/mycommand.go`).
2. Define a Cobra command struct and add it to `rootCmd` via `init()`:
   ```go
   var myCmd = &cobra.Command{
       Use:   "mycommand",
       Short: "Short description of mycommand",
       RunE: func(cmd *cobra.Command, args []string) error {
           return mypackage.Run()
       },
   }

   func init() {
       rootCmd.AddCommand(myCmd)
   }
   ```
3. Put the core logic in a package under `internal/`.
4. If the command should appear in the interactive command palette, register an `app.Action` in the TUI palette registry.
5. Regenerate markdown documentation:
   ```bash
   just gen-docs
   ```

### 2. Adding or Modifying a Supported Shell

1. Edit `internal/shell/shells.go`.
2. Add a new `ShellDef` entry to the `SupportedShells` list with the shell name, configuration path, and initialization command template.
3. If the shell has syntax restrictions (such as Nushell or PowerShell), add syntax generation in `internal/shell/`.
4. Validate changes using the shell test script:
   ```bash
   scripts/shell-experience.sh ./bluefin-cli <shell-name>
   ```

### 3. Adding a New Managed Tool

1. Add the tool definition to `Tools` in `internal/shell/tools.go`. Specify package names for Homebrew, Winget, and Alpine APK.
2. If the tool needs special init code for shells, update `internal/shell/shell.go`.
3. Add tests in `internal/shell/tools_test.go` and `internal/shell/shell_test.go`.

---

## Test Organization and Conventions

The project uses three layers of tests:

1. **Unit and Package Tests (`go test ./...`)**:
   - Each package contains unit tests (`*_test.go`) alongside the source code.
   - Run tests with race detection and the extra build tag:
     ```bash
     go test -tags extra -race ./...
     ```
2. **Integration and State Tests (`test/`)**:
   - Located in the `test/` directory (`integration_test.go`, `state_test.go`).
   - Verifies CLI execution against real filesystem state and configuration files.
   - Run via `just test` (runs inside a Linux container) or locally with compiled binaries.
3. **TUI Smoke and Experience Tests (`scripts/`)**:
   - `scripts/tui-smoke.sh`: Launches the TUI inside a tmux session and verifies key events and screens.
   - `scripts/shell-experience.sh`: Launches supported shells to verify aliases, prompt speed, and silent startup.
