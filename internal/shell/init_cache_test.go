package shell

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// An image-based system ships the binary with a fixed, ancient mtime, so a
// cache from the previous release always looks newer than the binary. The
// version stamp is what retires it.
func TestInitCacheIsKeyedOnVersion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	saved := Version
	t.Cleanup(func() { Version = saved })

	Version = "1.0.0"
	saveInitCache("bash", "echo old\n")
	path, err := getInitCachePath("bash")
	if err != nil {
		t.Fatal(err)
	}
	// Push the cache well past the test binary's mtime, as a stale cache on a
	// bootc system is.
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}

	got, ok := loadInitCache("bash")
	if !ok || got != "echo old\n" {
		t.Fatalf("same release: loadInitCache() = %q, %v; want the cached script", got, ok)
	}

	Version = "1.0.1"
	if got, ok := loadInitCache("bash"); ok {
		t.Fatalf("new release: loadInitCache() = %q, served a cache written by 1.0.0", got)
	}
}

// Caches written before the stamp existed have no header and must regenerate.
func TestInitCacheWithoutHeaderIsStale(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	path, err := getInitCachePath("zsh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("echo legacy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	if got, ok := loadInitCache("zsh"); ok {
		t.Fatalf("loadInitCache() = %q, served a cache with no release stamp", got)
	}
}
