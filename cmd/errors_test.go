package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"
)

// Run a fresh Fang/Cobra instance in a subprocess, so Fang sees the same
// colorprofile-wrapped stderr as the real application.
func TestFangErrorHelper(t *testing.T) {
	kind := os.Getenv("BLUEFIN_TEST_ERROR")
	if kind == "" {
		return
	}
	var err error
	switch kind {
	case "plain":
		err = errors.New("--clear failed.")
	case "multiline":
		err = errors.New("/tmp/MyPath failed\n\ntry --clear instead")
	case "crlf":
		err = errors.New("first\r\nsecond\r\n")
	case "joined":
		err = errors.Join(errors.New("first failed."), errors.New("--clear failed!"))
	case "domain":
		err = errors.New("accepts only --clear")
	default:
		panic("unknown test error")
	}
	root := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return err }}
	if executeWithFang(root) != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestFangErrorRendering(t *testing.T) {
	cases := []struct{ kind, message string }{
		{"plain", "--clear failed."},
		{"multiline", "/tmp/MyPath failed\n\ntry --clear instead"},
		{"crlf", "first\r\nsecond\r\n"},
		{"joined", "first failed.\n--clear failed!"},
		{"domain", "accepts only --clear"},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			env := []string{"BLUEFIN_TEST_ERROR=" + tc.kind}
			args := []string{"-test.run=^TestFangErrorHelper$"}
			out := runFailingCommand(t, false, env, os.Args[0], args...)
			want := tc.message
			if !strings.HasSuffix(want, "\n") {
				want += "\n"
			}
			if out != want {
				t.Errorf("redirected stderr = %q, want %q", out, want)
			}
			t.Run("terminal", func(t *testing.T) {
				out := terminalText(runFailingCommand(t, true, env, os.Args[0], args...))
				want := strings.ReplaceAll(tc.message, "\r", "")
				if !strings.Contains(out, "ERROR") || !strings.Contains(out, want) {
					t.Errorf("terminal output %q does not contain header and %q", out, want)
				}
				if strings.Contains(out, "failed..") || strings.Contains(out, "--Clear") || strings.Contains(out, "Try --help") {
					t.Errorf("terminal changed diagnostic: %q", out)
				}
			})
		})
	}
}

func TestBluefinErrorIntegration(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "bluefin-cli")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build Bluefin: %v\n%s", err, out)
	}
	home := t.TempDir()
	env := []string{"HOME=" + home, "USERPROFILE=" + home, "BLUEFIN_DISABLE_COUNTME=1"}
	for _, terminal := range []bool{false, true} {
		t.Run(fmt.Sprintf("terminal=%v", terminal), func(t *testing.T) {
			out := runFailingCommand(t, terminal, env, binary, "countme", "--clear")
			if terminal {
				out = terminalText(out)
			}
			if !strings.Contains(out, "unknown flag: --clear\n") || !strings.Contains(out, "Try --help for usage.") {
				t.Errorf("usage diagnostic: %q", out)
			}
			if !terminal && out != "unknown flag: --clear\nTry --help for usage.\n" {
				t.Errorf("redirected usage diagnostic: %q", out)
			}
			for _, args := range [][]string{{"init"}, {"countme", "--disable", "--enable"}} {
				out = runFailingCommand(t, terminal, env, binary, args...)
				if !strings.Contains(out, "Try --help for usage.") {
					t.Errorf("missing usage guidance for %v: %q", args, out)
				}
			}
			path := filepath.Join(home, "Missing-Path")
			out = runFailingCommand(t, terminal, env, binary, "brewfile", "list", "--file", path)
			if terminal {
				out = terminalText(out)
			}
			if !strings.Contains(out, "no Brewfile at "+path+" ") || strings.Contains(out, "Try --help") {
				t.Errorf("path diagnostic: %q", out)
			}
		})
	}
}

// script gives the child a real PTY; terminal tests are skipped where its
// Linux command-line interface is unavailable. Pipe tests run everywhere.
func runFailingCommand(t *testing.T, terminal bool, env []string, name string, args ...string) string {
	t.Helper()
	var cmd *exec.Cmd
	if terminal {
		if runtime.GOOS != "linux" {
			t.Skip("PTY test requires Linux script")
		}
		if _, err := exec.LookPath("script"); err != nil {
			t.Skip("script is unavailable")
		}
		parts := append([]string{name}, args...)
		for i, part := range parts {
			parts[i] = "'" + strings.ReplaceAll(part, "'", "'\\''") + "'"
		}
		cmd = exec.Command("script", "-q", "-e", "-c", strings.Join(parts, " "), "/dev/null")
	} else {
		cmd = exec.Command(name, args...)
	}
	cmd.Env = append(os.Environ(), append(env, "TERM=dumb", "NO_COLOR=1")...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if !terminal && stdout.Len() != 0 {
		t.Errorf("unexpected stdout: %q", stdout.String())
	}
	out := stderr.String()
	if terminal {
		out = stdout.String() + out // script forwards the PTY's combined streams to stdout.
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("command exit = %v, output %q", err, out)
	}
	return out
}

func terminalText(s string) string {
	return strings.ReplaceAll(ansi.Strip(s), "\r", "")
}
