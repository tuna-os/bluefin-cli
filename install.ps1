# Bluefin CLI installer for Windows (PowerShell 5.1+ / pwsh).
#
#   irm https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.ps1 | iex
#
# Options (environment variables):
#   BLUEFIN_CLI_VERSION - install a specific version (default: latest release)
#   BLUEFIN_CLI_PLUS    - set to 1 to install the plus binary (extra features)
#   BLUEFIN_CLI_BIN_DIR - install directory (default: %LOCALAPPDATA%\Programs\BluefinCLI)
$ErrorActionPreference = "Stop"

$repo = "tuna-os/bluefin-cli"
$binary = if ($env:BLUEFIN_CLI_PLUS -eq "1") { "bluefin-cli-plus" } else { "bluefin-cli" }

$arch = switch ((Get-CimInstance Win32_Processor).Architecture) {
    12 { "arm64" }
    default { "amd64" }
}

if ($env:BLUEFIN_CLI_VERSION) {
    $tag = "v" + $env:BLUEFIN_CLI_VERSION.TrimStart("v")
} else {
    $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
    $tag = $release.tag_name
}
$version = $tag.TrimStart("v")

$binDir = if ($env:BLUEFIN_CLI_BIN_DIR) { $env:BLUEFIN_CLI_BIN_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\BluefinCLI" }
$archive = "${binary}_${version}_windows_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$tag/$archive"
$checksumUrl = "https://github.com/$repo/releases/download/$tag/checksums.txt"

Write-Host "Downloading $binary $tag (windows/$arch)..."
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("bluefin-cli-install-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
try {
    $zip = Join-Path $tmp "bluefin-cli.zip"
    Invoke-WebRequest -Uri $url -OutFile $zip

    # Mirror install.sh: refuse to unpack an archive unless checksums.txt has a
    # valid SHA-256 entry for it and the digest matches.
    $checksums = Join-Path $tmp "checksums.txt"
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksums
    } catch {
        throw "could not download checksums for $archive"
    }

    $expectedSha = ""
    foreach ($line in Get-Content -Path $checksums) {
        $fields = $line.Trim() -split "\s+"
        if ($fields.Count -ge 2 -and $fields[1] -eq $archive) {
            $expectedSha = $fields[0]
            break
        }
    }
    if ($expectedSha -notmatch "^[0-9A-Fa-f]{64}$") {
        throw "checksums.txt has no valid SHA-256 entry for $archive"
    }

    $actualSha = (Get-FileHash -Path $zip -Algorithm SHA256).Hash
    if ($actualSha -ne $expectedSha) {
        throw "checksum verification failed for $archive"
    }

    Expand-Archive -Path $zip -DestinationPath $tmp -Force

    $exe = Join-Path $tmp "$binary.exe"
    if (-not (Test-Path $exe)) { throw "$binary.exe not found in archive" }

    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    Move-Item -Path $exe -Destination (Join-Path $binDir "$binary.exe") -Force

    Write-Host "Installed $binary $tag to $binDir"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (($userPath -split ";") -notcontains $binDir) {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$binDir", "User")
        $env:Path = "$env:Path;$binDir"
        Write-Host "Added $binDir to your user PATH (restart other terminals to pick it up)."
    }
    Write-Host "Run '$binary' to get started."
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
