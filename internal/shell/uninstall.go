package shell

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type UninstallOptions struct {
	Shells         []string
	RemoveSoftware bool
	RemoveModules  bool
	RemoveConfig   bool
}

func UninstallSetup(opts UninstallOptions) error {
	shells := opts.Shells
	if len(shells) == 0 {
		shells = []string{"powershell", "bash", "zsh", "fish"}
	}

	var failures []string

	for _, shellName := range shells {
		if err := Toggle(shellName, false); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unsupported shell") {
				continue
			}
			failures = append(failures, fmt.Sprintf("failed to disable %s: %v", shellName, err))
		}
	}

	if opts.RemoveSoftware {
		if err := uninstallManagedSoftware(); err != nil {
			failures = append(failures, err.Error())
		}
	}

	if opts.RemoveModules && runtime.GOOS == "windows" {
		if err := uninstallPowerShellModules(); err != nil {
			failures = append(failures, err.Error())
		}
	}

	if opts.RemoveConfig {
		configPath, err := getConfigPath()
		if err == nil {
			if removeErr := os.Remove(configPath); removeErr != nil && !os.IsNotExist(removeErr) {
				failures = append(failures, fmt.Sprintf("failed to remove config %s: %v", configPath, removeErr))
			}
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}

	fmt.Println(successStyle.Render("✓ Bluefin shell setup uninstalled"))
	return nil
}

func uninstallManagedSoftware() error {
	if runtime.GOOS == "windows" {
		return uninstallManagedWindowsSoftware()
	}

	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Println(infoStyle.Render("Homebrew not found; skipping managed package uninstall."))
		return nil
	}

	toolPkgs := dedupePackages(append(toolPackagesForCurrentPlatform(), "glow"))
	for _, pkg := range toolPkgs {
		cmd := exec.Command("brew", "uninstall", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}

	return nil
}

func uninstallManagedWindowsSoftware() error {
	if _, err := exec.LookPath("winget"); err != nil {
		fmt.Println(infoStyle.Render("winget not found; skipping managed package uninstall."))
		return nil
	}

	packages := dedupePackages(append(toolPackagesForCurrentPlatform(), "charmbracelet.glow"))
	for _, pkg := range packages {
		if pkg == "" {
			continue
		}

		cmd := exec.Command("winget", "uninstall", "--id", pkg, "--exact", "--source", "winget", "--accept-source-agreements", "--silent")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			continue
		}

		fallback := exec.Command("winget", "uninstall", "--name", pkg, "--source", "winget", "--accept-source-agreements", "--silent")
		fallback.Stdout = os.Stdout
		fallback.Stderr = os.Stderr
		_ = fallback.Run()
	}

	return nil
}

func uninstallPowerShellModules() error {
	powerShellExe := windowsPowerShellExe()
	modules := []string{"PSFzf", "Terminal-Icons"}

	for _, moduleName := range modules {
		script := fmt.Sprintf("$ErrorActionPreference='SilentlyContinue'; Uninstall-Module -Name '%s' -AllVersions -Force", moduleName)
		cmd := exec.Command(powerShellExe, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}

	return nil
}

func toolPackagesForCurrentPlatform() []string {
	packages := make([]string, 0)
	for _, tool := range toolsForCurrentPlatform() {
		packages = append(packages, strings.TrimSpace(tool.Pkg))
	}
	return packages
}

func dedupePackages(input []string) []string {
	seen := map[string]bool{}
	output := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		output = append(output, item)
	}
	return output
}
