package connsdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// decodeJSON decodes body into a generic value preserving numbers as json.Number.
func decodeJSON(body []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// RecordsAt extracts an array of records from a JSON body.
//
// path selects where the array lives using dotted notation:
//   - "" or "." selects the root (which may itself be an array, or an object
//     that is returned as a single one-element record set).
//   - "data" selects body["data"].
//   - "result.items" selects body["result"]["items"].
//
// If the selected value is an array, each element that is a JSON object becomes a
// Record. If it is a single object, it becomes a one-element result. This mirrors
// the variety of real REST payloads (top-level arrays, {data:[...]}, single
// resource responses).
func RecordsAt(body []byte, path string) ([]Record, error) {
	root, err := decodeJSON(body)
	if err != nil {
		return nil, err
	}
	node := selectPath(root, path)
	if node == nil {
		return nil, nil
	}
	switch v := node.(type) {
	case []any:
		out := make([]Record, 0, len(v))
		for _, item := range v {
			if obj, ok := item.(map[string]any); ok {
				out = append(out, Record(obj))
			}
		}
		return out, nil
	case map[string]any:
		return []Record{Record(v)}, nil
	default:
		return nil, nil
	}
}

// StringAt returns the value at a dotted path as a string. Numbers and booleans
// are stringified; missing or null values return "". Useful for reading the next
// cursor/page token out of a response body.
func StringAt(body []byte, path string) (string, error) {
	root, err := decodeJSON(body)
	if err != nil {
		return "", err
	}
	node := selectPath(root, path)
	return stringify(node), nil
}

// selectPath walks a decoded JSON value along a dotted path.
func selectPath(root any, path string) any {
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return root
	}
	cur := root
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = obj[seg]
		if !ok {
			return nil
		}
	}
	return cur
}

// stringify renders a scalar JSON value as a string.
func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
