package sunset

import (
	"fmt"
	"path/filepath"
)

// Reporter receives progress and warning messages produced by Apply. Cobra
// command handlers pass a reporter that prints to stdout; tests can pass one
// that records calls instead.
type Reporter interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
}

// PrintReporter is the default Reporter, used by the sunset command.
type PrintReporter struct{}

func (PrintReporter) Infof(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (PrintReporter) Warnf(format string, args ...any) {
	fmt.Printf("Warning: "+format+"\n", args...)
}

// Apply determines the current solar state for cfg's coordinates, applies
// the corresponding theme through operator, and applies a wallpaper chosen
// from cfg's manual override or, failing that, cfg's monthly wallpaper
// theme. Operator failures are reported through r as warnings rather than
// returned, matching the command's existing best-effort behavior: a failed
// theme or wallpaper change should not stop the other from being attempted.
//
// Apply reports cfg.Enabled == false as a no-op through r and returns nil;
// callers that want a hard error for a disabled config should check
// cfg.Enabled themselves before calling Apply.
func Apply(cfg *Config, operator ThemeOperator, r Reporter) error {
	if r == nil {
		r = PrintReporter{}
	}

	if !cfg.Enabled {
		r.Infof("Sunset theme switching is disabled. Use --enable to enable it.")
		return nil
	}

	state := GetSolarState(cfg.Latitude, cfg.Longitude, currentTime())
	isDay := state == StateDay

	r.Infof("Current solar state: %s", state)

	r.Infof("Applying %s theme...", state)
	if err := operator.SetTheme(isDay); err != nil {
		r.Warnf("failed to set theme: %v", err)
	}

	targetWallpaper := cfg.NightWallpaper
	if isDay {
		targetWallpaper = cfg.DayWallpaper
	}

	// If no manual wallpaper is set, try the monthly theme.
	if targetWallpaper == "" && cfg.WallpaperTheme != "" {
		r.Infof("Looking up monthly %s wallpaper for %s...", cfg.WallpaperTheme, state)
		path, err := GetMonthlyWallpaper(cfg.WallpaperTheme, isDay)
		if err != nil {
			r.Warnf("failed to find monthly wallpaper: %v", err)
		} else {
			targetWallpaper = path
		}
	}

	if targetWallpaper != "" {
		r.Infof("Applying wallpaper: %s", filepath.Base(targetWallpaper))
		if err := operator.SetWallpaper(targetWallpaper); err != nil {
			r.Warnf("failed to set wallpaper: %v", err)
		}
	}

	return nil
}
