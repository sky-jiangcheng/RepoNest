# RepoNest install script for Windows
# Installs both the desktop app (reponest.exe) and the MCP server (reponest-mcp.exe).
# Run in PowerShell: iwr -useb https://raw.githubusercontent.com/sky-jiangcheng/RepoNest/master/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$InstallDir = "$env:LOCALAPPDATA\RepoNest"
$BinaryName = "reponest.exe"
$McpBinaryName = "reponest-mcp.exe"
$Repo = "sky-jiangcheng/RepoNest"
$Target = "windows-amd64"

# ReleaseBase is the base URL for release assets. Overridable so CI can point
# it at a local fixture server and exercise the real download/extract/install
# path without a published release (see .github/workflows/install-smoke.yml).
$ReleaseBase = if ($env:REPO_NEST_RELEASE_BASE) { $env:REPO_NEST_RELEASE_BASE } else { "https://github.com/$Repo/releases/latest/download" }
$DesktopUrl = "$ReleaseBase/reponest-$Target.zip"
$McpUrl = "$ReleaseBase/reponest-mcp-$Target.zip"

Write-Host "Downloading RepoNest for Windows..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# The release workflow publishes Windows as a .zip (Compress-Archive of
# build/bin/reponest.exe), not a bare .exe. Download the archive and expand it
# rather than pointing at a .exe URL that 404s.
$DesktopZip = "$env:TEMP\reponest-$Target.zip"
try {
    Invoke-WebRequest -Uri $DesktopUrl -OutFile $DesktopZip
    Expand-Archive -Path $DesktopZip -DestinationPath $InstallDir -Force
    if (-not (Test-Path "$InstallDir\$BinaryName")) {
        throw "archive did not contain $BinaryName"
    }
} finally {
    if (Test-Path $DesktopZip) { Remove-Item $DesktopZip -Force }
}

# MCP server: separate release asset, so AI clients can run it without the desktop
# app. Non-fatal: older releases may not have the asset yet.
Write-Host "Downloading RepoNest MCP server for Windows..."
$McpZip = "$env:TEMP\reponest-mcp-$Target.zip"
try {
    Invoke-WebRequest -Uri $McpUrl -OutFile $McpZip
    Expand-Archive -Path $McpZip -DestinationPath $InstallDir -Force
    if (Test-Path "$InstallDir\$McpBinaryName") {
        Write-Host "RepoNest MCP server installed to $InstallDir\$McpBinaryName"
    } else {
        throw "archive did not contain $McpBinaryName"
    }
} catch {
    Write-Warning "Could not install reponest-mcp. The desktop app is installed and working."
    Write-Warning "Install the MCP server manually from $McpUrl"
} finally {
    if (Test-Path $McpZip) { Remove-Item $McpZip -Force }
}

# Add to PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to PATH"
}

Write-Host ""
Write-Host "RepoNest installed to $InstallDir"
Write-Host "Run 'reponest' in a new terminal to start!"
Write-Host ""
Write-Host "You can also create a desktop shortcut to: $InstallDir\$BinaryName"
