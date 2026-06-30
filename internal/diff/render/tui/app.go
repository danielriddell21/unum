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
		m := NewModel(diff, version)
		p := tea.NewProgram(m, tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			return fmt.Errorf("diff TUI: %w", err)
		}
		final, ok := result.(Model)
		if !ok || !final.ReloadRequest {
			return nil
		}

		cwd, _ := os.Getwd()
		fileA, err := pickFile("SELECT FILE A", cwd)
		if err != nil {
			return err
		}
		if fileA == "" {
			return nil
		}

		fileB, err := pickFile("SELECT FILE B", filepath.Dir(fileA))
		if err != nil {
			return err
		}
		if fileB == "" {
			return nil
		}

		newDiff, err := reloadDiff(fileA, fileB)
		if err != nil {
			return err
		}
		diff = newDiff
	}
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
