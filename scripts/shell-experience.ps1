#!/usr/bin/env pwsh
# Behavioral test of the PowerShell shell experience.
#
# The cross-shell suite in shell-experience.sh covers pwsh on Linux, which is
# enough to catch the scoping rules. This script exists to run the same
# assertions on Windows, where shell.ps1's executable lookup actually probes
# LOCALAPPDATA, the WinGet Links directory and ProgramFiles -- paths that do
# not exist anywhere else, so a lookup that silently finds nothing would look
# identical to a clean pass on Linux.
#
# Usage: scripts/shell-experience.ps1 <binary>

param([Parameter(Mandatory = $true)][string]$Binary)

$ErrorActionPreference = 'Stop'
$script:fail = 0

function Check($name, $cond) {
    if ($cond) { "ok   $name" } else { "FAIL $name"; $script:fail = 1 }
}
function Skip($name) { "skip $name" }

$bin = (Resolve-Path $Binary).Path
$sandbox = Join-Path ([System.IO.Path]::GetTempPath()) "bfc-shell-$([guid]::NewGuid())"
New-Item -ItemType Directory -Path $sandbox -Force | Out-Null
try {
    $env:HOME = $sandbox
    $env:USERPROFILE = $sandbox

    $initFile = Join-Path $sandbox 'init.ps1'
    & $bin init powershell 2>$null | Set-Content -Path $initFile -Encoding utf8

    # Loaded through Invoke-Expression, which is what $PROFILE is documented to
    # use. Dot-sourcing has different scoping rules, and that difference is
    # precisely where this script was broken: functions declared inside
    # bluefin_init were discarded when it returned.
    $probe = @'
param($InitFile)
Invoke-Expression (Get-Content $InitFile -Raw)
$r = [ordered]@{}
$r.ll    = [bool](Get-Command ll -CommandType Function -ErrorAction SilentlyContinue)
$r.ls    = [bool](Get-Command ls -CommandType Function -ErrorAction SilentlyContinue)
$r.z     = [bool](Get-Command z -ErrorAction SilentlyContinue)
$r.lsRan = $false
if ($r.ls) { try { $o = ls 2>&1; $r.lsRan = [bool]$o } catch { $r.lsRan = $false } }
$r | ConvertTo-Json -Compress
'@
    $probeFile = Join-Path $sandbox 'probe.ps1'
    Set-Content -Path $probeFile -Value $probe -Encoding utf8

    $out = & pwsh -NoProfile -File $probeFile -InitFile $initFile 2>&1 | ForEach-Object { "$_" }
    $json = $out | Where-Object { $_.TrimStart().StartsWith('{') } | Select-Object -Last 1
    if (-not $json) {
        "::error::the probe produced no result"
        $out
        exit 1
    }
    $r = $json | ConvertFrom-Json

    # Anything printed besides the result is startup noise the user would see.
    $noise = $out | Where-Object { -not $_.TrimStart().StartsWith('{') -and $_.Trim() }
    Check "startup is quiet" (-not $noise)
    if ($noise) { $noise | ForEach-Object { "     noise: $_" } }

    if (Get-Command eza -ErrorAction SilentlyContinue) {
        Check "ll is defined" $r.ll
        Check "ls is defined" $r.ls
        # Being defined is not the same as working: the functions call an
        # executable held in a variable, and an out-of-scope one makes them
        # run `& $null` while still existing.
        Check "the ls alias actually runs" $r.lsRan
    } else {
        Skip "eza is not installed"
    }

    if (Get-Command zoxide -ErrorAction SilentlyContinue) {
        Check "zoxide initialized" $r.z
    } else {
        Skip "zoxide is not installed"
    }
} finally {
    Remove-Item -Recurse -Force $sandbox -ErrorAction SilentlyContinue
}

""
if ($script:fail -ne 0) {
    "Shell experience (PowerShell): FAILURES"
    exit 1
}
"Shell experience (PowerShell): all checks passed"
