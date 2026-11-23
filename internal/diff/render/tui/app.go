package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/diff/format"
	"github.com/danielriddell21/unum/internal/diff/node"
	"github.com/danielriddell21/unum/internal/diff/parse"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

func Start(d *node.Diff, theme string, version string) error {
	applyTheme(theme)
	diff := d
	for {
		reload, err := runViewer(diff, version)
		if err != nil || !reload {
			return err
		}
		next, err := pickAndReload()
		if err != nil || next == nil {
			return err
		}
		diff = next
	}
}

func runViewer(d *node.Diff, version string) (bool, error) {
	p := tea.NewProgram(NewModel(d, version), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("diff TUI: %w", err)
	}
	final, ok := result.(Model)
	return ok && final.ReloadRequest, nil
}

func pickAndReload() (*node.Diff, error) {
	cwd, _ := os.Getwd()
	fileA, err := pickFile("SELECT FILE A", cwd)
	if err != nil || fileA == "" {
		return nil, err
	}
	fileB, err := pickFile("SELECT FILE B", filepath.Dir(fileA))
	if err != nil || fileB == "" {
		return nil, err
	}
	return reloadDiff(fileA, fileB)
}

func pickFile(title, startDir string) (string, error) {
	pm := newPickerModel(startDir, title)
	pp := tea.NewProgram(pm, tea.WithAltScreen())
	pr, err := pp.Run()
	if err != nil {
		return "", fmt.Errorf("file picker: %w", err)
	}
	picked, ok := pr.(pickerModel)
	if !ok || picked.Selected == "" {
		return "", nil
	}
	return picked.Selected, nil
}

func reloadDiff(fileA, fileB string) (*node.Diff, error) {
	dataA, err := os.ReadFile(fileA)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", fileA, err)
	}
	dataB, err := os.ReadFile(fileB)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", fileB, err)
	}
	diffFormat := format.Parse("", fileA, fileB)
	var d *node.Diff
	switch diffFormat {
	case node.FormatJSON:
		d, err = parse.JSON(dataA, dataB)
	case node.FormatYAML:
		d, err = parse.YAML(dataA, dataB)
	case node.FormatTerraform:
		d, err = parse.Terraform(dataA)
	default:
		d, err = parse.Text(dataA, dataB, 3)
	}
	if err != nil {
		return nil, fmt.Errorf("diff: %w", err)
	}
	d.FileA = fileA
	d.FileB = fileB
	d.Format = diffFormat
	return d, nil
}

func applyTheme(theme string) {
	ApplyPalette(panels.ResolvePalette(theme))
}
