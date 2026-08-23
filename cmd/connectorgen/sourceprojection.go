package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"unicode"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	sourceProjectionDefaultStringBytes               = 8 << 10
	sourceProjectionDefaultArrayItems                = 256
	sourceProjectionDefaultObjectProperties          = 256
	sourceOperationExecutionFoundation               = "closed-source-operation-execution-foundation-r1"
	sourceNonExecutableMutationDispositionFoundation = "source-cited-non-executable-mutation-foundation-r1"
	// sourceReadOnlyOperationFoundation is intentionally distinct from the
	// mutation disposition. A read-only declaration can never satisfy mutation
	// coverage, even when its endpoint currently lacks an executable action.
	sourceReadOnlyOperationFoundation = "source-read-only-operation-foundation-r1"
	sourceReadOnlyOperationModel      = "read_only"
	sourceReadOnlyPolicy              = "source-cited-read-only-operations-r1"
	// JSON-valued command flags carry a complete named field through the
	// declaration-owned body path. They are not an unbounded replacement for a
	// request body, so keep their encoded input explicitly bounded.
	sourceProjectionDefaultJSONBytes = 1 << 20
)

type sourceReadOnlyDeclaration struct {
	Policy string
	Reason string
}

var (
	sourceProjectionTemplateRE       = regexp.MustCompile(`\{\{\s*(?:config|record)\.([-A-Za-z0-9_]+)\s*\}\}`)
	sourceProjectionRecordTemplateRE = regexp.MustCompile(`\{\{\s*record\.([-A-Za-z0-9_]+)\s*\}\}`)
	sourceProjectionConfigTemplateRE = regexp.MustCompile(`\{\{\s*config\.([-A-Za-z0-9_]+)\s*\}\}`)
	sourceProjectionPathVariableRE   = regexp.MustCompile(`\{([-A-Za-z0-9_]+)\}`)
)

type sourceProjectionStats struct {
	Writes  int
	CLI     int
	Surface int
	Missing int
}

func (s sourceProjectionStats) Changed() bool { return s.Writes+s.CLI+s.Surface+s.Missing > 0 }

type sourceActionContract struct {
	Fields           map[string]any
	BareStringFields map[string]bool
	SecretFields     map[string]bool
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
	apiPath := filepath.Join(bundleDir, "api_surface.json")
	apiRaw, err := os.ReadFile(apiPath)
	if err != nil && !os.IsNotExist(err) {
		return sourceProjectionStats{}, err
	}
	var api orderedJSON
	if err == nil {
		if err := json.Unmarshal(apiRaw, &api); err != nil {
			return sourceProjectionStats{}, fmt.Errorf("api_surface.json: %w", err)
		}
	}
	declarationBundle := engine.Bundle{}
	if len(result.Operations) > 0 {
		declarationBundle.Name = result.Operations[0].Connector
	}
	if api.root != nil {
		raw, marshalErr := marshalNoEscapeHTML(api.root)
		if marshalErr != nil {
			return sourceProjectionStats{}, fmt.Errorf("encode api_surface.json: %w", marshalErr)
		}
		var surface engine.APISurface
		if unmarshalErr := json.Unmarshal(raw, &surface); unmarshalErr != nil {
			return sourceProjectionStats{}, fmt.Errorf("parse api_surface.json: %w", unmarshalErr)
		}
		declarationBundle.Surface = &surface
	}
	spec, err := sourceProjectionBundleSpec(bundleDir)
	if err != nil {
		return sourceProjectionStats{}, err
	}
	var declaredWrites struct {
		Actions []engine.WriteAction `json:"actions"`
	}
	if err := json.Unmarshal(writesRaw, &declaredWrites); err != nil {
		return sourceProjectionStats{}, fmt.Errorf("writes.json: %w", err)
	}
	var declaredCLI engine.CLISurface
	if err := json.Unmarshal(cliRaw, &declaredCLI); err != nil {
		return sourceProjectionStats{}, fmt.Errorf("cli_surface.json: %w", err)
	}
	declaredMutationBundle := engine.Bundle{Spec: spec, Writes: declaredWrites.Actions, CLISurface: &declaredCLI}
	blockedReads := sourceProjectionBlockedReadSources(result)
	reachableReads := sourceProjectionReachableReadSources(result)
	stats := sourceProjectionStats{CLI: sourceProjectionRestoreSourceBoundDirectReadPathFlagObjects(cli.root, spec, result)}
	stats.CLI += sourceProjectionDowngradeUnreachableReadCommands(cli.root, blockedReads)
	stats.CLI += sourceProjectionRestoreReachableReadCommands(cli.root, blockedReads, reachableReads)
	if api.root != nil {
		stats.Surface = sourceProjectionBlockUnreachableReadSurfaceEndpoints(api.root, blockedReads)
		stats.Surface += sourceProjectionRestoreReachableReadSurfaceEndpoints(api.root, cli.root, blockedReads, reachableReads)
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

	for _, operation := range result.Operations {
		_, readOnly, readOnlyErr := sourceProjectionReadOnlyDeclaration(declarationBundle, operation)
		if readOnlyErr != nil {
			return stats, fmt.Errorf("source operation %s: %w", operation.SourceID, readOnlyErr)
		}
		if sourceProjectionOperationMutates(operation) {
			if readOnly {
				return stats, fmt.Errorf("source operation %s: read-only declaration cannot cover a mutating source operation", operation.SourceID)
			}
			if sourceProjectionHasReadOnlyDisposition(operation) {
				stats.Missing++
				continue
			}
			if operation.Runtime.NonExecutableMutation != nil {
				if !sourceProjectionHasNonExecutableMutationDisposition(operation) {
					return stats, fmt.Errorf("source-cited non-executable mutation disposition is invalid: %s", operation.SourceID)
				}
				if sourceProjectionMutationActionIsComplete(declaredMutationBundle, operation) {
					return stats, fmt.Errorf("source-cited non-executable mutation disposition claims a complete executable action: %s", operation.SourceID)
				}
				if sourceProjectionMutationClaimsImplementedAction(declaredMutationBundle, operation) {
					return stats, fmt.Errorf("source-cited non-executable mutation disposition claims an implemented executable action: %s", operation.SourceID)
				}
				continue
			}
		}
		if readOnly {
			continue
		}
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
	if stats.Surface > 0 {
		if err := writeBundleJSON(apiPath, api, apiRaw); err != nil {
			return stats, err
		}
	}
	return stats, nil
}

func sourceProjectionBundleSpec(bundleDir string) (*engine.Schema, error) {
	raw, err := os.ReadFile(filepath.Join(bundleDir, "spec.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("spec.json: %w", err)
	}
	spec, err := engine.CompileSchema(json.RawMessage(raw))
	if err != nil {
		return nil, fmt.Errorf("spec.json: %w", err)
	}
	return spec, nil
}

type sourceNonExecutableMutationDispositionDocument struct {
	SchemaVersion int                                      `json:"schema_version"`
	Dispositions  []sourceNonExecutableMutationDisposition `json:"dispositions"`
}

// sourceProjectionReadNonExecutableMutationDispositions reads the connector-
// owned, source-cited mutation exceptions. They are deliberately separate from
// read-only coverage: every entry remains a mutation runtime gap until its
// provider action has a complete executable declaration.
func sourceProjectionReadNonExecutableMutationDispositions(bundleDir string) ([]sourceNonExecutableMutationDisposition, error) {
	connector := filepath.Base(filepath.Clean(bundleDir))
	path := filepath.Join(bundleDir, "sources", connector+"-mutation-dispositions.json")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var document sourceNonExecutableMutationDispositionDocument
	if err := decodeSourceStrictJSON(raw, &document); err != nil {
		return nil, fmt.Errorf("parse mutation dispositions: %w", err)
	}
	if document.SchemaVersion != 1 {
		return nil, fmt.Errorf("mutation dispositions schema_version = %d, want 1", document.SchemaVersion)
	}
	seen := make(map[string]bool, len(document.Dispositions))
	for _, disposition := range document.Dispositions {
		if err := sourceProjectionValidateNonExecutableMutationDispositionInput(disposition); err != nil {
			return nil, err
		}
		if seen[disposition.Source.SourceID] {
			return nil, fmt.Errorf("mutation dispositions duplicate source operation %q", disposition.Source.SourceID)
		}
		seen[disposition.Source.SourceID] = true
	}
	return document.Dispositions, nil
}

func sourceProjectionValidateNonExecutableMutationDispositionInput(disposition sourceNonExecutableMutationDisposition) error {
	for _, value := range []struct {
		name  string
		value string
		max   int
	}{
		{name: "source_id", value: disposition.Source.SourceID, max: 1024},
		{name: "method", value: disposition.Source.Method, max: 16},
		{name: "path", value: disposition.Source.Path, max: 4096},
		{name: "reason", value: disposition.Reason, max: 1024},
	} {
		if value.value == "" || value.value != strings.TrimSpace(value.value) || len(value.value) > value.max || strings.ContainsAny(value.value, "\r\n\x00") {
			return fmt.Errorf("mutation disposition %s is invalid", value.name)
		}
	}
	if !sourceProjectionMutationMethod(disposition.Source.Method) {
		return fmt.Errorf("mutation disposition source method %q is not mutating", disposition.Source.Method)
	}
	if !strings.HasPrefix(disposition.Source.Path, "/") {
		return fmt.Errorf("mutation disposition source path %q is invalid", disposition.Source.Path)
	}
	return nil
}

// sourceProjectionApplyNonExecutableMutationDispositions turns a connector-
// owned citation into an operation-owned runtime gap. The source operation and
// its exact HTTP citation are verified before any suppression can occur.
func sourceProjectionApplyNonExecutableMutationDispositions(bundle engine.Bundle, result *sourceImportResult, dispositions []sourceNonExecutableMutationDisposition) error {
	if len(dispositions) == 0 {
		return nil
	}
	if result == nil {
		return fmt.Errorf("mutation dispositions require source operations")
	}
	operations := sourceProjectionOperationsByID(*result)
	seen := make(map[string]bool, len(dispositions))
	for _, disposition := range dispositions {
		if err := sourceProjectionValidateNonExecutableMutationDispositionInput(disposition); err != nil {
			return err
		}
		if seen[disposition.Source.SourceID] {
			return fmt.Errorf("mutation dispositions duplicate source operation %q", disposition.Source.SourceID)
		}
		seen[disposition.Source.SourceID] = true
		operation, found := operations[disposition.Source.SourceID]
		if !found {
			return fmt.Errorf("mutation disposition cites unknown source operation %q", disposition.Source.SourceID)
		}
		if err := sourceProjectionValidateNonExecutableMutationDispositionCitation(operation, disposition); err != nil {
			return err
		}
		if sourceProjectionMutationActionIsComplete(bundle, operation) {
			return fmt.Errorf("mutation disposition source operation %q already has a complete executable action", operation.SourceID)
		}
		if sourceProjectionMutationClaimsImplementedAction(bundle, operation) {
			return fmt.Errorf("mutation disposition source operation %q already claims an implemented executable action", operation.SourceID)
		}
		for index := range result.Operations {
			if result.Operations[index].SourceID != operation.SourceID {
				continue
			}
			if result.Operations[index].Runtime.NonExecutableMutation != nil {
				return fmt.Errorf("source operation %q already has a non-executable mutation disposition", operation.SourceID)
			}
			copyDisposition := disposition
			result.Operations[index].Runtime.NonExecutableMutation = &copyDisposition
			result.Operations[index].Runtime.Gaps = sourceSortedGaps(append(result.Operations[index].Runtime.Gaps, sourceProjectionNonExecutableMutationRuntimeGap(result.Operations[index], copyDisposition)))
			result.Operations[index].Runtime.MergeBlocked = true
			break
		}
	}
	return nil
}

func sourceProjectionValidateNonExecutableMutationDispositionCitation(operation sourceOperationDescriptor, disposition sourceNonExecutableMutationDisposition) error {
	if !sourceProjectionOperationMutates(operation) {
		return fmt.Errorf("mutation disposition source operation %q is not mutating", operation.SourceID)
	}
	if operation.Source.URL == "" || operation.Source.SHA256 == "" || operation.Source.Bytes <= 0 || operation.Source.Location == "" {
		return fmt.Errorf("mutation disposition source operation %q lacks a provider source citation", operation.SourceID)
	}
	if disposition.Source.SourceID != operation.SourceID || !strings.EqualFold(disposition.Source.Method, operation.Method) || disposition.Source.Path != operation.Path {
		return fmt.Errorf("mutation disposition citation does not match provider source operation %q", operation.SourceID)
	}
	return nil
}

func sourceProjectionNonExecutableMutationRuntimeGap(operation sourceOperationDescriptor, disposition sourceNonExecutableMutationDisposition) sourceContractGap {
	return sourceContractGapFor(
		sourceNonExecutableMutationDispositionFoundation,
		"source operation "+operation.SourceID+" at "+operation.Source.URL+"#"+operation.Source.Location,
		"provider-cited mutation has no complete declaration-owned executable action: "+disposition.Reason,
	)
}

func sourceProjectionHasNonExecutableMutationDisposition(operation sourceOperationDescriptor) bool {
	disposition := operation.Runtime.NonExecutableMutation
	if disposition == nil || !operation.Runtime.MergeBlocked || sourceProjectionValidateNonExecutableMutationDispositionInput(*disposition) != nil || sourceProjectionValidateNonExecutableMutationDispositionCitation(operation, *disposition) != nil {
		return false
	}
	want := sourceProjectionNonExecutableMutationRuntimeGap(operation, *disposition)
	for _, gap := range operation.Runtime.Gaps {
		if gap == want {
			return true
		}
	}
	return false
}

func sourceProjectionHasReadOnlyDisposition(operation sourceOperationDescriptor) bool {
	return sourceOperationHasFoundationGap(operation, sourceReadOnlyOperationFoundation)
}

func sourceProjectionMutationActionIsComplete(bundle engine.Bundle, operation sourceOperationDescriptor) bool {
	commands := make(map[string]engine.CLICommand)
	if bundle.CLISurface != nil {
		for _, command := range bundle.CLISurface.Commands {
			if command.Write != "" && command.Availability == "implemented" {
				commands[command.Write] = command
			}
		}
	}
	for _, action := range bundle.Writes {
		if sourceProjectionEndpointKey(action.Method, sourceProjectionPath(action.Path)) == sourceProjectionEndpointKey(operation.Method, operation.Path) && sourceActionCoversOperation(action, commands[action.Name], operation) {
			return true
		}
	}
	return false
}

// sourceProjectionMutationClaimsImplementedAction recognizes any declared
// command claim for the provider mutation, whether that command's action
// contract happens to be complete or not. A disposition may retain only an
// absent, non-executable action; it cannot downgrade or hide a working command.
func sourceProjectionMutationClaimsImplementedAction(bundle engine.Bundle, operation sourceOperationDescriptor) bool {
	if bundle.CLISurface == nil {
		return false
	}
	endpoint := sourceProjectionEndpointKey(operation.Method, operation.Path)
	for _, command := range bundle.CLISurface.Commands {
		if command.Availability != "implemented" {
			continue
		}
		for _, surface := range command.APISurface {
			if sourceProjectionEndpointKey(surface.Method, surface.Path) == endpoint {
				return true
			}
		}
		if command.Write == "" {
			continue
		}
		for _, action := range bundle.Writes {
			if action.Name == command.Write && sourceProjectionEndpointKey(action.Method, sourceProjectionPath(action.Path)) == endpoint {
				return true
			}
		}
	}
	return false
}

// sourceProjectionBlockedReadSources indexes source operations which are
// explicitly blocked because no field-complete declaration-owned read route
// exists. It is source-derived and connector-neutral so the CLI and endpoint
// ledger receive the exact same disposition.
func sourceProjectionBlockedReadSources(result sourceImportResult) map[string]sourceOperationDescriptor {
	blocked := map[string]sourceOperationDescriptor{}
	for _, operation := range result.Operations {
		if operation.Protocol == "graphql" || sourceProjectionMutationMethod(operation.Method) || !sourceProjectionReadHasBlockingGap(operation) {
			continue
		}
		blocked[sourceProjectionEndpointKey(operation.Method, operation.Path)] = operation
	}
	return blocked
}

func sourceProjectionReachableReadSources(result sourceImportResult) map[string]sourceOperationDescriptor {
	reachable := map[string]sourceOperationDescriptor{}
	for _, operation := range result.Operations {
		if operation.Protocol == "graphql" || sourceProjectionMutationMethod(operation.Method) || sourceProjectionReadHasBlockingGap(operation) {
			continue
		}
		reachable[operation.SourceID] = operation
	}
	return reachable
}

// sourceProjectionReadHasBlockingGap retains every source gap for source
// validation, but a schema gap on an omitted optional input cannot make the
// zero-filter direct-read request unexecutable. The command surface does not
// invent an input for that gap, and the direct-read runner sends only declared
// caller values. Required inputs and all other source gaps remain blocking.
func sourceProjectionReadHasBlockingGap(operation sourceOperationDescriptor) bool {
	for _, gap := range operation.Runtime.Gaps {
		if sourceProjectionOptionalAmbiguousParameterGap(operation, gap) || sourceProjectionOmittedOptionalRequestBodySchemaGap(operation, gap) {
			continue
		}
		if sourceProjectionHasBlockingGap([]sourceContractGap{gap}) {
			return true
		}
	}
	return false
}

// sourceProjectionNormalizeNonBlockingReadGaps removes only a typed source
// gap on an omitted optional read input. Keeping it would make a descriptor
// claim the endpoint is unavailable despite the declared zero-input request
// being executable; required input and every other gap remain source-visible.
func sourceProjectionNormalizeNonBlockingReadGaps(result *sourceImportResult) {
	if result == nil {
		return
	}
	for index := range result.Operations {
		operation := &result.Operations[index]
		if operation.Protocol == "graphql" || sourceProjectionMutationMethod(operation.Method) || len(operation.Runtime.Gaps) == 0 {
			continue
		}
		gaps := operation.Runtime.Gaps[:0]
		for _, gap := range operation.Runtime.Gaps {
			if sourceProjectionOptionalAmbiguousParameterGap(*operation, gap) || sourceProjectionOmittedOptionalRequestBodySchemaGap(*operation, gap) {
				continue
			}
			gaps = append(gaps, gap)
		}
		operation.Runtime.Gaps = gaps
		operation.Runtime.MergeBlocked = len(gaps) > 0
	}
}

func sourceProjectionOmittedOptionalRequestBodySchemaGap(operation sourceOperationDescriptor, gap sourceContractGap) bool {
	return gap.Foundation == "cli-request-schema-foundation-r1" &&
		gap.Location == "request body" &&
		operation.Request.Body != nil &&
		!operation.Request.Body.Required
}

func sourceProjectionOptionalAmbiguousParameterGap(operation sourceOperationDescriptor, gap sourceContractGap) bool {
	if gap.Foundation != "cli-request-schema-foundation-r1" || !strings.HasPrefix(gap.Location, "parameter ") || !strings.Contains(gap.Reason, "ambiguous request schema uses ") {
		return false
	}
	name := strings.TrimPrefix(gap.Location, "parameter ")
	if name == "" {
		return false
	}
	for _, parameters := range [][]sourceParameterDescriptor{operation.Request.Path, operation.Request.Query, operation.Request.Header} {
		for _, parameter := range parameters {
			if parameter.Name == name {
				return !parameter.Required
			}
		}
	}
	return false
}

func sourceProjectionBlockedReadCommandNote(sourceID string) string {
	return "Blocked: locked source operation " + sourceID + " has no declaration-owned executable stream, direct-read, binary, or status route."
}

func sourceProjectionBlockedReadSurfaceReason(sourceID string) string {
	return "Locked source operation " + sourceID + " has no field-complete declaration-owned executable route."
}

func sourceProjectionBlockedReadSurfaceNote(sourceID string) string {
	return "Named dependency: source_operation=" + sourceID
}

// sourceProjectionDowngradeUnreachableReadCommands prevents a direct-read
// command from claiming implementation after the source projection has
// recorded that its provider operation has no field-complete
// declaration-owned route.
// The relationship comes entirely from the source endpoint and api_surface
// declaration; it intentionally has no connector-specific command knowledge.
func sourceProjectionDowngradeUnreachableReadCommands(cli *orderedObject, blocked map[string]sourceOperationDescriptor) int {
	changed := 0
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "availability") != "implemented" || stringField(command, "intent") != "direct_read" {
			continue
		}
		var source sourceOperationDescriptor
		matched := false
		for _, rawSurface := range arrayField(command, "api_surface") {
			surface, ok := rawSurface.(*orderedObject)
			if !ok {
				continue
			}
			source, matched = blocked[sourceProjectionEndpointKey(stringField(surface, "method"), stringField(surface, "path"))]
			if matched {
				break
			}
		}
		if !matched {
			continue
		}
		commandChanged := setOrderedIfDifferent(command, "availability", "partial")
		if setOrderedIfDifferent(command, "notes", sourceProjectionBlockedReadCommandNote(source.SourceID)) {
			commandChanged = true
		}
		if commandChanged {
			changed++
		}
	}
	return changed
}

// sourceProjectionRestoreReachableReadCommands restores an existing
// source-bound direct-read command when its endpoint's full required caller
// contract is declared. Certification cohorts choose which routes receive
// additional test evidence; they do not restrict the command's runtime
// authority or API-surface coverage.
func sourceProjectionRestoreReachableReadCommands(cli *orderedObject, blocked map[string]sourceOperationDescriptor, reachable map[string]sourceOperationDescriptor) int {
	changed := 0
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "availability") != "partial" || stringField(command, "intent") != "direct_read" {
			continue
		}
		sourceID, ok := sourceProjectionBlockedReadCommandSourceID(command)
		if !ok {
			continue
		}
		source, found := reachable[sourceID]
		if !found || sourceProjectionCommandHasBlockedEndpoint(command, blocked) || !sourceProjectionCommandHasEndpoint(command, sourceProjectionEndpointKey(source.Method, source.Path)) {
			continue
		}
		command.set("availability", "implemented")
		command.remove("notes")
		changed++
	}
	return changed
}

func sourceProjectionRestoreSourceBoundDirectReadPathFlagObjects(cli *orderedObject, spec *engine.Schema, result sourceImportResult) int {
	operations := sourceProjectionOperationsByID(result)
	changed := 0
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "intent") != "direct_read" || stringField(command, "availability") != "partial" || len(arrayField(command, "api_surface")) != 1 {
			continue
		}
		sourceID, ok := sourceProjectionBlockedReadCommandSourceID(command)
		if !ok {
			continue
		}
		source, found := operations[sourceID]
		if !found || source.Protocol == "graphql" || sourceProjectionMutationMethod(source.Method) || !sourceProjectionCommandHasEndpoint(command, sourceProjectionEndpointKey(source.Method, source.Path)) {
			continue
		}
		commandChanged := false
		for _, rawFlag := range arrayField(command, "flags") {
			flag, ok := rawFlag.(*orderedObject)
			if !ok || !sourceProjectionRequiredDirectReadPathParameter(source, stringField(flag, "maps_to")) {
				continue
			}
			if required, _ := flag.get("required"); required != true {
				flag.set("required", true)
				commandChanged = true
			}
		}
		flags := sourceProjectionOrderedCommandFlags(command)
		missing := sourceProjectionMissingRequiredDirectReadPathParameters(spec, source, flags)
		if len(missing) > 0 && setOrderedIfDifferent(command, "flags", sourceProjectionInsertDirectReadPathFlagObjects(arrayField(command, "flags"), source, missing)) {
			commandChanged = true
		}
		if commandChanged {
			changed++
		}
	}
	return changed
}

func sourceProjectionRequiredDirectReadPathParameter(source sourceOperationDescriptor, mapsTo string) bool {
	for _, parameter := range source.Request.Path {
		if parameter.Required && mapsTo == "path."+parameter.Name {
			return true
		}
	}
	return false
}

func sourceProjectionOrderedCommandFlags(command *orderedObject) []engine.CLIFlag {
	flags := make([]engine.CLIFlag, 0, len(arrayField(command, "flags")))
	for _, raw := range arrayField(command, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		flags = append(flags, engine.CLIFlag{MapsTo: stringField(flag, "maps_to")})
	}
	return flags
}

func sourceProjectionInsertDirectReadPathFlagObjects(flags []any, source sourceOperationDescriptor, missing []sourceParameterDescriptor) []any {
	missingByIndex := sourceProjectionMissingPathParametersByIndex(source, missing)
	out := make([]any, 0, len(flags)+len(missing))
	for _, raw := range flags {
		mapsTo := ""
		if flag, ok := raw.(*orderedObject); ok {
			mapsTo = stringField(flag, "maps_to")
		}
		for _, parameter := range sourceProjectionMissingPathParametersBefore(source, missingByIndex, mapsTo) {
			out = append(out, sourceProjectionOrderedDirectReadPathFlag(parameter))
		}
		out = append(out, raw)
	}
	for _, parameter := range sourceProjectionRemainingMissingPathParameters(source, missingByIndex) {
		out = append(out, sourceProjectionOrderedDirectReadPathFlag(parameter))
	}
	return out
}

func sourceProjectionOrderedDirectReadPathFlag(parameter sourceParameterDescriptor) *orderedObject {
	flag := sourceProjectionDirectReadPathFlag(parameter)
	object := newOrderedObject()
	object.set("name", flag.Name)
	object.set("type", flag.Type)
	object.set("summary", flag.Summary)
	if len(flag.Values) > 0 {
		values := make([]any, len(flag.Values))
		for index := range flag.Values {
			values[index] = flag.Values[index]
		}
		object.set("values", values)
	}
	object.set("maps_to", flag.MapsTo)
	object.set("required", true)
	if flag.MaxBytes > 0 {
		object.set("max_bytes", json.Number(strconv.Itoa(flag.MaxBytes)))
	}
	return object
}

func sourceProjectionBlockedReadCommandSourceID(command *orderedObject) (string, bool) {
	const prefix = "Blocked: locked source operation "
	const suffix = " has no declaration-owned executable stream, direct-read, binary, or status route."
	note := stringField(command, "notes")
	if !strings.HasPrefix(note, prefix) || !strings.HasSuffix(note, suffix) {
		return "", false
	}
	sourceID := strings.TrimSuffix(strings.TrimPrefix(note, prefix), suffix)
	if sourceID == "" || note != sourceProjectionBlockedReadCommandNote(sourceID) {
		return "", false
	}
	return sourceID, true
}

func sourceProjectionCommandHasBlockedEndpoint(command *orderedObject, blocked map[string]sourceOperationDescriptor) bool {
	for _, raw := range arrayField(command, "api_surface") {
		surface, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		if _, found := blocked[sourceProjectionEndpointKey(stringField(surface, "method"), stringField(surface, "path"))]; found {
			return true
		}
	}
	return false
}

func sourceProjectionCommandHasEndpoint(command *orderedObject, endpoint string) bool {
	for _, raw := range arrayField(command, "api_surface") {
		surface, ok := raw.(*orderedObject)
		if ok && sourceProjectionEndpointKey(stringField(surface, "method"), stringField(surface, "path")) == endpoint {
			return true
		}
	}
	return false
}

// sourceProjectionBlockUnreachableReadSurfaceEndpoints replaces stale
// direct_read coverage with the source-owned blocked-operation classification.
// A coverage row is an executable claim, so it cannot remain alongside the
// source gap or a partial command.
func sourceProjectionBlockUnreachableReadSurfaceEndpoints(surface *orderedObject, blocked map[string]sourceOperationDescriptor) int {
	changed := 0
	for _, raw := range arrayField(surface, "endpoints") {
		endpoint, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		source, found := blocked[sourceProjectionEndpointKey(stringField(endpoint, "method"), stringField(endpoint, "path"))]
		if !found {
			continue
		}
		endpointChanged := false
		stillCovered := false
		if rawCoverage, ok := endpoint.get("covered_by"); ok {
			if coverage, ok := rawCoverage.(*orderedObject); ok {
				removedDirectRead := coverage.remove("direct_read")
				removedDirectReads := coverage.remove("direct_reads")
				if removedDirectRead || removedDirectReads {
					endpointChanged = true
				}
				if len(coverage.keys) == 0 && endpoint.remove("covered_by") {
					endpointChanged = true
				}
				stillCovered = len(coverage.keys) > 0
			}
		}
		if stillCovered {
			if endpoint.remove("operation") {
				endpointChanged = true
			}
			if endpointChanged {
				changed++
			}
			continue
		}
		operation := newOrderedObject()
		operation.set("model", "direct_read")
		operation.set("status", "blocked")
		operation.set("risk", "low")
		operation.set("blocked_by_default", true)
		operation.set("reason", sourceProjectionBlockedReadSurfaceReason(source.SourceID))
		operation.set("notes", sourceProjectionBlockedReadSurfaceNote(source.SourceID))
		if setOrderedIfDifferent(endpoint, "operation", operation) {
			endpointChanged = true
		}
		if endpointChanged {
			changed++
		}
	}
	return changed
}

func sourceProjectionRestoreReachableReadSurfaceEndpoints(surface, cli *orderedObject, blocked map[string]sourceOperationDescriptor, reachable map[string]sourceOperationDescriptor) int {
	changed := 0
	for _, raw := range arrayField(surface, "endpoints") {
		endpoint, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		endpointKey := sourceProjectionEndpointKey(stringField(endpoint, "method"), stringField(endpoint, "path"))
		if _, stillBlocked := blocked[endpointKey]; stillBlocked {
			continue
		}
		sourceID, ok := sourceProjectionBlockedReadSurfaceSourceID(endpoint)
		if !ok {
			continue
		}
		source, found := reachable[sourceID]
		if !found || sourceProjectionEndpointKey(source.Method, source.Path) != endpointKey {
			continue
		}
		commands := sourceProjectionReadSurfaceCommandPaths(cli, endpointKey)
		if len(commands) == 0 {
			continue
		}
		endpoint.remove("operation")
		endpoint.set("covered_by", sourceProjectionReadSurfaceCoverage(source, cli, endpointKey, commands))
		changed++
	}
	return changed
}

func sourceProjectionBlockedReadSurfaceSourceID(endpoint *orderedObject) (string, bool) {
	rawOperation, ok := endpoint.get("operation")
	if !ok {
		return "", false
	}
	operation, ok := rawOperation.(*orderedObject)
	if !ok || stringField(operation, "model") != "direct_read" || stringField(operation, "status") != "blocked" {
		return "", false
	}
	const prefix = "Named dependency: source_operation="
	note := stringField(operation, "notes")
	if !strings.HasPrefix(note, prefix) {
		return "", false
	}
	sourceID := strings.TrimPrefix(note, prefix)
	if sourceID == "" || note != sourceProjectionBlockedReadSurfaceNote(sourceID) || stringField(operation, "reason") != sourceProjectionBlockedReadSurfaceReason(sourceID) {
		return "", false
	}
	return sourceID, true
}

// sourceProjectionReadSurfaceCommandPaths returns installed command paths for
// every executable read intent. API-surface coverage calls these historical
// binary and status routes direct_read, but their runtime executor remains the
// command's declared intent.
func sourceProjectionReadSurfaceCommandPaths(cli *orderedObject, endpoint string) []string {
	paths := map[string]bool{}
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || !engine.IsReadSurfaceIntent(stringField(command, "intent")) || stringField(command, "availability") != "implemented" || len(arrayField(command, "api_surface")) != 1 || !sourceProjectionCommandHasEndpoint(command, endpoint) {
			continue
		}
		if path := stringField(command, "path"); path != "" {
			paths[path] = true
		}
	}
	commands := make([]string, 0, len(paths))
	for path := range paths {
		commands = append(commands, path)
	}
	sort.Strings(commands)
	return commands
}

// sourceProjectionReadSurfaceCoverage preserves the established collection
// spelling for a repository-scoped, declaration-owned operation route. Those
// routes are intentionally alias-capable in the GitHub ledger, including a
// route with one current command; reducing them to direct_read during a
// temporary source block loses that source-owned surface shape. A source-only
// direct command without an operation remains a singular binding.
func sourceProjectionReadSurfaceCoverage(source sourceOperationDescriptor, cli *orderedObject, endpoint string, commands []string) *orderedObject {
	if len(commands) == 1 && strings.HasPrefix(source.Path, "/repos/{owner}/{repo}/") && sourceProjectionReadSurfaceCommandOwnsOperation(cli, endpoint, commands[0]) {
		coverage := newOrderedObject()
		coverage.set("direct_reads", []any{commands[0]})
		return coverage
	}
	return directReadCoverage(commands)
}

func sourceProjectionReadSurfaceCommandOwnsOperation(cli *orderedObject, endpoint, path string) bool {
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "path") != path || !sourceProjectionCommandHasEndpoint(command, endpoint) {
			continue
		}
		return strings.TrimSpace(stringField(command, "operation")) != ""
	}
	return false
}

func sourceOperationHasFoundationGap(operation sourceOperationDescriptor, foundation string) bool {
	for _, gap := range operation.Runtime.Gaps {
		if gap.Foundation == foundation {
			return true
		}
	}
	return false
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

func sourceReadOnlyOperationDeclaration(operation *engine.SurfaceOperation) (sourceReadOnlyDeclaration, bool, error) {
	if operation == nil || operation.Model != sourceReadOnlyOperationModel {
		return sourceReadOnlyDeclaration{}, false, nil
	}
	if operation.Status != "blocked" || !operation.BlockedByDefault {
		return sourceReadOnlyDeclaration{}, true, errors.New("read-only declaration must be blocked by default")
	}
	if strings.TrimSpace(operation.Reason) == "" {
		return sourceReadOnlyDeclaration{}, true, errors.New("read-only declaration lacks a reason")
	}
	wantNotes := "Named policy: " + sourceReadOnlyPolicy
	if operation.Notes != wantNotes {
		return sourceReadOnlyDeclaration{}, true, fmt.Errorf("read-only declaration notes = %q, want %q", operation.Notes, wantNotes)
	}
	return sourceReadOnlyDeclaration{Policy: sourceReadOnlyPolicy, Reason: operation.Reason}, true, nil
}

func sourceProjectionReadOnlyDeclaration(bundle engine.Bundle, source sourceOperationDescriptor) (sourceReadOnlyDeclaration, bool, error) {
	endpoint := sourceProjectionSurfaceEndpoint(bundle, source)
	declaration, declared, err := sourceReadOnlyOperationDeclaration(operationForSurfaceEndpoint(endpoint))
	if err != nil || !declared {
		return declaration, declared, err
	}
	if sourceProjectionMutationMethod(source.Method) {
		return sourceReadOnlyDeclaration{}, true, errors.New("read-only declaration cannot cover a mutating source operation")
	}
	return declaration, true, nil
}

func sourceProjectionSurfaceEndpoint(bundle engine.Bundle, source sourceOperationDescriptor) *engine.SurfaceEndpoint {
	if bundle.Surface == nil {
		return nil
	}
	for index := range bundle.Surface.Endpoints {
		endpoint := &bundle.Surface.Endpoints[index]
		if strings.EqualFold(endpoint.Method, source.Method) && endpoint.Path == source.Path {
			return endpoint
		}
	}
	return nil
}

func operationForSurfaceEndpoint(endpoint *engine.SurfaceEndpoint) *engine.SurfaceOperation {
	if endpoint == nil {
		return nil
	}
	return endpoint.Operation
}

func sourceProjectionOperationMutates(operation sourceOperationDescriptor) bool {
	if operation.Protocol == "graphql" {
		return operation.GraphQL != nil && strings.EqualFold(operation.GraphQL.Root, "mutation")
	}
	return sourceProjectionMutationMethod(operation.Method)
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
	contract := sourceActionContract{Fields: map[string]any{}, BareStringFields: map[string]bool{}, SecretFields: map[string]bool{}, Required: map[string]bool{}}
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
		contract.markSourceSecret(name, properties[name])
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
	contract := sourceActionContract{Fields: properties, BareStringFields: map[string]bool{}, SecretFields: map[string]bool{}, Required: sourceSchemaRequired(schema)}
	for _, name := range sortedSourceMapKeys(properties) {
		contract.markSourceSecret(name, rawProperties[name])
	}
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
	if sourceProjectionDeclaredSecret(schema) {
		out["x-secret"] = true
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
	if sourceProjectionDeclaredSecret(schema) {
		out["x-secret"] = true
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
		if contract.SecretFields[name] {
			flag.set("env_only", true)
		} else {
			flag.remove("env_only")
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
	contract.markSourceSecret(name, sourceSchema)
	if sourceProjectionFlagType(projectedSchema) == "json" && sourceProjectionContainsStringArm(sourceSchema) {
		contract.BareStringFields[name] = true
		return
	}
	delete(contract.BareStringFields, name)
}

func (contract *sourceActionContract) markSourceSecret(name string, sourceSchema any) {
	if sourceProjectionDeclaredSecret(sourceSchema) {
		contract.SecretFields[name] = true
	} else {
		delete(contract.SecretFields, name)
	}
}

func sourceProjectionDeclaredSecret(raw any) bool {
	switch schema := raw.(type) {
	case map[string]any:
		if secret, _ := schema["x-secret"].(bool); secret {
			return true
		}
		for _, value := range schema {
			if sourceProjectionDeclaredSecret(value) {
				return true
			}
		}
	case []any:
		for _, value := range schema {
			if sourceProjectionDeclaredSecret(value) {
				return true
			}
		}
	}
	return false
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
	wantSchemaVersion := 2
	if lock.SchemaVersion == 3 {
		wantSchemaVersion = 3
	}
	if descriptor.SchemaVersion != wantSchemaVersion {
		return []Finding{sourceProjectionFinding(connector, file, fmt.Sprintf("source descriptor schema_version = %d, want %d", descriptor.SchemaVersion, wantSchemaVersion))}
	}
	type expectedSource struct {
		source              sourceImportSource
		providerOperationID string
	}
	expected := map[string]expectedSource{}
	if lock.SchemaVersion == 3 {
		for _, document := range lock.Rest.SourceDocuments {
			for _, operation := range document.Operations {
				expected[operation.ID] = expectedSource{
					source: sourceImportSource{
						URL:                 document.Artifact.SourceURL,
						SHA256:              strings.ToLower(document.Artifact.SHA256),
						Bytes:               document.Artifact.Bytes,
						Location:            operation.SourceLocation,
						Form:                "openapi",
						Version:             document.Artifact.OpenAPI,
						DocumentID:          document.ID,
						PublishedURL:        document.PublishedSource.SourceURL,
						PublishedCaptureURL: document.PublishedSource.CaptureURL,
						PublishedSHA256:     strings.ToLower(document.PublishedSource.SHA256),
						PublishedBytes:      document.PublishedSource.Bytes,
						PublishedAdapter:    document.PublishedSource.Adapter,
					},
					providerOperationID: operation.OperationID,
				}
			}
		}
	} else {
		for _, operation := range lock.Rest.Operations {
			identity := operation.OperationID
			if identity == "" {
				identity = operation.ID
			}
			expected[identity] = expectedSource{source: sourceImportSource{SHA256: strings.ToLower(lock.Rest.SHA256), Bytes: lock.Rest.Bytes, Location: operation.SourceLocation}, providerOperationID: operation.OperationID}
		}
	}
	for _, field := range lock.GraphQL.QueryFields {
		expected[fmt.Sprintf("%s.graphql.query.%s", connector, field.Name)] = expectedSource{source: sourceImportSource{SHA256: strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)), Bytes: firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes)}}
	}
	for _, field := range lock.GraphQL.MutationFields {
		expected[fmt.Sprintf("%s.graphql.mutation.%s", connector, field.Name)] = expectedSource{source: sourceImportSource{SHA256: strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)), Bytes: firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes)}}
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
	for identity, expectedOperation := range expected {
		operation, ok := actual[identity]
		if !ok {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor is missing identity "+identity)}
		}
		if operation.Source.SHA256 != expectedOperation.source.SHA256 || operation.Source.Bytes != expectedOperation.source.Bytes || (expectedOperation.source.Location != "" && operation.Source.Location != expectedOperation.source.Location) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor provenance drift for "+identity)}
		}
		if lock.SchemaVersion == 3 && (operation.ProviderOperationID != expectedOperation.providerOperationID || operation.Source.URL != expectedOperation.source.URL || operation.Source.Form != expectedOperation.source.Form || operation.Source.Version != expectedOperation.source.Version || operation.Source.DocumentID != expectedOperation.source.DocumentID || operation.Source.PublishedURL != expectedOperation.source.PublishedURL || operation.Source.PublishedCaptureURL != expectedOperation.source.PublishedCaptureURL || operation.Source.PublishedSHA256 != expectedOperation.source.PublishedSHA256 || operation.Source.PublishedBytes != expectedOperation.source.PublishedBytes || operation.Source.PublishedAdapter != expectedOperation.source.PublishedAdapter) {
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
		if sourceProjectionOperationMutates(operation) {
			if operation.Runtime.NonExecutableMutation != nil {
				if !sourceProjectionHasNonExecutableMutationDisposition(operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited non-executable mutation disposition is invalid: "+operation.SourceID))
					continue
				}
				if sourceProjectionMutationActionIsComplete(bundle, operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited non-executable mutation disposition claims a complete executable action: "+operation.SourceID))
					continue
				}
				if sourceProjectionMutationClaimsImplementedAction(bundle, operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited non-executable mutation disposition claims an implemented executable action: "+operation.SourceID))
					continue
				}
				continue
			}
			if sourceProjectionHasReadOnlyDisposition(operation) {
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, "read-only disposition cannot cover a mutating source operation: "+operation.SourceID))
				continue
			}
		}
		if operation.Protocol == "graphql" {
			if sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
				continue
			}
			if !sourceGraphQLOperationIsReachable(bundle, operation) {
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source operation has no reachable executable operation: "+operation.SourceID))
			}
			continue
		}
		if !sourceProjectionMutationMethod(operation.Method) {
			_, readOnly, readOnlyErr := sourceProjectionReadOnlyDeclaration(bundle, operation)
			if readOnlyErr != nil {
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, readOnlyErr.Error()+": "+operation.SourceID))
				continue
			}
			if readOnly {
				if sourceProjectionReadHasBlockingGap(operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "read-only declaration conflicts with source-bound foundation gap: "+operation.SourceID))
					continue
				}
				if sourceRESTOperationIsReachable(bundle, operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "read-only declaration conflicts with executable operation: "+operation.SourceID))
				}
				continue
			}
			if sourceProjectionReadHasBlockingGap(operation) {
				if sourceGapDirectOperationIsImplementedIncompletely(bundle, operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "implemented source operation retains an unresolved source-bound gap: "+operation.SourceID))
				}
				continue
			}
			if !sourceRESTOperationIsReachable(bundle, operation) {
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source operation has no reachable executable operation: "+operation.SourceID))
			}
			continue
		}
		_, _, readOnlyErr := sourceProjectionReadOnlyDeclaration(bundle, operation)
		if readOnlyErr != nil {
			findings = append(findings, sourceProjectionFinding(bundle.Name, file, readOnlyErr.Error()+": "+operation.SourceID))
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

func sourceRESTOperationIsReachable(bundle engine.Bundle, source sourceOperationDescriptor) bool {
	return sourceRESTOperationIsDeclaredReachable(bundle, source, false)
}

// sourceRESTOperationIsDeclaredReachable evaluates the same closed route as
// validation. During source import it may additionally recognize a command
// which this projection itself previously marked partial for this exact source
// operation; otherwise repeated generation would turn an unrelated source gap
// into a cascading new execution gap.
func sourceRESTOperationIsDeclaredReachable(bundle engine.Bundle, source sourceOperationDescriptor, allowSourceBoundPartial bool) bool {
	endpoint := sourceProjectionEndpointKey(source.Method, source.Path)
	for _, operation := range bundle.Operations {
		if operation.REST != nil && sourceProjectionEndpointKey(operation.REST.Method, operation.REST.Path) == endpoint {
			flags, implemented := sourceProjectionOperationFlags(bundle, operation.ID, source.SourceID, allowSourceBoundPartial)
			if implemented && sourceRESTOperationCoversSource(bundle, operation.REST, source, flags) {
				return true
			}
		}
		if operation.Binary != nil && sourceProjectionEndpointKey(operation.Binary.Method, operation.Binary.Path) == endpoint {
			flags, implemented := sourceProjectionOperationFlags(bundle, operation.ID, source.SourceID, allowSourceBoundPartial)
			if implemented && sourceBinaryOperationCoversSource(bundle, operation.Binary, source, flags) {
				return true
			}
		}
	}
	for _, stream := range bundle.Streams {
		method := stream.Method
		if method == "" {
			method = "GET"
		}
		if sourceProjectionEndpointKey(method, stream.Path) == endpoint && sourceProjectionDeclaredStream(bundle, stream.Name, source.SourceID, allowSourceBoundPartial) {
			return true
		}
	}
	return sourceProjectionDeclaredDirectRead(bundle, source, allowSourceBoundPartial)
}

// sourceProjectionDeclaredDirectRead recognizes an existing closed direct-read
// route that binds exactly one locked source endpoint. Older bundles expressed
// those routes directly through cli_surface/api_surface rather than an
// operations.json entry; their declared config path fields and command flags
// are still the complete request contract. During projection, a route the
// projection itself marked partial may be re-evaluated only when its exact
// source-bound marker is present.
func sourceProjectionDeclaredDirectRead(bundle engine.Bundle, source sourceOperationDescriptor, allowSourceBoundPartial bool) bool {
	if bundle.CLISurface == nil {
		return false
	}
	endpoint := sourceProjectionEndpointKey(source.Method, source.Path)
	for _, command := range bundle.CLISurface.Commands {
		if command.Intent != "direct_read" || len(command.APISurface) != 1 || !sourceProjectionCommandHasEndpointRef(command, endpoint) {
			continue
		}
		if command.Availability != "implemented" && (!allowSourceBoundPartial || command.Availability != "partial" || !sourceProjectionCommandIsSourceBoundPartial(command, source.SourceID)) {
			continue
		}
		covered := sourceProjectionDeclaredConfigPathFields(bundle.Spec, source.Path)
		for _, flag := range command.Flags {
			covered[flag.MapsTo] = true
		}
		if sourceRequiredCallerFieldsCovered(source, covered) {
			return true
		}
	}
	return false
}

func sourceProjectionReadCandidateCommandDeclared(bundle engine.Bundle, path string) bool {
	if bundle.Certification == nil || bundle.Certification.DirectReadGeneration == nil {
		return false
	}
	for _, cohort := range bundle.Certification.DirectReadGeneration.Cohorts {
		for _, command := range cohort.Commands {
			if command == path {
				return true
			}
		}
	}
	return false
}

func sourceProjectionCommandHasEndpointRef(command engine.CLICommand, endpoint string) bool {
	for _, surface := range command.APISurface {
		if sourceProjectionEndpointKey(surface.Method, surface.Path) == endpoint {
			return true
		}
	}
	return false
}

// sourceProjectionRestoreSourceBoundDirectReadPathFlags restores only a
// required path input on a legacy direct-read route which this projection
// previously downgraded for the exact locked source operation. It preserves
// the closed route: config-bound path values and already-declared command
// flags remain authoritative, optional provider filters stay optional, and no
// new command or endpoint mapping is introduced.
func sourceProjectionRestoreSourceBoundDirectReadPathFlags(bundle *engine.Bundle, result sourceImportResult) int {
	if bundle == nil || bundle.CLISurface == nil {
		return 0
	}
	operations := sourceProjectionOperationsByID(result)
	changed := 0
	for index := range bundle.CLISurface.Commands {
		command := &bundle.CLISurface.Commands[index]
		sourceID, ok := sourceProjectionEngineSourceBoundDirectReadID(*command)
		if !ok || len(command.APISurface) != 1 {
			continue
		}
		if !sourceProjectionReadCandidateCommandDeclared(*bundle, command.Path) {
			continue
		}
		source, found := operations[sourceID]
		if !found || source.Protocol == "graphql" || sourceProjectionMutationMethod(source.Method) || !sourceProjectionCommandHasEndpointRef(*command, sourceProjectionEndpointKey(source.Method, source.Path)) {
			continue
		}
		missing := sourceProjectionMissingRequiredDirectReadPathParameters(bundle.Spec, source, command.Flags)
		if len(missing) == 0 {
			continue
		}
		command.Flags = sourceProjectionInsertDirectReadPathFlags(command.Flags, source, missing)
		changed++
	}
	return changed
}

func sourceProjectionEngineSourceBoundDirectReadID(command engine.CLICommand) (string, bool) {
	if command.Intent != "direct_read" || command.Availability != "partial" {
		return "", false
	}
	const prefix = "Blocked: locked source operation "
	const suffix = " has no declaration-owned executable stream, direct-read, binary, or status route."
	if !strings.HasPrefix(command.Notes, prefix) || !strings.HasSuffix(command.Notes, suffix) {
		return "", false
	}
	sourceID := strings.TrimSuffix(strings.TrimPrefix(command.Notes, prefix), suffix)
	if sourceID == "" || command.Notes != sourceProjectionBlockedReadCommandNote(sourceID) {
		return "", false
	}
	return sourceID, true
}

func sourceProjectionOperationsByID(result sourceImportResult) map[string]sourceOperationDescriptor {
	operations := make(map[string]sourceOperationDescriptor, len(result.Operations))
	for _, operation := range result.Operations {
		if operation.SourceID != "" {
			operations[operation.SourceID] = operation
		}
	}
	return operations
}

func sourceProjectionMissingRequiredDirectReadPathParameters(spec *engine.Schema, source sourceOperationDescriptor, flags []engine.CLIFlag) []sourceParameterDescriptor {
	covered := sourceProjectionDeclaredConfigPathFields(spec, source.Path)
	for _, flag := range flags {
		covered[flag.MapsTo] = true
	}
	missing := make([]sourceParameterDescriptor, 0, len(source.Request.Path))
	for _, parameter := range source.Request.Path {
		if parameter.Required && !covered["path."+parameter.Name] {
			missing = append(missing, parameter)
		}
	}
	return missing
}

func sourceProjectionInsertDirectReadPathFlags(flags []engine.CLIFlag, source sourceOperationDescriptor, missing []sourceParameterDescriptor) []engine.CLIFlag {
	missingByIndex := sourceProjectionMissingPathParametersByIndex(source, missing)
	out := make([]engine.CLIFlag, 0, len(flags)+len(missing))
	for _, flag := range flags {
		for _, parameter := range sourceProjectionMissingPathParametersBefore(source, missingByIndex, flag.MapsTo) {
			out = append(out, sourceProjectionDirectReadPathFlag(parameter))
		}
		out = append(out, flag)
	}
	for _, parameter := range sourceProjectionRemainingMissingPathParameters(source, missingByIndex) {
		out = append(out, sourceProjectionDirectReadPathFlag(parameter))
	}
	return out
}

func sourceProjectionMissingPathParametersByIndex(source sourceOperationDescriptor, missing []sourceParameterDescriptor) map[int]sourceParameterDescriptor {
	missingByName := make(map[string]bool, len(missing))
	for _, parameter := range missing {
		missingByName[parameter.Name] = true
	}
	missingByIndex := make(map[int]sourceParameterDescriptor, len(missing))
	for index, parameter := range source.Request.Path {
		if missingByName[parameter.Name] {
			missingByIndex[index] = parameter
		}
	}
	return missingByIndex
}

func sourceProjectionMissingPathParametersBefore(source sourceOperationDescriptor, missing map[int]sourceParameterDescriptor, mapsTo string) []sourceParameterDescriptor {
	position := len(source.Request.Path)
	for index, parameter := range source.Request.Path {
		if mapsTo == "path."+parameter.Name {
			position = index
			break
		}
	}
	return sourceProjectionTakeMissingPathParametersBefore(missing, position)
}

func sourceProjectionRemainingMissingPathParameters(source sourceOperationDescriptor, missing map[int]sourceParameterDescriptor) []sourceParameterDescriptor {
	return sourceProjectionTakeMissingPathParametersBefore(missing, len(source.Request.Path))
}

func sourceProjectionTakeMissingPathParametersBefore(missing map[int]sourceParameterDescriptor, position int) []sourceParameterDescriptor {
	parameters := make([]sourceParameterDescriptor, 0, len(missing))
	for index := 0; index < position; index++ {
		parameter, found := missing[index]
		if !found {
			continue
		}
		parameters = append(parameters, parameter)
		delete(missing, index)
	}
	return parameters
}

func sourceProjectionDirectReadPathFlag(parameter sourceParameterDescriptor) engine.CLIFlag {
	flagType := sourceProjectionFlagType(parameter.Schema)
	flag := engine.CLIFlag{
		Name:     strings.ReplaceAll(parameter.Name, "_", "-"),
		Type:     flagType,
		Summary:  "The " + parameter.Name + " path parameter.",
		MapsTo:   "path." + parameter.Name,
		Required: true,
	}
	if schema, ok := parameter.Schema.(map[string]any); ok {
		for _, value := range sourceAnySlice(schema["enum"]) {
			if name, ok := value.(string); ok {
				flag.Values = append(flag.Values, name)
			}
		}
	}
	if maxBytes := sourceProjectionFlagMaxBytes(parameter.Schema, flagType); maxBytes > 0 {
		flag.MaxBytes = int(maxBytes)
	}
	return flag
}

func sourceGraphQLOperationIsReachable(bundle engine.Bundle, source sourceOperationDescriptor) bool {
	if source.GraphQL == nil {
		return false
	}
	expectedID := source.Connector + ".graphql." + strings.ToLower(source.GraphQL.Root) + "." + sourceProjectionKebab(source.GraphQL.Name)
	for _, operation := range bundle.Operations {
		if operation.GraphQL == nil || (operation.ID != source.SourceID && operation.ID != expectedID) {
			continue
		}
		if operation.OutputPolicy == "" || (len(source.GraphQL.Arguments) > 0 && len(operation.GraphQL.VariablesSchema) == 0) {
			continue
		}
		if _, implemented := sourceProjectionImplementedOperationFlags(bundle, operation.ID); implemented {
			return true
		}
	}
	return false
}

func sourceProjectionImplementedOperationFlags(bundle engine.Bundle, operationID string) ([]engine.CLIFlag, bool) {
	return sourceProjectionOperationFlags(bundle, operationID, "", false)
}

func sourceProjectionOperationFlags(bundle engine.Bundle, operationID, sourceID string, allowSourceBoundPartial bool) ([]engine.CLIFlag, bool) {
	if bundle.CLISurface == nil {
		return nil, false
	}
	var flags []engine.CLIFlag
	implemented := false
	for _, command := range bundle.CLISurface.Commands {
		if command.Operation != operationID || (command.Availability != "implemented" && (!allowSourceBoundPartial || command.Availability != "partial" || !sourceProjectionCommandIsSourceBoundPartial(command, sourceID))) {
			continue
		}
		implemented = true
		flags = append(flags, command.Flags...)
	}
	return flags, implemented
}

func sourceProjectionDeclaredStream(bundle engine.Bundle, streamName, sourceID string, allowSourceBoundPartial bool) bool {
	if bundle.CLISurface == nil {
		return false
	}
	for _, command := range bundle.CLISurface.Commands {
		if command.Stream == streamName && (command.Availability == "implemented" || (allowSourceBoundPartial && command.Availability == "partial" && sourceProjectionCommandIsSourceBoundPartial(command, sourceID))) {
			return true
		}
	}
	return false
}

func sourceProjectionCommandIsSourceBoundPartial(command engine.CLICommand, sourceID string) bool {
	return sourceID != "" && strings.Contains(command.Notes, "locked source operation "+sourceID+" ")
}

func sourceRESTOperationCoversSource(bundle engine.Bundle, operation *engine.RESTOperationSpec, source sourceOperationDescriptor, flags []engine.CLIFlag) bool {
	covered := sourceProjectionDeclaredConfigPathFields(bundle.Spec, operation.Path)
	for _, parameter := range operation.Parameters {
		covered[parameter.In+"."+parameter.Name] = true
	}
	for _, parameter := range operation.PaginationParameters {
		covered[parameter.In+"."+parameter.Name] = true
	}
	if len(operation.BodySchema) > 0 {
		var schema map[string]any
		decoder := json.NewDecoder(bytes.NewReader(operation.BodySchema))
		decoder.UseNumber()
		if decoder.Decode(&schema) == nil {
			if properties, ok := schema["properties"].(map[string]any); ok {
				for name := range properties {
					covered["body."+name] = true
				}
			}
		}
	}
	for _, flag := range flags {
		covered[flag.MapsTo] = true
	}
	return sourceCallerFieldsCovered(source, covered)
}

func sourceBinaryOperationCoversSource(bundle engine.Bundle, operation *engine.BinaryOperationSpec, source sourceOperationDescriptor, flags []engine.CLIFlag) bool {
	covered := sourceProjectionDeclaredConfigPathFields(bundle.Spec, operation.Path)
	for _, parameter := range operation.Parameters {
		covered[parameter.In+"."+parameter.Name] = true
	}
	for _, flag := range flags {
		covered[flag.MapsTo] = true
	}
	return sourceCallerFieldsCovered(source, covered)
}

func sourceProjectionKebab(value string) string {
	var builder strings.Builder
	for index, runeValue := range value {
		if unicode.IsUpper(runeValue) {
			if index > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(unicode.ToLower(runeValue))
			continue
		}
		builder.WriteRune(runeValue)
	}
	return builder.String()
}

// sourceProjectionAnnotateUnreachableReadGaps makes an omitted locked read
// visible as a source-bound, non-implemented capability instead of allowing it
// to disappear from validation, generated guidance, and certification. It does
// not manufacture a generic request path: only a declaration-owned operation
// or stream with an implemented command can keep a source read executable.
func sourceProjectionAnnotateUnreachableReadGaps(bundle engine.Bundle, result *sourceImportResult) {
	if result == nil {
		return
	}
	for index := range result.Operations {
		operation := &result.Operations[index]
		if (operation.Protocol == "graphql" && len(operation.Runtime.Gaps) > 0) ||
			(operation.Protocol != "graphql" && (sourceProjectionMutationMethod(operation.Method) || sourceProjectionReadHasBlockingGap(*operation))) {
			continue
		}
		reachable := false
		if operation.Protocol == "graphql" {
			reachable = sourceGraphQLOperationIsReachable(bundle, *operation)
		} else {
			reachable = sourceRESTOperationIsDeclaredReachable(bundle, *operation, true)
		}
		if reachable {
			continue
		}
		operation.Runtime = sourceRuntimeReachability{
			MergeBlocked: true,
			Gaps: []sourceContractGap{sourceContractGapFor(
				sourceOperationExecutionFoundation,
				"source operation "+operation.SourceID,
				"locked provider operation has no field-complete declaration-owned executable stream, direct-read, binary, or status route",
			)},
		}
	}
}

// sourceProjectionExecutionSurface reads only the declaration-owned execution
// surfaces needed to classify a locked source operation. Source import must
// keep working for a descriptor-only fixture, so absent optional runtime files
// mean that no executable route has been declared rather than a raw fallback.
func sourceProjectionExecutionSurface(bundleDir, connector string) (engine.Bundle, error) {
	bundle := engine.Bundle{Name: connector}
	for _, document := range []struct {
		path   string
		decode func([]byte) error
	}{
		{path: "spec.json", decode: func(raw []byte) error {
			spec, err := engine.CompileSchema(json.RawMessage(raw))
			if err != nil {
				return err
			}
			bundle.Spec = spec
			return nil
		}},
		{path: "operations.json", decode: func(raw []byte) error {
			var value struct {
				Operations []engine.OperationSpec `json:"operations"`
			}
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			bundle.Operations = value.Operations
			return nil
		}},
		{path: "streams.json", decode: func(raw []byte) error {
			var value struct {
				Streams []engine.StreamSpec `json:"streams"`
			}
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			bundle.Streams = value.Streams
			return nil
		}},
		{path: "writes.json", decode: func(raw []byte) error {
			var value struct {
				Actions []engine.WriteAction `json:"actions"`
			}
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			bundle.Writes = value.Actions
			return nil
		}},
		{path: "cli_surface.json", decode: func(raw []byte) error {
			var value engine.CLISurface
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			bundle.CLISurface = &value
			return nil
		}},
		{path: "api_surface.json", decode: func(raw []byte) error {
			var value engine.APISurface
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			bundle.Surface = &value
			return nil
		}},
		{path: "certification.json", decode: func(raw []byte) error {
			var value engine.CertificationSpec
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			bundle.Certification = &value
			return nil
		}},
	} {
		raw, err := os.ReadFile(filepath.Join(bundleDir, document.path))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return engine.Bundle{}, fmt.Errorf("read %s: %w", document.path, err)
		}
		if err := document.decode(raw); err != nil {
			return engine.Bundle{}, fmt.Errorf("parse %s: %w", document.path, err)
		}
	}
	return bundle, nil
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

// sourceRequiredCallerFieldsCovered keeps legacy direct-read routes usable
// with provider defaults for optional filters and pagination while still
// requiring every source-required path, query, header, and required body
// field to have a declaration-owned binding. It is deliberately narrower than
// sourceCallerFieldsCovered: source projection uses that stricter predicate
// when producing new operation and action declarations.
func sourceRequiredCallerFieldsCovered(source sourceOperationDescriptor, covered map[string]bool) bool {
	for _, group := range []struct {
		prefix string
		items  []sourceParameterDescriptor
	}{{"path.", source.Request.Path}, {"query.", source.Request.Query}, {"header.", source.Request.Header}} {
		for _, parameter := range group.items {
			if parameter.Required && !covered[group.prefix+parameter.Name] {
				return false
			}
		}
	}
	if source.Request.Body == nil || !source.Request.Body.Required {
		return true
	}
	schema, ok := source.Request.Body.Schema.(map[string]any)
	if !ok {
		return false
	}
	for name := range sourceSchemaRequired(schema) {
		if !covered["body."+name] {
			return false
		}
	}
	return true
}

func sourceGapDirectOperationIsImplementedIncompletely(bundle engine.Bundle, source sourceOperationDescriptor) bool {
	if sourceGapHasImplementedAPICommand(bundle, source) {
		return true
	}
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
		for _, parameter := range operation.REST.PaginationParameters {
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

func sourceGapHasImplementedAPICommand(bundle engine.Bundle, source sourceOperationDescriptor) bool {
	if bundle.CLISurface == nil {
		return false
	}
	endpoint := sourceProjectionEndpointKey(source.Method, source.Path)
	for _, command := range bundle.CLISurface.Commands {
		if command.Availability != "implemented" || command.Intent != "direct_read" {
			continue
		}
		for _, surface := range command.APISurface {
			if sourceProjectionEndpointKey(surface.Method, surface.Path) == endpoint {
				return true
			}
		}
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
