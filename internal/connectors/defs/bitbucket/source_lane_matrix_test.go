package bitbucket

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

const (
	bitbucketSourceLaneMatrixPath = "sources/bitbucket-source-lane-matrix.json"
	bitbucketSourceLockPath       = "sources/bitbucket-operation-source-lock.json"
	bitbucketCrosswalkPath        = "sources/bitbucket-operation-crosswalk.json"
	bitbucketDirectReadPolicy     = "Provider-summary bounded-read action semantics classify direct-read candidates; HTTP method and source ID are retained facts, not lane selectors."
	bitbucketWriteReversePolicy   = "Provider-summary mutation action semantics classify direct-write and reverse-ETL candidates independently; HTTP method and source ID are retained facts, not lane selectors."
	bitbucketETLPolicy            = "Only a successful response schema resolved from the retained source contract with string next and array values declares continuation; request page or cursor controls are source evidence, while schema names and HTTP method are not selectors."
)

var bitbucketSourceLaneNames = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

// The source lock documents the exact binary response candidates below. The
// signal text is retained only to check that the mapping remains source-backed;
// it is not a runtime media allowance or executable command declaration.
var bitbucketBinaryDownloadEvidence = map[string]string{
	"bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/diff/{spec}":                          "raw git-style diff",
	"bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/downloads/{filename}":                 "actual file contents",
	"bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/patch/{spec}":                         "raw patch",
	"getPipelineStepLogForRepository":                                                               "log file",
	"getPipelineContainerLog":                                                                       "log file",
	"bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/diff":  "repository diff",
	"bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/patch": "repository patch",
	"bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/src/{commit}/{path}":                  "raw contents",
	"bitbucket.rest.get_/snippets/{workspace}/{encoded_id}/files/{path}":                            "raw files",
	"bitbucket.rest.get_/snippets/{workspace}/{encoded_id}/{node_id}/files/{path}":                  "raw contents",
	"bitbucket.rest.get_/snippets/{workspace}/{encoded_id}/{revision}/diff":                         "diff of the specified commit",
	"bitbucket.rest.get_/snippets/{workspace}/{encoded_id}/{revision}/patch":                        "patch of the specified commit",
	"bitbucket.rest.get_/workspaces/{workspace}/settings/gpg/public-key":                            "public GPG key",
}

var bitbucketBinaryUploadEvidence = map[string]string{
	"bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/downloads": "Upload new download artifacts",
	"bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/src":       "Create a commit by uploading a file",
}

// These provider operations configure delivery of selected Bitbucket events to
// an operator URL. They are source-backed sync candidates, but Track A records
// their unimplemented inbound receiver contract rather than pretending the
// generic outbound source transport is a webhook receiver.
var bitbucketWebhookDeliveryEvidence = map[string]string{
	"bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/hooks":      "events",
	"bitbucket.rest.put_/repositories/{workspace}/{repo_slug}/hooks/{uid}": "events",
	"bitbucket.rest.post_/workspaces/{workspace}/hooks":                    "events",
	"bitbucket.rest.put_/workspaces/{workspace}/hooks/{uid}":               "events",
}

var bitbucketWebhookCatalogIDs = map[string]struct{}{
	"bitbucket.rest.get_/hook_events":                {},
	"bitbucket.rest.get_/hook_events/{subject_type}": {},
}

type bitbucketSourceLaneMatrix struct {
	SchemaVersion                int                                    `json:"schema_version"`
	Connector                    string                                 `json:"connector"`
	Lanes                        []string                               `json:"lanes"`
	SourceLock                   bitbucketSourceLaneMatrixSourceLock    `json:"source_lock"`
	MappingPolicy                bitbucketSourceLaneMatrixMappingPolicy `json:"mapping_policy"`
	SourceBoundaryReconciliation bitbucketSourceBoundaryReconciliation  `json:"source_boundary_reconciliation"`
	SourceOperations             []bitbucketSourceLaneMatrixRow         `json:"source_operations"`
	Summary                      bitbucketSourceLaneMatrixSummary       `json:"summary"`
}

type bitbucketSourceLaneMatrixMappingPolicy struct {
	SourceAuthority          string `json:"source_authority"`
	DirectRead               string `json:"direct_read"`
	DirectWriteAndReverseETL string `json:"direct_write_and_reverse_etl"`
	Binary                   string `json:"binary"`
	ETL                      string `json:"etl"`
	SyncTransport            string `json:"sync_transport"`
}

type bitbucketSourceLaneMatrixSourceLock struct {
	Path           string `json:"path"`
	SchemaVersion  int    `json:"schema_version"`
	Connector      string `json:"connector"`
	SourceDocument struct {
		SourceURL      string `json:"source_url"`
		SHA256         string `json:"sha256"`
		Bytes          int    `json:"bytes"`
		OperationCount int    `json:"operation_count"`
	} `json:"source_document"`
}

type bitbucketSourceBoundaryReconciliation struct {
	Identity                    string                               `json:"identity"`
	SourceLockRows              int                                  `json:"source_lock_rows"`
	CrosswalkRows               int                                  `json:"crosswalk_rows"`
	CrosswalkOnlyRows           int                                  `json:"crosswalk_only_rows"`
	LockOnlyRows                int                                  `json:"lock_only_rows"`
	CrosswalkOnlySourceIdentity []bitbucketCrosswalkBoundaryIdentity `json:"crosswalk_only_source_identities"`
}

type bitbucketCrosswalkBoundaryIdentity struct {
	CrosswalkSourceID string `json:"crosswalk_source_id"`
	Method            string `json:"method"`
	Path              string `json:"path"`
	SourceLocation    string `json:"source_location"`
	Disposition       string `json:"disposition"`
	Reason            string `json:"reason"`
}

type bitbucketSourceLaneMatrixSummary struct {
	SourceRows             int                       `json:"source_rows"`
	SourceRowsWithAllLanes int                       `json:"source_rows_with_all_lanes"`
	LaneCounts             map[string]map[string]int `json:"lane_counts"`
}

type bitbucketSourceLaneMatrixRow struct {
	SourceID    string                                   `json:"source_id"`
	SourceFacts bitbucketSourceLaneMatrixSourceFacts     `json:"source_facts"`
	Lanes       map[string]bitbucketSourceLaneMatrixCell `json:"lanes"`
}

type bitbucketSourceLaneMatrixSourceFacts struct {
	Protocol    string `json:"protocol"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	OperationID string `json:"operation_id"`
	Deprecated  *bool  `json:"deprecated"`
	Citation    struct {
		SourceLocation string `json:"source_location"`
	} `json:"citation"`
	ScopeAndFanout struct {
		PathVariables   []string `json:"path_variables"`
		QueryParameters []string `json:"query_parameters"`
		Fanout          struct {
			State string `json:"state"`
		} `json:"fanout"`
	} `json:"scope_and_fanout"`
	Media struct {
		RequestMediaTypes         []string `json:"request_media_types"`
		SuccessResponseMediaTypes []string `json:"success_response_media_types"`
		BinarySignals             []string `json:"binary_signals"`
	} `json:"media"`
	Pagination struct {
		State                  string   `json:"state"`
		ResponseSchemaRefs     []string `json:"response_schema_refs"`
		RequestQueryParameters []string `json:"request_query_parameters"`
		ContinuationFields     []string `json:"continuation_fields"`
	} `json:"pagination"`
	EventCursor struct {
		State string `json:"state"`
	} `json:"event_cursor"`
	OperationSemantics struct {
		State         string `json:"state"`
		SummaryAction string `json:"summary_action"`
	} `json:"operation_semantics"`
}

type bitbucketSourceLaneMatrixCell struct {
	Applicability string          `json:"applicability"`
	Disposition   string          `json:"disposition"`
	Reason        string          `json:"reason"`
	Mapping       json.RawMessage `json:"mapping"`
}

type bitbucketSourceLaneLock struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	Counts        struct {
		Total int `json:"total"`
	} `json:"counts"`
	REST struct {
		SourceURL  string                           `json:"source_url"`
		SHA256     string                           `json:"sha256"`
		Bytes      int                              `json:"bytes"`
		Operations []bitbucketLockedSourceOperation `json:"operations"`
	} `json:"rest"`
	SourceContract bitbucketLockedSourceContract `json:"source_contract"`
}

type bitbucketLockedSourceContract struct {
	Components struct {
		Schemas map[string]json.RawMessage `json:"schemas"`
	} `json:"components"`
}

type bitbucketLockedSourceOperation struct {
	ID              string                           `json:"id"`
	Protocol        string                           `json:"protocol"`
	Method          string                           `json:"method"`
	Path            string                           `json:"path"`
	OperationID     string                           `json:"operation_id"`
	Deprecated      *bool                            `json:"deprecated"`
	SourceLocation  string                           `json:"source_location"`
	SourceOperation bitbucketLockedProviderOperation `json:"source_operation"`
}

type bitbucketLockedProviderOperation struct {
	Summary        string                             `json:"summary"`
	Description    string                             `json:"description"`
	PathParameters []bitbucketLockedSourceParameter   `json:"path_parameters"`
	Parameters     []bitbucketLockedSourceParameter   `json:"parameters"`
	RequestBody    *bitbucketLockedRequestBody        `json:"requestBody"`
	Responses      map[string]bitbucketLockedResponse `json:"responses"`
}

type bitbucketLockedSourceParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
}

type bitbucketLockedRequestBody struct {
	Content map[string]json.RawMessage `json:"content"`
}

type bitbucketLockedResponse struct {
	Content map[string]bitbucketLockedResponseMedia `json:"content"`
}

type bitbucketLockedResponseMedia struct {
	Schema json.RawMessage `json:"schema"`
}

type bitbucketPaginationEvidence struct {
	ResponseSchemaRefs     []string
	RequestQueryParameters []string
	ContinuationFields     []string
}

type bitbucketSourceLaneCrosswalk struct {
	SourceOperations []bitbucketCrosswalkSourceOperation `json:"source_operations"`
}

type bitbucketCrosswalkSourceOperation struct {
	SourceID       string `json:"source_id"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	SourceLocation string `json:"source_location"`
}

func TestBitbucketSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadBitbucketSourceLaneMatrix(t)
	lock := loadBitbucketSourceLaneLock(t)
	crosswalk := loadBitbucketSourceLaneCrosswalk(t)
	if err := validateBitbucketSourceLaneMatrix(matrix, lock, crosswalk); err != nil {
		t.Fatalf("validate Bitbucket source lane matrix: %v", err)
	}

	t.Run("rejects missing lane cell", func(t *testing.T) {
		broken := cloneBitbucketSourceLaneMatrix(matrix)
		delete(broken.SourceOperations[0].Lanes, "sync_transport")
		if err := validateBitbucketSourceLaneMatrix(broken, lock, crosswalk); err == nil || !strings.Contains(err.Error(), "missing lane cell") {
			t.Fatalf("missing-cell validation error = %v, want missing lane cell", err)
		}
	})

	t.Run("rejects hidden source row", func(t *testing.T) {
		broken := cloneBitbucketSourceLaneMatrix(matrix)
		broken.SourceOperations = broken.SourceOperations[1:]
		if err := validateBitbucketSourceLaneMatrix(broken, lock, crosswalk); err == nil || !strings.Contains(err.Error(), "source rows matrix=") {
			t.Fatalf("hidden-row validation error = %v, want source rows matrix", err)
		}
	})

	t.Run("rejects crosswalk boundary drop", func(t *testing.T) {
		broken := cloneBitbucketSourceLaneMatrix(matrix)
		broken.SourceBoundaryReconciliation.CrosswalkOnlySourceIdentity = broken.SourceBoundaryReconciliation.CrosswalkOnlySourceIdentity[1:]
		if err := validateBitbucketSourceLaneMatrix(broken, lock, crosswalk); err == nil || !strings.Contains(err.Error(), "crosswalk-only identities") {
			t.Fatalf("crosswalk-boundary validation error = %v, want crosswalk-only identities", err)
		}
	})

	for _, test := range []struct {
		name   string
		mutate func(*bitbucketSourceLaneMatrix)
	}{
		{
			name: "HTTP method direct-read selector",
			mutate: func(matrix *bitbucketSourceLaneMatrix) {
				matrix.MappingPolicy.DirectRead = "GET operations are direct-read candidates."
			},
		},
		{
			name: "HTTP method write selector",
			mutate: func(matrix *bitbucketSourceLaneMatrix) {
				matrix.MappingPolicy.DirectWriteAndReverseETL = "DELETE, PUT, and POST operations are direct-write and reverse-ETL candidates."
			},
		},
		{
			name: "schema-name ETL selector",
			mutate: func(matrix *bitbucketSourceLaneMatrix) {
				matrix.MappingPolicy.ETL = "Successful GET responses with paginated_ schema names are ETL candidates."
			},
		},
	} {
		t.Run("rejects "+test.name+" policy drift", func(t *testing.T) {
			broken := cloneBitbucketSourceLaneMatrix(matrix)
			test.mutate(&broken)
			if err := validateBitbucketSourceLaneMatrix(broken, lock, crosswalk); err == nil || !strings.Contains(err.Error(), "mapping policy") {
				t.Fatalf("mapping-policy validation error = %v, want mapping policy", err)
			}
		})
	}
}

func TestBitbucketSourceLaneMatrixSemanticSourceRules(t *testing.T) {
	lock := loadBitbucketSourceLaneLock(t)

	t.Run("documented POST list read is not a mutation", func(t *testing.T) {
		operation := findBitbucketLockedOperation(t, lock, "bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/commits")
		if operation.Method != "POST" {
			t.Fatalf("operation method = %q, want POST test fixture", operation.Method)
		}
		if !bitbucketIsDirectReadCandidate(operation) {
			t.Fatal("documented POST list read was not classified as a direct-read candidate")
		}
		if bitbucketIsMutationCandidate(operation) {
			t.Fatal("documented POST list read was incorrectly classified as a mutation candidate")
		}
	})

	t.Run("documented mutation remains a write and reverse-ETL candidate", func(t *testing.T) {
		operation := findBitbucketLockedOperation(t, lock, "bitbucket.rest.post_/repositories/{workspace}/{repo_slug}")
		if operation.SourceOperation.Summary != "Create a repository" {
			t.Fatalf("operation summary = %q, want Create a repository test fixture", operation.SourceOperation.Summary)
		}
		if bitbucketIsDirectReadCandidate(operation) {
			t.Fatal("documented create was incorrectly classified as a direct-read candidate")
		}
		if !bitbucketIsMutationCandidate(operation) {
			t.Fatal("documented create was not classified as a mutation candidate")
		}
	})

	t.Run("response schema spelling cannot hide documented search pagination", func(t *testing.T) {
		for _, sourceID := range []string{"searchTeam", "searchAccount", "searchWorkspace"} {
			operation := findBitbucketLockedOperation(t, lock, sourceID)
			pagination := bitbucketPaginationEvidenceForOperation(lock, operation)
			refs := pagination.ResponseSchemaRefs
			if !slices.Contains(refs, "#/components/schemas/search_result_page") {
				t.Errorf("%s pagination refs = %v, want source-backed search_result_page", sourceID, refs)
			}
			if !slices.Equal(pagination.RequestQueryParameters, []string{"page", "pagelen"}) ||
				!slices.Equal(pagination.ContinuationFields, []string{"next", "values"}) {
				t.Errorf("%s pagination evidence = %#v, want page/pagelen plus next/values", sourceID, pagination)
			}
		}
	})

	t.Run("values without a continuation link is not ETL", func(t *testing.T) {
		broken := bitbucketSourceLaneLock{}
		broken.SourceContract.Components.Schemas = map[string]json.RawMessage{
			"ordinary_collection": json.RawMessage(`{"type":"object","properties":{"values":{"type":"array"}}}`),
		}
		operation := bitbucketLockedSourceOperation{
			SourceOperation: bitbucketLockedProviderOperation{
				Responses: map[string]bitbucketLockedResponse{
					"200": {
						Content: map[string]bitbucketLockedResponseMedia{
							"application/json": {Schema: json.RawMessage(`{"$ref":"#/components/schemas/ordinary_collection"}`)},
						},
					},
				},
			},
		}
		if pagination := bitbucketPaginationEvidenceForOperation(broken, operation); len(pagination.ResponseSchemaRefs) != 0 {
			t.Fatalf("noncontinuable collection pagination refs = %v, want none", pagination.ResponseSchemaRefs)
		}
	})

	t.Run("noncontinuable list stays direct-only", func(t *testing.T) {
		operation := findBitbucketLockedOperation(t, lock, "bitbucket.rest.get_/repositories/{workspace}/{repo_slug}/downloads")
		if operation.SourceOperation.Summary != "List download artifacts" {
			t.Fatalf("operation summary = %q, want noncontinuable-list fixture", operation.SourceOperation.Summary)
		}
		if !bitbucketIsDirectReadCandidate(operation) {
			t.Fatal("documented list was not classified as a direct-read candidate")
		}
		if pagination := bitbucketPaginationEvidenceForOperation(lock, operation); len(pagination.ResponseSchemaRefs) != 0 {
			t.Fatalf("noncontinuable list pagination refs = %v, want none", pagination.ResponseSchemaRefs)
		}
	})

	t.Run("matrix has source-semantic backlinks and coverage totals", func(t *testing.T) {
		matrix := loadBitbucketSourceLaneMatrix(t)
		rows := make(map[string]bitbucketSourceLaneMatrixRow, len(matrix.SourceOperations))
		for _, row := range matrix.SourceOperations {
			rows[row.SourceID] = row
		}

		gotDirectRead := 0
		gotMutations := 0
		gotETL := 0
		for _, operation := range lock.REST.Operations {
			row, ok := rows[operation.ID]
			if !ok {
				t.Fatalf("matrix row missing for %s", operation.ID)
			}
			if bitbucketIsDirectReadCandidate(operation) {
				gotDirectRead++
				if err := validateBitbucketLaneMapping("direct_read", row.Lanes["direct_read"].Mapping, operation, lock); err != nil {
					t.Errorf("direct-read backlink %s: %v", operation.ID, err)
				}
			}
			if bitbucketIsMutationCandidate(operation) {
				gotMutations++
				for _, lane := range []string{"direct_write", "reverse_etl"} {
					if err := validateBitbucketLaneMapping(lane, row.Lanes[lane].Mapping, operation, lock); err != nil {
						t.Errorf("%s backlink %s: %v", lane, operation.ID, err)
					}
				}
			}
			if len(bitbucketPaginationEvidenceForOperation(lock, operation).ResponseSchemaRefs) > 0 {
				gotETL++
				if err := validateBitbucketLaneMapping("etl", row.Lanes["etl"].Mapping, operation, lock); err != nil {
					t.Errorf("ETL backlink %s: %v", operation.ID, err)
				}
			}
		}
		if gotDirectRead != 162 || gotMutations != 135 || gotETL != 73 {
			t.Fatalf("source semantic counts direct_read=%d mutations=%d etl=%d, want 162/135/73", gotDirectRead, gotMutations, gotETL)
		}
		for _, lane := range []string{"direct_read", "direct_write", "reverse_etl", "etl"} {
			if mapped := matrix.Summary.LaneCounts[lane]["mapped_unproven"]; mapped != map[string]int{"direct_read": 162, "direct_write": 135, "reverse_etl": 135, "etl": 73}[lane] {
				t.Errorf("summary %s mapped_unproven=%d, want source-semantic total", lane, mapped)
			}
		}
		searchTeam := rows["searchTeam"]
		if searchTeam.Lanes["etl"].Disposition != "mapped_unproven" || searchTeam.Lanes["sync_transport"].Disposition != "not_applicable" {
			t.Errorf("searchTeam ETL/sync dispositions = %q/%q, want mapped_unproven/not_applicable", searchTeam.Lanes["etl"].Disposition, searchTeam.Lanes["sync_transport"].Disposition)
		}
	})
}

func findBitbucketLockedOperation(t *testing.T, lock bitbucketSourceLaneLock, sourceID string) bitbucketLockedSourceOperation {
	t.Helper()
	for _, operation := range lock.REST.Operations {
		if operation.ID == sourceID {
			return operation
		}
	}
	t.Fatalf("source operation %q not present in locked Bitbucket source", sourceID)
	return bitbucketLockedSourceOperation{}
}

func loadBitbucketSourceLaneMatrix(t *testing.T) bitbucketSourceLaneMatrix {
	t.Helper()
	return loadBitbucketJSON[bitbucketSourceLaneMatrix](t, bitbucketSourceLaneMatrixPath)
}

func loadBitbucketSourceLaneLock(t *testing.T) bitbucketSourceLaneLock {
	t.Helper()
	return loadBitbucketJSON[bitbucketSourceLaneLock](t, bitbucketSourceLockPath)
}

func loadBitbucketSourceLaneCrosswalk(t *testing.T) bitbucketSourceLaneCrosswalk {
	t.Helper()
	return loadBitbucketJSON[bitbucketSourceLaneCrosswalk](t, bitbucketCrosswalkPath)
}

func loadBitbucketJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return decoded
}

func validateBitbucketSourceLaneMatrix(matrix bitbucketSourceLaneMatrix, lock bitbucketSourceLaneLock, crosswalk bitbucketSourceLaneCrosswalk) error {
	if matrix.SchemaVersion != 1 || matrix.Connector != "bitbucket" {
		return fmt.Errorf("matrix identity schema=%d connector=%q, want schema=1 connector=bitbucket", matrix.SchemaVersion, matrix.Connector)
	}
	if !slices.Equal(matrix.Lanes, bitbucketSourceLaneNames) {
		return fmt.Errorf("lane order = %v, want %v", matrix.Lanes, bitbucketSourceLaneNames)
	}
	if err := validateBitbucketMatrixLockBinding(matrix.SourceLock, lock); err != nil {
		return err
	}
	if err := validateBitbucketMappingPolicy(matrix.MappingPolicy); err != nil {
		return err
	}

	locked := make(map[string]bitbucketLockedSourceOperation, len(lock.REST.Operations))
	lockedMethodPaths := make(map[string]struct{}, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		if _, exists := locked[operation.ID]; exists {
			return fmt.Errorf("duplicate source lock ID %q", operation.ID)
		}
		locked[operation.ID] = operation
		lockedMethodPaths[bitbucketMethodPathKey(operation.Method, operation.Path)] = struct{}{}
	}
	if lock.Counts.Total != 297 || len(locked) != 297 {
		return fmt.Errorf("source lock denominator=%d unique=%d, want 297", lock.Counts.Total, len(locked))
	}
	if len(matrix.SourceOperations) != len(locked) {
		return fmt.Errorf("source rows matrix=%d lock=%d, want 297", len(matrix.SourceOperations), len(locked))
	}

	if err := validateBitbucketCrosswalkBoundary(matrix.SourceBoundaryReconciliation, crosswalk, lockedMethodPaths); err != nil {
		return err
	}

	counts := make(map[string]map[string]int, len(bitbucketSourceLaneNames))
	seen := make(map[string]struct{}, len(matrix.SourceOperations))
	for _, row := range matrix.SourceOperations {
		if _, exists := seen[row.SourceID]; exists {
			return fmt.Errorf("duplicate matrix source ID %q", row.SourceID)
		}
		seen[row.SourceID] = struct{}{}
		operation, ok := locked[row.SourceID]
		if !ok {
			return fmt.Errorf("matrix source ID %q is absent from the source lock", row.SourceID)
		}
		if err := validateBitbucketSourceFacts(row.SourceFacts, operation, lock); err != nil {
			return fmt.Errorf("source facts %q: %w", row.SourceID, err)
		}
		for _, lane := range bitbucketSourceLaneNames {
			cell, ok := row.Lanes[lane]
			if !ok {
				return fmt.Errorf("missing lane cell: %s %s", row.SourceID, lane)
			}
			if err := validateBitbucketLaneCell(row.SourceID, lane, cell, operation, lock); err != nil {
				return err
			}
			if counts[lane] == nil {
				counts[lane] = make(map[string]int)
			}
			counts[lane][cell.Disposition]++
		}
		if len(row.Lanes) != len(bitbucketSourceLaneNames) {
			return fmt.Errorf("lane cell count %s=%d, want %d", row.SourceID, len(row.Lanes), len(bitbucketSourceLaneNames))
		}
	}
	if len(seen) != len(locked) {
		return fmt.Errorf("matrix does not retain every locked source ID: matrix=%d lock=%d", len(seen), len(locked))
	}
	if matrix.Summary.SourceRows != len(locked) || matrix.Summary.SourceRowsWithAllLanes != len(locked) {
		return fmt.Errorf("summary source rows=%d rows_with_all_lanes=%d, want 297", matrix.Summary.SourceRows, matrix.Summary.SourceRowsWithAllLanes)
	}
	for _, lane := range bitbucketSourceLaneNames {
		if !equalBitbucketLaneCounts(matrix.Summary.LaneCounts[lane], counts[lane]) {
			return fmt.Errorf("summary %s counts=%v, computed=%v", lane, matrix.Summary.LaneCounts[lane], counts[lane])
		}
	}
	return nil
}

func validateBitbucketMatrixLockBinding(binding bitbucketSourceLaneMatrixSourceLock, lock bitbucketSourceLaneLock) error {
	if binding.Path != bitbucketSourceLockPath || binding.SchemaVersion != lock.SchemaVersion || binding.Connector != lock.Connector {
		return fmt.Errorf("source lock binding path=%q schema=%d connector=%q", binding.Path, binding.SchemaVersion, binding.Connector)
	}
	if binding.SourceDocument.SourceURL != lock.REST.SourceURL || binding.SourceDocument.SHA256 != lock.REST.SHA256 || binding.SourceDocument.Bytes != lock.REST.Bytes || binding.SourceDocument.OperationCount != lock.Counts.Total {
		return fmt.Errorf("source lock document binding drift")
	}
	return nil
}

func validateBitbucketMappingPolicy(policy bitbucketSourceLaneMatrixMappingPolicy) error {
	if strings.TrimSpace(policy.SourceAuthority) == "" || strings.TrimSpace(policy.Binary) == "" || strings.TrimSpace(policy.SyncTransport) == "" {
		return fmt.Errorf("mapping policy lacks retained source facts")
	}
	if policy.DirectRead != bitbucketDirectReadPolicy ||
		policy.DirectWriteAndReverseETL != bitbucketWriteReversePolicy ||
		policy.ETL != bitbucketETLPolicy {
		return fmt.Errorf("mapping policy does not match source-semantic lane rules")
	}
	return nil
}

func validateBitbucketCrosswalkBoundary(boundary bitbucketSourceBoundaryReconciliation, crosswalk bitbucketSourceLaneCrosswalk, lockedMethodPaths map[string]struct{}) error {
	if len(crosswalk.SourceOperations) != 331 {
		return fmt.Errorf("crosswalk denominator=%d, want 331", len(crosswalk.SourceOperations))
	}
	if boundary.Identity != "method + path" || boundary.SourceLockRows != 297 || boundary.CrosswalkRows != len(crosswalk.SourceOperations) || boundary.CrosswalkOnlyRows != 34 || boundary.LockOnlyRows != 0 {
		return fmt.Errorf("crosswalk boundary counts identity=%q lock=%d crosswalk=%d crosswalk_only=%d lock_only=%d", boundary.Identity, boundary.SourceLockRows, boundary.CrosswalkRows, boundary.CrosswalkOnlyRows, boundary.LockOnlyRows)
	}
	want := make(map[string]bitbucketCrosswalkSourceOperation)
	crosswalkMethodPaths := make(map[string]struct{}, len(crosswalk.SourceOperations))
	for _, operation := range crosswalk.SourceOperations {
		key := bitbucketMethodPathKey(operation.Method, operation.Path)
		if _, exists := crosswalkMethodPaths[key]; exists {
			return fmt.Errorf("duplicate crosswalk method/path identity %q", key)
		}
		crosswalkMethodPaths[key] = struct{}{}
		if _, locked := lockedMethodPaths[key]; !locked {
			want[key] = operation
		}
	}
	if len(want) != 34 {
		return fmt.Errorf("crosswalk-minus-lock identities=%d, want 34", len(want))
	}
	for key := range lockedMethodPaths {
		if _, present := crosswalkMethodPaths[key]; !present {
			return fmt.Errorf("lock-minus-crosswalk identity %q", key)
		}
	}
	got := make(map[string]bitbucketCrosswalkBoundaryIdentity, len(boundary.CrosswalkOnlySourceIdentity))
	for _, identity := range boundary.CrosswalkOnlySourceIdentity {
		key := bitbucketMethodPathKey(identity.Method, identity.Path)
		if _, locked := lockedMethodPaths[key]; locked {
			return fmt.Errorf("crosswalk-only identity %q is a retained source row", key)
		}
		if _, exists := got[key]; exists {
			return fmt.Errorf("duplicate crosswalk-only identity %q", key)
		}
		if identity.Disposition != "not_source_row" || strings.TrimSpace(identity.Reason) == "" {
			return fmt.Errorf("crosswalk-only identity %q lacks explicit boundary disposition", key)
		}
		got[key] = identity
	}
	if len(got) != len(want) {
		return fmt.Errorf("crosswalk-only identities matrix=%d want=%d", len(got), len(want))
	}
	for key, expected := range want {
		identity, ok := got[key]
		if !ok {
			return fmt.Errorf("crosswalk-only identities missing %q", key)
		}
		if identity.CrosswalkSourceID != expected.SourceID || identity.SourceLocation != expected.SourceLocation {
			return fmt.Errorf("crosswalk-only identity drift for %q", key)
		}
	}
	return nil
}

func validateBitbucketSourceFacts(facts bitbucketSourceLaneMatrixSourceFacts, operation bitbucketLockedSourceOperation, lock bitbucketSourceLaneLock) error {
	if facts.Protocol != operation.Protocol || facts.Method != operation.Method || facts.Path != operation.Path || facts.OperationID != operation.OperationID || facts.Citation.SourceLocation != operation.SourceLocation || !equalOptionalBool(facts.Deprecated, operation.Deprecated) {
		return fmt.Errorf("identity or citation drift")
	}
	if !slices.Equal(facts.ScopeAndFanout.PathVariables, bitbucketPathVariables(operation)) || !slices.Equal(facts.ScopeAndFanout.QueryParameters, bitbucketQueryParameters(operation)) || facts.ScopeAndFanout.Fanout.State != "not_declared" {
		return fmt.Errorf("scope or fanout facts drift")
	}
	requestMedia, responseMedia := bitbucketMediaTypes(operation)
	if !slices.Equal(facts.Media.RequestMediaTypes, requestMedia) || !slices.Equal(facts.Media.SuccessResponseMediaTypes, responseMedia) || !slices.Equal(facts.Media.BinarySignals, bitbucketBinarySignals(operation)) {
		return fmt.Errorf("media facts drift")
	}
	pagination := bitbucketPaginationEvidenceForOperation(lock, operation)
	wantPagination := "not_declared"
	if len(pagination.ResponseSchemaRefs) > 0 {
		wantPagination = "declared"
	}
	if facts.Pagination.State != wantPagination ||
		!slices.Equal(facts.Pagination.ResponseSchemaRefs, pagination.ResponseSchemaRefs) ||
		!slices.Equal(facts.Pagination.RequestQueryParameters, pagination.RequestQueryParameters) ||
		!slices.Equal(facts.Pagination.ContinuationFields, pagination.ContinuationFields) {
		return fmt.Errorf("pagination facts drift")
	}
	if facts.EventCursor.State != bitbucketEventCursorState(operation) ||
		facts.OperationSemantics.State != bitbucketOperationSemantics(operation) ||
		facts.OperationSemantics.SummaryAction != bitbucketSummaryAction(operation) {
		return fmt.Errorf("event/cursor or operation semantics facts drift")
	}
	if facts.OperationSemantics.State == "not_classified" {
		return fmt.Errorf("source summary action %q has no declared semantic classification", facts.OperationSemantics.SummaryAction)
	}
	if err := validateBitbucketCandidateEvidence(operation); err != nil {
		return err
	}
	return nil
}

func validateBitbucketCandidateEvidence(operation bitbucketLockedSourceOperation) error {
	sourceText := operation.SourceOperation.Summary + "\n" + operation.SourceOperation.Description
	for _, evidence := range []map[string]string{
		bitbucketBinaryDownloadEvidence,
		bitbucketBinaryUploadEvidence,
		bitbucketWebhookDeliveryEvidence,
	} {
		if required, ok := evidence[operation.ID]; ok && !strings.Contains(sourceText, required) {
			return fmt.Errorf("candidate evidence %q is absent from locked source text", required)
		}
	}
	return nil
}

func validateBitbucketLaneCell(sourceID, lane string, cell bitbucketSourceLaneMatrixCell, operation bitbucketLockedSourceOperation, lock bitbucketSourceLaneLock) error {
	wantApplicable, wantDisposition := bitbucketExpectedLane(operation, lock, lane)
	if cell.Applicability != wantApplicable || cell.Disposition != wantDisposition {
		return fmt.Errorf("lane %s %s applicability=%q disposition=%q, want applicability=%q disposition=%q", sourceID, lane, cell.Applicability, cell.Disposition, wantApplicable, wantDisposition)
	}
	if strings.TrimSpace(cell.Reason) == "" {
		return fmt.Errorf("lane %s %s lacks a reason", sourceID, lane)
	}
	if cell.Applicability == "not_applicable" {
		if cell.Disposition != "not_applicable" || len(cell.Mapping) != 0 {
			return fmt.Errorf("not-applicable lane promoted or mapped: %s %s", sourceID, lane)
		}
		return nil
	}
	if cell.Applicability != "applicable" || (cell.Disposition != "mapped_unproven" && cell.Disposition != "missing_foundation") {
		return fmt.Errorf("invalid applicable lane state: %s %s", sourceID, lane)
	}
	if len(cell.Mapping) == 0 || !json.Valid(cell.Mapping) {
		return fmt.Errorf("applicable lane lacks valid mapping evidence: %s %s", sourceID, lane)
	}
	if cell.Disposition == "missing_foundation" {
		if lane != "sync_transport" {
			return fmt.Errorf("non-sync missing-foundation lane: %s %s", sourceID, lane)
		}
	}
	if err := validateBitbucketLaneMapping(lane, cell.Mapping, operation, lock); err != nil {
		return fmt.Errorf("mapping evidence %s %s: %w", sourceID, lane, err)
	}
	return nil
}

func validateBitbucketLaneMapping(lane string, raw json.RawMessage, operation bitbucketLockedSourceOperation, lock bitbucketSourceLaneLock) error {
	switch lane {
	case "direct_read":
		var mapping struct {
			SourceFact struct {
				SemanticAction string `json:"semantic_action"`
				Classification string `json:"classification"`
				SourceSummary  string `json:"source_summary"`
			} `json:"source_fact"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.SourceFact.SemanticAction != bitbucketSummaryAction(operation) ||
			mapping.SourceFact.Classification != "read_candidate" ||
			mapping.SourceFact.SourceSummary != operation.SourceOperation.Summary ||
			strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("direct-read source mapping drift")
		}
	case "direct_write":
		var mapping struct {
			SourceFact struct {
				SemanticAction string `json:"semantic_action"`
				Classification string `json:"classification"`
				SourceSummary  string `json:"source_summary"`
			} `json:"source_fact"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.SourceFact.SemanticAction != bitbucketSummaryAction(operation) ||
			mapping.SourceFact.Classification != "mutation_candidate" ||
			mapping.SourceFact.SourceSummary != operation.SourceOperation.Summary ||
			strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("direct-write source mapping drift")
		}
	case "binary_download":
		return validateBitbucketBinaryLaneMapping(raw, bitbucketBinaryDownloadEvidence[operation.ID])
	case "binary_upload":
		return validateBitbucketBinaryLaneMapping(raw, bitbucketBinaryUploadEvidence[operation.ID])
	case "etl":
		var mapping struct {
			PaginationState        string   `json:"pagination_state"`
			ResponseSchemaRefs     []string `json:"response_schema_refs"`
			RequestQueryParameters []string `json:"request_query_parameters"`
			ContinuationFields     []string `json:"continuation_fields"`
			RuntimeClaim           string   `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		pagination := bitbucketPaginationEvidenceForOperation(lock, operation)
		if mapping.PaginationState != "declared" ||
			!slices.Equal(mapping.ResponseSchemaRefs, pagination.ResponseSchemaRefs) ||
			!slices.Equal(mapping.RequestQueryParameters, pagination.RequestQueryParameters) ||
			!slices.Equal(mapping.ContinuationFields, pagination.ContinuationFields) ||
			strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("ETL source mapping drift")
		}
	case "reverse_etl":
		var mapping struct {
			SourceFact struct {
				SemanticAction string `json:"semantic_action"`
				Classification string `json:"classification"`
				SourceSummary  string `json:"source_summary"`
			} `json:"source_fact"`
			RequiredFlow string `json:"required_flow"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.SourceFact.SemanticAction != bitbucketSummaryAction(operation) ||
			mapping.SourceFact.Classification != "mutation_candidate" ||
			mapping.SourceFact.SourceSummary != operation.SourceOperation.Summary ||
			strings.TrimSpace(mapping.RequiredFlow) == "" ||
			strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("reverse-ETL source mapping drift")
		}
	case "sync_transport":
		var mapping struct {
			FoundationID string `json:"foundation_id"`
			AtlasLookup  struct {
				ConsultedID    string `json:"consulted_id"`
				Classification string `json:"classification"`
			} `json:"atlas_lookup"`
			SourceEventEvidence string `json:"source_event_evidence"`
			RuntimeClaim        string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.FoundationID != "cli-webhook-event-surface-foundation-r1" || mapping.AtlasLookup.ConsultedID != "transport.sync-contract.v1" || mapping.AtlasLookup.Classification != "actual_gap" || mapping.SourceEventEvidence != bitbucketWebhookDeliveryEvidence[operation.ID] || strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("missing-foundation mapping drift")
		}
	default:
		return fmt.Errorf("unknown lane %q", lane)
	}
	return nil
}

func validateBitbucketBinaryLaneMapping(raw json.RawMessage, wantSignal string) error {
	var mapping struct {
		SourceBinarySignal string `json:"source_binary_signal"`
		RuntimeClaim       string `json:"runtime_claim"`
	}
	if err := json.Unmarshal(raw, &mapping); err != nil {
		return err
	}
	if mapping.SourceBinarySignal != wantSignal || strings.TrimSpace(mapping.RuntimeClaim) == "" {
		return fmt.Errorf("binary source mapping drift")
	}
	return nil
}

func bitbucketExpectedLane(operation bitbucketLockedSourceOperation, lock bitbucketSourceLaneLock, lane string) (string, string) {
	applicable := false
	switch lane {
	case "direct_read":
		applicable = bitbucketIsDirectReadCandidate(operation)
	case "direct_write", "reverse_etl":
		applicable = bitbucketIsMutationCandidate(operation)
	case "binary_download":
		_, applicable = bitbucketBinaryDownloadEvidence[operation.ID]
	case "binary_upload":
		_, applicable = bitbucketBinaryUploadEvidence[operation.ID]
	case "etl":
		applicable = len(bitbucketPaginationEvidenceForOperation(lock, operation).ResponseSchemaRefs) > 0
	case "sync_transport":
		_, applicable = bitbucketWebhookDeliveryEvidence[operation.ID]
	}
	if !applicable {
		return "not_applicable", "not_applicable"
	}
	if lane == "sync_transport" {
		return "applicable", "missing_foundation"
	}
	return "applicable", "mapped_unproven"
}

func bitbucketIsDirectReadCandidate(operation bitbucketLockedSourceOperation) bool {
	return bitbucketOperationSemantics(operation) == "read_candidate"
}

func bitbucketIsMutationCandidate(operation bitbucketLockedSourceOperation) bool {
	return bitbucketOperationSemantics(operation) == "mutation_candidate"
}

func bitbucketPathVariables(operation bitbucketLockedSourceOperation) []string {
	values := make([]string, 0, len(operation.SourceOperation.PathParameters))
	for _, parameter := range operation.SourceOperation.PathParameters {
		if parameter.Name != "" {
			values = append(values, parameter.Name)
		}
	}
	for _, parameter := range operation.SourceOperation.Parameters {
		if parameter.In == "path" && parameter.Name != "" {
			values = append(values, parameter.Name)
		}
	}
	return bitbucketSortedUnique(values)
}

func bitbucketQueryParameters(operation bitbucketLockedSourceOperation) []string {
	values := make([]string, 0, len(operation.SourceOperation.Parameters))
	for _, parameter := range operation.SourceOperation.Parameters {
		if parameter.In == "query" && parameter.Name != "" {
			values = append(values, parameter.Name)
		}
	}
	return bitbucketSortedUnique(values)
}

func bitbucketMediaTypes(operation bitbucketLockedSourceOperation) ([]string, []string) {
	request := make([]string, 0)
	if operation.SourceOperation.RequestBody != nil {
		for mediaType := range operation.SourceOperation.RequestBody.Content {
			request = append(request, mediaType)
		}
	}
	response := make([]string, 0)
	for status, result := range operation.SourceOperation.Responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		for mediaType := range result.Content {
			response = append(response, mediaType)
		}
	}
	return bitbucketSortedUnique(request), bitbucketSortedUnique(response)
}

func bitbucketBinarySignals(operation bitbucketLockedSourceOperation) []string {
	if signal, ok := bitbucketBinaryDownloadEvidence[operation.ID]; ok {
		return []string{signal}
	}
	if signal, ok := bitbucketBinaryUploadEvidence[operation.ID]; ok {
		return []string{signal}
	}
	return []string{}
}

// bitbucketPaginationEvidenceForOperation identifies pagination from the
// retained provider contract rather than from a generated schema name. A
// successful response must resolve to an object with both a string next link
// and an array of values; request-side page controls are retained as evidence,
// but are not required because the next link itself can be a valid continuation
// contract.
func bitbucketPaginationEvidenceForOperation(lock bitbucketSourceLaneLock, operation bitbucketLockedSourceOperation) bitbucketPaginationEvidence {
	refs := bitbucketSuccessfulResponseSchemaRefs(operation)
	paginated := make([]string, 0, len(refs))
	for _, ref := range refs {
		if bitbucketSourceSchemaHasContinuation(lock, ref) {
			paginated = append(paginated, ref)
		}
	}
	paginated = bitbucketSortedUnique(paginated)
	if len(paginated) == 0 {
		return bitbucketPaginationEvidence{
			ResponseSchemaRefs:     []string{},
			RequestQueryParameters: []string{},
			ContinuationFields:     []string{},
		}
	}
	return bitbucketPaginationEvidence{
		ResponseSchemaRefs:     paginated,
		RequestQueryParameters: bitbucketPaginationQueryParameters(operation),
		ContinuationFields:     []string{"next", "values"},
	}
}

func bitbucketSuccessfulResponseSchemaRefs(operation bitbucketLockedSourceOperation) []string {
	refs := make([]string, 0)
	for status, result := range operation.SourceOperation.Responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		for _, media := range result.Content {
			refs = append(refs, bitbucketJSONReferences(media.Schema)...)
		}
	}
	return bitbucketSortedUnique(refs)
}

func bitbucketSourceSchemaHasContinuation(lock bitbucketSourceLaneLock, ref string) bool {
	schema, ok := bitbucketSourceSchema(lock, ref)
	if !ok {
		return false
	}
	var contract struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(schema, &contract); err != nil {
		return false
	}
	return bitbucketSchemaPropertyHasType(contract.Properties["next"], "string") &&
		bitbucketSchemaPropertyHasType(contract.Properties["values"], "array")
}

func bitbucketSourceSchema(lock bitbucketSourceLaneLock, ref string) (json.RawMessage, bool) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, false
	}
	name := strings.TrimPrefix(ref, prefix)
	if name == "" || strings.Contains(name, "/") {
		return nil, false
	}
	name = strings.ReplaceAll(strings.ReplaceAll(name, "~1", "/"), "~0", "~")
	schema, ok := lock.SourceContract.Components.Schemas[name]
	return schema, ok
}

func bitbucketSchemaPropertyHasType(raw json.RawMessage, want string) bool {
	if len(raw) == 0 {
		return false
	}
	var property struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &property); err != nil {
		return false
	}
	return property.Type == want
}

func bitbucketPaginationQueryParameters(operation bitbucketLockedSourceOperation) []string {
	values := make([]string, 0)
	for _, parameter := range operation.SourceOperation.Parameters {
		if parameter.In != "query" || parameter.Name == "" {
			continue
		}
		fact := strings.ToLower(parameter.Name + " " + parameter.Description)
		if strings.Contains(fact, "page") || strings.Contains(fact, "cursor") || strings.Contains(fact, "offset") || strings.Contains(fact, "continuation") {
			values = append(values, parameter.Name)
		}
	}
	return bitbucketSortedUnique(values)
}

func bitbucketJSONReferences(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	refs := make([]string, 0)
	var visit func(any)
	visit = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			if ref, ok := typed["$ref"].(string); ok {
				refs = append(refs, ref)
			}
			for _, child := range typed {
				visit(child)
			}
		case []any:
			for _, child := range typed {
				visit(child)
			}
		}
	}
	visit(value)
	return refs
}

func bitbucketEventCursorState(operation bitbucketLockedSourceOperation) string {
	if _, ok := bitbucketWebhookDeliveryEvidence[operation.ID]; ok {
		return "webhook_subscription_delivery"
	}
	if _, ok := bitbucketWebhookCatalogIDs[operation.ID]; ok {
		return "webhook_event_catalog"
	}
	return "not_declared"
}

func bitbucketOperationSemantics(operation bitbucketLockedSourceOperation) string {
	switch bitbucketSummaryAction(operation) {
	case "get", "list", "search", "compare", "retrieve", "check":
		return "read_candidate"
	case "delete", "update", "create", "add", "remove", "unapprove", "approve", "watch", "set", "upload", "stop", "run", "resolve", "request", "reopen", "merge", "fork", "decline", "bulk":
		return "mutation_candidate"
	default:
		return "not_classified"
	}
}

// bitbucketSummaryAction is an intentionally small source-language classifier.
// It consumes only the provider's operation summary and makes an unknown verb
// explicit so a new source operation cannot silently become a read or write.
func bitbucketSummaryAction(operation bitbucketLockedSourceOperation) string {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(operation.SourceOperation.Summary)))
	if len(words) == 0 {
		return ""
	}
	return strings.Trim(words[0], ".,:;")
}

func bitbucketMethodPathKey(method, path string) string {
	return method + " " + path
}

func bitbucketSortedUnique(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return slices.Compact(copyValues)
}

func equalOptionalBool(left, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func equalBitbucketLaneCounts(got, want map[string]int) bool {
	if len(got) != len(want) {
		return false
	}
	for disposition, count := range want {
		if got[disposition] != count {
			return false
		}
	}
	return true
}

func cloneBitbucketSourceLaneMatrix(matrix bitbucketSourceLaneMatrix) bitbucketSourceLaneMatrix {
	clone := matrix
	clone.SourceOperations = append([]bitbucketSourceLaneMatrixRow(nil), matrix.SourceOperations...)
	for index := range clone.SourceOperations {
		clone.SourceOperations[index].Lanes = cloneBitbucketLaneCells(matrix.SourceOperations[index].Lanes)
	}
	clone.SourceBoundaryReconciliation.CrosswalkOnlySourceIdentity = append([]bitbucketCrosswalkBoundaryIdentity(nil), matrix.SourceBoundaryReconciliation.CrosswalkOnlySourceIdentity...)
	return clone
}

func cloneBitbucketLaneCells(cells map[string]bitbucketSourceLaneMatrixCell) map[string]bitbucketSourceLaneMatrixCell {
	clone := make(map[string]bitbucketSourceLaneMatrixCell, len(cells))
	for lane, cell := range cells {
		clone[lane] = cell
	}
	return clone
}
