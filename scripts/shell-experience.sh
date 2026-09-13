#!/usr/bin/env bash
# Behavioral end-to-end test of the shell experience, per shell.
#
# scripts/tui-smoke.sh asserts what is on screen and scripts/tui-state.sh
# asserts what lands in config files. Neither runs the init script that every
# user's shell actually evaluates at startup, and the CI syntax job only parses
# it. Parsing is a weak gate: it cannot tell you that `zoxide init posix` exits
# with "PWD hooks are not supported on POSIX shells", or that a bare `return`
# is an error at the top level of an `eval` -- both of which parsed fine and
# shipped.
#
# So this starts each shell for real, evaluates the init the same way the
# documented rc line does, and asserts the resulting experience: the aliases
# exist, the integrations loaded, PATH carries the uutils directories, and
# nothing was printed to stderr on the way.
#
# Usage: scripts/shell-experience.sh <binary> [shell ...]
set -u

BIN=$(cd "$(dirname "$1")" && pwd)/$(basename "${1:?usage: shell-experience.sh <binary> [shell ...]}")
shift || true

fail=0
ok()   { printf 'ok   %s\n' "$1"; }
bad()  { printf 'FAIL %s\n' "$1"; fail=1; }
skip() { printf 'skip %s\n' "$1"; }

# Each shell gets its own sandbox HOME so a real config never leaks in.
SANDBOX=$(mktemp -d)
trap 'rm -rf "$SANDBOX"' EXIT

# have <tool> -- is the tool on PATH? The assertions below are conditional on
# this, so the suite states a real invariant ("an alias exists exactly when its
# tool does") rather than requiring a fully provisioned machine.
have() { command -v "$1" >/dev/null 2>&1; }

# run_in <shell> <script> -- evaluate the init the way that shell's documented
# rc line does, then run <script>. stdout and stderr are captured separately so
# a clean startup can be asserted.
run_in() {
    _shell=$1; _probe=$2
    case "$_shell" in
      bash)  HOME=$SANDBOX bash -c "eval \"\$($BIN init bash 2>/dev/null)\"; $_probe" ;;
      zsh)   HOME=$SANDBOX zsh  -c "eval \"\$($BIN init zsh 2>/dev/null)\"; $_probe" ;;
      ash)   HOME=$SANDBOX busybox ash -c "eval \"\$($BIN init ash 2>/dev/null)\"; $_probe" ;;
      dash)  HOME=$SANDBOX dash -c "eval \"\$($BIN init ash 2>/dev/null)\"; $_probe" ;;
      # fish gates its tool initialization on `status is-interactive`, which is
      # right -- a script has no business starting a prompt -- so the probes
      # have to run in an interactive fish or every integration looks missing.
      fish)  HOME=$SANDBOX fish -i -c "$BIN init fish 2>/dev/null | source; $_probe" ;;
      nu)    HOME=$SANDBOX $BIN init nu >"$SANDBOX/init.nu" 2>/dev/null
             HOME=$SANDBOX nu -c "source $SANDBOX/init.nu; $_probe" ;;
    esac
}

# Per-shell probe syntax. Keeping these in one place is what lets the assertion
# bodies below stay shell-agnostic.
probe_alias() {   # is <name> defined as an alias/function?
    case "$1" in
      fish) printf 'functions -q %s; and echo YES' "$2" ;;
      nu)   printf 'if (scope aliases | where name == "%s" | is-not-empty) { print "YES" }' "$2" ;;
      *)    printf 'alias %s >/dev/null 2>&1 && echo YES' "$2" ;;
    esac
}
probe_var() {     # print the value of <name>
    case "$1" in
      fish) printf 'echo $%s' "$2" ;;
      nu)   printf 'print $env.%s' "$2" ;;
      *)    printf 'echo $%s' "$2" ;;
    esac
}
probe_cmd() {     # is <name> a defined command/function?
    case "$1" in
      fish) printf 'functions -q %s; or type -q %s; and echo YES' "$2" "$2" ;;
      nu)   printf 'if (which %s | is-not-empty) { print "YES" }' "$2" ;;
      *)    printf 'command -v %s >/dev/null 2>&1 && echo YES' "$2" ;;
    esac
}
probe_prompt() { # YES when the shell's prompt is driven by starship
    case "$1" in
      *) printf 'case "$PS1$PROMPT$PROMPT_COMMAND" in *starship*) echo YES ;; esac' ;;
    esac
}
probe_path() {    # print PATH
    case "$1" in
      fish) printf 'echo $PATH' ;;
      nu)   printf 'print ($env.PATH | str join ":")' ;;
      *)    printf 'echo $PATH' ;;
    esac
}

# uutils_enabled <shell> -- 1 when the generated init turns uutils-coreutils on.
uutils_enabled() {
    # Anchored to the generated preamble: each script also carries a
    # "default to 1 if unset" block further down that mentions the same
    # variable, and an unanchored match reads that as "enabled".
    case "$1" in
      nu)   $BIN init nu   2>/dev/null | grep -qx '\$env.BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS = "1"' && echo 1 || echo 0 ;;
      fish) $BIN init fish 2>/dev/null | grep -qx 'set -gx BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS 1'    && echo 1 || echo 0 ;;
      *)    $BIN init ash  2>/dev/null | grep -qx 'export BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS=1'     && echo 1 || echo 0 ;;
    esac
}

available() {
    case "$1" in
      ash) command -v busybox >/dev/null 2>&1 ;;
      nu)  command -v nu >/dev/null 2>&1 ;;
      *)   command -v "$1" >/dev/null 2>&1 ;;
    esac
}

test_shell() {
    sh_name=$1
    printf '\n── %s ──\n' "$sh_name"

    if ! available "$sh_name"; then
        skip "$sh_name is not installed"
        return
    fi

    # 1. A clean start prints nothing to stderr. This is the assertion that
    #    catches a tool rejecting the arguments we hand it: zoxide's POSIX
    #    complaint and bash's `return` error both showed up here and nowhere
    #    else, because both leave exit status 0.
    err=$(run_in "$sh_name" "true" 2>&1 >/dev/null)
    if [ -n "$err" ]; then
        bad "$sh_name: startup wrote to stderr: $err"
    else
        ok "$sh_name: startup writes nothing to stderr"
    fi

    #    stdout matters just as much and is easy to miss: zoxide's POSIX
    #    complaint goes to stdout, so a stderr-only check calls it quiet while
    #    the user gets a warning printed over their first prompt.
    out=$(run_in "$sh_name" "true" 2>/dev/null)
    if [ -n "$out" ]; then
        bad "$sh_name: startup wrote to stdout: $out"
    else
        ok "$sh_name: startup writes nothing to stdout"
    fi

    # 2. The init identifies the shell to itself, which is what the script uses
    #    to pick each tool's init target.
    got=$(run_in "$sh_name" "$(probe_var "$sh_name" BLING_SHELL)" 2>/dev/null | tr -d '\r')
    want=$sh_name
    [ "$sh_name" = "dash" ] && want=ash
    if [ "$got" = "$want" ]; then
        ok "$sh_name: BLING_SHELL=$got"
    else
        bad "$sh_name: BLING_SHELL=$got, want $want"
    fi

    # 3. An alias exists exactly when the tool backing it exists. A stale alias
    #    pointing at a missing binary is worse than no alias: it shadows the
    #    real command and fails on use.
    for pair in "eza:ll" "eza:ls" "eza:l1" "ug:grep"; do
        tool=${pair%%:*}; name=${pair#*:}
        [ "$sh_name" = "nu" ] && [ "$name" = "l1" ] && continue   # nu keeps ll/ls only
        got=$(run_in "$sh_name" "$(probe_alias "$sh_name" "$name")" 2>/dev/null | tr -d '\r')
        if have "$tool"; then
            [ "$got" = "YES" ] && ok "$sh_name: $name aliased ($tool present)" \
                               || bad "$sh_name: $name missing although $tool is installed"
        else
            [ "$got" = "YES" ] && bad "$sh_name: $name aliased but $tool is NOT installed" \
                               || ok "$sh_name: $name correctly absent ($tool missing)"
        fi
    done

    # 4. zoxide is the one integration every supported shell can load, so it is
    #    the check that the tool-init section ran rather than silently failing.
    if have zoxide; then
        got=$(run_in "$sh_name" "$(probe_cmd "$sh_name" __zoxide_z)" 2>/dev/null | tr -d '\r')
        [ "$got" = "YES" ] && ok "$sh_name: zoxide initialized" \
                           || bad "$sh_name: zoxide is installed but did not initialize"
    else
        skip "$sh_name: zoxide not installed"
    fi

    # 4b. starship is the prompt, and the one integration whose absence a user
    #     notices immediately. It has no ash/dash target, so on those shells the
    #     contract is the opposite: it must be skipped deliberately rather than
    #     handed a shell name it rejects. Both halves are asserted, because
    #     "not initialized" is only correct for the shells that cannot take it.
    if have starship; then
        # Asserted through the prompt itself rather than an internal symbol:
        # starship names its hooks differently per shell (starship_precmd on
        # bash, prompt_starship_precmd on zsh) and renames them between
        # releases, but "the prompt runs starship" is the thing the user
        # actually gets and does not churn.
        got=$(run_in "$sh_name" "$(probe_prompt "$sh_name")" 2>/dev/null | tr -d '\r')
        case "$sh_name" in
          ash|dash)
            [ "$got" = "YES" ] && bad "$sh_name: starship initialized, but it has no $sh_name target" \
                               || ok "$sh_name: starship correctly skipped (no $sh_name target)" ;;
          fish)
            # fish's starship integration replaces fish_prompt rather than
            # defining starship_precmd.
            got=$(run_in "$sh_name" "functions -q fish_prompt; and functions fish_prompt | grep -q starship; and echo YES" 2>/dev/null | tr -d '\r')
            [ "$got" = "YES" ] && ok "$sh_name: starship drives fish_prompt" \
                               || bad "$sh_name: starship is installed but did not take over fish_prompt" ;;
          nu)
            got=$(run_in "$sh_name" 'if (($nu.data-dir | path join "vendor" "autoload" "starship.nu") | path exists) { print "YES" }' 2>/dev/null | tr -d '\r')
            [ "$got" = "YES" ] && ok "$sh_name: starship written to the vendor autoload directory" \
                               || bad "$sh_name: starship is installed but no autoload script was written" ;;
          *)
            [ "$got" = "YES" ] && ok "$sh_name: starship initialized" \
                               || bad "$sh_name: starship is installed but did not initialize" ;;
        esac
    else
        skip "$sh_name: starship not installed"
    fi

    # 5. uutils ship in a libexec directory that is not on PATH by default, so
    #    an enabled uutils has to be prepended there or its binaries are
    #    unreachable. `init` disables tools it cannot find, so the invariant is
    #    conditional: the directory is on PATH exactly when the generated script
    #    turned the tool on.
    want_uutils=$(uutils_enabled "$sh_name")
    got=$(run_in "$sh_name" "$(probe_path "$sh_name")" 2>/dev/null)
    case "$got" in
      *uutils-coreutils/libexec/uubin*) have_uutils=1 ;;
      *) have_uutils=0 ;;
    esac
    if [ "$want_uutils" = "$have_uutils" ]; then
        [ "$want_uutils" = 1 ] && ok "$sh_name: uutils enabled and on PATH" \
                               || ok "$sh_name: uutils disabled and correctly absent from PATH"
    elif [ "$want_uutils" = 1 ]; then
        bad "$sh_name: uutils is enabled but its libexec directory is not on PATH"
    else
        bad "$sh_name: uutils is disabled but its libexec directory was added to PATH"
    fi
}

# 6. Loading twice must be quiet and must not run the body again -- re-running
#    re-initializes atuin, which is the bug the load-once guard exists for.
#    Only the POSIX shells carry that guard today.
test_double_load() {
    sh_name=$1
    available "$sh_name" || return
    case "$sh_name" in bash|zsh|ash|dash) ;; *) return ;; esac

    tag=$($BIN init "$([ "$sh_name" = dash ] && echo ash || echo "$sh_name")" 2>/dev/null >/dev/null; echo ok)
    [ "$tag" = ok ] || return

    case "$sh_name" in
      bash) sh_bin="bash"; arg=bash ;;
      zsh)  sh_bin="zsh";  arg=zsh ;;
      ash)  sh_bin="busybox ash"; arg=ash ;;
      dash) sh_bin="dash"; arg=ash ;;
    esac

    err=$(HOME=$SANDBOX $sh_bin -c "
        eval \"\$($BIN init $arg 2>/dev/null)\"
        eval \"\$($BIN init $arg 2>/dev/null)\"
    " 2>&1 >/dev/null)
    if [ -n "$err" ]; then
        bad "$sh_name: loading twice wrote to stderr: $err"
    else
        ok "$sh_name: loading twice is quiet"
    fi

    # The guard has to actually skip the body, not merely survive it. Removing
    # an alias and reloading proves which: if the body re-runs, it comes back.
    if have eza; then
        got=$(HOME=$SANDBOX $sh_bin -c "
            eval \"\$($BIN init $arg 2>/dev/null)\"
            unalias ll 2>/dev/null
            eval \"\$($BIN init $arg 2>/dev/null)\"
            alias ll >/dev/null 2>&1 && echo RERAN
        " 2>/dev/null | tr -d '\r')
        [ "$got" = "RERAN" ] && bad "$sh_name: the load-once guard did not stop the body re-running" \
                             || ok "$sh_name: load-once guard held"
    fi
}

SHELLS=${*:-"bash zsh ash dash fish nu"}
for s in $SHELLS; do test_shell "$s"; done
printf '\n── load-once guard ──\n'
for s in $SHELLS; do test_double_load "$s"; done

printf '\n'
if [ "$fail" -eq 0 ]; then
    echo "Shell experience: all checks passed"
else
    echo "Shell experience: FAILURES"
fi
exit "$fail"
