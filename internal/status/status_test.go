package status

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

var ansiRegex = regexp.MustCompile("[\u001b\u009b][\\[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")

func stripAnsi(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

func TestShow(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout on exit
	defer func() {
		os.Stdout = oldStdout
	}()

	err := Show("v0.1.0")
	if err != nil {
		t.Errorf("Show() returned error: %v", err)
	}

	// Read captured output
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := stripAnsi(buf.String())

	// Verify expected sections
	expectedStrings := []string{
		"Bluefin CLI Status",
		"Shell Experience:",
		"Message of the Day:",
		"Managed Tools:",
		"Package Manager:",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain %q", s)
		}
	}
}

func TestShowComponents(t *testing.T) {
	err := Show("v0.1.0")
	if err != nil {
		t.Fatalf("Show() failed: %v", err)
	}
}

func TestReportHelper(t *testing.T) {
	enabled := stripAnsi(report(true, "zsh", "enabled", " ★ (default)"))
	if !strings.Contains(enabled, "✓ zsh: enabled ★ (default)") {
		t.Errorf("Unexpected report output for enabled: %q", enabled)
	}

	disabled := stripAnsi(report(false, "bash", "disabled"))
	if !strings.Contains(disabled, "✗ bash: disabled") {
		t.Errorf("Unexpected report output for disabled: %q", disabled)
	}
}

func TestRenderLayouts(t *testing.T) {
	wide, err := Render(100, "1.0.0")
	if err != nil {
		t.Fatalf("Render wide failed: %v", err)
	}
	cleanWide := stripAnsi(wide)
	if !strings.Contains(cleanWide, "Bluefin CLI Status") {
		t.Errorf("Expected title in wide output")
	}

	narrow, err := Render(50, "1.0.0")
	if err != nil {
		t.Fatalf("Render narrow failed: %v", err)
	}
	cleanNarrow := stripAnsi(narrow)
	if !strings.Contains(cleanNarrow, "Bluefin CLI Status") {
		t.Errorf("Expected title in narrow output")
	}
}

func TestReportShellExperience(t *testing.T) {
	// Test empty shells
	emptyOut := stripAnsi(reportShellExperience(shellInfo{
		shellsToShow: nil,
	}))
	if !strings.Contains(emptyOut, "(no compatible shells found)") {
		t.Errorf("Expected no compatible shells notice, got %q", emptyOut)
	}

	// Test populated shells with markers
	populatedOut := stripAnsi(reportShellExperience(shellInfo{
		shellStatus:  map[string]bool{"zsh": true, "bash": false},
		shellsToShow: []string{"zsh", "bash"},
		defaultShell: "zsh",
		currentShell: "zsh",
	}))
	if !strings.Contains(populatedOut, "✓ zsh: enabled ★ (default, current)") {
		t.Errorf("Expected default and current marker for zsh, got %q", populatedOut)
	}
	if !strings.Contains(populatedOut, "✗ bash: disabled") {
		t.Errorf("Expected disabled bash, got %q", populatedOut)
	}
}

func TestReportMOTD(t *testing.T) {
	motdOut := stripAnsi(reportMOTD([]string{"bash"}))
	if !strings.Contains(motdOut, "Message of the Day:") {
		t.Errorf("Expected Message of the Day header, got %q", motdOut)
	}
	if !strings.Contains(motdOut, "bash:") {
		t.Errorf("Expected bash in MOTD report, got %q", motdOut)
	}
}

func TestReportManagedTools(t *testing.T) {
	toolsOut := stripAnsi(reportManagedTools("bash"))
	if !strings.Contains(toolsOut, "Managed Tools:") {
		t.Errorf("Expected Managed Tools header, got %q", toolsOut)
	}
}

func TestReportPackageManager(t *testing.T) {
	pmOut := stripAnsi(reportPackageManager())
	if !strings.Contains(pmOut, "Package Manager:") {
		t.Errorf("Expected Package Manager header, got %q", pmOut)
	}
}

func TestReportSunset(t *testing.T) {
	// Should not panic on any platform
	_ = reportSunset()
}
