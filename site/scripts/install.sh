#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPO="TheShiveshNetwork/dizz"
INSTALL_DIR="$HOME/.local/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}Installing Dizz...${NC}"

# Fetch latest release
echo "Fetching latest version..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest \
  | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${RED}Failed to detect latest release${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: $LATEST_RELEASE${NC}"

# Binary name (matches GoReleaser output)
BINARY="dizz-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$BINARY"

echo "Downloading: $DOWNLOAD_URL"

mkdir -p "$INSTALL_DIR"
TEMP_FILE="$(mktemp)"

# Download binary
if ! curl -fL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
    echo -e "${RED}Download failed${NC}"
    echo "Check: https://github.com/$REPO/releases/tag/$LATEST_RELEASE"
    exit 1
fi

# Install
chmod +x "$TEMP_FILE"
mv "$TEMP_FILE" "$INSTALL_DIR/dizz"

echo -e "${GREEN}Installed to $INSTALL_DIR/dizz${NC}"

# PATH check
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}$INSTALL_DIR is not in PATH${NC}"
    echo "Add this to your shell config:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

# Verify
if command -v dizz >/dev/null; then
    echo -e "${GREEN}Dizz installed successfully!${NC}"
    dizz version || true
else
    echo -e "${YELLOW}Installed, but dizz not found in PATH yet${NC}"
fi

