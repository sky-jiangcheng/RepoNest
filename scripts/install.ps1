# RepoNest install script for Windows
# Run in PowerShell: iwr -useb https://raw.githubusercontent.com/sky-jiangcheng/repo-nest/master/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$InstallDir = "$env:LOCALAPPDATA\RepoNest"
$BinaryName = "reponest.exe"
$Repo = "sky-jiangcheng/repo-nest"
$Target = "windows-amd64"

Write-Host "Downloading RepoNest for Windows..."
$DownloadUrl = "https://github.com/$Repo/releases/latest/download/reponest-$Target.exe"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Invoke-WebRequest -Uri $DownloadUrl -OutFile "$InstallDir\$BinaryName"

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
