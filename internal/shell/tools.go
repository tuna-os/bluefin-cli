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
	Default           bool            // Whether enabled by default
	ShellDefaults     map[string]bool // Per-shell default overrides
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

func (t Tool) SupportsShell(shell string) bool {
	if len(t.UnsupportedShells) == 0 {
		return true
	}

	normalized := strings.ToLower(strings.TrimSpace(shell))
	if normalized == "pwsh" {
		normalized = "powershell"
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
	{Name: "Eza", Description: "Modern, maintained replacement for ls", Binary: "eza", Pkg: "eza-community.eza", BrewPkg: "eza", Default: true},
	{Name: "Gsudo", Description: "sudo for Windows (run commands elevated)", Binary: "gsudo", Pkg: "gerardog.gsudo", Default: true, UnsupportedShells: map[string]bool{"bash": true, "zsh": true, "fish": true}},
	{Name: "Fzf", Description: "Command-line fuzzy finder", Binary: "fzf", Pkg: "junegunn.fzf", BrewPkg: "fzf", Default: true},
	{Name: "Ugrep", Description: "Ultra fast grep with interactive mode", Binary: "ug", Pkg: "ugrep", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "Bat", Description: "A cat clone with wings", Binary: "bat", Pkg: "sharkdp.bat", BrewPkg: "bat", Default: true},
	{Name: "Atuin", Description: "Magical shell history", Binary: "atuin", Pkg: "atuin", Default: false, ShellDefaults: map[string]bool{"zsh": true, "fish": true}},
	{Name: "Starship", Description: "The minimal, blazing-fast, and infinitely customizable prompt", Binary: "starship", Pkg: "Starship.Starship", BrewPkg: "starship", Default: true},
	{Name: "Zoxide", Description: "A smarter cd command", Binary: "zoxide", Pkg: "ajeetdsouza.zoxide", BrewPkg: "zoxide", Default: true},
	{Name: "UutilsCoreutils", Description: "Rust rewrite of GNU coreutils", Binary: "hashsum", Pkg: "uutils-coreutils", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "UutilsFindutils", Description: "Rust rewrite of GNU findutils", Binary: "ufind", Pkg: "uutils-findutils", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "UutilsDiffutils", Description: "Rust rewrite of GNU diffutils", Binary: "udiffutils", Pkg: "uutils-diffutils", Default: true, UnsupportedShells: map[string]bool{"powershell": true}},
	{Name: "Carapace", Description: "Multi-shell multi-command argument completer", Binary: "carapace", Pkg: "rsteube.carapace", BrewPkg: "carapace", Default: false},
	{Name: "Glow", Description: "Terminal markdown renderer for MOTD", Binary: "glow", Pkg: "charmbracelet.glow", BrewPkg: "glow", Default: true},
}
