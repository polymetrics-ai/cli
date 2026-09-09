package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/safety"
)

type vNextSourceProjectionInputs struct {
	ctx       context.Context
	directory *vNextPublicationDirectory
	retained  map[string][]byte
	bytes     int64
}

func (inputs *vNextSourceProjectionInputs) read(name string, limit int64) ([]byte, error) {
	if err := inputs.ctx.Err(); err != nil {
		return nil, err
	}
	if !sourceLaneRelativePath(name) || limit <= 0 || limit > 64<<20 {
		return nil, fmt.Errorf("invalid retained-source read")
	}
	file, err := inputs.directory.openRegular(name, "retained source", unix.O_RDONLY)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	if err := errors.Join(readErr, file.Close()); err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("retained source byte limit exceeded")
	}
	if previous, exists := inputs.retained[name]; exists {
		if !bytes.Equal(previous, data) {
			return nil, fmt.Errorf("retained source changed during admission")
		}
	} else {
		inputs.bytes += int64(len(data))
		if inputs.bytes > 512<<20 {
			return nil, fmt.Errorf("retained source aggregate byte limit exceeded")
		}
		inputs.retained[name] = append([]byte(nil), data...)
	}
	return data, nil
}

func (inputs *vNextSourceProjectionInputs) revalidate() error {
	names := make([]string, 0, len(inputs.retained))
	for name := range inputs.retained {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := inputs.read(name, 64<<20); err != nil {
			return err
		}
	}
	return nil
}

func (inputs *vNextSourceProjectionInputs) inventory(lock vNextSourceLock) (retainedSourceInventory, error) {
	cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "source-projection-v1"}
	for _, pin := range lock.SourceProjection.Inventories {
		data, err := inputs.read(pin.Path, 64<<20)
		if err != nil {
			return retainedSourceInventory{}, err
		}
		if int64(len(data)) != pin.Bytes || sourceBytesHash(data) != pin.SHA256 {
			return retainedSourceInventory{}, fmt.Errorf("source_projection inventory %q pin mismatch", pin.ID)
		}
		if _, err := countSourceJSONNodes(inputs.ctx, data, 1_000_000); err != nil {
			return retainedSourceInventory{}, err
		}
		var envelope struct {
			SchemaVersion int    `json:"schema_version"`
			Connector     string `json:"connector"`
			Rest          struct {
				Operations []struct {
					ID string `json:"id"`
				} `json:"operations"`
				Documents []struct {
					Artifact struct {
						SHA256 string `json:"sha256"`
						Bytes  int64  `json:"bytes"`
					} `json:"artifact"`
					Operations []struct {
						ID string `json:"id"`
					} `json:"operations"`
				} `json:"source_documents"`
			} `json:"rest"`
		}
		if err := decodeSourceJSON(data, &envelope); err != nil {
			return retainedSourceInventory{}, err
		}
		if (envelope.SchemaVersion != 2 && envelope.SchemaVersion != 3) || envelope.Connector != lock.Connector {
			return retainedSourceInventory{}, fmt.Errorf("source_projection inventory %q requires a supported retained envelope", pin.ID)
		}
		class, err := pin.Class.resolved()
		if err != nil {
			return retainedSourceInventory{}, err
		}
		anchor := sourceInventoryAnchor{Connector: lock.Connector, Inventory: pin.ID, Class: class, Path: pin.Path, SHA256: pin.SHA256}
		if envelope.SchemaVersion == 2 {
			for _, operation := range envelope.Rest.Operations {
				anchor.ExpectedIDs = append(anchor.ExpectedIDs, operation.ID)
			}
		} else {
			for _, document := range envelope.Rest.Documents {
				if !sourceLaneDigest(document.Artifact.SHA256) || document.Artifact.Bytes <= 0 || document.Artifact.Bytes > 64<<20 {
					return retainedSourceInventory{}, fmt.Errorf("source_projection inventory %q has an invalid raw-document pin", pin.ID)
				}
				// Retained source packages store immutable artifact objects next
				// to their inventory, named by their full content digest.
				artifactPath := path.Join(path.Dir(pin.Path), "artifacts", document.Artifact.SHA256+".artifact")
				anchor.Artifacts = append(anchor.Artifacts, sourceArtifactPin{Path: artifactPath, SHA256: document.Artifact.SHA256, Bytes: document.Artifact.Bytes})
				for _, operation := range document.Operations {
					anchor.ExpectedIDs = append(anchor.ExpectedIDs, operation.ID)
				}
			}
		}
		anchor.ExpectedCount = len(anchor.ExpectedIDs)
		cohort.Inventories = append(cohort.Inventories, anchor)
	}
	inventory := loadRetainedSourceInventoryUsingReader(inputs.ctx, "", cohort, 1_000_000, 8_000_000, inputs.read)
	if len(inventory.Diagnostics) != 0 {
		return retainedSourceInventory{}, fmt.Errorf("source_projection inventory refused: %s", inventory.Diagnostics[0].Code)
	}
	if err := inputs.loadDocuments(lock, &inventory); err != nil {
		return retainedSourceInventory{}, err
	}
	return inventory, nil
}

func lowerVNextSourceProjection(lock vNextSourceLock, inventory retainedSourceInventory) (vNextSourceLock, error) {
	semantics := map[vNextSourceProjectionKey]vNextSourceProjectionSemantic{}
	for _, semantic := range lock.SourceProjection.Semantics {
		semantics[semantic.Source] = semantic
	}
	documents := map[string]retainedSourceDocument{}
	for _, document := range inventory.Documents {
		documents[document.ID] = document
	}
	for _, semantic := range lock.SourceProjection.Semantics {
		for _, ref := range semantic.Evidence {
			doc, ok := documents[ref.DocumentID]
			if !ok {
				return vNextSourceLock{}, fmt.Errorf("source_projection citation document missing")
			}
			if err := validateSourceProjectionCitation(doc, ref); err != nil {
				return vNextSourceLock{}, err
			}
		}
	}
	lowered := lock
	lowered.SourceProjection = nil
	lowered.Schemas = map[string]json.RawMessage{}
	lowered.Operations = nil
	for _, row := range inventory.Operations {
		key := vNextSourceProjectionKey{Inventory: row.Key.Inventory, ID: row.Key.ID}
		semantic, exists := semantics[key]
		if !exists {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q has unresolved execution semantics", row.Key.ID)
		}
		delete(semantics, key)
		var rawDocument *retainedSourceDocument
		if row.RawDocumentID != "" {
			document, exists := documents[row.RawDocumentID]
			if !exists {
				return vNextSourceLock{}, fmt.Errorf("source_projection source %q raw document is absent", row.Key.ID)
			}
			rawDocument = &document
		}
		facts := normalizeSourceFacts(row, documents[row.DocumentID], rawDocument)
		if semantic.Write != nil {
			descriptor, err := sourceProjectionTypedWrite(facts, semantic.Write, row.Key)
			if err != nil {
				return vNextSourceLock{}, fmt.Errorf("source_projection source %q: %w", row.Key.ID, err)
			}
			lowered.Operations = append(lowered.Operations, descriptor)
			continue
		}
		if semantic.Collection == nil || (semantic.Effect != "" && semantic.Effect != "read") {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q has unresolved collection semantics", row.Key.ID)
		}
		if facts.Status != "available" || len(facts.Diagnostics) != 0 || facts.Protocol != "rest" || facts.Method != "GET" {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q has an unsupported source shape", row.Key.ID)
		}
		if body := facts.Groups["request_body"]; len(body) != 0 && string(body) != "null" {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q request body is not yet lowered", row.Key.ID)
		}
		response := semantic.Collection.Records.Response
		if response == nil || response.Media != "application/json" || response.Status != "200" {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q response variant is not yet lowered", row.Key.ID)
		}
		operation, ok := sourceResolveObject(facts, facts.Groups["source_operation"], map[string]bool{}, 0)
		if !ok {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q operation is unresolved", row.Key.ID)
		}
		name := facts.OperationID
		if inner, present := operation["operationId"]; present {
			var identity string
			if err := json.Unmarshal(inner, &identity); err != nil || identity == "" || (name != "" && name != identity) {
				return vNextSourceLock{}, fmt.Errorf("source_projection source %q has inconsistent operation identity", row.Key.ID)
			}
			name = identity
		}
		name = sourceProjectionDefaultName(name)
		if !namePattern.MatchString(name) {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q requires supported generated identity", row.Key.ID)
		}
		name = strings.ReplaceAll(name, "-", "_")
		recordSchema, recordPath, err := sourceProjectionRecordSchema(facts, semantic.Collection)
		if err != nil {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q: %w", row.Key.ID, err)
		}
		schemaPath := "schemas/" + sourceBytesHash(recordSchema) + ".json"
		lowered.Schemas[schemaPath] = recordSchema
		streamPath, query, configSchema, err := sourceProjectionStreamInputs(facts, lowered.ConfigSchema)
		if err != nil {
			return vNextSourceLock{}, fmt.Errorf("source_projection source %q: %w", row.Key.ID, err)
		}
		lowered.ConfigSchema = configSchema
		inputSchema, inputContract, err := sourceProjectionInputContract(facts)
		if err != nil {
			return vNextSourceLock{}, err
		}
		lowered.Schemas[inputContract.Schema] = inputSchema
		stream, err := json.Marshal(map[string]any{"name": name, "path": streamPath, "query": query, "request_inputs": inputContract, "records": map[string]string{"path": recordPath}, "schema": schemaPath})
		if err != nil {
			return vNextSourceLock{}, err
		}
		source, err := json.Marshal(row.Key)
		if err != nil {
			return vNextSourceLock{}, err
		}
		descriptor := vNextOperationDescriptor{ID: "stream:" + name, Source: source, Stream: stream, SchemaRefs: vNextSchemaReferences{Record: schemaPath, Input: inputContract.Schema}}
		if lock.Lanes["direct_read"] == "implemented" {
			operation, command, err := sourceProjectionDirectInput(facts, name)
			if err != nil {
				return vNextSourceLock{}, fmt.Errorf("source_projection source %q: %w", row.Key.ID, err)
			}
			var operationFields map[string]json.RawMessage
			if err := decodeSourceJSON(operation, &operationFields); err != nil {
				return vNextSourceLock{}, err
			}
			var restFields map[string]json.RawMessage
			if err := decodeSourceJSON(operationFields["rest"], &restFields); err != nil {
				return vNextSourceLock{}, err
			}
			restFields["request_inputs"], err = json.Marshal(inputContract)
			if err != nil {
				return vNextSourceLock{}, err
			}
			operationFields["rest"], err = json.Marshal(restFields)
			if err != nil {
				return vNextSourceLock{}, err
			}
			operation, err = json.Marshal(operationFields)
			if err != nil {
				return vNextSourceLock{}, err
			}
			descriptor.Operation = operation
			descriptor.Commands = []vNextCommandDescriptor{{Command: command}}
			if len(lowered.CLI) == 0 {
				lowered.CLI, err = json.Marshal(map[string]string{"usage": "pm " + lock.Connector + " <command>", "tagline": lock.Connector + " commands"})
				if err != nil {
					return vNextSourceLock{}, err
				}
			}
		}
		lowered.Operations = append(lowered.Operations, descriptor)
	}
	if len(semantics) != 0 {
		return vNextSourceLock{}, fmt.Errorf("source_projection semantics contain an unretained source key")
	}
	return lowered, nil
}

func sourceProjectionRecordSchema(facts sourceFacts, collection *vNextSourceProjectionCollection) (json.RawMessage, string, error) {
	coordinate := collection.Records.Response
	if coordinate == nil || coordinate.Pointer == nil {
		return nil, "", fmt.Errorf("collection requires a response coordinate")
	}
	var responses map[string]json.RawMessage
	if err := decodeSourceJSON(facts.Groups["responses"], &responses); err != nil {
		return nil, "", err
	}
	response, ok := sourceResolveObject(facts, responses[coordinate.Status], map[string]bool{}, 0)
	if !ok {
		return nil, "", fmt.Errorf("response status does not resolve")
	}
	var content map[string]json.RawMessage
	if err := decodeSourceJSON(response["content"], &content); err != nil {
		return nil, "", err
	}
	media, ok := sourceResolveObject(facts, content[coordinate.Media], map[string]bool{}, 0)
	if !ok {
		return nil, "", fmt.Errorf("response media does not resolve")
	}
	node := media["schema"]
	var fields []string
	if *coordinate.Pointer != "" {
		fields = strings.Split(strings.TrimPrefix(*coordinate.Pointer, "/"), "/")
	}
	for _, field := range fields {
		if !namePattern.MatchString(field) {
			return nil, "", fmt.Errorf("collection field requires a supported record path")
		}
		object, ok := sourceProjectionTypedSchema(facts, node, "object")
		if !ok {
			return nil, "", fmt.Errorf("collection field crosses an unsupported schema shape")
		}
		var properties map[string]json.RawMessage
		if err := decodeSourceJSON(object["properties"], &properties); err != nil {
			return nil, "", err
		}
		node = properties[field]
	}
	array, ok := sourceProjectionTypedSchema(facts, node, "array")
	if !ok {
		return nil, "", fmt.Errorf("collection coordinate does not select a supported array")
	}
	record, ok := sourceProjectionTypedSchema(facts, array["items"], "object")
	if !ok {
		return nil, "", fmt.Errorf("collection items do not select a supported record object")
	}
	var properties map[string]json.RawMessage
	if err := decodeSourceJSON(record["properties"], &properties); err != nil {
		return nil, "", err
	}
	keys := make([]string, 0, len(collection.PrimaryKey))
	for _, coordinate := range collection.PrimaryKey {
		key := strings.TrimPrefix(coordinate, "/")
		if !namePattern.MatchString(key) || len(properties[key]) == 0 {
			return nil, "", fmt.Errorf("primary key does not resolve to a supported record field")
		}
		keys = append(keys, key)
	}
	keyBytes, err := json.Marshal(keys)
	if err != nil {
		return nil, "", err
	}
	projected := make(map[string]json.RawMessage, len(record)+1)
	for key, value := range record {
		projected[key] = value
	}
	projected["x-primary-key"] = keyBytes
	encoded, err := json.Marshal(projected)
	return encoded, strings.Join(fields, "."), err
}

func sourceProjectionTypedSchema(facts sourceFacts, raw json.RawMessage, want string) (map[string]json.RawMessage, bool) {
	object, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
	if !ok {
		return nil, false
	}
	var kind string
	if err := json.Unmarshal(object["type"], &kind); err != nil || kind != want {
		return nil, false
	}
	for _, composition := range []string{"allOf", "anyOf", "oneOf", "not", "if", "then", "else"} {
		if _, exists := object[composition]; exists {
			return nil, false
		}
	}
	return object, true
}

// sourceProjectionStreamInputs derives saved-stream scoping from effective
// source parameters. Direct-operation flags have their own canonical consumer.
func sourceProjectionStreamInputs(facts sourceFacts, configSchema json.RawMessage) (string, map[string]any, json.RawMessage, error) {
	var config map[string]json.RawMessage
	if err := decodeSourceJSON(configSchema, &config); err != nil {
		return "", nil, nil, err
	}
	properties := map[string]json.RawMessage{}
	if raw := config["properties"]; len(raw) != 0 {
		if err := decodeSourceJSON(raw, &properties); err != nil {
			return "", nil, nil, err
		}
	}
	streamPath := facts.Path
	query := map[string]any{}
	locations := map[string]string{}
	for _, parameter := range facts.Parameters {
		if safety.ValidateIdentifier(parameter.Name, "source parameter") != nil || (parameter.In != "path" && parameter.In != "query") {
			return "", nil, nil, fmt.Errorf("parameter requires an unsupported stream binding")
		}
		if previous, exists := locations[parameter.Name]; exists && previous != parameter.In {
			return "", nil, nil, fmt.Errorf("parameter configuration names require source-coordinate alias reconciliation")
		}
		locations[parameter.Name] = parameter.In
		node, ok := sourceResolveObject(facts, parameter.Node, map[string]bool{}, 0)
		if !ok {
			return "", nil, nil, fmt.Errorf("parameter does not resolve")
		}
		schema, _, ok := sourceProjectionScalarSchema(facts, node["schema"])
		if !ok {
			return "", nil, nil, fmt.Errorf("parameter requires an unsupported stream schema")
		}
		// Nullable values and non-default serialization require their distinct
		// typed representation; never erase either to manufacture a string input.
		if raw := schema["nullable"]; len(raw) != 0 && string(raw) != "false" {
			return "", nil, nil, fmt.Errorf("nullable parameter representation is not yet lowered")
		}
		for _, field := range []string{"content", "style", "explode", "allowReserved", "allowEmptyValue"} {
			if len(node[field]) != 0 {
				return "", nil, nil, fmt.Errorf("parameter serialization is not yet lowered")
			}
		}
		rendered, err := json.Marshal(schema)
		if err != nil {
			return "", nil, nil, err
		}
		if existing, exists := properties[parameter.Name]; exists {
			var prior map[string]json.RawMessage
			if err := decodeSourceJSON(existing, &prior); err != nil {
				return "", nil, nil, err
			}
			canonical, err := json.Marshal(prior)
			if err != nil || !bytes.Equal(canonical, rendered) {
				return "", nil, nil, fmt.Errorf("source parameter conflicts with an existing configuration schema")
			}
		}
		template := "{{ config." + parameter.Name + " }}"
		if parameter.In == "path" {
			placeholder := "{" + parameter.Name + "}"
			if !parameter.Required || !strings.Contains(streamPath, placeholder) {
				return "", nil, nil, fmt.Errorf("path parameter has inconsistent source requiredness or location")
			}
			streamPath = strings.ReplaceAll(streamPath, placeholder, template)
		} else {
			query[parameter.Name] = map[string]any{"template": template, "omit_when_absent": !parameter.Required}
		}
	}
	rawProperties, err := json.Marshal(properties)
	if err != nil {
		return "", nil, nil, err
	}
	config["properties"] = rawProperties
	result, err := json.Marshal(config)
	return streamPath, query, result, err
}

func sourceProjectionDirectInput(facts sourceFacts, name string) (json.RawMessage, json.RawMessage, error) {
	parameters := make([]map[string]any, 0, len(facts.Parameters))
	flags := make([]map[string]any, 0, len(facts.Parameters))
	for _, p := range facts.Parameters {
		node, ok := sourceResolveObject(facts, p.Node, map[string]bool{}, 0)
		if !ok {
			return nil, nil, fmt.Errorf("direct parameter does not resolve")
		}
		schema, kind, ok := sourceProjectionScalarSchema(facts, node["schema"])
		if !ok {
			return nil, nil, fmt.Errorf("direct parameter schema is not yet lowered")
		}
		// Do not drop validation keywords while projecting a narrower parameter
		// dialect. Broader source schemas require the matching typed consumer.
		for field := range schema {
			switch field {
			case "type", "description", "title", "enum", "minimum", "maximum":
			default:
				return nil, nil, fmt.Errorf("direct parameter schema keyword %q is not yet lowered", field)
			}
		}
		parameter := map[string]any{"name": p.Name, "in": p.In, "type": kind, "required": p.Required}
		flag := map[string]any{"name": p.Name, "type": kind, "input_codec": "source_scalar_v1", "summary": p.Name, "required": p.Required, "maps_to": p.In + "." + p.Name}
		for _, bound := range []string{"minimum", "maximum"} {
			if raw := schema[bound]; len(raw) != 0 {
				parameter[bound], flag[bound] = raw, raw
			}
		}
		if raw := schema["enum"]; len(raw) != 0 {
			var values []string
			if err := decodeSourceJSON(raw, &values); err != nil || len(values) == 0 {
				return nil, nil, fmt.Errorf("direct parameter enum is not a nonempty string set")
			}
			parameter["values"], flag["values"] = values, values
		}
		parameters = append(parameters, parameter)
		flags = append(flags, flag)
	}
	operation, err := json.Marshal(map[string]any{"id": name, "kind": "rest_read", "summary": name, "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": facts.Method, "path": facts.Path, "max_bytes": commandrunner.MaxOperationDirectReadBytes, "parameters": parameters, "response": map[string]any{"success_statuses": []string{"200"}}}})
	if err != nil {
		return nil, nil, err
	}
	command, err := json.Marshal(map[string]any{"path": "api " + sourceProjectionDefaultName(name), "summary": name, "intent": "direct_read", "availability": "implemented", "operation": name, "api_surface": []map[string]string{{"method": facts.Method, "path": facts.Path}}, "output_policy": "json_redacted", "flags": flags})
	return operation, command, err
}

func sourceProjectionTypedWrite(facts sourceFacts, semantics *vNextSourceProjectionWrite, key sourceOperationKey) (vNextOperationDescriptor, error) {
	if facts.Status != "available" || len(facts.Diagnostics) != 0 || facts.Protocol != "rest" || facts.Method != "POST" || len(facts.Parameters) != 0 {
		return vNextOperationDescriptor{}, fmt.Errorf("typed mutation source shape is not yet lowered")
	}
	operation, ok := sourceResolveObject(facts, facts.Groups["source_operation"], map[string]bool{}, 0)
	if !ok {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation operation does not resolve")
	}
	name := facts.OperationID
	if raw, present := operation["operationId"]; present {
		var inner string
		if err := json.Unmarshal(raw, &inner); err != nil || inner == "" || (name != "" && name != inner) {
			return vNextOperationDescriptor{}, fmt.Errorf("inconsistent mutation operation identity")
		}
		name = inner
	}
	name = sourceProjectionDefaultName(name)
	if !namePattern.MatchString(name) {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation identity requires name reconciliation")
	}
	name = strings.ReplaceAll(name, "-", "_")
	body, ok := sourceResolveObject(facts, facts.Groups["request_body"], map[string]bool{}, 0)
	if !ok {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation request body does not resolve")
	}
	var content map[string]json.RawMessage
	if err := decodeSourceJSON(body["content"], &content); err != nil || len(content) != 1 {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation request media requires accounted variants")
	}
	media, ok := sourceResolveObject(facts, content["application/json"], map[string]bool{}, 0)
	if !ok {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation media is not yet lowered")
	}
	schema, ok := sourceProjectionTypedSchema(facts, media["schema"], "object")
	if !ok {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation schema requires supported object projection")
	}
	encoded, err := json.Marshal(schema)
	if err != nil {
		return vNextOperationDescriptor{}, err
	}
	if _, err := engine.CompileSchema(encoded); err != nil {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation schema projection: %w", err)
	}
	var properties map[string]json.RawMessage
	if err := decodeSourceJSON(schema["properties"], &properties); err != nil || len(properties) == 0 {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation requires concrete typed fields")
	}
	fields := make([]string, 0, len(properties))
	for field := range properties {
		if !namePattern.MatchString(field) {
			return vNextOperationDescriptor{}, fmt.Errorf("mutation field requires source-coordinate alias")
		}
		fields = append(fields, field)
	}
	sort.Strings(fields)
	var required bool
	if raw := body["required"]; len(raw) != 0 {
		if err := json.Unmarshal(raw, &required); err != nil {
			return vNextOperationDescriptor{}, err
		}
	}
	if !required {
		return vNextOperationDescriptor{}, fmt.Errorf("optional mutation body requires an accounted absent-body variant")
	}
	var responses map[string]json.RawMessage
	if err := decodeSourceJSON(facts.Groups["responses"], &responses); err != nil {
		return vNextOperationDescriptor{}, err
	}
	var statuses []int
	for status := range responses {
		code, err := strconv.Atoi(status)
		if err == nil && len(status) == 3 && code >= 200 && code < 300 {
			statuses = append(statuses, code)
		} else if strings.HasPrefix(status, "2") {
			return vNextOperationDescriptor{}, fmt.Errorf("mutation success status family is not yet lowered")
		}
	}
	if len(statuses) == 0 {
		return vNextOperationDescriptor{}, fmt.Errorf("mutation requires concrete source success statuses")
	}
	sort.Ints(statuses)
	// The existing unkeyed, non-delete writeRequester disables retries. No
	// source promise of provider idempotency or managed transport is invented.
	action, err := json.Marshal(engine.WriteAction{Name: name, Kind: "custom", Method: facts.Method, Path: facts.Path, Risk: semantics.Risk, Batchable: semantics.Batchable, BodyType: "json", BodyRequired: required, BodyFields: fields, BodySchema: encoded, RecordSchema: encoded, SuccessStatuses: statuses})
	if err != nil {
		return vNextOperationDescriptor{}, err
	}
	source, err := json.Marshal(key)
	if err != nil {
		return vNextOperationDescriptor{}, err
	}
	return vNextOperationDescriptor{ID: "write:" + name, Source: source, Write: action}, nil
}

// sourceProjectionDefaultName realizes the ordinary ASCII kebab spelling.
// Exact baseline aliases and collision reservations are separate source joins.
func sourceProjectionDefaultName(identity string) string {
	var name strings.Builder
	separator := false
	for i := 0; i < len(identity); i++ {
		c := identity[i]
		if c < 32 || c >= 127 {
			return ""
		}
		upper := c >= 'A' && c <= 'Z'
		lower := c >= 'a' && c <= 'z'
		digit := c >= '0' && c <= '9'
		if !upper && !lower && !digit {
			separator = name.Len() > 0
			continue
		}
		if upper && i > 0 {
			previous := identity[i-1]
			nextLower := i+1 < len(identity) && identity[i+1] >= 'a' && identity[i+1] <= 'z'
			if (previous >= 'a' && previous <= 'z') || (previous >= '0' && previous <= '9') || (previous >= 'A' && previous <= 'Z' && nextLower) {
				separator = name.Len() > 0
			}
		}
		if separator {
			name.WriteByte('-')
			separator = false
		}
		if upper {
			c += 'a' - 'A'
		}
		name.WriteByte(c)
	}
	return name.String()
}

// sourceProjectionScalarSchema preserves the declared scalar kind and exact
// numeric lexemes. Serialization and richer constraints retain their own gates.
func sourceProjectionScalarSchema(facts sourceFacts, raw json.RawMessage) (map[string]json.RawMessage, string, bool) {
	for _, kind := range []string{"string", "integer", "number", "boolean"} {
		if schema, ok := sourceProjectionTypedSchema(facts, raw, kind); ok {
			return schema, kind, true
		}
	}
	return nil, "", false
}
