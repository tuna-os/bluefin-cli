## bluefin-cli install wallpapers cleanup

Clean up the files of the wallpaper sync

### Synopsis

Remove the files that the wallpaper sync of Bluefin CLI made. In WSL this removes generated Windows themes, copied wallpaper folders, helper scripts, scheduled tasks, and state. Use --all to also uninstall the known casks of wallpapers and remove local wallpaper folders.

```
bluefin-cli install wallpapers cleanup [flags]
```

### Options

```
      --all    Also uninstall known wallpaper casks and remove local wallpaper folders
  -h, --help   help for cleanup
```

### SEE ALSO

* [`bluefin-cli install wallpapers`](bluefin-cli_install_wallpapers.md) — Install wallpaper casks from ublue-os/tap

