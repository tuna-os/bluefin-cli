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

func TestRendererPure(t *testing.T) {
	mockReport := Report{
		AppVersion: "1.2.3",
		Flavor:     "cyberpunk",
		Shells: []ShellItem{
			{Name: "bash", Enabled: true, IsDefault: true, IsCurrent: true},
			{Name: "zsh", Enabled: false, IsDefault: false, IsCurrent: false},
		},
		MOTD: map[string]bool{
			"bash": true,
			"zsh":  false,
		},
		Tools: []ToolItem{
			{Name: "starship", Binary: "starship", Installed: true},
			{Name: "atuin", Binary: "atuin", Installed: false},
		},
		PackageManager: PackageManagerStatus{
			IsWindows:     false,
			BrewInstalled: true,
			BrewVersion:   "Homebrew 4.2.0",
		},
		Sunset: &SunsetInfo{
			Enabled:        true,
			Latitude:       37.7749,
			Longitude:      -122.4194,
			WallpaperTheme: "mojave",
		},
	}

	renderer := NewRenderer(80)
	out := stripAnsi(renderer.Render(mockReport))

	for _, expected := range []string{
		"Bluefin CLI Status  v1.2.3 · flavor cyberpunk",
		"Shell Experience:",
		"bash: enabled ★ (default, current)",
		"zsh: disabled",
		"Message of the Day:",
		"Managed Tools:",
		"starship: installed",
		"atuin: not installed",
		"Homebrew: installed",
		"Homebrew 4.2.0",
		"Sunset Automation:",
		"Location: 37.7749, -122.4194",
		"Theme: mojave",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("Render() missing expected substring %q in:\n%s", expected, out)
		}
	}
}

func TestRendererNarrowWidth(t *testing.T) {
	mockReport := Report{
		AppVersion: "0.1.0",
		Flavor:     "classic",
		Shells: []ShellItem{
			{Name: "bash", Enabled: true},
		},
		MOTD: map[string]bool{"bash": true},
		PackageManager: PackageManagerStatus{
			IsWindows:       true,
			WindowsManagers: []string{"winget", "scoop"},
		},
	}

	renderer := NewRenderer(40)
	out := stripAnsi(renderer.Render(mockReport))

	if !strings.Contains(out, "Windows PMs: winget, scoop") {
		t.Errorf("Narrow Render() missing Windows PM info in:\n%s", out)
	}
}
