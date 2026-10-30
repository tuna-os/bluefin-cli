package commands

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// ShellController manages shell integration status and toggling.
type ShellController struct{}

// NewShellController returns a new ShellController.
func NewShellController() *ShellController {
	return &ShellController{}
}

// GetInstalledShells returns the list of detected shells on the machine.
func (c *ShellController) GetInstalledShells() []string {
	return shell.GetInstalledShells()
}

// CheckStatus returns a map of shell names to their integration status.
func (c *ShellController) CheckStatus() map[string]bool {
	return shell.CheckStatus()
}

// Toggle toggles the integration for a single shell.
func (c *ShellController) Toggle(sh string, enable bool) (Result, error) {
	if err := shell.Toggle(sh, enable); err != nil {
		return Result{Success: false, Message: err.Error()}, err
	}
	state := "disabled"
	if enable {
		state = "enabled"
	}
	return Result{
		Success: true,
		Message: fmt.Sprintf("Shell %s %s.", sh, state),
		Data:    enable,
	}, nil
}

// ApplyChanges applies desired state to installed shells and returns which shells were changed.
func (c *ShellController) ApplyChanges(want map[string]bool, initial map[string]bool, installed []string) ([]string, error) {
	var changed []string
	for _, sh := range installed {
		if want[sh] != initial[sh] {
			fmt.Printf("%s -> %v\n", sh, want[sh])
			if err := shell.Toggle(sh, want[sh]); err != nil {
				return changed, fmt.Errorf("toggling %s: %w", sh, err)
			}
			changed = append(changed, sh)
		}
	}
	return changed, nil
}
