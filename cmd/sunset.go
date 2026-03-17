package cmd

import (
	"fmt"
	"time"

	"github.com/hanthor/bluefin-cli/internal/sunset"
	"github.com/spf13/cobra"
)

var (
	lat      float64
	lon      float64
	dayWall  string
	nightWall string
	enable   bool
)

var sunsetCmd = &cobra.Command{
	Use:   "sunset",
	Short: "Manage solar-based theme and wallpaper switching",
	Long:  `Automatically switch between light and dark themes and different wallpapers based on sunrise and sunset times for your location.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

func runSunset(cfg *sunset.Config) error {
	if !cfg.Enabled {
		fmt.Println("Sunset theme switching is disabled. Use --enable to enable it.")
		return nil
	}

	now := time.Now()
	state := sunset.GetSolarState(cfg.Latitude, cfg.Longitude, now)
	
	fmt.Printf("Current solar state: %s\n", state)

	if state == sunset.StateDay {
		fmt.Println("Applying day theme...")
		if err := sunset.SetWindowsTheme(true); err != nil {
			fmt.Printf("Warning: failed to set light theme: %v\n", err)
		}
		if cfg.DayWallpaper != "" {
			fmt.Println("Applying day wallpaper...")
			if err := sunset.SetWallpaper(cfg.DayWallpaper); err != nil {
				fmt.Printf("Warning: failed to set day wallpaper: %v\n", err)
			}
		}
	} else {
		fmt.Println("Applying night theme...")
		if err := sunset.SetWindowsTheme(false); err != nil {
			fmt.Printf("Warning: failed to set dark theme: %v\n", err)
		}
		if cfg.NightWallpaper != "" {
			fmt.Println("Applying night wallpaper...")
			if err := sunset.SetWallpaper(cfg.NightWallpaper); err != nil {
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
