package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/render/tui/panels"
)

type viewMode int

const (
	viewUnified viewMode = iota
	viewSplit
)

type inputMode int

const (
	modeNormal inputMode = iota
	modeSearch
)

// Model is the root Bubble Tea model for the diff TUI.
type Model struct {
	diff *node.Diff

	view    viewMode
	unified panels.UnifiedPanel
	split   panels.SplitPanel
	info    panels.InfoPanel

	mode        inputMode
	searchInput textinput.Model

	width       int
	height      int
	initialized bool
}

// NewModel creates the diff TUI model.
func NewModel(d *node.Diff) Model {
	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 100

	return Model{
		diff:        d,
		view:        viewUnified,
		mode:        modeNormal,
		searchInput: si,
	}
}

func (m *Model) initPanels() {
	dw, dh, iw, ih := m.panelDimensions()
	m.unified = panels.NewUnifiedPanel(m.diff, dw, dh)
	m.split = panels.NewSplitPanel(m.diff, dw, dh)
	m.info = panels.NewInfoPanel(m.diff, iw, ih)
}

// panelDimensions returns (diffW, diffH, infoW, infoH).
// Layout: left 65% = diff, right 35% = info.
func (m *Model) panelDimensions() (dw, dh, iw, ih int) {
	statusH := 1
	usable := m.height - statusH

	dw = int(float64(m.width)*0.65) - 2
	iw = m.width - int(float64(m.width)*0.65) - 2
	dh = usable - 2
	ih = usable - 2

	if dw < 10 {
		dw = 10
	}
	if iw < 10 {
		iw = 10
	}
	if dh < 3 {
		dh = 3
	}
	if ih < 3 {
		ih = 3
	}
	return
}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.initialized {
			m.initPanels()
			m.initialized = true
		} else {
			m.resize()
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	if k == "ctrl+c" {
		return m, tea.Quit
	}

	// Search mode
	if m.mode == modeSearch {
		switch k {
		case "esc":
			m.mode = modeNormal
			m.searchInput.SetValue("")
			m.unified.SetSearch("")
			m.split.SetSearch("")
			m.searchInput.Blur()
		case "enter":
			m.mode = modeNormal
			m.searchInput.Blur()
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			q := m.searchInput.Value()
			m.unified.SetSearch(q)
			m.split.SetSearch(q)
			return m, cmd
		}
		return m, nil
	}

	// Normal mode
	switch k {
	case "q", "esc":
		return m, tea.Quit
	case "v":
		// Split view is only meaningful for text (hunk-based) diffs.
		if m.diff != nil && m.diff.Root == nil {
			if m.view == viewUnified {
				m.view = viewSplit
			} else {
				m.view = viewUnified
			}
		}
	case "/":
		m.mode = modeSearch
		m.searchInput.Focus()
		return m, textinput.Blink
	case "n":
		m.unified.NextMatch()
		m.split.NextMatch()
	case "N":
		m.unified.PrevMatch()
		m.split.PrevMatch()
	case "j", "down":
		m.unified.ScrollDown(1)
		m.split.ScrollDown(1)
	case "k", "up":
		m.unified.ScrollUp(1)
		m.split.ScrollUp(1)
	case "d", "ctrl+d":
		m.unified.HalfPageDown()
		m.split.HalfPageDown()
	case "u", "ctrl+u":
		m.unified.HalfPageUp()
		m.split.HalfPageUp()
	case "tab":
		m.split.CycleFocus()
	}

	return m, nil
}

func (m *Model) resize() {
	dw, dh, iw, ih := m.panelDimensions()
	m.unified.Resize(dw, dh)
	m.split.Resize(dw, dh)
	m.info.Resize(iw, ih)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	diffW := int(float64(m.width) * 0.65)
	infoW := m.width - diffW

	// Diff panel
	var diffContent string
	var viewLabel string
	if m.view == viewUnified {
		diffContent = m.unified.View()
		viewLabel = "unified"
	} else {
		diffContent = m.split.View()
		viewLabel = "split"
	}

	diffTitle := panelTitle("DIFF", true)
	diffPane := wrapPanel(diffTitle+"\n"+diffContent, true, diffW, m.height-1)

	// Info panel
	infoTitle := panelTitle("INFO", false)
	infoPane := wrapPanel(infoTitle+"\n"+m.info.View(), false, infoW, m.height-1)

	body := lipgloss.JoinHorizontal(lipgloss.Top, diffPane, infoPane)

	// Status bar
	searchQ := ""
	if m.mode == modeSearch {
		searchQ = m.searchInput.Value()
	}
	sb := panels.StatusBar(m.diff, viewLabel, m.width, m.mode == modeSearch, searchQ)
	statusStyled := lipgloss.NewStyle().
		Background(lipgloss.Color("#0D0D0D")).
		Foreground(lipgloss.Color("#3A3A3A")).
		Width(m.width).
		Render(sb)

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

// Search bar shown inline in status when mode=modeSearch
func (m Model) searchBar() string {
	if m.mode != modeSearch {
		return ""
	}
	return m.searchInput.View()
}

// ensure searchBar is used (it's referenced in View indirectly via status bar)
var _ = (&Model{}).searchBar

func (m Model) viewModeString() string {
	if m.view == viewSplit {
		return "split"
	}
	return "unified"
}

var _ = strings.Join // ensure strings import is used
