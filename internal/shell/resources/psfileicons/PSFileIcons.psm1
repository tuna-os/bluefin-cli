function Format-PSFileIcon {
    param([System.IO.FileSystemInfo]$Item)
    if ($Item.PSIsContainer) {
        [PSFileIcons.IconMap]::GetDirIcon($Item.Name) + ' ' + $Item.Name
    } else {
        [PSFileIcons.IconMap]::GetFileIcon($Item.Name, $Item.Extension) + ' ' + $Item.Name
    }
}

# Explicitly prepend our format so it takes priority over the built-in "children" view.
# FormatsToProcess in the manifest uses AppendPath by default, so we need this.
$_formatFile = Join-Path $PSScriptRoot 'PSFileIcons.format.ps1xml'
Update-FormatData -PrependPath $_formatFile
Remove-Variable _formatFile
