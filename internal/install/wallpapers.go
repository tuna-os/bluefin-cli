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
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hanthor/bluefin-cli/internal/env"
	"github.com/klauspost/compress/zstd"
)

const wallpapersTap = "ublue-os/tap"

const (
	wallpaperTapRawBasePrimary   = "https://raw.githubusercontent.com/ublue-os/homebrew-tap/main/Casks"
	wallpaperTapRawBaseFallback  = "https://raw.githubusercontent.com/ublue-os/tap/main/Casks"
	wallpaperUserInstallRootName = "BluefinCLI"
)

var (
	caskURLLine     = regexp.MustCompile(`(?m)^\s*url\s+"([^"]+)"`)
	caskVersionLine = regexp.MustCompile(`(?m)^\s*version\s+"([^"]+)"`)
)

var knownWallpaperCasks = []string{
	"bluefin-wallpapers",
	"aurora-wallpapers",
	"bazzite-wallpapers",
}

var (
	isWSL     = env.IsWSL
	isWindows = env.IsWindows
)

func EnsureBrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install Homebrew first: https://brew.sh")
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
	if runtime.GOOS == "windows" {
		return installWallpaperCasksWindows(casks)
	}

	if err := ensureTap(wallpapersTap); err != nil {
		return err
	}
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks selected")
	}
	args := []string{"install", "--cask"}
	for _, c := range casks {
		if strings.Contains(c, "/") {
			args = append(args, c)
		} else {
			args = append(args, wallpapersTap+"/"+c)
		}
	}
	cmd := exec.Command("brew", args...)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_ENV_HINTS=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install wallpaper casks: %w", err)
	}
	fmt.Println(successStyle.Render("✓ Wallpaper casks installed!"))

	postInstallWallpaperSetup(casks)

	// macOS specific instructions
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		fmt.Println("\n" + infoStyle.Render("Wallpapers installed to: "+filepath.Join(home, "Library/Desktop Pictures")))
		fmt.Println(infoStyle.Render("To use: System Settings > Wallpaper > Add Folder"))
	}

	return nil
}

func postInstallWallpaperSetup(casks []string) {
	// Legacy function, no-op now.
}

func CleanupWallpapers(all bool) error {
	if !all {
		return nil
	}

	if err := uninstallKnownWallpaperCasks(); err != nil {
		return err
	}

	if err := removeKnownLinuxWallpaperDirs(); err != nil {
		return err
	}

	return nil
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

func installWallpaperCasksWindows(casks []string) error {
	if len(casks) == 0 {
		return fmt.Errorf("no wallpaper casks selected")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}

	installRoot := filepath.Join(homeDir, "Pictures", wallpaperUserInstallRootName)
	if err := os.MkdirAll(installRoot, 0755); err != nil {
		return fmt.Errorf("failed to create wallpaper install directory: %w", err)
	}

	successChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	for _, c := range casks {
		normalized := normalizeCaskName(c)
		targetDir := ""

		done := make(chan struct{})
		go func() {
			defer close(done)
			archiveURL, err := fetchWallpaperArchiveURLFromTap(normalized)
			if err != nil {
				errChan <- err
				return
			}

			targetDirName := normalized
			if themeName, ok := detectThemeName(normalized); ok {
				targetDirName = strings.ToLower(themeName)
			}
			targetDir = filepath.Join(installRoot, targetDirName)

			if err := downloadAndExtractWallpaperArchive(archiveURL, targetDir); err != nil {
				errChan <- err
				return
			}
			successChan <- true
		}()

		// Throbber
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
				fmt.Printf("\r%s %s...", infoStyle.Foreground(lipgloss.Color("13")).Render(frames[i%len(frames)]), infoStyle.Render("Installing "+normalized))
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}

		select {
		case err := <-errChan:
			return fmt.Errorf("failed to install %s: %w", normalized, err)
		case <-successChan:
			fmt.Printf("\r\033[K") // Clear throbber line
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Installed %s wallpapers to %s", normalized, targetDir)))
		}
	}

	fmt.Println(infoStyle.Render("Wallpaper files downloaded without Homebrew."))
	fmt.Println(infoStyle.Render("You can apply them from Windows Settings > Personalization > Background."))
	return nil
}

func fetchWallpaperArchiveURLFromTap(cask string) (string, error) {
	cask = normalizeCaskName(cask)

	candidates := []string{
		fmt.Sprintf("%s/%s.rb", wallpaperTapRawBasePrimary, cask),
		fmt.Sprintf("%s/%s.rb", wallpaperTapRawBaseFallback, cask),
	}

	for _, rawURL := range candidates {
		resp, err := http.Get(rawURL)
		if err != nil {
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			continue
		}

		resolved := resolveWallpaperArchiveURLFromCask(body)
		if strings.TrimSpace(resolved) != "" {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("failed to resolve wallpaper archive URL for %s from tap metadata", cask)
}

func resolveWallpaperArchiveURLFromCask(body []byte) string {
	content := string(body)

	version := ""
	if m := caskVersionLine.FindStringSubmatch(content); len(m) == 2 {
		version = strings.TrimSpace(m[1])
	}

	urlMatches := caskURLLine.FindAllStringSubmatch(content, -1)
	if len(urlMatches) == 0 {
		return ""
	}

	urls := make([]string, 0, len(urlMatches))
	for _, m := range urlMatches {
		if len(m) != 2 {
			continue
		}
		u := strings.TrimSpace(m[1])
		if u == "" {
			continue
		}
		if version != "" {
			u = strings.ReplaceAll(u, "#{version}", version)
		}
		urls = append(urls, u)
	}

	if len(urls) == 0 {
		return ""
	}

	// Prefer variants based on platform.
	// For Windows, PNG is most compatible.
	if runtime.GOOS == "windows" {
		for _, u := range urls {
			if strings.Contains(strings.ToLower(u), "-png") {
				return u
			}
		}
	}

	// General preference: PNG/Gnome/Neutral over MacOS/KDE specific ones.
	for _, u := range urls {
		l := strings.ToLower(u)
		if strings.Contains(l, "wallpaper") && (strings.Contains(l, "png") || strings.Contains(l, "gnome")) {
			return u
		}
	}
	for _, u := range urls {
		l := strings.ToLower(u)
		if strings.Contains(l, "wallpaper") && !strings.Contains(l, "macos") && !strings.Contains(l, "kde") {
			return u
		}
	}

	return urls[0]
}

func downloadAndExtractWallpaperArchive(archiveURL, targetDir string) error {
	resp, err := http.Get(archiveURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	lowerURL := strings.ToLower(archiveURL)
	switch {
	case strings.HasSuffix(lowerURL, ".zip"):
		return extractImagesFromZip(payload, targetDir)
	case strings.HasSuffix(lowerURL, ".tar.zstd") || strings.HasSuffix(lowerURL, ".tar.zst"):
		if err := extractImagesFromTarZstd(payload, targetDir); err == nil {
			return nil
		}
		// Fallback to plain tar if zstd magic number mismatch
		return extractImagesFromTar(payload, targetDir)
	case strings.HasSuffix(lowerURL, ".tar.gz") || strings.HasSuffix(lowerURL, ".tgz"):
		if err := extractImagesFromTarGz(payload, targetDir); err == nil {
			return nil
		}
		// Fallback to plain tar
		return extractImagesFromTar(payload, targetDir)
	default:
		// Try common archive types for tap assets where extension can be omitted.
		if err := extractImagesFromZip(payload, targetDir); err == nil {
			return nil
		}
		if err := extractImagesFromTarZstd(payload, targetDir); err == nil {
			return nil
		}
		if err := extractImagesFromTarGz(payload, targetDir); err == nil {
			return nil
		}
		return extractImagesFromTar(payload, targetDir)
	}
}

func extractImagesFromZip(payload []byte, targetDir string) error {
	r, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}

	written := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() || !isImageFileName(f.Name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		name := filepath.Base(f.Name)
		outPath := filepath.Join(targetDir, name)
		if err := writeFileFromReader(outPath, rc); err == nil {
			written++
		}
		_ = rc.Close()
	}

	if written == 0 {
		return fmt.Errorf("no wallpaper image files found in archive")
	}

	return nil
}

func extractImagesFromTarGz(payload []byte, targetDir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	return extractImagesFromTarReader(tr, targetDir)
}

func extractImagesFromTarZstd(payload []byte, targetDir string) error {
	zr, err := zstd.NewReader(bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	return extractImagesFromTarReader(tr, targetDir)
}

func extractImagesFromTar(payload []byte, targetDir string) error {
	tr := tar.NewReader(bytes.NewReader(payload))
	return extractImagesFromTarReader(tr, targetDir)
}

func extractImagesFromTarReader(tr *tar.Reader, targetDir string) error {
	written := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if (h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA) || !isImageFileName(h.Name) {
			continue
		}

		name := filepath.Base(h.Name)
		outPath := filepath.Join(targetDir, name)
		if err := writeFileFromReader(outPath, tr); err == nil {
			written++
		}
	}

	if written == 0 {
		return fmt.Errorf("no wallpaper image files found in archive")
	}

	return nil
}

func isImageFileName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	return strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".bmp") || strings.HasSuffix(name, ".webp") || strings.HasSuffix(name, ".jxl") || strings.HasSuffix(name, ".heic") || strings.HasSuffix(name, ".avif")
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

func writeFileFromReader(path string, r io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}
