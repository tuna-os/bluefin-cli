package shell

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func toolsForCurrentPlatform() []Tool {
	if runtime.GOOS == "windows" {
		return ToolsForShell("powershell")
	}

	return Tools
}

// InstallTools iterates through the config and installs enabled tools
func InstallTools(cfg *Config) {
	// If MOTD is enabled, ensure Glow is also considered enabled for installation
	if cfg.IsEnabled("Motd") {
		cfg.SetEnabled("Glow", true)
	}

	tools := toolsForCurrentPlatform()

	// First check if we need to install anything
	needsInstall := false
	for _, tool := range tools {
		if cfg.IsEnabled(tool.Name) {
			if _, err := exec.LookPath(tool.Binary); err != nil {
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

func installToolsWindows(cfg *Config) {
	if err := ensurePowerShellModules(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to ensure PowerShell modules: %v", err)))
	}

	availableManagers := availableWindowsManagers()
	if len(availableManagers) == 0 {
		fmt.Println(errorStyle.Render("Skipping tool installation: winget not found"))
		return
	}

	fmt.Println(infoStyle.Render("Installing enabled components using winget."))

	tools := ToolsForShell("powershell")

	for _, tool := range tools {
		if strings.EqualFold(tool.Name, "Gsudo") && cfg.IsEnabled(tool.Name) {
			if err := ensureWindowsTool(tool.Binary, tool.Pkg, availableManagers); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install %s: %v", tool.Pkg, err)))
			}
			break
		}
	}

	if err := primeGsudoCache(); err != nil {
		fmt.Println(infoStyle.Render("Proceeding without gsudo elevation cache: " + err.Error()))
	}

	for _, tool := range tools {
		if strings.EqualFold(tool.Name, "Gsudo") {
			continue
		}

		if cfg.IsEnabled(tool.Name) {
			if err := ensureWindowsTool(tool.Binary, tool.Pkg, availableManagers); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install %s: %v", tool.Pkg, err)))
			}
		}
	}

	if cfg.IsEnabled("Motd") {
		if err := ensureWindowsTool("glow", "charmbracelet.glow", availableManagers); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install glow: %v", err)))
		}
	}
}

func ensurePowerShellModules() error {
	modules := []string{"PSReadLine", "Terminal-Icons", "PSFzf"}

	var failures []string
	for _, moduleName := range modules {
		if err := ensurePowerShellModule(moduleName); err != nil {
			failures = append(failures, fmt.Sprintf("%s (%v)", moduleName, err))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, ", "))
	}

	return nil
}

func ensurePowerShellModule(moduleName string) error {
	powerShellExe := windowsPowerShellExe()
	modulePathReset := `$env:PSModulePath = @("$env:USERPROFILE\Documents\WindowsPowerShell\Modules", "$env:ProgramFiles\WindowsPowerShell\Modules", "$env:WINDIR\System32\WindowsPowerShell\v1.0\Modules") -join ';'`

	checkScript := fmt.Sprintf("%s; if (Get-Module -ListAvailable -Name '%s' -ErrorAction SilentlyContinue) { exit 0 } else { exit 1 }", modulePathReset, moduleName)
	check := exec.Command(powerShellExe, "-NoProfile", "-NonInteractive", "-Command", checkScript)
	if err := check.Run(); err == nil {
		return nil
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Installing PowerShell module %s...", moduleName)))
	installScript := fmt.Sprintf("%s; $ErrorActionPreference='Stop'; Import-Module PackageManagement -ErrorAction Stop; Import-Module PowerShellGet -ErrorAction Stop; if (-not (Get-PackageProvider -Name NuGet -ListAvailable -ErrorAction SilentlyContinue)) { Install-PackageProvider -Name NuGet -MinimumVersion 2.8.5.201 -Scope CurrentUser -Force | Out-Null }; Set-PSRepository -Name PSGallery -InstallationPolicy Trusted -ErrorAction SilentlyContinue; Install-Module -Name '%s' -Scope CurrentUser -Repository PSGallery -Force -AllowClobber", modulePathReset, moduleName)
	install := exec.Command(powerShellExe, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", installScript)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("%w (try in Windows PowerShell: Install-Module -Name %s -Scope CurrentUser -Repository PSGallery -Force)", err, moduleName)
	}

	verify := exec.Command(powerShellExe, "-NoProfile", "-NonInteractive", "-Command", checkScript)
	if err := verify.Run(); err != nil {
		return fmt.Errorf("module %s install completed but module is still not discoverable", moduleName)
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ PowerShell module %s installed", moduleName)))
	return nil
}

func windowsPowerShellExe() string {
	if runtime.GOOS == "windows" {
		systemPath := `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`
		if _, err := os.Stat(systemPath); err == nil {
			return systemPath
		}
	}

	return "powershell.exe"
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
			os.Setenv("PATH", path+string(os.PathListSeparator)+filepath.Dir(p))
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
			os.Setenv("PATH", path+string(os.PathListSeparator)+filepath.Dir(p))
			fmt.Println(successStyle.Render("✓ Homebrew installed and added to PATH for this session."))
			return nil
		}
	}

	return fmt.Errorf("homebrew installed but not found in expected locations")
}

func ensureTool(binary, pkg string) error {
	if _, err := exec.LookPath(binary); err == nil {
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

func ensureWindowsTool(binary, pkg string, managers []string) error {
	if _, err := exec.LookPath(binary); err == nil {
		return nil
	}

	candidates := []string{pkg, binary}
	if pkg != strings.ToLower(pkg) {
		candidates = append(candidates, strings.ToLower(pkg))
	}

	seen := map[string]bool{}
	for _, manager := range managers {
		for _, candidate := range candidates {
			candidate = strings.TrimSpace(candidate)
			if candidate == "" {
				continue
			}

			key := manager + "::" + candidate
			if seen[key] {
				continue
			}
			seen[key] = true

			if err := tryInstallWithWindowsManager(manager, candidate); err == nil {
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s installed via %s", pkg, manager)))
				return nil
			}

			if _, err := exec.LookPath(binary); err == nil {
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s is already available", pkg)))
				return nil
			}
		}
	}

	return fmt.Errorf("no matching package found in available managers")
}

func availableWindowsManagers() []string {
	priority := []string{"winget"}
	available := make([]string, 0, len(priority))
	for _, manager := range priority {
		if _, err := exec.LookPath(manager); err == nil {
			available = append(available, manager)
		}
	}
	return available
}

func tryInstallWithWindowsManager(manager, candidate string) error {
	switch manager {
	case "winget":
		if wingetPackageInstalled(candidate) {
			return nil
		}

		if err := runWingetInstallWithOptionalGsudo("--id", candidate, "--exact", "--source", "winget", "--accept-source-agreements", "--accept-package-agreements", "--silent"); err == nil {
			return nil
		}

		if err := runWingetInstallWithOptionalGsudo("--name", candidate, "--source", "winget", "--accept-source-agreements", "--accept-package-agreements", "--silent"); err == nil {
			return nil
		}

		if wingetPackageInstalled(candidate) {
			return nil
		}

		return fmt.Errorf("winget install failed for %s", candidate)
	default:
		return fmt.Errorf("unsupported manager: %s", manager)
	}
}

func runWingetInstallWithOptionalGsudo(args ...string) error {
	wingetArgs := append([]string{"install"}, args...)

	wingetPath := resolveWindowsExecutable("winget")
	if wingetPath == "" {
		wingetPath = "winget"
	}

	if gsudoPath := resolveWindowsExecutable("gsudo"); gsudoPath != "" {
		commandArgs := append([]string{wingetPath}, wingetArgs...)
		cmd := exec.Command(gsudoPath, commandArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	cmd := exec.Command(wingetPath, wingetArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func primeGsudoCache() error {
	gsudoPath := resolveWindowsExecutable("gsudo")
	if gsudoPath == "" {
		return fmt.Errorf("gsudo not found")
	}

	cmd := exec.Command(gsudoPath, "cache", "on")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveWindowsExecutable(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}

	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData != "" {
		candidate := filepath.Join(localAppData, "Microsoft", "WinGet", "Links", name+".exe")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}

	return ""
}

func wingetPackageInstalled(candidate string) bool {
	cmd := exec.Command("winget", "list", "--id", candidate, "--exact", "--source", "winget", "--accept-source-agreements")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	text := strings.ToLower(string(out))
	if strings.Contains(text, "no installed package") {
		return false
	}

	return strings.Contains(text, strings.ToLower(candidate))
}

//go:embed resources/shell.sh
var shellShScript string

//go:embed resources/shell.fish
var shellFishScript string

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

	var configFile string
	var rcLine string

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	switch shell {
	case "bash":
		configFile = filepath.Join(home, ".bashrc")
		rcLine = fmt.Sprintf(`eval "$(bluefin-cli init bash)" %s`, shellMaker)
	case "zsh":
		configFile = filepath.Join(home, ".zshrc")
		rcLine = fmt.Sprintf(`eval "$(bluefin-cli init zsh)" %s`, shellMaker)
	case "fish":
		configFile = filepath.Join(home, ".config/fish/config.fish")
		rcLine = fmt.Sprintf(`bluefin-cli init fish | source %s`, shellMaker)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) && enable {
			// Create if doesn't exist and we are enabling
			// For fish, ensure dir exists
			if shell == "fish" {
				if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
					return err
				}
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
				InstallTools(cfg)
			}
			return nil
		}

		f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		prefix := "\n"
		if len(text) == 0 || strings.HasSuffix(text, "\n") {
			prefix = ""
		}

		if _, err := f.WriteString(prefix + rcLine + "\n"); err != nil {
			return err
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
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Disabled shell experience for %s", shell)))
	}

	if enable {
		if cfg, err := LoadConfig(shell); err == nil {
			InstallTools(cfg)
		}
	}

	return nil
}

func Init(shell string, config *Config) (string, error) {
	if config == nil {
		config = DefaultConfig(shell)
	}

	tools := ToolsForShell(shell)
	configUpdated := false

	// Synchronize configuration with installed tools in PATH
	for _, tool := range tools {
		isInstalled := false
		if _, err := exec.LookPath(tool.Binary); err == nil {
			isInstalled = true
		}

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
		return sb.String(), nil
	}

	for _, tool := range tools {
		enabled := config.IsEnabled(tool.Name)

		if shell == "fish" {
			fmt.Fprintf(&sb, "set -gx %s %d\n", tool.GetEnvVar(), boolToInt(enabled))
		} else {
			fmt.Fprintf(&sb, "export %s=%d\n", tool.GetEnvVar(), boolToInt(enabled))
		}
	}

	if shell == "fish" {
		fmt.Fprintf(&sb, "set -gx BLUEFIN_SHELL_ENABLE_MOTD %d\n", boolToInt(config.IsEnabled("Motd")))
		fmt.Fprintf(&sb, "set -gx BLING_SHELL %s\n", shell)
	} else {
		fmt.Fprintf(&sb, "export BLUEFIN_SHELL_ENABLE_MOTD=%d\n", boolToInt(config.IsEnabled("Motd")))
		fmt.Fprintf(&sb, "export BLING_SHELL=\"%s\"\n", shell)
	}

	sb.WriteString("\n")

	if shell == "fish" {
		sb.WriteString(shellFishScript)
	} else {
		sb.WriteString(shellShScript)
	}

	return sb.String(), nil
}

func CheckStatus() map[string]bool {
	status := make(map[string]bool)
	shells := []string{"bash", "zsh", "fish"}
	home, _ := os.UserHomeDir()

	for _, shell := range shells {
		var configFile string
		switch shell {
		case "bash":
			configFile = filepath.Join(home, ".bashrc")
		case "zsh":
			configFile = filepath.Join(home, ".zshrc")
		case "fish":
			configFile = filepath.Join(home, ".config/fish/config.fish")
		}

		content, err := os.ReadFile(configFile)
		if err != nil {
			status[shell] = false
			continue
		}

		status[shell] = strings.Contains(string(content), shellMaker) || strings.Contains(string(content), "# bluefin-cli bling")
	}

	status["powershell"] = checkPowerShellStatus()
	status["pwsh"] = status["powershell"]

	return status
}

func CheckDependencies() map[string]bool {
	status := make(map[string]bool)

	for _, tool := range toolsForCurrentPlatform() {
		_, err := exec.LookPath(tool.Binary)
		status[tool.Binary] = err == nil
	}

	return status
}

// GetInstalledShells returns a list of shells that are available in the PATH
func GetInstalledShells() []string {
	var installed []string
	shells := []string{"bash", "zsh", "fish"}

	for _, s := range shells {
		if _, err := exec.LookPath(s); err == nil {
			installed = append(installed, s)
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
