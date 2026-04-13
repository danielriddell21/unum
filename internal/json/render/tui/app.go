// Package tui provides the Bubble Tea interactive TUI for JSON navigation.
package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
	tuipanels "github.com/danielriddell21/unum/internal/tui/panels"
)

// applyTheme applies the named palette to all style vars.
func applyTheme(theme string) {
	ApplyPalette(tuipanels.ResolvePalette(theme))
}

// Start launches the TUI for the given node tree.
// It takes over the terminal (AltScreen) and restores it cleanly on exit.
// If the user presses 'o' to open a new file, the picker is shown and the
// explorer restarts with the selected file.
func Start(root *node.Node, filename string, theme string) error {
	applyTheme(theme)

	for { //nolint:dupl // mirrors startLoop body; Start applies theme first, startLoop skips it — extracting shared loop would require a theme parameter or closure
		m := NewModel(root, filename)
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

		result, err := p.Run()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			return fmt.Errorf("json TUI: %w", err)
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
			return fmt.Errorf("file picker: %w", err)
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
			return fmt.Errorf("parse: %w", err)
		}
		filename = picked.Selected
	}
}

// StartWithPicker launches a file picker TUI. Once a .json file is selected
// it parses it and transitions directly into the main JSON explorer.
func StartWithPicker(initialDir string, theme string) error {
	applyTheme(theme)

	pm := newPickerModel(initialDir)
	p := tea.NewProgram(pm, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("file picker: %w", err)
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
		return fmt.Errorf("parse: %w", err)
	}

	// Theme already applied — call Start with empty theme to skip re-applying
	return startLoop(root, picked.Selected)
}

// startLoop runs the main TUI loop without re-applying the theme.
func startLoop(root *node.Node, filename string) error {
	for { //nolint:dupl // mirrors Start body; Start applies theme first, startLoop skips it — extracting shared loop would require a theme parameter or closure
		m := NewModel(root, filename)
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

		result, err := p.Run()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			return fmt.Errorf("json TUI: %w", err)
		}

		final, ok := result.(Model)
		if !ok || !final.ReloadRequest {
			return nil
		}

		cwd, _ := os.Getwd()
		pm := newPickerModel(cwd)
		pp := tea.NewProgram(pm, tea.WithAltScreen())
		pr, err := pp.Run()
		if err != nil {
			return fmt.Errorf("file picker: %w", err)
		}
		picked, ok := pr.(pickerModel)
		if !ok || picked.Selected == "" {
			return nil
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
			return fmt.Errorf("parse: %w", err)
		}
		filename = picked.Selected
	}
}
