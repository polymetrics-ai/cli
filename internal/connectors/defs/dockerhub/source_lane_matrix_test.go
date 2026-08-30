package dockerhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	dockerHubMatrixPath          = "sources/dockerhub-source-lane-matrix.json"
	dockerHubSourceLockPath      = "sources/dockerhub-operation-source-lock.json"
	dockerHubCrosswalkPath       = "sources/dockerhub-operation-crosswalk.json"
	dockerHubDispositionPath     = "sources/dockerhub-declaration-disposition.json"
	dockerHubReverseETLAuditPath = "sources/dockerhub-reverse-etl-action-audit.json"
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

// dockerHubDisposition is deliberately a mapping-truth record. It does not
// declare runtime reachability: that belongs to the actual definition artifacts
// and their separate execution proof.
type dockerHubDisposition struct {
	SchemaVersion              int                             `json:"schema_version"`
	Connector                  string                          `json:"connector"`
	SourceBasis                dockerHubDispositionSourceBasis `json:"source_basis"`
	Mapping                    dockerHubDispositionMapping     `json:"mapping"`
	SharedExecutorCapabilities []dockerHubExecutorCapability   `json:"shared_executor_capabilities"`
	Notes                      []string                        `json:"notes"`
}

type dockerHubDispositionSourceBasis struct {
	SourceLock           string `json:"source_lock"`
	Crosswalk            string `json:"crosswalk"`
	SourceLaneMatrix     string `json:"source_lane_matrix"`
	SourceOperationCount int    `json:"source_operation_count"`
}

type dockerHubDispositionMapping struct {
	State            string                                    `json:"state"`
	LaneCells        map[string]dockerHubDispositionLaneCounts `json:"lane_cells"`
	RuntimeArtifacts map[string]dockerHubRuntimeArtifact       `json:"runtime_artifacts"`
}

type dockerHubDispositionLaneCounts struct {
	MappedUnproven int `json:"mapped_unproven"`
	NotApplicable  int `json:"not_applicable"`
}

type dockerHubRuntimeArtifact struct {
	Path        string `json:"path"`
	Present     bool   `json:"present"`
	RecordCount int    `json:"record_count"`
	Role        string `json:"role"`
}

type dockerHubExecutorCapability struct {
	Kind     string `json:"kind"`
	State    string `json:"state"`
	Evidence string `json:"evidence"`
}

// dockerHubReverseETLAudit lists source-backed candidates only. A candidate is
// never an invented write action or destination transport declaration.
type dockerHubReverseETLAudit struct {
	SchemaVersion    int                             `json:"schema_version"`
	Connector        string                          `json:"connector"`
	Purpose          string                          `json:"purpose"`
	SourceLock       string                          `json:"source_lock"`
	SourceLaneMatrix string                          `json:"source_lane_matrix"`
	Summary          dockerHubReverseETLAuditSummary `json:"summary"`
	WriteOperations  []dockerHubReverseETLCandidate  `json:"write_operations"`
}

type dockerHubReverseETLAuditSummary struct {
	SourceMutationOperations    int `json:"source_mutation_operations"`
	SourceDeleteOperations      int `json:"source_delete_operations"`
	MappedDirectWriteCells      int `json:"mapped_direct_write_cells"`
	MappedReverseETLCells       int `json:"mapped_reverse_etl_cells"`
	DeclaredWriteActions        int `json:"declared_write_actions"`
	DeclaredDestinationBindings int `json:"declared_destination_bindings"`
	ExecutableReverseETLActions int `json:"executable_reverse_etl_actions"`
}

type dockerHubReverseETLCandidate struct {
	SourceID       string   `json:"source_id"`
	Method         string   `json:"method"`
	Path           string   `json:"path"`
	SourceURL      string   `json:"source_url"`
	SourceLocation string   `json:"source_location"`
	Lanes          []string `json:"lanes"`
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
	APIRoutes                    map[string]struct{}
	Streams                      map[string]dockerHubStreamRecord
	Commands                     map[string]dockerHubCommandRecord
	OperationCount               int
	WritesArtifactPresent        bool
	SyncTransportArtifactPresent bool
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
	artifacts := readDockerHubArtifactRecords(t)
	if err := validateDockerHubSourceLaneMatrix(matrix, lock, crosswalk, artifacts); err != nil {
		t.Fatal(err)
	}
	if err := validateDockerHubDisposition(readDockerHubDisposition(t), matrix, lock, crosswalk, artifacts); err != nil {
		t.Fatal(err)
	}
	if err := validateDockerHubReverseETLAudit(readDockerHubReverseETLAudit(t), matrix, lock, crosswalk, artifacts); err != nil {
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
	if cell := dockerHubMatrixCellByLane(t, export, "binary_download"); cell.State != "mapped_unproven" || cell.Reason != "dockerhub.source.binary_download.text_export_response_media.v1" {
		t.Fatalf("text/csv member export binary_download = %s/%s, want source-evidenced mapped_unproven/text_export", cell.State, cell.Reason)
	}

	for _, sourceID := range []string{
		"dockerhub.rest.CheckRepository",
		"dockerhub.rest.head_/v2/namespaces/{namespace}/repositories/{repository}/tags",
		"dockerhub.rest.head_/v2/namespaces/{namespace}/repositories/{repository}/tags/{tag}",
	} {
		head := dockerHubMatrixOperationByID(t, matrix, sourceID)
		if head.Method != "HEAD" {
			t.Fatalf("source operation %s method = %q, want HEAD", sourceID, head.Method)
		}
		if cell := dockerHubMatrixCellByLane(t, head, "direct_read"); cell.State != "mapped_unproven" || cell.Reason != "dockerhub.source.direct_read.documented_head_status.v1" {
			t.Fatalf("HEAD source operation %s direct_read = %s/%s, want source-evidenced mapped_unproven/documented_head_status", sourceID, cell.State, cell.Reason)
		}
	}
}

func TestDockerHubSourceLaneMatrixRejectsArtifactBacklinkToUnmappedLane(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	sources, err := dockerHubSourceInfos(lock, crosswalk)
	if err != nil {
		t.Fatal(err)
	}
	operation := dockerHubMatrixOperationByID(t, matrix, "dockerhub.rest.CheckRepository")
	dockerHubMatrixCellByLane(t, operation, "direct_read").State = "not_applicable"
	if err := dockerHubValidateArtifactLinks(matrix, readDockerHubArtifactRecords(t), dockerHubMatrixByID(matrix), sources); err == nil || !strings.Contains(err.Error(), "does not point to a mapped lane") {
		t.Fatalf("artifact backlink validation error = %v, want mapped-lane rejection", err)
	}
}

func TestDockerHubDispositionRejectsStaleRuntimeClaims(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	artifacts := readDockerHubArtifactRecords(t)
	tests := []struct {
		name string
		edit func(*dockerHubDisposition)
		want string
	}{
		{
			name: "invented operation declarations",
			edit: func(disposition *dockerHubDisposition) {
				disposition.Mapping.RuntimeArtifacts["operations"] = dockerHubRuntimeArtifact{
					Path: "operations.json", Present: true, RecordCount: 45, Role: "declared_operations",
				}
			},
			want: "runtime artifact operations",
		},
		{
			name: "invented write actions",
			edit: func(disposition *dockerHubDisposition) {
				disposition.Mapping.RuntimeArtifacts["writes"] = dockerHubRuntimeArtifact{
					Path: "writes.json", Present: true, RecordCount: 20, Role: "declared_write_actions",
				}
			},
			want: "runtime artifact writes",
		},
		{
			name: "invented sync transport",
			edit: func(disposition *dockerHubDisposition) {
				disposition.Mapping.RuntimeArtifacts["sync_transport"] = dockerHubRuntimeArtifact{
					Path: "sync_transport.json", Present: true, RecordCount: 4, Role: "declared_sync_transport",
				}
			},
			want: "runtime artifact sync_transport",
		},
		{
			name: "misreports existing status capability as a foundation gap",
			edit: func(disposition *dockerHubDisposition) {
				for index := range disposition.SharedExecutorCapabilities {
					if disposition.SharedExecutorCapabilities[index].Kind == "rest_status" {
						disposition.SharedExecutorCapabilities[index].State = "missing_foundation"
					}
				}
			},
			want: "shared executor capability rest_status",
		},
		{
			name: "misreports existing text export capability as a foundation gap",
			edit: func(disposition *dockerHubDisposition) {
				for index := range disposition.SharedExecutorCapabilities {
					if disposition.SharedExecutorCapabilities[index].Kind == "text_export" {
						disposition.SharedExecutorCapabilities[index].State = "missing_foundation"
					}
				}
			},
			want: "shared executor capability text_export",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			disposition := readDockerHubDisposition(t)
			test.edit(disposition)
			err := validateDockerHubDisposition(disposition, matrix, lock, crosswalk, artifacts)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("disposition validation error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestDockerHubReverseETLAuditRejectsSourceOrLaneDrift(t *testing.T) {
	matrix, lock, crosswalk := readDockerHubMatrixInputs(t)
	artifacts := readDockerHubArtifactRecords(t)
	tests := []struct {
		name string
		edit func(*dockerHubReverseETLAudit)
		want string
	}{
		{
			name: "source path mismatch",
			edit: func(audit *dockerHubReverseETLAudit) {
				audit.WriteOperations[0].Path = "/not-the-pinned-source-path"
			},
			want: "does not preserve locked source facts",
		},
		{
			name: "non reverse etl backlink",
			edit: func(audit *dockerHubReverseETLAudit) {
				audit.WriteOperations[0].Lanes = []string{"direct_read"}
			},
			want: "does not preserve mapped direct-write/reverse-etl lanes",
		},
		{
			name: "invented executable action",
			edit: func(audit *dockerHubReverseETLAudit) {
				audit.Summary.ExecutableReverseETLActions = 1
			},
			want: "execution counts",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			audit := readDockerHubReverseETLAudit(t)
			test.edit(audit)
			err := validateDockerHubReverseETLAudit(audit, matrix, lock, crosswalk, artifacts)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("reverse-ETL audit validation error = %v, want substring %q", err, test.want)
			}
		})
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

func readDockerHubDisposition(t *testing.T) *dockerHubDisposition {
	t.Helper()
	disposition := new(dockerHubDisposition)
	readDockerHubStrictJSON(t, dockerHubDispositionPath, disposition)
	return disposition
}

func readDockerHubReverseETLAudit(t *testing.T) *dockerHubReverseETLAudit {
	t.Helper()
	audit := new(dockerHubReverseETLAudit)
	readDockerHubStrictJSON(t, dockerHubReverseETLAuditPath, audit)
	return audit
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

func readDockerHubStrictJSON(t *testing.T, path string, target any) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("strict decode %s: %v", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("strict decode %s has trailing JSON: %v", path, err)
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

func validateDockerHubDisposition(disposition *dockerHubDisposition, matrix *dockerHubMatrix, lock *dockerHubLock, crosswalk *dockerHubCrosswalk, artifacts dockerHubArtifactRecords) error {
	if disposition.SchemaVersion != 1 {
		return fmt.Errorf("disposition schema_version = %d, want 1", disposition.SchemaVersion)
	}
	if disposition.Connector != "dockerhub" {
		return fmt.Errorf("disposition connector = %q, want dockerhub", disposition.Connector)
	}
	if disposition.SourceBasis.SourceLock != dockerHubSourceLockPath ||
		disposition.SourceBasis.Crosswalk != dockerHubCrosswalkPath ||
		disposition.SourceBasis.SourceLaneMatrix != dockerHubMatrixPath {
		return errors.New("disposition does not name the exact Docker Hub source artifacts")
	}
	sources, err := dockerHubSourceInfos(lock, crosswalk)
	if err != nil {
		return err
	}
	if disposition.SourceBasis.SourceOperationCount != len(sources) || disposition.SourceBasis.SourceOperationCount != len(matrix.Operations) {
		return fmt.Errorf("disposition source operation count = %d, want %d", disposition.SourceBasis.SourceOperationCount, len(sources))
	}
	if disposition.Mapping.State != "source_mapped_runtime_unasserted" {
		return fmt.Errorf("disposition mapping state = %q, want source_mapped_runtime_unasserted", disposition.Mapping.State)
	}

	cells := dockerHubMatrixCellCounts(matrix)
	if len(disposition.Mapping.LaneCells) != len(dockerHubLaneOrder) {
		return fmt.Errorf("disposition lane cell count = %d, want %d", len(disposition.Mapping.LaneCells), len(dockerHubLaneOrder))
	}
	for _, lane := range dockerHubLaneOrder {
		counts, exists := disposition.Mapping.LaneCells[lane]
		if !exists {
			return fmt.Errorf("disposition has no %s lane count", lane)
		}
		if counts.MappedUnproven != cells[lane]["mapped_unproven"] || counts.NotApplicable != cells[lane]["not_applicable"] {
			return fmt.Errorf("disposition %s lane counts = mapped:%d not_applicable:%d, want mapped:%d not_applicable:%d", lane, counts.MappedUnproven, counts.NotApplicable, cells[lane]["mapped_unproven"], cells[lane]["not_applicable"])
		}
	}
	if err := dockerHubValidateRuntimeArtifactSnapshot(disposition.Mapping.RuntimeArtifacts, artifacts); err != nil {
		return err
	}
	if err := dockerHubValidateSharedExecutorCapabilities(disposition.SharedExecutorCapabilities); err != nil {
		return err
	}
	if len(disposition.Notes) == 0 {
		return errors.New("disposition must explain that source mapping is not runtime reachability")
	}
	return nil
}

func dockerHubValidateRuntimeArtifactSnapshot(snapshot map[string]dockerHubRuntimeArtifact, records dockerHubArtifactRecords) error {
	expected := map[string]dockerHubRuntimeArtifact{
		"api_surface": {
			Path:        "api_surface.json",
			Present:     true,
			RecordCount: len(records.APIRoutes),
			Role:        "source_inventory_not_execution_proof",
		},
		"streams": {
			Path:        "streams.json",
			Present:     true,
			RecordCount: len(records.Streams),
			Role:        "source_stream_declarations_not_execution_proof",
		},
		"operations": {
			Path:        "operations.json",
			Present:     true,
			RecordCount: records.OperationCount,
			Role:        "declared_operation_definitions",
		},
		"cli_surface": {
			Path:        "cli_surface.json",
			Present:     true,
			RecordCount: len(records.Commands),
			Role:        "declared_cli_surface_not_execution_proof",
		},
		"writes": {
			Path:        "writes.json",
			Present:     records.WritesArtifactPresent,
			RecordCount: 0,
			Role:        "no_declared_write_actions",
		},
		"sync_transport": {
			Path:        "sync_transport.json",
			Present:     records.SyncTransportArtifactPresent,
			RecordCount: 0,
			Role:        "no_declared_sync_transport",
		},
	}
	if len(snapshot) != len(expected) {
		return fmt.Errorf("runtime artifact snapshot count = %d, want %d", len(snapshot), len(expected))
	}
	for name, want := range expected {
		got, exists := snapshot[name]
		if !exists {
			return fmt.Errorf("runtime artifact %s is absent from the disposition", name)
		}
		if got != want {
			return fmt.Errorf("runtime artifact %s = %+v, want %+v", name, got, want)
		}
	}
	return nil
}

func dockerHubValidateSharedExecutorCapabilities(capabilities []dockerHubExecutorCapability) error {
	expected := map[string]struct{}{
		"rest_status": {},
		"text_export": {},
	}
	if len(capabilities) != len(expected) {
		return fmt.Errorf("shared executor capability count = %d, want %d", len(capabilities), len(expected))
	}
	byKind := make(map[string]dockerHubExecutorCapability, len(capabilities))
	for _, capability := range capabilities {
		if _, duplicate := byKind[capability.Kind]; duplicate {
			return fmt.Errorf("duplicate shared executor capability %q", capability.Kind)
		}
		byKind[capability.Kind] = capability
	}
	engineKinds, err := os.ReadFile("../../engine/operation_kind.go")
	if err != nil {
		return fmt.Errorf("read shared operation kinds: %w", err)
	}
	for kind := range expected {
		capability, exists := byKind[kind]
		if !exists || capability.State != "present" || strings.TrimSpace(capability.Evidence) == "" {
			return fmt.Errorf("shared executor capability %s = %+v, want present evidence", kind, capability)
		}
		if !strings.Contains(string(engineKinds), fmt.Sprintf("%q", kind)) {
			return fmt.Errorf("shared executor capability %s is not registered by the engine", kind)
		}
	}
	return nil
}

func validateDockerHubReverseETLAudit(audit *dockerHubReverseETLAudit, matrix *dockerHubMatrix, lock *dockerHubLock, crosswalk *dockerHubCrosswalk, artifacts dockerHubArtifactRecords) error {
	if audit.SchemaVersion != 1 {
		return fmt.Errorf("reverse-ETL audit schema_version = %d, want 1", audit.SchemaVersion)
	}
	if audit.Connector != "dockerhub" {
		return fmt.Errorf("reverse-ETL audit connector = %q, want dockerhub", audit.Connector)
	}
	if audit.Purpose != "Source-backed direct-write and reverse-ETL mapping candidates only; this is not a runtime action inventory." {
		return fmt.Errorf("reverse-ETL audit purpose = %q, want source-mapping-only purpose", audit.Purpose)
	}
	if audit.SourceLock != dockerHubSourceLockPath || audit.SourceLaneMatrix != dockerHubMatrixPath {
		return errors.New("reverse-ETL audit does not name the exact source lock and lane matrix")
	}
	if artifacts.WritesArtifactPresent || artifacts.SyncTransportArtifactPresent {
		return errors.New("reverse-ETL audit must be updated when Docker Hub declares write or sync transport artifacts")
	}

	sources, err := dockerHubSourceInfos(lock, crosswalk)
	if err != nil {
		return err
	}
	matrixByID := dockerHubMatrixByID(matrix)
	expectedCandidates := make(map[string]dockerHubSourceInfo)
	deleteCount := 0
	for sourceID, source := range sources {
		if !dockerHubIsMutation(source.Lock.Method) {
			continue
		}
		expectedCandidates[sourceID] = source
		if source.Lock.Method == "DELETE" {
			deleteCount++
		}
		operation := matrixByID[sourceID]
		if operation == nil || dockerHubMappedCellByLane(operation, "direct_write") == nil || dockerHubMappedCellByLane(operation, "reverse_etl") == nil {
			return fmt.Errorf("source mutation %s is missing mapped direct-write/reverse-etl lanes", sourceID)
		}
	}
	if audit.Summary.SourceMutationOperations != len(expectedCandidates) ||
		audit.Summary.SourceDeleteOperations != deleteCount ||
		audit.Summary.MappedDirectWriteCells != len(expectedCandidates) ||
		audit.Summary.MappedReverseETLCells != len(expectedCandidates) {
		return fmt.Errorf("reverse-ETL audit source mapping counts = %+v, want mutation:%d delete:%d direct_write:%d reverse_etl:%d", audit.Summary, len(expectedCandidates), deleteCount, len(expectedCandidates), len(expectedCandidates))
	}
	if audit.Summary.DeclaredWriteActions != 0 || audit.Summary.DeclaredDestinationBindings != 0 || audit.Summary.ExecutableReverseETLActions != 0 {
		return fmt.Errorf("reverse-ETL audit execution counts = actions:%d destination_bindings:%d executable:%d, want zero while definition artifacts are absent", audit.Summary.DeclaredWriteActions, audit.Summary.DeclaredDestinationBindings, audit.Summary.ExecutableReverseETLActions)
	}

	seen := make(map[string]struct{}, len(audit.WriteOperations))
	for _, candidate := range audit.WriteOperations {
		if _, duplicate := seen[candidate.SourceID]; duplicate {
			return fmt.Errorf("reverse-ETL audit duplicates source candidate %q", candidate.SourceID)
		}
		seen[candidate.SourceID] = struct{}{}
		source, exists := expectedCandidates[candidate.SourceID]
		if !exists {
			return fmt.Errorf("reverse-ETL audit source candidate %q is not a locked mutation", candidate.SourceID)
		}
		if candidate.Method != source.Lock.Method || candidate.Path != source.Lock.Path || candidate.SourceURL != source.SourceURL || candidate.SourceLocation != source.Lock.SourceLocation {
			return fmt.Errorf("reverse-ETL audit source candidate %s does not preserve locked source facts", candidate.SourceID)
		}
		if !dockerHubEqualStrings(candidate.Lanes, []string{"direct_write", "reverse_etl"}) {
			return fmt.Errorf("reverse-ETL audit source candidate %s does not preserve mapped direct-write/reverse-etl lanes", candidate.SourceID)
		}
		operation := matrixByID[candidate.SourceID]
		if operation == nil || dockerHubMappedCellByLane(operation, "direct_write") == nil || dockerHubMappedCellByLane(operation, "reverse_etl") == nil {
			return fmt.Errorf("reverse-ETL audit source candidate %s is not backed by mapped direct-write/reverse-etl cells", candidate.SourceID)
		}
	}
	if len(seen) != len(expectedCandidates) {
		return fmt.Errorf("reverse-ETL audit candidate count = %d, want %d", len(seen), len(expectedCandidates))
	}
	return nil
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
	isBinaryDownload := dockerHubHasBinaryDownload(source)
	isBinaryUpload := dockerHubHasBinaryUpload(source)
	requiresETL := dockerHubRequiresETL(source, sourceContract)

	expected := make(map[string]dockerHubExpectedCell, len(dockerHubLaneOrder))
	if source.Lock.Method == "GET" {
		expected["direct_read"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.direct_read.documented_get_response.v1"}
	} else if source.Lock.Method == "HEAD" {
		expected["direct_read"] = dockerHubExpectedCell{"mapped_unproven", "dockerhub.source.direct_read.documented_head_status.v1"}
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
		reason := "dockerhub.source.binary_download.binary_response_media.v1"
		if _, responseMedia := dockerHubSourceMedia(source); dockerHubHasTextExportMedia(responseMedia) {
			reason = "dockerhub.source.binary_download.text_export_response_media.v1"
		}
		expected["binary_download"] = dockerHubExpectedCell{"mapped_unproven", reason}
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
	return dockerHubHasBinaryMedia(responseMedia, false) || dockerHubHasTextExportMedia(responseMedia)
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

func dockerHubHasTextExportMedia(media []string) bool {
	for _, contentType := range media {
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		if contentType == "text/csv" {
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
		"direct_read":     27,
		"direct_write":    27,
		"binary_download": 1,
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
	if mappedTotal != 102 || notApplicableTotal != 276 || mappedTotal+notApplicableTotal != 378 {
		return fmt.Errorf("matrix cell totals = mapped:%d not_applicable:%d total:%d, want 102/276/378", mappedTotal, notApplicableTotal, mappedTotal+notApplicableTotal)
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
		APIRoutes:                    make(map[string]struct{}, len(api.Endpoints)),
		Streams:                      make(map[string]dockerHubStreamRecord, len(streams.Streams)),
		Commands:                     make(map[string]dockerHubCommandRecord, len(cli.Commands)),
		OperationCount:               len(operations.Operations),
		WritesArtifactPresent:        dockerHubFilePresent(t, "writes.json"),
		SyncTransportArtifactPresent: dockerHubFilePresent(t, "sync_transport.json"),
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

func dockerHubFilePresent(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	t.Fatalf("stat %s: %v", path, err)
	return false
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
		if err := dockerHubValidateArtifactMappedLane("api_surface.json", link.Record, operation, lane); err != nil {
			return err
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
			if err := dockerHubValidateArtifactMappedLane("streams.json", link.Record, operation, lane); err != nil {
				return err
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
	// A CLI command's pre-existing intent is not the source operation's lane
	// classification. The two detail commands currently carry intent "etl",
	// but their source-backed GET identities are one-resource direct reads.
	sourceByStream := map[string]struct {
		SourceID string
		Lanes    []string
	}{
		"repositories": {
			SourceID: "dockerhub.rest.listNamespaceRepositories",
			Lanes:    []string{"etl"},
		},
		"tags": {
			SourceID: "dockerhub.rest.ListRepositoryTags",
			Lanes:    []string{"etl"},
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
		want, exists := sourceByStream[command.Stream]
		if !exists || command.Intent != "etl" || link.SourceID != want.SourceID {
			return fmt.Errorf("artifact cli_surface.json link %q does not preserve its source-backed stream route", link.Record)
		}
		operation := matrixByID[link.SourceID]
		if operation == nil {
			return fmt.Errorf("artifact cli_surface.json link %q references nonexistent source cell owner", link.Record)
		}
		if !dockerHubEqualStrings(link.Lanes, want.Lanes) {
			return fmt.Errorf("artifact cli_surface.json link %q does not preserve its source-backed lane", link.Record)
		}
		for _, lane := range link.Lanes {
			if dockerHubMatrixCellByLaneNoTest(operation, lane) == nil {
				return fmt.Errorf("artifact cli_surface.json link %q references nonexistent cell %q", link.Record, lane)
			}
			if err := dockerHubValidateArtifactMappedLane("cli_surface.json", link.Record, operation, lane); err != nil {
				return err
			}
		}
	}
	return nil
}

func dockerHubValidateArtifactMappedLane(artifactPath, record string, operation *dockerHubMatrixOperation, lane string) error {
	cell := dockerHubMatrixCellByLaneNoTest(operation, lane)
	if cell == nil {
		return fmt.Errorf("artifact %s link %q references nonexistent cell %q", artifactPath, record, lane)
	}
	if cell.State != "mapped_unproven" {
		return fmt.Errorf("artifact %s link %q lane %q does not point to a mapped lane: state=%q", artifactPath, record, lane, cell.State)
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

func dockerHubMatrixByID(matrix *dockerHubMatrix) map[string]*dockerHubMatrixOperation {
	byID := make(map[string]*dockerHubMatrixOperation, len(matrix.Operations))
	for index := range matrix.Operations {
		byID[matrix.Operations[index].SourceID] = &matrix.Operations[index]
	}
	return byID
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

func dockerHubMappedCellByLane(operation *dockerHubMatrixOperation, lane string) *dockerHubMatrixCell {
	cell := dockerHubMatrixCellByLaneNoTest(operation, lane)
	if cell == nil || cell.State != "mapped_unproven" {
		return nil
	}
	return cell
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
