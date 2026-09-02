package cmd

import (
	"context"
	"fmt"
	"runtime/debug"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/countme"
	"github.com/tuna-os/bluefin-cli/internal/status"
)

var (
	version = "dev"
)

// reportedVersion returns the release version stamped at build time. For
// `go install ...@latest` builds (no ldflags stamp), it falls back to the
// Go module version so countme telemetry and --version don't collapse
// into an undifferentiated "dev".
func reportedVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok &&
		info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

var rootCmd = &cobra.Command{
	Use:   "bluefin-cli",
	Short: "A powerful CLI tool for managing Homebrew and shell customization",
	Long: `Bluefin CLI brings the bluefin terminal experience to you.

- Homebrew & Tool Management
- Shell Environment Configuration
- System Status & MOTD
- Starship Theme Management
- Automated Theme & Wallpaper Switching (Sunset)
- Automated Font Installation
- Monthly Wallpaper Themes`,
	Version: reportedVersion(),

	// Fire the countme ping in the background on every invocation.
	// This is a no-op if already counted this week, or if opted out.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip the ping when the user is managing countme itself, to avoid
		// a ping firing just before an explicit --disable.
		if cmd.Name() != "countme" {
			go countme.Count(reportedVersion())
		}
		applyThemeFlavor()
		return nil
	},
	// If no subcommand is provided, open the interactive main menu by default.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Defer to the interactive menu
		if menuCmd != nil && menuCmd.RunE != nil {
			return menuCmd.RunE(menuCmd, nil)
		}
		// Fallback: show help if menu is not available for some reason
		return cmd.Help()
	},
}

// Execute runs the root command via fang (styled help/errors, manpage and
// completion commands, --version).
func Execute() error {
	return fang.Execute(context.Background(), rootCmd, fang.WithVersion(reportedVersion()))
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bluefin-cli version %s\n", reportedVersion()))
	status.AppVersion = reportedVersion()
}
