package platform

import (
	"testing"
)

type mockPlatform struct {
	name  string
	goos  string
	isWin bool
	isWSL bool
	isAlp bool
	caps  Capabilities
}

func (m *mockPlatform) Name() string               { return m.name }
func (m *mockPlatform) GOOS() string               { return m.goos }
func (m *mockPlatform) IsWindows() bool            { return m.isWin }
func (m *mockPlatform) IsWSL() bool                { return m.isWSL }
func (m *mockPlatform) IsAlpine() bool             { return m.isAlp }
func (m *mockPlatform) Capabilities() Capabilities { return m.caps }

func TestPlatformDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		wantName string
		wantGOOS string
		wantWin  bool
		wantWSL  bool
		wantAlp  bool
		wantCaps Capabilities
	}{
		{
			name:     "windows",
			platform: NewWindowsPlatform(),
			wantName: "windows",
			wantGOOS: "windows",
			wantWin:  true,
			wantWSL:  false,
			wantAlp:  false,
			wantCaps: Capabilities{
				SupportsHomebrew:   false,
				SupportsPowerShell: true,
				SupportsSunset:     true,
				HasAPK:             false,
				PackageManager:     "winget",
			},
		},
		{
			name:     "wsl",
			platform: NewWSLPlatform(),
			wantName: "wsl",
			wantGOOS: "linux",
			wantWin:  false,
			wantWSL:  true,
			wantAlp:  false,
			wantCaps: Capabilities{
				SupportsHomebrew:   true,
				SupportsPowerShell: false,
				SupportsSunset:     true,
				HasAPK:             false,
				PackageManager:     "brew",
			},
		},
		{
			name:     "alpine",
			platform: NewAlpinePlatform(),
			wantName: "alpine",
			wantGOOS: "linux",
			wantWin:  false,
			wantWSL:  false,
			wantAlp:  true,
			wantCaps: Capabilities{
				SupportsHomebrew:   false,
				SupportsPowerShell: false,
				SupportsSunset:     false,
				HasAPK:             true,
				PackageManager:     "apk",
			},
		},
		{
			name:     "darwin",
			platform: NewUnixPlatform("darwin"),
			wantName: "darwin",
			wantGOOS: "darwin",
			wantWin:  false,
			wantWSL:  false,
			wantAlp:  false,
			wantCaps: Capabilities{
				SupportsHomebrew:   true,
				SupportsPowerShell: false,
				SupportsSunset:     false,
				HasAPK:             false,
				PackageManager:     "brew",
			},
		},
		{
			name:     "linux",
			platform: NewUnixPlatform("linux"),
			wantName: "linux",
			wantGOOS: "linux",
			wantWin:  false,
			wantWSL:  false,
			wantAlp:  false,
			wantCaps: Capabilities{
				SupportsHomebrew:   true,
				SupportsPowerShell: false,
				SupportsSunset:     false,
				HasAPK:             false,
				PackageManager:     "brew",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.Name(); got != tt.wantName {
				t.Errorf("Name() = %v, want %v", got, tt.wantName)
			}
			if got := tt.platform.GOOS(); got != tt.wantGOOS {
				t.Errorf("GOOS() = %v, want %v", got, tt.wantGOOS)
			}
			if got := tt.platform.IsWindows(); got != tt.wantWin {
				t.Errorf("IsWindows() = %v, want %v", got, tt.wantWin)
			}
			if got := tt.platform.IsWSL(); got != tt.wantWSL {
				t.Errorf("IsWSL() = %v, want %v", got, tt.wantWSL)
			}
			if got := tt.platform.IsAlpine(); got != tt.wantAlp {
				t.Errorf("IsAlpine() = %v, want %v", got, tt.wantAlp)
			}
			if got := tt.platform.Capabilities(); got != tt.wantCaps {
				t.Errorf("Capabilities() = %+v, want %+v", got, tt.wantCaps)
			}
		})
	}
}

func TestCurrentAndMockOverride(t *testing.T) {
	curr := Current()
	if curr == nil {
		t.Fatal("Current() returned nil")
	}

	mock := &mockPlatform{
		name:  "mock-os",
		goos:  "mock-goos",
		isWin: false,
		isWSL: true,
		isAlp: false,
		caps: Capabilities{
			SupportsHomebrew: false,
			PackageManager:   "custom-pm",
		},
	}

	restore := SetCurrent(mock)
	defer restore()

	if Name() != "mock-os" {
		t.Errorf("Name() = %v, want mock-os", Name())
	}
	if GOOS() != "mock-goos" {
		t.Errorf("GOOS() = %v, want mock-goos", GOOS())
	}
	if !IsWSL() {
		t.Errorf("IsWSL() = false, want true")
	}
	if IsWindows() {
		t.Errorf("IsWindows() = true, want false")
	}
	if IsAlpine() {
		t.Errorf("IsAlpine() = true, want false")
	}
	if GetCapabilities().PackageManager != "custom-pm" {
		t.Errorf("GetCapabilities().PackageManager = %v, want custom-pm", GetCapabilities().PackageManager)
	}

	restore()
	if Name() != curr.Name() {
		t.Errorf("After restore, Name() = %v, want %v", Name(), curr.Name())
	}
}
