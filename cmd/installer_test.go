package cmd

import (
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/install"
)

func TestInstallerInitialization(t *testing.T) {
	orig := install.GetInstaller()
	t.Cleanup(func() {
		install.SetInstaller(orig)
	})

	// Test default installer matches platform default
	initInstaller()
	got := install.GetInstaller()
	if got == nil {
		t.Fatal("expected non-nil installer after initInstaller()")
	}

	// Test BLUEFIN_INSTALLER=mock override
	t.Setenv("BLUEFIN_INSTALLER", "mock")
	initInstaller()
	mockInst := install.GetInstaller()
	if _, ok := mockInst.(*install.MockInstaller); !ok {
		t.Fatalf("expected *install.MockInstaller when BLUEFIN_INSTALLER=mock, got %T", mockInst)
	}

	// Test BLUEFIN_INSTALLER unset returns to default
	t.Setenv("BLUEFIN_INSTALLER", "")
	initInstaller()
	resetInst := install.GetInstaller()
	if _, ok := resetInst.(*install.MockInstaller); ok {
		t.Fatalf("expected platform installer when BLUEFIN_INSTALLER is empty, got %T", resetInst)
	}
}

func TestMockInstallerExecution(t *testing.T) {
	var bundleCalled, wallpaperCalled, cleanupCalled bool
	mock := &install.MockInstaller{
		InstallBundleFunc: func(nameOrPath ...string) error {
			bundleCalled = true
			return nil
		},
		InstallWallpapersFunc: func(casks []string) error {
			wallpaperCalled = true
			return nil
		},
		CleanupWallpapersFunc: func(all bool) error {
			cleanupCalled = true
			return nil
		},
	}

	orig := install.GetInstaller()
	t.Cleanup(func() {
		install.SetInstaller(orig)
	})

	install.SetInstaller(mock)

	if err := install.Bundle("cli"); err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}
	if !bundleCalled {
		t.Error("expected InstallBundle to be called on mock")
	}

	if err := install.InstallWallpaperCasks([]string{"test-cask"}); err != nil {
		t.Fatalf("InstallWallpaperCasks failed: %v", err)
	}
	if !wallpaperCalled {
		t.Error("expected InstallWallpapers to be called on mock")
	}

	if err := install.CleanupWallpapers(false); err != nil {
		t.Fatalf("CleanupWallpapers failed: %v", err)
	}
	if !cleanupCalled {
		t.Error("expected CleanupWallpapers to be called on mock")
	}
}
