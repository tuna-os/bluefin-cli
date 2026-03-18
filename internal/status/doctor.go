package status

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/shell"
	"github.com/hanthor/bluefin-cli/internal/tui"
)

// Check performs a health check of the environment
func Check() error {
	tui.RenderHeader("Bluefin CLI", "Doctor - Health Check")

	checks := []struct {
		name string
		fn   func() (bool, string)
	}{
		{"Homebrew", checkBrew},
		{"Shell Config", checkShellConfig},
		{"Development Container", checkContainer},
		{"Environment", checkEnv},
	}

	allPassed := true
	for _, c := range checks {
		pass, msg := c.fn()
		if pass {
			fmt.Printf("  ✓ %-20s %s\n", c.name, tui.SuccessStyle.Render(msg))
		} else {
			fmt.Printf("  ✗ %-20s %s\n", c.name, tui.ErrorStyle.Render(msg))
			allPassed = false
		}
	}

	fmt.Println()
	if allPassed {
		fmt.Println(tui.SuccessStyle.Render("Your environment is healthy! ✨"))
	} else {
		fmt.Println(tui.WarningStyle.Render("Some issues were found. Please check the suggestions above."))
	}

	return nil
}

func checkBrew() (bool, string) {
	if _, err := exec.LookPath("brew"); err != nil {
		return false, "Homebrew not found. Visit https://brew.sh to install."
	}

	cmd := exec.Command("brew", "doctor")
	if err := cmd.Run(); err != nil {
		return false, "Homebrew reports issues. Run 'brew doctor' for details."
	}

	return true, "Homebrew is healthy."
}

func checkShellConfig() (bool, string) {
	status := shell.CheckStatus()
	enabled := 0
	for _, v := range status {
		if v {
			enabled++
		}
	}

	if enabled == 0 {
		return false, "No shell enhancements enabled. Run 'bluefin-cli shell' to start."
	}

	return true, fmt.Sprintf("%d shell(s) configured.", enabled)
}

func checkContainer() (bool, string) {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true, "Running inside a container."
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return true, "Running inside a Podman container."
	}
	return true, "Running on host system."
}

func checkEnv() (bool, string) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	if env.IsWSL() {
		sb.WriteString(" (WSL)")
	}
	return true, sb.String()
}
