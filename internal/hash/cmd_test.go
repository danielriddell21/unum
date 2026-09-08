package hash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielriddell21/unum/internal/hash/history"
	"github.com/danielriddell21/unum/internal/hash/types"
)

func tempConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	t.Setenv("UNUM_NO_HASH_HISTORY", "")
	return dir
}

func historyExists(t *testing.T, dir string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, "unum", "hash-history.json"))
	return err == nil
}

func TestRecordHistory(t *testing.T) {
	tests := []struct {
		name      string
		noHistory bool
		historyOn bool
		wantFile  bool
	}{
		{name: "records by default", historyOn: true, wantFile: true},
		{name: "--no-history skips", noHistory: true, historyOn: true, wantFile: false},
		{name: "config off skips", historyOn: false, wantFile: false},
		{name: "both off skips", noHistory: true, historyOn: false, wantFile: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tempConfig(t)
			f := &flags{noHistory: tt.noHistory, historyOn: tt.historyOn}

			if err := f.recordHistory("some-service"); err != nil {
				t.Fatalf("recordHistory: %v", err)
			}
			if got := historyExists(t, dir); got != tt.wantFile {
				t.Errorf("history file exists=%v, want %v", got, tt.wantFile)
			}
		})
	}
}

func TestRecordHistory_StoresTheInput(t *testing.T) {
	tempConfig(t)
	f := &flags{historyOn: true}

	if err := f.recordHistory("my-service"); err != nil {
		t.Fatalf("recordHistory: %v", err)
	}
	entries := history.Load()
	if len(entries) != 1 || entries[0].Input != "my-service" {
		t.Errorf("history = %+v, want one entry for my-service", entries)
	}
}

func TestRenderHashResult_ReturnsSelectedField(t *testing.T) {
	r := types.Result{
		Input: "svc", Port: 8080, UUID: "uuid-value",
		Color: "#abcdef", Short: "abc123", Emoji: "🎉", Phrase: "a b c",
	}
	tests := []struct {
		name string
		f    *flags
		want string
	}{
		{"port", &flags{portOnly: true}, "port"},
		{"uuid", &flags{uuidOnly: true}, "uuid"},
		{"color", &flags{colorOnly: true}, "color"},
		{"short", &flags{shortOnly: true}, "short"},
		{"emoji", &flags{emojiOnly: true}, "emoji"},
		{"phrase", &flags{phraseOnly: true}, "phrase"},
		{"table", &flags{quiet: true, noColor: true}, "full"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderHashResult(tt.f, r); got != tt.want {
				t.Errorf("renderHashResult()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestActiveHashFlags(t *testing.T) {
	tests := []struct {
		name string
		f    *flags
		want []string
	}{
		{"none", &flags{}, nil},
		{"port", &flags{portOnly: true}, []string{"port"}},
		{"no-history", &flags{noHistory: true}, []string{"no-history"}},
		{
			name: "several",
			f:    &flags{uuidOnly: true, emojiOnly: true, noHistory: true},
			want: []string{"uuid", "emoji", "no-history"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := activeHashFlags(tt.f)
			if len(got) != len(tt.want) {
				t.Fatalf("activeHashFlags()=%v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("activeHashFlags()[%d]=%q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRunHash_ClearHistory(t *testing.T) {
	dir := tempConfig(t)

	f := &flags{historyOn: true}
	if err := f.recordHistory("something"); err != nil {
		t.Fatalf("recordHistory: %v", err)
	}
	if !historyExists(t, dir) {
		t.Fatal("history should exist before clearing")
	}

	if err := runHash(&flags{clearHistory: true}, nil); err != nil {
		t.Fatalf("runHash --clear-history: %v", err)
	}
	if historyExists(t, dir) {
		t.Error("--clear-history should remove the history file")
	}
}

func TestRunHash_ClearHistoryWhenAbsent(t *testing.T) {
	tempConfig(t)
	if err := runHash(&flags{clearHistory: true}, nil); err != nil {
		t.Errorf("clearing an absent history should succeed, got %v", err)
	}
}
