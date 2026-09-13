# bluefin-cli shell experience for Nushell.
#
# Nushell has no `eval` for strings, so this script is rendered to a file by
# `bluefin-cli shell enable nu` and sourced from config.nu, rather than being
# piped into the shell the way the POSIX and fish scripts are. Regenerate it
# with `bluefin-cli init nu | save -f ~/.config/nushell/bluefin-cli.nu`.
#
# Every BLUEFIN_SHELL_ENABLE_* variable this script reads is written by the
# preamble `bluefin-cli init nu` emits directly above it, so nothing here
# needs to defend against an unset one. That is also why the script is kept to
# plain, long-stable nushell: it is generated once and then parsed on every
# shell start, where a syntax error costs the user their prompt.

# The aliases are not here on purpose. Nushell's `alias` is a parse-time
# keyword, so one written inside an `if` belongs to that block's scope and is
# gone when the block ends -- these same aliases lived in conditionals here and
# defined nothing at all. `bluefin-cli init nu` now emits the ones that apply
# above this point, at the top level. See internal/shell/nu_aliases.go.

# Homebrew prefix, resolved the same way shell.sh resolves it.
if ($env.HOMEBREW_PREFIX? | is-empty) {
    $env.HOMEBREW_PREFIX = (
        if ("/opt/homebrew/bin/brew" | path exists) {
            "/opt/homebrew"
        } else if ("/usr/local/bin/brew" | path exists) {
            "/usr/local"
        } else {
            "/home/linuxbrew/.linuxbrew"
        }
    )
}

# uutils install into a non-standard libexec directory, so each enabled one is
# prepended to PATH rather than being picked up from <prefix>/bin.
if $env.BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS == "1" {
    $env.PATH = ($env.PATH | prepend ([$env.HOMEBREW_PREFIX "opt" "uutils-coreutils" "libexec" "uubin"] | path join))
}
if $env.BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS == "1" {
    $env.PATH = ($env.PATH | prepend ([$env.HOMEBREW_PREFIX "opt" "uutils-findutils" "libexec" "uubin"] | path join))
}
if $env.BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS == "1" {
    $env.PATH = ($env.PATH | prepend ([$env.HOMEBREW_PREFIX "opt" "uutils-diffutils" "libexec" "uubin"] | path join))
}

# atuin, starship, zoxide and carapace each generate nushell source of their
# own, and nushell can only `source` a path known at parse time. Each enabled
# one is therefore written into the vendor autoload directory, which nushell
# sources automatically on the next start -- so a freshly enabled tool becomes
# active one shell later. Set BLUEFIN_SKIP_AUTOLOAD_REFRESH=1 to skip this
# when startup latency matters more than picking changes up promptly.
if ($env.BLUEFIN_SKIP_AUTOLOAD_REFRESH? | default "0") != "1" {
    let autoload_dir = ($nu.data-dir | path join "vendor" "autoload")
    mkdir $autoload_dir

    if $env.BLUEFIN_SHELL_ENABLE_ATUIN == "1" and (which atuin | is-not-empty) {
        atuin init nu | save -f ($autoload_dir | path join "atuin.nu")
    }

    if $env.BLUEFIN_SHELL_ENABLE_STARSHIP == "1" and (which starship | is-not-empty) {
        starship init nu | save -f ($autoload_dir | path join "starship.nu")
    }

    if $env.BLUEFIN_SHELL_ENABLE_ZOXIDE == "1" and (which zoxide | is-not-empty) {
        zoxide init nushell | save -f ($autoload_dir | path join "zoxide.nu")
    }

    if $env.BLUEFIN_SHELL_ENABLE_CARAPACE == "1" and (which carapace | is-not-empty) {
        $env.CARAPACE_BRIDGES = "zsh,fish,bash,inshellisense"
        carapace _carapace nushell | save -f ($autoload_dir | path join "carapace.nu")
    }
}

if $env.BLUEFIN_SHELL_ENABLE_MOTD == "1" and (which bluefin-cli | is-not-empty) {
    bluefin-cli motd show
}
