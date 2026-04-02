# xParser CLI installer for Windows
# Usage: irm https://dllf.intsig.net/download/2026/Solution/xparser/install.ps1 | iex
#
# Environment variables:
#   XPARSER_VERSION   - version to install (default: "latest")
#   XPARSER_BASE_URL  - override download base URL
#   INSTALL_DIR       - install directory (default: $HOME\.xparser\bin)

$ErrorActionPreference = "Stop"

# Ensure TLS 1.2+ for older PowerShell versions
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13

$Version = if ($env:XPARSER_VERSION) { $env:XPARSER_VERSION } else { "latest" }
$BaseURL = if ($env:XPARSER_BASE_URL) { $env:XPARSER_BASE_URL } else { "https://dllf.intsig.net/download/2026/Solution/xparser" }
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$HOME\.xparser\bin" }

$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "32-bit Windows is not supported"; exit 1
}

$Binary = "xparser-windows-${Arch}.exe"
$URL = "${BaseURL}/${Version}/${Binary}"
$DestPath = Join-Path $InstallDir "xparser.exe"

Write-Host "Downloading xparser ${Version} for windows/${Arch}..."
Write-Host "  ${URL}"

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

try {
    Invoke-WebRequest -Uri $URL -OutFile $DestPath -UseBasicParsing

    if (Get-Command Unblock-File -ErrorAction SilentlyContinue) {
        Unblock-File -Path $DestPath -ErrorAction SilentlyContinue
    }
} catch {
    Write-Error "Download failed: $_"
    exit 1
}

if (-not (Test-Path $DestPath) -or (Get-Item $DestPath).Length -eq 0) {
    Write-Error "Download failed or file is empty"
    exit 1
}

# Add to user-level PATH (no admin required)
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    $NewPath = if ([string]::IsNullOrWhiteSpace($UserPath)) { $InstallDir } else { "$UserPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "Added $InstallDir to user PATH"
}

Write-Host ""
Write-Host "Installed successfully!"
Write-Host "=========================================================="
& $DestPath version
Write-Host "=========================================================="
Write-Host ""
Write-Host "Executable: $DestPath"
