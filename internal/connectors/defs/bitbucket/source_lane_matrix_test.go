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

// These two POST operations are source-documented list reads. They are kept
// separate from mutations so HTTP method alone never decides a Track A lane.
var bitbucketPostReadEvidence = map[string]string{
	"bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/commits":            "List commits with include/exclude",
	"bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/commits/{revision}": "List commits for revision using include/exclude",
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
	SchemaVersion                int                                   `json:"schema_version"`
	Connector                    string                                `json:"connector"`
	Lanes                        []string                              `json:"lanes"`
	SourceLock                   bitbucketSourceLaneMatrixSourceLock   `json:"source_lock"`
	SourceBoundaryReconciliation bitbucketSourceBoundaryReconciliation `json:"source_boundary_reconciliation"`
	SourceOperations             []bitbucketSourceLaneMatrixRow        `json:"source_operations"`
	Summary                      bitbucketSourceLaneMatrixSummary      `json:"summary"`
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
		State              string   `json:"state"`
		ResponseSchemaRefs []string `json:"response_schema_refs"`
	} `json:"pagination"`
	EventCursor struct {
		State string `json:"state"`
	} `json:"event_cursor"`
	OperationSemantics struct {
		State string `json:"state"`
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
	Name string `json:"name"`
	In   string `json:"in"`
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
		if err := validateBitbucketSourceFacts(row.SourceFacts, operation); err != nil {
			return fmt.Errorf("source facts %q: %w", row.SourceID, err)
		}
		for _, lane := range bitbucketSourceLaneNames {
			cell, ok := row.Lanes[lane]
			if !ok {
				return fmt.Errorf("missing lane cell: %s %s", row.SourceID, lane)
			}
			if err := validateBitbucketLaneCell(row.SourceID, lane, cell, operation); err != nil {
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

func validateBitbucketSourceFacts(facts bitbucketSourceLaneMatrixSourceFacts, operation bitbucketLockedSourceOperation) error {
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
	paginationRefs := bitbucketPaginatedResponseRefs(operation)
	wantPagination := "not_declared"
	if len(paginationRefs) > 0 {
		wantPagination = "declared"
	}
	if facts.Pagination.State != wantPagination || !slices.Equal(facts.Pagination.ResponseSchemaRefs, paginationRefs) {
		return fmt.Errorf("pagination facts drift")
	}
	if facts.EventCursor.State != bitbucketEventCursorState(operation) || facts.OperationSemantics.State != bitbucketOperationSemantics(operation) {
		return fmt.Errorf("event/cursor or operation semantics facts drift")
	}
	if err := validateBitbucketCandidateEvidence(operation); err != nil {
		return err
	}
	return nil
}

func validateBitbucketCandidateEvidence(operation bitbucketLockedSourceOperation) error {
	sourceText := operation.SourceOperation.Summary + "\n" + operation.SourceOperation.Description
	for _, evidence := range []map[string]string{
		bitbucketPostReadEvidence,
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

func validateBitbucketLaneCell(sourceID, lane string, cell bitbucketSourceLaneMatrixCell, operation bitbucketLockedSourceOperation) error {
	wantApplicable, wantDisposition := bitbucketExpectedLane(operation, lane)
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
	if err := validateBitbucketLaneMapping(lane, cell.Mapping, operation); err != nil {
		return fmt.Errorf("mapping evidence %s %s: %w", sourceID, lane, err)
	}
	return nil
}

func validateBitbucketLaneMapping(lane string, raw json.RawMessage, operation bitbucketLockedSourceOperation) error {
	switch lane {
	case "direct_read":
		var mapping struct {
			SourceFact struct {
				Method          string `json:"method"`
				Classification  string `json:"classification"`
				SummaryContains string `json:"summary_contains"`
			} `json:"source_fact"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		wantSummary := bitbucketPostReadEvidence[operation.ID]
		if mapping.SourceFact.Method != operation.Method || mapping.SourceFact.Classification != "read_candidate" || mapping.SourceFact.SummaryContains != wantSummary || strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("direct-read source mapping drift")
		}
	case "direct_write":
		var mapping struct {
			SourceFact struct {
				Method         string `json:"method"`
				Classification string `json:"classification"`
			} `json:"source_fact"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.SourceFact.Method != operation.Method || mapping.SourceFact.Classification != "mutation_candidate" || strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("direct-write source mapping drift")
		}
	case "binary_download":
		return validateBitbucketBinaryLaneMapping(raw, bitbucketBinaryDownloadEvidence[operation.ID])
	case "binary_upload":
		return validateBitbucketBinaryLaneMapping(raw, bitbucketBinaryUploadEvidence[operation.ID])
	case "etl":
		var mapping struct {
			PaginationState    string   `json:"pagination_state"`
			ResponseSchemaRefs []string `json:"response_schema_refs"`
			RuntimeClaim       string   `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.PaginationState != "declared" || !slices.Equal(mapping.ResponseSchemaRefs, bitbucketPaginatedResponseRefs(operation)) || strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("ETL source mapping drift")
		}
	case "reverse_etl":
		var mapping struct {
			SourceFact struct {
				Method         string `json:"method"`
				Classification string `json:"classification"`
			} `json:"source_fact"`
			RequiredFlow string `json:"required_flow"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.SourceFact.Method != operation.Method || mapping.SourceFact.Classification != "mutation_candidate" || strings.TrimSpace(mapping.RequiredFlow) == "" || strings.TrimSpace(mapping.RuntimeClaim) == "" {
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

func bitbucketExpectedLane(operation bitbucketLockedSourceOperation, lane string) (string, string) {
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
		applicable = len(bitbucketPaginatedResponseRefs(operation)) > 0
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
	if operation.Method == "GET" {
		return true
	}
	evidence, ok := bitbucketPostReadEvidence[operation.ID]
	return ok && strings.Contains(operation.SourceOperation.Summary, evidence)
}

func bitbucketIsMutationCandidate(operation bitbucketLockedSourceOperation) bool {
	if operation.Method == "DELETE" || operation.Method == "PUT" {
		return true
	}
	return operation.Method == "POST" && !bitbucketIsDirectReadCandidate(operation)
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

func bitbucketPaginatedResponseRefs(operation bitbucketLockedSourceOperation) []string {
	refs := make([]string, 0)
	for status, result := range operation.SourceOperation.Responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		for _, media := range result.Content {
			refs = append(refs, bitbucketJSONReferences(media.Schema)...)
		}
	}
	paginated := make([]string, 0, len(refs))
	for _, ref := range refs {
		if strings.HasPrefix(ref, "#/components/schemas/paginated_") {
			paginated = append(paginated, ref)
		}
	}
	return bitbucketSortedUnique(paginated)
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
	if bitbucketIsDirectReadCandidate(operation) {
		return "read_candidate"
	}
	if bitbucketIsMutationCandidate(operation) {
		return "mutation_candidate"
	}
	return "not_classified"
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
