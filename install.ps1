# Bluefin CLI — Windows one-liner installer
# Installs the 'plus' (full-featured) edition to %LOCALAPPDATA%\Programs\bluefin-cli\
# and adds it to the current user's PATH.
#
# Usage (run in PowerShell as normal user, no elevation required):
#   irm https://raw.githubusercontent.com/hanthor/bluefin-cli/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$repo    = 'hanthor/bluefin-cli'
$binName = 'bluefin-cli.exe'
$installDir = "$env:LOCALAPPDATA\Programs\bluefin-cli"

# Detect architecture
$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
} else {
    Write-Error 'Bluefin CLI requires a 64-bit Windows system.'
    exit 1
}

Write-Host "Fetching latest release..." -ForegroundColor Cyan
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name

# Prefer the 'plus' (extra) build; fall back to standard
$assetPattern = "bluefin-cli-plus_*_windows_${arch}.zip"
$asset = $release.assets | Where-Object { $_.name -like $assetPattern } | Select-Object -First 1
if (-not $asset) {
    $assetPattern = "bluefin-cli_*_windows_${arch}.zip"
    $asset = $release.assets | Where-Object { $_.name -like $assetPattern } | Select-Object -First 1
}
if (-not $asset) {
    Write-Error "No Windows $arch asset found in release $version."
    exit 1
}

Write-Host "Downloading $($asset.name) ($version)..." -ForegroundColor Cyan
$tmpZip = [System.IO.Path]::GetTempFileName() + '.zip'
$tmpDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), [System.IO.Path]::GetRandomFileName())

try {
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tmpZip -UseBasicParsing
    Expand-Archive -Path $tmpZip -DestinationPath $tmpDir -Force

    $null = New-Item -ItemType Directory -Path $installDir -Force

    # The archive contains bluefin-cli-plus.exe or bluefin-cli.exe
    $exeInZip = Get-ChildItem $tmpDir -Recurse -Filter '*.exe' | Select-Object -First 1
    if (-not $exeInZip) {
        Write-Error "No .exe found in downloaded archive."
        exit 1
    }
    Copy-Item $exeInZip.FullName "$installDir\$binName" -Force

    Write-Host "Installed to $installDir\$binName" -ForegroundColor Green
} finally {
    Remove-Item $tmpZip -ErrorAction SilentlyContinue
    Remove-Item $tmpDir -Recurse -ErrorAction SilentlyContinue
}

# Add to PATH (user scope only, no admin required)
$userPath = [System.Environment]::GetEnvironmentVariable('PATH', 'User')
if ($userPath -notlike "*$installDir*") {
    [System.Environment]::SetEnvironmentVariable('PATH', "$installDir;$userPath", 'User')
    $env:PATH = "$installDir;$env:PATH"
    Write-Host "Added $installDir to your PATH." -ForegroundColor Green
}

Write-Host ""
Write-Host "Bluefin CLI $version installed successfully!" -ForegroundColor Green
Write-Host "Run 'bluefin-cli' to get started. (You may need to restart your terminal.)" -ForegroundColor Cyan
