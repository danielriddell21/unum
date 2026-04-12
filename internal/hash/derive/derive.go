// Package derive implements the unum hash subcommand — a deterministic deriver
// that maps any string to a stable set of useful values via SHA256.
package derive

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sethvargo/go-diceware/diceware"

	"github.com/danielriddell21/unum/internal/hash/types"
)

// Result is an alias for types.Result so callers can use hash.Result directly.
type Result = types.Result

// Derive hashes input with SHA256 and derives all Result fields from the bytes.
func Derive(input string) Result {
	h := sha256.Sum256([]byte(input))

	return Result{
		Input:  input,
		Port:   derivePort(h),
		UUID:   deriveUUID(input),
		Color:  deriveColor(h),
		Short:  deriveShort(h),
		Emoji:  deriveEmoji(h),
		Phrase: derivePhrase(h),
	}
}

// derivePort maps bytes[0:2] to the range 1024–65535.
func derivePort(h [32]byte) uint16 {
	n := binary.BigEndian.Uint16(h[0:2])
	const lo, hi = 1024, 65535
	return lo + n%(hi-lo+1)
}

// deriveUUID returns a deterministic UUID v5 using the DNS namespace.
func deriveUUID(input string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(input)).String()
}

// deriveColor formats bytes[0:3] as a CSS hex color.
func deriveColor(h [32]byte) string {
	return fmt.Sprintf("#%s", hex.EncodeToString(h[0:3]))
}

// deriveShort returns the first 8 hex characters of the hash.
func deriveShort(h [32]byte) string {
	return hex.EncodeToString(h[0:4])
}

// deriveEmoji indexes into a curated emoji slice using byte[0].
func deriveEmoji(h [32]byte) string {
	return emojis[int(h[0])%len(emojis)]
}

// derivePhrase picks 3 words from the EFF large wordlist using bytes[0:6],
// 2 bytes per word, joined with hyphens.
func derivePhrase(h [32]byte) string {
	words := effWords()
	n := len(words)
	w1 := words[int(binary.BigEndian.Uint16(h[0:2]))%n]
	w2 := words[int(binary.BigEndian.Uint16(h[2:4]))%n]
	w3 := words[int(binary.BigEndian.Uint16(h[4:6]))%n]
	return w1 + "-" + w2 + "-" + w3
}

// ─── EFF wordlist ─────────────────────────────────────────────────────────────

var (
	effOnce  sync.Once
	effSlice []string
)

// effWords returns the EFF large wordlist as a flat []string, built once.
func effWords() []string {
	effOnce.Do(func() {
		wl := diceware.WordListEffLarge()
		effSlice = make([]string, 0, 7776)
		for d1 := 1; d1 <= 6; d1++ {
			for d2 := 1; d2 <= 6; d2++ {
				for d3 := 1; d3 <= 6; d3++ {
					for d4 := 1; d4 <= 6; d4++ {
						for d5 := 1; d5 <= 6; d5++ {
							code := d1*10000 + d2*1000 + d3*100 + d4*10 + d5
							effSlice = append(effSlice, wl.WordAt(code))
						}
					}
				}
			}
		}
	})
	return effSlice
}

// ─── Emoji list ───────────────────────────────────────────────────────────────

var emojis = []string{
	"🦊", "🐺", "🦁", "🐯", "🐻", "🐼", "🦝", "🦨",
	"🦡", "🦦", "🦥", "🐭", "🐹", "🐰", "🐸", "🐮",
	"🐷", "🐔", "🐧", "🐦", "🦆", "🦅", "🦉", "🦇",
	"🐺", "🐗", "🐴", "🦄", "🐝", "🪱", "🐛", "🦋",
	"🐌", "🐞", "🐜", "🪲", "🦟", "🦗", "🕷️", "🦂",
	"🐢", "🐍", "🦎", "🦖", "🦕", "🐙", "🦑", "🦐",
	"🦞", "🦀", "🐡", "🐟", "🐠", "🐬", "🐳", "🦈",
	"🐊", "🐅", "🐆", "🦓", "🦍", "🦧", "🐘", "🦏",
}
