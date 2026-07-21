# Dizz Installation Script for Windows (PowerShell)
param(
    [switch]$Force,
    [string]$InstallDir = "$env:USERPROFILE\.dizz\bin"
)

# Repo
$Repo = "TheShiveshNetwork/dizz"
$DizzConfigDir = "$env:USERPROFILE\.config\dizz"
$HooksDir = "$DizzConfigDir\hooks"

function Write-Info($msg)  { Write-Host $msg -ForegroundColor Cyan }
function Write-Ok($msg)    { Write-Host $msg -ForegroundColor Green }
function Write-Warn($msg)  { Write-Host $msg -ForegroundColor Yellow }
function Write-Err($msg)   { Write-Host $msg -ForegroundColor Red }

function Write-Utf8File($path, $content) {
    # Write a file as UTF-8 without BOM, with Unix line endings
    $utf8 = [System.Text.UTF8Encoding]::new($false)
    [System.IO.File]::WriteAllText($path, $content, $utf8)
}

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
        Write-Warn "Check: https://github.com/$Repo/releases"
        exit 1
    }

    Write-Ok "Latest version: $Tag"

    # Binary info
    $BinaryName = "dizz-windows-$Arch.exe"
    $DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$BinaryName"

    Write-Info "Downloading $BinaryName"
    $TempFile = Join-Path $env:TEMP "dizz.exe"

    try {
        Invoke-WebRequest `
            -Uri $DownloadUrl `
            -OutFile $TempFile `
            -UseBasicParsing `
            -ErrorAction Stop
    } catch {
        Write-Err "Download failed: $($_.Exception.Message)"
        Write-Warn "Check: https://github.com/$Repo/releases/tag/$Tag"
        exit 1
    }

    # Create install dir
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Ok "Created install directory: $InstallDir"
    }

    # Install binary
    $Dest = Join-Path $InstallDir "dizz.exe"
    Move-Item -Path $TempFile -Destination $Dest -Force
    Write-Ok "Installed dizz to $Dest"

    # Add to PATH (user-level, persists across restarts)
    $UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable(
            "PATH",
            "$UserPath;$InstallDir",
            "User"
        )
        # Also update current session PATH so verification works
        $env:Path = "$env:Path;$InstallDir"
        Write-Ok "Added $InstallDir to user PATH (persistent)"
    } else {
        Write-Ok "Install directory already in PATH"
    }

    # ──────────────────────────────────────────────────────────────
    # Install global dizz router hook
    # ──────────────────────────────────────────────────────────────
    Write-Info "Installing global git hook router..."

    if (-not (Test-Path $HooksDir)) {
        New-Item -ItemType Directory -Path $HooksDir -Force | Out-Null
    }

    $HookFile = Join-Path $HooksDir "post-commit"
    $HookContent = @'
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
'@

    Write-Utf8File $HookFile $HookContent
    Write-Ok "Global hook router written to $HookFile"

    # Set global hooks path via git config
    git config --global core.hooksPath $HooksDir 2>$null
    Write-Ok "Configured global core.hooksPath"

    # Verify install
    Write-Info "Verifying installation..."
    & $Dest version | Out-Host

    Write-Ok "Dizz installed successfully :)"
    Write-Info "Run: dizz --help"
    Write-Info "Next: run 'dizz install-skill' to enable agent skill discovery"

} catch {
    Write-Err "Installation failed: $($_.Exception.Message)"
    exit 1
}
