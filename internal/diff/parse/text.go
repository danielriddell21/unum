package parse

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/danielriddell21/unum/internal/diff/node"
)

func Text(a, b []byte, contextLines int) (*node.Diff, error) {
	dmp := diffmatchpatch.New()

	aStr := string(a)
	bStr := string(b)

	aChars, bChars, lineArray := dmp.DiffLinesToChars(aStr, bStr)
	rawDiffs := dmp.DiffMain(aChars, bChars, false)
	rawDiffs = dmp.DiffCharsToLines(rawDiffs, lineArray)

	hunks, added, removed := buildHunks(rawDiffs, contextLines)

	return &node.Diff{
		Format:  node.FormatText,
		Hunks:   hunks,
		Added:   added,
		Removed: removed,
	}, nil
}

func buildHunks(diffs []diffmatchpatch.Diff, ctx int) ([]node.Hunk, int, int) { //nolint:cyclop,gocognit // NOSONAR: processes every diff operation type in a single pass; branching on op kind is the algorithm
	// Expand diffs into a flat line list with change kind
	type rawLine struct {
		kind    node.ChangeKind
		content string
	}
	var raw []rawLine

	for _, d := range diffs {
		lines := splitLines(d.Text)
		var kind node.ChangeKind
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			kind = node.Added
		case diffmatchpatch.DiffDelete:
			kind = node.Removed
		default:
			kind = node.Unchanged
		}
		for _, l := range lines {
			raw = append(raw, rawLine{kind: kind, content: l})
		}
	}

	// Assign old/new line numbers
	type numberedLine struct {
		kind    node.ChangeKind
		oldNum  int
		newNum  int
		content string
	}
	numbered := make([]numberedLine, 0, len(raw))
	oldN, newN := 1, 1
	added, removed := 0, 0
	for _, r := range raw {
		nl := numberedLine{kind: r.kind, content: r.content}
		switch r.kind {
		case node.Added:
			nl.newNum = newN
			newN++
			added++
		case node.Removed:
			nl.oldNum = oldN
			oldN++
			removed++
		default:
			nl.oldNum = oldN
			nl.newNum = newN
			oldN++
			newN++
		}
		numbered = append(numbered, nl)
	}

	// Find changed line indices
	var changedIdx []int
	for i, l := range numbered {
		if l.kind != node.Unchanged {
			changedIdx = append(changedIdx, i)
		}
	}
	if len(changedIdx) == 0 {
		return nil, 0, 0
	}

	// Group changed indices into hunk ranges (with context)
	type hunkRange struct{ start, end int }
	var ranges []hunkRange
	cur := hunkRange{
		start: max(0, changedIdx[0]-ctx),
		end:   min(len(numbered)-1, changedIdx[0]+ctx),
	}
	for _, idx := range changedIdx[1:] {
		lo := max(0, idx-ctx)
		hi := min(len(numbered)-1, idx+ctx)
		if lo <= cur.end+1 {
			cur.end = hi
		} else {
			ranges = append(ranges, cur)
			cur = hunkRange{start: lo, end: hi}
		}
	}
	ranges = append(ranges, cur)

	// Build Hunks
	var hunks []node.Hunk
	for _, r := range ranges {
		slice := numbered[r.start : r.end+1]
		h := node.Hunk{}
		// Compute old/new start from first real old/new line in slice
		for _, l := range slice {
			if l.oldNum > 0 && h.OldStart == 0 {
				h.OldStart = l.oldNum
			}
			if l.newNum > 0 && h.NewStart == 0 {
				h.NewStart = l.newNum
			}
		}
		if h.OldStart == 0 {
			h.OldStart = 1
		}
		if h.NewStart == 0 {
			h.NewStart = 1
		}

		for _, l := range slice {
			if l.kind != node.Added {
				h.OldCount++
			}
			if l.kind != node.Removed {
				h.NewCount++
			}
			h.Lines = append(h.Lines, node.Line{
				Kind:    l.kind,
				OldNum:  l.oldNum,
				NewNum:  l.newNum,
				Content: l.content,
			})
		}
		hunks = append(hunks, h)
	}

	return hunks, added, removed
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// go-diff includes a trailing empty string when text ends with \n
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	// Strip \r so Windows CRLF line endings don't corrupt terminal output
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, "\r")
	}
	return lines
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
