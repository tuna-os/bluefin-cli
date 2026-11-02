# Windows Setup Guide

This guide shows how to install and set up `bluefin-cli` on Windows.

## Install

- PowerShell script:

  ```powershell
  irm https://raw.githubusercontent.com/tuna-os/bluefin-cli/main/install.ps1 | iex
  ```

- Winget:

  ```powershell
  winget install --id Hanthor.BluefinCLI --exact
  ```

- Archive: get the Windows `.zip` file from the GitHub releases page.

If you install with Winget, `bluefin-cli update` tells you to run
`winget upgrade Hanthor.BluefinCLI`.

## Files

- Settings: `%USERPROFILE%\.config\bluefin-cli\config.yaml`
- Sunset settings: `%USERPROFILE%\.config\bluefin-cli\sunset.json`
- PowerShell init cache: `%LOCALAPPDATA%\bluefin-cli\shell-cache\init.ps1`

## PowerShell

To turn on the shell tools, run this command:

```powershell
bluefin-cli shell powershell on
```

## WSL

In WSL, the `bluefin-cli sunset` command must use Windows APIs. The tool finds
`bluefin-cli.exe` on the Windows `PATH` or in the WinGet `Links` folder. Then it
sends the command to that Windows binary.

## Theme and wallpaper

- `bluefin-cli sunset` sets `AppsUseLightTheme` and `SystemUsesLightTheme` in
  `HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`.
- `bluefin-cli wallpaper` sets the wallpaper with `SystemParametersInfoW`.

## Check your setup

```powershell
bluefin-cli doctor
```

For more help, see the [help page](troubleshooting.md).
