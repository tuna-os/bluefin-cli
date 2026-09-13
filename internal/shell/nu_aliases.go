package shell

import (
	"fmt"
	"strings"
)

// Nushell's `alias` is a parse-time keyword, so an alias written inside an
// `if` block belongs to that block's scope and is gone by the time the block
// ends. The conditional alias blocks shell.nu used to carry therefore defined
// nothing at all: a nushell user got none of ll, ls, grep or cat, silently.
//
// The condition has to be resolved before the script is parsed, which means
// here: Init already knows which tools are enabled and present, so it emits
// exactly the aliases that apply, at the top level where nushell keeps them.
// The POSIX and fish scripts keep their runtime conditionals, which work
// correctly in those shells.

// nuAlias is one alias and the tool whose binary has to exist for it.
type nuAlias struct {
	toolName string // Tool.Name, matched against the shell config
	name     string // the alias
	body     string // what it expands to
}

// nuAliases mirrors the alias blocks in the POSIX and fish scripts.
var nuAliases = []nuAlias{
	{"Eza", "ll", "eza -l --icons=auto --group-directories-first"},
	{"Eza", "ls", "eza"},
	{"Eza", "l1", "eza -1"},
	{"Ugrep", "grep", "ug"},
	{"Ugrep", "egrep", "ug -E"},
	{"Ugrep", "fgrep", "ug -F"},
	{"Bat", "cat", "bat --style=plain --pager=never"},
}

// renderNuAliases returns the alias lines for every tool that is both enabled
// in the config and actually installed. A tool that is enabled but missing
// gets no alias: an alias pointing at an absent binary shadows the real
// command and fails only when the user runs it.
func renderNuAliases(config *Config) string {
	var sb strings.Builder

	byName := make(map[string]Tool, len(Tools))
	for _, tool := range Tools {
		byName[tool.Name] = tool
	}

	for _, alias := range nuAliases {
		tool, known := byName[alias.toolName]
		if !known || !config.IsEnabled(alias.toolName) || !isBinaryAvailable(tool) {
			continue
		}
		fmt.Fprintf(&sb, "alias %s = %s\n", alias.name, alias.body)
	}

	if sb.Len() > 0 {
		return "\n# Aliases, resolved at generation time -- see nu_aliases.go.\n" + sb.String()
	}
	return ""
}
