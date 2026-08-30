package dockerhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	dockerHubMatrixPath     = "sources/dockerhub-source-lane-matrix.json"
	dockerHubSourceLockPath = "sources/dockerhub-operation-source-lock.json"
	dockerHubCrosswalkPath  = "sources/dockerhub-operation-crosswalk.json"
)

var dockerHubLaneOrder = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

type dockerHubLock struct {
	Connector      string              `json:"connector"`
	CapturedAt     string              `json:"captured_at"`
	Counts         dockerHubLockCounts `json:"counts"`
	Rest           dockerHubRestLock   `json:"rest"`
	SourceContract map[string]any      `json:"source_contract"`
}

type dockerHubLockCounts struct {
	Rest  int `json:"rest"`
	Total int `json:"total"`
}

type dockerHubRestLock struct {
	SourceURL  string                     `json:"source_url"`
	SHA256     string                     `json:"sha256"`
	Operations []dockerHubSourceOperation `json:"operations"`
}

type dockerHubSourceOperation struct {
	ID             string         `json:"id"`
	Protocol       string         `json:"protocol"`
	Method         string         `json:"method"`
	Path           string         `json:"path"`
	OperationID    string         `json:"operation_id"`
	SourceLocation string         `json:"source_location"`
	Operation      map[string]any `json:"source_operation"`
}

type dockerHubCrosswalk struct {
	Connector        string                     `json:"connector"`
	SourceLock       string                     `json:"source_lock"`
	Accounting       dockerHubCrosswalkCounts   `json:"accounting"`
	SourceOperations []dockerHubCrosswalkRecord `json:"source_operations"`
}

type dockerHubCrosswalkCounts struct {
	SourceOperations           int `json:"source_operations"`
	SourceUniqueMethodPath     int `json:"source_unique_method_path"`
	APISurfaceEndpoints        int `json:"api_surface_endpoints"`
	APISurfaceUniqueMethodPath int `json:"api_surface_unique_method_path"`
	ExactSourceToSurface       int `json:"exact_source_to_surface"`
	SourceOnly                 int `json:"source_only"`
	SurfaceOnly                int `json:"surface_only"`
}

type dockerHubCrosswalkRecord struct {
	ID             string                     `json:"id"`
	SourceURL      string                     `json:"source_url"`
	Method         string                     `json:"method"`
	Path           string                     `json:"path"`
	OperationID    string                     `json:"operation_id"`
	SourceLocation string                     `json:"source_location"`
	Request        dockerHubCrosswalkRequest  `json:"request"`
	Response       dockerHubCrosswalkResponse `json:"response"`
}

type dockerHubCrosswalkRequest struct {
	PathParameters   []dockerHubParameter `json:"path_parameters"`
	QueryParameters  []dockerHubParameter `json:"query_parameters"`
	BodyContentTypes []string             `json:"body_content_types"`
}

type dockerHubCrosswalkResponse struct {
	ContentTypes []string `json:"content_types"`
}

type dockerHubParameter struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type dockerHubMatrix struct {
	SchemaVersion int                        `json:"schema_version"`
	Connector     string                     `json:"connector"`
	SourceLock    dockerHubMatrixSourceLock  `json:"source_lock"`
	Lanes         []string                   `json:"lanes"`
	Operations    []dockerHubMatrixOperation `json:"operations"`
	Artifacts     []dockerHubMatrixArtifact  `json:"artifacts"`
	BacklinkGaps  []any                      `json:"backlink_gaps"`
}

type dockerHubMatrixSourceLock struct {
	Path                 string `json:"path"`
	SourceURL            string `json:"source_url"`
	SHA256               string `json:"sha256"`
	CapturedAt           string `json:"captured_at"`
	SourceOperationCount int    `json:"source_operation_count"`
}

type dockerHubMatrixOperation struct {
	SourceID       string                `json:"source_id"`
	Protocol       string                `json:"protocol"`
	Method         string                `json:"method"`
	Path           string                `json:"path"`
	OperationID    string                `json:"operation_id"`
	SourceURL      string                `json:"source_url"`
	SourceLocation string                `json:"source_location"`
	Facts          dockerHubMatrixFacts  `json:"facts"`
	Cells          []dockerHubMatrixCell `json:"cells"`
}

type dockerHubMatrixFacts struct {
	Classification string                      `json:"classification"`
	Pagination     dockerHubPaginationFact     `json:"pagination"`
	Scope          dockerHubScopeFact          `json:"scope"`
	Media          dockerHubMediaFact          `json:"media"`
	EventCursor    dockerHubEventCursorFact    `json:"event_cursor"`
	Extractability dockerHubExtractabilityFact `json:"extractability"`
}

type dockerHubPaginationFact struct {
	Kind            string   `json:"kind"`
	QueryParameters []string `json:"query_parameters"`
	Description     string   `json:"description"`
	SourceLocation  string   `json:"source_location"`
}

type dockerHubScopeFact struct {
	PathParameters          []string `json:"path_parameters"`
	RequiredPathParameters  []string `json:"required_path_parameters"`
	QueryParameters         []string `json:"query_parameters"`
	RequiredQueryParameters []string `json:"required_query_parameters"`
	SourceLocation          string   `json:"source_location"`
}

type dockerHubMediaFact struct {
	Request        []string `json:"request"`
	Response       []string `json:"response"`
	SourceLocation string   `json:"source_location"`
}

type dockerHubEventCursorFact struct {
	Kind                string `json:"kind"`
	Parameter           string `json:"parameter"`
	Description         string `json:"description"`
	DocumentedCallbacks int    `json:"documented_callbacks"`
	SourceLocation      string `json:"source_location"`
}

type dockerHubExtractabilityFact struct {
	Kind             string   `json:"kind"`
	ResponseStatuses []string `json:"response_statuses"`
	SourceLocation   string   `json:"source_location"`
}

type dockerHubMatrixCell struct {
	Lane           string `json:"lane"`
	State          string `json:"state"`
	Reason         string `json:"reason"`
	SourceLocation string `json:"source_location"`
	SourceURL      string `json:"source_url"`
}

type dockerHubMatrixArtifact struct {
	Path        string                  `json:"path"`
	RecordCount int                     `json:"record_count"`
	Links       []dockerHubArtifactLink `json:"links"`
}

type dockerHubArtifactLink struct {
	Record   string   `json:"record"`
	SourceID string   `json:"source_id"`
	Lanes    []string `json:"lanes"`
}

type dockerHubSourceInfo struct {
	Lock      dockerHubSourceOperation
	Crosswalk dockerHubCrosswalkRecord
	SourceURL string
}

type dockerHubArtifactRecords struct {
	APIRoutes      map[string]struct{}
	Streams        map[string]dockerHubStreamRecord
	Commands       map[string]dockerHubCommandRecord
	OperationCount int
}

type dockerHubAPISurfaceDocument struct {
	Endpoints []dockerHubRoute `json:"endpoints"`
}

type dockerHubStreamsDocument struct {
	Streams []dockerHubStreamRecord `json:"streams"`
}

type dockerHubStreamRecord struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type dockerHubOperationsDocument struct {
	Operations []json.RawMessage `json:"operations"`
}

type dockerHubCLISurfaceDocument struct {
	Commands []dockerHubCommandRecord `json:"commands"`
}

type dockerHubCommandRecord struct {
	Path       string           `json:"path"`
	Stream     string           `json:"stream"`
	Intent     string           `json:"intent"`
	APISurface []dockerHubRoute `json:"api_surface"`
}

type dockerHubRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

func TestDockerHubSourceLaneMatrixContract(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	if err := validateDockerHubSourceLaneMatrix(matrix, lock, crosswalk, readDockerHubArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}
}

func TestDockerHubSourceLaneMatrixRejectsHiddenSourceRow(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	matrix.Operations = matrix.Operations[:len(matrix.Operations)-1]

	assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "source row absent from matrix")
}

func TestDockerHubSourceLaneMatrixRejectsDuplicateSourceID(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	matrix.Operations = append(matrix.Operations, matrix.Operations[0])

	assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "duplicate matrix source id")
}

func TestDockerHubSourceLaneMatrixRejectsInvalidArtifactBacklink(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	dockerHubMatrixArtifactByPath(t, matrix, "api_surface.json").Links[0].SourceID = "dockerhub.rest.NoSuchOperation"

	assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "artifact api_surface.json link")
}

func TestDockerHubSourceLaneMatrixRejectsNonexistentArtifactCell(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	dockerHubMatrixArtifactByPath(t, matrix, "api_surface.json").Links[0].Lanes = []string{"no_such_lane"}

	assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "references nonexistent cell")
}

func TestDockerHubSourceLaneMatrixRejectsMissingSourceFact(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	matrix.Operations[0].Facts.Scope.SourceLocation = ""

	assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "does not preserve exact scope/path evidence")
}

func TestDockerHubSourceLaneMatrixRejectsMissingPagingOrExtractableETLDisposition(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	sources, err := dockerHubSourceInfos(lock, crosswalk)
	if err != nil {
		t.Fatal(err)
	}
	for index := range matrix.Operations {
		source := sources[matrix.Operations[index].SourceID]
		if !dockerHubRequiresETL(source, lock.SourceContract) {
			continue
		}
		dockerHubMatrixCellByLane(t, &matrix.Operations[index], "etl").State = "not_applicable"
		assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "pageable or extractable source operation")
		return
	}
	t.Fatal("matrix has no pageable or extractable source operation")
}

func TestDockerHubSourceLaneMatrixRejectsMissingMutationDirectWriteDisposition(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	for index := range matrix.Operations {
		if !dockerHubIsMutation(matrix.Operations[index].Method) {
			continue
		}
		dockerHubMatrixCellByLane(t, &matrix.Operations[index], "direct_write").State = "not_applicable"
		assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "mutation source operation")
		return
	}
	t.Fatal("matrix has no mutation source operation")
}

func TestDockerHubSourceLaneMatrixRejectsMissingMutationReverseETLDisposition(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	for index := range matrix.Operations {
		if !dockerHubIsMutation(matrix.Operations[index].Method) {
			continue
		}
		dockerHubMatrixCellByLane(t, &matrix.Operations[index], "reverse_etl").State = "not_applicable"
		assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "mutation source operation")
		return
	}
	t.Fatal("matrix has no mutation source operation")
}

func TestDockerHubSourceLaneMatrixRejectsSourceCountMismatch(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	matrix.SourceLock.SourceOperationCount = 53

	assertDockerHubMatrixValidationError(t, matrix, lock, crosswalk, "source lock operation count")
}

func TestDockerHubSourceLaneMatrixPreservesRetainedSurface(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	if err := validateDockerHubSourceLaneMatrix(matrix, lock, crosswalk, readDockerHubArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}

	sources, err := dockerHubSourceInfos(lock, crosswalk)
	if err != nil {
		t.Fatal(err)
	}
	var scimOperations int
	for _, source := range sources {
		if strings.Contains(source.Lock.Path, "/scim/2.0/") {
			scimOperations++
			_, responseMedia := dockerHubSourceMedia(source)
			if !dockerHubContainsString(responseMedia, "application/scim+json") {
				t.Fatalf("SCIM source operation %s does not preserve application/scim+json response evidence", source.Lock.ID)
			}
		}
	}
	if scimOperations != 9 {
		t.Fatalf("SCIM source operation count = %d, want 9", scimOperations)
	}

	export := dockerHubMatrixOperationByID(t, matrix, "dockerhub.rest.get_/v2/orgs/{org_name}/members/export")
	if !dockerHubContainsString(export.Facts.Media.Response, "text/csv") {
		t.Fatal("member export source media does not retain text/csv")
	}
	if cell := dockerHubMatrixCellByLane(t, export, "binary_download"); cell.State != "not_applicable" {
		t.Fatalf("text/csv member export binary_download state = %q, want source-evidenced not_applicable", cell.State)
	}
}

func readDockerHubMatrixInputs(t *testing.T) (*dockerHubMatrix, *dockerHubLock, *dockerHubCrosswalk) {
	t.Helper()
	matrix := new(dockerHubMatrix)
	lock := new(dockerHubLock)
	crosswalk := new(dockerHubCrosswalk)
	readDockerHubJSON(t, dockerHubMatrixPath, matrix)
	readDockerHubJSON(t, dockerHubSourceLockPath, lock)
	readDockerHubJSON(t, dockerHubCrosswalkPath, crosswalk)
	return matrix, lock, crosswalk
}

func readDockerHubJSON(t *testing.T, path string, target any) {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(bytes, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func assertDockerHubMatrixValidationError(t *testing.T, matrix *dockerHubMatrix, lock *dockerHubLock, crosswalk *dockerHubCrosswalk, want string) {
	t.Helper()
	err := validateDockerHubSourceLaneMatrix(matrix, lock, crosswalk, readDockerHubArtifactRecords(t))
	if err == nil {
		t.Fatalf("matrix validation unexpectedly passed, want error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("matrix validation error = %q, want substring %q", err, want)
	}
}

func validateDockerHubSourceLaneMatrix(matrix *dockerHubMatrix, lock *dockerHubLock, crosswalk *dockerHubCrosswalk, artifacts dockerHubArtifactRecords) error {
	if matrix.SchemaVersion != 1 {
		return fmt.Errorf("matrix schema_version = %d, want 1", matrix.SchemaVersion)
	}
	if matrix.Connector != "dockerhub" {
		return fmt.Errorf("matrix connector = %q, want dockerhub", matrix.Connector)
	}
	if !dockerHubEqualStrings(matrix.Lanes, dockerHubLaneOrder) {
		return fmt.Errorf("matrix lanes = %v, want %v", matrix.Lanes, dockerHubLaneOrder)
	}

	sources, err := dockerHubSourceInfos(lock, crosswalk)
	if err != nil {
		return err
	}
	if err := dockerHubValidateSourceLockReference(matrix.SourceLock, lock, len(sources)); err != nil {
		return err
	}

	matrixByID := make(map[string]*dockerHubMatrixOperation, len(matrix.Operations))
	for index := range matrix.Operations {
		operation := &matrix.Operations[index]
		if _, exists := matrixByID[operation.SourceID]; exists {
			return fmt.Errorf("duplicate matrix source id %q", operation.SourceID)
		}
		matrixByID[operation.SourceID] = operation
	}
	for _, sourceID := range dockerHubSortedSourceIDs(sources) {
		operation, exists := matrixByID[sourceID]
		if !exists {
			return fmt.Errorf("source row absent from matrix: %s", sourceID)
		}
		if err := dockerHubValidateOperation(operation, sources[sourceID], lock.SourceContract); err != nil {
			return err
		}
	}
	for sourceID := range matrixByID {
		if _, exists := sources[sourceID]; !exists {
			return fmt.Errorf("matrix source id %q is not in the source lock", sourceID)
		}
	}
	if len(matrixByID) != len(sources) {
		return fmt.Errorf("matrix operation count = %d, source lock count = %d", len(matrixByID), len(sources))
	}
	if err := dockerHubValidateCountReconciliation(matrix, lock, sources); err != nil {
		return err
	}
	return dockerHubValidateArtifactLinks(matrix, artifacts, matrixByID, sources)
}

func dockerHubSourceInfos(lock *dockerHubLock, crosswalk *dockerHubCrosswalk) (map[string]dockerHubSourceInfo, error) {
	if lock.Connector != "dockerhub" || crosswalk.Connector != "dockerhub" {
		return nil, errors.New("Docker Hub source artifacts have the wrong connector identity")
	}
	if crosswalk.SourceLock != dockerHubSourceLockPath {
		return nil, fmt.Errorf("Docker Hub crosswalk source lock = %q, want %q", crosswalk.SourceLock, dockerHubSourceLockPath)
	}
	if len(lock.Rest.Operations) != len(crosswalk.SourceOperations) {
		return nil, fmt.Errorf("Docker Hub source artifact operation counts differ: lock:%d crosswalk:%d", len(lock.Rest.Operations), len(crosswalk.SourceOperations))
	}
	accounting := crosswalk.Accounting
	if accounting.SourceOperations != len(lock.Rest.Operations) ||
		accounting.SourceUniqueMethodPath != len(lock.Rest.Operations) ||
		accounting.APISurfaceEndpoints != len(lock.Rest.Operations) ||
		accounting.APISurfaceUniqueMethodPath != len(lock.Rest.Operations) ||
		accounting.ExactSourceToSurface != len(lock.Rest.Operations) ||
		accounting.SourceOnly != 0 ||
		accounting.SurfaceOnly != 0 {
		return nil, fmt.Errorf("Docker Hub crosswalk accounting does not reconcile the retained source rows: %+v", accounting)
	}
	crosswalkByID := make(map[string]dockerHubCrosswalkRecord, len(crosswalk.SourceOperations))
	for _, operation := range crosswalk.SourceOperations {
		if _, exists := crosswalkByID[operation.ID]; exists {
			return nil, fmt.Errorf("duplicate Docker Hub crosswalk source id %q", operation.ID)
		}
		crosswalkByID[operation.ID] = operation
	}
	sources := make(map[string]dockerHubSourceInfo, len(lock.Rest.Operations))
	for _, operation := range lock.Rest.Operations {
		if _, exists := sources[operation.ID]; exists {
			return nil, fmt.Errorf("duplicate Docker Hub lock source id %q", operation.ID)
		}
		crosswalkOperation, exists := crosswalkByID[operation.ID]
		if !exists {
			return nil, fmt.Errorf("Docker Hub source lock operation %s is absent from the crosswalk", operation.ID)
		}
		if crosswalkOperation.SourceURL != lock.Rest.SourceURL ||
			crosswalkOperation.Method != operation.Method ||
			crosswalkOperation.Path != operation.Path ||
			crosswalkOperation.OperationID != operation.OperationID ||
			crosswalkOperation.SourceLocation != operation.SourceLocation {
			return nil, fmt.Errorf("Docker Hub source artifacts disagree about operation identity for %s", operation.ID)
		}
		sources[operation.ID] = dockerHubSourceInfo{
			Lock:      operation,
			Crosswalk: crosswalkOperation,
			SourceURL: lock.Rest.SourceURL,
		}
	}
	return sources, nil
}

func dockerHubValidateSourceLockReference(reference dockerHubMatrixSourceLock, lock *dockerHubLock, sourceCount int) error {
	if reference.Path != dockerHubSourceLockPath {
		return fmt.Errorf("matrix source lock path = %q, want %q", reference.Path, dockerHubSourceLockPath)
	}
	if reference.SourceURL != lock.Rest.SourceURL || reference.SHA256 != lock.Rest.SHA256 || reference.CapturedAt != lock.CapturedAt {
		return errors.New("matrix source lock identity does not match the pinned Docker Hub source")
	}
	if reference.SourceOperationCount != sourceCount || reference.SourceOperationCount != lock.Counts.Rest || reference.SourceOperationCount != lock.Counts.Total {
		return fmt.Errorf("source lock operation count = %d, want %d", reference.SourceOperationCount, sourceCount)
	}
	return nil
}

func dockerHubValidateOperation(operation *dockerHubMatrixOperation, source dockerHubSourceInfo, sourceContract map[string]any) error {
	if operation.Protocol != source.Lock.Protocol ||
		operation.Method != source.Lock.Method ||
		operation.Path != source.Lock.Path ||
		operation.OperationID != source.Lock.OperationID ||
		operation.SourceURL != source.SourceURL {
		return fmt.Errorf("source operation %s does not preserve the locked operation identity", source.Lock.ID)
	}
	if operation.SourceLocation == "" || operation.SourceLocation != source.Lock.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve its cited source location", source.Lock.ID)
	}
	if err := dockerHubValidateFacts(operation.Facts, source, sourceContract); err != nil {
		return err
	}

	expected := dockerHubExpectedCells(source, sourceContract)
	if len(operation.Cells) != len(dockerHubLaneOrder) {
		return fmt.Errorf("source operation %s has %d cells, want %d", source.Lock.ID, len(operation.Cells), len(dockerHubLaneOrder))
	}
	seen := make(map[string]struct{}, len(operation.Cells))
	for _, cell := range operation.Cells {
		if _, duplicate := seen[cell.Lane]; duplicate {
			return fmt.Errorf("source operation %s has duplicate %s cell", source.Lock.ID, cell.Lane)
		}
		seen[cell.Lane] = struct{}{}
		want, exists := expected[cell.Lane]
		if !exists {
			return fmt.Errorf("source operation %s has unknown lane %q", source.Lock.ID, cell.Lane)
		}
		if cell.State != want.State || cell.Reason != want.Reason {
			if cell.Lane == "etl" && dockerHubRequiresETL(source, sourceContract) {
				return fmt.Errorf("pageable or extractable source operation %s has %s cell = %s/%s, want %s/%s", source.Lock.ID, cell.Lane, cell.State, cell.Reason, want.State, want.Reason)
			}
			if (cell.Lane == "direct_write" || cell.Lane == "reverse_etl") && dockerHubIsMutation(source.Lock.Method) {
				return fmt.Errorf("mutation source operation %s has %s cell = %s/%s, want %s/%s", source.Lock.ID, cell.Lane, cell.State, cell.Reason, want.State, want.Reason)
			}
			return fmt.Errorf("source operation %s has %s cell = %s/%s, want %s/%s", source.Lock.ID, cell.Lane, cell.State, cell.Reason, want.State, want.Reason)
		}
		if cell.SourceLocation != source.Lock.SourceLocation || cell.SourceURL != source.SourceURL {
			return fmt.Errorf("source operation %s has uncited %s cell", source.Lock.ID, cell.Lane)
		}
		if cell.State != "mapped_unproven" && cell.State != "not_applicable" {
			return fmt.Errorf("source operation %s has unsupported mapping-only state %q", source.Lock.ID, cell.State)
		}
	}
	for _, lane := range dockerHubLaneOrder {
		if _, exists := seen[lane]; !exists {
			return fmt.Errorf("source operation %s has no %s cell", source.Lock.ID, lane)
		}
	}
	return nil
}

func dockerHubValidateFacts(facts dockerHubMatrixFacts, source dockerHubSourceInfo, sourceContract map[string]any) error {
	classification := "head_probe"
	switch {
	case source.Lock.Method == "GET":
		classification = "read"
	case dockerHubIsMutation(source.Lock.Method):
		classification = "mutation"
	}
	if facts.Classification != classification {
		return fmt.Errorf("source operation %s classification = %q, want %q", source.Lock.ID, facts.Classification, classification)
	}

	wantPagination := dockerHubSourcePagination(source)
	if facts.Pagination.Kind != wantPagination.Kind ||
		!dockerHubEqualStrings(facts.Pagination.QueryParameters, wantPagination.QueryParameters) ||
		facts.Pagination.Description != wantPagination.Description ||
		facts.Pagination.SourceLocation != source.Lock.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact pagination evidence", source.Lock.ID)
	}

	wantScope := dockerHubSourceScope(source)
	if !dockerHubEqualStrings(facts.Scope.PathParameters, wantScope.PathParameters) ||
		!dockerHubEqualStrings(facts.Scope.RequiredPathParameters, wantScope.RequiredPathParameters) ||
		!dockerHubEqualStrings(facts.Scope.QueryParameters, wantScope.QueryParameters) ||
		!dockerHubEqualStrings(facts.Scope.RequiredQueryParameters, wantScope.RequiredQueryParameters) ||
		facts.Scope.SourceLocation != source.Lock.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact scope/path evidence", source.Lock.ID)
	}

	requestMedia, responseMedia := dockerHubSourceMedia(source)
	if !dockerHubEqualStrings(facts.Media.Request, requestMedia) ||
		!dockerHubEqualStrings(facts.Media.Response, responseMedia) ||
		facts.Media.SourceLocation != source.Lock.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact media evidence", source.Lock.ID)
	}

	callbackCount := dockerHubDocumentedCallbackCount(source)
	if facts.EventCursor.Kind != "not_documented" ||
		facts.EventCursor.Parameter != "" ||
		facts.EventCursor.Description != "" ||
		facts.EventCursor.DocumentedCallbacks != callbackCount ||
		facts.EventCursor.SourceLocation != source.Lock.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve its event/cursor evidence", source.Lock.ID)
	}

	wantExtractability := dockerHubSourceExtractability(source, sourceContract)
	if facts.Extractability.Kind != wantExtractability.Kind ||
		!dockerHubEqualStrings(facts.Extractability.ResponseStatuses, wantExtractability.ResponseStatuses) ||
		facts.Extractability.SourceLocation != source.Lock.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact collection-response evidence: got %s/%v/%q want %s/%v/%q",
			source.Lock.ID,
			facts.Extractability.Kind,
			facts.Extractability.ResponseStatuses,
			facts.Extractability.SourceLocation,
			wantExtractability.Kind,
			wantExtractability.ResponseStatuses,
			source.Lock.SourceLocation,
		)
	}
	return nil
}

func dockerHubSourcePagination(source dockerHubSourceInfo) dockerHubPaginationFact {
	if source.Lock.Method != "GET" {
		return dockerHubPaginationFact{
			Kind:            "not_documented",
			QueryParameters: []string{},
			SourceLocation:  source.Lock.SourceLocation,
		}
	}
	queryNames := dockerHubParameterNames(source.Crosswalk.Request.QueryParameters, false)
	kind, parameters := "not_documented", []string{}
	switch {
	case dockerHubContainsString(queryNames, "page") && dockerHubContainsString(queryNames, "page_size"):
		kind, parameters = "page_number", []string{"page", "page_size"}
	case dockerHubContainsString(queryNames, "startIndex") && dockerHubContainsString(queryNames, "count"):
		kind, parameters = "start_index", []string{"count", "startIndex"}
	}
	description := ""
	if kind != "not_documented" {
		description = dockerHubStringValue(source.Lock.Operation["description"])
	}
	return dockerHubPaginationFact{
		Kind:            kind,
		QueryParameters: parameters,
		Description:     description,
		SourceLocation:  source.Lock.SourceLocation,
	}
}

func dockerHubSourceScope(source dockerHubSourceInfo) dockerHubScopeFact {
	return dockerHubScopeFact{
		PathParameters:          dockerHubParameterNames(source.Crosswalk.Request.PathParameters, false),
		RequiredPathParameters:  dockerHubParameterNames(source.Crosswalk.Request.PathParameters, true),
		QueryParameters:         dockerHubParameterNames(source.Crosswalk.Request.QueryParameters, false),
		RequiredQueryParameters: dockerHubParameterNames(source.Crosswalk.Request.QueryParameters, true),
		SourceLocation:          source.Lock.SourceLocation,
	}
}

func dockerHubParameterNames(parameters []dockerHubParameter, requiredOnly bool) []string {
	names := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		if requiredOnly && !parameter.Required {
			continue
		}
		names = append(names, parameter.Name)
	}
	return dockerHubSortedStrings(names)
}

func dockerHubSourceMedia(source dockerHubSourceInfo) ([]string, []string) {
	return dockerHubSortedStrings(source.Crosswalk.Request.BodyContentTypes), dockerHubSortedStrings(source.Crosswalk.Response.ContentTypes)
}

func dockerHubDocumentedCallbackCount(source dockerHubSourceInfo) int {
	callbacks, ok := source.Lock.Operation["callbacks"].(map[string]any)
	if !ok {
		return 0
	}
	return len(callbacks)
}

func dockerHubSourceExtractability(source dockerHubSourceInfo, sourceContract map[string]any) dockerHubExtractabilityFact {
	statuses := make([]string, 0)
	responses, ok := source.Lock.Operation["responses"].(map[string]any)
	if ok {
		for status, response := range responses {
			if dockerHubResponseHasTopLevelArraySchema(response, sourceContract, map[string]bool{}) {
				statuses = append(statuses, status)
			}
		}
	}
	statuses = dockerHubSortedStrings(statuses)
	kind := "not_documented"
	if len(statuses) > 0 {
		kind = "array_response"
	}
	return dockerHubExtractabilityFact{
		Kind:             kind,
		ResponseStatuses: statuses,
		SourceLocation:   source.Lock.SourceLocation,
	}
}

func dockerHubResponseHasTopLevelArraySchema(value any, sourceContract map[string]any, seenReferences map[string]bool) bool {
	response := dockerHubDereferenceObject(value, sourceContract, seenReferences)
	content, ok := response["content"].(map[string]any)
	if !ok {
		return false
	}
	for _, media := range content {
		mediaObject, ok := media.(map[string]any)
		if !ok {
			continue
		}
		if dockerHubSchemaHasTopLevelArray(mediaObject["schema"], sourceContract, map[string]bool{}) {
			return true
		}
	}
	return false
}

func dockerHubSchemaHasTopLevelArray(value any, sourceContract map[string]any, seenReferences map[string]bool) bool {
	schema := dockerHubDereferenceObject(value, sourceContract, seenReferences)
	if dockerHubStringValue(schema["type"]) == "array" {
		return true
	}
	for _, composition := range []string{"allOf", "anyOf", "oneOf"} {
		branches, ok := schema[composition].([]any)
		if !ok {
			continue
		}
		for _, branch := range branches {
			if dockerHubSchemaHasTopLevelArray(branch, sourceContract, seenReferences) {
				return true
			}
		}
	}
	return false
}

func dockerHubDereferenceObject(value any, sourceContract map[string]any, seenReferences map[string]bool) map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	reference := dockerHubStringValue(object["$ref"])
	if reference == "" || seenReferences[reference] {
		return object
	}
	next, found := dockerHubJSONPointer(sourceContract, reference)
	if !found {
		return object
	}
	nextSeen := make(map[string]bool, len(seenReferences)+1)
	for seen := range seenReferences {
		nextSeen[seen] = true
	}
	nextSeen[reference] = true
	return dockerHubDereferenceObject(next, sourceContract, nextSeen)
}

func dockerHubJSONPointer(root map[string]any, reference string) (any, bool) {
	if !strings.HasPrefix(reference, "#/") {
		return nil, false
	}
	var current any = root
	for _, rawSegment := range strings.Split(strings.TrimPrefix(reference, "#/"), "/") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		segment := strings.ReplaceAll(strings.ReplaceAll(rawSegment, "~1", "/"), "~0", "~")
		next, ok := object[segment]
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

func dockerHubExpectedCells(source dockerHubSourceInfo, sourceContract map[string]any) map[string]dockerHubExpectedCell {
	isMutation := dockerHubIsMutation(source.Lock.Method)
	isRead := source.Lock.Method == "GET"
	isBinaryDownload := dockerHubHasBinaryDownload(source)
	isBinaryUpload := dockerHubHasBinaryUpload(source)
	requiresETL := dockerHubRequiresETL(source, sourceContract)

	expected := make(map[string]dockerHubExpectedCell, len(dockerHubLaneOrder))
	if isRead {
		expected["direct_read"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.direct_read.documented_get_response.v1"}
	} else {
		expected["direct_read"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.direct_read.non_get_not_applicable.v1"}
	}
	if isMutation {
		expected["direct_write"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.direct_write.mutation_verb.v1"}
		expected["reverse_etl"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.reverse_etl.mutation_verb.v1"}
	} else {
		expected["direct_write"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.direct_write.non_mutation_not_applicable.v1"}
		expected["reverse_etl"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.reverse_etl.non_mutation_not_applicable.v1"}
	}
	if isBinaryDownload {
		expected["binary_download"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.binary_download.binary_response_media.v1"}
	} else {
		expected["binary_download"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.binary_download.no_binary_response_media.v1"}
	}
	if isBinaryUpload {
		expected["binary_upload"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.binary_upload.binary_request_media.v1"}
	} else {
		expected["binary_upload"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.binary_upload.no_binary_request_media.v1"}
	}
	if requiresETL {
		expected["etl"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.etl.pageable_or_extractable_collection_read.v1"}
		expected["sync_transport"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.sync_transport.pageable_or_extractable_collection_read.v1"}
	} else {
		expected["etl"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.etl.no_pageable_or_extractable_collection_read.v1"}
		expected["sync_transport"] = dockerHubExpectedCell{"not_applicable", "dockerhub.source.sync_transport.no_pageable_or_extractable_collection_read.v1"}
	}
	return expected
}

type dockerHubExpectedCell struct {
	State  string
	Reason string
}

func dockerHubIsMutation(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func dockerHubRequiresETL(source dockerHubSourceInfo, sourceContract map[string]any) bool {
	if source.Lock.Method != "GET" {
		return false
	}
	return dockerHubSourcePagination(source).Kind != "not_documented" ||
		dockerHubSourceExtractability(source, sourceContract).Kind == "array_response"
}

func dockerHubHasBinaryDownload(source dockerHubSourceInfo) bool {
	_, responseMedia := dockerHubSourceMedia(source)
	return dockerHubHasBinaryMedia(responseMedia, false)
}

func dockerHubHasBinaryUpload(source dockerHubSourceInfo) bool {
	requestMedia, _ := dockerHubSourceMedia(source)
	return dockerHubHasBinaryMedia(requestMedia, true)
}

func dockerHubHasBinaryMedia(media []string, includeMultipart bool) bool {
	for _, contentType := range media {
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		switch contentType {
		case "application/octet-stream", "application/pdf", "application/zip", "application/gzip", "application/x-gzip", "application/x-tar":
			return true
		case "multipart/form-data":
			if includeMultipart {
				return true
			}
		}
		if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "audio/") || strings.HasPrefix(contentType, "video/") {
			return true
		}
	}
	return false
}

func dockerHubValidateCountReconciliation(matrix *dockerHubMatrix, lock *dockerHubLock, sources map[string]dockerHubSourceInfo) error {
	if len(matrix.Operations) != 54 || len(sources) != 54 || lock.Counts.Rest != 54 || lock.Counts.Total != 54 {
		return fmt.Errorf("retained source count reconciliation = matrix:%d lock:%d rest:%d total:%d, want 54", len(matrix.Operations), len(sources), lock.Counts.Rest, lock.Counts.Total)
	}

	var getCount, mutationCount, deleteCount, headCount, pagingCount, arrayReadCount, etlCount, scimCount, csvCount, callbackCount int
	for _, source := range sources {
		if source.Lock.Method == "GET" {
			getCount++
		}
		if dockerHubIsMutation(source.Lock.Method) {
			mutationCount++
		}
		if source.Lock.Method == "DELETE" {
			deleteCount++
		}
		if source.Lock.Method == "HEAD" {
			headCount++
		}
		pagination := dockerHubSourcePagination(source)
		extractability := dockerHubSourceExtractability(source, lock.SourceContract)
		if pagination.Kind != "not_documented" {
			pagingCount++
		}
		if source.Lock.Method == "GET" && extractability.Kind == "array_response" {
			arrayReadCount++
		}
		if dockerHubRequiresETL(source, lock.SourceContract) {
			etlCount++
		}
		if strings.Contains(source.Lock.Path, "/scim/2.0/") {
			scimCount++
		}
		_, responseMedia := dockerHubSourceMedia(source)
		if dockerHubContainsString(responseMedia, "text/csv") {
			csvCount++
		}
		callbackCount += dockerHubDocumentedCallbackCount(source)
	}
	if getCount != 24 || mutationCount != 27 || deleteCount != 6 || headCount != 3 ||
		pagingCount != 9 || arrayReadCount != 2 || etlCount != 10 || scimCount != 9 ||
		csvCount != 1 || callbackCount != 0 {
		return fmt.Errorf("source fact counts = GET:%d mutation:%d delete:%d head:%d paging:%d array_read:%d etl:%d scim:%d csv:%d callbacks:%d",
			getCount, mutationCount, deleteCount, headCount, pagingCount, arrayReadCount, etlCount, scimCount, csvCount, callbackCount)
	}

	cells := dockerHubMatrixCellCounts(matrix)
	wantMapped := map[string]int{
		"direct_read":     24,
		"direct_write":    27,
		"binary_download": 0,
		"binary_upload":   0,
		"etl":             10,
		"reverse_etl":     27,
		"sync_transport":  10,
	}
	mappedTotal, notApplicableTotal := 0, 0
	for _, lane := range dockerHubLaneOrder {
		if cells[lane]["mapped_unproven"] != wantMapped[lane] {
			return fmt.Errorf("%s mapped_unproven count = %d, want %d", lane, cells[lane]["mapped_unproven"], wantMapped[lane])
		}
		if cells[lane]["not_applicable"] != 54-wantMapped[lane] {
			return fmt.Errorf("%s not_applicable count = %d, want %d", lane, cells[lane]["not_applicable"], 54-wantMapped[lane])
		}
		if cells[lane]["implemented"] != 0 || cells[lane]["missing_foundation"] != 0 {
			return fmt.Errorf("%s contains an execution or foundation state", lane)
		}
		mappedTotal += cells[lane]["mapped_unproven"]
		notApplicableTotal += cells[lane]["not_applicable"]
	}
	if mappedTotal != 98 || notApplicableTotal != 280 || mappedTotal+notApplicableTotal != 378 {
		return fmt.Errorf("matrix cell totals = mapped:%d not_applicable:%d total:%d, want 98/280/378", mappedTotal, notApplicableTotal, mappedTotal+notApplicableTotal)
	}
	return nil
}

func dockerHubMatrixCellCounts(matrix *dockerHubMatrix) map[string]map[string]int {
	counts := make(map[string]map[string]int, len(dockerHubLaneOrder))
	for _, lane := range dockerHubLaneOrder {
		counts[lane] = make(map[string]int)
	}
	for _, operation := range matrix.Operations {
		for _, cell := range operation.Cells {
			counts[cell.Lane][cell.State]++
		}
	}
	return counts
}

func readDockerHubArtifactRecords(t *testing.T) dockerHubArtifactRecords {
	t.Helper()
	var api dockerHubAPISurfaceDocument
	var streams dockerHubStreamsDocument
	var operations dockerHubOperationsDocument
	var cli dockerHubCLISurfaceDocument
	readDockerHubJSON(t, "api_surface.json", &api)
	readDockerHubJSON(t, "streams.json", &streams)
	readDockerHubJSON(t, "operations.json", &operations)
	readDockerHubJSON(t, "cli_surface.json", &cli)

	records := dockerHubArtifactRecords{
		APIRoutes:      make(map[string]struct{}, len(api.Endpoints)),
		Streams:        make(map[string]dockerHubStreamRecord, len(streams.Streams)),
		Commands:       make(map[string]dockerHubCommandRecord, len(cli.Commands)),
		OperationCount: len(operations.Operations),
	}
	for _, endpoint := range api.Endpoints {
		record := dockerHubRouteRecord(endpoint.Method, endpoint.Path)
		if _, exists := records.APIRoutes[record]; exists {
			t.Fatalf("api_surface.json duplicates endpoint %q", record)
		}
		records.APIRoutes[record] = struct{}{}
	}
	for _, stream := range streams.Streams {
		if _, exists := records.Streams[stream.Name]; exists {
			t.Fatalf("streams.json duplicates stream %q", stream.Name)
		}
		records.Streams[stream.Name] = stream
	}
	for _, command := range cli.Commands {
		if _, exists := records.Commands[command.Path]; exists {
			t.Fatalf("cli_surface.json duplicates command %q", command.Path)
		}
		records.Commands[command.Path] = command
	}
	return records
}

func dockerHubValidateArtifactLinks(matrix *dockerHubMatrix, records dockerHubArtifactRecords, matrixByID map[string]*dockerHubMatrixOperation, sources map[string]dockerHubSourceInfo) error {
	artifacts := make(map[string]*dockerHubMatrixArtifact, len(matrix.Artifacts))
	for index := range matrix.Artifacts {
		artifact := &matrix.Artifacts[index]
		if _, exists := artifacts[artifact.Path]; exists {
			return fmt.Errorf("duplicate artifact matrix %q", artifact.Path)
		}
		artifacts[artifact.Path] = artifact
	}
	if len(artifacts) != 4 {
		return fmt.Errorf("matrix artifact count = %d, want 4", len(artifacts))
	}
	if err := dockerHubValidateAPISurfaceLinks(artifacts["api_surface.json"], records, matrixByID, sources); err != nil {
		return err
	}
	if err := dockerHubValidateStreamLinks(artifacts["streams.json"], records, matrixByID); err != nil {
		return err
	}
	if err := dockerHubValidateOperationLinks(artifacts["operations.json"], records); err != nil {
		return err
	}
	if err := dockerHubValidateCommandLinks(artifacts["cli_surface.json"], records, matrixByID); err != nil {
		return err
	}
	if len(matrix.BacklinkGaps) != 0 {
		return fmt.Errorf("Docker Hub has %d unexpected backlink gaps despite exact source-to-artifact coverage", len(matrix.BacklinkGaps))
	}
	return nil
}

func dockerHubValidateAPISurfaceLinks(artifact *dockerHubMatrixArtifact, records dockerHubArtifactRecords, matrixByID map[string]*dockerHubMatrixOperation, sources map[string]dockerHubSourceInfo) error {
	if artifact == nil {
		return errors.New("matrix is missing api_surface.json artifact")
	}
	if artifact.RecordCount != len(records.APIRoutes) || len(artifact.Links) != len(records.APIRoutes) {
		return fmt.Errorf("artifact api_surface.json link count = records:%d links:%d, want %d", artifact.RecordCount, len(artifact.Links), len(records.APIRoutes))
	}
	seen := make(map[string]struct{}, len(artifact.Links))
	for _, link := range artifact.Links {
		if _, exists := records.APIRoutes[link.Record]; !exists {
			return fmt.Errorf("artifact api_surface.json link references nonexistent record %q", link.Record)
		}
		if _, duplicate := seen[link.Record]; duplicate {
			return fmt.Errorf("artifact api_surface.json link duplicates record %q", link.Record)
		}
		seen[link.Record] = struct{}{}
		operation, exists := matrixByID[link.SourceID]
		if !exists {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent source cell owner %q", link.Record, link.SourceID)
		}
		source, exists := sources[link.SourceID]
		if !exists {
			return fmt.Errorf("artifact api_surface.json link %q creates source id %q outside the source lock", link.Record, link.SourceID)
		}
		if link.Record != dockerHubRouteRecord(source.Lock.Method, source.Lock.Path) ||
			link.Record != dockerHubRouteRecord(operation.Method, operation.Path) {
			return fmt.Errorf("artifact api_surface.json link %q does not preserve source route", link.Record)
		}
		lane := "direct_read"
		if dockerHubIsMutation(source.Lock.Method) {
			lane = "direct_write"
		}
		if !dockerHubEqualStrings(link.Lanes, []string{lane}) || dockerHubMatrixCellByLaneNoTest(operation, lane) == nil {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent cell %q", link.Record, lane)
		}
	}
	return nil
}

func dockerHubValidateStreamLinks(artifact *dockerHubMatrixArtifact, records dockerHubArtifactRecords, matrixByID map[string]*dockerHubMatrixOperation) error {
	expected := map[string]struct {
		SourceID string
		Lanes    []string
	}{
		"repositories": {
			SourceID: "dockerhub.rest.listNamespaceRepositories",
			Lanes:    []string{"direct_read", "etl", "sync_transport"},
		},
		"tags": {
			SourceID: "dockerhub.rest.ListRepositoryTags",
			Lanes:    []string{"direct_read", "etl", "sync_transport"},
		},
		"repository_detail": {
			SourceID: "dockerhub.rest.GetRepository",
			Lanes:    []string{"direct_read"},
		},
		"tag_detail": {
			SourceID: "dockerhub.rest.GetRepositoryTag",
			Lanes:    []string{"direct_read"},
		},
	}
	if artifact == nil {
		return errors.New("matrix is missing streams.json artifact")
	}
	if artifact.RecordCount != len(records.Streams) || len(artifact.Links) != len(expected) || len(artifact.Links) != len(records.Streams) {
		return errors.New("artifact streams.json link count does not reconcile")
	}
	seen := make(map[string]struct{}, len(artifact.Links))
	for _, link := range artifact.Links {
		if _, exists := records.Streams[link.Record]; !exists {
			return fmt.Errorf("artifact streams.json link references nonexistent stream %q", link.Record)
		}
		if _, duplicate := seen[link.Record]; duplicate {
			return fmt.Errorf("artifact streams.json link duplicates stream %q", link.Record)
		}
		seen[link.Record] = struct{}{}
		want, exists := expected[link.Record]
		if !exists || link.SourceID != want.SourceID {
			return fmt.Errorf("artifact streams.json link %q has incorrect source id", link.Record)
		}
		operation := matrixByID[link.SourceID]
		if operation == nil {
			return fmt.Errorf("artifact streams.json link %q references nonexistent source cell owner", link.Record)
		}
		if !dockerHubEqualStrings(link.Lanes, want.Lanes) {
			return fmt.Errorf("artifact streams.json link %q has incorrect lanes", link.Record)
		}
		for _, lane := range link.Lanes {
			if dockerHubMatrixCellByLaneNoTest(operation, lane) == nil {
				return fmt.Errorf("artifact streams.json link %q references nonexistent cell %q", link.Record, lane)
			}
		}
	}
	return nil
}

func dockerHubValidateOperationLinks(artifact *dockerHubMatrixArtifact, records dockerHubArtifactRecords) error {
	if artifact == nil {
		return errors.New("matrix is missing operations.json artifact")
	}
	if artifact.RecordCount != records.OperationCount || artifact.RecordCount != 0 || len(artifact.Links) != 0 {
		return errors.New("artifact operations.json link count does not reconcile")
	}
	return nil
}

func dockerHubValidateCommandLinks(artifact *dockerHubMatrixArtifact, records dockerHubArtifactRecords, matrixByID map[string]*dockerHubMatrixOperation) error {
	sourceByStream := map[string]string{
		"repositories":      "dockerhub.rest.listNamespaceRepositories",
		"tags":              "dockerhub.rest.ListRepositoryTags",
		"repository_detail": "dockerhub.rest.GetRepository",
		"tag_detail":        "dockerhub.rest.GetRepositoryTag",
	}
	if artifact == nil {
		return errors.New("matrix is missing cli_surface.json artifact")
	}
	if artifact.RecordCount != len(records.Commands) || len(artifact.Links) != len(records.Commands) {
		return errors.New("artifact cli_surface.json link count does not reconcile")
	}
	seen := make(map[string]struct{}, len(artifact.Links))
	for _, link := range artifact.Links {
		command, exists := records.Commands[link.Record]
		if !exists {
			return fmt.Errorf("artifact cli_surface.json link references nonexistent command %q", link.Record)
		}
		if _, duplicate := seen[link.Record]; duplicate {
			return fmt.Errorf("artifact cli_surface.json link duplicates command %q", link.Record)
		}
		seen[link.Record] = struct{}{}
		if command.Intent != "etl" || link.SourceID != sourceByStream[command.Stream] {
			return fmt.Errorf("artifact cli_surface.json link %q does not preserve its source-backed stream route", link.Record)
		}
		operation := matrixByID[link.SourceID]
		if operation == nil {
			return fmt.Errorf("artifact cli_surface.json link %q references nonexistent source cell owner", link.Record)
		}
		if !dockerHubEqualStrings(link.Lanes, []string{"etl"}) || dockerHubMatrixCellByLaneNoTest(operation, "etl") == nil {
			return fmt.Errorf("artifact cli_surface.json link %q references nonexistent cell %q", link.Record, "etl")
		}
	}
	return nil
}

func dockerHubMatrixArtifactByPath(t *testing.T, matrix *dockerHubMatrix, path string) *dockerHubMatrixArtifact {
	t.Helper()
	for index := range matrix.Artifacts {
		if matrix.Artifacts[index].Path == path {
			return &matrix.Artifacts[index]
		}
	}
	t.Fatalf("matrix has no %s artifact", path)
	return nil
}

func dockerHubMatrixOperationByID(t *testing.T, matrix *dockerHubMatrix, sourceID string) *dockerHubMatrixOperation {
	t.Helper()
	for index := range matrix.Operations {
		if matrix.Operations[index].SourceID == sourceID {
			return &matrix.Operations[index]
		}
	}
	t.Fatalf("matrix has no %s source operation", sourceID)
	return nil
}

func dockerHubMatrixCellByLane(t *testing.T, operation *dockerHubMatrixOperation, lane string) *dockerHubMatrixCell {
	t.Helper()
	cell := dockerHubMatrixCellByLaneNoTest(operation, lane)
	if cell == nil {
		t.Fatalf("source operation %s has no %s cell", operation.SourceID, lane)
	}
	return cell
}

func dockerHubMatrixCellByLaneNoTest(operation *dockerHubMatrixOperation, lane string) *dockerHubMatrixCell {
	for index := range operation.Cells {
		if operation.Cells[index].Lane == lane {
			return &operation.Cells[index]
		}
	}
	return nil
}

func dockerHubRouteRecord(method, path string) string {
	return method + " " + path
}

func dockerHubSortedSourceIDs(sources map[string]dockerHubSourceInfo) []string {
	ids := make([]string, 0, len(sources))
	for sourceID := range sources {
		ids = append(ids, sourceID)
	}
	return dockerHubSortedStrings(ids)
}

func dockerHubSortedStrings(values []string) []string {
	sorted := append([]string{}, values...)
	sort.Strings(sorted)
	return sorted
}

func dockerHubEqualStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func dockerHubContainsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func dockerHubStringValue(value any) string {
	stringValue, _ := value.(string)
	return stringValue
}
