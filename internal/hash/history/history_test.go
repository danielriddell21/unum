package history_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielriddell21/unum/internal/hash/history"
)

const (
	svcA = "service-a"
	svcB = "service-b"
)

func withTempHistory(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	orig, _ := os.UserConfigDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	_ = orig
}

func TestHistory_AppendAndLoad(t *testing.T) {
	withTempHistory(t)

	if err := history.Append(svcA); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}
	if err := history.Append(svcB); err != nil {
		t.Fatalf("AppendHistory: %v", err)
	}

	entries := history.Load()
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	// Newest first
	if entries[0].Input != svcB {
		t.Errorf("entries[0].Input = %q, want service-b", entries[0].Input)
	}
	if entries[1].Input != svcA {
		t.Errorf("entries[1].Input = %q, want service-a", entries[1].Input)
	}
}

func TestHistory_Deduplication(t *testing.T) {
	withTempHistory(t)

	_ = history.Append(svcA)
	_ = history.Append(svcB)
	_ = history.Append(svcA) // re-add — should move to front

	entries := history.Load()
	if len(entries) != 2 {
		t.Fatalf("want 2 entries after dedup, got %d", len(entries))
	}
	if entries[0].Input != svcA {
		t.Errorf("entries[0].Input = %q, want service-a", entries[0].Input)
	}
}

func TestHistory_Cap(t *testing.T) {
	withTempHistory(t)

	for i := 0; i < history.Max+10; i++ {
		_ = history.Append(filepath.Join("svc", string(rune('a'+i%26))))
	}

	entries := history.Load()
	if len(entries) > history.Max {
		t.Errorf("history len %d exceeds cap %d", len(entries), history.Max)
	}
}

func TestHistory_EmptyFile(t *testing.T) {
	withTempHistory(t)
	entries := history.Load()
	if len(entries) != 0 {
		t.Errorf("expected empty history, got %v", entries)
	}
}

func TestHistory_MalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Write corrupt JSON directly to the history file.
	dir := tmp
	histDir := filepath.Join(dir, "unum")
	if err := os.MkdirAll(histDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(histDir, "hash-history.json"), []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}

	entries := history.Load()
	if len(entries) != 0 {
		t.Errorf("malformed JSON should return empty slice, got %d entries", len(entries))
	}
}

func TestHistory_Clear(t *testing.T) {
	withTempHistory(t)

	if err := history.Append(svcA); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if len(history.Load()) == 0 {
		t.Fatal("history should have an entry before clearing")
	}

	if err := history.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if got := history.Load(); len(got) != 0 {
		t.Errorf("history should be empty after Clear, got %d entries", len(got))
	}

	path, err := history.Path()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("history file should be removed")
	}
}

func TestHistory_ClearWhenAbsent(t *testing.T) {
	withTempHistory(t)
	if err := history.Clear(); err != nil {
		t.Errorf("Clear on a missing file should succeed, got %v", err)
	}
}
