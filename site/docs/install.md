# Installation Guide

Dizz supports multiple installation methods to suit different platforms and preferences.

## Quick Install (Recommended)

### Unix/macOS

```bash
curl -fsSL https://your-domain.com/install.sh | sh
```

### Windows (PowerShell)

```powershell
iwr -useb https://your-domain.com/install.ps1 | iex
```

## Manual Installation

### Download from GitHub Releases

1. Visit the [releases page](https://github.com/yourusername/dizz/releases)
2. Download the appropriate binary for your platform:
   - `dizz-{version}-darwin-amd64` for macOS Intel
   - `dizz-{version}-darwin-arm64` for macOS Apple Silicon
   - `dizz-{version}-linux-amd64` for Linux x86_64
   - `dizz-{version}-linux-arm64` for Linux ARM64
   - `dizz-{version}-windows-amd64.zip` for Windows x86_64
3. Extract the archive and move the binary to your PATH

### Using Package Managers

#### Homebrew (macOS)

```bash
brew tap yourusername/dizz
brew install dizz
```

#### Scoop (Windows)

```bash
scoop bucket add dizz https://github.com/yourusername/dizz-scoop
scoop install dizz
```

#### Debian/Ubuntu

```bash
wget -qO- https://your-domain.com/apt/gpg.key | sudo apt-key add -
echo "deb https://your-domain.com/apt/ stable main" | sudo tee /etc/apt/sources.list.d/dizz.list
sudo apt update
sudo apt install dizz
```

#### Arch Linux

```bash
yay -S dizz-bin
```

### Building from Source

If you prefer to build from source:

```bash
git clone https://github.com/yourusername/dizz.git
cd dizz
make build
sudo make install
```

## Verification

After installation, verify that Dizz is working:

```bash
dizz version
```

This should display the installed version.

## Upgrading

### Using Install Scripts (Recommended)

Simply run the installation command again to upgrade to the latest version.

### Using Package Managers

```bash
# Homebrew
brew upgrade dizz

# Scoop
scoop update dizz

# APT
sudo apt update && sudo apt upgrade dizz
```

### Manual Upgrade

Download the latest release and replace the existing binary.

## Uninstallation

### Manual Removal

```bash
# Remove the binary
rm -f /usr/local/bin/dizz

# Or if installed to ~/.local/bin
rm -f ~/.local/bin/dizz
```

### Package Manager

```bash
# Homebrew
brew uninstall dizz

# Scoop
scoop uninstall dizz

# APT
sudo apt remove dizz
```

## Configuration

After installation, Dizz creates a configuration directory at:

- Unix/macOS: `~/.config/dizz/`
- Windows: `%APPDATA%\dizz\`

You can customize Dizz behavior by editing the configuration file.

## Troubleshooting

### Command Not Found

If you get "command not found" after installation:

1. Ensure the installation directory is in your PATH
2. Restart your terminal
3. Check the installation logs for any errors

### Permission Denied

If you get permission errors:

1. Make sure the binary is executable: `chmod +x dizz`
2. On Unix systems, you may need sudo to install to system directories

### Network Issues

If the install scripts fail due to network issues:

1. Check your internet connection
2. Try downloading manually from GitHub releases
3. Use a VPN if you're behind a corporate firewall

For additional help, please open an issue on GitHub.