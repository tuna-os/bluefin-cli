#!/usr/bin/env fish

# Source the configuration environment file if it exists


# Default to enabled if variable is not set (backwards compatibility)
if not set -q BLUEFIN_SHELL_ENABLE_EZA
    set BLUEFIN_SHELL_ENABLE_EZA 1
end
if not set -q BLUEFIN_SHELL_ENABLE_UGREP
    set BLUEFIN_SHELL_ENABLE_UGREP 1
end
if not set -q BLUEFIN_SHELL_ENABLE_BAT
    set BLUEFIN_SHELL_ENABLE_BAT 1
end
if not set -q BLUEFIN_SHELL_ENABLE_ATUIN
    set BLUEFIN_SHELL_ENABLE_ATUIN 1
end
if not set -q BLUEFIN_SHELL_ENABLE_STARSHIP
    set BLUEFIN_SHELL_ENABLE_STARSHIP 1
end
if not set -q BLUEFIN_SHELL_ENABLE_ZOXIDE
    set BLUEFIN_SHELL_ENABLE_ZOXIDE 1
end
if not set -q BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS
    set BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS 1
end
if not set -q BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS
    set BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS 1
end
if not set -q BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS
    set BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS 1
end
if not set -q BLUEFIN_SHELL_ENABLE_CARAPACE
    set BLUEFIN_SHELL_ENABLE_CARAPACE 0
end
if not set -q BLUEFIN_SHELL_ENABLE_DIRENV
    set BLUEFIN_SHELL_ENABLE_DIRENV 0
end
if not set -q BLUEFIN_SHELL_ENABLE_MISE
    set BLUEFIN_SHELL_ENABLE_MISE 0
end
if not set -q BLUEFIN_SHELL_ENABLE_MOTD
    set BLUEFIN_SHELL_ENABLE_MOTD 1
end

# ls aliases
if test "$BLUEFIN_SHELL_ENABLE_EZA" -eq 1; and type -q eza
    alias ll='eza -l --icons=auto --group-directories-first'
    alias l.='eza -d .*'
    alias ls='eza'
    alias l1='eza -1'
end

# ugrep for grep
if test "$BLUEFIN_SHELL_ENABLE_UGREP" -eq 1; and type -q ug
    alias grep='ug'
    alias egrep='ug -E'
    alias fgrep='ug -F'
    alias xzgrep='ug -z'
    alias xzegrep='ug -zE'
    alias xzfgrep='ug -zF'
end

# bat for cat, only when bat is there: a cached init script can outlive it
if test "$BLUEFIN_SHELL_ENABLE_BAT" -eq 1; and type -q bat
    alias cat='bat --style=plain --pager=never'
end

# open -> xdg-open, where no `open` command exists already
if type -q xdg-open; and not test -e /usr/bin/open
    function open --wraps xdg-open --description 'Open files and URLs with xdg-open'
        xdg-open $argv >/dev/null 2>&1
    end
end

if not set -q HOMEBREW_PREFIX
    if test -d "/opt/homebrew"
        set HOMEBREW_PREFIX "/opt/homebrew"
    else if test -d "/usr/local/Homebrew"
        set HOMEBREW_PREFIX "/usr/local"
    else
        set HOMEBREW_PREFIX "/home/linuxbrew/.linuxbrew"
    end
end

# uutils
test "$BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS" -eq 1; and fish_add_path --prepend "$HOMEBREW_PREFIX/opt/uutils-coreutils/libexec/uubin"
test "$BLUEFIN_SHELL_ENABLE_UUTILSFINDUTILS" -eq 1; and fish_add_path --prepend "$HOMEBREW_PREFIX/opt/uutils-findutils/libexec/uubin"
test "$BLUEFIN_SHELL_ENABLE_UUTILSDIFFUTILS" -eq 1; and fish_add_path --prepend "$HOMEBREW_PREFIX/opt/uutils-diffutils/libexec/uubin" 

if status is-interactive
    if test "$BLUEFIN_SHELL_ENABLE_DIRENV" -eq 1; and type -q direnv
        direnv hook fish | source
    end

    # Initialize atuin before starship to ensure proper command history capture
    # Atuin allows these flags: "--disable-up-arrow" and/or "--disable-ctrl-r"
    # Use by setting a universal variable, e.g. set -U ATUIN_INIT_FLAGS "--disable-up-arrow"
    # Or set in config.fish before this file is sourced
    if test "$BLUEFIN_SHELL_ENABLE_ATUIN" -eq 1; and type -q atuin
        atuin init fish $ATUIN_INIT_FLAGS | source
    end

    if test "$BLUEFIN_SHELL_ENABLE_STARSHIP" -eq 1; and type -q starship
        starship init fish | source
    end

    if test "$BLUEFIN_SHELL_ENABLE_ZOXIDE" -eq 1; and type -q zoxide
        zoxide init fish | source
    end

    if test "$BLUEFIN_SHELL_ENABLE_CARAPACE" -eq 1; and type -q carapace
        set -x CARAPACE_BRIDGES 'zsh,fish,bash,inshellisense'
        carapace _carapace | source
    end

    # MISE_FISH_AUTO_ACTIVATE=0 keeps mise out of fish, as bling.fish allowed
    if test "$BLUEFIN_SHELL_ENABLE_MISE" -eq 1; and type -q mise; and test "$MISE_FISH_AUTO_ACTIVATE" != 0
        mise activate fish | source
    end

    if test "$BLUEFIN_SHELL_ENABLE_MOTD" -eq 1; and type -q bluefin-cli
        bluefin-cli motd show
    end
end
