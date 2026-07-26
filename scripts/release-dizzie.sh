#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TAG="$1"

if [ -z "$TAG" ]; then
    echo -e "${RED}Usage: $0 <tag>${NC}"
    echo -e "  Example: ${YELLOW}$0 dizzie-v1.0.0${NC}"
    exit 1
fi

if ! echo "$TAG" | grep -qE '^dizzie-v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$'; then
    echo -e "${RED}Invalid tag format: $TAG${NC}"
    echo -e "  Expected: ${YELLOW}dizzie-v<major>.<minor>.<patch>[-prerelease]${NC}"
    echo -e "  Example:  ${YELLOW}dizzie-v1.0.0${NC}"
    exit 1
fi

echo -e "${BLUE}Releasing dizzie ${TAG}...${NC}"

echo "Generating embed files..."
go generate ./internal/defaults/

echo "Building dizzie..."
cd tui && go build -trimpath -ldflags "-s -w -X main.version=${TAG#dizzie-}" -o /tmp/dizzie . && cd ..

echo "Running tests..."
cd tui && go test ./... && cd ..

echo "Creating git tag..."
git tag -a "$TAG" -m "Release $TAG"

echo "Pushing tag..."
git push origin "$TAG"

echo -e "${GREEN}Tag $TAG pushed.${NC}"
echo -e "${BLUE}GitHub Actions will build and publish the release.${NC}"
echo -e "  Track: ${YELLOW}https://github.com/TheShiveshNetwork/dizz/actions${NC}"
