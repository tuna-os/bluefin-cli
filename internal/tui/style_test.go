package tui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// ── Style presence tests ─────────────────────────────────────────────────────

func TestStylesExist(t *testing.T) {
	tests := []struct {
		name  string
		style interface{}
	}{
		{"TitleStyle", TitleStyle},
		{"SubtitleStyle", SubtitleStyle},
		{"SuccessStyle", SuccessStyle},
		{"ErrorStyle", ErrorStyle},
		{"WarningStyle", WarningStyle},
		{"InfoStyle", InfoStyle},
		{"PopupStyle", PopupStyle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.style == nil {
				t.Errorf("%s is nil", tt.name)
			}
		})
	}
}

func TestStyleRendering(t *testing.T) {
	// Verify styles render without panic
	tests := []struct {
		name  string
		value string
		fn    func(...string) string
	}{
		{"TitleStyle", "Test Title", TitleStyle.Render},
		{"SuccessStyle", "✓ Success", SuccessStyle.Render},
		{"ErrorStyle", "✗ Error", ErrorStyle.Render},
		{"WarningStyle", "⚠ Warning", WarningStyle.Render},
		{"InfoStyle", "ℹ Info", InfoStyle.Render},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.value)
			if result == "" {
				t.Errorf("%s.Render(%q) returned empty string", tt.name, tt.value)
			}
			if !strings.Contains(result, tt.value) {
				t.Errorf("%s.Render(%q) = %q, should contain %q", tt.name, tt.value, result, tt.value)
			}
		})
	}
}

// TestApplyTheme_RebuildsStyles guards against the shared styles freezing the
// theme captured at package init: after a flavor change they must render with
// the new theme's tokens.
func TestApplyTheme_RebuildsStyles(t *testing.T) {
	savedFlavor, savedTheme := theme.Flavor, theme.DefaultTheme
	t.Cleanup(func() {
		theme.Flavor = savedFlavor
		ApplyTheme(savedTheme)
	})

	// Start from the package-init default (Mocha on a dark terminal).
	theme.Flavor = ""
	mocha := theme.Resolve(true)
	ApplyTheme(mocha)

	theme.Flavor = "latte"
	latte := theme.Resolve(true)
	if sameColor(mocha.SuccessText, latte.SuccessText) {
		t.Fatal("test precondition: latte and mocha SuccessText must differ")
	}
	ApplyTheme(latte)

	if !sameColor(theme.DefaultTheme.Bg, latte.Bg) {
		t.Error("ApplyTheme did not set theme.DefaultTheme")
	}
	if !sameColor(CurrentTheme.Bg, latte.Bg) {
		t.Error("ApplyTheme did not set CurrentTheme")
	}

	tests := []struct {
		name string
		got  color.Color
		want color.Color
	}{
		{"TitleStyle", TitleStyle.GetForeground(), latte.PrimaryBorder},
		{"SubtitleStyle", SubtitleStyle.GetForeground(), latte.SecondaryText},
		{"SuccessStyle", SuccessStyle.GetForeground(), latte.SuccessText},
		{"ErrorStyle", ErrorStyle.GetForeground(), latte.ErrorText},
		{"WarningStyle", WarningStyle.GetForeground(), latte.WarningText},
		{"InfoStyle", InfoStyle.GetForeground(), latte.InfoText},
		{"PopupStyle", PopupStyle.GetBorderTopForeground(), latte.ErrorText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !sameColor(tt.got, tt.want) {
				t.Errorf("%s color = %v, want %v (latte)", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ── KeyMap tests ─────────────────────────────────────────────────────────────

func TestMenuKeyMap(t *testing.T) {
	km := MenuKeyMap()
	if km == nil {
		t.Fatal("MenuKeyMap() returned nil")
	}

	// Verify Quit binding has all expected keys
	quitKeys := km.Quit.Keys()
	expectedQuit := []string{"esc", "ctrl+c", "left", "backspace"}
	for _, k := range expectedQuit {
		if !hasKey(quitKeys, k) {
			t.Errorf("MenuKeyMap Quit should bind %s, got keys: %v", k, quitKeys)
		}
	}

	// Verify Select Submit has enter and right
	selectKeys := km.Select.Submit.Keys()
	if !hasKey(selectKeys, "enter") || !hasKey(selectKeys, "right") {
		t.Errorf("MenuKeyMap Select Submit should bind enter and right, got: %v", selectKeys)
	}

	// Verify MultiSelect Submit has enter only
	multiKeys := km.MultiSelect.Submit.Keys()
	if !hasKey(multiKeys, "enter") {
		t.Errorf("MenuKeyMap MultiSelect Submit should bind enter, got: %v", multiKeys)
	}
	if hasKey(multiKeys, "right") {
		t.Error("MenuKeyMap MultiSelect Submit should NOT bind right")
	}
}

func TestConfirmKeyMap(t *testing.T) {
	km := ConfirmKeyMap()
	if km == nil {
		t.Fatal("ConfirmKeyMap() returned nil")
	}

	// Verify Quit binding only has ctrl+c (no left/backspace)
	quitKeys := km.Quit.Keys()
	expectedQuit := []string{"ctrl+c"}
	for _, k := range expectedQuit {
		if !hasKey(quitKeys, k) {
			t.Errorf("ConfirmKeyMap Quit should bind %s, got keys: %v", k, quitKeys)
		}
	}
	unexpectedQuit := []string{"left", "backspace", "esc"}
	for _, k := range unexpectedQuit {
		if hasKey(quitKeys, k) {
			t.Errorf("ConfirmKeyMap Quit should NOT bind %s, got keys: %v", k, quitKeys)
		}
	}

	// Verify Confirm Accept/Reject bindings
	acceptKeys := km.Confirm.Accept.Keys()
	if !hasKey(acceptKeys, "y") || !hasKey(acceptKeys, "Y") {
		t.Errorf("ConfirmKeyMap Accept should bind y/Y, got: %v", acceptKeys)
	}

	rejectKeys := km.Confirm.Reject.Keys()
	if !hasKey(rejectKeys, "n") || !hasKey(rejectKeys, "N") {
		t.Errorf("ConfirmKeyMap Reject should bind n/N, got: %v", rejectKeys)
	}

	// Verify Confirm Submit has enter
	submitKeys := km.Confirm.Submit.Keys()
	if !hasKey(submitKeys, "enter") {
		t.Errorf("ConfirmKeyMap Submit should bind enter, got: %v", submitKeys)
	}
}

// ── AppTheme tests ───────────────────────────────────────────────────────────

func TestAppTheme(t *testing.T) {
	theme := AppTheme
	if theme == nil {
		t.Error("AppTheme is nil")
	}
	// Theme should return a non-nil styleset
	s := theme.Theme(true)
	if s == nil {
		t.Error("AppTheme.Theme(true) returned nil")
	}
	s = theme.Theme(false)
	if s == nil {
		t.Error("AppTheme.Theme(false) returned nil")
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func hasKey(keys []string, expected string) bool {
	for _, k := range keys {
		if k == expected {
			return true
		}
	}
	return false
}

// Ensure fmt is used (for RenderHeader)
var _ = fmt.Println
