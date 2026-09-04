package tui

import (
	"bytes"
	"fmt"
	"image"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/danielriddell21/unum/internal/image/optimize"
	imagepanels "github.com/danielriddell21/unum/internal/image/render/tui/panels"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

const (
	qualityStep = 5
	scaleStep   = 0.05
	minScale    = 0.05
)

type Deps struct {
	Optimize func(optimize.Source, optimize.Options) (optimize.Result, error)
	Ladder   func(optimize.Source, optimize.Options, []int) ([]optimize.Step, error)
	Save     func(path string, data []byte) error
}

func DefaultDeps() Deps {
	return Deps{
		Optimize: optimize.Optimize,
		Ladder:   optimize.Ladder,
		Save: func(path string, data []byte) error {
			if err := os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec // user-facing artifact, not a secret
				return fmt.Errorf("write %s: %w", path, err)
			}
			return nil
		},
	}
}

type optimizedMsg struct {
	seq     int
	result  optimize.Result
	preview image.Image
	err     error
}

type ladderMsg struct {
	seq   int
	steps []optimize.Step
	err   error
}

type Model struct {
	src     optimize.Source
	opts    optimize.Options
	output  string
	version string
	deps    Deps

	result   *optimize.Result
	steps    []optimize.Step
	preview  image.Image
	encoding bool
	err      error
	saved    string
	seq      int

	previewPanel  imagepanels.PreviewPanel
	settingsPanel imagepanels.SettingsPanel

	showHelp bool
	width    int
	height   int
}

func NewModel(src optimize.Source, opts optimize.Options, output, version string, deps Deps) Model {
	opts.Format = optimize.ResolveOutputFormat(opts.Format, src.Format)
	if opts.Scale <= 0 {
		opts.Scale = 1
	}
	opts.Quality = optimize.ClampQuality(opts.Quality)

	return Model{
		src:           src,
		opts:          opts,
		output:        output,
		version:       version,
		deps:          deps,
		preview:       src.Image,
		encoding:      true,
		previewPanel:  imagepanels.NewPreviewPanel(0, 0),
		settingsPanel: imagepanels.NewSettingsPanel(0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.WindowSize(), m.encodeCmd(), m.ladderCmd())
}

func (m Model) encodeCmd() tea.Cmd {
	seq, src, opts, deps := m.seq, m.src, m.opts, m.deps
	return func() tea.Msg {
		result, err := deps.Optimize(src, opts)
		if err != nil {
			return optimizedMsg{seq: seq, err: err}
		}
		// Preview what was actually encoded, so quality loss is visible rather
		// than implied.
		preview, _, decodeErr := image.Decode(bytes.NewReader(result.Data))
		if decodeErr != nil {
			preview = src.Image
		}
		return optimizedMsg{seq: seq, result: result, preview: preview}
	}
}

func (m Model) ladderCmd() tea.Cmd {
	seq, src, opts, deps := m.seq, m.src, m.opts, m.deps
	return func() tea.Msg {
		steps, err := deps.Ladder(src, opts, optimize.DefaultLadder)
		return ladderMsg{seq: seq, steps: steps, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil

	case optimizedMsg:
		if msg.seq != m.seq {
			return m, nil
		}
		m.encoding = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		result := msg.result
		m.result = &result
		m.preview = msg.preview
		return m, nil

	case ladderMsg:
		if msg.seq == m.seq && msg.err == nil {
			m.steps = msg.steps
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) resize() {
	lw, rw, h := m.panelDimensions()
	m.previewPanel.Resize(lw, h)
	m.settingsPanel.Resize(rw, h)
}

func (m *Model) panelDimensions() (lw, rw, h int) {
	lw = int(float64(m.width)*0.50) - 2
	rw = m.width - int(float64(m.width)*0.50) - 2
	h = m.height - 3
	return max(lw, 10), max(rw, 10), max(h, 3)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "left", "h":
		return m.adjust(-qualityStep, 0)
	case "right", "l":
		return m.adjust(qualityStep, 0)
	case "down", "j":
		return m.adjust(0, -scaleStep)
	case "up", "k":
		return m.adjust(0, scaleStep)
	case "f":
		return m.cycleFormat()
	case "r":
		return m.reset()
	case "s":
		return m.save()
	}
	return m, nil
}

func (m Model) adjust(quality int, scale float64) (tea.Model, tea.Cmd) {
	m.opts.Quality = optimize.ClampQuality(m.opts.Quality + quality)
	m.opts.Scale = min(max(m.opts.Scale+scale, minScale), 1)
	return m.recompute()
}

func (m Model) cycleFormat() (tea.Model, tea.Cmd) {
	formats := optimize.EncodableFormats()
	next := 0
	for i, f := range formats {
		if f == m.opts.Format {
			next = (i + 1) % len(formats)
			break
		}
	}
	m.opts.Format = formats[next]
	m.output = replaceExt(m.output, m.opts.Format.Ext())
	return m.recompute()
}

func (m Model) reset() (tea.Model, tea.Cmd) {
	m.opts.Quality = 80
	m.opts.Scale = 1
	return m.recompute()
}

func (m Model) recompute() (tea.Model, tea.Cmd) {
	m.seq++
	m.encoding = true
	m.saved = ""
	return m, tea.Batch(m.encodeCmd(), m.ladderCmd())
}

func (m Model) save() (tea.Model, tea.Cmd) {
	if m.result == nil || m.deps.Save == nil {
		return m, nil
	}
	if err := m.deps.Save(m.output, m.result.Data); err != nil {
		m.err = err
		return m, nil
	}
	m.saved = m.output
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftW := int(float64(m.width) * 0.50)
	rightW := m.width - leftW

	previewContent := tuipanels.PanelTitle("PREVIEW", true) + "\n" + m.previewPanel.View(m.preview)
	left := tuipanels.WrapPanel(previewContent, true, leftW, m.height-1)

	w, h := optimize.TargetDimensions(m.src.Width, m.src.Height, m.opts.Scale, m.opts.MaxWidth, m.opts.MaxHeight)
	settingsContent := tuipanels.PanelTitle("SETTINGS", false) + "\n" + m.settingsPanel.View(imagepanels.SettingsView{
		Source:   m.src,
		Format:   m.opts.Format,
		Quality:  m.opts.Quality,
		Scale:    m.opts.Scale,
		Width:    w,
		Height:   h,
		Result:   m.result,
		Steps:    m.steps,
		Encoding: m.encoding,
		Err:      m.err,
		Saved:    m.saved,
	})
	right := tuipanels.WrapPanel(settingsContent, false, rightW, m.height-1)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	status := lipgloss.NewStyle().
		Background(lipgloss.Color(tuipanels.ColorBG)).
		Width(m.width).
		Render(statusBar(m, m.width))

	if m.showHelp {
		return tuipanels.HelpOverlay(body+"\n"+status, helpText(), tuipanels.ColorActiveBorder, 46)
	}
	return body + "\n" + status
}

func helpText() string {
	return fmt.Sprintf(`%s

%s
  ←/→  h/l   Quality down / up (steps of %d)
  ↑/↓  k/j   Scale up / down (steps of %d%%)
  f           Cycle output format
  r           Reset to quality 80 at full size

%s
  s           Save the current result
  ?           Toggle this help
  q / esc     Quit`,
		tuipanels.StyleTitle.Render("UNUM IMAGE — keyboard reference"),
		tuipanels.StyleTitle.Render("ADJUST"),
		qualityStep,
		int(scaleStep*100),
		tuipanels.StyleTitle.Render("ACTIONS"),
	)
}
