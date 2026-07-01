package shell

import (
	"testing"
)

func TestToolGetEnvVar(t *testing.T) {
	tool := Tool{Name: "Eza"}
	got := tool.GetEnvVar()
	if got != "BLUEFIN_SHELL_ENABLE_EZA" {
		t.Errorf("GetEnvVar() = %q, want %q", got, "BLUEFIN_SHELL_ENABLE_EZA")
	}
}

func TestToolGetEnvVar_MultiWord(t *testing.T) {
	tool := Tool{Name: "UutilsCoreutils"}
	got := tool.GetEnvVar()
	if got != "BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS" {
		t.Errorf("GetEnvVar() = %q, want %q", got, "BLUEFIN_SHELL_ENABLE_UUTILSCOREUTILS")
	}
}

func TestToolGetBrewPkg_HasBrewPkg(t *testing.T) {
	tool := Tool{Name: "Eza", Pkg: "eza-community.eza", BrewPkg: "eza"}
	got := tool.GetBrewPkg()
	if got != "eza" {
		t.Errorf("GetBrewPkg() = %q, want %q (brew pkg)", got, "eza")
	}
}

func TestToolGetBrewPkg_FallbackToPkg(t *testing.T) {
	tool := Tool{Name: "Bat", Pkg: "sharkdp.bat", BrewPkg: ""}
	got := tool.GetBrewPkg()
	if got != "sharkdp.bat" {
		t.Errorf("GetBrewPkg() = %q, want %q (fallback to Pkg)", got, "sharkdp.bat")
	}
}

func TestToolGetBrewPkg_BothEmpty(t *testing.T) {
	tool := Tool{Name: "Empty"}
	got := tool.GetBrewPkg()
	if got != "" {
		t.Errorf("expected empty when both fields empty, got %q", got)
	}
}

func TestToolSupportsShell_NoRestrictions(t *testing.T) {
	// Tool with no UnsupportedShells should support all shells
	tool := Tool{Name: "Eza", Default: true}
	if !tool.SupportsShell("bash") {
		t.Error("Eza should support bash")
	}
	if !tool.SupportsShell("zsh") {
		t.Error("Eza should support zsh")
	}
	if !tool.SupportsShell("powershell") {
		t.Error("Eza should support powershell")
	}
	if !tool.SupportsShell("fish") {
		t.Error("Eza should support fish")
	}
}

func TestToolSupportsShell_Unsupported(t *testing.T) {
	// Gsudo doesn't support bash/zsh/fish (it's a Windows-only tool)
	tool := Tool{
		Name:              "Gsudo",
		UnsupportedShells: map[string]bool{"bash": true, "zsh": true, "fish": true},
	}
	if tool.SupportsShell("bash") {
		t.Error("Gsudo should NOT support bash")
	}
	if tool.SupportsShell("zsh") {
		t.Error("Gsudo should NOT support zsh")
	}
	if !tool.SupportsShell("powershell") {
		t.Error("Gsudo should support powershell")
	}
}

func TestToolSupportsShell_PartialUnsupported(t *testing.T) {
	// Ugrep doesn't support powershell but supports others
	tool := Tool{
		Name:              "Ugrep",
		UnsupportedShells: map[string]bool{"powershell": true},
	}
	if !tool.SupportsShell("bash") {
		t.Error("Ugrep should support bash")
	}
	if !tool.SupportsShell("zsh") {
		t.Error("Ugrep should support zsh")
	}
	if tool.SupportsShell("powershell") {
		t.Error("Ugrep should NOT support powershell")
	}
}

func TestToolsForShell_Bash(t *testing.T) {
	tools := ToolsForShell("bash")
	// Most tools should support bash
	for _, tool := range tools {
		if !tool.SupportsShell("bash") {
			t.Errorf("Tool %s should support bash but ToolsForShell returned it", tool.Name)
		}
	}
}

func TestToolsForShell_Powershell(t *testing.T) {
	tools := ToolsForShell("powershell")
	// Powershell should exclude Ugrep and Uutils* tools
	for _, tool := range tools {
		if !tool.SupportsShell("powershell") {
			t.Errorf("Tool %s should support powershell but was returned", tool.Name)
		}
	}
}

func TestToolsForShell_AllToolsFiltered(t *testing.T) {
	// Every tool should appear in at least one shell's results
	bashTools := ToolsForShell("bash")
	pwshTools := ToolsForShell("powershell")
	allTools := append(bashTools, pwshTools...)
	seen := map[string]bool{}
	for _, tool := range allTools {
		seen[tool.Name] = true
	}
	// Gsudo only supports powershell, so it should appear
	if !seen["Gsudo"] {
		t.Error("Gsudo should appear in powershell ToolsForShell")
	}
}
