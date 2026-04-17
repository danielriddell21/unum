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

// Start launches the diff TUI.
func Start(d *node.Diff, theme string, version string) error { //nolint:cyclop,gocognit // NOSONAR: file-picker reload loop branches on multiple picker/parse outcomes; each case is necessary
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

		// Pick file A
		cwd, _ := os.Getwd()
		pmA := newPickerModel(cwd, "SELECT FILE A")
		ppA := tea.NewProgram(pmA, tea.WithAltScreen())
		prA, err := ppA.Run()
		if err != nil {
			return fmt.Errorf("file picker: %w", err)
		}
		pickedA, ok := prA.(pickerModel)
		if !ok || pickedA.Selected == "" {
			return nil
		}

		// Pick file B (start in same dir as file A)
		pmB := newPickerModel(filepath.Dir(pickedA.Selected), "SELECT FILE B")
		ppB := tea.NewProgram(pmB, tea.WithAltScreen())
		prB, err := ppB.Run()
		if err != nil {
			return fmt.Errorf("file picker: %w", err)
		}
		pickedB, ok := prB.(pickerModel)
		if !ok || pickedB.Selected == "" {
			return nil
		}

		// Re-read and re-diff
		dataA, err := os.ReadFile(pickedA.Selected)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", pickedA.Selected, err)
		}
		dataB, err := os.ReadFile(pickedB.Selected)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", pickedB.Selected, err)
		}

		diffFormat := format.Parse("", pickedA.Selected, pickedB.Selected)
		var newDiff *node.Diff
		switch diffFormat {
		case node.FormatJSON:
			newDiff, err = parse.JSON(dataA, dataB)
		case node.FormatYAML:
			newDiff, err = parse.YAML(dataA, dataB)
		case node.FormatTerraform:
			newDiff, err = parse.Terraform(dataA)
		default:
			newDiff, err = parse.Text(dataA, dataB, 3)
		}
		if err != nil {
			return fmt.Errorf("diff: %w", err)
		}
		newDiff.FileA = pickedA.Selected
		newDiff.FileB = pickedB.Selected
		newDiff.Format = diffFormat
		diff = newDiff
	}
}

func applyTheme(theme string) {
	ApplyPalette(panels.ResolvePalette(theme))
}
