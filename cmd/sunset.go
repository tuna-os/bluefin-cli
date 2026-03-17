package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/hanthor/bluefin-cli/internal/sunset"
	"github.com/spf13/cobra"
)

var (
	lat       float64
	lon       float64
	dayWall   string
	nightWall string
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
	for _, arg := range os.Args[2:] {
		args = append(args, arg)
	}

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
	
	fmt.Printf("Current solar state: %s\n", state)

	operator := sunset.NewThemeOperator()

	if state == sunset.StateDay {
		fmt.Println("Applying day theme...")
		if err := operator.SetTheme(true); err != nil {
			fmt.Printf("Warning: failed to set light theme: %v\n", err)
		}
		if cfg.DayWallpaper != "" {
			fmt.Println("Applying day wallpaper...")
			if err := operator.SetWallpaper(cfg.DayWallpaper); err != nil {
				fmt.Printf("Warning: failed to set day wallpaper: %v\n", err)
			}
		}
	} else {
		fmt.Println("Applying night theme...")
		if err := operator.SetTheme(false); err != nil {
			fmt.Printf("Warning: failed to set dark theme: %v\n", err)
		}
		if cfg.NightWallpaper != "" {
			fmt.Println("Applying night wallpaper...")
			if err := operator.SetWallpaper(cfg.NightWallpaper); err != nil {
				fmt.Printf("Warning: failed to set night wallpaper: %v\n", err)
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(sunsetCmd)

	sunsetCmd.Flags().Float64Var(&lat, "latitude", 0, "Latitude for solar calculations")
	sunsetCmd.Flags().Float64Var(&lon, "longitude", 0, "Longitude for solar calculations")
	sunsetCmd.Flags().StringVar(&dayWall, "day-wallpaper", "", "Path to the day wallpaper image")
	sunsetCmd.Flags().StringVar(&nightWall, "night-wallpaper", "", "Path to the night wallpaper image")
	sunsetCmd.Flags().BoolVar(&enable, "enable", false, "Enable or disable sunset switching")
}
