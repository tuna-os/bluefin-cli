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

	err := Show()
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
	// This test mainly verifies that Show runs without panicking
	// We'll trust TestShow to verify the output content
	err := Show()
	if err != nil {
		t.Fatalf("Show() failed: %v", err)
	}
}
