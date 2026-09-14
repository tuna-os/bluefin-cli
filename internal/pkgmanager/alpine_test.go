//go:build !windows

// These tests put POSIX shell stubs on PATH to assert the exact commands the
// Alpine backend runs. A `#!/bin/sh` stub is not executable on Windows, and
// neither coldbrew nor sudo apk exists there -- installToolsWindows handles
// that platform through winget -- so the file is Unix-only rather than
// skipped at runtime.
package pkgmanager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExecutable(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestDetectAlpinePrefersColdbrew(t *testing.T) {
	dir := t.TempDir()
	writeExecutable(t, dir, "apk", "#!/bin/sh\n")
	writeExecutable(t, dir, "coldbrew", "#!/bin/sh\n")
	t.Setenv("PATH", dir)

	if got := DetectAlpine(); got != Coldbrew {
		t.Fatalf("DetectAlpine() = %q, want %q", got, Coldbrew)
	}
}

func TestInstallAlpineColdbrewInstallsAndWraps(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExecutable(t, dir, "coldbrew", "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$CALL_LOG\"\n")
	t.Setenv("PATH", dir)
	t.Setenv("CALL_LOG", log)

	if err := InstallAlpine(Coldbrew, "starship"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "install starship\nwrap starship\n" {
		t.Fatalf("calls = %q", got)
	}
}

func TestInstallAlpineAPKUsesSudo(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExecutable(t, dir, "sudo", "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$CALL_LOG\"\n")
	t.Setenv("PATH", dir)
	t.Setenv("CALL_LOG", log)

	if err := InstallAlpine(APK, "starship"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != "apk add starship" {
		t.Fatalf("call = %q", got)
	}
}
