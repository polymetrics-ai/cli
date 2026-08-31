package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// retainedSourceMappingCohortManifest is deliberately fixed to the reviewed
// Batch R1 denominator. This command is a read-only authoring check: it does
// not write a descriptor, an enabled contract, or a runtime declaration.
const retainedSourceMappingCohortManifest = "data/connector-canon/batch1-source-operation-mapping-cohort.json"

var retainedSourceMappingLaneOrder = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

// retainedSourceMappingPartitionOrder is only a deterministic accounting
// partition for EnabledConnectorContract reconciliation. It does not select a
// runtime lane, infer provider semantics, or make a command executable.
var retainedSourceMappingPartitionOrder = []string{
	"direct_read",
	"reverse_etl",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"sync_transport",
}

type retainedSourceMappingOptions struct {
	Connector string
	Check     bool
}

type retainedSourceMappingResult struct {
	Contract connectors.EnabledConnectorContract
	Report   retainedSourceMappingReport
}

type retainedSourceMappingSourceProof struct {
	OperationIDs []string
}

// retainedSourceMappingReport intentionally contains only source-accounting
// facts. It is terminal output, not a definition artifact and not an
// executable-capability declaration.
type retainedSourceMappingReport struct {
	Connector                string                            `json:"connector"`
	MappingOnly              bool                              `json:"mapping_only"`
	SourceLock               string                            `json:"source_lock"`
	SourceOperations         int                               `json:"source_operations"`
	VerifiedSourceOperations int                               `json:"verified_source_operations"`
	ExecutableDeclarations   int                               `json:"executable_declarations"`
	AccountingPartition      string                            `json:"accounting_partition"`
	Lanes                    []retainedSourceMappingLaneReport `json:"lanes"`
}

type retainedSourceMappingLaneReport struct {
	Name               string `json:"name"`
	Expected           int    `json:"expected"`
	MappedUnproven     int    `json:"mapped_unproven"`
	DeferredFoundation int    `json:"deferred_foundation"`
	Unsupported        int    `json:"unsupported_with_provider_evidence"`
	Partition          bool   `json:"partition"`
}

type retainedSourceMappingMatrix struct {
	Connector string
	Rows      map[string]map[string]string
}

func runRetainedSourceMapping(args []string, stdout, stderr io.Writer) int {
	if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
		logln(stdout, retainedSourceMappingUsage())
		return 0
	}
	opts, err := parseRetainedSourceMappingOptions(args)
	if err != nil {
		logf(stderr, "connectorgen retained-source-mapping: %v\n", err)
		return 2
	}
	root, err := repoRoot()
	if err != nil {
		logf(stderr, "connectorgen retained-source-mapping: resolve repository root: %v\n", err)
		return 1
	}
	result, err := retainedSourceMappingFromRepository(root, opts.Connector)
	if err != nil {
		logf(stderr, "connectorgen retained-source-mapping: %v\n", err)
		return 1
	}
	logf(stdout, "connectorgen retained-source-mapping: %s: mapping-only; %d source operation(s), %d verified source operation(s), 0 executable declaration(s), %d lane(s), 0 finding(s)\n", result.Report.Connector, result.Report.SourceOperations, result.Report.VerifiedSourceOperations, len(result.Report.Lanes))
	return 0
}

func retainedSourceMappingUsage() string {
	return "usage: connectorgen retained-source-mapping <connector> [--check]"
}

func parseRetainedSourceMappingOptions(args []string) (retainedSourceMappingOptions, error) {
	var options retainedSourceMappingOptions
	for _, argument := range args[1:] {
		switch argument {
		case "--check":
			if options.Check {
				return retainedSourceMappingOptions{}, fmt.Errorf("--check may only be specified once")
			}
			options.Check = true
		default:
			if strings.HasPrefix(argument, "-") {
				return retainedSourceMappingOptions{}, fmt.Errorf("unknown flag %q", argument)
			}
			if options.Connector != "" {
				return retainedSourceMappingOptions{}, fmt.Errorf("only one connector may be checked at a time")
			}
			options.Connector = argument
		}
	}
	if options.Connector == "" {
		return retainedSourceMappingOptions{}, fmt.Errorf("a connector name is required")
	}
	if err := validateSourceImportConnector(options.Connector); err != nil {
		return retainedSourceMappingOptions{}, err
	}
	return options, nil
}

func retainedSourceMappingFromRepository(root, connector string) (retainedSourceMappingResult, error) {
	if err := validateSourceImportConnector(connector); err != nil {
		return retainedSourceMappingResult{}, err
	}
	manifestPath := filepath.Join(root, filepath.FromSlash(retainedSourceMappingCohortManifest))
	cohort, err := sourceOperationMappingCohortPathCheck(root, manifestPath)
	if err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("verify frozen Batch R1 cohort: %w", err)
	}
	if len(cohort.Findings) != 0 {
		return retainedSourceMappingResult{}, fmt.Errorf("frozen Batch R1 cohort has %d finding(s)", len(cohort.Findings))
	}

	entry, err := retainedSourceMappingCohortEntry(manifestPath, connector)
	if err != nil {
		return retainedSourceMappingResult{}, err
	}

	lockPath, err := sourceOperationMappingCohortOwnedPath(root, connector, entry.Path, "-operation-source-lock.json")
	if err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("resolve source lock: %w", err)
	}
	matrixPath, err := retainedSourceMappingOwnedMatrixPath(root, connector, entry.MatrixPath)
	if err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("resolve source-lane matrix: %w", err)
	}
	lockRaw, err := os.ReadFile(lockPath)
	if err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("read source lock: %w", err)
	}
	lock, err := parseSourceImportLock(lockRaw, connector)
	if err != nil {
		return retainedSourceMappingResult{}, err
	}
	if err := retainedSourceMappingEligible(lock); err != nil {
		return retainedSourceMappingResult{}, err
	}
	proof, err := retainedSourceMappingVerifySourceEvidence(lock)
	if err != nil {
		return retainedSourceMappingResult{}, err
	}
	matrixRaw, err := os.ReadFile(matrixPath)
	if err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("read source-lane matrix: %w", err)
	}
	matrix, err := decodeRetainedSourceMappingMatrix(matrixRaw, connector)
	if err != nil {
		return retainedSourceMappingResult{}, err
	}
	artifact := filepath.ToSlash(filepath.Join("sources", connector+"-operation-source-lock.json"))
	return buildRetainedSourceMapping(lock, artifact, matrix, proof)
}

func retainedSourceMappingCohortEntry(manifestPath, connector string) (sourceOperationMappingCohortSourceLock, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return sourceOperationMappingCohortSourceLock{}, fmt.Errorf("read cohort manifest: %w", err)
	}
	var manifest sourceOperationMappingCohortManifest
	if err := decodeSourceStrictJSON(raw, &manifest); err != nil {
		return sourceOperationMappingCohortSourceLock{}, fmt.Errorf("decode cohort manifest: %w", err)
	}
	for _, entry := range manifest.SourceLocks {
		if entry.Connector == connector {
			return entry, nil
		}
	}
	return sourceOperationMappingCohortSourceLock{}, fmt.Errorf("connector %q is not in the frozen Batch R1 cohort", connector)
}

func retainedSourceMappingOwnedMatrixPath(root, connector, raw string) (string, error) {
	if err := sourceOperationMappingCohortMatrixPath(connector, raw); err != nil {
		return "", err
	}
	return sourceOperationMappingCohortOwnedPath(root, connector, raw, "-source-lane-matrix.json")
}

// retainedSourceMappingEligible deliberately does not reuse or relax
// sourceImportLockHasCanonicalEvidenceContract. Normal canonical-evidence
// source import stays closed to its explicit legacy admission marker.
func retainedSourceMappingEligible(lock sourceImportLock) error {
	if lock.SchemaVersion != 2 {
		return fmt.Errorf("source lock schema version %d is not an eligible retained v2 mapping lock", lock.SchemaVersion)
	}
	if lock.isLegacySourceReference() {
		return fmt.Errorf("source lock is reference-only and cannot supply retained provider contract mapping")
	}
	if lock.Rest.CanonicalEvidence {
		return fmt.Errorf("source lock declares canonical_evidence and belongs to normal source-import admission")
	}
	if lock.Counts.GraphQLQuery != 0 || lock.Counts.GraphQLMutation != 0 || len(lock.GraphQL.QueryFields) != 0 || len(lock.GraphQL.MutationFields) != 0 {
		return fmt.Errorf("retained-source mapping accepts zero GraphQL operations only")
	}
	if lock.SourceContract == nil || len(bytes.TrimSpace(lock.SourceContract.Raw)) == 0 {
		return fmt.Errorf("source lock has no retained source_contract object")
	}
	if len(lock.Rest.Operations) == 0 || lock.Counts.REST != len(lock.Rest.Operations) {
		return fmt.Errorf("source lock REST inventory is incomplete")
	}
	for _, operation := range lock.Rest.Operations {
		if operation.SourceOperation == nil || len(bytes.TrimSpace(operation.SourceOperation.Raw)) == 0 {
			return fmt.Errorf("source lock operation %q has no retained source_operation object", operation.ID)
		}
	}
	return nil
}

// retainedSourceMappingVerifySourceEvidence validates retained provider
// grammar entirely in memory. It calls neither runSourceImport nor a fetcher,
// and it deliberately avoids descriptor/schema materialization: mapping
// evidence must not be rejected merely because a later runtime projection
// cannot resolve an unrelated provider schema. The proof is limited to exact
// source-operation identity, path, method, and provider-operation-ID facts.
func retainedSourceMappingVerifySourceEvidence(lock sourceImportLock) (retainedSourceMappingSourceProof, error) {
	if err := retainedSourceMappingEligible(lock); err != nil {
		return retainedSourceMappingSourceProof{}, err
	}
	if err := validateSourceImportArtifact(lock.Rest.sourceImportArtifact); err != nil {
		return retainedSourceMappingSourceProof{}, err
	}
	if err := validateSourceImportPathBridge(lock.Rest.PathBridge); err != nil {
		return retainedSourceMappingSourceProof{}, err
	}
	limits := defaultSourceImportLimits()
	if int64(len(lock.SourceContract.Raw)) > limits.MaxArtifactBytes {
		return retainedSourceMappingSourceProof{}, fmt.Errorf("retained source contract byte limit exceeded")
	}
	document, form, err := sourceImportCanonicalEvidenceDocument(lock)
	if err != nil {
		return retainedSourceMappingSourceProof{}, fmt.Errorf("assemble retained source contract: %w", err)
	}
	if err := validateSourceImportArtifactForm(lock.Rest.sourceImportArtifact, form); err != nil {
		return retainedSourceMappingSourceProof{}, fmt.Errorf("retained source contract form: %w", err)
	}
	paths, err := retainedSourceMappingObject(document["paths"], "retained source contract paths")
	if err != nil {
		return retainedSourceMappingSourceProof{}, err
	}
	proof := retainedSourceMappingSourceProof{OperationIDs: make([]string, 0, len(lock.Rest.Operations))}
	seen := make(map[string]bool, len(lock.Rest.Operations))
	for _, locked := range lock.Rest.Operations {
		if seen[locked.ID] {
			return retainedSourceMappingSourceProof{}, fmt.Errorf("source lock duplicates retained source ID %q", locked.ID)
		}
		seen[locked.ID] = true
		pathItem, err := retainedSourceMappingObject(paths[locked.Path], fmt.Sprintf("retained source contract path %q", locked.Path))
		if err != nil {
			return retainedSourceMappingSourceProof{}, err
		}
		operation, err := retainedSourceMappingObject(pathItem[strings.ToLower(locked.Method)], fmt.Sprintf("retained source contract %s %s", strings.ToUpper(locked.Method), locked.Path))
		if err != nil {
			return retainedSourceMappingSourceProof{}, err
		}
		if rawOperationID, exists := operation["operationId"]; exists {
			providerOperationID, ok := rawOperationID.(string)
			if !ok || providerOperationID != locked.OperationID {
				return retainedSourceMappingSourceProof{}, fmt.Errorf("retained source operation %q operationId conflicts with source lock", locked.ID)
			}
		} else if locked.OperationID != "" {
			return retainedSourceMappingSourceProof{}, fmt.Errorf("retained source operation %q omits locked operationId", locked.ID)
		}
		proof.OperationIDs = append(proof.OperationIDs, locked.ID)
	}
	sort.Strings(proof.OperationIDs)
	return proof, nil
}

func decodeRetainedSourceMappingMatrix(raw []byte, connector string) (retainedSourceMappingMatrix, error) {
	var root any
	if err := decodeSourceJSON(raw, &root); err != nil {
		return retainedSourceMappingMatrix{}, fmt.Errorf("decode source-lane matrix: %w", err)
	}
	object, err := retainedSourceMappingObject(root, "source-lane matrix")
	if err != nil {
		return retainedSourceMappingMatrix{}, err
	}
	if version, err := retainedSourceMappingInteger(object["schema_version"], "schema_version"); err != nil || version != 1 {
		if err != nil {
			return retainedSourceMappingMatrix{}, err
		}
		return retainedSourceMappingMatrix{}, fmt.Errorf("unsupported source-lane matrix schema version %d", version)
	}
	actualConnector, err := retainedSourceMappingString(object["connector"], "connector")
	if err != nil {
		return retainedSourceMappingMatrix{}, err
	}
	if actualConnector != connector {
		return retainedSourceMappingMatrix{}, fmt.Errorf("source-lane matrix connector %q does not match requested connector %q", actualConnector, connector)
	}
	lanes, err := retainedSourceMappingStringArray(object["lanes"], "lanes")
	if err != nil {
		return retainedSourceMappingMatrix{}, err
	}
	if !retainedSourceMappingExactLaneSet(lanes) {
		return retainedSourceMappingMatrix{}, fmt.Errorf("source-lane matrix must declare exactly the seven fixed lanes")
	}

	_, hasSourceOperations := object["source_operations"]
	_, hasOperations := object["operations"]
	if hasSourceOperations == hasOperations {
		return retainedSourceMappingMatrix{}, fmt.Errorf("source-lane matrix must declare exactly one of source_operations or operations")
	}
	rowsValue := object["source_operations"]
	form := "source_operations"
	if hasOperations {
		rowsValue = object["operations"]
		form = "operations"
	}
	rows, err := retainedSourceMappingArray(rowsValue, form)
	if err != nil {
		return retainedSourceMappingMatrix{}, err
	}
	if len(rows) == 0 {
		return retainedSourceMappingMatrix{}, fmt.Errorf("source-lane matrix has no operation rows")
	}

	decoded := retainedSourceMappingMatrix{Connector: connector, Rows: make(map[string]map[string]string, len(rows))}
	for index, rawRow := range rows {
		row, err := retainedSourceMappingObject(rawRow, fmt.Sprintf("%s[%d]", form, index))
		if err != nil {
			return retainedSourceMappingMatrix{}, err
		}
		if form == "source_operations" {
			if _, exists := row["source_operation_id"]; exists {
				return retainedSourceMappingMatrix{}, fmt.Errorf("source_operations[%d] must use source_id, not source_operation_id", index)
			}
			sourceID, err := retainedSourceMappingString(row["source_id"], fmt.Sprintf("source_operations[%d].source_id", index))
			if err != nil {
				return retainedSourceMappingMatrix{}, err
			}
			if _, hasCells := row["cells"]; hasCells {
				return retainedSourceMappingMatrix{}, fmt.Errorf("source_operations[%d] cannot mix cells with lanes", index)
			}
			states, err := retainedSourceMappingLanes(row["lanes"], fmt.Sprintf("source_operations[%d].lanes", index))
			if err != nil {
				return retainedSourceMappingMatrix{}, err
			}
			if _, duplicate := decoded.Rows[sourceID]; duplicate {
				return retainedSourceMappingMatrix{}, fmt.Errorf("source-lane matrix duplicates source ID %q", sourceID)
			}
			decoded.Rows[sourceID] = states
			continue
		}

		_, hasSourceID := row["source_id"]
		_, hasSourceOperationID := row["source_operation_id"]
		if hasSourceID == hasSourceOperationID {
			return retainedSourceMappingMatrix{}, fmt.Errorf("operations[%d] must declare exactly one of source_id or source_operation_id", index)
		}
		field := "source_id"
		if hasSourceOperationID {
			field = "source_operation_id"
		}
		sourceID, err := retainedSourceMappingString(row[field], fmt.Sprintf("operations[%d].%s", index, field))
		if err != nil {
			return retainedSourceMappingMatrix{}, err
		}
		if _, hasLanes := row["lanes"]; hasLanes {
			return retainedSourceMappingMatrix{}, fmt.Errorf("operations[%d] cannot mix lanes with cells", index)
		}
		states, err := retainedSourceMappingCells(row["cells"], fmt.Sprintf("operations[%d].cells", index))
		if err != nil {
			return retainedSourceMappingMatrix{}, err
		}
		if _, duplicate := decoded.Rows[sourceID]; duplicate {
			return retainedSourceMappingMatrix{}, fmt.Errorf("source-lane matrix duplicates source ID %q", sourceID)
		}
		decoded.Rows[sourceID] = states
	}
	return decoded, nil
}

func retainedSourceMappingLanes(value any, field string) (map[string]string, error) {
	lanes, err := retainedSourceMappingObject(value, field)
	if err != nil {
		return nil, err
	}
	if len(lanes) != len(retainedSourceMappingLaneOrder) {
		return nil, fmt.Errorf("%s must declare exactly seven lane cells", field)
	}
	states := make(map[string]string, len(lanes))
	for _, lane := range retainedSourceMappingLaneOrder {
		rawCell, exists := lanes[lane]
		if !exists {
			return nil, fmt.Errorf("%s omits lane %q", field, lane)
		}
		cell, err := retainedSourceMappingObject(rawCell, field+"."+lane)
		if err != nil {
			return nil, err
		}
		applicability, err := retainedSourceMappingString(cell["applicability"], field+"."+lane+".applicability")
		if err != nil {
			return nil, err
		}
		disposition, err := retainedSourceMappingString(cell["disposition"], field+"."+lane+".disposition")
		if err != nil {
			return nil, err
		}
		state, err := retainedSourceMappingDisposition(applicability, disposition, field+"."+lane)
		if err != nil {
			return nil, err
		}
		if state != "" {
			states[lane] = state
		}
	}
	for lane := range lanes {
		if !retainedSourceMappingKnownLane(lane) {
			return nil, fmt.Errorf("%s has unknown lane %q", field, lane)
		}
	}
	return states, nil
}

func retainedSourceMappingCells(value any, field string) (map[string]string, error) {
	cells, err := retainedSourceMappingArray(value, field)
	if err != nil {
		return nil, err
	}
	if len(cells) != len(retainedSourceMappingLaneOrder) {
		return nil, fmt.Errorf("%s must declare exactly seven lane cells", field)
	}
	states := make(map[string]string, len(cells))
	seen := make(map[string]bool, len(cells))
	for index, rawCell := range cells {
		cell, err := retainedSourceMappingObject(rawCell, fmt.Sprintf("%s[%d]", field, index))
		if err != nil {
			return nil, err
		}
		lane, err := retainedSourceMappingString(cell["lane"], fmt.Sprintf("%s[%d].lane", field, index))
		if err != nil {
			return nil, err
		}
		if !retainedSourceMappingKnownLane(lane) || seen[lane] {
			return nil, fmt.Errorf("%s has an unknown or duplicate lane %q", field, lane)
		}
		seen[lane] = true
		state, err := retainedSourceMappingString(cell["state"], fmt.Sprintf("%s[%d].state", field, index))
		if err != nil {
			return nil, err
		}
		state, err = retainedSourceMappingState(state, fmt.Sprintf("%s[%d].state", field, index))
		if err != nil {
			return nil, err
		}
		if state != "" {
			states[lane] = state
		}
	}
	return states, nil
}

func retainedSourceMappingDisposition(applicability, disposition, field string) (string, error) {
	switch applicability {
	case "not_applicable":
		if disposition != "not_applicable" {
			return "", fmt.Errorf("%s has not_applicable applicability with disposition %q", field, disposition)
		}
		return "", nil
	case "applicable", "source_candidate":
		return retainedSourceMappingState(disposition, field+".disposition")
	default:
		return "", fmt.Errorf("%s has unknown applicability %q", field, applicability)
	}
}

func retainedSourceMappingState(value, field string) (string, error) {
	switch value {
	case "not_applicable":
		return "", nil
	case "mapped_unproven":
		return connectors.EnabledLaneMappedUnproven, nil
	case "missing_foundation", connectors.EnabledLaneDeferred:
		return connectors.EnabledLaneDeferred, nil
	case connectors.EnabledLaneImplemented, "unmapped_mapping":
		return "", fmt.Errorf("%s cannot claim %q in a retained-source mapping", field, value)
	default:
		return "", fmt.Errorf("%s has unknown disposition/state %q", field, value)
	}
}

func retainedSourceMappingObject(value any, field string) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return nil, fmt.Errorf("%s must be an object", field)
	}
	return object, nil
}

func retainedSourceMappingArray(value any, field string) ([]any, error) {
	array, ok := value.([]any)
	if !ok || array == nil {
		return nil, fmt.Errorf("%s must be an array", field)
	}
	return array, nil
}

func retainedSourceMappingString(value any, field string) (string, error) {
	stringValue, ok := value.(string)
	if !ok || strings.TrimSpace(stringValue) == "" {
		return "", fmt.Errorf("%s must be a non-empty string", field)
	}
	return stringValue, nil
}

func retainedSourceMappingStringArray(value any, field string) ([]string, error) {
	array, err := retainedSourceMappingArray(value, field)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(array))
	for index, raw := range array {
		value, err := retainedSourceMappingString(raw, fmt.Sprintf("%s[%d]", field, index))
		if err != nil {
			return nil, err
		}
		values[index] = value
	}
	return values, nil
}

func retainedSourceMappingInteger(value any, field string) (int, error) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s must be an integer", field)
	}
	parsed, err := strconv.Atoi(number.String())
	if err != nil || strconv.Itoa(parsed) != number.String() {
		return 0, fmt.Errorf("%s must be an integer", field)
	}
	return parsed, nil
}

func retainedSourceMappingExactLaneSet(values []string) bool {
	if len(values) != len(retainedSourceMappingLaneOrder) {
		return false
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if !retainedSourceMappingKnownLane(value) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func retainedSourceMappingKnownLane(value string) bool {
	for _, lane := range retainedSourceMappingLaneOrder {
		if value == lane {
			return true
		}
	}
	return false
}

func buildRetainedSourceMapping(lock sourceImportLock, artifact string, matrix retainedSourceMappingMatrix, proof retainedSourceMappingSourceProof) (retainedSourceMappingResult, error) {
	if err := retainedSourceMappingEligible(lock); err != nil {
		return retainedSourceMappingResult{}, err
	}
	if matrix.Connector != lock.Connector {
		return retainedSourceMappingResult{}, fmt.Errorf("source-lane matrix connector %q does not match source lock connector %q", matrix.Connector, lock.Connector)
	}
	if filepath.ToSlash(artifact) != filepath.ToSlash(filepath.Join("sources", lock.Connector+"-operation-source-lock.json")) {
		return retainedSourceMappingResult{}, fmt.Errorf("retained source mapping artifact must be the connector-owned source lock")
	}
	operations := enabledContractSourceOperations(lock)
	if len(operations) != lock.Counts.REST {
		return retainedSourceMappingResult{}, fmt.Errorf("source lock retained %d REST operations, want %d", len(operations), lock.Counts.REST)
	}
	if err := retainedSourceMappingProofIdentities(lock, proof); err != nil {
		return retainedSourceMappingResult{}, err
	}
	if len(matrix.Rows) != len(operations) {
		return retainedSourceMappingResult{}, fmt.Errorf("source-lane matrix has %d rows, want %d source operations", len(matrix.Rows), len(operations))
	}

	lanes, err := retainedSourceMappingEmptyLanes(lock.Connector, artifact, lock.Rest.SourceURL)
	if err != nil {
		return retainedSourceMappingResult{}, err
	}
	known := make(map[string]bool, len(operations))
	for _, operation := range operations {
		known[operation.ID] = true
		states, exists := matrix.Rows[operation.ID]
		if !exists {
			return retainedSourceMappingResult{}, fmt.Errorf("source-lane matrix omits retained source ID %q", operation.ID)
		}
		primaryLane, primaryState := retainedSourceMappingPrimaryCell(states)
		if err := retainedSourceMappingAddCell(lanes[primaryLane], operation.ID, primaryState, true); err != nil {
			return retainedSourceMappingResult{}, err
		}
		for _, lane := range retainedSourceMappingLaneOrder {
			state, exists := states[lane]
			if !exists || lane == primaryLane {
				continue
			}
			if err := retainedSourceMappingAddCell(lanes[lane], operation.ID, state, false); err != nil {
				return retainedSourceMappingResult{}, err
			}
		}
	}
	for sourceID := range matrix.Rows {
		if !known[sourceID] {
			return retainedSourceMappingResult{}, fmt.Errorf("source-lane matrix references unknown source ID %q", sourceID)
		}
	}

	contract := connectors.EnabledConnectorContract{
		SchemaVersion: 1,
		Connector:     lock.Connector,
		RetentionOnly: true,
		SourceLock: connectors.EnabledContractSourceLock{
			Path:              artifact,
			SHA256:            lock.Rest.SHA256,
			Bytes:             lock.Rest.Bytes,
			CanonicalEvidence: false,
		},
		Lanes: make([]connectors.EnabledConnectorLane, 0, len(retainedSourceMappingLaneOrder)),
	}
	for _, name := range retainedSourceMappingLaneOrder {
		retainedSourceMappingFinalizeLane(lanes[name])
		contract.Lanes = append(contract.Lanes, *lanes[name])
	}
	if err := contract.ValidateRetentionOnly(); err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("validate source-only retained contract: %w", err)
	}
	if err := contract.ReconcileSourceOperations(operations); err != nil {
		return retainedSourceMappingResult{}, fmt.Errorf("reconcile exact source-operation IDs: %w", err)
	}

	report := retainedSourceMappingReport{
		Connector:                lock.Connector,
		MappingOnly:              true,
		SourceLock:               artifact,
		SourceOperations:         len(operations),
		VerifiedSourceOperations: len(proof.OperationIDs),
		ExecutableDeclarations:   0,
		AccountingPartition:      "fixed source-ID accounting only; not a runtime lane selection",
		Lanes:                    make([]retainedSourceMappingLaneReport, 0, len(contract.Lanes)),
	}
	for _, lane := range contract.Lanes {
		report.Lanes = append(report.Lanes, retainedSourceMappingLaneReport{
			Name:               lane.Name,
			Expected:           lane.Source.Expected,
			MappedUnproven:     lane.Source.MappedUnproven,
			DeferredFoundation: lane.Source.DeferredFoundation,
			Unsupported:        lane.Source.Unsupported,
			Partition:          lane.Source.Partition,
		})
	}
	return retainedSourceMappingResult{Contract: contract, Report: report}, nil
}

func retainedSourceMappingProofIdentities(lock sourceImportLock, proof retainedSourceMappingSourceProof) error {
	if len(proof.OperationIDs) != len(lock.Rest.Operations) {
		return fmt.Errorf("retained source evidence has %d REST operations, want %d", len(proof.OperationIDs), len(lock.Rest.Operations))
	}
	locked := make(map[string]bool, len(lock.Rest.Operations))
	for _, operation := range lock.Rest.Operations {
		locked[operation.ID] = true
	}
	seen := make(map[string]bool, len(proof.OperationIDs))
	for _, sourceID := range proof.OperationIDs {
		if seen[sourceID] {
			return fmt.Errorf("retained source evidence duplicates source ID %q", sourceID)
		}
		seen[sourceID] = true
		if !locked[sourceID] {
			return fmt.Errorf("retained source evidence emitted unknown source ID %q", sourceID)
		}
	}
	for sourceID := range locked {
		if !seen[sourceID] {
			return fmt.Errorf("retained source evidence omits source ID %q", sourceID)
		}
	}
	return nil
}

func retainedSourceMappingEmptyLanes(connector, artifact, sourceURL string) (map[string]*connectors.EnabledConnectorLane, error) {
	if strings.TrimSpace(sourceURL) == "" {
		return nil, fmt.Errorf("source lock has no source URL for source-only citations")
	}
	lanes := make(map[string]*connectors.EnabledConnectorLane, len(retainedSourceMappingLaneOrder))
	for _, name := range retainedSourceMappingLaneOrder {
		lanes[name] = &connectors.EnabledConnectorLane{
			Name:      name,
			State:     connectors.EnabledLaneUnsupported,
			Reason:    "The retained source-lane matrix records this provider lane as not applicable.",
			Citations: []connectors.EnabledContractCitation{{URL: sourceURL, Location: "source-lane-matrix"}},
			Artifacts: []string{artifact},
			Source: connectors.EnabledContractSourceCoverage{
				Coverage: connectors.EnabledCoverageNotApplicable,
			},
		}
	}
	return lanes, nil
}

func retainedSourceMappingPrimaryCell(states map[string]string) (string, string) {
	for _, lane := range retainedSourceMappingPartitionOrder {
		if state, exists := states[lane]; exists {
			return lane, state
		}
	}
	return "direct_read", connectors.EnabledLaneUnsupported
}

func retainedSourceMappingAddCell(lane *connectors.EnabledConnectorLane, sourceID, state string, partition bool) error {
	if lane == nil {
		return fmt.Errorf("source ID %q maps to an unknown lane", sourceID)
	}
	if partition {
		lane.Source.Partition = true
	}
	lane.Source.OperationIDs = append(lane.Source.OperationIDs, sourceID)
	lane.Source.Expected++
	switch state {
	case connectors.EnabledLaneMappedUnproven:
		lane.Source.MappedUnproven++
	case connectors.EnabledLaneDeferred:
		lane.Source.DeferredFoundation++
	case connectors.EnabledLaneUnsupported:
		lane.Source.Unsupported++
	default:
		return fmt.Errorf("source ID %q lane %q has non-retainable state %q", sourceID, lane.Name, state)
	}
	return nil
}

func retainedSourceMappingFinalizeLane(lane *connectors.EnabledConnectorLane) {
	sort.Strings(lane.Source.OperationIDs)
	if lane.Source.Expected == 0 {
		return
	}
	lane.Source.Coverage = connectors.EnabledCoveragePartial
	switch {
	case lane.Source.MappedUnproven > 0:
		lane.State = connectors.EnabledLaneMappedUnproven
		lane.Reason = "The retained source-lane matrix retains source semantics without an executable declaration."
	case lane.Source.DeferredFoundation > 0:
		lane.State = connectors.EnabledLaneDeferred
		lane.Reason = "The retained source-lane matrix retains a named foundation gap without an executable declaration."
	default:
		lane.State = connectors.EnabledLaneUnsupported
		lane.Reason = "The retained source-lane matrix retains provider-evidenced non-applicability without an executable declaration."
	}
}

// retainedSourceMappingContractJSON provides deterministic bytes for a future
// connector-owned sidecar without writing one. Keeping serialization pure lets
// tests prove reproducibility while this bridge remains mapping-only.
func retainedSourceMappingContractJSON(contract connectors.EnabledConnectorContract) ([]byte, error) {
	raw, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
