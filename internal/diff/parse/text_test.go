package parse

import (
	"testing"

	"github.com/danielriddell21/unum/internal/diff/node"
)

const (
	threeLinesText = "line1\nline2\nline3\n"
	hunksWant1     = "got %d hunks, want 1"
)

func TestText_Identical(t *testing.T) {
	a := []byte(threeLinesText)
	d, err := Text(a, a, 3)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 0 {
		t.Errorf("added=%d removed=%d, want 0 0", d.Added, d.Removed)
	}
	if len(d.Hunks) != 0 {
		t.Errorf("got %d hunks, want 0", len(d.Hunks))
	}
}

func TestText_AddedLines(t *testing.T) {
	a := []byte("line1\nline2\n")
	b := []byte(threeLinesText)
	d, err := Text(a, b, 3)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 || d.Removed != 0 {
		t.Errorf("added=%d removed=%d, want 1 0", d.Added, d.Removed)
	}
	if len(d.Hunks) != 1 {
		t.Fatalf(hunksWant1, len(d.Hunks))
	}
	// The added line should have a new line number
	var found bool
	for _, l := range d.Hunks[0].Lines {
		if l.Kind == node.Added {
			found = true
			if l.NewNum == 0 {
				t.Error("added line has NewNum=0, want >0")
			}
		}
	}
	if !found {
		t.Error("no Added line found in hunk")
	}
}

func TestText_RemovedLines(t *testing.T) {
	a := []byte(threeLinesText)
	b := []byte("line1\nline3\n")
	d, err := Text(a, b, 3)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 0 || d.Removed != 1 {
		t.Errorf("added=%d removed=%d, want 0 1", d.Added, d.Removed)
	}
}

func TestText_ModifiedLine(t *testing.T) {
	// A changed line appears as one Removed + one Added
	a := []byte("aaa\nbbb\nccc\n")
	b := []byte("aaa\nBBB\nccc\n")
	d, err := Text(a, b, 0)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 || d.Removed != 1 {
		t.Errorf("added=%d removed=%d, want 1 1", d.Added, d.Removed)
	}
	if len(d.Hunks) != 1 {
		t.Fatalf(hunksWant1, len(d.Hunks))
	}
	h := d.Hunks[0]
	var gotRemoved, gotAdded bool
	for _, l := range h.Lines {
		if l.Kind == node.Removed && l.Content == "bbb" {
			gotRemoved = true
		}
		if l.Kind == node.Added && l.Content == "BBB" {
			gotAdded = true
		}
	}
	if !gotRemoved {
		t.Error("missing Removed line with content 'bbb'")
	}
	if !gotAdded {
		t.Error("missing Added line with content 'BBB'")
	}
}

func TestText_LineNumbers(t *testing.T) {
	// "b" on line 2 becomes "X"; verify OldNum/NewNum assignments.
	a := []byte("a\nb\nc\n")
	b := []byte("a\nX\nc\n")
	d, err := Text(a, b, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) != 1 {
		t.Fatalf(hunksWant1, len(d.Hunks))
	}
	h := d.Hunks[0]
	for _, l := range h.Lines {
		switch l.Kind {
		case node.Removed:
			if l.OldNum != 2 {
				t.Errorf("removed line OldNum=%d, want 2", l.OldNum)
			}
		case node.Added:
			if l.NewNum != 2 {
				t.Errorf("added line NewNum=%d, want 2", l.NewNum)
			}
		}
	}
}

func TestText_HunkCounts(t *testing.T) {
	// Verify OldStart, OldCount, NewStart, NewCount on a simple hunk.
	a := []byte("x\ny\nz\n") // 3 lines
	b := []byte("x\nY\nz\n") // line 2 changed
	d, err := Text(a, b, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) != 1 {
		t.Fatalf(hunksWant1, len(d.Hunks))
	}
	h := d.Hunks[0]
	// With context=1: hunk covers lines 1-3 (all of them)
	if h.OldStart != 1 {
		t.Errorf("OldStart=%d, want 1", h.OldStart)
	}
	if h.NewStart != 1 {
		t.Errorf("NewStart=%d, want 1", h.NewStart)
	}
	// OldCount = unchanged + removed = 3; NewCount = unchanged + added = 3
	if h.OldCount != 3 {
		t.Errorf("OldCount=%d, want 3", h.OldCount)
	}
	if h.NewCount != 3 {
		t.Errorf("NewCount=%d, want 3", h.NewCount)
	}
}

func TestText_MultipleHunks(t *testing.T) {
	// Changes at lines 1 and 20 are far enough apart (>2*ctx) to produce 2 hunks.
	lines := make([]byte, 0, 100)
	for i := 0; i < 20; i++ {
		lines = append(lines, []byte("unchanged\n")...)
	}
	a := append([]byte("first\n"), lines...)
	a = append(a, []byte("last\n")...)
	b := append([]byte("FIRST\n"), lines...)
	b = append(b, []byte("LAST\n")...)

	d, err := Text(a, b, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) != 2 {
		t.Errorf("got %d hunks, want 2", len(d.Hunks))
	}
}

func TestText_HunksAreMergedWhenClose(t *testing.T) {
	// Two changes with fewer unchanged lines between them than 2*context
	// collapse into one hunk.
	a := []byte("A\nb\nC\n") // lines 1-3; 1 and 3 will change
	b := []byte("X\nb\nY\n")
	d, err := Text(a, b, 3) // context=3 — the gap of 1 line is within context
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) != 1 {
		t.Errorf("got %d hunks, want 1 (changes too close to split)", len(d.Hunks))
	}
}

func TestText_CRLFStripped(t *testing.T) {
	// CRLF line endings must not appear in line content.
	a := []byte("line1\r\nline2\r\n")
	b := []byte("line1\r\nLINE2\r\n")
	d, err := Text(a, b, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Hunks) != 1 {
		t.Fatalf(hunksWant1, len(d.Hunks))
	}
	for _, l := range d.Hunks[0].Lines {
		for _, ch := range l.Content {
			if ch == '\r' {
				t.Errorf("line content contains \\r: %q", l.Content)
			}
		}
	}
}

func TestText_EmptyFiles(t *testing.T) {
	a := []byte("")
	b := []byte("new line\n")
	d, err := Text(a, b, 3)
	if err != nil {
		t.Fatal(err)
	}
	if d.Added != 1 {
		t.Errorf("added=%d, want 1", d.Added)
	}
}
