package cmd

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/doctor"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common problems with your Bluefin CLI setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bench, _ := cmd.Flags().GetBool("bench"); bench {
			return runBench()
		}
		if fix, _ := cmd.Flags().GetBool("fix"); fix {
			runDoctorFixes()
		}
		report, failures := doctorReport()
		fmt.Println(report)
		if failures > 0 {
			return fmt.Errorf("%d check(s) failed", failures)
		}
		return nil
	},
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Apply safe automatic fixes before reporting")
	doctorCmd.Flags().Bool("bench", false, "Benchmark interactive shell startup vs a bare shell")
	rootCmd.AddCommand(doctorCmd)
	app.Register(app.Action{
		ID: "doctor", Icon: "🩺", Label: "Doctor", Section: "Home",
		Do: doctorScreenCmd,
	})
}

func defaultDoctorOptions() doctor.Options {
	return doctor.Options{
		CurrentShell: currentShellName(),
		Version:      version,
	}
}

// runDoctorFixes applies the remediations that are safe to automate:
// enabling shell integration and installing missing managed tools.
func runDoctorFixes() {
	doctor.RunFixes(doctor.FixOptions{
		CurrentShell: currentShellName(),
	})
}

// runBench measures interactive shell startup with the current rc files
// against a bare --norc baseline.
func runBench() error {
	res, err := doctor.RunBench(doctor.BenchOptions{
		Shell: currentShellName(),
	})
	if err != nil {
		return err
	}
	fmt.Print(doctor.RenderBench(res))
	return nil
}

// doctorReport runs all checks and renders the results; failures counts
// hard failures (warnings excluded).
func doctorReport() (string, int) {
	report := doctor.Diagnose(defaultDoctorOptions())
	return report.Render(), report.Failures
}

// doctorScreenCmd runs the doctor checks in a RunnerScreen (they hit the
// network, so they must not block the render loop).
func doctorScreenCmd() tea.Cmd {
	return app.Push(app.NewRunner("Doctor", func() error {
		report, _ := doctorReport()
		fmt.Println(report)
		return nil
	}))
}
