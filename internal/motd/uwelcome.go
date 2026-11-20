package motd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// uwelcome is the terminal greeting Bluefin and Universal Blue images ship
// (github.com/projectbluefin/uwelcome). Where it is installed it owns the
// greeting: bluefin-cli shows and toggles it instead of printing a second
// one, and its own MOTD is the fallback for systems without it (macOS, other
// distributions).
//
// uwelcome keeps its on/off state in a file, <user config dir>/uwelcome/disabled,
// which `uwelcome enable|disable|toggle` and its greeting hooks read.
var uwelcomeBinary = "uwelcome"

// Uwelcome reports whether uwelcome is installed.
func Uwelcome() bool {
	_, err := exec.LookPath(uwelcomeBinary)
	return err == nil
}

// UwelcomeEnabled reports whether uwelcome greets: its disabled file is absent.
func UwelcomeEnabled() bool {
	path, err := uwelcomeDisabledFile()
	if err != nil {
		return true
	}
	_, err = os.Stat(path)
	return errors.Is(err, os.ErrNotExist)
}

func uwelcomeDisabledFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "uwelcome", "disabled"), nil
}

// SetUwelcome turns the uwelcome greeting on or off through uwelcome itself.
// Its own confirmation is discarded: callers print theirs, and the menu runs
// this while the TUI owns the terminal.
func SetUwelcome(enable bool) error {
	action := "disable"
	if enable {
		action = "enable"
	}
	out, err := exec.Command(uwelcomeBinary, action).CombinedOutput()
	if err != nil {
		return fmt.Errorf("uwelcome %s: %w: %s", action, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// showUwelcome prints uwelcome's greeting. uwelcome prints nothing while it
// is disabled, so say why rather than show an empty screen.
func showUwelcome() error {
	if !UwelcomeEnabled() {
		fmt.Fprintln(os.Stderr, "The uwelcome greeting is off. Turn it on with: bluefin-cli motd toggle all on")
		return nil
	}
	cmd := exec.Command(uwelcomeBinary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
