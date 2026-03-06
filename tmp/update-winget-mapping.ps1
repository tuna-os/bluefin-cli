$ErrorActionPreference = 'Stop'

$packagesPath = 'tmp/windows-packages.txt'
$mappingPath = 'internal/install/windows_mapping.json'
$reportPath = 'tmp/winget-mapping-report.csv'
$unmatchedPath = 'tmp/winget-unmatched.txt'

if (-not (Test-Path $packagesPath)) {
    throw "Package list not found: $packagesPath"
}
if (-not (Test-Path $mappingPath)) {
    throw "Mapping JSON not found: $mappingPath"
}

$rawMapping = Get-Content $mappingPath -Raw
$existing = ConvertFrom-Json -InputObject $rawMapping -AsHashtable
if ($null -eq $existing) {
    $existing = @{}
}

function Get-UniqueStrings {
    param([string[]]$Values)
    $seen = @{}
    $out = New-Object System.Collections.Generic.List[string]
    foreach ($v in $Values) {
        $vv = "$v".Trim()
        if ([string]::IsNullOrWhiteSpace($vv)) { continue }
        if ($seen.ContainsKey($vv)) { continue }
        $seen[$vv] = $true
        [void]$out.Add($vv)
    }
    return $out.ToArray()
}

function Find-FirstWingetIdFromOutput {
    param([string[]]$Lines)

    $separatorSeen = $false
    foreach ($line in $Lines) {
        if (-not $separatorSeen) {
            if ($line -match '^\s*-{5,}\s*$') {
                $separatorSeen = $true
            }
            continue
        }

        if ($line -match '^\s*(.+?)\s{2,}([A-Za-z0-9][A-Za-z0-9\._-]+)\s{2,}(\S+)\s*$') {
            return $Matches[2]
        }
    }

    return $null
}

function Try-ResolveWingetId {
    param([string]$Package)

    $baseName = ($Package -split '/')[-1]
    $noLinux = $baseName -replace '-linux$', ''
    $strippedPrefixes = $baseName
    if ($baseName -match '^[a-z0-9-]+-(.+)$') {
        $strippedPrefixes = $Matches[1]
    }

    $exactCandidates = Get-UniqueStrings @($Package, $baseName, $noLinux, $strippedPrefixes)
    foreach ($candidate in $exactCandidates) {
        if ($candidate.Contains('/')) { continue }
        $out = & winget search --id $candidate --exact --source winget --accept-source-agreements 2>&1
        if ($LASTEXITCODE -eq 0) {
            return @{ Id = $candidate; Method = 'exact-id' }
        }
    }

    $queryCandidates = Get-UniqueStrings @($baseName, $noLinux, $strippedPrefixes, $Package)
    foreach ($query in $queryCandidates) {
        if ([string]::IsNullOrWhiteSpace($query)) { continue }
        $out = & winget search --query $query --source winget --accept-source-agreements -n 8 2>&1
        if ($LASTEXITCODE -ne 0) { continue }

        $id = Find-FirstWingetIdFromOutput -Lines $out
        if (-not [string]::IsNullOrWhiteSpace($id)) {
            return @{ Id = $id; Method = "query:$query" }
        }
    }

    return $null
}

$packages = Get-Content $packagesPath | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
$packages = $packages | Sort-Object -Unique

$report = New-Object System.Collections.Generic.List[object]
$unmatched = New-Object System.Collections.Generic.List[string]

foreach ($pkg in $packages) {
    $pkgKey = $pkg.Trim()
    if ($existing.ContainsKey($pkgKey)) {
        $curr = @($existing[$pkgKey])
        $resolved = if ($curr.Count -gt 0) { $curr[0] } else { '' }
        [void]$report.Add([pscustomobject]@{ Package = $pkgKey; Query = ''; Match = $resolved; Method = 'existing' })
        continue
    }

    $resolved = Try-ResolveWingetId -Package $pkgKey
    if ($null -ne $resolved) {
        $existing[$pkgKey] = @($resolved.Id)
        [void]$report.Add([pscustomobject]@{ Package = $pkgKey; Query = ''; Match = $resolved.Id; Method = $resolved.Method })
    } else {
        [void]$report.Add([pscustomobject]@{ Package = $pkgKey; Query = ''; Match = ''; Method = 'unmatched' })
        [void]$unmatched.Add($pkgKey)
    }
}

$ordered = [ordered]@{}
foreach ($k in ($existing.Keys | Sort-Object)) {
    $ordered[$k] = @($existing[$k] | ForEach-Object { "$_" })
}

$ordered | ConvertTo-Json -Depth 8 | Set-Content $mappingPath
$report | Sort-Object Package | Export-Csv -Path $reportPath -NoTypeInformation
$unmatched | Sort-Object -Unique | Set-Content $unmatchedPath

Write-Output "packages=$($packages.Count)"
Write-Output "mappings=$($ordered.Count)"
Write-Output "unmatched=$((Get-Content $unmatchedPath | Where-Object { $_.Trim() -ne '' }).Count)"
