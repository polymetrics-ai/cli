package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"polymetrics.ai/internal/connectors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FormEncoding is an opt-in, bounded source-selected form dialect. Fields name
// schema properties, never wire-path aliases. Absent metadata keeps legacy forms.
type FormEncoding struct {
	Version    int                          `json:"version"`
	MaxDepth   int                          `json:"max_depth"`
	MaxMembers int                          `json:"max_members"`
	MaxItems   int                          `json:"max_items"`
	MaxPairs   int                          `json:"max_pairs"`
	MaxBytes   int                          `json:"max_bytes"`
	Fields     map[string]FormFieldEncoding `json:"fields"`
}

// FormFieldEncoding controls a field and its nested values. Arrays inside a
// brackets field require Arrays; no implicit array-of-object dialect exists.
type FormFieldEncoding struct {
	Mode       string `json:"mode"`
	Arrays     string `json:"arrays,omitempty"`
	EmptyArray string `json:"empty_array,omitempty"`
	Null       string `json:"null,omitempty"`
}

func validateFormEncoding(f *FormEncoding, raw json.RawMessage, excluded []string) error {
	if f == nil {
		return nil
	}
	if f.Version != 1 {
		return fmt.Errorf("form_encoding version must be 1")
	}
	if f.MaxDepth < 1 || f.MaxDepth > 32 || f.MaxMembers < 1 || f.MaxMembers > 10000 || f.MaxItems < 1 || f.MaxItems > 10000 || f.MaxPairs < 1 || f.MaxPairs > 10000 || f.MaxBytes < 1 || f.MaxBytes > 16<<20 {
		return fmt.Errorf("form_encoding requires finite depth/member/item/pair/byte limits within implementation guards")
	}
	if _, err := CompileSchema(raw); err != nil {
		return fmt.Errorf("form_encoding schema: %w", err)
	}
	var root struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &root); err != nil || root.Type != "object" {
		return fmt.Errorf("form_encoding requires an object schema")
	}
	skip := map[string]bool{}
	for _, k := range excluded {
		skip[k] = true
	}
	for k, v := range f.Fields {
		if !validFormMember(k) || skip[k] || root.Properties[k] == nil {
			return fmt.Errorf("form_encoding field %q is not an unambiguous body schema property", k)
		}
		switch v.Mode {
		case "scalar", "json", "brackets", "repeated", "indexed", "bracket_repeated":
		default:
			return fmt.Errorf("form_encoding field %q has unsupported mode", k)
		}
		if v.Arrays != "" && v.Arrays != "indexed" && v.Arrays != "repeated" && v.Arrays != "bracket_repeated" {
			return fmt.Errorf("form_encoding field %q has unsupported arrays", k)
		}
		if v.EmptyArray != "" && v.EmptyArray != "omit" && v.EmptyArray != "empty" {
			return fmt.Errorf("form_encoding field %q has unsupported empty_array", k)
		}
		if v.Null != "" && v.Null != "reject" && v.Null != "empty" && v.Null != "literal" {
			return fmt.Errorf("form_encoding field %q has unsupported null", k)
		}

		var node map[string]any
		if err := json.Unmarshal(root.Properties[k], &node); err != nil {
			return err
		}
		if err := validateFormFieldShape(node, v, v.Mode); err != nil {
			return fmt.Errorf("form_encoding field %q: %w", k, err)
		}
	}
	for k := range root.Properties {
		if !skip[k] {
			if _, ok := f.Fields[k]; !ok {
				return fmt.Errorf("form_encoding missing field %q", k)
			}
		}
	}
	return nil
}
func validFormMember(k string) bool {
	return k != "" && utf8.ValidString(k) && !strings.ContainsAny(k, "[]\x00\r\n")
}

type formEncodingState struct {
	spec                         *FormEncoding
	pairs                        url.Values
	members, items, count, bytes int
	inputBytes                   int
}

func (s *formEncodingState) add(key, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("form value is not valid UTF-8")
	}
	size := len(url.QueryEscape(key)) + 1 + len(url.QueryEscape(value))
	if s.count > 0 {
		size++
	}
	if s.count >= s.spec.MaxPairs || size > s.spec.MaxBytes-s.bytes {
		return fmt.Errorf("form encoded pair/byte limit exceeded")
	}
	s.bytes += size
	s.count++
	s.pairs.Add(key, value)
	return nil
}

// encodeTypedForm is shared by preparation callers. Sending consumes sealed bytes.
// json.Number is the compatible exact-number boundary for the shared input codec.
func encodeTypedForm(f *FormEncoding, schema json.RawMessage, body map[string]any, excluded []string) (url.Values, string, error) {
	if err := validateFormEncoding(f, schema, excluded); err != nil {
		return nil, "", err
	}
	if f == nil {
		return nil, "", fmt.Errorf("typed form requires form_encoding")
	}
	if body == nil {
		body = map[string]any{}
	}
	// Bound and normalize before schema validation or JSON-valued field marshaling.
	state := &formEncodingState{spec: f, pairs: url.Values{}}
	normalized, err := state.normalize(reflect.ValueOf(body), 0)
	if err != nil {
		return nil, "", err
	}
	sch, err := CompileSchema(schema)
	if err != nil {
		return nil, "", err
	}
	if err = sch.Validate(normalized); err != nil {
		return nil, "", err
	}
	skip := map[string]bool{}
	for _, k := range excluded {
		skip[k] = true
	}
	m := normalized.(map[string]any)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if skip[k] {
			continue
		}
		dialect, ok := f.Fields[k]
		if !ok {
			return nil, "", fmt.Errorf("form field %q has no encoding", k)
		}
		if err := state.field(k, m[k], dialect, dialect.Mode); err != nil {
			return nil, "", fmt.Errorf("form field %q: %w", k, err)
		}
	}
	return state.pairs, state.pairs.Encode(), nil
}
func (s *formEncodingState) normalize(v reflect.Value, depth int) (any, error) {
	if depth > s.spec.MaxDepth {
		return nil, fmt.Errorf("form depth limit exceeded")
	}
	if !v.IsValid() {
		return nil, nil
	}
	if v.CanInterface() {
		if n, ok := v.Interface().(json.Number); ok {

			if len(n.String()) > 1024 {
				return nil, fmt.Errorf("oversized form number")
			}
			// The shared scalar codec owns numeric grammar and exponent work.
			// Form keeps its narrower lexeme budget and selected schema permission.
			if _, err := connectors.DecodeSourceScalar("number", n.String(), 4096); err != nil {
				return nil, fmt.Errorf("invalid form number: %w", err)
			}
			if err := s.accountInput(len(n.String())); err != nil {
				return nil, err
			}
			return n, nil
		}
		if _, ok := v.Interface().(json.Marshaler); ok {
			return nil, fmt.Errorf("form custom JSON marshalers unsupported")
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return nil, nil
		}
		return s.normalize(v.Elem(), depth+1)
	case reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		return s.normalize(v.Elem(), depth)
	case reflect.Map:
		if v.IsNil() {
			return nil, nil
		}
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("form map keys must be strings")
		}
		if v.Len() > s.spec.MaxMembers-s.members {
			return nil, fmt.Errorf("form member limit exceeded")
		}
		s.members += v.Len()
		out := map[string]any{}
		for _, k := range v.MapKeys() {
			key := k.String()
			if err := s.accountInput(len(key)); err != nil {
				return nil, err
			}
			if !utf8.ValidString(key) {
				return nil, fmt.Errorf("ambiguous form member %q", key)
			}
			w, err := s.normalize(v.MapIndex(k), depth+1)
			if err != nil {
				return nil, err
			}
			out[key] = w
		}
		return out, nil
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil, nil
		}
		if v.Len() > s.spec.MaxItems-s.items {
			return nil, fmt.Errorf("form array item limit exceeded")
		}
		s.items += v.Len()
		out := make([]any, v.Len())
		for i := range out {
			w, err := s.normalize(v.Index(i), depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = w
		}
		return out, nil
	case reflect.String:
		if !utf8.ValidString(v.String()) || len(v.String()) > s.spec.MaxBytes {
			return nil, fmt.Errorf("form string invalid or too large")
		}
		if err := s.accountInput(len(v.String())); err != nil {
			return nil, err
		}
		return v.String(), nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return json.Number(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return json.Number(strconv.FormatUint(v.Uint(), 10)), nil
	case reflect.Float32, reflect.Float64:
		n := v.Float()
		if math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) > 1<<53 {
			return nil, fmt.Errorf("form number requires exact json.Number representation")
		}
		return json.Number(strconv.FormatFloat(n, 'g', -1, v.Type().Bits())), nil
	default:
		return nil, fmt.Errorf("unsupported form value type %s", v.Type())
	}
}
func (s *formEncodingState) field(key string, v any, d FormFieldEncoding, mode string) error {
	if v == nil {
		switch d.Null {
		case "empty":
			return s.add(key, "")
		case "literal":
			return s.add(key, "null")
		default:
			return fmt.Errorf("null has no declared wire encoding")
		}
	}
	if mode == "json" {
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return s.add(key, string(raw))
	}
	switch n := v.(type) {
	case map[string]any:
		if mode != "brackets" && mode != "indexed" {
			return fmt.Errorf("object requires bracket dialect")
		}
		keys := make([]string, 0, len(n))
		for k := range n {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !validFormMember(k) {
				return fmt.Errorf("ambiguous bracket member %q", k)
			}
			if err := s.field(key+"["+k+"]", n[k], d, "brackets"); err != nil {
				return err
			}
		}
		return nil
	case []any:
		arrays := mode
		if mode == "brackets" {
			arrays = d.Arrays
		}
		if arrays != "indexed" && arrays != "repeated" && arrays != "bracket_repeated" {
			return fmt.Errorf("array requires explicit indexed or repeated dialect")
		}
		if len(n) == 0 {
			if d.EmptyArray == "empty" {
				if arrays == "bracket_repeated" {
					key += "[]"
				}
				return s.add(key, "")
			}
			return nil
		}
		for i, item := range n {
			next := key
			childmode := "scalar"
			if arrays == "bracket_repeated" {
				next += "[]"
			}
			if arrays == "indexed" {
				next += "[" + strconv.Itoa(i) + "]"
				childmode = "indexed"
			}
			if err := s.field(next, item, d, childmode); err != nil {
				return err
			}
		}
		return nil
	case string:
		return s.add(key, n)
	case bool:
		return s.add(key, strconv.FormatBool(n))
	case json.Number:
		return s.add(key, n.String())
	default:
		return fmt.Errorf("unsupported canonical form value")
	}
}

func validateFormFieldShape(node map[string]any, d FormFieldEncoding, mode string) error {
	for _, arm := range asFormSchemas(node["oneOf"]) {
		if err := validateFormFieldShape(arm, d, mode); err != nil {
			return err
		}
	}
	var types []string
	switch value := node["type"].(type) {
	case string:
		types = []string{value}
	case []any:
		for _, v := range value {
			if t, ok := v.(string); ok {
				types = append(types, t)
			}
		}
	}
	if len(types) == 0 && len(asFormSchemas(node["oneOf"])) == 0 {
		return fmt.Errorf("form field requires explicit source type or oneOf alternatives")
	}
	for _, typ := range types {
		if typ == "null" && d.Null == "" {
			return fmt.Errorf("nullable field requires explicit null wire rule")
		}
		if mode == "json" {
			continue
		}
		switch typ {
		case "object":
			if mode != "brackets" && mode != "indexed" {
				return fmt.Errorf("object requires bracket dialect")
			}
			for _, group := range []string{"properties", "patternProperties"} {
				if props, ok := node[group].(map[string]any); ok {
					for name, value := range props {
						if group == "properties" && !validFormMember(name) {
							return fmt.Errorf("ambiguous bracket property")
						}
						if child, ok := value.(map[string]any); ok {
							if err := validateFormFieldShape(child, d, "brackets"); err != nil {
								return err
							}
						}
					}
				}
			}
			if closed, ok := node["additionalProperties"].(bool); !ok || closed {
				return fmt.Errorf("bracket objects require declared properties or typed patternProperties and additionalProperties false")
			}
		case "array":
			arrays := mode
			if mode == "brackets" {
				arrays = d.Arrays
			}
			if arrays != "indexed" && arrays != "repeated" && arrays != "bracket_repeated" {
				return fmt.Errorf("array requires explicit indexed or repeated dialect")
			}
			child, ok := node["items"].(map[string]any)
			if !ok {
				return fmt.Errorf("form array requires one typed items schema")
			}
			next := "scalar"
			if arrays == "indexed" {
				next = "indexed"
			}
			if err := validateFormFieldShape(child, d, next); err != nil {
				return err
			}
		}
	}
	return nil
}
func asFormSchemas(value any) []map[string]any {
	var out []map[string]any
	if arms, ok := value.([]any); ok {
		for _, v := range arms {
			if node, ok := v.(map[string]any); ok {
				out = append(out, node)
			}
		}
	}
	return out
}

// accountInput bounds aggregate scalar/key memory before JSON-valued encoding.
func (s *formEncodingState) accountInput(size int) error {
	if size > s.spec.MaxBytes-s.inputBytes {
		return fmt.Errorf("form input scalar/key byte limit exceeded")
	}
	s.inputBytes += size
	return nil
}
