// Shell enablement, init-script rendering and cache, and status reporting.
//
// Package installation lives next door: installers.go owns the Homebrew and
// cross-platform paths, windows_tools.go the winget/PowerShell ones, and
// install_alpine.go the coldbrew/apk ones. Before that split
// (tuna-os/bluefin-cli#217) all of it shared one module, so a change to a
// platform backend sat in the same file as the interactive startup path and
// neither could be read or evolved without the other.
package shell

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

//go:embed resources/shell.sh
var shellShScript string

//go:embed resources/shell.fish
var shellFishScript string

//go:embed resources/shell.nu
var shellNuScript string

//go:embed resources/shell.ps1
var shellPowerShellScript string

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

const shellMaker = "# bluefin-cli shell-config"
const blingMarker = "# bluefin-cli bling"

func Toggle(shell string, enable bool) error {
	if isPowerShellShell(shell) {
		return togglePowerShell(enable)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	spec, ok := LookupShell(shell)
	if !ok {
		return fmt.Errorf("unsupported shell: %s (supported: %s)", shell, strings.Join(ManagedShells(), ", "))
	}
	shell = spec.Name
	configFile := spec.ConfigPath(home)
	rcLine := spec.RCLine(home)

	// Nushell cannot evaluate a string, so config.nu sources a file instead of
	// piping `bluefin-cli init` into the shell. Render that file here, before
	// the line that sources it is written.
	if enable && spec.GeneratedInit != "" {
		if err := writeGeneratedInit(spec, home); err != nil {
			return err
		}
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) && enable {
			// Create if it doesn't exist and we are enabling. fish and nushell
			// both keep their config under ~/.config, which may be absent.
			if err := spec.ensureRCDir(home); err != nil {
				return err
			}
			content = []byte("")
		} else if os.IsNotExist(err) && !enable {
			fmt.Println(infoStyle.Render(fmt.Sprintf("%s is already disabled for %s", shell, shell)))
			return nil
		} else {
			return err
		}
	}

	text := string(content)
	hasLine := strings.Contains(text, shellMaker)

	if enable {
		if hasLine {
			fmt.Println(infoStyle.Render(fmt.Sprintf("%s is already enabled for %s", shell, shell)))
			if cfg, err := LoadConfig(shell); err == nil {
				InstallTools(shell, cfg)
			}
			return nil
		}

		f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		prefix := "\n"
		if len(text) == 0 || strings.HasSuffix(text, "\n") {
			prefix = ""
		}

		if _, err := f.WriteString(prefix + rcLine + "\n"); err != nil {
			return err
		}
		// ash has no rc file of its own: an interactive shell reads whatever
		// $ENV names. Writing ~/.ashrc alone leaves it dormant, so point $ENV
		// at it from ~/.profile too.
		if shell == "ash" {
			if err := ensureAshENV(home); err != nil {
				return err
			}
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Enabled shell experience for %s", shell)))
	} else {
		if !hasLine {
			fmt.Println(infoStyle.Render(fmt.Sprintf("%s is already disabled for %s", shell, shell)))
			return nil
		}

		// Remove the lines containing the marker
		lines := strings.Split(text, "\n")
		var newLines []string
		for _, line := range lines {
			if !strings.Contains(line, shellMaker) && !strings.Contains(line, blingMarker) {
				newLines = append(newLines, line)
			}
		}

		output := strings.Join(newLines, "\n")
		// Trim extra newlines at the end
		output = strings.TrimRight(output, "\n") + "\n"

		if err := os.WriteFile(configFile, []byte(output), 0644); err != nil {
			return err
		}
		if shell == "ash" {
			if err := removeAshENV(home); err != nil {
				return err
			}
		}
		if path := spec.GeneratedInitPath(home); path != "" {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Disabled shell experience for %s", shell)))
	}

	if enable {
		if cfg, err := LoadConfig(shell); err == nil {
			InstallTools(shell, cfg)
		}
	}

	return nil
}

// initCacheMtime returns the modification time of the shell config file
// (used to decide whether the cached init output is still valid).
func initCacheMtime(shell string) (time.Time, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return time.Time{}, err
	}
	info, err := os.Stat(configPath)
	if err != nil {
		// Config doesn't exist yet — treat as epoch so any cache is
		// considered newer and we skip the expensive tool checks.
		return time.Time{}, nil
	}
	return info.ModTime(), nil
}

// loadInitCache returns the cached init script for shell if it's still fresh
// (binary and config haven't changed since the cache was written).
func loadInitCache(shell string) (string, bool) {
	cachePath, err := getInitCachePath(shell)
	if err != nil {
		return "", false
	}
	cacheInfo, err := os.Stat(cachePath)
	if err != nil {
		return "", false
	}
	cacheMtime := cacheInfo.ModTime()

	// Must be newer than the bluefin-cli binary.
	if exe, err := os.Executable(); err == nil {
		if exeInfo, err := os.Stat(exe); err == nil {
			if !cacheMtime.After(exeInfo.ModTime()) {
				return "", false
			}
		}
	}

	// Must be newer than the shell config file.
	if cfgMtime, err := initCacheMtime(shell); err == nil {
		if !cacheMtime.After(cfgMtime) {
			return "", false
		}
	}

	content, err := os.ReadFile(cachePath)
	if err != nil {
		return "", false
	}
	return string(content), true
}

// saveInitCache writes the generated init script to a cache file.
func saveInitCache(shell, script string) {
	cachePath, err := getInitCachePath(shell)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return
	}
	_ = os.WriteFile(cachePath, []byte(script), 0644)
}

// getInitCachePath returns the path to the cached init script for a shell.
func getInitCachePath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "bluefin-cli", fmt.Sprintf("init-%s.sh", shell)), nil
}

func Init(shell string, config *Config) (string, error) {
	if config == nil {
		config = DefaultConfig(shell)
	}

	// Try to serve a cached init script to avoid expensive tool-availability
	// checks on every interactive shell startup (issue #59). The cache is
	// invalidated when the bluefin-cli binary or the shell config changes.
	if cached, ok := loadInitCache(shell); ok {
		return cached, nil
	}

	tools := ToolsForShell(shell)
	configUpdated := false

	// Synchronize configuration with installed tools in PATH
	for _, tool := range tools {
		isInstalled := isBinaryAvailable(tool)

		isEnabled := config.IsEnabled(tool.Name)

		if isEnabled && !isInstalled {
			fmt.Fprintf(os.Stderr, "bluefin-cli: %s is enabled but '%s' was not found in PATH.\n", tool.Name, tool.Binary)
			fmt.Fprintf(os.Stderr, "  - Install it: brew install %s\n", tool.GetBrewPkg())
			fmt.Fprintf(os.Stderr, "  - Or disable it: bluefin-cli shell config\n")

			config.SetEnabled(tool.Name, false)
			configUpdated = true

			// If Glow is disabled, also disable MOTD
			if tool.Name == "Glow" {
				config.SetEnabled("Motd", false)
			}
		} else if !isEnabled && isInstalled {
			// Auto-enable if found and previously disabled
			config.SetEnabled(tool.Name, true)
			configUpdated = true
			fmt.Fprintf(os.Stderr, "bluefin-cli: %s found in PATH, automatically enabling.\n", tool.Name)

			// If Glow is found, also enable MOTD
			if tool.Name == "Glow" {
				config.SetEnabled("Motd", true)
			}
		}
	}

	// MOTD consistency check (if manual override happened)
	if config.IsEnabled("Motd") && !config.IsEnabled("Glow") {
		// If MOTD enabled but Glow disabled, try to enable Glow
		if _, err := exec.LookPath("glow"); err == nil {
			config.SetEnabled("Glow", true)
			configUpdated = true
		} else {
			// Glow missing, must disable MOTD
			config.SetEnabled("Motd", false)
			configUpdated = true
		}
	}

	if configUpdated {
		if err := SaveConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "bluefin-cli: failed to save updated config: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "bluefin-cli: configuration synchronized with installed components.\n")
		}
	}

	var sb strings.Builder

	if isPowerShellShell(shell) {
		for _, tool := range tools {
			enabled := config.IsEnabled(tool.Name)
			fmt.Fprintf(&sb, "$env:%s = \"%d\"\n", tool.GetEnvVar(), boolToInt(enabled))
		}
		fmt.Fprintf(&sb, "$env:BLUEFIN_SHELL_ENABLE_MOTD = \"%d\"\n", boolToInt(config.IsEnabled("Motd")))

		sb.WriteString("\n")
		sb.WriteString(shellPowerShellScript)
		result := sb.String()
		saveInitCache(shell, result)
		return result, nil
	}

	// Unknown shells still render the POSIX script, as they always have: an
	// unrecognized name reaching `init` is far more likely to be a
	// sh-compatible shell than anything else.
	spec, ok := LookupShell(shell)
	if !ok {
		spec = Shell{Name: shell, Flavor: "posix"}
	}

	for _, tool := range tools {
		sb.WriteString(spec.exportVar(tool.GetEnvVar(), boolToInt(config.IsEnabled(tool.Name))))
	}
	sb.WriteString(spec.exportVar("BLUEFIN_SHELL_ENABLE_MOTD", boolToInt(config.IsEnabled("Motd"))))
	sb.WriteString(spec.exportString("BLING_SHELL", spec.Name))

	sb.WriteString("\n")
	sb.WriteString(spec.script())

	result := sb.String()
	saveInitCache(shell, result)
	return result, nil
}

func CheckStatus() map[string]bool {
	status := make(map[string]bool)
	home, _ := os.UserHomeDir()

	for _, spec := range registry {
		content, err := os.ReadFile(spec.ConfigPath(home))
		if err != nil {
			status[spec.Name] = false
			continue
		}

		status[spec.Name] = strings.Contains(string(content), shellMaker) || strings.Contains(string(content), blingMarker)
	}
	// Nushell is spelled both ways in the wild; report it under both.
	status["nushell"] = status["nu"]

	status["powershell"] = checkPowerShellStatus()
	status["pwsh"] = status["powershell"]

	return status
}

// GetInstalledShells returns a list of shells that are available in the PATH
func GetInstalledShells() []string {
	var installed []string

	for _, spec := range registry {
		if spec.IsInstalled() {
			installed = append(installed, spec.Name)
		}
	}

	if _, err := exec.LookPath("pwsh"); err == nil {
		installed = append(installed, "pwsh")
	} else if _, err := exec.LookPath("powershell"); err == nil {
		installed = append(installed, "powershell")
	} else if _, err := exec.LookPath("powershell.exe"); err == nil {
		installed = append(installed, "powershell")
	}

	return installed
}

// writeGeneratedInit renders the init script for a shell that sources a file
// rather than evaluating a pipe, and writes it next to that shell's config.
func writeGeneratedInit(spec Shell, home string) error {
	path := spec.GeneratedInitPath(home)
	if path == "" {
		return nil
	}
	if err := spec.ensureRCDir(home); err != nil {
		return err
	}
	cfg, err := LoadConfig(spec.Name)
	if err != nil {
		cfg = DefaultConfig(spec.Name)
	}
	script, err := Init(spec.Name, cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(script), 0644)
}

// ensureAshENV points $ENV at ~/.ashrc from ~/.profile, which is what makes
// an interactive ash read the rc file at all. It is a no-op once the line is
// present.
func ensureAshENV(home string) error {
	profile := filepath.Join(home, ".profile")
	content, err := os.ReadFile(profile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	text := string(content)
	if strings.Contains(text, ashENVLine) {
		return nil
	}
	prefix := "\n"
	if len(text) == 0 || strings.HasSuffix(text, "\n") {
		prefix = ""
	}
	f, err := os.OpenFile(profile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(prefix + ashENVLine + "\n")
	return err
}

// removeAshENV takes the $ENV line back out of ~/.profile. Only lines
// carrying both the export and the bluefin marker are touched, so a user's
// own $ENV setup survives.
func removeAshENV(home string) error {
	profile := filepath.Join(home, ".profile")
	content, err := os.ReadFile(profile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var kept []string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, shellMaker) && strings.Contains(line, "ENV=") {
			continue
		}
		kept = append(kept, line)
	}
	output := strings.TrimRight(strings.Join(kept, "\n"), "\n") + "\n"
	return os.WriteFile(profile, []byte(output), 0644)
}
