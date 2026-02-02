# Dizz Installation Script for Windows (PowerShell)
param(
    [switch]$Force,
    [string]$InstallDir = "$env:USERPROFILE\.dizz\bin"
)

# Repo
$Repo = "TheShiveshNetwork/dizz"

function Write-Info($msg)  { Write-Host $msg -ForegroundColor Cyan }
function Write-Ok($msg)    { Write-Host $msg -ForegroundColor Green }
function Write-Warn($msg)  { Write-Host $msg -ForegroundColor Yellow }
function Write-Err($msg)   { Write-Host $msg -ForegroundColor Red }

try {
    Write-Info "Installing Dizz CLI..."

    # Detect architecture
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { $Arch = "amd64" }
        "ARM64" { $Arch = "arm64" }
        default {
            Write-Err "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
            exit 1
        }
    }

    $OS = "windows"
    Write-Info "Detected platform: $OS-$Arch"

    # Fetch latest release
    Write-Info "Fetching latest version..."
    try {
        $Release = Invoke-RestMethod `
            -Uri "https://api.github.com/repos/$Repo/releases/latest" `
            -Headers @{ "User-Agent" = "dizz-installer" }
        $Tag = $Release.tag_name
    } catch {
        Write-Err "Failed to fetch latest release"
        exit 1
    }

    Write-Ok "Latest version: $Tag"

    # Binary info
    $BinaryName = "dizz-windows-$Arch.exe"
    $DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$BinaryName"

    Write-Info "Downloading $BinaryName"
    $TempFile = Join-Path $env:TEMP "dizz.exe"

    Invoke-WebRequest `
        -Uri $DownloadUrl `
        -OutFile $TempFile `
        -UseBasicParsing `
        -ErrorAction Stop

    # Create install dir
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Ok "Created install directory: $InstallDir"
    }

    # Install binary
    $Dest = Join-Path $InstallDir "dizz.exe"
    Move-Item -Path $TempFile -Destination $Dest -Force
    Write-Ok "Installed dizz to $Dest"

    # Add to PATH (user-level)
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable(
            "PATH",
            "$UserPath;$InstallDir",
            "User"
        )
        Write-Warn "Added to PATH (restart terminal required)"
    } else {
        Write-Ok "Install directory already in PATH"
    }

    # Verify install
    Write-Info "Verifying installation..."
    & $Dest version | Out-Host

    Write-Ok "Dizz installed successfully 🎉"
    Write-Info "Run: dizz --help"

} catch {
    Write-Err "Installation failed: $($_.Exception.Message)"
    exit 1
}

