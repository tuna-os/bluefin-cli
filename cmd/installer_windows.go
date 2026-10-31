//go:build windows

package cmd

import "github.com/tuna-os/bluefin-cli/internal/install"

func defaultPlatformInstaller() install.Installer {
	return &install.WindowsInstaller{}
}
