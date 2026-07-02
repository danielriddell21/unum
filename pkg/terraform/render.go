package terraform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type planLine struct {
	indent int
	marker string
	text   string
}

// RenderOptions configures [Plan.RenderDiff]. The zero value renders authentic,
// indented terraform plan output; callers apply their own colour.
type RenderOptions struct {
	// MarkerFirst places each +/-/~ change marker in column 0 instead of an
	// indented gutter, so that a GitHub-flavoured diff code block colours the
	// added and removed lines. The zero value keeps terraform's indentation.
	MarkerFirst bool

	// BangUpdates, in [RenderOptions.MarkerFirst] mode, marks update-in-place
	// lines with "!" instead of "~". A "!" prefix is the context-diff marker for
	// a changed line, which diff highlighters (Pygments, Rouge, GitHub) colour
	// while leaving "~" untouched. It has no effect unless MarkerFirst is set.
	BangUpdates bool
}

// RenderDiff renders only the per-resource changes as plain, uncoloured
// terraform plan-style text: per-resource headers, ~/+/- markers, "old -> new"
// transitions, "(known after apply)" values, nested blocks and
// "# (N unchanged … hidden)" counts.
//
// It emits no "Terraform will perform…" preamble and no "Plan:" footer, and it
// applies no colour; a caller composes the heading and footer (using the count
// accessors such as [Plan.AddCount]) and colours the output itself. RenderDiff
// returns the empty string when the plan has no changes. See [RenderOptions]
// for layout control.
func (p *Plan) RenderDiff(opts RenderOptions) string {
	changed := p.Changed()
	if len(changed) == 0 {
		return ""
	}

	var b strings.Builder
	for i, rc := range changed {
		if i > 0 {
			b.WriteByte('\n')
		}
		for _, ln := range rc.lines() {
			b.WriteString(formatLine(ln, opts))
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func formatLine(ln planLine, opts RenderOptions) string {
	lead := strings.Repeat("  ", ln.indent+1)

	if opts.MarkerFirst {
		switch ln.marker {
		case "+", "-":
			return ln.marker + " " + lead + ln.text
		case "~":
			marker := "~"
			if opts.BangUpdates {
				marker = "!"
			}
			return marker + " " + lead + ln.text
		case "#":
			return "  " + lead + "# " + ln.text
		default:
			return "  " + lead + ln.text
		}
	}

	gutter := "  "
	switch ln.marker {
	case "":
	case "#":
		gutter = "# "
	default:
		gutter = ln.marker + " "
	}
	return lead + gutter + ln.text
}

func (rc ResourceChange) lines() []planLine {
	marker := resourceMarker(rc.Action)
	body, _ := objectBody(rc.Before, rc.After, rc.AfterUnknown, rc.AfterSensitive, 1)
	out := make([]planLine, 0, len(body)+3)
	out = append(out,
		planLine{indent: 0, marker: "#", text: fmt.Sprintf("%s %s", rc.Address, rc.Action.phrase())},
		planLine{indent: 0, marker: marker, text: fmt.Sprintf("resource %q %q {", rc.Type, rc.Name)},
	)
	out = append(out, body...)
	out = append(out, planLine{indent: 0, marker: marker, text: "}"})
	return out
}

func resourceMarker(a Action) string {
	switch a {
	case ActionCreate:
		return "+"
	case ActionDelete:
		return "-"
	default:
		return "~"
	}
}

func objectBody(before, after, unknown, sensitive map[string]any, indent int) ([]planLine, bool) {
	keys := unionKeys(before, after, unknown)
	slices.Sort(keys)
	width := alignWidth(keys, before, after)

	var lines []planLine
	hidden := 0
	changed := false
	for _, k := range keys {
		bv, bok := before[k]
		av, aok := after[k]
		if !bok && !aok && valueAt(unknown, k) != true {
			continue
		}
		ls, ch := attribute(k, bv, bok, av, aok, valueAt(unknown, k), valueAt(sensitive, k), width, indent)
		if !ch {
			hidden++
			continue
		}
		changed = true
		lines = append(lines, ls...)
	}
	if hidden > 0 {
		noun := "attribute"
		if hidden > 1 {
			noun = "attributes"
		}
		lines = append(lines, planLine{indent: indent, marker: "#", text: fmt.Sprintf("(%d unchanged %s hidden)", hidden, noun)})
	}
	return lines, changed
}

func attribute(key string, bv any, bok bool, av any, aok bool, unknown, sensitive any, width, indent int) ([]planLine, bool) { //nolint:cyclop // branching on presence/type/equality is inherent to a per-attribute diff
	pad := padKey(key, width)

	if isTrue(unknown) {
		if bok {
			return []planLine{{indent, "~", fmt.Sprintf("%s = %s -> (known after apply)", pad, fmtValue(bv, false))}}, true
		}
		return []planLine{{indent, "+", fmt.Sprintf("%s = (known after apply)", pad)}}, true
	}

	switch {
	case !bok && aok:
		return addedLines(key, pad, av, sensitive, indent), true
	case bok && !aok:
		return removedLines(key, pad, bv, sensitive, indent), true
	}

	bMap, bIsMap := bv.(map[string]any)
	aMap, aIsMap := av.(map[string]any)
	if bIsMap && aIsMap {
		body, ch := objectBody(bMap, aMap, toMap(unknown), toMap(sensitive), indent+1)
		if !ch {
			return nil, false
		}
		return block(indent, "~", key, body), true
	}

	bList, bIsList := bv.([]any)
	aList, aIsList := av.([]any)
	if bIsList && aIsList {
		if reflect.DeepEqual(bList, aList) {
			return nil, false
		}
		return listDiff(pad, bList, aList, indent), true
	}

	if reflect.DeepEqual(bv, av) {
		return nil, false
	}
	sens := isTrue(sensitive)
	return []planLine{{indent, "~", fmt.Sprintf("%s = %s -> %s", pad, fmtValue(bv, sens), fmtValue(av, sens))}}, true
}

func block(indent int, marker, key string, body []planLine) []planLine {
	out := make([]planLine, 0, len(body)+2)
	out = append(out, planLine{indent, marker, key + " {"})
	out = append(out, body...)
	out = append(out, planLine{indent, marker, "}"})
	return out
}

func addedLines(key, pad string, v any, sensitive any, indent int) []planLine {
	if m, ok := v.(map[string]any); ok {
		body, _ := objectBody(nil, m, nil, toMap(sensitive), indent+1)
		return block(indent, "+", key, body)
	}
	if l, ok := v.([]any); ok {
		return listAll(pad, l, "+", indent)
	}
	return []planLine{{indent, "+", fmt.Sprintf("%s = %s", pad, fmtValue(v, isTrue(sensitive)))}}
}

func removedLines(key, pad string, v any, sensitive any, indent int) []planLine {
	if m, ok := v.(map[string]any); ok {
		body, _ := objectBody(m, nil, nil, toMap(sensitive), indent+1)
		return block(indent, "-", key, body)
	}
	if l, ok := v.([]any); ok {
		return listAll(pad, l, "-", indent)
	}
	return []planLine{{indent, "-", fmt.Sprintf("%s = %s -> null", pad, fmtValue(v, isTrue(sensitive)))}}
}

func listAll(pad string, list []any, marker string, indent int) []planLine {
	out := make([]planLine, 0, len(list)+2)
	out = append(out, planLine{indent, marker, pad + " = ["})
	for _, el := range list {
		out = append(out, planLine{indent + 1, marker, fmtValue(el, false) + ","})
	}
	out = append(out, planLine{indent, marker, "]"})
	return out
}

func listDiff(pad string, before, after []any, indent int) []planLine {
	out := []planLine{{indent, "~", pad + " = ["}}
	n := max(len(before), len(after))
	for i := range n {
		switch {
		case i < len(before) && i < len(after):
			if reflect.DeepEqual(before[i], after[i]) {
				out = append(out, planLine{indent + 1, "", fmtValue(after[i], false) + ","})
				continue
			}
			out = append(out, planLine{indent + 1, "-", fmtValue(before[i], false) + ","})
			out = append(out, planLine{indent + 1, "+", fmtValue(after[i], false) + ","})
		case i < len(before):
			out = append(out, planLine{indent + 1, "-", fmtValue(before[i], false) + ","})
		default:
			out = append(out, planLine{indent + 1, "+", fmtValue(after[i], false) + ","})
		}
	}
	out = append(out, planLine{indent, "~", "]"})
	return out
}

func fmtValue(v any, sensitive bool) string {
	if sensitive {
		return "(sensitive value)"
	}
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return strconv.Quote(t)
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case map[string]any, []any:
		return compactJSON(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func compactJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func padKey(key string, width int) string {
	if len(key) >= width {
		return key
	}
	return key + strings.Repeat(" ", width-len(key))
}

func alignWidth(keys []string, before, after map[string]any) int {
	w := 0
	for _, k := range keys {
		if isBlock(before[k]) || isBlock(after[k]) {
			continue
		}
		w = max(w, len(k))
	}
	return w
}

func isBlock(v any) bool {
	_, ok := v.(map[string]any)
	return ok
}

func unionKeys(sources ...map[string]any) []string {
	seen := map[string]bool{}
	var keys []string
	for _, m := range sources {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	return keys
}

func valueAt(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}

func toMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func isTrue(v any) bool {
	b, ok := v.(bool)
	return ok && b
}
