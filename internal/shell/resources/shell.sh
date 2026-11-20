#!/usr/bin/env sh

# Load once per shell, so that re-reading the rc file does not initialize atuin
# a second time: https://github.com/atuinsh/atuin/issues/380#issuecomment-1594014644
#
# This is a conditional rather than an early `return` because bash and zsh are
# documented to run this through `eval "$(bluefin-cli init bash)"`, and `return`
# at the top level of an eval is an error -- "can only `return' from a function
# or sourced script" -- which bash printed on every `source ~/.bashrc` while
# carrying on to run the body anyway. So the guard never actually guarded
# anything on the path almost every user takes.
if [ "${SOURCED_BLUEFIN_SHELL:-0}" != 1 ]; then
SOURCED_BLUEFIN_SHELL=1

# Default to enabled if variable is not set (backwards compatibility)
: "${BLUEFIN_SHELL_ENABLE_EZA:=1}"
: "${BLUEFIN_SHELL_ENABLE_BAT:=1}"
: "${BLUEFIN_SHELL_ENABLE_STARSHIP:=1}"
: "${BLUEFIN_SHELL_ENABLE_ZOXIDE:=1}"
: "${BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS:=1}"
: "${BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS:=1}"
: "${BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS:=1}"
: "${BLUEFIN_SHELL_ENABLE_MOTD:=1}"

# Default disabled tools
: "${BLUEFIN_SHELL_ENABLE_CARAPACE:=0}"
: "${BLUEFIN_SHELL_ENABLE_DIRENV:=0}"
: "${BLUEFIN_SHELL_ENABLE_MISE:=0}"
: "${BLUEFIN_SHELL_ENABLE_UGREP:=0}"
: "${BLUEFIN_SHELL_ENABLE_ATUIN:=0}"

# eza
# ls aliases
if [ "$BLUEFIN_SHELL_ENABLE_EZA" -eq 1 ] && [ "$(command -v eza)" ]; then
    alias ll='eza -l --icons=auto --group-directories-first'
    alias l.='eza -d .*'
    alias ls='eza'
    alias l1='eza -1'
fi

# ugrep 
# for grep
if [ "$BLUEFIN_SHELL_ENABLE_UGREP" -eq 1 ] && [ "$(command -v ug)" ]; then
    alias grep='ug'
    alias egrep='ug -E'
    alias fgrep='ug -F'
    alias xzgrep='ug -z'
    alias xzegrep='ug -zE'
    alias xzfgrep='ug -zF'
fi

# bat for cat. Check the binary too: `cat` is the one alias a missing tool
# must never break, and a cached init script can outlive an uninstalled bat.
if [ "$BLUEFIN_SHELL_ENABLE_BAT" -eq 1 ] && [ "$(command -v bat)" ]; then
    alias cat='bat --style=plain --pager=never'
fi

# open -> xdg-open, the macOS-style opener. Only where no `open` command
# exists already: macOS has its own, and Debian's is an alternatives link.
if [ "$(command -v xdg-open)" ] && [ ! -e /usr/bin/open ]; then
    alias open='xdg-open >/dev/null 2>&1'
fi

HOMEBREW_PREFIX="${HOMEBREW_PREFIX:-}"
if [ -z "$HOMEBREW_PREFIX" ]; then
    [ -x "/opt/homebrew/bin/brew" ] && HOMEBREW_PREFIX="/opt/homebrew" || \
    [ -x "/usr/local/bin/brew" ] && HOMEBREW_PREFIX="/usr/local" || \
    HOMEBREW_PREFIX="/home/linuxbrew/.linuxbrew"
fi

# uutils
[ "$BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS" -eq 1 ] && PATH="${HOMEBREW_PREFIX}/opt/uutils-coreutils/libexec/uubin:$PATH"
[ "$BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS" -eq 1 ] && PATH="${HOMEBREW_PREFIX}/opt/uutils-findutils/libexec/uubin:$PATH"
[ "$BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS" -eq 1 ] && PATH="${HOMEBREW_PREFIX}/opt/uutils-diffutils/libexec/uubin:$PATH"

# set ATUIN_INIT_FLAGS in your ~/.bashrc before ublue-bling is sourced.
# Atuin allows these flags: "--disable-up-arrow" and/or "--disable-ctrl-r"
ATUIN_INIT_FLAGS=${ATUIN_INIT_FLAGS:-""}

# Detect shell (macOS/Linux compatible)
if [ -z "$BLING_SHELL" ]; then
    if [ -n "$BASH_VERSION" ]; then
        BLING_SHELL="bash"
    elif [ -n "$ZSH_VERSION" ]; then
        BLING_SHELL="zsh"
    else
        BLING_SHELL="$(ps -p $$ -o comm= 2>/dev/null | sed 's/^-//' | xargs basename 2>/dev/null)"
    fi
fi

# Not every prompt or history tool speaks every POSIX shell. ash (busybox --
# the default on Alpine and postmarketOS) has no init target in atuin,
# starship or carapace, so those are skipped there instead of being handed a
# shell name they reject. zoxide does support it, under the name "posix", but
# only with an explicit prompt hook: `zoxide init posix` on its own fails with
# "PWD hooks are not supported on POSIX shells", printing that at every shell
# start and leaving zoxide unloaded. Anything else keeps the shell's own name
# and zoxide's default hook, exactly as before.
# A dumb terminal cannot draw a prompt, and starship says so loudly: it exits
# with "Under a 'dumb' terminal (TERM=dumb)" printed in red, on every single
# shell start. Emacs' M-x shell sets TERM=dumb, as do a number of editor
# terminals and CI runners, so this is a real place people live rather than an
# edge case. Prompt and completion tools are skipped there; the aliases and
# PATH still apply, which is the part that works without a capable terminal.
BLUEFIN_SHELL_HAS_TERMINAL=1
case "${TERM:-dumb}" in
    dumb|"") BLUEFIN_SHELL_HAS_TERMINAL=0 ;;
esac

ZOXIDE_TARGET="$BLING_SHELL"
ZOXIDE_HOOK_ARGS=""
case "$BLING_SHELL" in
    ash|dash|sh)
        BLUEFIN_SHELL_HAS_NATIVE_INIT=0
        ZOXIDE_TARGET="posix"
        ZOXIDE_HOOK_ARGS="--hook prompt"
        ;;
    *)
        BLUEFIN_SHELL_HAS_NATIVE_INIT=1
        ;;
esac

if [ "${BLING_SHELL}" = "zsh" ]; then
    # Ensure compdef is available for zoxide/carapace.
    #
    # -i matters: a bare compinit that finds a group-writable directory on
    # fpath stops to ask the user what to do, on every shell start, and when it
    # cannot ask -- a non-interactive shell -- it prints "not interactive and
    # can't open terminal" and aborts, leaving no completions at all. -i skips
    # those directories silently, which is what the shell frameworks do, and
    # keeps the insecure ones out rather than loading them.
    if ! command -v compdef >/dev/null 2>&1; then
        autoload -Uz compinit && compinit -i
    fi
fi

# direnv goes before bash-preexec: both write PROMPT_COMMAND, and direnv's hook
# has to be in place first (https://github.com/rcaloras/bash-preexec/pull/143).
[ "$BLUEFIN_SHELL_HAS_NATIVE_INIT" -eq 1 ] && [ "$BLUEFIN_SHELL_ENABLE_DIRENV" -eq 1 ] && [ "$(command -v direnv)" ] && eval "$(direnv hook ${BLING_SHELL})"

if [ "${BLING_SHELL}" = "bash" ]; then
    [ -f "/etc/profile.d/bash-preexec.sh" ] && . "/etc/profile.d/bash-preexec.sh"
    [ -f "/usr/share/bash-prexec" ] && . "/usr/share/bash-prexec"
    [ -f "/usr/share/bash-prexec.sh" ] && . "/usr/share/bash-prexec.sh"
    [ -f "${HOMEBREW_PREFIX}/etc/profile.d/bash-preexec.sh" ] && . "${HOMEBREW_PREFIX}/etc/profile.d/bash-preexec.sh"
fi

# Initialize atuin before starship to ensure proper command history capture
# See: https://github.com/atuinsh/atuin/issues/2804 
[ "$BLUEFIN_SHELL_HAS_NATIVE_INIT" -eq 1 ] && [ "$BLUEFIN_SHELL_ENABLE_ATUIN" -eq 1 ] && [ "$(command -v atuin)" ] && eval "$(atuin init ${BLING_SHELL} ${ATUIN_INIT_FLAGS})"

[ "$BLUEFIN_SHELL_HAS_NATIVE_INIT" -eq 1 ] && [ "$BLUEFIN_SHELL_HAS_TERMINAL" -eq 1 ] && [ "$BLUEFIN_SHELL_ENABLE_STARSHIP" -eq 1 ] && [ "$(command -v starship)" ] && eval "$(starship init ${BLING_SHELL})"

[ "$BLUEFIN_SHELL_ENABLE_ZOXIDE" -eq 1 ] && [ "$(command -v zoxide)" ] && eval "$(zoxide init ${ZOXIDE_TARGET} ${ZOXIDE_HOOK_ARGS})"

[ "$BLUEFIN_SHELL_HAS_NATIVE_INIT" -eq 1 ] && [ "$BLUEFIN_SHELL_ENABLE_CARAPACE" -eq 1 ] && [ "$(command -v carapace)" ] && eval "$(carapace _carapace ${BLING_SHELL})"

# mise: MISE_BASH_AUTO_ACTIVATE=0 or MISE_ZSH_AUTO_ACTIVATE=0 keeps it out of
# that shell, as Universal Blue's bling.sh allowed.
if [ "$BLUEFIN_SHELL_ENABLE_MISE" -eq 1 ] && [ "$(command -v mise)" ]; then
    case "$BLING_SHELL" in
        bash) [ "${MISE_BASH_AUTO_ACTIVATE:-1}" != 0 ] && eval "$(mise activate bash)" ;;
        zsh) [ "${MISE_ZSH_AUTO_ACTIVATE:-1}" != 0 ] && eval "$(mise activate zsh)" ;;
    esac
fi

# Greet once per terminal: a shell started from a greeted one (bash starting
# zsh, or a login hook that already ran `bluefin-cli motd show --if-enabled`)
# inherits BLUEFIN_MOTD_SHOWN and stays quiet. Where uwelcome is installed the
# greeting is uwelcome's: its own hooks export UWELCOME_SHOWN when they greet,
# and when they have not, `motd show` runs uwelcome, so it appears once.
if [ "$BLUEFIN_SHELL_ENABLE_MOTD" -eq 1 ] && [ -z "${BLUEFIN_MOTD_SHOWN:-}" ] && [ -z "${UWELCOME_SHOWN:-}" ] && [ -n "$PS1" ] && [ -t 1 ] && [ "$(command -v bluefin-cli)" ]; then
    BLUEFIN_MOTD_SHOWN=1
    UWELCOME_SHOWN=1
    export BLUEFIN_MOTD_SHOWN UWELCOME_SHOWN
    bluefin-cli motd show --if-enabled
fi
fi
