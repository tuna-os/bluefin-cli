package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// FixOptions configures automatic remediation.
type FixOptions struct {
	CurrentShell    string
	IsAlpine        func() bool
	CommandExists   func(bin string) bool
	CheckStatus     func() map[string]bool
	ToggleShell     func(sh string, enable bool) error
	EnsureColdbrew  func() bool
	EnsureInstalled func(tool string) error
	Output          io.Writer
}

func (o FixOptions) withDefaults() FixOptions {
	opts := o
	if opts.IsAlpine == nil {
		opts.IsAlpine = env.IsAlpine
	}
	if opts.CommandExists == nil {
		opts.CommandExists = func(bin string) bool {
			_, err := exec.LookPath(bin)
			return err == nil
		}
	}
	if opts.CheckStatus == nil {
		opts.CheckStatus = shell.CheckStatus
	}
	if opts.ToggleShell == nil {
		opts.ToggleShell = shell.Toggle
	}
	if opts.EnsureColdbrew == nil {
		opts.EnsureColdbrew = shell.EnsureColdbrew
	}
	if opts.EnsureInstalled == nil {
		opts.EnsureInstalled = shell.EnsureInstalled
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	return opts
}

// RunFixes applies the remediations that are safe to automate:
// enabling shell integration and installing missing managed tools.
func RunFixes(opts FixOptions) {
	o := opts.withDefaults()
	current := o.CurrentShell
	status := o.CheckStatus()
	if status == nil || !status[current] {
		_, _ = fmt.Fprintln(o.Output, "fix: enabling shell integration for "+current)
		if err := o.ToggleShell(current, true); err != nil {
			_, _ = fmt.Fprintln(o.Output, "  failed: "+err.Error())
		}
	}
	if o.IsAlpine() && !o.CommandExists("coldbrew") {
		_, _ = fmt.Fprintln(o.Output, "fix: setting up coldbrew")
		_ = o.EnsureColdbrew()
	}
	for _, tool := range []string{"eza", "fzf", "starship"} {
		if !o.CommandExists(tool) {
			_, _ = fmt.Fprintln(o.Output, "fix: installing "+tool)
			if err := o.EnsureInstalled(tool); err != nil {
				_, _ = fmt.Fprintf(o.Output, "  failed: %v\n", err)
			}
		}
	}
}
