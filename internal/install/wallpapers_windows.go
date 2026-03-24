//go:build windows

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
	"path/filepath"
	"regexp"
	"strings"
)

var (
	wallpaperTapRawBasePrimary  = "https://raw.githubusercontent.com/ublue-os/homebrew-tap/main/Casks"
	wallpaperTapRawBaseFallback = "https://raw.githubusercontent.com/ublue-os/tap/main/Casks"
	caskURLLine                 = regexp.MustCompile(`(?m)^\s*url\s+"([^"]+)"`)
)

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
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil || resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
			continue
		}
		matches := caskURLLine.FindStringSubmatch(string(body))
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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = rc.Close() }()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, rc) //nolint:gosec
	return err
}

func extractTarGz(data []byte, targetDir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to open gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

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
			if err := out.Close(); err != nil && copyErr == nil {
				return err
			}
			if copyErr != nil {
				return copyErr
			}
		}
	}
	return nil
}
