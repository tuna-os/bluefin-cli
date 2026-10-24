package shell

import (
	"os"
	"os/exec"
	"strings"
)

// CheckStatus returns a map of shell names to their enablement status.
func CheckStatus() map[string]bool {
	status := make(map[string]bool)
	home, _ := os.UserHomeDir()

	for _, spec := range registry {
		content, err := os.ReadFile(spec.ConfigPath(home))
		if err != nil {
			status[spec.Name] = false
			continue
		}

		status[spec.Name] = strings.Contains(string(content), shellMaker) || strings.Contains(string(content), blingMarker)
	}
	// Nushell is spelled both ways in the wild; report it under both.
	status["nushell"] = status["nu"]

	status["powershell"] = checkPowerShellStatus()
	status["pwsh"] = status["powershell"]

	return status
}

// GetInstalledShells returns a list of shells that are available in the PATH.
func GetInstalledShells() []string {
	var installed []string

	for _, spec := range registry {
		if spec.IsInstalled() {
			installed = append(installed, spec.Name)
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

// CheckDependencies checks availability of managed shell tools on the current platform.
func CheckDependencies() map[string]bool {
	status := make(map[string]bool)

	for _, tool := range toolsForCurrentPlatform() {
		status[tool.Binary] = isBinaryAvailable(tool)
	}

	return status
}
