package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/sunset"
	"github.com/hanthor/bluefin-cli/internal/tui"
	"github.com/spf13/cobra"
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
			return handleWSLDelegation()
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

		return runSunset(cfg)
	},
}

func handleWSLDelegation() error {
	winExe := findWindowsCLI()
	if winExe == "" {
		fmt.Println("WSL detected. Solar theme switching requires the Windows version of Bluefin CLI.")
		fmt.Println("Please install bluefin-cli.exe in Windows and ensure it is in your Windows PATH.")
		fmt.Println("\nYou can download it from the GitHub releases page.")
		return nil
	}

	fmt.Printf("Delegating sunset command to Windows CLI: %s\n", winExe)

	// Reconstruct the command for the Windows side
	args := []string{"sunset"}
	args = append(args, os.Args[2:]...)

	// In WSL, running .exe works directly if it's in the Windows path or referenced by full path
	command := exec.Command(winExe, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	return command.Run()
}

func findWindowsCLI() string {
	// Try looking in PATH first
	if path, err := exec.LookPath("bluefin-cli.exe"); err == nil {
		return path
	}

	// Try common locations if not in PATH
	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData != "" {
		// Check for wine path style if env is not translated properly,
		// but usually in WSL it is if configured.
		// However, WSL often has LOCALAPPDATA translated.
		candidate := filepath.Join(localAppData, "Microsoft", "WinGet", "Links", "bluefin-cli.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func runSunset(cfg *sunset.Config) error {
	if !cfg.Enabled {
		fmt.Println("Sunset theme switching is disabled. Use --enable to enable it.")
		return nil
	}

	now := time.Now()
	state := sunset.GetSolarState(cfg.Latitude, cfg.Longitude, now)
	isDay := state == sunset.StateDay

	fmt.Printf("Current solar state: %s\n", state)

	operator := sunset.NewThemeOperator()

	fmt.Printf("Applying %s theme...\n", state)
	if err := operator.SetTheme(isDay); err != nil {
		fmt.Printf("Warning: failed to set theme: %v\n", err)
	}

	targetWallpaper := ""
	if isDay {
		targetWallpaper = cfg.DayWallpaper
	} else {
		targetWallpaper = cfg.NightWallpaper
	}

	// If no manual wallpaper is set, try the monthly theme
	if targetWallpaper == "" && cfg.WallpaperTheme != "" {
		fmt.Printf("Looking up monthly %s wallpaper for %s...\n", cfg.WallpaperTheme, state)
		path, err := sunset.GetMonthlyWallpaper(cfg.WallpaperTheme, isDay)
		if err != nil {
			fmt.Printf("Warning: failed to find monthly wallpaper: %v\n", err)
		} else {
			targetWallpaper = path
		}
	}

	if targetWallpaper != "" {
		fmt.Printf("Applying wallpaper: %s\n", filepath.Base(targetWallpaper))
		if err := operator.SetWallpaper(targetWallpaper); err != nil {
			fmt.Printf("Warning: failed to set wallpaper: %v\n", err)
		}
	}

	return nil
}

func RunSunsetSetupFlow() error {
	// Handle WSL Delegation
	if env.IsWSL() {
		return handleWSLDelegation()
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
	).WithTheme(tui.AppTheme).WithKeyMap(tui.MenuKeyMap())

	if err := confForm.Run(); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Setup cancelled.")
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
	return runSunset(cfg)
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
