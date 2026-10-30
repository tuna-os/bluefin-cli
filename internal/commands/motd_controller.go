package commands

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/motd"
	"github.com/tuna-os/bluefin-cli/internal/shell"
)

// MOTDController manages MOTD configuration and display actions.
type MOTDController struct{}

// NewMOTDController returns a new MOTDController.
func NewMOTDController() *MOTDController {
	return &MOTDController{}
}

// IsEnabled checks if MOTD is currently enabled for the given shell.
func (c *MOTDController) IsEnabled(sh string) bool {
	cfg, err := shell.LoadConfig(sh)
	if err != nil {
		cfg = shell.DefaultConfig(sh)
	}
	return cfg.IsEnabled("Motd")
}

// Toggle toggles the MOTD setting for the given shell.
// It returns a Result describing the new state or an error if persistence failed.
func (c *MOTDController) Toggle(sh string) (Result, error) {
	cfg, err := shell.LoadConfig(sh)
	if err != nil {
		cfg = shell.DefaultConfig(sh)
	}
	enable := !cfg.IsEnabled("Motd")
	cfg.SetEnabled("Motd", enable)
	if err := shell.SaveConfig(cfg); err != nil {
		return Result{Success: false, Message: "Error: " + err.Error()}, fmt.Errorf("saving %s config: %w", sh, err)
	}
	if enable {
		return Result{Success: true, Message: "MOTD enabled.", Data: true}, nil
	}
	return Result{Success: true, Message: "MOTD disabled.", Data: false}, nil
}

// SetEnabled explicitly sets MOTD enabled state for the specified shells.
func (c *MOTDController) SetEnabled(shells []string, enable bool) (Result, error) {
	for _, sh := range shells {
		cfg, err := shell.LoadConfig(sh)
		if err != nil {
			cfg = shell.DefaultConfig(sh)
		}
		cfg.SetEnabled("Motd", enable)
		if err := shell.SaveConfig(cfg); err != nil {
			return Result{Success: false}, fmt.Errorf("saving %s config: %w", sh, err)
		}
	}
	state := "disabled"
	if enable {
		state = "enabled"
	}
	return Result{
		Success: true,
		Message: fmt.Sprintf("MOTD %s for %v.", state, shells),
		Data:    enable,
	}, nil
}

// Show renders the MOTD output to stdout.
func (c *MOTDController) Show() error {
	return motd.Show()
}
