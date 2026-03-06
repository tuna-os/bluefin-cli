package install

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestPostInstallWallpaperSetup_PlatformGate(t *testing.T) {
	originalIsWSL := isWSL
	originalIsWindows := isWindows
	originalSync := syncWallpapersToWindowsWSL
	originalWinSync := syncWallpapersToWindowsWin
	defer func() {
		isWSL = originalIsWSL
		isWindows = originalIsWindows
		syncWallpapersToWindowsWSL = originalSync
		syncWallpapersToWindowsWin = originalWinSync
	}()

	wslCalled := false
	winCalled := false
	isWSL = func() bool { return false }
	isWindows = func() bool { return false }
	syncWallpapersToWindowsWSL = func(casks []string) error {
		wslCalled = true
		return nil
	}
	syncWallpapersToWindowsWin = func(casks []string) error {
		winCalled = true
		return nil
	}

	postInstallWallpaperSetup([]string{"bluefin-wallpaper"})
	if wslCalled || winCalled {
		t.Fatal("expected no Windows sync path to run when not in WSL/Windows")
	}

	isWSL = func() bool { return true }
	postInstallWallpaperSetup([]string{"bluefin-wallpaper"})
	if !wslCalled {
		t.Fatal("expected Windows sync to run in WSL")
	}

	wslCalled = false
	winCalled = false
	isWSL = func() bool { return false }
	isWindows = func() bool { return true }
	postInstallWallpaperSetup([]string{"bluefin-wallpaper"})
	if !winCalled {
		t.Fatal("expected Windows sync to run on native Windows")
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

	if err := ConfigureWindowsThemeAutomation(false, ThemeSyncTriggerPolling); err != nil {
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

	err := ConfigureWindowsThemeAutomation(false, ThemeSyncTriggerPolling)
	if err == nil {
		t.Fatal("expected non-access-denied error to be returned")
	}
}

func TestSplitTaskCommand(t *testing.T) {
	execPath, args, err := splitTaskCommand(`powershellw.exe -NoProfile -File "C:\\path\\sync.ps1"`)
	if err != nil {
		t.Fatalf("splitTaskCommand returned error: %v", err)
	}
	if execPath != "powershellw.exe" {
		t.Fatalf("unexpected execute path: got %q", execPath)
	}
	if args != `-NoProfile -File "C:\\path\\sync.ps1"` {
		t.Fatalf("unexpected args: got %q", args)
	}

	if _, _, err := splitTaskCommand("   "); err == nil {
		t.Fatal("expected empty command to fail")
	}
}

func TestPowerShellTriggerScript(t *testing.T) {
	tests := []struct {
		name         string
		scheduleArgs []string
		contains     []string
		wantErr      string
	}{
		{
			name:         "minute trigger",
			scheduleArgs: []string{"/SC", "MINUTE", "/MO", "5"},
			contains: []string{
				"New-ScheduledTaskTrigger -Once",
				"New-TimeSpan -Minutes 5",
			},
		},
		{
			name:         "daily trigger",
			scheduleArgs: []string{"/SC", "DAILY", "/ST", "18:00"},
			contains: []string{
				"New-ScheduledTaskTrigger -Daily",
				"[TimeSpan]::Parse('18:00')",
			},
		},
		{
			name:         "onlogon trigger",
			scheduleArgs: []string{"/SC", "ONLOGON"},
			contains: []string{"New-ScheduledTaskTrigger -AtLogOn"},
		},
		{
			name:         "daily missing time",
			scheduleArgs: []string{"/SC", "DAILY"},
			wantErr:      "daily schedule missing /ST time",
		},
		{
			name:         "unsupported schedule",
			scheduleArgs: []string{"/SC", "WEEKLY"},
			wantErr:      "unsupported task schedule type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := powershellTriggerScript(tt.scheduleArgs)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("powershellTriggerScript returned error: %v", err)
			}

			for _, want := range tt.contains {
				if !strings.Contains(script, want) {
					t.Fatalf("expected script to contain %q, got %q", want, script)
				}
			}
		})
	}
}

func TestConfigureWindowsThemeAutomation_EnableAutoDarkLight_RegistersAndRuns(t *testing.T) {
	originalLookPath := lookPathWSL
	originalRegister := registerWindowsTaskWSL
	originalRun := runWindowsTaskWSL
	originalDelete := deleteWindowsTaskWSL
	originalEnsureStartup := ensureStartupRunEntryWSL
	originalDeleteStartup := deleteStartupRunEntryWSL
	originalNow := nowWSL
	defer func() {
		lookPathWSL = originalLookPath
		registerWindowsTaskWSL = originalRegister
		runWindowsTaskWSL = originalRun
		deleteWindowsTaskWSL = originalDelete
		ensureStartupRunEntryWSL = originalEnsureStartup
		deleteStartupRunEntryWSL = originalDeleteStartup
		nowWSL = originalNow
	}()

	lookPathWSL = func(file string) (string, error) { return "C:/Windows/System32/schtasks.exe", nil }

	registered := make(map[string][]string)
	registerWindowsTaskWSL = func(taskName string, scheduleArgs []string, taskCommand string) error {
		registered[taskName] = append([]string{}, scheduleArgs...)
		if strings.TrimSpace(taskCommand) == "" {
			return fmt.Errorf("task command for %s is empty", taskName)
		}
		if !strings.Contains(taskCommand, "powershellw.exe") {
			return fmt.Errorf("task command for %s should use powershellw.exe", taskName)
		}
		return nil
	}

	runCalls := []string{}
	runWindowsTaskWSL = func(taskName string) error {
		runCalls = append(runCalls, taskName)
		return nil
	}

	deleteCalls := []string{}
	deleteWindowsTaskWSL = func(taskName string) error {
		deleteCalls = append(deleteCalls, taskName)
		return nil
	}

	ensureStartupRunEntryWSL = func(valueName, command string) error {
		if valueName != startupRunValueName {
			return fmt.Errorf("unexpected startup value: %s", valueName)
		}
		if !strings.Contains(command, "theme-mode-sync.ps1") {
			return fmt.Errorf("unexpected startup command: %s", command)
		}
		return nil
	}
	deleteStartupRunEntryWSL = func(valueName string) error { return nil }

	// 8 PM should immediately run dark-mode task after registration.
	nowWSL = func() time.Time {
		return time.Date(2026, time.March, 6, 20, 0, 0, 0, time.UTC)
	}

	if err := ConfigureWindowsThemeAutomation(true, ThemeSyncTriggerPolling); err != nil {
		t.Fatalf("ConfigureWindowsThemeAutomation(true) returned error: %v", err)
	}

	if _, ok := registered[taskThemeModeSync]; !ok {
		t.Fatalf("expected %s to be registered", taskThemeModeSync)
	}
	if _, ok := registered[taskSetLightAt6AM]; !ok {
		t.Fatalf("expected %s to be registered", taskSetLightAt6AM)
	}
	if _, ok := registered[taskSetDarkAt6PM]; !ok {
		t.Fatalf("expected %s to be registered", taskSetDarkAt6PM)
	}

	if len(runCalls) < 2 {
		t.Fatalf("expected at least 2 task runs (sync + mode), got %d", len(runCalls))
	}
	if runCalls[0] != taskThemeModeSync {
		t.Fatalf("expected first run to be %s, got %s", taskThemeModeSync, runCalls[0])
	}
	if runCalls[1] != taskSetDarkAt6PM {
		t.Fatalf("expected second run to be %s at night, got %s", taskSetDarkAt6PM, runCalls[1])
	}

	if len(deleteCalls) != 0 {
		t.Fatalf("did not expect any cleanup task deletions when auto dark/light is enabled, got: %v", deleteCalls)
	}
}

func TestConfigureWindowsThemeAutomation_StartupTriggerSkipsPollingTask(t *testing.T) {
	originalLookPath := lookPathWSL
	originalRegister := registerWindowsTaskWSL
	originalRun := runWindowsTaskWSL
	originalDelete := deleteWindowsTaskWSL
	originalEnsureStartup := ensureStartupRunEntryWSL
	originalDeleteStartup := deleteStartupRunEntryWSL
	originalExec := execWSL
	defer func() {
		lookPathWSL = originalLookPath
		registerWindowsTaskWSL = originalRegister
		runWindowsTaskWSL = originalRun
		deleteWindowsTaskWSL = originalDelete
		ensureStartupRunEntryWSL = originalEnsureStartup
		deleteStartupRunEntryWSL = originalDeleteStartup
		execWSL = originalExec
	}()

	lookPathWSL = func(file string) (string, error) { return "C:/Windows/System32/schtasks.exe", nil }

	registered := []string{}
	registerWindowsTaskWSL = func(taskName string, scheduleArgs []string, taskCommand string) error {
		registered = append(registered, taskName)
		return nil
	}

	runWindowsTaskWSL = func(taskName string) error {
		return nil
	}

	deleteWindowsTaskWSL = func(taskName string) error { return nil }
	ensureStartupRunEntryWSL = func(valueName, command string) error { return nil }
	deleteStartupRunEntryWSL = func(valueName string) error { return nil }

	execWSL = func(name string, arg ...string) *exec.Cmd {
		// Simulate successful helper script checks and immediate script run.
		return exec.Command("cmd", "/c", "exit", "0")
	}

	if err := ConfigureWindowsThemeAutomation(false, ThemeSyncTriggerStartup); err != nil {
		t.Fatalf("ConfigureWindowsThemeAutomation(startup) returned error: %v", err)
	}

	for _, taskName := range registered {
		if taskName == taskThemeModeSync {
			t.Fatal("did not expect polling task registration in startup mode")
		}
	}
}

func TestParseThemeSyncTriggerSource(t *testing.T) {
	tests := []struct {
		input   string
		expect  ThemeSyncTriggerSource
		wantErr bool
	}{
		{input: "", expect: ThemeSyncTriggerPolling},
		{input: "polling", expect: ThemeSyncTriggerPolling},
		{input: "startup", expect: ThemeSyncTriggerStartup},
		{input: "autodarkmode", expect: ThemeSyncTriggerAutoDarkMode},
		{input: "AUTOdarkMODE", expect: ThemeSyncTriggerAutoDarkMode},
		{input: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		got, err := ParseThemeSyncTriggerSource(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("ParseThemeSyncTriggerSource(%q) expected error", tt.input)
			}
			continue
		}

		if err != nil {
			t.Fatalf("ParseThemeSyncTriggerSource(%q) returned error: %v", tt.input, err)
		}

		if got != tt.expect {
			t.Fatalf("ParseThemeSyncTriggerSource(%q) = %q, want %q", tt.input, got, tt.expect)
		}
	}
}
