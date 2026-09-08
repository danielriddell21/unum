package web

import (
	"testing"
	"time"

	"github.com/danielriddell21/unum/internal/json/node"
)

func TestDocCache_PutGet(t *testing.T) {
	c := &docCache{items: make(map[string]cacheEntry)}
	root := &node.Node{}
	c.put("k", root, []byte("payload"))

	e, ok := c.get("k")
	if !ok {
		t.Fatal("entry should be present")
	}
	if e.root != root {
		t.Error("root should round-trip")
	}
	if string(e.payload) != "payload" {
		t.Errorf("payload=%q", e.payload)
	}
}

func TestDocCache_MissingKey(t *testing.T) {
	c := &docCache{items: make(map[string]cacheEntry)}
	if _, ok := c.get("nope"); ok {
		t.Error("missing key should not be found")
	}
}

func TestDocCache_EvictsOldestBeyondMax(t *testing.T) {
	c := &docCache{items: make(map[string]cacheEntry)}
	for i := range cacheMaxEntries + 5 {
		c.put(string(rune('a'+i)), &node.Node{}, []byte{byte(i)})
	}

	if len(c.items) != cacheMaxEntries {
		t.Errorf("len(items)=%d, want %d", len(c.items), cacheMaxEntries)
	}
	if _, ok := c.get("a"); ok {
		t.Error("oldest entry should have been evicted")
	}
	if _, ok := c.get(string(rune('a' + cacheMaxEntries + 4))); !ok {
		t.Error("newest entry should still be present")
	}
}

func TestDocCache_EvictsExpired(t *testing.T) {
	c := &docCache{items: make(map[string]cacheEntry)}
	c.put("stale", &node.Node{}, []byte("x"))

	c.mu.Lock()
	e := c.items["stale"]
	e.stored = time.Now().Add(-cacheTTL - time.Minute)
	c.items["stale"] = e
	c.mu.Unlock()

	if _, ok := c.get("stale"); ok {
		t.Error("expired entry should not be returned")
	}
	if len(c.items) != 0 {
		t.Errorf("expired entry should be dropped, len=%d", len(c.items))
	}
	if len(c.order) != 0 {
		t.Errorf("order should be pruned, len=%d", len(c.order))
	}
}

func TestDocCache_FreshEntrySurvivesEviction(t *testing.T) {
	c := &docCache{items: make(map[string]cacheEntry)}
	c.put("old", &node.Node{}, []byte("x"))
	c.put("new", &node.Node{}, []byte("y"))

	c.mu.Lock()
	e := c.items["old"]
	e.stored = time.Now().Add(-cacheTTL - time.Minute)
	c.items["old"] = e
	c.mu.Unlock()

	if _, ok := c.get("new"); !ok {
		t.Error("fresh entry should survive an expiry sweep")
	}
}
