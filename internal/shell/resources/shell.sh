#!/usr/bin/env sh

# Check if Bluefin shell has already been sourced so that we dont break atuin. https://github.com/atuinsh/atuin/issues/380#issuecomment-1594014644
[ "${SOURCED_BLUEFIN_SHELL:-0}" -eq 1 ] && return 
SOURCED_BLUEFIN_SHELL=1

# Default to enabled if variable is not set (backwards compatibility)
: "${BLUEFIN_SHELL_ENABLE_EZA:=1}"
: "${BLUEFIN_SHELL_ENABLE_UGREP:=1}"
: "${BLUEFIN_SHELL_ENABLE_BAT:=1}"
: "${BLUEFIN_SHELL_ENABLE_ATUIN:=1}"
: "${BLUEFIN_SHELL_ENABLE_STARSHIP:=1}"
: "${BLUEFIN_SHELL_ENABLE_ZOXIDE:=1}"
: "${BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS:=1}"
: "${BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS:=1}"
: "${BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS:=1}"
: "${BLUEFIN_SHELL_ENABLE_MOTD:=1}"

# Default disabled tools
: "${BLUEFIN_SHELL_ENABLE_CARAPACE:=0}"

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

# bat for cat
if [ "$BLUEFIN_SHELL_ENABLE_BAT" -eq 1 ]; then
    alias cat='bat --style=plain --pager=never' 2>/dev/null
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

# Detect shell
[ -n "$BASH_VERSION" ] && BLING_SHELL="bash"
[ -n "$ZSH_VERSION" ] && BLING_SHELL="zsh"

if [ "${BLING_SHELL}" = "bash" ]; then
    [ -f "/etc/profile.d/bash-preexec.sh" ] && . "/etc/profile.d/bash-preexec.sh"
    [ -f "/usr/share/bash-prexec" ] && . "/usr/share/bash-prexec"
    [ -f "/usr/share/bash-prexec.sh" ] && . "/usr/share/bash-prexec.sh"
    [ -f "${HOMEBREW_PREFIX}/etc/profile.d/bash-preexec.sh" ] && . "${HOMEBREW_PREFIX}/etc/profile.d/bash-preexec.sh"
fi

# Initialize atuin before starship to ensure proper command history capture
# See: https://github.com/atuinsh/atuin/issues/2804 
[ "$BLUEFIN_SHELL_ENABLE_ATUIN" -eq 1 ] && [ "$(command -v atuin)" ] && eval "$(atuin init ${BLING_SHELL} ${ATUIN_INIT_FLAGS})"

[ "$BLUEFIN_SHELL_ENABLE_STARSHIP" -eq 1 ] && [ "$(command -v starship)" ] && eval "$(starship init ${BLING_SHELL})"

[ "$BLUEFIN_SHELL_ENABLE_ZOXIDE" -eq 1 ] && [ "$(command -v zoxide)" ] && eval "$(zoxide init ${BLING_SHELL})"

[ "$BLUEFIN_SHELL_ENABLE_CARAPACE" -eq 1 ] && [ "$(command -v carapace)" ] && eval "$(carapace _carapace ${BLING_SHELL})"

if [ "$BLUEFIN_SHELL_ENABLE_MOTD" -eq 1 ] && [ -n "$PS1" ] && [ -t 1 ] && [ "$(command -v bluefin-cli)" ]; then
    bluefin-cli motd show
fi



