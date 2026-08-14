package install

// Tests for the install.go core helpers that had zero coverage:
// IsLinux, IsGnome, CheckFlatpak, EnsureFlathub, CheckBbrew, EnsureBbrew,
// RunBbrew, Bundle/CustomBundles, and GetBrewfile error paths.
// Uses PATH manipulation and fake command stubs — no production changes.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── IsLinux / IsGnome ──────────────────────────────────────────────────────

func TestIsLinux(t *testing.T) {
	got := IsLinux()
	if got != (os.PathSeparator == '/' || os.Getenv("GOOS") == "linux") {
		t.Errorf("IsLinux() = %v", got)
	}
}

func TestIsGnome_EnvBased(t *testing.T) {
	t.Setenv("XDG_CURRENT_DESKTOP", "GNOME")
	if !IsGnome() {
		t.Error("IsGnome() = false with XDG_CURRENT_DESKTOP=GNOME")
	}

	t.Setenv("XDG_CURRENT_DESKTOP", "gnome")
	if !IsGnome() {
		t.Error("IsGnome() = false with lowercase gnome")
	}

	t.Setenv("XDG_CURRENT_DESKTOP", "KDE")
	if IsGnome() {
		t.Error("IsGnome() = true with XDG_CURRENT_DESKTOP=KDE")
	}

	t.Setenv("XDG_CURRENT_DESKTOP", "")
	if IsGnome() {
		t.Error("IsGnome() = true with empty XDG_CURRENT_DESKTOP")
	}
}

// ── CheckFlatpak / EnsureFlathub ───────────────────────────────────────────

func TestCheckFlatpak_Present(t *testing.T) {
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "flatpak"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := CheckFlatpak(); err != nil {
		t.Errorf("CheckFlatpak with flatpak in PATH: %v", err)
	}
}

func TestCheckFlatpak_Missing(t *testing.T) {
	t.Setenv("PATH", "/nonexistent-dir")

	if err := CheckFlatpak(); err == nil {
		t.Error("CheckFlatpak without flatpak: expected error")
	}
}

func TestEnsureFlathub_MissingFlatpak(t *testing.T) {
	t.Setenv("PATH", "/nonexistent-dir")

	err := EnsureFlathub()
	if err == nil {
		t.Fatal("EnsureFlathub without flatpak: expected error")
	}
	if !strings.Contains(err.Error(), "flatpak not found") {
		t.Errorf("error = %v, want 'flatpak not found'", err)
	}
}

func TestEnsureFlathub_AlreadyPresent(t *testing.T) {
	// fake flatpak whose remote-list prints flathub → no remote-add needed.
	binDir := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = \"remote-list\" ]; then echo flathub; fi\n"
	if err := os.WriteFile(filepath.Join(binDir, "flatpak"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := EnsureFlathub(); err != nil {
		t.Fatalf("EnsureFlathub with flathub present: %v", err)
	}
}

func TestEnsureFlathub_AddsRemote(t *testing.T) {
	// remote-list without flathub → remote-add must run.
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "flatpak-calls.log")
	script := "#!/bin/sh\necho \"$0 $@\" >> \"" + logPath + "\"\nif [ \"$1\" = \"remote-list\" ]; then echo \"other-remote\"; fi\n"
	if err := os.WriteFile(filepath.Join(binDir, "flatpak"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := EnsureFlathub(); err != nil {
		t.Fatalf("EnsureFlathub: %v", err)
	}

	data, _ := os.ReadFile(logPath)
	joined := string(data)
	if !strings.Contains(joined, "remote-add --if-not-exists flathub") {
		t.Errorf("missing remote-add call, got: %s", joined)
	}
}

// ── CheckBbrew / EnsureBbrew / RunBbrew ────────────────────────────────────

func TestCheckBbrew_Present(t *testing.T) {
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "bbrew"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := CheckBbrew(); err != nil {
		t.Errorf("CheckBbrew with bbrew in PATH: %v", err)
	}
}

func TestEnsureBbrew_AlreadyInstalled(t *testing.T) {
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "bbrew"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := EnsureBbrew(); err != nil {
		t.Fatalf("EnsureBbrew with bbrew present: %v", err)
	}
}

func TestRunBbrew(t *testing.T) {
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "bbrew-calls.log")
	script := "#!/bin/sh\necho \"$0 $@\" >> \"" + logPath + "\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "bbrew"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	if err := RunBbrew("/tmp/test.Brewfile"); err != nil {
		t.Fatalf("RunBbrew: %v", err)
	}
	data, _ := os.ReadFile(logPath)
	if !strings.Contains(string(data), "-f /tmp/test.Brewfile") {
		t.Errorf("bbrew call = %q, want '-f /tmp/test.Brewfile'", strings.TrimSpace(string(data)))
	}
}

// ── Bundle helpers ─────────────────────────────────────────────────────────

func TestBundle_UnknownName(t *testing.T) {
	_, ok := bundles["no-such-bundle"]
	if ok {
		t.Fatal("unexpected: no-such-bundle exists in bundles map")
	}
}

func TestCustomBundles_NoCustomDir(t *testing.T) {
	t.Setenv("HOME", filepath.Join(t.TempDir(), "missing"))

	// Missing dir → nil (callers treat nil/empty the same).
	if got := CustomBundles(); len(got) != 0 {
		t.Errorf("CustomBundles() with missing dir = %v, want nil or empty", got)
	}
}

func TestCustomBundles_FindsBrewfiles(t *testing.T) {
	home := t.TempDir()
	bundleDir := filepath.Join(home, ".config", "bluefin-cli", "bundles")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"mine.Brewfile", "theirs.Brewfile", "not-a-brewfile.txt"} {
		if err := os.WriteFile(filepath.Join(bundleDir, name), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)

	got := CustomBundles()
	if len(got) != 2 {
		t.Fatalf("CustomBundles() = %v, want [mine theirs]", got)
	}
	if got[0] != "mine" || got[1] != "theirs" {
		t.Errorf("CustomBundles() = %v, want sorted [mine theirs]", got)
	}
}

func TestGetBrewfile_MissingLocalPath(t *testing.T) {
	_, _, err := GetBrewfile("/does/not/exist.Brewfile")
	if err == nil {
		t.Fatal("GetBrewfile with missing path: expected error")
	}
	if !strings.Contains(err.Error(), "brewfile not found") {
		t.Errorf("error = %v, want 'brewfile not found'", err)
	}
}

func TestGetBrewfile_UnknownBundle(t *testing.T) {
	_, _, err := GetBrewfile("no-such-bundle")
	if err == nil {
		t.Fatal("GetBrewfile with unknown bundle: expected error")
	}
	if !strings.Contains(err.Error(), "unknown bundle") {
		t.Errorf("error = %v, want 'unknown bundle'", err)
	}
}
