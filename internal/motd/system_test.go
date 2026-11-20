package motd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// pointSystemPaths redirects the /etc and /usr/share sources into a temp dir.
func pointSystemPaths(t *testing.T) (etc, usr string) {
	t.Helper()
	root := t.TempDir()
	etc = filepath.Join(root, "etc")
	usr = filepath.Join(root, "usr")
	for _, dir := range []string{filepath.Join(etc, "tips"), filepath.Join(usr, "tips")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	savedTemplates, savedTips, savedInfo := systemTemplatePaths, systemTipsDirs, imageInfoPath
	t.Cleanup(func() { systemTemplatePaths, systemTipsDirs, imageInfoPath = savedTemplates, savedTips, savedInfo })
	systemTemplatePaths = []string{filepath.Join(etc, "motd.md"), filepath.Join(usr, "motd.md")}
	systemTipsDirs = []string{filepath.Join(etc, "tips"), filepath.Join(usr, "tips")}
	imageInfoPath = filepath.Join(root, "image-info.json")
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("HOMEBREW_PREFIX", "")
	return etc, usr
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRenderUsesTheDistributionTemplate(t *testing.T) {
	_, usr := pointSystemPaths(t)
	write(t, filepath.Join(usr, "motd.md"), "# Welcome to {{.ImageName}} ({{.ImageTag}})\n{{.Tip}}\n")
	write(t, filepath.Join(usr, "tips", "one.md"), "Run `ujust` for system tasks")
	write(t, imageInfoPath, `{"image-name": "utah", "image-tag": "testing"}`)

	got := Render()
	want := "# Welcome to utah (testing)\n"
	if runtime.GOOS != "linux" {
		// image-info.json is a Linux convention; elsewhere only the shape matters.
		want = "# Welcome to "
	}
	if !strings.HasPrefix(got, want) {
		t.Errorf("Render() = %q, want prefix %q", got, want)
	}
	if !strings.Contains(got, "Run `ujust` for system tasks") {
		t.Errorf("Render() = %q, want the /usr/share tip", got)
	}
}

func TestRenderEtcOverridesUsrShare(t *testing.T) {
	etc, usr := pointSystemPaths(t)
	write(t, filepath.Join(usr, "motd.md"), "distribution\n")
	write(t, filepath.Join(etc, "motd.md"), "administrator\n")
	write(t, filepath.Join(usr, "tips", "a.md"), "distribution tip")
	write(t, filepath.Join(etc, "tips", "a.md"), "administrator tip")

	if got := Render(); got != "administrator\n" {
		t.Errorf("Render() = %q, want /etc's template", got)
	}
	if got := pickTip(DefaultConfig()); !strings.Contains(got, "administrator tip") {
		t.Errorf("pickTip() = %q, want /etc's tips", got)
	}
}

func TestRenderSkipsABrokenTemplate(t *testing.T) {
	etc, usr := pointSystemPaths(t)
	write(t, filepath.Join(etc, "motd.md"), "{{.NoSuchField}}\n")
	write(t, filepath.Join(usr, "motd.md"), "fallback\n")

	if got := Render(); got != "fallback\n" {
		t.Errorf("Render() = %q, want the next template after a broken one", got)
	}
}

func TestRenderWithoutSystemFilesKeepsTheBuiltin(t *testing.T) {
	pointSystemPaths(t)
	if got := Render(); !strings.Contains(got, "Welcome to Bluefin CLI") {
		t.Errorf("Render() = %q, want the built-in template", got)
	}
}
