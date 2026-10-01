package doctor

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/update"
)

type mockFileInfo struct {
	fs.FileInfo
	name string
}

func (m mockFileInfo) Name() string { return m.name }

func TestCheckBrew(t *testing.T) {
	tests := []struct {
		name       string
		isAlpine   bool
		lookPath   func(string) (string, error)
		stat       func(string) (os.FileInfo, error)
		wantOK     bool
		wantWarn   bool
		wantInName string
		wantInNote string
	}{
		{
			name:     "alpine with coldbrew",
			isAlpine: true,
			lookPath: func(bin string) (string, error) {
				if bin == "coldbrew" {
					return "/bin/coldbrew", nil
				}
				return "", os.ErrNotExist
			},
			wantOK:     true,
			wantInName: "coldbrew",
		},
		{
			name:     "alpine with apk fallback",
			isAlpine: true,
			lookPath: func(bin string) (string, error) {
				if bin == "apk" {
					return "/sbin/apk", nil
				}
				return "", os.ErrNotExist
			},
			wantWarn:   true,
			wantInName: "coldbrew not set up yet",
			wantInNote: "doctor --fix",
		},
		{
			name:     "alpine neither found",
			isAlpine: true,
			lookPath: func(string) (string, error) {
				return "", os.ErrNotExist
			},
			wantWarn:   true,
			wantInName: "Package manager",
			wantInNote: "neither coldbrew nor apk found",
		},
		{
			name:     "standard with brew on path",
			isAlpine: false,
			lookPath: func(bin string) (string, error) {
				if bin == "brew" {
					return "/usr/local/bin/brew", nil
				}
				return "", os.ErrNotExist
			},
			wantOK:     true,
			wantInName: "Homebrew on PATH",
		},
		{
			name:     "standard with brew in /opt/homebrew not on path",
			isAlpine: false,
			lookPath: func(string) (string, error) {
				return "", os.ErrNotExist
			},
			stat: func(path string) (os.FileInfo, error) {
				if path == "/opt/homebrew/bin/brew" {
					return mockFileInfo{name: "brew"}, nil
				}
				return nil, os.ErrNotExist
			},
			wantWarn:   true,
			wantInName: "Homebrew on PATH",
			wantInNote: "/opt/homebrew/bin/brew",
		},
		{
			name:     "standard brew not found",
			isAlpine: false,
			lookPath: func(string) (string, error) {
				return "", os.ErrNotExist
			},
			stat: func(string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			wantWarn:   true,
			wantInName: "Homebrew on PATH",
			wantInNote: "https://brew.sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				IsAlpine: func() bool { return tt.isAlpine },
				LookPath: tt.lookPath,
				Stat:     tt.stat,
			}
			res := CheckBrew(opts)
			if res.OK != tt.wantOK {
				t.Errorf("CheckBrew().OK = %v, want %v", res.OK, tt.wantOK)
			}
			if res.Warn != tt.wantWarn {
				t.Errorf("CheckBrew().Warn = %v, want %v", res.Warn, tt.wantWarn)
			}
			if !strings.Contains(res.Name, tt.wantInName) {
				t.Errorf("CheckBrew().Name = %q, want containing %q", res.Name, tt.wantInName)
			}
			if tt.wantInNote != "" && !strings.Contains(res.Note, tt.wantInNote) {
				t.Errorf("CheckBrew().Note = %q, want containing %q", res.Note, tt.wantInNote)
			}
		})
	}
}

func TestCheckShellIntegration(t *testing.T) {
	tests := []struct {
		name       string
		shell      string
		status     map[string]bool
		wantOK     bool
		wantWarn   bool
		wantInName string
	}{
		{
			name:       "shell enabled",
			shell:      "zsh",
			status:     map[string]bool{"zsh": true},
			wantOK:     true,
			wantWarn:   false,
			wantInName: "zsh",
		},
		{
			name:       "shell disabled",
			shell:      "bash",
			status:     map[string]bool{"bash": false},
			wantOK:     false,
			wantWarn:   true,
			wantInName: "bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				CurrentShell: tt.shell,
				ShellStatus:  func() map[string]bool { return tt.status },
			}
			res := CheckShellIntegration(opts)
			if res.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v", res.OK, tt.wantOK)
			}
			if res.Warn != tt.wantWarn {
				t.Errorf("Warn = %v, want %v", res.Warn, tt.wantWarn)
			}
			if !strings.Contains(res.Name, tt.wantInName) {
				t.Errorf("Name = %q, want containing %q", res.Name, tt.wantInName)
			}
		})
	}
}

func TestCheckTools(t *testing.T) {
	t.Run("all installed", func(t *testing.T) {
		opts := Options{
			LookPath: func(string) (string, error) { return "/bin/tool", nil },
		}
		res := CheckTools(opts)
		if !res.OK {
			t.Errorf("expected OK, got %+v", res)
		}
	})

	t.Run("some missing", func(t *testing.T) {
		opts := Options{
			LookPath: func(tool string) (string, error) {
				if tool == "fzf" {
					return "", os.ErrNotExist
				}
				return "/bin/" + tool, nil
			},
		}
		res := CheckTools(opts)
		if res.OK || !res.Warn {
			t.Errorf("expected warn result, got %+v", res)
		}
		if !strings.Contains(res.Note, "fzf") {
			t.Errorf("expected note to mention fzf, got %q", res.Note)
		}
	})
}

func TestCheckNetwork(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		opts := Options{
			CheckNetwork: func() error { return nil },
		}
		res := CheckNetwork(opts)
		if !res.OK {
			t.Errorf("expected OK, got %+v", res)
		}
	})

	t.Run("failure", func(t *testing.T) {
		opts := Options{
			CheckNetwork: func() error { return errors.New("connection refused") },
		}
		res := CheckNetwork(opts)
		if res.OK || !res.Warn {
			t.Errorf("expected warn, got %+v", res)
		}
		if !strings.Contains(res.Note, "connection refused") {
			t.Errorf("expected note to mention connection refused, got %q", res.Note)
		}
	})
}

func TestCheckSelfUpdate(t *testing.T) {
	t.Run("managed by package manager", func(t *testing.T) {
		opts := Options{
			DetectUpdate: func() update.InstallMethod { return update.MethodHomebrew },
		}
		res := CheckSelfUpdate(opts)
		if !res.OK || !strings.Contains(res.Name, "homebrew") {
			t.Errorf("expected OK with homebrew, got %+v", res)
		}
	})

	t.Run("direct writable", func(t *testing.T) {
		opts := Options{
			DetectUpdate: func() update.InstallMethod { return update.MethodDirect },
			Executable:   func() (string, error) { return "/opt/bluefin/bin/bluefin-cli", nil },
			CreateTempFile: func(dir, pattern string) (*os.File, error) {
				return os.CreateTemp("", pattern)
			},
			RemoveTempFile: func(name string) error { return os.Remove(name) },
		}
		res := CheckSelfUpdate(opts)
		if !res.OK {
			t.Errorf("expected OK, got %+v", res)
		}
	})

	t.Run("direct not writable", func(t *testing.T) {
		opts := Options{
			DetectUpdate: func() update.InstallMethod { return update.MethodDirect },
			Executable:   func() (string, error) { return "/usr/bin/bluefin-cli", nil },
			CreateTempFile: func(dir, pattern string) (*os.File, error) {
				return nil, os.ErrPermission
			},
		}
		res := CheckSelfUpdate(opts)
		if res.OK || !res.Warn {
			t.Errorf("expected warn, got %+v", res)
		}
		if !strings.Contains(res.Note, "not writable") {
			t.Errorf("expected not writable note, got %q", res.Note)
		}
	})
}

func TestCheckVersion(t *testing.T) {
	t.Run("dev build", func(t *testing.T) {
		res := CheckVersion(Options{Version: "dev"})
		if !res.OK || !strings.Contains(res.Name, "dev") {
			t.Errorf("expected OK dev build, got %+v", res)
		}
	})

	t.Run("latest version", func(t *testing.T) {
		opts := Options{
			Version: "v1.0.0",
			LatestRelease: func() (*update.Release, error) {
				return &update.Release{TagName: "v1.0.0"}, nil
			},
			IsNewer: func(cur, latest string) bool { return false },
		}
		res := CheckVersion(opts)
		if !res.OK || !strings.Contains(res.Name, "latest") {
			t.Errorf("expected OK latest, got %+v", res)
		}
	})

	t.Run("outdated version", func(t *testing.T) {
		opts := Options{
			Version: "v1.0.0",
			LatestRelease: func() (*update.Release, error) {
				return &update.Release{TagName: "v1.1.0"}, nil
			},
			IsNewer:    func(cur, latest string) bool { return true },
			UpdateHint: func(m update.InstallMethod) string { return "bluefin-cli update" },
		}
		res := CheckVersion(opts)
		if res.OK || !res.Warn {
			t.Errorf("expected warn for outdated, got %+v", res)
		}
		if !strings.Contains(res.Note, "v1.1.0 available") {
			t.Errorf("expected note with v1.1.0, got %q", res.Note)
		}
	})
}

func TestDiagnoseAndRender(t *testing.T) {
	opts := Options{
		CurrentShell: "bash",
		Version:      "dev",
		LookPath:     func(string) (string, error) { return "/bin/tool", nil },
		ShellStatus:  func() map[string]bool { return map[string]bool{"bash": true} },
		CheckNetwork: func() error { return nil },
		DetectUpdate: func() update.InstallMethod { return update.MethodHomebrew },
		IsAlpine:     func() bool { return false },
	}

	report := Diagnose(opts)
	if report.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", report.Failures)
	}
	rendered := report.Render()
	if !strings.Contains(rendered, "All checks passed") {
		t.Errorf("expected 'All checks passed' in output, got:\n%s", rendered)
	}
}

func TestRunFixes(t *testing.T) {
	var buf bytes.Buffer
	toggled := false
	ensuredColdbrew := false
	installed := []string{}

	opts := FixOptions{
		CurrentShell: "bash",
		IsAlpine:     func() bool { return true },
		CommandExists: func(bin string) bool {
			return bin == "eza" // coldbrew, fzf, starship missing
		},
		CheckStatus: func() map[string]bool {
			return map[string]bool{"bash": false}
		},
		ToggleShell: func(sh string, enable bool) error {
			if sh == "bash" && enable {
				toggled = true
			}
			return nil
		},
		EnsureColdbrew: func() bool {
			ensuredColdbrew = true
			return true
		},
		EnsureInstalled: func(tool string) error {
			installed = append(installed, tool)
			return nil
		},
		Output: &buf,
	}

	RunFixes(opts)

	if !toggled {
		t.Errorf("expected ToggleShell to be called")
	}
	if !ensuredColdbrew {
		t.Errorf("expected EnsureColdbrew to be called")
	}
	if len(installed) != 2 || installed[0] != "fzf" || installed[1] != "starship" {
		t.Errorf("unexpected installed tools: %v", installed)
	}
	out := buf.String()
	if !strings.Contains(out, "enabling shell integration") || !strings.Contains(out, "setting up coldbrew") {
		t.Errorf("unexpected fix output: %s", out)
	}
}

func TestRunBench(t *testing.T) {
	t.Run("unsupported shell", func(t *testing.T) {
		_, err := RunBench(BenchOptions{Shell: "tcsh"})
		if err == nil {
			t.Errorf("expected error for unsupported shell")
		}
	})

	t.Run("missing shell binary", func(t *testing.T) {
		opts := BenchOptions{
			Shell:    "bash",
			LookPath: func(string) (string, error) { return "", os.ErrNotExist },
		}
		_, err := RunBench(opts)
		if err == nil {
			t.Errorf("expected error when shell not on PATH")
		}
	})

	t.Run("successful bench execution", func(t *testing.T) {
		currentTime := time.Now()
		runs := 0
		opts := BenchOptions{
			Shell:    "bash",
			LookPath: func(string) (string, error) { return "/bin/bash", nil },
			Runner: func(sh string, args ...string) error {
				runs++
				return nil
			},
			Now: func() time.Time {
				currentTime = currentTime.Add(10 * time.Millisecond)
				return currentTime
			},
			Runs: 3,
		}

		res, err := RunBench(opts)
		if err != nil {
			t.Fatalf("RunBench failed: %v", err)
		}
		if res.Shell != "bash" {
			t.Errorf("res.Shell = %q, want bash", res.Shell)
		}
		rendered := RenderBench(res)
		if !strings.Contains(rendered, "Benchmarking bash startup") {
			t.Errorf("unexpected render output: %s", rendered)
		}
	})
}
