package shell

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestGetInstalledShells(t *testing.T) {
	installed := GetInstalledShells()
	// Should return a slice without panicking
	if installed == nil {
		// Even if empty, could be nil or empty slice, but we can verify it doesn't fail
		t.Log("No installed shells detected in current environment")
	}
}
