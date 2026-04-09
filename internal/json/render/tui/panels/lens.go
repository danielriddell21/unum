package panels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/lens/query"
	"github.com/danielriddell21/unum/internal/json/lens/stats"
	"github.com/danielriddell21/unum/internal/json/lens/transform"
	"github.com/danielriddell21/unum/internal/json/lens/typegen"
	"github.com/danielriddell21/unum/internal/json/node"
)

// LensID identifies an active lens.
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

// TypeGenSubMode selects the typegen output language.
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

// LensPanel is the bottom-right panel — shows output from the active analysis lens.
type LensPanel struct {
	root         *node.Node
	cursorNode   *node.Node
	active       LensID
	typegenMode  TypeGenSubMode
	queryInput string
	content    string // cached rendered output
	viewport     viewport.Model
	width        int
	height       int
	focused      bool
	autoSwitched bool // true if current lens was auto-selected by cursor movement
}

// NewLensPanel creates a lens panel.
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

func (p *LensPanel) SetFocused(f bool)          { p.focused = f }
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

// HandleKey processes keypresses when the lens panel is focused.
// Returns true if a key was consumed.
func (p *LensPanel) HandleKey(key string) bool {
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

// ActiveLens returns the current lens ID.
func (p *LensPanel) ActiveLens() LensID { return p.active }

// View renders the lens panel.
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

	nc, ok := target.GetAnnotation(stats.Lens, "numeric_count")
	if !ok || nc.(int) == 0 {
		return styleMuted.Render(fmt.Sprintf("Array at %s has no numeric values.", target.Path()))
	}

	var sb strings.Builder
	path := styleHash.Render(target.Path())
	fmt.Fprintf(&sb, "Array: %s\n", path)

	count, _ := target.GetAnnotation(stats.Lens, "count")
	numCount, _ := target.GetAnnotation(stats.Lens, "numeric_count")
	fmt.Fprintf(&sb, "\nItems:   %v  (numeric: %v)\n", count, numCount)

	if v, ok := target.GetAnnotation(stats.Lens, "min"); ok {
		fmt.Fprintf(&sb, "Min:     %s\n", styleStats.Render(fmtF(v.(float64))))
	}
	if v, ok := target.GetAnnotation(stats.Lens, "max"); ok {
		fmt.Fprintf(&sb, "Max:     %s\n", styleStats.Render(fmtF(v.(float64))))
	}
	if v, ok := target.GetAnnotation(stats.Lens, "mean"); ok {
		fmt.Fprintf(&sb, "Mean:    %s\n", styleStats.Render(fmtF(v.(float64))))
	}
	if v, ok := target.GetAnnotation(stats.Lens, "stddev"); ok {
		fmt.Fprintf(&sb, "Std Dev: %s\n", styleStats.Render(fmtF(v.(float64))))
	}
	fmt.Fprintln(&sb)
	if v, ok := target.GetAnnotation(stats.Lens, "p50"); ok {
		fmt.Fprintf(&sb, "p50:     %s\n", styleMuted.Render(fmtF(v.(float64))))
	}
	if v, ok := target.GetAnnotation(stats.Lens, "p95"); ok {
		fmt.Fprintf(&sb, "p95:     %s\n", styleMuted.Render(fmtF(v.(float64))))
	}
	if v, ok := target.GetAnnotation(stats.Lens, "p99"); ok {
		fmt.Fprintf(&sb, "p99:     %s\n", styleMuted.Render(fmtF(v.(float64))))
	}

	return sb.String()
}

func fmtF(f float64) string {
	return fmt.Sprintf("%.4g", f)
}

// prettyJSON pretty-prints a node's value for the preview panel.
func PrettyNode(n *node.Node) string {
	b, err := n.MarshalJSON()
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	_ = json.Indent(&buf, b, "", "  ")
	return buf.String()
}

var (
	styleTabActive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4FF")).Bold(true)
	styleTabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("#3A3A3A"))
	styleHash        = lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD"))
	styleStats       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))
	styleSearchPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
)
