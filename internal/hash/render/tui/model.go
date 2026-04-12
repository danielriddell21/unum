// Package tui provides a Bubble Tea TUI for the hash derivation tool.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/hash/types"
	"github.com/danielriddell21/unum/internal/tui/panels"
)


// deriveFunc and historyFuncs are injected at startup to avoid import cycles.
var (
	deriveFn        func(input string) types.Result
	appendHistoryFn func(input string) error
	loadHistoryFn   func() []types.HistoryEntry
)

// SetFuncs wires in the derivation and history functions.
func SetFuncs(
	derive func(string) types.Result,
	append func(string) error,
	load func() []types.HistoryEntry,
) {
	deriveFn = derive
	appendHistoryFn = append
	loadHistoryFn = load
}

type focus int

const (
	focusInput focus = iota
	focusHistory
)

// Model is the root Bubble Tea model for the hash TUI.
type Model struct {
	input   textinput.Model
	result  *types.Result
	history []types.HistoryEntry
	focused focus

	// historyCursor is the selected index into m.history (0 = newest).
	historyCursor int

	showHelp bool

	// Layout
	width  int
	height int
}

// NewModel creates the root model, loading existing history.
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "enter text to hash..."
	ti.CharLimit = 200
	ti.Focus()

	var hist []types.HistoryEntry
	if loadHistoryFn != nil {
		hist = loadHistoryFn()
	}

	return Model{
		input:   ti,
		history: hist,
		focused: focusInput,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.WindowSize(), textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Global
	switch k {
	case "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "tab":
		return m.switchFocus(), textinput.Blink
	}

	if m.focused == focusHistory {
		return m.handleHistoryKey(k)
	}
	return m.handleInputKey(msg, k)
}

func (m Model) handleInputKey(msg tea.KeyMsg, k string) (tea.Model, tea.Cmd) {
	switch k {
	case "esc":
		return m, tea.Quit

	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			return m, nil
		}
		m.deriveAndSave(text)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleHistoryKey(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "esc", "q":
		return m, tea.Quit

	case "up", "k":
		if m.historyCursor > 0 {
			m.historyCursor--
		}

	case "down", "j":
		if m.historyCursor < len(m.history)-1 {
			m.historyCursor++
		}

	case "enter":
		if len(m.history) > 0 {
			selected := m.history[m.historyCursor].Input
			m.input.SetValue(selected)
			m.deriveAndSave(selected)
			// Switch back to input so the user can edit
			m.focused = focusInput
			m.input.Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m *Model) switchFocus() Model {
	if m.focused == focusInput {
		m.focused = focusHistory
		m.input.Blur()
		// Start cursor at top of history
		m.historyCursor = 0
	} else {
		m.focused = focusInput
		m.input.Focus()
	}
	return *m
}

func (m *Model) deriveAndSave(text string) {
	if deriveFn != nil {
		r := deriveFn(text)
		m.result = &r
	}
	if appendHistoryFn != nil {
		_ = appendHistoryFn(text)
	}
	if loadHistoryFn != nil {
		m.history = loadHistoryFn()
	}
	m.historyCursor = 0
}

func (m *Model) historyCapacity() int {
	usable := m.height - 5
	if usable < 1 {
		return 1
	}
	return usable
}

// historyViewport returns the slice of history to display and the offset within it
// of the cursor, keeping the cursor visible.
func (m *Model) historyViewport() (visible []types.HistoryEntry, cursorInView int) {
	cap := m.historyCapacity()
	total := len(m.history)
	if total == 0 {
		return nil, 0
	}

	// Scroll window to keep cursor visible
	start := m.historyCursor - cap + 1
	if start < 0 {
		start = 0
	}
	if start > total-cap {
		start = total - cap
	}
	if start < 0 {
		start = 0
	}

	end := start + cap
	if end > total {
		end = total
	}

	return m.history[start:end], m.historyCursor - start
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftW := int(float64(m.width) * 0.60)
	rightW := m.width - leftW

	left := m.leftPanel(leftW)
	right := m.rightPanel(rightW)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	statusStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBG)).
		Width(m.width).
		Render(m.statusBar())

	if m.showHelp {
		return m.helpOverlay(body + "\n" + statusStyled)
	}
	return body + "\n" + statusStyled
}

func (m Model) statusBar() string {
	var left string
	if m.result != nil {
		left = " " + styleTitle.Render("[ "+m.result.Input+" ]")
	} else {
		left = " " + styleHint.Render("no input yet")
	}

	var hints string
	if m.focused == focusInput {
		hints = styleHint.Render("enter:derive  tab:history  ?:help  esc:quit") + " "
	} else {
		hints = styleHint.Render("↑↓:navigate  enter:re-derive  tab:back  ?:help  q:quit") + " "
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(hints)
	mid := m.width - leftW - rightW
	if mid < 1 {
		mid = 1
	}
	return left + strings.Repeat(" ", mid) + hints
}

func (m Model) helpOverlay(base string) string {
	return panels.HelpOverlay(base, hashHelpText(), colorActive, 44)
}

func hashHelpText() string {
	return fmt.Sprintf(`%s

%s
  enter       Derive hash values
  tab         Switch to history panel

%s (when focused)
  ↑/↓  j/k   Navigate entries
  enter       Re-derive selected

%s
  ?           Toggle this help
  esc / q     Quit`,
		styleTitle.Render("UNUM HASH — keyboard reference"),
		styleTitle.Render("INPUT"),
		styleTitle.Render("HISTORY"),
		styleTitle.Render("ACTIONS"),
	)
}

func (m Model) leftPanel(w int) string {
	innerW := w - 2
	if innerW < 10 {
		innerW = 10
	}

	var title string
	if m.focused == focusInput {
		title = styleTitle.Render(" HASH ")
	} else {
		title = styleTitleDim.Render(" HASH ")
	}

	sep := styleHint.Render(strings.Repeat("─", innerW))
	inputLine := m.input.View()

	var resultsBlock string
	if m.result != nil {
		rows := []struct{ label, value string }{
			{"input", m.result.Input},
			{"port", fmt.Sprintf("%d", m.result.Port)},
			{"uuid", m.result.UUID},
			{"color", m.result.Color},
			{"short", m.result.Short},
			{"emoji", m.result.Emoji},
			{"phrase", m.result.Phrase},
		}
		lines := make([]string, 0, len(rows)+2)
		lines = append(lines, sep)
		for _, row := range rows {
			label := styleLabel.Render(fmt.Sprintf("  %-7s", row.label))
			value := styleValue.Render(row.value)
			lines = append(lines, fmt.Sprintf("%s  %s", label, value))
		}
		lines = append(lines, sep)
		resultsBlock = strings.Join(lines, "\n")
	} else {
		resultsBlock = styleHint.Render("  type something and press enter")
	}

	content := strings.Join([]string{title, "", inputLine, "", resultsBlock}, "\n")

	active := m.focused == focusInput
	border := borderDim
	if active {
		border = borderActive
	}
	return border.Width(innerW).Height(m.height - 3).Render(content)
}

func (m Model) rightPanel(w int) string {
	innerW := w - 2
	if innerW < 8 {
		innerW = 8
	}

	histFocused := m.focused == focusHistory
	var title string
	if histFocused {
		title = styleTitle.Render(" HISTORY ")
	} else {
		title = styleTitleDim.Render(" HISTORY ")
	}

	visible, cursorInView := m.historyViewport()

	lines := make([]string, 0, len(visible)+1)
	lines = append(lines, title)

	if len(visible) == 0 {
		lines = append(lines, styleHint.Render("  no history"))
	}
	for i, e := range visible {
		text := e.Input
		maxLen := innerW - 4
		if maxLen < 1 {
			maxLen = 1
		}
		if len([]rune(text)) > maxLen {
			runes := []rune(text)
			text = string(runes[:maxLen]) + "…"
		}
		entry := fmt.Sprintf("  %s", text)
		if histFocused && i == cursorInView {
			lines = append(lines, styleHistorySelected.Render("> "+strings.TrimPrefix(entry, "  ")))
		} else {
			lines = append(lines, styleValue.Render(entry))
		}
	}

	content := strings.Join(lines, "\n")
	border := borderDim
	if histFocused {
		border = borderActive
	}
	return border.Width(innerW).Height(m.height - 3).Render(content)
}
