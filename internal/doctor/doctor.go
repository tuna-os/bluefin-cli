package doctor

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
	"github.com/tuna-os/bluefin-cli/internal/update"
)

// CheckResult represents the outcome of a single diagnostic probe.
type CheckResult struct {
	OK   bool   `json:"ok"`
	Warn bool   `json:"warn"`
	Name string `json:"name"`
	Note string `json:"note,omitempty"` // failure detail or fix hint
}

// Report holds the outcome of running all diagnostic checks.
type Report struct {
	Checks   []CheckResult `json:"checks"`
	Failures int           `json:"failures"`
}

// Options configures diagnostic collection with injectable seams for testing.
type Options struct {
	CurrentShell   string
	Version        string
	LookPath       func(file string) (string, error)
	Stat           func(name string) (os.FileInfo, error)
	IsAlpine       func() bool
	CheckNetwork   func() error
	ShellStatus    func() map[string]bool
	DetectUpdate   func() update.InstallMethod
	Executable     func() (string, error)
	CreateTempFile func(dir, pattern string) (*os.File, error)
	RemoveTempFile func(name string) error
	LatestRelease  func() (*update.Release, error)
	IsNewer        func(current, latest string) bool
	UpdateHint     func(method update.InstallMethod) string
}

func (o Options) withDefaults() Options {
	opts := o
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.Stat == nil {
		opts.Stat = os.Stat
	}
	if opts.IsAlpine == nil {
		opts.IsAlpine = env.IsAlpine
	}
	if opts.CheckNetwork == nil {
		opts.CheckNetwork = func() error {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Head("https://api.github.com")
			if err != nil {
				return err
			}
			_ = resp.Body.Close()
			return nil
		}
	}
	if opts.ShellStatus == nil {
		opts.ShellStatus = shell.CheckStatus
	}
	if opts.DetectUpdate == nil {
		opts.DetectUpdate = update.Detect
	}
	if opts.Executable == nil {
		opts.Executable = os.Executable
	}
	if opts.CreateTempFile == nil {
		opts.CreateTempFile = os.CreateTemp
	}
	if opts.RemoveTempFile == nil {
		opts.RemoveTempFile = os.Remove
	}
	if opts.LatestRelease == nil {
		opts.LatestRelease = update.Latest
	}
	if opts.IsNewer == nil {
		opts.IsNewer = update.IsNewer
	}
	if opts.UpdateHint == nil {
		opts.UpdateHint = func(m update.InstallMethod) string {
			return m.UpdateHint()
		}
	}
	return opts
}

// Diagnose executes all standard diagnostic probes and compiles a Report.
func Diagnose(opts Options) Report {
	o := opts.withDefaults()

	checks := []CheckResult{
		CheckBrew(o),
		CheckShellIntegration(o),
		CheckTools(o),
		CheckNetwork(o),
		CheckVersion(o),
		CheckSelfUpdate(o),
	}

	failures := 0
	for _, c := range checks {
		if !c.OK && !c.Warn {
			failures++
		}
	}

	return Report{
		Checks:   checks,
		Failures: failures,
	}
}

// Render formats the report into a styled terminal string.
func (r Report) Render() string {
	t := theme.DefaultTheme
	pass := lipgloss.NewStyle().Foreground(t.Success).Render("✓")
	warn := lipgloss.NewStyle().Foreground(t.Warning).Render("!")
	fail := lipgloss.NewStyle().Foreground(t.Error).Render("✗")
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	var b strings.Builder
	b.WriteString("\n")
	for _, c := range r.Checks {
		mark := pass
		switch {
		case !c.OK && c.Warn:
			mark = warn
		case !c.OK:
			mark = fail
		}
		fmt.Fprintf(&b, "  %s %s\n", mark, c.Name)
		if c.Note != "" {
			fmt.Fprintf(&b, "    %s\n", dim.Render(c.Note))
		}
	}
	if r.Failures == 0 {
		b.WriteString("\n" + dim.Render("  All checks passed."))
	}
	return b.String()
}
