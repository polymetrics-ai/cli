package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestidentity"
)

// vNextCanonicalGraph is the typed authoring graph built from one strict
// schema-4 source lock. It supports in-memory semantic admission only; it
// never constructs a registry connection, credential, transport, or provider
// request.
type vNextCanonicalGraph struct {
	ProviderEvidence json.RawMessage
	AuthoredCLI      json.RawMessage
	Lanes            map[string]string
	Metadata         engine.Metadata
	HTTP             *engine.HTTPBase
	Config           *engine.Schema
	CLI              *engine.CLISurface
	Schemas          map[string]*engine.Schema

	Operations []vNextCanonicalOperation
	Identity   manifestidentity.Identity
}

// vNextCanonicalOperation keeps immutable source identity separate from typed
// lane nodes. Index is the authored array position for diagnostics;
// CanonicalIndex is assigned after source-ID sorting for staged provenance.
type vNextCanonicalOperation struct {
	Index          int
	CanonicalIndex int
	ID             string
	Source         json.RawMessage
	SchemaRefs     vNextSchemaReferences
	Stream         *vNextCanonicalStream
	Write          *vNextCanonicalWrite
	Operation      *vNextCanonicalRuntimeOperation
	Commands       []vNextCanonicalCommand
}

type vNextCanonicalStream struct {
	Order int
	Raw   json.RawMessage
	Spec  engine.StreamSpec
}

type vNextCanonicalWrite struct {
	Order int
	Raw   json.RawMessage
	Spec  engine.WriteAction
}

type vNextCanonicalRuntimeOperation struct {
	Order int
	Raw   json.RawMessage
	Spec  engine.OperationSpec
}

type vNextCanonicalCommand struct {
	Index int
	Order int
	Raw   json.RawMessage
	Spec  engine.CLICommand
}

func buildVNextCanonicalGraph(descriptor vNextCanonicalDescriptor, providerEvidence, authoredCLI json.RawMessage) (vNextCanonicalGraph, error) {
	graph := vNextCanonicalGraph{Lanes: make(map[string]string, len(descriptor.Lanes)), Schemas: make(map[string]*engine.Schema, len(descriptor.Schemas))}
	for lane, state := range descriptor.Lanes {
		graph.Lanes[lane] = state
	}
	if len(authoredCLI) != 0 {
		if err := validateRawJSONObject("cli", authoredCLI); err != nil {
			return vNextCanonicalGraph{}, vNextGraphError("/cli", err)
		}
		graph.AuthoredCLI = cloneRawJSON(authoredCLI)
	}
	if len(providerEvidence) != 0 {
		if err := validateRawJSONObject("provider_evidence", providerEvidence); err != nil {
			return vNextCanonicalGraph{}, vNextGraphError("/provider_evidence", err)
		}
		graph.ProviderEvidence = cloneRawJSON(providerEvidence)
	}
	if err := decodeStrictJSON(descriptor.Metadata, &graph.Metadata); err != nil {
		return vNextCanonicalGraph{}, vNextGraphError("/metadata", err)
	}
	config, err := engine.CompileSchema(descriptor.ConfigSchema)
	if err != nil {
		return vNextCanonicalGraph{}, vNextGraphError("/config_schema", err)
	}
	graph.Config = config
	if len(descriptor.HTTP) != 0 {
		var http engine.HTTPBase
		if err := decodeStrictJSON(descriptor.HTTP, &http); err != nil {
			return vNextCanonicalGraph{}, vNextGraphError("/http", err)
		}
		graph.HTTP = &http
	}
	if len(descriptor.CLI) != 0 {
		var cli engine.CLISurface
		if err := decodeStrictJSON(descriptor.CLI, &cli); err != nil {
			return vNextCanonicalGraph{}, vNextGraphError("/cli", err)
		}
		graph.CLI = &cli
	}
	schemaNames := make([]string, 0, len(descriptor.Schemas))
	for name := range descriptor.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)
	for _, name := range schemaNames {
		raw := descriptor.Schemas[name]
		schema, err := engine.CompileSchema(raw)
		if err != nil {
			return vNextCanonicalGraph{}, vNextGraphError("/schemas/"+vNextJSONPointerSegment(name), err)
		}
		graph.Schemas[name] = schema
	}
	for index, authored := range descriptor.Operations {
		operation, err := buildVNextCanonicalOperation(index, authored)
		if err != nil {
			return vNextCanonicalGraph{}, err
		}
		if err := validateVNextSchemaRoles(operation, descriptor.Schemas); err != nil {
			return vNextCanonicalGraph{}, err
		}
		graph.Operations = append(graph.Operations, operation)
	}
	sort.Slice(graph.Operations, func(left, right int) bool {
		return graph.Operations[left].ID < graph.Operations[right].ID
	})
	for index := range graph.Operations {
		graph.Operations[index].CanonicalIndex = index
	}
	if err := validateVNextCanonicalAliases(graph.Operations); err != nil {
		return vNextCanonicalGraph{}, err
	}
	return graph, nil
}

func buildVNextCanonicalOperation(index int, authored vNextOperationDescriptor) (vNextCanonicalOperation, error) {
	operation := vNextCanonicalOperation{
		Index:      index,
		ID:         authored.ID,
		SchemaRefs: authored.SchemaRefs,
		Source:     cloneRawJSON(authored.Source),
	}
	if len(operation.Source) != 0 {
		if err := validateRawJSONObject("source", operation.Source); err != nil {
			return vNextCanonicalOperation{}, vNextGraphError(vNextOperationPointer(index, "source"), err)
		}
	}
	if len(authored.Stream) != 0 {
		var stream engine.StreamSpec
		if err := decodeStrictJSON(authored.Stream, &stream); err != nil {
			return vNextCanonicalOperation{}, vNextGraphError(vNextOperationPointer(index, "stream"), err)
		}
		operation.Stream = &vNextCanonicalStream{Order: authored.StreamOrder, Raw: cloneRawJSON(authored.Stream), Spec: stream}
	}
	if len(authored.Write) != 0 {
		var write engine.WriteAction
		if err := decodeStrictJSON(authored.Write, &write); err != nil {
			return vNextCanonicalOperation{}, vNextGraphError(vNextOperationPointer(index, "write"), err)
		}
		operation.Write = &vNextCanonicalWrite{Order: authored.WriteOrder, Raw: cloneRawJSON(authored.Write), Spec: write}
	}
	if len(authored.Operation) != 0 {
		var runtimeOperation engine.OperationSpec
		if err := decodeStrictJSON(authored.Operation, &runtimeOperation); err != nil {
			return vNextCanonicalOperation{}, vNextGraphError(vNextOperationPointer(index, "operation"), err)
		}
		operation.Operation = &vNextCanonicalRuntimeOperation{Order: authored.OperationOrder, Raw: cloneRawJSON(authored.Operation), Spec: runtimeOperation}
	}
	operation.Commands = make([]vNextCanonicalCommand, 0, len(authored.Commands))
	for commandIndex, authoredCommand := range authored.Commands {
		var command engine.CLICommand
		if err := decodeStrictJSON(authoredCommand.Command, &command); err != nil {
			return vNextCanonicalOperation{}, vNextGraphError(vNextOperationPointer(index, "commands", fmt.Sprint(commandIndex)), err)
		}
		operation.Commands = append(operation.Commands, vNextCanonicalCommand{
			Index: commandIndex,
			Order: authoredCommand.Order,
			Raw:   cloneRawJSON(authoredCommand.Command),
			Spec:  command,
		})
	}
	return operation, nil
}

func validateVNextSchemaRoles(operation vNextCanonicalOperation, schemas map[string]json.RawMessage) error {
	if reference := operation.SchemaRefs.Request; reference != "" {
		if operation.Write == nil && operation.Operation == nil {
			return vNextGraphError(vNextOperationPointer(operation.Index, "schema_refs", "request"), fmt.Errorf("request schema %q has no write or operation binding", reference))
		}
		if operation.Write != nil && operation.Operation == nil && len(operation.Write.Spec.RecordSchema) == 0 {
			return vNextGraphError(vNextOperationPointer(operation.Index, "schema_refs", "request"), fmt.Errorf("request schema %q has no write record_schema binding", reference))
		}
	}
	if reference := operation.SchemaRefs.Response; reference != "" && operation.Stream == nil && operation.Operation == nil {
		return vNextGraphError(vNextOperationPointer(operation.Index, "schema_refs", "response"), fmt.Errorf("response schema %q has no stream or operation binding", reference))
	}
	if reference := operation.SchemaRefs.Record; reference != "" {
		if operation.Stream == nil {
			return vNextGraphError(vNextOperationPointer(operation.Index, "schema_refs", "record"), fmt.Errorf("record schema %q has no stream binding", reference))
		}
		if operation.Stream.Spec.SchemaRef != reference {
			return vNextGraphError(vNextOperationPointer(operation.Index, "schema_refs", "record"), fmt.Errorf("record schema %q does not bind stream schema %q", reference, operation.Stream.Spec.SchemaRef))
		}
	}
	for _, role := range []struct {
		name      string
		reference string
	}{
		{name: "request", reference: operation.SchemaRefs.Request},
		{name: "response", reference: operation.SchemaRefs.Response},
		{name: "record", reference: operation.SchemaRefs.Record},
	} {
		if role.reference == "" {
			continue
		}
		if _, exists := schemas[role.reference]; !exists {
			return vNextGraphError(vNextOperationPointer(operation.Index, "schema_refs", role.name), fmt.Errorf("%s schema reference %q is missing", role.name, role.reference))
		}
	}
	return nil
}

func validateVNextCanonicalAliases(operations []vNextCanonicalOperation) error {
	commands := make(map[string]vNextCanonicalCommand)
	for _, operation := range operations {
		for _, command := range operation.Commands {
			alias := vNextCommandAlias(command.Spec.Path)
			if alias == "" {
				return vNextGraphError(vNextOperationPointer(operation.Index, "commands", fmt.Sprint(command.Index), "path"), fmt.Errorf("command path is required"))
			}
			if previous, exists := commands[alias]; exists {
				return vNextGraphError(vNextOperationPointer(operation.Index, "commands", fmt.Sprint(command.Index), "path"), fmt.Errorf("command path %q aliases command at index %d", command.Spec.Path, previous.Index))
			}
			commands[alias] = command
		}
	}
	return nil
}

func vNextCommandAlias(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func validateVNextCanonicalGraph(descriptor *vNextCanonicalDescriptor) error {
	stage, err := admitVNextCanonicalDescriptor(*descriptor, vNextSemanticAdmissionInput{})
	if err != nil {
		return err
	}
	descriptor.Graph.Identity = stage.Identity
	descriptor.Staged = stage
	return nil
}

func vNextStaticValidationPointer(descriptor vNextCanonicalDescriptor, message string) string {
	type collection struct {
		file       string
		name       string
		lane       string
		operations []int
	}
	collections := []collection{
		{file: "streams.json", name: "streams", lane: "stream", operations: vNextOrderedLaneOperationIndexes(descriptor.Operations, func(operation vNextOperationDescriptor) (int, json.RawMessage) {
			return operation.StreamOrder, operation.Stream
		})},
		{file: "writes.json", name: "actions", lane: "write", operations: vNextOrderedLaneOperationIndexes(descriptor.Operations, func(operation vNextOperationDescriptor) (int, json.RawMessage) {
			return operation.WriteOrder, operation.Write
		})},
		{file: "operations.json", name: "operations", lane: "operation", operations: vNextOrderedLaneOperationIndexes(descriptor.Operations, func(operation vNextOperationDescriptor) (int, json.RawMessage) {
			return operation.OperationOrder, operation.Operation
		})},
	}
	for _, candidate := range collections {
		index, suffix, found := vNextStaticValidationIndex(message, candidate.file, candidate.name)
		if !found || index >= len(candidate.operations) {
			continue
		}
		return vNextOperationPointer(candidate.operations[index], candidate.lane) + suffix
	}
	index, suffix, found := vNextStaticValidationIndex(message, "cli_surface.json", "commands")
	if found {
		commands := vNextOrderedCommandOperationIndexes(descriptor.Operations)
		if index < len(commands) {
			return vNextOperationPointer(commands[index].operation, "commands", fmt.Sprint(commands[index].command)) + suffix
		}
	}
	return "/operations"
}

func vNextStaticValidationIndex(message, file, collection string) (int, string, bool) {
	marker := file + ": /" + collection + "/"
	start := strings.Index(message, marker)
	if start < 0 {
		return 0, "", false
	}
	value := message[start+len(marker):]
	indexText, suffix, nested := strings.Cut(value, "/")
	if !nested {
		indexText, suffix, _ = strings.Cut(value, ":")
	}
	index, err := strconv.Atoi(indexText)
	if err != nil {
		return 0, "", false
	}
	if suffix == "" {
		return index, "", true
	}
	if colon := strings.IndexByte(suffix, ':'); colon >= 0 {
		suffix = suffix[:colon]
	}
	return index, "/" + suffix, true
}

func vNextOrderedLaneOperationIndexes(operations []vNextOperationDescriptor, lane func(vNextOperationDescriptor) (int, json.RawMessage)) []int {
	type item struct {
		index    int
		order    int
		sourceID string
	}
	items := make([]item, 0, len(operations))
	for index, operation := range operations {
		order, raw := lane(operation)
		if len(raw) != 0 {
			items = append(items, item{index: index, order: order, sourceID: operation.ID})
		}
	}
	sort.Slice(items, func(left, right int) bool {
		if items[left].order != items[right].order {
			return items[left].order < items[right].order
		}
		return items[left].sourceID < items[right].sourceID
	})
	indexes := make([]int, len(items))
	for index, item := range items {
		indexes[index] = item.index
	}
	return indexes
}

type vNextCommandOperationIndex struct {
	operation int
	command   int
	order     int
	path      string
	sourceID  string
}

func vNextOrderedCommandOperationIndexes(operations []vNextOperationDescriptor) []vNextCommandOperationIndex {
	items := make([]vNextCommandOperationIndex, 0)
	for operationIndex, operation := range operations {
		for commandIndex, command := range operation.Commands {
			commandPath, _ := vNextRawOptionalString(command.Command, "path")
			items = append(items, vNextCommandOperationIndex{
				operation: operationIndex,
				command:   commandIndex,
				order:     command.Order,
				path:      vNextCommandAlias(commandPath),
				sourceID:  operation.ID,
			})
		}
	}
	sort.Slice(items, func(left, right int) bool {
		if items[left].order != items[right].order {
			return items[left].order < items[right].order
		}
		if items[left].path != items[right].path {
			return items[left].path < items[right].path
		}
		return items[left].sourceID < items[right].sourceID
	})
	return items
}

func vNextGraphError(pointer string, err error) error {
	return fmt.Errorf("source lock %s: %w", pointer, err)
}

func vNextOperationPointer(index int, segments ...string) string {
	pointer := fmt.Sprintf("/operations/%d", index)
	for _, segment := range segments {
		pointer += "/" + vNextJSONPointerSegment(segment)
	}
	return pointer
}

func vNextJSONPointerSegment(value string) string {
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(value)
}

// vNextExecutionFS gives the static engine loader a read-only in-memory view
// of rendered bytes. It retains the caller-owned byte slices and performs no
// filesystem write or copy while validating a source graph.
type vNextExecutionFS struct {
	files map[string][]byte
}

func newVNextExecutionFS(connector string, outputs map[string][]byte) vNextExecutionFS {
	files := make(map[string][]byte, len(outputs))
	for name, contents := range outputs {
		files[path.Join(connector, name)] = contents
	}
	return vNextExecutionFS{files: files}
}

func (fsys vNextExecutionFS) Open(name string) (fs.File, error) {
	info, err := fsys.Stat(name)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return &vNextExecutionFile{Reader: bytes.NewReader(nil), info: info}, nil
	}
	return &vNextExecutionFile{Reader: bytes.NewReader(fsys.files[name]), info: info}, nil
}

func (fsys vNextExecutionFS) Stat(name string) (fs.FileInfo, error) {
	if name != "." && !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}
	if contents, exists := fsys.files[name]; exists {
		return vNextExecutionInfo{name: path.Base(name), size: int64(len(contents))}, nil
	}
	if name == "." || fsys.isDirectory(name) {
		return vNextExecutionInfo{name: path.Base(name), directory: true}, nil
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

func (fsys vNextExecutionFS) ReadDir(name string) ([]fs.DirEntry, error) {
	info, err := fsys.Stat(name)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	prefix := ""
	if name != "." {
		prefix = name + "/"
	}
	entries := map[string]vNextExecutionInfo{}
	for filename, contents := range fsys.files {
		if !strings.HasPrefix(filename, prefix) {
			continue
		}
		remainder := strings.TrimPrefix(filename, prefix)
		entryName, _, nested := strings.Cut(remainder, "/")
		if entryName == "" {
			continue
		}
		if nested {
			entries[entryName] = vNextExecutionInfo{name: entryName, directory: true}
			continue
		}
		entries[entryName] = vNextExecutionInfo{name: entryName, size: int64(len(contents))}
	}
	names := make([]string, 0, len(entries))
	for entryName := range entries {
		names = append(names, entryName)
	}
	sort.Strings(names)
	result := make([]fs.DirEntry, 0, len(names))
	for _, entryName := range names {
		result = append(result, vNextExecutionEntry{info: entries[entryName]})
	}
	return result, nil
}

func (fsys vNextExecutionFS) isDirectory(name string) bool {
	prefix := name
	if prefix != "" && prefix != "." {
		prefix += "/"
	} else {
		prefix = ""
	}
	for filename := range fsys.files {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}
	return false
}

type vNextExecutionFile struct {
	*bytes.Reader
	info fs.FileInfo
}

func (file *vNextExecutionFile) Close() error {
	return nil
}

func (file *vNextExecutionFile) Stat() (fs.FileInfo, error) {
	return file.info, nil
}

type vNextExecutionInfo struct {
	name      string
	size      int64
	directory bool
}

func (info vNextExecutionInfo) Name() string { return info.name }
func (info vNextExecutionInfo) Size() int64  { return info.size }
func (info vNextExecutionInfo) Mode() fs.FileMode {
	if info.directory {
		return fs.ModeDir | 0o555
	}
	return 0o444
}
func (info vNextExecutionInfo) ModTime() time.Time { return time.Time{} }
func (info vNextExecutionInfo) IsDir() bool        { return info.directory }
func (info vNextExecutionInfo) Sys() any           { return nil }

type vNextExecutionEntry struct {
	info vNextExecutionInfo
}

func (entry vNextExecutionEntry) Name() string               { return entry.info.Name() }
func (entry vNextExecutionEntry) IsDir() bool                { return entry.info.IsDir() }
func (entry vNextExecutionEntry) Type() fs.FileMode          { return entry.info.Mode().Type() }
func (entry vNextExecutionEntry) Info() (fs.FileInfo, error) { return entry.info, nil }
