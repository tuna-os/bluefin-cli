package terminal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// setHomeEnv points os.UserHomeDir() at dir on every platform: Unix reads
// HOME, Windows reads USERPROFILE.
func setHomeEnv(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestPreferredName(t *testing.T) {
	got := PreferredName()
	want := "Ghostty"
	if runtime.GOOS == "windows" {
		want = "WezTerm"
	}
	if got != want {
		t.Errorf("PreferredName() = %q, want %q", got, want)
	}
}

func TestGhosttyConfigDir(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	got, err := ghosttyConfigDir()
	if err != nil {
		t.Fatalf("ghosttyConfigDir() error = %v", err)
	}
	want := filepath.Join(home, ".config", "ghostty")
	if got != want {
		t.Errorf("ghosttyConfigDir() = %q, want %q", got, want)
	}
}

func TestDetectNerdFontFound(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	fontDir := filepath.Join(home, ".local", "share", "fonts")
	if err := os.MkdirAll(fontDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fontDir, "HackNerdFont-Regular.ttf"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := DetectNerdFont()
	if got != "Hack Nerd Font" {
		t.Errorf("DetectNerdFont() = %q, want %q", got, "Hack Nerd Font")
	}
}

func TestDetectNerdFontNone(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	if got := DetectNerdFont(); got != "" {
		t.Errorf("DetectNerdFont() = %q, want empty", got)
	}
}

func TestWriteGhosttyConfigUnix(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)

	if err := writeGhosttyConfigUnix(); err != nil {
		t.Fatalf("writeGhosttyConfigUnix() error = %v", err)
	}
	cfgPath := filepath.Join(home, ".config", "ghostty", "config")
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if !strings.Contains(string(content), markerStart) || !strings.Contains(string(content), markerEnd) {
		t.Errorf("config missing managed block markers: %s", content)
	}
	if !strings.Contains(string(content), "catppuccin-mocha") {
		t.Errorf("config missing expected theme content: %s", content)
	}

	// Writing again should replace the old managed block, not duplicate it.
	if err := writeGhosttyConfigUnix(); err != nil {
		t.Fatalf("second writeGhosttyConfigUnix() error = %v", err)
	}
	content2, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if strings.Count(string(content2), markerStart) != 1 {
		t.Errorf("expected exactly one managed block after rewrite, got content: %s", content2)
	}
}

func TestWriteGhosttyConfigUnixPreservesUserContent(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	dir := filepath.Join(home, ".config", "ghostty")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "config")
	if err := os.WriteFile(cfgPath, []byte("font-size = 14\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := writeGhosttyConfigUnix(); err != nil {
		t.Fatalf("writeGhosttyConfigUnix() error = %v", err)
	}
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "font-size = 14") {
		t.Errorf("existing user config was dropped: %s", content)
	}
	if !strings.Contains(string(content), markerStart) {
		t.Errorf("managed block not appended: %s", content)
	}
}

func TestWriteWezTermConfigFresh(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)

	if err := writeWezTermConfig(); err != nil {
		t.Fatalf("writeWezTermConfig() error = %v", err)
	}
	cfgPath := filepath.Join(home, ".wezterm.lua")
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if !strings.Contains(string(content), "Catppuccin Mocha") {
		t.Errorf("config missing expected theme content: %s", content)
	}
}

func TestWriteWezTermConfigLeavesExisting(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)
	cfgPath := filepath.Join(home, ".wezterm.lua")
	if err := os.WriteFile(cfgPath, []byte("-- my config\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := writeWezTermConfig(); err != nil {
		t.Fatalf("writeWezTermConfig() error = %v", err)
	}
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "-- my config\n" {
		t.Errorf("existing wezterm config was overwritten: %s", content)
	}
}

func TestWriteGhosttyConfigDispatch(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)

	if err := WriteGhosttyConfig(); err != nil {
		t.Fatalf("WriteGhosttyConfig() error = %v", err)
	}
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(filepath.Join(home, ".wezterm.lua")); err != nil {
			t.Errorf("expected wezterm config to be written: %v", err)
		}
		return
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "ghostty", "config")); err != nil {
		t.Errorf("expected ghostty config to be written: %v", err)
	}
}

func TestGhosttyInstalledFalseWhenAbsent(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("PATH-based LookPath stub is Linux-specific")
	}
	emptyDir := t.TempDir()
	t.Setenv("PATH", emptyDir)
	if GhosttyInstalled() {
		t.Error("GhosttyInstalled() = true, want false with empty PATH")
	}
}

func TestGhosttyInstalledTrueWhenOnPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("PATH-based LookPath stub is Linux-specific")
	}
	binDir := t.TempDir()
	stub := filepath.Join(binDir, "ghostty")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	if !GhosttyInstalled() {
		t.Error("GhosttyInstalled() = false, want true when ghostty is on PATH")
	}
}

func TestPinToDockNonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("darwin-only guard is bypassed on darwin")
	}
	if err := PinToDock("/Applications/Ghostty.app"); err == nil {
		t.Error("PinToDock() on non-darwin should error, got nil")
	}
}

func TestInstallGhosttyAlreadyInstalled(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("PATH-based LookPath stub is Linux-specific")
	}
	binDir := t.TempDir()
	stub := filepath.Join(binDir, "ghostty")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	if err := InstallGhostty(); err != nil {
		t.Errorf("InstallGhostty() error = %v, want nil when already installed", err)
	}
}

func TestDetectTerminals(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)

	if got := DetectTerminals(); len(got) != 0 {
		t.Errorf("DetectTerminals() on empty HOME = %v, want empty", got)
	}

	if err := os.MkdirAll(filepath.Join(home, ".config", "alacritty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "alacritty", "alacritty.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".wezterm.lua"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeGhosttyConfigUnix(); err != nil {
		t.Fatal(err)
	}

	got := DetectTerminals()
	names := map[string]bool{}
	for _, term := range got {
		names[term.Name] = true
		if !term.IsReady || term.ConfigDir == "" {
			t.Errorf("terminal %q should be ready with a config dir, got %+v", term.Name, term)
		}
	}
	for _, want := range []string{"Alacritty", "WezTerm", "Ghostty"} {
		if !names[want] {
			t.Errorf("DetectTerminals() missing %q, got %v", want, got)
		}
	}
}

func TestSetTerminalFontDispatch(t *testing.T) {
	home := t.TempDir()
	setHomeEnv(t, home)

	if err := SetTerminalFont(ConfiguredTerminal{Name: "Ghostty"}, "JetBrainsMono Nerd Font"); err != nil {
		t.Errorf("SetTerminalFont(Ghostty) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "ghostty", "config")); err != nil {
		t.Errorf("expected ghostty config to be written: %v", err)
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if err := SetTerminalFont(ConfiguredTerminal{Name: "SomeOtherTerminal"}, "Hack Nerd Font"); err != nil {
			t.Errorf("SetTerminalFont(default) error = %v, want nil from platform fallback", err)
		}
	}
}

func TestInstallGhosttyMissingBrew(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("PATH-based LookPath stub is Linux-specific")
	}
	emptyDir := t.TempDir()
	t.Setenv("PATH", emptyDir)
	err := InstallGhostty()
	if err == nil {
		t.Fatal("InstallGhostty() error = nil, want error when brew is missing")
	}
	if !strings.Contains(err.Error(), "homebrew") {
		t.Errorf("InstallGhostty() error = %v, want mention of homebrew", err)
	}
}
