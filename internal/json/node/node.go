package node

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Kind represents the JSON value type of a node.
type Kind uint8

const (
	KindNull Kind = iota
	KindBool
	KindNumber
	KindString
	KindArray
	KindObject
)

func (k Kind) String() string {
	switch k {
	case KindNull:
		return "null"
	case KindBool:
		return "bool"
	case KindNumber:
		return "number"
	case KindString:
		return "string"
	case KindArray:
		return "array"
	case KindObject:
		return "object"
	default:
		return "unknown"
	}
}

// AnnotationKey namespaces annotations to prevent collisions between lenses.
type AnnotationKey struct {
	Lens string // e.g. "stats", "merkle", "decode"
	Name string // e.g. "mean", "hash", "type"
}

// Annotation is a single piece of derived data attached to a node by a lens.
type Annotation struct {
	Key   AnnotationKey
	Value any
}

// Node is the core data structure. Parse once, annotate many times.
// All lenses read from the same tree and write their own annotation keys.
type Node struct {
	Kind  Kind
	Key   string // object key if this is an object value; "" otherwise
	Index int    // array index if this is an array element; -1 otherwise
	Raw   string // raw JSON token for leaf nodes (e.g. `"hello"`, `42`, `true`, `null`)

	Children []*Node
	Parent   *Node

	mu          sync.RWMutex
	annotations []Annotation
}

// Annotate stamps an annotation onto this node. Safe for concurrent use.
func (n *Node) Annotate(a Annotation) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.annotations = append(n.annotations, a)
}

// GetAnnotation returns the value for a specific lens+name key.
func (n *Node) GetAnnotation(lens, name string) (any, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	key := AnnotationKey{Lens: lens, Name: name}
	for _, a := range n.annotations {
		if a.Key == key {
			return a.Value, true
		}
	}
	return nil, false
}

// GetAnnotations returns all annotations for a given lens.
func (n *Node) GetAnnotations(lens string) []Annotation {
	n.mu.RLock()
	defer n.mu.RUnlock()
	var out []Annotation
	for _, a := range n.annotations {
		if a.Key.Lens == lens {
			out = append(out, a)
		}
	}
	return out
}

// Walk visits every node depth-first pre-order (parent before children).
// Return false from fn to skip the subtree rooted at that node.
func (n *Node) Walk(fn func(*Node) bool) {
	if !fn(n) {
		return
	}
	for _, c := range n.Children {
		c.Walk(fn)
	}
}

// WalkPost visits every node depth-first post-order (children before parent).
// Required by merkle hashing, where children must be hashed before their parent.
func (n *Node) WalkPost(fn func(*Node)) {
	for _, c := range n.Children {
		c.WalkPost(fn)
	}
	fn(n)
}

// Path returns the dot-notation path from root to this node (e.g. ".users[0].name").
func (n *Node) Path() string {
	if n.Parent == nil {
		return "."
	}
	parts := make([]string, 0, 8)
	cur := n
	for cur.Parent != nil {
		if cur.Index >= 0 {
			parts = append(parts, fmt.Sprintf("[%d]", cur.Index))
		} else {
			parts = append(parts, "."+cur.Key)
		}
		cur = cur.Parent
	}
	// reverse
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	p := strings.Join(parts, "")
	if !strings.HasPrefix(p, ".") {
		return "." + p
	}
	return p
}

// CountNodes returns the total number of nodes in this subtree (including self).
func (n *Node) CountNodes() int {
	count := 1
	for _, c := range n.Children {
		count += c.CountNodes()
	}
	return count
}

// ToAny converts the node back to a Go any value suitable for json.Marshal.
func (n *Node) ToAny() any {
	switch n.Kind {
	case KindNull:
		return nil
	case KindBool:
		return n.Raw == "true"
	case KindNumber:
		f := json.Number(n.Raw)
		return f
	case KindString:
		var s string
		_ = json.Unmarshal([]byte(n.Raw), &s)
		return s
	case KindArray:
		arr := make([]any, len(n.Children))
		for i, c := range n.Children {
			arr[i] = c.ToAny()
		}
		return arr
	case KindObject:
		// Use ordered pairs to produce consistent output
		type kv struct {
			k string
			v any
		}
		pairs := make([]kv, len(n.Children))
		for i, c := range n.Children {
			pairs[i] = kv{c.Key, c.ToAny()}
		}
		// Marshal to ordered JSON then unmarshal to map (loses order in map,
		// but ToAny is used for YAML/query interop where order matters less)
		m := make(map[string]any, len(n.Children))
		for _, p := range pairs {
			m[p.k] = p.v
		}
		return m
	}
	return nil
}

// MarshalJSON produces compact JSON for this node, preserving key order.
func (n *Node) MarshalJSON() ([]byte, error) { //nolint:gocognit // NOSONAR: serialises every node kind with different JSON shapes; the branches reflect the type system
	switch n.Kind {
	case KindNull, KindBool, KindNumber:
		return []byte(n.Raw), nil
	case KindString:
		return []byte(n.Raw), nil
	case KindArray:
		if len(n.Children) == 0 {
			return []byte("[]"), nil
		}
		var sb strings.Builder
		sb.WriteByte('[')
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteByte(',')
			}
			b, err := c.MarshalJSON()
			if err != nil {
				return nil, err
			}
			sb.Write(b)
		}
		sb.WriteByte(']')
		return []byte(sb.String()), nil
	case KindObject:
		if len(n.Children) == 0 {
			return []byte("{}"), nil
		}
		var sb strings.Builder
		sb.WriteByte('{')
		for i, c := range n.Children {
			if i > 0 {
				sb.WriteByte(',')
			}
			keyBytes, err := json.Marshal(c.Key)
			if err != nil {
				return nil, fmt.Errorf("marshal child: %w", err)
			}
			sb.Write(keyBytes)
			sb.WriteByte(':')
			valBytes, err := c.MarshalJSON()
			if err != nil {
				return nil, err
			}
			sb.Write(valBytes)
		}
		sb.WriteByte('}')
		return []byte(sb.String()), nil
	}
	return nil, fmt.Errorf("unknown kind: %d", n.Kind)
}
