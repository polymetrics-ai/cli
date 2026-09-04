package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestidentity"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/syncplan"
)

// vNextStagedGeneration is the complete semantic-admission result for one
// source lock. It is an in-memory handoff to CP11; it never creates a staging
// directory or activates an execution generation.
type vNextStagedGeneration struct {
	Outputs    map[string][]byte
	Identity   manifestidentity.Identity
	Manifest   manifestindex.Entry
	Index      manifestindex.Index
	Provenance []vNextSourceExecutionProvenance
	Sync       []vNextResolvedSyncAdmission
}

// vNextSourceExecutionProvenance preserves the exact source field responsible
// for one rendered runtime identity. It is deterministic and authoring-only.
type vNextSourceExecutionProvenance struct {
	SourceID   string
	FieldPath  string
	TargetKind string
	TargetID   string
}

// vNextSemanticAdmissionInput supplies facts that cannot truthfully be
// inferred from a single source lock. In particular, saved synchronization
// requires an independently declared destination and Foundation Atlas fact.
type vNextSemanticAdmissionInput struct {
	Manifest *manifestindex.Entry
	Sync     []vNextSyncAdmission
}

type vNextSyncAdmission struct {
	SourceID  string
	FieldPath string
	Plan      syncplan.Plan
	Budget    synccontract.Budget
}

type vNextResolvedSyncAdmission struct {
	SourceID  string
	FieldPath string
	Result    syncplan.Result
}

func admitVNextCanonicalDescriptor(descriptor vNextCanonicalDescriptor, input vNextSemanticAdmissionInput) (vNextStagedGeneration, error) {
	outputs, err := renderVNextExecutionBundle(descriptor)
	if err != nil {
		return vNextStagedGeneration{}, err
	}
	bundle, err := engine.Load(newVNextExecutionFS(descriptor.Connector, outputs), descriptor.Connector)
	if err != nil {
		return vNextStagedGeneration{}, vNextGraphError(vNextStaticValidationPointer(descriptor, err.Error()), fmt.Errorf("static execution validation: %w", err))
	}

	runtime := engine.New(bundle, nil)
	provenance, err := vNextBuildSemanticProvenance(descriptor, outputs, bundle, runtime)
	if err != nil {
		return vNextStagedGeneration{}, err
	}
	executor, err := vNextSelectedExecutor(descriptor.Connector)
	if err != nil {
		return vNextStagedGeneration{}, vNextSemanticRootError(descriptor, "/connector", "select executor: %v", err)
	}
	manifest := vNextManifestEntry(bundle, runtime, executor)
	if input.Manifest != nil {
		if err := vNextValidateManifestBinding(descriptor, *input.Manifest, manifest); err != nil {
			return vNextStagedGeneration{}, err
		}
	}
	index, err := manifestindex.New([]manifestindex.Entry{manifest}, 1)
	if err != nil {
		return vNextStagedGeneration{}, vNextSemanticRootError(descriptor, "/connector", "manifest index admission: %v", err)
	}
	stage := vNextStagedGeneration{
		Outputs:    outputs,
		Identity:   bundle.Identity,
		Manifest:   manifest,
		Index:      index,
		Provenance: provenance,
	}
	if err := vNextValidateStagedGeneration(descriptor, stage); err != nil {
		return vNextStagedGeneration{}, err
	}
	syncAdmissions, err := vNextResolveSuppliedSyncAdmissions(descriptor, manifest, input.Sync)
	if err != nil {
		return vNextStagedGeneration{}, err
	}
	stage.Sync = syncAdmissions
	return stage, nil
}

func vNextBuildSemanticProvenance(descriptor vNextCanonicalDescriptor, outputs map[string][]byte, bundle engine.Bundle, runtime *engine.Connector) ([]vNextSourceExecutionProvenance, error) {
	streams := make(map[string]engine.StreamSpec, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		if _, duplicate := streams[stream.Name]; duplicate {
			return nil, vNextSemanticRootError(descriptor, "/operations", "loaded bundle duplicates stream %q", stream.Name)
		}
		streams[stream.Name] = stream
	}
	writes := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, write := range bundle.Writes {
		if _, duplicate := writes[write.Name]; duplicate {
			return nil, vNextSemanticRootError(descriptor, "/operations", "loaded bundle duplicates write action %q", write.Name)
		}
		writes[write.Name] = write
	}
	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		if _, duplicate := operations[operation.ID]; duplicate {
			return nil, vNextSemanticRootError(descriptor, "/operations", "loaded bundle duplicates operation %q", operation.ID)
		}
		operations[operation.ID] = operation
	}
	commands := make(map[string]engine.CLICommand)
	if bundle.CLISurface != nil {
		for _, command := range bundle.CLISurface.Commands {
			alias := vNextCommandAlias(command.Path)
			if _, duplicate := commands[alias]; duplicate {
				return nil, vNextSemanticRootError(descriptor, "/cli", "loaded bundle duplicates command path %q", command.Path)
			}
			commands[alias] = command
		}
	}
	runtimeCommands := make(map[string]connectors.CommandSurfaceCommand)
	if surface := runtime.CommandSurface(); surface != nil {
		for _, command := range surface.Commands {
			alias := vNextCommandAlias(command.Path)
			if _, duplicate := runtimeCommands[alias]; duplicate {
				return nil, vNextSemanticRootError(descriptor, "/cli", "runtime command surface duplicates command path %q", command.Path)
			}
			runtimeCommands[alias] = command
		}
	}

	provenance := make([]vNextSourceExecutionProvenance, 0, len(descriptor.Graph.Operations)*5+len(descriptor.Lanes))
	for _, lane := range vNextLaneNames {
		state := descriptor.Lanes[lane]
		provenance = append(provenance, vNextSourceExecutionProvenance{
			SourceID: "connector:" + descriptor.Connector, FieldPath: "/lanes/" + lane, TargetKind: "lane", TargetID: state,
		})
	}
	for _, source := range descriptor.Graph.Operations {
		if err := vNextValidateSourceSchemas(descriptor, outputs, source, &provenance); err != nil {
			return nil, err
		}
		if source.Stream != nil {
			loaded, found := streams[source.Stream.Spec.Name]
			if !found || !vNextJSONEquivalent(loaded, source.Stream.Spec) {
				return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "stream"), "stream %q does not exactly bind the loaded execution stream", source.Stream.Spec.Name)
			}
			provenance = append(provenance, vNextSourceExecutionProvenance{SourceID: source.ID, FieldPath: vNextOperationPointer(source.Index, "stream"), TargetKind: "stream", TargetID: source.Stream.Spec.Name})
		}
		if source.Write != nil {
			loaded, found := writes[source.Write.Spec.Name]
			if !found || !vNextJSONEquivalent(loaded, source.Write.Spec) {
				return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "write"), "write action %q does not exactly bind the loaded execution action", source.Write.Spec.Name)
			}
			provenance = append(provenance, vNextSourceExecutionProvenance{SourceID: source.ID, FieldPath: vNextOperationPointer(source.Index, "write"), TargetKind: "write", TargetID: source.Write.Spec.Name})
		}
		if source.Operation != nil {
			loaded, found := operations[source.Operation.Spec.ID]
			if !found || !vNextJSONEquivalent(loaded, source.Operation.Spec) {
				return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "operation"), "operation %q does not exactly bind the loaded execution operation", source.Operation.Spec.ID)
			}
			provenance = append(provenance, vNextSourceExecutionProvenance{SourceID: source.ID, FieldPath: vNextOperationPointer(source.Index, "operation"), TargetKind: "operation", TargetID: source.Operation.Spec.ID})
		}
		sourceFacts, err := vNextDecodeSourceFacts(source)
		if err != nil {
			return nil, err
		}
		if err := vNextValidateSourceFacts(source, sourceFacts); err != nil {
			return nil, err
		}
		for _, command := range source.Commands {
			alias := vNextCommandAlias(command.Spec.Path)
			loaded, found := commands[alias]
			if !found || !vNextJSONEquivalent(loaded, command.Spec) {
				return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index), "path"), "command path %q does not exactly bind the loaded command surface", command.Spec.Path)
			}
			if err := vNextValidateCommandParent(source, command); err != nil {
				return nil, err
			}
			if command.Spec.Availability == "implemented" {
				runtimeCommand, found := runtimeCommands[alias]
				if !found {
					return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index), "path"), "command path %q is absent from the runtime command surface", command.Spec.Path)
				}
				resolved, err := engine.ResolveImplementedCommandBinding(bundle, runtimeCommand)
				if err != nil {
					return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index)), "runtime binding resolution for command %q: %v", command.Spec.Path, err)
				}
				if err := commandrunner.Preflight(runtime, strings.Fields(command.Spec.Path)); err != nil {
					return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index)), "runtime preflight for command %q: %v", command.Spec.Path, err)
				}
				provenance = append(provenance, vNextSourceExecutionProvenance{SourceID: source.ID, FieldPath: vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index), "binding"), TargetKind: "binding:" + resolved.Binding.Kind, TargetID: resolved.Binding.ID})
			}
			if err := vNextValidateSourceCommandFacts(source, command, sourceFacts); err != nil {
				return nil, err
			}
			provenance = append(provenance, vNextSourceExecutionProvenance{SourceID: source.ID, FieldPath: vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index)), TargetKind: "command", TargetID: command.Spec.Path})
		}
	}
	sort.Slice(provenance, func(left, right int) bool {
		if provenance[left].SourceID != provenance[right].SourceID {
			return provenance[left].SourceID < provenance[right].SourceID
		}
		if provenance[left].FieldPath != provenance[right].FieldPath {
			return provenance[left].FieldPath < provenance[right].FieldPath
		}
		if provenance[left].TargetKind != provenance[right].TargetKind {
			return provenance[left].TargetKind < provenance[right].TargetKind
		}
		return provenance[left].TargetID < provenance[right].TargetID
	})
	return provenance, nil
}

func vNextValidateSourceSchemas(descriptor vNextCanonicalDescriptor, outputs map[string][]byte, source vNextCanonicalOperation, provenance *[]vNextSourceExecutionProvenance) error {
	for _, reference := range []struct {
		role string
		name string
	}{
		{role: "request", name: source.SchemaRefs.Request},
		{role: "response", name: source.SchemaRefs.Response},
		{role: "record", name: source.SchemaRefs.Record},
	} {
		if reference.name == "" {
			continue
		}
		if _, exists := descriptor.Schemas[reference.name]; !exists {
			return vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "schema_refs", reference.role), "%s schema %q is absent from the canonical registry", reference.role, reference.name)
		}
		if _, exists := outputs[reference.name]; !exists {
			return vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "schema_refs", reference.role), "%s schema %q is absent from the staged execution set", reference.role, reference.name)
		}
		*provenance = append(*provenance, vNextSourceExecutionProvenance{SourceID: source.ID, FieldPath: vNextOperationPointer(source.Index, "schema_refs", reference.role), TargetKind: "schema", TargetID: reference.name})
	}
	return nil
}

func vNextValidateCommandParent(source vNextCanonicalOperation, command vNextCanonicalCommand) error {
	pointer := vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index))
	bindings := 0
	if command.Spec.Stream != "" {
		bindings++
		if source.Stream == nil || source.Stream.Spec.Name != command.Spec.Stream {
			return vNextSemanticOperationError(source, pointer+"/stream", "command stream %q is not the source operation's stream binding", command.Spec.Stream)
		}
	}
	if command.Spec.Write != "" {
		bindings++
		if source.Write == nil || source.Write.Spec.Name != command.Spec.Write {
			return vNextSemanticOperationError(source, pointer+"/write", "command write action %q is not the source operation's write binding", command.Spec.Write)
		}
	}
	if command.Spec.Operation != "" {
		bindings++
		if source.Operation == nil || source.Operation.Spec.ID != command.Spec.Operation {
			return vNextSemanticOperationError(source, pointer+"/operation", "command operation %q is not the source operation's exact execution operation", command.Spec.Operation)
		}
	}
	if bindings > 1 {
		return vNextSemanticOperationError(source, pointer, "command has multiple runtime bindings")
	}
	return nil
}

type vNextSourceFacts map[string]json.RawMessage

func vNextValidateSourceFacts(source vNextCanonicalOperation, facts vNextSourceFacts) error {
	providerOperation, hasProviderOperation, err := vNextSourceFact(source, facts, "provider_operation")
	if err != nil {
		return err
	}
	method, hasMethod, err := vNextSourceFact(source, facts, "method")
	if err != nil {
		return err
	}
	path, hasPath, err := vNextSourceFact(source, facts, "path")
	if err != nil {
		return err
	}
	if source.Operation == nil {
		return nil
	}
	if source.Operation.Spec.GraphQL != nil && hasProviderOperation && strings.TrimSpace(providerOperation) != "" && source.Operation.Spec.GraphQL.OperationName != providerOperation {
		return vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "source", "provider_operation"), "provider operation %q does not match GraphQL operation %q", providerOperation, source.Operation.Spec.GraphQL.OperationName)
	}
	actualMethod, actualPath, hasEndpoint := vNextOperationEndpoint(source.Operation.Spec)
	if !hasEndpoint {
		return nil
	}
	if hasMethod && strings.TrimSpace(method) != "" && strings.ToUpper(method) != actualMethod {
		return vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "source", "method"), "source method %q does not match execution method %q", method, actualMethod)
	}
	if hasPath && strings.TrimSpace(path) != "" && vNextSourcePath(path) != vNextSourcePath(actualPath) {
		return vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "source", "path"), "source path %q does not match execution path %q", path, actualPath)
	}
	return nil
}

func vNextValidateSourceCommandFacts(source vNextCanonicalOperation, command vNextCanonicalCommand, facts vNextSourceFacts) error {
	if command.Spec.Unsupported == nil {
		return nil
	}
	providerOperation, hasProviderOperation, err := vNextSourceFact(source, facts, "provider_operation")
	if err != nil {
		return err
	}
	method, hasMethod, err := vNextSourceFact(source, facts, "method")
	if err != nil {
		return err
	}
	path, hasPath, err := vNextSourceFact(source, facts, "path")
	if err != nil {
		return err
	}
	target := command.Spec.Unsupported.Target
	base := vNextOperationPointer(source.Index, "commands", fmt.Sprint(command.Index), "unsupported_disposition", "target")
	if hasProviderOperation && strings.TrimSpace(providerOperation) != "" && target.ProviderOperationID != providerOperation {
		return vNextSemanticOperationError(source, base+"/operation_id", "unsupported command operation %q does not match provider operation %q", target.ProviderOperationID, providerOperation)
	}
	if hasMethod && strings.TrimSpace(method) != "" && strings.ToUpper(target.Method) != strings.ToUpper(method) {
		return vNextSemanticOperationError(source, base+"/method", "unsupported command method %q does not match source method %q", target.Method, method)
	}
	if hasPath && strings.TrimSpace(path) != "" && vNextSourcePath(target.Path) != vNextSourcePath(path) {
		return vNextSemanticOperationError(source, base+"/path", "unsupported command path %q does not match source path %q", target.Path, path)
	}
	return nil
}

func vNextDecodeSourceFacts(source vNextCanonicalOperation) (vNextSourceFacts, error) {
	if len(source.Source) == 0 {
		return nil, nil
	}
	var facts vNextSourceFacts
	if err := json.Unmarshal(source.Source, &facts); err != nil {
		return nil, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "source"), "decode source facts: %v", err)
	}
	return facts, nil
}

func vNextSourceFact(source vNextCanonicalOperation, facts vNextSourceFacts, field string) (string, bool, error) {
	raw, found := facts[field]
	if !found {
		return "", false, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false, vNextSemanticOperationError(source, vNextOperationPointer(source.Index, "source", field), "source fact must be a string")
	}
	return value, true, nil
}

func vNextOperationEndpoint(operation engine.OperationSpec) (method, path string, ok bool) {
	switch {
	case operation.REST != nil:
		return strings.ToUpper(operation.REST.Method), operation.REST.Path, true
	case operation.GraphQL != nil:
		return http.MethodPost, operation.GraphQL.Path, true
	case operation.Binary != nil:
		return strings.ToUpper(operation.Binary.Method), operation.Binary.Path, true
	default:
		return "", "", false
	}
}

func vNextSourcePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return "/" + strings.TrimLeft(value, "/")
}

func vNextSelectedExecutor(connector string) (string, error) {
	executors, err := generatedNativeExecutors()
	if err != nil {
		return "", err
	}
	if executor := executors[connector]; executor != "" {
		return executor, nil
	}
	return "api_engine.v1", nil
}

func vNextManifestEntry(bundle engine.Bundle, runtime *engine.Connector, executor string) manifestindex.Entry {
	entry := manifestindex.Entry{
		Connector:  bundle.Identity.Connector,
		Generation: bundle.Identity.Generation,
		Digest:     bundle.Identity.Digest,
		Executor:   executor,
		Metadata:   runtime.Metadata(),
		Bytes:      bundle.Identity.Bytes,
	}
	if bundle.CLISurface != nil {
		entry.CommandUsage = bundle.CLISurface.Usage
		entry.CommandTagline = bundle.CLISurface.Tagline
	}
	return entry
}

func vNextValidateManifestBinding(descriptor vNextCanonicalDescriptor, expected, actual manifestindex.Entry) error {
	for _, field := range []struct {
		name string
		got  string
		want string
	}{
		{name: "connector", got: expected.Connector, want: actual.Connector},
		{name: "generation", got: expected.Generation, want: actual.Generation},
		{name: "digest", got: expected.Digest, want: actual.Digest},
		{name: "executor", got: expected.Executor, want: actual.Executor},
		{name: "extension", got: expected.Extension, want: actual.Extension},
		{name: "command_usage", got: expected.CommandUsage, want: actual.CommandUsage},
		{name: "command_tagline", got: expected.CommandTagline, want: actual.CommandTagline},
	} {
		if field.got != field.want {
			return vNextSemanticRootError(descriptor, "/connector", "manifest %s %q does not match staged %q", field.name, field.got, field.want)
		}
	}
	if expected.Bytes != actual.Bytes {
		return vNextSemanticRootError(descriptor, "/connector", "manifest byte charge %d does not match staged %d", expected.Bytes, actual.Bytes)
	}
	if !reflect.DeepEqual(expected.Metadata, actual.Metadata) {
		return vNextSemanticRootError(descriptor, "/metadata", "manifest metadata does not match staged metadata")
	}
	return nil
}

func vNextJSONEquivalent(left, right any) bool {
	leftRaw, leftErr := json.Marshal(left)
	rightRaw, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	var leftValue, rightValue any
	if json.Unmarshal(leftRaw, &leftValue) != nil || json.Unmarshal(rightRaw, &rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func vNextValidateStagedGeneration(descriptor vNextCanonicalDescriptor, stage vNextStagedGeneration) error {
	if stage.Identity.Connector != descriptor.Connector || stage.Identity.Generation == "" || stage.Identity.Digest == "" || stage.Identity.Bytes <= 0 {
		return vNextSemanticRootError(descriptor, "/connector", "staged identity is incomplete or cross-connector")
	}
	if err := vNextValidateManifestBinding(descriptor, stage.Manifest, manifestindex.Entry{
		Connector: stage.Identity.Connector, Generation: stage.Identity.Generation, Digest: stage.Identity.Digest,
		Executor: stage.Manifest.Executor, Extension: stage.Manifest.Extension, CommandUsage: stage.Manifest.CommandUsage,
		CommandTagline: stage.Manifest.CommandTagline, Metadata: stage.Manifest.Metadata, Bytes: stage.Identity.Bytes,
	}); err != nil {
		return err
	}
	indexed, found := stage.Index.Lookup(descriptor.Connector)
	if !found || !reflect.DeepEqual(indexed, stage.Manifest) {
		return vNextSemanticRootError(descriptor, "/connector", "staged manifest index does not retain the exact connector entry")
	}
	seen := make(map[string]struct{}, len(stage.Provenance))
	for _, row := range stage.Provenance {
		if row.SourceID == "" || row.FieldPath == "" || row.TargetKind == "" || row.TargetID == "" {
			return vNextSemanticRootError(descriptor, "/operations", "staged provenance has an incomplete row")
		}
		key := row.SourceID + "\x00" + row.FieldPath + "\x00" + row.TargetKind + "\x00" + row.TargetID
		if _, duplicate := seen[key]; duplicate {
			return vNextSemanticRootError(descriptor, "/operations", "staged provenance duplicates %q", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func vNextResolveSuppliedSyncAdmissions(descriptor vNextCanonicalDescriptor, manifest manifestindex.Entry, admissions []vNextSyncAdmission) ([]vNextResolvedSyncAdmission, error) {
	if len(admissions) == 0 {
		return nil, nil
	}
	bySource := make(map[string]vNextCanonicalOperation, len(descriptor.Graph.Operations))
	for _, operation := range descriptor.Graph.Operations {
		bySource[operation.ID] = operation
	}
	order := make([]int, len(admissions))
	for index := range admissions {
		order[index] = index
	}
	sort.Slice(order, func(left, right int) bool {
		first, second := admissions[order[left]], admissions[order[right]]
		if first.SourceID != second.SourceID {
			return first.SourceID < second.SourceID
		}
		return first.FieldPath < second.FieldPath
	})
	out := make([]vNextResolvedSyncAdmission, 0, len(admissions))
	seen := make(map[string]struct{}, len(admissions))
	for _, index := range order {
		admission := admissions[index]
		source, found := bySource[admission.SourceID]
		if !found {
			return nil, fmt.Errorf("source operation %q field %s: supplied sync admission names no canonical source operation", admission.SourceID, admission.FieldPath)
		}
		pointer := admission.FieldPath
		if pointer == "" {
			pointer = vNextOperationPointer(source.Index, "stream")
		}
		key := admission.SourceID + "\x00" + pointer
		if _, duplicate := seen[key]; duplicate {
			return nil, vNextSemanticOperationError(source, pointer, "duplicate supplied sync admission")
		}
		seen[key] = struct{}{}
		if source.Stream == nil || admission.Plan.Source.Kind != synccontract.BindingKindStream || admission.Plan.Source.ID != source.Stream.Spec.Name {
			return nil, vNextSemanticOperationError(source, pointer, "sync source binding %s/%q does not match the source operation stream", admission.Plan.Source.Kind, admission.Plan.Source.ID)
		}
		if !vNextPlanUsesManifestSource(admission.Plan, manifest) {
			return nil, vNextSemanticOperationError(source, pointer, "sync source executor does not match staged manifest executor and digest")
		}
		if admission.Plan.GenerationDigest != manifest.Digest {
			return nil, vNextSemanticOperationError(source, pointer, "sync generation digest does not match staged manifest digest")
		}
		result := syncplan.Resolve(admission.Plan, admission.Budget)
		if result.Kind != syncplan.ResultKindExecutable {
			return nil, vNextSemanticOperationError(source, pointer, "sync resolver did not admit supplied source/destination/Atlas facts: %s", result.Kind)
		}
		out = append(out, vNextResolvedSyncAdmission{SourceID: admission.SourceID, FieldPath: pointer, Result: result})
	}
	return out, nil
}

func vNextPlanUsesManifestSource(plan syncplan.Plan, manifest manifestindex.Entry) bool {
	for _, executor := range plan.Executors {
		if executor.Role == syncplan.ExecutorRoleSource && executor.ID == manifest.Executor && executor.Digest == manifest.Digest {
			return true
		}
	}
	return false
}

func vNextSemanticOperationError(source vNextCanonicalOperation, pointer, format string, args ...any) error {
	return fmt.Errorf("source operation %q field %s: %s", source.ID, pointer, fmt.Sprintf(format, args...))
}

func vNextSemanticRootError(descriptor vNextCanonicalDescriptor, pointer, format string, args ...any) error {
	return fmt.Errorf("source connector %q field %s: %s", descriptor.Connector, pointer, fmt.Sprintf(format, args...))
}
