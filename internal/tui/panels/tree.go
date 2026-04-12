// Package panels contains shared TUI panel components and the palette system.
package panels

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/json/node"
)

// TreeNode is a flattened representation of a node for the tree panel.
// The flat list is pre-computed on expand/collapse; View() just slices it.
type TreeNode struct {
	Node     *node.Node
	Depth    int
	IsLast   bool
	Collapsed bool // only valid for KindObject and KindArray
}

// TreePanel is the left panel — a navigable, collapsible JSON tree.
type TreePanel struct {
	root      *node.Node
	flat      []TreeNode
	collapsed map[*node.Node]bool
	cursor    int
	viewport  viewport.Model
	width     int
	height    int
	focused   bool
	search    string // active search filter (empty = no filter)
}

// NewTreePanel creates a tree panel from the root node.
func NewTreePanel(root *node.Node, w, h int) TreePanel {
	p := TreePanel{
		root:      root,
		collapsed: make(map[*node.Node]bool),
		width:     w,
		height:    h,
	}
	p.viewport = viewport.New(w, h)
	p.reflatten()
	return p
}

// CursorNode returns the currently highlighted node.
func (p *TreePanel) CursorNode() *node.Node {
	if p.cursor >= 0 && p.cursor < len(p.flat) {
		return p.flat[p.cursor].Node
	}
	return nil
}

func (p *TreePanel) SetFocused(f bool) { p.focused = f }
func (p *TreePanel) SetSearch(s string) {
	p.search = s
	p.reflatten()
	if p.cursor >= len(p.flat) {
		p.cursor = max(0, len(p.flat)-1)
	}
}
func (p *TreePanel) Resize(w, h int) {
	p.width = w
	p.height = h
	p.viewport.Width = w
	p.viewport.Height = h
}

// Update handles keystrokes when the tree panel is focused.
func (p *TreePanel) Update(msg tea.KeyMsg) (changed bool) {
	switch msg.String() {
	case "j", "down":
		if p.cursor < len(p.flat)-1 {
			p.cursor++
			changed = true
		}
	case "k", "up":
		if p.cursor > 0 {
			p.cursor--
			changed = true
		}
	case "l", "right", "enter":
		if n := p.CursorNode(); n != nil {
			if n.Kind == node.KindObject || n.Kind == node.KindArray {
				p.collapsed[n] = false
				p.reflatten()
				changed = true
			}
		}
	case " ":
		if n := p.CursorNode(); n != nil {
			if n.Kind == node.KindObject || n.Kind == node.KindArray {
				p.collapsed[n] = !p.collapsed[n]
				p.reflatten()
				changed = true
			}
		}
	case "h", "left":
		if n := p.CursorNode(); n != nil {
			if (n.Kind == node.KindObject || n.Kind == node.KindArray) && !p.collapsed[n] {
				p.collapsed[n] = true
				p.reflatten()
				changed = true
			} else if n.Parent != nil {
				// Jump to parent
				for i, tn := range p.flat {
					if tn.Node == n.Parent {
						p.cursor = i
						break
					}
				}
				changed = true
			}
		}
	}
	return changed
}

// View renders the tree panel content.
func (p *TreePanel) View() string {
	if len(p.flat) == 0 {
		return styleMuted.Render("  (empty)")
	}

	var sb strings.Builder
	for i, tn := range p.flat {
		line := p.renderTreeNode(tn, i == p.cursor)
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	p.viewport.SetContent(sb.String())
	// Keep cursor visible
	p.scrollToCursor()

	return p.viewport.View()
}

func (p *TreePanel) scrollToCursor() {
	// Each line is 1 row. Scroll so cursor is visible.
	if p.cursor < p.viewport.YOffset {
		p.viewport.YOffset = p.cursor
	} else if p.cursor >= p.viewport.YOffset+p.height {
		p.viewport.YOffset = p.cursor - p.height + 1
	}
}

func (p *TreePanel) renderTreeNode(tn TreeNode, selected bool) string {
	n := tn.Node
	indent := strings.Repeat("  ", tn.Depth)

	// Connector
	var connector string
	if tn.Depth == 0 {
		connector = ""
	} else if tn.IsLast {
		connector = connectorLast
	} else {
		connector = connectorMid
	}

	// Key/index prefix
	var prefix string
	if n.Key != "" {
		prefix = styleObjectKey.Render(`"`+n.Key+`"`) + styleMuted.Render(": ")
	} else if n.Index >= 0 {
		prefix = styleArrayIdx.Render(fmt.Sprintf("[%d]", n.Index)) + styleMuted.Render(" ")
	}

	// Value
	var value string
	switch n.Kind {
	case node.KindObject:
		if tn.Collapsed {
			value = styleMuted.Render(fmt.Sprintf("{%d}", len(n.Children)))
		} else {
			value = connectorOpen + styleMuted.Render("{")
		}
	case node.KindArray:
		if tn.Collapsed {
			value = styleMuted.Render(fmt.Sprintf("[%d]", len(n.Children)))
		} else {
			value = connectorOpen + styleMuted.Render("[")
		}
	case node.KindString:
		var s string
		if json.Unmarshal([]byte(n.Raw), &s) == nil {
			display := s
			if len(display) > 40 {
				display = display[:37] + "..."
			}
			value = styleString.Render(`"` + display + `"`)
		} else {
			value = styleString.Render(n.Raw)
		}
	case node.KindNumber:
		value = styleNumber.Render(n.Raw)
	case node.KindBool:
		if n.Raw == "true" {
			value = styleBoolTrue.Render("true")
		} else {
			value = styleBoolFalse.Render("false")
		}
	case node.KindNull:
		value = styleNull.Render("null")
	}

	// Collapsed containers show ▶ instead of connector
	if (n.Kind == node.KindObject || n.Kind == node.KindArray) && tn.Collapsed {
		connector = strings.Replace(connector, "▶", "▶", 1)
		if tn.Depth > 0 {
			if tn.IsLast {
				connector = styleMuted.Render("└─▷ ")
			} else {
				connector = styleMuted.Render("├─▶ ")
			}
		}
		value = connectorClosed + value
	}

	line := indent + connector + prefix + value

	// Truncate to panel width
	w := p.width - 2
	if lipgloss.Width(line) > w && w > 3 {
		line = truncate(line, w)
	}

	if selected {
		line = styleCursor.Width(p.width - 2).Render(line)
	}

	return line
}

// reflatten recomputes the flat list from the root, applying collapse state and search.
func (p *TreePanel) reflatten() {
	p.flat = p.flat[:0]
	p.flatten(p.root, 0, true)
}

func (p *TreePanel) flatten(n *node.Node, depth int, isLast bool) {
	collapsed := p.collapsed[n] && (n.Kind == node.KindObject || n.Kind == node.KindArray)

	tn := TreeNode{
		Node:      n,
		Depth:     depth,
		IsLast:    isLast,
		Collapsed: collapsed,
	}

	// Search filter: skip nodes that don't match (but keep their ancestors)
	if p.search != "" && !p.nodeMatches(n) && !p.anyChildMatches(n) {
		return
	}

	p.flat = append(p.flat, tn)

	if !collapsed && (n.Kind == node.KindObject || n.Kind == node.KindArray) {
		for i, child := range n.Children {
			p.flatten(child, depth+1, i == len(n.Children)-1)
		}
	}
}

func (p *TreePanel) nodeMatches(n *node.Node) bool {
	if p.search == "" {
		return true
	}
	q := strings.ToLower(p.search)
	if strings.Contains(strings.ToLower(n.Key), q) {
		return true
	}
	if strings.Contains(strings.ToLower(n.Raw), q) {
		return true
	}
	return false
}

func (p *TreePanel) anyChildMatches(n *node.Node) bool {
	for _, c := range n.Children {
		if p.nodeMatches(c) || p.anyChildMatches(c) {
			return true
		}
	}
	return false
}

func truncate(s string, maxWidth int) string {
	w := lipgloss.Width(s)
	if w <= maxWidth {
		return s
	}
	// Strip ANSI and truncate
	plain := lipgloss.NewStyle().Render(s)
	if len(plain) > maxWidth-1 {
		plain = plain[:maxWidth-1]
	}
	return plain + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

