package starship

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hanthor/bluefin-cli/internal/tui"
)

var (
	// For testing
	execCommand = exec.Command
	runCommand  = func(cmd *exec.Cmd) error {
		return cmd.Run()
	}
	lookPath = exec.LookPath
)

// Install downloads and installs Starship
func Install() error {
	// Check if already installed
	if _, err := lookPath("starship"); err == nil {
		fmt.Println(tui.SuccessStyle.Render("✓ Starship is already installed"))
		return nil
	}

	fmt.Println(tui.InfoStyle.Render("⬇️  Installing Starship..."))

	// Use Homebrew if available
	if _, err := lookPath("brew"); err == nil {
		cmd := execCommand("brew", "install", "starship")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("brew install failed: %w", err)
		}

		fmt.Println(tui.SuccessStyle.Render("✓ Starship installed successfully!"))
		return nil
	}

	// Fallback to official installer
	cmd := execCommand("sh", "-c", "curl -sS https://starship.rs/install.sh | sh -s -- -y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println(tui.SuccessStyle.Render("✓ Starship installed successfully!"))
	return nil
}

// ApplyTheme applies a Starship preset theme
func ApplyTheme(themeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")
	starshipConfig := filepath.Join(configDir, "starship.toml")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Download and apply the preset
	cmd := execCommand("starship", "preset", themeName, "-o", starshipConfig)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	return nil
}
