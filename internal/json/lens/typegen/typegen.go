// Package typegen generates Go structs, TypeScript interfaces, and JSON Schema
// from an inferred TypeInfo derived from the parsed node tree.
package typegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/typeinfo"
)

// Target selects the output language.
type Target string

const (
	TargetGo         Target = "go"
	TargetTypeScript Target = "ts"
	TargetJSONSchema Target = "jsonschema"
)

// Options controls code generation.
type Options struct {
	Target      Target
	PackageName string // Go only; defaults to "main"
	TypeName    string // root type name; defaults to "Root"
}

// Generate produces source code or a JSON Schema from the node tree.
func Generate(root *node.Node, opts Options) (string, error) {
	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	if opts.TypeName == "" {
		opts.TypeName = "Root"
	}

	ti := typeinfo.Infer(root)

	switch opts.Target {
	case TargetGo:
		return generateGo(ti, opts)
	case TargetTypeScript:
		return generateTS(ti, opts)
	case TargetJSONSchema:
		return generateJSONSchema(ti, opts)
	default:
		return "", fmt.Errorf("unknown typegen target: %s", opts.Target)
	}
}

// ─── Go ──────────────────────────────────────────────────────────────────────

func generateGo(ti *typeinfo.TypeInfo, opts Options) (string, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "package %s\n\n", opts.PackageName)

	structs := &strings.Builder{}
	collectGoStructs(structs, opts.TypeName, ti, make(map[string]bool))

	sb.WriteString(structs.String())
	return sb.String(), nil
}

func collectGoStructs(out *strings.Builder, name string, ti *typeinfo.TypeInfo, seen map[string]bool) {
	if ti.Kind != typeinfo.TypeObject {
		return
	}
	if seen[name] {
		return
	}
	seen[name] = true

	fmt.Fprintf(out, "type %s struct {\n", name)
	for _, f := range ti.Fields {
		nestedName := name + f.GoName
		goType := goTypeName(nestedName, f.Type, f.Nullable)
		fmt.Fprintf(out, "\t%s %s `json:\"%s\"`\n", f.GoName, goType, f.Name)
	}
	out.WriteString("}\n\n")

	// Recurse for nested objects
	for _, f := range ti.Fields {
		nestedName := name + f.GoName
		if f.Type.Kind == typeinfo.TypeObject {
			collectGoStructs(out, nestedName, f.Type, seen)
		} else if f.Type.Kind == typeinfo.TypeArray && f.Type.Elem != nil && f.Type.Elem.Kind == typeinfo.TypeObject {
			collectGoStructs(out, nestedName, f.Type.Elem, seen)
		}
	}
}

func goTypeName(fullName string, ti *typeinfo.TypeInfo, nullable bool) string {
	ptr := ""
	if nullable && ti.Kind != typeinfo.TypeNull {
		ptr = "*"
	}

	switch ti.Kind {
	case typeinfo.TypeNull:
		return "any"
	case typeinfo.TypeBool:
		return ptr + "bool"
	case typeinfo.TypeNumber:
		return ptr + "float64"
	case typeinfo.TypeString:
		return ptr + "string"
	case typeinfo.TypeMixed:
		return "any"
	case typeinfo.TypeArray:
		if ti.Elem == nil {
			return "[]any"
		}
		return "[]" + goTypeName(fullName, ti.Elem, false)
	case typeinfo.TypeObject:
		return ptr + fullName
	}
	return "any"
}

// ─── TypeScript ───────────────────────────────────────────────────────────────

func generateTS(ti *typeinfo.TypeInfo, opts Options) (string, error) {
	var sb strings.Builder
	collectTSInterfaces(&sb, opts.TypeName, ti, make(map[string]bool))
	return sb.String(), nil
}

func collectTSInterfaces(out *strings.Builder, name string, ti *typeinfo.TypeInfo, seen map[string]bool) {
	if ti.Kind != typeinfo.TypeObject {
		return
	}
	if seen[name] {
		return
	}
	seen[name] = true

	fmt.Fprintf(out, "interface %s {\n", name)
	for _, f := range ti.Fields {
		optional := ""
		if f.Nullable {
			optional = "?"
		}
		nestedName := name + f.GoName
		tsType := tsTypeName(nestedName, f.Type)
		fmt.Fprintf(out, "  %s%s: %s;\n", f.Name, optional, tsType)
	}
	out.WriteString("}\n\n")

	for _, f := range ti.Fields {
		nestedName := name + f.GoName
		if f.Type.Kind == typeinfo.TypeObject {
			collectTSInterfaces(out, nestedName, f.Type, seen)
		} else if f.Type.Kind == typeinfo.TypeArray && f.Type.Elem != nil && f.Type.Elem.Kind == typeinfo.TypeObject {
			collectTSInterfaces(out, nestedName, f.Type.Elem, seen)
		}
	}
}

func tsTypeName(fullName string, ti *typeinfo.TypeInfo) string {
	nullable := ""
	if ti.Nullable {
		nullable = " | null"
	}

	switch ti.Kind {
	case typeinfo.TypeNull:
		return "null"
	case typeinfo.TypeBool:
		return "boolean" + nullable
	case typeinfo.TypeNumber:
		return "number" + nullable
	case typeinfo.TypeString:
		return "string" + nullable
	case typeinfo.TypeMixed:
		return "unknown"
	case typeinfo.TypeArray:
		if ti.Elem == nil {
			return "unknown[]"
		}
		return tsTypeName(fullName, ti.Elem) + "[]"
	case typeinfo.TypeObject:
		return fullName + nullable
	}
	return "unknown"
}

// ─── JSON Schema ──────────────────────────────────────────────────────────────

func generateJSONSchema(ti *typeinfo.TypeInfo, opts Options) (string, error) {
	schema := buildSchema(ti)
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["title"] = opts.TypeName

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func buildSchema(ti *typeinfo.TypeInfo) map[string]any {
	schema := make(map[string]any)

	if ti.Nullable {
		// JSON Schema 2020-12: use anyOf with null type
		inner := buildSchemaInner(ti)
		schema["anyOf"] = []any{inner, map[string]any{"type": "null"}}
		return schema
	}

	return buildSchemaInner(ti)
}

func buildSchemaInner(ti *typeinfo.TypeInfo) map[string]any {
	schema := make(map[string]any)
	switch ti.Kind {
	case typeinfo.TypeNull:
		schema["type"] = "null"
	case typeinfo.TypeBool:
		schema["type"] = "boolean"
	case typeinfo.TypeNumber:
		schema["type"] = "number"
	case typeinfo.TypeString:
		schema["type"] = "string"
	case typeinfo.TypeMixed:
		// no type constraint
	case typeinfo.TypeArray:
		schema["type"] = "array"
		if ti.Elem != nil {
			schema["items"] = buildSchema(ti.Elem)
		}
	case typeinfo.TypeObject:
		schema["type"] = "object"
		if len(ti.Fields) > 0 {
			props := make(map[string]any)
			var required []string
			for _, f := range ti.Fields {
				props[f.Name] = buildSchema(f.Type)
				if !f.Nullable {
					required = append(required, f.Name)
				}
			}
			schema["properties"] = props
			if len(required) > 0 {
				schema["required"] = required
			}
		}
	}
	return schema
}
