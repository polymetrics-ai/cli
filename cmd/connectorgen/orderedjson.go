package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// orderedJSON is a JSON document that round-trips with its object key order
// and exact number formatting intact.
//
// encoding/json decodes objects into map[string]any, which loses key order and
// re-emits keys sorted, and decodes every number as float64, which would
// rewrite 104857600 as 1.048576e+08. Either behaviour would turn a four-field
// edit into a whole-file rewrite and bury the real change in noise, so the
// generator decodes through the token stream instead and preserves both.
type orderedJSON struct {
	root *orderedObject
}

// orderedObject is a JSON object that remembers the order its keys appeared in.
type orderedObject struct {
	keys   []string
	values map[string]any
}

func newOrderedObject() *orderedObject {
	return &orderedObject{values: map[string]any{}}
}

func (o *orderedObject) get(key string) (any, bool) {
	v, ok := o.values[key]
	return v, ok
}

// set overwrites an existing key in place, or appends a new one at the end.
func (o *orderedObject) set(key string, value any) {
	if _, exists := o.values[key]; !exists {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

// remove deletes a key, keeping the remaining keys in their original order. It
// reports whether the key was present.
func (o *orderedObject) remove(key string) bool {
	if _, exists := o.values[key]; !exists {
		return false
	}
	delete(o.values, key)
	for i, existing := range o.keys {
		if existing == key {
			o.keys = append(o.keys[:i:i], o.keys[i+1:]...)
			break
		}
	}
	return true
}

func (o *orderedObject) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range o.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		encoded, err := marshalNoEscapeHTML(key)
		if err != nil {
			return nil, err
		}
		buf.Write(encoded)
		buf.WriteByte(':')
		encoded, err = marshalNoEscapeHTML(o.values[key])
		if err != nil {
			return nil, err
		}
		buf.Write(encoded)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// marshalNoEscapeHTML encodes a value without turning <, > and & into <,
// > and &.
//
// json.Marshal always escapes those, and the outer encoder's
// SetEscapeHTML(false) does not reach values encoded inside a custom
// MarshalJSON. Connector summaries and approval strings are full of angle
// brackets ("pm gong <command>", "plan -> preview"), so escaping them here
// would rewrite hundreds of untouched lines in every bundle it wrote.
func marshalNoEscapeHTML(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func (d *orderedJSON) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeOrderedValue(decoder)
	if err != nil {
		return err
	}
	obj, ok := value.(*orderedObject)
	if !ok {
		return fmt.Errorf("expected a JSON object at the document root")
	}
	d.root = obj
	return nil
}

func (d orderedJSON) MarshalJSON() ([]byte, error) {
	if d.root == nil {
		return []byte("null"), nil
	}
	return d.root.MarshalJSON()
}

// decodeOrderedValue reads exactly one JSON value from the token stream,
// recursing into objects and arrays.
func decodeOrderedValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	return decodeOrderedFrom(decoder, token)
}

func decodeOrderedFrom(decoder *json.Decoder, token json.Token) (any, error) {
	switch tok := token.(type) {
	case json.Delim:
		switch tok {
		case '{':
			obj := newOrderedObject()
			for {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				if delim, ok := keyToken.(json.Delim); ok && delim == '}' {
					return obj, nil
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, fmt.Errorf("expected object key, got %v", keyToken)
				}
				value, err := decodeOrderedValue(decoder)
				if err != nil {
					return nil, err
				}
				obj.set(key, value)
			}
		case '[':
			items := []any{}
			for {
				itemToken, err := decoder.Token()
				if err != nil {
					if err == io.EOF {
						return nil, fmt.Errorf("unexpected end of array")
					}
					return nil, err
				}
				if delim, ok := itemToken.(json.Delim); ok && delim == ']' {
					return items, nil
				}
				value, err := decodeOrderedFrom(decoder, itemToken)
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
		default:
			return nil, fmt.Errorf("unexpected delimiter %v", tok)
		}
	default:
		return token, nil
	}
}
