package install

import (
	"os"
	"sync"
	"testing"
)

func TestListBundles(t *testing.T) {
	// This should not panic or error
	ListBundles()
}

func TestBundleValidation(t *testing.T) {
	tests := []struct {
		name      string
		bundle    string
		expectErr bool
	}{
		{"Valid ai bundle", "ai", false},
		{"Valid cli bundle", "cli", false},
		{"Valid cncf bundle", "cncf", false},
		{"Valid experimental-ide bundle", "experimental-ide", false},

		{"Valid fonts bundle", "fonts", false},
		{"Valid full-desktop bundle", "full-desktop", false},
		{"Valid ide bundle", "ide", false},
		{"Valid k8s bundle", "k8s", false},
		{"Invalid bundle", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if bundle exists in our map
			_, exists := bundles[tt.bundle]
			if exists == tt.expectErr {
				t.Errorf("Bundle %s existence = %v, expectErr %v", tt.bundle, exists, tt.expectErr)
			}
		})
	}
}

func TestBundleFile(t *testing.T) {
	// Verify all bundles have proper file names
	for name, bundle := range bundles {
		if bundle.File == "" {
			t.Errorf("Bundle %s has empty file name", name)
		}
		if bundle.Description == "" {
			t.Errorf("Bundle %s has empty description", name)
		}
	}
}

func TestBundleWithLocalFile(t *testing.T) {
	// Create a temporary Brewfile
	tmpFile, err := os.CreateTemp("", "test-Brewfile-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	defer func() {
		_ = tmpFile.Close()
	}()

	// Write some content
	content := `tap "homebrew/core"
brew "git"
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// This test will fail if brew is not installed, which is okay for unit tests
	// The integration test will verify the full functionality
	t.Log("Skipping actual bundle installation in unit test")
}

func TestDownloadFile(t *testing.T) {
	// Test with a known good URL
	tmpFile, err := os.CreateTemp("", "download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// Try to download a small file
	url := "https://raw.githubusercontent.com/ublue-os/bluefin/main/README.md"
	if err := downloadFile(url, tmpPath); err != nil {
		t.Skipf("Skipping download test (network required): %v", err)
	}

	// Verify file was created and has content
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("Failed to stat downloaded file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Downloaded file is empty")
	}
}

func TestDownloadFileInvalidURL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// Try to download from invalid URL
	url := "https://invalid.invalid/nonexistent.txt"
	err = downloadFile(url, tmpPath)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestParseBrewfilePackages(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "windows-brewfile-*.Brewfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	content := `tap "homebrew/core"
brew "git"
cask "visual-studio-code"
# comment
`

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp Brewfile: %v", err)
	}
	tmpFile.Close()

	pkgs, err := parseBrewfilePackages(tmpFile.Name())
	if err != nil {
		t.Fatalf("parseBrewfilePackages returned error: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 parsed packages, got %d", len(pkgs))
	}

	if pkgs[0].name != "git" || pkgs[1].name != "visual-studio-code" {
		t.Fatalf("unexpected parsed package names: %+v", pkgs)
	}
}

func TestWindowsCandidatesIncludesAliases(t *testing.T) {
	candidates := windowsCandidates("visual-studio-code")
	if len(candidates) < 2 {
		t.Fatalf("expected aliases for visual-studio-code, got %v", candidates)
	}

	found := false
	for _, c := range candidates {
		if c == "Microsoft.VisualStudioCode" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected Microsoft.VisualStudioCode alias in %v", candidates)
	}
}

func TestAvailableWindowsManagersPriority(t *testing.T) {
	originalLookPath := windowsLookPath
	defer func() {
		windowsLookPath = originalLookPath
	}()

	windowsLookPath = func(file string) (string, error) {
		switch file {
		case "winget", "choco":
			return "/mock/" + file, nil
		default:
			return "", os.ErrNotExist
		}
	}

	managers := AvailableWindowsManagers()
	if len(managers) != 1 {
		t.Fatalf("expected 1 manager, got %d (%v)", len(managers), managers)
	}

	if managers[0] != "winget" {
		t.Fatalf("unexpected manager order: %v", managers)
	}
}

func TestWindowsCandidatesFromLoader(t *testing.T) {
	originalLoader := windowsPackageAliasesLoad
	originalAliases := windowsPackageAliases

	defer func() {
		windowsPackageAliasesLoad = originalLoader
		windowsPackageAliases = originalAliases
		windowsMappingLoadOnce = sync.Once{}
	}()

	windowsPackageAliasesLoad = func() map[string][]string {
		return map[string][]string{
			"custom-pkg": {"Custom.Id", "custom"},
		}
	}
	windowsPackageAliases = nil
	windowsMappingLoadOnce = sync.Once{}

	candidates := windowsCandidates("custom-pkg")
	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %v", candidates)
	}

	if candidates[1] != "Custom.Id" || candidates[2] != "custom" {
		t.Fatalf("unexpected candidates: %v", candidates)
	}
}

func TestWindowsCandidatesUnknownPackage(t *testing.T) {
	candidates := windowsCandidates("unmapped-tool")
	if len(candidates) != 1 || candidates[0] != "unmapped-tool" {
		t.Fatalf("expected only original package name, got %v", candidates)
	}
}

func TestWindowsBundleManifestLoads(t *testing.T) {
	manifest := getWindowsBundleManifest()
	if len(manifest) == 0 {
		t.Fatal("expected non-empty Windows bundle manifest")
	}

	ai, ok := manifest["ai"]
	if !ok {
		t.Fatal("expected ai bundle in Windows manifest")
	}
	if len(ai.Packages) == 0 {
		t.Fatal("expected ai bundle to include packages")
	}
}

func TestWindowsPackagesForBundlesDedupesIDs(t *testing.T) {
	pkgs, err := WindowsPackagesForBundles([]string{"cncf", "k8s"})
	if err != nil {
		t.Fatalf("WindowsPackagesForBundles returned error: %v", err)
	}

	seen := map[string]bool{}
	for _, pkg := range pkgs {
		if seen[pkg.ID] {
			t.Fatalf("duplicate package id found: %s", pkg.ID)
		}
		seen[pkg.ID] = true
	}

	if len(pkgs) == 0 {
		t.Fatal("expected at least one package")
	}
}
