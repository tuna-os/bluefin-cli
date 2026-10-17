## bluefin-cli uninstall

Remove the Bluefin shell setup and managed tools

### Synopsis

Remove shell initialization for powershell, bash, zsh, and fish.

By default this command also tries to uninstall managed software:
  - Windows: winget-managed shell tools
  - Linux/macOS/WSL: Homebrew-managed shell tools

Use flags to keep software/modules/config as needed.

```
bluefin-cli uninstall [flags]
```

### Options

```
  -h, --help          help for uninstall
      --keep-config   Keep Bluefin shell preferences JSON
      --modules       Uninstall PowerShell modules managed by Bluefin CLI (default true)
      --software      Uninstall managed software packages (default true)
```

### SEE ALSO

* [bluefin-cli](bluefin-cli.md)	 — CLI tool for Homebrew and shell setup

