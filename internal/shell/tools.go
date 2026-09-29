package shell

import (
	"fmt"
	"strings"
)

// Tool represents a CLI tool that can be managed by bluefin-cli
type Tool struct {
	Name              string          // Display name
	Description       string          // Short description
	Binary            string          // Binary name to check for
	Pkg               string          // Default/WinGet package name
	BrewPkg           string          // Homebrew package name (fallback to Pkg if empty)
	ApkPkg            string          // Alpine/apk package name (Alpine, postmarketOS); see GetApkPkg for the fallback order
	Default           bool            // Whether enabled by default
	ShellDefaults     map[string]bool // Per-shell default overrides
	SupportedShells   map[string]bool // Optional allowlist for shell-specific tools
	UnsupportedShells map[string]bool // Shells where this tool should not be managed
}

// GetEnvVar returns the environment variable name for this tool
func (t Tool) GetEnvVar() string {
	return fmt.Sprintf("BLUEFIN_SHELL_ENABLE_%s", strings.ToUpper(t.Name))
}

func (t Tool) GetBrewPkg() string {
	if t.BrewPkg != "" {
		return t.BrewPkg
	}
	return t.Pkg
}

// GetApkPkg resolves the name to hand `coldbrew install` / `apk add`.
//
// The fallback used to be t.Binary, which is wrong whenever a package ships a
// binary under a different name: the uutils trio installs as
// uutils-coreutils but provides ucat, so Alpine users got
// "ucat (no such package)" for every one of them
// (tuna-os/bluefin-cli#261). BrewPkg is the better fallback -- it is already a
// plain package name in the same naming tradition Alpine follows, whereas
// t.Pkg is a WinGet identifier ("sharkdp.bat") that apk has never heard of.
// Binary remains the last resort for tools that declare nothing else.
func (t Tool) GetApkPkg() string {
	if t.ApkPkg != "" {
		return t.ApkPkg
	}
	if t.BrewPkg != "" {
		return t.BrewPkg
	}
	return t.Binary
}

func (t Tool) SupportsShell(shell string) bool {
	normalized := NormalizeShell(shell)
	if len(t.SupportedShells) > 0 {
		return t.SupportedShells[normalized]
	}

	return !t.UnsupportedShells[normalized]
}

func ToolsForShell(shell string) []Tool {
	filtered := make([]Tool, 0, len(Tools))
	for _, tool := range Tools {
		if tool.SupportsShell(shell) {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

// Tools is the list of managed tools
var Tools = []Tool{
	{Name: "Eza", Description: "Modern, maintained replacement for ls", Binary: "eza", Pkg: "eza-community.eza", BrewPkg: "eza", ApkPkg: "eza", Default: true},
	{Name: "Gsudo", Description: "sudo for Windows (run commands elevated)", Binary: "gsudo", Pkg: "gerardog.gsudo", Default: true, SupportedShells: map[string]bool{"powershell": true}},
	{Name: "Fzf", Description: "Command-line fuzzy finder", Binary: "fzf", Pkg: "junegunn.fzf", BrewPkg: "fzf", ApkPkg: "fzf", Default: true},
	{Name: "Ugrep", Description: "Ultra fast grep with interactive mode", Binary: "ug", Pkg: "ugrep", ApkPkg: "ugrep", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "Bat", Description: "A cat clone with wings", Binary: "bat", Pkg: "sharkdp.bat", BrewPkg: "bat", ApkPkg: "bat", Default: true},
	{Name: "Atuin", Description: "Magical shell history", Binary: "atuin", Pkg: "atuin", ApkPkg: "atuin", Default: false, ShellDefaults: map[string]bool{"zsh": true, "fish": true}},
	{Name: "Starship", Description: "A minimal, fast, and customizable prompt", Binary: "starship", Pkg: "Starship.Starship", BrewPkg: "starship", ApkPkg: "starship", Default: true},
	{Name: "Zoxide", Description: "A smarter cd command", Binary: "zoxide", Pkg: "ajeetdsouza.zoxide", BrewPkg: "zoxide", ApkPkg: "zoxide", Default: true},
	{Name: "UutilsCoreutils", Description: "Rust rewrite of GNU coreutils", Binary: "ucat", Pkg: "uutils-coreutils", BrewPkg: "uutils-coreutils", ApkPkg: "uutils-coreutils", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "UutilsFindutils", Description: "Rust rewrite of GNU findutils", Binary: "ufind", Pkg: "uutils-findutils", BrewPkg: "uutils-findutils", ApkPkg: "uutils-findutils", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "UutilsDiffutils", Description: "Rust rewrite of GNU diffutils", Binary: "udiffutils", Pkg: "uutils-diffutils", BrewPkg: "uutils-diffutils", ApkPkg: "uutils-diffutils", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "Carapace", Description: "Argument completion for many shells and commands", Binary: "carapace", Pkg: "rsteube.carapace", BrewPkg: "carapace", Default: false},
	{Name: "Glow", Description: "Terminal markdown renderer for MOTD", Binary: "glow", Pkg: "charmbracelet.glow", BrewPkg: "glow", Default: true},
}
