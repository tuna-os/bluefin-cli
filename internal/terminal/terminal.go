package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/install"
)

// IsWSL is a convenience re-export from env.
func IsWSL() bool { return env.IsWSL() }

// TerminalType identifies a supported terminal or IDE.
type TerminalType string

const (
	WindowsTerminal TerminalType = "Windows Terminal"
	Ghostty         TerminalType = "Ghostty"
	ITerm2          TerminalType = "iTerm2"
	TerminalApp     TerminalType = "Terminal.app"
	Ptyxis          TerminalType = "Ptyxis"
	GnomeConsole    TerminalType = "Gnome Console"
	Konsole         TerminalType = "Konsole"
	VSCode          TerminalType = "VS Code"
	VSCodium        TerminalType = "VSCodium"
	Antigravity     TerminalType = "Antigravity"
	Unknown         TerminalType = "Unknown"
)

// SetFontAndTheme configures all discovered terminals with the given font and Catppuccin theme.
func SetFontAndTheme(font install.NerdFont, theme string) error {
	targets := DiscoverTerminals()
	if len(targets) == 0 {
		return fmt.Errorf("no supported terminals discovered")
	}
	return SetFontAndThemeToTargets(targets, font, theme)
}

// SetFontAndThemeToTargets applies font/theme settings to specific terminal types.
func SetFontAndThemeToTargets(targets []TerminalType, font install.NerdFont, theme string) error {
	var errs []error
	for _, term := range targets {
		var err error
		switch term {
		case WindowsTerminal:
			err = setWindowsTerminal(font, theme)
		case Ghostty:
			err = setGhostty(font, theme)
		case VSCode, VSCodium, Antigravity:
			err = setVSCodeFlavor(term, font, theme)
		default:
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", term, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors configuring terminals: %v", errs)
	}
	return nil
}

// SetLiveFontAndTheme applies settings only to the currently running terminal.
func SetLiveFontAndTheme(font install.NerdFont, theme string) error {
	for _, term := range GetActiveTerminals() {
		switch term {
		case WindowsTerminal:
			_ = setWindowsTerminal(font, theme)
		case Ghostty:
			_ = setGhostty(font, theme)
		case VSCode, VSCodium, Antigravity:
			_ = setVSCodeFlavor(term, font, theme)
		}
	}
	return nil
}

// DiscoverTerminals scans the system for supported terminal/IDE installations.
func DiscoverTerminals() []TerminalType {
	var found []TerminalType
	seen := map[TerminalType]bool{}

	add := func(t TerminalType) {
		if !seen[t] {
			seen[t] = true
			found = append(found, t)
		}
	}

	// VS Code family — check PATH and common Windows install dirs
	for _, check := range []struct {
		term       TerminalType
		bin        string
		folderName string
	}{
		{VSCode, "code", "Code"},
		{VSCodium, "codium", "VSCodium"},
		{Antigravity, "antigravity", "Antigravity"},
	} {
		if _, err := resolveVSCodeBinary(check.term); err == nil {
			add(check.term)
		} else if runtime.GOOS == "windows" {
			appData := os.Getenv("APPDATA")
			if _, err := os.Stat(filepath.Join(appData, check.folderName)); err == nil {
				add(check.term)
			}
		}
	}

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		wt := filepath.Join(localAppData, "Packages/Microsoft.WindowsTerminal_8wekyb3d8bbwe/LocalState/settings.json")
		if _, err := os.Stat(wt); err == nil {
			add(WindowsTerminal)
		}
	} else {
		if _, err := exec.LookPath("ghostty"); err == nil {
			add(Ghostty)
		}
	}

	return found
}

// DetectTerminal returns the terminal emulator the CLI is currently running inside.
func DetectTerminal() TerminalType {
	if runtime.GOOS == "windows" && os.Getenv("WT_SESSION") != "" {
		return WindowsTerminal
	}
	if os.Getenv("ANTIGRAVITY_EDITOR_APP_ROOT") != "" {
		return Antigravity
	}
	if os.Getenv("VSCODE_VSC_BASE_URL") != "" || os.Getenv("VSCODE_VSC_BASE") != "" {
		return VSCodium
	}
	switch strings.ToLower(os.Getenv("TERM_PROGRAM")) {
	case "ghostty":
		return Ghostty
	case "vscode":
		cacheDir := strings.ToLower(os.Getenv("VSCODE_CODE_CACHE_PATH"))
		if strings.Contains(cacheDir, "antigravity") {
			return Antigravity
		}
		if strings.Contains(cacheDir, "codium") {
			return VSCodium
		}
		return VSCode
	case "vscodium", "codium":
		return VSCodium
	case "antigravity", "google-antigravity":
		return Antigravity
	case "iterm.app":
		return ITerm2
	case "apple_terminal":
		return TerminalApp
	}
	if runtime.GOOS == "linux" && os.Getenv("PTYXIS_VERSION") != "" {
		return Ptyxis
	}
	return Unknown
}

// GetActiveTerminals returns terminals/IDEs currently running on this machine.
func GetActiveTerminals() []TerminalType {
	active := map[TerminalType]bool{}

	if t := DetectTerminal(); t != Unknown {
		active[t] = true
	}

	if runtime.GOOS == "windows" {
		if out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output(); err == nil {
			s := string(out)
			if strings.Contains(s, `"Code.exe"`) {
				active[VSCode] = true
			}
			if strings.Contains(s, `"VSCodium.exe"`) {
				active[VSCodium] = true
			}
			if strings.Contains(s, `"Antigravity.exe"`) {
				active[Antigravity] = true
			}
		}
		localAppData := os.Getenv("LOCALAPPDATA")
		wt := filepath.Join(localAppData, "Packages/Microsoft.WindowsTerminal_8wekyb3d8bbwe/LocalState/settings.json")
		if _, err := os.Stat(wt); err == nil {
			active[WindowsTerminal] = true
		}
	}

	var result []TerminalType
	for t := range active {
		result = append(result, t)
	}
	return result
}

// --- Windows Terminal --------------------------------------------------------

func setWindowsTerminal(font install.NerdFont, theme string) error {
	localAppData := os.Getenv("LOCALAPPDATA")
	paths := []string{
		filepath.Join(localAppData, "Packages/Microsoft.WindowsTerminal_8wekyb3d8bbwe/LocalState/settings.json"),
		filepath.Join(localAppData, "Packages/Microsoft.WindowsTerminalPreview_8wekyb3d8bbwe/LocalState/settings.json"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return setWindowsTerminalPath(p, font, theme)
		}
	}
	return fmt.Errorf("Windows Terminal settings.json not found")
}

func setWindowsTerminalPath(settingsPath string, font install.NerdFont, theme string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	profiles, _ := settings["profiles"].(map[string]interface{})
	if profiles == nil {
		return fmt.Errorf("invalid settings.json: profiles not found")
	}
	defaults, _ := profiles["defaults"].(map[string]interface{})
	if defaults == nil {
		defaults = map[string]interface{}{}
		profiles["defaults"] = defaults
	}
	fontSettings, _ := defaults["font"].(map[string]interface{})
	if fontSettings == nil {
		fontSettings = map[string]interface{}{}
		defaults["font"] = fontSettings
	}
	fontSettings["face"] = font.Face

	if theme != "" {
		if colors := getCatppuccinScheme(theme); colors != nil {
			schemeName := colors["name"]
			schemes, _ := settings["schemes"].([]interface{})
			found := false
			for i, s := range schemes {
				if scheme, ok := s.(map[string]interface{}); ok && scheme["name"] == schemeName {
					schemes[i] = colors
					found = true
					break
				}
			}
			if !found {
				settings["schemes"] = append(schemes, colors)
			}
			defaults["colorScheme"] = schemeName
		}
	}

	newData, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, newData, 0644)
}

// --- Ghostty ----------------------------------------------------------------

func setGhostty(font install.NerdFont, theme string) error {
	if runtime.GOOS == "windows" {
		return nil // Ghostty not on Windows
	}
	home, _ := os.UserHomeDir()
	return setGhosttyPath(filepath.Join(home, ".config/ghostty/config"), font, theme)
}

func setGhosttyPath(configPath string, font install.NerdFont, theme string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			return err
		}
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	out := make([]string, 0, len(lines))
	fontSet, themeSet := false, false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "font-family =") {
			out = append(out, "font-family = "+font.Face)
			fontSet = true
		} else if theme != "" && strings.HasPrefix(trimmed, "theme =") {
			out = append(out, "theme = "+theme)
			themeSet = true
		} else {
			out = append(out, line)
		}
	}
	if !fontSet {
		out = append(out, "font-family = "+font.Face)
	}
	if theme != "" && !themeSet {
		out = append(out, "theme = "+theme)
	}
	return os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0644)
}

// --- VS Code family ---------------------------------------------------------

func setVSCodeFlavor(flavor TerminalType, font install.NerdFont, theme string) error {
	var folderName string
	switch flavor {
	case VSCodium:
		folderName = "VSCodium"
	case Antigravity:
		folderName = "Antigravity"
	default:
		folderName = "Code"
	}

	home, _ := os.UserHomeDir()
	var settingsPath string
	switch runtime.GOOS {
	case "windows":
		settingsPath = filepath.Join(os.Getenv("APPDATA"), folderName, "User/settings.json")
	case "darwin":
		settingsPath = filepath.Join(home, "Library/Application Support", folderName, "User/settings.json")
	default:
		settingsPath = filepath.Join(home, ".config", folderName, "User/settings.json")
	}
	return setVSCodePath(settingsPath, font, theme)
}

func setVSCodePath(settingsPath string, font install.NerdFont, theme string) error {
	// Infer flavor from path for background extension installation
	var flavor TerminalType = VSCode
	lp := strings.ToLower(settingsPath)
	if strings.Contains(lp, "codium") {
		flavor = VSCodium
	} else if strings.Contains(lp, "antigravity") {
		flavor = Antigravity
	}
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(settingsPath, []byte("{}"), 0644); err != nil {
			return err
		}
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	settings["terminal.integrated.fontFamily"] = font.Face

	if theme != "" {
		themeMap := map[string]string{
			"catppuccin-frappe":    "Catppuccin Frappe",
			"catppuccin-macchiato": "Catppuccin Macchiato",
			"catppuccin-mocha":     "Catppuccin Mocha",
			"catppuccin-latte":     "Catppuccin Latte",
		}
		vsTheme := themeMap[theme]
		if vsTheme == "" {
			vsTheme = "Catppuccin Frappe"
		}
		settings["workbench.colorTheme"] = vsTheme
		// Install Catppuccin extension in the background
		go func() {
			_ = installVSCodeExtension(flavor, "catppuccin.catppuccin-vsc")
			_ = installVSCodeExtension(flavor, "catppuccin.catppuccin-vsc-icons")
		}()
	}

	newData, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, newData, 0644)
}

func installVSCodeExtension(flavor TerminalType, extensionID string) error {
	bin, err := resolveVSCodeBinary(flavor)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(bin), ".cmd") {
		cmd = exec.Command("cmd", "/c", bin, "--install-extension", extensionID)
	} else {
		cmd = exec.Command(bin, "--install-extension", extensionID)
	}
	return cmd.Start()
}

func resolveVSCodeBinary(flavor TerminalType) (string, error) {
	var binName, folderName string
	switch flavor {
	case VSCodium:
		binName, folderName = "codium", "VSCodium"
	case Antigravity:
		binName, folderName = "antigravity", "Antigravity"
	default:
		binName, folderName = "code", "Microsoft VS Code"
	}

	if path, err := exec.LookPath(binName); err == nil {
		return path, nil
	}
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		for _, base := range []string{localAppData, os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
			p := filepath.Join(base, "Programs", folderName, "bin", binName+".cmd")
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("could not find CLI binary for %s", flavor)
}

// --- Catppuccin color schemes -----------------------------------------------

func getCatppuccinScheme(theme string) map[string]string {
	switch theme {
	case "catppuccin-frappe":
		return map[string]string{
			"name": "Catppuccin Frappe", "background": "#303446", "foreground": "#c6d0f5",
			"black": "#51576d", "blue": "#8caaee", "brightBlack": "#626880", "brightBlue": "#8caaee",
			"brightCyan": "#81c8be", "brightGreen": "#a6d189", "brightPurple": "#f4b8e4",
			"brightRed": "#e78284", "brightWhite": "#a5adce", "brightYellow": "#e5c890",
			"cursorColor": "#f2d5cf", "cyan": "#81c8be", "green": "#a6d189", "purple": "#f4b8e4",
			"red": "#e78284", "selectionBackground": "#414559", "white": "#b5bfe3", "yellow": "#e5c890",
		}
	case "catppuccin-macchiato":
		return map[string]string{
			"name": "Catppuccin Macchiato", "background": "#24273a", "foreground": "#cad3f5",
			"black": "#494d64", "blue": "#8aadf4", "brightBlack": "#5b6078", "brightBlue": "#8aadf4",
			"brightCyan": "#8bd5ca", "brightGreen": "#a6da95", "brightPurple": "#f5bde6",
			"brightRed": "#ed8796", "brightWhite": "#a5adcb", "brightYellow": "#eed49f",
			"cursorColor": "#f4dbd6", "cyan": "#8bd5ca", "green": "#a6da95", "purple": "#f5bde6",
			"red": "#ed8796", "selectionBackground": "#363a4f", "white": "#b8c0e0", "yellow": "#eed49f",
		}
	case "catppuccin-mocha":
		return map[string]string{
			"name": "Catppuccin Mocha", "background": "#1e1e2e", "foreground": "#cdd6f4",
			"black": "#45475a", "blue": "#89b4fa", "brightBlack": "#585b70", "brightBlue": "#89b4fa",
			"brightCyan": "#94e2d5", "brightGreen": "#a6e3a1", "brightPurple": "#f5c2e7",
			"brightRed": "#f38ba8", "brightWhite": "#a6adc8", "brightYellow": "#f9e2af",
			"cursorColor": "#f5e0dc", "cyan": "#94e2d5", "green": "#a6e3a1", "purple": "#f5c2e7",
			"red": "#f38ba8", "selectionBackground": "#313244", "white": "#bac2de", "yellow": "#f9e2af",
		}
	case "catppuccin-latte":
		return map[string]string{
			"name": "Catppuccin Latte", "background": "#eff1f5", "foreground": "#4c4f69",
			"black": "#dce0e8", "blue": "#1e66f5", "brightBlack": "#bcc0cc", "brightBlue": "#1e66f5",
			"brightCyan": "#179287", "brightGreen": "#40a02b", "brightPurple": "#ea76cb",
			"brightRed": "#d20f39", "brightWhite": "#9ca0b0", "brightYellow": "#df8e1d",
			"cursorColor": "#dc8a78", "cyan": "#179287", "green": "#40a02b", "purple": "#ea76cb",
			"red": "#d20f39", "selectionBackground": "#ccd0da", "white": "#acb0be", "yellow": "#df8e1d",
		}
	}
	return nil
}
