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
    go build -ldflags "-s -w -X github.com/tuna-os/bluefin-cli/cmd.version=$(git describe --tags --always --dirty)" -o bluefin-cli

# Build the plus binary (Everything)
build-plus:
    @echo "Building plus binary..."
    go build -tags extra -ldflags "-s -w -X github.com/tuna-os/bluefin-cli/cmd.version=$(git describe --tags --always --dirty)" -o bluefin-cli-plus

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

# Open shell in container with the shell experience already enabled (manual testing)
shell-with-experience: build-container build
    #!/usr/bin/env bash
    echo "Setting up container with the shell experience enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli shell enable bash && \
                 ./bluefin-cli shell enable zsh && \
                 ./bluefin-cli shell enable fish && \
                 ./bluefin-cli motd toggle bash on && \
                 echo "" && \
                 echo "=== Shell experience has been enabled ===" && \
                 echo "Binary: ./bluefin-cli" && \
                 echo "Configs: ~/.bashrc, ~/.zshrc, ~/.config/fish/config.fish" && \
                 echo "Init script: bluefin-cli init bash" && \
                 echo "" && \
                 echo "Try: ./bluefin-cli status" && \
                 echo "     cat ~/.bashrc" && \
                 echo "     ./bluefin-cli init bash" && \
                 echo "" && \
                 bash'

# Open bash with the shell experience enabled and sourced
bash: build-container build
    #!/usr/bin/env bash
    echo "Launching bash with the shell experience enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        -e PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli shell enable bash > /dev/null 2>&1 && \
                 echo "✓ Shell experience enabled - Tools: starship=$(command -v starship), eza=$(command -v eza)" && \
                 exec bash'

# Open zsh with the shell experience enabled and sourced
zsh: build-container build
    #!/usr/bin/env bash
    echo "Launching zsh with the shell experience enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        -e SHELL=/bin/zsh \
        -e PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 rm -f ~/.zshrc && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli shell enable zsh > /dev/null 2>&1 && \
                 echo "✓ Shell experience enabled - Tools: starship=$(command -v starship), eza=$(command -v eza)" && \
                 ZDOTDIR=/root exec zsh'

# Open fish with the shell experience enabled and sourced
fish: build-container build
    #!/usr/bin/env bash
    echo "Launching fish with the shell experience enabled..."
    podman run --rm -it \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        -e PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli shell enable fish > /dev/null 2>&1 && \
                 exec fish'

# Inspect what the shell experience writes into a shell's config
inspect-shell: build-container build
    @echo "Inspecting shell experience files in container..."
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        -e HOME=/root \
        bluefin-cli-dev \
        bash -c 'mkdir -p ~/.config/fish && \
                 touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish && \
                 ./bluefin-cli shell enable bash && \
                 ./bluefin-cli shell enable zsh && \
                 ./bluefin-cli shell enable fish && \
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
                 echo "=== Generated init scripts ===" && \
                 echo "" && \
                 echo "--- init bash (first 50 lines) ---" && \
                 ./bluefin-cli init bash | head -50 && \
                 echo "" && \
                 echo "--- init fish (first 30 lines) ---" && \
                 ./bluefin-cli init fish | head -30'

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

# Update embedded resources (Brewfiles, wallpaper casks) from upstream.
# Also rewrites internal/install/resources/PROVENANCE.json, which records the
# upstream commit and a digest for every embedded file; CI holds the tree to it.
update-resources:
    @python3 scripts/update-resources.py

# Regenerate the embedded Homebrew-to-Winget package mapping on Windows.
update-winget-mapping:
    pwsh -NoProfile -File scripts/winget/update-mapping.ps1

# Summarize coverage of the embedded Winget package mapping on Windows.
summarize-winget-mapping:
    pwsh -NoProfile -File scripts/winget/summarize-mapping.ps1

# Show Go module info
mod-info: build-container
    @echo "Go module information:"
    podman run --rm \
        -v "$(pwd):/workspace:Z" \
        -w /workspace \
        bluefin-cli-dev \
        go version

# Run the end-to-end TUI smoke test in tmux (drives real keys, asserts screens)
tui-smoke:
    go build -tags extra -o tmp/bfc-smoke .
    ./scripts/tui-smoke.sh tmp/bfc-smoke
    rm -f tmp/bfc-smoke

# Full local verification gauntlet — what CI checks, without the queue.
# GOTMPDIR keeps test binaries off /tmp, which is noexec on some setups.
verify:
    mkdir -p tmp/gotmp
    go build ./...
    go build -tags extra ./...
    gofmt -l . | grep -v '^tmp/' | (! grep .) || (echo "gofmt needed" && exit 1)
    GOTMPDIR=$PWD/tmp/gotmp go vet -tags extra ./...
    command -v golangci-lint >/dev/null && GOTMPDIR=$PWD/tmp/gotmp golangci-lint run --build-tags extra ./... || echo "note: golangci-lint not installed, skipping lint"
    go build -tags extra -o bluefin-cli .
    GOTMPDIR=$PWD/tmp/gotmp go test -tags extra ./...
    rm -f bluefin-cli
    just tui-smoke
    go build -tags extra -o tmp/bfc-state .
    ./scripts/tui-state.sh tmp/bfc-state
    rm -f tmp/bfc-state

# Heavy verification on a build host (default himachal), tmux suites local
# against the remotely built binary. Keeps the low-spec VPS responsive.
verify-remote host="himachal":
    rsync -az --delete --exclude tmp --exclude .git ./ {{host}}:dev/bluefin-cli-ci/
    ssh {{host}} 'export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH; set -e; cd dev/bluefin-cli-ci; go build ./...; go build -tags extra ./...; gofmt -l cmd internal test | (! grep .); go vet -tags extra ./...; command -v golangci-lint >/dev/null && golangci-lint run --build-tags extra ./... || true; go build -tags extra -o bluefin-cli .; go test -tags extra ./...; rm -f bluefin-cli; go build -tags extra -o /tmp/bfc-remote .'
    mkdir -p tmp
    scp -q {{host}}:/tmp/bfc-remote tmp/bfc-remote
    ./scripts/tui-smoke.sh tmp/bfc-remote
    ./scripts/tui-state.sh tmp/bfc-remote
    rm -f tmp/bfc-remote
