package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// --- RunnerScreen -------------------------------------------------------

// driveRunner pumps runner events synchronously until done or timeout.
func driveRunner(t *testing.T, r *RunnerScreen) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for !r.done {
		select {
		case ev := <-r.ch:
			r.Update(runnerMsg{s: r, ev: ev})
		case <-deadline:
			t.Fatal("runner did not finish in time")
		}
	}
}

func TestRunnerCapturesOutputAndError(t *testing.T) {
	r := NewRunner("test", func() error {
		fmt.Println("line one")
		fmt.Println("line two")
		return errors.New("boom")
	})
	if cmd := r.Init(); cmd == nil {
		t.Fatal("Init returned no command")
	}
	driveRunner(t, r)

	joined := strings.Join(r.lines, "\n")
	for _, want := range []string{"line one", "line two"} {
		if !strings.Contains(joined, want) {
			t.Errorf("captured log missing %q (got %q)", want, joined)
		}
	}
	if r.err == nil || r.err.Error() != "boom" {
		t.Errorf("err = %v, want boom", r.err)
	}
	if r.CapturingInput() {
		t.Error("finished runner should not capture input")
	}
	if view := stripAnsi(r.View(80, 20)); !strings.Contains(view, "boom") {
		t.Errorf("view missing error text: %q", view)
	}
}

func TestRunnerBlocksNavigationWhileRunning(t *testing.T) {
	release := make(chan struct{})
	r := NewRunner("test", func() error { <-release; return nil })
	r.Init()
	if !r.CapturingInput() {
		t.Error("running runner must capture input (blocks esc/q)")
	}
	close(release)
	driveRunner(t, r)
}

func TestRunnerWithPostCapturesInputWhenDone(t *testing.T) {
	release := make(chan struct{})
	r := NewRunnerWithPost("test", func() error { <-release; return nil }, func() tea.Cmd { return nil })
	r.Init()
	close(release)
	driveRunner(t, r)
	if !r.CapturingInput() {
		t.Error("runner with onDone must keep capturing input after done to intercept esc")
	}
}

func TestRunnerWithPostInterceptsEsc(t *testing.T) {
	called := false
	r := NewRunnerWithPost("test", func() error { return nil }, func() tea.Cmd {
		called = true
		return nil
	})
	r.Init()
	driveRunner(t, r)
	if !r.done {
		t.Fatal("runner did not finish")
	}
	// Simulate pressing esc on the finished runner.
	_, cmd := r.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc on runner with onDone should return a command")
	}
	if !called {
		t.Error("esc on finished runner with onDone did not call the callback")
	}
}

// --- TextScreen ---------------------------------------------------------

func TestTextScreenScrollsAndClamps(t *testing.T) {
	var lines []string
	for i := 1; i <= 30; i++ {
		lines = append(lines, fmt.Sprintf("row-%d", i))
	}
	s := NewText("test", func(int) (string, error) {
		return strings.Join(lines, "\n"), nil
	})

	view := s.View(80, 10)
	if !strings.Contains(view, "row-1") || strings.Contains(view, "row-20") {
		t.Errorf("initial view should start at the top: %q", view)
	}
	if !strings.Contains(stripAnsi(view), "of 30") {
		t.Error("long content should show a position indicator")
	}

	s.scroll = 1 << 30 // as set by G/end
	view = s.View(80, 10)
	if !strings.Contains(view, "row-30") {
		t.Errorf("end scroll should show last row: %q", view)
	}

	s.scroll = -5
	view = s.View(80, 10)
	if !strings.Contains(view, "row-1") {
		t.Errorf("negative scroll should clamp to top: %q", view)
	}
}

func TestTextScreenErrorRendered(t *testing.T) {
	s := NewText("test", func(int) (string, error) { return "", errors.New("fetch failed") })
	if view := stripAnsi(s.View(80, 10)); !strings.Contains(view, "fetch failed") {
		t.Errorf("view should surface fetch error: %q", view)
	}
}

// --- GameScreen ---------------------------------------------------------

func TestGameKelpCollisionGrounded(t *testing.T) {
	g := NewGame(0, nil)
	g.width = 80
	g.obstacles = []obstacle{{x: gameDinoX + 2}}
	if !g.collided() {
		t.Error("grounded dino should collide with kelp in its footprint")
	}
	g.air = gameJumpLen - 3 // mid-jump, lift 3
	if g.collided() {
		t.Error("airborne dino should clear kelp")
	}
}

func TestGameFishCollisionAirborne(t *testing.T) {
	g := NewGame(0, nil)
	g.width = 80
	g.obstacles = []obstacle{{x: gameDinoX + 2, fly: true}}
	if g.collided() {
		t.Error("grounded dino should pass under a fish")
	}
	g.air = gameJumpLen - 3
	if !g.collided() {
		t.Error("jumping into a fish should collide")
	}
}

func TestGameHighScorePersists(t *testing.T) {
	saved := -1
	g := NewGame(10, func(n int) { saved = n })
	g.width = 80
	g.score = 25
	g.obstacles = []obstacle{{x: gameDinoX + 2}}
	g.advanceGame()
	if !g.over {
		t.Fatal("game should be over after collision")
	}
	if saved != 26 { // advanceGame increments score first
		t.Errorf("saveBest got %d, want 26", saved)
	}

	// A worse run must not overwrite the best.
	saved = -1
	g.restart()
	g.score = 3
	g.obstacles = []obstacle{{x: gameDinoX + 2}}
	g.advanceGame()
	if saved != -1 {
		t.Errorf("saveBest should not fire for a lower score, got %d", saved)
	}
}
