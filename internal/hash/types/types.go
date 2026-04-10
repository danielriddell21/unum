// Package types defines the shared data types for the hash tool.
package types

import "time"

// Result holds all values deterministically derived from an input string.
type Result struct {
	Input  string
	Port   uint16
	UUID   string
	Color  string // "#rrggbb"
	Short  string // first 8 hex chars of hash
	Emoji  string
	Phrase string // "word-word-word"
}

// HistoryEntry is one item in the persistent hash history.
type HistoryEntry struct {
	Input string    `json:"input"`
	Time  time.Time `json:"time"`
}
