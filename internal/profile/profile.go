// Package profile captures a machine's bluefin-cli setup as a portable
// document and replays it elsewhere — the "make this machine feel like my
// machine" feature. Export/Import round-trip through the same public APIs
// the CLI and TUI use, so an imported profile reaches real state: rc lines,
// tool config, theme.
package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"

	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// Profile is the portable setup document.
type Profile struct {
	Version       int             `json:"version"`
	Flavor        string          `json:"flavor,omitempty"`
	Tools         map[string]bool `json:"tools,omitempty"`
	EnabledShells []string        `json:"enabled_shells"`
}

// reconciledShells are the shells import will reconcile exactly on this
// platform (Windows manages the PowerShell profile instead of rc files).
func reconciledShells() []string {
	if runtime.GOOS == "windows" {
		return []string{"powershell"}
	}
	return []string{"bash", "zsh", "fish"}
}

// Export captures the current setup.
func Export(currentShell string) (*Profile, error) {
	cfg, err := shell.LoadConfig(currentShell)
	if err != nil {
		cfg = shell.DefaultConfig(currentShell)
	}
	status := shell.CheckStatus()
	enabled := make([]string, 0, len(status))
	for sh, on := range status {
		if on {
			enabled = append(enabled, sh)
		}
	}
	sort.Strings(enabled)

	return &Profile{
		Version:       1,
		Flavor:        viper.GetString("ui.flavor"),
		Tools:         map[string]bool(*cfg),
		EnabledShells: enabled,
	}, nil
}

// Save writes the profile as indented JSON; path "-" writes to stdout.
func (p *Profile) Save(path string) error {
	data, err := p.Marshal()
	if err != nil {
		return err
	}
	if path == "-" || path == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Marshal returns the canonical portable profile document.
func (p *Profile) Marshal() ([]byte, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// Load reads a profile document.
func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse validates a profile document received from any storage or transport.
func Parse(data []byte) (*Profile, error) {
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	if p.Version != 1 {
		return nil, fmt.Errorf("unsupported profile version %d", p.Version)
	}
	return &p, nil
}

// Apply drives the machine to the profile's state: theme flavor, tool
// config, and shell integration (the unix shells are reconciled exactly —
// enabled if listed, disabled otherwise). It reports each step to progress
// output via fmt (captured by the TUI RunnerScreen when run from the menu).
func (p *Profile) Apply() error {
	if p.Flavor != "" {
		viper.Set("ui.flavor", p.Flavor)
		if err := config.Save(); err != nil {
			return fmt.Errorf("saving flavor: %w", err)
		}
		fmt.Printf("flavor: %s\n", p.Flavor)
	}

	if len(p.Tools) > 0 {
		cfg := shell.Config(p.Tools)
		if err := shell.SaveConfig(&cfg); err != nil {
			return fmt.Errorf("saving tool config: %w", err)
		}
		fmt.Printf("tools: %d configured\n", len(p.Tools))
	}

	want := map[string]bool{}
	for _, sh := range p.EnabledShells {
		want[sh] = true
	}
	current := shell.CheckStatus()
	for _, sh := range reconciledShells() {
		if current[sh] == want[sh] {
			continue
		}
		if err := shell.Toggle(sh, want[sh]); err != nil {
			return fmt.Errorf("toggling %s: %w", sh, err)
		}
		fmt.Printf("shell %s: %v\n", sh, want[sh])
	}
	return nil
}

// Diff reports the changes Apply would make to reach want from current —
// empty means no drift.
func Diff(current, want *Profile) []string {
	var out []string

	if want.Flavor != "" && want.Flavor != current.Flavor {
		out = append(out, fmt.Sprintf("flavor: %s -> %s", current.Flavor, want.Flavor))
	}

	names := make([]string, 0, len(want.Tools))
	for name := range want.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if current.Tools[name] != want.Tools[name] {
			state := "off -> on"
			if !want.Tools[name] {
				state = "on -> off"
			}
			out = append(out, fmt.Sprintf("tool %s: %s", name, state))
		}
	}

	cur := map[string]bool{}
	for _, sh := range current.EnabledShells {
		cur[sh] = true
	}
	wanted := map[string]bool{}
	for _, sh := range want.EnabledShells {
		wanted[sh] = true
		if !cur[sh] {
			out = append(out, fmt.Sprintf("shell %s: disabled -> enabled", sh))
		}
	}
	// Diff is a pure comparison of the two documents (Apply filters by
	// platform); anything enabled now but absent from the profile is drift.
	for _, sh := range current.EnabledShells {
		if !wanted[sh] {
			out = append(out, fmt.Sprintf("shell %s: enabled -> disabled", sh))
		}
	}
	return out
}
