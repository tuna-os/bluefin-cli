function bluefin_init {
    function Get-BluefinExecutable {
        param([string]$Name)

        $command = Get-Command $Name -ErrorAction SilentlyContinue
        if ($command) {
            return $command.Source
        }

        $candidatePaths = @(
            "$env:LOCALAPPDATA\Microsoft\WinGet\Links\$Name.exe",
            "$env:LOCALAPPDATA\Programs\$Name\bin\$Name.exe",
            "$env:ProgramFiles\$Name\$Name.exe",
            "$env:ProgramFiles\$Name\bin\$Name.exe"
        )

        foreach ($candidate in $candidatePaths) {
            if (Test-Path $candidate) {
                return $candidate
            }
        }

        $wingetPackagesRoot = "$env:LOCALAPPDATA\Microsoft\WinGet\Packages"
        if (Test-Path $wingetPackagesRoot) {
            $resolved = Get-ChildItem -Path $wingetPackagesRoot -Filter "$Name.exe" -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($resolved) {
                return $resolved.FullName
            }
        }

        return $null
    }

    $wingetLinksPath = "$env:LOCALAPPDATA\Microsoft\WinGet\Links"
    if ((Test-Path $wingetLinksPath) -and ($env:PATH -notlike "*$wingetLinksPath*")) {
        $env:PATH = "$wingetLinksPath;$env:PATH"
    }

    $windowsModulePaths = @(
        "$HOME\Documents\WindowsPowerShell\Modules",
        "$env:ProgramFiles\WindowsPowerShell\Modules",
        "$env:WINDIR\System32\WindowsPowerShell\v1.0\Modules"
    )

    foreach ($modulePath in $windowsModulePaths) {
        if ((Test-Path $modulePath) -and ($env:PSModulePath -notlike "*$modulePath*")) {
            $env:PSModulePath = "$env:PSModulePath;$modulePath"
        }
    }

    if (Get-Module -ListAvailable -Name PSReadLine -ErrorAction SilentlyContinue) {
        Import-Module PSReadLine -ErrorAction SilentlyContinue
    }

    if (Get-Module -ListAvailable -Name Terminal-Icons -ErrorAction SilentlyContinue) {
        Import-Module Terminal-Icons -ErrorAction SilentlyContinue
    }

    $fzfExe = Get-BluefinExecutable "fzf"
    if ($fzfExe) {
        $fzfDir = Split-Path -Path $fzfExe -Parent
        if ($fzfDir -and ($env:PATH -notlike "*$fzfDir*")) {
            $env:PATH = "$fzfDir;$env:PATH"
        }
    }

    if ((Get-Module -ListAvailable -Name PSFzf -ErrorAction SilentlyContinue) -and $fzfExe) {
        Import-Module PSFzf -ErrorAction SilentlyContinue
    }

    $zoxideExe = Get-BluefinExecutable "zoxide"
    if ($zoxideExe) {
        Invoke-Expression (& { (& $zoxideExe init powershell | Out-String) })
    }

    $atuinExe = Get-BluefinExecutable "atuin"
    if ($atuinExe) {
        Invoke-Expression (& { (& $atuinExe init powershell | Out-String) })
    }

    $starshipExe = Get-BluefinExecutable "starship"
    if ($starshipExe) {
        Invoke-Expression (& { (& $starshipExe init powershell | Out-String) })
    }

    $ezaExe = Get-BluefinExecutable "eza"
    if ($env:BLUEFIN_SHELL_ENABLE_EZA -eq "1") {
        if ($ezaExe) {
            function ll { & $script:ezaExe -al --group-directories-first }
            function ls { & $script:ezaExe --group-directories-first }
        }
    }

    $batExe = Get-BluefinExecutable "bat"
    if ($env:BLUEFIN_SHELL_ENABLE_BAT -eq "1") {
        if ($batExe) {
            function cat { & $script:batExe @Args }
        }
    }

    $ugrepExe = Get-BluefinExecutable "ug"
    if ($env:BLUEFIN_SHELL_ENABLE_UGREP -eq "1") {
        if ($ugrepExe) {
            function grep { & $script:ugrepExe @Args }
        }
    }

    $gsudoExe = Get-BluefinExecutable "gsudo"
    if ($env:BLUEFIN_SHELL_ENABLE_GSUDO -eq "1" -and $gsudoExe) {
        function sudo { & $script:gsudoExe @Args }
    }

    $bluefinCliExe = Get-BluefinExecutable "bluefin-cli"
    if ($env:BLUEFIN_SHELL_ENABLE_MOTD -eq "1" -and $Host.Name -ne 'ServerRemoteHost' -and $bluefinCliExe) {
        & $bluefinCliExe motd show
    }
}

bluefin_init
