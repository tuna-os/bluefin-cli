package sunset

import (
	"errors"
	"testing"
)

func withLookPath(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	original := lookPath
	lookPath = fn
	t.Cleanup(func() { lookPath = original })
}

func TestFindWindowsCLI_FoundOnPath(t *testing.T) {
	withLookPath(t, func(file string) (string, error) {
		if file == "bluefin-cli.exe" {
			return "/mnt/c/tools/bluefin-cli.exe", nil
		}
		return "", errors.New("not found")
	})

	got := FindWindowsCLI()
	if got != "/mnt/c/tools/bluefin-cli.exe" {
		t.Errorf("FindWindowsCLI() = %q, want path from PATH lookup", got)
	}
}

func TestFindWindowsCLI_NotFoundAnywhere(t *testing.T) {
	withLookPath(t, func(file string) (string, error) {
		return "", errors.New("not found")
	})

	original := getenvVar
	getenvVar = func(key string) string { return "" }
	t.Cleanup(func() { getenvVar = original })

	got := FindWindowsCLI()
	if got != "" {
		t.Errorf("FindWindowsCLI() = %q, want empty when PATH and LOCALAPPDATA both miss", got)
	}
}

func TestDelegateToWindowsCLI_ReportsWhenMissing(t *testing.T) {
	withLookPath(t, func(file string) (string, error) {
		return "", errors.New("not found")
	})
	originalGetenv := getenvVar
	getenvVar = func(key string) string { return "" }
	t.Cleanup(func() { getenvVar = originalGetenv })

	r := &recordingReporter{}
	err := DelegateToWindowsCLI([]string{"sunset"}, RunnerIO{}, r)
	if err != nil {
		t.Fatalf("DelegateToWindowsCLI() = %v, want nil when no Windows CLI is found", err)
	}
	if len(r.infos) == 0 {
		t.Errorf("expected an informational message when no Windows CLI is found")
	}
}
