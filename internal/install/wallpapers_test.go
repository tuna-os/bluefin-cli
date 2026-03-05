package install

import (
	"errors"
	"testing"
	"time"
)

func TestPostInstallWallpaperSetup_WSLGate(t *testing.T) {
	originalIsWSL := isWSL
	originalSync := syncWallpapersToWindowsWSL
	defer func() {
		isWSL = originalIsWSL
		syncWallpapersToWindowsWSL = originalSync
	}()

	called := false
	isWSL = func() bool { return false }
	syncWallpapersToWindowsWSL = func(casks []string) error {
		called = true
		return nil
	}

	postInstallWallpaperSetup([]string{"bluefin-wallpaper"})
	if called {
		t.Fatal("expected Windows sync not to run when not in WSL")
	}

	isWSL = func() bool { return true }
	postInstallWallpaperSetup([]string{"bluefin-wallpaper"})
	if !called {
		t.Fatal("expected Windows sync to run in WSL")
	}
}

func TestNormalizeCaskName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "bluefin-wallpaper", want: "bluefin-wallpaper"},
		{input: "ublue-os/tap/bluefin-wallpaper", want: "bluefin-wallpaper"},
	}

	for _, tt := range tests {
		if got := normalizeCaskName(tt.input); got != tt.want {
			t.Fatalf("normalizeCaskName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDetectThemeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: "bluefin-wallpaper", want: "Bluefin", ok: true},
		{input: "aurora-dynamic-wallpaper", want: "Aurora", ok: true},
		{input: "bazzite-wallpaper", want: "Bazzite", ok: true},
		{input: "some-other-wallpaper", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := detectThemeName(tt.input)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("detectThemeName(%q) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

func TestThemesFromWallpaperCasks(t *testing.T) {
	casks := []string{
		"ublue-os/tap/bluefin-wallpapers",
		"aurora-wallpapers",
		"bazzite-wallpapers",
		"bluefin-wallpapers",
		"something-else",
	}

	themes := ThemesFromWallpaperCasks(casks)
	if len(themes) != 3 {
		t.Fatalf("expected 3 themes, got %d (%v)", len(themes), themes)
	}
}

func TestSelectMonthlyWallpaper(t *testing.T) {
	paths := []string{
		"/tmp/01-bluefin-day.png",
		"/tmp/01-bluefin-night.png",
		"/tmp/02-bluefin-day.png",
		"/tmp/02-bluefin-night.png",
	}

	day := time.Date(2026, time.February, 10, 9, 0, 0, 0, time.UTC)
	if got := selectMonthlyWallpaper(paths, day); got != "/tmp/02-bluefin-day.png" {
		t.Fatalf("expected February day wallpaper, got %s", got)
	}

	night := time.Date(2026, time.February, 10, 21, 0, 0, 0, time.UTC)
	if got := selectMonthlyWallpaper(paths, night); got != "/tmp/02-bluefin-night.png" {
		t.Fatalf("expected February night wallpaper, got %s", got)
	}
}

func TestSupportsMonthlyWallpapers(t *testing.T) {
	monthly := []string{
		"/tmp/01-bluefin-day.png",
		"/tmp/01-bluefin-night.png",
		"/tmp/02-bluefin-day.png",
	}
	if !supportsMonthlyWallpapers(monthly) {
		t.Fatal("expected monthly wallpapers to be detected")
	}

	nonMonthly := []string{
		"/tmp/bazzite-main.png",
		"/tmp/bazzite-alt.png",
	}
	if supportsMonthlyWallpapers(nonMonthly) {
		t.Fatal("expected non-monthly wallpapers not to be detected")
	}
}

func TestIsWindowsAccessDenied(t *testing.T) {
	if isWindowsAccessDenied(nil) {
		t.Fatal("expected nil error not to be access denied")
	}

	if !isWindowsAccessDenied(errors.New("ERROR: Access is denied.")) {
		t.Fatal("expected access denied error to be detected")
	}

	if isWindowsAccessDenied(errors.New("some other scheduler error")) {
		t.Fatal("expected non-access denied error not to be detected")
	}
}

func TestConfigureWindowsThemeAutomation_AccessDeniedIsNonFatal(t *testing.T) {
	originalLookPath := lookPathWSL
	originalRegister := registerWindowsTaskWSL
	originalRun := runWindowsTaskWSL
	originalDelete := deleteWindowsTaskWSL
	originalEnsureStartup := ensureStartupRunEntryWSL
	originalDeleteStartup := deleteStartupRunEntryWSL
	defer func() {
		lookPathWSL = originalLookPath
		registerWindowsTaskWSL = originalRegister
		runWindowsTaskWSL = originalRun
		deleteWindowsTaskWSL = originalDelete
		ensureStartupRunEntryWSL = originalEnsureStartup
		deleteStartupRunEntryWSL = originalDeleteStartup
	}()

	lookPathWSL = func(file string) (string, error) { return "/mnt/c/Windows/System32/schtasks.exe", nil }
	registerWindowsTaskWSL = func(taskName string, scheduleArgs []string, taskCommand string) error {
		return nil
	}
	runWindowsTaskWSL = func(taskName string) error { return nil }
	deleteWindowsTaskWSL = func(taskName string) error { return nil }
	ensureStartupRunEntryWSL = func(valueName, command string) error {
		return errors.New("ERROR: Access is denied.")
	}
	deleteStartupRunEntryWSL = func(valueName string) error { return nil }

	if err := ConfigureWindowsThemeAutomation(false); err != nil {
		t.Fatalf("expected access denied to be non-fatal, got error: %v", err)
	}
}

func TestConfigureWindowsThemeAutomation_UnexpectedErrorIsFatal(t *testing.T) {
	originalLookPath := lookPathWSL
	originalRegister := registerWindowsTaskWSL
	originalRun := runWindowsTaskWSL
	originalDelete := deleteWindowsTaskWSL
	originalEnsureStartup := ensureStartupRunEntryWSL
	originalDeleteStartup := deleteStartupRunEntryWSL
	defer func() {
		lookPathWSL = originalLookPath
		registerWindowsTaskWSL = originalRegister
		runWindowsTaskWSL = originalRun
		deleteWindowsTaskWSL = originalDelete
		ensureStartupRunEntryWSL = originalEnsureStartup
		deleteStartupRunEntryWSL = originalDeleteStartup
	}()

	lookPathWSL = func(file string) (string, error) { return "/mnt/c/Windows/System32/schtasks.exe", nil }
	registerWindowsTaskWSL = func(taskName string, scheduleArgs []string, taskCommand string) error {
		if taskName == taskThemeModeSync {
			return errors.New("ERROR: The parameter is incorrect.")
		}
		return nil
	}
	runWindowsTaskWSL = func(taskName string) error { return nil }
	deleteWindowsTaskWSL = func(taskName string) error { return nil }
	ensureStartupRunEntryWSL = func(valueName, command string) error { return nil }
	deleteStartupRunEntryWSL = func(valueName string) error { return nil }

	err := ConfigureWindowsThemeAutomation(false)
	if err == nil {
		t.Fatal("expected non-access-denied error to be returned")
	}
}
