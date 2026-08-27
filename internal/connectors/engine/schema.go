package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
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
	patternProperties    []schemaPatternProperty
	items                *schemaNode
	prefixItems          []*schemaNode
	oneOf                []*schemaNode
	anyOf                []*schemaNode
	allOf                []*schemaNode
	enum                 []any
	pattern              *regexp.Regexp
	format               string
	minLength            int
	hasMinLength         bool
	maxLength            int
	hasMaxLength         bool
	minProperties        int
	hasMinProperties     bool
	maxProperties        int
	hasMaxProperties     bool
	additionalProperties bool // true unless explicitly set to false
	hasAdditionalProps   bool

	// minItems/maxItems are the array-cardinality pair. Each carries its own
	// has* flag rather than relying on the zero value, because an EXPLICIT
	// "minItems": 0 (a documented "may be empty") must stay distinguishable
	// from an absent keyword — the same distinction PaginationSpec.StartPage
	// solves one layer up with a pointer.
	//
	// Bounded provider search depends on the same pair: an "ids[] bounded to
	// 100" contract has to be declarable in the bundle and enforceable at load
	// time, and unknown keywords are a compile error in this dialect.
	minItems    int
	hasMinItems bool
	maxItems    int
	hasMaxItems bool

	minimum    *big.Rat
	hasMinimum bool
	maximum    *big.Rat
	hasMaximum bool

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

type schemaPatternProperty struct {
	pattern *regexp.Regexp
	schema  *schemaNode
}

// annotationKeywords are accepted by the schema compiler. Supported formats
// participate in instance validation, and configuration-time validation
// evaluates a bundled spec's declared top-level format constraints.
var annotationKeywords = map[string]bool{
	"format":  true,
	"default": true,
	// example is an OpenAPI/JSON Schema annotation. It documents a provider
	// value but never changes whether an instance validates, so compilation
	// accepts it without treating the example itself as executable input.
	"example":     true,
	"title":       true,
	"description": true,
	"$schema":     true,
	// Source import preserves an OpenAPI discriminator as provenance for a
	// closed oneOf contract. oneOf still supplies executable selection; this
	// annotation retains the provider's declared selector without widening it.
	"x-source-discriminator": true,
}

// structuralKeywords are the only keywords this dialect understands
// structurally.
var structuralKeywords = map[string]bool{
	"type":                 true,
	"required":             true,
	"properties":           true,
	"patternProperties":    true,
	"items":                true,
	"oneOf":                true,
	"anyOf":                true,
	"allOf":                true,
	"enum":                 true,
	"pattern":              true,
	"minLength":            true,
	"maxLength":            true,
	"minProperties":        true,
	"maxProperties":        true,
	"minItems":             true,
	"maxItems":             true,
	"minimum":              true,
	"maximum":              true,
	"additionalProperties": true,
	"x-secret":             true,
	"x-primary-key":        true,
	"x-cursor-field":       true,
	// Dynamic catalog schemas retain these provider-derived annotations. The
	// engine's executable static sync contract remains x-primary-key and
	// x-cursor-field, while the additional fields let catalog consumers keep
	// the same draft-07 document regardless of schema origin.
	"x-stream_name":                true,
	"x-supported_sync_modes":       true,
	"x-default_sync_mode":          true,
	"x-source_defined_primary_key": true,
	"x-source_defined_cursor":      true,
	"x-default_cursor_field":       true,
	"x-references":                 true,
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
	return compileSchema(raw, false)
}

func compileStructuredRESTBodySchemaDocument(raw json.RawMessage) (*Schema, error) {
	return compileSchema(raw, true)
}

func compileSchema(raw json.RawMessage, allowPrefixItems bool) (*Schema, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("compile schema: invalid json: %w", err)
	}
	node, err := compileNode(m, allowPrefixItems)
	if err != nil {
		return nil, err
	}
	return &Schema{node: node}, nil
}

func compileNode(m map[string]json.RawMessage, allowPrefixItems bool) (*schemaNode, error) {
	for k := range m {
		if annotationKeywords[k] || structuralKeywords[k] || (allowPrefixItems && k == "prefixItems") {
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
			child, err := compileNode(sub, allowPrefixItems)
			if err != nil {
				return nil, fmt.Errorf("compile schema: properties.%s: %w", name, err)
			}
			n.properties[name] = child
		}
	}

	if raw, ok := m["patternProperties"]; ok {
		var props map[string]map[string]json.RawMessage
		if err := json.Unmarshal(raw, &props); err != nil || props == nil {
			if err != nil {
				return nil, fmt.Errorf("compile schema: patternProperties: %w", err)
			}
			return nil, fmt.Errorf("compile schema: patternProperties must be an object")
		}
		patterns := make([]string, 0, len(props))
		for pattern := range props {
			patterns = append(patterns, pattern)
		}
		sort.Strings(patterns)
		n.patternProperties = make([]schemaPatternProperty, 0, len(patterns))
		for _, pattern := range patterns {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile schema: patternProperties.%q: %w", pattern, err)
			}
			child, err := compileNode(props[pattern], allowPrefixItems)
			if err != nil {
				return nil, fmt.Errorf("compile schema: patternProperties.%s: %w", pattern, err)
			}
			n.patternProperties = append(n.patternProperties, schemaPatternProperty{pattern: compiled, schema: child})
		}
	}

	if raw, ok := m["items"]; ok {
		var sub map[string]json.RawMessage
		if err := json.Unmarshal(raw, &sub); err != nil {
			return nil, fmt.Errorf("compile schema: items: %w", err)
		}
		child, err := compileNode(sub, allowPrefixItems)
		if err != nil {
			return nil, fmt.Errorf("compile schema: items: %w", err)
		}
		n.items = child
	}

	for _, composition := range []string{"oneOf", "anyOf", "allOf"} {
		raw, ok := m[composition]
		if !ok {
			continue
		}
		children, err := compileComposition(raw, composition, allowPrefixItems)
		if err != nil {
			return nil, err
		}
		switch composition {
		case "oneOf":
			n.oneOf = children
		case "anyOf":
			n.anyOf = children
		case "allOf":
			if err := validateAllOfConsistency(children); err != nil {
				return nil, err
			}
			n.allOf = children
		}
	}

	if raw, ok := m["prefixItems"]; ok {
		var subs []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &subs); err != nil || subs == nil {
			if err != nil {
				return nil, fmt.Errorf("compile schema: prefixItems: %w", err)
			}
			return nil, fmt.Errorf("compile schema: prefixItems must be an array")
		}
		n.prefixItems = make([]*schemaNode, len(subs))
		for index, sub := range subs {
			if sub == nil {
				return nil, fmt.Errorf("compile schema: prefixItems.%d must be a schema object", index)
			}
			child, err := compileNode(sub, allowPrefixItems)
			if err != nil {
				return nil, fmt.Errorf("compile schema: prefixItems.%d: %w", index, err)
			}
			n.prefixItems[index] = child
		}
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
	if err := compileStringLength(m, n); err != nil {
		return nil, err
	}

	if raw, ok := m["format"]; ok {
		if err := json.Unmarshal(raw, &n.format); err != nil {
			return nil, fmt.Errorf("compile schema: format: %w", err)
		}
	}

	if raw, ok := m["minProperties"]; ok {
		mp, err := compileNonNegativeInt(raw, "minProperties")
		if err != nil {
			return nil, err
		}
		n.minProperties = mp
		n.hasMinProperties = true
	}
	if raw, ok := m["maxProperties"]; ok {
		mp, err := compileNonNegativeInt(raw, "maxProperties")
		if err != nil {
			return nil, err
		}
		n.maxProperties = mp
		n.hasMaxProperties = true
	}
	if n.hasMinProperties && n.hasMaxProperties && n.maxProperties < n.minProperties {
		return nil, fmt.Errorf("compile schema: maxProperties %d is below minProperties %d", n.maxProperties, n.minProperties)
	}

	if err := compileArrayCardinality(m, n); err != nil {
		return nil, err
	}
	if err := compileNumericRange(m, n); err != nil {
		return nil, err
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

// compileComposition compiles the three JSON Schema composition keywords into
// the same closed node dialect as ordinary properties. A duplicate alternative
// is not merely redundant here: it makes oneOf selection ambiguous and hides
// provider-spec drift in anyOf/allOf, so declaration admission rejects it.
func compileComposition(raw json.RawMessage, keyword string, allowPrefixItems bool) ([]*schemaNode, error) {
	var subs []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &subs); err != nil || len(subs) == 0 {
		if err != nil {
			return nil, fmt.Errorf("compile schema: %s: %w", keyword, err)
		}
		return nil, fmt.Errorf("compile schema: %s must contain at least one schema object", keyword)
	}

	children := make([]*schemaNode, len(subs))
	seen := make(map[string]int, len(subs))
	for index, sub := range subs {
		if sub == nil {
			return nil, fmt.Errorf("compile schema: %s.%d must be a schema object", keyword, index)
		}
		canonical, err := json.Marshal(sub)
		if err != nil {
			return nil, fmt.Errorf("compile schema: %s.%d canonicalize: %w", keyword, index, err)
		}
		if prior, duplicate := seen[string(canonical)]; duplicate {
			return nil, fmt.Errorf("compile schema: %s.%d duplicates alternative %d", keyword, index, prior)
		}
		seen[string(canonical)] = index
		child, err := compileNode(sub, allowPrefixItems)
		if err != nil {
			return nil, fmt.Errorf("compile schema: %s.%d: %w", keyword, index, err)
		}
		children[index] = child
	}
	return children, nil
}

// validateAllOfConsistency rejects intersections that can never match before
// a request reaches a provider. More involved intersections still retain every
// child validator and therefore fail in Schema.Validate before transport.
func validateAllOfConsistency(children []*schemaNode) error {
	var types []string
	var enum []any
	for _, child := range children {
		if len(child.types) > 0 {
			if types == nil {
				types = append([]string(nil), child.types...)
			} else {
				types = intersectSchemaTypes(types, child.types)
				if len(types) == 0 {
					return fmt.Errorf("compile schema: allOf has contradictory type constraints")
				}
			}
		}
		if len(child.enum) == 0 {
			continue
		}
		if enum == nil {
			enum = append([]any(nil), child.enum...)
			continue
		}
		intersection := make([]any, 0, len(enum))
		for _, candidate := range enum {
			for _, permitted := range child.enum {
				if enumEquals(candidate, permitted) {
					intersection = append(intersection, candidate)
					break
				}
			}
		}
		if len(intersection) == 0 {
			return fmt.Errorf("compile schema: allOf has contradictory enum constraints")
		}
		enum = intersection
	}
	return nil
}

func intersectSchemaTypes(left, right []string) []string {
	seen := make(map[string]bool, len(left))
	for _, leftType := range left {
		for _, rightType := range right {
			switch {
			case leftType == rightType:
				seen[leftType] = true
			case (leftType == "number" && rightType == "integer") || (leftType == "integer" && rightType == "number"):
				seen["integer"] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for typeName := range seen {
		out = append(out, typeName)
	}
	sort.Strings(out)
	return out
}

// compileStringLength accepts the draft-07 string cardinality pair. Header
// parameter schemas need the same declaration-owned length semantics as any
// other scalar schema; keeping it in the shared compiler prevents a parallel
// one-off validator from drifting.
func compileStringLength(m map[string]json.RawMessage, n *schemaNode) error {
	if raw, ok := m["minLength"]; ok {
		v, err := compileNonNegativeInt(raw, "minLength")
		if err != nil {
			return err
		}
		n.minLength = v
		n.hasMinLength = true
	}
	if raw, ok := m["maxLength"]; ok {
		v, err := compileNonNegativeInt(raw, "maxLength")
		if err != nil {
			return err
		}
		n.maxLength = v
		n.hasMaxLength = true
	}
	if n.hasMinLength && n.hasMaxLength && n.maxLength < n.minLength {
		return fmt.Errorf("compile schema: maxLength %d is below minLength %d", n.maxLength, n.minLength)
	}
	return nil
}

// compileArrayCardinality compiles the draft-07 minItems/maxItems pair.
//
// The engine dialect deliberately had no array-cardinality keyword, which made
// "this documented request array must not be empty" inexpressible: a bundle
// could declare an array field required, but not that it carries at least one
// element, so an operation whose provider rejects an empty array could not be
// declared executable without risking a malformed request. Two bundles already
// document the gap in prose (defs/drip/writes.json, defs/zoho-bigin/writes.json)
// and Airtable's operation ledger blocks 25 endpoints on it by name.
//
// Both bounds must be non-negative integers (draft-07), and a declared maxItems
// below a declared minItems is unsatisfiable — always an authoring mistake, so
// it fails at compile time rather than rejecting every instance at runtime.
func compileArrayCardinality(m map[string]json.RawMessage, n *schemaNode) error {
	if raw, ok := m["minItems"]; ok {
		v, err := compileNonNegativeInt(raw, "minItems")
		if err != nil {
			return err
		}
		n.minItems = v
		n.hasMinItems = true
	}
	if raw, ok := m["maxItems"]; ok {
		v, err := compileNonNegativeInt(raw, "maxItems")
		if err != nil {
			return err
		}
		n.maxItems = v
		n.hasMaxItems = true
	}
	if n.hasMinItems && n.hasMaxItems && n.maxItems < n.minItems {
		return fmt.Errorf("compile schema: maxItems %d is below minItems %d", n.maxItems, n.minItems)
	}
	return nil
}

// compileNumericRange keeps numeric provider domains exact. In particular,
// GraphQL Int is a signed 32-bit value even on a 64-bit host, so float-based
// range storage would make the declaration's wire contract platform-dependent.
func compileNumericRange(m map[string]json.RawMessage, n *schemaNode) error {
	compile := func(keyword string) (*big.Rat, bool, error) {
		raw, ok := m[keyword]
		if !ok {
			return nil, false, nil
		}
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, false, fmt.Errorf("compile schema: %s: %w", keyword, err)
		}
		if err := requireEOF(decoder, "compile schema: "+keyword); err != nil {
			return nil, false, err
		}
		number, ok := value.(json.Number)
		if !ok {
			return nil, false, fmt.Errorf("compile schema: %s must be a number", keyword)
		}
		rat, ok := exactNumber(number)
		if !ok {
			return nil, false, fmt.Errorf("compile schema: %s must be a finite number", keyword)
		}
		return rat, true, nil
	}
	minimum, hasMinimum, err := compile("minimum")
	if err != nil {
		return err
	}
	maximum, hasMaximum, err := compile("maximum")
	if err != nil {
		return err
	}
	if hasMinimum && hasMaximum && maximum.Cmp(minimum) < 0 {
		return fmt.Errorf("compile schema: maximum %s is below minimum %s", maximum.RatString(), minimum.RatString())
	}
	n.minimum, n.hasMinimum = minimum, hasMinimum
	n.maximum, n.hasMaximum = maximum, hasMaximum
	return nil
}

func requireEOF(decoder *json.Decoder, label string) error {
	var trailing any
	if err := decoder.Decode(&trailing); err == io.EOF {
		return nil
	} else if err == nil {
		return fmt.Errorf("%s contains multiple JSON values", label)
	} else {
		return fmt.Errorf("%s: %w", label, err)
	}
}

func compileNonNegativeInt(raw json.RawMessage, keyword string) (int, error) {
	var v int
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0, fmt.Errorf("compile schema: %s: %w", keyword, err)
	}
	if v < 0 {
		return 0, fmt.Errorf("compile schema: %s must be non-negative, got %d", keyword, v)
	}
	return v, nil
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
	if len(n.oneOf) != 0 {
		matches := 0
		for _, alternative := range n.oneOf {
			if err := alternative.validate(v, path); err == nil {
				matches++
			}
		}
		if matches != 1 {
			return fmt.Errorf("%s: oneOf expected exactly one schema match, got %d", displayPath(path), matches)
		}
	}
	if len(n.anyOf) != 0 {
		matches := 0
		for _, alternative := range n.anyOf {
			if err := alternative.validate(v, path); err == nil {
				matches++
			}
		}
		if matches == 0 {
			return fmt.Errorf("%s: anyOf expected at least one schema match", displayPath(path))
		}
	}
	for index, alternative := range n.allOf {
		if err := alternative.validate(v, path); err != nil {
			return fmt.Errorf("%s: allOf alternative %d: %w", displayPath(path), index, err)
		}
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
	if number, ok := exactNumber(v); ok {
		if n.hasMinimum && number.Cmp(n.minimum) < 0 {
			return fmt.Errorf("%s: minimum %s not satisfied (got %s)", displayPath(path), n.minimum.RatString(), number.RatString())
		}
		if n.hasMaximum && number.Cmp(n.maximum) > 0 {
			return fmt.Errorf("%s: maximum %s exceeded (got %s)", displayPath(path), n.maximum.RatString(), number.RatString())
		}
	}

	switch val := v.(type) {
	case string:
		if n.pattern != nil && !n.pattern.MatchString(val) {
			return fmt.Errorf("%s: value does not match pattern %q", displayPath(path), n.pattern.String())
		}
		if n.hasMinLength && len([]rune(val)) < n.minLength {
			return fmt.Errorf("%s: minLength %d not satisfied (got %d)", displayPath(path), n.minLength, len([]rune(val)))
		}
		if n.hasMaxLength && len([]rune(val)) > n.maxLength {
			return fmt.Errorf("%s: maxLength %d exceeded (got %d)", displayPath(path), n.maxLength, len([]rune(val)))
		}
		if n.format == "uri" && !validURI(val) {
			return fmt.Errorf("%s: value does not match format %q", displayPath(path), n.format)
		}
	case map[string]any:
		if err := n.validateObject(val, path); err != nil {
			return err
		}
	}
	if elems, ok := arrayElements(v); ok {
		// Cardinality applies to array INSTANCES only, per draft-07
		// applicability. "required and non-empty" is therefore the composition
		// required + minItems, exactly as it is in real draft-07 — enforcing on
		// an absent value instead would silently change the meaning of every
		// optional array field already declared in a bundle.
		if err := n.validateArrayCardinality(len(elems), path); err != nil {
			return err
		}
		for i, elem := range elems {
			if i < len(n.prefixItems) {
				if err := n.prefixItems[i].validate(elem, fmt.Sprintf("%s/%d", path, i)); err != nil {
					return err
				}
				continue
			}
			if n.items != nil {
				if err := n.items.validate(elem, fmt.Sprintf("%s/%d", path, i)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validURI(value string) bool {
	if strings.Contains(value, `\`) || strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.IsAbs()
}

func (n *schemaNode) validateArrayCardinality(count int, path string) error {
	if n.hasMinItems && count < n.minItems {
		return fmt.Errorf("%s: minItems %d not satisfied (got %d)", displayPath(path), n.minItems, count)
	}
	if n.hasMaxItems && count > n.maxItems {
		return fmt.Errorf("%s: maxItems %d exceeded (got %d)", displayPath(path), n.maxItems, count)
	}
	return nil
}

func (n *schemaNode) validateObject(obj map[string]any, path string) error {
	if n.hasMinProperties && len(obj) < n.minProperties {
		return fmt.Errorf("%s: minProperties %d not satisfied (got %d)", displayPath(path), n.minProperties, len(obj))
	}
	if n.hasMaxProperties && len(obj) > n.maxProperties {
		return fmt.Errorf("%s: maxProperties %d exceeded (got %d)", displayPath(path), n.maxProperties, len(obj))
	}

	for _, req := range n.required {
		if _, ok := obj[req]; !ok {
			return fmt.Errorf("%s/%s: required property missing", displayPath(path), req)
		}
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		declared := false
		if child, ok := n.properties[k]; ok {
			declared = true
			if err := child.validate(obj[k], path+"/"+k); err != nil {
				return err
			}
		}
		for _, pattern := range n.patternProperties {
			if !pattern.pattern.MatchString(k) {
				continue
			}
			declared = true
			if err := pattern.schema.validate(obj[k], path+"/"+k); err != nil {
				return err
			}
		}
		if n.hasAdditionalProps && !n.additionalProperties && !declared {
			return fmt.Errorf("%s/%s: additional property not allowed", displayPath(path), k)
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
	_, ok := exactNumber(v)
	return ok
}

func isIntegerNumber(v any) bool {
	n, ok := exactNumber(v)
	return ok && n.IsInt()
}

func enumEquals(v, want any) bool {
	vn, vok := exactNumber(v)
	wn, wok := exactNumber(want)
	if vok && wok {
		return vn.Cmp(wn) == 0
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

func exactNumber(v any) (*big.Rat, bool) {
	integer := func(value string) (*big.Rat, bool) {
		n := new(big.Int)
		if _, ok := n.SetString(value, 10); !ok {
			return nil, false
		}
		return new(big.Rat).SetInt(n), true
	}
	switch n := v.(type) {
	case json.Number:
		r, ok := new(big.Rat).SetString(n.String())
		return r, ok
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return nil, false
		}
		return new(big.Rat).SetFloat64(n), true
	case float32:
		f := float64(n)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, false
		}
		return new(big.Rat).SetFloat64(f), true
	case int:
		return integer(strconv.FormatInt(int64(n), 10))
	case int8:
		return integer(strconv.FormatInt(int64(n), 10))
	case int16:
		return integer(strconv.FormatInt(int64(n), 10))
	case int32:
		return integer(strconv.FormatInt(int64(n), 10))
	case int64:
		return integer(strconv.FormatInt(n, 10))
	case uint:
		return integer(strconv.FormatUint(uint64(n), 10))
	case uint8:
		return integer(strconv.FormatUint(uint64(n), 10))
	case uint16:
		return integer(strconv.FormatUint(uint64(n), 10))
	case uint32:
		return integer(strconv.FormatUint(uint64(n), 10))
	case uint64:
		return integer(strconv.FormatUint(n, 10))
	default:
		return nil, false
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
	// Raw is the original draft-07 record contract. Catalog projections use
	// this rather than re-deriving a lossy Stream from the compiled schema, so
	// static and provider-discovered streams share one downstream shape.
	Raw json.RawMessage
}
