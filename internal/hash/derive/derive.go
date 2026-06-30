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

type Result = types.Result

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

func derivePort(h [32]byte) uint16 {
	n := binary.BigEndian.Uint16(h[0:2])
	const lo, hi = 1024, 65535
	return lo + n%(hi-lo+1)
}

func deriveUUID(input string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(input)).String()
}

func deriveColor(h [32]byte) string {
	return fmt.Sprintf("#%s", hex.EncodeToString(h[0:3]))
}

func deriveShort(h [32]byte) string {
	return hex.EncodeToString(h[0:4])
}

func deriveEmoji(h [32]byte) string {
	return emojis[int(h[0])%len(emojis)]
}

func derivePhrase(h [32]byte) string {
	words := effWords()
	n := len(words)
	w1 := words[int(binary.BigEndian.Uint16(h[0:2]))%n]
	w2 := words[int(binary.BigEndian.Uint16(h[2:4]))%n]
	w3 := words[int(binary.BigEndian.Uint16(h[4:6]))%n]
	return w1 + "-" + w2 + "-" + w3
}

var (
	effOnce  sync.Once
	effSlice []string
)

func effWords() []string {
	effOnce.Do(func() {
		wl := diceware.WordListEffLarge()
		effSlice = make([]string, 0, 7776)
		for i := 0; i < 7776; i++ {
			d1 := i/1296 + 1
			d2 := (i/216)%6 + 1
			d3 := (i/36)%6 + 1
			d4 := (i/6)%6 + 1
			d5 := i%6 + 1
			code := d1*10000 + d2*1000 + d3*100 + d4*10 + d5
			effSlice = append(effSlice, wl.WordAt(code))
		}
	})
	return effSlice
}

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
