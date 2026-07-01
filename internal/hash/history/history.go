package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/danielriddell21/unum/internal/hash/types"
)

const Max = 50

type HistoryEntry = types.HistoryEntry

func historyPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, "unum", "hash-history.json"), nil
}

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

func Append(input string) error {
	path, err := historyPath()
	if err != nil {
		return err
	}

	entries := Load()

	// Deduplicate: remove existing entry for this input so it moves to front.
	filtered := slices.DeleteFunc(entries, func(e HistoryEntry) bool { return e.Input == input })

	// Prepend new entry.
	entries = append([]HistoryEntry{{Input: input, Time: time.Now()}}, filtered...)

	// Cap length.
	if len(entries) > Max {
		entries = entries[:Max]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}
