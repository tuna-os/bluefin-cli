## bluefin-cli uninstall

Uninstall the Bluefin shell setup and its tools

### Synopsis

Remove the Bluefin shell setup from powershell, bash, zsh, and fish.

By default this command also tries to uninstall the software it manages:
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

* [`bluefin-cli`](bluefin-cli.md) — A CLI tool to manage Homebrew and customize your shell

