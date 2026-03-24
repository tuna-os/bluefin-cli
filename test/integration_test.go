package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func getBinaryPath() string {
	path := "../bluefin-cli"
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	return path
}

// Shell configuration for parameterized tests
type ShellConfig struct {
	Name         string
	ConfigFile   string
	ShellPattern string
	ShellScript  string
	InitShell    func() error
}

var shells = []ShellConfig{
	{
		Name:         "bash",
		ConfigFile:   ".bashrc",
		ShellPattern: "shell.sh",
		ShellScript:  "shell.sh",
		InitShell:    func() error { return touchFile(filepath.Join(os.Getenv("HOME"), ".bashrc")) },
	},
	{
		Name:         "zsh",
		ConfigFile:   ".zshrc",
		ShellPattern: "shell.sh",
		ShellScript:  "shell.sh",
		InitShell:    func() error { return touchFile(filepath.Join(os.Getenv("HOME"), ".zshrc")) },
	},
	{
		Name:         "fish",
		ConfigFile:   ".config/fish/config.fish",
		ShellPattern: "shell.fish",
		ShellScript:  "shell.fish",
		InitShell: func() error {
			dir := filepath.Join(os.Getenv("HOME"), ".config/fish")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return touchFile(filepath.Join(dir, "config.fish"))
		},
	},
	{
		Name:         "powershell",
		ConfigFile:   "Documents/PowerShell/Microsoft.PowerShell_profile.ps1",
		ShellPattern: "bluefin-cli init powershell",
		ShellScript:  "bluefin-cli init powershell",
		InitShell: func() error {
			dir := filepath.Join(os.Getenv("HOME"), "Documents", "PowerShell")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return touchFile(filepath.Join(dir, "Microsoft.PowerShell_profile.ps1"))
		},
	},
}

// Tool configuration for testing shell script content
type ToolConfig struct {
	Name    string
	Pattern string
	Shell   string // "bash", "zsh", or "fish"
}

var tools = []ToolConfig{
	{Name: "eza", Pattern: "alias ll='eza", Shell: "bash"},
	{Name: "bat", Pattern: "alias cat='bat", Shell: "bash"},
	{Name: "starship-bash", Pattern: "starship init ${BLING_SHELL}", Shell: "bash"},
	{Name: "starship-zsh", Pattern: "starship init ${BLING_SHELL}", Shell: "zsh"},
	{Name: "starship-fish", Pattern: "starship init fish", Shell: "fish"},
	{Name: "zoxide", Pattern: "zoxide init", Shell: "bash"},
	{Name: "atuin", Pattern: "atuin init", Shell: "bash"},
}

func TestMain(m *testing.M) {
	// Setup
	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "bluefin-test-home-*")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpHome)
	}()

	if err := os.Setenv("HOME", tmpHome); err != nil {
		fmt.Printf("Warning: failed to set mock HOME: %v\n", err)
	}

	// Initialize shell configs
	for _, shell := range shells {
		if err := shell.InitShell(); err != nil {
			panic(err)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := os.Setenv("HOME", originalHome); err != nil {
		fmt.Printf("Warning: failed to restore HOME: %v\n", err)
	}
	os.Exit(code)
}

func touchFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

var ansiRegex = regexp.MustCompile("[\u001b\u009b][\\[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")

func stripAnsi(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

func runCommand(t *testing.T, args ...string) (string, error) {
	cmd := exec.Command(getBinaryPath(), args...)
	output, err := cmd.CombinedOutput()
	return stripAnsi(string(output)), err
}

func fileContains(t *testing.T, filepath, pattern string) bool {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), pattern)
}

func TestBinaryExecutes(t *testing.T) {
	_, err := runCommand(t, "--version")
	if err != nil {
		t.Fatalf("Binary failed to execute: %v", err)
	}
}

func TestStatusCommand(t *testing.T) {
	_, err := runCommand(t, "status")
	if err != nil {
		t.Fatalf("Status command failed: %v", err)
	}
}

func TestShellEnableForAllShells(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell enabling tests on native Windows")
	}
	for _, shell := range shells {
		t.Run(shell.Name, func(t *testing.T) {
			// Enable shell
			_, err := runCommand(t, "shell", shell.Name, "on")
			if err != nil {
				t.Fatalf("Failed to enable shell for %s: %v", shell.Name, err)
			}

			// Verify config file contains the managed init line.
			// PowerShell uses a caching profile that doesn't include "bluefin-cli init"
			// literally; check for the shell-config marker instead.
			configPath := filepath.Join(os.Getenv("HOME"), shell.ConfigFile)
			expected := "bluefin-cli init"
			if shell.Name == "powershell" {
				expected = "# bluefin-cli shell-config"
			}
			if !fileContains(t, configPath, expected) {
				t.Errorf("Config file %s doesn't contain init command %s", configPath, expected)
			}
		})
	}
}

func TestShellScriptSourcing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell script sourcing tests on native Windows")
	}
	for _, shell := range shells {
		t.Run(shell.Name, func(t *testing.T) {
			// Run init command and check if it outputs the script content
			output, err := runCommand(t, "init", shell.Name)
			if err != nil {
				t.Fatalf("Failed to run init: %v", err)
			}

			// We check for some shell specific syntax or standard env vars
			if !strings.Contains(output, "BLUEFIN_SHELL_ENABLE_EZA") {
				t.Errorf("Init output doesn't seem to contain shell script logic")
			}
		})
	}
}

func TestShellSyntax(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell syntax tests on native Windows")
	}
	tests := []struct {
		shell      string
		configFile string
		validator  string
	}{
		{"bash", ".bashrc", "bash"},
		{"zsh", ".zshrc", "zsh"},
		{"fish", ".config/fish/config.fish", "fish"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			if _, err := exec.LookPath(tt.validator); err != nil {
				t.Skipf("Skipping %s syntax check: validator %q not found", tt.shell, tt.validator)
			}
			configPath := filepath.Join(os.Getenv("HOME"), tt.configFile)
			cmd := exec.Command(tt.validator, "-n", configPath)
			if err := cmd.Run(); err != nil {
				t.Errorf("%s config has syntax errors: %v", tt.shell, err)
			}
		})
	}
}

func TestShellToolConfigurations(t *testing.T) {
	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			// Run init to get the script content
			output, err := runCommand(t, "init", tool.Shell)
			if err != nil {
				t.Fatalf("Failed to run init: %v", err)
			}

			if !strings.Contains(output, tool.Pattern) {
				t.Errorf("Shell initialization script doesn't contain configuration for %s (pattern: %s)", tool.Name, tool.Pattern)
			}
		})
	}
}

func TestMOTDSystem(t *testing.T) {
	t.Run("MOTDInBashrc", func(t *testing.T) {
		output, _ := runCommand(t, "init", "bash")
		if !strings.Contains(output, "bluefin-cli motd show") {
			t.Error("MOTD hook missing from init output")
		}
	})

	// MOTD resources check removed as it depends on external setup not controlled by CLI logic under test

	t.Run("MOTDShowCommand", func(t *testing.T) {
		output, _ := runCommand(t, "motd", "show")
		if !strings.Contains(output, "Bluefin") {
			t.Error("MOTD show command didn't display expected content")
		}
	})
}

func TestStatusReflectsChanges(t *testing.T) {
	targetShell := "bash"
	if runtime.GOOS == "windows" {
		targetShell = "powershell"
	}

	// Enable shell
	_, err := runCommand(t, "shell", targetShell, "on")
	if err != nil {
		t.Fatalf("Failed to enable shell %s: %v", targetShell, err)
	}

	output, err := runCommand(t, "status")
	if err != nil {
		t.Fatalf("Status command failed: %v", err)
	}

	t.Logf("Status Output:\n%s", output)

	found := false
	if targetShell == "powershell" {
		if strings.Contains(output, "powershell: enabled") || strings.Contains(output, "pwsh: enabled") {
			found = true
		}
	} else {
		if strings.Contains(output, fmt.Sprintf("%s: enabled", targetShell)) {
			found = true
		}
	}

	if !found {
		t.Errorf("Status doesn't show %s as enabled", targetShell)
	}

	if !strings.Contains(output, "Message of the Day") {
		t.Error("Status doesn't show MOTD section")
	}
}

func TestShellDisable(t *testing.T) {
	_, err := runCommand(t, "shell", "bash", "off")
	if err != nil {
		t.Fatalf("Failed to disable shell: %v", err)
	}

	bashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
	if fileContains(t, bashrc, "# bluefin-cli shell-config") {
		t.Error("Shell marker still present in bashrc after disable")
	}
}

func TestInstallList(t *testing.T) {
	output, err := runCommand(t, "install", "list")
	if err != nil {
		t.Fatalf("Install list command failed: %v", err)
	}

	if !strings.Contains(output, "Available Bundles") {
		t.Error("Install list doesn't show expected content")
	}
}
