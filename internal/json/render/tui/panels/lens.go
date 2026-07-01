package panels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"

	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/lens/transform"
	"github.com/danielriddell21/unum/internal/json/lens/typegen"
	"github.com/danielriddell21/unum/internal/json/node"
)

type LensID int

const (
	LensTypeGen LensID = iota + 1
	LensSchema
	LensYAML
	LensMerkle
	LensStats
	LensQuery
)

var lensNames = map[LensID]string{
	LensTypeGen: "TypeGen",
	LensSchema:  "Schema",
	LensYAML:    "YAML",
	LensMerkle:  "Merkle",
	LensStats:   "Stats",
	LensQuery:   "Query",
}

type TypeGenSubMode int

const (
	TypeGenGo TypeGenSubMode = iota
	TypeGenTS
	TypeGenJSONSchema
)

var typegenLabels = map[TypeGenSubMode]string{
	TypeGenGo:         "Go",
	TypeGenTS:         "TypeScript",
	TypeGenJSONSchema: "JSON Schema",
}

type LensPanel struct {
	root         *node.Node
	cursorNode   *node.Node
	active       LensID
	typegenMode  TypeGenSubMode
	queryInput   string
	content      string
	viewport     viewport.Model
	width        int
	height       int
	focused      bool
	autoSwitched bool
}

func NewLensPanel(root *node.Node, w, h int) LensPanel {
	p := LensPanel{
		root:   root,
		active: LensTypeGen,
		width:  w,
		height: h,
	}
	p.viewport = viewport.New(w, h-3) // reserve rows for tab strip
	p.refresh(root)
	return p
}

func (p *LensPanel) SetFocused(f bool) { p.focused = f }
func (p *LensPanel) SetCursorNode(n *node.Node) {
	prev := p.cursorNode
	p.cursorNode = n

	// Auto-switch to Stats when cursor lands on a numeric array
	if n != nil && n != prev && !p.autoSwitched {
		if n.Kind == node.KindArray {
			if v, ok := n.GetAnnotation(stats.Lens, "numeric_count"); ok && v.(int) > 0 {
				p.active = LensStats
				p.autoSwitched = true
				p.refresh(n)
				return
			}
		}
	}
	p.autoSwitched = false
	p.refresh(n)
}

func (p *LensPanel) Resize(w, h int) {
	p.width = w
	p.height = h
	p.viewport.Width = w
	p.viewport.Height = max(1, h-3)
}

func (p *LensPanel) HandleKey(key string) bool { //nolint:cyclop // key dispatch for every lens action; each case handles a distinct user intent
	// Lens switching via digits
	switch key {
	case "1":
		p.active = LensTypeGen
		p.autoSwitched = false
		p.refresh(p.cursorNode)
		return true
	case "2":
		p.active = LensSchema
		p.autoSwitched = false
		p.refresh(p.cursorNode)
		return true
	case "3":
		p.active = LensYAML
		p.autoSwitched = false
		p.refresh(p.cursorNode)
		return true
	case "4":
		p.active = LensMerkle
		p.autoSwitched = false
		p.refresh(p.cursorNode)
		return true
	case "5":
		p.active = LensStats
		p.autoSwitched = false
		p.refresh(p.cursorNode)
		return true
	case "6":
		p.active = LensQuery
		p.autoSwitched = false
		p.refresh(p.cursorNode)
		return true
	}

	// TypeGen sub-mode (when TypeGen is active)
	if p.active == LensTypeGen {
		switch key {
		case "g":
			p.typegenMode = TypeGenGo
			p.refresh(p.cursorNode)
			return true
		case "t":
			p.typegenMode = TypeGenTS
			p.refresh(p.cursorNode)
			return true
		case "j":
			p.typegenMode = TypeGenJSONSchema
			p.refresh(p.cursorNode)
			return true
		}
	}

	// Query input
	if p.active == LensQuery {
		switch key {
		case "enter":
			p.refresh(p.cursorNode)
			return true
		case "backspace":
			if len(p.queryInput) > 0 {
				p.queryInput = p.queryInput[:len(p.queryInput)-1]
			}
			return true
		case "esc", "ctrl+c":
			p.queryInput = ""
			p.refresh(p.cursorNode)
			return false // let parent handle quit
		default:
			if len(key) == 1 {
				p.queryInput += key
				return true
			}
		}
	}

	// Scrolling
	switch key {
	case "j", "down":
		p.viewport.ScrollDown(1)
		return true
	case "k", "up":
		p.viewport.ScrollUp(1)
		return true
	case "d", "ctrl+d":
		p.viewport.HalfPageDown()
		return true
	case "u", "ctrl+u":
		p.viewport.HalfPageUp()
		return true
	}

	return false
}

func (p *LensPanel) ActiveLens() LensID { return p.active }

func (p *LensPanel) View() string {
	tabs := p.renderTabs()
	p.viewport.SetContent(p.content)
	return tabs + "\n" + p.viewport.View()
}

func (p *LensPanel) renderTabs() string {
	var parts []string
	for _, id := range []LensID{LensTypeGen, LensSchema, LensYAML, LensMerkle, LensStats, LensQuery} {
		num := int(id)
		name := lensNames[id]
		label := fmt.Sprintf("[%d:%s]", num, name)

		if id == p.active {
			extra := ""
			if id == LensTypeGen {
				extra = " > " + typegenLabels[p.typegenMode]
			}
			parts = append(parts, styleTabActive.Render(label+extra))
		} else {
			parts = append(parts, styleTabInactive.Render(label))
		}
	}
	return " " + strings.Join(parts, styleMuted.Render("  "))
}

func (p *LensPanel) refresh(contextNode *node.Node) {
	target := p.root
	if contextNode != nil {
		target = contextNode
	}

	var out string
	var err error

	switch p.active {
	case LensTypeGen:
		tgt := typegen.TargetGo
		switch p.typegenMode {
		case TypeGenTS:
			tgt = typegen.TargetTypeScript
		case TypeGenJSONSchema:
			tgt = typegen.TargetJSONSchema
		}
		out, err = typegen.Generate(p.root, typegen.Options{Target: tgt})

	case LensSchema:
		out, err = typegen.Generate(p.root, typegen.Options{Target: typegen.TargetJSONSchema})

	case LensYAML:
		var b []byte
		b, err = transform.ToYAML(p.root)
		if err == nil {
			out = string(b)
		}

	case LensMerkle:
		out = renderMerkle(p.root)

	case LensStats:
		out = renderStats(target)

	case LensQuery:
		if p.queryInput == "" {
			out = styleSearchPrompt.Render("[ SCANNING > ]") + "\n\nEnter a jq expression below.\n\n" +
				styleMuted.Render("Examples:\n  .users[].name\n  .[] | select(.active == true)\n  keys")
		} else {
			result, qErr := query.Execute(p.root, p.queryInput)
			if qErr != nil {
				out = styleError.Render("✗ "+qErr.Error()) + "\n\nExpression: " + styleSearchPrompt.Render(p.queryInput)
			} else {
				out = result
			}
		}
	}

	if err != nil {
		p.content = styleError.Render("✗ " + err.Error())
	} else {
		p.content = out
	}
	p.viewport.GotoTop()
}

func renderMerkle(root *node.Node) string {
	rootHash := merkle.RootHash(root)
	if rootHash == "" {
		return styleMuted.Render("Run with --merkle to enable hash annotations.\n\nOr launch with: unum json <file> --ui --merkle")
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Root: %s\n\n", styleHash.Render(rootHash))
	fmt.Fprintln(&sb, "Selected subtree hashes:")
	fmt.Fprintln(&sb, styleMuted.Render("(navigate the tree to inspect hashes)"))
	return sb.String()
}

func renderStats(n *node.Node) string {
	if n == nil {
		return styleMuted.Render("Navigate to an array node to see statistics.")
	}

	// Find the nearest array ancestor or self
	target := n
	for target != nil && target.Kind != node.KindArray {
		target = target.Parent
	}
	if target == nil {
		return styleMuted.Render("No array node at or above cursor.\n\nNavigate to an array with numeric values.")
	}

	s, ok := stats.Get(target)
	if !ok || s.NumericCount == 0 {
		return styleMuted.Render(fmt.Sprintf("Array at %s has no numeric values.", target.Path()))
	}

	var sb strings.Builder
	path := styleHash.Render(target.Path())
	fmt.Fprintf(&sb, "Array: %s\n", path)
	fmt.Fprintf(&sb, "\nItems:   %d  (numeric: %d)\n", s.Count, s.NumericCount)
	fmt.Fprintf(&sb, "Min:     %s\n", styleStats.Render(fmtF(s.Min)))
	fmt.Fprintf(&sb, "Max:     %s\n", styleStats.Render(fmtF(s.Max)))
	fmt.Fprintf(&sb, "Mean:    %s\n", styleStats.Render(fmtF(s.Mean)))
	fmt.Fprintf(&sb, "Std Dev: %s\n", styleStats.Render(fmtF(s.Stddev)))
	fmt.Fprintln(&sb)
	fmt.Fprintf(&sb, "p50:     %s\n", styleMuted.Render(fmtF(s.P50)))
	fmt.Fprintf(&sb, "p95:     %s\n", styleMuted.Render(fmtF(s.P95)))
	fmt.Fprintf(&sb, "p99:     %s\n", styleMuted.Render(fmtF(s.P99)))

	return sb.String()
}

func fmtF(f float64) string {
	return fmt.Sprintf("%.4g", f)
}

func PrettyNode(n *node.Node) string {
	b, err := n.MarshalJSON()
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	_ = json.Indent(&buf, b, "", "  ")
	return buf.String()
}
