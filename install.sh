#!/bin/bash

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
fi

# Set version
VERSION="0.1.0"
BINARY_NAME="zap-${OS}-${ARCH}"

if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

# Download URL
DOWNLOAD_URL="https://github.com/marpit19/zap-pm/releases/download/v${VERSION}/${BINARY_NAME}"

# Installation directory
INSTALL_DIR="/usr/local/bin"
mkdir -p "$INSTALL_DIR"

# Download and install
echo "Downloading Zap Package Manager..."
curl -L "$DOWNLOAD_URL" -o "$INSTALL_DIR/zap"
chmod +x "$INSTALL_DIR/zap"

echo "Zap Package Manager installed successfully!"
echo "Run 'zap --version' to verify installation"
