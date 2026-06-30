package types

import "time"

type Result struct {
	Input  string
	Port   uint16
	UUID   string
	Color  string
	Short  string
	Emoji  string
	Phrase string
}

type HistoryEntry struct {
	Input string    `json:"input"`
	Time  time.Time `json:"time"`
}
