package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tuna-os/bluefin-cli/internal/doctor"
)

func TestDoctorJSONSubprocess(t *testing.T) {
	// All current probes report warnings rather than hard failures. Inject a
	// report only in this subprocess, while exercising the real Cobra/Fang
	// command and its process exit behavior.
	if scenario := os.Getenv("DOCTOR_JSON_TEST_REPORT"); scenario != "" {
		report := doctor.Report{Checks: []doctor.CheckResult{{OK: true, Name: "healthy"}}}
		switch scenario {
		case "warning":
			report.Checks = []doctor.CheckResult{{Warn: true, Name: "warning", Note: "hint"}}
		case "failure":
			report.Checks = []doctor.CheckResult{{Name: "failed", Note: "detail"}}
			report.Failures = 1
		}
		doctorCmd.RunE = newDoctorCommand(func() doctor.Report { return report }).RunE
		args := []string{"doctor", "--json"}
		switch scenario {
		case "human":
			args = []string{"doctor"}
		case "fix", "bench":
			args = append(args, "--"+scenario)
		}
		rootCmd.SetArgs(args)
		if err := Execute(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	for _, scenario := range []string{"healthy", "warning", "failure"} {
		t.Run(scenario, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestDoctorJSONSubprocess$")
			command.Env = append(os.Environ(), "DOCTOR_JSON_TEST_REPORT="+scenario, "BLUEFIN_DISABLE_COUNTME=1")
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			if (err != nil) != (scenario == "failure") {
				t.Fatalf("exit error = %v, stderr = %s", err, stderr.String())
			}
			if scenario == "failure" && !strings.Contains(stderr.String(), "1 check(s) failed") {
				t.Fatalf("missing failure error on stderr: %q", stderr.String())
			}
			if bytes.Contains(stdout.Bytes(), []byte("\x1b")) {
				t.Fatalf("ANSI escape in stdout: %q", stdout.String())
			}
			var got struct {
				SchemaVersion int `json:"schema_version"`
				doctor.Report
			}
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON %q: %v", stdout.String(), err)
			}
			if got.SchemaVersion != 1 || len(got.Checks) != 1 {
				t.Fatalf("unexpected report: %+v", got)
			}
			check := got.Checks[0]
			if scenario == "healthy" && (!check.OK || got.Failures != 0) ||
				scenario == "warning" && (!check.Warn || check.OK || got.Failures != 0 || check.Note != "hint") ||
				scenario == "failure" && (check.OK || check.Warn || got.Failures != 1 || check.Note != "detail") {
				t.Fatalf("unexpected report: %+v", got)
			}
		})
	}
}

func TestDoctorHumanAndFlagErrorsSubprocess(t *testing.T) {
	for _, scenario := range []string{"human", "fix", "bench"} {
		t.Run(scenario, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestDoctorJSONSubprocess$")
			command.Env = append(os.Environ(), "DOCTOR_JSON_TEST_REPORT="+scenario, "BLUEFIN_DISABLE_COUNTME=1")
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			if scenario == "human" {
				if err != nil || !strings.Contains(stdout.String(), "healthy") || !strings.Contains(stdout.String(), "All checks passed.") || json.Valid(stdout.Bytes()) {
					t.Fatalf("human output: exit=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
				}
			} else if err == nil || stdout.Len() != 0 || !strings.Contains(stderr.String(), "cannot be combined") {
				t.Fatalf("invalid flags: exit=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
			}
		})
	}
}

func TestDoctorJSONSchema(t *testing.T) {
	var out bytes.Buffer
	report := doctor.Report{Checks: []doctor.CheckResult{
		{OK: true, Name: "passed"},
		{Warn: true, Name: "warning", Note: "escape: \x1b"},
	}}
	if err := writeDoctorJSON(&out, report); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out.Bytes(), []byte("\x1b")) {
		t.Fatalf("unescaped control character: %q", out.String())
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if string(got["schema_version"]) != "1" || string(got["failures"]) != "0" {
		t.Fatalf("unexpected envelope: %s", out.String())
	}
	var checks []map[string]json.RawMessage
	if err := json.Unmarshal(got["checks"], &checks); err != nil {
		t.Fatal(err)
	}
	if len(checks) != 2 || len(checks[0]) != 3 || len(checks[1]) != 4 || checks[0]["note"] != nil || checks[1]["note"] == nil {
		t.Fatalf("unexpected checks: %s", got["checks"])
	}
}

type shortDoctorWriter struct{}

func (shortDoctorWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }

type failingDoctorWriter struct{}

func (failingDoctorWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestDoctorJSONWriteErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		writer io.Writer
		want   error
	}{
		{"short write", shortDoctorWriter{}, io.ErrShortWrite},
		{"write error", failingDoctorWriter{}, io.ErrClosedPipe},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newDoctorCommand(func() doctor.Report { return doctor.Report{} })
			cmd.Flags().Bool("json", true, "")
			cmd.Flags().Bool("bench", false, "")
			cmd.Flags().Bool("fix", false, "")
			cmd.SetOut(tc.writer)
			if err := cmd.RunE(cmd, nil); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDoctorJSONRejectsIncompatibleFlags(t *testing.T) {
	for _, flag := range []string{"fix", "bench"} {
		t.Run(flag, func(t *testing.T) {
			called := false
			cmd := newDoctorCommand(func() doctor.Report { called = true; return doctor.Report{} })
			cmd.Flags().Bool("json", true, "")
			cmd.Flags().Bool("fix", flag == "fix", "")
			cmd.Flags().Bool("bench", flag == "bench", "")
			var out bytes.Buffer
			cmd.SetOut(&out)
			err := cmd.RunE(cmd, nil)
			if err == nil || !strings.Contains(err.Error(), "--json cannot be combined") || called || out.Len() != 0 {
				t.Fatalf("error=%v, diagnosed=%v, stdout=%q", err, called, out.String())
			}
		})
	}
}
