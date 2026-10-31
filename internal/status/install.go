package status

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/install"
)

func reportPackageManager() string {
	var out string
	out += labelStyle.Render("Package Manager:") + "\n"
	if env.IsWindows() {
		managers := install.AvailableWindowsManagers()
		if len(managers) == 0 {
			out += report(false, "Windows PMs", "none detected")
			out += "    Install winget, scoop, or chocolatey\n"
		} else {
			out += report(true, "Windows PMs", strings.Join(managers, ", "))
		}
	} else {
		if _, err := exec.LookPath("brew"); err == nil {
			out += report(true, "Homebrew", "installed")

			ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
			defer cancel()
			if output, err := exec.CommandContext(ctx, "brew", "--version").Output(); err == nil {
				version := string(output)
				if len(version) > 0 {
					out += fmt.Sprintf("    %s\n", version[:len(version)-1])
				}
			}
		} else {
			out += report(false, "Homebrew", "not installed")
			out += "    Install from: https://brew.sh\n"
		}
	}
	return out
}
