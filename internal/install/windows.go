package install

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
)

type brewfilePackage struct {
	kind string
	name string
}

type windowsManager string

const (
	managerWinget windowsManager = "winget"
)

var (
	windowsManagerPriority = []windowsManager{managerWinget}
	windowsLookPath        = exec.LookPath
	windowsRunCommand      = func(cmd *exec.Cmd) error {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	windowsPromptConfirm = func(title, description string) (bool, error) {
		var confirm bool
		err := huh.NewConfirm().
			Title(title).
			Description(description).
			Value(&confirm).
			Run()
		return confirm, err
	}
	windowsBootstrapManagersOnce sync.Once
	windowsBootstrapManagers     []string
	windowsBootstrapErr          error
	brewDeclLine                 = regexp.MustCompile(`^\s*(brew|cask)\s+["']([^"']+)["']`)
)

type packageResolver interface {
	Candidates(name string) []string
}

type packageExecutor interface {
	AvailableManagers() []string
	Install(manager, candidate string) error
}

type windowsResolver struct{}

func (windowsResolver) Candidates(name string) []string {
	return windowsCandidates(name)
}

type windowsExecutor struct{}

func (windowsExecutor) AvailableManagers() []string {
	return AvailableWindowsManagers()
}

func (windowsExecutor) Install(manager, candidate string) error {
	return tryInstallWithManager(manager, candidate)
}

func BundleWindows(nameOrPath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("windows bundle flow is only available on Windows")
	}

	if strings.Contains(nameOrPath, "/") || strings.Contains(nameOrPath, "\\") {
		return fmt.Errorf("custom Brewfiles are not supported on Windows; use bundle names or interactive install")
	}

	pkgs, err := WindowsPackagesForBundles([]string{nameOrPath})
	if err != nil {
		return err
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages configured for bundle %s", nameOrPath)
	}

	return InstallWindowsPackages(pkgs)
}

func WindowsPackagesForBundles(bundleNames []string) ([]WindowsPackage, error) {
	manifest := getWindowsBundleManifest()
	if len(bundleNames) == 0 {
		return nil, fmt.Errorf("no bundles selected")
	}

	seen := map[string]bool{}
	packages := make([]WindowsPackage, 0)
	for _, bundleName := range bundleNames {
		bundleName = strings.TrimSpace(bundleName)
		if bundleName == "" {
			continue
		}

		bundle, ok := manifest[bundleName]
		if !ok {
			return nil, fmt.Errorf("unknown Windows bundle: %s", bundleName)
		}

		for _, pkg := range bundle.Packages {
			id := strings.TrimSpace(pkg.ID)
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

func InstallWindowsPackages(pkgs []WindowsPackage) error {
	if len(pkgs) == 0 {
		return fmt.Errorf("no Windows packages selected")
	}

	availableManagers, err := ensureWindowsManagersBootstrap(AvailableWindowsManagers)
	if err != nil {
		return err
	}
	if len(availableManagers) == 0 {
		return fmt.Errorf("winget is required on Windows and was not found")
	}

	fmt.Println(infoStyle.Render("📦 Installing selected Windows packages..."))
	fmt.Println(infoStyle.Render("Manager: winget"))

	var unmatched []string
	for _, pkg := range pkgs {
		if err := installWindowsManifestPackage(availableManagers, pkg); err != nil {
			if strings.TrimSpace(pkg.Name) != "" {
				unmatched = append(unmatched, fmt.Sprintf("%s (%s)", pkg.Name, pkg.ID))
			} else {
				unmatched = append(unmatched, pkg.ID)
			}
		}
	}

	if len(unmatched) > 0 {
		unavailablePath, writeErr := writeWingetUnavailableList(unmatched)
		fmt.Println()
		fmt.Println(errorStyle.Render("⚠ Some packages are unavailable in winget and were skipped:"))
		for _, name := range unmatched {
			fmt.Printf("  - %s\n", name)
		}
		if unavailablePath != "" {
			fmt.Println(infoStyle.Render("Unavailable list saved to: " + unavailablePath))
		}
		if writeErr != nil {
			fmt.Println(errorStyle.Render("Warning: could not write unavailable list: " + writeErr.Error()))
		}
	}

	fmt.Println(successStyle.Render("✓ Windows package installation complete"))
	return nil
}

func installWindowsManifestPackage(availableManagers []string, pkg WindowsPackage) error {
	candidates := sanitizeWindowsCandidates(append([]string{pkg.ID}, pkg.Aliases...))
	if len(candidates) == 0 {
		return fmt.Errorf("no viable package id for %s", pkg.Name)
	}

	for _, manager := range availableManagers {
		for _, candidate := range candidatesForManager(manager, candidates) {
			if tryInstallWithManager(manager, candidate) == nil {
				label := pkg.Name
				if strings.TrimSpace(label) == "" {
					label = pkg.ID
				}
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s (%s)", label, manager)))
				return nil
			}
		}
	}

	return fmt.Errorf("no package match for %s", pkg.ID)
}

func AvailableWindowsManagers() []string {
	available := make([]string, 0, len(windowsManagerPriority))
	for _, manager := range windowsManagerPriority {
		if isWindowsManagerAvailable(manager) {
			available = append(available, string(manager))
		}
	}
	return available
}

func isWindowsManagerAvailable(manager windowsManager) bool {
	bin := string(manager)
	if manager == managerWinget {
		bin = "winget"
	}

	_, err := windowsLookPath(bin)
	return err == nil
}

func installWindowsPackage(availableManagers []string, pkg brewfilePackage, resolver packageResolver, executor packageExecutor) error {
	if isUnsupportedWindowsPackage(pkg.name) {
		return fmt.Errorf("unsupported package for Windows: %s", pkg.name)
	}

	candidates := sanitizeWindowsCandidates(resolver.Candidates(pkg.name))
	if len(candidates) == 0 {
		return fmt.Errorf("no viable Windows candidates for %s", pkg.name)
	}

	for _, manager := range availableManagers {
		for _, candidate := range candidatesForManager(manager, candidates) {
			if executor.Install(manager, candidate) == nil {
				fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s (%s)", pkg.name, manager)))
				return nil
			}
		}
	}
	return fmt.Errorf("no package match for %s", pkg.name)
}

func tryInstallWithManager(manager, candidate string) error {
	switch manager {
	case string(managerWinget):
		cmd := exec.Command("winget", "install", "--id", candidate, "--exact", "--source", "winget", "--accept-source-agreements", "--accept-package-agreements", "--silent")
		if err := windowsRunCommand(cmd); err == nil {
			return nil
		}

		fallback := exec.Command("winget", "install", "--name", candidate, "--source", "winget", "--accept-source-agreements", "--accept-package-agreements", "--silent")
		return windowsRunCommand(fallback)
	default:
		return fmt.Errorf("unsupported manager: %s", manager)
	}
}

func ensureWindowsManagersBootstrap(getAvailable func() []string) ([]string, error) {
	windowsBootstrapManagersOnce.Do(func() {
		available := getAvailable()
		if len(available) > 0 {
			windowsBootstrapManagers = available
			return
		}

		fmt.Println(infoStyle.Render("No Windows package manager detected. Bluefin CLI can help bootstrap one now."))

		for _, manager := range windowsManagerPriority {
			installNow, err := promptInstallWindowsManager(manager)
			if err != nil {
				windowsBootstrapErr = err
				return
			}

			if !installNow {
				continue
			}

			if err := bootstrapWindowsManager(manager); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Failed to install %s: %v", manager, err)))
			}

			available = getAvailable()
			if len(available) > 0 {
				windowsBootstrapManagers = available
				return
			}
		}

		windowsBootstrapManagers = getAvailable()
	})

	return windowsBootstrapManagers, windowsBootstrapErr
}

func promptInstallWindowsManager(manager windowsManager) (bool, error) {
	switch manager {
	case managerWinget:
		return windowsPromptConfirm(
			"winget is not installed. Install it now?",
			"Opens the Microsoft App Installer page.",
		)
	default:
		return false, fmt.Errorf("unsupported manager: %s", manager)
	}
}

func bootstrapWindowsManager(manager windowsManager) error {
	switch manager {
	case managerWinget:
		fmt.Println(infoStyle.Render("Opening Microsoft App Installer page for winget..."))
		fmt.Println(infoStyle.Render("If it does not open automatically, install from: https://apps.microsoft.com/detail/9NBLGGH4NNS1"))
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://apps.microsoft.com/detail/9NBLGGH4NNS1")
		if err := windowsRunCommand(cmd); err != nil {
			return fmt.Errorf("could not open App Installer page: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported manager: %s", manager)
	}
}

func isUnsupportedWindowsPackage(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return true
	}

	if strings.Contains(normalized, "/") {
		return true
	}

	if strings.Contains(normalized, "linux") {
		return true
	}

	return false
}

func sanitizeWindowsCandidates(candidates []string) []string {
	seen := map[string]bool{}
	cleaned := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || strings.Contains(candidate, "/") {
			continue
		}
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		cleaned = append(cleaned, candidate)
	}
	return cleaned
}

func candidatesForManager(manager string, candidates []string) []string {
	filtered := make([]string, 0, len(candidates))
	filtered = append(filtered, candidates...)

	if manager != string(managerWinget) {
		return filtered
	}

	ids := make([]string, 0, len(filtered))
	names := make([]string, 0, len(filtered))
	for _, candidate := range filtered {
		if strings.Contains(candidate, ".") {
			ids = append(ids, candidate)
			continue
		}
		names = append(names, candidate)
	}

	return append(ids, names...)
}

func writeWingetUnavailableList(unmatched []string) (string, error) {
	if len(unmatched) == 0 {
		return "", nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(configDir, "BluefinCLI")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, "winget-unavailable.txt")
	existing := map[string]bool{}

	if content, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			name := strings.TrimSpace(line)
			if name != "" {
				existing[name] = true
			}
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	for _, name := range unmatched {
		name = strings.TrimSpace(name)
		if name != "" {
			existing[name] = true
		}
	}

	all := make([]string, 0, len(existing))
	for name := range existing {
		all = append(all, name)
	}
	sort.Strings(all)

	content := strings.Join(all, "\n")
	if content != "" {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func parseBrewfilePackages(path string) ([]brewfilePackage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var packages []brewfilePackage
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := brewDeclLine.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		packages = append(packages, brewfilePackage{
			kind: matches[1],
			name: strings.TrimSpace(matches[2]),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return packages, nil
}
