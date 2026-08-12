package install

// Tests for the Brewfile editing helpers (AddToBrewfile / RemoveFromBrewfile
// / HomeBrewfile). These manipulate the user's Brewfile — the untested
// module was the whole brewfile.go, which is pure file logic and runs
// without Homebrew installed.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddToBrewfile_RejectsInvalidKind(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Brewfile")
	err := AddToBrewfile(path, "git", "bogus")
	if err == nil {
		t.Fatal("AddToBrewfile with invalid kind expected error")
	}
	if !strings.Contains(err.Error(), "kind must be one of") {
		t.Errorf("error = %v, want 'kind must be one of'", err)
	}
	// The invalid call must not create the file.
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Error("invalid AddToBrewfile created the Brewfile")
	}
}

func TestAddToBrewfile_AppendsAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Brewfile")

	if err := AddToBrewfile(path, "git", "brew"); err != nil {
		t.Fatalf("AddToBrewfile(brew git): %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Brewfile: %v", err)
	}
	if string(data) != "brew \"git\"\n" {
		t.Errorf("Brewfile = %q, want %q", data, "brew \"git\"\n")
	}

	// Duplicate is a no-op.
	if err := AddToBrewfile(path, "git", "brew"); err != nil {
		t.Fatalf("AddToBrewfile duplicate: %v", err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "brew \"git\"\n" {
		t.Errorf("duplicate add changed Brewfile: %q", data)
	}

	// Different kind appends.
	if err := AddToBrewfile(path, "firefox", "cask"); err != nil {
		t.Fatalf("AddToBrewfile(cask firefox): %v", err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "brew \"git\"\ncask \"firefox\"\n" {
		t.Errorf("Brewfile = %q, want both entries", data)
	}
}

func TestAddToBrewfile_AppendsToExistingEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Brewfile")
	if err := os.WriteFile(path, []byte("tap \"foo/bar\"\nbrew \"jq\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AddToBrewfile(path, "git", "brew"); err != nil {
		t.Fatalf("AddToBrewfile: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "tap \"foo/bar\"\nbrew \"jq\"\nbrew \"git\"\n" {
		t.Errorf("Brewfile = %q, want appended without disturbing existing lines", data)
	}
}

func TestRemoveFromBrewfile_RemovesOnlyMatchingEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Brewfile")
	content := "tap \"foo/bar\"\nbrew \"git\"\nbrew \"jq\"\ncask \"firefox\"\n# keep me\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveFromBrewfile(path, []string{"git", "firefox"}); err != nil {
		t.Fatalf("RemoveFromBrewfile: %v", err)
	}
	data, _ := os.ReadFile(path)
	got := string(data)
	for _, want := range []string{"tap \"foo/bar\"", "brew \"jq\"", "# keep me"} {
		if !strings.Contains(got, want) {
			t.Errorf("removed line %q should remain: %q", want, got)
		}
	}
	for _, banned := range []string{"brew \"git\"", "cask \"firefox\""} {
		if strings.Contains(got, banned) {
			t.Errorf("line %q should have been removed: %q", banned, got)
		}
	}
}

func TestRemoveFromBrewfile_NoMatchErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Brewfile")
	if err := os.WriteFile(path, []byte("brew \"git\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := RemoveFromBrewfile(path, []string{"nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "no matching entries") {
		t.Errorf("RemoveFromBrewfile no-match error = %v, want 'no matching entries'", err)
	}
	// File must be untouched.
	data, _ := os.ReadFile(path)
	if string(data) != "brew \"git\"\n" {
		t.Errorf("failed removal modified the file: %q", data)
	}
}

func TestRemoveFromBrewfile_MissingFileErrors(t *testing.T) {
	err := RemoveFromBrewfile(filepath.Join(t.TempDir(), "nope"), []string{"git"})
	if err == nil {
		t.Error("RemoveFromBrewfile on missing file expected error")
	}
}

func TestHomeBrewfile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// No Brewfile yet → default path under HOME.
	got := HomeBrewfile()
	want := filepath.Join(tmpHome, "Brewfile")
	if got != want {
		t.Errorf("HomeBrewfile() = %q, want %q", got, want)
	}

	// Existing .Brewfile is preferred.
	dot := filepath.Join(tmpHome, ".Brewfile")
	if err := os.WriteFile(dot, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := HomeBrewfile(); got != dot {
		t.Errorf("HomeBrewfile() = %q, want existing .Brewfile %q", got, dot)
	}
}
