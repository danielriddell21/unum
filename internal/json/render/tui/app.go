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

const (
	errFilePickerFmt = "file picker: %w"
	errParseFmt      = "parse: %w"
	errCannotReadFmt = "cannot read %s: %w"
	msgInvalidJSON   = "invalid JSON: %v\n"
)

// applyTheme applies the named palette to all style vars.
func applyTheme(theme string) {
	ApplyPalette(tuipanels.ResolvePalette(theme))
}

// Start launches the TUI for the given node tree.
// It takes over the terminal (AltScreen) and restores it cleanly on exit.
// If the user presses 'o' to open a new file, the picker is shown and the
// explorer restarts with the selected file.
func Start(root *node.Node, filename, theme, version string) error {
	applyTheme(theme)
	return startLoop(root, filename, version)
}

// StartWithPicker launches a file picker TUI. Once a .json file is selected
// it parses it and transitions directly into the main JSON explorer.
func StartWithPicker(initialDir, theme, version string) error {
	applyTheme(theme)

	pm := newPickerModel(initialDir)
	p := tea.NewProgram(pm, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		return fmt.Errorf(errFilePickerFmt, err)
	}

	picked, ok := result.(pickerModel)
	if !ok || picked.Selected == "" {
		return nil // user quit without selecting
	}

	data, err := os.ReadFile(picked.Selected)
	if err != nil {
		return fmt.Errorf(errCannotReadFmt, picked.Selected, err)
	}
	if err := parse.Validate(data); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, msgInvalidJSON, err)
		os.Exit(1)
	}
	root, err := parse.Parse(data)
	if err != nil {
		return fmt.Errorf(errParseFmt, err)
	}

	return startLoop(root, picked.Selected, version)
}

// startLoop runs the main TUI loop. Theme must already have been applied.
func startLoop(root *node.Node, filename, version string) error {
	for {
		m := NewModel(root, filename, version)
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

		newRoot, newFilename, ok, err := promptAndReload()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		root, filename = newRoot, newFilename
	}
}

// promptAndReload shows the file picker and parses the selected JSON file.
// Returns ok=false (with nil err) when the user cancels or validation fails —
// the caller should stop the reload loop without surfacing an error.
func promptAndReload() (*node.Node, string, bool, error) {
	cwd, _ := os.Getwd()
	pm := newPickerModel(cwd)
	pp := tea.NewProgram(pm, tea.WithAltScreen())
	pr, err := pp.Run()
	if err != nil {
		return nil, "", false, fmt.Errorf(errFilePickerFmt, err)
	}
	picked, ok := pr.(pickerModel)
	if !ok || picked.Selected == "" {
		return nil, "", false, nil
	}

	data, err := os.ReadFile(picked.Selected)
	if err != nil {
		return nil, "", false, fmt.Errorf(errCannotReadFmt, picked.Selected, err)
	}
	if err := parse.Validate(data); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, msgInvalidJSON, err)
		return nil, "", false, nil
	}
	root, err := parse.Parse(data)
	if err != nil {
		return nil, "", false, fmt.Errorf(errParseFmt, err)
	}
	return root, picked.Selected, true, nil
}
