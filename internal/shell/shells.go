package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Shell describes one shell bluefin-cli can manage. Adding a shell used to
// mean editing four parallel switch statements (Toggle, CheckStatus,
// GetInstalledShells, Init) and hoping none was missed; the registry below is
// the single place that knows about a shell, and those functions read from it.
//
// PowerShell is deliberately absent: its config lives in a
// platform-dependent profile path discovered at runtime, and it is handled by
// powershell.go rather than by a static entry here.
type Shell struct {
	// Name is the canonical name, as accepted by `bluefin-cli init <shell>`.
	Name string

	// Aliases are other names that resolve to this shell. Nushell's binary is
	// `nu` while its ecosystem writes "nushell"; both must work.
	Aliases []string

	// Binaries are the executables to look for when detecting whether this
	// shell is installed. Usually just Name.
	Binaries []string

	// RCPath is the config file the init line is written into, relative to
	// $HOME.
	RCPath string

	// RCLine renders the line appended to RCPath. It takes $HOME so shells
	// that source a generated file can spell out an absolute path.
	RCLine func(home string) string

	// Flavor selects the embedded script and the variable syntax used to
	// export tool toggles: "posix", "fish" or "nu".
	Flavor string

	// GeneratedInit, when non-empty, is a path relative to $HOME that Toggle
	// writes the rendered init script to. Nushell cannot eval a string at
	// runtime the way `eval "$(...)"` does in a POSIX shell, so its startup
	// file has to source a real file on disk.
	GeneratedInit string
}

// registry is the set of managed non-PowerShell shells.
var registry = []Shell{
	{
		Name:     "bash",
		Binaries: []string{"bash"},
		RCPath:   ".bashrc",
		RCLine:   func(string) string { return `eval "$(bluefin-cli init bash)" ` + shellMaker },
		Flavor:   "posix",
	},
	{
		Name:     "zsh",
		Binaries: []string{"zsh"},
		RCPath:   ".zshrc",
		RCLine:   func(string) string { return `eval "$(bluefin-cli init zsh)" ` + shellMaker },
		Flavor:   "posix",
	},
	{
		Name:     "fish",
		Binaries: []string{"fish"},
		RCPath:   ".config/fish/config.fish",
		RCLine:   func(string) string { return `bluefin-cli init fish | source ` + shellMaker },
		Flavor:   "fish",
	},
	{
		// ash is busybox's shell and the default on Alpine and postmarketOS
		// (tuna-os/bluefin-cli#260). It reads the file named by $ENV at
		// interactive startup and has no conventional rc file of its own, so
		// Toggle writes ~/.ashrc and points $ENV at it from ~/.profile.
		Name:     "ash",
		Aliases:  []string{"dash", "sh"},
		Binaries: []string{"ash", "busybox"},
		RCPath:   ".ashrc",
		RCLine:   func(string) string { return `eval "$(bluefin-cli init ash)" ` + shellMaker },
		Flavor:   "posix",
	},
	{
		// Nushell's own syntax, its own config file, and no runtime eval:
		// config.nu sources a file that `shell enable` renders ahead of time.
		Name:     "nu",
		Aliases:  []string{"nushell"},
		Binaries: []string{"nu"},
		RCPath:   ".config/nushell/config.nu",
		RCLine: func(home string) string {
			return "source " + filepath.Join(home, generatedNuInit) + " " + shellMaker
		},
		Flavor:        "nu",
		GeneratedInit: generatedNuInit,
	},
}

// generatedNuInit is where the rendered nushell init script is written.
const generatedNuInit = ".config/nushell/bluefin-cli.nu"

// ashENVLine is appended to ~/.profile so an interactive ash actually reads
// ~/.ashrc. Without it the rc file exists and is never sourced.
const ashENVLine = `export ENV="$HOME/.ashrc" ` + shellMaker

// NormalizeShell maps a user-supplied name onto its canonical form. Unknown
// names are returned lowercased and untouched so callers can report them.
func NormalizeShell(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "pwsh" {
		return "powershell"
	}
	for _, s := range registry {
		if normalized == s.Name {
			return s.Name
		}
		for _, alias := range s.Aliases {
			if normalized == alias {
				return s.Name
			}
		}
	}
	return normalized
}

// LookupShell finds a managed shell by name or alias.
func LookupShell(name string) (Shell, bool) {
	canonical := NormalizeShell(name)
	for _, s := range registry {
		if s.Name == canonical {
			return s, true
		}
	}
	return Shell{}, false
}

// ManagedShells returns the canonical names of every managed non-PowerShell
// shell, in registry order.
func ManagedShells() []string {
	names := make([]string, 0, len(registry))
	for _, s := range registry {
		names = append(names, s.Name)
	}
	return names
}

// ConfigPath is the absolute path of the shell's startup file.
func (s Shell) ConfigPath(home string) string {
	return filepath.Join(home, filepath.FromSlash(s.RCPath))
}

// GeneratedInitPath is the absolute path of the rendered init script, or ""
// for shells that evaluate it inline.
func (s Shell) GeneratedInitPath(home string) string {
	if s.GeneratedInit == "" {
		return ""
	}
	return filepath.Join(home, filepath.FromSlash(s.GeneratedInit))
}

// IsInstalled reports whether the shell is on PATH.
func (s Shell) IsInstalled() bool {
	for _, binary := range s.Binaries {
		if _, err := exec.LookPath(binary); err == nil {
			return true
		}
	}
	return false
}

// exportVar renders one `NAME=value` assignment in the shell's own syntax.
func (s Shell) exportVar(name string, value int) string {
	switch s.Flavor {
	case "fish":
		return fmt.Sprintf("set -gx %s %d\n", name, value)
	case "nu":
		return fmt.Sprintf("$env.%s = \"%d\"\n", name, value)
	default:
		return fmt.Sprintf("export %s=%d\n", name, value)
	}
}

// exportString renders one string assignment in the shell's own syntax.
func (s Shell) exportString(name, value string) string {
	switch s.Flavor {
	case "fish":
		return fmt.Sprintf("set -gx %s %s\n", name, value)
	case "nu":
		return fmt.Sprintf("$env.%s = \"%s\"\n", name, value)
	default:
		return fmt.Sprintf("export %s=\"%s\"\n", name, value)
	}
}

// script returns the embedded runtime script for this shell's flavor.
func (s Shell) script() string {
	switch s.Flavor {
	case "fish":
		return shellFishScript
	case "nu":
		return shellNuScript
	default:
		return shellShScript
	}
}

// ensureRCDir creates the parent directory of the shell's config file. fish
// and nushell both keep theirs under ~/.config, which may not exist yet.
func (s Shell) ensureRCDir(home string) error {
	return os.MkdirAll(filepath.Dir(s.ConfigPath(home)), 0755)
}
