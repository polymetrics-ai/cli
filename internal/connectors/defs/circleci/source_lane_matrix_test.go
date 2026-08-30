package circleci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	circleCISourceLockPath   = "sources/circleci-operation-source-lock.json"
	circleCISourceMatrixPath = "sources/circleci-source-lane-matrix.json"
)

var circleCILanes = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

var circleCIPagingOperationIDs = map[string]string{
	"circleci.rest.getProjectWorkflowMetrics":    "query_parameter",
	"circleci.rest.getProjectWorkflowRuns":       "query_parameter",
	"circleci.rest.getProjectWorkflowJobMetrics": "query_parameter",
	"circleci.rest.listPipelines":                "query_parameter",
	"circleci.rest.listWorkflowsByPipelineId":    "query_parameter",
	"circleci.rest.listPipelinesForProject":      "query_parameter",
	"circleci.rest.listMyPipelines":              "query_parameter",
	"circleci.rest.listSchedulesForProject":      "query_parameter",
	"circleci.rest.listWorkflowJobs":             "openapi_link",
}

var circleCIMutationOperationIDs = map[string]struct{}{
	"circleci.rest.cancelJobByJobID":                     {},
	"circleci.rest.createOrganization":                   {},
	"circleci.rest.deleteOrganization":                   {},
	"circleci.rest.createProject":                        {},
	"circleci.rest.createURLOrbAllowListEntry":           {},
	"circleci.rest.removeURLOrbAllowListEntry":           {},
	"circleci.rest.continuePipeline":                     {},
	"circleci.rest.deleteProjectBySlug":                  {},
	"circleci.rest.createCheckoutKey":                    {},
	"circleci.rest.deleteCheckoutKey":                    {},
	"circleci.rest.createEnvVar":                         {},
	"circleci.rest.deleteEnvVar":                         {},
	"circleci.rest.cancelJobByJobNumber":                 {},
	"circleci.rest.triggerPipeline":                      {},
	"circleci.rest.createSchedule":                       {},
	"circleci.rest.deleteScheduleById":                   {},
	"circleci.rest.updateSchedule":                       {},
	"circleci.rest.approvePendingApprovalJobById":        {},
	"circleci.rest.cancelWorkflow":                       {},
	"circleci.rest.rerunWorkflow":                        {},
	"circleci.rest.DeleteOrgClaims":                      {},
	"circleci.rest.PatchOrgClaims":                       {},
	"circleci.rest.DeleteProjectClaims":                  {},
	"circleci.rest.PatchProjectClaims":                   {},
	"circleci.rest.MakeDecision":                         {},
	"circleci.rest.SetDecisionSettings":                  {},
	"circleci.rest.CreatePolicyBundle":                   {},
	"circleci.rest.createContext":                        {},
	"circleci.rest.deleteContext":                        {},
	"circleci.rest.addEnvironmentVariableToContext":      {},
	"circleci.rest.deleteEnvironmentVariableFromContext": {},
	"circleci.rest.createContextRestriction":             {},
	"circleci.rest.deleteContextRestriction":             {},
	"circleci.rest.patchProjectSettings":                 {},
	"circleci.rest.createOrganizationGroup":              {},
	"circleci.rest.deleteGroup":                          {},
	"circleci.rest.createUsageExport":                    {},
	"circleci.rest.triggerPipelineRun":                   {},
	"circleci.rest.createPipelineDefinition":             {},
	"circleci.rest.updatePipelineDefinition":             {},
	"circleci.rest.deletePipelineDefinition":             {},
	"circleci.rest.createTrigger":                        {},
	"circleci.rest.updateTrigger":                        {},
	"circleci.rest.deleteTrigger":                        {},
	"circleci.rest.rollbackProject":                      {},
	"circleci.rest.createWebhook":                        {},
	"circleci.rest.updateWebhook":                        {},
	"circleci.rest.deleteWebhook":                        {},
	"circleci.rest.createOtelExporter":                   {},
	"circleci.rest.deleteOtelExporter":                   {},
}

type circleCISourceLock struct {
	Connector string `json:"connector"`
	Counts    struct {
		REST  int `json:"rest"`
		Total int `json:"total"`
	} `json:"counts"`
	REST struct {
		SourceURL  string                    `json:"source_url"`
		SHA256     string                    `json:"sha256"`
		Bytes      int                       `json:"bytes"`
		Operations []circleCISourceOperation `json:"operations"`
	} `json:"rest"`
	SourceContract json.RawMessage `json:"source_contract"`
}

type circleCISourceOperation struct {
	ID              string          `json:"id"`
	Protocol        string          `json:"protocol"`
	Method          string          `json:"method"`
	Path            string          `json:"path"`
	OperationID     string          `json:"operation_id"`
	SourceLocation  string          `json:"source_location"`
	SourceOperation json.RawMessage `json:"source_operation"`
}

type circleCISourceLaneMatrix struct {
	SchemaVersion int                       `json:"schema_version"`
	Connector     string                    `json:"connector"`
	SourceLock    circleCIMatrixSourceLock  `json:"source_lock"`
	Lanes         []string                  `json:"lanes"`
	Operations    []circleCIMatrixOperation `json:"operations"`
	Artifacts     []circleCIMatrixArtifact  `json:"artifacts"`
}

type circleCIMatrixSourceLock struct {
	Path                 string `json:"path"`
	SourceURL            string `json:"source_url"`
	SHA256               string `json:"sha256"`
	Bytes                int    `json:"bytes"`
	SourceOperationCount int    `json:"source_operation_count"`
}

type circleCIMatrixOperation struct {
	SourceOperationID   string                 `json:"source_operation_id"`
	ProviderOperationID string                 `json:"provider_operation_id"`
	Method              string                 `json:"method"`
	Path                string                 `json:"path"`
	Citation            circleCISourceCitation `json:"citation"`
	Facts               circleCIMatrixFacts    `json:"facts"`
	Cells               []circleCIMatrixCell   `json:"cells"`
}

type circleCISourceCitation struct {
	SourceURL string `json:"source_url"`
	Location  string `json:"location"`
}

type circleCIMatrixFacts struct {
	Pagination  circleCIPaginationFact `json:"pagination"`
	Scope       circleCIScopeFact      `json:"scope"`
	Media       circleCIMediaFact      `json:"media"`
	EventCursor circleCIEventCursor    `json:"event_cursor"`
	Write       circleCIWriteFact      `json:"write"`
}

type circleCIPaginationFact struct {
	Kind             string                 `json:"kind"`
	RequestParameter string                 `json:"request_parameter"`
	ResponseField    string                 `json:"response_field"`
	ContinuationKind string                 `json:"continuation_kind"`
	Citation         circleCISourceCitation `json:"citation"`
}

type circleCIScopeFact struct {
	PathParameters  []string               `json:"path_parameters"`
	QueryParameters []string               `json:"query_parameters"`
	ParameterRefs   []string               `json:"parameter_refs"`
	Citation        circleCISourceCitation `json:"citation"`
}

type circleCIMediaFact struct {
	Request  []string               `json:"request"`
	Response []string               `json:"response"`
	Citation circleCISourceCitation `json:"citation"`
}

type circleCIEventCursor struct {
	Kind     string                 `json:"kind"`
	Citation circleCISourceCitation `json:"citation"`
}

type circleCIWriteFact struct {
	Kind               string                      `json:"kind"`
	Summary            string                      `json:"summary"`
	RequestBodyPresent bool                        `json:"request_body_present"`
	RequestSchemaRefs  []string                    `json:"request_schema_refs"`
	SecretShapedFields []circleCISecretShapedField `json:"secret_shaped_fields"`
	Citation           circleCISourceCitation      `json:"citation"`
}

type circleCISecretShapedField struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type circleCIMatrixCell struct {
	Lane           string                  `json:"lane"`
	State          string                  `json:"state"`
	Citation       circleCISourceCitation  `json:"citation"`
	SourceEvidence *circleCISourceEvidence `json:"source_evidence,omitempty"`
}

type circleCISourceEvidence struct {
	Kind     string                 `json:"kind"`
	Citation circleCISourceCitation `json:"citation"`
}

type circleCIMatrixArtifact struct {
	Path        string                       `json:"path"`
	RecordCount int                          `json:"record_count"`
	Links       []circleCIMatrixArtifactLink `json:"links"`
}

type circleCIMatrixArtifactLink struct {
	Record            string `json:"record"`
	SourceOperationID string `json:"source_operation_id"`
	Lane              string `json:"lane"`
}

type circleCIArtifactNamedRecord struct {
	Name string `json:"name"`
}

type circleCIOperationDocument struct {
	Summary     string                             `json:"summary"`
	Parameters  []circleCIOpenAPIParameter         `json:"parameters"`
	RequestBody *circleCIOpenAPIRequestBody        `json:"requestBody"`
	Responses   map[string]circleCIOpenAPIResponse `json:"responses"`
}

type circleCIOpenAPIParameter struct {
	Ref  string `json:"$ref"`
	In   string `json:"in"`
	Name string `json:"name"`
}

type circleCIOpenAPIRequestBody struct {
	Content map[string]circleCIOpenAPIContent `json:"content"`
}

type circleCIOpenAPIResponse struct {
	Content map[string]circleCIOpenAPIContent `json:"content"`
	Links   map[string]circleCIOpenAPILink    `json:"links"`
}

type circleCIOpenAPIContent struct {
	Schema json.RawMessage `json:"schema"`
}

type circleCIOpenAPILink struct {
	Parameters map[string]string `json:"parameters"`
}

type circleCIOpenAPIContract struct {
	Components struct {
		Parameters map[string]circleCIOpenAPIParameter `json:"parameters"`
		Schemas    map[string]circleCIOpenAPISchema    `json:"schemas"`
	} `json:"components"`
}

type circleCIOpenAPISchema struct {
	Required   []string                   `json:"required"`
	Properties map[string]json.RawMessage `json:"properties"`
}

func TestCircleCISourceLaneMatrixReconcilesPinnedSourceLock(t *testing.T) {
	lock := loadCircleCISourceLock(t)
	matrix := loadCircleCISourceLaneMatrix(t)
	if err := validateCircleCISourceLaneMatrix(lock, matrix); err != nil {
		t.Fatal(err)
	}

	if got, want := len(matrix.Operations), 111; got != want {
		t.Fatalf("matrix operation rows = %d, want %d", got, want)
	}
	cellCount := 0
	mappedUnproven := 0
	notApplicable := 0
	for _, operation := range matrix.Operations {
		cellCount += len(operation.Cells)
		for _, cell := range operation.Cells {
			switch cell.State {
			case "mapped_unproven":
				mappedUnproven++
			case "not_applicable":
				notApplicable++
			}
		}
	}
	if got, want := cellCount, 777; got != want {
		t.Fatalf("matrix cell count = %d, want %d", got, want)
	}
	if got, want := mappedUnproven, 179; got != want {
		t.Fatalf("mapped_unproven cells = %d, want %d", got, want)
	}
	if got, want := notApplicable, 598; got != want {
		t.Fatalf("not_applicable cells = %d, want %d", got, want)
	}
}

func TestCircleCISourceLaneMatrixRejectsHiddenSourceRows(t *testing.T) {
	lock := loadCircleCISourceLock(t)
	matrix := cloneCircleCISourceLaneMatrix(t, loadCircleCISourceLaneMatrix(t))
	removed := matrix.Operations[len(matrix.Operations)-1].SourceOperationID
	matrix.Operations = matrix.Operations[:len(matrix.Operations)-1]

	err := validateCircleCISourceLaneMatrix(lock, matrix)
	if err == nil || !strings.Contains(err.Error(), "source row absent from matrix: "+removed) {
		t.Fatalf("hidden source row error = %v, want missing %q", err, removed)
	}
}

func TestCircleCISourceLaneMatrixRejectsInvalidArtifactBacklink(t *testing.T) {
	lock := loadCircleCISourceLock(t)
	matrix := cloneCircleCISourceLaneMatrix(t, loadCircleCISourceLaneMatrix(t))
	link := firstCircleCIArtifactLink(t, &matrix)
	link.SourceOperationID = "circleci.rest.not-a-retained-source-row"

	err := validateCircleCISourceLaneMatrix(lock, matrix)
	if err == nil || !strings.Contains(err.Error(), "artifact backlink") {
		t.Fatalf("invalid artifact backlink error = %v, want backlink rejection", err)
	}
}

func TestCircleCISourceLaneMatrixRejectsMissingPagingETLDisposition(t *testing.T) {
	lock := loadCircleCISourceLock(t)
	matrix := cloneCircleCISourceLaneMatrix(t, loadCircleCISourceLaneMatrix(t))
	operation := findCircleCIMatrixOperation(t, &matrix, "circleci.rest.listPipelines")
	cell := findCircleCIMatrixCell(t, operation, "etl")
	cell.State = "not_applicable"
	cell.SourceEvidence = &circleCISourceEvidence{
		Kind:     "circleci.source.test-only.v1",
		Citation: operation.Citation,
	}

	err := validateCircleCISourceLaneMatrix(lock, matrix)
	if err == nil || !strings.Contains(err.Error(), "paging source operation") {
		t.Fatalf("missing paging ETL disposition error = %v, want paging rejection", err)
	}
}

func TestCircleCISourceLaneMatrixRejectsMissingMutationReverseETLDisposition(t *testing.T) {
	lock := loadCircleCISourceLock(t)
	matrix := cloneCircleCISourceLaneMatrix(t, loadCircleCISourceLaneMatrix(t))
	operation := findCircleCIMatrixOperation(t, &matrix, "circleci.rest.createSchedule")
	cell := findCircleCIMatrixCell(t, operation, "reverse_etl")
	cell.State = "not_applicable"
	cell.SourceEvidence = &circleCISourceEvidence{
		Kind:     "circleci.source.test-only.v1",
		Citation: operation.Citation,
	}

	err := validateCircleCISourceLaneMatrix(lock, matrix)
	if err == nil || !strings.Contains(err.Error(), "mutation source operation") {
		t.Fatalf("missing mutation reverse-ETL disposition error = %v, want mutation rejection", err)
	}
}

func TestCircleCISourceLaneMatrixPreservesSourceFacts(t *testing.T) {
	lock := loadCircleCISourceLock(t)
	matrix := loadCircleCISourceLaneMatrix(t)
	if err := validateCircleCISourceLaneMatrix(lock, matrix); err != nil {
		t.Fatal(err)
	}

	for _, sourceID := range []string{"circleci.rest.createWebhook", "circleci.rest.updateWebhook"} {
		operation := findCircleCIMatrixOperation(t, &matrix, sourceID)
		if len(operation.Facts.Write.SecretShapedFields) != 1 || operation.Facts.Write.SecretShapedFields[0].Name != "signing-secret" {
			t.Fatalf("%s does not preserve its source-declared signing-secret field", sourceID)
		}
	}
	if findCircleCIMatrixOperation(t, &matrix, "circleci.rest.createWebhook").Facts.Write.SecretShapedFields[0].Required != true {
		t.Fatal("createWebhook signing-secret requiredness was not preserved")
	}
	if findCircleCIMatrixOperation(t, &matrix, "circleci.rest.updateWebhook").Facts.Write.SecretShapedFields[0].Required != false {
		t.Fatal("updateWebhook signing-secret requiredness was not preserved")
	}

	for _, operation := range matrix.Operations {
		if strings.Contains(operation.Path, "{project-slug}") && !containsString(operation.Facts.Scope.PathParameters, "project-slug") {
			t.Fatalf("%s dropped source path parameter project-slug", operation.SourceOperationID)
		}
	}
}

func loadCircleCISourceLock(t *testing.T) circleCISourceLock {
	t.Helper()
	raw, err := os.ReadFile(circleCISourceLockPath)
	if err != nil {
		t.Fatalf("read CircleCI source lock: %v", err)
	}
	var lock circleCISourceLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("unmarshal CircleCI source lock: %v", err)
	}
	return lock
}

func loadCircleCISourceLaneMatrix(t *testing.T) circleCISourceLaneMatrix {
	t.Helper()
	raw, err := os.ReadFile(circleCISourceMatrixPath)
	if err != nil {
		t.Fatalf("read CircleCI source-lane matrix: %v", err)
	}
	var matrix circleCISourceLaneMatrix
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&matrix); err != nil {
		t.Fatalf("decode CircleCI source-lane matrix: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("decode CircleCI source-lane matrix: trailing JSON value: %v", err)
	}
	return matrix
}

func cloneCircleCISourceLaneMatrix(t *testing.T, matrix circleCISourceLaneMatrix) circleCISourceLaneMatrix {
	t.Helper()
	raw, err := json.Marshal(matrix)
	if err != nil {
		t.Fatalf("marshal matrix clone: %v", err)
	}
	var clone circleCISourceLaneMatrix
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatalf("unmarshal matrix clone: %v", err)
	}
	return clone
}

func validateCircleCISourceLaneMatrix(lock circleCISourceLock, matrix circleCISourceLaneMatrix) error {
	if lock.Connector != "circleci" || lock.Counts.REST != 111 || lock.Counts.Total != 111 || len(lock.REST.Operations) != 111 {
		return fmt.Errorf("unexpected CircleCI source-lock inventory: connector=%q rest=%d total=%d operations=%d", lock.Connector, lock.Counts.REST, lock.Counts.Total, len(lock.REST.Operations))
	}
	if matrix.SchemaVersion != 1 || matrix.Connector != lock.Connector {
		return fmt.Errorf("matrix identity does not match CircleCI source lock")
	}
	if matrix.SourceLock.Path != circleCISourceLockPath || matrix.SourceLock.SourceURL != lock.REST.SourceURL || matrix.SourceLock.SHA256 != lock.REST.SHA256 || matrix.SourceLock.Bytes != lock.REST.Bytes || matrix.SourceLock.SourceOperationCount != len(lock.REST.Operations) {
		return fmt.Errorf("matrix source-lock reference does not match pinned CircleCI source lock")
	}
	if !equalStringSets(matrix.Lanes, circleCILanes) || len(matrix.Lanes) != len(circleCILanes) {
		return fmt.Errorf("matrix lanes = %v, want exactly %v", matrix.Lanes, circleCILanes)
	}

	contract, err := circleCIParameterContract(lock)
	if err != nil {
		return err
	}
	sourceByID := make(map[string]circleCISourceOperation, len(lock.REST.Operations))
	mutationCount := 0
	for _, source := range lock.REST.Operations {
		if source.ID == "" || source.SourceLocation == "" || source.OperationID == "" {
			return fmt.Errorf("source lock has incomplete source row: %q", source.ID)
		}
		if _, exists := sourceByID[source.ID]; exists {
			return fmt.Errorf("duplicate source operation ID in lock: %s", source.ID)
		}
		sourceByID[source.ID] = source
		if source.Method != "GET" {
			mutationCount++
			if _, known := circleCIMutationOperationIDs[source.ID]; !known {
				return fmt.Errorf("mutation source row is not explicitly dispositioned: %s", source.ID)
			}
		}
	}
	if mutationCount != len(circleCIMutationOperationIDs) {
		return fmt.Errorf("mutation source rows = %d, want %d", mutationCount, len(circleCIMutationOperationIDs))
	}
	for sourceID := range circleCIMutationOperationIDs {
		if _, exists := sourceByID[sourceID]; !exists {
			return fmt.Errorf("expected mutation source row missing from lock: %s", sourceID)
		}
	}

	pagingByID, err := circleCISourcePagingEvidence(lock, contract)
	if err != nil {
		return err
	}
	if len(pagingByID) != len(circleCIPagingOperationIDs) {
		return fmt.Errorf("source-documented paging rows = %d, want %d", len(pagingByID), len(circleCIPagingOperationIDs))
	}
	for sourceID, continuationKind := range circleCIPagingOperationIDs {
		fact, exists := pagingByID[sourceID]
		if !exists || fact.ContinuationKind != continuationKind {
			return fmt.Errorf("source-documented paging evidence mismatch for %s", sourceID)
		}
	}

	matrixByID := make(map[string]circleCIMatrixOperation, len(matrix.Operations))
	cellBySourceAndLane := make(map[string]map[string]circleCIMatrixCell, len(matrix.Operations))
	for _, operation := range matrix.Operations {
		if operation.SourceOperationID == "" {
			return fmt.Errorf("matrix contains a row without a source operation ID")
		}
		if _, exists := matrixByID[operation.SourceOperationID]; exists {
			return fmt.Errorf("duplicate source operation ID in matrix: %s", operation.SourceOperationID)
		}
		source, exists := sourceByID[operation.SourceOperationID]
		if !exists {
			return fmt.Errorf("matrix row has unknown source operation ID: %s", operation.SourceOperationID)
		}
		cells, err := validateCircleCIMatrixOperation(lock, contract, source, operation, pagingByID[operation.SourceOperationID])
		if err != nil {
			return err
		}
		matrixByID[operation.SourceOperationID] = operation
		cellBySourceAndLane[operation.SourceOperationID] = cells
	}
	for _, source := range lock.REST.Operations {
		if _, exists := matrixByID[source.ID]; !exists {
			return fmt.Errorf("source row absent from matrix: %s", source.ID)
		}
	}
	if len(matrixByID) != len(sourceByID) {
		return fmt.Errorf("matrix source row count = %d, want %d", len(matrixByID), len(sourceByID))
	}

	return validateCircleCIArtifactLinks(matrix.Artifacts, cellBySourceAndLane)
}

func validateCircleCIMatrixOperation(lock circleCISourceLock, contract circleCIOpenAPIContract, source circleCISourceOperation, operation circleCIMatrixOperation, paging circleCIPaginationFact) (map[string]circleCIMatrixCell, error) {
	if operation.ProviderOperationID != source.OperationID || operation.Method != source.Method || operation.Path != source.Path {
		return nil, fmt.Errorf("matrix row %s does not preserve source operation identity", source.ID)
	}
	if err := validateCircleCICitation(lock, source, operation.Citation); err != nil {
		return nil, fmt.Errorf("matrix row %s citation: %w", source.ID, err)
	}
	if err := validateCircleCIMatrixFacts(lock, contract, source, operation.Facts, paging); err != nil {
		return nil, fmt.Errorf("matrix row %s facts: %w", source.ID, err)
	}

	cells := make(map[string]circleCIMatrixCell, len(operation.Cells))
	for _, cell := range operation.Cells {
		if !containsString(circleCILanes, cell.Lane) {
			return nil, fmt.Errorf("matrix row %s has unknown lane %q", source.ID, cell.Lane)
		}
		if _, exists := cells[cell.Lane]; exists {
			return nil, fmt.Errorf("matrix row %s has duplicate lane %q", source.ID, cell.Lane)
		}
		if err := validateCircleCICitation(lock, source, cell.Citation); err != nil {
			return nil, fmt.Errorf("matrix row %s lane %s citation: %w", source.ID, cell.Lane, err)
		}
		switch cell.State {
		case "implemented", "mapped_unproven", "missing_foundation", "not_applicable":
		default:
			return nil, fmt.Errorf("matrix row %s lane %s has invalid state %q", source.ID, cell.Lane, cell.State)
		}
		if cell.State == "not_applicable" {
			if cell.SourceEvidence == nil || !isStableCircleCISourceReason(cell.SourceEvidence.Kind) {
				return nil, fmt.Errorf("matrix row %s lane %s has source-evidenced not_applicable without a stable reason", source.ID, cell.Lane)
			}
			if err := validateCircleCICitation(lock, source, cell.SourceEvidence.Citation); err != nil {
				return nil, fmt.Errorf("matrix row %s lane %s not_applicable evidence: %w", source.ID, cell.Lane, err)
			}
		}
		if cell.State == "mapped_unproven" && cell.SourceEvidence != nil {
			return nil, fmt.Errorf("matrix row %s lane %s has source evidence attached to mapped_unproven", source.ID, cell.Lane)
		}
		if cell.State == "implemented" || cell.State == "missing_foundation" {
			return nil, fmt.Errorf("matrix row %s lane %s claims runtime state %q in source-only Track A", source.ID, cell.Lane, cell.State)
		}
		cells[cell.Lane] = cell
	}
	if !equalStringSets(sortedKeys(cells), circleCILanes) || len(cells) != len(circleCILanes) {
		return nil, fmt.Errorf("matrix row %s does not contain exactly one of every lane", source.ID)
	}

	if source.Method == "GET" {
		if err := requireCircleCICellState(source.ID, cells, "direct_read", "mapped_unproven"); err != nil {
			return nil, err
		}
		if err := requireCircleCICellState(source.ID, cells, "direct_write", "not_applicable"); err != nil {
			return nil, err
		}
		if err := requireCircleCICellState(source.ID, cells, "reverse_etl", "not_applicable"); err != nil {
			return nil, err
		}
	} else {
		if err := requireCircleCICellState(source.ID, cells, "direct_read", "not_applicable"); err != nil {
			return nil, err
		}
		if err := requireCircleCICellState(source.ID, cells, "direct_write", "mapped_unproven"); err != nil {
			return nil, err
		}
		if err := requireCircleCICellState(source.ID, cells, "reverse_etl", "mapped_unproven"); err != nil {
			return nil, fmt.Errorf("mutation source operation %s requires independent reverse_etl mapped_unproven: %w", source.ID, err)
		}
	}
	for _, lane := range []string{"binary_download", "binary_upload"} {
		if err := requireCircleCICellState(source.ID, cells, lane, "not_applicable"); err != nil {
			return nil, err
		}
	}
	if paging.Kind == "cursor" {
		for _, lane := range []string{"etl", "sync_transport"} {
			if err := requireCircleCICellState(source.ID, cells, lane, "mapped_unproven"); err != nil {
				return nil, fmt.Errorf("paging source operation %s requires explicit %s mapped_unproven: %w", source.ID, lane, err)
			}
		}
	} else {
		for _, lane := range []string{"etl", "sync_transport"} {
			if err := requireCircleCICellState(source.ID, cells, lane, "not_applicable"); err != nil {
				return nil, err
			}
		}
	}
	return cells, nil
}

func validateCircleCIMatrixFacts(lock circleCISourceLock, contract circleCIOpenAPIContract, source circleCISourceOperation, facts circleCIMatrixFacts, paging circleCIPaginationFact) error {
	for _, citation := range []circleCISourceCitation{facts.Pagination.Citation, facts.Scope.Citation, facts.Media.Citation, facts.EventCursor.Citation, facts.Write.Citation} {
		if err := validateCircleCICitation(lock, source, citation); err != nil {
			return err
		}
	}
	document, err := circleCIOperationDocumentFor(source)
	if err != nil {
		return err
	}
	pathParameters, queryParameters, parameterRefs, err := circleCIScopeFor(contract, source, document)
	if err != nil {
		return err
	}
	if !equalStringSets(facts.Scope.PathParameters, pathParameters) || !equalStringSets(facts.Scope.QueryParameters, queryParameters) || !equalStringSets(facts.Scope.ParameterRefs, parameterRefs) {
		return fmt.Errorf("scope facts do not match retained source path and parameters")
	}
	requestMedia, responseMedia := circleCIMediaFor(document)
	if !equalStringSets(facts.Media.Request, requestMedia) || !equalStringSets(facts.Media.Response, responseMedia) {
		return fmt.Errorf("media facts do not match retained source operation")
	}
	if circleCIHasBinaryMedia(requestMedia) || circleCIHasBinaryMedia(responseMedia) {
		return fmt.Errorf("binary media is present in the retained source operation but both binary lanes are not_applicable")
	}
	if facts.EventCursor.Kind != "not_documented" {
		return fmt.Errorf("event/cursor fact must remain not_documented for this retained operation")
	}
	if facts.Write.Kind != map[bool]string{true: "read", false: "mutation"}[source.Method == "GET"] || facts.Write.Summary != document.Summary || facts.Write.RequestBodyPresent != (document.RequestBody != nil) {
		return fmt.Errorf("write/read source facts do not match retained operation")
	}
	if paging.Kind == "cursor" {
		if facts.Pagination.Kind != "cursor" || facts.Pagination.RequestParameter != paging.RequestParameter || facts.Pagination.ResponseField != paging.ResponseField || facts.Pagination.ContinuationKind != paging.ContinuationKind {
			return fmt.Errorf("paging facts do not match source-documented continuation")
		}
	} else {
		expectedKind := "not_documented"
		if circleCIResponseHasNextPageToken(document) {
			expectedKind = "response_token_without_operation_continuation"
		}
		if facts.Pagination.Kind != expectedKind || facts.Pagination.RequestParameter != "" || facts.Pagination.ResponseField != "" || facts.Pagination.ContinuationKind != "" {
			return fmt.Errorf("pagination facts do not preserve the retained source evidence")
		}
	}
	requestSchemaRefs := circleCIRequestSchemaReferences(document)
	if !equalStringSets(facts.Write.RequestSchemaRefs, requestSchemaRefs) {
		return fmt.Errorf("request schema references do not match retained source operation")
	}
	if err := validateCircleCIWebhookSecretFact(contract, source.ID, facts.Write.SecretShapedFields, requestSchemaRefs); err != nil {
		return err
	}
	return nil
}

func validateCircleCIWebhookSecretFact(contract circleCIOpenAPIContract, sourceID string, fields []circleCISecretShapedField, schemaRefs []string) error {
	expected, err := circleCISecretShapedFields(contract, schemaRefs)
	if err != nil {
		return err
	}
	if len(fields) != len(expected) {
		return fmt.Errorf("%s secret-shaped source facts = %d, want %d", sourceID, len(fields), len(expected))
	}
	for index := range expected {
		if fields[index] != expected[index] {
			return fmt.Errorf("%s secret-shaped source fact %d = %+v, want %+v", sourceID, index, fields[index], expected[index])
		}
	}
	return nil
}

func circleCISecretShapedFields(contract circleCIOpenAPIContract, schemaRefs []string) ([]circleCISecretShapedField, error) {
	fields := []circleCISecretShapedField{}
	for _, ref := range schemaRefs {
		const prefix = "#/components/schemas/"
		if !strings.HasPrefix(ref, prefix) {
			return nil, fmt.Errorf("unsupported request schema reference %q", ref)
		}
		schema, exists := contract.Components.Schemas[strings.TrimPrefix(ref, prefix)]
		if !exists {
			return nil, fmt.Errorf("request schema reference %q not found", ref)
		}
		if _, exists := schema.Properties["signing-secret"]; exists {
			fields = append(fields, circleCISecretShapedField{Name: "signing-secret", Required: containsString(schema.Required, "signing-secret")})
		}
	}
	return fields, nil
}

func validateCircleCIArtifactLinks(artifacts []circleCIMatrixArtifact, cells map[string]map[string]circleCIMatrixCell) error {
	wantArtifactPaths := map[string]struct{}{
		"api_surface.json": {},
		"streams.json":     {},
		"writes.json":      {},
	}
	if len(artifacts) != len(wantArtifactPaths) {
		return fmt.Errorf("matrix artifact inventory count = %d, want %d", len(artifacts), len(wantArtifactPaths))
	}
	seenArtifacts := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Path == "" || artifact.RecordCount < 0 {
			return fmt.Errorf("invalid artifact inventory row")
		}
		if _, exists := wantArtifactPaths[artifact.Path]; !exists {
			return fmt.Errorf("matrix references unsupported artifact %s", artifact.Path)
		}
		if _, exists := seenArtifacts[artifact.Path]; exists {
			return fmt.Errorf("matrix has duplicate artifact inventory %s", artifact.Path)
		}
		seenArtifacts[artifact.Path] = struct{}{}
		records, recordCount, err := circleCIArtifactRecords(artifact.Path)
		if err != nil {
			return err
		}
		if artifact.RecordCount != recordCount {
			return fmt.Errorf("artifact %s record count = %d, want %d from the retained artifact", artifact.Path, artifact.RecordCount, recordCount)
		}
		for _, link := range artifact.Links {
			if link.Record == "" || link.SourceOperationID == "" || link.Lane == "" {
				return fmt.Errorf("artifact backlink in %s is incomplete", artifact.Path)
			}
			if len(records) == 0 || !records[link.Record] {
				return fmt.Errorf("artifact backlink in %s references unknown artifact record %s", artifact.Path, link.Record)
			}
			operationCells, exists := cells[link.SourceOperationID]
			if !exists {
				return fmt.Errorf("artifact backlink in %s record %s references unknown source operation %s", artifact.Path, link.Record, link.SourceOperationID)
			}
			if _, exists := operationCells[link.Lane]; !exists {
				return fmt.Errorf("artifact backlink in %s record %s references nonexistent cell %s/%s", artifact.Path, link.Record, link.SourceOperationID, link.Lane)
			}
		}
	}
	if len(seenArtifacts) != len(wantArtifactPaths) {
		return fmt.Errorf("matrix artifact inventory is incomplete")
	}
	return nil
}

func circleCIArtifactRecords(path string) (map[string]bool, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("read retained artifact %s: %w", path, err)
	}
	switch path {
	case "api_surface.json":
		var artifact struct {
			Endpoints []json.RawMessage `json:"endpoints"`
		}
		if err := json.Unmarshal(raw, &artifact); err != nil {
			return nil, 0, fmt.Errorf("decode retained artifact %s: %w", path, err)
		}
		return map[string]bool{}, len(artifact.Endpoints), nil
	case "streams.json":
		var artifact struct {
			Streams []circleCIArtifactNamedRecord `json:"streams"`
		}
		if err := json.Unmarshal(raw, &artifact); err != nil {
			return nil, 0, fmt.Errorf("decode retained artifact %s: %w", path, err)
		}
		return namedCircleCIArtifactRecords(artifact.Streams), len(artifact.Streams), nil
	case "writes.json":
		var artifact struct {
			Actions []circleCIArtifactNamedRecord `json:"actions"`
		}
		if err := json.Unmarshal(raw, &artifact); err != nil {
			return nil, 0, fmt.Errorf("decode retained artifact %s: %w", path, err)
		}
		return namedCircleCIArtifactRecords(artifact.Actions), len(artifact.Actions), nil
	default:
		return nil, 0, fmt.Errorf("unsupported retained artifact %s", path)
	}
}

func namedCircleCIArtifactRecords(records []circleCIArtifactNamedRecord) map[string]bool {
	result := make(map[string]bool, len(records))
	for _, record := range records {
		if record.Name != "" {
			result[record.Name] = true
		}
	}
	return result
}

func circleCIParameterContract(lock circleCISourceLock) (circleCIOpenAPIContract, error) {
	var contract circleCIOpenAPIContract
	if err := json.Unmarshal(lock.SourceContract, &contract); err != nil {
		return circleCIOpenAPIContract{}, fmt.Errorf("decode retained CircleCI source contract: %w", err)
	}
	return contract, nil
}

func circleCIOperationDocumentFor(source circleCISourceOperation) (circleCIOperationDocument, error) {
	var document circleCIOperationDocument
	if err := json.Unmarshal(source.SourceOperation, &document); err != nil {
		return circleCIOperationDocument{}, fmt.Errorf("decode retained source operation %s: %w", source.ID, err)
	}
	return document, nil
}

func circleCIScopeFor(contract circleCIOpenAPIContract, source circleCISourceOperation, document circleCIOperationDocument) ([]string, []string, []string, error) {
	pathParameters := circleCIPathParameters(source.Path)
	queryParameters := []string{}
	parameterRefs := []string{}
	for _, parameter := range document.Parameters {
		if parameter.Ref != "" {
			parameterRefs = append(parameterRefs, parameter.Ref)
			resolved, err := circleCIResolveParameter(contract, parameter.Ref)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("resolve %s parameter %s: %w", source.ID, parameter.Ref, err)
			}
			parameter = resolved
		}
		switch parameter.In {
		case "query":
			queryParameters = append(queryParameters, parameter.Name)
		case "path":
			// Path variables are represented from the authoritative route template above.
		}
	}
	return uniqueSorted(pathParameters), uniqueSorted(queryParameters), uniqueSorted(parameterRefs), nil
}

func circleCIResolveParameter(contract circleCIOpenAPIContract, ref string) (circleCIOpenAPIParameter, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return circleCIOpenAPIParameter{}, fmt.Errorf("unsupported parameter reference %q", ref)
	}
	parameter, exists := contract.Components.Parameters[strings.TrimPrefix(ref, prefix)]
	if !exists {
		return circleCIOpenAPIParameter{}, fmt.Errorf("parameter reference %q not found", ref)
	}
	return parameter, nil
}

func circleCIMediaFor(document circleCIOperationDocument) ([]string, []string) {
	request := []string{}
	if document.RequestBody != nil {
		for mediaType := range document.RequestBody.Content {
			request = append(request, mediaType)
		}
	}
	response := []string{}
	for _, item := range document.Responses {
		for mediaType := range item.Content {
			response = append(response, mediaType)
		}
	}
	return uniqueSorted(request), uniqueSorted(response)
}

func circleCIRequestSchemaReferences(document circleCIOperationDocument) []string {
	if document.RequestBody == nil {
		return []string{}
	}
	references := []string{}
	for _, content := range document.RequestBody.Content {
		var schema struct {
			Ref string `json:"$ref"`
		}
		if err := json.Unmarshal(content.Schema, &schema); err == nil && schema.Ref != "" {
			references = append(references, schema.Ref)
		}
	}
	return uniqueSorted(references)
}

func circleCIHasBinaryMedia(mediaTypes []string) bool {
	for _, mediaType := range mediaTypes {
		if strings.HasPrefix(mediaType, "image/") || strings.HasPrefix(mediaType, "audio/") || strings.HasPrefix(mediaType, "video/") || strings.HasPrefix(mediaType, "multipart/") || strings.Contains(mediaType, "octet-stream") || strings.Contains(mediaType, "pdf") || strings.Contains(mediaType, "zip") {
			return true
		}
	}
	return false
}

func circleCISourcePagingEvidence(lock circleCISourceLock, contract circleCIOpenAPIContract) (map[string]circleCIPaginationFact, error) {
	results := make(map[string]circleCIPaginationFact)
	for _, source := range lock.REST.Operations {
		if source.Method != "GET" {
			continue
		}
		document, err := circleCIOperationDocumentFor(source)
		if err != nil {
			return nil, err
		}
		_, queryParameters, _, err := circleCIScopeFor(contract, source, document)
		if err != nil {
			return nil, err
		}
		if !circleCIResponseHasNextPageToken(document) {
			continue
		}
		continuationKind := ""
		if containsString(queryParameters, "page-token") {
			continuationKind = "query_parameter"
		} else if circleCIHasNextPageLink(document) {
			continuationKind = "openapi_link"
		}
		if continuationKind != "" {
			results[source.ID] = circleCIPaginationFact{
				Kind:             "cursor",
				RequestParameter: "page-token",
				ResponseField:    "next_page_token",
				ContinuationKind: continuationKind,
				Citation: circleCISourceCitation{
					SourceURL: lock.REST.SourceURL,
					Location:  source.SourceLocation,
				},
			}
		}
	}
	return results, nil
}

func circleCIResponseHasNextPageToken(document circleCIOperationDocument) bool {
	for _, response := range document.Responses {
		for _, content := range response.Content {
			if circleCISchemaHasProperty(content.Schema, "next_page_token") {
				return true
			}
		}
	}
	return false
}

func circleCISchemaHasProperty(raw json.RawMessage, property string) bool {
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return false
	}
	_, exists := schema.Properties[property]
	return exists
}

func circleCIHasNextPageLink(document circleCIOperationDocument) bool {
	for _, response := range document.Responses {
		for _, link := range response.Links {
			if link.Parameters["page-token"] == "$response.body#/next_page_token" {
				return true
			}
		}
	}
	return false
}

func requireCircleCICellState(sourceID string, cells map[string]circleCIMatrixCell, lane, state string) error {
	cell, exists := cells[lane]
	if !exists || cell.State != state {
		got := "missing"
		if exists {
			got = cell.State
		}
		return fmt.Errorf("source operation %s lane %s = %s, want %s", sourceID, lane, got, state)
	}
	return nil
}

func validateCircleCICitation(lock circleCISourceLock, source circleCISourceOperation, citation circleCISourceCitation) error {
	if citation.SourceURL != lock.REST.SourceURL || citation.Location != source.SourceLocation {
		return fmt.Errorf("citation = %q at %q, want %q at %q", citation.SourceURL, citation.Location, lock.REST.SourceURL, source.SourceLocation)
	}
	return nil
}

func findCircleCIMatrixOperation(t *testing.T, matrix *circleCISourceLaneMatrix, sourceID string) *circleCIMatrixOperation {
	t.Helper()
	for index := range matrix.Operations {
		if matrix.Operations[index].SourceOperationID == sourceID {
			return &matrix.Operations[index]
		}
	}
	t.Fatalf("matrix operation %s not found", sourceID)
	return nil
}

func findCircleCIMatrixCell(t *testing.T, operation *circleCIMatrixOperation, lane string) *circleCIMatrixCell {
	t.Helper()
	for index := range operation.Cells {
		if operation.Cells[index].Lane == lane {
			return &operation.Cells[index]
		}
	}
	t.Fatalf("matrix cell %s/%s not found", operation.SourceOperationID, lane)
	return nil
}

func firstCircleCIArtifactLink(t *testing.T, matrix *circleCISourceLaneMatrix) *circleCIMatrixArtifactLink {
	t.Helper()
	for artifactIndex := range matrix.Artifacts {
		if len(matrix.Artifacts[artifactIndex].Links) > 0 {
			return &matrix.Artifacts[artifactIndex].Links[0]
		}
	}
	t.Fatal("matrix has no artifact links")
	return nil
}

func circleCIPathParameters(path string) []string {
	parameters := []string{}
	for rest := path; ; {
		start := strings.IndexByte(rest, '{')
		if start < 0 {
			break
		}
		endOffset := strings.IndexByte(rest[start+1:], '}')
		if endOffset < 0 {
			break
		}
		end := start + 1 + endOffset
		parameters = append(parameters, rest[start+1:end])
		rest = rest[end+1:]
	}
	return uniqueSorted(parameters)
}

func equalStringSets(got, want []string) bool {
	return strings.Join(uniqueSorted(got), "\x00") == strings.Join(uniqueSorted(want), "\x00")
}

func uniqueSorted(values []string) []string {
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return compactStrings(copyValues)
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func sortedKeys(values map[string]circleCIMatrixCell) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return uniqueSorted(keys)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func isStableCircleCISourceReason(reason string) bool {
	return strings.HasPrefix(reason, "circleci.source.") && strings.HasSuffix(reason, ".v1")
}
