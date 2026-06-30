package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"

	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

type pickerModel struct {
	fp       filepicker.Model
	Selected string
	height   int
}

func newPickerModel(dir string) pickerModel {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.CurrentDirectory = dir
	fp.ShowHidden = false
	fp.SetHeight(20)
	return pickerModel{fp: fp}
}

func (m pickerModel) Init() tea.Cmd {
	return m.fp.Init()
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.fp.SetHeight(m.height - 6)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.fp, cmd = m.fp.Update(msg)

	if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
		m.Selected = path
		return m, tea.Quit
	}

	return m, cmd
}

func (m pickerModel) View() string {
	header := tuipanels.StyleTitle.Render(" [ UNUM ] SELECT JSON FILE ") + "\n\n"
	hint := tuipanels.StyleHint.Render(fmt.Sprintf("  dir: %s\n\n  ↑↓ navigate · enter select · q quit", m.fp.CurrentDirectory))
	return header + m.fp.View() + "\n" + hint
}
