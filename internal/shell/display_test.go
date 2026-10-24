package shell

import (
	"strings"
	"testing"
)

func TestFormatToggleResult(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		enabled  bool
		changed  bool
		contains string
	}{
		{
			name:     "enabled with change",
			shell:    "bash",
			enabled:  true,
			changed:  true,
			contains: "Enabled shell experience for bash",
		},
		{
			name:     "enabled without change",
			shell:    "bash",
			enabled:  true,
			changed:  false,
			contains: "bash is already enabled for bash",
		},
		{
			name:     "disabled with change",
			shell:    "zsh",
			enabled:  false,
			changed:  true,
			contains: "Disabled shell experience for zsh",
		},
		{
			name:     "disabled without change",
			shell:    "zsh",
			enabled:  false,
			changed:  false,
			contains: "zsh is already disabled for zsh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatToggleResult(tt.shell, tt.enabled, tt.changed)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("FormatToggleResult(%q, %v, %v) = %q; want to contain %q", tt.shell, tt.enabled, tt.changed, got, tt.contains)
			}
		})
	}
}

func TestStyles(t *testing.T) {
	if SuccessStyle().GetForeground() == nil {
		t.Error("expected non-nil foreground for SuccessStyle")
	}
	if ErrorStyle().GetForeground() == nil {
		t.Error("expected non-nil foreground for ErrorStyle")
	}
	if InfoStyle().GetForeground() == nil {
		t.Error("expected non-nil foreground for InfoStyle")
	}
}
