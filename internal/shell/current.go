package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

// "Which shell is the user in?" has two answers, and callers must pick the
// one they mean: the login shell ($SHELL) and the shell that started this
// process. They differ whenever the user runs a shell other than their login
// shell. Both rules live here so that no command grows a third.

// runningNameTimeout bounds the parent-process lookup.
const runningNameTimeout = 1500 * time.Millisecond

// LoginName returns the base name of the login shell from $SHELL. When $SHELL
// is unset it falls back to "powershell" on Windows and "bash" elsewhere.
func LoginName() string {
	name := filepath.Base(os.Getenv("SHELL"))
	if name == "" || name == "." {
		if env.IsWindows() {
			return "powershell"
		}
		return "bash"
	}
	return name
}

// RunningName returns the command name of this process's parent, which is
// the shell the user typed the command into. It returns "" when the parent
// cannot be determined, including where ps is unavailable.
func RunningName() string {
	ppid := os.Getppid()
	if ppid <= 0 {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), runningNameTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=").Output()
	if err != nil {
		return ""
	}
	return processShellName(string(out))
}

// processShellName normalises a ps comm field: a full path (/bin/zsh) or a
// login-shell name (-zsh) both become "zsh".
func processShellName(comm string) string {
	comm = strings.TrimPrefix(strings.TrimSpace(comm), "-")
	if comm == "" {
		return ""
	}
	return filepath.Base(comm)
}

// ActiveName returns the shell the user is actually in: the running shell
// when it is a managed shell, otherwise the login shell. This is the rule
// status uses for its "current" marker; doctor acts on the same shell.
func ActiveName() string {
	return pickActiveName(RunningName(), LoginName())
}

func pickActiveName(running, login string) string {
	if _, ok := LookupShell(running); ok || isPowerShellShell(running) {
		return NormalizeShell(running)
	}
	return login
}
