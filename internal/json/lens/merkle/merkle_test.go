package merkle_test

import (
	"context"
	"testing"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/lens/merkle"
	"github.com/danielriddell21/unum/internal/json/parse"
)

func runMerkle(t *testing.T, src string) string {
	t.Helper()
	root, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if err := (merkle.Analyzer{}).Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Fatal(err)
	}
	return merkle.RootHash(root)
}

func TestMerkleDeterministic(t *testing.T) {
	h1 := runMerkle(t, `{"a": 1, "b": 2}`)
	h2 := runMerkle(t, `{"a": 1, "b": 2}`)
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %s vs %s", h1, h2)
	}
}

func TestMerkleChangeSensitive(t *testing.T) {
	h1 := runMerkle(t, `{"a": 1}`)
	h2 := runMerkle(t, `{"a": 2}`)
	if h1 == h2 {
		t.Error("different values produced the same hash")
	}
}

func TestMerkleKeyOrderIndependent(t *testing.T) {
	// Object key order must not affect the hash (keys are sorted before hashing)
	h1 := runMerkle(t, `{"a": 1, "b": 2}`)
	h2 := runMerkle(t, `{"b": 2, "a": 1}`)
	if h1 != h2 {
		t.Errorf("object key order should not affect hash: %s vs %s", h1, h2)
	}
}

func TestMerkleArrayOrderMatters(t *testing.T) {
	// Array element order must affect the hash
	h1 := runMerkle(t, `[1, 2, 3]`)
	h2 := runMerkle(t, `[3, 2, 1]`)
	if h1 == h2 {
		t.Error("different array orders produced the same hash")
	}
}

func TestMerkleHashNonEmpty(t *testing.T) {
	h := runMerkle(t, `{"x": "hello"}`)
	if h == "" {
		t.Error("expected non-empty hash")
	}
	if len(h) != 64 {
		t.Errorf("expected 64-char hex SHA256, got %d chars: %s", len(h), h)
	}
}

func TestMerkleShort(t *testing.T) {
	if s := merkle.Short("abcdef0123456789"); s != "abcdef01" {
		t.Errorf("Short: got %q, want %q", s, "abcdef01")
	}
	if s := merkle.Short("abc"); s != "abc" {
		t.Errorf("Short on short string: got %q", s)
	}
}
