# Profile Workflow Guide

A profile is a JSON file. It records your theme flavor, your tool settings, and
the shells that have `bluefin-cli` integration. Use a profile to copy your
setup to a different machine.

## 1. Profile contents

`bluefin-cli profile export` writes a JSON file with this shape:

```json
{
  "version": 1,
  "flavor": "mocha",
  "tools": {
    "atuin": true,
    "bat": true,
    "carapace": false,
    "eza": true,
    "fzf": true,
    "glow": true,
    "gsudo": false,
    "motd": true,
    "starship": true,
    "ugrep": true,
    "uutilscoreutils": true,
    "uutilsdiffutils": true,
    "uutilsfindutils": true,
    "zoxide": true
  },
  "enabled_shells": [
    "bash",
    "fish",
    "zsh"
  ]
}
```

The fields are:

- `version`: The schema version. The only valid value is `1`.
- `flavor`: The value of `ui.flavor`. Valid values are `auto`, `latte`,
  `frappe`, `macchiato`, and `mocha`.
- `tools`: The tool keys and their state. A key is the tool name in lowercase,
  for example `uutilscoreutils`.
- `enabled_shells`: The shells that have `bluefin-cli` integration.

An import changes these items:

- `ui.flavor` in `~/.config/bluefin-cli/config.yaml`.
- The tool settings in `~/.config/bluefin-cli/shell.json`.
- The `bluefin-cli` lines in your shell rc files. Your other rc lines do not
  change.

A profile does not keep passwords, tokens, or SSH keys.

## 2. Export and import

These commands use local files. They do not need a GitHub account.

To show the profile in the terminal:

```bash
bluefin-cli profile export
```

To write the profile to a file:

```bash
bluefin-cli profile export my-profile.json
```

### Find the differences

Before you apply a profile, compare it with this machine:

```bash
bluefin-cli profile diff my-profile.json
```

If there are no differences, the output is:

```text
✓ No drift — this machine matches the profile.
```

If there are differences, the command shows each change:

```text
Import would change:
  flavor: latte -> mocha
  tool atuin: off -> on
  shell fish: disabled -> enabled
```

Add `--exit-code` to get a non-zero exit status when differences exist. Scripts
can use this status:

```bash
if ! bluefin-cli profile diff my-profile.json --exit-code; then
  echo "The setup is different from the profile."
fi
```

### Apply a profile

```bash
bluefin-cli profile import my-profile.json
```

## 3. Sync with a GitHub gist

`profile push` and `profile pull` keep the profile in a private GitHub gist.
These commands use the GitHub CLI (`gh`). Install `gh` and run `gh auth login`
first.

### Push

```bash
bluefin-cli profile push
```

The first push makes a private gist. The gist contains the file
`bluefin-profile.json`. `bluefin-cli` keeps the gist ID in
`~/.config/bluefin-cli/config.yaml` as `profile.gist_id`. The output shows the
ID:

```text
✓ Profile pushed to new private gist 7a8b9c0d1e2f3a4b5c6d (saved in config)
```

Each subsequent push on that machine replaces the contents of the same gist.

### Pull

On a different machine, give the gist ID:

```bash
bluefin-cli profile pull 7a8b9c0d1e2f3a4b5c6d
```

`bluefin-cli` gets `bluefin-profile.json` from the gist and applies it. If the
machine has no gist ID yet, `bluefin-cli` keeps this ID as `profile.gist_id`.
After that, you can pull without the ID:

```bash
bluefin-cli profile pull
```

### Prevent lost changes

A gist does not merge changes. The last push replaces the gist contents.
Before you push from a second machine, compare the gist with that machine:

```bash
gh gist view 7a8b9c0d1e2f3a4b5c6d --filename bluefin-profile.json --raw > /tmp/remote-profile.json
bluefin-cli profile diff /tmp/remote-profile.json
```

Then select one of these steps:

- To use the gist settings, run `bluefin-cli profile pull`.
- To keep the settings of this machine, run `bluefin-cli profile push`.

## 4. Different operating systems

The flavor and the tool settings have the same meaning on all platforms. The
shell step uses the operating system of the target machine:

| Platform | Shells that import changes |
|---|---|
| Linux, macOS, and WSL | `bash`, `zsh`, `fish` |
| Windows | `powershell` |

On Linux, macOS, and WSL, import enables each of `bash`, `zsh`, and `fish`
if `enabled_shells` lists it. It disables each one that the list does not
include. On Windows, import does the same for `powershell`. Import ignores
the other shells in the list.

## 5. Keep a profile in Git

You can keep a profile in a dotfiles repository:

```bash
bluefin-cli profile export ~/dotfiles/bluefin/profile.json
cd ~/dotfiles
git add bluefin/profile.json
git commit -m "Update bluefin-cli profile"
```

If a profile has a different `version`, import stops with an error:

```text
unsupported profile version 2
```

## 6. Troubleshooting

### Push or pull shows "profile sync uses the GitHub CLI"

`gh` is not on your `PATH`. Install the GitHub CLI and run `gh auth login`.

### Pull shows "no gist configured"

You did not give a gist ID, and `profile.gist_id` has no value. Give the ID:

```bash
bluefin-cli profile pull <gist-id>
```

### A tool does not start after import

A profile keeps the tool settings, not the tool programs. Run
`bluefin-cli doctor` to find the missing programs.
