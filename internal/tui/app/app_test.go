package app

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	teatest "github.com/charmbracelet/x/exp/teatest/v2"
)

// ansiRE matches CSI sequences and OSC strings (BEL- or ST-terminated) so
// plain-text assertions survive styling like the gradient wordmark, which
// inserts escape codes between adjacent letters.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)

func stripAnsi(s string) string { return ansiRE.ReplaceAllString(s, "") }

func key(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func special(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

// testTree builds Home -> (Alpha -> [One, Two], Beta) for navigation tests.
func testTree() Screen {
	alpha := NewMenu("Alpha", []MenuItem{
		{Label: "One", Value: "one"},
		{Label: "Two", Value: "two"},
	}, nil, func(MenuItem) tea.Cmd { return Toast("picked", false) })

	return NewMenu("Home", []MenuItem{
		{Label: "Alpha", Value: "alpha", Submenu: true},
		{Label: "Beta", Value: "beta"},
	}, nil, func(it MenuItem) tea.Cmd {
		if it.Value == "alpha" {
			return Push(alpha)
		}
		return nil
	})
}

// drive feeds messages through Update synchronously, executing any returned
// commands inline and feeding their messages back. This is deterministic:
// the real program runs commands in goroutines, so message ordering there
// races with test input. Time-based commands (Toast expiry) will block for
// their duration, so drive is for navigation paths, not toast flows.
func drive(root Screen, msgs ...tea.Msg) Model {
	var m tea.Model = New(root)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, msg := range msgs {
		m = step(m, msg, 0)
	}
	return m.(Model)
}

func step(m tea.Model, msg tea.Msg, depth int) tea.Model {
	if msg == nil || depth > 8 {
		return m
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil {
				m = step(m, c(), depth+1)
			}
		}
		return m
	}
	var cmd tea.Cmd
	m, cmd = m.Update(msg)
	if cmd != nil {
		m = step(m, cmd(), depth+1)
	}
	return m
}

// TestViewRendersMenu checks the composed frame directly — header, items,
// footer — without going through a terminal.
func TestViewRendersMenu(t *testing.T) {
	var m tea.Model = New(testTree())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	frame := stripAnsi(m.(Model).View().Content)

	for _, want := range []string{"Bluefin CLI", "Home", "Alpha", "Beta", "move", "select", "filter", "help", "quit"} {
		if !strings.Contains(frame, want) {
			t.Errorf("rendered frame missing %q", want)
		}
	}
	if !strings.Contains(frame, "▎") {
		t.Error("rendered frame missing cursor accent bar")
	}
}

func TestDrillDownPushesScreen(t *testing.T) {
	m := drive(testTree(), special(tea.KeyEnter))
	if len(m.stack) != 2 {
		t.Fatalf("stack depth = %d, want 2", len(m.stack))
	}
	if got := m.top().Title(); got != "Alpha" {
		t.Errorf("top screen = %q, want Alpha", got)
	}
}

func TestEscapePopsBackHome(t *testing.T) {
	m := drive(testTree(), special(tea.KeyEnter), special(tea.KeyEscape))
	if len(m.stack) != 1 {
		t.Fatalf("stack depth = %d, want 1", len(m.stack))
	}
	if got := m.top().Title(); got != "Home" {
		t.Errorf("top screen = %q, want Home", got)
	}
}

func TestVimKeysMoveCursor(t *testing.T) {
	m := drive(testTree(), key('j'))
	menu := m.top().(*MenuScreen)
	if menu.cursor != 1 {
		t.Errorf("cursor = %d after j, want 1", menu.cursor)
	}
}

func TestFilterNarrowsItems(t *testing.T) {
	m := drive(testTree(), key('/'), key('b'), key('e'))
	menu := m.top().(*MenuScreen)
	if !menu.filtering || menu.query != "be" {
		t.Fatalf("filtering=%v query=%q, want active filter %q", menu.filtering, menu.query, "be")
	}
	vis := menu.visible()
	if len(vis) != 1 || menu.items[vis[0]].Label != "Beta" {
		t.Errorf("filter %q matched %d items, want only Beta", menu.query, len(vis))
	}
}

func TestHelpOverlayToggles(t *testing.T) {
	m := drive(testTree(), key('?'))
	if !m.showHelp {
		t.Error("help overlay not shown after ?")
	}
}

func TestPaletteOpensFromRegistry(t *testing.T) {
	orig := SnapshotRegistry()
	ResetRegistry()
	Register(Action{ID: "test-action", Label: "Do The Thing", Do: func() tea.Cmd { return nil }})
	t.Cleanup(func() { _ = RestoreRegistry(orig) })

	m := drive(testTree(), tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}, key('t'), key('h'))
	menu, ok := m.top().(*MenuScreen)
	if !ok || menu.Title() != "Palette" {
		t.Fatalf("top screen = %q, want Palette", m.top().Title())
	}
	if !menu.filtering || menu.query != "th" {
		t.Errorf("palette filter = %q (filtering=%v), want th", menu.query, menu.filtering)
	}
	vis := menu.visible()
	if len(vis) != 1 || menu.items[vis[0]].Label != "Do The Thing" {
		t.Errorf("palette should list the registered action, got %d matches", len(vis))
	}
}

func TestQuitAtRootViaEscape(t *testing.T) {
	tm := teatest.NewTestModel(t, New(testTree()), teatest.WithInitialTermSize(80, 24))
	tm.Send(special(tea.KeyEscape)) // pop at root quits the program
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
}

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		hay, needle string
		want        bool
	}{
		{"install apps", "sta", true},
		{"status", "sta", true},
		{"wallpapers", "wal", true},
		{"fonts", "wal", false},
		{"anything", "", true},
	}
	for _, c := range cases {
		if got := fuzzyMatch(c.hay, c.needle); got != c.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", c.hay, c.needle, got, c.want)
		}
	}
}
