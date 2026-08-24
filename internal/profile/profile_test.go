package profile

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// withTempHome isolates HOME (and USERPROFILE, for the Windows-profile code
// paths shared with internal/shell) so Export/Apply's filesystem side
// effects never touch the real user's dotfiles. Mirrors the pattern in
// internal/shell/shell_test.go.
func withTempHome(t *testing.T) string {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	return tmpHome
}

func TestDiff(t *testing.T) {
	current := &Profile{
		Version:       1,
		Flavor:        "auto",
		Tools:         map[string]bool{"eza": true, "bat": false},
		EnabledShells: []string{"bash", "zsh"},
	}
	want := &Profile{
		Version:       1,
		Flavor:        "mocha",
		Tools:         map[string]bool{"eza": true, "bat": true},
		EnabledShells: []string{"bash", "fish"},
	}

	diff := Diff(current, want)
	joined := strings.Join(diff, "\n")
	for _, expect := range []string{
		"flavor: auto -> mocha",
		"tool bat: off -> on",
		"shell fish: disabled -> enabled",
		"shell zsh: enabled -> disabled",
	} {
		if !strings.Contains(joined, expect) {
			t.Errorf("diff missing %q:\n%s", expect, joined)
		}
	}
	if strings.Contains(joined, "eza") || strings.Contains(joined, "bash") {
		t.Errorf("diff contains unchanged entries:\n%s", joined)
	}

	if d := Diff(want, want); len(d) != 0 {
		t.Errorf("self-diff should be empty, got %v", d)
	}
}

func TestSaveWritesIndentedJSONWithTrailingNewline(t *testing.T) {
	p := &Profile{
		Version:       1,
		Flavor:        "mocha",
		Tools:         map[string]bool{"eza": true},
		EnabledShells: []string{"bash"},
	}
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := p.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved profile: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Errorf("Save() output missing trailing newline: %q", data)
	}
	if !strings.Contains(string(data), "\n  \"version\": 1") {
		t.Errorf("Save() output not indented JSON: %q", data)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	want := &Profile{
		Version:       1,
		Flavor:        "latte",
		Tools:         map[string]bool{"bat": true, "eza": false},
		EnabledShells: []string{"fish", "zsh"},
	}
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := want.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load() round trip = %+v, want %+v", got, want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("Load() on a missing file should error")
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "parsing profile") {
		t.Fatalf("Load() error = %v, want it to wrap a parse error", err)
	}
}

func TestLoadUnsupportedVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(path, []byte(`{"version":2,"enabled_shells":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported profile version 2") {
		t.Fatalf("Load() error = %v, want unsupported-version error", err)
	}
}

func TestExportCapturesEnabledShellsAndTools(t *testing.T) {
	withTempHome(t)

	if err := shell.Toggle("bash", true); err != nil {
		t.Fatalf("Toggle(bash, true): %v", err)
	}
	if err := shell.Toggle("zsh", true); err != nil {
		t.Fatalf("Toggle(zsh, true): %v", err)
	}

	got, err := Export("bash")
	if err != nil {
		t.Fatalf("Export() error: %v", err)
	}
	if got.Version != 1 {
		t.Errorf("Export().Version = %d, want 1", got.Version)
	}

	shells := append([]string{}, got.EnabledShells...)
	sort.Strings(shells)
	if !reflect.DeepEqual(shells, []string{"bash", "zsh"}) {
		t.Errorf("Export().EnabledShells = %v, want [bash zsh]", shells)
	}
	if got.Tools == nil {
		t.Error("Export().Tools should be populated from the shell config")
	}
}

func TestApplyReconcilesShellIntegration(t *testing.T) {
	withTempHome(t)

	// Start with bash enabled, zsh disabled.
	if err := shell.Toggle("bash", true); err != nil {
		t.Fatalf("Toggle(bash, true): %v", err)
	}

	// Apply a profile that wants zsh on and bash off; leave Flavor/Tools
	// empty so Apply skips the viper-backed config.Save branch (covered
	// separately — this test isolates the shell-reconciliation branch).
	p := &Profile{Version: 1, EnabledShells: []string{"zsh"}}
	if err := p.Apply(); err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	status := shell.CheckStatus()
	if !status["zsh"] {
		t.Error("Apply() should have enabled zsh")
	}
	if status["bash"] {
		t.Error("Apply() should have disabled bash (absent from the profile)")
	}
	// fish was never touched and stays off.
	if status["fish"] {
		t.Error("Apply() should not enable shells absent from both current and wanted state")
	}
}
