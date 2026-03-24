@{
    ModuleVersion     = '0.1.0'
    GUID              = '4a7b2c8d-3f1e-4a9b-8c7d-2e5f6a3b4c8d'
    Author            = 'James Reilly'
    CompanyName       = ''
    Copyright         = '(c) 2026 James Reilly. MIT License.'
    Description       = 'Fast Nerd Font file icons for Get-ChildItem. A lightweight replacement for Terminal-Icons — all icon mappings are compiled into a C# dictionary with no runtime file I/O.'
    PowerShellVersion = '5.1'

    # The root module (psm1) defines Format-PSFileIcon used by the format XML
    RootModule         = 'PSFileIcons.psm1'

    # The compiled DLL is loaded before the psm1 and format data
    RequiredAssemblies = @('PSFileIcons.dll')

    # Format is registered via Update-FormatData -PrependPath in PSFileIcons.psm1
    # (not via FormatsToProcess, which would append instead of prepend)
    FormatsToProcess   = @()

    FunctionsToExport  = @('Format-PSFileIcon')
    CmdletsToExport    = @()
    VariablesToExport  = @()
    AliasesToExport    = @()

    PrivateData = @{
        PSData = @{
            Tags         = @('Icons', 'NerdFonts', 'FileSystem', 'Terminal', 'Colors', 'Files', 'Prompt')
            LicenseUri   = 'https://github.com/hanthor/PSFileIcons/blob/main/LICENSE'
            ProjectUri   = 'https://github.com/hanthor/PSFileIcons'
            ReleaseNotes = 'Initial release.'
        }
    }
}
