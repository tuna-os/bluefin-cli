$ErrorActionPreference = 'Stop'

$packagesPath = 'tmp/windows-packages.txt'
$mappingPath = 'internal/install/windows_mapping.json'
$reportPath = 'tmp/winget-mapping-report.csv'
$unmatchedPath = 'tmp/winget-unmatched.txt'

$packages = Get-Content $packagesPath | Where-Object { $_.Trim() -ne '' } | Sort-Object -Unique
$mapping = Get-Content $mappingPath -Raw | ConvertFrom-Json -AsHashtable

$rows = foreach ($pkg in $packages) {
    if ($mapping.ContainsKey($pkg)) {
        $arr = @($mapping[$pkg])
        [pscustomobject]@{
            Package = $pkg
            Match = ($arr -join ';')
            Status = 'mapped'
        }
    } else {
        [pscustomobject]@{
            Package = $pkg
            Match = ''
            Status = 'unmapped'
        }
    }
}

$rows | Export-Csv -Path $reportPath -NoTypeInformation
$rows | Where-Object { $_.Status -eq 'unmapped' } | Select-Object -ExpandProperty Package | Set-Content $unmatchedPath

$mapped = ($rows | Where-Object { $_.Status -eq 'mapped' }).Count
$unmapped = ($rows | Where-Object { $_.Status -eq 'unmapped' }).Count
Write-Output "packages=$($packages.Count)"
Write-Output "mapped=$mapped"
Write-Output "unmapped=$unmapped"
