//go:build extra

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/env"
	"github.com/tuna-os/bluefin-cli/internal/sunset"
	"github.com/tuna-os/bluefin-cli/internal/tui"
)

var (
	lat       float64
	lon       float64
	dayWall   string
	nightWall string
	wallTheme string
	enable    bool
)

var sunsetCmd = &cobra.Command{
	Use:   "sunset",
	Short: "Manage solar-based theme and wallpaper switching",
	Long:  `Automatically switch between light and dark themes and different wallpapers based on sunrise and sunset times for your location.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle WSL Delegation
		if env.IsWSL() {
			return sunset.DelegateToWindowsCLI(
				append([]string{"sunset"}, os.Args[2:]...),
				sunset.RunnerIO{Stdout: os.Stdout, Stderr: os.Stderr, Stdin: os.Stdin},
				nil,
			)
		}

		// Non-Windows guard (native Linux)
		if runtime.GOOS != "windows" {
			fmt.Println("Solar theme/wallpaper switching is only supported on Windows or via WSL.")
			return nil
		}

		cfg, err := sunset.LoadConfig()
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("latitude") {
			cfg.Latitude = lat
		}
		if cmd.Flags().Changed("longitude") {
			cfg.Longitude = lon
		}
		if cmd.Flags().Changed("day-wallpaper") {
			cfg.DayWallpaper = dayWall
		}
		if cmd.Flags().Changed("night-wallpaper") {
			cfg.NightWallpaper = nightWall
		}
		if cmd.Flags().Changed("wallpaper-theme") {
			cfg.WallpaperTheme = wallTheme
		}
		if cmd.Flags().Changed("enable") {
			cfg.Enabled = enable
		}

		if cmd.Flags().NFlag() > 0 {
			if err := sunset.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Println("Configuration updated.")
		}

		return sunset.Apply(cfg, sunset.NewThemeOperator(), nil)
	},
}

func RunSunsetSetupFlow() error {
	// Handle WSL Delegation
	if env.IsWSL() {
		return sunset.DelegateToWindowsCLI(
			append([]string{"sunset"}, os.Args[2:]...),
			sunset.RunnerIO{Stdout: os.Stdout, Stderr: os.Stderr, Stdin: os.Stdin},
			nil,
		)
	}

	// Non-Windows guard (native Linux)
	if runtime.GOOS != "windows" {
		fmt.Println("Solar theme/wallpaper switching is only supported on Windows or via WSL.")
		return nil
	}

	cfg, err := sunset.LoadConfig()
	if err != nil {
		return err
	}

	var cityName string
	var themeChoice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter your city").
				Placeholder("e.g. New York, London, Tokyo").
				Value(&cityName).
				Validate(func(s string) error {
					if len(s) < 2 {
						return fmt.Errorf("please enter a valid city name")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Choose a wallpaper theme").
				Options(
					huh.NewOption("Bluefin", "bluefin"),
					huh.NewOption("Aurora", "aurora"),
					huh.NewOption("Bazzite", "bazzite"),
					huh.NewOption("None (Keep current)", ""),
				).
				Value(&themeChoice),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := form.Run(); err != nil {
		return err
	}

	fmt.Printf("Searching for %s...\n", cityName)
	result, err := sunset.GeocodeCity(cityName)
	if err != nil {
		return fmt.Errorf("failed to resolve city: %w", err)
	}

	fmt.Printf("Resolved: %s, %s, %s (Lat: %.4f, Long: %.4f)\n",
		result.Name, result.Admin1, result.Country, result.Latitude, result.Longitude)

	confirm := true
	confForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use these coordinates?").
				Value(&confirm),
		),
	).WithTheme(tui.AppTheme).WithKeyMap(tui.ConfirmKeyMap())

	if err := confForm.Run(); err != nil {
		return err
	}

	if !confirm {
		return nil
	}

	cfg.Latitude = result.Latitude
	cfg.Longitude = result.Longitude
	cfg.WallpaperTheme = themeChoice
	cfg.Enabled = true

	if err := sunset.SaveConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Configuration updated and feature enabled!")
	return sunset.Apply(cfg, sunset.NewThemeOperator(), nil)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup for your location",
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunSunsetSetupFlow()
	},
}

func init() {
	rootCmd.AddCommand(sunsetCmd)
	sunsetCmd.AddCommand(setupCmd)

	sunsetCmd.Flags().Float64Var(&lat, "latitude", 0, "Latitude for solar calculations")
	sunsetCmd.Flags().Float64Var(&lon, "longitude", 0, "Longitude for solar calculations")
	sunsetCmd.Flags().StringVar(&dayWall, "day-wallpaper", "", "Path to the day wallpaper image")
	sunsetCmd.Flags().StringVar(&nightWall, "night-wallpaper", "", "Path to the night wallpaper image")
	sunsetCmd.Flags().StringVar(&wallTheme, "wallpaper-theme", "", "Theme for monthly wallpapers (bluefin, aurora, bazzite)")
	sunsetCmd.Flags().BoolVar(&enable, "enable", false, "Enable or disable sunset switching")
}
