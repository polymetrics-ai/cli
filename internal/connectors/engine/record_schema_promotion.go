package engine

import (
	"encoding/json"
	"fmt"
)

// RecordSchemaShape is the promotion-relevant shape of a declarative write
// record schema. It intentionally examines raw JSON Schema before CompileSchema
// consumes it: a provider union has to be expanded before its concrete arms can
// be measured, while the engine's executable schema dialect deliberately does
// not accept unions at its command boundary.
type RecordSchemaShape struct {
	RootUnion             string
	ArmCount              int
	SubstantiveArmCount   int
	AdmitsOnlyEmptyObject bool
}

// InspectRecordSchema expands top-level oneOf/anyOf arms before measuring their
// named fields. A top-level union is not executable as one record schema, but
// its expanded shape makes the required remediation unambiguous: separate
// named actions rather than a hollow derived object.
func InspectRecordSchema(raw json.RawMessage) (RecordSchemaShape, error) {
	if len(raw) == 0 {
		return RecordSchemaShape{}, nil
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return RecordSchemaShape{}, fmt.Errorf("invalid record_schema: %w", err)
	}
	if root == nil {
		return RecordSchemaShape{}, fmt.Errorf("record_schema must be an object")
	}

	rootUnion, arms, err := expandRecordSchemaArms(root)
	if err != nil {
		return RecordSchemaShape{}, err
	}
	shape := RecordSchemaShape{RootUnion: rootUnion, ArmCount: len(arms)}
	for _, arm := range arms {
		if recordSchemaArmHasNamedFields(arm) {
			shape.SubstantiveArmCount++
		}
	}
	if rootUnion == "" && len(arms) == 1 {
		empty, err := recordSchemaArmAdmitsOnlyEmptyObject(arms[0])
		if err != nil {
			return RecordSchemaShape{}, err
		}
		shape.AdmitsOnlyEmptyObject = empty
	}
	return shape, nil
}

// ValidatePromotableRecordSchema is the runtime promotion gate for an
// implemented declarative reverse-ETL command. It prevents a command from
// being promoted merely because a deriver discarded a provider union's fields.
func ValidatePromotableRecordSchema(raw json.RawMessage) error {
	if len(raw) == 0 {
		// Hook-owned actions can intentionally have no declarative record schema.
		return nil
	}
	shape, err := InspectRecordSchema(raw)
	if err != nil {
		return err
	}
	if shape.RootUnion != "" {
		return fmt.Errorf("record_schema root %s expands to %d arm(s), %d with named fields; declare a separate named write action for each arm", shape.RootUnion, shape.ArmCount, shape.SubstantiveArmCount)
	}
	if shape.AdmitsOnlyEmptyObject {
		return fmt.Errorf("record_schema admits only an empty object ({})")
	}
	return nil
}

// ValidateRecordSchemaField proves that field is an exact top-level property
// of one concrete write record schema. It deliberately preserves case: JSON
// Schema property names are provider-owned and a snake_case-to-camelCase
// conversion would select a different request field. It does not accept
// additionalProperties as authority for a transport mapping; a closed typed
// destination must point at a named schema property.
func ValidateRecordSchemaField(raw json.RawMessage, field string) error {
	if field == "" {
		return fmt.Errorf("record field %q must map to one top-level record property", field)
	}
	properties, _, err := recordSchemaTopLevelProperties(raw)
	if err != nil {
		return err
	}
	if _, ok := properties[field]; !ok {
		return fmt.Errorf("record field %q is not declared in record_schema", field)
	}
	return nil
}

// ValidateRecordSchemaFieldMapping proves that declaration-owned fields map
// exactly to one concrete write action and cover every required top-level
// property of its record schema.
func ValidateRecordSchemaFieldMapping(raw json.RawMessage, fields []string) error {
	properties, required, err := recordSchemaTopLevelProperties(raw)
	if err != nil {
		return err
	}
	mapped := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if field == "" {
			return fmt.Errorf("record field %q must map to one top-level record property", field)
		}
		if _, duplicate := mapped[field]; duplicate {
			return fmt.Errorf("record schema mapping duplicates field %q", field)
		}
		if _, declared := properties[field]; !declared {
			return fmt.Errorf("record field %q is not declared in record_schema", field)
		}
		mapped[field] = struct{}{}
	}
	for _, field := range required {
		if _, declared := properties[field]; !declared {
			return fmt.Errorf("record_schema requires undeclared field %q", field)
		}
		if _, present := mapped[field]; !present {
			return fmt.Errorf("record schema required field %q is not mapped", field)
		}
	}
	return nil
}

// ValidateStructuredJSONRecordField is the shared declaration gate for a
// command flag which supplies one structured JSON value to a reverse-ETL
// record. A JSON flag is deliberately narrower than a generic request body:
// it must name exactly one top-level field of a concrete write action, and
// that field must itself be an explicitly typed object or array. The runner
// calls this through Connector during runtime preflight; connectorgen calls
// the same function while validating authored CLI metadata.
func ValidateStructuredJSONRecordField(raw json.RawMessage, field string) error {
	if field == "" {
		return fmt.Errorf("structured JSON field %q must map to one top-level record property", field)
	}
	properties, _, err := recordSchemaTopLevelProperties(raw)
	if err != nil {
		return err
	}
	property, ok := properties[field]
	if !ok {
		return fmt.Errorf("record field %q is not declared in record_schema", field)
	}

	var schema struct {
		Type json.RawMessage `json:"type"`
	}
	if err := json.Unmarshal(property, &schema); err != nil {
		return fmt.Errorf("structured JSON field %q has invalid schema: %w", field, err)
	}
	var types []string
	var single string
	if err := json.Unmarshal(schema.Type, &single); err == nil {
		types = []string{single}
	} else if err := json.Unmarshal(schema.Type, &types); err != nil {
		return fmt.Errorf("structured JSON field %q has invalid schema: %w", field, err)
	}
	structured := false
	for _, typeName := range types {
		if typeName == "object" || typeName == "array" {
			structured = true
			continue
		}
		if typeName != "null" {
			return fmt.Errorf("structured JSON field %q must declare type object or array (optionally null)", field)
		}
	}
	if !structured {
		return fmt.Errorf("structured JSON field %q must declare type object or array", field)
	}
	return nil
}

func recordSchemaTopLevelProperties(raw json.RawMessage) (map[string]json.RawMessage, []string, error) {
	if err := ValidatePromotableRecordSchema(raw); err != nil {
		return nil, nil, err
	}
	if _, err := CompileSchema(raw); err != nil {
		return nil, nil, fmt.Errorf("record_schema: %w", err)
	}
	var root struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, nil, fmt.Errorf("invalid record_schema: %w", err)
	}
	return root.Properties, root.Required, nil
}

// expandRecordSchemaArms only treats a union at the current root as a branch;
// recursive calls expand nested top-level unions after the enclosing arm has
// been merged. That avoids the old failure mode of counting a union wrapper's
// own empty properties instead of the properties on its arms.
func expandRecordSchemaArms(root map[string]json.RawMessage) (string, []map[string]json.RawMessage, error) {
	for _, keyword := range []string{"oneOf", "anyOf"} {
		rawArms, ok := root[keyword]
		if !ok {
			continue
		}
		var arms []map[string]json.RawMessage
		if err := json.Unmarshal(rawArms, &arms); err != nil {
			return "", nil, fmt.Errorf("record_schema %s must be an array of schema objects: %w", keyword, err)
		}
		if len(arms) == 0 {
			return "", nil, fmt.Errorf("record_schema %s must contain at least one arm", keyword)
		}
		base := cloneRecordSchemaMap(root)
		delete(base, keyword)
		var expanded []map[string]json.RawMessage
		for _, arm := range arms {
			if arm == nil {
				return "", nil, fmt.Errorf("record_schema %s arm must be an object", keyword)
			}
			_, nested, err := expandRecordSchemaArms(mergeRecordSchemaMaps(base, arm))
			if err != nil {
				return "", nil, err
			}
			expanded = append(expanded, nested...)
		}
		return keyword, expanded, nil
	}
	return "", []map[string]json.RawMessage{root}, nil
}

func cloneRecordSchemaMap(in map[string]json.RawMessage) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(in))
	for key, value := range in {
		out[key] = append(json.RawMessage(nil), value...)
	}
	return out
}

func mergeRecordSchemaMaps(base, arm map[string]json.RawMessage) map[string]json.RawMessage {
	out := cloneRecordSchemaMap(base)
	for key, value := range arm {
		if key == "properties" {
			if merged, ok := mergeRecordSchemaProperties(out[key], value); ok {
				out[key] = merged
				continue
			}
		}
		if key == "required" {
			if merged, ok := mergeRecordSchemaRequired(out[key], value); ok {
				out[key] = merged
				continue
			}
		}
		out[key] = append(json.RawMessage(nil), value...)
	}
	return out
}

// mergeRecordSchemaProperties preserves fields declared alongside a root union
// as well as fields declared by the selected arm. A straight map overwrite
// would reproduce the measuring bug for schemas that put shared fields on the
// wrapper and arm-specific fields inside oneOf/anyOf.
func mergeRecordSchemaProperties(base, arm json.RawMessage) (json.RawMessage, bool) {
	if len(base) == 0 {
		return nil, false
	}
	var baseProperties, armProperties map[string]json.RawMessage
	if err := json.Unmarshal(base, &baseProperties); err != nil {
		return nil, false
	}
	if err := json.Unmarshal(arm, &armProperties); err != nil {
		return nil, false
	}
	for name, schema := range armProperties {
		baseProperties[name] = schema
	}
	merged, err := json.Marshal(baseProperties)
	if err != nil {
		return nil, false
	}
	return merged, true
}

// mergeRecordSchemaRequired preserves wrapper-required fields in each arm.
// It keeps their source order stable to make generated diagnostics predictable.
func mergeRecordSchemaRequired(base, arm json.RawMessage) (json.RawMessage, bool) {
	if len(base) == 0 {
		return nil, false
	}
	var baseRequired, armRequired []string
	if err := json.Unmarshal(base, &baseRequired); err != nil {
		return nil, false
	}
	if err := json.Unmarshal(arm, &armRequired); err != nil {
		return nil, false
	}
	// Required-field arrays originate in provider schemas. Let Go grow these
	// collections as needed rather than combining untrusted lengths for an
	// allocation capacity.
	seen := make(map[string]struct{})
	mergedRequired := make([]string, 0)
	for _, required := range [][]string{baseRequired, armRequired} {
		for _, name := range required {
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			mergedRequired = append(mergedRequired, name)
		}
	}
	merged, err := json.Marshal(mergedRequired)
	if err != nil {
		return nil, false
	}
	return merged, true
}

func recordSchemaArmHasNamedFields(arm map[string]json.RawMessage) bool {
	rawProperties, ok := arm["properties"]
	if !ok {
		return false
	}
	var properties map[string]json.RawMessage
	return json.Unmarshal(rawProperties, &properties) == nil && len(properties) > 0
}

func recordSchemaArmAdmitsOnlyEmptyObject(arm map[string]json.RawMessage) (bool, error) {
	if !recordSchemaIsOnlyObjectType(arm["type"]) {
		return false, nil
	}
	if recordSchemaArmHasNamedFields(arm) {
		return false, nil
	}
	if rawRequired, ok := arm["required"]; ok {
		var required []string
		if err := json.Unmarshal(rawRequired, &required); err != nil {
			return false, fmt.Errorf("record_schema required must be an array of strings: %w", err)
		}
		if len(required) > 0 {
			return false, nil
		}
	}
	rawAdditional, ok := arm["additionalProperties"]
	if !ok {
		return false, nil
	}
	var additionalProperties bool
	if err := json.Unmarshal(rawAdditional, &additionalProperties); err != nil {
		return false, fmt.Errorf("record_schema additionalProperties must be boolean: %w", err)
	}
	return !additionalProperties, nil
}

func recordSchemaIsOnlyObjectType(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return single == "object"
	}
	var types []string
	if err := json.Unmarshal(raw, &types); err != nil {
		return false
	}
	return len(types) == 1 && types[0] == "object"
}
