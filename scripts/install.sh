#!/bin/sh
set -e

# PingForge automated installer script
VERSION="1.0.0"
REPO="sujanv/PingForge"
INSTALL_DIR="/usr/local/bin"

echo "⚡ Installing PingForge v${VERSION}..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Detected OS: $OS, Architecture: $ARCH"

if command -v go >/dev/null 2>&1; then
  echo "Building directly from source with Go..."
  go install github.com/sujanv/pingforge/cmd/pingforge@latest
  echo "✔ Successfully installed to $(go env GOPATH)/bin/pingforge"
  exit 0
fi

echo "Please ensure Go is installed to compile PingForge."
exit 1
