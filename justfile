# justfile for bluefin-cli development

# Default recipe - show available commands
default:
    @just --list

deps:
    @echo "Installing dependencies..."
    @brew install go gum zoxide atuin starship eza bat ugrep 

# Run Go tests (canonical test suite)
test: build-container build
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running Go tests in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        go test -v ./test/...
    echo "Go tests completed!"

# Build the development container image (if not exists or force rebuild)
build-container:
    #!/usr/bin/env bash
    if ! podman image exists bluefin-cli-dev; then
        echo "Building development container image..."
        podman build -t bluefin-cli-dev -f Containerfile.dev .
    else
        echo "Development container image already exists (use 'just rebuild-container' to force rebuild)"
    fi

# Force rebuild the development container
rebuild-container:
    @echo "Rebuilding development container image..."
    podman build -t bluefin-cli-dev -f Containerfile.dev .

# Run unit tests in container
unit-test: build-container
    @echo "Running unit tests in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go test ./... -v

motd-test: build-container build
    @echo "Running motd tests in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go test -v ./internal/motd/...
    @echo "Motd tests completed!"

# Build the standard binary (Standard features only)
build-standard:
    @echo "Building standard binary..."
    go build -o bluefin-cli

# Build the plus binary (Everything)
build-plus:
    @echo "Building plus binary..."
    go build -tags extra -o bluefin-cli-plus

# Build both binaries
build-all: build-standard build-plus

# Build the binary locally (default to both)
build: build-all


# Open an interactive shell in the development container
shell: build-container build
    @echo "Opening interactive shell in development container..."
    @echo "Binary is ready at: ./bluefin-cli-plus"
    @echo ""
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash

# Open shell in container with bling already enabled (for manual testing)
shell-with-bling: build-container build
    #!/usr/bin/env bash
    echo "Setting up container with bling enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli bling bash on && \
                 ./bluefin-cli bling zsh on && \
                 ./bluefin-cli bling fish on && \
                 ./bluefin-cli motd toggle bash on && \
                 echo "" && \
                 echo "=== Bling has been enabled ===" && \
                 echo "Binary: ./bluefin-cli" && \
                 echo "Configs: ~/.bashrc, ~/.zshrc, ~/.config/fish/config.fish" && \
                 echo "Bling scripts: ~/.local/share/bluefin-cli/bling/" && \
                 echo "" && \
                 echo "Try: ./bluefin-cli status" && \
                 echo "     cat ~/.bashrc" && \
                 echo "     cat ~/.local/share/bluefin-cli/bling/bling.sh" && \
                 echo "" && \
                 bash'

# Open bash with bling enabled and sourced
bash: build-container build
    #!/usr/bin/env bash
    echo "Launching bash with bling enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        -e PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli bling bash on > /dev/null 2>&1 && \
                 echo "✓ Bling enabled - Tools: starship=$(command -v starship), eza=$(command -v eza)" && \
                 exec bash'

# Open zsh with bling enabled and sourced
zsh: build-container build
    #!/usr/bin/env bash
    echo "Launching zsh with bling enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        -e SHELL=/bin/zsh \
        -e PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish ~/.local/share/bluefin-cli/bling && \
                 rm -f ~/.zshrc ~/.local/share/bluefin-cli/bling/bling.sh && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli bling zsh on > /dev/null 2>&1 && \
                 echo "✓ Bling enabled - Tools: starship=$(command -v starship), eza=$(command -v eza)" && \
                 ZDOTDIR=/root exec zsh'

# Open fish with bling enabled and sourced
fish: build-container build
    #!/usr/bin/env bash
    echo "Launching fish with bling enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        -e PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli bling fish on > /dev/null 2>&1 && \
                 exec fish'

# Inspect what files were created by bling
inspect-bling: build-container build
    @echo "Inspecting bling files in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli bling bash on && \
                 ./bluefin-cli bling zsh on && \
                 ./bluefin-cli bling fish on && \
                 echo "=== Shell Configs ===" && \
                 echo "" && \
                 echo "--- ~/.bashrc ---" && \
                 cat ~/.bashrc && \
                 echo "" && \
                 echo "--- ~/.zshrc ---" && \
                 cat ~/.zshrc && \
                 echo "" && \
                 echo "--- ~/.config/fish/config.fish ---" && \
                 cat ~/.config/fish/config.fish && \
                 echo "" && \
                 echo "=== Bling Scripts ===" && \
                 echo "" && \
                 echo "--- bling.sh (first 50 lines) ---" && \
                 head -50 ~/.local/share/bluefin-cli/bling/bling.sh && \
                 echo "" && \
                 echo "--- bling.fish (first 30 lines) ---" && \
                 head -30 ~/.local/share/bluefin-cli/bling/bling.fish'

# Show what MOTD looks like
inspect-motd: build-container build
    @echo "Inspecting MOTD in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash -c 'touch ~/.bashrc && \
                 ./bluefin-cli motd toggle bash on && \
                 echo "=== MOTD Configuration ===" && \
                 echo "" && \
                 echo "--- Tips available ---" && \
                 ls -1 ~/.local/share/bluefin-cli/motd/tips/ && \
                 echo "" && \
                 echo "--- Sample tip (01-tip.md) ---" && \
                 cat ~/.local/share/bluefin-cli/motd/tips/01-tip.md && \
                 echo "" && \
                 echo "--- MOTD show output ---" && \
                 ./bluefin-cli motd show'

# Run a specific command in container for debugging
run CMD: build-container build
    @echo "Running command in container: {{CMD}}"
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash -c '{{CMD}}'

# Clean up built artifacts
clean:
    @echo "Cleaning up built artifacts..."
    rm -f bluefin-cli
    go clean

# Clean up container images
clean-containers:
    @echo "Removing development container image..."
    -podman rmi bluefin-cli-dev
    -podman rmi bluefin-cli-test

# Full clean (artifacts + containers)
clean-all: clean clean-containers

# Run linter in container
lint: build-container
    @echo "Running linter in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        bash -c 'command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping..."'

# Format code
fmt:
    @echo "Formatting Go code..."
    go fmt ./...

# Generate Markdown documentation for all commands
gen-docs: build
    @./bluefin-cli docs --dest ./docs/commands

# Update embedded resources (Brewfiles) from upstream
update-resources:
    #!/usr/bin/env bash
    set -euo pipefail
    BASE_URL="https://raw.githubusercontent.com/projectbluefin/common/main/system_files"
    DEFAULT_PATH="shared/usr/share/ublue-os/homebrew"
    BLUEFIN_PATH="bluefin/usr/share/ublue-os/homebrew"
    
    mkdir -p internal/install/resources/brewfiles
    
    FILES=(
        "ai-tools.Brewfile"
        "cli.Brewfile"
        "cncf.Brewfile"
        "experimental-ide.Brewfile"
        "fonts.Brewfile"
        "ide.Brewfile"
        "k8s-tools.Brewfile"
    )
    
    echo "Updating common Brewfiles..."
    for file in "${FILES[@]}"; do
        echo "  -> $file"
        curl -sSfL "$BASE_URL/$DEFAULT_PATH/$file" -o "internal/install/resources/brewfiles/$file"
    done
    
    echo "Updating Bluefin-specific Brewfiles..."
    echo "  -> full-desktop.Brewfile"
    curl -sSfL "$BASE_URL/$BLUEFIN_PATH/full-desktop.Brewfile" -o "internal/install/resources/brewfiles/full-desktop.Brewfile"
    echo "Update complete!"

# Show Go module info
mod-info: build-container
    @echo "Go module information:"
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go version
