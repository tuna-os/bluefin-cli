# Windows Setup Guide

This guide covers installation options, shell integration, terminal configuration,
WSL interop, and automation features for `bluefin-cli` on Windows.

## Installation

### Installation options

- **PowerShell installation script**:
  ```powershell
  irm https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.ps1 | iex
  ```
- **Windows Package Manager (Winget)**:
  ```cmd
  winget install Hanthor.BluefinCLI
  ```
- **Scoop**:
  ```cmd
  scoop bucket add tuna-os https://github.com/tuna-os/scoop-bucket
  scoop install bluefin-cli
  ```
- **Direct download**: Pre-compiled `.zip` packages for Windows (x86_64, arm64)
  are published on GitHub releases.

### Package manager availability and update behavior

- **Winget**: Default package backend for installing system tools and terminal emulators.
  Binaries install under `%LOCALAPPDATA%\Microsoft\WinGet\Links`.
- **Self-update**: `bluefin-cli update` verifies binary checksums for direct script
  installs. When installed via a package manager, it reports the appropriate upgrade command:
  ```cmd
  winget upgrade Hanthor.BluefinCLI
  ```
- **UAC and permissions**: Installing packages that install system-wide may require
  administrator rights. `bluefin-cli` uses `gsudo` if available when elevation is needed,
  or prompts through standard Windows UAC dialogs.

## Configuration resolution

On Windows, configuration directory paths resolve as follows:

1. `%USERPROFILE%\.config\bluefin-cli\` if that directory exists.
2. `%HOMEBREW_PREFIX%\etc\bluefin-cli\` if `HOMEBREW_PREFIX` is set.
3. `%USERPROFILE%\.config\bluefin-cli\` (created on first run).

Configuration files:
- Settings file: `%USERPROFILE%\.config\bluefin-cli\config.yaml`
- Sunset configuration: `%USERPROFILE%\.config\bluefin-cli\sunset.json`
- Windows Winget unmatched package logs: `%USERPROFILE%\.config\bluefin-cli\winget-unavailable.txt`

## Shell setup and differences

`bluefin-cli` supports PowerShell (`pwsh` and Windows PowerShell 5.1+), as well
as cross-platform shells (Bash, Fish, Nushell) running on Windows.

### PowerShell integration

Enable the shell integration in PowerShell:
```powershell
bluefin-cli shell powershell on
```

Key differences and implementation details:
- **Scoping**: All user-facing functions in `shell.ps1` (`ll`, `ls`, `cat`, `grep`, `sudo`)
  are declared in `global:` scope to prevent function loss after initialization.
- **Path resolution**: PowerShell initialization probes `%LOCALAPPDATA%\Microsoft\WinGet\Links`,
  `%LOCALAPPDATA%\Programs`, and `%ProgramFiles%` to ensure Winget-installed tools are reachable.
- **Startup cache**: Shell tool initialization (`starship`, `zoxide`) is cached under
  `$env:LOCALAPPDATA\bluefin-cli\shell-cache` to maintain fast terminal startup times.

### cmd.exe vs. PowerShell vs. Windows Terminal

- `cmd.exe` does not support interactive ANSI styling or the rich shell experience.
- PowerShell provides full support for prompt theming, aliases, and completions.
- **Windows Terminal** is the recommended terminal emulator:
  - Settings are detected at `%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json`
    (MS Store install) or `%LOCALAPPDATA%\Microsoft\Windows Terminal\settings.json` (unpackaged/Winget).
  - Font installation (`bluefin-cli fonts`) automatically configures Nerd Fonts in Windows Terminal settings.

## WSL (Windows Subsystem for Linux) and Sunset Switching

When running inside WSL (Ubuntu, Debian, Fedora Remix):
- `bluefin-cli` detects WSL environment indicators (`WSL_DISTRO_NAME`, `/proc/sys/kernel/osrelease`).
- **Sunset automation delegation**: Windows desktop theme switching and wallpaper settings
  require Windows APIs. When running `bluefin-cli sunset` commands inside WSL, the command
  locates `bluefin-cli.exe` on Windows (via Windows PATH or `%LOCALAPPDATA%\Microsoft\WinGet\Links\bluefin-cli.exe`)
  and delegates execution to the Windows host binary.

## Registry integration and automation

- **Theme switching**: `bluefin-cli sunset` modifies `AppsUseLightTheme` and `SystemUsesLightTheme`
  in `HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, then broadcasts
  `WM_SETTINGCHANGE` (`ImmersiveColorSet`) to update open windows instantly.
- **Wallpaper automation**: Wallpapers are applied natively using `SystemParametersInfoW`
  (`SPI_SETDESKWALLPAPER`) and stored in `HKCU\Control Panel\Desktop\Wallpaper`.

## Verification and troubleshooting

Check your installation status:
```cmd
bluefin-cli doctor
```

Common resolutions:
- Ensure `%LOCALAPPDATA%\Microsoft\WinGet\Links` is in your `PATH` environment variable.
- For execution policy restrictions in PowerShell, run:
  ```powershell
  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```
- See [Troubleshooting](troubleshooting.md) for additional error solutions.
