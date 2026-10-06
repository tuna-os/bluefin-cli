package sunset

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Swappable for tests.
var (
	lookPath  = exec.LookPath
	statPath  = os.Stat
	getenvVar = os.Getenv
)

// FindWindowsCLI locates the Windows build of this CLI from within WSL:
// first on PATH, then at the WinGet Links location under %LOCALAPPDATA%.
// It returns "" if neither is found.
func FindWindowsCLI() string {
	if path, err := lookPath("bluefin-cli.exe"); err == nil {
		return path
	}

	localAppData := strings.TrimSpace(getenvVar("LOCALAPPDATA"))
	if localAppData != "" {
		candidate := filepath.Join(localAppData, "Microsoft", "WinGet", "Links", "bluefin-cli.exe")
		if _, err := statPath(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// RunnerIO bundles the standard streams a delegated process should inherit.
type RunnerIO struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// DelegateToWindowsCLI re-execs the Windows build of this CLI (located by
// FindWindowsCLI) with args, wiring streams's streams straight through. It
// returns an error describing how to install the Windows build if no
// Windows executable can be found, and the exec error otherwise.
func DelegateToWindowsCLI(args []string, streams RunnerIO, r Reporter) error {
	if r == nil {
		r = PrintReporter{}
	}

	winExe := FindWindowsCLI()
	if winExe == "" {
		r.Infof("WSL detected. Solar theme switching requires the Windows version of Bluefin CLI.")
		r.Infof("Please install bluefin-cli.exe in Windows and ensure it is in your Windows PATH.")
		r.Infof("You can download it from the GitHub releases page.")
		return nil
	}

	r.Infof("Delegating sunset command to Windows CLI: %s", winExe)

	command := exec.Command(winExe, args...)
	command.Stdout = streams.Stdout
	command.Stderr = streams.Stderr
	command.Stdin = streams.Stdin
	if err := command.Run(); err != nil {
		return fmt.Errorf("running Windows CLI: %w", err)
	}
	return nil
}
