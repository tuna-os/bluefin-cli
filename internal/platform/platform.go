package platform

import (
	"runtime"
	"sync"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

// Capabilities describes feature and toolchain support for a platform.
type Capabilities struct {
	SupportsHomebrew   bool
	SupportsPowerShell bool
	SupportsSunset     bool
	HasAPK             bool
	PackageManager     string // "brew", "winget", "apk", "coldbrew", etc.
}

// Platform defines the interface for platform identification and capabilities.
type Platform interface {
	Name() string
	GOOS() string
	IsWindows() bool
	IsWSL() bool
	IsAlpine() bool
	Capabilities() Capabilities
}

type basePlatform struct {
	name  string
	goos  string
	caps  Capabilities
	isWin bool
	isWSL bool
	isAlp bool
}

func (b *basePlatform) Name() string               { return b.name }
func (b *basePlatform) GOOS() string               { return b.goos }
func (b *basePlatform) IsWindows() bool            { return b.isWin }
func (b *basePlatform) IsWSL() bool                { return b.isWSL }
func (b *basePlatform) IsAlpine() bool             { return b.isAlp }
func (b *basePlatform) Capabilities() Capabilities { return b.caps }

type WindowsPlatform struct {
	basePlatform
}

func NewWindowsPlatform() *WindowsPlatform {
	return &WindowsPlatform{
		basePlatform: basePlatform{
			name:  "windows",
			goos:  "windows",
			isWin: true,
			caps: Capabilities{
				SupportsHomebrew:   false,
				SupportsPowerShell: true,
				SupportsSunset:     true,
				HasAPK:             false,
				PackageManager:     "winget",
			},
		},
	}
}

type WSLPlatform struct {
	basePlatform
}

func NewWSLPlatform() *WSLPlatform {
	return &WSLPlatform{
		basePlatform: basePlatform{
			name:  "wsl",
			goos:  "linux",
			isWSL: true,
			caps: Capabilities{
				SupportsHomebrew:   true,
				SupportsPowerShell: false,
				SupportsSunset:     true,
				HasAPK:             false,
				PackageManager:     "brew",
			},
		},
	}
}

type AlpinePlatform struct {
	basePlatform
}

func NewAlpinePlatform() *AlpinePlatform {
	return &AlpinePlatform{
		basePlatform: basePlatform{
			name:  "alpine",
			goos:  "linux",
			isAlp: true,
			caps: Capabilities{
				SupportsHomebrew:   false,
				SupportsPowerShell: false,
				SupportsSunset:     false,
				HasAPK:             true,
				PackageManager:     "apk",
			},
		},
	}
}

type UnixPlatform struct {
	basePlatform
}

func NewUnixPlatform(goos string) *UnixPlatform {
	return &UnixPlatform{
		basePlatform: basePlatform{
			name: goos,
			goos: goos,
			caps: Capabilities{
				SupportsHomebrew:   true,
				SupportsPowerShell: false,
				SupportsSunset:     false,
				HasAPK:             false,
				PackageManager:     "brew",
			},
		},
	}
}

var (
	mu              sync.RWMutex
	currentPlatform Platform
)

// Detect detects the current runtime platform using host environment and runtime GOOS.
func Detect() Platform {
	if runtime.GOOS == "windows" {
		return NewWindowsPlatform()
	}
	if env.IsWSL() {
		return NewWSLPlatform()
	}
	if env.IsAlpine() {
		return NewAlpinePlatform()
	}
	return NewUnixPlatform(runtime.GOOS)
}

// Current returns the active Platform instance (detecting it if not already set).
func Current() Platform {
	mu.RLock()
	p := currentPlatform
	mu.RUnlock()
	if p != nil {
		return p
	}

	mu.Lock()
	defer mu.Unlock()
	if currentPlatform == nil {
		currentPlatform = Detect()
	}
	return currentPlatform
}

// SetCurrent allows overriding the active Platform (primarily for testing and mocking).
// Returns a restore function that reverts to the previous Platform.
func SetCurrent(p Platform) func() {
	mu.Lock()
	prev := currentPlatform
	currentPlatform = p
	mu.Unlock()
	return func() {
		mu.Lock()
		currentPlatform = prev
		mu.Unlock()
	}
}

// Convenience helpers delegating to Current()

func Name() string                  { return Current().Name() }
func GOOS() string                  { return Current().GOOS() }
func IsWindows() bool               { return Current().IsWindows() }
func IsWSL() bool                   { return Current().IsWSL() }
func IsAlpine() bool                { return Current().IsAlpine() }
func GetCapabilities() Capabilities { return Current().Capabilities() }
