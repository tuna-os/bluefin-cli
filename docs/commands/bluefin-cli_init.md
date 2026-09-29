## bluefin-cli init

Generate the init script for a shell

### Synopsis

Generate the shell initialization script for bluefin-cli.
Add the following to your shell configuration file:

Bash (~/.bashrc):
  eval "$(bluefin-cli init bash)"

Zsh (~/.zshrc):
  eval "$(bluefin-cli init zsh)"

Fish (~/.config/fish/config.fish):
  bluefin-cli init fish | source

PowerShell ($PROFILE):
  Invoke-Expression (& bluefin-cli init powershell)

Ash (~/.ashrc, with ENV="$HOME/.ashrc" exported from ~/.profile):
  eval "$(bluefin-cli init ash)"

Nushell (~/.config/nushell/config.nu) — nushell cannot evaluate a string.
Save the script to a file, and source that file:
  bluefin-cli init nu | save -f ~/.config/nushell/bluefin-cli.nu
  source ~/.config/nushell/bluefin-cli.nu

'bluefin-cli shell enable <shell>' wires any of these up for you.


```
bluefin-cli init [bash|zsh|fish|ash|nu|powershell|pwsh] [flags]
```

### Options

```
      --atuin             Enable Atuin
      --bat               Enable Bat (default true)
      --carapace          Enable Carapace
      --eza               Enable Eza (default true)
      --fzf               Enable Fzf (default true)
      --glow              Enable Glow (default true)
      --gsudo             Enable Gsudo (default true)
  -h, --help              help for init
      --motd              Enable MOTD (default true)
      --starship          Enable Starship (default true)
      --ugrep             Enable Ugrep (default true)
      --uutilscoreutils   Enable UutilsCoreutils (default true)
      --uutilsdiffutils   Enable UutilsDiffutils (default true)
      --uutilsfindutils   Enable UutilsFindutils (default true)
      --zoxide            Enable Zoxide (default true)
```

### SEE ALSO

* [`bluefin-cli`](bluefin-cli.md) — A CLI tool to manage Homebrew and customize your shell

