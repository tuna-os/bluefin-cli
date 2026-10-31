package render

import (
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/shell"
)

func TestToggleStatusRendersAllStates(t *testing.T) {
	tests := []struct {
		name     string
		res      shell.ToggleResult
		contains string
	}{
		{
			name:     "enabled changed",
			res:      shell.ToggleResult{Shell: "bash", Enabled: true, Changed: true},
			contains: "✓ Enabled shell experience for bash",
		},
		{
			name:     "disabled changed",
			res:      shell.ToggleResult{Shell: "zsh", Enabled: false, Changed: true},
			contains: "✓ Disabled shell experience for zsh",
		},
		{
			name:     "already enabled",
			res:      shell.ToggleResult{Shell: "fish", Enabled: true, Changed: false},
			contains: "fish is already enabled for fish",
		},
		{
			name:     "already disabled",
			res:      shell.ToggleResult{Shell: "nu", Enabled: false, Changed: false},
			contains: "nu is already disabled for nu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := ToggleStatus(tt.res)
			if !strings.Contains(out, tt.contains) {
				t.Fatalf("expected output to contain %q, got %q", tt.contains, out)
			}
		})
	}
}

func TestUninstallSuccess(t *testing.T) {
	out := UninstallSuccess()
	if !strings.Contains(out, "✓ Bluefin shell setup uninstalled") {
		t.Fatalf("unexpected uninstall success output: %s", out)
	}
}
