package sunset

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGetSolarState(t *testing.T) {
	// NYC Coordinates
	lat := 40.7128
	lon := -74.0060

	// Test noon (Day)
	noon := time.Date(2024, 6, 20, 12, 0, 0, 0, time.UTC)
	state := GetSolarState(lat, lon, noon)
	if state != StateDay {
		t.Errorf("Expected StateDay for NYC noon in June")
	}

	// Test midnight (Night in NYC)
	midnight := time.Date(2024, 6, 20, 4, 0, 0, 0, time.UTC)
	state = GetSolarState(lat, lon, midnight)
	if state != StateNight {
		t.Errorf("Expected StateNight for NYC midnight in June (04:00 UTC)")
	}
}

func TestConfigLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sunset.json")

	cfg := &Config{
		Enabled:        true,
		Latitude:       40.7128,
		Longitude:      -74.0060,
		DayWallpaper:   "day.jpg",
		NightWallpaper: "night.jpg",
		WallpaperTheme: "bluefin",
	}

	err := cfg.SaveTo(configPath)
	if err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded := &Config{}
	err = loaded.LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.Latitude != cfg.Latitude || loaded.WallpaperTheme != cfg.WallpaperTheme {
		t.Errorf("Loaded config does not match saved config")
	}
}
