# Configuration

`bluefin-cli` stores its settings in a YAML config file and reads a small set
of environment variables. This page documents both.

## Config file

The config file lives at `~/.config/bluefin-cli/config.yaml` and is managed
with [Viper](https://github.com/spf13/viper). You normally never edit it by
hand — commands like `bluefin-cli theme <flavor>` and the TUI's settings
screen call `viper.Set` and persist the file for you — but it is plain YAML if
you want to inspect or script against it.

### Location

`bluefin-cli` resolves the config directory in this order ([`internal/env.GetConfigDir`](../internal/env/env.go)):

1. `~/.config/bluefin-cli` if that directory already exists.
2. `$HOMEBREW_PREFIX/etc/bluefin-cli` if the `HOMEBREW_PREFIX` environment
   variable is set.
3. `~/.config/bluefin-cli` otherwise (created on first write).

### Keys

| Key | Default | Set by |
|---|---|---|
| `theme` | `catppuccin` | Reserved default; the active flavor is tracked under `ui.flavor` below. |
| `ui.flavor` | `auto` | `bluefin-cli theme <flavor>`, or the Settings screen in the TUI menu. `auto` follows your terminal's reported scheme. |
| `ui.dark_mode` | `true` | The Settings screen in the TUI menu. |
| `bundles.base_url` | `https://raw.githubusercontent.com/projectbluefin/common/main/system_files` | Not currently exposed as a command flag; override via config file or environment variable (see below) if you need a different source for downloaded (non-embedded) bundles. |
| `bundles.default_path` | `shared/usr/share/ublue-os/homebrew` | Same as above. |
| `game.high_score` | `0` | The hidden dino game in the TUI header. |
| `update.last_check` | unset | `bluefin-cli menu`, to throttle the update check to once per 24h. |
| `profile.gist_id` | unset | `bluefin-cli profile sync`, to remember which GitHub Gist a profile is synced to. |

### Example

```yaml
theme: catppuccin
ui:
  flavor: mocha
  dark_mode: true
game:
  high_score: 1337
```

## Environment variables

### `BLUEFIN_*` config overrides

Viper is configured with `SetEnvPrefix("BLUEFIN")` and `AutomaticEnv()`
([`internal/config/config.go`](../internal/config/config.go)), so **any** key
in the table above can be overridden by an environment variable: uppercase the
key and replace `.` with `_`. An environment variable always wins over the
config file.

| Variable | Overrides |
|---|---|
| `BLUEFIN_UI_FLAVOR` | `ui.flavor` |
| `BLUEFIN_UI_DARK_MODE` | `ui.dark_mode` |
| `BLUEFIN_THEME` | `theme` |
| `BLUEFIN_BUNDLES_BASE_URL` | `bundles.base_url` |
| `BLUEFIN_BUNDLES_DEFAULT_PATH` | `bundles.default_path` |

This mechanism is generic — it is not a hardcoded list of `os.Getenv` calls,
it is Viper reading any `BLUEFIN_<KEY>` variable that is set at process start.

### Other environment variables

These are read directly in Go code and are not part of the Viper config:

| Variable | Effect |
|---|---|
| `HOMEBREW_PREFIX` | When set, `bluefin-cli` adds `$HOMEBREW_PREFIX/etc/bluefin-cli` as a config search path and prefers it as the config directory if `~/.config/bluefin-cli` does not exist. Also consulted when locating `brew` itself. |
| `BLUEFIN_DISABLE_COUNTME` | Set to any non-empty value to opt out of the anonymous countme usage ping ([`internal/countme/countme.go`](../internal/countme/countme.go)). |
| `WSL_DISTRO_NAME`, `WSL_INTEROP` | Used (alongside `/proc` markers) to detect WSL; affects Windows-specific install paths. |
| `LOCALAPPDATA` | Used on Windows to locate per-user app data for terminal/font registration. |
| `SHELL` | Used to detect your current shell for `bluefin-cli status` and the shell-integration menu. |
| `XDG_CURRENT_DESKTOP` | Consulted by the installer to tailor desktop-specific install steps on Linux. |

### XDG Base Directory support

`bluefin-cli` does not currently implement full XDG Base Directory
specification support (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, etc.). The config
directory is hardcoded to `~/.config/bluefin-cli` (or the Homebrew prefix
path above), independent of `XDG_CONFIG_HOME`.

## User-defined bundles

Bundles are Homebrew `Brewfile`s. Curated bundles ship embedded in the
binary, but you can add your own:

```bash
mkdir -p ~/.config/bluefin-cli/bundles
cat > ~/.config/bluefin-cli/bundles/my-tools.Brewfile <<'EOF'
brew "ripgrep"
brew "fd"
EOF

bluefin-cli install my-tools
```

Any `*.Brewfile` in that directory is picked up automatically
([`internal/install/install.go`](../internal/install/install.go)) and listed
alongside the curated bundles (`ai`, `cli`, `cncf`, `experimental-ide`,
`fonts`, `full-desktop`, `ide`, `k8s`, `all`) when you run
`bluefin-cli install --list` or the TUI's Install Apps menu. User-defined
bundles extend the curated set — they do not shadow or replace any of it.
