package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// setTestHome points os.UserHomeDir at dir for the duration of the test.
//
// Setting HOME alone is not enough: on Windows os.UserHomeDir reads
// %USERPROFILE% and ignores HOME entirely, so these tests failed the
// windows job with "%userprofile% is not defined" while passing on Linux.
// Both are set so the same test body works on either platform.
func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	}
}

func TestNormalizeShell(t *testing.T) {
	cases := map[string]string{
		"bash":    "bash",
		"BASH":    "bash",
		" zsh ":   "zsh",
		"fish":    "fish",
		"ash":     "ash",
		"sh":      "ash",
		"dash":    "ash",
		"nu":      "nu",
		"nushell": "nu",
		"NuShell": "nu",
		"pwsh":    "powershell",
		"elvish":  "elvish", // unknown names pass through lowercased
	}
	for in, want := range cases {
		if got := NormalizeShell(in); got != want {
			t.Errorf("NormalizeShell(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLookupShell_KnownAndUnknown(t *testing.T) {
	for _, name := range []string{"bash", "zsh", "fish", "ash", "sh", "nu", "nushell"} {
		if _, ok := LookupShell(name); !ok {
			t.Errorf("LookupShell(%q) reported unknown", name)
		}
	}
	if _, ok := LookupShell("elvish"); ok {
		t.Error("LookupShell(\"elvish\") should report unknown")
	}
	// PowerShell is handled by powershell.go, not the registry.
	if _, ok := LookupShell("powershell"); ok {
		t.Error("powershell should not be in the rc-file registry")
	}
}

func TestManagedShells_IncludesAshAndNu(t *testing.T) {
	got := strings.Join(ManagedShells(), ",")
	for _, want := range []string{"bash", "zsh", "fish", "ash", "nu"} {
		if !strings.Contains(got, want) {
			t.Errorf("ManagedShells() = %v, missing %q", got, want)
		}
	}
}

// TestInit_NuUsesNushellSyntax the nushell preamble must assign through
// `$env.NAME = "0|1"`; a POSIX `export` line is a parse error in nushell and
// would break every startup.
func TestInit_NuUsesNushellSyntax(t *testing.T) {
	setTestHome(t, t.TempDir())
	script, err := Init("nu", DefaultConfig("nu"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, `$env.BLUEFIN_SHELL_ENABLE_MOTD = `) {
		t.Errorf("nu init script has no $env assignment for MOTD:\n%s", firstLines(script, 12))
	}
	if !strings.Contains(script, `$env.BLING_SHELL = "nu"`) {
		t.Errorf("nu init script does not set BLING_SHELL:\n%s", firstLines(script, 12))
	}
	for _, forbidden := range []string{"export BLUEFIN_", "set -gx "} {
		if strings.Contains(script, forbidden) {
			t.Errorf("nu init script contains non-nushell syntax %q", forbidden)
		}
	}
}

// TestInit_AshUsesPosixSyntaxAndScript ash shares the POSIX script but must
// identify itself, so the script can skip the tools with no ash init target.
func TestInit_AshUsesPosixSyntaxAndScript(t *testing.T) {
	setTestHome(t, t.TempDir())
	script, err := Init("ash", DefaultConfig("ash"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, `export BLING_SHELL="ash"`) {
		t.Errorf("ash init script does not set BLING_SHELL=ash:\n%s", firstLines(script, 12))
	}
	if !strings.Contains(script, "BLUEFIN_SHELL_HAS_NATIVE_INIT") {
		t.Error("ash init script is missing the native-init guard from shell.sh")
	}
	if !strings.Contains(script, `ZOXIDE_TARGET="posix"`) {
		t.Error("ash init script does not remap zoxide onto its posix target")
	}
}

// TestInit_AliasRendersCanonicalName `init sh` and `init nushell` must render
// the same script as their canonical names rather than a third variant.
func TestInit_AliasRendersCanonicalName(t *testing.T) {
	setTestHome(t, t.TempDir())
	for alias, canonical := range map[string]string{"sh": "ash", "nushell": "nu"} {
		viaAlias, err := Init(alias, DefaultConfig(alias))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(viaAlias, canonical) {
			t.Errorf("init %q did not render as %q:\n%s", alias, canonical, firstLines(viaAlias, 6))
		}
	}
}

// TestToggle_AshWiresUpENV enabling ash has to do two things: write ~/.ashrc
// and export ENV from ~/.profile. Without the second, ash never reads the
// first and the feature silently does nothing.
func TestToggle_AshWiresUpENV(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)

	if err := Toggle("ash", true); err != nil {
		t.Fatalf("Toggle(ash, true): %v", err)
	}

	ashrc := readFile(t, filepath.Join(home, ".ashrc"))
	if !strings.Contains(ashrc, "bluefin-cli init ash") {
		t.Errorf("~/.ashrc has no init line:\n%s", ashrc)
	}
	profile := readFile(t, filepath.Join(home, ".profile"))
	if !strings.Contains(profile, `ENV="$HOME/.ashrc"`) {
		t.Errorf("~/.profile does not export ENV:\n%s", profile)
	}

	// Enabling twice must not duplicate the ENV line.
	if err := Toggle("ash", true); err != nil {
		t.Fatalf("second Toggle(ash, true): %v", err)
	}
	if n := strings.Count(readFile(t, filepath.Join(home, ".profile")), "ENV="); n != 1 {
		t.Errorf("~/.profile has %d ENV lines, want 1", n)
	}

	if err := Toggle("ash", false); err != nil {
		t.Fatalf("Toggle(ash, false): %v", err)
	}
	if got := readFile(t, filepath.Join(home, ".profile")); strings.Contains(got, "ENV=") {
		t.Errorf("disable left the ENV line behind:\n%s", got)
	}
}

// TestToggle_AshPreservesUnrelatedProfileLines disabling must only remove the
// line bluefin-cli wrote, never a user's own $ENV setup.
func TestToggle_AshPreservesUnrelatedProfileLines(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)

	userLine := `export ENV="$HOME/.my-own-rc"`
	writeFile(t, filepath.Join(home, ".profile"), userLine+"\n")

	if err := Toggle("ash", true); err != nil {
		t.Fatal(err)
	}
	if err := Toggle("ash", false); err != nil {
		t.Fatal(err)
	}

	if got := readFile(t, filepath.Join(home, ".profile")); !strings.Contains(got, userLine) {
		t.Errorf("disable removed the user's own ENV line:\n%s", got)
	}
}

// TestToggle_NuRendersSourcedFile nushell sources a real file, so enabling has
// to produce that file and a config.nu line pointing at it. A config.nu that
// sources a file which was never written breaks every nushell start.
func TestToggle_NuRendersSourcedFile(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)

	if err := Toggle("nu", true); err != nil {
		t.Fatalf("Toggle(nu, true): %v", err)
	}

	generated := filepath.Join(home, ".config", "nushell", "bluefin-cli.nu")
	script := readFile(t, generated)
	if !strings.Contains(script, "$env.BLING_SHELL") {
		t.Errorf("generated nu script is not a nushell init script:\n%s", firstLines(script, 10))
	}

	configNu := readFile(t, filepath.Join(home, ".config", "nushell", "config.nu"))
	if !strings.Contains(configNu, "source "+generated) {
		t.Errorf("config.nu does not source the generated script:\n%s", configNu)
	}

	if err := Toggle("nu", false); err != nil {
		t.Fatalf("Toggle(nu, false): %v", err)
	}
	if _, err := os.Stat(generated); !os.IsNotExist(err) {
		t.Error("disable left the generated nu script on disk")
	}
}

// TestToggle_UnsupportedShellNamesTheSupportedOnes a typo should say what is
// actually accepted rather than just rejecting the input.
func TestToggle_UnsupportedShellNamesTheSupportedOnes(t *testing.T) {
	setTestHome(t, t.TempDir())
	err := Toggle("elvish", true)
	if err == nil {
		t.Fatal("Toggle on an unsupported shell should fail")
	}
	if !strings.Contains(err.Error(), "ash") || !strings.Contains(err.Error(), "nu") {
		t.Errorf("error does not list the supported shells: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
