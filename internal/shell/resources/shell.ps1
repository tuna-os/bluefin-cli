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

        return $null
    }

    # Cache the output of `<tool> init powershell` to avoid spawning a process on every startup.
    # The cache is keyed by the executable's last-write time, so it auto-refreshes on upgrade.
    function Invoke-CachedInit {
        param([string]$Exe)

        $cacheDir = "$env:LOCALAPPDATA\bluefin-cli\shell-cache"
        if (-not (Test-Path $cacheDir)) {
            $null = New-Item -ItemType Directory -Path $cacheDir -Force
        }

        $exeName = [System.IO.Path]::GetFileNameWithoutExtension($Exe)
        $exeMtime = (Get-Item $Exe).LastWriteTimeUtc.Ticks
        $cacheFile = "$cacheDir\$exeName-$exeMtime.ps1"

        if (-not (Test-Path $cacheFile)) {
            # Remove stale cache files for this executable before writing the new one
            Get-ChildItem "$cacheDir\$exeName-*.ps1" -ErrorAction SilentlyContinue | Remove-Item -Force
            & $Exe init powershell | Out-File $cacheFile -Encoding utf8
        }

        . $cacheFile
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

    if (-not (Get-Module -Name PSReadLine)) {
        Import-Module PSReadLine -ErrorAction SilentlyContinue
    }
    if (-not (Get-Module -Name PSFileIcons)) {
        Import-Module PSFileIcons -ErrorAction SilentlyContinue
    }

    $fzfExe = Get-BluefinExecutable "fzf"
    if ($fzfExe) {
        $fzfDir = Split-Path -Path $fzfExe -Parent
        if ($fzfDir -and ($env:PATH -notlike "*$fzfDir*")) {
            $env:PATH = "$fzfDir;$env:PATH"
        }
    }

    $zoxideExe = Get-BluefinExecutable "zoxide"
    if ($zoxideExe) {
        Invoke-CachedInit $zoxideExe
    }

    $atuinExe = Get-BluefinExecutable "atuin"
    if ($atuinExe) {
        Invoke-CachedInit $atuinExe
    }

    $starshipExe = Get-BluefinExecutable "starship"
    if ($starshipExe) {
        Invoke-CachedInit $starshipExe
    }

    $ezaExe = Get-BluefinExecutable "eza"
    if ($env:BLUEFIN_SHELL_ENABLE_EZA -eq "1") {
        if ($ezaExe) {
            function ll { & $script:ezaExe -al --icons=auto --group-directories-first }
            function ls { & $script:ezaExe --icons=auto --group-directories-first }
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
