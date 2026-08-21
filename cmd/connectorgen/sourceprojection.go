package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	sourceProjectionDefaultStringBytes      = 8 << 10
	sourceProjectionDefaultArrayItems       = 256
	sourceProjectionDefaultObjectProperties = 256
	// JSON-valued command flags carry a complete named field through the
	// declaration-owned body path. They are not an unbounded replacement for a
	// request body, so keep their encoded input explicitly bounded.
	sourceProjectionDefaultJSONBytes = 1 << 20
)

var (
	sourceProjectionTemplateRE       = regexp.MustCompile(`\{\{\s*(?:config|record)\.([-A-Za-z0-9_]+)\s*\}\}`)
	sourceProjectionRecordTemplateRE = regexp.MustCompile(`\{\{\s*record\.([-A-Za-z0-9_]+)\s*\}\}`)
	sourceProjectionConfigTemplateRE = regexp.MustCompile(`\{\{\s*config\.([-A-Za-z0-9_]+)\s*\}\}`)
	sourceProjectionPathVariableRE   = regexp.MustCompile(`\{([-A-Za-z0-9_]+)\}`)
)

type sourceProjectionStats struct {
	Writes  int
	CLI     int
	Missing int
}

func (s sourceProjectionStats) Changed() bool { return s.Writes+s.CLI+s.Missing > 0 }

type sourceActionContract struct {
	Fields           map[string]any
	BareStringFields map[string]bool
	Required         map[string]bool
	PathFields       []string
	Query            []sourceParameterDescriptor
	BodyFields       []string
	BodyType         string
	Binary           bool
}

// projectSourceDescriptorToBundle carries source-owned request fields into the
// broad executable action and its CLI command. It updates one action per
// method/path union: semantic aliases remain narrow, while the broad action
// proves the provider contract is reachable without double-counting aliases.
func projectSourceDescriptorToBundle(bundleDir string, result sourceImportResult, check bool) (sourceProjectionStats, error) {
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writesRaw, err := os.ReadFile(writesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return sourceProjectionStats{}, nil
		}
		return sourceProjectionStats{}, err
	}
	cliRaw, err := os.ReadFile(cliPath)
	if err != nil {
		if os.IsNotExist(err) {
			return sourceProjectionStats{}, nil
		}
		return sourceProjectionStats{}, err
	}
	var writes, cli orderedJSON
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		return sourceProjectionStats{}, fmt.Errorf("writes.json: %w", err)
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		return sourceProjectionStats{}, fmt.Errorf("cli_surface.json: %w", err)
	}

	actionsByEndpoint := map[string][]*orderedObject{}
	for _, raw := range arrayField(writes.root, "actions") {
		action, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		key := sourceProjectionEndpointKey(stringField(action, "method"), sourceProjectionPath(stringField(action, "path")))
		actionsByEndpoint[key] = append(actionsByEndpoint[key], action)
	}
	commandsByWrite := map[string][]*orderedObject{}
	for _, raw := range arrayField(cli.root, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "write") == "" {
			continue
		}
		write := stringField(command, "write")
		commandsByWrite[write] = append(commandsByWrite[write], command)
	}

	stats := sourceProjectionStats{}
	for _, operation := range result.Operations {
		if operation.Protocol == "graphql" || !sourceProjectionMutationMethod(operation.Method) {
			continue
		}
		candidates := actionsByEndpoint[sourceProjectionEndpointKey(operation.Method, operation.Path)]
		if len(candidates) == 0 {
			if sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
				continue
			}
			stats.Missing++
			continue
		}
		actions := []*orderedObject{sourceProjectionBroadAction(candidates)}
		if sourceProjectionRequiresConcreteVariants(operation) {
			actions = candidates
		}
		for _, action := range actions {
			var contract sourceActionContract
			var err error
			if sourceProjectionRequiresConcreteVariants(operation) {
				contract, err = sourceContractForConcreteVariant(operation, action)
			} else {
				contract, err = sourceContractForAction(operation, action)
			}
			sealedExistingAction := false
			if err != nil {
				// The imported descriptor retains the typed gap. Generation never
				// invents a lossy record shape for a contract it cannot express. A
				// pre-existing closed action can still be an executable declared
				// variant, including when its CLI command has not been generated yet.
				if sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
					sealed, sealErr := sourceProjectionSealExistingActionRecordSchema(action)
					sealedExistingAction = sealed
					if existing, existingErr := sourceContractForExistingClosedAction(operation, action); sealErr == nil && existingErr == nil {
						contract = existing
						err = nil
					} else if sealErr != nil {
						err = sealErr
					} else {
						err = existingErr
					}
					if err != nil {
						continue
					}
				}
				if err != nil {
					stats.Missing++
					continue
				}
			}
			if sourceProjectAction(action, contract) || sealedExistingAction {
				stats.Writes++
			}
			createdCommand := false
			if len(commandsByWrite[stringField(action, "name")]) == 0 {
				command := sourceProjectionNewCommand(operation, action)
				commands := append(arrayField(cli.root, "commands"), command)
				cli.root.set("commands", commands)
				commandsByWrite[stringField(action, "name")] = []*orderedObject{command}
				createdCommand = true
			}
			projectedCommand := false
			for _, command := range commandsByWrite[stringField(action, "name")] {
				commandChanged := sourceProjectCommand(command, contract)
				if sourceProjectionRefreshGeneratedCommandMetadata(command, operation, action) {
					commandChanged = true
				}
				if commandChanged {
					stats.CLI++
					projectedCommand = true
				}
			}
			if createdCommand && !projectedCommand {
				stats.CLI++
			}
		}
	}
	if stats.Missing > 0 {
		return stats, fmt.Errorf("%d source operation(s) have no complete executable action", stats.Missing)
	}

	if check || !stats.Changed() {
		return stats, nil
	}
	if stats.Writes > 0 {
		if err := writeBundleJSON(writesPath, writes, writesRaw); err != nil {
			return stats, err
		}
	}
	if stats.CLI > 0 {
		if err := writeBundleJSON(cliPath, cli, cliRaw); err != nil {
			return stats, err
		}
	}
	return stats, nil
}

// sourceProjectionSealExistingActionRecordSchema closes an action root only
// when it already has a declaration-owned object schema and merely omitted the
// JSON Schema default. This lets its named fields remain reachable without
// turning an incomplete provider body into a generic object escape hatch.
func sourceProjectionSealExistingActionRecordSchema(action *orderedObject) (bool, error) {
	rawSchema, exists := action.get("record_schema")
	if !exists {
		return false, fmt.Errorf("action has no record schema")
	}
	schema, ok := rawSchema.(*orderedObject)
	if !ok || stringField(schema, "type") != "object" {
		return false, fmt.Errorf("action record schema is not an object")
	}
	if additional, exists := schema.get("additionalProperties"); exists {
		if closed, ok := additional.(bool); !ok || closed {
			return false, fmt.Errorf("action record schema is not closed")
		}
		return false, nil
	}
	schema.set("additionalProperties", false)
	return true, nil
}

func sourceProjectionMutationMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func sourceProjectionEndpointKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func sourceProjectionPath(path string) string {
	return sourceProjectionTemplateRE.ReplaceAllString(path, `{$1}`)
}

func sourceProjectionHasBlockingGap(gaps []sourceContractGap) bool {
	for _, gap := range gaps {
		if gap.Foundation == "cli-operation-route-override-foundation-r1" && strings.Contains(gap.Reason, "fixed") {
			continue
		}
		return true
	}
	return false
}

func sourceProjectionBroadAction(actions []*orderedObject) *orderedObject {
	sort.SliceStable(actions, func(i, j int) bool {
		left, right := sourceProjectionActionPropertyCount(actions[i]), sourceProjectionActionPropertyCount(actions[j])
		if left != right {
			return left > right
		}
		return stringField(actions[i], "name") < stringField(actions[j], "name")
	})
	return actions[0]
}

// sourceProjectionRequiresConcreteVariants identifies a provider body whose
// root object has explicit oneOf/anyOf arms. A single broad projection would
// erase the arm-required fields and turn several legal actions into an empty
// request. These operations are represented only by their existing named,
// closed action variants.
func sourceProjectionRequiresConcreteVariants(operation sourceOperationDescriptor) bool {
	if !sourceProjectionHasBlockingGap(operation.Runtime.Gaps) || operation.Request.Body == nil {
		return false
	}
	body, ok := operation.Request.Body.Schema.(map[string]any)
	if !ok || sourceSchemaType(body) != "object" {
		return false
	}
	_, oneOf := body["oneOf"]
	_, anyOf := body["anyOf"]
	return oneOf || anyOf
}

func sourceProjectionActionPropertyCount(action *orderedObject) int {
	raw, _ := action.get("record_schema")
	schema, _ := raw.(*orderedObject)
	propsRaw, _ := schema.get("properties")
	props, _ := propsRaw.(*orderedObject)
	if props == nil {
		return 0
	}
	return len(props.keys)
}

func sourceContractForAction(operation sourceOperationDescriptor, action *orderedObject) (sourceActionContract, error) {
	contract := sourceActionContract{Fields: map[string]any{}, BareStringFields: map[string]bool{}, Required: map[string]bool{}}
	// A retained source gap does not authorize a raw body. It does permit a
	// source-owned named field to use the bounded JSON flag path when its exact
	// nested shape is not expressible by the closed record-schema subset.
	allowBoundedNamedJSON := sourceProjectionHasBlockingGap(operation.Runtime.Gaps)
	pathFields := map[string]bool{}
	for _, match := range sourceProjectionRecordTemplateRE.FindAllStringSubmatch(stringField(action, "path"), -1) {
		pathFields[match[1]] = true
	}
	for _, parameter := range operation.Request.Path {
		if !pathFields[parameter.Name] {
			continue
		}
		converted, err := sourceProjectionSchema(parameter.Schema)
		if err != nil && allowBoundedNamedJSON {
			converted, err = sourceProjectionBoundedNamedJSONSchema(parameter.Schema)
		}
		if err != nil {
			return sourceActionContract{}, err
		}
		contract.setSourceField(parameter.Name, parameter.Schema, converted)
		contract.Required[parameter.Name] = true
		contract.PathFields = append(contract.PathFields, parameter.Name)
	}
	for _, parameter := range operation.Request.Query {
		converted, err := sourceProjectionSchema(parameter.Schema)
		if err != nil && allowBoundedNamedJSON {
			converted, err = sourceProjectionBoundedNamedJSONSchema(parameter.Schema)
		}
		if err != nil {
			return sourceActionContract{}, err
		}
		contract.setSourceField(parameter.Name, parameter.Schema, converted)
		// One record field can legitimately bind more than one declared input
		// location (for example a path `name` plus an optional body `name`).
		// Requiredness is the union of those declarations: an optional second
		// occurrence must never downgrade the structural path requirement.
		contract.Required[parameter.Name] = contract.Required[parameter.Name] || parameter.Required
		contract.Query = append(contract.Query, parameter)
	}
	if operation.Request.Body != nil {
		media := sourceNormalizedMediaType(operation.Request.MediaType)
		if media == "application/octet-stream" {
			contract.Binary = true
			if stringField(action, "body_type") != "binary_upload" {
				return sourceActionContract{}, fmt.Errorf("binary source requires binary_upload")
			}
			rawUpload, _ := action.get("binary_upload")
			upload, _ := rawUpload.(*orderedObject)
			field := stringField(upload, "source_field")
			if field == "" {
				return sourceActionContract{}, fmt.Errorf("binary upload has no source field")
			}
			contract.Fields[field] = map[string]any{"type": "string", "minLength": 1, "maxLength": 4096}
			contract.Required[field] = true
			return contract, nil
		}
		if !sourceJSONMediaType(media) {
			return sourceActionContract{}, fmt.Errorf("request media %q is not projectable", media)
		}
		body, ok := operation.Request.Body.Schema.(map[string]any)
		if !ok || sourceSchemaType(body) != "object" {
			return sourceActionContract{}, fmt.Errorf("request body is not an object")
		}
		if _, oneOf := body["oneOf"]; oneOf {
			return sourceActionContract{}, fmt.Errorf("request body oneOf requires a concrete closed action variant")
		}
		if _, anyOf := body["anyOf"]; anyOf {
			return sourceActionContract{}, fmt.Errorf("request body anyOf requires a concrete closed action variant")
		}
		if _, err := sourceProjectionSchema(body); err != nil && !allowBoundedNamedJSON {
			return sourceActionContract{}, err
		}
		properties, _ := body["properties"].(map[string]any)
		required := sourceSchemaRequired(body)
		for _, name := range sortedSourceMapKeys(properties) {
			converted, err := sourceProjectionSchema(properties[name])
			if err != nil && allowBoundedNamedJSON {
				converted, err = sourceProjectionBoundedNamedJSONSchema(properties[name])
			}
			if err != nil {
				return sourceActionContract{}, err
			}
			contract.setSourceField(name, properties[name], converted)
			// Preserve a required path/query occurrence when the provider also
			// exposes the same field as an optional body member.
			contract.Required[name] = contract.Required[name] || required[name]
			contract.BodyFields = append(contract.BodyFields, name)
		}
	}
	if err := sourceProjectionRetainDeclaredHookFields(action, &contract); err != nil {
		return sourceActionContract{}, err
	}
	sort.Strings(contract.PathFields)
	sort.Slice(contract.Query, func(i, j int) bool { return contract.Query[i].Name < contract.Query[j].Name })
	sort.Strings(contract.BodyFields)
	return contract, nil
}

// sourceProjectionRetainDeclaredHookFields preserves an explicit, closed list
// of fields that a declaration-owned compound hook consumes after the
// provider operation's own request body. A source descriptor owns the direct
// provider fields; it cannot silently erase a different declared follow-up
// route. The fields remain record-schema and command fields, but are never
// appended to the direct operation's body_fields, so this is neither a raw
// body channel nor connector-name-specific generation.
func sourceProjectionRetainDeclaredHookFields(action *orderedObject, contract *sourceActionContract) error {
	rawNames, declared := action.get("hook_fields")
	if !declared {
		return nil
	}
	if strings.TrimSpace(stringField(action, "hook")) == "" {
		return fmt.Errorf("hook_fields requires a declaration-owned hook")
	}
	rawSchema, exists := action.get("record_schema")
	if !exists {
		return fmt.Errorf("hook_fields requires an action record schema")
	}
	encoded, err := marshalNoEscapeHTML(rawSchema)
	if err != nil {
		return fmt.Errorf("encode hook record schema: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var schema map[string]any
	if err := decoder.Decode(&schema); err != nil {
		return fmt.Errorf("decode hook record schema: %w", err)
	}
	if sourceSchemaType(schema) != "object" || schema["additionalProperties"] != false {
		return fmt.Errorf("hook_fields requires a closed action record schema")
	}
	properties, _ := schema["properties"].(map[string]any)
	required := sourceSchemaRequired(schema)
	seen := map[string]bool{}
	for _, rawName := range sourceAnySlice(rawNames) {
		name, ok := rawName.(string)
		if !ok || strings.TrimSpace(name) == "" || strings.TrimSpace(name) != name {
			return fmt.Errorf("hook_fields contains an invalid field name")
		}
		if seen[name] {
			return fmt.Errorf("hook_fields duplicates %q", name)
		}
		seen[name] = true
		if _, overlapsProviderBody := contract.Fields[name]; overlapsProviderBody {
			return fmt.Errorf("hook field %q overlaps the provider operation contract", name)
		}
		rawProperty, exists := properties[name]
		if !exists {
			return fmt.Errorf("hook field %q is missing from the closed action record schema", name)
		}
		converted, err := sourceProjectionSchema(rawProperty)
		if err != nil {
			return fmt.Errorf("project hook field %q: %w", name, err)
		}
		contract.setSourceField(name, properties[name], converted)
		contract.Required[name] = required[name]
	}
	return nil
}

// sourceContractForConcreteVariant projects one named declaration-owned action
// for a root object union. Its fields come only from the provider schema; its
// arm is selected only when the existing action's declared identity or fields
// distinguish one arm. An ambiguous action fails closed and is left for the
// retained closed-action fallback rather than becoming a generic body route.
func sourceContractForConcreteVariant(operation sourceOperationDescriptor, action *orderedObject) (sourceActionContract, error) {
	if operation.Request.Body == nil || !sourceJSONMediaType(sourceNormalizedMediaType(operation.Request.MediaType)) {
		return sourceActionContract{}, fmt.Errorf("concrete variant has no JSON request body")
	}
	body, ok := operation.Request.Body.Schema.(map[string]any)
	if !ok || sourceSchemaType(body) != "object" {
		return sourceActionContract{}, fmt.Errorf("concrete variant request body is not an object")
	}
	arms, err := sourceProjectionRootUnionArms(body)
	if err != nil {
		return sourceActionContract{}, err
	}
	variant, err := sourceProjectionVariant(action, arms)
	if err != nil {
		return sourceActionContract{}, err
	}

	withoutBody := operation
	withoutBody.Request.Body = nil
	contract, err := sourceContractForAction(withoutBody, action)
	if err != nil {
		return sourceActionContract{}, err
	}
	properties := map[string]any{}
	for name, schema := range sourceProjectionObjectProperties(body) {
		properties[name] = schema
	}
	for name, schema := range sourceProjectionObjectProperties(variant) {
		properties[name] = schema
	}
	rootRequired := sourceSchemaRequired(body)
	variantRequired := sourceSchemaRequired(variant)
	for name := range variantRequired {
		if _, declared := properties[name]; !declared {
			return sourceActionContract{}, fmt.Errorf("concrete variant %q requires undeclared field %q", stringField(action, "name"), name)
		}
	}
	for _, name := range sortedSourceMapKeys(properties) {
		converted, err := sourceProjectionSchema(properties[name])
		if err != nil {
			converted, err = sourceProjectionBoundedNamedJSONSchema(properties[name])
		}
		if err != nil {
			return sourceActionContract{}, err
		}
		contract.Fields[name] = converted
		contract.Required[name] = contract.Required[name] || rootRequired[name] || variantRequired[name]
		contract.BodyFields = append(contract.BodyFields, name)
	}
	sort.Strings(contract.BodyFields)
	return contract, nil
}

func sourceProjectionRootUnionArms(body map[string]any) ([]map[string]any, error) {
	for _, keyword := range []string{"oneOf", "anyOf"} {
		raw, exists := body[keyword]
		if !exists {
			continue
		}
		values, ok := raw.([]any)
		if !ok || len(values) == 0 {
			return nil, fmt.Errorf("request body %s must be a non-empty array", keyword)
		}
		arms := make([]map[string]any, 0, len(values))
		for _, value := range values {
			arm, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("request body %s arm is not an object", keyword)
			}
			arms = append(arms, arm)
		}
		return arms, nil
	}
	return nil, fmt.Errorf("request body has no root union")
}

func sourceProjectionVariant(action *orderedObject, arms []map[string]any) (map[string]any, error) {
	identity := sourceProjectionIdentifierTokens(stringField(action, "name"))
	declared := sourceProjectionActionRequiredFields(action)
	bestIdentity, bestDeclared, bestIndex, ties := 0, 0, -1, 0
	for index, arm := range arms {
		required := sourceSchemaRequired(arm)
		if len(required) == 0 {
			continue
		}
		identityScore, declaredScore := 0, 0
		for field := range required {
			if declared[field] {
				declaredScore++
			}
			for token := range sourceProjectionIdentifierTokens(field) {
				if identity[token] {
					identityScore++
				}
			}
		}
		if identityScore > bestIdentity || (identityScore == bestIdentity && declaredScore > bestDeclared) {
			bestIdentity, bestDeclared, bestIndex, ties = identityScore, declaredScore, index, 1
		} else if identityScore == bestIdentity && declaredScore == bestDeclared && (identityScore > 0 || declaredScore > 0) {
			ties++
		}
	}
	if bestIndex < 0 || (bestIdentity == 0 && bestDeclared == 0) || ties != 1 {
		return nil, fmt.Errorf("action %q does not identify exactly one root union arm", stringField(action, "name"))
	}
	return arms[bestIndex], nil
}

func sourceProjectionObjectProperties(schema map[string]any) map[string]any {
	properties, _ := schema["properties"].(map[string]any)
	return properties
}

func sourceProjectionActionRequiredFields(action *orderedObject) map[string]bool {
	fields := map[string]bool{}
	rawSchema, _ := action.get("record_schema")
	schema, _ := rawSchema.(*orderedObject)
	if schema == nil {
		return fields
	}
	for _, raw := range arrayField(schema, "required") {
		if name, ok := raw.(string); ok {
			fields[name] = true
		}
	}
	return fields
}

func sourceProjectionIdentifierTokens(value string) map[string]bool {
	tokens := map[string]bool{}
	for _, token := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return r < 'a' || r > 'z'
	}) {
		if token != "" {
			tokens[token] = true
		}
	}
	return tokens
}

// sourceContractForExistingClosedAction derives flags only from a pre-existing
// action's closed record schema. It is used when the provider descriptor keeps
// a typed source gap: the action is not expanded from that incomplete source,
// but callers can still reach every field the action already declares.
func sourceContractForExistingClosedAction(operation sourceOperationDescriptor, action *orderedObject) (sourceActionContract, error) {
	rawSchema, exists := action.get("record_schema")
	if !exists {
		return sourceActionContract{}, fmt.Errorf("action has no record schema")
	}
	encoded, err := marshalNoEscapeHTML(rawSchema)
	if err != nil {
		return sourceActionContract{}, fmt.Errorf("encode action record schema: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var schema map[string]any
	if err := decoder.Decode(&schema); err != nil {
		return sourceActionContract{}, fmt.Errorf("decode action record schema: %w", err)
	}
	if sourceSchemaType(schema) != "object" || schema["additionalProperties"] != false {
		return sourceActionContract{}, fmt.Errorf("action record schema is not closed")
	}
	rawProperties, _ := schema["properties"].(map[string]any)
	properties := make(map[string]any, len(rawProperties))
	for _, name := range sortedSourceMapKeys(rawProperties) {
		converted, err := sourceProjectionSchema(rawProperties[name])
		if err != nil {
			// A retained closed action can contain a source-owned field whose
			// nested provider shape is unavailable (for example an inconsistent
			// oneOf arm). Preserve only that named field as bounded JSON; never
			// replace the closed root with a generic body object.
			converted, err = sourceProjectionBoundedNamedJSONSchema(rawProperties[name])
		}
		if err != nil {
			return sourceActionContract{}, fmt.Errorf("project action record property %q: %w", name, err)
		}
		properties[name] = converted
	}
	contract := sourceActionContract{Fields: properties, Required: sourceSchemaRequired(schema)}
	contract.BodyType = stringField(action, "body_type")
	for _, raw := range arrayField(action, "body_fields") {
		if name, ok := raw.(string); ok {
			contract.BodyFields = append(contract.BodyFields, name)
		}
	}
	pathFields := map[string]bool{}
	for _, match := range sourceProjectionRecordTemplateRE.FindAllStringSubmatch(stringField(action, "path"), -1) {
		if _, exists := properties[match[1]]; !exists {
			return sourceActionContract{}, fmt.Errorf("path field %q is missing from action record schema", match[1])
		}
		pathFields[match[1]] = true
		contract.PathFields = append(contract.PathFields, match[1])
	}
	for _, parameter := range operation.Request.Path {
		if pathFields[parameter.Name] {
			// A provider path parameter is structurally required. Preserve that
			// requirement even when the retained source gap prevents us from
			// replacing the rest of the action's closed record shape.
			contract.Required[parameter.Name] = true
		}
	}
	// A retained closed action is a bounded declaration-owned alternative for a
	// source body with a typed gap; it is not a reason to send the provider an
	// empty request. A legacy action can have recorded its named body fields in
	// the closed record schema while still saying body_type:none. The immutable
	// source proves this route has a JSON request body, so promote only those
	// already-declared, non-path fields to the named JSON body. No open object
	// or caller-selected body channel is introduced.
	if operation.Request.Body != nil && sourceJSONMediaType(sourceNormalizedMediaType(operation.Request.MediaType)) {
		bodyFields := make([]string, 0, len(properties))
		for _, name := range sortedSourceMapKeys(properties) {
			if !pathFields[name] {
				bodyFields = append(bodyFields, name)
			}
		}
		if len(bodyFields) == 0 {
			return sourceActionContract{}, fmt.Errorf("closed action has no declared JSON body field")
		}
		contract.BodyType = "json"
		contract.BodyFields = bodyFields
	}
	sort.Strings(contract.PathFields)
	sort.Strings(contract.BodyFields)
	return contract, nil
}

func sourceSchemaType(schema map[string]any) string {
	typeName, _ := schema["type"].(string)
	return typeName
}

func sourceSchemaRequired(schema map[string]any) map[string]bool {
	result := map[string]bool{}
	for _, raw := range sourceAnySlice(schema["required"]) {
		if name, ok := raw.(string); ok {
			result[name] = true
		}
	}
	return result
}

func sourceAnySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []string:
		out := make([]any, len(typed))
		for index := range typed {
			out[index] = typed[index]
		}
		return out
	default:
		return nil
	}
}

func sourceProjectionSchema(raw any) (any, error) {
	schema, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("source schema is not an object")
	}
	if _, exists := schema["oneOf"]; exists {
		return nil, fmt.Errorf("oneOf requires a typed gap")
	}
	if _, exists := schema["anyOf"]; exists {
		return nil, fmt.Errorf("anyOf requires a typed gap")
	}
	types := []string{}
	switch typed := schema["type"].(type) {
	case string:
		types = append(types, typed)
	case []any:
		for _, value := range typed {
			name, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("source schema has a non-string type")
			}
			types = append(types, name)
		}
	default:
		return nil, fmt.Errorf("source schema has no type")
	}
	if nullable, _ := schema["nullable"].(bool); nullable && !sourceProjectionContainsString(types, "null") {
		types = append(types, "null")
	}
	if len(types) == 0 {
		return nil, fmt.Errorf("source schema has no type")
	}
	primary := types[0]
	out := map[string]any{}
	if len(types) == 1 {
		out["type"] = primary
	} else {
		values := make([]any, len(types))
		for index := range types {
			values[index] = types[index]
		}
		out["type"] = values
	}
	if enum, ok := schema["enum"].([]any); ok && len(enum) > 0 {
		out["enum"] = enum
	}
	for _, key := range []string{"pattern", "minLength", "maxLength", "minItems", "maxItems", "minProperties", "maxProperties"} {
		if value, exists := schema[key]; exists {
			out[key] = value
		}
	}
	switch primary {
	case "string":
		if _, bounded := out["maxLength"]; !bounded {
			out["maxLength"] = json.Number(fmt.Sprintf("%d", sourceProjectionDefaultStringBytes))
		}
	case "array":
		items, exists := schema["items"]
		if !exists {
			return nil, fmt.Errorf("array schema has no items")
		}
		converted, err := sourceProjectionSchema(items)
		if err != nil {
			return nil, err
		}
		out["items"] = converted
		if _, bounded := out["maxItems"]; !bounded {
			out["maxItems"] = json.Number(fmt.Sprintf("%d", sourceProjectionDefaultArrayItems))
		}
	case "object":
		if additional, exists := schema["additionalProperties"]; exists && additional != false {
			return nil, fmt.Errorf("dynamic additionalProperties requires a typed gap")
		}
		properties, _ := schema["properties"].(map[string]any)
		convertedProperties := map[string]any{}
		for _, name := range sortedSourceMapKeys(properties) {
			converted, err := sourceProjectionSchema(properties[name])
			if err != nil {
				return nil, fmt.Errorf("property %q: %w", name, err)
			}
			convertedProperties[name] = converted
		}
		out["properties"] = convertedProperties
		out["additionalProperties"] = false
		if required := sourceAnySlice(schema["required"]); len(required) > 0 {
			out["required"] = required
		}
	}
	return out, nil
}

// sourceProjectionBoundedNamedJSONSchema retains a source-owned value whose
// recursive shape cannot yet be represented by the closed record-schema
// subset. The enclosing action root remains closed and the generated command
// can map only this one named field. Commandrunner applies the explicit JSON
// byte cap before the value reaches the reverse-ETL plan, so this is not a
// generic request-body escape hatch.
func sourceProjectionBoundedNamedJSONSchema(raw any) (any, error) {
	schema, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("source schema is not an object")
	}
	types, err := sourceProjectionJSONTypes(schema)
	if err != nil {
		return nil, err
	}
	if len(types) == 0 {
		return nil, fmt.Errorf("source schema has no JSON type")
	}
	out := map[string]any{}
	if len(types) == 1 {
		out["type"] = types[0]
	} else {
		values := make([]any, len(types))
		for index := range types {
			values[index] = types[index]
		}
		out["type"] = values
	}
	if sourceProjectionContainsString(types, "string") {
		out["maxLength"] = sourceProjectionBoundedCardinality(schema, "maxLength", sourceProjectionDefaultJSONBytes)
	}
	if sourceProjectionContainsString(types, "array") {
		out["maxItems"] = sourceProjectionBoundedCardinality(schema, "maxItems", sourceProjectionDefaultArrayItems)
		// `{}` retains all provider-owned element shapes. The field remains
		// named, its cardinality is finite, and commandrunner rejects an
		// over-limit JSON value before decoding; this is not a body escape.
		out["items"] = map[string]any{}
	}
	if sourceProjectionContainsString(types, "object") {
		out["maxProperties"] = sourceProjectionBoundedCardinality(schema, "maxProperties", sourceProjectionDefaultObjectProperties)
		// This fallback represents a dynamic or ambiguous nested provider
		// object. The record root stays closed; all retained provider keys live
		// only in this named field and receive finite key and byte bounds.
		out["additionalProperties"] = true
	}
	return out, nil
}

func sourceProjectionBoundedCardinality(schema map[string]any, key string, fallback int) json.Number {
	if value, exists := schema[key]; exists {
		switch typed := value.(type) {
		case json.Number:
			if integer, err := typed.Int64(); err == nil && integer > 0 {
				return json.Number(strconv.FormatInt(integer, 10))
			}
		case int:
			if typed > 0 {
				return json.Number(strconv.Itoa(typed))
			}
		case int64:
			if typed > 0 {
				return json.Number(strconv.FormatInt(typed, 10))
			}
		case float64:
			if typed > 0 && typed == math.Trunc(typed) && typed <= math.MaxInt64 {
				return json.Number(strconv.FormatInt(int64(typed), 10))
			}
		}
	}
	return json.Number(strconv.Itoa(fallback))
}

func sourceProjectionJSONTypes(schema map[string]any) ([]string, error) {
	seen := map[string]bool{}
	var add func(map[string]any) error
	add = func(node map[string]any) error {
		switch typed := node["type"].(type) {
		case string:
			seen[typed] = true
		case []any:
			for _, value := range typed {
				name, ok := value.(string)
				if !ok {
					return fmt.Errorf("source schema has a non-string type")
				}
				seen[name] = true
			}
		case nil:
		default:
			return fmt.Errorf("source schema has an invalid type")
		}
		if nullable, _ := node["nullable"].(bool); nullable {
			seen["null"] = true
		}
		for _, keyword := range []string{"oneOf", "anyOf"} {
			rawArms, exists := node[keyword]
			if !exists {
				continue
			}
			arms, ok := rawArms.([]any)
			if !ok || len(arms) == 0 {
				return fmt.Errorf("source schema %s must be a non-empty array", keyword)
			}
			for _, rawArm := range arms {
				arm, ok := rawArm.(map[string]any)
				if !ok {
					return fmt.Errorf("source schema %s arm is not an object", keyword)
				}
				if err := add(arm); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := add(schema); err != nil {
		return nil, err
	}
	types := make([]string, 0, len(seen))
	for name := range seen {
		types = append(types, name)
	}
	sort.Strings(types)
	return types, nil
}

func sourceProjectionContainsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func sourceProjectAction(action *orderedObject, contract sourceActionContract) bool {
	properties := newOrderedObject()
	for _, name := range sortedAnyMapKeys(contract.Fields) {
		properties.set(name, orderedFromAny(contract.Fields[name]))
	}
	requiredNames := []string{}
	for name, required := range contract.Required {
		if required {
			requiredNames = append(requiredNames, name)
		}
	}
	sort.Strings(requiredNames)
	required := make([]any, len(requiredNames))
	for index := range requiredNames {
		required[index] = requiredNames[index]
	}
	schema := newOrderedObject()
	schema.set("$schema", "http://json-schema.org/draft-07/schema#")
	schema.set("type", "object")
	schema.set("additionalProperties", false)
	schema.set("required", required)
	schema.set("properties", properties)
	changed := setOrderedIfDifferent(action, "record_schema", schema)
	pathFields := make([]any, len(contract.PathFields))
	for index := range contract.PathFields {
		pathFields[index] = contract.PathFields[index]
	}
	if len(pathFields) == 0 {
		changed = action.remove("path_fields") || changed
	} else {
		changed = setOrderedIfDifferent(action, "path_fields", pathFields) || changed
	}

	query := newOrderedObject()
	for _, parameter := range contract.Query {
		if parameter.Required {
			query.set(parameter.Name, "{{ record."+parameter.Name+" }}")
			continue
		}
		value := newOrderedObject()
		value.set("template", "{{ record."+parameter.Name+" }}")
		value.set("omit_when_absent", true)
		query.set(parameter.Name, value)
	}
	if len(query.keys) == 0 {
		changed = action.remove("query") || changed
	} else {
		changed = setOrderedIfDifferent(action, "query", query) || changed
	}
	if !contract.Binary {
		if contract.BodyType != "" {
			changed = setOrderedIfDifferent(action, "body_type", contract.BodyType) || changed
			if len(contract.BodyFields) > 0 {
				values := make([]any, len(contract.BodyFields))
				for index := range contract.BodyFields {
					values[index] = contract.BodyFields[index]
				}
				changed = setOrderedIfDifferent(action, "body_fields", values) || changed
			} else {
				changed = action.remove("body_fields") || changed
			}
			return changed
		}
		bodyType := "none"
		if len(contract.BodyFields) > 0 {
			bodyType = "json"
		}
		changed = setOrderedIfDifferent(action, "body_type", bodyType) || changed
		if len(contract.BodyFields) > 0 {
			values := make([]any, len(contract.BodyFields))
			for index := range contract.BodyFields {
				values[index] = contract.BodyFields[index]
			}
			changed = setOrderedIfDifferent(action, "body_fields", values) || changed
		} else {
			changed = action.remove("body_fields") || changed
		}
	}
	return changed
}

func sourceProjectCommand(command *orderedObject, contract sourceActionContract) bool {
	existing := map[string]*orderedObject{}
	var preserved []any
	for _, raw := range arrayField(command, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok {
			preserved = append(preserved, raw)
			continue
		}
		target := stringField(flag, "maps_to")
		name, record := strings.CutPrefix(target, "record.")
		if !record {
			preserved = append(preserved, flag)
			continue
		}
		existing[name] = flag
	}
	flags := append([]any{}, preserved...)
	for _, name := range sortedAnyMapKeys(contract.Fields) {
		flag := newOrderedObject()
		flag.set("name", strings.ReplaceAll(name, "_", "-"))
		if prior := existing[name]; prior != nil {
			if summary := stringField(prior, "summary"); summary != "" {
				flag.set("summary", summary)
			}
		}
		flagType := sourceProjectionFlagType(contract.Fields[name])
		flag.set("type", flagType)
		if contract.BareStringFields[name] {
			flag.set("allow_bare_string", true)
		} else {
			flag.remove("allow_bare_string")
		}
		if schema, ok := contract.Fields[name].(map[string]any); ok {
			if values, ok := schema["enum"].([]any); ok && len(values) > 0 {
				flag.set("values", orderedFromAny(values))
			}
		}
		flag.set("maps_to", "record."+name)
		if contract.Required[name] {
			flag.set("required", true)
		} else {
			flag.remove("required")
		}
		if maxBytes := sourceProjectionFlagMaxBytes(contract.Fields[name], flagType); maxBytes > 0 {
			flag.set("max_bytes", json.Number(fmt.Sprintf("%d", maxBytes)))
		} else {
			flag.remove("max_bytes")
		}
		flags = append(flags, flag)
	}
	return setOrderedIfDifferent(command, "flags", flags)
}

func sourceProjectionNewCommand(operation sourceOperationDescriptor, action *orderedObject) *orderedObject {
	command := newOrderedObject()
	command.set("path", sourceProjectionGeneratedCommandPath(operation))
	command.set("summary", strings.ToUpper(operation.Method)+" "+operation.Path)
	command.set("intent", "reverse_etl")
	command.set("availability", "implemented")
	command.set("write", stringField(action, "name"))
	if operation.Source.URL != "" {
		command.set("source_url", operation.Source.URL)
	}
	command.set("risk", stringField(action, "risk"))
	command.set("approval", sourceProjectionApproval(action))
	endpoint := newOrderedObject()
	endpoint.set("method", strings.ToUpper(operation.Method))
	endpoint.set("path", operation.Path)
	command.set("api_surface", []any{endpoint})
	return command
}

func sourceProjectionGeneratedCommandPath(operation sourceOperationDescriptor) string {
	path := strings.NewReplacer("/", " ", "_", "-").Replace(operation.SourceID)
	return "api " + path
}

// sourceProjectionRefreshGeneratedCommandMetadata refreshes only the command
// identity generated from the immutable source descriptor. Author-owned aliases
// keep their own prose, while the generated command must continue to publish
// the declaration's actual approval lifecycle after a later projection pass.
func sourceProjectionRefreshGeneratedCommandMetadata(command *orderedObject, operation sourceOperationDescriptor, action *orderedObject) bool {
	if stringField(command, "path") != sourceProjectionGeneratedCommandPath(operation) {
		return false
	}
	return setOrderedIfDifferent(command, "approval", sourceProjectionApproval(action))
}

// sourceProjectionApproval publishes the same plan lifecycle the declaration
// will enforce at execution. An unconfirmed non-DELETE write mints an approval
// token at plan time, so its preview is optional; destructive declarations
// withhold that token until preview/confirmation. New source-derived commands
// must not collapse those two distinct contracts into a blanket sentence.
func sourceProjectionApproval(action *orderedObject) string {
	if strings.EqualFold(stringField(action, "method"), "DELETE") || strings.TrimSpace(stringField(action, "confirm")) != "" {
		return "Reverse ETL writes require plan, preview, approval, execute."
	}
	if confirmation, declared := action.get("confirmation"); declared && confirmation != nil {
		return "Reverse ETL writes require plan, preview, approval, execute."
	}
	return "Reverse ETL writes require plan, approval, execute; preview is optional."
}

func sourceProjectionFlagType(schema any) string {
	object, _ := schema.(map[string]any)
	if enum, ok := object["enum"].([]any); ok && len(enum) > 0 {
		return "enum"
	}
	var types []string
	switch rawType := object["type"].(type) {
	case string:
		if rawType != "null" {
			types = append(types, rawType)
		}
	case []any:
		for _, raw := range rawType {
			name, ok := raw.(string)
			if ok && name != "null" {
				types = append(types, name)
			}
		}
	}
	if len(types) != 1 {
		return "json"
	}
	switch types[0] {
	case "integer":
		return "integer"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "string":
		return "string"
	default:
		return "json"
	}
}

// setSourceField retains the complete projected provider schema and marks a
// source-declared multi-kind field whose named JSON flag may accept its string
// arm as ordinary CLI text. The field remains a bounded, declaration-owned
// JSON value: objects and arrays still require JSON, and the record schema
// validates the final value before any provider I/O.
func (contract *sourceActionContract) setSourceField(name string, sourceSchema, projectedSchema any) {
	contract.Fields[name] = projectedSchema
	if sourceProjectionFlagType(projectedSchema) == "json" && sourceProjectionContainsStringArm(sourceSchema) {
		contract.BareStringFields[name] = true
		return
	}
	delete(contract.BareStringFields, name)
}

func sourceProjectionContainsStringArm(raw any) bool {
	schema, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	types, err := sourceProjectionJSONTypes(schema)
	if err != nil {
		return false
	}
	return sourceProjectionContainsString(types, "string")
}

func sourceProjectionSchemaMaxBytes(schema any) int64 {
	object, _ := schema.(map[string]any)
	value, exists := object["maxLength"]
	if !exists {
		return 0
	}
	switch typed := value.(type) {
	case json.Number:
		integer, _ := typed.Int64()
		return integer
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func sourceProjectionFlagMaxBytes(schema any, flagType string) int64 {
	if flagType == "string" {
		// JSON Schema maxLength counts Unicode code points. The runner's cap is
		// bytes, so reserve UTF-8's maximum width to keep every source-valid
		// non-ASCII value reachable.
		if maxRunes := sourceProjectionSchemaMaxBytes(schema); maxRunes > 0 && maxRunes <= math.MaxInt64/utf8.UTFMax {
			return maxRunes * utf8.UTFMax
		}
	}
	if flagType == "json" {
		return sourceProjectionDefaultJSONBytes
	}
	return 0
}

func sortedAnyMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func orderedFromAny(value any) any {
	raw, _ := marshalNoEscapeHTML(value)
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	converted, err := decodeOrderedValue(decoder)
	if err != nil {
		return value
	}
	return converted
}

func setOrderedIfDifferent(object *orderedObject, key string, wanted any) bool {
	current, exists := object.get(key)
	if exists && orderedSemanticEqual(current, wanted) {
		return false
	}
	object.set(key, wanted)
	return true
}

func orderedSemanticEqual(left, right any) bool {
	leftRaw, leftErr := marshalNoEscapeHTML(left)
	rightRaw, rightErr := marshalNoEscapeHTML(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	// orderedObject preserves source formatting when we write, but object-member
	// order is not a declaration change. Comparing the wire bytes here made a
	// derived flag that a later synchronizer completed look different forever
	// merely because that field had been appended after maps_to. Decode through
	// UseNumber so equality remains lossless for source numeric literals while
	// maps compare independently of their member order.
	var leftValue, rightValue any
	leftDecoder := json.NewDecoder(bytes.NewReader(leftRaw))
	leftDecoder.UseNumber()
	rightDecoder := json.NewDecoder(bytes.NewReader(rightRaw))
	rightDecoder.UseNumber()
	if leftDecoder.Decode(&leftValue) != nil || rightDecoder.Decode(&rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

// checkSourceProjection proves the checked-in descriptor is the exact lock
// projection and every executable source request is field-complete at both the
// action and CLI boundary. Operations with an explicit typed gap stay visible
// and intentionally do not masquerade as implemented coverage.
func checkSourceProjection(fsys fs.FS, bundle engine.Bundle) []Finding {
	lockPath := filepath.ToSlash(filepath.Join(bundle.Name, "sources", bundle.Name+"-operation-source-lock.json"))
	lockRaw, err := fs.ReadFile(fsys, lockPath)
	if err != nil {
		return nil
	}
	lock, err := parseSourceImportLock(lockRaw, bundle.Name)
	if err != nil {
		return []Finding{sourceProjectionFinding(bundle.Name, lockPath, err.Error())}
	}
	descriptorPath := filepath.ToSlash(filepath.Join(bundle.Name, "sources", bundle.Name+"-operation-descriptor.json"))
	descriptorRaw, err := fs.ReadFile(fsys, descriptorPath)
	if err != nil {
		return []Finding{sourceProjectionFinding(bundle.Name, descriptorPath, "canonical source descriptor is missing")}
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
		return []Finding{sourceProjectionFinding(bundle.Name, descriptorPath, "parse canonical source descriptor: "+err.Error())}
	}
	findings := validateSourceDescriptorAgainstLock(bundle.Name, descriptorPath, lock, descriptor)
	findings = append(findings, validateSourceExecutableCoverage(bundle, descriptorPath, descriptor)...)
	return findings
}

func sourceProjectionFinding(connector, file, message string) Finding {
	return Finding{Connector: connector, File: strings.TrimPrefix(file, connector+"/"), Rule: ruleSourceProjection, Message: message}
}

func validateSourceDescriptorAgainstLock(connector, file string, lock sourceImportLock, descriptor sourceImportDescriptorDocument) []Finding {
	if descriptor.SchemaVersion != 2 {
		return []Finding{sourceProjectionFinding(connector, file, fmt.Sprintf("source descriptor schema_version = %d, want 2", descriptor.SchemaVersion))}
	}
	expected := map[string]sourceImportSource{}
	for _, operation := range lock.Rest.Operations {
		identity := operation.OperationID
		if identity == "" {
			identity = operation.ID
		}
		expected[identity] = sourceImportSource{SHA256: strings.ToLower(lock.Rest.SHA256), Bytes: lock.Rest.Bytes, Location: operation.SourceLocation}
	}
	for _, field := range lock.GraphQL.QueryFields {
		expected[fmt.Sprintf("%s.graphql.query.%s", connector, field.Name)] = sourceImportSource{SHA256: strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)), Bytes: firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes)}
	}
	for _, field := range lock.GraphQL.MutationFields {
		expected[fmt.Sprintf("%s.graphql.mutation.%s", connector, field.Name)] = sourceImportSource{SHA256: strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)), Bytes: firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes)}
	}
	actual := map[string]sourceOperationDescriptor{}
	for _, operation := range descriptor.Operations {
		if _, duplicate := actual[operation.SourceID]; duplicate {
			return []Finding{sourceProjectionFinding(connector, file, "duplicate source identity "+operation.SourceID)}
		}
		actual[operation.SourceID] = operation
	}
	if len(actual) != len(expected) {
		return []Finding{sourceProjectionFinding(connector, file, fmt.Sprintf("source descriptor has %d identities, lock requires %d", len(actual), len(expected)))}
	}
	for identity, source := range expected {
		operation, ok := actual[identity]
		if !ok {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor is missing identity "+identity)}
		}
		if operation.Source.SHA256 != source.SHA256 || operation.Source.Bytes != source.Bytes || (source.Location != "" && operation.Source.Location != source.Location) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor provenance drift for "+identity)}
		}
		if operation.Runtime.MergeBlocked != (len(operation.Runtime.Gaps) > 0) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor gap state is inconsistent for "+identity)}
		}
	}
	return nil
}

func validateSourceExecutableCoverage(bundle engine.Bundle, file string, descriptor sourceImportDescriptorDocument) []Finding {
	actions := map[string][]engine.WriteAction{}
	for _, action := range bundle.Writes {
		key := sourceProjectionEndpointKey(action.Method, sourceProjectionPath(action.Path))
		actions[key] = append(actions[key], action)
	}
	commands := map[string]engine.CLICommand{}
	if bundle.CLISurface != nil {
		for _, command := range bundle.CLISurface.Commands {
			if command.Write != "" && command.Availability == "implemented" {
				commands[command.Write] = command
			}
		}
	}
	var findings []Finding
	for _, operation := range descriptor.Operations {
		if operation.Protocol == "graphql" {
			continue
		}
		if !sourceProjectionMutationMethod(operation.Method) {
			if sourceProjectionHasBlockingGap(operation.Runtime.Gaps) && sourceGapDirectOperationIsImplementedIncompletely(bundle, operation) {
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, "implemented source operation retains an unresolved source-bound gap: "+operation.SourceID))
			}
			continue
		}
		candidates := actions[sourceProjectionEndpointKey(operation.Method, operation.Path)]
		if len(candidates) == 0 {
			if sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
				continue
			}
			findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source operation has no executable action: "+operation.SourceID))
			continue
		}
		complete := false
		for _, action := range candidates {
			if sourceActionCoversOperation(action, commands[action.Name], operation) {
				complete = true
				break
			}
		}
		if !complete {
			if sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
				if !sourceGapMutationOperationIsImplementedIncompletely(bundle, candidates, commands, operation) {
					continue
				}
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, "implemented source operation retains an unresolved source-bound gap: "+operation.SourceID))
				continue
			}
			findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source request fields are missing from action/CLI union: "+operation.SourceID))
		}
	}
	return findings
}

func sourceGapMutationOperationIsImplementedIncompletely(bundle engine.Bundle, candidates []engine.WriteAction, commands map[string]engine.CLICommand, source sourceOperationDescriptor) bool {
	for _, action := range candidates {
		command := commands[action.Name]
		mapped := map[string]bool{}
		for _, flag := range command.Flags {
			if field, ok := strings.CutPrefix(flag.MapsTo, "record."); ok {
				mapped[field] = true
			}
		}
		covered := map[string]bool{}
		for field := range sourceProjectionDeclaredConfigPathFields(bundle.Spec, action.Path) {
			covered[field] = true
		}
		pathFields := map[string]bool{}
		for _, field := range action.PathFields {
			pathFields[field] = true
			covered["path."+field] = mapped[field]
		}
		queryFields := map[string]bool{}
		for field := range action.Query {
			queryFields[field] = true
			covered["query."+field] = mapped[field]
		}
		bodyFields := map[string]bool{}
		for _, field := range action.BodyFields {
			bodyFields[field] = true
		}
		var schema map[string]any
		decoder := json.NewDecoder(bytes.NewReader(action.RecordSchema))
		decoder.UseNumber()
		if decoder.Decode(&schema) != nil {
			continue
		}
		properties, _ := schema["properties"].(map[string]any)
		for field := range properties {
			if pathFields[field] || queryFields[field] {
				continue
			}
			if len(bodyFields) == 0 || bodyFields[field] {
				covered["body."+field] = mapped[field]
			}
		}
		if sourceCallerFieldsCovered(source, covered) {
			return false
		}
	}
	return true
}

func sourceCallerFieldsCovered(source sourceOperationDescriptor, covered map[string]bool) bool {
	for _, group := range []struct {
		prefix string
		items  []sourceParameterDescriptor
	}{{"path.", source.Request.Path}, {"query.", source.Request.Query}, {"header.", source.Request.Header}} {
		for _, parameter := range group.items {
			if !covered[group.prefix+parameter.Name] {
				return false
			}
		}
	}
	if source.Request.Body != nil {
		if schema, ok := source.Request.Body.Schema.(map[string]any); ok {
			if properties, ok := schema["properties"].(map[string]any); ok {
				for name := range properties {
					if !covered["body."+name] {
						return false
					}
				}
			}
		}
	}
	return true
}

func sourceGapDirectOperationIsImplementedIncompletely(bundle engine.Bundle, source sourceOperationDescriptor) bool {
	implemented := map[string][]engine.CLIFlag{}
	if bundle.CLISurface != nil {
		for _, command := range bundle.CLISurface.Commands {
			if command.Availability == "implemented" && command.Operation != "" {
				implemented[command.Operation] = append(implemented[command.Operation], command.Flags...)
			}
		}
	}
	for _, operation := range bundle.Operations {
		if operation.REST == nil || len(implemented[operation.ID]) == 0 && !sourceOperationHasNoCallerFields(source) {
			continue
		}
		if sourceProjectionEndpointKey(operation.REST.Method, operation.REST.Path) != sourceProjectionEndpointKey(source.Method, source.Path) {
			continue
		}
		covered := map[string]bool{}
		for field := range sourceProjectionDeclaredConfigPathFields(bundle.Spec, operation.REST.Path) {
			covered[field] = true
		}
		for _, parameter := range operation.REST.Parameters {
			covered[parameter.In+"."+parameter.Name] = true
		}
		if len(operation.REST.BodySchema) > 0 {
			var schema map[string]any
			decoder := json.NewDecoder(bytes.NewReader(operation.REST.BodySchema))
			decoder.UseNumber()
			if decoder.Decode(&schema) == nil {
				if properties, ok := schema["properties"].(map[string]any); ok {
					for name := range properties {
						covered["body."+name] = true
					}
				}
			}
		}
		for _, flag := range implemented[operation.ID] {
			covered[flag.MapsTo] = true
		}
		return !sourceCallerFieldsCovered(source, covered)
	}
	return false
}

// sourceProjectionDeclaredConfigPathFields recognizes only path values backed
// by an exact declared config property. A connector config is validated before
// dispatch and the engine resolves it before CLI path parameters, so this is a
// closed binding rather than an omitted caller input or a raw path escape.
func sourceProjectionDeclaredConfigPathFields(spec *engine.Schema, path string) map[string]bool {
	covered := map[string]bool{}
	if spec == nil {
		return covered
	}
	declared := map[string]bool{}
	for _, name := range spec.Properties() {
		declared[name] = true
	}
	for _, match := range sourceProjectionConfigTemplateRE.FindAllStringSubmatch(path, -1) {
		if declared[match[1]] {
			covered["path."+match[1]] = true
		}
	}
	for _, match := range sourceProjectionPathVariableRE.FindAllStringSubmatch(path, -1) {
		if declared[match[1]] {
			covered["path."+match[1]] = true
		}
	}
	return covered
}

func sourceOperationHasNoCallerFields(operation sourceOperationDescriptor) bool {
	return len(operation.Request.Path) == 0 && len(operation.Request.Query) == 0 && len(operation.Request.Header) == 0 && operation.Request.Body == nil
}

func sourceActionCoversOperation(action engine.WriteAction, command engine.CLICommand, operation sourceOperationDescriptor) bool {
	actionObject := newOrderedObject()
	actionObject.set("body_type", action.BodyType)
	actionObject.set("path", action.Path)
	if action.BinaryUpload != nil {
		upload := newOrderedObject()
		upload.set("source_field", action.BinaryUpload.SourceField)
		actionObject.set("binary_upload", upload)
	}
	pathFields := make([]any, len(action.PathFields))
	for index := range action.PathFields {
		pathFields[index] = action.PathFields[index]
	}
	actionObject.set("path_fields", pathFields)
	contract, err := sourceContractForAction(operation, actionObject)
	if err != nil {
		return false
	}
	schema, err := engine.CompileSchema(action.RecordSchema)
	if err != nil {
		return false
	}
	var schemaDocument map[string]any
	decoder := json.NewDecoder(bytes.NewReader(action.RecordSchema))
	decoder.UseNumber()
	if err := decoder.Decode(&schemaDocument); err != nil {
		return false
	}
	recordProperties, _ := schemaDocument["properties"].(map[string]any)
	properties := map[string]bool{}
	for _, name := range schema.Properties() {
		properties[name] = true
	}
	required := map[string]bool{}
	for _, name := range schema.RequiredKeys() {
		required[name] = true
	}
	flags := map[string]engine.CLIFlag{}
	for _, flag := range command.Flags {
		if name, ok := strings.CutPrefix(flag.MapsTo, "record."); ok {
			flags[name] = flag
		}
	}
	for name := range contract.Fields {
		flag := flags[name]
		flagType := sourceProjectionFlagType(contract.Fields[name])
		if !properties[name] || flag.MapsTo == "" || !sourceProjectionFieldEquivalent(recordProperties[name], contract.Fields[name]) ||
			flag.Type != flagType || flag.MaxBytes != int(sourceProjectionFlagMaxBytes(contract.Fields[name], flagType)) ||
			flag.AllowBareString != contract.BareStringFields[name] ||
			!sourceProjectionFlagEnumEquivalent(flag.Values, contract.Fields[name]) ||
			(contract.Required[name] && (!required[name] || !flag.Required)) || (!contract.Required[name] && (required[name] || flag.Required)) {
			return false
		}
	}
	if contract.Binary {
		fixedOrigin := sourceProjectionFixedOrigin(operation)
		return fixedOrigin != "" && strings.TrimRight(action.BaseURL, "/") == fixedOrigin && action.BodyType == "binary_upload" && action.BinaryUpload != nil
	}
	for _, parameter := range contract.Query {
		query, ok := action.Query[parameter.Name]
		if !ok || query.Template != "{{ record."+parameter.Name+" }}" || query.OmitWhenAbsent == parameter.Required || query.Default != "" {
			return false
		}
	}
	bodyFields := map[string]bool{}
	for _, name := range action.BodyFields {
		bodyFields[name] = true
	}
	for _, name := range contract.BodyFields {
		if !bodyFields[name] {
			return false
		}
	}
	return true
}

func sourceProjectionFieldEquivalent(actual, expected any) bool {
	actualRaw, actualErr := marshalNoEscapeHTML(actual)
	expectedRaw, expectedErr := marshalNoEscapeHTML(expected)
	return actualErr == nil && expectedErr == nil && bytes.Equal(actualRaw, expectedRaw)
}

func sourceProjectionFlagEnumEquivalent(actual []string, schema any) bool {
	object, _ := schema.(map[string]any)
	raw, _ := object["enum"].([]any)
	if len(raw) == 0 {
		return len(actual) == 0
	}
	expected := make([]string, len(raw))
	for index, value := range raw {
		text, ok := value.(string)
		if !ok {
			return false
		}
		expected[index] = text
	}
	return slices.Equal(actual, expected)
}

func sourceProjectionFixedOrigin(operation sourceOperationDescriptor) string {
	for _, layer := range []sourceServerLayer{operation.Servers.Operation, operation.Servers.PathItem} {
		if !sourceServerLayerHasFixedOrigin(layer) {
			continue
		}
		items := layer.Servers.([]any)
		server := items[0].(map[string]any)
		return strings.TrimRight(server["url"].(string), "/")
	}
	return ""
}
