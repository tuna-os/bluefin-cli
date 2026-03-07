package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPowerShellProfilePaths(t *testing.T) {
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	if err := os.Setenv("USERPROFILE", tmpHome); err != nil {
		t.Fatalf("failed to set USERPROFILE: %v", err)
	}
	defer func() {
		if originalHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", originalHome)
		}

		if originalUserProfile == "" {
			_ = os.Unsetenv("USERPROFILE")
		} else {
			_ = os.Setenv("USERPROFILE", originalUserProfile)
		}
	}()

	profiles, err := powerShellProfilePaths()
	if err != nil {
		t.Fatalf("powerShellProfilePaths() returned error: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profile paths, got %d", len(profiles))
	}

	expectedPwsh := filepath.Join(tmpHome, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	expectedWindowsPowerShell := filepath.Join(tmpHome, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	if profiles[0] != expectedPwsh {
		t.Fatalf("unexpected pwsh profile path: got %q, want %q", profiles[0], expectedPwsh)
	}
	if profiles[1] != expectedWindowsPowerShell {
		t.Fatalf("unexpected Windows PowerShell profile path: got %q, want %q", profiles[1], expectedWindowsPowerShell)
	}
}

func TestSyncPowerShellProfileEnableDisable(t *testing.T) {
	tmpDir := t.TempDir()
	profile := filepath.Join(tmpDir, "Microsoft.PowerShell_profile.ps1")
	rcLine := `if (Get-Command bluefin-cli -ErrorAction SilentlyContinue) { Invoke-Expression ((& bluefin-cli init powershell | Out-String)) } # bluefin-cli shell-config`

	changed, err := syncPowerShellProfile(profile, rcLine, true)
	if err != nil {
		t.Fatalf("syncPowerShellProfile(enable) returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected enable to report changed=true")
	}

	content, err := os.ReadFile(profile)
	if err != nil {
		t.Fatalf("failed to read profile after enable: %v", err)
	}
	if !strings.Contains(string(content), shellMaker) {
		t.Fatalf("expected profile to contain %q", shellMaker)
	}

	changed, err = syncPowerShellProfile(profile, rcLine, true)
	if err != nil {
		t.Fatalf("syncPowerShellProfile(second enable) returned error: %v", err)
	}
	if changed {
		t.Fatalf("expected second enable to report changed=false")
	}

	changed, err = syncPowerShellProfile(profile, rcLine, false)
	if err != nil {
		t.Fatalf("syncPowerShellProfile(disable) returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected disable to report changed=true")
	}

	content, err = os.ReadFile(profile)
	if err != nil {
		t.Fatalf("failed to read profile after disable: %v", err)
	}
	if strings.Contains(string(content), shellMaker) {
		t.Fatalf("expected profile to not contain %q after disable", shellMaker)
	}
}

func TestSyncPowerShellProfileCleansLegacyLines(t *testing.T) {
	tmpDir := t.TempDir()
	profile := filepath.Join(tmpDir, "Microsoft.PowerShell_profile.ps1")
	rcLine := `if (Get-Command bluefin-cli -ErrorAction SilentlyContinue) { Invoke-Expression ((& bluefin-cli init powershell | Out-String)) } # bluefin-cli shell-config`

	legacy := strings.Join([]string{
		"if (Test-Path 'C:\\Users\\james\\dev\\bluefin-cli\\bluefin-cli.exe')",
		"    { Invoke-Expression (& 'C:\\Users\\james\\dev\\bluefin-cli\\bluefin-cli.exe' init powershell) }",
		"elseif (Get-Command bluefin-cli -ErrorAction SilentlyContinue)",
		"$env:FOO = 'bar'",
	}, "\n") + "\n"

	if err := os.WriteFile(profile, []byte(legacy), 0644); err != nil {
		t.Fatalf("failed to write initial profile: %v", err)
	}

	changed, err := syncPowerShellProfile(profile, rcLine, true)
	if err != nil {
		t.Fatalf("syncPowerShellProfile(enable) returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected enable to report changed=true for legacy cleanup")
	}

	content, err := os.ReadFile(profile)
	if err != nil {
		t.Fatalf("failed to read profile after cleanup enable: %v", err)
	}

	got := string(content)
	if strings.Contains(got, "elseif (Get-Command bluefin-cli") {
		t.Fatalf("expected malformed legacy line to be removed")
	}
	if strings.Count(got, shellMaker) != 1 {
		t.Fatalf("expected exactly one managed shell marker, got %d", strings.Count(got, shellMaker))
	}
	if !strings.Contains(got, "$env:FOO = 'bar'") {
		t.Fatalf("expected unrelated profile content to be preserved")
	}

	changed, err = syncPowerShellProfile(profile, rcLine, false)
	if err != nil {
		t.Fatalf("syncPowerShellProfile(disable) returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected disable to report changed=true")
	}

	content, err = os.ReadFile(profile)
	if err != nil {
		t.Fatalf("failed to read profile after disable: %v", err)
	}
	if strings.Contains(string(content), "bluefin-cli") {
		t.Fatalf("expected all managed bluefin-cli profile lines to be removed")
	}
	if !strings.Contains(string(content), "$env:FOO = 'bar'") {
		t.Fatalf("expected unrelated profile content to remain after disable")
	}
}
