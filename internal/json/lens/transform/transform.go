package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/danielriddell21/unum/internal/json/node"
)

func ToYAML(root *node.Node) ([]byte, error) {
	// Round-trip through JSON → any → YAML.
	// We marshal from the node tree to preserve key order as much as yaml.v3 allows.
	jsonBytes, err := marshalOrdered(root)
	if err != nil {
		return nil, fmt.Errorf("transform: %w", err)
	}

	var v any
	if err := json.Unmarshal(jsonBytes, &v); err != nil {
		return nil, fmt.Errorf("transform unmarshal: %w", err)
	}

	// Convert to a yaml.Node to better preserve ordering
	yamlNode, err := toYAMLNode(root)
	if err != nil {
		return nil, fmt.Errorf("transform yaml node: %w", err)
	}

	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{yamlNode},
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("transform encode: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("transform close: %w", err)
	}
	return buf.Bytes(), nil
}

func toYAMLNode(n *node.Node) (*yaml.Node, error) {
	switch n.Kind {
	case node.KindNull:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil

	case node.KindBool:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: n.Raw}, nil

	case node.KindNumber:
		// Use !!int for integers, !!float for decimals
		tag := "!!float"
		if !strings.ContainsAny(n.Raw, ".eE") {
			tag = "!!int"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: n.Raw}, nil

	case node.KindString:
		var s string
		if err := json.Unmarshal([]byte(n.Raw), &s); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}, nil

	case node.KindArray:
		yn := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, c := range n.Children {
			child, err := toYAMLNode(c)
			if err != nil {
				return nil, err
			}
			yn.Content = append(yn.Content, child)
		}
		return yn, nil

	case node.KindObject:
		yn := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, c := range n.Children {
			key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: c.Key}
			val, err := toYAMLNode(c)
			if err != nil {
				return nil, err
			}
			yn.Content = append(yn.Content, key, val)
		}
		return yn, nil
	}

	return nil, fmt.Errorf("unknown kind: %d", n.Kind)
}

func marshalOrdered(n *node.Node) ([]byte, error) {
	b, err := n.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return b, nil
}
