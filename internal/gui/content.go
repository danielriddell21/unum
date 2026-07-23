package gui

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/danielriddell21/unum/internal/diff/format"
	diffnode "github.com/danielriddell21/unum/internal/diff/node"
	diffparse "github.com/danielriddell21/unum/internal/diff/parse"
	diffstatic "github.com/danielriddell21/unum/internal/diff/render/static"
	"github.com/danielriddell21/unum/internal/hash/derive"
	jsonparse "github.com/danielriddell21/unum/internal/json/parse"
	jsonstatic "github.com/danielriddell21/unum/internal/json/render/static"
	"github.com/danielriddell21/unum/pkg/terraform"
)

type LineKind int

const (
	KindPlain LineKind = iota
	KindDim
	KindAdded
	KindRemoved
	KindModified
)

type Line struct {
	Text string
	Kind LineKind
}

func JSONLines(file string) []Line {
	data, err := os.ReadFile(file)
	if err != nil {
		return errorLines(err)
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if err := jsonparse.Validate(data); err != nil {
		return errorLines(err)
	}
	root, err := jsonparse.Parse(data)
	if err != nil {
		return errorLines(err)
	}

	// A zero Theme leaves every lipgloss style unset, so Render emits plain
	// text regardless of the terminal profile the process started under.
	var buf bytes.Buffer
	if err := jsonstatic.Render(&buf, root, jsonstatic.Options{NoColor: true, Quiet: true, ShowLineNums: true}); err != nil {
		return errorLines(err)
	}
	return plainLines(buf.String())
}

func DiffLines(fileA, fileB string) []Line {
	dataA, err := os.ReadFile(fileA)
	if err != nil {
		return errorLines(err)
	}
	dataB, err := os.ReadFile(fileB)
	if err != nil {
		return errorLines(err)
	}

	diffFormat := format.Parse("", fileA, fileB)
	if diffFormat == diffnode.FormatTerraform {
		return terraformLines(dataA)
	}

	var d *diffnode.Diff
	switch diffFormat {
	case diffnode.FormatJSON:
		d, err = diffparse.JSON(dataA, dataB)
	case diffnode.FormatYAML:
		d, err = diffparse.YAML(dataA, dataB)
	default:
		d, err = diffparse.Text(dataA, dataB, 3)
	}
	if err != nil {
		return errorLines(err)
	}
	d.FileA = fileA
	d.FileB = fileB

	var buf bytes.Buffer
	if err := diffstatic.Render(&buf, d, diffstatic.Options{NoColor: true, Context: 3}); err != nil {
		return errorLines(err)
	}
	lines := splitLines(buf.String())
	out := make([]Line, len(lines))
	for i, ln := range lines {
		out[i] = Line{Text: ln, Kind: classifyDiffLine(ln)}
	}
	return out
}

func terraformLines(data []byte) []Line {
	plan, err := terraform.Parse(data)
	if err != nil {
		return errorLines(err)
	}
	body := plan.RenderDiff(terraform.RenderOptions{})
	if body == "" {
		return []Line{{Text: "No changes.", Kind: KindDim}}
	}
	lines := splitLines(body)
	out := make([]Line, len(lines))
	for i, ln := range lines {
		out[i] = Line{Text: ln, Kind: classifyTerraformLine(ln)}
	}
	return out
}

func classifyDiffLine(ln string) LineKind {
	switch {
	case strings.HasPrefix(ln, "--- ") || strings.HasPrefix(ln, "+++ ") || strings.HasPrefix(ln, "@@"):
		return KindDim
	case strings.HasPrefix(ln, "+"):
		return KindAdded
	case strings.HasPrefix(ln, "-"):
		return KindRemoved
	case strings.HasPrefix(ln, "!"):
		return KindModified
	}
	// Hunk body lines carry a "old new " gutter before the change marker.
	if len(ln) > 10 {
		switch ln[10] {
		case '+':
			return KindAdded
		case '-':
			return KindRemoved
		}
	}
	return KindPlain
}

func classifyTerraformLine(ln string) LineKind {
	trimmed := strings.TrimLeft(ln, " ")
	if trimmed == "" {
		return KindPlain
	}
	switch trimmed[0] {
	case '+':
		return KindAdded
	case '-':
		return KindRemoved
	case '~':
		return KindModified
	case '#':
		return KindDim
	}
	return KindPlain
}

func HashLines(input string) []Line {
	if input == "" {
		return []Line{{Text: "type to derive", Kind: KindDim}}
	}
	r := derive.Derive(input)
	// The emoji row stays out: the window's 7x13 face has no emoji glyphs.
	rows := []struct{ label, value string }{
		{"port", fmt.Sprintf("%d", r.Port)},
		{"uuid", r.UUID},
		{"color", r.Color},
		{"short", r.Short},
		{"phrase", r.Phrase},
	}
	out := make([]Line, len(rows))
	for i, row := range rows {
		out[i] = Line{Text: fmt.Sprintf("%-7s  %s", row.label, row.value)}
	}
	return out
}

func errorLines(err error) []Line {
	return []Line{{Text: "error: " + err.Error(), Kind: KindRemoved}}
}

func plainLines(s string) []Line {
	lines := splitLines(s)
	out := make([]Line, len(lines))
	for i, ln := range lines {
		out[i] = Line{Text: ln}
	}
	return out
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}
