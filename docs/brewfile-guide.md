# Brewfile Guide

One Brewfile can list the packages for Linux, macOS, and Windows.
This guide shows the syntax, what `bluefin-cli brewfile install` does on each
platform, and some examples.

The default Brewfile is `~/Brewfile`. If `~/Brewfile` does not exist and
`~/.Brewfile` does, Bluefin CLI uses `~/.Brewfile`. Use `--file` to select a
different file.

## 1. Syntax

Bluefin CLI reads five package kinds from a Brewfile:

```ruby
# Homebrew (macOS, Linux)
brew "ripgrep"
cask "ghostty"

# Windows managers
winget "Microsoft.VisualStudioCode"
scoop "neovim"
choco "git"
```

| Kind | Target | Install command |
|---|---|---|
| `brew` | macOS, Linux | `brew bundle install` |
| `cask` | macOS, Linux | `brew bundle install` |
| `winget` | Windows | `winget install --exact --accept-package-agreements --accept-source-agreements --id <id>` |
| `scoop` | Windows | `cmd /c scoop install <id>` |
| `choco` | Windows | `choco install -y <id>` |

The tool does not read `tap` lines, `flatpak` lines, comments (`#`), or
blank lines into the package list. `brew bundle install` still reads the `tap` lines.

## 2. One File for All Platforms

When you run `bluefin-cli brewfile install`:

1. If the file has `brew` or `cask` lines, Bluefin CLI runs
   `brew bundle install --file=<path>`. If `brew` is not on your `PATH`, it
   shows this message and continues:

   ```text
   skipping brew/cask entries: homebrew not found (https://brew.sh)
   ```

2. For each `winget`, `scoop`, or `choco` line, Bluefin CLI runs the
   install command for that manager. If the manager is not on your `PATH`, it
   shows this message and continues:

   ```text
   skipping winget "Microsoft.VisualStudioCode": winget not installed
   ```

Thus you can keep Homebrew lines and Windows lines in the same file.

## 3. Errors

- If `brew bundle install` fails, the command stops with that error.
- Before each Windows package, Bluefin CLI shows a line like
  `winget: installing Microsoft.VisualStudioCode`.
- If the install of a Windows package fails, Bluefin CLI shows
  `winget <id> failed: <error>` and continues with the next package.

## 4. Record Your Packages

`bluefin-cli brewfile dump` writes the installed packages to the Brewfile:

```bash
bluefin-cli brewfile dump --file ./Team.Brewfile
```

On macOS and Linux, this command runs `brew bundle dump --force`. On Windows,
it writes the `winget`, `scoop`, and `choco` packages that it finds.

Other subcommands: `add`, `remove`, `list`, and `search`. Refer to the
[brewfile command reference](commands/bluefin-cli_brewfile.md).

## 5. Custom Bundles

Put extra bundle files in this directory:

```text
~/.config/bluefin-cli/bundles/<name>.Brewfile
```

These bundles show in `bluefin-cli install list` and in the TUI menu
**Install Apps**. Refer to [Configuration](configuration.md).

## 6. Use With Profiles

A profile keeps your shells, tools, and theme. A Brewfile keeps your packages.
To copy a setup to a new machine, use the two together.

On the old machine:

```bash
bluefin-cli profile export > my-profile.json
bluefin-cli brewfile dump --file ~/MyBrewfile
```

On the new machine:

```bash
bluefin-cli profile import my-profile.json
bluefin-cli brewfile install --file ~/MyBrewfile
```

To compare the machine with the profile:

```bash
bluefin-cli profile diff my-profile.json
```

## 7. Example

```ruby
# Command-line tools (macOS, Linux)
brew "bat"
brew "eza"
brew "fzf"
brew "ripgrep"
brew "starship"
brew "zoxide"

# Applications (macOS)
cask "ghostty"
cask "visual-studio-code"

# Windows
winget "Microsoft.VisualStudioCode"
winget "Git.Git"
winget "Starship.Starship"
scoop "7zip"
choco "curl"
```
