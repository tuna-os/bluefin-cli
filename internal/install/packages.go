package install

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Package is a cross-platform representation of a single installable package.
type Package struct {
	Name        string // display name
	ID          string // brew cask/formula name, or winget ID
	Kind        string // "brew", "cask", or "winget"
	Installed   bool
	Description string
}

// GetBundlePackages returns the packages for a named bundle.
// On Windows it reads from the Windows manifest; on Unix it parses the embedded Brewfile.
func GetBundlePackages(bundleName string) ([]Package, error) {
	if runtime.GOOS == "windows" {
		return getWindowsBundlePackages(bundleName)
	}
	return getUnixBundlePackages(bundleName)
}

// GetBrewfilePackages parses any Brewfile on disk into packages (brew and
// cask lines), for user Brewfiles and custom bundles.
func GetBrewfilePackages(path string) ([]Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseBrewfileBytes(data), nil
}

func getUnixBundlePackages(bundleName string) ([]Package, error) {
	bundle, ok := bundles[bundleName]
	if !ok {
		return nil, fmt.Errorf("unknown bundle: %s", bundleName)
	}
	data, err := EmbeddedBrewfiles.ReadFile("resources/brewfiles/" + bundle.File)
	if err != nil {
		return nil, fmt.Errorf("could not read bundle %s: %w", bundleName, err)
	}
	return parseBrewfileBytes(data), nil
}

func getWindowsBundlePackages(bundleName string) ([]Package, error) {
	manifest := getWindowsBundleManifest()
	bundle, ok := manifest[bundleName]
	if !ok {
		return nil, fmt.Errorf("unknown Windows bundle: %s", bundleName)
	}
	pkgs := make([]Package, 0, len(bundle.Packages))
	for _, wp := range bundle.Packages {
		name := wp.Name
		if name == "" {
			name = wp.ID
		}
		pkgs = append(pkgs, Package{
			Name:        name,
			ID:          wp.ID,
			Kind:        "winget",
			Description: wp.Description,
		})
	}
	return pkgs, nil
}

// parseBrewfileBytes parses brew/cask lines from a Brewfile, skipping taps and flatpaks.
func parseBrewfileBytes(data []byte) []Package {
	var pkgs []Package
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "tap ") || strings.HasPrefix(line, "flatpak ") {
			continue
		}
		var kind string
		switch {
		case strings.HasPrefix(line, "brew "):
			kind = "brew"
		case strings.HasPrefix(line, "cask "):
			kind = "cask"
		case strings.HasPrefix(line, "winget "):
			kind = "winget"
		case strings.HasPrefix(line, "scoop "):
			kind = "scoop"
		case strings.HasPrefix(line, "choco "):
			kind = "choco"
		default:
			continue
		}
		name := extractQuotedName(line)
		if name == "" {
			continue
		}
		// Use the last path segment as the display ID (e.g. "tap/pkg" → "pkg")
		displayID := name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			displayID = name[idx+1:]
		}
		pkgs = append(pkgs, Package{Name: displayID, ID: name, Kind: kind})
	}
	return pkgs
}

func extractQuotedName(line string) string {
	start := strings.Index(line, `"`)
	if start < 0 {
		return ""
	}
	end := strings.Index(line[start+1:], `"`)
	if end < 0 {
		return ""
	}
	return line[start+1 : start+1+end]
}

// MarkInstalled annotates packages with their installed status.
func MarkInstalled(pkgs []Package) []Package {
	if runtime.GOOS == "windows" {
		return markInstalledWinget(pkgs)
	}
	return markInstalledBrew(pkgs)
}

// InstalledCasks returns the set of installed Homebrew casks (short names);
// empty on error or when brew is absent.
func InstalledCasks() map[string]bool {
	installed := make(map[string]bool)
	out, err := exec.Command("brew", "list", "--cask").Output()
	if err != nil {
		return installed
	}
	for _, name := range strings.Fields(string(out)) {
		installed[strings.TrimSpace(name)] = true
	}
	return installed
}

func markInstalledBrew(pkgs []Package) []Package {
	installed := make(map[string]bool)
	for _, args := range [][]string{{"list", "--formula"}, {"list", "--cask"}} {
		out, err := exec.Command("brew", args...).Output()
		if err != nil {
			continue
		}
		for _, name := range strings.Fields(string(out)) {
			installed[strings.TrimSpace(name)] = true
		}
	}
	result := make([]Package, len(pkgs))
	for i, p := range pkgs {
		// Match on the short name (last segment) since brew list returns short names
		shortName := p.Name
		p.Installed = installed[shortName]
		result[i] = p
	}
	return result
}

func markInstalledWinget(pkgs []Package) []Package {
	installed := make(map[string]bool)
	out, err := exec.Command("winget", "list", "--source", "winget", "--accept-source-agreements").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			for _, f := range fields {
				installed[strings.ToLower(strings.TrimSpace(f))] = true
			}
		}
	}
	result := make([]Package, len(pkgs))
	for i, p := range pkgs {
		p.Installed = installed[strings.ToLower(p.ID)]
		result[i] = p
	}
	return result
}

// GetInstalledBrewPackages returns the names of all currently installed brew formulae and casks.
func GetInstalledBrewPackages() ([]string, error) {
	var all []string
	for _, args := range [][]string{{"list", "--formula"}, {"list", "--cask"}} {
		out, err := exec.Command("brew", args...).Output()
		if err != nil {
			continue
		}
		for _, name := range strings.Fields(string(out)) {
			name = strings.TrimSpace(name)
			if name != "" {
				all = append(all, name)
			}
		}
	}
	return all, nil
}

// UninstallBrewPackages removes brew formulae/casks by name.
func UninstallBrewPackages(names []string) error {
	if len(names) == 0 {
		return nil
	}
	args := append([]string{"uninstall"}, names...)
	cmd := exec.Command("brew", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// UninstallWingetPackages removes winget packages by ID in a single command.
func UninstallWingetPackages(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		cmd := exec.Command("winget", "uninstall", "--id", id, "--exact", "--accept-source-agreements", "--silent")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall %s: %v\n", id, err)
		}
	}
	return nil
}

// InstallBrewPackages installs a list of brew formulae/casks.
// Formulae and casks are batched into separate single commands.
func InstallBrewPackages(pkgs []Package) error {
	var formulae, casks []string
	for _, p := range pkgs {
		if p.Kind == "cask" {
			casks = append(casks, p.ID)
		} else {
			formulae = append(formulae, p.ID)
		}
	}
	if len(formulae) > 0 {
		args := append([]string{"install"}, formulae...)
		cmd := exec.Command("brew", args...)
		cmd.Env = append(os.Environ(), "HOMEBREW_NO_ENV_HINTS=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew install failed: %w", err)
		}
	}
	if len(casks) > 0 {
		args := append([]string{"install", "--cask"}, casks...)
		cmd := exec.Command("brew", args...)
		cmd.Env = append(os.Environ(), "HOMEBREW_NO_ENV_HINTS=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew install --cask failed: %w", err)
		}
	}
	return nil
}
