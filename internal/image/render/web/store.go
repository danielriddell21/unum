package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/danielriddell21/unum/internal/image/optimize"
)

const maxStored = 8

// entry keeps the original bytes alongside the decoded image so the browser can
// be handed the source back to display, rather than making an object URL out of
// the File the user picked.
type entry struct {
	src optimize.Source
	raw []byte
}

type store struct {
	mu    sync.RWMutex
	items map[string]entry
	order []string
}

func newStore() *store {
	return &store{items: make(map[string]entry)}
}

func (s *store) put(name string, data []byte) (string, optimize.Source, error) {
	sum := sha256.Sum256(data)
	id := hex.EncodeToString(sum[:8])

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.items[id]; ok {
		return id, existing.src, nil
	}

	src, err := optimize.Decode(name, data)
	if err != nil {
		return "", optimize.Source{}, fmt.Errorf("decode %s: %w", name, err)
	}

	s.items[id] = entry{src: src, raw: data}
	s.order = append(s.order, id)
	for len(s.order) > maxStored {
		delete(s.items, s.order[0])
		s.order = s.order[1:]
	}
	return id, src, nil
}

func (s *store) get(id string) (optimize.Source, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[id]
	return e.src, ok
}

// raw returns the bytes exactly as they were uploaded.
func (s *store) raw(id string) ([]byte, optimize.Format, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[id]
	if !ok {
		return nil, optimize.FormatUnknown, false
	}
	return e.raw, e.src.Format, true
}
