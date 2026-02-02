# Installation Guide

## Quick Install

### Linux/macOS
```bash
curl -sSL https://dizz.shitworks.co/install.sh | bash
```

### Windows (PowerShell)
```powershell
powershell -c "irm https://dizz.shitworks.co/install.ps1 | iex"
```

## Manual Installation

1. Visit [releases](https://github.com/TheShiveshNetwork/dizz/releases)
2. Download binary for your platform:
   - `dizz-{version}-darwin-amd64` (macOS Intel)
   - `dizz-{version}-darwin-arm64` (macOS Apple Silicon)  
   - `dizz-{version}-linux-amd64` (Linux x86_64)
   - `dizz-{version}-linux-arm64` (Linux ARM64)
   - `dizz-{version}-windows-amd64.zip` (Windows)
3. Move binary to your PATH

## Build from Source

```bash
git clone https://github.com/TheShiveshNetwork/dizz.git
cd dizz
go build -o dizz .
```

## Verify Installation

```bash
dizz version
```

## Upgrade

Run the install command again to upgrade to the latest version.

## Uninstall

Remove the binary from your system location.