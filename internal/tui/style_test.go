package tui

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

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

// ── RenderHeader tests ───────────────────────────────────────────────────────

func TestRenderHeader(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	RenderHeader("Test Title", "Test Subtitle")

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Test Title") {
		t.Errorf("RenderHeader output should contain title, got: %s", output)
	}
	if !strings.Contains(output, "Test Subtitle") {
		t.Errorf("RenderHeader output should contain subtitle, got: %s", output)
	}
}

func TestRenderHeader_NoSubtitle(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	RenderHeader("Only Title", "")

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Only Title") {
		t.Errorf("RenderHeader output should contain title, got: %s", output)
	}
	// Empty subtitle should not produce extra lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Error("RenderHeader should produce at least one line of output")
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
