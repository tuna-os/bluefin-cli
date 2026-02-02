# Containerfile for bluefin-cli integration testing
# Using Fedora as base since bluefin-cli is designed for Fedora-based systems

# Builder stage - install Go and dependencies first (cached)
FROM registry.fedoraproject.org/fedora:44 AS builder

# Install Go and build dependencies (this layer is cached unless Fedora updates)
RUN dnf install -y golang git && dnf clean all

# Set GOTOOLCHAIN to auto to allow downloading newer Go if needed
# Enable checksum database for toolchain download
ENV GOTOOLCHAIN=auto
ENV GOSUMDB=sum.golang.org

WORKDIR /app

# Copy go.mod and go.sum first for better caching of dependencies
COPY go.mod go.sum ./
RUN go mod download

# Now copy source and build (only this layer rebuilds when code changes)
COPY . .
RUN go build -o bluefin-cli

# Test stage - set up base environment first (cached)
FROM registry.fedoraproject.org/fedora:44

# Install runtime dependencies for testing (cached layer)
RUN dnf install -y bash zsh fish curl git grep findutils coreutils && dnf clean all

# Create test user with home directory (cached layer)
RUN useradd -m -s /bin/bash testuser && \
    mkdir -p /home/testuser/.config/fish

# Set up test environment as testuser (cached layer)
USER testuser
WORKDIR /home/testuser
RUN touch ~/.bashrc ~/.zshrc ~/.config/fish/config.fish

# Switch back to root to copy files
USER root

# Copy built binary (only rebuilds when binary changes)
COPY --from=builder /app/bluefin-cli /usr/local/bin/bluefin-cli

# Copy test script (only rebuilds when test script changes)
COPY test-container.sh /usr/local/bin/test-container.sh
RUN chmod +x /usr/local/bin/test-container.sh

# Switch back to testuser for running tests
USER testuser
WORKDIR /home/testuser

# Run comprehensive tests
CMD ["/usr/local/bin/test-container.sh"]
