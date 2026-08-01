package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

type Info struct {
	File   string
	Lang   string
	Source string
	ASCII  string
	SVG    []byte
	PNG    []byte
	Drawio []byte
	Width  int
	Height int
	Shapes int
	Conns  int
	Err    error
}

var (
	styleStruct = lipgloss.NewStyle()
	styleText   = lipgloss.NewStyle()
)

type Model struct {
	info     Info
	lines    []string
	scroll   int
	status   string
	width    int
	height   int
	version  string
	showHelp bool
}

func Start(info Info, themeName, version string) error {
	pal := tuipanels.ResolvePalette(themeName)
	tuipanels.ApplyBaseStyles(pal)
	styleStruct = lipgloss.NewStyle().Foreground(lipgloss.Color(pal.AccentPrimary))
	styleText = lipgloss.NewStyle().Foreground(lipgloss.Color(pal.Text))
	p := tea.NewProgram(NewModel(version, info), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("render TUI: %w", err)
	}
	return nil
}

func NewModel(version string, info Info) Model {
	return Model{
		info:    info,
		lines:   strings.Split(strings.TrimRight(info.Source, "\n"), "\n"),
		version: version,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg.String())
	}
	return m, nil
}

func (m Model) handleKey(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
	case "up", "k":
		if m.scroll > 0 {
			m.scroll--
		}
	case "down", "j":
		if m.scroll < len(m.lines)-1 {
			m.scroll++
		}
	case "s":
		m.status = save("diagram.svg", m.info.SVG)
	case "p":
		m.status = save("diagram.png", m.info.PNG)
	case "d":
		m.status = save("diagram.drawio", m.info.Drawio)
	}
	return m, nil
}

func save(name string, data []byte) string {
	if len(data) == 0 {
		return "nothing to write for " + name
	}
	if err := os.WriteFile(name, data, 0o644); err != nil { //nolint:gosec // user-facing artifact, not a secret
		return "error: " + err.Error()
	}
	return fmt.Sprintf("wrote %s (%d bytes)", name, len(data))
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftW := int(float64(m.width) * 0.42)
	rightW := m.width - leftW
	bodyH := m.height - 1

	srcTitle := tuipanels.PanelTitle("SOURCE", true)
	left := tuipanels.WrapPanel(srcTitle+"\n"+m.sourceView(leftW-4, bodyH-3), true, leftW, bodyH)

	renTitle := tuipanels.PanelTitle("RENDER", false)
	right := tuipanels.WrapPanel(renTitle+"\n"+m.renderView(rightW-4, bodyH-4), false, rightW, bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	status := lipgloss.NewStyle().
		Background(lipgloss.Color(tuipanels.ColorBG)).
		Width(m.width).
		Render(m.statusBar())

	view := body + "\n" + status
	if m.showHelp {
		return tuipanels.HelpOverlay(view, helpText(), tuipanels.ColorActiveBorder, 46)
	}
	return view
}

func (m Model) sourceView(w, h int) string {
	if h < 1 {
		h = 1
	}
	var b strings.Builder
	for i := m.scroll; i < len(m.lines) && i < m.scroll+h; i++ {
		line := m.lines[i]
		if len(line) > w && w > 1 {
			line = line[:w-1] + "…"
		}
		num := tuipanels.StyleTitleDim.Render(fmt.Sprintf("%3d ", i+1))
		fmt.Fprintf(&b, "%s%s\n", num, styleText.Render(line))
	}
	return b.String()
}

func (m Model) renderView(w, h int) string {
	caption := tuipanels.StyleTitleDim.Render(m.caption())
	if m.info.Err != nil {
		return caption + "\n\n" + tuipanels.StyleHint.Render("preview unavailable:\n"+m.info.Err.Error())
	}
	if m.info.ASCII != "" {
		return caption + "\n" + m.asciiView(w, h-1)
	}
	note := "no terminal preview for this diagram type\n\nopen it with --web, or export --format svg / png"
	return caption + "\n\n" + tuipanels.StyleHint.Render(note)
}

func (m Model) asciiView(w, h int) string {
	if w < 1 || h < 1 {
		return ""
	}
	lines := strings.Split(strings.TrimRight(m.info.ASCII, "\n"), "\n")
	rows := make([][]rune, len(lines))
	contentW := 0
	for i, line := range lines {
		rows[i] = []rune(line)
		if len(rows[i]) > contentW {
			contentW = len(rows[i])
		}
	}

	startRow, padTop := windowRange(len(rows), h)
	startCol, padLeft := windowRange(contentW, w)

	var b strings.Builder
	for i := 0; i < padTop; i++ {
		b.WriteByte('\n')
	}
	pad := strings.Repeat(" ", padLeft)
	for i := startRow; i < len(rows) && i < startRow+h-padTop; i++ {
		rs := rows[i]
		if startCol < len(rs) {
			rs = rs[startCol:]
		} else {
			rs = nil
		}
		if len(rs) > w-padLeft {
			rs = rs[:w-padLeft]
		}
		b.WriteString(pad)
		b.WriteString(colorizeASCII(rs))
		b.WriteByte('\n')
	}
	return b.String()
}

func windowRange(content, view int) (start, pad int) {
	if content <= view {
		// Smaller than the pane: no scroll, pad to centre it.
		return 0, (view - content) / 2
	}
	// Larger than the pane: start at the centred offset, no pad.
	return (content - view) / 2, 0
}

func colorizeASCII(rs []rune) string {
	var b strings.Builder
	for _, r := range rs {
		switch {
		case r == ' ':
			b.WriteRune(r)
		case isStructRune(r):
			b.WriteString(styleStruct.Render(string(r)))
		default:
			b.WriteString(styleText.Render(string(r)))
		}
	}
	return b.String()
}

func isStructRune(r rune) bool {
	if r >= 0x2500 && r <= 0x257F { // box-drawing block
		return true
	}
	switch r {
	case '▲', '▼', '◄', '►', '◀', '▶', '↑', '↓', '←', '→':
		return true
	}
	return false
}

func (m Model) caption() string {
	parts := []string{m.info.Lang}
	if m.info.Width > 0 && m.info.Height > 0 {
		parts = append(parts, fmt.Sprintf("%d×%d", m.info.Width, m.info.Height))
	}
	if m.info.Shapes >= 0 {
		parts = append(parts, fmt.Sprintf("%d shapes", m.info.Shapes), fmt.Sprintf("%d edges", m.info.Conns))
	}
	return strings.Join(parts, " · ")
}

func (m Model) statusBar() string {
	left := tuipanels.StyleTitle.Render(" [ "+m.info.File+" ] ") + tuipanels.StyleHint.Render(m.info.Lang)
	right := "s:svg · p:png · d:drawio · ?:help · q:quit "
	if m.status != "" {
		right = m.status + "  "
	}
	return tuipanels.Bar(left, tuipanels.StyleHint.Render(right), m.width)
}

func helpText() string {
	return fmt.Sprintf(`%s

%s
  ↑/↓  j/k   Scroll source
  s           Write diagram.svg
  p           Write diagram.png
  d           Write diagram.drawio
  ?           Toggle this help
  esc / q     Quit`,
		tuipanels.StyleTitle.Render("UNUM RENDER — keyboard reference"),
		tuipanels.StyleTitle.Render("ACTIONS"),
	)
}
