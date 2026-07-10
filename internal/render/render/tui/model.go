package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

type Info struct {
	File   string
	Lang   string
	Source string
	Width  int
	Height int
	Bytes  int
	Shapes int
	Conns  int
	Err    error
}

type Model struct {
	info     Info
	lines    []string
	scroll   int
	width    int
	height   int
	version  string
	showHelp bool
}

func Start(version, themeName string, info Info) error {
	tuipanels.ApplyBaseStyles(tuipanels.ResolvePalette(themeName))
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
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftW := int(float64(m.width) * 0.55)
	rightW := m.width - leftW
	bodyH := m.height - 1

	srcTitle := tuipanels.PanelTitle("SOURCE", true)
	left := tuipanels.WrapPanel(srcTitle+"\n"+m.sourceView(leftW-4, bodyH-3), true, leftW, bodyH)

	renTitle := tuipanels.PanelTitle("RENDER", false)
	right := tuipanels.WrapPanel(renTitle+"\n"+m.renderView(), false, rightW, bodyH)

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
		fmt.Fprintf(&b, "%s%s\n", num, line)
	}
	return b.String()
}

func (m Model) renderView() string {
	if m.info.Err != nil {
		return tuipanels.StyleHint.Render("render failed:\n" + m.info.Err.Error())
	}
	rows := [][2]string{
		{"language", m.info.Lang},
		{"formats", "svg · png · drawio"},
		{"svg size", fmt.Sprintf("%d × %d", m.info.Width, m.info.Height)},
		{"svg bytes", fmt.Sprintf("%d", m.info.Bytes)},
	}
	if m.info.Shapes >= 0 {
		rows = append(rows,
			[2]string{"shapes", fmt.Sprintf("%d", m.info.Shapes)},
			[2]string{"edges", fmt.Sprintf("%d", m.info.Conns)},
		)
	}
	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "%s  %s\n", tuipanels.StyleTitle.Render(fmt.Sprintf("%-9s", r[0])), r[1])
	}
	b.WriteString("\n")
	b.WriteString(tuipanels.StyleHint.Render("terminals can't show images —\nuse --web for a live preview or\n-o to write svg/png/drawio."))
	return b.String()
}

func (m Model) statusBar() string {
	left := tuipanels.StyleTitle.Render(" [ "+m.info.File+" ] ") + tuipanels.StyleHint.Render(m.info.Lang)
	right := tuipanels.StyleHint.Render(fmt.Sprintf("%d×%d · ?:help · q:quit · v%s ", m.info.Width, m.info.Height, m.version))
	return tuipanels.Bar(left, right, m.width)
}

func helpText() string {
	return fmt.Sprintf(`%s

%s
  ↑/↓  j/k   Scroll source
  ?           Toggle this help
  esc / q     Quit`,
		tuipanels.StyleTitle.Render("UNUM RENDER — keyboard reference"),
		tuipanels.StyleTitle.Render("ACTIONS"),
	)
}
