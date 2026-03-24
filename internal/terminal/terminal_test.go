package terminal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanthor/bluefin-cli/internal/install"
)

func TestSetWindowsTerminalPath(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	initial := map[string]interface{}{
		"profiles": map[string]interface{}{
			"defaults": map[string]interface{}{},
		},
		"schemes": []interface{}{},
	}
	data, _ := json.MarshalIndent(initial, "", "    ")
	os.WriteFile(settingsPath, data, 0644)

	font := install.NerdFont{Name: "JetBrains Mono", Face: "JetBrainsMonoNF"}
	if err := setWindowsTerminalPath(settingsPath, font, "catppuccin-frappe"); err != nil {
		t.Fatalf("setWindowsTerminalPath failed: %v", err)
	}

	newData, _ := os.ReadFile(settingsPath)
	var settings map[string]interface{}
	json.Unmarshal(newData, &settings)

	profiles := settings["profiles"].(map[string]interface{})
	defaults := profiles["defaults"].(map[string]interface{})
	fontSettings := defaults["font"].(map[string]interface{})

	if fontSettings["face"] != "JetBrainsMonoNF" {
		t.Errorf("expected font face JetBrainsMonoNF, got %v", fontSettings["face"])
	}
	if defaults["colorScheme"] != "Catppuccin Frappe" {
		t.Errorf("expected color scheme Catppuccin Frappe, got %v", defaults["colorScheme"])
	}
}

func TestSetGhosttyPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	font := install.NerdFont{Name: "JetBrains Mono", Face: "JetBrainsMonoNF"}

	if err := setGhosttyPath(configPath, font, "catppuccin-frappe"); err != nil {
		t.Fatalf("setGhosttyPath failed for new file: %v", err)
	}
	content, _ := os.ReadFile(configPath)
	if !strings.Contains(string(content), "font-family = JetBrainsMonoNF") {
		t.Errorf("expected font-family in config, got %s", content)
	}
	if !strings.Contains(string(content), "theme = catppuccin-frappe") {
		t.Errorf("expected theme in config, got %s", content)
	}

	// Update existing file
	os.WriteFile(configPath, []byte("font-family = OldFont\ntheme = old-theme\nother-setting = true"), 0644)
	if err := setGhosttyPath(configPath, font, "catppuccin-frappe"); err != nil {
		t.Fatalf("setGhosttyPath failed for existing file: %v", err)
	}
	content, _ = os.ReadFile(configPath)
	if !strings.Contains(string(content), "font-family = JetBrainsMonoNF") {
		t.Errorf("expected updated font-family, got %s", content)
	}
	if strings.Contains(string(content), "font-family = OldFont") {
		t.Error("old font-family still present")
	}
	if !strings.Contains(string(content), "other-setting = true") {
		t.Error("other settings lost in update")
	}
}

func TestSetVSCodePath(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")
	font := install.NerdFont{Name: "JetBrains Mono", Face: "JetBrainsMonoNF"}

	if err := setVSCodePath(settingsPath, font, "catppuccin-frappe"); err != nil {
		t.Fatalf("setVSCodePath failed for new file: %v", err)
	}
	data, _ := os.ReadFile(settingsPath)
	var settings map[string]interface{}
	json.Unmarshal(data, &settings)

	if settings["terminal.integrated.fontFamily"] != "JetBrainsMonoNF" {
		t.Errorf("expected terminal.integrated.fontFamily JetBrainsMonoNF, got %v", settings["terminal.integrated.fontFamily"])
	}
	if settings["workbench.colorTheme"] != "Catppuccin Frappe" {
		t.Errorf("expected workbench.colorTheme Catppuccin Frappe, got %v", settings["workbench.colorTheme"])
	}

	// Update existing file
	initial := map[string]interface{}{"terminal.integrated.fontFamily": "OldFont", "other.setting": true}
	data, _ = json.MarshalIndent(initial, "", "    ")
	os.WriteFile(settingsPath, data, 0644)

	if err := setVSCodePath(settingsPath, font, "catppuccin-frappe"); err != nil {
		t.Fatalf("setVSCodePath failed for existing file: %v", err)
	}
	data, _ = os.ReadFile(settingsPath)
	json.Unmarshal(data, &settings)

	if settings["terminal.integrated.fontFamily"] != "JetBrainsMonoNF" {
		t.Errorf("expected updated font family, got %v", settings["terminal.integrated.fontFamily"])
	}
	if settings["other.setting"] != true {
		t.Error("other settings lost in update")
	}
}

func TestDetectTerminal(t *testing.T) {
	tests := []struct {
		termProgram string
		expected    TerminalType
	}{
		{"vscode", VSCode},
		{"VSCode", VSCode},
		{"vscodium", VSCodium},
		{"antigravity", Antigravity},
		{"google-antigravity", Antigravity},
		{"ghostty", Ghostty},
		{"something-else", Unknown},
	}
	oldWT := os.Getenv("WT_SESSION")
	os.Setenv("WT_SESSION", "")
	defer os.Setenv("WT_SESSION", oldWT)

	for _, tt := range tests {
		t.Run(tt.termProgram, func(t *testing.T) {
			os.Setenv("TERM_PROGRAM", tt.termProgram)
			if got := DetectTerminal(); got != tt.expected {
				t.Errorf("DetectTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}
