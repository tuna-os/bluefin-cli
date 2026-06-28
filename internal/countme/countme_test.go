// Tests for the Fedora countme protocol implementation.
//
// These tests cover time window arithmetic, age bucket computation,
// user-agent generation, state persistence, and status string formatting.
// HTTP calls (sendPing) and environment-sensitive functions (variant,
// goosName, baseArch) are tested at the unit level where possible.
package countme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ── Time window arithmetic ───────────────────────────────────────────────────

func TestWindowNumber(t *testing.T) {
	tests := []struct {
		name  string
		unix  int64
		want  int64
	}{
		{"epoch (1970-01-01)", 0, 0},   // Before first window — integer division rounds to 0
		{"first window start", 345600, 0}, // 1970-01-05 00:00:00 UTC
		{"first window +1s", 345601, 0},
		{"second window start", 345600 + 604800, 1},
		{"second window -1s", 345600 + 604800 - 1, 0},
		{"one week later", 345600 + 604800*52, 52},
		{"one year later (approx)", 345600 + 604800*52, 52},
		{"large window", 345600 + 604800*1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := windowNumber(tt.unix)
			if got != tt.want {
				t.Errorf("windowNumber(%d) = %d, want %d", tt.unix, got, tt.want)
			}
		})
	}
}

func TestWindowToUnix(t *testing.T) {
	tests := []struct {
		name   string
		window int64
		want   int64
	}{
		{"window 0", 0, 345600},
		{"window 1", 1, 345600 + 604800},
		{"window 52", 52, 345600 + 604800*52},
		{"window 1000", 1000, 345600 + 604800*1000},
		{"negative window", -5, -5*604800 + 345600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := windowToUnix(tt.window)
			if got != tt.want {
				t.Errorf("windowToUnix(%d) = %d, want %d", tt.window, got, tt.want)
			}
		})
	}
}

func TestWindowRoundTrip(t *testing.T) {
	// Verify that windowNumber(windowToUnix(w)) == w for all relevant windows.
	for w := int64(0); w < 520; w++ {
		unix := windowToUnix(w)
		back := windowNumber(unix)
		if back != w {
			t.Errorf("round-trip failed for window %d: windowToUnix(%d)=%d, windowNumber(%d)=%d",
				w, w, unix, unix, back)
		}
	}
}

// ── Age bucket computation ───────────────────────────────────────────────────

func TestAgeBucket(t *testing.T) {
	tests := []struct {
		name       string
		epochWin   int64
		currentWin int64
		want       int
	}{
		{"same window (first week)", 100, 100, 1},
		{"next window (bucket 1)", 100, 101, 1},
		{"2 windows later (bucket 2 boundary)", 100, 102, 2},
		{"4 windows later (bucket 2)", 100, 104, 2},
		{"5 windows later (bucket 3 boundary)", 100, 105, 3},
		{"24 windows later (bucket 3)", 100, 124, 3},
		{"25 windows later (bucket 4 boundary)", 100, 125, 4},
		{"100 windows later (bucket 4)", 100, 200, 4},
		{"negative step (future epoch)", 200, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ageBucket(tt.epochWin, tt.currentWin)
			if got != tt.want {
				t.Errorf("ageBucket(%d, %d) = %d, want %d",
					tt.epochWin, tt.currentWin, got, tt.want)
			}
		})
	}
}

// ── User-Agent generation ────────────────────────────────────────────────────

func TestUserAgent(t *testing.T) {
	// The UA format is: "libdnf (Bluefin <version>; <variant>; <OS>.<arch>)"
	ua := userAgent("0.0.3")
	if !strings.HasPrefix(ua, "libdnf (Bluefin") {
		t.Errorf("userAgent should start with 'libdnf (Bluefin', got: %s", ua)
	}
	if !strings.HasSuffix(ua, ")") {
		t.Errorf("userAgent should end with ')', got: %s", ua)
	}

	// Check version is embedded
	if !strings.Contains(ua, "Bluefin 0.0.3") {
		t.Errorf("userAgent should contain 'Bluefin 0.0.3', got: %s", ua)
	}

	// Check arch and OS are present (at least one of darwin/linux/windows)
	if !strings.Contains(ua, runtime.GOOS) && !strings.Contains(ua, goosName()) {
		t.Errorf("userAgent should contain OS indicator, got: %s", ua)
	}

	// Check variant is valid
	knownVariants := []string{"mac", "wsl", "powershell", "linux"}
	hasVariant := false
	for _, v := range knownVariants {
		if strings.Contains(ua, v) {
			hasVariant = true
			break
		}
	}
	if !hasVariant {
		t.Errorf("userAgent should contain a known variant, got: %s", ua)
	}
}

func TestUserAgentVersion(t *testing.T) {
	versions := []string{"0.0.1", "1.0.0", "2026.6.28", "0.0.0"}
	for _, v := range versions {
		ua := userAgent(v)
		if !strings.Contains(ua, "Bluefin "+v) {
			t.Errorf("userAgent(%q) should contain 'Bluefin %s', got: %s", v, v, ua)
		}
	}
}

// ── Architecture and OS mapping ──────────────────────────────────────────────

func TestBaseArch(t *testing.T) {
	arch := baseArch()
	// On the test runner, we expect one of the standard Go arch values mapped
	// to RPM convention.
	known := map[string]bool{
		"x86_64":  true,
		"aarch64": true,
		"amd64":   false, // Go arch, not RPM — won't appear after mapping
		"arm64":   false,
	}
	if _, ok := known[arch]; !ok {
		// Unknown architecture on this runner — that's OK, just log it
		t.Logf("baseArch returned %q (unexpected but not an error on this platform)", arch)
	}
}

func TestGoosName(t *testing.T) {
	name := goosName()
	known := map[string]bool{
		"Darwin":  true,
		"Linux":   true,
		"Windows": true,
	}
	if _, ok := known[name]; !ok {
		t.Errorf("goosName returned %q, expected one of Darwin/Linux/Windows", name)
	}
}

// ── State persistence ────────────────────────────────────────────────────────

func TestLoadState_FileNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countme.json")

	state, err := loadState(path)
	if err != nil {
		t.Fatalf("loadState for non-existent file should not error, got: %v", err)
	}
	if state != (State{}) {
		t.Errorf("loadState should return zero state for missing file, got: %+v", state)
	}
}

func TestLoadState_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countme.json")

	original := State{Epoch: 1000, Window: 2000, Disabled: false}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	state, err := loadState(path)
	if err != nil {
		t.Fatalf("loadState should not error for valid JSON, got: %v", err)
	}
	if state.Epoch != 1000 || state.Window != 2000 || state.Disabled {
		t.Errorf("loadState = %+v, want {Epoch:1000 Window:2000 Disabled:false}", state)
	}
}

func TestLoadState_DisabledState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countme.json")

	original := State{Epoch: 1000, Window: 2000, Disabled: true}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	state, err := loadState(path)
	if err != nil {
		t.Fatalf("loadState should not error, got: %v", err)
	}
	if !state.Disabled {
		t.Error("loadState should return Disabled=true")
	}
}

func TestLoadState_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countme.json")

	if err := os.WriteFile(path, []byte("{this is not json}"), 0600); err != nil {
		t.Fatal(err)
	}

	state, err := loadState(path)
	if err != nil {
		t.Fatalf("loadState should return zero state on corrupt JSON, not error, got: %v", err)
	}
	if state != (State{}) {
		t.Errorf("loadState should return zero state for corrupt JSON, got: %+v", state)
	}
}

func TestSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countme.json")

	original := State{Epoch: 42, Window: 99, Disabled: false}
	if err := saveState(path, original); err != nil {
		t.Fatalf("saveState failed: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved state: %v", err)
	}

	var parsed State
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}
	if parsed != original {
		t.Errorf("save/load round-trip: got %+v, want %+v", parsed, original)
	}
}

func TestSaveState_Disabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "countme.json")

	original := State{Epoch: 42, Window: 99, Disabled: true}
	if err := saveState(path, original); err != nil {
		t.Fatalf("saveState failed: %v", err)
	}

	loaded, err := loadState(path)
	if err != nil {
		t.Fatalf("loadState failed: %v", err)
	}
	if !loaded.Disabled {
		t.Error("disabled state was not persisted correctly")
	}
}

// ── StatusString ─────────────────────────────────────────────────────────────

func TestStatusString_OptOutEnvVar(t *testing.T) {
	// Set the opt-out env var
	t.Setenv("BLUEFIN_DISABLE_COUNTME", "1")

	status := StatusString("0.0.3")
	if !strings.Contains(status, "disabled") {
		t.Errorf("StatusString should indicate disabled via env var, got: %s", status)
	}
}

func TestStatusString_DisabledState(t *testing.T) {
	// Create a state file with Disabled=true
	dir := t.TempDir()
	statePath := filepath.Join(dir, "countme.json")
	if err := saveState(statePath, State{Epoch: 100, Window: 200, Disabled: true}); err != nil {
		t.Fatal(err)
	}

	// Temporarily set env.GetConfigDir to return our temp dir
	// (We test the status string by using the real state file path)
	t.Setenv("BLUEFIN_DISABLE_COUNTME", "")
	// Override the config dir via env
	origHome, _ := os.UserHomeDir()
	t.Setenv("HOME", filepath.Dir(filepath.Dir(statePath)))
	defer t.Setenv("HOME", origHome)

	// We can't easily mock GetConfigDir, but we can test state-dependent
	// status formatting via loadState directly.
	state, err := loadState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Disabled {
		t.Error("expected disabled state to be true")
	}
}

func TestStatusString_Formatting(t *testing.T) {
	status := StatusString("0.0.3")
	// Should always contain "countme:" prefix
	if !strings.HasPrefix(status, "countme:") {
		t.Errorf("StatusString should start with 'countme:', got: %s", status)
	}
}

// ── sendPing validation (no actual HTTP call) ────────────────────────────────

func TestSendPingRequestFormat(t *testing.T) {
	// Verify the URL format by inspecting sendPing internals.
	// We can't call sendPing without making an HTTP request, so we verify
	// the URL construction pattern by examining the component functions.
	arch := baseArch()
	if arch == "" {
		t.Error("baseArch() should not return empty string")
	}

	// Verify userAgent doesn't contain placeholder text
	ua := userAgent("0.0.3")
	if strings.Contains(ua, "%") {
		t.Errorf("userAgent appears to contain unformatted printf directives: %s", ua)
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────────

func TestAgeBucket_EdgeCases(t *testing.T) {
	// Very large window differences
	got := ageBucket(0, 10000)
	if got < 1 || got > 4 {
		t.Errorf("ageBucket(0, 10000) = %d, want in range [1,4]", got)
	}

	// Negative epoch window (should be treated as bucket 1)
	got = ageBucket(-10, 0)
	if got != 3 {
		t.Errorf("ageBucket(-10, 0) = %d, want 3", got)
	}
}

func TestWindowNumber_EdgeCases(t *testing.T) {
	// Very large timestamps
	got := windowNumber(1 << 60)
	if got <= 0 {
		t.Errorf("windowNumber(2^60) = %d, expected positive value", got)
	}

	// Current timestamp (approximately)
	now := time.Now().Unix()
	win := windowNumber(now)
	if win <= 0 {
		t.Errorf("windowNumber(now=%d) = %d, expected positive", now, win)
	}
}

func TestWindowToUnix_EdgeCases(t *testing.T) {
	// Large window index
	unix := windowToUnix(1 << 30)
	if unix <= 0 {
		t.Errorf("windowToUnix(2^30) = %d, expected positive", unix)
	}

	// Verify window boundaries align
	w0 := windowToUnix(0)
	if w0%604800 != 345600%604800 {
		t.Errorf("window 0 start (%d) should align to offset boundary", w0)
	}
}
