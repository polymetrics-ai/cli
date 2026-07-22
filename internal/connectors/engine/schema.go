package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// Schema is a compiled instance of the engine's minimal draft-07 subset. It is
// compiled once (CompileSchema) and validated many times (Validate). The
// compiled form is opaque outside the package; callers only see the accessor
// methods below.
type Schema struct {
	node *schemaNode
}

// schemaNode is the compiled representation of one (sub-)schema object.
type schemaNode struct {
	// types holds the accepted JSON types ("string", "number", "integer",
	// "boolean", "object", "array", "null"); empty means "any type".
	types []string

	required             []string
	properties           map[string]*schemaNode
	items                *schemaNode
	enum                 []any
	pattern              *regexp.Regexp
	minProperties        int
	hasMinProperties     bool
	additionalProperties bool // true unless explicitly set to false
	hasAdditionalProps   bool

	// extensions
	secret      bool     // x-secret
	primaryKey  []string // x-primary-key (only meaningful at the root)
	cursorField string   // x-cursor-field (only meaningful at the root)

	// defaultVal/hasDefault capture the "default" annotation keyword's raw
	// decoded value (gap-loop cycle-1 item 6, REVIEW-A.md C3): "default" was
	// previously accepted-but-only-preserved (never read back out anywhere);
	// this is now consumed by Defaults() so the engine can materialize a
	// spec.json property's default into RuntimeConfig.Config when a caller's
	// config omits that key entirely.
	defaultVal any
	hasDefault bool
}

// annotationKeywords are accepted but only preserved, never enforced.
var annotationKeywords = map[string]bool{
	"format":      true,
	"default":     true,
	"title":       true,
	"description": true,
	"$schema":     true,
}

// structuralKeywords are the only keywords this dialect understands
// structurally.
var structuralKeywords = map[string]bool{
	"type":                 true,
	"required":             true,
	"properties":           true,
	"items":                true,
	"enum":                 true,
	"pattern":              true,
	"minProperties":        true,
	"additionalProperties": true,
	"x-secret":             true,
	"x-primary-key":        true,
	"x-cursor-field":       true,
}

var validTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"integer": true,
	"boolean": true,
	"object":  true,
	"array":   true,
	"null":    true,
}

// CompileSchema parses and compiles a draft-07 subset schema document. Unknown
// keywords are a compile error, keeping bundles honest.
func CompileSchema(raw json.RawMessage) (*Schema, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("compile schema: invalid json: %w", err)
	}
	node, err := compileNode(m)
	if err != nil {
		return nil, err
	}
	return &Schema{node: node}, nil
}

func compileNode(m map[string]json.RawMessage) (*schemaNode, error) {
	for k := range m {
		if annotationKeywords[k] || structuralKeywords[k] {
			continue
		}
		return nil, fmt.Errorf("compile schema: unknown keyword %q", k)
	}

	n := &schemaNode{additionalProperties: true}

	if raw, ok := m["type"]; ok {
		types, err := compileTypes(raw)
		if err != nil {
			return nil, err
		}
		n.types = types
	}

	if raw, ok := m["required"]; ok {
		var req []string
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("compile schema: required: %w", err)
		}
		n.required = req
	}

	if raw, ok := m["properties"]; ok {
		var props map[string]map[string]json.RawMessage
		if err := json.Unmarshal(raw, &props); err != nil {
			return nil, fmt.Errorf("compile schema: properties: %w", err)
		}
		n.properties = make(map[string]*schemaNode, len(props))
		for name, sub := range props {
			child, err := compileNode(sub)
			if err != nil {
				return nil, fmt.Errorf("compile schema: properties.%s: %w", name, err)
			}
			n.properties[name] = child
		}
	}

	if raw, ok := m["items"]; ok {
		var sub map[string]json.RawMessage
		if err := json.Unmarshal(raw, &sub); err != nil {
			return nil, fmt.Errorf("compile schema: items: %w", err)
		}
		child, err := compileNode(sub)
		if err != nil {
			return nil, fmt.Errorf("compile schema: items: %w", err)
		}
		n.items = child
	}

	if raw, ok := m["enum"]; ok {
		var vals []any
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.UseNumber()
		if err := dec.Decode(&vals); err != nil {
			return nil, fmt.Errorf("compile schema: enum: %w", err)
		}
		n.enum = vals
	}

	if raw, ok := m["pattern"]; ok {
		var pat string
		if err := json.Unmarshal(raw, &pat); err != nil {
			return nil, fmt.Errorf("compile schema: pattern: %w", err)
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("compile schema: pattern %q: %w", pat, err)
		}
		n.pattern = re
	}

	if raw, ok := m["minProperties"]; ok {
		var mp int
		if err := json.Unmarshal(raw, &mp); err != nil {
			return nil, fmt.Errorf("compile schema: minProperties: %w", err)
		}
		n.minProperties = mp
		n.hasMinProperties = true
	}

	if raw, ok := m["default"]; ok {
		var def any
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.UseNumber()
		if err := dec.Decode(&def); err != nil {
			return nil, fmt.Errorf("compile schema: default: %w", err)
		}
		n.defaultVal = def
		n.hasDefault = true
	}

	if raw, ok := m["additionalProperties"]; ok {
		var ap bool
		if err := json.Unmarshal(raw, &ap); err != nil {
			return nil, fmt.Errorf("compile schema: additionalProperties: only bool form supported: %w", err)
		}
		n.additionalProperties = ap
		n.hasAdditionalProps = true
	}

	if raw, ok := m["x-secret"]; ok {
		var secret bool
		if err := json.Unmarshal(raw, &secret); err != nil {
			return nil, fmt.Errorf("compile schema: x-secret: %w", err)
		}
		n.secret = secret
	}

	if raw, ok := m["x-primary-key"]; ok {
		var pk []string
		if err := json.Unmarshal(raw, &pk); err != nil {
			return nil, fmt.Errorf("compile schema: x-primary-key: %w", err)
		}
		n.primaryKey = pk
	}

	if raw, ok := m["x-cursor-field"]; ok {
		var cf string
		if err := json.Unmarshal(raw, &cf); err != nil {
			return nil, fmt.Errorf("compile schema: x-cursor-field: %w", err)
		}
		n.cursorField = cf
	}

	return n, nil
}

func compileTypes(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		if !validTypes[single] {
			return nil, fmt.Errorf("compile schema: unknown type %q", single)
		}
		return []string{single}, nil
	}
	var multi []string
	if err := json.Unmarshal(raw, &multi); err != nil {
		return nil, fmt.Errorf("compile schema: type: %w", err)
	}
	for _, t := range multi {
		if !validTypes[t] {
			return nil, fmt.Errorf("compile schema: unknown type %q", t)
		}
	}
	return multi, nil
}

// Validate checks v (already decoded via encoding/json, ideally with
// UseNumber for integer fidelity) against the compiled schema. Errors name a
// JSON-pointer-ish path to the offending value.
func (s *Schema) Validate(v any) error {
	return s.node.validate(v, "")
}

func (n *schemaNode) validate(v any, path string) error {
	if len(n.types) > 0 && !typeMatches(v, n.types) {
		return fmt.Errorf("%s: value does not match type %v", displayPath(path), n.types)
	}

	if len(n.enum) > 0 {
		matched := false
		for _, want := range n.enum {
			if enumEquals(v, want) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%s: value not in enum %v", displayPath(path), n.enum)
		}
	}

	switch val := v.(type) {
	case string:
		if n.pattern != nil && !n.pattern.MatchString(val) {
			return fmt.Errorf("%s: value does not match pattern %q", displayPath(path), n.pattern.String())
		}
	case map[string]any:
		if err := n.validateObject(val, path); err != nil {
			return err
		}
	}
	if elems, ok := arrayElements(v); ok && n.items != nil {
		for i, elem := range elems {
			if err := n.items.validate(elem, fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *schemaNode) validateObject(obj map[string]any, path string) error {
	if n.hasMinProperties && len(obj) < n.minProperties {
		return fmt.Errorf("%s: minProperties %d not satisfied (got %d)", displayPath(path), n.minProperties, len(obj))
	}

	for _, req := range n.required {
		if _, ok := obj[req]; !ok {
			return fmt.Errorf("%s/%s: required property missing", displayPath(path), req)
		}
	}

	if n.hasAdditionalProps && !n.additionalProperties {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if _, declared := n.properties[k]; !declared {
				return fmt.Errorf("%s/%s: additional property not allowed", displayPath(path), k)
			}
		}
	}

	if n.properties != nil {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child, ok := n.properties[k]
			if !ok {
				continue
			}
			if err := child.validate(obj[k], path+"/"+k); err != nil {
				return err
			}
		}
	}

	return nil
}

func displayPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

// typeMatches reports whether v's JSON-decoded runtime type is one of types.
func typeMatches(v any, types []string) bool {
	for _, t := range types {
		if valueIsType(v, t) {
			return true
		}
	}
	return false
}

func arrayElements(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}
	if elems, ok := v.([]any); ok {
		return elems, true
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	elems := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elems[i] = rv.Index(i).Interface()
	}
	return elems, true
}

func valueIsType(v any, t string) bool {
	switch t {
	case "null":
		return v == nil
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := arrayElements(v)
		return ok
	case "integer":
		return isIntegerNumber(v)
	case "number":
		return isNumber(v)
	default:
		return false
	}
}

func isNumber(v any) bool {
	switch v.(type) {
	case json.Number, float64, float32, int, int64:
		return true
	default:
		return false
	}
}

func isIntegerNumber(v any) bool {
	switch n := v.(type) {
	case json.Number:
		if _, err := n.Int64(); err == nil {
			return true
		}
		f, err := n.Float64()
		return err == nil && f == float64(int64(f))
	case float64:
		return n == float64(int64(n))
	case int, int64:
		return true
	default:
		return false
	}
}

func enumEquals(v, want any) bool {
	vn, vok := normalizeNumber(v)
	wn, wok := normalizeNumber(want)
	if vok && wok {
		return vn == wn
	}
	return fmt.Sprint(v) == fmt.Sprint(want) && sameKind(v, want)
}

func sameKind(a, b any) bool {
	switch a.(type) {
	case string:
		_, ok := b.(string)
		return ok
	case bool:
		_, ok := b.(bool)
		return ok
	default:
		return true
	}
}

func normalizeNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// SecretKeys returns the top-level property names marked x-secret: true.
func (s *Schema) SecretKeys() []string {
	if s.node.properties == nil {
		return nil
	}
	var out []string
	for name, child := range s.node.properties {
		if child.secret {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// Properties returns the top-level declared property names.
func (s *Schema) Properties() []string {
	if s.node.properties == nil {
		return nil
	}
	out := make([]string, 0, len(s.node.properties))
	for name := range s.node.properties {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Defaults returns the top-level property-name -> stringified-default map
// for every root property that declares a JSON Schema "default" annotation
// (gap-loop cycle-1 item 6, REVIEW-A.md C3). The stored default's raw
// decoded JSON value is stringified via the same rules
// engine/interpolate.go's stringify() uses for every other config-shaped
// value (a JSON string returns verbatim; a number/bool/other value is
// formatted via fmt.Sprint) so a materialized default slots into
// RuntimeConfig.Config — a map[string]string — exactly like any other
// caller-supplied config value would. A property with no "default" key at
// all is simply absent from the returned map (never a zero-value entry).
func (s *Schema) Defaults() map[string]string {
	if s.node.properties == nil {
		return nil
	}
	var out map[string]string
	for name, child := range s.node.properties {
		if !child.hasDefault {
			continue
		}
		if out == nil {
			out = make(map[string]string, len(s.node.properties))
		}
		out[name] = stringify(child.defaultVal)
	}
	return out
}

// DefaultTypeMismatches returns the sorted list of root property names whose
// declared "default" JSON value does NOT type-check against that same
// property's declared "type" (gap-loop cycle-1 item 6 validate rule,
// REVIEW-A.md C3: "default must type-check"). Used by
// `cmd/connectorgen validate` to hard-fail a spec.json whose default would
// silently materialize a value of the wrong shape into config (e.g. a
// default: 100 on a "type":"string" property, or a default:"yes" on a
// "type":"boolean" property) — connectorgen's own compiled *Schema already
// enforces "properties" structurally, so this reuses the same type-matching
// logic Validate() uses for ordinary instance data, applied to the default
// value itself.
func (s *Schema) DefaultTypeMismatches() []string {
	if s.node.properties == nil {
		return nil
	}
	var out []string
	for name, child := range s.node.properties {
		if !child.hasDefault {
			continue
		}
		if len(child.types) > 0 && !typeMatches(child.defaultVal, child.types) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// RequiredKeys returns the root-level "required" property name list (F4,
// REVIEW.md: resolveHeaders needs to distinguish a declared-but-OPTIONAL
// config key, whose runtime absence is tolerated by omitting the header that
// references it — e.g. Stripe-Account/account_id — from a declared-and-
// REQUIRED one, whose absence is a hard configuration error, never silently
// swallowed).
func (s *Schema) RequiredKeys() []string {
	return s.node.required
}

// PrimaryKeys returns the root-level x-primary-key list.
func (s *Schema) PrimaryKeys() []string {
	return s.node.primaryKey
}

// CursorFieldName returns the root-level x-cursor-field value ("" when unset).
func (s *Schema) CursorFieldName() string {
	return s.node.cursorField
}

// StreamSchema pairs a compiled record schema with its primary key and cursor
// field extensions, extracted once at bundle-load time for convenient reuse.
type StreamSchema struct {
	*Schema
	PrimaryKey  []string // x-primary-key
	CursorField string   // x-cursor-field
}
