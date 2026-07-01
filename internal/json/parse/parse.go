package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/danielriddell21/unum/internal/json/node"
)

func Validate(data []byte) error {
	if json.Valid(data) {
		return nil
	}
	var syntaxErr *json.SyntaxError
	var v any
	if err := json.Unmarshal(data, &v); errors.As(err, &syntaxErr) {
		line, col := offsetToLineCol(data, syntaxErr.Offset)
		return fmt.Errorf("syntax error at line %d, col %d: %s", line, col, syntaxErr.Error())
	}
	return fmt.Errorf("invalid JSON")
}

func Parse(data []byte) (*node.Node, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber() // preserve number formatting (e.g. "1.0" stays "1.0")
	return parseValue(dec, "", -1, nil)
}

const errTokenFmt = "token: %w"

func parseValue(dec *json.Decoder, key string, index int, parent *node.Node) (*node.Node, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf(errTokenFmt, err)
	}

	n := &node.Node{Key: key, Index: index, Parent: parent}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			n.Kind = node.KindObject
			if err := parseObject(dec, n); err != nil {
				return nil, err
			}
		case '[':
			n.Kind = node.KindArray
			if err := parseArray(dec, n); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected delimiter: %s", v)
		}

	case json.Number:
		n.Kind = node.KindNumber
		n.Raw = v.String()

	case string:
		n.Kind = node.KindString
		b, _ := json.Marshal(v)
		n.Raw = string(b)

	case bool:
		n.Kind = node.KindBool
		if v {
			n.Raw = "true"
		} else {
			n.Raw = "false"
		}

	case nil:
		n.Kind = node.KindNull
		n.Raw = "null"

	default:
		return nil, fmt.Errorf("unexpected token type: %T", tok)
	}

	return n, nil
}

func parseObject(dec *json.Decoder, n *node.Node) error {
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return fmt.Errorf(errTokenFmt, err)
		}
		childKey, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("expected object key, got %T", keyTok)
		}
		child, err := parseValue(dec, childKey, -1, n)
		if err != nil {
			return err
		}
		n.Children = append(n.Children, child)
	}
	if _, err := dec.Token(); err != nil { // consume '}'
		return fmt.Errorf(errTokenFmt, err)
	}
	return nil
}

func parseArray(dec *json.Decoder, n *node.Node) error {
	for i := 0; dec.More(); i++ {
		child, err := parseValue(dec, "", i, n)
		if err != nil {
			return err
		}
		n.Children = append(n.Children, child)
	}
	if _, err := dec.Token(); err != nil { // consume ']'
		return fmt.Errorf(errTokenFmt, err)
	}
	return nil
}

func offsetToLineCol(data []byte, offset int64) (line, col int) {
	line = 1
	col = 1
	for i := int64(0); i < offset && i < int64(len(data)); i++ {
		if data[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return
}
