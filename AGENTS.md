# AGENTS.md

## 🤖 Project Overview
**Bluefin CLI** is a Go-based command-line tool designed for the Bluefin OS (and other Universal Blue derivatives). It serves as a unified interface for system customization, "bling" (shell enhancements), and software installation via Homebrew bundles.

## 🏗 Architecture
The project follows a standard Go CLI structure:
- **`cmd/`**: Contains the main entry point and Cobra commands. Each command file (e.g., `bling.go`, `install.go`) typically handles the CLI arguments and delegates logic to `internal/`.
- **`internal/`**: Contains the core business logic, separated by domain (`bling`, `install`, `motd`, etc.).
- **TUI**: Heavily uses [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss), and [Huh](https://github.com/charmbracelet/huh) for interactive menus.

### Key Components
1.  **Bling (`internal/bling`)**: Manages shell configuration files (`.bashrc`, `.zshrc`, `config.fish`). It embeds shell scripts (`resources/bling.sh`, `resources/bling.fish`) and sources them in the user's shell config.
2.  **Bundles (`internal/install`)**: Defines a hardcoded list of "bundles" (e.g., ai, k8s) mapping to remote Brewfiles hosted on GitHub. It downloads these Brewfiles and runs `brew bundle install`.

## 🛠 Development
The project uses `just` as a task runner.

### Common Commands
- **`just build`**: Builds the binary locally.
- **`just test`**: Runs integration tests inside a container.
- **`just unit-test`**: Runs all Go unit tests inside a container.
- **`just shell-with-bling`**: Spawns a container with the CLI pre-installed and "bling" enabled for manual testing.
- **`just inspect-bling`**: Verifies that shell configuration files are correctly modified.

### Testing Guidelines
- **Unit Tests**: Add unit tests for new logic in the respective `internal/` package. Run them locally with `go test ./internal/...` or via `just unit-test`.
- **Integration Tests**: Integration tests are located in `test/integration_test.go`. Run them via `just test`.
- **Manual Verification**: Use `just shell-with-bling` to verify UI and shell configuration changes in an isolated environment.

### Guidelines for Agents
- **Changing Bundles**: Update `internal/install/install.go` to add/remove bundles or change the source URL.
- **Modifying Bling**:
    - If changing the shell script logic, edit `internal/bling/resources/bling.sh` or `.fish`.
    - If changing how it hooks into shells, edit `internal/bling/bling.go`.
- **UI Changes**: TUI logic is often inline in `cmd/` for simple commands or in `cmd/menu.go` for the main menu. Use `lipgloss` styles defined in the respective files.
