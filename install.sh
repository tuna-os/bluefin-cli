#!/bin/sh
# Bluefin CLI installer for Linux and macOS.
#
#   curl -fsSL https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.sh | sh
#
# Options (environment variables):
#   BLUEFIN_CLI_VERSION  - install a specific version (default: latest release)
#   BLUEFIN_CLI_PLUS     - set to 1 to install the plus binary (extra features)
#   BLUEFIN_CLI_BIN_DIR  - install directory (default: ~/.local/bin, or
#                          /usr/local/bin when run as root)
set -eu

REPO="tuna-os/bluefin-cli"
BINARY="bluefin-cli"
[ "${BLUEFIN_CLI_PLUS:-0}" = "1" ] && BINARY="bluefin-cli-plus"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux|darwin) ;;
  *) echo "error: unsupported OS: $os (use install.ps1 on Windows)" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "error: unsupported architecture: $arch" >&2; exit 1 ;;
esac

if [ -n "${BLUEFIN_CLI_VERSION:-}" ]; then
  tag="v${BLUEFIN_CLI_VERSION#v}"
else
  tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$tag" ] || { echo "error: could not determine latest release" >&2; exit 1; }
fi
version=${tag#v}

if [ -n "${BLUEFIN_CLI_BIN_DIR:-}" ]; then
  bin_dir="$BLUEFIN_CLI_BIN_DIR"
elif [ "$(id -u)" = "0" ]; then
  bin_dir="/usr/local/bin"
else
  bin_dir="$HOME/.local/bin"
fi

archive="${BINARY}_${version}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$tag/$archive"
checksum_url="https://github.com/$REPO/releases/download/$tag/checksums.txt"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $BINARY $tag ($os/$arch)..."
curl -fsSL "$url" -o "$tmp/bluefin-cli.tar.gz"

if curl -fsSL "$checksum_url" -o "$tmp/checksums.txt" 2>/dev/null; then
  expected_sha=$(grep "  ${archive}$" "$tmp/checksums.txt" | awk '{print $1}')
  if [ -n "$expected_sha" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual_sha=$(sha256sum "$tmp/bluefin-cli.tar.gz" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
      actual_sha=$(shasum -a 256 "$tmp/bluefin-cli.tar.gz" | awk '{print $1}')
    else
      actual_sha=""
    fi
    if [ -n "$actual_sha" ] && [ "$actual_sha" != "$expected_sha" ]; then
      echo "error: checksum verification failed for $archive" >&2
      exit 1
    fi
  fi
fi

tar -xzf "$tmp/bluefin-cli.tar.gz" -C "$tmp"

[ -f "$tmp/$BINARY" ] || { echo "error: $BINARY not found in archive" >&2; exit 1; }

mkdir -p "$bin_dir"
install -m 0755 "$tmp/$BINARY" "$bin_dir/$BINARY"

echo "Installed $BINARY $tag to $bin_dir/$BINARY"
case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) echo "note: $bin_dir is not on your PATH — add it to your shell config" ;;
esac
echo "Run '$BINARY' to get started."
