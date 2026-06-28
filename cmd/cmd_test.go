package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ── Command tree verification ────────────────────────────────────────────────

func TestRootCommand_HasSubcommands(t *testing.T) {
	expected := []string{
		"cleanup",
		"countme",
		"docs",
		"fonts",
		"help",
		"init",
		"install",
		"install-wallpapers",
		"install-wallpapers-cleanup",
		"menu",
		"motd",
		"shell",
		"starship",
		"status",
		"sunset",
		"uninstall",
	}
	for _, name := range expected {
		cmd, _, err := rootCmd.Find([]string{name})
		if err != nil || cmd == rootCmd || cmd == nil {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestRootCommand_HasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("rootCmd should have a version set")
	}
}

func TestMenuCommand_IsDefault(t *testing.T) {
	// When no subcommand is provided, rootCmd.RunE should exist
	if rootCmd.RunE == nil {
		t.Error("rootCmd should have a RunE function (default menu)")
	}
}

func TestRootCommand_HasPersistentPreRun(t *testing.T) {
	if rootCmd.PersistentPreRunE == nil {
		t.Error("rootCmd should have PersistentPreRunE (countme ping)")
	}
}

// ── Subcommand structure ─────────────────────────────────────────────────────

func TestMotdCommand_HasSubcommands(t *testing.T) {
	expected := []string{"config", "show", "toggle"}
	cmd, _, err := rootCmd.Find([]string{"motd"})
	if err != nil || cmd == rootCmd {
		t.Fatal("motd command not found")
	}
	for _, name := range expected {
		sub, _, err := cmd.Find([]string{name})
		if err != nil || sub == cmd || sub == nil {
			t.Errorf("missing motd subcommand %q", name)
		}
	}
}

func TestShellCommand_HasSubcommands(t *testing.T) {
	expected := []string{"config"}
	cmd, _, err := rootCmd.Find([]string{"shell"})
	if err != nil || cmd == rootCmd {
		t.Fatal("shell command not found")
	}
	for _, name := range expected {
		sub, _, err := cmd.Find([]string{name})
		if err != nil || sub == cmd || sub == nil {
			t.Errorf("missing shell subcommand %q", name)
		}
	}
}

func TestStarshipCommand_HasSubcommands(t *testing.T) {
	expected := []string{"install", "theme"}
	cmd, _, err := rootCmd.Find([]string{"starship"})
	if err != nil || cmd == rootCmd {
		t.Fatal("starship command not found")
	}
	for _, name := range expected {
		sub, _, err := cmd.Find([]string{name})
		if err != nil || sub == cmd || sub == nil {
			t.Errorf("missing starship subcommand %q", name)
		}
	}
}

func TestSunsetCommand_HasSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"sunset"})
	if err != nil || cmd == rootCmd {
		t.Fatal("sunset command not found")
	}
	// Should have setup subcommand
	sub, _, err := cmd.Find([]string{"setup"})
	if err != nil || sub == cmd || sub == nil {
		t.Error("missing sunset subcommand 'setup'")
	}
}

// ── Help output ──────────────────────────────────────────────────────────────

func TestHelpOutput(t *testing.T) {
	// Use Help() directly instead of Execute() to avoid triggering
	// PersistentPreRunE (which spawns countme goroutines).
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	if err := rootCmd.Help(); err != nil {
		t.Fatalf("rootCmd.Help() failed: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "bluefin-cli") {
		t.Error("help output should contain 'bluefin-cli'")
	}
	if !strings.Contains(output, "Usage:") {
		t.Error("help output should contain 'Usage:'")
	}
}

func TestSubcommandHelpOutput(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		expect string
	}{
		{"install", "install"},
		{"motd", "motd"},
		{"shell", "shell"},
		{"starship", "starship"},
		{"status", "status"},
		{"countme", "countme"},
		{"fonts", "fonts"},
		{"docs", "docs"},
		{"cleanup", "cleanup"},
		{"uninstall", "uninstall"},
		{"menu", "menu"},
		{"init", "init"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{tt.args})
			if err != nil || cmd == rootCmd || cmd == nil {
				t.Fatalf("command %s not found", tt.name)
			}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			if err := cmd.Help(); err != nil {
				t.Fatalf("%s.Help() failed: %v", tt.name, err)
			}
			output := buf.String()
			if !strings.Contains(output, tt.expect) {
				t.Errorf("help output for %s should contain %q, got:\n%s", tt.name, tt.expect, output)
			}
			if !strings.Contains(output, "Usage:") {
				t.Errorf("help output for %s should contain 'Usage:'", tt.name)
			}
		})
	}
}

// ── Flag verification ────────────────────────────────────────────────────────

func TestInstallCommand_HasRequiredFlags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"install"})
	if err != nil || cmd == rootCmd {
		t.Fatal("install command not found")
	}
	expectedFlags := []string{"non-interactive", "yes"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("install command missing flag --%s", flag)
		}
	}
}

func TestCountmeCommand_HasRequiredFlags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"countme"})
	if err != nil || cmd == rootCmd {
		t.Fatal("countme command not found")
	}
	expectedFlags := []string{"disable", "enable", "status"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("countme command missing flag --%s", flag)
		}
	}
}

func TestShellCommand_HasRequiredFlags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"shell"})
	if err != nil || cmd == rootCmd {
		t.Fatal("shell command not found")
	}
	shellCmd, _, _ := cmd.Find([]string{"config"})
	if shellCmd != nil && shellCmd != cmd {
		if shellCmd.Flags().Lookup("shell") == nil {
			t.Error("shell config command missing flag --shell")
		}
	}
}

func TestVersionTemplate(t *testing.T) {
	// Verify version template is set on the root command
	if rootCmd.VersionTemplate() == "" {
		t.Error("rootCmd should have a version template")
	}
	// The template should contain the version string somewhere
	tpl := rootCmd.VersionTemplate()
	if !strings.Contains(tpl, "version") {
		t.Errorf("version template should contain 'version', got: %s", tpl)
	}
}

// ── Nested command tree helpers ──────────────────────────────────────────────

func countCommands(cmd *cobra.Command) int {
	count := 1 // count this command
	for _, sub := range cmd.Commands() {
		count += countCommands(sub)
	}
	return count
}

func TestCommandTreeSize(t *testing.T) {
	total := countCommands(rootCmd)
	// Expect at least 15 top-level + subcommands (excluding help/completion)
	if total < 15 {
		t.Errorf("command tree seems too small: %d commands total", total)
	}
}
