package web

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/danielriddell21/unum/internal/image/optimize"
)

const maxStored = 8

type store struct {
	mu    sync.RWMutex
	items map[string]optimize.Source
	order []string
}

func newStore() *store {
	return &store{items: make(map[string]optimize.Source)}
}

func (s *store) put(name string, data []byte) (string, optimize.Source, error) {
	sum := sha256.Sum256(data)
	id := hex.EncodeToString(sum[:8])

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.items[id]; ok {
		return id, existing, nil
	}

	src, err := optimize.Decode(name, data)
	if err != nil {
		return "", optimize.Source{}, fmt.Errorf("decode %s: %w", name, err)
	}

	s.items[id] = src
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
	src, ok := s.items[id]
	return src, ok
}
