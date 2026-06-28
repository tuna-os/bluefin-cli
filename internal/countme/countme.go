// Package countme implements the Fedora countme protocol for bluefin-cli.
//
// This sends a single lightweight GET request to Fedora's metalink
// infrastructure at most once per week. It uses the standard libdnf
// User-Agent format so that bluefin-cli users appear in Fedora's public
// aggregate dataset (https://data-analysis.fedoraproject.org/) alongside
// native Bluefin Linux installs, broken down by platform variant.
//
// No personal data, IP address, or machine identifier is transmitted.
// The only information sent is:
//   - The binary name and version ("Bluefin 0.0.3")
//   - The platform variant ("mac", "wsl", "powershell")
//   - The OS and architecture ("Darwin.arm64", "Linux.x86_64", etc.)
//   - A coarse system age bucket (1–4, representing weeks of use)
//
// Opt out by setting BLUEFIN_DISABLE_COUNTME=1 in your environment,
// or by running: bluefin-cli countme --disable
package countme

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tuna-os/bluefin-cli/internal/env"
)

const (
	// Weekly window constants from the Fedora countme spec.
	// Windows are 604800-second (7-day) periods aligned to 1970-01-05 00:00:00 UTC.
	countmeWindow = 7 * 24 * 60 * 60 // 604800 seconds
	countmeOffset = 345600             // offset to 1970-01-05 (Monday)

	// Age bucket boundaries (number of windows elapsed since first count).
	// Bucket 1: first week, 2: first month, 3: first 6 months, 4: older.
	bucket2Threshold = 2
	bucket3Threshold = 5
	bucket4Threshold = 25

	// fedoraRelease is the Fedora version used in the metalink request URL.
	// Update this alongside Bluefin's base Fedora release.
	fedoraRelease = "42"

	stateFile    = "countme.json"
	optOutEnvVar = "BLUEFIN_DISABLE_COUNTME"
)

// State is persisted to disk to track epoch (first use) and last counted window.
type State struct {
	// Epoch is the Unix timestamp of the start of the first-ever counted week.
	// It is used to compute the system age bucket and never changes after first set.
	Epoch int64 `json:"epoch"`
	// Window is the Unix timestamp of the last successfully counted window.
	Window int64 `json:"window"`
	// Disabled is set to true when the user explicitly opts out via CLI.
	Disabled bool `json:"disabled,omitempty"`
}

// Count sends a countme ping if a new weekly window has started since the last
// ping. It is safe to call on every startup — it is a no-op unless needed.
// Network errors are silently swallowed; the state is only updated on success.
//
// This function is intended to be called in a goroutine so it does not block
// the CLI startup.
func Count(cliVersion string) {
	if os.Getenv(optOutEnvVar) != "" {
		return
	}

	configDir, err := env.GetConfigDir()
	if err != nil {
		return
	}

	statePath := filepath.Join(configDir, stateFile)

	state, err := loadState(statePath)
	if err != nil {
		return
	}

	if state.Disabled {
		return
	}

	curWindow := windowNumber(time.Now().Unix())

	// Initialise epoch on first run.
	if state.Epoch == 0 {
		state.Epoch = windowToUnix(curWindow)
	}

	epochWindow := windowNumber(state.Epoch)
	lastWindow := windowNumber(state.Window)

	// Already counted this window.
	if state.Window > 0 && lastWindow >= curWindow {
		return
	}

	bucket := ageBucket(epochWindow, curWindow)

	if err := sendPing(cliVersion, bucket); err != nil {
		// Network failure — leave state unchanged so we retry next run.
		return
	}

	state.Window = time.Now().Unix()
	_ = saveState(statePath, state)
}

// Disable writes a state file marking countme as disabled. This is the
// persistent opt-out used by "bluefin-cli countme --disable".
func Disable() error {
	configDir, err := env.EnsureConfigDir()
	if err != nil {
		return err
	}
	statePath := filepath.Join(configDir, stateFile)
	state, err := loadState(statePath)
	if err != nil {
		return err
	}
	state.Disabled = true
	return saveState(statePath, state)
}

// Enable removes the persistent opt-out.
func Enable() error {
	configDir, err := env.EnsureConfigDir()
	if err != nil {
		return err
	}
	statePath := filepath.Join(configDir, stateFile)
	state, err := loadState(statePath)
	if err != nil {
		return err
	}
	state.Disabled = false
	return saveState(statePath, state)
}

// windowNumber returns the countme window index for a Unix timestamp.
func windowNumber(unixTime int64) int64 {
	return (unixTime - countmeOffset) / countmeWindow
}

// windowToUnix converts a window index back to a Unix timestamp (start of window).
func windowToUnix(window int64) int64 {
	return window*countmeWindow + countmeOffset
}

// ageBucket returns the countme age bucket (1–4) given the epoch window and
// the current window.
func ageBucket(epochWindow, curWindow int64) int {
	step := curWindow - epochWindow
	switch {
	case step < bucket2Threshold:
		return 1
	case step < bucket3Threshold:
		return 2
	case step < bucket4Threshold:
		return 3
	default:
		return 4
	}
}

// variant returns the platform string embedded in the User-Agent.
func variant() string {
	switch {
	case runtime.GOOS == "darwin":
		return "mac"
	case env.IsWSL():
		return "wsl"
	case env.IsWindows():
		return "powershell"
	default:
		return "linux"
	}
}

// goosName returns the OS name component of the User-Agent.
func goosName() string {
	switch runtime.GOOS {
	case "darwin":
		return "Darwin"
	case "windows":
		return "Windows"
	default:
		return "Linux"
	}
}

// baseArch maps Go arch identifiers to the RPM base arch convention used
// in the countme User-Agent and metalink URL.
func baseArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}

// userAgent constructs the libdnf-format User-Agent string.
// Format: libdnf (NAME VERSION; VARIANT; OS.ARCH)
// Example: libdnf (Bluefin 0.0.3; mac; Darwin.aarch64)
func userAgent(cliVersion string) string {
	return fmt.Sprintf("libdnf (Bluefin %s; %s; %s.%s)",
		cliVersion, variant(), goosName(), baseArch())
}

// sendPing fires the countme GET request. The response body is discarded;
// the server-side infrastructure logs the request from the access log.
func sendPing(cliVersion string, bucket int) error {
	arch := baseArch()
	url := fmt.Sprintf(
		"https://mirrors.fedoraproject.org/metalink?repo=fedora-%s&arch=%s&countme=%d",
		fedoraRelease, arch, bucket,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent(cliVersion))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return nil
}

func loadState(path string) (State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		// Corrupt state file — start fresh rather than failing.
		return State{}, nil
	}
	return s, nil
}

func saveState(path string, s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	// Write atomically via a temp file in the same directory.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// StatusString returns a human-readable summary for "bluefin-cli countme --status".
func StatusString(cliVersion string) string {
	if os.Getenv(optOutEnvVar) != "" {
		return fmt.Sprintf("countme: disabled via %s environment variable", optOutEnvVar)
	}

	configDir, err := env.GetConfigDir()
	if err != nil {
		return "countme: unable to determine config directory"
	}

	state, err := loadState(filepath.Join(configDir, stateFile))
	if err != nil {
		return "countme: error reading state"
	}

	if state.Disabled {
		return "countme: disabled (run 'bluefin-cli countme --enable' to re-enable)"
	}

	var sb strings.Builder
	sb.WriteString("countme: enabled\n")
	fmt.Fprintf(&sb, "  user-agent : %s\n", userAgent(cliVersion))
	fmt.Fprintf(&sb, "  endpoint   : https://mirrors.fedoraproject.org/metalink?repo=fedora-%s&arch=%s\n",
		fedoraRelease, baseArch())
	if state.Epoch > 0 {
		fmt.Fprintf(&sb, "  first seen : %s\n", time.Unix(state.Epoch, 0).Format(time.DateOnly))
		fmt.Fprintf(&sb, "  age bucket : %d\n", ageBucket(windowNumber(state.Epoch), windowNumber(time.Now().Unix())))
	} else {
		sb.WriteString("  first seen : not yet counted\n")
	}
	if state.Window > 0 {
		fmt.Fprintf(&sb, "  last ping  : %s\n", time.Unix(state.Window, 0).Format(time.DateTime))
	} else {
		sb.WriteString("  last ping  : never\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}
