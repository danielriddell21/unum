package tui

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/json/node"
	jsonpanels "github.com/danielriddell21/unum/internal/json/render/tui/panels"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

// focusedPanel identifies which panel is active.
type focusedPanel int

const (
	focusTree focusedPanel = iota
	focusPreview
	focusLens
)

// mode is the current interaction mode.
type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeHelp
)

// yankFeedbackMsg clears the yank toast after a delay.
type yankFeedbackMsg struct{}

// Model is the root Bubble Tea model. It owns all three panels and dispatches
// keystrokes to the focused sub-panel.
type Model struct {
	root     *node.Node
	filename string

	// Panels
	tree    jsonpanels.TreePanel
	preview jsonpanels.PreviewPanel
	lens    jsonpanels.LensPanel

	// State
	focused      focusedPanel
	currentMode  mode
	searchInput  textinput.Model
	yankFeedback string

	// Set to true when the user wants to pick a new file (causes clean quit).
	ReloadRequest bool

	// Layout
	width  int
	height int
}

// NewModel creates the root model with the parsed tree.
func NewModel(root *node.Node, filename string) Model {
	si := textinput.New()
	si.Placeholder = "filter..."
	si.CharLimit = 100

	m := Model{
		root:     root,
		filename: filename,
		focused:  focusTree,
	}

	m.searchInput = si
	return m
}

func (m *Model) initPanels() {
	treeW, treeH, previewW, previewH, lensW, lensH := m.panelDimensions()

	m.tree = jsonpanels.NewTreePanel(m.root, treeW, treeH)
	m.preview = jsonpanels.NewPreviewPanel(previewW, previewH)
	m.lens = jsonpanels.NewLensPanel(m.root, lensW, lensH)

	m.tree.SetFocused(true)
	m.syncPanels()
}

// panelDimensions returns (treeW, treeH, previewW, previewH, lensW, lensH).
// Layout: left column = tree, right column top = preview, right column bottom = lens.
// Right column split: preview ≈ 35%, lens ≈ 65%.
func (m *Model) panelDimensions() (tw, th, pw, ph, lw, lh int) {
	statusH := 1
	usable := m.height - statusH

	tw = int(float64(m.width) * 0.38)
	rw := m.width - tw

	// Remove borders from usable height
	ph = int(float64(usable) * 0.38)
	lh = usable - ph

	// Border widths (each border takes 2 chars: left+right)
	tw -= 2
	rw -= 2
	pw = rw
	lw = rw

	th = usable - 2
	ph -= 2
	lh -= 2

	if tw < 10 {
		tw = 10
	}
	if pw < 10 {
		pw = 10
	}
	if lh < 5 {
		lh = 5
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
		if m.tree.CursorNode() == nil {
			m.initPanels()
		} else {
			m.resize()
		}
		return m, nil

	case yankFeedbackMsg:
		m.yankFeedback = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) { //nolint:cyclop // key dispatch switch covers every navigation + action key; extracting helpers would obscure control flow
	k := msg.String()

	// Global quit (always)
	if k == "ctrl+c" {
		return m, tea.Quit
	}

	// Search mode
	if m.currentMode == modeSearch {
		switch k {
		case "esc":
			m.currentMode = modeNormal
			m.searchInput.SetValue("")
			m.tree.SetSearch("")
			m.searchInput.Blur()
		case "enter":
			m.currentMode = modeNormal
			m.searchInput.Blur()
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.tree.SetSearch(m.searchInput.Value())
			m.syncPanels()
			return m, cmd
		}
		return m, nil
	}

	// Help mode
	if m.currentMode == modeHelp {
		if k == "?" || k == "q" || k == "esc" {
			m.currentMode = modeNormal
		}
		return m, nil
	}

	// Normal mode global keys
	switch k {
	case "q", "esc":
		return m, tea.Quit

	case "o":
		m.ReloadRequest = true
		return m, tea.Quit

	case "?":
		m.currentMode = modeHelp
		return m, nil

	case "/":
		m.currentMode = modeSearch
		m.searchInput.Focus()
		return m, textinput.Blink

	case "y":
		if n := m.tree.CursorNode(); n != nil {
			path := n.Path()
			_ = clipboard.WriteAll(path)
			m.yankFeedback = "path yanked!"
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return yankFeedbackMsg{}
			})
		}
		return m, nil

	case "tab":
		m.cyclePanel(1)
		return m, nil

	case "shift+tab":
		m.cyclePanel(-1)
		return m, nil

	// Lens switching (global — works from any panel)
	case "1", "2", "3", "4", "5", "6":
		m.lens.HandleKey(k)
		return m, nil
	}

	// Panel-specific keys
	switch m.focused {
	case focusTree:
		changed := m.tree.Update(msg)
		if changed {
			m.syncPanels()
		}
	case focusPreview:
		m.preview.HandleKey(k)
	case focusLens:
		m.lens.HandleKey(k)
	}

	return m, nil
}

func (m *Model) cyclePanel(dir int) {
	panels := []focusedPanel{focusTree, focusPreview, focusLens}
	cur := int(m.focused)
	next := (cur + dir + len(panels)) % len(panels)
	m.focused = focusedPanel(next)

	m.tree.SetFocused(m.focused == focusTree)
	m.preview.SetFocused(m.focused == focusPreview)
	m.lens.SetFocused(m.focused == focusLens)
}

func (m *Model) syncPanels() {
	n := m.tree.CursorNode()
	m.preview.SetNode(n)
	m.lens.SetCursorNode(n)
}

func (m *Model) resize() {
	tw, th, pw, ph, lw, lh := m.panelDimensions()
	m.tree.Resize(tw, th)
	m.preview.Resize(pw, ph)
	m.lens.Resize(lw, lh)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Build left column (tree)
	leftTitle := panelTitle("JSON TREE", m.focused == focusTree)
	treeContent := m.tree.View()
	leftPane := wrapPanel(leftTitle+"\n"+treeContent, m.focused == focusTree,
		int(float64(m.width)*0.38)-0, m.height-1)

	// Build right column
	// Top: preview
	previewTitle := panelTitle("PREVIEW", m.focused == focusPreview)
	previewContent := m.preview.View()
	previewPane := wrapPanel(previewTitle+"\n"+previewContent, m.focused == focusPreview,
		m.width-int(float64(m.width)*0.38)-0, int(float64(m.height-1)*0.38))

	// Bottom: lens
	lensContent := m.lens.View()
	lensPane := wrapPanel(lensContent, m.focused == focusLens,
		m.width-int(float64(m.width)*0.38)-0, m.height-1-int(float64(m.height-1)*0.38))

	rightCol := lipgloss.JoinVertical(lipgloss.Left, previewPane, lensPane)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightCol)

	// Status bar
	cursorNode := m.tree.CursorNode()
	statusBar := jsonpanels.StatusBar(cursorNode, m.width, m.currentMode == modeSearch,
		m.searchInput.Value(), m.yankFeedback)
	statusStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBG)).
		Width(m.width).
		Render(statusBar)

	// Help overlay
	if m.currentMode == modeHelp {
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
	return tuipanels.HelpOverlay(base, helpText(), colorActiveBorder, 50)
}

func helpText() string {
	return fmt.Sprintf(`%s

%s
  j/k  ↑/↓   Navigate tree
  l/→  enter  Expand node
  h/←         Collapse node
  tab         Next panel
  shift+tab   Prev panel

%s
  1-6         Switch lens
  g/t/j       TypeGen: Go/TS/Schema
  j/k         Scroll lens

%s
  /           Enter search mode
  y           Yank jq path to clipboard
  o           Open a different file
  ?           Toggle this help
  q / esc     Quit`,
		styleTitle.Render("UNUM — keyboard reference"),
		styleTitle.Render("NAVIGATION"),
		styleTitle.Render("LENSES"),
		styleTitle.Render("ACTIONS"),
	)
}
