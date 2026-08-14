package install

// Tests for the Unix (non-Windows) installer path (internal/install/
// unix_installer.go), which had zero coverage: InstallBundle routing
// (Alpine vs Homebrew), the homebrew-not-found error, and wallpaper cask
// argument construction. A fake brew stub in PATH records invocations;
// local Brewfile paths bypass the network/bundle lookup entirely.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installFakeBrew writes a brew stub into a temp bin dir that appends its
// invocation to calls.log. Returns the bin dir.
func installFakeBrew(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "brew-calls.log")
	script := "#!/bin/sh\necho \"$0 $@\" >> \"" + logPath + "\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "brew"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", orig) })
	os.Setenv("PATH", binDir+":"+orig)
	return binDir
}

func readBrewCalls(t *testing.T, binDir string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(binDir, "brew-calls.log"))
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

func writeLocalBrewfile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "local.Brewfile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ── InstallBundle ──────────────────────────────────────────────────────────

func TestUnixInstaller_InstallBundle_NoBrew(t *testing.T) {
	orig := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", orig) })
	os.Setenv("PATH", "/nonexistent-dir")

	i := &UnixInstaller{}
	err := i.InstallBundle(writeLocalBrewfile(t, "brew \"git\"\n"))
	if err == nil {
		t.Fatal("expected error when brew is missing")
	}
	if !strings.Contains(err.Error(), "homebrew not found") {
		t.Errorf("error = %v, want 'homebrew not found'", err)
	}
}

func TestUnixInstaller_InstallBundle_RunsBrewBundleInstall(t *testing.T) {
	binDir := installFakeBrew(t)
	i := &UnixInstaller{}

	bf := writeLocalBrewfile(t, "brew \"git\"\nbrew \"jq\"\n")
	if err := i.InstallBundle(bf); err != nil {
		t.Fatalf("InstallBundle: %v", err)
	}

	calls := readBrewCalls(t, binDir)
	if len(calls) != 1 {
		t.Fatalf("expected 1 brew call, got %v", calls)
	}
	if !strings.Contains(calls[0], "bundle install") {
		t.Errorf("call = %q, want 'brew bundle install'", calls[0])
	}
	if !strings.Contains(calls[0], "--file=") {
		t.Errorf("call = %q, want --file flag", calls[0])
	}
}

func TestUnixInstaller_InstallBundle_MergesMultipleBrewfiles(t *testing.T) {
	binDir := installFakeBrew(t)
	i := &UnixInstaller{}

	bf1 := writeLocalBrewfile(t, "brew \"git\"\n")
	bf2 := writeLocalBrewfile(t, "brew \"jq\"\n")
	if err := i.InstallBundle(bf1, bf2); err != nil {
		t.Fatalf("InstallBundle: %v", err)
	}

	calls := readBrewCalls(t, binDir)
	if len(calls) != 1 {
		t.Fatalf("expected 1 brew call after merging, got %v", calls)
	}
}

func TestUnixInstaller_InstallBundle_UnknownBundle(t *testing.T) {
	installFakeBrew(t)
	i := &UnixInstaller{}

	err := i.InstallBundle("no-such-bundle-name")
	if err == nil {
		t.Fatal("expected error for unknown bundle")
	}
	if !strings.Contains(err.Error(), "unknown bundle") {
		t.Errorf("error = %v, want 'unknown bundle'", err)
	}
}

// ── InstallWallpapers ──────────────────────────────────────────────────────

func TestUnixInstaller_InstallWallpapers_NoCasks(t *testing.T) {
	installFakeBrew(t)
	i := &UnixInstaller{}

	err := i.InstallWallpapers(nil)
	if err == nil {
		t.Fatal("expected error with no casks selected")
	}
	if !strings.Contains(err.Error(), "no wallpaper casks selected") {
		t.Errorf("error = %v, want 'no wallpaper casks selected'", err)
	}
}

func TestUnixInstaller_InstallWallpapers_QualifiesTap(t *testing.T) {
	binDir := installFakeBrew(t)
	i := &UnixInstaller{}

	if err := i.InstallWallpapers([]string{"bluefin-wallpapers"}); err != nil {
		t.Fatalf("InstallWallpapers: %v", err)
	}

	calls := readBrewCalls(t, binDir)
	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "brew tap ublue-os/tap") {
		t.Errorf("missing brew tap call, got: %v", calls)
	}
	if !strings.Contains(joined, "brew install --cask ublue-os/tap/bluefin-wallpapers") {
		t.Errorf("cask not qualified with tap, got: %v", calls)
	}
}

func TestUnixInstaller_InstallWallpapers_KeepsQualifiedCask(t *testing.T) {
	binDir := installFakeBrew(t)
	i := &UnixInstaller{}

	if err := i.InstallWallpapers([]string{"some-other/tap/cask"}); err != nil {
		t.Fatalf("InstallWallpapers: %v", err)
	}

	joined := strings.Join(readBrewCalls(t, binDir), "\n")
	if !strings.Contains(joined, "brew install --cask some-other/tap/cask") {
		t.Errorf("already-qualified cask should be passed as-is, got: %v", joined)
	}
}

// ── CleanupWallpapers ──────────────────────────────────────────────────────

func TestUnixInstaller_CleanupWallpapers_NonAllIsNoop(t *testing.T) {
	installFakeBrew(t)
	i := &UnixInstaller{}

	if err := i.CleanupWallpapers(false); err != nil {
		t.Fatalf("CleanupWallpapers(false): %v", err)
	}
	if calls := readBrewCalls(t, ""); len(calls) > 0 {
		t.Errorf("CleanupWallpapers(false) ran %d brew commands; must be a no-op", len(calls))
	}
}

func TestUnixInstaller_CleanupWallpapers_AllUninstalls(t *testing.T) {
	binDir := installFakeBrew(t)
	i := &UnixInstaller{}

	if err := i.CleanupWallpapers(true); err != nil {
		t.Fatalf("CleanupWallpapers(true): %v", err)
	}

	joined := strings.Join(readBrewCalls(t, binDir), "\n")
	for _, cask := range []string{"bluefin-wallpapers", "aurora-wallpapers", "bazzite-wallpapers"} {
		if !strings.Contains(joined, "ublue-os/tap/"+cask) {
			t.Errorf("missing uninstall of %s, calls: %v", cask, joined)
		}
	}
	if !strings.Contains(joined, "brew uninstall --cask") {
		t.Errorf("missing brew uninstall --cask call, calls: %v", joined)
	}
}
