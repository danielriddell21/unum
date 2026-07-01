package panels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"

	"github.com/danielriddell21/unum/internal/json/node"
)

type PreviewPanel struct {
	viewport viewport.Model
	width    int
	height   int
	focused  bool
}

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

func (p *PreviewPanel) SetNode(n *node.Node) {
	if n == nil {
		p.viewport.SetContent(styleMuted.Render("(empty)"))
		return
	}
	p.viewport.SetContent(p.renderNode(n))
	p.viewport.GotoTop()
}

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
		return styleError.Render("error: " + err.Error())
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, b, "", "  "); err != nil {
		return string(b)
	}
	return colorizeJSON(pretty.String())
}

func (p *PreviewPanel) renderLeaf(n *node.Node) string {
	var typeLabel string
	var valueStr string

	switch n.Kind {
	case node.KindString:
		typeLabel = styleMuted.Render("STRING")
		var s string
		_ = json.Unmarshal([]byte(n.Raw), &s)
		valueStr = styleString.Render(`"` + s + `"`)
	case node.KindNumber:
		typeLabel = styleMuted.Render("NUMBER")
		valueStr = styleNumber.Render(n.Raw)
	case node.KindBool:
		typeLabel = styleMuted.Render("BOOL")
		if n.Raw == "true" {
			valueStr = styleBoolTrue.Render("true")
		} else {
			valueStr = styleBoolFalse.Render("false")
		}
	case node.KindNull:
		typeLabel = styleMuted.Render("NULL")
		valueStr = styleNull.Render("null")
	}

	path := stylePath.Render(n.Path())
	return fmt.Sprintf("%s\n\n%s\n\n%s", typeLabel, valueStr, path)
}

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
			coloredKey := styleObjectKey.Render(key)
			coloredRest := colorizeValue(strings.TrimSpace(rest[1:])) // strip ":"
			return indent + coloredKey + styleMuted.Render(": ") + coloredRest
		}
	}

	return indent + colorizeValue(trimmed)
}

func colorizeValue(s string) string { //nolint:cyclop // colorizes every JSON value kind by prefix-matching; each case maps a syntax token to a colour
	// Strip trailing comma
	suffix := ""
	plain := s
	if strings.HasSuffix(plain, ",") {
		suffix = styleMuted.Render(",")
		plain = plain[:len(plain)-1]
	}

	switch {
	case strings.HasPrefix(plain, `"`):
		return styleString.Render(plain) + suffix
	case plain == "true":
		return styleBoolTrue.Render(plain) + suffix
	case plain == "false":
		return styleBoolFalse.Render(plain) + suffix
	case plain == "null":
		return styleNull.Render(plain) + suffix
	case plain == "{" || plain == "}" || plain == "[" || plain == "]" ||
		plain == "{}" || plain == "[]":
		return styleMuted.Render(plain) + suffix
	default:
		// Assume number
		if len(plain) > 0 && (plain[0] == '-' || (plain[0] >= '0' && plain[0] <= '9')) {
			return styleNumber.Render(plain) + suffix
		}
		return plain + suffix
	}
}
