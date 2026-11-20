## bluefin-cli motd toggle

Set MOTD display for shells

### Synopsis

Enable or disable MOTD display on shell startup for bash, zsh, fish, or all shells.

Where uwelcome (the terminal welcome tool of Bluefin) is installed,
this turns uwelcome on or off for every shell. Without on|off, it flips
the current state of uwelcome.

```
bluefin-cli motd toggle [shell|all] [on|off] [flags]
```

### Options

```
  -h, --help   help for toggle
```

### SEE ALSO

* [bluefin-cli motd](bluefin-cli_motd.md)	 — Set up daily shell messages

