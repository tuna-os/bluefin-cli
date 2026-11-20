
# Re-arm the bash-preexec DEBUG trap on every prompt. Bash only: Init appends
# this after the shared POSIX script, whose ash and dash readers would reject
# the array syntax below even in a branch they never take.
#
# Fedora (bash >= 5.1) exposes PROMPT_COMMAND as an array. bash-preexec 0.6.0
# defers its own installation by appending a string to PROMPT_COMMAND, and its
# __bp_install only ever removes that string from element [0]. Once another
# hook (direnv, starship, mise, zoxide, vte) has pushed it into a later
# element, it runs on every prompt, clears the DEBUG trap, and every preexec
# consumer -- atuin's history recording first -- silently stops firing.
# Re-arming the trap at the end of each prompt restores the invariant
# bash-preexec assumes. Ported from Universal Blue's bling.sh:
#   https://github.com/projectbluefin/common/issues/869
#   https://github.com/rcaloras/bash-preexec/issues/188
if [[ "$(type -t __bp_preexec_invoke_exec)" == "function" ]] &&
    (unset PROMPT_COMMAND) 2>/dev/null; then
    __bluefin_rearm_bp_debug_trap() {
        trap '__bp_preexec_invoke_exec "$_"' DEBUG
    }
    if [[ "${PROMPT_COMMAND[*]-}" != *__bluefin_rearm_bp_debug_trap* ]]; then
        if ((BASH_VERSINFO[0] > 5 || (BASH_VERSINFO[0] == 5 && BASH_VERSINFO[1] >= 1))); then
            PROMPT_COMMAND+=('__bluefin_rearm_bp_debug_trap')
        elif [[ -n "${PROMPT_COMMAND:-}" ]]; then
            PROMPT_COMMAND="${PROMPT_COMMAND}"$'\n'"__bluefin_rearm_bp_debug_trap"
        else
            PROMPT_COMMAND="__bluefin_rearm_bp_debug_trap"
        fi
    fi
fi
