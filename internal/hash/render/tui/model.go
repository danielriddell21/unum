// Package tui provides a Bubble Tea TUI for the hash derivation tool.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	hashpanels "github.com/danielriddell21/unum/internal/hash/render/tui/panels"
	"github.com/danielriddell21/unum/internal/hash/types"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
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
	input         textinput.Model
	result        *types.Result
	focused       focus
	hashPanel     hashpanels.HashPanel
	historyPanel  hashpanels.HistoryPanel
	showHelp      bool
	width         int
	height        int
}

// NewModel creates the root model, loading existing history.
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "enter text to hash..."
	ti.CharLimit = 200
	ti.Focus()

	hp := hashpanels.NewHistoryPanel(0, 0)
	if loadHistoryFn != nil {
		hp.SetHistory(loadHistoryFn())
	}

	return Model{
		input:        ti,
		hashPanel:    hashpanels.NewHashPanel(0, 0),
		historyPanel: hp,
		focused:      focusInput,
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
		m.resize()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) resize() {
	lw, rw, h := m.panelDimensions()
	m.hashPanel.Resize(lw, h)
	m.historyPanel.Resize(rw, h)
}

func (m *Model) panelDimensions() (lw, rw, h int) {
	lw = int(float64(m.width)*0.60) - 2
	rw = m.width - int(float64(m.width)*0.60) - 2
	h = m.height - 3
	if lw < 10 {
		lw = 10
	}
	if rw < 8 {
		rw = 8
	}
	if h < 3 {
		h = 3
	}
	return
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
		m.historyPanel.ScrollUp()

	case "down", "j":
		m.historyPanel.ScrollDown()

	case "enter":
		if selected := m.historyPanel.SelectedInput(); selected != "" {
			m.input.SetValue(selected)
			m.deriveAndSave(selected)
			m.focused = focusInput
			m.historyPanel.SetFocused(false)
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
		m.historyPanel.ResetCursor()
		m.historyPanel.SetFocused(true)
	} else {
		m.focused = focusInput
		m.historyPanel.SetFocused(false)
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
		m.historyPanel.SetHistory(loadHistoryFn())
	}
	m.historyPanel.ResetCursor()
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftW := int(float64(m.width) * 0.60)
	rightW := m.width - leftW

	hashTitle := panelTitle("HASH", m.focused == focusInput)
	hashContent := hashTitle + "\n" + m.hashPanel.View(m.input.View(), m.result)
	left := wrapPanel(hashContent, m.focused == focusInput, leftW, m.height-1)

	histTitle := panelTitle("HISTORY", m.focused == focusHistory)
	histContent := histTitle + "\n" + m.historyPanel.View()
	right := wrapPanel(histContent, m.focused == focusHistory, rightW, m.height-1)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	statusStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBG)).
		Width(m.width).
		Render(statusBar(m.result, m.focused, m.width))

	if m.showHelp {
		return m.helpOverlay(body + "\n" + statusStyled)
	}
	return body + "\n" + statusStyled
}

func panelTitle(title string, active bool) string {
	if active {
		return styleTitle.Render(" " + title + " ")
	}
	return styleTitleDim.Render(" " + title + " ")
}

func wrapPanel(content string, active bool, width, height int) string {
	if width < 4 {
		width = 4
	}
	if height < 4 {
		height = 4
	}
	style := borderDim.Width(width - 2).Height(height - 2)
	if active {
		style = borderActive.Width(width - 2).Height(height - 2)
	}
	return style.Render(content)
}

func (m Model) helpOverlay(base string) string {
	return tuipanels.HelpOverlay(base, hashHelpText(), colorActive, 44)
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
