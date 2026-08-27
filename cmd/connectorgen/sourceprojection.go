package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"net/http"
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
	"polymetrics.ai/internal/safety"
)

const (
	sourceProjectionDefaultStringBytes               = 8 << 10
	sourceProjectionDefaultArrayItems                = 256
	sourceProjectionDefaultObjectProperties          = 256
	sourceOperationExecutionFoundation               = "closed-source-operation-execution-foundation-r1"
	sourceNonExecutableMutationDispositionFoundation = "source-cited-non-executable-mutation-foundation-r1"
	sourcePartialMutationCoverageFoundation          = "source-cited-partial-mutation-coverage-foundation-r1"
	// sourceReadOnlyOperationFoundation is intentionally distinct from the
	// mutation disposition. A read-only declaration can never satisfy mutation
	// coverage, even when its endpoint currently lacks an executable action.
	sourceReadOnlyOperationFoundation = "source-read-only-operation-foundation-r1"
	sourceReadOnlyOperationModel      = "read_only"
	sourceReadOnlyPolicy              = "source-cited-read-only-operations-r1"
	// sourceWriteDisabledMutationArtifactReason is attached only to a mutation
	// whose connector explicitly declares no write capability. The matching
	// provider citation and named non-executable foundation make that absence
	// auditable without fabricating an action or a runnable command.
	sourceWriteDisabledMutationArtifactReason = "connector declares no write capability and has no complete declaration-owned executable action"
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
	Writes     int
	Operations int
	Streams    int
	CLI        int
	Surface    int
	Missing    int
}

func (s sourceProjectionStats) Changed() bool {
	return s.Writes+s.Operations+s.Streams+s.CLI+s.Surface+s.Missing > 0
}

// sourceProjectionMaterializeDirectReadSurfaceEndpoints replaces only a
// legacy blocked API record once its exact canonical command has already been
// materialized. Source metadata by itself cannot promote a route; the command
// must be implemented, source-bound, and name the same method/path. Stream
// coverage remains stream-owned.
func sourceProjectionMaterializeDirectReadSurfaceEndpoints(surface, cli *orderedObject, result sourceImportResult) int {
	byEndpoint := map[string]sourceOperationDescriptor{}
	for _, source := range result.Operations {
		if source.Protocol == "graphql" || sourceProjectionMutationMethod(source.Method) || sourceProjectionReadHasBlockingGap(source) {
			continue
		}
		endpoint := sourceProjectionEndpointKey(source.Method, source.Path)
		if _, duplicate := byEndpoint[endpoint]; duplicate {
			delete(byEndpoint, endpoint)
			continue
		}
		byEndpoint[endpoint] = source
	}
	changed := 0
	for _, raw := range arrayField(surface, "endpoints") {
		endpoint, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		key := sourceProjectionEndpointKey(stringField(endpoint, "method"), stringField(endpoint, "path"))
		source, found := byEndpoint[key]
		if !found {
			continue
		}
		paths := sourceProjectionImplementedDirectReadCommandPaths(cli, key, source.SourceID)
		if len(paths) == 0 {
			continue
		}
		coverage := directReadCoverage(paths)
		current, hasCoverage := endpoint.get("covered_by")
		if hasCoverage && orderedSemanticEqual(current, coverage) {
			if endpoint.remove("operation") {
				changed++
			}
			continue
		}
		endpoint.remove("operation")
		endpoint.set("covered_by", coverage)
		changed++
	}
	return changed
}

func sourceProjectionImplementedDirectReadCommandPaths(cli *orderedObject, endpoint, sourceID string) []string {
	paths := make([]string, 0)
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "intent") != "direct_read" || stringField(command, "availability") != "implemented" || stringField(command, "source_operation") != sourceID || !sourceProjectionCommandHasEndpoint(command, endpoint) {
			continue
		}
		if path := stringField(command, "path"); path != "" {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

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
	return projectSourceDescriptorToBundleMode(bundleDir, result, check, false, true)
}

// projectSourceBoundReadDescriptorToBundle is the explicit, read-only source
// binding lane. Generic source import deliberately leaves provider reads alone:
// it must remain byte-stable for connectors that have not opted into this
// foundation.
func projectSourceBoundReadDescriptorToBundle(bundleDir string, result sourceImportResult, check bool) (sourceProjectionStats, error) {
	return projectSourceDescriptorToBundleMode(bundleDir, result, check, true, true)
}

// projectSourceMutationMappingsToBundle verifies source-cited mutation
// dispositions from an already checked-in descriptor without rewriting the
// connector's separately planned read surface. surface-sync uses this narrow
// mode so its generic mutation parity check cannot replace author-owned
// declaration-pending read explanations.
func projectSourceMutationMappingsToBundle(bundleDir string, result sourceImportResult, check bool) (sourceProjectionStats, error) {
	return projectSourceDescriptorToBundleMode(bundleDir, result, check, false, false)
}

func projectSourceDescriptorToBundleMode(bundleDir string, result sourceImportResult, check, materializeReads, reconcileReadSurface bool) (sourceProjectionStats, error) {
	// A cited-only descriptor contributes source-to-declaration evidence, but it
	// has no executable contract. Remove it before *any* mode-specific
	// projection transform so source references cannot become an execution gate.
	result = sourceProjectionMaterializableResult(result)
	if len(result.Operations) == 0 {
		return sourceProjectionStats{}, nil
	}
	if err := validateSourceProjectionExecutionEnvelopes(result); err != nil {
		return sourceProjectionStats{}, err
	}
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
	operationsPath := filepath.Join(bundleDir, "operations.json")
	operationsRaw, operationsErr := os.ReadFile(operationsPath)
	if operationsErr != nil && !os.IsNotExist(operationsErr) {
		return sourceProjectionStats{}, operationsErr
	}
	var operations orderedJSON
	hasOperations := operationsErr == nil
	if hasOperations {
		if err := json.Unmarshal(operationsRaw, &operations); err != nil {
			return sourceProjectionStats{}, fmt.Errorf("operations.json: %w", err)
		}
	}
	streamsPath := filepath.Join(bundleDir, "streams.json")
	streamsRaw, streamsErr := os.ReadFile(streamsPath)
	if streamsErr != nil && !os.IsNotExist(streamsErr) {
		return sourceProjectionStats{}, streamsErr
	}
	var streams orderedJSON
	if streamsErr == nil {
		if err := json.Unmarshal(streamsRaw, &streams); err != nil {
			return sourceProjectionStats{}, fmt.Errorf("streams.json: %w", err)
		}
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
	executionSurface, err := sourceProjectionExecutionSurface(bundleDir, declarationBundle.Name)
	if err != nil {
		return sourceProjectionStats{}, err
	}
	declarationBundle.Metadata = executionSurface.Metadata
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
	stats := sourceProjectionStats{}
	if reconcileReadSurface {
		blockedReads := sourceProjectionBlockedReadSources(result)
		reachableReads := sourceProjectionReachableReadSources(result)
		stats.CLI = sourceProjectionRestoreSourceBoundDirectReadPathFlagObjects(cli.root, spec, result)
		stats.CLI += sourceProjectionDowngradeUnreachableReadCommands(cli.root, blockedReads)
		stats.CLI += sourceProjectionRestoreReachableReadCommands(cli.root, blockedReads, reachableReads)
		if api.root != nil {
			stats.Surface = sourceProjectionBlockUnreachableReadSurfaceEndpoints(api.root, blockedReads)
			stats.Surface += sourceProjectionRestoreReachableReadSurfaceEndpoints(api.root, cli.root, blockedReads, reachableReads)
		}
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
	noBodyStats, err := sourceProjectionMaterializeNoBodyMutationActions(writes.root, cli.root, operations.root, result)
	if err != nil {
		return stats, err
	}
	stats.Writes += noBodyStats.Writes
	stats.CLI += noBodyStats.CLI
	stats.Operations += noBodyStats.Operations
	if noBodyStats.Writes > 0 || noBodyStats.CLI > 0 {
		actionsByEndpoint = map[string][]*orderedObject{}
		for _, raw := range arrayField(writes.root, "actions") {
			action, ok := raw.(*orderedObject)
			if !ok {
				continue
			}
			key := sourceProjectionEndpointKey(stringField(action, "method"), sourceProjectionPath(stringField(action, "path")))
			actionsByEndpoint[key] = append(actionsByEndpoint[key], action)
		}
		commandsByWrite = map[string][]*orderedObject{}
		for _, raw := range arrayField(cli.root, "commands") {
			command, ok := raw.(*orderedObject)
			if !ok || stringField(command, "write") == "" {
				continue
			}
			write := stringField(command, "write")
			commandsByWrite[write] = append(commandsByWrite[write], command)
		}
	}
	if api.root != nil {
		stats.Surface += sourceProjectionMaterializeNoBodyMutationSurfaceEndpoints(api.root, writes.root, cli.root, result)
	}
	gapStats := sourceProjectionAnnotateMutationFoundationGaps(operations.root, cli.root, api.root, result)
	stats.Operations += gapStats.Operations
	stats.CLI += gapStats.CLI
	stats.Surface += gapStats.Surface
	readGapStats := sourceProjectionAnnotateReadFoundationGaps(operations.root, cli.root, api.root, result)
	stats.Operations += readGapStats.Operations
	stats.CLI += readGapStats.CLI
	stats.Surface += readGapStats.Surface

	for _, operation := range result.Operations {
		// A source reference carries operation identity and a cited provider URL,
		// not a request/response contract. It may update evidence and validation,
		// but it must never materialize or mutate a declaration-owned action.
		if sourceOperationHasFoundationGap(operation, sourceContractUnavailableFoundation) {
			continue
		}
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
				if err := sourceProjectionValidateWriteDisabledMutationArtifact(declarationBundle, operation); err != nil {
					return stats, fmt.Errorf("%w: %s", err, operation.SourceID)
				}
				if sourceProjectionMutationActionIsComplete(declaredMutationBundle, operation) {
					return stats, fmt.Errorf("source-cited non-executable mutation disposition claims a complete executable action: %s", operation.SourceID)
				}
				if sourceProjectionMutationClaimsImplementedAction(declaredMutationBundle, operation) {
					return stats, fmt.Errorf("source-cited non-executable mutation disposition claims an implemented executable action: %s", operation.SourceID)
				}
				continue
			}
			if operation.Runtime.PartialCoverageMutation != nil {
				if !sourceProjectionHasPartialMutationCoverageDisposition(operation) {
					return stats, fmt.Errorf("source-cited partial mutation coverage disposition is invalid: %s", operation.SourceID)
				}
				if !sourceProjectionPartialCoverageFoundationMatchesOperation(declaredMutationBundle, operation, *operation.Runtime.PartialCoverageMutation) {
					return stats, fmt.Errorf("source-cited partial mutation coverage disposition has no matching missing foundation: %s", operation.SourceID)
				}
				if sourceProjectionMutationActionIsComplete(declaredMutationBundle, operation) {
					return stats, fmt.Errorf("source-cited partial mutation coverage disposition claims a complete executable action: %s", operation.SourceID)
				}
				if !sourceProjectionMutationClaimsImplementedAction(declaredMutationBundle, operation) {
					return stats, fmt.Errorf("source-cited partial mutation coverage disposition has no implemented declared action: %s", operation.SourceID)
				}
				continue
			}
		}
		if readOnly {
			continue
		}
		if !sourceProjectionOperationMutates(operation) {
			if !materializeReads || !hasOperations {
				continue
			}
			readStats, err := sourceProjectionMaterializeReadOperation(operations.root, cli.root, streams.root, operation, result.Operations)
			if err != nil {
				return stats, fmt.Errorf("source operation %s: %w", operation.SourceID, err)
			}
			stats.Operations += readStats.Operations
			stats.Streams += readStats.Streams
			stats.CLI += readStats.CLI
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
	if materializeReads && api.root != nil {
		stats.Surface += sourceProjectionMaterializeDirectReadSurfaceEndpoints(api.root, cli.root, result)
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
	if stats.Operations > 0 {
		if err := writeBundleJSON(operationsPath, operations, operationsRaw); err != nil {
			return stats, err
		}
	}
	if stats.Streams > 0 {
		if err := writeBundleJSON(streamsPath, streams, streamsRaw); err != nil {
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

func sourceProjectionMaterializableResult(result sourceImportResult) sourceImportResult {
	filtered := result
	filtered.Operations = make([]sourceOperationDescriptor, 0, len(result.Operations))
	for _, operation := range result.Operations {
		if sourceOperationHasFoundationGap(operation, sourceContractUnavailableFoundation) {
			continue
		}
		filtered.Operations = append(filtered.Operations, operation)
	}
	return filtered
}

// sourceProjectionMaterializeNoBodyMutationActions promotes a source-complete
// fixed mutation only when its existing rest_write operation and canonical CLI
// command already identify the exact endpoint. It supplies the established
// reverse-ETL action envelope (including destructive approval), never a
// generic HTTP write route.
func sourceProjectionMaterializeNoBodyMutationActions(writes, cli, operations *orderedObject, result sourceImportResult) (sourceProjectionStats, error) {
	stats := sourceProjectionStats{}
	for _, source := range result.Operations {
		if !sourceProjectionNoBodyMutationActionEligible(source) {
			continue
		}
		operation := sourceProjectionRestWriteOperationForEndpoint(operations, source.Method, source.Path)
		if operation == nil {
			// The source projection never invents an operation just to promote a
			// mutation. Connectors which do not yet declare this endpoint keep
			// their own explicit foundation disposition; the coverage gate below
			// proves every eligible endpoint has a real lane.
			continue
		}
		operationID := stringField(operation, "id")
		command := sourceProjectionReverseETLCommandForOperation(cli, operationID)
		if operationID == "" || command == nil {
			continue
		}
		if stringField(command, "write") != "" {
			if sourceProjectionReplaceLegacyPromotedMutationSummary(command) {
				stats.CLI++
			}
			if setOrderedIfDifferent(command, "notes", "Implemented source-bound provider mutation through the declared reverse-ETL action.") {
				stats.CLI++
			}
			if sourceProjectionRemoveApprovalConfirmFlag(command) {
				stats.CLI++
			}
			continue
		}
		if sourceProjectionActionForEndpoint(writes, source.Method, source.Path) != nil {
			continue
		}
		action := newOrderedObject()
		action.set("name", operationID)
		if strings.EqualFold(source.Method, http.MethodDelete) {
			action.set("kind", "delete")
		} else {
			action.set("kind", "update")
		}
		action.set("method", strings.ToUpper(source.Method))
		action.set("path", sourceProjectionMutationRecordPath(source.Path))
		pathFields := make([]any, 0, len(source.Request.Path))
		redactFields := make([]any, 0, len(source.Request.Path))
		for _, parameter := range source.Request.Path {
			if !parameter.Required || !sourceScalarWireSchema(parameter.Schema) {
				return stats, fmt.Errorf("source operation %s has no bounded typed path action contract", source.SourceID)
			}
			pathFields = append(pathFields, parameter.Name)
			redactFields = append(redactFields, parameter.Name)
		}
		action.set("path_fields", pathFields)
		action.set("body_type", "none")
		if strings.EqualFold(source.Method, http.MethodDelete) {
			deleteSpec := newOrderedObject()
			deleteSpec.set("missing_ok_status", []any{404})
			deleteSpec.set("idempotent", true)
			action.set("delete", deleteSpec)
		}
		action.set("risk", "destructive mutation; requires reverse ETL plan -> preview -> explicit approval -> execute and typed confirm destructive")
		action.set("confirm", "destructive")
		action.set("redact_fields", redactFields)
		writes.set("actions", append(arrayField(writes, "actions"), action))

		command.set("availability", "implemented")
		command.set("write", operationID)
		command.remove("operation")
		sourceProjectionRemoveApprovalConfirmFlag(command)
		sourceProjectionReplaceLegacyPromotedMutationSummary(command)
		command.set("notes", "Implemented source-bound provider mutation through the declared reverse-ETL action.")
		operation.set("description", "Implemented source-bound provider operation mapped to declared reverse-ETL action "+operationID+".")
		stats.Writes++
		stats.CLI++
		stats.Operations++
	}
	return stats, nil
}

// sourceProjectionRemoveApprovalConfirmFlag removes a legacy command-owned
// echo of the reverse-ETL confirmation carrier. `--confirm` is handled by the
// shared plan/approval lifecycle, not a record or command mapping; retaining
// it on a newly implemented command would make the declared CLI invalid and
// suggest a second execution channel.
func sourceProjectionRemoveApprovalConfirmFlag(command *orderedObject) bool {
	flags := arrayField(command, "flags")
	kept := make([]any, 0, len(flags))
	changed := false
	for _, raw := range flags {
		flag, ok := raw.(*orderedObject)
		if ok && stringField(flag, "maps_to") == "approval.confirm" {
			changed = true
			continue
		}
		kept = append(kept, raw)
	}
	if changed {
		command.set("flags", kept)
	}
	return changed
}

// sourceProjectionReplaceLegacyPromotedMutationSummary replaces a historical
// planning label only once the source-complete operation has a declared,
// executable reverse-ETL action. It deliberately retains the author-owned
// operation wording after the standard source-bound prefix.
func sourceProjectionReplaceLegacyPromotedMutationSummary(command *orderedObject) bool {
	const prefix = "Planned fixed-target "
	summary := stringField(command, "summary")
	if !strings.HasPrefix(summary, prefix) {
		return false
	}
	return setOrderedIfDifferent(command, "summary", "Implemented source-bound "+strings.TrimPrefix(summary, prefix))
}

func sourceProjectionRestWriteOperationForEndpoint(operations *orderedObject, method, path string) *orderedObject {
	endpoint := sourceProjectionEndpointKey(method, path)
	var match *orderedObject
	for _, raw := range arrayField(operations, "operations") {
		operation, ok := raw.(*orderedObject)
		if !ok || stringField(operation, "kind") != "rest_write" {
			continue
		}
		restRaw, declared := operation.get("rest")
		rest, ok := restRaw.(*orderedObject)
		if !declared || !ok || sourceProjectionEndpointKey(stringField(rest, "method"), stringField(rest, "path")) != endpoint {
			continue
		}
		if match != nil {
			return nil
		}
		match = operation
	}
	return match
}

func sourceProjectionReverseETLCommandForOperation(cli *orderedObject, operationID string) *orderedObject {
	var match *orderedObject
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "intent") != "reverse_etl" {
			continue
		}
		if stringField(command, "operation") != operationID && stringField(command, "write") != operationID {
			continue
		}
		if match != nil {
			return nil
		}
		match = command
	}
	return match
}

func sourceProjectionNoBodyMutationActionEligible(source sourceOperationDescriptor) bool {
	return source.Protocol != "graphql" && sourceProjectionMutationMethod(source.Method) && source.Runtime.NonExecutableMutation == nil && source.Runtime.PartialCoverageMutation == nil && !sourceProjectionHasBlockingGap(source.Runtime.Gaps) && source.Request.Body == nil && len(source.Request.Header) == 0 && len(source.Request.Media) == 0 && len(source.Request.Path) > 0
}

func sourceProjectionActionForEndpoint(writes *orderedObject, method, path string) *orderedObject {
	endpoint := sourceProjectionEndpointKey(method, path)
	for _, raw := range arrayField(writes, "actions") {
		action, ok := raw.(*orderedObject)
		if ok && sourceProjectionEndpointKey(stringField(action, "method"), sourceProjectionPath(stringField(action, "path"))) == endpoint {
			return action
		}
	}
	return nil
}

func sourceProjectionMutationRecordPath(path string) string {
	return sourceProjectionPathVariableRE.ReplaceAllString(path, "{{ record.$1 }}")
}

// sourceProjectionMaterializeNoBodyMutationSurfaceEndpoints promotes the
// source-ledger row only after its exact action and implemented reverse-ETL
// command exist. This is the mutation equivalent of the source-bound direct
// read projection: api_surface is a declaration of executable coverage, never
// a way to make an endpoint executable by itself.
func sourceProjectionMaterializeNoBodyMutationSurfaceEndpoints(surface, writes, cli *orderedObject, result sourceImportResult) int {
	actions := map[string]string{}
	for _, raw := range arrayField(writes, "actions") {
		action, ok := raw.(*orderedObject)
		if !ok || stringField(action, "name") == "" {
			continue
		}
		key := sourceProjectionEndpointKey(stringField(action, "method"), sourceProjectionPath(stringField(action, "path")))
		if _, duplicate := actions[key]; duplicate {
			delete(actions, key)
			continue
		}
		actions[key] = stringField(action, "name")
	}
	commands := map[string]bool{}
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "intent") != "reverse_etl" || stringField(command, "availability") != "implemented" || stringField(command, "write") == "" {
			continue
		}
		commands[stringField(command, "write")] = true
	}
	sources := map[string]sourceOperationDescriptor{}
	for _, source := range result.Operations {
		if sourceProjectionNoBodyMutationActionEligible(source) {
			sources[sourceProjectionEndpointKey(source.Method, source.Path)] = source
		}
	}
	changed := 0
	for _, raw := range arrayField(surface, "endpoints") {
		endpoint, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		key := sourceProjectionEndpointKey(stringField(endpoint, "method"), stringField(endpoint, "path"))
		if _, sourceComplete := sources[key]; !sourceComplete {
			continue
		}
		write := actions[key]
		if write == "" || !commands[write] {
			continue
		}
		coverage := newOrderedObject()
		coverage.set("write", write)
		current, hasCoverage := endpoint.get("covered_by")
		if hasCoverage && orderedSemanticEqual(current, coverage) {
			if endpoint.remove("operation") {
				changed++
			}
			continue
		}
		endpoint.remove("operation")
		endpoint.set("covered_by", coverage)
		changed++
	}
	return changed
}

// sourceProjectionAnnotateMutationFoundationGaps preserves every
// source-backed mutation that is not executable yet, but replaces inherited
// generic planned wording with the exact current source foundations. These are
// not a promotion path: availability stays non-executable until the declared
// action can satisfy those foundations.
func sourceProjectionAnnotateMutationFoundationGaps(operations, cli, surface *orderedObject, result sourceImportResult) sourceProjectionStats {
	stats := sourceProjectionStats{}
	for _, source := range result.Operations {
		if !sourceProjectionOperationMutates(source) {
			continue
		}
		operation := sourceProjectionRestWriteOperationForEndpoint(operations, source.Method, source.Path)
		operationID := ""
		if operation != nil {
			operationID = stringField(operation, "id")
		}
		command := sourceProjectionCommandForOperation(cli, operationID, source.Method, source.Path)
		if command == nil {
			command = sourceProjectionMutationCommandForEndpoint(cli, source.Method, source.Path)
			if command != nil {
				operationID = firstNonEmpty(stringField(command, "operation"), stringField(command, "write"))
				operation = sourceProjectionOperationByID(operations, operationID)
			}
		}
		if command == nil {
			continue
		}
		availability := stringField(command, "availability")
		if availability != "planned" && availability != "unsupported_api" {
			continue
		}
		note := ""
		if sourceProjectionHasBlockingGap(source.Runtime.Gaps) && availability == "planned" {
			note = sourceProjectionMutationFoundationNote(source)
		} else if availability == "unsupported_api" && source.Path == "/batch" {
			note = "not_applicable=generic_batch_wrapper; source_operation=" + source.SourceID
		} else {
			continue
		}
		if setOrderedIfDifferent(command, "notes", note) {
			stats.CLI++
		}
		if sourceProjectionReplaceLegacyUnavailableMutationSummary(command, availability == "unsupported_api") {
			stats.CLI++
		}
		description := "Unavailable source-bound provider mutation: " + note
		if availability == "unsupported_api" {
			description = "Not-applicable source-bound provider generic wrapper: " + note
		}
		if operation != nil && setOrderedIfDifferent(operation, "description", description) {
			stats.Operations++
		}
		if surface == nil {
			continue
		}
		for _, raw := range arrayField(surface, "endpoints") {
			endpoint, ok := raw.(*orderedObject)
			if !ok || sourceProjectionEndpointKey(stringField(endpoint, "method"), stringField(endpoint, "path")) != sourceProjectionEndpointKey(source.Method, source.Path) {
				continue
			}
			operationRaw, declared := endpoint.get("operation")
			classified, ok := operationRaw.(*orderedObject)
			if !declared || !ok {
				continue
			}
			if setOrderedIfDifferent(classified, "reason", note) {
				stats.Surface++
			}
			if setOrderedIfDifferent(classified, "notes", "source_operation="+source.SourceID) {
				stats.Surface++
			}
		}
	}
	return stats
}

// sourceProjectionMutationCommandForEndpoint locates a declared write command
// by the provider endpoint it names. File and composite operations do not have
// a rest_write declaration, but they are still source-ledger rows and must
// retain their exact unavailable classification rather than old planning prose.
func sourceProjectionMutationCommandForEndpoint(cli *orderedObject, method, path string) *orderedObject {
	endpoint := sourceProjectionEndpointKey(method, path)
	var match *orderedObject
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || !sourceProjectionCommandHasEndpoint(command, endpoint) {
			continue
		}
		intent := stringField(command, "intent")
		if intent != "reverse_etl" && intent != "direct_write" {
			continue
		}
		if match != nil {
			return nil
		}
		match = command
	}
	return match
}

func sourceProjectionOperationByID(operations *orderedObject, operationID string) *orderedObject {
	if operationID == "" {
		return nil
	}
	for _, raw := range arrayField(operations, "operations") {
		operation, ok := raw.(*orderedObject)
		if ok && stringField(operation, "id") == operationID {
			return operation
		}
	}
	return nil
}

func sourceProjectionMutationFoundationNote(source sourceOperationDescriptor) string {
	foundations := map[string]bool{}
	for _, gap := range source.Runtime.Gaps {
		if sourceProjectionHasBlockingGap([]sourceContractGap{gap}) && gap.Foundation != "" {
			foundations[gap.Foundation] = true
		}
	}
	names := make([]string, 0, len(foundations))
	for foundation := range foundations {
		names = append(names, foundation)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names)+1)
	for _, foundation := range names {
		parts = append(parts, "missing_foundation="+foundation)
	}
	if len(parts) == 0 {
		parts = append(parts, "missing_foundation="+sourceOperationExecutionFoundation)
	}
	parts = append(parts, "source_operation="+source.SourceID)
	return strings.Join(parts, "; ")
}

// sourceProjectionAnnotateReadFoundationGaps gives a retained
// source-bound read the same precise provenance as a mutation gap. The command
// remains unavailable; this removes only inherited planning prose, not the
// declared source requirement or the blocked-by-default guard.
func sourceProjectionAnnotateReadFoundationGaps(operations, cli, surface *orderedObject, result sourceImportResult) sourceProjectionStats {
	stats := sourceProjectionStats{}
	for _, source := range result.Operations {
		if sourceProjectionOperationMutates(source) || !sourceProjectionHasBlockingGap(source.Runtime.Gaps) {
			continue
		}
		operation := sourceProjectionOperationForEndpoint(operations, source.Method, source.Path)
		if operation == nil {
			continue
		}
		operationID := stringField(operation, "id")
		command := sourceProjectionCommandForOperation(cli, operationID, source.Method, source.Path)
		if command == nil || stringField(command, "availability") != "planned" {
			continue
		}
		note := sourceProjectionMutationFoundationNote(source)
		if setOrderedIfDifferent(command, "notes", note) {
			stats.CLI++
		}
		const plannedPrefix = "Planned fixed-target "
		if summary := stringField(command, "summary"); strings.HasPrefix(summary, plannedPrefix) && setOrderedIfDifferent(command, "summary", "Unavailable source-bound "+strings.TrimPrefix(summary, plannedPrefix)) {
			stats.CLI++
		}
		if setOrderedIfDifferent(operation, "description", "Unavailable source-bound provider read: "+note) {
			stats.Operations++
		}
		if surface == nil {
			continue
		}
		for _, raw := range arrayField(surface, "endpoints") {
			endpoint, ok := raw.(*orderedObject)
			if !ok || sourceProjectionEndpointKey(stringField(endpoint, "method"), stringField(endpoint, "path")) != sourceProjectionEndpointKey(source.Method, source.Path) {
				continue
			}
			blocked, declared := endpoint.get("operation")
			classified, ok := blocked.(*orderedObject)
			if !declared || !ok {
				continue
			}
			if setOrderedIfDifferent(classified, "reason", note) {
				stats.Surface++
			}
			if setOrderedIfDifferent(classified, "notes", "source_operation="+source.SourceID) {
				stats.Surface++
			}
		}
	}
	return stats
}

func sourceProjectionReplaceLegacyUnavailableMutationSummary(command *orderedObject, notApplicable bool) bool {
	const prefix = "Planned fixed-target "
	summary := stringField(command, "summary")
	if !strings.HasPrefix(summary, prefix) {
		return false
	}
	replacement := "Unavailable source-bound "
	if notApplicable {
		replacement = "Not-applicable source-bound "
	}
	return setOrderedIfDifferent(command, "summary", replacement+strings.TrimPrefix(summary, prefix))
}

func validateSourceProjectionExecutionEnvelopes(result sourceImportResult) error {
	if result.DescriptorSchemaVersion < 3 {
		return nil
	}
	limits := sourceImportLimits{UseExecutionEnvelopes: true}
	for _, operation := range result.Operations {
		if operation.Protocol == "graphql" {
			continue
		}
		for _, group := range []struct {
			location   string
			parameters []sourceParameterDescriptor
		}{
			{location: "path", parameters: operation.Request.Path},
			{location: "query", parameters: operation.Request.Query},
			{location: "header", parameters: operation.Request.Header},
		} {
			for _, parameter := range group.parameters {
				if err := validateSourceParameterExecutionEnvelope(parameter, group.location, limits); err != nil {
					return fmt.Errorf("source operation %q parameter %q: %w", operation.SourceID, parameter.Name, err)
				}
				if group.location == "header" && sourceScalarWireSchema(parameter.Schema) && sourceBoundedHeaderMaxBytes(parameter.Schema) == 0 && parameter.ExecutionEnvelope == nil && !operation.Runtime.MergeBlocked {
					return fmt.Errorf("source operation %q unbounded header parameter %q is neither enveloped nor merge-blocked", operation.SourceID, parameter.Name)
				}
			}
		}
		if operation.Request.Body != nil {
			if err := validateSourceRequestBodyExecutionEnvelope(operation.Request.Body.ExecutionEnvelope, operation.Request.MediaType, limits); err != nil {
				return fmt.Errorf("source operation %q: %w", operation.SourceID, err)
			}
		}
		for _, media := range operation.Request.Media {
			if err := validateSourceRequestBodyExecutionEnvelope(media.ExecutionEnvelope, media.MediaType, limits); err != nil {
				return fmt.Errorf("source operation %q media %q: %w", operation.SourceID, media.MediaType, err)
			}
		}
	}
	return nil
}

// sourceProjectionReadStats deliberately counts operations separately from
// commands. A read becomes callable only when both sides carry the same locked
// source operation identity; changing just one would be an executable-claim
// drift, not partial progress.
type sourceProjectionReadStats struct {
	Operations int
	Streams    int
	CLI        int
}

const sourceBoundReadMissingFoundationPrefix = "missing_foundation=source-bound-read-execution-r1: "
const sourceProjectionLegacyPlannedReadNote = "Planned ETL/direct read metadata only; no raw provider request execution is exposed."

// sourceProjectionMaterializeReadOperation is the narrow bridge from a
// locked, non-mutating source descriptor to an existing declaration-owned
// operation. It never creates an HTTP route, derives a URL, or turns a list
// into ETL. The inventory operation and command must already name the exact
// method/path; source projection merely seals their common source identity and
// promotes the one executor whose prerequisites are present.
func sourceProjectionMaterializeReadOperation(operations, cli, streams *orderedObject, source sourceOperationDescriptor, inventory []sourceOperationDescriptor) (sourceProjectionReadStats, error) {
	if source.Protocol == "graphql" || !strings.EqualFold(source.Method, "GET") || source.SourceID == "" {
		return sourceProjectionReadStats{}, nil
	}
	if sourceProjectionReadHasBlockingGap(source) {
		return sourceProjectionMarkReadMissingFoundation(cli, source, sourceProjectionReadFoundationReason(source)), nil
	}
	if sourceProjectionSourceEndpointCount(inventory, source.Method, source.Path) != 1 {
		return sourceProjectionMarkReadMissingFoundation(cli, source, "exact source binding is ambiguous for this method/path"), nil
	}
	op := sourceProjectionOperationForEndpoint(operations, source.Method, source.Path)
	if op == nil {
		op = sourceProjectionStreamOperationForEndpoint(operations, cli, streams, source.Method, source.Path)
	}
	if op == nil {
		return sourceProjectionMarkReadMissingFoundation(cli, source, "no declaration-owned REST or stream operation has this exact method/path"), nil
	}
	operationID := stringField(op, "id")
	if operationID == "" {
		return sourceProjectionReadStats{}, fmt.Errorf("operation at %s %s has no id", source.Method, source.Path)
	}
	command := sourceProjectionCommandForOperation(cli, operationID, source.Method, source.Path)
	if command == nil && stringField(op, "kind") == "stream_etl" {
		command = sourceProjectionCommandForStreamOperation(cli, op, source.Method, source.Path)
	}
	if command == nil && stringField(op, "kind") != "stream_etl" {
		return sourceProjectionMarkReadMissingFoundation(cli, source, "no unambiguous canonical command is bound to declaration operation "+operationID), nil
	}

	switch stringField(op, "kind") {
	case "rest_read":
		return sourceProjectionMaterializeDirectRead(op, command, source)
	case "stream_etl":
		return sourceProjectionMaterializeStreamRead(op, command, streams, source)
	default:
		return sourceProjectionMarkReadMissingFoundation(cli, source, "operation "+operationID+" has no read executor kind"), nil
	}
}

func sourceProjectionSourceEndpointCount(inventory []sourceOperationDescriptor, method, path string) int {
	count := 0
	endpoint := sourceProjectionEndpointKey(method, path)
	for _, operation := range inventory {
		if operation.Protocol == "graphql" || sourceProjectionEndpointKey(operation.Method, operation.Path) != endpoint {
			continue
		}
		count++
	}
	return count
}

func sourceProjectionOperationForEndpoint(operations *orderedObject, method, path string) *orderedObject {
	endpoint := sourceProjectionEndpointKey(method, path)
	var match *orderedObject
	for _, raw := range arrayField(operations, "operations") {
		operation, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		restRaw, declared := operation.get("rest")
		rest, ok := restRaw.(*orderedObject)
		if !declared || !ok || sourceProjectionEndpointKey(stringField(rest, "method"), stringField(rest, "path")) != endpoint {
			continue
		}
		// Two declaration operations must not silently share one source route:
		// there would be no way for a caller to select its source identity.
		if match != nil {
			return nil
		}
		match = operation
	}
	return match
}

func sourceProjectionStreamOperationForEndpoint(operations, cli, streams *orderedObject, method, path string) *orderedObject {
	endpoint := sourceProjectionEndpointKey(method, path)
	streamName := ""
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || !sourceProjectionCommandHasEndpoint(command, endpoint) || stringField(command, "stream") == "" {
			continue
		}
		name := stringField(command, "stream")
		if streamName != "" && streamName != name {
			return nil
		}
		streamName = name
	}
	var match *orderedObject
	for _, raw := range arrayField(operations, "operations") {
		operation, ok := raw.(*orderedObject)
		if !ok || stringField(operation, "kind") != "stream_etl" {
			continue
		}
		selectedStream := sourceProjectionOperationStreamName(operation)
		if selectedStream == "" || (streamName != "" && selectedStream != streamName) {
			continue
		}
		if streamName == "" && !sourceProjectionDeclaredStreamEndpointMatches(streams, selectedStream, method, path) {
			continue
		}
		if match != nil {
			return nil
		}
		match = operation
	}
	return match
}

func sourceProjectionDeclaredStreamEndpointMatches(streams *orderedObject, streamName, method, path string) bool {
	for _, raw := range arrayField(streams, "streams") {
		stream, ok := raw.(*orderedObject)
		if !ok || stringField(stream, "name") != streamName {
			continue
		}
		streamMethod := stringField(stream, "method")
		if streamMethod == "" {
			streamMethod = "GET"
		}
		return strings.EqualFold(streamMethod, method) && sourceProjectionStreamPathMatchesSourcePath(stringField(stream, "path"), path)
	}
	return false
}

func sourceProjectionCommandForOperation(cli *orderedObject, operationID, method, path string) *orderedObject {
	endpoint := sourceProjectionEndpointKey(method, path)
	var match *orderedObject
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "operation") != operationID || !sourceProjectionCommandHasEndpoint(command, endpoint) {
			continue
		}
		if match != nil {
			return nil
		}
		match = command
	}
	return match
}

func sourceProjectionCommandForStreamOperation(cli, operation *orderedObject, method, path string) *orderedObject {
	streamName := sourceProjectionOperationStreamName(operation)
	if streamName == "" {
		return nil
	}
	endpoint := sourceProjectionEndpointKey(method, path)
	var match *orderedObject
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || stringField(command, "stream") != streamName || !sourceProjectionCommandHasEndpoint(command, endpoint) {
			continue
		}
		if match != nil {
			return nil
		}
		match = command
	}
	return match
}

func sourceProjectionOperationStreamName(operation *orderedObject) string {
	compositeRaw, declared := operation.get("composite")
	composite, ok := compositeRaw.(*orderedObject)
	if !declared || !ok {
		return ""
	}
	for _, raw := range arrayField(composite, "steps") {
		if step, ok := raw.(string); ok && strings.HasPrefix(step, "stream:") && strings.TrimPrefix(step, "stream:") != "" {
			return strings.TrimPrefix(step, "stream:")
		}
	}
	return ""
}

func sourceProjectionMaterializeDirectRead(operation, command *orderedObject, source sourceOperationDescriptor) (sourceProjectionReadStats, error) {
	restRaw, _ := operation.get("rest")
	rest, _ := restRaw.(*orderedObject)
	stats := sourceProjectionReadStats{}
	if source.Output.Class != sourceOutputJSON {
		return sourceProjectionMarkOneReadMissingFoundation(command, source, "the source response is not a bounded JSON direct-read contract"), nil
	}
	if len(source.Request.Header) != 0 || source.Request.Body != nil || len(source.Request.Media) != 0 {
		return sourceProjectionMarkOneReadMissingFoundation(command, source, "typed request header or body execution is not available for source-bound GET reads"), nil
	}
	if !positiveNumberValue(rest, "max_bytes") || stringField(operation, "output_policy") == "" {
		return sourceProjectionMarkOneReadMissingFoundation(command, source, "operation lacks a bounded response cap or output policy"), nil
	}
	if sourceProjectionSyncReadParameters(rest, source) {
		stats.Operations++
	}
	if changed, err := sourceProjectionSyncReadPagination(rest, source); err != nil {
		return stats, err
	} else if changed {
		stats.Operations++
	}
	if reason := sourceProjectionReadParametersComplete(rest, source); reason != "" {
		missing := sourceProjectionMarkOneReadMissingFoundation(command, source, reason)
		stats.CLI += missing.CLI
		return stats, nil
	}

	if sourceProjectionSetSourceOperation(operation, source) {
		stats.Operations++
	}
	commandChanged := setOrderedIfDifferent(command, "intent", "direct_read")
	if setOrderedIfDifferent(command, "availability", "implemented") {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "source_operation", source.SourceID) {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "api_surface", derivedAPISurface(source.Method, source.Path)) {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "output_policy", stringField(operation, "output_policy")) {
		commandChanged = true
	}
	if deriveCommandParameterFlags(command, rest) > 0 {
		commandChanged = true
	}
	if sourceProjectionRequireSourceReadPathFlags(command, source) {
		commandChanged = true
	}
	if sourceProjectionClearReadMissingFoundation(command) || sourceProjectionClearBlockedReadNote(command, source.SourceID) || sourceProjectionClearHistoricalBlockedReadNote(command) {
		commandChanged = true
	}
	if sourceProjectionClearLegacyPlannedReadNote(command) {
		commandChanged = true
	}
	if sourceProjectionClearLegacyPlannedReadSummary(command) {
		commandChanged = true
	}
	if commandChanged {
		stats.CLI++
	}
	return stats, nil
}

func sourceProjectionMaterializeStreamRead(operation, command, streams *orderedObject, source sourceOperationDescriptor) (sourceProjectionReadStats, error) {
	streamName := ""
	if command == nil || stringField(command, "stream") == "" {
		streamName = sourceProjectionOperationStreamName(operation)
	} else {
		streamName = stringField(command, "stream")
	}
	if streamName == "" {
		return sourceProjectionMarkOneReadMissingFoundation(command, source, "source-backed ETL requires an already implemented named stream"), nil
	}
	if source.Pagination == nil {
		return sourceProjectionMarkOneReadMissingFoundation(command, source, "source pagination semantics are absent; the operation may only be a bounded direct read"), nil
	}
	stats := sourceProjectionReadStats{}
	if sourceProjectionSyncStreamPagination(streams, streamName, source.Pagination) {
		stats.Streams++
	}
	if !sourceProjectionStreamMatchesReadOperation(operation, streams, streamName, source) {
		missing := sourceProjectionMarkOneReadMissingFoundation(command, source, "named stream lacks exact source path, record schema, or pagination semantics")
		missing.Streams += stats.Streams
		return missing, nil
	}
	if sourceProjectionSetSourceOperation(operation, source) {
		stats.Operations++
	}
	if command == nil {
		return stats, nil
	}
	commandChanged := command.remove("operation")
	if setOrderedIfDifferent(command, "intent", "etl") {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "availability", "implemented") {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "stream", streamName) {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "source_operation", source.SourceID) {
		commandChanged = true
	}
	if setOrderedIfDifferent(command, "api_surface", derivedAPISurface(source.Method, source.Path)) {
		commandChanged = true
	}
	if command.remove("flags") {
		commandChanged = true
	}
	if sourceProjectionClearReadMissingFoundation(command) || sourceProjectionClearHistoricalBlockedReadNote(command) {
		commandChanged = true
	}
	if sourceProjectionClearLegacyPlannedReadNote(command) {
		commandChanged = true
	}
	if sourceProjectionClearLegacyPlannedReadSummary(command) {
		commandChanged = true
	}
	if commandChanged {
		stats.CLI++
	}
	return stats, nil
}

// sourceProjectionSyncStreamPagination admits a source-derived pagination
// extension only when the existing declared stream already proves the same
// response-owned pager after those closed controls are removed. This keeps the
// pre-existing ETL route/record contract intact while making the initial page
// use the bounded provider window needed to obtain its next URL.
func sourceProjectionSyncStreamPagination(streams *orderedObject, streamName string, pagination any) bool {
	for _, raw := range arrayField(streams, "streams") {
		stream, ok := raw.(*orderedObject)
		if !ok || stringField(stream, "name") != streamName {
			continue
		}
		if current, declared := stream.get("pagination"); declared {
			if !sourceProjectionPaginationAddsOnlyClosedControls(current, pagination) {
				return false
			}
			return setOrderedIfDifferent(stream, "pagination", pagination)
		}
		baseRaw, declared := streams.get("base")
		base, ok := baseRaw.(*orderedObject)
		if !declared || !ok {
			return false
		}
		current, declared := base.get("pagination")
		if !declared || !sourceProjectionPaginationAddsOnlyClosedControls(current, pagination) {
			return false
		}
		return setOrderedIfDifferent(base, "pagination", pagination)
	}
	return false
}

func sourceProjectionPaginationAddsOnlyClosedControls(current, source any) bool {
	if orderedSemanticEqual(current, source) {
		return true
	}
	derived, ok := source.(map[string]any)
	if !ok {
		return false
	}
	base := make(map[string]any, len(derived))
	for key, value := range derived {
		switch key {
		case "size_param", "limit_param", "offset_param", "page_size":
			continue
		default:
			base[key] = value
		}
	}
	return orderedSemanticEqual(current, base)
}

// sourceProjectionClearLegacyPlannedReadSummary removes the generated
// declaration-era prefix only after the same command has been materialized as
// a source-bound executor.  It deliberately leaves an author-owned summary
// alone unless it matches the historical generated form, so projection never
// uses source import as a general documentation rewriter.
func sourceProjectionClearLegacyPlannedReadSummary(command *orderedObject) bool {
	summary := stringField(command, "summary")
	const prefix = "Planned fixed-target "
	if !strings.HasPrefix(summary, prefix) {
		return false
	}
	readPrefixEnd := strings.Index(summary[len(prefix):], " read: ")
	if readPrefixEnd < 0 {
		return false
	}
	materialized := summary[len(prefix)+readPrefixEnd+len(" read: "):]
	if materialized == "" {
		return false
	}
	return setOrderedIfDifferent(command, "summary", materialized)
}

func sourceProjectionStreamMatchesReadOperation(operation, streams *orderedObject, streamName string, source sourceOperationDescriptor) bool {
	compositeRaw, declared := operation.get("composite")
	composite, ok := compositeRaw.(*orderedObject)
	if !declared || !ok {
		return false
	}
	foundStep := false
	for _, raw := range arrayField(composite, "steps") {
		if name, ok := raw.(string); ok && name == "stream:"+streamName {
			foundStep = true
		}
	}
	if !foundStep {
		return false
	}
	for _, raw := range arrayField(streams, "streams") {
		stream, ok := raw.(*orderedObject)
		if !ok || stringField(stream, "name") != streamName {
			continue
		}
		method := stringField(stream, "method")
		if method == "" {
			method = "GET"
		}
		if !strings.EqualFold(method, source.Method) || !sourceProjectionStreamPathMatchesSourcePath(stringField(stream, "path"), source.Path) {
			return false
		}
		recordsRaw, recordsDeclared := stream.get("records")
		records, recordsOK := recordsRaw.(*orderedObject)
		if !recordsDeclared || !recordsOK || stringField(records, "path") == "" || stringField(stream, "schema") == "" {
			return false
		}
		if pagination, ownPagination := stream.get("pagination"); ownPagination {
			return orderedSemanticEqual(pagination, source.Pagination)
		}
		baseRaw, baseDeclared := streams.get("base")
		base, baseOK := baseRaw.(*orderedObject)
		if !baseDeclared || !baseOK {
			return false
		}
		pagination, inheritedPagination := base.get("pagination")
		return inheritedPagination && orderedSemanticEqual(pagination, source.Pagination)
	}
	return false
}

// sourceProjectionStreamPathMatchesSourcePath permits a stream to use a typed
// connection setting for one complete source path segment.  It does not make
// that setting a route escape hatch: every literal segment must still be the
// locked source path, and both variable spellings must occupy a whole segment.
// This preserves long-standing streams whose configured `workspace_id` is the
// provider's `workspace_gid`, while retaining the provider's exact method/path
// in source_operation and cli_surface.
func sourceProjectionStreamPathMatchesSourcePath(streamPath, sourcePath string) bool {
	streamParts := strings.Split(strings.Trim(streamPath, "/"), "/")
	sourceParts := strings.Split(strings.Trim(sourcePath, "/"), "/")
	if len(streamParts) != len(sourceParts) {
		return false
	}
	for index := range streamParts {
		if streamParts[index] == sourceParts[index] {
			continue
		}
		if (sourceProjectionConfigPathSegment(streamParts[index]) || sourceProjectionFanOutPathSegment(streamParts[index])) && sourceProjectionSourcePathSegment(sourceParts[index]) {
			continue
		}
		return false
	}
	return true
}

func sourceProjectionConfigPathSegment(segment string) bool {
	return strings.HasPrefix(segment, "{{ config.") && strings.HasSuffix(segment, " }}") && len(strings.TrimSuffix(strings.TrimPrefix(segment, "{{ config."), " }}")) > 0
}

func sourceProjectionFanOutPathSegment(segment string) bool {
	return segment == "{{ fanout.id }}"
}

func sourceProjectionSourcePathSegment(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")) > 0
}

func sourceProjectionReadParametersComplete(rest *orderedObject, source sourceOperationDescriptor) string {
	if _, declared := rest.get("parameters"); !declared {
		return "typed operation parameters have not been imported from the locked provider definition"
	}
	parameters := map[string]*orderedObject{}
	for _, raw := range arrayField(rest, "parameters") {
		parameter, ok := raw.(*orderedObject)
		if !ok {
			continue
		}
		location := stringField(parameter, "in")
		name := stringField(parameter, "name")
		if (location == "path" || location == "query") && name != "" {
			parameters[location+"."+name] = parameter
		}
	}
	for _, candidate := range []struct {
		location   string
		parameters []sourceParameterDescriptor
	}{
		{location: "path", parameters: source.Request.Path},
		{location: "query", parameters: source.Request.Query},
	} {
		for _, parameter := range candidate.parameters {
			if candidate.location == "query" && sourceProjectionProviderPagingParameter(parameter) {
				continue
			}
			// A bounded direct read may intentionally omit an optional provider
			// filter. Its command then exposes no lossy substitute for that
			// field; only a caller-required source input must have an executable
			// typed command contract before this operation can advance.
			if !parameter.Required {
				continue
			}
			if !sourceScalarWireSchema(parameter.Schema) {
				return "typed " + candidate.location + " parameter " + parameter.Name + " requires unsupported serialization"
			}
			declared, found := parameters[candidate.location+"."+parameter.Name]
			if !found {
				return "typed " + candidate.location + " parameter " + parameter.Name + " has not been imported from the locked provider definition"
			}
			required, _ := declared.get("required")
			if required != true {
				return "typed " + candidate.location + " parameter " + parameter.Name + " lost requiredness"
			}
		}
	}
	return ""
}

// sourceProjectionSyncReadParameters imports every scalar non-paging caller
// input from the canonical descriptor, including optional filters. Paging
// controls are deliberately excluded: the declared stream/operation pagination
// contract is navigated only by --page or --page-cursor. A zero-input read
// receives an explicit empty parameter list so its bounded direct-read contract
// is distinguishable from one whose typed inputs were never imported. Optional
// non-scalar filters deliberately remain omitted: the one-page direct read
// stays useful without inventing a serialization surface that the source
// contract did not prove PM can execute.
func sourceProjectionSyncReadParameters(rest *orderedObject, source sourceOperationDescriptor) bool {
	parameters := arrayField(rest, "parameters")
	_, declaredParameters := rest.get("parameters")
	changed := !declaredParameters
	kept := parameters[:0]
	for _, raw := range parameters {
		parameter, ok := raw.(*orderedObject)
		if !ok {
			kept = append(kept, raw)
			continue
		}
		location := stringField(parameter, "in")
		name := stringField(parameter, "name")
		if (location == "path" || location == "query" || location == "header") && !sourceProjectionReadParameterAdmitted(source, location, name) {
			changed = true
			continue
		}
		kept = append(kept, raw)
	}
	parameters = kept
	for _, group := range []struct {
		location   string
		parameters []sourceParameterDescriptor
	}{
		{location: "path", parameters: source.Request.Path},
		{location: "query", parameters: source.Request.Query},
	} {
		for _, sourceParameter := range group.parameters {
			if group.location == "query" && sourceProjectionProviderPagingParameter(sourceParameter) {
				continue
			}
			if !sourceScalarWireSchema(sourceParameter.Schema) {
				continue
			}
			typeName := sourceProjectionReadParameterType(sourceParameter.Schema)
			if typeName == "" {
				continue
			}
			var declared *orderedObject
			for _, raw := range parameters {
				parameter, ok := raw.(*orderedObject)
				if ok && stringField(parameter, "in") == group.location && stringField(parameter, "name") == sourceParameter.Name {
					declared = parameter
					break
				}
			}
			if declared == nil {
				declared = newOrderedObject()
				declared.set("name", sourceParameter.Name)
				declared.set("in", group.location)
				parameters = append(parameters, declared)
				changed = true
			}
			if setOrderedIfDifferent(declared, "type", typeName) {
				changed = true
			}
			if sourceParameter.Required {
				if required, _ := declared.get("required"); required != true {
					declared.set("required", true)
					changed = true
				}
			} else if declared.remove("required") {
				changed = true
			}
			if sourceProjectionSetReadParameterEnum(declared, sourceParameter.Schema) {
				changed = true
			}
		}
	}
	if sourceProjectionSyncReadStaticQuery(rest, source) {
		changed = true
	}
	if sourceProjectionSyncReadBody(rest, source) {
		changed = true
	}
	if changed {
		rest.set("parameters", parameters)
	}
	return changed
}

// sourceProjectionReadParameterAdmitted is the local-to-source half of the
// source-bound read closure. A source descriptor is allowed to omit optional
// caller filters from a bounded direct read, but a local request parameter
// must never manufacture a path/query/header channel the retained provider
// operation did not declare. Paging is deliberately excluded here: its only
// admitted representation is the operation's closed pagination contract.
func sourceProjectionReadParameterAdmitted(source sourceOperationDescriptor, location, name string) bool {
	if name == "" {
		return false
	}
	for _, group := range []struct {
		location   string
		parameters []sourceParameterDescriptor
	}{
		{location: "path", parameters: source.Request.Path},
		{location: "query", parameters: source.Request.Query},
		{location: "header", parameters: source.Request.Header},
	} {
		if location != group.location {
			continue
		}
		for _, parameter := range group.parameters {
			if parameter.Name != name {
				continue
			}
			return location != "query" || !sourceProjectionProviderPagingParameter(parameter)
		}
	}
	return false
}

func sourceProjectionSyncReadStaticQuery(rest *orderedObject, source sourceOperationDescriptor) bool {
	raw, declared := rest.get("query")
	if !declared {
		return false
	}
	query, ok := raw.(*orderedObject)
	if !ok {
		return false
	}
	changed := false
	for _, name := range append([]string(nil), query.keys...) {
		if sourceProjectionReadParameterAdmitted(source, "query", name) {
			continue
		}
		query.remove(name)
		changed = true
	}
	if len(query.keys) == 0 {
		return rest.remove("query") || changed
	}
	return changed
}

// A source-bound GET whose retained request has no body/media cannot gain a
// declaration-owned body or content-type channel. If the source does declare
// either class, materialization stops at its explicit shared-foundation gap
// before reaching this direct-read projection.
func sourceProjectionSyncReadBody(rest *orderedObject, source sourceOperationDescriptor) bool {
	if source.Request.Body != nil || len(source.Request.Media) != 0 || source.Request.MediaType != "" {
		return false
	}
	changed := rest.remove("body")
	if rest.remove("body_schema") {
		changed = true
	}
	if rest.remove("content_type") {
		changed = true
	}
	return changed
}

// sourceProjectionSyncReadPagination materializes source-proven pagination in
// the operation-owned block. The controls stay out of rest.parameters, so
// surface-sync can never make a raw paging flag; pagination_parameters proves
// that the closed engine dialect is backed by this operation's provider
// contract.
func sourceProjectionSyncReadPagination(rest *orderedObject, source sourceOperationDescriptor) (bool, error) {
	if source.Pagination == nil {
		changed := rest.remove("pagination")
		if rest.remove("pagination_parameters") {
			changed = true
		}
		return changed, nil
	}
	pagination, ok := source.Pagination.(map[string]any)
	if !ok {
		return false, fmt.Errorf("source pagination is not an object")
	}
	names := sourceProjectionPaginationParameterNames(pagination)
	byName := make(map[string]sourceParameterDescriptor, len(source.Request.Query))
	for _, parameter := range source.Request.Query {
		byName[parameter.Name] = parameter
	}
	parameters := make([]any, 0, len(names))
	for _, name := range names {
		parameter, found := byName[name]
		if !found || !sourceScalarWireSchema(parameter.Schema) {
			return false, fmt.Errorf("source pagination parameter %q is not a scalar query parameter", name)
		}
		entry := newOrderedObject()
		entry.set("name", name)
		entry.set("in", "query")
		if typeName := sourceProjectionReadParameterType(parameter.Schema); typeName != "" {
			entry.set("type", typeName)
		}
		if parameter.Required {
			entry.set("required", true)
		}
		parameters = append(parameters, entry)
	}

	changed := setOrderedIfDifferent(rest, "pagination", pagination)
	if len(parameters) == 0 {
		if rest.remove("pagination_parameters") {
			changed = true
		}
		return changed, nil
	}
	if setOrderedIfDifferent(rest, "pagination_parameters", parameters) {
		changed = true
	}
	return changed, nil
}

func sourceProjectionPaginationParameterNames(pagination map[string]any) []string {
	field := func(key string) string {
		value, _ := pagination[key].(string)
		return strings.TrimSpace(value)
	}
	names := make([]string, 0, 3)
	add := func(name string) {
		if name == "" {
			return
		}
		for _, existing := range names {
			if existing == name {
				return
			}
		}
		names = append(names, name)
	}
	switch field("type") {
	case "page_number":
		add(field("page_param"))
		add(field("size_param"))
	case "offset_limit":
		add(field("offset_param"))
		add(field("limit_param"))
		add(field("size_param"))
	case "cursor":
		add(field("cursor_param"))
		add(field("size_param"))
	case "start_index":
		add(field("start_index_param"))
		add(field("count_param"))
		add(field("size_param"))
	case "next_url":
		// A next URL owns the next page, but a documented initial size and
		// continuation offset are still part of its closed contract.
		add(field("size_param"))
		add(field("limit_param"))
		add(field("offset_param"))
	case "link_header":
		add(field("size_param"))
	}
	return names
}

// sourceProjectionProviderPagingParameter keeps source-projected reads on the
// same closed paging classifier as params-import. Source descriptors retain
// only typed parameter shape, so provider paging descriptions are unavailable
// here; the shared semantic-name classifier still prevents a second raw
// navigation channel.
func sourceProjectionProviderPagingParameter(parameter sourceParameterDescriptor) bool {
	return isProviderPagingParameter(openAPIParameter{Name: parameter.Name, In: "query"})
}

func sourceProjectionReadParameterType(schema any) string {
	if sourceStringScalarWireUnion(schema) {
		return "string"
	}
	return sourceProjectionFlagType(schema)
}

func sourceProjectionSetReadParameterEnum(parameter *orderedObject, schema any) bool {
	object, _ := schema.(map[string]any)
	values := sourceAnySlice(object["enum"])
	if len(values) == 0 {
		return parameter.remove("values")
	}
	text := make([]any, 0, len(values))
	for _, value := range values {
		stringValue, ok := value.(string)
		if !ok {
			return parameter.remove("values")
		}
		text = append(text, stringValue)
	}
	return setOrderedIfDifferent(parameter, "values", text)
}

func sourceProjectionSetSourceOperation(operation *orderedObject, source sourceOperationDescriptor) bool {
	binding := newOrderedObject()
	binding.set("id", source.SourceID)
	binding.set("method", strings.ToUpper(strings.TrimSpace(source.Method)))
	binding.set("path", source.Path)
	return setOrderedIfDifferent(operation, "source_operation", binding)
}

func sourceProjectionRequireSourceReadPathFlags(command *orderedObject, source sourceOperationDescriptor) bool {
	changed := false
	required := map[string]bool{}
	for _, parameter := range source.Request.Path {
		if parameter.Required {
			required["path."+parameter.Name] = true
		}
	}
	for _, raw := range arrayField(command, "flags") {
		flag, ok := raw.(*orderedObject)
		if !ok || !required[stringField(flag, "maps_to")] {
			continue
		}
		if present, _ := flag.get("required"); present != true {
			flag.set("required", true)
			changed = true
		}
	}
	return changed
}

func sourceProjectionMarkReadMissingFoundation(cli *orderedObject, source sourceOperationDescriptor, reason string) sourceProjectionReadStats {
	stats := sourceProjectionReadStats{}
	for _, raw := range arrayField(cli, "commands") {
		command, ok := raw.(*orderedObject)
		if !ok || !sourceProjectionCommandHasEndpoint(command, sourceProjectionEndpointKey(source.Method, source.Path)) {
			continue
		}
		child := sourceProjectionMarkOneReadMissingFoundation(command, source, reason)
		stats.CLI += child.CLI
	}
	return stats
}

func sourceProjectionMarkOneReadMissingFoundation(command *orderedObject, source sourceOperationDescriptor, reason string) sourceProjectionReadStats {
	changed := false
	if stringField(command, "availability") == "implemented" {
		if stringField(command, "source_operation") != source.SourceID {
			// A pre-foundation command may already have its own working executor.
			// Do not downgrade it merely because this source-bound bridge cannot
			// yet describe the full provider contract. Only a command the bridge
			// previously bound to this exact source identity can be reclassified.
			return sourceProjectionReadStats{}
		}
		// An implemented command may retain a source_operation binding from an
		// earlier projection while its present locked source facts no longer
		// establish the executor it claims (for example a stream whose declared
		// pagination differs from the provider response).  It must stop claiming
		// implementation until the declaration and source contract agree again.
		command.set("availability", "planned")
		changed = true
	}
	note := sourceBoundReadMissingFoundationPrefix + reason + "; source_operation=" + source.SourceID
	if sourceProjectionReadHasBlockingGap(source) {
		note = sourceProjectionMutationFoundationNote(source)
		const plannedPrefix = "Planned fixed-target "
		if summary := stringField(command, "summary"); strings.HasPrefix(summary, plannedPrefix) {
			command.set("summary", "Unavailable source-bound "+strings.TrimPrefix(summary, plannedPrefix))
			changed = true
		}
	}
	if setOrderedIfDifferent(command, "notes", note) {
		changed = true
	}
	if changed {
		return sourceProjectionReadStats{CLI: 1}
	}
	return sourceProjectionReadStats{}
}

func sourceProjectionReadFoundationReason(source sourceOperationDescriptor) string {
	foundations := make([]string, 0, len(source.Runtime.Gaps))
	for _, gap := range source.Runtime.Gaps {
		if !sourceProjectionHasBlockingGap([]sourceContractGap{gap}) {
			continue
		}
		foundations = append(foundations, gap.Foundation)
	}
	sort.Strings(foundations)
	if len(foundations) == 0 {
		return "missing foundation " + sourceOperationExecutionFoundation
	}
	return "missing foundation " + strings.Join(foundations, ",")
}

func sourceProjectionClearReadMissingFoundation(command *orderedObject) bool {
	notes := stringField(command, "notes")
	if strings.HasPrefix(notes, sourceBoundReadMissingFoundationPrefix) || strings.HasPrefix(notes, "missing_foundation=closed-source-operation-execution-foundation-r1; source_operation=") {
		return command.remove("notes")
	}
	return false
}

func sourceProjectionClearLegacyPlannedReadNote(command *orderedObject) bool {
	if stringField(command, "notes") != sourceProjectionLegacyPlannedReadNote {
		return false
	}
	return command.remove("notes")
}

// sourceProjectionClearBlockedReadNote removes only the exact source-derived
// block that a previous projection wrote.  It must not erase an author note:
// those can describe a broader workflow dependency than this one operation.
func sourceProjectionClearBlockedReadNote(command *orderedObject, sourceID string) bool {
	if stringField(command, "notes") != sourceProjectionBlockedReadCommandNote(sourceID) {
		return false
	}
	return command.remove("notes")
}

// sourceProjectionClearHistoricalBlockedReadNote removes the old generated
// blocker wording only after this same command has been promoted from a
// complete source contract. A real current foundation remains represented by
// the exact missing_foundation note and is never cleared here.
func sourceProjectionClearHistoricalBlockedReadNote(command *orderedObject) bool {
	if !strings.HasPrefix(stringField(command, "notes"), "Blocked until ") {
		return false
	}
	return command.remove("notes")
}

func positiveNumberValue(object *orderedObject, key string) bool {
	value, _ := object.get(key)
	return positiveNumber(value)
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
	SchemaVersion               int                                        `json:"schema_version"`
	Dispositions                []sourceNonExecutableMutationDisposition   `json:"dispositions"`
	PartialCoverageDispositions []sourcePartialMutationCoverageDisposition `json:"partial_coverage_dispositions,omitempty"`
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

// sourceProjectionReadPartialMutationCoverageDispositions reads the
// operation-granular exception for a working action whose complete provider
// request contract needs a missing shared foundation. It is intentionally
// separate from a non-executable disposition: neither form can disguise the
// other, nor an unsupported provider operation.
func sourceProjectionReadPartialMutationCoverageDispositions(bundleDir string) ([]sourcePartialMutationCoverageDisposition, error) {
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
	seen := make(map[string]bool, len(document.PartialCoverageDispositions))
	for _, disposition := range document.PartialCoverageDispositions {
		if err := sourceProjectionValidatePartialMutationCoverageDispositionInput(disposition); err != nil {
			return nil, err
		}
		if seen[disposition.Source.SourceID] {
			return nil, fmt.Errorf("partial mutation coverage dispositions duplicate source operation %q", disposition.Source.SourceID)
		}
		seen[disposition.Source.SourceID] = true
	}
	return document.PartialCoverageDispositions, nil
}

func sourceProjectionValidateNonExecutableMutationDispositionInput(disposition sourceNonExecutableMutationDisposition) error {
	return sourceProjectionValidateMutationDispositionInput("mutation disposition", disposition.Source, disposition.Reason)
}

func sourceProjectionValidatePartialMutationCoverageDispositionInput(disposition sourcePartialMutationCoverageDisposition) error {
	if err := sourceProjectionValidateMutationDispositionInput("partial mutation coverage disposition", disposition.Source, disposition.Reason); err != nil {
		return err
	}
	if disposition.Foundation == "" || disposition.Foundation != strings.TrimSpace(disposition.Foundation) || len(disposition.Foundation) > 256 || strings.ContainsAny(disposition.Foundation, "\r\n\x00") {
		return fmt.Errorf("partial mutation coverage disposition foundation is invalid")
	}
	return nil
}

func sourceProjectionValidateMutationDispositionInput(kind string, source sourceOperationCitation, reason string) error {
	for _, value := range []struct {
		name  string
		value string
		max   int
	}{
		{name: "source_id", value: source.SourceID, max: 1024},
		{name: "method", value: source.Method, max: 16},
		{name: "path", value: source.Path, max: 4096},
		{name: "reason", value: reason, max: 1024},
	} {
		if value.value == "" || value.value != strings.TrimSpace(value.value) || len(value.value) > value.max || strings.ContainsAny(value.value, "\r\n\x00") {
			return fmt.Errorf("%s %s is invalid", kind, value.name)
		}
	}
	if !sourceProjectionMutationMethod(source.Method) {
		return fmt.Errorf("%s source method %q is not mutating", kind, source.Method)
	}
	if !strings.HasPrefix(source.Path, "/") {
		return fmt.Errorf("%s source path %q is invalid", kind, source.Path)
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
			if result.Operations[index].Runtime.NonExecutableMutation != nil || result.Operations[index].Runtime.PartialCoverageMutation != nil {
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

// sourceProjectionApplyPartialMutationCoverageDispositions attaches an exact
// provider citation to an implemented mutation whose declared action remains
// deliberately narrower than the source request contract. It cannot suppress
// an absent action, a source-complete action, an unsupported provider route,
// or a read.
func sourceProjectionApplyPartialMutationCoverageDispositions(bundle engine.Bundle, result *sourceImportResult, dispositions []sourcePartialMutationCoverageDisposition) error {
	if len(dispositions) == 0 {
		return nil
	}
	if result == nil {
		return fmt.Errorf("partial mutation coverage dispositions require source operations")
	}
	operations := sourceProjectionOperationsByID(*result)
	seen := make(map[string]bool, len(dispositions))
	for _, disposition := range dispositions {
		if err := sourceProjectionValidatePartialMutationCoverageDispositionInput(disposition); err != nil {
			return err
		}
		if seen[disposition.Source.SourceID] {
			return fmt.Errorf("partial mutation coverage dispositions duplicate source operation %q", disposition.Source.SourceID)
		}
		seen[disposition.Source.SourceID] = true
		operation, found := operations[disposition.Source.SourceID]
		if !found {
			return fmt.Errorf("partial mutation coverage disposition cites unknown source operation %q", disposition.Source.SourceID)
		}
		if err := sourceProjectionValidatePartialMutationCoverageDispositionCitation(operation, disposition); err != nil {
			return err
		}
		if sourceProjectionMutationActionIsComplete(bundle, operation) {
			return fmt.Errorf("partial mutation coverage disposition source operation %q already has a complete executable action", operation.SourceID)
		}
		if !sourceProjectionMutationClaimsImplementedAction(bundle, operation) {
			return fmt.Errorf("partial mutation coverage disposition source operation %q has no implemented declared action", operation.SourceID)
		}
		if !sourceProjectionPartialCoverageFoundationMatchesOperation(bundle, operation, disposition) {
			return fmt.Errorf("partial mutation coverage disposition source operation %q has no matching missing foundation %q", operation.SourceID, disposition.Foundation)
		}
		for index := range result.Operations {
			if result.Operations[index].SourceID != operation.SourceID {
				continue
			}
			if result.Operations[index].Runtime.NonExecutableMutation != nil || result.Operations[index].Runtime.PartialCoverageMutation != nil {
				return fmt.Errorf("source operation %q already has a mutation disposition", operation.SourceID)
			}
			copyDisposition := disposition
			result.Operations[index].Runtime.PartialCoverageMutation = &copyDisposition
			result.Operations[index].Runtime.Gaps = sourceSortedGaps(append(result.Operations[index].Runtime.Gaps, sourceProjectionPartialMutationCoverageRuntimeGap(result.Operations[index], copyDisposition)))
			result.Operations[index].Runtime.MergeBlocked = true
			break
		}
	}
	return nil
}

// sourceProjectionApplyWriteDisabledMutationArtifacts retains a provider
// mutation when the connector has explicitly opted out of writes. It uses the
// same source-cited non-executable mutation artifact as a manual disposition,
// but derives it from the locked provider operation on every source import so
// no local source document, action, request schema, transport, or CLI command
// has to be invented merely to make the read surface validate.
//
// An executable action or implemented action claim always wins. Those
// operations remain ordinary executable coverage, including delete and reverse
// ETL actions; the write-disabled declaration is never a safety suppression.
func sourceProjectionApplyWriteDisabledMutationArtifacts(bundle engine.Bundle, result *sourceImportResult) int {
	if result == nil || !sourceProjectionExplicitWriteDisabled(bundle) {
		return 0
	}

	artifacts := 0
	for index := range result.Operations {
		operation := &result.Operations[index]
		if !sourceProjectionOperationMutates(*operation) || operation.Runtime.NonExecutableMutation != nil {
			continue
		}
		if _, declared, err := sourceProjectionReadOnlyDeclaration(bundle, *operation); err != nil || declared {
			// A malformed or mutating `read_only` row must remain a validation
			// failure. Do not cover it with a separate mutation artifact.
			continue
		}
		if sourceProjectionMutationActionIsComplete(bundle, *operation) || sourceProjectionMutationClaimsImplementedAction(bundle, *operation) {
			continue
		}
		disposition := sourceNonExecutableMutationDisposition{
			Source: sourceOperationCitation{
				SourceID: operation.SourceID,
				Method:   operation.Method,
				Path:     operation.Path,
			},
			Reason: sourceWriteDisabledMutationArtifactReason,
		}
		if sourceProjectionValidateNonExecutableMutationDispositionCitation(*operation, disposition) != nil {
			// Source import normally guarantees this provenance. Keep the helper
			// fail-closed as well so callers cannot synthesize an artifact from a
			// mutation-shaped value without a retained provider citation.
			continue
		}
		operation.Runtime.NonExecutableMutation = &disposition
		operation.Runtime.Gaps = sourceSortedGaps(append(operation.Runtime.Gaps, sourceProjectionNonExecutableMutationRuntimeGap(*operation, disposition)))
		operation.Runtime.MergeBlocked = true
		artifacts++
	}
	return artifacts
}

func sourceProjectionValidateNonExecutableMutationDispositionCitation(operation sourceOperationDescriptor, disposition sourceNonExecutableMutationDisposition) error {
	return sourceProjectionValidateMutationDispositionCitation(operation, disposition.Source, "mutation disposition")
}

func sourceProjectionValidatePartialMutationCoverageDispositionCitation(operation sourceOperationDescriptor, disposition sourcePartialMutationCoverageDisposition) error {
	return sourceProjectionValidateMutationDispositionCitation(operation, disposition.Source, "partial mutation coverage disposition")
}

func sourceProjectionValidateMutationDispositionCitation(operation sourceOperationDescriptor, citation sourceOperationCitation, kind string) error {
	if !sourceProjectionOperationMutates(operation) {
		return fmt.Errorf("%s source operation %q is not mutating", kind, operation.SourceID)
	}
	if operation.Source.URL == "" || operation.Source.Location == "" {
		return fmt.Errorf("%s source operation %q lacks a provider source citation", kind, operation.SourceID)
	}
	if citation.SourceID != operation.SourceID || !strings.EqualFold(citation.Method, operation.Method) || citation.Path != operation.Path {
		return fmt.Errorf("%s citation does not match provider source operation %q", kind, operation.SourceID)
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

func sourceProjectionPartialMutationCoverageRuntimeGap(operation sourceOperationDescriptor, disposition sourcePartialMutationCoverageDisposition) sourceContractGap {
	return sourceContractGapFor(
		sourcePartialMutationCoverageFoundation,
		"source operation "+operation.SourceID+" at "+operation.Source.URL+"#"+operation.Source.Location,
		"missing_foundation: provider-cited implemented mutation retains only its declared typed subset: "+disposition.Reason,
	)
}

func sourceProjectionHasPartialMutationCoverageDisposition(operation sourceOperationDescriptor) bool {
	disposition := operation.Runtime.PartialCoverageMutation
	if disposition == nil || !operation.Runtime.MergeBlocked || sourceProjectionValidatePartialMutationCoverageDispositionInput(*disposition) != nil || sourceProjectionValidatePartialMutationCoverageDispositionCitation(operation, *disposition) != nil {
		return false
	}
	want := sourceProjectionPartialMutationCoverageRuntimeGap(operation, *disposition)
	for _, gap := range operation.Runtime.Gaps {
		if gap == want {
			return true
		}
	}
	return false
}

func sourceProjectionPartialCoverageFoundationMatchesOperation(bundle engine.Bundle, operation sourceOperationDescriptor, disposition sourcePartialMutationCoverageDisposition) bool {
	if disposition.Foundation == "source-path-parameter-alias-foundation-r1" {
		return sourceProjectionMutationHasImplementedPathParameterAlias(bundle, operation)
	}
	for _, gap := range operation.Runtime.Gaps {
		if gap.Foundation == disposition.Foundation && sourceProjectionHasBlockingGap([]sourceContractGap{gap}) {
			return true
		}
	}
	return false
}

// sourceProjectionMutationHasImplementedPathParameterAlias is the one legacy
// partial-coverage category that is not represented by an importer gap: older
// actions may name their typed record path field differently from the locked
// provider parameter (for example, local gid versus provider project_gid).
// Require that exact evidence from the declared, implemented action; an
// arbitrary incomplete mutation cannot select this foundation name.
func sourceProjectionMutationHasImplementedPathParameterAlias(bundle engine.Bundle, operation sourceOperationDescriptor) bool {
	if bundle.CLISurface == nil || len(operation.Request.Path) == 0 {
		return false
	}
	endpoint := sourceProjectionEndpointKey(operation.Method, operation.Path)
	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}
	for _, command := range bundle.CLISurface.Commands {
		if command.Availability != "implemented" || command.Write == "" {
			continue
		}
		matchesSourceEndpoint := false
		for _, surface := range command.APISurface {
			if sourceProjectionEndpointKey(surface.Method, surface.Path) == endpoint {
				matchesSourceEndpoint = true
				break
			}
		}
		if !matchesSourceEndpoint {
			continue
		}
		action, found := actions[command.Write]
		if !found {
			continue
		}
		pathFields := make(map[string]bool, len(action.PathFields))
		for _, field := range action.PathFields {
			pathFields[field] = true
		}
		for _, parameter := range operation.Request.Path {
			if !pathFields[parameter.Name] {
				return true
			}
		}
	}
	return false
}

// sourceProjectionValidateWriteDisabledMutationArtifact ensures the automatic
// artifact cannot be copied into a write-capable bundle to waive source
// executable coverage. Manually authored non-executable mutation dispositions
// retain their existing policy; this applies only to the exact automatic
// write-disabled reason.
func sourceProjectionValidateWriteDisabledMutationArtifact(bundle engine.Bundle, operation sourceOperationDescriptor) error {
	disposition := operation.Runtime.NonExecutableMutation
	if disposition == nil || disposition.Reason != sourceWriteDisabledMutationArtifactReason {
		return nil
	}
	if !sourceProjectionExplicitWriteDisabled(bundle) {
		return errors.New("automatic write-disabled mutation artifact requires connector metadata capabilities.write=false")
	}
	return nil
}

func sourceProjectionExplicitWriteDisabled(bundle engine.Bundle) bool {
	return bundle.Metadata.Name != "" && bundle.Metadata.Capabilities.WriteDeclared && !bundle.Metadata.Capabilities.Write
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
		if sourceProjectionOptionalParameterSchemaGap(operation, gap) || sourceProjectionOmittedOptionalRequestBodySchemaGap(operation, gap) {
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
			if sourceProjectionOptionalParameterSchemaGap(*operation, gap) || sourceProjectionOmittedOptionalRequestBodySchemaGap(*operation, gap) {
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

func sourceProjectionOptionalParameterSchemaGap(operation sourceOperationDescriptor, gap sourceContractGap) bool {
	if gap.Foundation != "cli-request-schema-foundation-r1" || !strings.HasPrefix(gap.Location, "parameter ") {
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
		reason := sourceProjectionBlockedReadSurfaceReason(source.SourceID)
		note := sourceProjectionBlockedReadSurfaceNote(source.SourceID)
		if sourceProjectionReadHasBlockingGap(source) {
			reason = sourceProjectionMutationFoundationNote(source)
			note = "source_operation=" + source.SourceID
		}
		operation.set("reason", reason)
		operation.set("notes", note)
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
	if sourceProjectionOperationMutates(source) {
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
	// Source IDs are provider-owned provenance, not a command grammar: several
	// OpenAPI importers use the raw method/path as their ID, including `{name}`
	// parameter syntax that commandrunner rejects. Encode the normalized endpoint
	// instead. Hex is injective over the complete endpoint key, so a literal and
	// a path parameter (or two parameter names) cannot collapse to one command.
	endpoint := sourceProjectionEndpointKey(operation.Method, operation.Path)
	return "api op-" + hex.EncodeToString([]byte(endpoint))
}

func sourceProjectionLegacyGeneratedCommandPath(operation sourceOperationDescriptor) string {
	path := strings.NewReplacer("/", " ", "_", "-").Replace(operation.SourceID)
	return "api " + path
}

func sourceProjectionCommandPathIsRuntimeValid(path string) bool {
	for index, segment := range strings.Fields(path) {
		if err := safety.ValidateIdentifier(segment, fmt.Sprintf("command path segment %d", index+1)); err != nil {
			return false
		}
	}
	return strings.TrimSpace(path) != ""
}

// sourceProjectionRefreshGeneratedCommandMetadata refreshes only the command
// identity generated from the immutable source descriptor. Author-owned aliases
// keep their own prose, while the generated command must continue to publish
// the declaration's actual approval lifecycle after a later projection pass.
func sourceProjectionRefreshGeneratedCommandMetadata(command *orderedObject, operation sourceOperationDescriptor, action *orderedObject) bool {
	currentPath := stringField(command, "path")
	generatedPath := sourceProjectionGeneratedCommandPath(operation)
	if currentPath == generatedPath {
		return setOrderedIfDifferent(command, "approval", sourceProjectionApproval(action))
	}
	if currentPath != sourceProjectionLegacyGeneratedCommandPath(operation) {
		return false
	}
	if sourceProjectionCommandPathIsRuntimeValid(currentPath) {
		return setOrderedIfDifferent(command, "approval", sourceProjectionApproval(action))
	}
	command.set("path", generatedPath)
	setOrderedIfDifferent(command, "approval", sourceProjectionApproval(action))
	return true
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
	if unavailable, exists := sourceImportUnavailableDocument(lock); exists {
		return []Finding{sourceProjectionFinding(bundle.Name, lockPath, sourceImportUnavailableFindingMessage(unavailable))}
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

func sourceImportUnavailableDocument(lock sourceImportLock) (sourceImportRESTDocument, bool) {
	if lock.SchemaVersion != 3 {
		return sourceImportRESTDocument{}, false
	}
	for _, document := range lock.Rest.SourceDocuments {
		if document.isUnavailable() {
			return document, true
		}
	}
	return sourceImportRESTDocument{}, false
}

func sourceImportUnavailableFindingMessage(document sourceImportRESTDocument) string {
	if document.PublishedSource.SourceURL != "" {
		return fmt.Sprintf("source inventory is unavailable: document %q cites %s: %s", document.ID, document.PublishedSource.SourceURL, document.UnavailableReason)
	}
	return fmt.Sprintf("source inventory is unavailable: document %q: %s", document.ID, document.UnavailableReason)
}

func sourceProjectionFinding(connector, file, message string) Finding {
	return Finding{Connector: connector, File: strings.TrimPrefix(file, connector+"/"), Rule: ruleSourceProjection, Message: message}
}

func validateSourceDescriptorAgainstLock(connector, file string, lock sourceImportLock, descriptor sourceImportDescriptorDocument) []Finding {
	if unavailable, exists := sourceImportUnavailableDocument(lock); exists {
		return []Finding{sourceProjectionFinding(connector, file, sourceImportUnavailableFindingMessage(unavailable))}
	}
	wantSchemaVersion := 2
	if lock.SchemaVersion == 3 || lock.isLegacySourceReference() {
		wantSchemaVersion = 3
	}
	if descriptor.SchemaVersion != wantSchemaVersion {
		return []Finding{sourceProjectionFinding(connector, file, fmt.Sprintf("source descriptor schema_version = %d, want %d", descriptor.SchemaVersion, wantSchemaVersion))}
	}
	type expectedSource struct {
		source              sourceImportSource
		providerOperationID string
		reference           *sourceOperationDescriptor
		method              string
		path                string
	}
	expected := map[string]expectedSource{}
	referenceGaps := []sourceContractGap{}
	referenceCount := 0
	if lock.SchemaVersion == 3 {
		for _, document := range lock.Rest.SourceDocuments {
			for _, operation := range document.Operations {
				source := sourceImportExpectedV3DescriptorProvenance(document, operation)
				expectedOperation := expectedSource{
					source:              source,
					providerOperationID: operation.OperationID,
					method:              strings.ToUpper(operation.Method),
					path:                operation.Path,
				}
				if document.isSourceReference() {
					reference := sourceImportReferenceOperation(lock.Connector, operation, document.sourceArtifact(), sourceImportReferenceForm(document.sourceArtifact()))
					reference.Source.DocumentID = document.ID
					expectedOperation.reference = &reference
					referenceGaps = append(referenceGaps, reference.Runtime.Gaps...)
					referenceCount++
				}
				expected[operation.ID] = expectedOperation
			}
		}
	} else if lock.isLegacySourceReference() {
		artifacts, err := sourceImportLegacyReferenceArtifacts(lock)
		if err != nil {
			return []Finding{sourceProjectionFinding(connector, file, "source-reference provenance: "+err.Error())}
		}
		for _, operation := range lock.Rest.Operations {
			artifact, found := artifacts[operation.SourceURL]
			if !found {
				return []Finding{sourceProjectionFinding(connector, file, "source-reference operation "+operation.ID+" cites an undeclared source URL")}
			}
			reference := sourceImportReferenceOperation(lock.Connector, operation, artifact, sourceImportReferenceForm(artifact))
			referenceGaps = append(referenceGaps, reference.Runtime.Gaps...)
			referenceCount++
			expected[operation.ID] = expectedSource{
				source:              sourceImportReferenceProvenance(artifact, operation.SourceLocation, sourceImportReferenceForm(artifact)),
				providerOperationID: operation.OperationID,
				reference:           &reference,
			}
		}
	} else {
		for _, operation := range lock.Rest.Operations {
			identity := operation.OperationID
			if identity == "" {
				identity = operation.ID
			}
			expected[identity] = expectedSource{source: sourceImportSource{URL: lock.Rest.SourceURL, Location: operation.SourceLocation}, method: strings.ToUpper(operation.Method), path: operation.Path}
		}
	}
	for _, field := range lock.GraphQL.QueryFields {
		expected[fmt.Sprintf("%s.graphql.query.%s", connector, field.Name)] = expectedSource{source: sourceImportSource{SHA256: strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)), Bytes: firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes)}}
	}
	for _, field := range lock.GraphQL.MutationFields {
		expected[fmt.Sprintf("%s.graphql.mutation.%s", connector, field.Name)] = expectedSource{source: sourceImportSource{SHA256: strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)), Bytes: firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes)}}
	}
	if referenceCount > 0 && !descriptor.MergeBlocked {
		return []Finding{sourceProjectionFinding(connector, file, "source descriptor reference contract drift: merge_blocked must remain true")}
	}
	if referenceCount == len(expected) && !reflect.DeepEqual(sourceSortedGaps(descriptor.Gaps), sourceSortedGaps(referenceGaps)) {
		return []Finding{sourceProjectionFinding(connector, file, "source descriptor reference contract drift: descriptor gaps do not match the cited source contract")}
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
		if expectedOperation.reference != nil && !reflect.DeepEqual(operation, *expectedOperation.reference) {
			// A cited-only source has no request or response contract to repair
			// downstream. Its exact operation descriptor is therefore the closed
			// lock projection: identity, provenance, empty execution fields, and
			// the sole source_contract_unavailable gap all travel together.
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor reference contract drift for "+identity)}
		}
		if operation.Source.SHA256 != expectedOperation.source.SHA256 || operation.Source.Bytes != expectedOperation.source.Bytes || (expectedOperation.source.Location != "" && operation.Source.Location != expectedOperation.source.Location) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor provenance drift for "+identity)}
		}
		if (lock.SchemaVersion == 3 || lock.isLegacySourceReference()) && (operation.ProviderOperationID != expectedOperation.providerOperationID || operation.Source.URL != expectedOperation.source.URL || operation.Source.Form != expectedOperation.source.Form || operation.Source.Version != expectedOperation.source.Version || operation.Source.DocumentID != expectedOperation.source.DocumentID || operation.Source.PublishedURL != expectedOperation.source.PublishedURL || operation.Source.PublishedCaptureURL != expectedOperation.source.PublishedCaptureURL || operation.Source.PublishedSHA256 != expectedOperation.source.PublishedSHA256 || operation.Source.PublishedBytes != expectedOperation.source.PublishedBytes || operation.Source.PublishedAdapter != expectedOperation.source.PublishedAdapter || operation.Source.ContentType != expectedOperation.source.ContentType || operation.Source.CitationURL != expectedOperation.source.CitationURL) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor provenance drift for "+identity)}
		}
		if (expectedOperation.source.URL != "" && operation.Source.URL != expectedOperation.source.URL) || (expectedOperation.source.Location != "" && operation.Source.Location != expectedOperation.source.Location) || (expectedOperation.method != "" && (!strings.EqualFold(operation.Method, expectedOperation.method) || operation.Path != expectedOperation.path)) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor provider contract drift for "+identity)}
		}
		if lock.SchemaVersion == 3 && operation.ProviderOperationID != expectedOperation.providerOperationID {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor provider contract drift for "+identity)}
		}
		if operation.Runtime.MergeBlocked != (len(operation.Runtime.Gaps) > 0) {
			return []Finding{sourceProjectionFinding(connector, file, "source descriptor gap state is inconsistent for "+identity)}
		}
	}
	return nil
}

func sourceImportExpectedV3DescriptorProvenance(document sourceImportRESTDocument, operation sourceImportRESTOperation) sourceImportSource {
	if document.isSourceReference() {
		source := sourceImportReferenceProvenance(document.sourceArtifact(), operation.SourceLocation, sourceImportReferenceForm(document.sourceArtifact()))
		source.DocumentID = document.ID
		return source
	}
	form := document.sourceKind()
	version := ""
	if form == sourceImportDocumentKindOpenAPI {
		if document.Artifact.Swagger != "" {
			form = "swagger"
			version = document.Artifact.Swagger
		} else {
			version = document.Artifact.OpenAPI
		}
	}
	return sourceImportSource{
		URL:                 document.Artifact.SourceURL,
		SHA256:              strings.ToLower(document.Artifact.SHA256),
		Bytes:               document.Artifact.Bytes,
		Location:            operation.SourceLocation,
		Form:                form,
		Version:             version,
		DocumentID:          document.ID,
		PublishedURL:        document.PublishedSource.SourceURL,
		PublishedCaptureURL: document.PublishedSource.CaptureURL,
		PublishedSHA256:     strings.ToLower(document.PublishedSource.SHA256),
		PublishedBytes:      document.PublishedSource.Bytes,
		PublishedAdapter:    document.PublishedSource.Adapter,
		ContentType:         document.ContentType,
		CitationURL:         operation.CitationURL,
	}
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
				if err := sourceProjectionValidateWriteDisabledMutationArtifact(bundle, operation); err != nil {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, err.Error()+": "+operation.SourceID))
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
			if operation.Runtime.PartialCoverageMutation != nil {
				if !sourceProjectionHasPartialMutationCoverageDisposition(operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited partial mutation coverage disposition is invalid: "+operation.SourceID))
					continue
				}
				if !sourceProjectionPartialCoverageFoundationMatchesOperation(bundle, operation, *operation.Runtime.PartialCoverageMutation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited partial mutation coverage disposition has no matching missing foundation: "+operation.SourceID))
					continue
				}
				if sourceProjectionMutationActionIsComplete(bundle, operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited partial mutation coverage disposition claims a complete executable action: "+operation.SourceID))
					continue
				}
				if !sourceProjectionMutationClaimsImplementedAction(bundle, operation) {
					findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-cited partial mutation coverage disposition has no implemented declared action: "+operation.SourceID))
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
			if violation := sourceProjectionReadInputClosure(bundle, operation); violation != "" {
				findings = append(findings, sourceProjectionFinding(bundle.Name, file, "source-bound request input absent from locked source contract: "+operation.SourceID+": "+violation))
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
		if strings.EqualFold(method, source.Method) && sourceProjectionStreamPathMatchesSourcePath(stream.Path, source.Path) && sourceProjectionDeclaredStream(bundle, stream.Name, source.SourceID, allowSourceBoundPartial) {
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
	// A direct read is allowed to omit a provider's optional filters and use
	// the provider default page.  Requiring those optional inputs here made a
	// fully typed required input look unreachable, which in turn downgraded an
	// otherwise closed command on every generator pass.
	return sourceRequiredCallerFieldsCovered(source, covered)
}

// sourceProjectionReadInputClosure is deliberately stricter than the
// required-input coverage test above. Coverage answers whether the source's
// required fields have a declaration-owned route. Closure also answers the
// inverse question: whether a source-bound read has added an input that its
// retained source never admitted. This keeps a local operation-only parameter
// from becoming an unchecked provider request channel even when no CLI flag
// exposes it.
func sourceProjectionReadInputClosure(bundle engine.Bundle, source sourceOperationDescriptor) string {
	for _, operation := range bundle.Operations {
		if operation.SourceOperation == nil || operation.SourceOperation.ID != source.SourceID || operation.REST == nil {
			continue
		}
		if violation := sourceProjectionRESTReadInputClosure(operation.REST, source); violation != "" {
			return "operation " + operation.ID + " " + violation
		}
	}
	if bundle.CLISurface == nil {
		return ""
	}
	for _, command := range bundle.CLISurface.Commands {
		if command.SourceOperation != source.SourceID {
			continue
		}
		for _, flag := range command.Flags {
			if violation := sourceProjectionReadFlagInputClosure(flag, source); violation != "" {
				return "command " + command.Path + " " + violation
			}
		}
	}
	return ""
}

func sourceProjectionRESTReadInputClosure(rest *engine.RESTOperationSpec, source sourceOperationDescriptor) string {
	for _, parameter := range rest.Parameters {
		if sourceProjectionReadParameterAdmitted(source, parameter.In, parameter.Name) {
			continue
		}
		return "parameter " + parameter.In + "." + parameter.Name
	}
	for _, parameter := range rest.PaginationParameters {
		if sourceProjectionReadPaginationParameterAdmitted(source, parameter) {
			continue
		}
		return "pagination parameter " + parameter.In + "." + parameter.Name
	}
	for name := range rest.Query {
		if sourceProjectionReadParameterAdmitted(source, "query", name) {
			continue
		}
		return "static query " + name
	}
	if violation := sourceProjectionReadBodyInputClosure(rest, source); violation != "" {
		return violation
	}
	return ""
}

func sourceProjectionReadPaginationParameterAdmitted(source sourceOperationDescriptor, parameter engine.OperationParameter) bool {
	if parameter.In != "query" || source.Pagination == nil {
		return false
	}
	pagination, ok := source.Pagination.(map[string]any)
	if !ok {
		return false
	}
	for _, name := range sourceProjectionPaginationParameterNames(pagination) {
		if name != parameter.Name {
			continue
		}
		for _, sourceParameter := range source.Request.Query {
			if sourceParameter.Name == name {
				return true
			}
		}
	}
	return false
}

func sourceProjectionReadFlagInputClosure(flag engine.CLIFlag, source sourceOperationDescriptor) string {
	mapping := strings.TrimSpace(flag.MapsTo)
	for _, location := range []string{"path", "query", "header"} {
		prefix := location + "."
		if !strings.HasPrefix(mapping, prefix) {
			continue
		}
		if sourceProjectionReadParameterAdmitted(source, location, strings.TrimPrefix(mapping, prefix)) {
			return ""
		}
		return "flag --" + flag.Name + " mapping " + mapping
	}
	if strings.HasPrefix(mapping, "body") {
		field := strings.TrimPrefix(strings.TrimPrefix(mapping, "body"), ".")
		if sourceProjectionReadBodyFieldAdmitted(source, field) {
			return ""
		}
		return "flag --" + flag.Name + " mapping " + mapping
	}
	return ""
}

func sourceProjectionReadBodyInputClosure(rest *engine.RESTOperationSpec, source sourceOperationDescriptor) string {
	if source.Request.Body == nil {
		if len(rest.Body) != 0 || len(rest.BodySchema) != 0 || rest.ContentType != "" {
			return "body"
		}
		return ""
	}
	for name := range rest.Body {
		if !sourceProjectionReadBodyFieldAdmitted(source, name) {
			return "body." + name
		}
	}
	if len(rest.BodySchema) == 0 {
		return ""
	}
	var schema map[string]any
	decoder := json.NewDecoder(bytes.NewReader(rest.BodySchema))
	decoder.UseNumber()
	if decoder.Decode(&schema) != nil {
		return "body schema"
	}
	properties, _ := schema["properties"].(map[string]any)
	for name := range properties {
		if !sourceProjectionReadBodyFieldAdmitted(source, name) {
			return "body schema." + name
		}
	}
	return ""
}

func sourceProjectionReadBodyFieldAdmitted(source sourceOperationDescriptor, field string) bool {
	if source.Request.Body == nil || field == "" {
		return false
	}
	schema, ok := source.Request.Body.Schema.(map[string]any)
	if !ok {
		return false
	}
	properties, _ := schema["properties"].(map[string]any)
	_, admitted := properties[field]
	return admitted
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

// sourceProjectionReadOnlyResult retains the complete provider descriptor but
// selects every non-mutating GET for the closed materializer. The materializer
// alone decides whether an exact declaration-owned REST operation can become a
// bounded direct read, an existing stream proves ETL semantics, or a named
// foundation must remain visible. Pre-filtering against a historical status
// would let planned/deferred metadata hide an otherwise complete provider
// contract.
func sourceProjectionReadOnlyResult(bundleDir string, result sourceImportResult) (sourceImportResult, error) {
	_ = bundleDir // Callers retain the argument while this lane becomes capability-based.
	filtered := result
	filtered.Operations = make([]sourceOperationDescriptor, 0, len(result.Operations))
	for _, operation := range result.Operations {
		if sourceProjectionOperationMutates(operation) || operation.Protocol == "graphql" || !strings.EqualFold(operation.Method, "GET") {
			continue
		}
		filtered.Operations = append(filtered.Operations, operation)
	}
	return filtered, nil
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
		{path: "metadata.json", decode: func(raw []byte) error {
			var value engine.Metadata
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			if !value.Capabilities.WriteDeclared {
				return errors.New("metadata capabilities.write must be explicitly declared")
			}
			bundle.Metadata = value
			return nil
		}},
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
