package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/shell"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
	"github.com/tuna-os/bluefin-cli/internal/update"

	"charm.land/lipgloss/v2"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Find common problems in your setup of Bluefin CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bench, _ := cmd.Flags().GetBool("bench"); bench {
			return runBench()
		}
		if fix, _ := cmd.Flags().GetBool("fix"); fix {
			runDoctorFixes()
		}
		report, failures := doctorReport()
		fmt.Println(report)
		if failures > 0 {
			return fmt.Errorf("%d check(s) failed", failures)
		}
		return nil
	},
}

// runDoctorFixes applies the remediations that are safe to automate:
// enabling shell integration and installing missing managed tools. The
// report afterwards shows what remains.
func runDoctorFixes() {
	current := currentShellName()
	if !shell.CheckStatus()[current] {
		fmt.Println("fix: enabling shell integration for " + current)
		if err := shell.Toggle(current, true); err != nil {
			fmt.Println("  failed: " + err.Error())
		}
	}
	if env.IsAlpine() && !commandExists("coldbrew") {
		fmt.Println("fix: setting up coldbrew")
		shell.EnsureColdbrew()
	}
	for _, tool := range []string{"eza", "fzf", "starship"} {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Println("fix: installing " + tool)
			if err := shell.EnsureInstalled(tool); err != nil {
				fmt.Printf("  failed: %v\n", err)
			}
		}
	}
}

// runBench measures interactive shell startup with the current rc files
// against a bare --norc baseline, so the cost of the shell experience is a
// number instead of a feeling.
func runBench() error {
	sh := currentShellName()
	var withRC, bare []string
	switch sh {
	case "bash":
		withRC, bare = []string{"-i", "-c", "exit"}, []string{"--norc", "--noprofile", "-i", "-c", "exit"}
	case "zsh":
		withRC, bare = []string{"-i", "-c", "exit"}, []string{"-f", "-i", "-c", "exit"}
	case "fish":
		withRC, bare = []string{"-i", "-c", "exit"}, []string{"--no-config", "-i", "-c", "exit"}
	default:
		return fmt.Errorf("bench supports bash, zsh, and fish (current: %s)", sh)
	}
	if _, err := exec.LookPath(sh); err != nil {
		return fmt.Errorf("%s not found on PATH", sh)
	}

	median := func(args []string) (time.Duration, error) {
		const runs = 7
		times := make([]time.Duration, 0, runs)
		for i := 0; i < runs; i++ {
			start := time.Now()
			err := exec.Command(sh, args...).Run()
			var exitErr *exec.ExitError
			if err != nil && !errors.As(err, &exitErr) {
				// A non-zero exit is fine (interactive shells without a TTY
				// often grumble); failing to start is not.
				return 0, err
			}
			times = append(times, time.Since(start))
		}
		sort.Slice(times, func(a, b int) bool { return times[a] < times[b] })
		return times[len(times)/2], nil
	}

	fmt.Printf("Benchmarking %s startup (median of 7 runs)...\n", sh)
	full, err := median(withRC)
	if err != nil {
		return fmt.Errorf("running %s: %w", sh, err)
	}
	base, err := median(bare)
	if err != nil {
		return fmt.Errorf("running bare %s: %w", sh, err)
	}
	overhead := full - base
	fmt.Printf("  your config:  %s\n", full.Round(time.Millisecond))
	fmt.Printf("  bare shell:   %s\n", base.Round(time.Millisecond))
	fmt.Printf("  rc overhead:  %s\n", overhead.Round(time.Millisecond))
	if overhead > 300*time.Millisecond {
		fmt.Println(tui.WarningStyle.Render("  Startup overhead is noticeable — try disabling components in the Shell menu."))
	} else {
		fmt.Println(tui.SuccessStyle.Render("  Snappy."))
	}
	return nil
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Apply safe automatic fixes before reporting")
	doctorCmd.Flags().Bool("bench", false, "Benchmark interactive shell startup vs a bare shell")
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	ok   bool
	warn bool
	name string
	note string // failure detail or fix hint
}

// doctorReport runs all checks and renders the results; failures counts
// hard failures (warnings excluded).
func doctorReport() (string, int) {
	t := theme.DefaultTheme
	pass := lipgloss.NewStyle().Foreground(t.Success).Render("✓")
	warn := lipgloss.NewStyle().Foreground(t.Warning).Render("!")
	fail := lipgloss.NewStyle().Foreground(t.Error).Render("✗")
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	checks := []checkResult{
		checkBrew(),
		checkShellIntegration(),
		checkTools(),
		checkNetwork(),
		checkVersion(),
		checkSelfUpdate(),
	}

	var b strings.Builder
	failures := 0
	b.WriteString("\n")
	for _, c := range checks {
		mark := pass
		switch {
		case !c.ok && c.warn:
			mark = warn
		case !c.ok:
			mark = fail
			failures++
		}
		fmt.Fprintf(&b, "  %s %s\n", mark, c.name)
		if c.note != "" {
			fmt.Fprintf(&b, "    %s\n", dim.Render(c.note))
		}
	}
	if failures == 0 {
		b.WriteString("\n" + dim.Render("  All checks passed."))
	}
	return b.String(), failures
}

func checkBrew() checkResult {
	// Homebrew requires glibc (its portable Ruby bootstrap can't even start
	// on musl) and isn't the right answer here even if a stray `brew` shim
	// is on PATH from another system's dotfiles — check package manager
	// fitness first, not just binary presence.
	if env.IsAlpine() {
		switch {
		case commandExists("coldbrew"):
			return checkResult{ok: true, name: "Package manager: coldbrew"}
		case commandExists("apk"):
			return checkResult{name: "Package manager: coldbrew not set up yet", warn: true,
				note: "run 'bluefin-cli doctor --fix' to install it (rootless, sandboxed); falls back to 'sudo apk add' per package until then"}
		default:
			return checkResult{name: "Package manager", warn: true,
				note: "neither coldbrew nor apk found — shell tool installs will be skipped"}
		}
	}
	if _, err := exec.LookPath("brew"); err == nil {
		return checkResult{ok: true, name: "Homebrew on PATH"}
	}
	// The classic macOS papercut: brew is installed but the shell was never
	// taught about it.
	for _, p := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew", "/home/linuxbrew/.linuxbrew/bin/brew"} {
		if _, err := os.Stat(p); err == nil {
			return checkResult{name: "Homebrew on PATH", warn: true,
				note: fmt.Sprintf("brew is installed at %s but not on PATH — add: eval \"$(%s shellenv)\"", p, p)}
		}
	}
	return checkResult{name: "Homebrew on PATH", warn: true,
		note: "brew not found — Install Apps bundles need it: https://brew.sh"}
}

func commandExists(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func checkShellIntegration() checkResult {
	current := currentShellName()
	if shell.CheckStatus()[current] {
		return checkResult{ok: true, name: fmt.Sprintf("Shell integration enabled (%s)", current)}
	}
	return checkResult{name: fmt.Sprintf("Shell integration enabled (%s)", current), warn: true,
		note: fmt.Sprintf("run: bluefin-cli shell %s on", current)}
}

func checkTools() checkResult {
	missing := []string{}
	for _, tool := range []string{"eza", "fzf", "starship"} {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) == 0 {
		return checkResult{ok: true, name: "Managed tools installed (eza, fzf, starship)"}
	}
	return checkResult{name: "Managed tools installed", warn: true,
		note: fmt.Sprintf("missing: %v — install via the Bluefin Shell menu", missing)}
}

func checkNetwork() checkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head("https://api.github.com")
	if err != nil {
		return checkResult{name: "GitHub reachable", warn: true,
			note: "updates and bundle installs need network access: " + err.Error()}
	}
	_ = resp.Body.Close()
	return checkResult{ok: true, name: "GitHub reachable"}
}

// checkSelfUpdate verifies a direct install can actually replace its own
// binary (package-manager installs update through the manager instead).
func checkSelfUpdate() checkResult {
	if update.Detect() != update.MethodDirect {
		return checkResult{ok: true, name: "Updates managed by " + string(update.Detect())}
	}
	exe, err := os.Executable()
	if err != nil {
		return checkResult{name: "Self-update viable", warn: true, note: err.Error()}
	}
	dir := filepath.Dir(exe)
	f, err := os.CreateTemp(dir, ".bluefin-doctor-*")
	if err != nil {
		return checkResult{name: "Self-update viable", warn: true,
			note: dir + " is not writable — 'bluefin-cli update' would need elevated permissions"}
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return checkResult{ok: true, name: "Self-update viable (install dir writable)"}
}

func checkVersion() checkResult {
	if version == "dev" {
		return checkResult{ok: true, name: "Version: dev build"}
	}
	rel, err := update.Latest()
	if err != nil {
		return checkResult{ok: true, name: "Version " + version,
			note: "could not check for updates: " + err.Error()}
	}
	if update.IsNewer(version, rel.TagName) {
		hint := "bluefin-cli update"
		if h := update.Detect().UpdateHint(); h != "" {
			hint = h
		}
		return checkResult{name: "Version " + version, warn: true,
			note: fmt.Sprintf("%s available — run: %s", rel.TagName, hint)}
	}
	return checkResult{ok: true, name: "Version " + version + " (latest)"}
}
