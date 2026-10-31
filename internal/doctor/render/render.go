// Package render formats doctor reports for terminal output.
package render

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/doctor"
	"github.com/tuna-os/bluefin-cli/internal/tui"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// ReportText formats a diagnostic report with the application's terminal styles.
func ReportText(report doctor.Report) string {
	t := theme.DefaultTheme
	pass := lipgloss.NewStyle().Foreground(t.Success).Render("✓")
	warn := lipgloss.NewStyle().Foreground(t.Warning).Render("!")
	fail := lipgloss.NewStyle().Foreground(t.Error).Render("✗")
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	var b strings.Builder
	b.WriteString("\n")
	for _, check := range report.Checks {
		mark := pass
		switch {
		case !check.OK && check.Warn:
			mark = warn
		case !check.OK:
			mark = fail
		}
		fmt.Fprintf(&b, "  %s %s\n", mark, check.Name)
		if check.Note != "" {
			fmt.Fprintf(&b, "    %s\n", dim.Render(check.Note))
		}
	}
	if report.Failures == 0 {
		b.WriteString("\n" + dim.Render("  All checks passed."))
	}
	return b.String()
}

// BenchText formats shell startup benchmark results with the application's
// terminal styles.
func BenchText(result *doctor.BenchResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Benchmarking %s startup (median of 7 runs)...\n", result.Shell)
	fmt.Fprintf(&b, "  your config:  %s\n", result.Full.Round(time.Millisecond))
	fmt.Fprintf(&b, "  bare shell:   %s\n", result.Base.Round(time.Millisecond))
	fmt.Fprintf(&b, "  rc overhead:  %s\n", result.Overhead.Round(time.Millisecond))
	if result.Overhead > 300*time.Millisecond {
		b.WriteString(tui.WarningStyle.Render("  Startup overhead is noticeable — try disabling components in the Shell menu.") + "\n")
	} else {
		b.WriteString(tui.SuccessStyle.Render("  Snappy.") + "\n")
	}
	return b.String()
}
