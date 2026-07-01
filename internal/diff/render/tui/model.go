package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/render/tui/panels"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

type viewMode int

const (
	viewSemantic viewMode = iota
	viewUnified
	viewSplit
)

type inputMode int

const (
	modeNormal inputMode = iota
	modeSearch
	modeHelp
)

type Model struct {
	diff *node.Diff

	view    viewMode
	unified panels.UnifiedPanel
	split   panels.SplitPanel
	info    panels.InfoPanel

	mode        inputMode
	searchInput textinput.Model

	width         int
	height        int
	initialized   bool
	ReloadRequest bool
	version       string
}

func NewModel(d *node.Diff, version string) Model {
	si := textinput.New()
	si.Placeholder = "search..."
	si.CharLimit = 100

	return Model{
		diff:        d,
		mode:        modeNormal,
		searchInput: si,
		version:     version,
	}
}

func (m *Model) initPanels() {
	dw, dh, iw, ih := m.panelDimensions()
	m.unified = panels.NewUnifiedPanel(m.diff, dw, dh)
	m.split = panels.NewSplitPanel(m.diff, dw, dh)
	m.info = panels.NewInfoPanel(m.diff, iw, ih)
	if m.diff.Root == nil {
		m.view = viewUnified
	}
}

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

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) { //nolint:cyclop // NOSONAR: key dispatch switch covers every navigation + action key; extracting helpers would obscure control flow
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

	// Help mode
	if m.mode == modeHelp {
		if k == "?" || k == "q" || k == "esc" {
			m.mode = modeNormal
		}
		return m, nil
	}

	// Normal mode
	switch k {
	case "q", "esc":
		return m, tea.Quit
	case "?":
		m.mode = modeHelp
		return m, nil
	case "o":
		m.ReloadRequest = true
		return m, tea.Quit
	case "v":
		hasSemantic := m.diff.Root != nil
		hasText := len(m.diff.Hunks) > 0
		if hasSemantic && hasText {
			switch m.view {
			case viewSemantic:
				m.view = viewUnified
			case viewUnified:
				m.view = viewSplit
			default:
				m.view = viewSemantic
			}
		} else if hasText {
			if m.view == viewSplit {
				m.view = viewUnified
			} else {
				m.view = viewSplit
			}
		}
		// semantic-only (no hunks): v does nothing
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
	switch m.view {
	case viewSemantic:
		m.unified.SetTextOnly(false)
		diffContent = m.unified.View()
		viewLabel = "semantic"
	case viewUnified:
		m.unified.SetTextOnly(true)
		diffContent = m.unified.View()
		viewLabel = "unified"
	default:
		diffContent = m.split.View()
		viewLabel = "split"
	}

	diffTitle := tuipanels.PanelTitle("DIFF", true)
	diffPane := tuipanels.WrapPanel(diffTitle+"\n"+diffContent, true, diffW, m.height-1)

	// Info panel
	infoTitle := tuipanels.PanelTitle("INFO", false)
	infoPane := tuipanels.WrapPanel(infoTitle+"\n"+m.info.View(), false, infoW, m.height-1)

	body := lipgloss.JoinHorizontal(lipgloss.Top, diffPane, infoPane)

	// Status bar
	searchQ := ""
	if m.mode == modeSearch {
		searchQ = m.searchInput.Value()
	}
	mc := m.unified.MatchCount()
	if m.view == viewSplit {
		mc = m.split.MatchCount()
	}

	sb := panels.StatusBar(m.diff, viewLabel, m.width, m.mode == modeSearch, searchQ, mc, m.version)
	statusStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(tuipanels.ColorBG)).
		Width(m.width).
		Render(sb)

	if m.mode == modeHelp {
		return m.helpOverlay(body + "\n" + statusStyled)
	}

	return body + "\n" + statusStyled
}

func (m Model) helpOverlay(base string) string {
	return tuipanels.HelpOverlay(base, helpText(), tuipanels.ColorActiveBorder, 52)
}

func helpText() string {
	return fmt.Sprintf(`%s

%s
  j/k  ↑/↓      Scroll diff
  d/u  ctrl+D/U  Half-page scroll
  n / N          Next / previous search match

%s
  v              Cycle view (semantic / unified / split)
  tab            Cycle focus (split view)

%s
  /              Enter search mode
  o              Open new files
  ?              Toggle this help
  q / esc        Quit`,
		tuipanels.StyleTitle.Render("UNUM DIFF — keyboard reference"),
		tuipanels.StyleTitle.Render("NAVIGATION"),
		tuipanels.StyleTitle.Render("VIEWS"),
		tuipanels.StyleTitle.Render("ACTIONS"),
	)
}
