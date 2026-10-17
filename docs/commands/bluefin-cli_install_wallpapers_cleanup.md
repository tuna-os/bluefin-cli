## bluefin-cli install wallpapers cleanup

Remove sync files for wallpapers

### Synopsis

Remove wallpaper sync files. In WSL, this removes generated Windows themes, copied wallpaper folders, helper scripts, scheduled tasks, and state. With --all, the command also removes wallpaper casks and local wallpaper folders.

```
bluefin-cli install wallpapers cleanup [flags]
```

### Options

```
      --all    Also uninstall known wallpaper casks and remove local wallpaper folders
  -h, --help   help for cleanup
```

### SEE ALSO

* [bluefin-cli install wallpapers](bluefin-cli_install_wallpapers.md)	 — Install wallpaper casks from ublue-os/tap

