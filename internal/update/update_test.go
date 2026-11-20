package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"0.6.4", "v0.7.0", true},
		{"0.7.0", "v0.7.0", false},
		{"0.8.0", "v0.7.0", false},
		{"dev", "v99.0.0", false},
		{"v0.6.4", "0.7.0", true},
		{"garbage", "v0.7.0", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.latest); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

func checksumRelease(t *testing.T, body string) *Release {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	rel := &Release{TagName: "v1.0.0"}
	rel.Assets = append(rel.Assets, struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{Name: "checksums.txt", BrowserDownloadURL: srv.URL})
	return rel
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("archive-bytes")
	sum := sha256.Sum256(data)
	good := hex.EncodeToString(sum[:])

	rel := checksumRelease(t, good+"  app_1.0.0_linux_amd64.tar.gz\n")
	if err := verifyChecksum(rel, "app_1.0.0_linux_amd64.tar.gz", data); err != nil {
		t.Errorf("valid checksum rejected: %v", err)
	}

	rel = checksumRelease(t, good+"  app_1.0.0_linux_amd64.tar.gz\n")
	if err := verifyChecksum(rel, "app_1.0.0_linux_amd64.tar.gz", []byte("tampered")); err == nil {
		t.Error("tampered archive accepted")
	}

	rel = checksumRelease(t, good+"  some_other_asset.tar.gz\n")
	if err := verifyChecksum(rel, "app_1.0.0_linux_amd64.tar.gz", data); err == nil {
		t.Error("missing checksum entry accepted")
	}

	if err := verifyChecksum(&Release{TagName: "v1.0.0"}, "x.tar.gz", data); err == nil {
		t.Error("release without checksums.txt accepted")
	}
}

func TestDetectPath(t *testing.T) {
	cases := map[string]InstallMethod{
		"/usr/bin/bluefin-cli":                                  MethodSystem,
		"/usr/sbin/bluefin-cli":                                 MethodSystem,
		"/usr/local/bin/bluefin-cli":                            MethodDirect,
		"/home/me/.local/bin/bluefin-cli":                       MethodDirect,
		"/home/linuxbrew/.linuxbrew/Cellar/bluefin-cli/1/bin/x": MethodHomebrew,
		"/home/me/go/bin/bluefin-cli":                           MethodGoInstall,
	}
	for path, want := range cases {
		if got := detectPath(path); got != want {
			t.Errorf("detectPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestSystemUpdateHint(t *testing.T) {
	none := func(string) bool { return false }
	only := func(want string) func(string) bool { return func(s string) bool { return s == want } }

	if got := systemUpdateHint(only("/run/ostree-booted"), only("dnf")); !strings.HasPrefix(got, "sudo bootc upgrade") {
		t.Errorf("image-based system: hint = %q, want the image update", got)
	}
	if got := systemUpdateHint(none, only("dnf")); got != "sudo dnf upgrade bluefin-cli" {
		t.Errorf("dnf system: hint = %q", got)
	}
	if got := systemUpdateHint(none, only("apt-get")); got != "sudo apt-get install --only-upgrade bluefin-cli" {
		t.Errorf("apt system: hint = %q", got)
	}
	if got := systemUpdateHint(none, none); got == "" {
		t.Error("a system-owned binary must never fall through to self-update")
	}
}
