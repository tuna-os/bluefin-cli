package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// runTheme runs themeCmd's RunE directly. (Calling themeCmd.Execute() here is
// avoided: cobra walks back up to the root command, which defaults to the
// interactive bubbletea menu and requires a TTY.)
func runTheme(t *testing.T, args []string) error {
	t.Helper()
	return themeCmd.RunE(themeCmd, args)
}

// TestThemeCommand_Registered ensures the theme command is wired into the
// root command tree and accepts exactly the supported Catppuccin flavors.
func TestThemeCommand_Registered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"theme"})
	if err != nil || cmd == rootCmd || cmd == nil {
		t.Fatal("theme command not found under root")
	}
	flavors := theme.Flavors()
	if len(cmd.ValidArgs) != len(flavors) {
		t.Errorf("theme ValidArgs = %v, want %v", cmd.ValidArgs, flavors)
	}
	for _, f := range flavors {
		if !slicesContains(cmd.ValidArgs, f) {
			t.Errorf("theme ValidArgs missing supported flavor %q", f)
		}
	}
}

// TestThemeCommand_ShowFlavor exercises the read path (no args) — it must
// complete without error.
func TestThemeCommand_ShowFlavor(t *testing.T) {
	if err := runTheme(t, nil); err != nil {
		t.Fatalf("theme with no args failed: %v", err)
	}
}

// TestThemeCommand_UnknownFlavor_Errors verifies invalid flavors are
// rejected before any config mutation happens.
func TestThemeCommand_UnknownFlavor_Errors(t *testing.T) {
	err := runTheme(t, []string{"neon"})
	if err == nil {
		t.Fatal("expected an error for unknown flavor")
	}
	if !strings.Contains(err.Error(), `unknown flavor "neon"`) {
		t.Errorf("error = %q, want unknown-flavor message", err.Error())
	}
	if !strings.Contains(err.Error(), "options:") {
		t.Errorf("error should list supported options: %q", err.Error())
	}
}

// TestThemeCommand_SetFlavor_Persists sets a valid flavor and verifies the
// config write path (isolated HOME so nothing outside the test is touched).
func TestThemeCommand_SetFlavor_Persists(t *testing.T) {
	home := t.TempDir()
	// os.UserHomeDir() honours $USERPROFILE on Windows and $HOME elsewhere;
	// set whichever the platform uses so the config write lands in the
	// isolated temp dir on every OS.
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	} else {
		t.Setenv("HOME", home)
	}
	t.Cleanup(config.Reset)

	if err := runTheme(t, []string{"mocha"}); err != nil {
		t.Fatalf("theme mocha failed: %v", err)
	}
	if got := config.GetUIFlavor(); got != "mocha" {
		t.Errorf("config ui.flavor = %q, want mocha", got)
	}
	cfg := filepath.Join(home, ".config", "bluefin-cli", "config.yaml")
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatalf("expected config file at %s: %v", cfg, err)
	}
	if !strings.Contains(string(data), "mocha") {
		t.Errorf("config file does not contain the pinned flavor:\n%s", data)
	}
}

// TestApplyThemeFlavor_Noop guards the auto/empty fast path: with no pinned
// flavor the theme package must keep its defaults and the call must not panic.
func TestApplyThemeFlavor_Noop(t *testing.T) {
	t.Cleanup(config.Reset)
	config.SetUIFlavor("auto")
	applyThemeFlavor() // must not panic, must not change the global flavor
	if theme.Flavor != "" {
		t.Errorf("theme.Flavor = %q after auto no-op, want empty", theme.Flavor)
	}
}

// TestApplyThemeFlavor_PinsFlavor verifies a configured flavor is applied to
// the theme package.
func TestApplyThemeFlavor_PinsFlavor(t *testing.T) {
	t.Cleanup(func() {
		config.Reset()
		theme.Flavor = ""
		theme.DefaultTheme = theme.Resolve(true)
	})
	config.SetUIFlavor("mocha")
	applyThemeFlavor()
	if theme.Flavor != "mocha" {
		t.Errorf("theme.Flavor = %q, want mocha", theme.Flavor)
	}
}

func slicesContains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
