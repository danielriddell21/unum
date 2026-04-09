package panels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/json/node"
)

// PreviewPanel shows the value of the currently selected tree node.
type PreviewPanel struct {
	viewport viewport.Model
	width    int
	height   int
	focused  bool
}

// NewPreviewPanel creates a preview panel.
func NewPreviewPanel(w, h int) PreviewPanel {
	return PreviewPanel{
		viewport: viewport.New(w, h),
		width:    w,
		height:   h,
	}
}

func (p *PreviewPanel) SetFocused(f bool) { p.focused = f }

func (p *PreviewPanel) Resize(w, h int) {
	p.width = w
	p.height = h
	p.viewport.Width = w
	p.viewport.Height = h
}

// SetNode updates the preview to show the given node's value.
func (p *PreviewPanel) SetNode(n *node.Node) {
	if n == nil {
		p.viewport.SetContent(styleMuted.Render("(empty)"))
		return
	}
	p.viewport.SetContent(p.renderNode(n))
	p.viewport.GotoTop()
}

// HandleKey handles scrolling when this panel is focused.
func (p *PreviewPanel) HandleKey(key string) bool {
	switch key {
	case "j", "down":
		p.viewport.ScrollDown(1)
		return true
	case "k", "up":
		p.viewport.ScrollUp(1)
		return true
	}
	return false
}

// View returns the rendered preview content.
func (p *PreviewPanel) View() string {
	return p.viewport.View()
}

func (p *PreviewPanel) renderNode(n *node.Node) string {
	switch n.Kind {
	case node.KindObject, node.KindArray:
		return p.renderContainer(n)
	default:
		return p.renderLeaf(n)
	}
}

func (p *PreviewPanel) renderContainer(n *node.Node) string {
	b, err := n.MarshalJSON()
	if err != nil {
		return styleError2.Render("error: " + err.Error())
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, b, "", "  "); err != nil {
		return string(b)
	}
	// Apply basic coloring to the pretty-printed JSON
	return colorizeJSON(pretty.String())
}

func (p *PreviewPanel) renderLeaf(n *node.Node) string {
	var typeLabel string
	var valueStr string

	switch n.Kind {
	case node.KindString:
		typeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render("STRING")
		var s string
		_ = json.Unmarshal([]byte(n.Raw), &s)
		valueStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")).Render(`"` + s + `"`)
	case node.KindNumber:
		typeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render("NUMBER")
		valueStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).Render(n.Raw)
	case node.KindBool:
		typeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render("BOOL")
		if n.Raw == "true" {
			valueStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#56B6C2")).Render("true")
		} else {
			valueStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Render("false")
		}
	case node.KindNull:
		typeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render("NULL")
		valueStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")).Render("null")
	}

	path := lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")).Render(n.Path())

	return fmt.Sprintf("%s\n\n%s\n\n%s", typeLabel, valueStr, path)
}

// colorizeJSON applies basic ANSI colors to a pretty-printed JSON string.
// This is a lightweight line-by-line colorizer (not a full lexer).
func colorizeJSON(s string) string {
	lines := strings.Split(s, "\n")
	colored := make([]string, 0, len(lines))
	for _, line := range lines {
		colored = append(colored, colorizeLine(line))
	}
	return strings.Join(colored, "\n")
}

func colorizeLine(line string) string {
	trimmed := strings.TrimSpace(line)
	indent := line[:len(line)-len(strings.TrimLeft(line, " "))]

	// Object key: `"key": ...`
	if strings.HasPrefix(trimmed, `"`) && strings.Contains(trimmed, `":`) {
		colonIdx := strings.Index(trimmed, `":`)
		if colonIdx > 0 {
			key := trimmed[:colonIdx+1]  // includes closing quote
			rest := trimmed[colonIdx+1:] // ": value..."
			coloredKey := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")).Render(key)
			coloredRest := colorizeValue(strings.TrimSpace(rest[1:])) // strip ":"
			return indent + coloredKey + lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render(": ") + coloredRest
		}
	}

	return indent + colorizeValue(trimmed)
}

func colorizeValue(s string) string {
	// Strip trailing comma
	suffix := ""
	plain := s
	if strings.HasSuffix(plain, ",") {
		suffix = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render(",")
		plain = plain[:len(plain)-1]
	}

	switch {
	case strings.HasPrefix(plain, `"`):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")).Render(plain) + suffix
	case plain == "true":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#56B6C2")).Render(plain) + suffix
	case plain == "false":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Render(plain) + suffix
	case plain == "null":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")).Render(plain) + suffix
	case plain == "{" || plain == "}" || plain == "[" || plain == "]" ||
		plain == "{}" || plain == "[]":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A")).Render(plain) + suffix
	default:
		// Assume number
		if len(plain) > 0 && (plain[0] == '-' || (plain[0] >= '0' && plain[0] <= '9')) {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).Render(plain) + suffix
		}
		return plain + suffix
	}
}

var styleError2 = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))
