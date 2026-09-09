package connectors

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// DecodeSourceStructured is the bounded value codec for a selected object or
// array input. It preserves numbers and rejects ambiguous duplicate members;
// it does not authorize a query field or choose a serialization dialect.
func DecodeSourceStructured(kind, raw string, maxBytes int) (any, error) {
	if (kind != "object" && kind != "array") || maxBytes <= 0 || len(raw) > maxBytes || !utf8.ValidString(raw) {
		return nil, fmt.Errorf("invalid structured source input kind, encoding or byte budget")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	nodes := 0
	var read func(int) (any, error)
	read = func(depth int) (any, error) {
		nodes++
		if depth > 32 || nodes > 10000 {
			return nil, fmt.Errorf("structured source input exceeds depth or node budget")
		}
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid structured source JSON")
		}
		switch value := token.(type) {
		case json.Delim:
			switch value {
			case '{':
				object := map[string]any{}
				for decoder.More() {
					token, err := decoder.Token()
					if err != nil {
						return nil, fmt.Errorf("invalid structured source member")
					}
					key, ok := token.(string)
					if !ok {
						return nil, fmt.Errorf("invalid structured source member")
					}
					if _, exists := object[key]; exists {
						return nil, fmt.Errorf("duplicate structured source member")
					}
					child, err := read(depth + 1)
					if err != nil {
						return nil, err
					}
					object[key] = child
				}
				close, err := decoder.Token()
				if err != nil || close != json.Delim('}') {
					return nil, fmt.Errorf("invalid structured source object")
				}
				return object, nil
			case '[':
				array := []any{}
				for decoder.More() {
					child, err := read(depth + 1)
					if err != nil {
						return nil, err
					}
					array = append(array, child)
				}
				close, err := decoder.Token()
				if err != nil || close != json.Delim(']') {
					return nil, fmt.Errorf("invalid structured source array")
				}
				return array, nil
			default:
				return nil, fmt.Errorf("invalid structured source delimiter")
			}
		case json.Number:
			return DecodeSourceScalar("number", value.String(), maxBytes)
		case string, bool, nil:
			return value, nil
		default:
			return nil, fmt.Errorf("invalid structured source value")
		}
	}
	value, err := read(0)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("structured source input must contain exactly one value")
	}
	if kind == "object" {
		if _, ok := value.(map[string]any); !ok {
			return nil, fmt.Errorf("source input must be object")
		}
	} else {
		if _, ok := value.([]any); !ok {
			return nil, fmt.Errorf("source input must be array")
		}
	}
	return value, nil
}
