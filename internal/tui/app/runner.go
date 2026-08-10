package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// RunnerScreen executes a task natively inside the shell, capturing anything
// it prints to stdout/stderr into a scrolling log view — no terminal
// handover. While the task runs, back/quit keys are held; when it finishes,
// esc returns (or fires onDone if set).
//
// Capture works by swapping the os.Stdout/os.Stderr variables to a pipe for
// the duration of the task: the bubbletea renderer holds its own reference
// to the real terminal, so frames still reach the screen while fmt.Print
// output from the task lands in the log.
type RunnerScreen struct {
	title   string
	run     func() error
	onDone  func() tea.Cmd
	ch      chan runnerEvent
	lines   []string
	done    bool
	err     error
	started time.Time
	elapsed time.Duration
	spin    int
	scroll  int // manual scroll offset once done; tails while running
}

type runnerEvent struct {
	line string
	done bool
	err  error
}

type runnerMsg struct {
	s  *RunnerScreen
	ev runnerEvent
}

type runnerTickMsg struct{ s *RunnerScreen }

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// maxRunnerLines bounds the log buffer.
const maxRunnerLines = 500

// NewRunner creates a screen that runs the task when pushed.
func NewRunner(title string, run func() error) *RunnerScreen {
	return &RunnerScreen{title: title, run: run}
}

// NewRunnerWithPost is like NewRunner but fires onDone when the user
// dismisses the finished runner view (esc). While onDone is set the runner
// keeps capturing input even after the task completes so it can intercept
// the dismiss key and chain into the next screen.
func NewRunnerWithPost(title string, run func() error, onDone func() tea.Cmd) *RunnerScreen {
	return &RunnerScreen{title: title, run: run, onDone: onDone}
}

func (s *RunnerScreen) Title() string { return s.title }

// CapturingInput blocks navigation while the task is still running, and
// also while onDone is set (so the runner can intercept esc and chain).
func (s *RunnerScreen) CapturingInput() bool { return !s.done || s.onDone != nil }

func (s *RunnerScreen) Init() tea.Cmd {
	s.ch = make(chan runnerEvent, 64)
	s.started = time.Now()
	go s.exec()
	return tea.Batch(s.wait(), s.tick())
}

func (s *RunnerScreen) tick() tea.Cmd {
	return tea.Tick(time.Second/10, func(time.Time) tea.Msg { return runnerTickMsg{s: s} })
}

func (s *RunnerScreen) wait() tea.Cmd {
	return func() tea.Msg { return runnerMsg{s: s, ev: <-s.ch} }
}

func (s *RunnerScreen) exec() {
	r, w, err := os.Pipe()
	if err != nil {
		s.ch <- runnerEvent{done: true, err: s.run()}
		return
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w

	scanDone := make(chan struct{})
	go func() {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
		for sc.Scan() {
			s.ch <- runnerEvent{line: sc.Text()}
		}
		close(scanDone)
	}()

	runErr := s.run()
	os.Stdout, os.Stderr = oldOut, oldErr
	_ = w.Close()
	<-scanDone
	_ = r.Close()
	s.ch <- runnerEvent{done: true, err: runErr}
}

func (s *RunnerScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && s.done {
		switch key.String() {
		case "up", "k":
			s.scroll = max(s.scroll-1, 0)
		case "down", "j":
			s.scroll++
		case "g", "home":
			s.scroll = 0
		case "G", "end":
			s.scroll = 1 << 30
		case "esc":
			if s.onDone != nil {
				return s, tea.Sequence(s.onDone(), Pop())
			}
		}
		return s, nil
	}
	if t, ok := msg.(runnerTickMsg); ok && t.s == s {
		if s.done {
			return s, nil
		}
		s.spin++
		s.elapsed = time.Since(s.started)
		return s, s.tick()
	}
	m, ok := msg.(runnerMsg)
	if !ok || m.s != s {
		return s, nil
	}
	if m.ev.done {
		s.done, s.err = true, m.ev.err
		s.elapsed = time.Since(s.started)
		s.scroll = 1 << 30 // start at the tail; j/k adjust from there
		return s, nil
	}
	s.lines = append(s.lines, m.ev.line)
	if len(s.lines) > maxRunnerLines {
		s.lines = s.lines[len(s.lines)-maxRunnerLines:]
	}
	return s, s.wait()
}

func (s *RunnerScreen) View(width, height int) string {
	t := theme.DefaultTheme
	dim := lipgloss.NewStyle().Foreground(t.TextFaint)

	body := max(height-2, 3)
	start := max(len(s.lines)-body, 0)
	if s.done {
		s.scroll = max(min(s.scroll, len(s.lines)-body), 0)
		start = s.scroll
	}
	end := min(start+body, len(s.lines))
	out := make([]string, 0, body+2)
	out = append(out, "")
	for _, l := range s.lines[start:end] {
		out = append(out, " "+lipgloss.NewStyle().MaxWidth(max(width-2, 10)).Render(dim.Render(l)))
	}

	elapsed := ""
	if s.elapsed >= time.Second {
		elapsed = fmt.Sprintf("  %.0fs", s.elapsed.Seconds())
	}
	var status string
	switch {
	case !s.done:
		frame := spinnerFrames[s.spin%len(spinnerFrames)]
		status = lipgloss.NewStyle().Foreground(t.Info).Render(" "+frame+" working…") + dim.Render(elapsed)
	case s.err != nil:
		status = lipgloss.NewStyle().Foreground(t.Error).Render(" ✗ "+s.err.Error()) + dim.Render(elapsed)
	default:
		status = lipgloss.NewStyle().Foreground(t.Success).Render(" ✓ done") +
			dim.Render(elapsed+"  —  esc to go back")
	}
	out = append(out, status)
	return strings.Join(out, "\n")
}

func (s *RunnerScreen) KeyHints() []KeyHint {
	if s.done {
		return []KeyHint{{"jk", "scroll"}, {"esc", "back"}}
	}
	return []KeyHint{}
}
