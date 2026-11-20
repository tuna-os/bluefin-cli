package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripUblueBling(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"ublue-bling block": {
			in:   "alias k=kubectl\n### bling.sh source start\ntest -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh\n### bling.sh source end\nexport EDITOR=nano\n",
			want: "alias k=kubectl\nexport EDITOR=nano\n",
		},
		"fish block at end of file without newline": {
			in:   "set -x A 1\n### bling.fish source start\ntest -f /usr/share/ublue-os/bling/bling.fish && source /usr/share/ublue-os/bling/bling.fish\n### bling.fish source end",
			want: "set -x A 1\n",
		},
		"hand-written source line": {
			in:   "source /usr/share/ublue-os/bling/bling.sh\nalias x=y\n",
			want: "alias x=y\n",
		},
		"dot-sourced line": {
			in:   ". /usr/share/ublue-os/bling/bling.sh\n",
			want: "",
		},
		"unrelated rc is untouched": {
			in:   "source ~/.bling-of-my-own.sh\neval \"$(bluefin-cli init bash)\" # bluefin-cli shell-config\n",
			want: "source ~/.bling-of-my-own.sh\neval \"$(bluefin-cli init bash)\" # bluefin-cli shell-config\n",
		},
	}
	for name, c := range cases {
		if got := stripUblueBling(c.in); got != c.want {
			t.Errorf("%s: stripUblueBling() = %q, want %q", name, got, c.want)
		}
	}
}

// Turning the shell experience on replaces ublue-bling's block with the
// managed line; turning it off leaves neither.
func TestToggleReplacesUblueBling(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOMEBREW_PREFIX", "")
	// Every tool off, MOTD too, so "on" has nothing to install.
	off := Config{"motd": false}
	for _, tool := range Tools {
		off.SetEnabled(tool.Name, false)
	}
	if err := os.MkdirAll(filepath.Join(home, ".config", "bluefin-cli"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SaveConfig(&off); err != nil {
		t.Fatal(err)
	}

	rc := filepath.Join(home, ".bashrc")
	legacy := "### bling.sh source start\ntest -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh\n### bling.sh source end\n"
	if err := os.WriteFile(rc, []byte("alias k=kubectl\n"+legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Toggle("bash", true); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(rc)
	if strings.Contains(string(got), "ublue-os/bling") || !strings.Contains(string(got), shellMaker) {
		t.Fatalf("after on, rc = %q; want the managed line and no bling", got)
	}

	if err := os.WriteFile(rc, append(got, []byte(legacy)...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Toggle("bash", false); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(rc)
	if string(got) != "alias k=kubectl\n" {
		t.Fatalf("after off, rc = %q; want only the user's own line", got)
	}
}
