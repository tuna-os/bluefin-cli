package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const wallpapersTap = "ublue-os/tap"

const (
	wallpaperTapRawBasePrimary   = "https://raw.githubusercontent.com/ublue-os/homebrew-tap/main/Casks"
	wallpaperTapRawBaseFallback  = "https://raw.githubusercontent.com/ublue-os/tap/main/Casks"
	wallpaperUserInstallRootName = "BluefinCLI"
)

var (
	CaskURLLine     = regexp.MustCompile(`(?m)^\s*url\s+"([^"]+)"`)
	CaskVersionLine = regexp.MustCompile(`(?m)^\s*version\s+"([^"]+)"`)
)

var knownWallpaperCasks = []string{
	"bluefin-wallpapers",
	"aurora-wallpapers",
	"bazzite-wallpapers",
}

func EnsureBrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found. Please install Homebrew first: https://brew.sh")
	}
	return nil
}

func ensureTap(tap string) error {
	if err := EnsureBrew(); err != nil {
		return err
	}
	cmd := exec.Command("brew", "tap", tap)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetWallpaperCasks() ([]string, error) {
	if runtime.GOOS == "windows" {
		casks := append([]string{}, knownWallpaperCasks...)
		sort.Strings(casks)
		return casks, nil
	}

	if err := ensureTap(wallpapersTap); err != nil {
		return nil, err
	}

	cmd := exec.Command("brew", "--repository", wallpapersTap)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tap repository path: %w", err)
	}

	tapPath := strings.TrimSpace(string(out))
	casksDir := filepath.Join(tapPath, "Casks")

	entries, err := os.ReadDir(casksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read casks directory at %s: %w", casksDir, err)
	}

	var casks []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".rb") {
			caskName := strings.TrimSuffix(name, ".rb")
			if strings.Contains(strings.ToLower(caskName), "wallpaper") {
				casks = append(casks, caskName)
			}
		}
	}

	return casks, nil
}

func InstallWallpaperCasks(casks []string) error {
	return GetInstaller().InstallWallpapers(casks)
}

func postInstallWallpaperSetup(casks []string) {
	// Legacy function, no-op now.
}

func CleanupWallpapers(all bool) error {
	return GetInstaller().CleanupWallpapers(all)
}

func uninstallKnownWallpaperCasks() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	if err := ensureTap(wallpapersTap); err != nil {
		return err
	}

	args := []string{"uninstall", "--cask"}
	for _, cask := range knownWallpaperCasks {
		args = append(args, wallpapersTap+"/"+cask)
	}

	cmd := exec.Command("brew", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.ToLower(string(out))
		if strings.Contains(message, "is not installed") {
			return nil
		}
		return fmt.Errorf("failed to uninstall wallpaper casks: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

func removeKnownLinuxWallpaperDirs() error {
	if runtime.GOOS == "windows" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home directory: %w", err)
		}
		windowsDir := filepath.Join(homeDir, "Pictures", wallpaperUserInstallRootName)
		if err := os.RemoveAll(windowsDir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", windowsDir, err)
		}
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	dirs := []string{
		filepath.Join(homeDir, ".local", "share", "backgrounds", "bluefin"),
		filepath.Join(homeDir, ".local", "share", "backgrounds", "aurora"),
		filepath.Join(homeDir, ".local", "share", "backgrounds", "bazzite"),
		filepath.Join(homeDir, ".local", "share", "wallpapers", "bluefin"),
		filepath.Join(homeDir, ".local", "share", "wallpapers", "aurora"),
		filepath.Join(homeDir, ".local", "share", "wallpapers", "bazzite"),
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
	}

	return nil
}

func normalizeCaskName(cask string) string {
	parts := strings.Split(cask, "/")
	return parts[len(parts)-1]
}

func detectThemeName(cask string) (string, bool) {
	name := strings.ToLower(cask)
	switch {
	case strings.Contains(name, "bluefin"):
		return "Bluefin", true
	case strings.Contains(name, "aurora"):
		return "Aurora", true
	case strings.Contains(name, "bazzite"):
		return "Bazzite", true
	default:
		return "", false
	}
}

func ThemesFromWallpaperCasks(casks []string) []string {
	seen := map[string]struct{}{}
	themes := make([]string, 0, len(casks))

	for _, cask := range casks {
		normalized := normalizeCaskName(cask)
		themeName, ok := detectThemeName(normalized)
		if !ok {
			continue
		}
		if _, exists := seen[themeName]; exists {
			continue
		}
		seen[themeName] = struct{}{}
		themes = append(themes, themeName)
	}

	sort.Strings(themes)
	return themes
}

// fetchWallpaperArchiveURLFromTap fetches the cask file from the ublue-os tap on
// GitHub and extracts the direct download URL for the wallpaper archive.
func fetchWallpaperArchiveURLFromTap(caskName string) (string, error) {
	candidates := []string{
		fmt.Sprintf("%s/%s.rb", wallpaperTapRawBasePrimary, caskName),
		fmt.Sprintf("%s/%s.rb", wallpaperTapRawBaseFallback, caskName),
	}

	var lastErr error
	for _, url := range candidates {
		resp, err := http.Get(url) //nolint:gosec
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
			continue
		}
		matches := CaskURLLine.FindStringSubmatch(string(body))
		if len(matches) < 2 {
			lastErr = fmt.Errorf("no url field found in cask at %s", url)
			continue
		}
		return matches[1], nil
	}
	return "", fmt.Errorf("could not fetch cask for %q: %w", caskName, lastErr)
}

// downloadAndExtractWallpaperArchive downloads an archive from archiveURL and
// extracts its contents into targetDir, supporting .zip and .tar.gz formats.
func downloadAndExtractWallpaperArchive(archiveURL, targetDir string) error {
	resp, err := http.Get(archiveURL) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to download wallpaper archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP %d downloading %s", resp.StatusCode, archiveURL)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read archive: %w", err)
	}

	lower := strings.ToLower(archiveURL)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(data, targetDir)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(data, targetDir)
	default:
		return fmt.Errorf("unsupported archive format for URL: %s", archiveURL)
	}
}

func extractZip(data []byte, targetDir string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	for _, f := range r.File {
		if err := extractZipEntry(f, targetDir); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(f *zip.File, targetDir string) error {
	dest := filepath.Join(targetDir, filepath.FromSlash(f.Name))
	if !strings.HasPrefix(dest, filepath.Clean(targetDir)+string(os.PathSeparator)) {
		return fmt.Errorf("zip entry %q would escape target directory", f.Name)
	}
	if f.FileInfo().IsDir() {
		return os.MkdirAll(dest, 0755)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc) //nolint:gosec
	return err
}

func extractTarGz(data []byte, targetDir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to open gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}
		dest := filepath.Join(targetDir, filepath.FromSlash(hdr.Name))
		if !strings.HasPrefix(dest, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("tar entry %q would escape target directory", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				return err
			}
			out, err := os.Create(dest)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr) //nolint:gosec
			out.Close()
			if copyErr != nil {
				return copyErr
			}
		}
	}
	return nil
}

