package sunset

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeOperator struct {
	setThemeErr     error
	setWallpaperErr error
	themeCalls      []bool
	wallpaperCalls  []string
}

func (f *fakeOperator) SetTheme(isLight bool) error {
	f.themeCalls = append(f.themeCalls, isLight)
	return f.setThemeErr
}

func (f *fakeOperator) SetWallpaper(path string) error {
	f.wallpaperCalls = append(f.wallpaperCalls, path)
	return f.setWallpaperErr
}

type recordingReporter struct {
	infos []string
	warns []string
}

func (r *recordingReporter) Infof(format string, args ...any) {
	r.infos = append(r.infos, format)
}

func (r *recordingReporter) Warnf(format string, args ...any) {
	r.warns = append(r.warns, format)
}

func withFixedTime(t *testing.T, when time.Time) {
	t.Helper()
	original := currentTime
	currentTime = func() time.Time { return when }
	t.Cleanup(func() { currentTime = original })
}

func TestApply_Disabled(t *testing.T) {
	cfg := &Config{Enabled: false}
	op := &fakeOperator{}
	r := &recordingReporter{}

	if err := Apply(cfg, op, r); err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}
	if len(op.themeCalls) != 0 || len(op.wallpaperCalls) != 0 {
		t.Errorf("Apply() on disabled config called the operator: %+v", op)
	}
}

func TestApply_DayWithManualWallpaper(t *testing.T) {
	// NYC noon in June is day (see sunset_test.go's TestGetSolarState).
	withFixedTime(t, time.Date(2024, 6, 20, 12, 0, 0, 0, time.UTC))

	cfg := &Config{
		Enabled:      true,
		Latitude:     40.7128,
		Longitude:    -74.0060,
		DayWallpaper: "/wallpapers/day.jpg",
	}
	op := &fakeOperator{}
	r := &recordingReporter{}

	if err := Apply(cfg, op, r); err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}
	if len(op.themeCalls) != 1 || op.themeCalls[0] != true {
		t.Errorf("themeCalls = %v, want [true]", op.themeCalls)
	}
	if len(op.wallpaperCalls) != 1 || op.wallpaperCalls[0] != "/wallpapers/day.jpg" {
		t.Errorf("wallpaperCalls = %v, want [/wallpapers/day.jpg]", op.wallpaperCalls)
	}
}

func TestApply_NightWithMonthlyFallback(t *testing.T) {
	withFixedTime(t, time.Date(2024, 6, 20, 4, 0, 0, 0, time.UTC)) // NYC midnight-ish, night

	tmpHome := t.TempDir()
	originalGetHomeDir := getHomeDir
	getHomeDir = func() (string, error) { return tmpHome, nil }
	t.Cleanup(func() { getHomeDir = originalGetHomeDir })

	wallpaperDir := filepath.Join(tmpHome, "Pictures", "BluefinCLI", "bluefin")
	if err := os.MkdirAll(wallpaperDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	monthly := filepath.Join(wallpaperDir, "06-night.jpg")
	if err := os.WriteFile(monthly, []byte("fake"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := &Config{
		Enabled:        true,
		Latitude:       40.7128,
		Longitude:      -74.0060,
		WallpaperTheme: "bluefin",
	}
	op := &fakeOperator{}
	r := &recordingReporter{}

	if err := Apply(cfg, op, r); err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}
	if len(op.themeCalls) != 1 || op.themeCalls[0] != false {
		t.Errorf("themeCalls = %v, want [false]", op.themeCalls)
	}
	if len(op.wallpaperCalls) != 1 || op.wallpaperCalls[0] != monthly {
		t.Errorf("wallpaperCalls = %v, want [%s]", op.wallpaperCalls, monthly)
	}
}

func TestApply_NoWallpaperConfigured(t *testing.T) {
	withFixedTime(t, time.Date(2024, 6, 20, 12, 0, 0, 0, time.UTC))

	cfg := &Config{Enabled: true, Latitude: 40.7128, Longitude: -74.0060}
	op := &fakeOperator{}
	r := &recordingReporter{}

	if err := Apply(cfg, op, r); err != nil {
		t.Fatalf("Apply() = %v, want nil", err)
	}
	if len(op.wallpaperCalls) != 0 {
		t.Errorf("wallpaperCalls = %v, want none", op.wallpaperCalls)
	}
}

func TestApply_OperatorErrorsAreWarningsNotFailures(t *testing.T) {
	withFixedTime(t, time.Date(2024, 6, 20, 12, 0, 0, 0, time.UTC))

	cfg := &Config{
		Enabled:      true,
		Latitude:     40.7128,
		Longitude:    -74.0060,
		DayWallpaper: "/wallpapers/day.jpg",
	}
	op := &fakeOperator{
		setThemeErr:     errors.New("boom theme"),
		setWallpaperErr: errors.New("boom wallpaper"),
	}
	r := &recordingReporter{}

	if err := Apply(cfg, op, r); err != nil {
		t.Fatalf("Apply() = %v, want nil (operator errors become warnings)", err)
	}
	if len(r.warns) != 2 {
		t.Errorf("warns = %v, want 2 entries", r.warns)
	}
}
