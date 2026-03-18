package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToggle(t *testing.T) {
	// Toggle is now a no-op that prints to stdout, so we just check it doesn't error
	err := Toggle("bash", true)
	if err != nil {
		t.Errorf("Toggle() returned error: %v", err)
	}
}

func TestInit(t *testing.T) {
	tmpHome := t.TempDir()
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("Failed to set mock HOME: %v", err)
	}
	if err := os.Setenv("USERPROFILE", tmpHome); err != nil {
		t.Fatalf("Failed to set mock USERPROFILE: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("HOME")
		_ = os.Unsetenv("USERPROFILE")
	}()

	tests := []struct {
		name    string
		shell   string
		wantIn  []string
		wantErr bool
	}{
		{
			"Bash init",
			"bash",
			[]string{"export BLUEFIN_SHELL_ENABLE_EZA=", "shell.sh"},
			false,
		},
		{
			"Fish init",
			"fish",
			[]string{"set -gx BLUEFIN_SHELL_ENABLE_EZA", "shell.fish"},
			false,
		},
		{
			"Zsh init",
			"zsh",
			[]string{"export BLUEFIN_SHELL_ENABLE_EZA=", "shell.sh"},
			false,
		},
		{
			"PowerShell init",
			"powershell",
			[]string{"$env:BLUEFIN_SHELL_ENABLE_EZA", "bluefin_init"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Init(tt.shell, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.wantIn {
				// We check if the expected strings (like export commands or script content parts) are present
				if want == "shell.sh" || want == "shell.fish" {
					// Check for a known variable that should be in the script
					want = "BLUEFIN_SHELL_ENABLE_EZA"
				}

				if !strings.Contains(got, want) {
					t.Errorf("Init() output missing %q", want)
				}
			}
		})
	}
}

func TestCheckStatus(t *testing.T) {
	// Create temporary home directory
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	os.Setenv("USERPROFILE", tmpHome)
	defer os.Unsetenv("HOME")
	defer os.Unsetenv("USERPROFILE")

	// Manually create a bashrc with the marker
	bashrc := filepath.Join(tmpHome, ".bashrc")
	content := "# bluefin-cli shell-config\n"
	if err := os.WriteFile(bashrc, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create mock bashrc: %v", err)
	}

	status := CheckStatus()

	if !status["bash"] {
		t.Error("Expected bash shell experience to be enabled")
	}
	if status["zsh"] {
		t.Error("Expected zsh shell experience to be disabled")
	}
}

func TestCheckDependencies(t *testing.T) {
	deps := CheckDependencies()

	if deps == nil {
		t.Error("Expected non-nil dependencies map")
	}

	for _, tool := range toolsForCurrentPlatform() {
		if _, exists := deps[tool.Binary]; !exists {
			t.Errorf("Expected tool %s to be in dependencies map", tool.Binary)
		}
	}
}
