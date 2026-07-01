package install

import (
	"testing"
)

func TestExtractQuotedName_Basic(t *testing.T) {
	got := extractQuotedName(`brew "git"`)
	if got != "git" {
		t.Errorf("extractQuotedName = %q, want %q", got, "git")
	}
}

func TestExtractQuotedName_Cask(t *testing.T) {
	got := extractQuotedName(`cask "visual-studio-code"`)
	if got != "visual-studio-code" {
		t.Errorf("extractQuotedName = %q, want %q", got, "visual-studio-code")
	}
}

func TestExtractQuotedName_Tap(t *testing.T) {
	got := extractQuotedName(`tap "homebrew/core"`)
	if got != "homebrew/core" {
		t.Errorf("extractQuotedName = %q, want %q", got, "homebrew/core")
	}
}

func TestExtractQuotedName_NoQuotes(t *testing.T) {
	got := extractQuotedName(`brew git`)
	if got != "" {
		t.Errorf("expected empty for line without quotes, got %q", got)
	}
}

func TestExtractQuotedName_EmptyString(t *testing.T) {
	got := extractQuotedName("")
	if got != "" {
		t.Errorf("expected empty for empty string, got %q", got)
	}
}

func TestExtractQuotedName_UnclosedQuote(t *testing.T) {
	got := extractQuotedName(`brew "git`)
	if got != "" {
		t.Errorf("expected empty for unclosed quote, got %q", got)
	}
}

func TestExtractQuotedName_MultipleQuotes(t *testing.T) {
	got := extractQuotedName(`brew "git" "extra"`)
	if got != "git" {
		t.Errorf("expected first quoted string, got %q", got)
	}
}

func TestParseBrewfileBytes_Empty(t *testing.T) {
	pkgs := parseBrewfileBytes([]byte{})
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages for empty input, got %d", len(pkgs))
	}
}

func TestParseBrewfileBytes_OnlyComments(t *testing.T) {
	data := []byte("# this is a comment\n# another one\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages for comments-only, got %d", len(pkgs))
	}
}

func TestParseBrewfileBytes_OnlyTapsAndFlatpaks(t *testing.T) {
	data := []byte("tap \"homebrew/core\"\nflatpak \"some.app\"\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages for taps/flatpaks only, got %d", len(pkgs))
	}
}

func TestParseBrewfileBytes_BrewAndCask(t *testing.T) {
	data := []byte("brew \"git\"\ncask \"visual-studio-code\"\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "git" {
		t.Errorf("pkgs[0].Name = %q, want %q", pkgs[0].Name, "git")
	}
	if pkgs[0].Kind != "brew" {
		t.Errorf("pkgs[0].Kind = %q, want %q", pkgs[0].Kind, "brew")
	}
	if pkgs[1].Name != "visual-studio-code" {
		t.Errorf("pkgs[1].Name = %q, want %q", pkgs[1].Name, "visual-studio-code")
	}
	if pkgs[1].Kind != "cask" {
		t.Errorf("pkgs[1].Kind = %q, want %q", pkgs[1].Kind, "cask")
	}
}

func TestParseBrewfileBytes_UnknownLine(t *testing.T) {
	data := []byte("brew \"git\"\nsomething \"else\"\ncask \"code\"\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (unknown lines skipped), got %d", len(pkgs))
	}
}

func TestParseBrewfileBytes_TapPrefixName(t *testing.T) {
	data := []byte("brew \"homebrew/core/git\"\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	// With a tap prefix, display ID should be the last segment
	if pkgs[0].Name != "git" {
		t.Errorf("pkgs[0].Name = %q, want %q (last path segment)", pkgs[0].Name, "git")
	}
	if pkgs[0].ID != "homebrew/core/git" {
		t.Errorf("pkgs[0].ID = %q, want %q", pkgs[0].ID, "homebrew/core/git")
	}
}

func TestParseBrewfileBytes_WhitespaceLines(t *testing.T) {
	data := []byte("brew \"git\"\n  \n\ncask \"code\"\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (whitespace lines skipped), got %d", len(pkgs))
	}
}

func TestParseBrewfileBytes_MixedContent(t *testing.T) {
	data := []byte("# Bluefin Brewfile\ntap \"homebrew/core\"\nbrew \"git\"\nbrew \"fish\"\ncask \"docker\"\nflatpak \"org.mozilla.firefox\"\n")
	pkgs := parseBrewfileBytes(data)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages (brew + cask only), got %d: %+v", len(pkgs), pkgs)
	}
}
