#!/bin/bash
set -e

REPO="WHQ25/rawgenai"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin|linux) ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Failed to get latest version"
    exit 1
fi

# Build filename
if [ "$OS" = "windows" ]; then
    FILENAME="rawgenai_${VERSION}_${OS}_${ARCH}.zip"
else
    FILENAME="rawgenai_${VERSION}_${OS}_${ARCH}.tar.gz"
fi

URL="https://github.com/$REPO/releases/download/v${VERSION}/${FILENAME}"

echo "Installing rawgenai v${VERSION} (${OS}/${ARCH})..."

# Download and install
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

cd "$TMPDIR"
curl -fsSL -o "$FILENAME" "$URL"

if [ "$OS" = "windows" ]; then
    unzip -q "$FILENAME"
else
    tar -xzf "$FILENAME"
fi

# Install binary
if [ -w "$INSTALL_DIR" ]; then
    mv rawgenai "$INSTALL_DIR/"
else
    sudo mv rawgenai "$INSTALL_DIR/"
fi

echo "Installed rawgenai to $INSTALL_DIR/rawgenai"
rawgenai version
