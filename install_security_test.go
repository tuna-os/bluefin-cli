package main

import (
	"os"
	"strings"
	"testing"
)

func TestShellInstallersRequireChecksums(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"install.sh", "src/bluefin-cli/install.sh"} {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			script := string(data)
			for _, required := range []string{
				"checksums.txt",
				"expected_sha=",
				"${#expected_sha}",
				"sha256sum",
				"shasum -a 256",
				"actual_sha",
				"checksum verification failed",
			} {
				if !strings.Contains(script, required) {
					t.Errorf("installer does not enforce %q", required)
				}
			}
		})
	}
}

func TestPowerShellInstallerRequiresChecksum(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("install.ps1")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, required := range []string{
		"checksums.txt",
		"$expectedSha",
		"[0-9A-Fa-f]{64}",
		"Get-FileHash",
		"-Algorithm SHA256",
		"$actualSha",
		"checksum verification failed",
	} {
		if !strings.Contains(script, required) {
			t.Errorf("install.ps1 does not enforce %q", required)
		}
	}

	verify := strings.Index(script, "checksum verification failed")
	extract := strings.Index(script, "Expand-Archive")
	if verify < 0 || extract < 0 || verify > extract {
		t.Error("install.ps1 must verify the checksum before Expand-Archive")
	}
}
