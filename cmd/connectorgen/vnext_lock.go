package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
)

const vNextSourceLockSchemaVersion = 4

var vNextLaneNames = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

// vNextSourceLock is immutable authoring input. ProviderEvidence is deliberately
// absent from canonicalDescriptor and therefore cannot affect runtime output.
// Execution facts are authored once per operation; commands and lanes project
// into the engine's existing execution JSON files.
type vNextSourceLock struct {
	SchemaVersion    int                        `json:"schema_version"`
	Connector        string                     `json:"connector"`
	Lanes            map[string]string          `json:"lanes"`
	ProviderEvidence json.RawMessage            `json:"provider_evidence,omitempty"`
	Metadata         json.RawMessage            `json:"metadata"`
	ConfigSchema     json.RawMessage            `json:"config_schema"`
	HTTP             json.RawMessage            `json:"http,omitempty"`
	Schemas          map[string]json.RawMessage `json:"schemas,omitempty"`
	Operations       []vNextOperationDescriptor `json:"operations,omitempty"`
	CLI              json.RawMessage            `json:"cli,omitempty"`
	Execution        map[string]json.RawMessage `json:"execution,omitempty"`
}

// vNextOperationDescriptor is the canonical per-operation authoring unit. A
// single provider operation may populate multiple lanes without duplicating its
// source identity or shared schema references.
type vNextOperationDescriptor struct {
	ID             string                   `json:"id"`
	Source         json.RawMessage          `json:"source,omitempty"`
	SchemaRefs     vNextSchemaReferences    `json:"schema_refs,omitempty"`
	Stream         json.RawMessage          `json:"stream,omitempty"`
	StreamOrder    int                      `json:"stream_order,omitempty"`
	Write          json.RawMessage          `json:"write,omitempty"`
	WriteOrder     int                      `json:"write_order,omitempty"`
	Operation      json.RawMessage          `json:"operation,omitempty"`
	OperationOrder int                      `json:"operation_order,omitempty"`
	Commands       []vNextCommandDescriptor `json:"commands,omitempty"`
}

type vNextCommandDescriptor struct {
	Order   int             `json:"order"`
	Command json.RawMessage `json:"command"`
}

type vNextSchemaReferences struct {
	Request  string `json:"request,omitempty"`
	Response string `json:"response,omitempty"`
	Record   string `json:"record,omitempty"`
}

type vNextCanonicalDescriptor struct {
	Connector    string
	Metadata     json.RawMessage
	ConfigSchema json.RawMessage
	HTTP         json.RawMessage
	Schemas      map[string]json.RawMessage
	Operations   []vNextOperationDescriptor
	CLI          json.RawMessage
	Execution    map[string]json.RawMessage
}

func canonicalizeVNextSourceLock(lock vNextSourceLock) (vNextCanonicalDescriptor, error) {
	if lock.SchemaVersion != vNextSourceLockSchemaVersion {
		return vNextCanonicalDescriptor{}, fmt.Errorf("source lock schema_version %d is unsupported; want %d", lock.SchemaVersion, vNextSourceLockSchemaVersion)
	}
	if !namePattern.MatchString(lock.Connector) {
		return vNextCanonicalDescriptor{}, fmt.Errorf("source lock connector %q is invalid", lock.Connector)
	}
	observedLanes := observedVNextLanes(lock)
	if len(lock.Lanes) != len(vNextLaneNames) {
		return vNextCanonicalDescriptor{}, fmt.Errorf("source lock requires explicit declarations for all %d execution lanes", len(vNextLaneNames))
	}
	for _, lane := range vNextLaneNames {
		state, ok := lock.Lanes[lane]
		if !ok {
			return vNextCanonicalDescriptor{}, fmt.Errorf("source lock lane %q is undeclared", lane)
		}
		if state != "implemented" && state != "unsupported" {
			return vNextCanonicalDescriptor{}, fmt.Errorf("source lock lane %q has invalid state %q", lane, state)
		}
		if observedLanes[lane] != (state == "implemented") {
			return vNextCanonicalDescriptor{}, fmt.Errorf("source lock lane %q is %s but its authored execution content says implemented=%t", lane, state, observedLanes[lane])
		}
	}
	for lane := range lock.Lanes {
		if !slicesContains(vNextLaneNames, lane) {
			return vNextCanonicalDescriptor{}, fmt.Errorf("source lock declares unknown lane %q", lane)
		}
	}
	if err := validateRawJSONObject("metadata", lock.Metadata); err != nil {
		return vNextCanonicalDescriptor{}, err
	}
	if err := validateRawJSONObject("config_schema", lock.ConfigSchema); err != nil {
		return vNextCanonicalDescriptor{}, err
	}
	if len(lock.HTTP) != 0 {
		if err := validateRawJSONObject("http", lock.HTTP); err != nil {
			return vNextCanonicalDescriptor{}, err
		}
	}
	schemas := make(map[string]json.RawMessage, len(lock.Schemas))
	for name, schema := range lock.Schemas {
		if !validVNextSchemaPath(name) {
			return vNextCanonicalDescriptor{}, fmt.Errorf("shared schema path %q is invalid", name)
		}
		if err := validateRawJSONObject(name, schema); err != nil {
			return vNextCanonicalDescriptor{}, err
		}
		schemas[name] = cloneRawJSON(schema)
	}
	operations := append([]vNextOperationDescriptor(nil), lock.Operations...)
	seen := make(map[string]struct{}, len(operations))
	for index := range operations {
		descriptor := &operations[index]
		if strings.TrimSpace(descriptor.ID) == "" {
			return vNextCanonicalDescriptor{}, fmt.Errorf("operation %d requires an id", index)
		}
		if _, duplicate := seen[descriptor.ID]; duplicate {
			return vNextCanonicalDescriptor{}, fmt.Errorf("source lock duplicates operation %q", descriptor.ID)
		}
		seen[descriptor.ID] = struct{}{}
		if len(descriptor.Stream) == 0 && len(descriptor.Write) == 0 && len(descriptor.Operation) == 0 && len(descriptor.Commands) == 0 {
			return vNextCanonicalDescriptor{}, fmt.Errorf("operation %q has no execution lane", descriptor.ID)
		}
		for role, reference := range map[string]string{
			"request": descriptor.SchemaRefs.Request, "response": descriptor.SchemaRefs.Response, "record": descriptor.SchemaRefs.Record,
		} {
			if reference == "" {
				continue
			}
			if _, ok := schemas[reference]; !ok {
				return vNextCanonicalDescriptor{}, fmt.Errorf("operation %q %s schema reference %q is missing", descriptor.ID, role, reference)
			}
		}
		for lane, raw := range map[string]json.RawMessage{"stream": descriptor.Stream, "write": descriptor.Write, "operation": descriptor.Operation} {
			if len(raw) != 0 {
				if err := validateRawJSONObject("operation "+descriptor.ID+" "+lane, raw); err != nil {
					return vNextCanonicalDescriptor{}, err
				}
				if err := rejectVNextLegacyExecutionEvidence("operation "+descriptor.ID+" "+lane, raw); err != nil {
					return vNextCanonicalDescriptor{}, err
				}
			}
		}
		for commandIndex := range descriptor.Commands {
			if err := validateRawJSONObject(fmt.Sprintf("operation %s command %d", descriptor.ID, commandIndex), descriptor.Commands[commandIndex].Command); err != nil {
				return vNextCanonicalDescriptor{}, err
			}
			if err := rejectVNextLegacyExecutionEvidence(fmt.Sprintf("operation %s command %d", descriptor.ID, commandIndex), descriptor.Commands[commandIndex].Command); err != nil {
				return vNextCanonicalDescriptor{}, err
			}
			descriptor.Commands[commandIndex].Command = cloneRawJSON(descriptor.Commands[commandIndex].Command)
		}
		descriptor.Source = cloneRawJSON(descriptor.Source)
		descriptor.Stream = cloneRawJSON(descriptor.Stream)
		descriptor.Write = cloneRawJSON(descriptor.Write)
		descriptor.Operation = cloneRawJSON(descriptor.Operation)
	}
	execution := make(map[string]json.RawMessage, len(lock.Execution))
	for name, raw := range lock.Execution {
		if !isVNextOptionalExecutionFile(name) {
			return vNextCanonicalDescriptor{}, fmt.Errorf("unsupported execution artifact %q", name)
		}
		if err := validateRawJSONObject(name, raw); err != nil {
			return vNextCanonicalDescriptor{}, err
		}
		if err := rejectVNextLegacyExecutionEvidence(name, raw); err != nil {
			return vNextCanonicalDescriptor{}, err
		}
		execution[name] = cloneRawJSON(raw)
	}
	if len(lock.CLI) != 0 {
		if err := validateRawJSONObject("cli", lock.CLI); err != nil {
			return vNextCanonicalDescriptor{}, err
		}
		var cli map[string]json.RawMessage
		if err := json.Unmarshal(lock.CLI, &cli); err != nil {
			return vNextCanonicalDescriptor{}, err
		}
		// Provider CLI references are authoring evidence only. They stay in the
		// immutable source lock and never enter the runtime execution bundle.
		delete(cli, "source_cli")
		delete(cli, "destination_cli")
		lock.CLI, _ = json.Marshal(cli)
	}
	return vNextCanonicalDescriptor{
		Connector: lock.Connector, Metadata: cloneRawJSON(lock.Metadata), ConfigSchema: cloneRawJSON(lock.ConfigSchema), HTTP: cloneRawJSON(lock.HTTP),
		Schemas: schemas, Operations: operations, CLI: cloneRawJSON(lock.CLI), Execution: execution,
	}, nil
}

func renderVNextExecutionBundle(descriptor vNextCanonicalDescriptor) (map[string][]byte, error) {
	outputs := make(map[string][]byte, 8+len(descriptor.Schemas)+len(descriptor.Execution))
	var err error
	if outputs["metadata.json"], err = renderRawJSON(descriptor.Metadata); err != nil {
		return nil, err
	}
	if outputs["spec.json"], err = renderRawJSON(descriptor.ConfigSchema); err != nil {
		return nil, err
	}
	for name, schema := range descriptor.Schemas {
		if outputs[name], err = renderRawJSON(schema); err != nil {
			return nil, err
		}
	}
	streams := orderedVNextLane(descriptor.Operations, func(operation vNextOperationDescriptor) (int, json.RawMessage) {
		return operation.StreamOrder, operation.Stream
	})
	if len(descriptor.HTTP) != 0 || len(streams) != 0 {
		if outputs["streams.json"], err = renderJSONObject(map[string]any{"base": rawJSONValue(descriptor.HTTP), "streams": rawJSONValues(streams)}); err != nil {
			return nil, err
		}
	}
	writes := orderedVNextLane(descriptor.Operations, func(operation vNextOperationDescriptor) (int, json.RawMessage) {
		return operation.WriteOrder, operation.Write
	})
	if len(writes) != 0 {
		if outputs["writes.json"], err = renderJSONObject(map[string]any{"actions": rawJSONValues(writes)}); err != nil {
			return nil, err
		}
	}
	operations := orderedVNextLane(descriptor.Operations, func(operation vNextOperationDescriptor) (int, json.RawMessage) {
		return operation.OperationOrder, operation.Operation
	})
	if len(operations) != 0 {
		if outputs["operations.json"], err = renderJSONObject(map[string]any{"operations": rawJSONValues(operations)}); err != nil {
			return nil, err
		}
	}
	commands := make([]vNextCommandDescriptor, 0)
	for _, operation := range descriptor.Operations {
		commands = append(commands, operation.Commands...)
	}
	if len(descriptor.CLI) != 0 || len(commands) != 0 {
		sort.SliceStable(commands, func(left, right int) bool { return commands[left].Order < commands[right].Order })
		var cli map[string]any
		if len(descriptor.CLI) != 0 {
			if err := json.Unmarshal(descriptor.CLI, &cli); err != nil {
				return nil, err
			}
		} else {
			cli = make(map[string]any)
		}
		commandValues := make([]any, 0, len(commands))
		for _, command := range commands {
			commandValues = append(commandValues, rawJSONValue(command.Command))
		}
		cli["commands"] = commandValues
		if outputs["cli_surface.json"], err = renderJSONObject(cli); err != nil {
			return nil, err
		}
	}
	for name, raw := range descriptor.Execution {
		if outputs[name], err = renderRawJSON(raw); err != nil {
			return nil, err
		}
	}
	return outputs, nil
}

func observedVNextLanes(lock vNextSourceLock) map[string]bool {
	lanes := make(map[string]bool, len(vNextLaneNames))
	for _, operation := range lock.Operations {
		if len(operation.Stream) != 0 {
			lanes["etl"] = true
		}
		for _, command := range operation.Commands {
			intent, _ := vNextRawOptionalString(command.Command, "intent")
			if slicesContains(vNextLaneNames, intent) {
				lanes[intent] = true
			}
		}
	}
	if len(lock.Execution["sync_transport.json"]) != 0 {
		lanes["sync_transport"] = true
	}
	return lanes
}

func slicesContains(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func orderedVNextLane(operations []vNextOperationDescriptor, lane func(vNextOperationDescriptor) (int, json.RawMessage)) []json.RawMessage {
	type item struct {
		order int
		raw   json.RawMessage
	}
	items := make([]item, 0, len(operations))
	for _, operation := range operations {
		order, raw := lane(operation)
		if len(raw) != 0 {
			items = append(items, item{order: order, raw: raw})
		}
	}
	sort.SliceStable(items, func(left, right int) bool { return items[left].order < items[right].order })
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		out = append(out, item.raw)
	}
	return out
}

func rejectVNextLegacyExecutionEvidence(name string, raw json.RawMessage) error {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	if key := findVNextLegacyEvidenceKey(value); key != "" {
		return fmt.Errorf("%s contains legacy runtime evidence field %q", name, key)
	}
	return nil
}

func findVNextLegacyEvidenceKey(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"conformance", "source_operation", "source_cli_path"} {
			if _, exists := typed[key]; exists {
				return key
			}
		}
		for _, child := range typed {
			if key := findVNextLegacyEvidenceKey(child); key != "" {
				return key
			}
		}
	case []any:
		for _, child := range typed {
			if key := findVNextLegacyEvidenceKey(child); key != "" {
				return key
			}
		}
	}
	return ""
}

func vNextRawArray(raw json.RawMessage, key string) ([]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	var values []json.RawMessage
	if err := json.Unmarshal(object[key], &values); err != nil {
		return nil, fmt.Errorf("%s must be an array: %w", key, err)
	}
	return values, nil
}

func vNextRawString(raw json.RawMessage, key string) (string, error) {
	value, err := vNextRawOptionalString(raw, key)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func vNextRawOptionalString(raw json.RawMessage, key string) (string, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return "", err
	}
	valueRaw := object[key]
	if len(valueRaw) == 0 {
		return "", nil
	}
	var value string
	if err := json.Unmarshal(valueRaw, &value); err != nil {
		return "", fmt.Errorf("%s must be a string", key)
	}
	return value, nil
}

func validateRawJSONObject(name string, raw json.RawMessage) error {
	if len(raw) == 0 || !json.Valid(raw) {
		return fmt.Errorf("%s must be valid JSON", name)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return fmt.Errorf("%s must be a JSON object", name)
	}
	return nil
}

func validVNextSchemaPath(name string) bool {
	return strings.HasPrefix(name, "schemas/") && strings.HasSuffix(name, ".json") && path.Clean(name) == name && !strings.Contains(name, "..")
}

func isVNextOptionalExecutionFile(name string) bool {
	switch name {
	case "changefeed.json", "polling_watermark.json", "sync_transport.json", "rate_limits.json", "database.json":
		return true
	default:
		return false
	}
}

func rawJSONValue(raw json.RawMessage) any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var value any
	_ = json.Unmarshal(raw, &value)
	return value
}

func rawJSONValues(raws []json.RawMessage) []any {
	values := make([]any, 0, len(raws))
	for _, raw := range raws {
		values = append(values, rawJSONValue(raw))
	}
	return values
}

func renderRawJSON(raw json.RawMessage) ([]byte, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return renderJSONObject(value)
}

func renderJSONObject(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func executionBundlesEqual(left, right map[string][]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for name, leftRaw := range left {
		if !bytes.Equal(leftRaw, right[name]) {
			return false
		}
	}
	return true
}

func cloneRawJSON(raw json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), raw...)
}
