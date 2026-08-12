package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToggle(t *testing.T) {
	// Toggle rewrites the shell rc file under $HOME. Use a temp home so the
	// test never touches the real user's dotfiles (and fails cleanly when
	// that home is not writable).
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Enable creates ~/.bashrc with the init line.
	if err := Toggle("bash", true); err != nil {
		t.Fatalf("Toggle(bash, true) error: %v", err)
	}
	bashrc := filepath.Join(tmpHome, ".bashrc")
	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatalf("read %s: %v", bashrc, err)
	}
	if !strings.Contains(string(data), "bluefin-cli init bash") {
		t.Errorf("enabled .bashrc missing init line: %q", data)
	}

	// Enabling again is idempotent and must not duplicate the line.
	if err := Toggle("bash", true); err != nil {
		t.Errorf("Toggle(bash, true) second call error: %v", err)
	}
	data, _ = os.ReadFile(bashrc)
	if strings.Count(string(data), "bluefin-cli init bash") != 1 {
		t.Errorf("enabling twice duplicated the init line: %q", data)
	}

	// Disable removes the marker.
	if err := Toggle("bash", false); err != nil {
		t.Fatalf("Toggle(bash, false) error: %v", err)
	}
	data, _ = os.ReadFile(bashrc)
	if strings.Contains(string(data), "bluefin-cli init bash") {
		t.Errorf("disabled .bashrc still contains init line: %q", data)
	}

	// Disabling when already disabled is a no-op success.
	if err := Toggle("bash", false); err != nil {
		t.Errorf("Toggle(bash, false) second call error: %v", err)
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
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

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
