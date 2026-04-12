package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/danielriddell21/unum/internal/hash/types"
)

const Max = 50

// HistoryEntry is an alias for types.HistoryEntry so callers can use hash.HistoryEntry directly.
type HistoryEntry = types.HistoryEntry

func historyPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "unum", "hash-history.json"), nil
}

// Load reads the history file and returns entries newest-first.
// Returns an empty slice if the file is absent or malformed.
func Load() []HistoryEntry {
	path, err := historyPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

// Append adds input to the history file, capping at maxHistory entries.
func Append(input string) error {
	path, err := historyPath()
	if err != nil {
		return err
	}

	entries := Load()

	// Deduplicate: remove existing entry for this input so it moves to front.
	filtered := entries[:0]
	for _, e := range entries {
		if e.Input != input {
			filtered = append(filtered, e)
		}
	}

	// Prepend new entry.
	entries = append([]HistoryEntry{{Input: input, Time: time.Now()}}, filtered...)

	// Cap length.
	if len(entries) > Max {
		entries = entries[:Max]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
