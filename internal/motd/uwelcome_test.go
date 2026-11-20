package motd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeUwelcome puts a uwelcome on PATH that logs its arguments, and points
// the user config dir at a temp dir. It returns the log and the disabled file.
func fakeUwelcome(t *testing.T) (log, disabled string) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("uwelcome is a Linux greeting; os.UserConfigDir only follows XDG_CONFIG_HOME there")
	}
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	log = filepath.Join(root, "calls")
	script := "#!/bin/sh\necho \"uwelcome $*\" >>" + log + "\n"
	if err := os.WriteFile(filepath.Join(bin, "uwelcome"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	return log, filepath.Join(root, "config", "uwelcome", "disabled")
}

func calls(t *testing.T, log string) string {
	t.Helper()
	data, err := os.ReadFile(log)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(data)
}

func TestUwelcomeAbsent(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if Uwelcome() {
		t.Error("Uwelcome() = true with no uwelcome on PATH")
	}
}

func TestUwelcomeEnabledFollowsItsDisabledFile(t *testing.T) {
	_, disabled := fakeUwelcome(t)
	if !Uwelcome() {
		t.Fatal("Uwelcome() = false with uwelcome on PATH")
	}
	if !UwelcomeEnabled() {
		t.Error("UwelcomeEnabled() = false without a disabled file")
	}
	if err := os.MkdirAll(filepath.Dir(disabled), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(disabled, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if UwelcomeEnabled() {
		t.Error("UwelcomeEnabled() = true with ~/.config/uwelcome/disabled present")
	}
}

func TestSetUwelcomeDrivesUwelcome(t *testing.T) {
	log, _ := fakeUwelcome(t)
	if err := SetUwelcome(false); err != nil {
		t.Fatal(err)
	}
	if err := SetUwelcome(true); err != nil {
		t.Fatal(err)
	}
	if got, want := calls(t, log), "uwelcome disable\nuwelcome enable\n"; got != want {
		t.Errorf("uwelcome calls = %q, want %q", got, want)
	}
}

func TestShowDelegatesToUwelcome(t *testing.T) {
	log, disabled := fakeUwelcome(t)
	if err := Show(); err != nil {
		t.Fatal(err)
	}
	if got, want := calls(t, log), "uwelcome \n"; got != want {
		t.Errorf("uwelcome calls = %q, want %q (one greeting, no bluefin-cli MOTD)", got, want)
	}

	// Turned off: uwelcome would print nothing, so it is not run at all.
	if err := os.MkdirAll(filepath.Dir(disabled), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(disabled, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(log); err != nil {
		t.Fatal(err)
	}
	if err := Show(); err != nil {
		t.Fatal(err)
	}
	if got := calls(t, log); got != "" {
		t.Errorf("uwelcome calls = %q with the greeting off, want none", got)
	}
}
