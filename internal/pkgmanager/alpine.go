package pkgmanager

import (
	"fmt"
	"os"
	"os/exec"
)

// AlpineManager identifies a supported package manager on Alpine-family hosts.
type AlpineManager string

const (
	Coldbrew AlpineManager = "coldbrew"
	APK      AlpineManager = "apk"
)

// DetectAlpine prefers the rootless coldbrew backend and falls back to apk.
func DetectAlpine() AlpineManager {
	if _, err := exec.LookPath(string(Coldbrew)); err == nil {
		return Coldbrew
	}
	if _, err := exec.LookPath(string(APK)); err == nil {
		return APK
	}
	return ""
}

// InstallAlpine installs one package through the selected backend.
func InstallAlpine(manager AlpineManager, pkg string) error {
	switch manager {
	case Coldbrew:
		if err := run("coldbrew", "install", pkg); err != nil {
			return err
		}
		// A wrapped package is visible on PATH without `coldbrew run`.
		return run("coldbrew", "wrap", pkg)
	case APK:
		if err := run("sudo", "apk", "add", pkg); err != nil {
			return fmt.Errorf("%w (needs passwordless sudo for apk; run manually with a password if this fails)", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown package manager %q", manager)
	}
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
