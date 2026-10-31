# Quick start

Follow this guide after [installation](../README.md#-installation) to set up your
shell, install tools, and save your setup for another machine.

## After installation

Open a terminal and check your setup:

```sh
bluefin-cli status
bluefin-cli doctor
```

`status` shows enabled shells and installed tools. `doctor` reports problems
and suggests fixes. If the command is not found, open a new terminal and check
that the binary's directory is on your `PATH`.

Start the interactive menu to explore the available actions:

```sh
bluefin-cli menu
```

`bluefin-cli` with no arguments opens the same menu.

Both the standard `bluefin-cli` and `bluefin-cli-plus` support this guide.
If you installed the plus binary, use `bluefin-cli-plus` in place of
`bluefin-cli` in the examples. Plus adds features such as sunset automation;
menu entries depend on your build and platform.
See [standard and plus builds](vanilla-vs-extra.md) for details.

## Interactive menu or CLI commands?

Use the terminal user interface (TUI) to discover features, browse choices,
and select individual packages. Use direct commands when you know the task
or want to repeat it in a script.

| Task | Menu | Direct command |
| --- | --- | --- |
| Inspect your setup | Status | `bluefin-cli status` |
| Set up your shell | Bluefin Shell | `bluefin-cli shell bash on` |
| Choose packages | Install Apps | `bluefin-cli install cli` installs the CLI bundle |
| Choose a prompt | Starship Theme | `bluefin-cli starship theme` opens the theme selector |

Direct commands can still open a selector or request confirmation. Check
`bluefin-cli <command> --help` for supported arguments and flags before you
use a command in a script.

In menus, use the arrow keys to move and Enter to select. Esc goes back;
`q` quits. Press `/` to filter choices, `Ctrl+p` to search all actions, or
`?` for help. See the [menu guide](menus.md) for the full navigation map.

## Recommended first steps

### 1. Enable the shell experience

The shell experience adds aliases and initializes tools such as eza, bat,
zoxide, and Starship. In **Bluefin Shell**, select the shells and components
you want to use. Or enable your current shell directly:

```sh
bluefin-cli shell bash on
```

Replace `bash` with `zsh`, `fish`, `ash`, `nu`, or `powershell` as appropriate.
For example, in Windows PowerShell:

```powershell
bluefin-cli shell powershell on
```

This updates your shell startup file and tries to install enabled tools.
Open a new terminal for the changes to take effect, then run `bluefin-cli status`
to check the result. To disable the integration, use the same command with `off`.

### 2. Install a tool bundle

A **bundle** is a named collection of packages for a task, such as CLI work,
AI development, or Kubernetes. List the choices, then install the CLI bundle:

```sh
bluefin-cli install list
bluefin-cli install cli
```

Bundle installation uses Homebrew on Linux and macOS, and winget on Windows.
Make sure the package manager is available first. In **Install Apps**, you can
choose individual packages instead of the whole bundle. Review the proposed
changes before you confirm. If you deselect an installed package, Bluefin CLI can remove it.
See [available tools](tools.md) for more choices.

### 3. Explore Starship themes

If Starship is missing, install it, then open the prompt theme selector:

```sh
bluefin-cli starship install
bluefin-cli starship theme
```

Choose a preset and open a new terminal to see your prompt. Starship must
be enabled in your shell components for its prompt to appear.

## Save a profile

A **profile** is a JSON file with your enabled shells, tool selections, and
interface theme flavor. Export it after you finish setup:

```sh
bluefin-cli profile export setup.json
```

On another machine with Bluefin CLI installed, copy the file and compare it
with that machine's setup before you import it:

```sh
bluefin-cli profile diff setup.json
bluefin-cli profile import setup.json
```

Import applies the saved tool settings and changes shell integration for the
platform. It can disable shells absent from the profile. A profile does not
copy your installed packages or all dotfiles. Use
[My Brewfile](commands/bluefin-cli_brewfile.md) to record and install packages
separately. See the [profile reference](commands/bluefin-cli_profile.md) for
more options.

## Next steps

- [Configuration](configuration.md): file locations, tool settings, and custom bundles.
- [Setup problems](troubleshooting.md): diagnosis and fixes for common setup problems.
- [Command reference](commands/bluefin-cli.md): all commands and flags for more advanced tasks.
- [Interactive menu guide](menus.md): menu actions and keyboard shortcuts.
