package sunset

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/wallpaper"
)

func TestGeocodeCity(t *testing.T) {
	// Mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openMeteoResponse{
			Results: []GeocodeResult{
				{Name: "New York", Latitude: 40.7128, Longitude: -74.0060},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Mock the base URL
	originalURL := GeocodingBaseURL
	GeocodingBaseURL = server.URL
	defer func() { GeocodingBaseURL = originalURL }()

	result, err := GeocodeCity("New York")
	if err != nil {
		t.Fatalf("GeocodeCity failed: %v", err)
	}
	if result.Name != "New York" || result.Latitude != 40.7128 {
		t.Errorf("Unexpected geocoding result: %+v", result)
	}
}

func TestGetMonthlyWallpaper(t *testing.T) {
	tmpHome := t.TempDir()

	// Mock getHomeDir
	originalGetHomeDir := getHomeDir
	getHomeDir = func() (string, error) { return tmpHome, nil }
	defer func() { getHomeDir = originalGetHomeDir }()

	// Mock currentTime
	testTime := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	originalCurrentTime := currentTime
	currentTime = func() time.Time { return testTime }
	defer func() { currentTime = originalCurrentTime }()

	theme := "bluefin"
	wallpaperDir := filepath.Join(tmpHome, "Pictures", "BluefinCLI", theme)
	if err := os.MkdirAll(wallpaperDir, 0755); err != nil {
		t.Fatalf("Failed to create mock wallpaper dir: %v", err)
	}

	// Create a mock wallpaper file
	mockFile := filepath.Join(wallpaperDir, "03-day.jpg")
	if err := os.WriteFile(mockFile, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("Failed to create mock wallpaper file: %v", err)
	}

	path, err := GetMonthlyWallpaper(theme, true)
	if err != nil {
		t.Fatalf("GetMonthlyWallpaper failed: %v", err)
	}
	if filepath.Base(path) != "03-day.jpg" {
		t.Errorf("Expected 03-day.jpg, got %s", filepath.Base(path))
	}
}

// Sunset reads monthly wallpapers from the directory the installer writes
// each collection to, whichever form of the theme name the config holds.
func TestGetMonthlyWallpaper_ReadsInstalledCollectionDir(t *testing.T) {
	tmpHome := t.TempDir()
	originalGetHomeDir := getHomeDir
	getHomeDir = func() (string, error) { return tmpHome, nil }
	t.Cleanup(func() { getHomeDir = originalGetHomeDir })

	originalCurrentTime := currentTime
	currentTime = func() time.Time { return time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { currentTime = originalCurrentTime })

	for _, c := range wallpaper.Collections {
		dir := wallpaper.CollectionDir(tmpHome, c.ID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "03-night.jpg")
		if err := os.WriteFile(want, []byte("fake"), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, theme := range []string{c.ID, c.Name} {
			got, err := GetMonthlyWallpaper(theme, false)
			if err != nil {
				t.Fatalf("GetMonthlyWallpaper(%q): %v", theme, err)
			}
			if got != want {
				t.Errorf("GetMonthlyWallpaper(%q) = %q, want %q", theme, got, want)
			}
		}
	}
}
