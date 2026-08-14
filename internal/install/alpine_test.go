package install

// Tests for the Alpine/musl code path (internal/install/alpine.go), which
// had zero coverage: installBundleAlpine, alpinePackageManager, and
// installAlpinePkg.
//
// No production code changes: a fake coldbrew/apk pair is written into a
// temp bin dir that is prepended to PATH, and the scripts append their
// invocations to a log file — so installAlpinePkg / installBundleAlpine run
// against real exec.Command with observable, controllable behaviour.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installFakeAlpineTools writes coldbrew/apk stub scripts into a temp bin
// dir; every invocation is appended to <bin>/calls.log. Returns the dir.
func installFakeAlpineTools(t *testing.T, failPkgs ...string) string {
	t.Helper()
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "calls.log")

	failSet := make(map[string]bool, len(failPkgs))
	for _, p := range failPkgs {
		failSet[p] = true
	}
	var failExpr strings.Builder
	for _, p := range failPkgs {
		failExpr.WriteString("  if [ \"$2\" = \"" + p + "\" ]; then exit 1; fi\n")
	}

	script := "#!/bin/sh\n" +
		"echo \"$0 $@\" >> \"" + logPath + "\"\n" +
		failExpr.String() + "\n"

	for _, name := range []string{"coldbrew", "apk", "sudo"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	orig := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+orig)
	return binDir
}

func readCalls(t *testing.T, binDir string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(binDir, "calls.log"))
	if err != nil {
		return nil
	}
	var calls []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) != "" {
			calls = append(calls, strings.TrimSpace(line))
		}
	}
	return calls
}

func writeBrewfile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.Brewfile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ── alpinePackageManager ───────────────────────────────────────────────────

func TestAlpinePackageManager_NoneAvailable(t *testing.T) {
	t.Setenv("PATH", "/nonexistent-dir")

	if got := alpinePackageManager(); got != "" {
		t.Errorf("alpinePackageManager with empty PATH = %q, want \"\"", got)
	}
}

func TestAlpinePackageManager_PrefersColdbrew(t *testing.T) {
	binDir := t.TempDir()
	for _, name := range []string{"coldbrew", "apk"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)

	if got := alpinePackageManager(); got != "coldbrew" {
		t.Errorf("alpinePackageManager = %q, want coldbrew (preferred over apk)", got)
	}
}

func TestAlpinePackageManager_FallsBackToApk(t *testing.T) {
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "apk"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if got := alpinePackageManager(); got != "apk" {
		t.Errorf("alpinePackageManager = %q, want apk fallback", got)
	}
}

// ── installAlpinePkg ───────────────────────────────────────────────────────

func TestInstallAlpinePkg_UnknownManager(t *testing.T) {
	err := installAlpinePkg("yum", "git")
	if err == nil {
		t.Fatal("installAlpinePkg with unknown manager: expected error")
	}
	if !strings.Contains(err.Error(), "unknown package manager") {
		t.Errorf("error = %v, want 'unknown package manager'", err)
	}
}

func TestInstallAlpinePkg_ColdbrewInstallAndWrap(t *testing.T) {
	binDir := installFakeAlpineTools(t)
	if err := installAlpinePkg("coldbrew", "git"); err != nil {
		t.Fatalf("installAlpinePkg(coldbrew, git): %v", err)
	}
	calls := readCalls(t, binDir)
	if len(calls) < 2 {
		t.Fatalf("expected install + wrap calls, got: %v", calls)
	}
	if !strings.Contains(strings.Join(calls, "\n"), "coldbrew install git") {
		t.Errorf("missing coldbrew install call, got: %v", calls)
	}
	if !strings.Contains(strings.Join(calls, "\n"), "coldbrew wrap git") {
		t.Errorf("missing coldbrew wrap call, got: %v", calls)
	}
}

func TestInstallAlpinePkg_ApkAdd(t *testing.T) {
	binDir := installFakeAlpineTools(t)
	if err := installAlpinePkg("apk", "git"); err != nil {
		t.Fatalf("installAlpinePkg(apk, git): %v", err)
	}
	calls := readCalls(t, binDir)
	if len(calls) != 1 || !strings.Contains(calls[0], "apk add git") {
		t.Errorf("expected single 'apk add git' call, got: %v", calls)
	}
}

func TestInstallAlpinePkg_FailedInstallPropagates(t *testing.T) {
	// Script exits 1 for the named package.
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "calls.log")
	script := "#!/bin/sh\necho \"$0 $@\" >> \"" + logPath + "\"\nif [ \"$2\" = \"git\" ]; then exit 1; fi\n"
	if err := os.WriteFile(filepath.Join(binDir, "coldbrew"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+orig)

	err := installAlpinePkg("coldbrew", "git")
	if err == nil {
		t.Fatal("expected error when coldbrew install fails")
	}
}

// ── installBundleAlpine ────────────────────────────────────────────────────

func TestInstallBundleAlpine_NoManager(t *testing.T) {
	t.Setenv("PATH", "/nonexistent-dir")

	path := writeBrewfile(t, "brew \"git\"\n")
	err := installBundleAlpine(path)
	if err == nil {
		t.Fatal("expected error when neither coldbrew nor apk is available")
	}
	if !strings.Contains(err.Error(), "coldbrew or apk") {
		t.Errorf("error = %v, want mention of missing coldbrew/apk", err)
	}
}

func TestInstallBundleAlpine_InstallsBrewFormulae(t *testing.T) {
	binDir := installFakeAlpineTools(t)

	path := writeBrewfile(t, "brew \"git\"\nbrew \"jq\"\n")
	if err := installBundleAlpine(path); err != nil {
		t.Fatalf("installBundleAlpine: %v", err)
	}

	calls := readCalls(t, binDir)
	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "coldbrew install git") {
		t.Errorf("missing install of git, calls: %v", calls)
	}
	if !strings.Contains(joined, "coldbrew install jq") {
		t.Errorf("missing install of jq, calls: %v", calls)
	}
}

func TestInstallBundleAlpine_SkipsCasks(t *testing.T) {
	binDir := installFakeAlpineTools(t)

	path := writeBrewfile(t, "brew \"git\"\ncask \"firefox\"\n")
	if err := installBundleAlpine(path); err != nil {
		t.Fatalf("installBundleAlpine: %v", err)
	}

	for _, c := range readCalls(t, binDir) {
		if strings.Contains(c, "firefox") {
			t.Errorf("cask should be skipped on Alpine, got install call: %s", c)
		}
	}
}

func TestInstallBundleAlpine_FailedPackageReported(t *testing.T) {
	// git fails to install; with nothing skipped the function prints the
	// failure but returns nil (the all-skipped error is only for casks etc.).
	installFakeAlpineTools(t, "git")

	path := writeBrewfile(t, "brew \"git\"\n")
	if err := installBundleAlpine(path); err != nil {
		t.Fatalf("installBundleAlpine with only-failing brew formula: %v", err)
	}
}

func TestInstallBundleAlpine_AllSkippedIsError(t *testing.T) {
	installFakeAlpineTools(t)

	path := writeBrewfile(t, "cask \"firefox\"\n")
	err := installBundleAlpine(path)
	if err == nil {
		t.Fatal("expected error when every package is skipped")
	}
	if !strings.Contains(err.Error(), "all 1 packages were skipped") {
		t.Errorf("error = %v, want all-skipped message", err)
	}
}

func TestInstallBundleAlpine_EmptyBundle(t *testing.T) {
	installFakeAlpineTools(t)

	path := writeBrewfile(t, "")
	err := installBundleAlpine(path)
	if err == nil {
		t.Fatal("expected error for an empty bundle")
	}
	if !strings.Contains(err.Error(), "no installable packages") {
		t.Errorf("error = %v, want no-installable-packages message", err)
	}
}
