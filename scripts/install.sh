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
DIZZ_CONFIG_DIR="$HOME/.config/dizz"
HOOKS_DIR="$DIZZ_CONFIG_DIR/hooks"

# Detect OS and architecture
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
    linux)  OS="linux"  ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}Installing Dizz...${NC}"

# Fetch latest release
echo "Fetching latest version..."
LATEST_RELEASE=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') || true

if [ -z "$LATEST_RELEASE" ]; then
    # Fallback: try tag-based lookup via redirect
    LATEST_RELEASE=$(curl -sIL "https://github.com/$REPO/releases/latest" \
        | grep -i '^location:' | sed -E 's/.*\/tag\/(v?[0-9]+\.[0-9]+\.[0-9]+).*/\1/' | head -1) || true
fi

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${RED}Failed to detect latest release${NC}"
    echo -e "  ${YELLOW}Check: https://github.com/$REPO/releases${NC}"
    echo -e "  You can also download the binary manually from the releases page."
    exit 1
fi

echo -e "${GREEN}Latest version: $LATEST_RELEASE${NC}"

# Binary name (matches GoReleaser output)
BINARY="dizz-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$BINARY"

echo "Downloading: $DOWNLOAD_URL"

mkdir -p "$INSTALL_DIR"

# Download to temp file, then atomic move
TEMP_FILE="$(mktemp)"
trap 'rm -f "$TEMP_FILE"' EXIT

if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
    echo -e "${RED}Download failed${NC}"
    echo "Check: https://github.com/$REPO/releases/tag/$LATEST_RELEASE"
    exit 1
fi

chmod +x "$TEMP_FILE"
mv "$TEMP_FILE" "$INSTALL_DIR/dizz"
trap - EXIT

echo -e "${GREEN}Installed to $INSTALL_DIR/dizz${NC}"

# ─────────────────────────────────────────────────────────────────
# macOS: remove Gatekeeper quarantine attribute
# Without this, macOS blocks the binary with an unverified developer
# warning when the user tries to run it.
# ─────────────────────────────────────────────────────────────────
if [ "$(uname -s)" = "Darwin" ]; then
    xattr -d com.apple.quarantine "$INSTALL_DIR/dizz" 2>/dev/null || true
    echo -e "  ${BLUE}macOS Gatekeeper:${NC} quarantine attribute removed"
fi

# ─────────────────────────────────────────────────────────────────
# Ensure INSTALL_DIR is in PATH (persistent, across restarts)
# Detects the user's preferred shell RC file and appends the export
# if it's not already present.
# ─────────────────────────────────────────────────────────────────
add_to_path() {
    local rc_file="$1"
    local line="export PATH=\"\$PATH:$INSTALL_DIR\""

    if [ ! -f "$rc_file" ]; then
        touch "$rc_file" 2>/dev/null || return 1
    fi

    if ! grep -qsF "$INSTALL_DIR" "$rc_file" 2>/dev/null; then
        echo "" >> "$rc_file"
        echo "# Added by dizz installer" >> "$rc_file"
        echo "$line" >> "$rc_file"
        echo -e "  ${BLUE}Added to PATH${NC} in $rc_file"
        return 0
    fi
    return 1
}

# Determine which RC file to use (order of preference)
PATH_UPDATED=false
if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
    add_to_path "$HOME/.zshrc" && PATH_UPDATED=true
fi
if [ "$PATH_UPDATED" = false ] && ( [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ] ); then
    add_to_path "$HOME/.bashrc" && PATH_UPDATED=true
fi
if [ "$PATH_UPDATED" = false ] && [ -f "$HOME/.bash_profile" ]; then
    add_to_path "$HOME/.bash_profile" && PATH_UPDATED=true
fi
if [ "$PATH_UPDATED" = false ] && [ -f "$HOME/.profile" ]; then
    add_to_path "$HOME/.profile" && PATH_UPDATED=true
fi

# Also add to PATH for the current session so `dizz version` below works
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    export PATH="$PATH:$INSTALL_DIR"
fi

# ─────────────────────────────────────────────────────────────────
# Install global dizz router hook
# This ensures dizz post-commit hooks fire on every repo that has
# .dizz/hooks/post-commit, even on the very first commit after clone.
# ─────────────────────────────────────────────────────────────────
echo -e "${BLUE}Installing global git hook router...${NC}"

mkdir -p "$HOOKS_DIR"

cat > "$HOOKS_DIR/post-commit" << 'HOOK'
#!/usr/bin/env sh

# dizz global router hook
# Delegates to .dizz/hooks/post-commit if it exists in the current repo.
# After first delegation, it sets local core.hooksPath so future commits
# bypass the router and use the repo's tracked hooks directly.

DIZZ_HOOKS=".dizz/hooks/post-commit"

if [ -f "$DIZZ_HOOKS" ] && [ -x "$DIZZ_HOOKS" ]; then
    git config core.hooksPath ".dizz/hooks" 2>/dev/null || true
    exec "$DIZZ_HOOKS"
fi
HOOK

chmod +x "$HOOKS_DIR/post-commit"

# Set global hooks path
git config --global core.hooksPath "$HOOKS_DIR" 2>/dev/null || true

echo -e "${GREEN}✓ Global hook router installed${NC}"
echo -e "  ${BLUE}Location:${NC} $HOOKS_DIR/post-commit"

# Verify
if command -v dizz >/dev/null 2>&1; then
    echo -e "${GREEN}Dizz installed successfully!${NC}"
    dizz version || true
else
    echo -e "${YELLOW}Installed to $INSTALL_DIR/dizz${NC}"
    echo -e "${YELLOW}Restart your terminal or run: export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
fi

echo ""
echo -e "${BLUE}Next:${NC} run 'dizz install-skill' to enable agent skill discovery"
