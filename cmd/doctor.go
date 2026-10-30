package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/tuna-os/bluefin-cli/internal/doctor"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

var doctorCmd = newDoctorCommand(func() doctor.Report {
	return doctor.Diagnose(defaultDoctorOptions())
})

func newDoctorCommand(diagnose func() doctor.Report) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose common problems with your Bluefin CLI setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			bench, _ := cmd.Flags().GetBool("bench")
			fix, _ := cmd.Flags().GetBool("fix")
			if jsonOutput && (bench || fix) {
				return fmt.Errorf("--json cannot be combined with --bench or --fix")
			}
			if bench {
				return runBench()
			}
			if fix {
				runDoctorFixes()
			}
			report := diagnose()
			if jsonOutput {
				if err := writeDoctorJSON(cmd.OutOrStdout(), report); err != nil {
					return fmt.Errorf("write doctor JSON: %w", err)
				}
			} else {
				fmt.Println(report.Render())
			}
			if report.Failures > 0 {
				return fmt.Errorf("%d check(s) failed", report.Failures)
			}
			return nil
		},
	}
}

// writeDoctorJSON writes a versioned envelope around the existing report.
func writeDoctorJSON(w io.Writer, report doctor.Report) error {
	data, err := json.Marshal(struct {
		SchemaVersion int `json:"schema_version"`
		doctor.Report
	}{SchemaVersion: 1, Report: report})
	if err != nil {
		return err
	}
	data = append(data, '\n')
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Apply safe automatic fixes before reporting")
	doctorCmd.Flags().Bool("bench", false, "Benchmark interactive shell startup vs a bare shell")
	doctorCmd.Flags().Bool("json", false, "Output a versioned JSON diagnostic report (incompatible with --fix and --bench)")
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
