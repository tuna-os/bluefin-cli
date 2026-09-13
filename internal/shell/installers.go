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

// brewFallbackPaths are the per-platform prefixes Homebrew installs into. A
// brew that is installed but not yet on PATH -- the state right after an
// install, and in any shell that has not re-sourced its profile -- is found
// here rather than by LookPath.
var brewFallbackPaths = []string{
	"/home/linuxbrew/.linuxbrew/bin/brew",
	"/opt/homebrew/bin/brew",
	"/usr/local/bin/brew",
}

// findBrew returns the directory holding brew, and whether it found one. A
// brew already on PATH reports an empty directory: there is nothing to add.
func findBrew() (dir string, found bool) {
	if _, err := exec.LookPath("brew"); err == nil {
		return "", true
	}
	for _, p := range brewFallbackPaths {
		if _, err := os.Stat(p); err == nil {
			return filepath.Dir(p), true
		}
	}
	return "", false
}

// HomebrewAvailable reports whether this machine has Homebrew, on PATH or in
// one of its install prefixes. Callers that are about to start work which
// needs brew use it to settle the question up front -- in particular the TUI,
// which has to decide whether to ask the user *before* it opens a runner that
// cannot take input (tuna-os/bluefin-cli#273).
func HomebrewAvailable() bool {
	_, found := findBrew()
	return found
}

// InstallHomebrew runs the upstream install script and puts the resulting brew
// on PATH for this process. It never prompts: when the TUI owns the terminal it
// runs the script with NONINTERACTIVE=1 and no stdin, because the script's own
// "press RETURN to continue" would hang exactly as a prompt here would.
func InstallHomebrew() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("homebrew is not available on Windows; install shell tools with winget")
	}

	fmt.Println(infoStyle.Render("⬇️  Installing Homebrew..."))

	cmd := exec.Command("/bin/bash", "-c", "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if env.TerminalOwned() {
		cmd.Env = append(os.Environ(), "NONINTERACTIVE=1")
	} else {
		cmd.Stdin = os.Stdin
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install homebrew: %w", err)
	}

	if err := addBrewToPath(); err != nil {
		return err
	}
	fmt.Println(successStyle.Render("✓ Homebrew installed and added to PATH for this session."))
	return nil
}

// addBrewToPath puts a just-installed brew on PATH for the rest of this
// process, so the tool installs that follow can use it without a restart.
func addBrewToPath() error {
	dir, found := findBrew()
	if !found {
		return fmt.Errorf("homebrew installed but not found in expected locations")
	}
	if dir == "" {
		return nil
	}
	if err := os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+dir); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}
	return nil
}

func ensureHomebrew() error {
	if dir, found := findBrew(); found {
		if dir == "" {
			return nil
		}
		return addBrewToPath()
	}

	if runtime.GOOS == "windows" {
		return fmt.Errorf("homebrew not found on Windows; install shell tools with winget")
	}

	fmt.Println(infoStyle.Render("Homebrew is missing. It is required to install enabled components."))

	// Inside a TUI runner the keyboard belongs to Bubble Tea, so a confirm here
	// would render nowhere and wait forever (tuna-os/bluefin-cli#273). The TUI
	// asks on the menu thread instead, before it opens the runner; reaching
	// this point means the user was already asked, or the caller is not the
	// TUI. Either way, say what to do rather than block.
	if env.TerminalOwned() {
		return fmt.Errorf("homebrew is not installed; install it, then run 'bluefin-cli doctor' to check the rest")
	}

	var install bool
	if err := huh.NewConfirm().
		Title("Would you like to install Homebrew?").
		Value(&install).
		Run(); err != nil {
		return err
	}
	if !install {
		return fmt.Errorf("homebrew installation declined")
	}

	return InstallHomebrew()
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
