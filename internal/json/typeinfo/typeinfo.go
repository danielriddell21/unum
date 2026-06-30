package typeinfo

import (
	"strings"

	"github.com/danielriddell21/unum/internal/json/node"
)

type TypeKind uint8

const (
	TypeNull TypeKind = iota
	TypeBool
	TypeNumber
	TypeString
	TypeArray
	TypeObject
	TypeMixed
)

type FieldInfo struct {
	Name     string
	GoName   string
	Type     *TypeInfo
	Nullable bool
}

type TypeInfo struct {
	Kind     TypeKind
	Nullable bool

	Fields []*FieldInfo

	Elem *TypeInfo
}

func Infer(n *node.Node) *TypeInfo {
	return inferNode(n)
}

func inferNode(n *node.Node) *TypeInfo {
	switch n.Kind {
	case node.KindNull:
		return &TypeInfo{Kind: TypeNull, Nullable: true}
	case node.KindBool:
		return &TypeInfo{Kind: TypeBool}
	case node.KindNumber:
		return &TypeInfo{Kind: TypeNumber}
	case node.KindString:
		return &TypeInfo{Kind: TypeString}

	case node.KindArray:
		if len(n.Children) == 0 {
			return &TypeInfo{Kind: TypeArray, Elem: &TypeInfo{Kind: TypeMixed}}
		}
		elem := inferNode(n.Children[0])
		for _, c := range n.Children[1:] {
			elem = merge(elem, inferNode(c))
		}
		return &TypeInfo{Kind: TypeArray, Elem: elem}

	case node.KindObject:
		ti := &TypeInfo{Kind: TypeObject}
		// Track field order and merge shapes
		seen := make(map[string]int) // key → index in Fields
		for _, c := range n.Children {
			ft := inferNode(c)
			if idx, exists := seen[c.Key]; exists {
				ti.Fields[idx].Type = merge(ti.Fields[idx].Type, ft)
				if ft.Nullable {
					ti.Fields[idx].Nullable = true
				}
			} else {
				seen[c.Key] = len(ti.Fields)
				ti.Fields = append(ti.Fields, &FieldInfo{
					Name:     c.Key,
					GoName:   toPascalCase(c.Key),
					Type:     ft,
					Nullable: ft.Nullable,
				})
			}
		}
		return ti
	}
	return &TypeInfo{Kind: TypeMixed}
}

func merge(a, b *TypeInfo) *TypeInfo {
	if a.Kind == b.Kind {
		result := &TypeInfo{Kind: a.Kind, Nullable: a.Nullable || b.Nullable}
		if a.Kind == TypeObject {
			result.Fields = mergeFields(a.Fields, b.Fields)
		}
		if a.Kind == TypeArray && a.Elem != nil && b.Elem != nil {
			result.Elem = merge(a.Elem, b.Elem)
		}
		return result
	}
	if a.Kind == TypeNull {
		b.Nullable = true
		return b
	}
	if b.Kind == TypeNull {
		a.Nullable = true
		return a
	}
	return &TypeInfo{Kind: TypeMixed, Nullable: a.Nullable || b.Nullable}
}

func mergeFields(a, b []*FieldInfo) []*FieldInfo {
	seen := make(map[string]int)
	result := make([]*FieldInfo, len(a))
	copy(result, a)
	for i, f := range result {
		seen[f.Name] = i
	}
	for _, bf := range b {
		if idx, exists := seen[bf.Name]; exists {
			result[idx].Type = merge(result[idx].Type, bf.Type)
			result[idx].Nullable = result[idx].Nullable || bf.Nullable
		} else {
			// Field present in b but not a → nullable (absent = null)
			bf2 := *bf
			bf2.Nullable = true
			seen[bf.Name] = len(result)
			result = append(result, &bf2)
		}
	}
	// Fields in a not in b → also mark nullable
	bKeys := make(map[string]bool)
	for _, bf := range b {
		bKeys[bf.Name] = true
	}
	for i, f := range result {
		if !bKeys[f.Name] {
			result[i].Nullable = true
		}
	}
	return result
}

func toPascalCase(s string) string {
	if s == "" {
		return "Field"
	}
	parts := splitWords(s)
	var sb strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		sb.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			sb.WriteString(p[1:])
		}
	}
	result := sb.String()
	if result == "" {
		return "Field"
	}
	return result
}

func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	for i, r := range s {
		switch {
		case r == '_' || r == '-' || r == ' ':
			if cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
			}
		case i > 0 && r >= 'A' && r <= 'Z':
			if cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
			}
			cur.WriteRune(r)
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	return words
}
