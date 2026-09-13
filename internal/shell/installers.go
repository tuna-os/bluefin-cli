// Tool installation: the shared entry points and the Homebrew backend.
//
// Split out of shell.go (tuna-os/bluefin-cli#217), which had grown to route
// three package-manager backends alongside shell rendering and status
// reporting. The public API is unchanged -- InstallTools and EnsureInstalled
// are still the way callers install things, and they still dispatch by
// platform to the Windows and Alpine backends in their own files.
package shell

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"charm.land/huh/v2"
	"github.com/tuna-os/bluefin-cli/internal/env"
)

func toolsForCurrentPlatform() []Tool {
	if runtime.GOOS == "windows" {
		return ToolsForShell("powershell")
	}

	// For non-Windows platforms, filter out tools that are explicitly unsupported on common shells.
	// This ensures tools like gsudo (Windows-only) are not included.
	filtered := make([]Tool, 0, len(Tools))
	for _, tool := range Tools {
		if tool.SupportsShell("bash") || tool.SupportsShell("zsh") || tool.SupportsShell("fish") {
			filtered = append(filtered, tool)
		}
	}

	return filtered
}

// InstallTools iterates through the config and installs enabled tools
func InstallTools(shell string, cfg *Config) {
	// If MOTD is enabled, ensure Glow is also considered enabled for installation
	if cfg.IsEnabled("Motd") {
		cfg.SetEnabled("Glow", true)
	}

	tools := ToolsForShell(shell)

	// First check if we need to install anything
	needsInstall := false
	for _, tool := range tools {
		if cfg.IsEnabled(tool.Name) {
			if !isBinaryAvailable(tool) {
				needsInstall = true
				break
			}
		}
	}

	if !needsInstall {
		return
	}

	if runtime.GOOS == "windows" {
		installToolsWindows(cfg)
		return
	}

	if env.IsAlpine() {
		installToolsAlpine(tools, cfg)
		return
	}

	// Ensure Homebrew is available
	if err := ensureHomebrew(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Skipping tool installation: %v", err)))
		return
	}

	for _, tool := range tools {
		if cfg.IsEnabled(tool.Name) {
			if err := ensureTool(tool.Binary, tool.GetBrewPkg()); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install %s: %v", tool.GetBrewPkg(), err)))
			}
		}
	}
}

func ensureHomebrew() error {
	if _, err := exec.LookPath("brew"); err == nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		return fmt.Errorf("homebrew not found on Windows; install shell tools with winget")
	}

	commonPaths := []string{"/home/linuxbrew/.linuxbrew/bin/brew", "/opt/homebrew/bin/brew", "/usr/local/bin/brew"}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			path := os.Getenv("PATH")
			if err := os.Setenv("PATH", path+string(os.PathListSeparator)+filepath.Dir(p)); err != nil {
				return fmt.Errorf("failed to update PATH: %w", err)
			}
			return nil
		}
	}

	fmt.Println(infoStyle.Render("Homebrew is missing. It is required to install enabled components."))
	var install bool
	err := huh.NewConfirm().
		Title("Would you like to install Homebrew?").
		Value(&install).
		Run()
	if err != nil {
		return err
	}

	if !install {
		return fmt.Errorf("homebrew installation declined")
	}

	fmt.Println(infoStyle.Render("⬇️  Installing Homebrew..."))

	cmd := exec.Command("/bin/bash", "-c", "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install homebrew: %w", err)
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			path := os.Getenv("PATH")
			if err := os.Setenv("PATH", path+string(os.PathListSeparator)+filepath.Dir(p)); err != nil {
				return fmt.Errorf("failed to update PATH: %w", err)
			}
			fmt.Println(successStyle.Render("✓ Homebrew installed and added to PATH for this session."))
			return nil
		}
	}

	return fmt.Errorf("homebrew installed but not found in expected locations")
}

// EnsureInstalled installs a single tool (matched by Binary in Tools) via
// the best available manager on this platform, if it isn't already on
// PATH. Used by `doctor --fix`, which needs a one-off install outside the
// InstallTools bulk flow.
func EnsureInstalled(binary string) error {
	if _, err := exec.LookPath(binary); err == nil {
		return nil
	}
	var tool *Tool
	for i := range Tools {
		if Tools[i].Binary == binary {
			tool = &Tools[i]
			break
		}
	}
	if tool == nil {
		return fmt.Errorf("unknown tool %q", binary)
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("install %s from the Shell menu on Windows", binary)
	}
	if env.IsAlpine() {
		mgr := alpinePackageManager()
		if mgr == "" {
			return fmt.Errorf("neither coldbrew nor apk found")
		}
		return installAlpinePkg(mgr, tool.GetApkPkg())
	}
	return ensureTool(tool.Binary, tool.GetBrewPkg())
}

func ensureTool(binary, pkg string) error {
	// Check both PATH and the brew opt prefix (uutils tools use libexec/uubin).
	if checkBinary := func() bool {
		if _, err := exec.LookPath(binary); err == nil {
			return true
		}
		prefix := homebrewPrefix()
		if prefix == "" {
			return false
		}
		if _, err := os.Stat(filepath.Join(prefix, "opt", pkg, "libexec", "uubin", binary)); err == nil {
			return true
		}
		return false
	}(); checkBinary {
		return nil
	}

	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("brew not found")
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Installing %s via Homebrew...", pkg)))
	cmd := exec.Command("brew", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", pkg, err)
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s installed successfully!", pkg)))
	return nil
}

// homebrewPrefix resolves the Homebrew prefix using the same logic as the
// embedded shell.sh: respect HOMEBREW_PREFIX env var, probe common paths,
// and fall back to the Linuxbrew default. Returns "" when brew is absent.
func homebrewPrefix() string {
	if p := os.Getenv("HOMEBREW_PREFIX"); p != "" {
		return p
	}
	for _, p := range []string{"/opt/homebrew", "/usr/local", "/home/linuxbrew/.linuxbrew"} {
		if _, err := os.Stat(filepath.Join(p, "bin", "brew")); err == nil {
			return p
		}
	}
	return ""
}

// isBinaryAvailable reports whether tool.Binary can be found on PATH or,
// for uutils tools that Homebrew installs into a non-standard libexec
// directory, inside the brew opt prefix.
func isBinaryAvailable(tool Tool) bool {
	if _, err := exec.LookPath(tool.Binary); err == nil {
		return true
	}
	prefix := homebrewPrefix()
	if prefix == "" {
		return false
	}
	// uutils-coreutils / uutils-findutils / uutils-diffutils all follow
	// the same Homebrew layout: <prefix>/opt/<pkg>/libexec/uubin/<binary>
	optBin := filepath.Join(prefix, "opt", tool.GetBrewPkg(), "libexec", "uubin", tool.Binary)
	if _, err := os.Stat(optBin); err == nil {
		return true
	}
	return false
}

func CheckDependencies() map[string]bool {
	status := make(map[string]bool)

	for _, tool := range toolsForCurrentPlatform() {
		status[tool.Binary] = isBinaryAvailable(tool)
	}

	return status
}
