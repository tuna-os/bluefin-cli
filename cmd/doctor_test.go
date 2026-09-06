package cmd

import (
	"strings"
	"testing"
)

// ── commandExists ─────────────────────────────────────────────────────────

func TestCommandExists(t *testing.T) {
	if !commandExists("sh") {
		t.Error("expected 'sh' to be found on PATH")
	}
	if commandExists("definitely-not-a-real-binary-xyz") {
		t.Error("expected a made-up binary name to be reported as missing")
	}
}

// ── checkVersion ──────────────────────────────────────────────────────────

func TestCheckVersion_DevBuild(t *testing.T) {
	old := version
	version = "dev"
	defer func() { version = old }()

	r := checkVersion()
	if !r.ok {
		t.Error("dev build should always report ok")
	}
	if !strings.Contains(r.name, "dev") {
		t.Errorf("expected name to mention dev build, got %q", r.name)
	}
}

// ── checkSelfUpdate ───────────────────────────────────────────────────────

func TestCheckSelfUpdate_ReturnsResult(t *testing.T) {
	// The test binary isn't installed via Homebrew/Scoop/Winget, so this
	// exercises the direct-install branch without touching the network.
	r := checkSelfUpdate()
	if r.name == "" {
		t.Error("expected a non-empty check name")
	}
}

// ── checkBrew ─────────────────────────────────────────────────────────────

func TestCheckBrew_ReturnsResult(t *testing.T) {
	r := checkBrew()
	if r.name == "" {
		t.Error("expected a non-empty check name")
	}
	if r.ok && r.warn {
		t.Error("a check result should not be both ok and warn")
	}
}

// ── checkTools ────────────────────────────────────────────────────────────

func TestCheckTools_ReturnsResult(t *testing.T) {
	r := checkTools()
	if !strings.Contains(r.name, "Managed tools") {
		t.Errorf("unexpected name: %q", r.name)
	}
	if !r.ok && r.note == "" {
		t.Error("a non-ok result should explain what's missing")
	}
}

// ── checkShellIntegration ─────────────────────────────────────────────────

func TestCheckShellIntegration_ReturnsResult(t *testing.T) {
	r := checkShellIntegration()
	current := currentShellName()
	if !strings.Contains(r.name, current) {
		t.Errorf("expected name to mention current shell %q, got %q", current, r.name)
	}
	if !r.ok && r.note == "" {
		t.Error("a non-ok result should include a fix hint")
	}
}

// ── checkNetwork ──────────────────────────────────────────────────────────

func TestCheckNetwork_ReturnsResult(t *testing.T) {
	r := checkNetwork()
	if r.name != "GitHub reachable" {
		t.Errorf("unexpected name: %q", r.name)
	}
	// Outcome depends on network access in the test environment; only the
	// shape of the result is asserted.
	if !r.ok && r.note == "" {
		t.Error("a failed network check should explain why")
	}
}

// ── doctorReport ──────────────────────────────────────────────────────────

func TestDoctorReport_ProducesOutput(t *testing.T) {
	report, failures := doctorReport()
	if report == "" {
		t.Error("expected non-empty report text")
	}
	if failures < 0 {
		t.Error("failures count should never be negative")
	}
	if failures == 0 && !strings.Contains(report, "All checks passed") {
		t.Error("zero failures should render the all-clear message")
	}
}
