//go:build extra

package cmd

import (
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/platform"
)

// TestSupportsWindowsThemePostInstall checks the platform guard stays in
// sync with the platform package's detection.
func TestSupportsWindowsThemePostInstall(t *testing.T) {
	got := supportsWindowsThemePostInstall()
	want := platform.GetCapabilities().SupportsSunset
	if got != want {
		t.Errorf("supportsWindowsThemePostInstall() = %v, want %v", got, want)
	}
}

// TestMaybeHandleWindowsThemePostInstall_NonWindows covers the early-return
// path: outside Windows/WSL the helper must be a no-op even with a nil cmd.
func TestMaybeHandleWindowsThemePostInstall_NonWindows(t *testing.T) {
	if supportsWindowsThemePostInstall() {
		t.Skip("Windows/WSL path would prompt for sunset setup")
	}
	if err := maybeHandleWindowsThemePostInstall(nil, nil); err != nil {
		t.Fatalf("expected nil error on non-Windows, got %v", err)
	}
}

// TestMaybeHandleWindowsThemePostInstall_YesFlag covers the --yes shortcut:
// sunset setup runs (non-blocking outside Windows) and returns nil.
func TestMaybeHandleWindowsThemePostInstall_YesFlag(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows path runs the interactive sunset form")
	}
	if !supportsWindowsThemePostInstall() {
		t.Skip("non-Windows early-return path covered elsewhere")
	}
	c := &cobra.Command{Use: "test"}
	c.Flags().Bool("non-interactive", false, "")
	c.Flags().Bool("yes", false, "")
	if err := c.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	if err := maybeHandleWindowsThemePostInstall(c, nil); err != nil {
		t.Fatalf("--yes path failed: %v", err)
	}
}

// TestMaybeHandleWindowsThemePostInstall_NonInteractive covers the
// non-interactive (but not --yes) path: must return nil without prompting.
func TestMaybeHandleWindowsThemePostInstall_NonInteractive(t *testing.T) {
	if !supportsWindowsThemePostInstall() {
		t.Skip("non-Windows early-return path covered elsewhere")
	}
	c := &cobra.Command{Use: "test"}
	c.Flags().Bool("non-interactive", false, "")
	c.Flags().Bool("yes", false, "")
	if err := c.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := maybeHandleWindowsThemePostInstall(c, nil); err != nil {
		t.Fatalf("non-interactive path failed: %v", err)
	}
}

// TestMaybeHandlePostFontInstall ensures the post-install hook is a no-op
// that never errors.
func TestMaybeHandlePostFontInstall(t *testing.T) {
	if err := maybeHandlePostFontInstall(); err != nil {
		t.Fatalf("maybeHandlePostFontInstall() = %v, want nil", err)
	}
}

// TestRunSunsetSetupFlow_NonWindows exercises the native-Linux guard inside
// the sunset setup flow (non-blocking outside Windows).
func TestRunSunsetSetupFlow_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows path runs the interactive sunset form")
	}
	if platform.IsWSL() {
		t.Skip("WSL path delegates to a Windows-side runner")
	}
	if err := RunSunsetSetupFlow(); err != nil {
		t.Fatalf("RunSunsetSetupFlow() = %v, want nil", err)
	}
}
