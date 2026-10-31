package commands

import (
	"fmt"

	"github.com/tuna-os/bluefin-cli/internal/starship"
)

// StarshipController manages Starship installation and theme application.
type StarshipController struct{}

// NewStarshipController returns a new StarshipController.
func NewStarshipController() *StarshipController {
	return &StarshipController{}
}

// Install ensures the starship binary is installed.
func (c *StarshipController) Install() error {
	return starship.Install()
}

// ApplyTheme installs Starship if needed and applies the specified preset theme.
func (c *StarshipController) ApplyTheme(preset string) (Result, error) {
	if err := starship.Install(); err != nil {
		return Result{Success: false, Message: err.Error()}, fmt.Errorf("installing starship: %w", err)
	}
	if err := starship.ApplyTheme(preset); err != nil {
		return Result{Success: false, Message: err.Error()}, fmt.Errorf("applying starship theme %q: %w", preset, err)
	}
	return Result{
		Success: true,
		Message: "Applied Starship preset: " + preset,
		Data:    preset,
	}, nil
}
