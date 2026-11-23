package web

import (
	"sync"
	"time"

	"github.com/danielriddell21/unum/internal/json/node"
)

const (
	// Each entry can retain the node tree, the marshalled payload and, once
	// queried, the any-form copy gojq runs on, so the cap is kept modest.
	cacheMaxEntries = 16
	cacheTTL        = 30 * time.Minute
)

type queryDoc struct {
	root  *node.Node
	once  sync.Once
	input any
}

func newQueryDoc(root *node.Node) *queryDoc {
	return &queryDoc{root: root}
}

// anyForm converts lazily so documents that are never queried do not pay for
// the extra copy; concurrent first queries share one conversion.
func (d *queryDoc) anyForm() any {
	d.once.Do(func() { d.input = d.root.ToAny() })
	return d.input
}

type cacheEntry struct {
	doc     *queryDoc
	payload []byte
	stored  time.Time
}

type docCache struct {
	mu    sync.Mutex
	items map[string]cacheEntry
	order []string
}

var uploads = &docCache{items: make(map[string]cacheEntry)}

func (c *docCache) put(key string, doc *queryDoc, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictExpired()
	c.items[key] = cacheEntry{doc: doc, payload: payload, stored: time.Now()}
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
