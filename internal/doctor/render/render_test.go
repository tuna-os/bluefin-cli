package render

import (
	"strings"
	"testing"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/doctor"
)

func TestReportTextRendersSuccessAndFailureStates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		out := ReportText(doctor.Report{
			Checks: []doctor.CheckResult{{OK: true, Name: "Homebrew on PATH"}},
		})
		if !strings.Contains(out, "✓") || !strings.Contains(out, "Homebrew on PATH") {
			t.Fatalf("success output = %q", out)
		}
		if !strings.Contains(out, "All checks passed") {
			t.Fatalf("success output missing completion message: %q", out)
		}
	})

	t.Run("failure", func(t *testing.T) {
		out := ReportText(doctor.Report{
			Checks:   []doctor.CheckResult{{Name: "GitHub reachable", Note: "connection refused"}},
			Failures: 1,
		})
		if !strings.Contains(out, "✗") || !strings.Contains(out, "GitHub reachable") {
			t.Fatalf("failure output = %q", out)
		}
		if !strings.Contains(out, "connection refused") {
			t.Fatalf("failure output missing note: %q", out)
		}
		if strings.Contains(out, "All checks passed") {
			t.Fatalf("failure output incorrectly reported success: %q", out)
		}
	})
}

func TestBenchTextShowsOverheadThreshold(t *testing.T) {
	base := doctor.BenchResult{
		Shell:    "bash",
		Full:     500 * time.Millisecond,
		Base:     400 * time.Millisecond,
		Overhead: 100 * time.Millisecond,
	}
	if out := BenchText(&base); !strings.Contains(out, "Snappy.") {
		t.Fatalf("below-threshold output = %q", out)
	}

	over := base
	over.Overhead = 301 * time.Millisecond
	if out := BenchText(&over); !strings.Contains(out, "Startup overhead is noticeable") {
		t.Fatalf("above-threshold output = %q", out)
	}
}
