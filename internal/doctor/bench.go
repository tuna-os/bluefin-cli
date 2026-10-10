package doctor

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/tui"
)

// BenchResult records shell startup benchmark measurements.
type BenchResult struct {
	Shell    string
	Full     time.Duration
	Base     time.Duration
	Overhead time.Duration
}

// BenchOptions configures shell startup benchmarking.
type BenchOptions struct {
	Shell    string
	LookPath func(file string) (string, error)
	Runner   func(sh string, args ...string) error
	Now      func() time.Time
	Runs     int
}

func (o BenchOptions) withDefaults() BenchOptions {
	opts := o
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.Runner == nil {
		opts.Runner = func(sh string, args ...string) error {
			err := exec.Command(sh, args...).Run()
			var exitErr *exec.ExitError
			if err != nil && !errors.As(err, &exitErr) {
				return err
			}
			return nil
		}
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Runs <= 0 {
		opts.Runs = 7
	}
	return opts
}

// RunBench measures interactive shell startup with the current rc files
// against a bare baseline.
func RunBench(opts BenchOptions) (*BenchResult, error) {
	o := opts.withDefaults()
	sh := o.Shell
	var withRC, bare []string
	switch sh {
	case "bash":
		withRC, bare = []string{"-i", "-c", "exit"}, []string{"--norc", "--noprofile", "-i", "-c", "exit"}
	case "zsh":
		withRC, bare = []string{"-i", "-c", "exit"}, []string{"-f", "-i", "-c", "exit"}
	case "fish":
		withRC, bare = []string{"-i", "-c", "exit"}, []string{"--no-config", "-i", "-c", "exit"}
	default:
		return nil, fmt.Errorf("bench supports bash, zsh, and fish (current: %s)", sh)
	}

	if _, err := o.LookPath(sh); err != nil {
		return nil, fmt.Errorf("%s not found on PATH", sh)
	}

	median := func(args []string) (time.Duration, error) {
		times := make([]time.Duration, 0, o.Runs)
		for i := 0; i < o.Runs; i++ {
			start := o.Now()
			if err := o.Runner(sh, args...); err != nil {
				return 0, err
			}
			times = append(times, o.Now().Sub(start))
		}
		sort.Slice(times, func(a, b int) bool { return times[a] < times[b] })
		return times[len(times)/2], nil
	}

	full, err := median(withRC)
	if err != nil {
		return nil, fmt.Errorf("running %s: %w", sh, err)
	}
	base, err := median(bare)
	if err != nil {
		return nil, fmt.Errorf("running bare %s: %w", sh, err)
	}
	overhead := full - base

	return &BenchResult{
		Shell:    sh,
		Full:     full,
		Base:     base,
		Overhead: overhead,
	}, nil
}

// RenderBench formats benchmark results.
func RenderBench(res *BenchResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Benchmarking %s startup (median of 7 runs)...\n", res.Shell)
	fmt.Fprintf(&b, "  your config:  %s\n", res.Full.Round(time.Millisecond))
	fmt.Fprintf(&b, "  bare shell:   %s\n", res.Base.Round(time.Millisecond))
	fmt.Fprintf(&b, "  rc overhead:  %s\n", res.Overhead.Round(time.Millisecond))
	if res.Overhead > 300*time.Millisecond {
		b.WriteString(tui.WarningStyle.Render("  Startup overhead is noticeable — try disabling components in the Shell menu.") + "\n")
	} else {
		b.WriteString(tui.SuccessStyle.Render("  Snappy.") + "\n")
	}
	return b.String()
}
