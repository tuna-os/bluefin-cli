package shell

import (
	"strings"
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

// TestGetApkPkg_UutilsResolveToPackageNames is the regression guard for
// tuna-os/bluefin-cli#261. GetApkPkg used to fall back to Tool.Binary, so the
// uutils trio was handed to coldbrew/apk as "ucat", "ufind" and "udiffutils"
// -- binary names that no Alpine repository has ever carried. Every Alpine
// user saw three "no such package" errors on every shell setup.
func TestGetApkPkg_UutilsResolveToPackageNames(t *testing.T) {
	want := map[string]string{
		"UutilsCoreutils": "uutils-coreutils",
		"UutilsFindutils": "uutils-findutils",
		"UutilsDiffutils": "uutils-diffutils",
	}
	for _, tool := range Tools {
		expected, ok := want[tool.Name]
		if !ok {
			continue
		}
		if got := tool.GetApkPkg(); got != expected {
			t.Errorf("%s.GetApkPkg() = %q, want %q", tool.Name, got, expected)
		}
		delete(want, tool.Name)
	}
	for name := range want {
		t.Errorf("tool %s is no longer in Tools — update this test", name)
	}
}

// TestGetApkPkg_NeverFallsBackToAWingetID Tool.Pkg is a WinGet identifier for
// most tools ("sharkdp.bat"). apk has never heard of one, so the fallback
// chain must not reach it.
func TestGetApkPkg_NeverFallsBackToAWingetID(t *testing.T) {
	for _, tool := range Tools {
		if got := tool.GetApkPkg(); strings.Contains(got, ".") {
			t.Errorf("%s.GetApkPkg() = %q, which looks like a WinGet ID rather than an apk package", tool.Name, got)
		}
	}
}

// TestGetApkPkg_FallbackOrder pins the documented precedence:
// ApkPkg, then BrewPkg, then Binary.
func TestGetApkPkg_FallbackOrder(t *testing.T) {
	cases := []struct {
		name string
		tool Tool
		want string
	}{
		{"explicit apk name wins", Tool{Binary: "b", Pkg: "vendor.p", BrewPkg: "brew", ApkPkg: "apk"}, "apk"},
		{"brew name is the fallback", Tool{Binary: "b", Pkg: "vendor.p", BrewPkg: "brew"}, "brew"},
		{"binary is the last resort", Tool{Binary: "b", Pkg: "vendor.p"}, "b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.tool.GetApkPkg(); got != tc.want {
				t.Errorf("GetApkPkg() = %q, want %q", got, tc.want)
			}
		})
	}
}
