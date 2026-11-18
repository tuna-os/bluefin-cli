package shell

import (
	"runtime"
	"testing"
)

func TestLoginName(t *testing.T) {
	fallback := "bash"
	if runtime.GOOS == "windows" {
		fallback = "powershell"
	}
	cases := []struct {
		shell string
		want  string
	}{
		{"/usr/bin/zsh", "zsh"},
		{"/bin/bash", "bash"},
		{"fish", "fish"},
		{"", fallback},
	}
	for _, c := range cases {
		t.Setenv("SHELL", c.shell)
		if got := LoginName(); got != c.want {
			t.Errorf("SHELL=%q: LoginName() = %q, want %q", c.shell, got, c.want)
		}
	}
}

func TestProcessShellName(t *testing.T) {
	cases := map[string]string{
		"zsh\n":      "zsh",
		"-zsh\n":     "zsh",
		"/bin/fish":  "fish",
		"-/bin/bash": "bash",
		"  nu  \n":   "nu",
		"":           "",
		"\n":         "",
		"-":          "",
	}
	for in, want := range cases {
		if got := processShellName(in); got != want {
			t.Errorf("processShellName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPickActiveName(t *testing.T) {
	cases := []struct{ running, login, want string }{
		{"fish", "bash", "fish"},
		{"pwsh", "bash", "powershell"},
		{"", "bash", "bash"},
		{"sudo", "zsh", "zsh"},
	}
	for _, c := range cases {
		if got := pickActiveName(c.running, c.login); got != c.want {
			t.Errorf("pickActiveName(%q, %q) = %q, want %q", c.running, c.login, got, c.want)
		}
	}
}
