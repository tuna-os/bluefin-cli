// Package update implements self-updating for script-installed binaries.
//
// Binaries installed through a package manager (Homebrew, Winget, Scoop)
// should be updated through that manager, so Detect reports the install
// method and Apply is only used for direct/script installs.
package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"golang.org/x/mod/semver"
)

const repo = "tuna-os/bluefin-cli"

// maxDownloadSize caps release archive downloads (largest archive is ~10MB).
const maxDownloadSize = 200 << 20

// InstallMethod describes how the running binary was installed.
type InstallMethod string

const (
	MethodDirect    InstallMethod = "direct"
	MethodHomebrew  InstallMethod = "homebrew"
	MethodScoop     InstallMethod = "scoop"
	MethodWinget    InstallMethod = "winget"
	MethodGoInstall InstallMethod = "go install"
	// MethodSystem is a binary the operating system owns: the deb or rpm
	// package, or a distribution image that ships bluefin-cli. Overwriting it
	// would fight the package database, and on an image-based system /usr is
	// read-only anyway.
	MethodSystem InstallMethod = "system package"
)

// UpdateHint returns the command a user should run to update for a
// package-manager-owned install, or "" when self-update is appropriate.
func (m InstallMethod) UpdateHint() string {
	switch m {
	case MethodHomebrew:
		return "brew upgrade bluefin-cli"
	case MethodScoop:
		return "scoop update bluefin-cli"
	case MethodWinget:
		return "winget upgrade Hanthor.BluefinCLI"
	case MethodGoInstall:
		return "go install github.com/tuna-os/bluefin-cli@latest"
	case MethodSystem:
		return systemUpdateHint(fileExists, lookPath)
	default:
		return ""
	}
}

var (
	fileExists = func(path string) bool { _, err := os.Stat(path); return err == nil }
	lookPath   = func(name string) bool { _, err := exec.LookPath(name); return err == nil }
)

// systemUpdateHint names the tool that updates a system-owned binary: the
// image update on an ostree or bootc system (/run/ostree-booted), otherwise
// the package manager that installed the deb or rpm.
func systemUpdateHint(exists, has func(string) bool) string {
	switch {
	case exists("/run/ostree-booted"):
		return "sudo bootc upgrade   (bluefin-cli ships with the system image)"
	case has("dnf"):
		return "sudo dnf upgrade bluefin-cli"
	case has("apt-get"):
		return "sudo apt-get install --only-upgrade bluefin-cli"
	default:
		return "update bluefin-cli with your system package manager"
	}
}

// Detect determines the install method from the running executable's path.
func Detect() InstallMethod {
	exe, err := os.Executable()
	if err != nil {
		return MethodDirect
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return detectPath(exe)
}

// detectPath maps an executable path to the install method that owns it.
func detectPath(exe string) InstallMethod {
	p := strings.ToLower(filepath.ToSlash(exe))
	switch {
	case strings.Contains(p, "/cellar/") || strings.Contains(p, "/linuxbrew/") || strings.Contains(p, "/homebrew/"):
		return MethodHomebrew
	case strings.Contains(p, "/scoop/"):
		return MethodScoop
	case strings.Contains(p, "/winget/") || strings.Contains(p, "/windowsapps/"):
		return MethodWinget
	case strings.Contains(p, "/go/bin/"):
		return MethodGoInstall
	// The deb and rpm packages install to /usr/bin. install.sh never does: it
	// uses ~/.local/bin, or /usr/local/bin as root, which stays self-updating.
	case strings.HasPrefix(p, "/usr/bin/") || strings.HasPrefix(p, "/usr/sbin/") ||
		strings.HasPrefix(p, "/bin/") || strings.HasPrefix(p, "/sbin/"):
		return MethodSystem
	default:
		return MethodDirect
	}
}

// Release is the subset of the GitHub release API response we need.
type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Latest fetches the latest release metadata from GitHub.
func Latest() (*Release, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checking for updates: GitHub API returned %s", resp.Status)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("parsing release info: %w", err)
	}
	return &rel, nil
}

// IsNewer reports whether latest (a "vX.Y.Z" tag) is newer than current.
// Dev builds ("dev") are never considered outdated.
func IsNewer(current, latest string) bool {
	if current == "dev" {
		return false
	}
	c := "v" + strings.TrimPrefix(current, "v")
	l := "v" + strings.TrimPrefix(latest, "v")
	if !semver.IsValid(c) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(l, c) > 0
}

// Apply downloads the release archive for this OS/arch, extracts the binary
// matching the running executable's name, and atomically replaces it.
func Apply(rel *Release, progress io.Writer) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	binName := filepath.Base(exe)
	binName = strings.TrimSuffix(binName, ".exe")

	version := strings.TrimPrefix(rel.TagName, "v")
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	// Archive name matches the binary: bluefin-cli_... or bluefin-cli-plus_...
	wanted := fmt.Sprintf("%s_%s_%s_%s.%s", binName, version, runtime.GOOS, runtime.GOARCH, ext)

	var url string
	for _, a := range rel.Assets {
		if a.Name == wanted {
			url = a.BrowserDownloadURL
			break
		}
	}
	if url == "" {
		return fmt.Errorf("no release asset %q for %s/%s", wanted, runtime.GOOS, runtime.GOARCH)
	}

	_, _ = fmt.Fprintf(progress, "Downloading %s...\n", wanted)
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading update: server returned %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize))
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	if err := verifyChecksum(rel, wanted, data); err != nil {
		return err
	}

	bin, err := extractBinary(data, ext, binName)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(progress, "Installing %s %s...\n", binName, rel.TagName)
	if err := selfupdate.Apply(bytes.NewReader(bin), selfupdate.Options{}); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("update failed and rollback failed (reinstall manually): %w", rerr)
		}
		return fmt.Errorf("update failed (rolled back): %w", err)
	}
	return nil
}

// verifyChecksum checks the downloaded archive against the release's
// checksums.txt (published by goreleaser). A release without a checksums
// asset is rejected: an attacker who can tamper with archives could also
// drop the checksum file, so its absence must not disable verification.
func verifyChecksum(rel *Release, assetName string, data []byte) error {
	var url string
	for _, a := range rel.Assets {
		if a.Name == "checksums.txt" {
			url = a.BrowserDownloadURL
			break
		}
	}
	if url == "" {
		return fmt.Errorf("release %s has no checksums.txt; refusing unverified update", rel.TagName)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("fetching checksums: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching checksums: server returned %s", resp.Status)
	}

	want := ""
	sc := bufio.NewScanner(io.LimitReader(resp.Body, 1<<20))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && fields[1] == assetName {
			want = strings.ToLower(fields[0])
			break
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("reading checksums: %w", err)
	}
	if want == "" {
		return fmt.Errorf("checksums.txt has no entry for %s; refusing unverified update", assetName)
	}

	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", assetName, got, want)
	}
	return nil
}

func extractBinary(archive []byte, ext, binName string) ([]byte, error) {
	target := binName
	if ext == "zip" {
		target += ".exe"
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, fmt.Errorf("reading update archive: %w", err)
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == target {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer func() { _ = rc.Close() }()
				return io.ReadAll(io.LimitReader(rc, maxDownloadSize))
			}
		}
		return nil, fmt.Errorf("%s not found in update archive", target)
	}

	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("reading update archive: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading update archive: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == target {
			return io.ReadAll(io.LimitReader(tr, maxDownloadSize))
		}
	}
	return nil, fmt.Errorf("%s not found in update archive", target)
}
