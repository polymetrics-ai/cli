package gitlab

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	gitLabSourceLaneMatrixPath = "sources/gitlab-source-lane-matrix.json"
	gitLabSourceLockPath       = "sources/gitlab-operation-source-lock.json"
	gitLabBinaryLockPath       = "sources/gitlab-binary-operation-source-lock.json"
	gitLabCrosswalkPath        = "sources/gitlab-operation-crosswalk.json"
	gitLabDescriptorPath       = "sources/gitlab-operation-descriptor.json"
	gitLabRetainedArtifacts    = "sources/gitlab-retained-artifacts.json"
	gitLabStreamsPath          = "streams.json"
	gitLabSnapshotCommit       = "dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db"
	gitLabRuntimeClaim         = "Source-backed mapping only; no runtime execution, certification, or availability proof is claimed."
)

var gitLabLanes = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

var gitLabExpectedCounts = map[string]map[string]int{
	"direct_read":     {"mapped_unproven": 747, "not_applicable": 1007},
	"direct_write":    {"mapped_unproven": 1004, "not_applicable": 750},
	"binary_download": {"mapped_unproven": 1, "not_applicable": 1753},
	"binary_upload":   {"mapped_unproven": 46, "not_applicable": 1708},
	"etl":             {"mapped_unproven": 255, "not_applicable": 1499},
	"reverse_etl":     {"mapped_unproven": 1004, "not_applicable": 750},
	"sync_transport":  {"missing_foundation": 3, "not_applicable": 1751},
}

var gitLabLegacyStreams = map[string]map[string]string{
	"gitlab.rest.getApiV4Projects": {"stream": "projects", "path": "/projects"},
	"gitlab.rest.getApiV4Groups":   {"stream": "groups", "path": "/groups"},
	"gitlab.rest.getApiV4Users":    {"stream": "users", "path": "/users"},
	"gitlab.rest.getApiV4Issues":   {"stream": "issues", "path": "/issues"},
}

var gitLabPathRestrictionSpecs = []map[string]string{
	{
		"record_id":              "gitlab.required-path-parameter.postApiV4GroupsIdDashEpicsEpicIidIssuesIssueId",
		"source_id":              "gitlab.rest.postApiV4GroupsIdDashEpicsEpicIidIssuesIssueId",
		"descriptor_location":    "paths[\"/api/v4/groups/{id}/(-/)epics/{epic_iid}/issues/{epic_issue_id}\"].post.request.path",
		"missing_placeholder":    "epic_issue_id",
		"source_location_suffix": ".post",
	},
	{
		"record_id":              "gitlab.required-path-parameter.getApiV4JobsIdSbomScansSbomScanId",
		"source_id":              "gitlab.rest.getApiV4JobsIdSbomScansSbomScanId",
		"descriptor_location":    "paths[\"/api/v4/jobs/{id}/sbom_scans/{sbom_digest}\"].get.request.path",
		"missing_placeholder":    "sbom_digest",
		"source_location_suffix": ".get",
	},
}

func TestGitLabSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadGitLabObject(t, gitLabSourceLaneMatrixPath)
	lock := loadGitLabObject(t, gitLabSourceLockPath)
	binaryLock := loadGitLabObject(t, gitLabBinaryLockPath)
	crosswalk := loadGitLabObject(t, gitLabCrosswalkPath)
	descriptor := loadGitLabObject(t, gitLabDescriptorPath)
	retainedArtifacts := loadGitLabObject(t, gitLabRetainedArtifacts)
	streams := loadGitLabObject(t, gitLabStreamsPath)
	if err := validateGitLabSourceLaneMatrix(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err != nil {
		t.Fatalf("validate GitLab source lane matrix: %v", err)
	}

	t.Run("rejects hidden source row", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		rows := mustGitLabArray(t, broken["source_operations"])
		broken["source_operations"] = rows[1:]
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "source rows matrix=") {
			t.Fatalf("hidden-row validation error = %v, want source rows matrix", err)
		}
	})

	t.Run("rejects missing lane cell", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		rows := mustGitLabArray(t, broken["source_operations"])
		delete(mustGitLabObject(t, rows[0])["lanes"].(map[string]any), "etl")
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("missing-cell validation error = %v, want lane cells", err)
		}
	})

	t.Run("rejects invalid mutation disposition", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		row := gitLabMatrixRow(t, broken, "gitlab.rest.deleteApiV4AdminActiveContextDeadQueue")
		mustGitLabObject(t, mustGitLabObject(t, row["lanes"])["direct_write"])["disposition"] = "not_applicable"
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("mutation-disposition validation error = %v, want lane cells", err)
		}
	})

	t.Run("rejects crosswalk boundary drop", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		boundary := mustGitLabObject(t, broken["source_boundary_reconciliation"])
		records := mustGitLabArray(t, boundary["crosswalk_only_source_identities"])
		boundary["crosswalk_only_source_identities"] = records[1:]
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "crosswalk boundary") {
			t.Fatalf("boundary validation error = %v, want crosswalk boundary", err)
		}
	})

	t.Run("rejects executable promotion", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		row := gitLabMatrixRow(t, broken, "gitlab.rest.getApiV4Projects")
		mustGitLabObject(t, mustGitLabObject(t, row["lanes"])["direct_read"])["disposition"] = "implemented"
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("executable-promotion validation error = %v, want lane cells", err)
		}
	})
}

func validateGitLabSourceLaneMatrix(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams map[string]any) error {
	if numberAt(matrix, "schema_version") != 1 || stringAt(matrix, "connector") != "gitlab" {
		return fmt.Errorf("matrix identity drift")
	}
	if !reflect.DeepEqual(stringSlice(matrix["lanes"]), gitLabLanes) {
		return fmt.Errorf("lane order drift")
	}
	if err := validateGitLabSnapshotBinding(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts); err != nil {
		return err
	}

	primary := mustMapSlice(objectAt(lock, "rest")["operations"])
	descriptors := mustMapSlice(descriptor["operations"])
	descriptorByID := make(map[string]map[string]any, len(descriptors))
	for _, operation := range descriptors {
		id := stringAt(operation, "operation_id")
		if _, exists := descriptorByID[id]; exists {
			return fmt.Errorf("duplicate descriptor operation ID %q", id)
		}
		descriptorByID[id] = operation
	}

	primaryByID := make(map[string]map[string]any, len(primary))
	for _, operation := range primary {
		id := stringAt(operation, "id")
		if _, exists := primaryByID[id]; exists {
			return fmt.Errorf("duplicate primary source lock ID %q", id)
		}
		if descriptorByID[id] == nil {
			return fmt.Errorf("primary source lock ID %q has no descriptor", id)
		}
		primaryByID[id] = operation
	}
	if numberAt(objectAt(lock, "counts"), "total") != 1752 || len(primaryByID) != 1752 || len(descriptorByID) != 1752 {
		return fmt.Errorf("primary source lock or descriptor denominator drift")
	}
	if err := validateGitLabPaginationReconciliation(matrix, primaryByID, descriptorByID); err != nil {
		return err
	}

	supplemental := gitLabSupplementalOperations(binaryLock)
	if numberAt(objectAt(binaryLock, "counts"), "total") != 2 || len(supplemental) != 2 {
		return fmt.Errorf("supplemental binary source lock denominator drift")
	}

	if err := validateGitLabCrosswalkBoundary(matrix, crosswalk, primaryByID, supplemental); err != nil {
		return err
	}
	if err := validateGitLabMappingRestrictions(matrix, lock, descriptor, primaryByID); err != nil {
		return err
	}
	if err := validateGitLabFoundationAtlas(matrix); err != nil {
		return err
	}
	if err := validateGitLabLegacyStreamBacklinks(matrix, primaryByID, descriptorByID, streams); err != nil {
		return err
	}

	rows := mustMapSlice(matrix["source_operations"])
	if len(rows) != 1754 {
		return fmt.Errorf("source rows matrix=%d, want 1754", len(rows))
	}
	counts := make(map[string]map[string]int, len(gitLabLanes))
	seen := make(map[string]struct{}, len(rows))
	rest := objectAt(lock, "rest")
	for _, row := range rows {
		id := stringAt(row, "source_id")
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("duplicate matrix source ID %q", id)
		}
		seen[id] = struct{}{}

		var expectedFacts map[string]any
		expectedOrigin := "rendered_reference_supplement"
		if strings.HasPrefix(id, "gitlab.rest.") {
			operationID := strings.TrimPrefix(id, "gitlab.rest.")
			operation := primaryByID[operationID]
			if operation == nil {
				return fmt.Errorf("matrix primary source ID %q is absent from source lock", id)
			}
			expectedFacts = expectedGitLabPrimaryFacts(operation, descriptorByID[operationID], rest)
			expectedOrigin = "primary_openapi"
		} else {
			operation := supplemental[id]
			if operation == nil {
				return fmt.Errorf("matrix supplemental source ID %q is absent from binary source lock", id)
			}
			expectedFacts = expectedGitLabSupplementalFacts(operation)
		}
		if stringAt(row, "source_origin") != expectedOrigin {
			return fmt.Errorf("source origin %q drift", id)
		}
		if !sameGitLabJSON(objectAt(row, "source_facts"), expectedFacts) {
			return fmt.Errorf("source facts %q drift", id)
		}

		lanes := objectAt(row, "lanes")
		if len(lanes) != len(gitLabLanes) {
			return fmt.Errorf("lane cells %s=%d, want %d", id, len(lanes), len(gitLabLanes))
		}
		wantLanes := expectedGitLabLanes(id, expectedFacts)
		if !sameGitLabJSON(lanes, wantLanes) {
			return fmt.Errorf("lane cells %q drift", id)
		}
		for _, lane := range gitLabLanes {
			cell := objectAt(lanes, lane)
			disposition := stringAt(cell, "disposition")
			if counts[lane] == nil {
				counts[lane] = make(map[string]int)
			}
			counts[lane][disposition]++
		}
	}
	if len(seen) != 1754 {
		return fmt.Errorf("matrix source rows retained=%d, want 1754", len(seen))
	}
	for id := range primaryByID {
		if _, exists := seen["gitlab.rest."+id]; !exists {
			return fmt.Errorf("primary source ID hidden from matrix %q", id)
		}
	}
	for id := range supplemental {
		if _, exists := seen[id]; !exists {
			return fmt.Errorf("supplemental source ID hidden from matrix %q", id)
		}
	}

	summary := objectAt(matrix, "summary")
	if numberAt(summary, "source_rows") != 1754 || numberAt(summary, "source_rows_with_all_lanes") != 1754 || numberAt(summary, "total_lane_cells") != 12278 {
		return fmt.Errorf("matrix summary source accounting drift")
	}
	for _, lane := range gitLabLanes {
		if !equalGitLabCounts(counts[lane], gitLabExpectedCounts[lane]) {
			return fmt.Errorf("computed %s counts=%v, want %v", lane, counts[lane], gitLabExpectedCounts[lane])
		}
		if !equalGitLabCounts(numberMap(objectAt(summary, "lane_counts")[lane]), counts[lane]) {
			return fmt.Errorf("summary %s lane counts drift", lane)
		}
	}
	return nil
}

func validateGitLabSnapshotBinding(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts map[string]any) error {
	snapshot := objectAt(matrix, "source_snapshot")
	if stringAt(snapshot, "source_snapshot_ref") != "fm/cli-top100-declaration-batch-r1" || stringAt(snapshot, "source_snapshot_commit") != gitLabSnapshotCommit || stringAt(snapshot, "materialization") != "git_archive_byte_identical" {
		return fmt.Errorf("source snapshot identity drift")
	}
	files := mustMapSlice(snapshot["retained_files"])
	wantFiles := []map[string]any{
		{"path": gitLabSourceLockPath, "git_blob_sha1": "d874f0d462bc054d3065a41e32b0bb1b1675a84c", "bytes": float64(5783052)},
		{"path": gitLabCrosswalkPath, "git_blob_sha1": "20353c196663b425280bc0f63ee09dddb5bdc913", "bytes": float64(3004301)},
		{"path": gitLabDescriptorPath, "git_blob_sha1": "19d74be07d6862ef2492c9fb6ff9e6467b67df96", "bytes": float64(18119458)},
		{"path": gitLabBinaryLockPath, "git_blob_sha1": "ee538b7ce20912ea13e95d87854b0c014928231d", "bytes": float64(3058)},
		{"path": gitLabRetainedArtifacts, "git_blob_sha1": "1d7104c91be20f4ab73920389e681c7a0fb6bc56", "bytes": float64(678)},
		{"path": "sources/artifacts/53244a720b8509536290e0058c946a246817c775c797df36f4c9aa1225fdf0a4.artifact", "git_blob_sha1": "2328f5833aa166b7f44fae2ce1cebad2528c163f", "bytes": float64(102560)},
		{"path": "sources/artifacts/f59c93194c095d0e925a5751a08eb7a2176a26c6b5f38bda52f805154219d0f0.artifact", "git_blob_sha1": "a3a5ba1e5121258588101f1f9c33ebf226d7909b", "bytes": float64(103189)},
	}
	if len(files) != len(wantFiles) {
		return fmt.Errorf("retained source file count=%d, want %d", len(files), len(wantFiles))
	}
	for i, want := range wantFiles {
		got := files[i]
		if !sameGitLabJSON(got, want) {
			return fmt.Errorf("retained source file metadata drift at index %d", i)
		}
		contents, err := os.ReadFile(stringAt(got, "path"))
		if err != nil {
			return fmt.Errorf("read retained source file %q: %w", stringAt(got, "path"), err)
		}
		if len(contents) != numberAt(got, "bytes") || gitLabBlobSHA1(contents) != stringAt(got, "git_blob_sha1") {
			return fmt.Errorf("retained source file byte/blob identity drift %q", stringAt(got, "path"))
		}
	}
	if stringAt(lock, "connector") != "gitlab" || stringAt(binaryLock, "connector") != "gitlab" || stringAt(crosswalk, "connector") != "gitlab" || numberAt(descriptor, "schema_version") < 1 || stringAt(retainedArtifacts, "connector") != "gitlab" {
		return fmt.Errorf("retained source artifact identity drift")
	}
	for _, artifact := range mustMapSlice(retainedArtifacts["artifacts"]) {
		sha := stringAt(artifact, "sha256")
		contents, err := os.ReadFile("sources/artifacts/" + sha + ".artifact")
		if err != nil {
			return fmt.Errorf("read retained artifact %q: %w", sha, err)
		}
		digest := sha256.Sum256(contents)
		if hex.EncodeToString(digest[:]) != sha || len(contents) != numberAt(artifact, "bytes") {
			return fmt.Errorf("retained artifact digest/bytes drift %q", sha)
		}
	}
	return nil
}

func validateGitLabCrosswalkBoundary(matrix, crosswalk map[string]any, primary map[string]map[string]any, supplemental map[string]map[string]any) error {
	if stringAt(crosswalk, "source_lock") != gitLabSourceLockPath || len(mustMapSlice(crosswalk["source_operations"])) != 1755 {
		return fmt.Errorf("crosswalk identity or denominator drift")
	}
	boundary := objectAt(matrix, "source_boundary_reconciliation")
	if stringAt(boundary, "identity") != "source_id" || numberAt(boundary, "primary_source_rows") != 1752 || numberAt(boundary, "supplemental_binary_source_rows") != 2 || numberAt(boundary, "retained_source_rows") != 1754 || numberAt(boundary, "crosswalk_rows") != 1755 || numberAt(boundary, "crosswalk_only_rows") != 3 || numberAt(boundary, "supplemental_not_in_primary_crosswalk_rows") != 2 {
		return fmt.Errorf("crosswalk boundary accounting drift")
	}
	primaryIDs := make(map[string]struct{}, len(primary))
	for id := range primary {
		primaryIDs["gitlab.rest."+id] = struct{}{}
	}
	for id := range supplemental {
		if _, inPrimary := primaryIDs[id]; inPrimary {
			return fmt.Errorf("supplemental source ID %q unexpectedly appears in primary source namespace", id)
		}
	}

	want := make([]any, 0, 3)
	for _, entry := range mustMapSlice(crosswalk["source_operations"]) {
		id := stringAt(entry, "source_id")
		if _, present := primaryIDs[id]; present {
			continue
		}
		want = append(want, map[string]any{
			"source_id":       id,
			"operation_id":    stringAt(entry, "operation_id"),
			"method":          stringAt(entry, "method"),
			"path":            stringAt(entry, "path"),
			"source_location": stringAt(entry, "source_location"),
			"disposition":     "not_source_row",
			"reason":          "Present in the retained GitLab crosswalk but absent from the immutable primary and supplemental source locks; this is boundary evidence, not a matrix source row.",
			"crosswalk_entry": entry,
		})
	}
	if len(want) != 3 {
		return fmt.Errorf("crosswalk-minus-primary source accounting=%d, want 3", len(want))
	}
	if !sameGitLabJSON(boundary["crosswalk_only_source_identities"], want) {
		return fmt.Errorf("crosswalk boundary identities drift")
	}
	return nil
}

func validateGitLabMappingRestrictions(matrix, lock, descriptor map[string]any, primary map[string]map[string]any) error {
	want := expectedGitLabMappingRestrictions(lock, descriptor, primary)
	if !sameGitLabJSON(matrix["mapping_restrictions"], want) {
		return fmt.Errorf("source-visible mapping restrictions drift")
	}
	return nil
}

func validateGitLabFoundationAtlas(matrix map[string]any) error {
	atlas := objectAt(matrix, "foundation_atlas")
	if stringAt(atlas, "consulted_snapshot_ref") != "fm/cli-top100-declaration-batch-r1" || stringAt(atlas, "consulted_snapshot_commit") != gitLabSnapshotCommit || stringAt(atlas, "catalog_path") != "docs/connector-canon/foundations/catalog.json" || stringAt(atlas, "usage") != "authoring_only_not_a_runtime_loader" {
		return fmt.Errorf("Foundation Atlas provenance drift")
	}
	wantReuse := []any{"source.retention-import.v1", "source.projection-admission.v1", "runtime.direct-execution.v1", "warehouse.stage-etl.v1", "warehouse.reverse-etl.v1", "transport.sync-contract.v1"}
	if !sameGitLabJSON(atlas["consulted_capabilities"], wantReuse) {
		return fmt.Errorf("Foundation Atlas capability lookup drift")
	}
	gap := objectAt(atlas, "sync_actual_gap")
	if stringAt(gap, "gap_id") != "gitlab-inbound-webhook-source-executor-r1" || stringAt(gap, "lane") != "sync_transport" || stringAt(gap, "consulted_atlas_id") != "transport.sync-contract.v1" || stringAt(gap, "status") != "recorded_only_requires_captain_approval_before_implementation" || !sameGitLabJSON(gap["source_ids"], []any{"gitlab.rest.postApiV4GroupsIdHooks", "gitlab.rest.postApiV4Hooks", "gitlab.rest.postApiV4ProjectsIdHooks"}) || strings.TrimSpace(stringAt(gap, "missing_capability")) == "" || strings.TrimSpace(stringAt(gap, "why_existing_capability_is_insufficient")) == "" || strings.TrimSpace(stringAt(gap, "proof_test_idea")) == "" {
		return fmt.Errorf("Foundation Atlas sync-gap record drift")
	}
	return nil
}

func validateGitLabPaginationReconciliation(matrix map[string]any, primary, descriptors map[string]map[string]any) error {
	kinds := make(map[string]int)
	for operationID, operation := range primary {
		facts := gitLabPaginationFacts(operation, descriptors[operationID])
		kinds[stringAt(facts, "kind")]++
	}
	if kinds["page_per_page"] != 253 || kinds["page_token"] != 2 || kinds["page_per_page"]+kinds["page_token"] != 255 {
		return fmt.Errorf("source pagination candidate accounting=%v, want page_per_page=253 page_token=2", kinds)
	}
	reconciliation := objectAt(matrix, "source_paging_reconciliation")
	if stringAt(reconciliation, "criterion") != "GET with explicit source query controls page+per_page or page_token" || numberAt(reconciliation, "page_per_page_candidates") != 253 || numberAt(reconciliation, "page_token_candidates") != 2 || numberAt(reconciliation, "total_candidates") != 255 {
		return fmt.Errorf("matrix source pagination reconciliation drift")
	}
	return nil
}

func validateGitLabLegacyStreamBacklinks(matrix map[string]any, primary, descriptors map[string]map[string]any, streams map[string]any) error {
	streamPaths := make(map[string]string)
	for _, stream := range mustMapSlice(streams["streams"]) {
		streamPaths[stringAt(stream, "name")] = stringAt(stream, "path")
	}
	for id, want := range gitLabLegacyStreams {
		op := primary[strings.TrimPrefix(id, "gitlab.rest.")]
		desc := descriptors[strings.TrimPrefix(id, "gitlab.rest.")]
		if op == nil || desc == nil || stringAt(gitLabPaginationFacts(op, desc), "kind") == "not_documented" {
			return fmt.Errorf("legacy stream source lookup drift %q", id)
		}
		if stringAt(op, "method") != "GET" || streamPaths[want["stream"]] != want["path"] {
			return fmt.Errorf("legacy stream source/definition drift %q", id)
		}
	}
	return nil
}

func expectedGitLabPrimaryFacts(operation, descriptor, rest map[string]any) map[string]any {
	id := stringAt(operation, "id")
	params := mapSliceOrEmpty(objectAt(operation, "source_operation")["parameters"])
	return map[string]any{
		"source_kind":    "primary_openapi",
		"source_lock":    gitLabSourceLockPath,
		"source_id":      "gitlab.rest." + id,
		"operation_id":   id,
		"protocol":       stringAt(operation, "protocol"),
		"method":         stringAt(operation, "method"),
		"path":           stringAt(operation, "path"),
		"mapping_path":   stringAt(descriptor, "mapping_path"),
		"deprecated":     operation["deprecated"],
		"source_summary": stringAt(objectAt(operation, "source_operation"), "summary"),
		"citation": map[string]any{
			"url":      stringAt(rest, "source_url"),
			"sha256":   stringAt(rest, "sha256"),
			"bytes":    rest["bytes"],
			"location": stringAt(operation, "source_location"),
		},
		"scope_fanout": map[string]any{
			"path_parameters":  sourceParameterNames(params, "path"),
			"query_parameters": sourceParameterNames(params, "query"),
		},
		"request_media_types":          sourceRequestMediaTypes(operation),
		"success_response_media_types": sourceSuccessResponseMediaTypes(operation),
		"pagination":                   gitLabPaginationFacts(operation, descriptor),
		"binary": map[string]any{
			"request_binary_fields": gitLabBinaryRequestFields(descriptor),
			"download_state":        "not_declared_by_primary_source",
		},
		"event_cursor":                   gitLabEventFacts(descriptor),
		"mapping_restriction_record_ids": gitLabRestrictionIDs("gitlab.rest." + id),
		"crosswalk_state":                "primary_crosswalk_exact",
	}
}

func expectedGitLabSupplementalFacts(operation map[string]any) map[string]any {
	id := stringAt(operation, "id")
	role := "generic_package_file_upload"
	if id == "gitlab.docs.repository_files.raw_download" {
		role = "repository_file_raw_download"
	}
	return map[string]any{
		"source_kind": "rendered_reference_supplement",
		"source_lock": gitLabBinaryLockPath,
		"source_id":   id,
		"protocol":    stringAt(operation, "protocol"),
		"method":      stringAt(operation, "method"),
		"path":        stringAt(operation, "path"),
		"citation": map[string]any{
			"url":      stringAt(operation, "citation_url"),
			"sha256":   operation["document_sha256"],
			"bytes":    operation["document_bytes"],
			"location": stringAt(operation, "source_location"),
		},
		"scope_fanout": map[string]any{
			"path_parameters":  pathTemplateParameters(stringAt(operation, "path")),
			"query_parameters": []any{},
		},
		"request_media_types":          []any{},
		"success_response_media_types": []any{},
		"pagination":                   map[string]any{"kind": "not_documented", "controls": []any{}},
		"binary": map[string]any{
			"rendered_reference_role": role,
		},
		"event_cursor":    map[string]any{"state": "not_documented"},
		"crosswalk_state": "supplemental_source_not_in_primary_crosswalk",
	}
}

func expectedGitLabLanes(id string, facts map[string]any) map[string]any {
	method := stringAt(facts, "method")
	lockPath := stringAt(facts, "source_lock")
	backlink := map[string]any{"kind": "source_lock", "path": lockPath, "source_id": id}
	lanes := make(map[string]any, len(gitLabLanes))
	if method == "GET" {
		lanes["direct_read"] = gitLabMappedCell("Locked GET source row; source-backed direct-read candidate only.", "read_verb_candidate", backlink)
	} else {
		lanes["direct_read"] = gitLabNotApplicableCell("Locked method " + method + " is not a GET direct-read candidate.")
	}
	if gitLabIsMutation(method) {
		cell := gitLabMappedCell("Locked "+method+" source row; source-backed mutation candidate only.", "mutation_verb_candidate", backlink)
		lanes["direct_write"] = cell
		lanes["reverse_etl"] = cloneGitLabMap(cell)
	} else {
		lanes["direct_write"] = gitLabNotApplicableCell("Locked method " + method + " is not a provider-mutation candidate.")
		lanes["reverse_etl"] = gitLabNotApplicableCell("Locked method " + method + " is not a provider-mutation candidate.")
	}

	binary := objectAt(facts, "binary")
	if stringAt(binary, "rendered_reference_role") == "repository_file_raw_download" {
		lanes["binary_download"] = gitLabMappedCell("Retained rendered GitLab reference explicitly documents repository-file raw download.", "binary_download_candidate", backlink)
	} else {
		lanes["binary_download"] = gitLabNotApplicableCell("No retained source evidence declares a binary-download candidate.")
	}
	if fields := stringSliceOrEmpty(binary["request_binary_fields"]); len(fields) > 0 {
		lanes["binary_upload"] = gitLabMappedCell("Resolved source request schema declares binary field(s): "+strings.Join(fields, ", ")+".", "binary_upload_candidate", backlink)
	} else if stringAt(binary, "rendered_reference_role") == "generic_package_file_upload" {
		lanes["binary_upload"] = gitLabMappedCell("Retained rendered GitLab reference explicitly documents generic-package file upload.", "binary_upload_candidate", backlink)
	} else {
		lanes["binary_upload"] = gitLabNotApplicableCell("No retained source evidence declares a binary-upload candidate.")
	}

	pagination := objectAt(facts, "pagination")
	if kind := stringAt(pagination, "kind"); kind != "not_documented" {
		mapping := backlink
		if stream, ok := gitLabLegacyStreams[id]; ok {
			mapping = map[string]any{"kind": "existing_stream", "path": gitLabStreamsPath, "stream": stream["stream"], "stream_path": stream["path"]}
		}
		lanes["etl"] = gitLabMappedCell("Source declares explicit pagination controls: "+strings.Join(stringSlice(pagination["controls"]), ", ")+".", "pageable_extractable_collection_candidate", mapping)
	} else {
		lanes["etl"] = gitLabNotApplicableCell("No explicit source pagination controls match the Track A extraction criterion.")
	}

	event := objectAt(facts, "event_cursor")
	if stringAt(event, "state") == "webhook_registration" {
		lanes["sync_transport"] = map[string]any{
			"applicability": "source_candidate",
			"disposition":   "missing_foundation",
			"reason":        "Source documents webhook registration with a required URL and event selectors; the consulted Atlas has no closed inbound GitLab webhook source executor.",
			"mapping": map[string]any{
				"source_fact":         map[string]any{"classification": "webhook_registration_candidate", "source_id": id},
				"definition_backlink": backlink,
				"foundation_gap_id":   "gitlab-inbound-webhook-source-executor-r1",
				"consulted_atlas_id":  "transport.sync-contract.v1",
				"runtime_claim":       "No inbound GitLab webhook receiver or selected source executor is claimed.",
			},
		}
	} else {
		lanes["sync_transport"] = gitLabNotApplicableCell("No retained source fact documents a webhook registration with required URL and event selectors.")
	}
	return lanes
}

func gitLabMappedCell(reason, classification string, backlink map[string]any) map[string]any {
	return map[string]any{
		"applicability": "source_candidate",
		"disposition":   "mapped_unproven",
		"reason":        reason,
		"mapping": map[string]any{
			"source_fact":         map[string]any{"classification": classification},
			"definition_backlink": backlink,
			"runtime_claim":       gitLabRuntimeClaim,
		},
	}
}

func gitLabNotApplicableCell(reason string) map[string]any {
	return map[string]any{"applicability": "not_applicable", "disposition": "not_applicable", "reason": reason}
}

func expectedGitLabMappingRestrictions(lock, descriptor map[string]any, primary map[string]map[string]any) []any {
	rest := objectAt(lock, "rest")
	gaps := mustMapSlice(descriptor["gaps"])
	result := make([]any, 0, len(gitLabPathRestrictionSpecs))
	for _, spec := range gitLabPathRestrictionSpecs {
		operationID := strings.TrimPrefix(spec["source_id"], "gitlab.rest.")
		op := primary[operationID]
		if op == nil || !strings.HasSuffix(stringAt(op, "source_location"), spec["source_location_suffix"]) {
			panic("invalid static GitLab path restriction source ID")
		}
		var descriptorGap map[string]any
		for _, gap := range gaps {
			if stringAt(gap, "foundation") == "cli-malformed-path-parameter-foundation-r1" && stringAt(gap, "location") == spec["descriptor_location"] {
				descriptorGap = gap
				break
			}
		}
		if descriptorGap == nil {
			panic("missing expected static GitLab descriptor gap")
		}
		result = append(result, map[string]any{
			"record_id":             spec["record_id"],
			"state":                 "mapping_restriction",
			"source_id":             spec["source_id"],
			"source_location":       stringAt(op, "source_location"),
			"descriptor_gap":        descriptorGap,
			"missing_placeholder":   spec["missing_placeholder"],
			"rest_path_bridge":      rest["path_bridge"],
			"atlas_lookup":          "source.projection-admission.v1",
			"needed_mapping_repair": "Projection/import mapping must preserve the source path and record the missing required placeholder visibly; it must not synthesize a binding or erase the source row.",
			"status":                "recorded_only_no_shared_code_changed",
		})
	}
	return result
}

func gitLabPaginationFacts(operation, descriptor map[string]any) map[string]any {
	if stringAt(operation, "method") != "GET" {
		return map[string]any{"kind": "not_documented", "controls": []any{}}
	}
	query := mustMapSlice(objectAt(descriptor, "request")["query"])
	names := make(map[string]struct{}, len(query))
	for _, parameter := range query {
		names[stringAt(parameter, "name")] = struct{}{}
	}
	if _, page := names["page"]; page {
		if _, perPage := names["per_page"]; perPage {
			return map[string]any{"kind": "page_per_page", "controls": []any{"page", "per_page"}}
		}
	}
	if _, pageToken := names["page_token"]; pageToken {
		controls := []any{"page_token"}
		if _, maxResults := names["max_results"]; maxResults {
			controls = append(controls, "max_results")
		}
		return map[string]any{"kind": "page_token", "controls": controls}
	}
	return map[string]any{"kind": "not_documented", "controls": []any{}}
}

func gitLabBinaryRequestFields(descriptor map[string]any) []any {
	request := objectAt(descriptor, "request")
	body, exists := request["body"]
	if !exists {
		return []any{}
	}
	fields := make([]string, 0)
	gitLabFindBinaryFields(objectAt(mustGitLabObjectNoTest(body), "schema"), "body", &fields)
	sort.Strings(fields)
	return stringsToAny(fields)
}

func gitLabFindBinaryFields(schema map[string]any, path string, fields *[]string) {
	if stringAt(schema, "format") == "binary" {
		*fields = append(*fields, path)
	}
	if properties, ok := schema["properties"].(map[string]any); ok {
		keys := make([]string, 0, len(properties))
		for name := range properties {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		for _, name := range keys {
			child, ok := properties[name].(map[string]any)
			if ok {
				gitLabFindBinaryFields(child, path+"."+name, fields)
			}
		}
	}
	if items, ok := schema["items"].(map[string]any); ok {
		gitLabFindBinaryFields(items, path+"[]", fields)
	}
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		if values, ok := schema[key].([]any); ok {
			for _, value := range values {
				if child, ok := value.(map[string]any); ok {
					gitLabFindBinaryFields(child, path, fields)
				}
			}
		}
	}
}

func gitLabEventFacts(descriptor map[string]any) map[string]any {
	if !strings.HasSuffix(stringAt(descriptor, "path"), "/hooks") {
		return map[string]any{"state": "not_documented"}
	}
	request := objectAt(descriptor, "request")
	body, exists := request["body"]
	if !exists {
		return map[string]any{"state": "not_documented"}
	}
	bodyMap := mustGitLabObjectNoTest(body)
	schema := objectAt(bodyMap, "schema")
	required := make(map[string]struct{})
	for _, name := range stringSlice(schema["required"]) {
		required[name] = struct{}{}
	}
	if _, urlRequired := required["url"]; !urlRequired {
		return map[string]any{"state": "not_documented"}
	}
	eventFields := make([]string, 0)
	for name := range objectAt(schema, "properties") {
		if strings.HasSuffix(name, "_events") {
			eventFields = append(eventFields, name)
		}
	}
	if len(eventFields) == 0 {
		return map[string]any{"state": "not_documented"}
	}
	sort.Strings(eventFields)
	return map[string]any{"state": "webhook_registration", "url_required": true, "event_selectors": stringsToAny(eventFields)}
}

func gitLabSupplementalOperations(binaryLock map[string]any) map[string]map[string]any {
	result := make(map[string]map[string]any, 2)
	for _, document := range mustMapSlice(objectAt(binaryLock, "rest")["source_documents"]) {
		artifact := objectAt(document, "artifact")
		for _, rawOperation := range mustMapSlice(document["operations"]) {
			operation := cloneGitLabMap(rawOperation)
			operation["document_sha256"] = artifact["sha256"]
			operation["document_bytes"] = artifact["bytes"]
			id := stringAt(operation, "id")
			if _, exists := result[id]; exists {
				panic("duplicate supplemental GitLab source ID")
			}
			result[id] = operation
		}
	}
	return result
}

func sourceParameterNames(parameters []map[string]any, location string) []any {
	names := make([]string, 0)
	for _, parameter := range parameters {
		if stringAt(parameter, "in") == location {
			names = append(names, stringAt(parameter, "name"))
		}
	}
	sort.Strings(names)
	return stringsToAny(names)
}

func sourceRequestMediaTypes(operation map[string]any) []any {
	body, exists := objectAt(operation, "source_operation")["requestBody"]
	if !exists {
		return []any{}
	}
	return sortedObjectKeys(objectAt(mustGitLabObjectNoTest(body), "content"))
}

func sourceSuccessResponseMediaTypes(operation map[string]any) []any {
	media := make(map[string]struct{})
	responses, _ := objectAt(operation, "source_operation")["responses"].(map[string]any)
	for status, rawResponse := range responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		response := mustGitLabObjectNoTest(rawResponse)
		content, _ := response["content"].(map[string]any)
		for _, mediaType := range sortedObjectKeys(content) {
			media[mediaType.(string)] = struct{}{}
		}
	}
	keys := make([]string, 0, len(media))
	for mediaType := range media {
		keys = append(keys, mediaType)
	}
	sort.Strings(keys)
	return stringsToAny(keys)
}

func gitLabRestrictionIDs(sourceID string) []any {
	for _, spec := range gitLabPathRestrictionSpecs {
		if spec["source_id"] == sourceID {
			return []any{spec["record_id"]}
		}
	}
	return []any{}
}

func pathTemplateParameters(path string) []any {
	params := make([]string, 0)
	for remaining := path; ; {
		open := strings.IndexByte(remaining, '{')
		if open < 0 {
			break
		}
		remaining = remaining[open+1:]
		close := strings.IndexByte(remaining, '}')
		if close < 0 {
			break
		}
		params = append(params, remaining[:close])
		remaining = remaining[close+1:]
	}
	return stringsToAny(params)
}

func gitLabIsMutation(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
}

func gitLabMatrixRow(t *testing.T, matrix map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, row := range mustGitLabArray(t, matrix["source_operations"]) {
		object := mustGitLabObject(t, row)
		if stringAt(object, "source_id") == sourceID {
			return object
		}
	}
	t.Fatalf("matrix source row %q not found", sourceID)
	return nil
}

func loadGitLabObject(t *testing.T, path string) map[string]any {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value map[string]any
	if err := json.Unmarshal(contents, &value); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return value
}

func cloneGitLabObject(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	contents, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal clone: %v", err)
	}
	var clone map[string]any
	if err := json.Unmarshal(contents, &clone); err != nil {
		t.Fatalf("unmarshal clone: %v", err)
	}
	return clone
}

func cloneGitLabMap(value map[string]any) map[string]any {
	contents, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(contents, &clone); err != nil {
		panic(err)
	}
	return clone
}

func mustGitLabArray(t *testing.T, value any) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("array type %T", value)
	}
	return array
}

func mustMapSlice(value any) []map[string]any {
	array, ok := value.([]any)
	if !ok {
		panic(fmt.Sprintf("array type %T", value))
	}
	result := make([]map[string]any, len(array))
	for i, item := range array {
		object, ok := item.(map[string]any)
		if !ok {
			panic(fmt.Sprintf("array object type %T", item))
		}
		result[i] = object
	}
	return result
}

func mapSliceOrEmpty(value any) []map[string]any {
	if value == nil {
		return []map[string]any{}
	}
	return mustMapSlice(value)
}

func mustGitLabObject(t *testing.T, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("object type %T", value)
	}
	return object
}

func mustGitLabObjectNoTest(value any) map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("object type %T", value))
	}
	return object
}

func objectAt(value map[string]any, key string) map[string]any {
	object, ok := value[key].(map[string]any)
	if !ok {
		panic(fmt.Sprintf("object %q type %T", key, value[key]))
	}
	return object
}

func stringAt(value map[string]any, key string) string {
	if value[key] == nil {
		return ""
	}
	stringValue, ok := value[key].(string)
	if !ok {
		panic(fmt.Sprintf("string %q type %T", key, value[key]))
	}
	return stringValue
}

func numberAt(value map[string]any, key string) int {
	number, ok := value[key].(float64)
	if !ok {
		panic(fmt.Sprintf("number %q type %T", key, value[key]))
	}
	return int(number)
}

func stringSlice(value any) []string {
	array, ok := value.([]any)
	if !ok {
		panic(fmt.Sprintf("string array type %T", value))
	}
	result := make([]string, len(array))
	for i, item := range array {
		stringValue, ok := item.(string)
		if !ok {
			panic(fmt.Sprintf("string array item type %T", item))
		}
		result[i] = stringValue
	}
	return result
}

func stringSliceOrEmpty(value any) []string {
	if value == nil {
		return []string{}
	}
	return stringSlice(value)
}

func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}

func sortedObjectKeys(value map[string]any) []any {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return stringsToAny(keys)
}

func numberMap(value any) map[string]int {
	object := mustGitLabObjectNoTest(value)
	result := make(map[string]int, len(object))
	for key, raw := range object {
		number, ok := raw.(float64)
		if !ok {
			panic(fmt.Sprintf("number map %q type %T", key, raw))
		}
		result[key] = int(number)
	}
	return result
}

func equalGitLabCounts(got, want map[string]int) bool {
	return reflect.DeepEqual(got, want)
}

func sameGitLabJSON(got, want any) bool {
	return reflect.DeepEqual(got, want)
}

func gitLabBlobSHA1(contents []byte) string {
	prefix := []byte(fmt.Sprintf("blob %d\x00", len(contents)))
	digest := sha1.Sum(append(prefix, contents...))
	return hex.EncodeToString(digest[:])
}
