// Package app is the persistent TUI shell: one bubbletea program hosting a
// stack of screens with a shared header (breadcrumb + dino), a contextual
// footer, adaptive theming, toasts, and a help overlay.
//
// Navigation model: screens are pushed onto the stack to drill down and
// popped with esc/left/backspace (k9s-style). Everything renders natively;
// RunExternal exists only for genuinely external interactive programs.
package app

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/tuna-os/bluefin-cli/internal/tui/theme"
)

// KeyHint is a single entry in the contextual footer / help overlay.
type KeyHint struct {
	Keys string
	Desc string
}

// Screen is one level of the navigation stack.
type Screen interface {
	Title() string
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	// View renders the screen body within the given content area.
	View(width, height int) string
	KeyHints() []KeyHint
}

// InputCapturer is implemented by screens that are currently consuming text
// input (e.g. an active filter or an embedded form); while capturing, the
// shell passes keys through instead of treating them as navigation.
type InputCapturer interface {
	CapturingInput() bool
}

// Reloader is implemented by screens whose content should refresh when they
// become the top of the stack again (e.g. enabled/disabled labels).
type Reloader interface {
	Reload() tea.Cmd
}

// Messages for stack navigation and feedback.
type (
	PushMsg  struct{ Screen Screen }
	PopMsg   struct{}
	ToastMsg struct {
		Text  string
		IsErr bool
	}
	toastExpireMsg struct{}
	dinoTickMsg    struct{}
	reloadTopMsg   struct{}
)

// ReloadTop returns a command that refreshes the current screen's content
// (e.g. after an in-place toggle changed a label).
func ReloadTop() tea.Cmd {
	return func() tea.Msg { return reloadTopMsg{} }
}

// Push returns a command that pushes a screen onto the stack.
func Push(s Screen) tea.Cmd {
	return func() tea.Msg { return PushMsg{Screen: s} }
}

// Pop returns a command that pops the top screen.
func Pop() tea.Cmd {
	return func() tea.Msg { return PopMsg{} }
}

// Toast returns a command that shows a transient status message.
func Toast(text string, isErr bool) tea.Cmd {
	return func() tea.Msg { return ToastMsg{Text: text, IsErr: isErr} }
}

const (
	headerHeight = 5 // title row, three sky rows (dino), seafloor row
	footerHeight = 2 // blank + hints
	minBodyLines = 3
)

// toastDuration is how long a toast stays visible; a variable so tests can
// shrink it instead of sleeping.
var toastDuration = 4 * time.Second

// Model is the root shell model.
type Model struct {
	stack     []Screen
	width     int
	height    int
	theme     theme.Theme
	showHelp  bool
	toast     string
	toastErr  bool
	dino      dino
	blurred   bool
	quitting  bool
	extraInit []tea.Cmd
}

// New creates the shell with the given root screen. Any extra commands run
// alongside the root screen's Init (e.g. a background update check).
func New(root Screen, extraInit ...tea.Cmd) Model {
	return Model{
		stack:     []Screen{root},
		theme:     theme.DefaultTheme,
		extraInit: extraInit,
	}
}

// Run starts the shell program on the terminal.
func Run(root Screen, extraInit ...tea.Cmd) error {
	_, err := tea.NewProgram(New(root, extraInit...)).Run()
	return err
}

func (m Model) top() Screen { return m.stack[len(m.stack)-1] }

func (m Model) capturing() bool {
	if c, ok := m.top().(InputCapturer); ok {
		return c.CapturingInput()
	}
	return false
}

func (m Model) Init() tea.Cmd {
	cmds := append([]tea.Cmd{m.top().Init(), m.dino.tick(), tea.RequestBackgroundColor}, m.extraInit...)
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m.delegate(msg)

	case tea.BackgroundColorMsg:
		m.theme = theme.Resolve(isDarkColor(msg))
		theme.DefaultTheme = m.theme
		return m, nil

	case dinoTickMsg:
		if m.blurred {
			// Terminal lost focus: stop animating (and burning battery);
			// FocusMsg restarts the tick loop.
			return m, nil
		}
		m.dino.advance(m.width)
		return m, m.dino.tick()

	case tea.BlurMsg:
		m.blurred = true
		return m, nil

	case tea.FocusMsg:
		if m.blurred {
			m.blurred = false
			return m, m.dino.tick()
		}
		return m, nil

	case PushMsg:
		m.stack = append(m.stack, msg.Screen)
		m.showHelp = false
		m.dino.boost()
		return m, msg.Screen.Init()

	case PopMsg:
		if len(m.stack) <= 1 {
			m.quitting = true
			return m, tea.Quit
		}
		m.stack = m.stack[:len(m.stack)-1]
		if r, ok := m.top().(Reloader); ok {
			return m, r.Reload()
		}
		return m, nil

	case ToastMsg:
		m.toast, m.toastErr = msg.Text, msg.IsErr
		return m, tea.Tick(toastDuration, func(time.Time) tea.Msg { return toastExpireMsg{} })

	case toastExpireMsg:
		m.toast = ""
		return m, nil

	case reloadTopMsg:
		if r, ok := m.top().(Reloader); ok {
			return m, r.Reload()
		}
		return m, nil

	case ExternalDoneMsg:
		cmds := []tea.Cmd{}
		if r, ok := m.top().(Reloader); ok {
			cmds = append(cmds, r.Reload())
		}
		if msg.Err != nil {
			cmds = append(cmds, Toast("Error: "+msg.Err.Error(), true))
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if !m.capturing() {
			switch key {
			case "esc", "left", "backspace", "h":
				return m.Update(PopMsg{})
			case "q":
				m.quitting = true
				return m, tea.Quit
			case "?":
				m.showHelp = true
				return m, nil
			case "ctrl+p":
				if defaultRegistry.Count() > 0 {
					return m.Update(PushMsg{Screen: newPalette()})
				}
			}
		}
		return m.delegate(msg)
	}

	return m.delegate(msg)
}

func (m Model) delegate(msg tea.Msg) (tea.Model, tea.Cmd) {
	s, cmd := m.top().Update(msg)
	m.stack[len(m.stack)-1] = s
	return m, cmd
}

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.width == 0 {
		// Terminal size not reported yet; render for a conventional 80x24.
		m.width, m.height = 80, 24
	}
	t := m.theme

	header := m.renderHeader()
	footer := m.renderFooter()

	bodyHeight := max(m.height-headerHeight-footerHeight, minBodyLines)
	var body string
	if m.showHelp {
		body = m.renderHelp(bodyHeight)
	} else {
		body = m.top().View(m.width, bodyHeight)
	}
	body = lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(body)

	view := tea.NewView(strings.Join([]string{header, body, footer}, "\n"))
	view.BackgroundColor = t.Bg
	view.ReportFocus = true
	return view
}

func (m Model) renderHeader() string {
	t := m.theme
	crumbs := make([]string, 0, len(m.stack))
	for _, s := range m.stack {
		crumbs = append(crumbs, s.Title())
	}

	title := " " + gradient("Bluefin", t.Accent, t.Success) +
		lipgloss.NewStyle().Foreground(t.TextFaint).Render(" CLI")
	sep := lipgloss.NewStyle().Foreground(t.TextFaint).Render(" › ")
	crumb := lipgloss.NewStyle().Foreground(t.TextMuted).Render(strings.Join(crumbs, " › "))
	titleRow := title + sep + crumb

	// The dino occupies four rows: three sky lines for head, body, and
	// tail, and a full-width braille seafloor that doubles as the header
	// rule.
	return titleRow + "\n" + m.dino.renderGround(max(m.width, 1), t)
}

// gradient renders s with a per-rune color blend from one theme color to
// another — the wordmark treatment.
func gradient(s string, from, to color.Color) string {
	runes := []rune(s)
	cf, okf := colorful.MakeColor(from)
	ct, okt := colorful.MakeColor(to)
	if !okf || !okt || len(runes) == 0 {
		return lipgloss.NewStyle().Bold(true).Render(s)
	}
	var b strings.Builder
	for i, r := range runes {
		frac := 0.0
		if len(runes) > 1 {
			frac = float64(i) / float64(len(runes)-1)
		}
		c := cf.BlendLuv(ct, frac)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Bold(true).Render(string(r)))
	}
	return b.String()
}

func (m Model) renderFooter() string {
	t := m.theme
	if m.toast != "" {
		style := lipgloss.NewStyle().Foreground(t.Success)
		if m.toastErr {
			style = lipgloss.NewStyle().Foreground(t.Error)
		}
		return "\n " + style.Render(m.toast)
	}

	hints := m.top().KeyHints()
	hints = append(hints, globalHints(len(m.stack) > 1)...)

	keyStyle := lipgloss.NewStyle().Foreground(t.Accent)
	descStyle := lipgloss.NewStyle().Foreground(t.TextFaint)
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, keyStyle.Render(h.Keys)+" "+descStyle.Render(h.Desc))
	}
	line := " " + strings.Join(parts, descStyle.Render("  ·  "))
	return "\n" + lipgloss.NewStyle().MaxWidth(m.width).Render(line)
}

// helpSafe rewrites arrow glyphs to ASCII words inside the bordered help
// panel: ↑↓←→ are East-Asian-ambiguous width, and terminals that render
// them double-wide would break the panel's right border alignment.
var helpSafe = strings.NewReplacer("↑↓", "up/down", "→", "right", "←", "left")

func (m Model) renderHelp(height int) string {
	t := m.theme
	hints := append(m.top().KeyHints(), globalHints(len(m.stack) > 1)...)

	keyStyle := lipgloss.NewStyle().Foreground(t.Accent).Width(16)
	descStyle := lipgloss.NewStyle().Foreground(t.TextBase)
	rows := make([]string, 0, len(hints)+2)
	rows = append(rows, lipgloss.NewStyle().Foreground(t.TextMuted).Bold(true).Render("Keys"), "")
	for _, h := range hints {
		rows = append(rows, keyStyle.Render(helpSafe.Replace(h.Keys))+descStyle.Render(h.Desc))
	}

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Overlay).
		Padding(1, 3).
		Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, height, lipgloss.Center, lipgloss.Center, panel)
}

func globalHints(canGoBack bool) []KeyHint {
	hints := []KeyHint{}
	if canGoBack {
		hints = append(hints, KeyHint{"esc/←", "back"})
	}
	if defaultRegistry.Count() > 0 {
		hints = append(hints, KeyHint{"ctrl+p", "palette"})
	}
	hints = append(hints,
		KeyHint{"?", "help"},
		KeyHint{"q", "quit"},
	)
	return hints
}

// isDarkColor reports whether a terminal background color is dark, using
// relative luminance.
func isDarkColor(c color.Color) bool {
	if c == nil {
		return true
	}
	r, g, b, _ := c.RGBA()
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	return lum < 0.5*0xffff
}
