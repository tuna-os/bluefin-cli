package install

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// installBundleAlpine is the Alpine/musl code path for InstallBundle.
// It parses the Brewfile, skips taps/casks/flatpaks/vscode (which have no
// apk equivalent), and installs brew formulae via coldbrew (rootless,
// sandboxed) or apk (needs sudo). A summary is always printed so the user
// knows which packages were installed and which were skipped.
func installBundleAlpine(packages ...string) error {
	mgr := alpinePackageManager()
	if mgr == "" {
		return fmt.Errorf("bundle installation requires coldbrew or apk, but neither was found — please install coldbrew first: sudo apk add coldbrew")
	}

	var brewfiles []string
	var cleanups []func()
	defer func() {
		for _, c := range cleanups {
			c()
		}
	}()

	for _, pkg := range packages {
		path, cleanup, err := GetBrewfile(pkg)
		if err != nil {
			return err
		}
		cleanups = append(cleanups, cleanup)
		brewfiles = append(brewfiles, path)
	}

	merged, cleanup, err := MergeBrewfiles(brewfiles)
	if err != nil {
		return err
	}
	defer cleanup()

	pkgs := GetBrewfilePackages(merged)
	if len(pkgs) == 0 {
		return fmt.Errorf("no installable packages found in the bundle")
	}

	var installed, skipped, failed []string

	for _, p := range pkgs {
		switch p.Kind {
		case "brew":
			// Brew formulae may have Alpine equivalents; try to install.
			apkName := p.Name
			fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Installing %s via %s...", apkName, mgr)))
			if err := installAlpinePkg(mgr, apkName); err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ %s: not available via %s", apkName, mgr)))
				failed = append(failed, apkName)
			} else {
				fmt.Println(successStyle.Render(fmt.Sprintf("  ✓ %s installed", apkName)))
				installed = append(installed, apkName)
			}
		case "cask":
			skipped = append(skipped, fmt.Sprintf("%s (cask)", p.Name))
		case "winget", "scoop", "choco":
			skipped = append(skipped, fmt.Sprintf("%s (%s)", p.Name, p.Kind))
		default:
			skipped = append(skipped, fmt.Sprintf("%s (%s)", p.Name, p.Kind))
		}
	}

	// Print summary
	fmt.Println()
	if len(installed) > 0 {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Installed %d package(s): %s", len(installed), strings.Join(installed, ", "))))
	}
	if len(skipped) > 0 {
		fmt.Println(infoStyle.Render(fmt.Sprintf("⊘ Skipped %d package(s) (no apk equivalent): %s", len(skipped), strings.Join(skipped, ", "))))
	}
	if len(failed) > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %d package(s) not found in Alpine repos: %s", len(failed), strings.Join(failed, ", "))))
		fmt.Println(infoStyle.Render("These may need to be installed manually or from source."))
	}

	if len(installed) == 0 && len(skipped) > 0 {
		return fmt.Errorf("no packages could be installed from this bundle on Alpine — all %d packages were skipped", len(skipped))
	}

	return nil
}

// alpinePackageManager picks the best available installer on Alpine-family
// systems: coldbrew (rootless, sandboxed) when available, falling back to
// apk (needs sudo). Returns "" if neither is available.
func alpinePackageManager() string {
	if _, err := exec.LookPath("coldbrew"); err == nil {
		return "coldbrew"
	}
	if _, err := exec.LookPath("apk"); err == nil {
		return "apk"
	}
	return ""
}

// installAlpinePkg installs a single package via coldbrew or apk.
func installAlpinePkg(mgr, pkg string) error {
	switch mgr {
	case "coldbrew":
		install := exec.Command("coldbrew", "install", pkg)
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr
		if err := install.Run(); err != nil {
			return err
		}
		wrap := exec.Command("coldbrew", "wrap", pkg)
		wrap.Stdout = os.Stdout
		wrap.Stderr = os.Stderr
		return wrap.Run()
	case "apk":
		cmd := exec.Command("sudo", "apk", "add", pkg)
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%w (needs passwordless sudo for apk; run manually with a password if this fails)", err)
		}
		return nil
	}
	return fmt.Errorf("unknown package manager %q", mgr)
}
