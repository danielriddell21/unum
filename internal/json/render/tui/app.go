// Package tui provides the Bubble Tea interactive TUI for JSON navigation.
package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

// Start launches the TUI for the given node tree.
// It takes over the terminal (AltScreen) and restores it cleanly on exit.
// If the user presses 'o' to open a new file, the picker is shown and the
// explorer restarts with the selected file.
func Start(root *node.Node, filename string) error {
	for {
		m := NewModel(root, filename)
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

		result, err := p.Run()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			return err
		}

		final, ok := result.(Model)
		if !ok || !final.ReloadRequest {
			return nil
		}

		// User wants to pick a different file.
		cwd, _ := os.Getwd()
		pm := newPickerModel(cwd)
		pp := tea.NewProgram(pm, tea.WithAltScreen())
		pr, err := pp.Run()
		if err != nil {
			return err
		}
		picked, ok := pr.(pickerModel)
		if !ok || picked.Selected == "" {
			return nil // picker dismissed — exit cleanly
		}

		data, err := os.ReadFile(picked.Selected)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", picked.Selected, err)
		}
		if err := parse.Validate(data); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "invalid JSON: %v\n", err)
			return nil
		}
		root, err = parse.Parse(data)
		if err != nil {
			return err
		}
		filename = picked.Selected
	}
}

// StartWithPicker launches a file picker TUI. Once a .json file is selected
// it parses it and transitions directly into the main JSON explorer.
func StartWithPicker(initialDir string) error {
	pm := newPickerModel(initialDir)
	p := tea.NewProgram(pm, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		return err
	}

	picked, ok := result.(pickerModel)
	if !ok || picked.Selected == "" {
		return nil // user quit without selecting
	}

	data, err := os.ReadFile(picked.Selected)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", picked.Selected, err)
	}
	if err := parse.Validate(data); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid JSON: %v\n", err)
		os.Exit(1)
	}
	root, err := parse.Parse(data)
	if err != nil {
		return err
	}

	return Start(root, picked.Selected)
}
