package notion

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	notionSourceLaneMatrixPath = "sources/notion-source-lane-matrix.json"
	notionSourceLockPath       = "sources/notion-operation-source-lock.json"
	notionCrosswalkPath        = "sources/notion-operation-crosswalk.json"
)

var notionLaneNames = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

// These are manually source-reviewed classifications, not output from a
// connector generator. The test proves every ID remains in the frozen lock.
// A direct read is a single bounded provider read. GET is common but not a
// prerequisite: a source-documented POST query or introspection response can
// also be a bounded direct read when the retained contract proves its read
// semantics and response shape.
var notionGETDirectReadIDs = stringSet(
	"notion.rest.get-self",
	"notion.rest.get-user",
	"notion.rest.get-users",
	"notion.rest.retrieve-a-page",
	"notion.rest.retrieve-a-page-property",
	"notion.rest.retrieve-page-markdown",
	"notion.rest.retrieve-async-task",
	"notion.rest.retrieve-a-block",
	"notion.rest.get-block-children",
	"notion.rest.retrieve-a-data-source",
	"notion.rest.list-data-source-templates",
	"notion.rest.retrieve-database",
	"notion.rest.list-comments",
	"notion.rest.retrieve-comment",
	"notion.rest.list-file-uploads",
	"notion.rest.retrieve-file-upload",
	"notion.rest.list-custom-emojis",
	"notion.rest.list-views",
	"notion.rest.retrieve-a-view",
	"notion.rest.get-view-query-results",
)

var notionSemanticReadNonGETIDs = stringSet(
	"notion.rest.post-database-query",
	"notion.rest.post-search",
	"notion.rest.query-meeting-notes",
	"notion.rest.introspect-token",
)

func notionDirectReadCandidate(sourceID string) bool {
	return notionGETDirectReadIDs[sourceID] || notionSemanticReadNonGETIDs[sourceID]
}

func notionDirectReadClassification(sourceID string) string {
	if notionSemanticReadNonGETIDs[sourceID] {
		return "source_semantic_post_read_candidate"
	}
	return "locked_get_source_candidate"
}

var notionMutationIDs = stringSet(
	"notion.rest.post-page",
	"notion.rest.patch-page",
	"notion.rest.move-page",
	"notion.rest.update-page-markdown",
	"notion.rest.update-a-block",
	"notion.rest.delete-a-block",
	"notion.rest.patch-block-children",
	"notion.rest.update-a-data-source",
	"notion.rest.create-a-database",
	"notion.rest.update-database",
	"notion.rest.create-database",
	"notion.rest.create-a-comment",
	"notion.rest.update-a-comment",
	"notion.rest.delete-a-comment",
	"notion.rest.create-file",
	"notion.rest.upload-file",
	"notion.rest.complete-file-upload",
	"notion.rest.create-view",
	"notion.rest.update-a-view",
	"notion.rest.delete-view",
	"notion.rest.create-view-query",
	"notion.rest.delete-view-query",
	"notion.rest.create-meeting-note",
	"notion.rest.create-a-token",
	"notion.rest.revoke-token",
)

// ETL is source-semantic, not GET-only: the two cursor-body POST queries and
// the bounded meeting-notes collection remain visible. The last has an
// explicit source-mapping restriction because it advertises has_more but no
// continuation input.
var notionETLIDs = stringSet(
	"notion.rest.get-users",
	"notion.rest.retrieve-a-page-property",
	"notion.rest.get-block-children",
	"notion.rest.post-database-query",
	"notion.rest.list-data-source-templates",
	"notion.rest.post-search",
	"notion.rest.list-comments",
	"notion.rest.list-file-uploads",
	"notion.rest.list-custom-emojis",
	"notion.rest.list-views",
	"notion.rest.get-view-query-results",
	"notion.rest.query-meeting-notes",
)

var notionBinaryUploadIDs = stringSet("notion.rest.upload-file")

func TestNotionSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadNotionJSON(t, notionSourceLaneMatrixPath)
	lock := loadNotionJSON(t, notionSourceLockPath)
	if err := validateNotionMatrix(matrix, lock); err != nil {
		t.Fatalf("validate Notion source lane matrix: %v", err)
	}

	t.Run("rejects hidden source row", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		rows := objectArray(broken["source_operations"])
		broken["source_operations"] = rows[1:]
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "source rows matrix=") {
			t.Fatalf("hidden-row error = %v, want source rows matrix", err)
		}
	})

	t.Run("rejects missing lane cell", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		row := objectArray(broken["source_operations"])[0]
		delete(object(row["lanes"]), "sync_transport")
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "missing lane cell") {
			t.Fatalf("missing-lane error = %v, want missing lane cell", err)
		}
	})

	t.Run("rejects missing source backlink", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		row := objectArray(broken["source_operations"])[0]
		facts := object(row["source_facts"])
		citation := object(facts["citation"])
		citation["source_location"] = ""
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "citation") {
			t.Fatalf("backlink error = %v, want citation", err)
		}
	})

	t.Run("rejects executable promotion", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		row := objectArray(broken["source_operations"])[0]
		cell := object(object(row["lanes"])["direct_read"])
		cell["disposition"] = "implemented"
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "implemented disposition") {
			t.Fatalf("promotion error = %v, want implemented disposition", err)
		}
	})

	t.Run("accepts source-documented bounded semantic POST reads", func(t *testing.T) {
		for _, sourceID := range []string{
			"notion.rest.post-database-query",
			"notion.rest.post-search",
			"notion.rest.query-meeting-notes",
			"notion.rest.introspect-token",
		} {
			t.Run(sourceID, func(t *testing.T) {
				row := notionMatrixRow(t, matrix, sourceID)
				cell := object(object(row["lanes"])["direct_read"])
				if stringValue(cell["applicability"]) != "source_candidate" ||
					stringValue(cell["disposition"]) != "mapped_unproven" ||
					stringValue(object(object(cell["mapping"])["source_fact"])["classification"]) != "source_semantic_post_read_candidate" {
					t.Fatalf("semantic POST direct-read cell %s = %#v, want source-backed mapped_unproven", sourceID, cell)
				}
			})
		}
	})

	t.Run("rejects mutation POST direct-read mapping without source read facts", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		row := notionMatrixRow(t, broken, "notion.rest.post-page")
		cell := object(object(row["lanes"])["direct_read"])
		cell["applicability"] = "source_candidate"
		cell["disposition"] = "mapped_unproven"
		cell["reason"] = "Incorrectly treating a mutation as a direct read."
		cell["mapping"] = map[string]any{
			"source_fact": map[string]any{"classification": "source_semantic_post_read_candidate"},
			"definition_backlink": map[string]any{
				"kind":      "source_lock",
				"path":      notionSourceLockPath,
				"source_id": "notion.rest.post-page",
			},
			"mapping_restriction": "",
			"runtime_claim":       "Source-backed mapping only; no runtime execution, certification, or availability proof is claimed.",
		}
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "must be explicit not_applicable") {
			t.Fatalf("mutation POST direct-read error = %v, want explicit not_applicable", err)
		}
	})

	t.Run("rejects semantic POST direct read without its retained response fact", func(t *testing.T) {
		brokenLock := cloneNotionJSON(t, lock)
		operation := notionLockOperation(t, brokenLock, "notion.rest.post-search")
		delete(object(object(operation["source_operation"])["responses"]), "200")
		if err := validateNotionMatrix(matrix, brokenLock); err == nil || !strings.Contains(err.Error(), "requires a 200 application/json response") {
			t.Fatalf("semantic POST response error = %v, want retained response fact", err)
		}
	})

	t.Run("preserves meeting-notes as both a direct read and restricted ETL candidate", func(t *testing.T) {
		row := notionMatrixRow(t, matrix, "notion.rest.query-meeting-notes")
		directRead := object(object(row["lanes"])["direct_read"])
		etl := object(object(row["lanes"])["etl"])
		if stringValue(directRead["disposition"]) != "mapped_unproven" ||
			stringValue(etl["disposition"]) != "mapped_unproven" ||
			stringValue(object(etl["mapping"])["mapping_restriction"]) != "source declares a bounded collection and has_more but no retained cursor, offset, or continuation request input" {
			t.Fatalf("meeting-notes lanes direct_read=%#v etl=%#v, want direct read plus restricted ETL", directRead, etl)
		}
	})

	t.Run("rejects crosswalk boundary drop", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		boundary := object(broken["source_boundary_reconciliation"])
		identities := objectArray(boundary["crosswalk_only_source_identities"])
		boundary["crosswalk_only_source_identities"] = identities[1:]
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "crosswalk-only identities") {
			t.Fatalf("boundary error = %v, want crosswalk-only identities", err)
		}
	})

	t.Run("rejects a crosswalk-only identity that overlaps the retained source lock", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		boundary := object(broken["source_boundary_reconciliation"])
		identity := object(objectArray(boundary["crosswalk_only_source_identities"])[0])
		operation := objectArray(object(lock["rest"])["operations"])[0]
		identity["method"] = stringValue(operation["method"])
		identity["path"] = stringValue(operation["path"])
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "incorrectly a source row") {
			t.Fatalf("source-overlap boundary error = %v, want non-source identity rejection", err)
		}
	})

	t.Run("rejects mapping-control restriction drop", func(t *testing.T) {
		broken := cloneNotionJSON(t, matrix)
		delete(broken, "mapping_control_restrictions")
		if err := validateNotionMatrix(broken, lock); err == nil || !strings.Contains(err.Error(), "mapping-control restriction") {
			t.Fatalf("restriction error = %v, want mapping-control restriction", err)
		}
	})
}

func validateNotionMatrix(matrix, lock map[string]any) error {
	if intValue(matrix["schema_version"]) != 1 || stringValue(matrix["connector"]) != "notion" {
		return fmt.Errorf("matrix identity mismatch")
	}
	if !sameStrings(stringsValue(matrix["lanes"]), notionLaneNames) {
		return fmt.Errorf("matrix lanes=%v", stringsValue(matrix["lanes"]))
	}
	if !sameStrings(stringsValue(matrix["disposition_vocabulary"]), []string{"implemented", "mapped_unproven", "missing_foundation", "not_applicable"}) {
		return fmt.Errorf("matrix disposition vocabulary=%v", stringsValue(matrix["disposition_vocabulary"]))
	}
	if err := validateNotionSnapshot(object(matrix["source_snapshot"])); err != nil {
		return err
	}
	if err := validateNotionAtlas(object(matrix["foundation_atlas"])); err != nil {
		return err
	}

	rest := object(lock["rest"])
	operations := objectArray(rest["operations"])
	lockedByID := make(map[string]map[string]any, len(operations))
	for _, operation := range operations {
		id := stringValue(operation["id"])
		if id == "" || lockedByID[id] != nil {
			return fmt.Errorf("invalid locked source id %q", id)
		}
		lockedByID[id] = operation
	}
	if intValue(object(lock["counts"])["total"]) != len(lockedByID) || len(lockedByID) != 49 {
		return fmt.Errorf("frozen source denominator is not 49")
	}
	if err := validateNotionSourceLockHeader(object(matrix["source_lock"]), lock, len(lockedByID)); err != nil {
		return err
	}
	if err := validateNotionMappingControlRestrictions(objectArray(matrix["mapping_control_restrictions"]), len(lockedByID)); err != nil {
		return err
	}
	if err := validateNotionBoundary(object(matrix["source_boundary_reconciliation"]), lock, lockedByID); err != nil {
		return err
	}
	if err := validateNotionCohorts(lockedByID); err != nil {
		return err
	}

	rows := objectArray(matrix["source_operations"])
	if len(rows) != len(lockedByID) {
		return fmt.Errorf("source rows matrix=%d lock=%d", len(rows), len(lockedByID))
	}
	seen := map[string]bool{}
	counts := map[string]map[string]int{}
	for _, row := range rows {
		sourceID := stringValue(row["source_id"])
		locked := lockedByID[sourceID]
		if locked == nil || seen[sourceID] {
			return fmt.Errorf("matrix source identity %q is absent or duplicated", sourceID)
		}
		seen[sourceID] = true
		if err := validateNotionFacts(sourceID, object(row["source_facts"]), locked, rest); err != nil {
			return err
		}
		lanes := object(row["lanes"])
		for _, lane := range notionLaneNames {
			cell, ok := lanes[lane]
			if !ok {
				return fmt.Errorf("source %s missing lane cell %s", sourceID, lane)
			}
			if err := validateNotionCell(sourceID, lane, object(cell)); err != nil {
				return err
			}
			if counts[lane] == nil {
				counts[lane] = map[string]int{}
			}
			counts[lane][stringValue(object(cell)["disposition"])]++
		}
		if len(lanes) != len(notionLaneNames) {
			return fmt.Errorf("source %s lane count=%d", sourceID, len(lanes))
		}
	}
	if len(seen) != len(lockedByID) {
		return fmt.Errorf("source identity reconciliation matrix=%d lock=%d", len(seen), len(lockedByID))
	}

	summary := object(matrix["summary"])
	if intValue(summary["source_rows"]) != 49 || intValue(summary["source_rows_with_all_lanes"]) != 49 || intValue(summary["total_lane_cells"]) != 343 {
		return fmt.Errorf("matrix summary rows/cells is not 49 x 7")
	}
	wantCounts := map[string]map[string]int{
		"direct_read":     {"mapped_unproven": 24, "not_applicable": 25},
		"direct_write":    {"mapped_unproven": 25, "not_applicable": 24},
		"binary_download": {"not_applicable": 49},
		"binary_upload":   {"mapped_unproven": 1, "not_applicable": 48},
		"etl":             {"mapped_unproven": 12, "not_applicable": 37},
		"reverse_etl":     {"mapped_unproven": 25, "not_applicable": 24},
		"sync_transport":  {"not_applicable": 49},
	}
	summaryCounts := object(summary["lane_counts"])
	for lane, want := range wantCounts {
		if !sameCounts(counts[lane], want) || !sameCounts(intMap(object(summaryCounts[lane])), want) {
			return fmt.Errorf("lane %s counts got=%v summary=%v want=%v", lane, counts[lane], intMap(object(summaryCounts[lane])), want)
		}
	}
	return nil
}

func validateNotionSnapshot(snapshot map[string]any) error {
	if stringValue(snapshot["source_snapshot_ref"]) != "fm/cli-top100-declaration-batch-r1" ||
		stringValue(snapshot["source_snapshot_commit"]) != "dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db" ||
		stringValue(snapshot["materialization"]) != "git_archive_byte_identical" {
		return fmt.Errorf("source snapshot provenance mismatch")
	}
	want := map[string]struct {
		blob  string
		bytes int
	}{
		notionSourceLockPath: {blob: "d70571f1c8f7c66f955de15af5f8fc08669a60e6", bytes: 1257564},
		notionCrosswalkPath:  {blob: "505e24e16371038ca6368342a3a7a2bc537e95ab", bytes: 43198},
	}
	files := objectArray(snapshot["retained_files"])
	if len(files) != len(want) {
		return fmt.Errorf("retained source files=%d", len(files))
	}
	for _, file := range files {
		path := stringValue(file["path"])
		expected, ok := want[path]
		if !ok || stringValue(file["git_blob_sha1"]) != expected.blob || intValue(file["bytes"]) != expected.bytes {
			return fmt.Errorf("retained source file %q is not byte-identified", path)
		}
		delete(want, path)
	}
	if len(want) != 0 {
		return fmt.Errorf("missing retained source file identities")
	}
	return nil
}

func validateNotionAtlas(atlas map[string]any) error {
	if stringValue(atlas["consulted_snapshot_ref"]) != "fm/cli-top100-declaration-batch-r1" ||
		stringValue(atlas["consulted_snapshot_commit"]) != "dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db" ||
		stringValue(atlas["catalog_path"]) != "docs/connector-canon/foundations/catalog.json" ||
		stringValue(atlas["usage"]) != "authoring_only_not_a_runtime_loader" {
		return fmt.Errorf("Foundation Atlas provenance mismatch")
	}
	want := stringSet(
		"source.retention-import.v1",
		"source.projection-admission.v1",
		"runtime.direct-execution.v1",
		"warehouse.stage-etl.v1",
		"warehouse.reverse-etl.v1",
		"transport.sync-contract.v1",
	)
	for _, row := range objectArray(atlas["classifications"]) {
		id := stringValue(row["id"])
		if !want[id] || stringValue(row["classification"]) != "reuse" || stringValue(row["reason"]) == "" {
			return fmt.Errorf("Foundation Atlas classification %q is not a source-mapping reuse", id)
		}
		delete(want, id)
	}
	if len(want) != 0 {
		return fmt.Errorf("missing Foundation Atlas classifications")
	}
	return nil
}

func validateNotionSourceLockHeader(header, lock map[string]any, denominator int) error {
	rest := object(lock["rest"])
	document := object(header["source_document"])
	if stringValue(header["path"]) != notionSourceLockPath ||
		intValue(header["schema_version"]) != intValue(lock["schema_version"]) ||
		stringValue(header["connector"]) != "notion" ||
		stringValue(document["source_url"]) != stringValue(rest["source_url"]) ||
		stringValue(document["sha256"]) != stringValue(rest["sha256"]) ||
		intValue(document["bytes"]) != intValue(rest["bytes"]) ||
		intValue(document["operation_count"]) != denominator {
		return fmt.Errorf("source lock header mismatch")
	}
	return nil
}

func validateNotionMappingControlRestrictions(rows []map[string]any, denominator int) error {
	type expectedRestriction struct {
		artifact string
		command  string
		result   string
		repair   string
	}
	want := map[string]expectedRestriction{
		"notion.source_projection.v2_embedded_source_operation_payload": {
			artifact: notionSourceLockPath,
			command:  "go run ./cmd/connectorgen validate internal/connectors/defs/notion --json",
			result:   "parse source lock: json: unknown field \"source_operation\"",
			repair:   "Extend the strict source-projection parser to accept and preserve the retained v2 source_operation field without dropping any locked row.",
		},
		"notion.source_projection.canonical_descriptor_absent": {
			artifact: "sources/notion-operation-descriptor.json",
			command:  "go run ./cmd/connectorgen surface-sync internal/connectors/defs --check",
			result:   "notion source projection: canonical source descriptor is missing",
			repair:   "Establish a source-preserving canonical descriptor from the frozen v2 source lock without discarding any locked row or embedded source_operation fact.",
		},
	}
	if len(rows) != len(want) {
		return fmt.Errorf("mapping-control restrictions=%d", len(rows))
	}
	for _, restriction := range rows {
		id := stringValue(restriction["id"])
		expected, ok := want[id]
		if !ok ||
			stringValue(restriction["classification"]) != "mapping_restriction" ||
			stringValue(restriction["affected_artifact"]) != expected.artifact ||
			stringValue(restriction["affected_source_scope"]) != "all_retained_source_rows" ||
			intValue(restriction["affected_source_count"]) != denominator ||
			stringValue(restriction["observed_command"]) != expected.command ||
			stringValue(restriction["observed_result"]) != expected.result ||
			stringValue(restriction["required_mapping_repair"]) != expected.repair ||
			restriction["runtime_foundation"] != false {
			return fmt.Errorf("mapping-control restriction %q is incomplete or misclassified", id)
		}
		delete(want, id)
	}
	if len(want) != 0 {
		return fmt.Errorf("mapping-control restriction is missing")
	}
	return nil
}

// The retained matrix includes the precise historical crosswalk-only boundary
// sidecar and pins the original crosswalk blob in source_snapshot. It lets this
// Batch R1 mapping test prove the source/non-source split without importing the
// entire legacy crosswalk artifact.
func validateNotionBoundary(boundary, lock map[string]any, lockedByID map[string]map[string]any) error {
	if stringValue(boundary["identity"]) != "source_id" ||
		intValue(boundary["source_lock_rows"]) != len(lockedByID) ||
		intValue(boundary["crosswalk_rows"]) != len(lockedByID) ||
		intValue(boundary["exact_source_rows"]) != len(lockedByID) ||
		intValue(boundary["lock_only_rows"]) != 0 {
		return fmt.Errorf("source boundary counts mismatch")
	}
	lockedMethodPaths := make(map[string]struct{}, len(lockedByID))
	for sourceID, operation := range lockedByID {
		key := stringValue(operation["method"]) + " " + stringValue(operation["path"])
		if _, duplicate := lockedMethodPaths[key]; duplicate {
			return fmt.Errorf("source lock method/path identity %q is duplicated", key)
		}
		lockedMethodPaths[key] = struct{}{}
		if sourceID == "" {
			return fmt.Errorf("source lock has an empty source ID")
		}
	}
	only := objectArray(boundary["crosswalk_only_source_identities"])
	want := map[string]string{
		"GET /v1/databases":                      "https://developers.notion.com/reference/get-databases",
		"POST /v1/databases/{database_id}/query": "https://developers.notion.com/reference/post-database-query",
	}
	if intValue(boundary["crosswalk_only_rows"]) != len(want) || len(only) != len(want) {
		return fmt.Errorf("crosswalk-only identities=%d want=%d", len(only), len(want))
	}
	for _, sourceOnly := range only {
		key := stringValue(sourceOnly["method"]) + " " + stringValue(sourceOnly["path"])
		if _, sourceRow := lockedMethodPaths[key]; sourceRow {
			return fmt.Errorf("crosswalk-only identity %q is incorrectly a source row", key)
		}
		wantURL, ok := want[key]
		if !ok ||
			stringValue(sourceOnly["source_url"]) != wantURL ||
			stringValue(sourceOnly["disposition"]) != "not_source_row" ||
			stringValue(sourceOnly["reason"]) == "" {
			return fmt.Errorf("crosswalk-only identity %q lost its source boundary", key)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		return fmt.Errorf("crosswalk-only identities are missing from the retained boundary")
	}
	eventOnly := object(boundary["event_schema_only"])
	if stringValue(eventOnly["source_contract_path"]) != "source_contract.components.schemas" ||
		intValue(eventOnly["webhook_schema_count"]) != countWebhookSchemas(lock) ||
		stringValue(eventOnly["disposition"]) != "not_source_row" ||
		stringValue(eventOnly["reason"]) == "" {
		return fmt.Errorf("event schema boundary mismatch")
	}
	return nil
}

func validateNotionCohorts(locked map[string]map[string]any) error {
	if len(notionGETDirectReadIDs) != 20 || len(notionSemanticReadNonGETIDs) != 4 || len(notionMutationIDs) != 25 || len(notionETLIDs) != 12 || len(notionBinaryUploadIDs) != 1 {
		return fmt.Errorf("manual cohort cardinality changed")
	}
	for id, operation := range locked {
		method := stringValue(operation["method"])
		if method == "GET" && !notionGETDirectReadIDs[id] {
			return fmt.Errorf("GET source %s is missing direct-read classification", id)
		}
		if notionGETDirectReadIDs[id] && method != "GET" {
			return fmt.Errorf("GET direct-read source %s has method %s", id, method)
		}
		if method != "GET" && !notionSemanticReadNonGETIDs[id] && !notionMutationIDs[id] {
			return fmt.Errorf("non-GET source %s is not classified", id)
		}
		if notionSemanticReadNonGETIDs[id] {
			if err := validateNotionSemanticPOSTReadEvidence(id, operation); err != nil {
				return err
			}
		}
	}
	wantSemanticSummary := map[string]string{
		"notion.rest.post-database-query": "Query a data source",
		"notion.rest.post-search":         "Search by title",
		"notion.rest.query-meeting-notes": "Query meeting notes",
		"notion.rest.introspect-token":    "Introspect a token",
	}
	for id, summary := range wantSemanticSummary {
		if stringValue(object(locked[id]["source_operation"])["summary"]) != summary {
			return fmt.Errorf("semantic read %s lost its source summary", id)
		}
	}
	if !strings.Contains(stringify(object(locked["notion.rest.upload-file"]["source_operation"])), "multipart/form-data") {
		return fmt.Errorf("binary upload source lacks multipart source media")
	}
	return nil
}

func validateNotionFacts(sourceID string, facts, locked, rest map[string]any) error {
	if stringValue(facts["source_lock"]) != notionSourceLockPath ||
		stringValue(facts["protocol"]) != stringValue(locked["protocol"]) ||
		stringValue(facts["method"]) != stringValue(locked["method"]) ||
		stringValue(facts["path"]) != stringValue(locked["path"]) ||
		stringValue(facts["operation_id"]) != stringValue(locked["operation_id"]) {
		return fmt.Errorf("source %s identity facts mismatch", sourceID)
	}
	citation := object(facts["citation"])
	if stringValue(citation["source_url"]) != stringValue(rest["source_url"]) ||
		stringValue(citation["sha256"]) != stringValue(rest["sha256"]) ||
		intValue(citation["bytes"]) != intValue(rest["bytes"]) ||
		stringValue(citation["source_location"]) != stringValue(locked["source_location"]) ||
		stringValue(citation["source_location"]) == "" {
		return fmt.Errorf("source %s citation mismatch", sourceID)
	}
	sourceOperation := object(locked["source_operation"])
	scope := object(facts["scope_and_fanout"])
	path, query := operationParameters(sourceOperation)
	if !sameStrings(stringsValue(scope["path_parameters"]), path) ||
		!sameStrings(stringsValue(scope["query_parameters"]), query) ||
		stringValue(scope["fanout_state"]) != "not_documented_by_locked_operation" {
		return fmt.Errorf("source %s scope/fanout facts mismatch", sourceID)
	}
	media := object(facts["media"])
	request, response := operationMedia(sourceOperation)
	if !sameStrings(stringsValue(media["request_media_types"]), request) ||
		!sameStrings(stringsValue(media["success_response_media_types"]), response) {
		return fmt.Errorf("source %s media facts mismatch", sourceID)
	}
	wantBinary := "not_documented_by_locked_operation"
	if notionBinaryUploadIDs[sourceID] {
		wantBinary = "multipart_form_data_file_upload"
	}
	if stringValue(media["binary_signal"]) != wantBinary {
		return fmt.Errorf("source %s binary fact mismatch", sourceID)
	}
	semantics := object(facts["operation_semantics"])
	wantSemantics := "mutation_candidate"
	if notionGETDirectReadIDs[sourceID] {
		wantSemantics = "read_get_candidate"
	} else if notionSemanticReadNonGETIDs[sourceID] {
		wantSemantics = "semantic_read_non_get"
	}
	if stringValue(semantics["state"]) != wantSemantics || stringValue(semantics["source_summary"]) != stringValue(sourceOperation["summary"]) {
		return fmt.Errorf("source %s semantic facts mismatch", sourceID)
	}
	event := object(facts["event_cursor"])
	if len(stringsValue(event["callback_names"])) != 0 || stringValue(event["state"]) != "no_operation_callback_or_webhook_registration_in_locked_rows" {
		return fmt.Errorf("source %s event facts mismatch", sourceID)
	}
	return validateNotionPagination(sourceID, object(facts["pagination"]), sourceOperation)
}

func validateNotionSemanticPOSTReadEvidence(sourceID string, locked map[string]any) error {
	if stringValue(locked["method"]) != "POST" {
		return fmt.Errorf("semantic POST read %s has method %q", sourceID, stringValue(locked["method"]))
	}
	operation := object(locked["source_operation"])
	requestSchema := object(object(object(operation["requestBody"])["content"])["application/json"])
	requestSchema = object(requestSchema["schema"])
	responseSchema := object(object(object(object(operation["responses"])["200"])["content"])["application/json"])
	responseSchema = object(responseSchema["schema"])
	if len(requestSchema) == 0 {
		return fmt.Errorf("semantic POST read %s requires a documented application/json request", sourceID)
	}
	if len(responseSchema) == 0 {
		return fmt.Errorf("semantic POST read %s requires a 200 application/json response", sourceID)
	}
	switch sourceID {
	case "notion.rest.post-database-query", "notion.rest.post-search":
		if !notionSchemaHasProperties(requestSchema, "start_cursor", "page_size") ||
			!notionSchemaHasProperties(responseSchema, "results", "next_cursor", "has_more") {
			return fmt.Errorf("semantic POST read %s lacks retained one-page request/response facts", sourceID)
		}
	case "notion.rest.query-meeting-notes":
		limit := object(object(requestSchema["properties"])["limit"])
		if intValue(limit["maximum"]) != 50 || !notionSchemaHasProperties(responseSchema, "results", "has_more") {
			return fmt.Errorf("semantic POST read %s lacks retained bounded meeting-notes response facts", sourceID)
		}
	case "notion.rest.introspect-token":
		if !notionSchemaRequires(requestSchema, "token") || !notionSchemaRequires(responseSchema, "active") {
			return fmt.Errorf("semantic POST read %s lacks retained token/request response facts", sourceID)
		}
	default:
		return fmt.Errorf("semantic POST read %s is not an explicitly reviewed source candidate", sourceID)
	}
	return nil
}

func notionSchemaHasProperties(schema map[string]any, names ...string) bool {
	properties := object(schema["properties"])
	for _, name := range names {
		if _, ok := properties[name]; !ok {
			return false
		}
	}
	return true
}

func notionSchemaRequires(schema map[string]any, name string) bool {
	for _, required := range stringsValue(schema["required"]) {
		if required == name {
			return true
		}
	}
	return false
}

func validateNotionPagination(sourceID string, pagination, sourceOperation map[string]any) error {
	if sourceID == "notion.rest.query-meeting-notes" {
		if stringValue(pagination["state"]) != "collection_response_has_more_without_retained_continuation" ||
			!sameStrings(stringsValue(pagination["request_controls"]), []string{"limit"}) ||
			!sameStrings(stringsValue(pagination["response_controls"]), []string{"results", "has_more"}) ||
			stringValue(pagination["mapping_restriction"]) != "source declares a bounded collection and has_more but no retained cursor, offset, or continuation request input" ||
			!strings.Contains(stringify(sourceOperation), "has_more") || !strings.Contains(stringify(sourceOperation), "limit") {
			return fmt.Errorf("source %s pagination restriction mismatch", sourceID)
		}
		return nil
	}
	if notionETLIDs[sourceID] {
		if stringValue(pagination["state"]) != "cursor_pagination_candidate" ||
			!sameStrings(stringsValue(pagination["request_controls"]), []string{"start_cursor", "page_size"}) ||
			!sameStrings(stringsValue(pagination["response_controls"]), []string{"results", "next_cursor", "has_more"}) ||
			stringValue(pagination["mapping_restriction"]) != "" {
			return fmt.Errorf("source %s cursor pagination facts mismatch", sourceID)
		}
		return nil
	}
	if sourceID == "notion.rest.patch-block-children" {
		if stringValue(pagination["state"]) != "response_cursor_metadata_on_mutation" ||
			len(stringsValue(pagination["request_controls"])) != 0 ||
			!sameStrings(stringsValue(pagination["response_controls"]), []string{"next_cursor", "has_more"}) {
			return fmt.Errorf("source %s mutation cursor facts mismatch", sourceID)
		}
		return nil
	}
	if stringValue(pagination["state"]) != "not_documented_by_locked_operation" ||
		len(stringsValue(pagination["request_controls"])) != 0 ||
		len(stringsValue(pagination["response_controls"])) != 0 ||
		stringValue(pagination["mapping_restriction"]) != "" {
		return fmt.Errorf("source %s unexpected pagination facts", sourceID)
	}
	return nil
}

func validateNotionCell(sourceID, lane string, cell map[string]any) error {
	if stringValue(cell["reason"]) == "" {
		return fmt.Errorf("source %s lane %s has empty reason", sourceID, lane)
	}
	if stringValue(cell["disposition"]) == "implemented" {
		return fmt.Errorf("source %s lane %s has unsupported implemented disposition", sourceID, lane)
	}
	expected := false
	classification := ""
	switch lane {
	case "direct_read":
		expected, classification = notionDirectReadCandidate(sourceID), notionDirectReadClassification(sourceID)
	case "direct_write", "reverse_etl":
		expected, classification = notionMutationIDs[sourceID], "source_semantic_mutation_candidate"
	case "binary_upload":
		expected, classification = notionBinaryUploadIDs[sourceID], "multipart_form_data_source_candidate"
	case "etl":
		expected, classification = notionETLIDs[sourceID], "source_collection_read_candidate"
	case "binary_download", "sync_transport":
	default:
		return fmt.Errorf("unknown lane %s", lane)
	}
	if !expected {
		if stringValue(cell["applicability"]) != "not_applicable" || stringValue(cell["disposition"]) != "not_applicable" || cell["mapping"] != nil {
			return fmt.Errorf("source %s lane %s must be explicit not_applicable", sourceID, lane)
		}
		return nil
	}
	if stringValue(cell["applicability"]) != "source_candidate" || stringValue(cell["disposition"]) != "mapped_unproven" || cell["mapping"] == nil {
		return fmt.Errorf("source %s lane %s must be mapped_unproven", sourceID, lane)
	}
	mapping := object(cell["mapping"])
	if stringValue(object(mapping["source_fact"])["classification"]) != classification ||
		stringValue(object(mapping["definition_backlink"])["kind"]) != "source_lock" ||
		stringValue(object(mapping["definition_backlink"])["path"]) != notionSourceLockPath ||
		stringValue(object(mapping["definition_backlink"])["source_id"]) != sourceID ||
		stringValue(mapping["runtime_claim"]) != "Source-backed mapping only; no runtime execution, certification, or availability proof is claimed." {
		return fmt.Errorf("source %s lane %s has invalid source backlink", sourceID, lane)
	}
	restriction := stringValue(mapping["mapping_restriction"])
	if sourceID == "notion.rest.query-meeting-notes" && lane == "etl" {
		if restriction != "source declares a bounded collection and has_more but no retained cursor, offset, or continuation request input" {
			return fmt.Errorf("meeting-notes ETL restriction missing")
		}
	} else if restriction != "" {
		return fmt.Errorf("source %s lane %s has unexpected mapping restriction", sourceID, lane)
	}
	return nil
}

func operationParameters(operation map[string]any) ([]string, []string) {
	var path, query []string
	for _, parameter := range objectArray(operation["parameters"]) {
		switch stringValue(parameter["in"]) {
		case "path":
			path = append(path, stringValue(parameter["name"]))
		case "query":
			query = append(query, stringValue(parameter["name"]))
		}
	}
	return path, query
}

func operationMedia(operation map[string]any) ([]string, []string) {
	var request []string
	if body, ok := operation["requestBody"]; ok {
		for media := range object(object(body)["content"]) {
			request = append(request, media)
		}
	}
	responses := map[string]bool{}
	for status, response := range object(operation["responses"]) {
		if strings.HasPrefix(status, "2") {
			for media := range object(object(response)["content"]) {
				responses[media] = true
			}
		}
	}
	sort.Strings(request)
	return request, setKeys(responses)
}

func countWebhookSchemas(lock map[string]any) int {
	count := 0
	for name := range object(object(object(lock["source_contract"])["components"])["schemas"]) {
		if strings.Contains(strings.ToLower(name), "webhook") {
			count++
		}
	}
	return count
}

func loadNotionJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value map[string]any
	if err := json.Unmarshal(contents, &value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return value
}

func notionMatrixRow(t *testing.T, matrix map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, row := range objectArray(matrix["source_operations"]) {
		if stringValue(row["source_id"]) == sourceID {
			return row
		}
	}
	t.Fatalf("matrix source row %q not found", sourceID)
	return nil
}

func notionLockOperation(t *testing.T, lock map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, operation := range objectArray(object(lock["rest"])["operations"]) {
		if stringValue(operation["id"]) == sourceID {
			return operation
		}
	}
	t.Fatalf("lock source operation %q not found", sourceID)
	return nil
}

func cloneNotionJSON(t *testing.T, value map[string]any) map[string]any {
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

func object(value any) map[string]any {
	got, _ := value.(map[string]any)
	return got
}

func objectArray(value any) []map[string]any {
	values, _ := value.([]any)
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, object(value))
	}
	return result
}

func stringsValue(value any) []string {
	values, _ := value.([]any)
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, stringValue(value))
	}
	return result
}

func stringValue(value any) string {
	got, _ := value.(string)
	return got
}

func intValue(value any) int {
	got, _ := value.(float64)
	return int(got)
}

func intMap(value map[string]any) map[string]int {
	result := make(map[string]int, len(value))
	for key, value := range value {
		result[key] = intValue(value)
	}
	return result
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sameCounts(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func stringSet(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		if result[value] {
			panic("duplicate static source ID: " + value)
		}
		result[value] = true
	}
	return result
}

func setKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringify(value any) string {
	contents, _ := json.Marshal(value)
	return string(contents)
}
