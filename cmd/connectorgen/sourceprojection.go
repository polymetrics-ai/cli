package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	sourceProjectionDefaultStringBytes = 8 << 10
	sourceProjectionDefaultArrayItems  = 256
)

var sourceProjectionTemplateRE = regexp.MustCompile(`\{\{\s*(?:config|record)\.([-A-Za-z0-9_]+)\s*\}\}`)

type sourceProjectionStats struct {
	Writes  int
	CLI     int
	Missing int
}

func (s sourceProjectionStats) Changed() bool { return s.Writes+s.CLI+s.Missing > 0 }

type sourceActionContract struct {
	Fields     map[string]any
	Required   map[string]bool
	Query      []sourceParameterDescriptor
	BodyFields []string
	Binary     bool
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
		if operation.Protocol == "graphql" || !sourceProjectionMutationMethod(operation.Method) || sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
			continue
		}
		candidates := actionsByEndpoint[sourceProjectionEndpointKey(operation.Method, operation.Path)]
		if len(candidates) == 0 {
			stats.Missing++
			continue
		}
		action := sourceProjectionBroadAction(candidates)
		contract, err := sourceContractForAction(operation, action)
		if err != nil {
			// The imported descriptor retains the typed gap. Generation never
			// invents a lossy record shape for a contract it cannot express.
			stats.Missing++
			continue
		}
		if sourceProjectAction(action, contract) {
			stats.Writes++
		}
		if len(commandsByWrite[stringField(action, "name")]) == 0 {
			command := sourceProjectionNewCommand(operation, action)
			commands := append(arrayField(cli.root, "commands"), command)
			cli.root.set("commands", commands)
			commandsByWrite[stringField(action, "name")] = []*orderedObject{command}
			stats.CLI++
		}
		for _, command := range commandsByWrite[stringField(action, "name")] {
			if sourceProjectCommand(command, contract) {
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
	contract := sourceActionContract{Fields: map[string]any{}, Required: map[string]bool{}}
	pathFields := map[string]bool{}
	for _, raw := range arrayField(action, "path_fields") {
		if name, ok := raw.(string); ok {
			pathFields[name] = true
		}
	}
	for _, parameter := range operation.Request.Path {
		if !pathFields[parameter.Name] {
			continue
		}
		converted, err := sourceProjectionSchema(parameter.Schema)
		if err != nil {
			return sourceActionContract{}, err
		}
		contract.Fields[parameter.Name] = converted
		contract.Required[parameter.Name] = true
	}
	for _, parameter := range operation.Request.Query {
		converted, err := sourceProjectionSchema(parameter.Schema)
		if err != nil {
			return sourceActionContract{}, err
		}
		contract.Fields[parameter.Name] = converted
		contract.Required[parameter.Name] = parameter.Required
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
		properties, _ := body["properties"].(map[string]any)
		required := sourceSchemaRequired(body)
		for _, name := range sortedSourceMapKeys(properties) {
			converted, err := sourceProjectionSchema(properties[name])
			if err != nil {
				return sourceActionContract{}, err
			}
			contract.Fields[name] = converted
			contract.Required[name] = required[name]
			contract.BodyFields = append(contract.BodyFields, name)
		}
	}
	sort.Slice(contract.Query, func(i, j int) bool { return contract.Query[i].Name < contract.Query[j].Name })
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
	for _, key := range []string{"pattern", "minLength", "maxLength", "minItems", "maxItems", "minProperties"} {
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
		flag.set("type", sourceProjectionFlagType(contract.Fields[name]))
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
		if maxBytes := sourceProjectionSchemaMaxBytes(contract.Fields[name]); maxBytes > 0 {
			flag.set("max_bytes", json.Number(fmt.Sprintf("%d", maxBytes)))
		}
		flags = append(flags, flag)
	}
	return setOrderedIfDifferent(command, "flags", flags)
}

func sourceProjectionNewCommand(operation sourceOperationDescriptor, action *orderedObject) *orderedObject {
	command := newOrderedObject()
	path := strings.NewReplacer("/", " ", "_", "-").Replace(operation.SourceID)
	command.set("path", "api "+path)
	command.set("summary", strings.ToUpper(operation.Method)+" "+operation.Path)
	command.set("intent", "reverse_etl")
	command.set("availability", "implemented")
	command.set("write", stringField(action, "name"))
	if operation.Source.URL != "" {
		command.set("source_url", operation.Source.URL)
	}
	command.set("risk", stringField(action, "risk"))
	command.set("approval", "Reverse ETL writes require plan, preview, approval, and execute.")
	endpoint := newOrderedObject()
	endpoint.set("method", strings.ToUpper(operation.Method))
	endpoint.set("path", operation.Path)
	command.set("api_surface", []any{endpoint})
	return command
}

func sourceProjectionFlagType(schema any) string {
	object, _ := schema.(map[string]any)
	if enum, ok := object["enum"].([]any); ok && len(enum) > 0 {
		return "enum"
	}
	typeName := object["type"]
	if values, ok := typeName.([]any); ok {
		for _, value := range values {
			if value != "null" {
				typeName = value
				break
			}
		}
	}
	switch typeName {
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
	return leftErr == nil && rightErr == nil && bytes.Equal(leftRaw, rightRaw)
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
		expected[fmt.Sprintf("%s.graphql.query.%s", connector, field.Name)] = sourceImportSource{SHA256: strings.ToLower(lock.GraphQL.SHA256), Bytes: lock.GraphQL.Bytes}
	}
	for _, field := range lock.GraphQL.MutationFields {
		expected[fmt.Sprintf("%s.graphql.mutation.%s", connector, field.Name)] = sourceImportSource{SHA256: strings.ToLower(lock.GraphQL.SHA256), Bytes: lock.GraphQL.Bytes}
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
	for _, operation := range descriptor.Operations {
		if operation.Protocol == "graphql" || !sourceProjectionMutationMethod(operation.Method) || sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
			continue
		}
		candidates := actions[sourceProjectionEndpointKey(operation.Method, operation.Path)]
		if len(candidates) == 0 {
			return []Finding{sourceProjectionFinding(bundle.Name, file, "source operation has no executable action: "+operation.SourceID)}
		}
		complete := false
		for _, action := range candidates {
			if sourceActionCoversOperation(action, commands[action.Name], operation) {
				complete = true
				break
			}
		}
		if !complete {
			return []Finding{sourceProjectionFinding(bundle.Name, file, "source request fields are missing from action/CLI union: "+operation.SourceID)}
		}
	}
	return nil
}

func sourceActionCoversOperation(action engine.WriteAction, command engine.CLICommand, operation sourceOperationDescriptor) bool {
	actionObject := newOrderedObject()
	actionObject.set("body_type", action.BodyType)
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
		if !properties[name] || flag.MapsTo == "" || !sourceProjectionFieldEquivalent(recordProperties[name], contract.Fields[name]) ||
			flag.Type != sourceProjectionFlagType(contract.Fields[name]) || flag.MaxBytes != int(sourceProjectionSchemaMaxBytes(contract.Fields[name])) ||
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
