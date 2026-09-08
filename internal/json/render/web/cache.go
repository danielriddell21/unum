package web

import (
	"sync"
	"time"

	"github.com/danielriddell21/unum/internal/json/node"
)

const (
	cacheMaxEntries = 32
	cacheTTL        = 30 * time.Minute
)

type cacheEntry struct {
	root    *node.Node
	payload []byte
	stored  time.Time
}

type docCache struct {
	mu    sync.Mutex
	items map[string]cacheEntry
	order []string
}

var uploads = &docCache{items: make(map[string]cacheEntry)}

func (c *docCache) put(key string, root *node.Node, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictExpired()
	c.items[key] = cacheEntry{root: root, payload: payload, stored: time.Now()}
	c.order = append(c.order, key)

	for len(c.order) > cacheMaxEntries {
		delete(c.items, c.order[0])
		c.order = c.order[1:]
	}
}

func (c *docCache) get(key string) (cacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictExpired()
	e, ok := c.items[key]
	return e, ok
}

func (c *docCache) evictExpired() {
	cutoff := time.Now().Add(-cacheTTL)
	kept := c.order[:0]
	for _, k := range c.order {
		if e, ok := c.items[k]; ok && e.stored.Before(cutoff) {
			delete(c.items, k)
			continue
		}
		kept = append(kept, k)
	}
	c.order = kept
}
