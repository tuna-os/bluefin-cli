# Note on scoping, which this file got wrong for a long time: PowerShell scopes
# a `function` declared inside another function to that parent, and discards it
# when the parent returns. Every user-facing function below was therefore gone
# by the time bluefin_init finished, so ll, ls, cat, grep and sudo were never
# actually defined for anyone. They are declared `global:` for that reason, and
# the executables they call are held in global variables too -- a `$script:`
# reference to a variable that was assigned locally reads as empty, so the
# functions would have run `& $null` even if they had survived.
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

        # PowerShell runs on Linux and macOS too, where LOCALAPPDATA does not
        # exist. Interpolating it empty produced "\bluefin-cli\shell-cache",
        # which is an absolute path at the filesystem root there: a normal user
        # got "Access to the path '/bluefin-cli' is denied" printed at every
        # shell start with zoxide and starship then never initializing, and a
        # root shell quietly created /bluefin-cli instead. Join-Path keeps the
        # separators right on each platform.
        $cacheRoot = $env:LOCALAPPDATA
        if (-not $cacheRoot) { $cacheRoot = $env:XDG_CACHE_HOME }
        if (-not $cacheRoot) {
            $home_ = if ($env:HOME) { $env:HOME } else { $env:USERPROFILE }
            if ($home_) { $cacheRoot = Join-Path $home_ '.cache' }
        }
        if (-not $cacheRoot) { $cacheRoot = [System.IO.Path]::GetTempPath() }

        $cacheDir = Join-Path (Join-Path $cacheRoot 'bluefin-cli') 'shell-cache'
        # Exposed so the location can be inspected -- by a user debugging a
        # stale cache, and by scripts/shell-experience.sh, which asserts it
        # lands somewhere writable rather than re-deriving the path itself.
        $global:BluefinShellCacheDir = $cacheDir
        if (-not (Test-Path $cacheDir)) {
            $null = New-Item -ItemType Directory -Path $cacheDir -Force -ErrorAction SilentlyContinue
        }
        if (-not (Test-Path $cacheDir)) {
            # Without a usable cache, still initialize the tool -- just without
            # the startup saving. Failing to cache must not cost the feature.
            return (& $Exe init powershell | Out-String)
        }

        $exeName = [System.IO.Path]::GetFileNameWithoutExtension($Exe)
        $exeMtime = (Get-Item $Exe).LastWriteTimeUtc.Ticks
        $cacheFile = Join-Path $cacheDir "$exeName-$exeMtime.ps1"

        if (-not (Test-Path $cacheFile)) {
            # Remove stale cache files for this executable before writing the new one
            Get-ChildItem (Join-Path $cacheDir "$exeName-*.ps1") -ErrorAction SilentlyContinue | Remove-Item -Force
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

    $global:BluefinEzaExe = Get-BluefinExecutable "eza"
    if ($env:BLUEFIN_SHELL_ENABLE_EZA -eq "1") {
        if ($global:BluefinEzaExe) {
            function global:ll { & $global:BluefinEzaExe -al --icons=auto --group-directories-first }
            function global:ls { & $global:BluefinEzaExe --icons=auto --group-directories-first }
        }
    }

    $global:BluefinBatExe = Get-BluefinExecutable "bat"
    if ($env:BLUEFIN_SHELL_ENABLE_BAT -eq "1") {
        if ($global:BluefinBatExe) {
            function global:cat { & $global:BluefinBatExe @Args }
        }
    }

    $global:BluefinUgrepExe = Get-BluefinExecutable "ug"
    if ($env:BLUEFIN_SHELL_ENABLE_UGREP -eq "1") {
        if ($global:BluefinUgrepExe) {
            function global:grep { & $global:BluefinUgrepExe @Args }
        }
    }

    $global:BluefinGsudoExe = Get-BluefinExecutable "gsudo"
    if ($env:BLUEFIN_SHELL_ENABLE_GSUDO -eq "1" -and $global:BluefinGsudoExe) {
        function global:sudo { & $global:BluefinGsudoExe @Args }
    }

    $bluefinCliExe = Get-BluefinExecutable "bluefin-cli"
    if ($env:BLUEFIN_SHELL_ENABLE_MOTD -eq "1" -and $Host.Name -ne 'ServerRemoteHost' -and $bluefinCliExe) {
        & $bluefinCliExe motd show
    }
}

bluefin_init
