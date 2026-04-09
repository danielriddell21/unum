// Package tui provides the Bubble Tea interactive TUI for JSON navigation.
package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/danielriddell21/unum/internal/json/node"
)

// Start launches the TUI for the given node tree.
// It takes over the terminal (AltScreen) and restores it cleanly on exit.
func Start(root *node.Node, filename string) error {
	m := NewModel(root, filename)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // take over the full terminal, restore on exit
		tea.WithMouseCellMotion(), // optional: enable mouse support
	)

	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		return err
	}
	return nil
}
