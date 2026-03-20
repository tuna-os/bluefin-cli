package motd

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestToggle(t *testing.T) {
	// Toggle is deprecated and prints information, so just check it doesn't error
	err := Toggle("bash", true)
	if err != nil {
		t.Errorf("Toggle() returned error: %v", err)
	}
}

func TestGetImageInfo(t *testing.T) {
	// This function uses getImageInfo which is internal, but the test file is in 'package motd'
	// so it should have access if it was exported or if test is in same package.
	// Since getImageInfo is in motd.go and is package private 'getImageInfo', it is accessible here.
	info := getImageInfo()

	if info.ImageName == "" {
		t.Error("Expected ImageName to be set")
	}

	if info.ImageTag == "" {
		t.Error("Expected ImageTag to be set")
	}
}

func TestCheckStatus(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Manually inject legacy marker
	bashrc := filepath.Join(tmpHome, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("# bluefin-cli motd\n"), 0644); err != nil {
		t.Fatalf("Failed to write bashrc: %v", err)
	}

	status := CheckStatus()

	if !status["bash"] {
		t.Error("Expected bash MOTD to be enabled (legacy detection)")
	}
	if status["zsh"] {
		t.Error("Expected zsh MOTD to be disabled")
	}
}

func TestShow(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

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
	// ReadAll from pipe
	out, _ := io.ReadAll(r)
	output := string(out)

	// Verify expected output
	expectedStrings := []string{
		"Welcome to Bluefin CLI",
		"Command",
		"Description",
		"bluefin-cli",
		"GitHub Issues",
	}

	strippedOutput := stripAnsi(output)

	for _, s := range expectedStrings {
		if !strings.Contains(strippedOutput, s) {
			t.Errorf("Expected output to contain %q, got:\n%s", s, strippedOutput)
		}
	}
}

func stripAnsi(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
