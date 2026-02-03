#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Root of repo
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Paths (absolute, safe)
SITE_DIR="$ROOT_DIR/site"
PUBLIC_DIR="$ROOT_DIR/site/public"
SERVER_DIR="$ROOT_DIR/site/server"
SCRIPTS_DIR="$ROOT_DIR/site/scripts"

echo -e "${BLUE}Building Dizz WebAssembly project...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
  echo -e "${RED}Error: Go is not installed${NC}"
  exit 1
fi

echo -e "${GREEN}Go version: $(go version)${NC}"

# Ensure directories exist
mkdir -p "$PUBLIC_DIR/docs"

# -----------------------
# Build WASM
# -----------------------
echo -e "${BLUE}Building WebAssembly module...${NC}"

cd "$SITE_DIR"
GOOS=js GOARCH=wasm go build -tags wasm -o "$PUBLIC_DIR/dizz.wasm" wasm.go
cd - > /dev/null

echo -e "${GREEN}✓ WebAssembly module built${NC}"

# -----------------------
# Copy wasm_exec.js
# -----------------------
echo -e "${BLUE}Copying wasm_exec.js...${NC}"

GOROOT="$(go env GOROOT)"

if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/misc/wasm/wasm_exec.js" "$PUBLIC_DIR/"
elif [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/lib/wasm/wasm_exec.js" "$PUBLIC_DIR/"
else
  echo -e "${RED}✗ wasm_exec.js not found in Go installation${NC}"
  exit 1
fi

echo -e "${GREEN}✓ wasm_exec.js copied${NC}"

# -----------------------
# Generate HTML from template
# -----------------------
echo -e "${BLUE}Generating HTML from template...${NC}"

cd "$SITE_DIR"
go run generate.go
cd - > /dev/null

echo -e "${GREEN}✓ HTML generated${NC}"

# -----------------------
# Build web server
# -----------------------
echo -e "${BLUE}Building HTTP server...${NC}"

echo -e "${GREEN}✓ HTTP server built${NC}"

# -----------------------
# Copy assets
# -----------------------
ASSETS_DIR="$SITE_DIR/assets"
if [ -d "$ASSETS_DIR" ]; then
  cp -r "$ASSETS_DIR" "$PUBLIC_DIR/"
  echo -e "${GREEN}✓ Assets copied to public directory${NC}"
fi

# -----------------------
# Copy install scripts
# -----------------------
if [ -d "$SCRIPTS_DIR" ]; then
  chmod +x "$SCRIPTS_DIR"/*.sh || true
  cp "$SCRIPTS_DIR"/install.sh "$PUBLIC_DIR/"
  cp "$SCRIPTS_DIR"/install.ps1 "$PUBLIC_DIR/"
  echo -e "${GREEN}✓ Install scripts copied to public directory${NC}"
fi

# -----------------------
# Summary
# -----------------------
echo
echo -e "${BLUE}Build Summary${NC}"
echo -e "${GREEN}✓ WASM: $PUBLIC_DIR/dizz.wasm${NC}"
echo -e "${GREEN}✓ wasm_exec.js${NC}"
echo
echo -e "${BLUE}Run locally:${NC}"
echo "go run site/server/server.go"
echo "→ http://localhost:8080"
echo
echo -e "${GREEN}Build completed successfully!${NC}"

