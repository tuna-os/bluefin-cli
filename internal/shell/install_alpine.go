package shell

import (
	"fmt"
	"os"
	"os/exec"
)

// EnsureColdbrew makes coldbrew available, bootstrapping it via apk if it's
// not already installed (coldbrew ships as a normal Alpine package, so this
// is a one-time `sudo apk add coldbrew`). Returns true once coldbrew is
// ready to use. This is the default path on Alpine-family systems: after
// the one-time bootstrap, every tool install is rootless and sandboxed
// instead of needing sudo per package.
func EnsureColdbrew() bool {
	if _, err := exec.LookPath("coldbrew"); err == nil {
		return true
	}
	if _, err := exec.LookPath("apk"); err != nil {
		return false
	}
	fmt.Println(infoStyle.Render("⬇️  Setting up coldbrew (rootless package manager)..."))
	cmd := exec.Command("sudo", "apk", "add", "coldbrew")
	cmd.Stdin = nil
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(errorStyle.Render("Could not set up coldbrew: " + err.Error()))
		return false
	}
	if _, err := exec.LookPath("coldbrew"); err != nil {
		return false
	}
	fmt.Println(successStyle.Render("✓ coldbrew ready"))
	return true
}

// alpinePackageManager picks the best available installer on Alpine-family
// systems, defaulting to coldbrew (rootless, sandboxed via bubblewrap — the
// path postmarketOS's Duranium variant is standardizing on), bootstrapping
// it on first use when possible. Falls back to apk directly (needs root;
// passwordless sudo is assumed since interactive password prompts have no
// TTY to answer them from here) only if coldbrew can't be set up.
func alpinePackageManager() string {
	if EnsureColdbrew() {
		return "coldbrew"
	}
	if _, err := exec.LookPath("apk"); err == nil {
		return "apk"
	}
	return ""
}

func installToolsAlpine(tools []Tool, cfg *Config) {
	mgr := alpinePackageManager()
	if mgr == "" {
		fmt.Println(errorStyle.Render("Skipping tool installation: neither coldbrew nor apk found"))
		return
	}
	for _, tool := range tools {
		if !cfg.IsEnabled(tool.Name) {
			continue
		}
		if _, err := exec.LookPath(tool.Binary); err == nil {
			continue
		}
		pkg := tool.GetApkPkg()
		fmt.Println(infoStyle.Render(fmt.Sprintf("⬇️  Installing %s via %s...", pkg, mgr)))
		if err := installAlpinePkg(mgr, pkg); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: Failed to install %s: %v", pkg, err)))
			continue
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s installed successfully!", pkg)))
	}
}

func installAlpinePkg(mgr, pkg string) error {
	switch mgr {
	case "coldbrew":
		install := exec.Command("coldbrew", "install", pkg)
		install.Stdout, install.Stderr = os.Stdout, os.Stderr
		if err := install.Run(); err != nil {
			return err
		}
		// wrap creates a real PATH-visible shim; without it the binary is
		// only reachable via `coldbrew run <pkg>`, and our tool-detection
		// (exec.LookPath) would never see it as installed.
		wrap := exec.Command("coldbrew", "wrap", pkg)
		wrap.Stdout, wrap.Stderr = os.Stdout, os.Stderr
		return wrap.Run()
	case "apk":
		cmd := exec.Command("sudo", "apk", "add", pkg)
		cmd.Stdin = nil
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%w (needs passwordless sudo for apk; run manually with a password if this fails)", err)
		}
		return nil
	}
	return fmt.Errorf("unknown package manager %q", mgr)
}
