package vercel

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

const (
	vercelSourceLaneMatrixPath = "sources/vercel-source-lane-matrix.json"
	vercelSourceLockPath       = "sources/vercel-operation-source-lock.json"
	vercelCrosswalkPath        = "sources/vercel-operation-crosswalk.json"
	vercelStreamsPath          = "streams.json"
)

var vercelLanes = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

var vercelExpectedCounts = map[string]map[string]int{
	"direct_read":     {"mapped_unproven": 162, "not_applicable": 238},
	"direct_write":    {"mapped_unproven": 235, "not_applicable": 165},
	"binary_download": {"mapped_unproven": 1, "not_applicable": 399},
	"binary_upload":   {"mapped_unproven": 4, "not_applicable": 396},
	"etl":             {"mapped_unproven": 22, "not_applicable": 378},
	"reverse_etl":     {"mapped_unproven": 235, "not_applicable": 165},
	"sync_transport":  {"missing_foundation": 1, "not_applicable": 399},
}

var vercelLegacyETLStreams = map[string]map[string]string{
	"vercel.rest.getDomains":  {"stream": "domains", "path": "/v5/domains"},
	"vercel.rest.getProjects": {"stream": "projects", "path": "/v10/projects"},
	"vercel.rest.getTeams":    {"stream": "teams", "path": "/v2/teams"},
	"vercel.rest.listAliases": {"stream": "aliases", "path": "/v4/aliases"},
}

func TestVercelSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadVercelObject(t, vercelSourceLaneMatrixPath)
	lock := loadVercelObject(t, vercelSourceLockPath)
	crosswalk := loadVercelObject(t, vercelCrosswalkPath)
	streams := loadVercelObject(t, vercelStreamsPath)
	if err := validateVercelSourceLaneMatrix(matrix, lock, crosswalk, streams); err != nil {
		t.Fatalf("validate Vercel source lane matrix: %v", err)
	}

	t.Run("rejects missing lane cell", func(t *testing.T) {
		broken := cloneVercelObject(t, matrix)
		rows := mustVercelArray(t, broken["source_operations"])
		delete(mustVercelObject(t, rows[0])["lanes"].(map[string]any), "sync_transport")
		if err := validateVercelSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "missing lane cell") {
			t.Fatalf("missing-cell validation error = %v, want missing lane cell", err)
		}
	})

	t.Run("rejects hidden source row", func(t *testing.T) {
		broken := cloneVercelObject(t, matrix)
		rows := mustVercelArray(t, broken["source_operations"])
		broken["source_operations"] = rows[1:]
		if err := validateVercelSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "source rows matrix=") {
			t.Fatalf("hidden-row validation error = %v, want source rows matrix", err)
		}
	})

	t.Run("rejects legacy ETL backlink drift", func(t *testing.T) {
		broken := cloneVercelObject(t, matrix)
		row := vercelMatrixRow(t, broken, "vercel.rest.getDomains")
		lanes := mustVercelObject(t, row["lanes"])
		etl := mustVercelObject(t, lanes["etl"])
		mapping := mustVercelObject(t, etl["mapping"])
		mustVercelObject(t, mapping["definition_backlink"])["path"] = "wrong-streams.json"
		if err := validateVercelSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "legacy stream backlink") {
			t.Fatalf("backlink validation error = %v, want legacy stream backlink", err)
		}
	})

	t.Run("rejects crosswalk boundary drop", func(t *testing.T) {
		broken := cloneVercelObject(t, matrix)
		boundary := mustVercelObject(t, broken["source_boundary_reconciliation"])
		identities := mustVercelArray(t, boundary["crosswalk_only_source_identities"])
		boundary["crosswalk_only_source_identities"] = identities[1:]
		if err := validateVercelSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "crosswalk boundary identities") {
			t.Fatalf("boundary validation error = %v, want crosswalk boundary identities", err)
		}
	})

	t.Run("rejects invalid executable disposition", func(t *testing.T) {
		broken := cloneVercelObject(t, matrix)
		row := vercelMatrixRow(t, broken, "vercel.rest.readAccessGroup")
		lanes := mustVercelObject(t, row["lanes"])
		mustVercelObject(t, lanes["direct_read"])["disposition"] = "implemented"
		if err := validateVercelSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "lane direct_read") {
			t.Fatalf("invalid-disposition validation error = %v, want lane direct_read", err)
		}
	})
}

func TestVercelSourceLaneMatrixUsesRetainedSourceSemanticsBeyondHTTPVerb(t *testing.T) {
	matrix := loadVercelObject(t, vercelSourceLaneMatrixPath)
	lock := loadVercelObject(t, vercelSourceLockPath)
	locked := make(map[string]map[string]any)
	for _, operation := range mustMapSlice(objectAt(lock, "rest")["operations"]) {
		locked[stringAt(operation, "id")] = operation
	}

	cases := []struct {
		name               string
		sourceID           string
		lane               string
		wantApplicability  string
		wantDisposition    string
		wantOperationState string
	}{
		{
			name:               "HEAD equivalent-to-GET existence check is a direct read",
			sourceID:           "vercel.rest.artifactExists",
			lane:               "direct_read",
			wantApplicability:  "applicable",
			wantDisposition:    "mapped_unproven",
			wantOperationState: "source_get_equivalent_read_candidate",
		},
		{
			name:               "POST query is a direct read",
			sourceID:           "vercel.rest.artifactQuery",
			lane:               "direct_read",
			wantApplicability:  "applicable",
			wantDisposition:    "mapped_unproven",
			wantOperationState: "source_declared_read_candidate",
		},
		{
			name:               "POST file read is a direct read",
			sourceID:           "vercel.rest.readSessionFile",
			lane:               "direct_read",
			wantApplicability:  "applicable",
			wantDisposition:    "mapped_unproven",
			wantOperationState: "source_declared_read_candidate",
		},
		{
			name:              "POST file read is a source-cited binary download",
			sourceID:          "vercel.rest.readSessionFile",
			lane:              "binary_download",
			wantApplicability: "applicable",
			wantDisposition:   "mapped_unproven",
		},
		{
			name:              "documented gzip tarball write is a source-cited binary upload",
			sourceID:          "vercel.rest.writeSessionFiles",
			lane:              "binary_upload",
			wantApplicability: "applicable",
			wantDisposition:   "mapped_unproven",
		},
		{
			name:              "ordinary mutation POST stays out of direct read",
			sourceID:          "vercel.rest.createAccessGroup",
			lane:              "direct_read",
			wantApplicability: "not_applicable",
			wantDisposition:   "not_applicable",
		},
		{
			name:              "ordinary mutation POST remains direct write",
			sourceID:          "vercel.rest.createAccessGroup",
			lane:              "direct_write",
			wantApplicability: "applicable",
			wantDisposition:   "mapped_unproven",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operation := locked[tc.sourceID]
			if operation == nil {
				t.Fatalf("source lock operation %q is absent", tc.sourceID)
			}
			gotApplicability, gotDisposition := vercelExpectedLane(operation, tc.lane)
			if gotApplicability != tc.wantApplicability || gotDisposition != tc.wantDisposition {
				t.Fatalf("semantic lane %s %s = (%s, %s), want (%s, %s)", tc.sourceID, tc.lane, gotApplicability, gotDisposition, tc.wantApplicability, tc.wantDisposition)
			}

			row := vercelMatrixRow(t, matrix, tc.sourceID)
			cell := mustVercelObject(t, mustVercelObject(t, row["lanes"])[tc.lane])
			if stringAt(cell, "applicability") != tc.wantApplicability || stringAt(cell, "disposition") != tc.wantDisposition {
				t.Fatalf("matrix lane %s %s = (%s, %s), want (%s, %s)", tc.sourceID, tc.lane, stringAt(cell, "applicability"), stringAt(cell, "disposition"), tc.wantApplicability, tc.wantDisposition)
			}
			if tc.wantOperationState != "" {
				gotState := stringAt(mustVercelObject(t, objectAt(row, "source_facts")["operation_semantics"]), "state")
				if gotState != tc.wantOperationState {
					t.Fatalf("operation semantics %s = %q, want %q", tc.sourceID, gotState, tc.wantOperationState)
				}
			}
		})
	}
}

func TestVercelSourceSemanticLaneEvidenceRejectsIncompleteOrContradictoryFacts(t *testing.T) {
	lock := loadVercelObject(t, vercelSourceLockPath)
	locked := make(map[string]map[string]any)
	for _, operation := range mustMapSlice(objectAt(lock, "rest")["operations"]) {
		locked[stringAt(operation, "id")] = operation
	}

	cases := []struct {
		name                string
		sourceID            string
		mutate              func(map[string]any)
		wantDirectRead      string
		wantDirectWrite     string
		wantDownloadSignals []string
		wantUploadSignals   []string
	}{
		{
			name:     "query POST without successful response is not a direct read",
			sourceID: "vercel.rest.artifactQuery",
			mutate: func(operation map[string]any) {
				objectAt(operation, "source_operation")["responses"] = map[string]any{}
			},
			wantDirectRead:  "not_applicable",
			wantDirectWrite: "applicable",
		},
		{
			name:     "query wording contradicted by creation wording stays mutation",
			sourceID: "vercel.rest.artifactQuery",
			mutate: func(operation map[string]any) {
				objectAt(operation, "source_operation")["description"] = "Query information and create an artifact."
			},
			wantDirectRead:  "not_applicable",
			wantDirectWrite: "applicable",
		},
		{
			name:     "binary schema under JSON response media is not binary download evidence",
			sourceID: "vercel.rest.readSessionFile",
			mutate: func(operation map[string]any) {
				response := objectAt(objectAt(objectAt(operation, "source_operation"), "responses"), "200")
				response["content"] = map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"type": "string", "format": "binary"},
					},
				}
			},
			wantDownloadSignals: []string{},
		},
		{
			name:     "binary response media without binary schema is not binary download evidence",
			sourceID: "vercel.rest.readSessionFile",
			mutate: func(operation map[string]any) {
				response := objectAt(objectAt(objectAt(operation, "source_operation"), "responses"), "200")
				response["content"] = map[string]any{
					"application/octet-stream": map[string]any{
						"schema": map[string]any{"type": "string"},
					},
				}
			},
			wantDownloadSignals: []string{},
		},
		{
			name:     "gzip content type without binary payload documentation is not binary upload evidence",
			sourceID: "vercel.rest.writeSessionFiles",
			mutate: func(operation map[string]any) {
				objectAt(operation, "source_operation")["description"] = "The Content-Type header must be application/gzip."
			},
			wantUploadSignals: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operation := cloneVercelObject(t, locked[tc.sourceID])
			tc.mutate(operation)
			if tc.wantDirectRead != "" {
				got, _ := vercelExpectedLane(operation, "direct_read")
				if got != tc.wantDirectRead {
					t.Fatalf("direct-read applicability = %q, want %q", got, tc.wantDirectRead)
				}
			}
			if tc.wantDirectWrite != "" {
				got, _ := vercelExpectedLane(operation, "direct_write")
				if got != tc.wantDirectWrite {
					t.Fatalf("direct-write applicability = %q, want %q", got, tc.wantDirectWrite)
				}
			}
			if tc.wantDownloadSignals != nil {
				if got := vercelBinaryDownloadSignals(operation); !slices.Equal(got, tc.wantDownloadSignals) {
					t.Fatalf("binary-download signals = %v, want %v", got, tc.wantDownloadSignals)
				}
			}
			if tc.wantUploadSignals != nil {
				if got := vercelBinaryUploadSignals(operation); !slices.Equal(got, tc.wantUploadSignals) {
					t.Fatalf("binary-upload signals = %v, want %v", got, tc.wantUploadSignals)
				}
			}
		})
	}
}

func validateVercelSourceLaneMatrix(matrix, lock, crosswalk, streams map[string]any) error {
	if numberAt(matrix, "schema_version") != 1 || stringAt(matrix, "connector") != "vercel" {
		return fmt.Errorf("matrix identity drift")
	}
	if !slices.Equal(stringSlice(matrix["lanes"]), vercelLanes) {
		return fmt.Errorf("lane order drift")
	}
	if err := validateVercelLockBinding(matrix, lock); err != nil {
		return err
	}
	if err := validateVercelFoundationAtlas(matrix); err != nil {
		return err
	}

	operations := mustMapSlice(objectAt(lock, "rest")["operations"])
	locked := make(map[string]map[string]any, len(operations))
	lockedMethodPaths := make(map[string]struct{}, len(operations))
	for _, operation := range operations {
		id := stringAt(operation, "id")
		if _, exists := locked[id]; exists {
			return fmt.Errorf("duplicate source lock ID %q", id)
		}
		locked[id] = operation
		lockedMethodPaths[vercelMethodPathKey(stringAt(operation, "method"), stringAt(operation, "path"))] = struct{}{}
	}
	if numberAt(objectAt(lock, "counts"), "total") != 400 || len(locked) != 400 {
		return fmt.Errorf("source lock denominator=%d unique=%d, want 400", numberAt(objectAt(lock, "counts"), "total"), len(locked))
	}

	rows := mustMapSlice(matrix["source_operations"])
	if len(rows) != len(locked) {
		return fmt.Errorf("source rows matrix=%d lock=%d, want 400", len(rows), len(locked))
	}
	if err := validateVercelCrosswalkBoundary(objectAt(matrix, "source_boundary_reconciliation"), crosswalk, lockedMethodPaths); err != nil {
		return err
	}
	if err := validateVercelLegacyETLReconciliation(objectAt(matrix, "legacy_etl_reconciliation"), locked, streams); err != nil {
		return err
	}

	counts := make(map[string]map[string]int, len(vercelLanes))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		id := stringAt(row, "source_id")
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("duplicate matrix source ID %q", id)
		}
		seen[id] = struct{}{}
		operation, ok := locked[id]
		if !ok {
			return fmt.Errorf("matrix source ID %q is absent from the source lock", id)
		}
		if got, want := objectAt(row, "source_facts"), vercelExpectedSourceFacts(operation); !sameVercelJSON(got, want) {
			return fmt.Errorf("source facts %q drift", id)
		}
		lanes := objectAt(row, "lanes")
		for _, lane := range vercelLanes {
			rawCell, ok := lanes[lane]
			if !ok {
				return fmt.Errorf("missing lane cell: %s %s", id, lane)
			}
			cell := mustVercelObjectFrom(rawCell)
			if err := validateVercelLaneCell(id, lane, cell, operation); err != nil {
				return err
			}
			if counts[lane] == nil {
				counts[lane] = make(map[string]int)
			}
			counts[lane][stringAt(cell, "disposition")]++
		}
		if len(lanes) != len(vercelLanes) {
			return fmt.Errorf("lane cell count %s=%d, want %d", id, len(lanes), len(vercelLanes))
		}
	}
	if len(seen) != len(locked) {
		return fmt.Errorf("matrix does not retain every locked source ID: matrix=%d lock=%d", len(seen), len(locked))
	}
	summary := objectAt(matrix, "summary")
	if numberAt(summary, "source_rows") != 400 || numberAt(summary, "source_rows_with_all_lanes") != 400 {
		return fmt.Errorf("summary source rows drift")
	}
	summaryCounts := objectAt(summary, "lane_counts")
	for _, lane := range vercelLanes {
		if !equalVercelCounts(counts[lane], vercelExpectedCounts[lane]) {
			return fmt.Errorf("expected %s counts=%v, computed=%v", lane, vercelExpectedCounts[lane], counts[lane])
		}
		if !equalVercelCounts(numberMap(objectAt(summaryCounts, lane)), counts[lane]) {
			return fmt.Errorf("summary %s counts drift", lane)
		}
	}
	return nil
}

func validateVercelLockBinding(matrix, lock map[string]any) error {
	binding := objectAt(matrix, "source_lock")
	rest := objectAt(lock, "rest")
	document := objectAt(binding, "source_document")
	if stringAt(binding, "path") != vercelSourceLockPath || numberAt(binding, "schema_version") != numberAt(lock, "schema_version") || stringAt(binding, "connector") != stringAt(lock, "connector") || stringAt(document, "source_url") != stringAt(rest, "source_url") || stringAt(document, "sha256") != stringAt(rest, "sha256") || numberAt(document, "bytes") != numberAt(rest, "bytes") || numberAt(document, "operation_count") != numberAt(objectAt(lock, "counts"), "total") {
		return fmt.Errorf("source lock document binding drift")
	}
	return nil
}

func validateVercelFoundationAtlas(matrix map[string]any) error {
	atlas := objectAt(matrix, "foundation_atlas")
	if stringAt(atlas, "catalog_path") != "docs/connector-canon/foundations/catalog.json" || !slices.Equal(stringSlice(atlas["reuse_for_mapping_only"]), []string{"source.retention-import.v1", "runtime.direct-execution.v1", "warehouse.stage-etl.v1", "warehouse.reverse-etl.v1"}) {
		return fmt.Errorf("Foundation Atlas mapping-only reuse drift")
	}
	gap := objectAt(atlas, "sync_actual_gap")
	if stringAt(gap, "source_id") != "vercel.rest.createWebhook" || stringAt(gap, "lane") != "sync_transport" || stringAt(gap, "consulted_atlas_id") != "transport.sync-contract.v1" || strings.TrimSpace(stringAt(gap, "missing_capability")) == "" || !slices.Equal(stringSlice(gap["owner_symbols"]), []string{"internal/connectors/sync_transport.go#SyncTransportDescriptor", "internal/synctransport/orchestrator.go#(*Orchestrator).Run"}) || strings.TrimSpace(stringAt(gap, "proof_test_idea")) == "" || stringAt(gap, "status") != "recorded_only_requires_captain_approval_before_implementation" {
		return fmt.Errorf("Foundation Atlas sync-gap record drift")
	}
	return nil
}

func validateVercelCrosswalkBoundary(boundary, crosswalk map[string]any, lockedMethodPaths map[string]struct{}) error {
	if stringAt(crosswalk, "connector") != "vercel" || stringAt(crosswalk, "source_lock") != vercelSourceLockPath || len(mustMapSlice(crosswalk["source_operations"])) != 400 {
		return fmt.Errorf("crosswalk identity or denominator drift")
	}
	if stringAt(boundary, "identity") != "method + path" || numberAt(boundary, "source_lock_rows") != 400 || numberAt(boundary, "crosswalk_rows") != 400 || numberAt(boundary, "api_surface_rows") != 422 || numberAt(boundary, "crosswalk_only_rows") != 22 || numberAt(boundary, "lock_only_rows") != 0 {
		return fmt.Errorf("crosswalk boundary counts drift")
	}
	accounting := objectAt(crosswalk, "accounting")
	if numberAt(accounting, "source_operations") != 400 || numberAt(accounting, "source_unique_method_path") != 400 || numberAt(accounting, "api_surface_endpoints") != 422 || numberAt(accounting, "exact_source_to_surface") != 400 || numberAt(accounting, "source_only") != 0 || numberAt(accounting, "surface_only") != 22 {
		return fmt.Errorf("crosswalk accounting drift")
	}

	crosswalkMethodPaths := make(map[string]struct{}, 400)
	for _, operation := range mustMapSlice(crosswalk["source_operations"]) {
		crosswalkMethodPaths[vercelMethodPathKey(stringAt(operation, "method"), stringAt(operation, "path"))] = struct{}{}
	}
	for key := range lockedMethodPaths {
		if _, present := crosswalkMethodPaths[key]; !present {
			return fmt.Errorf("lock-minus-crosswalk identity %q", key)
		}
	}
	for key := range crosswalkMethodPaths {
		if _, present := lockedMethodPaths[key]; !present {
			return fmt.Errorf("crosswalk-minus-lock identity %q", key)
		}
	}

	want := make(map[string]map[string]any, 22)
	for _, entry := range mustMapSlice(crosswalk["surface_only"]) {
		want[vercelMethodPathKey(stringAt(entry, "method"), stringAt(entry, "path"))] = entry
	}
	identities := mustMapSlice(boundary["crosswalk_only_source_identities"])
	if len(identities) != len(want) {
		return fmt.Errorf("crosswalk boundary identities matrix=%d crosswalk=%d", len(identities), len(want))
	}
	seen := make(map[string]struct{}, len(identities))
	for _, identity := range identities {
		key := vercelMethodPathKey(stringAt(identity, "method"), stringAt(identity, "path"))
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("duplicate crosswalk boundary identity %q", key)
		}
		seen[key] = struct{}{}
		if stringAt(identity, "disposition") != "not_source_row" || strings.TrimSpace(stringAt(identity, "reason")) == "" || !reflect.DeepEqual(objectAt(identity, "crosswalk_entry"), want[key]) {
			return fmt.Errorf("crosswalk boundary identity drift %q", key)
		}
		if _, sourceRow := lockedMethodPaths[key]; sourceRow {
			return fmt.Errorf("crosswalk boundary identity %q is incorrectly a source row", key)
		}
	}
	return nil
}

func validateVercelLegacyETLReconciliation(reconciliation map[string]any, locked map[string]map[string]any, streams map[string]any) error {
	const criterion = "GET operation with required pagination response wrapper and retained limit+(since|until|cursor|from) or page+per_page query controls"
	if stringAt(reconciliation, "source_paging_candidate_criterion") != criterion || numberAt(reconciliation, "source_paging_candidates") != 22 || numberAt(reconciliation, "remaining_paging_candidates") != 18 {
		return fmt.Errorf("legacy ETL paging reconciliation counts or criterion drift")
	}
	links := mustMapSlice(reconciliation["legacy_stream_backlinks"])
	if len(links) != len(vercelLegacyETLStreams) {
		return fmt.Errorf("legacy stream backlink count=%d, want %d", len(links), len(vercelLegacyETLStreams))
	}
	streamPaths := make(map[string]string)
	for _, stream := range mustMapSlice(streams["streams"]) {
		streamPaths[stringAt(stream, "name")] = stringAt(stream, "path")
	}
	seen := make(map[string]struct{}, len(links))
	for _, link := range links {
		id := stringAt(link, "source_id")
		expected, ok := vercelLegacyETLStreams[id]
		if !ok {
			return fmt.Errorf("unexpected legacy stream backlink source ID %q", id)
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("duplicate legacy stream backlink source ID %q", id)
		}
		seen[id] = struct{}{}
		operation := locked[id]
		if stringAt(operation, "method") != "GET" || stringAt(operation, "path") != expected["path"] || !vercelIsETLCandidate(operation) || stringAt(link, "stream") != expected["stream"] || stringAt(link, "path") != expected["path"] || streamPaths[expected["stream"]] != expected["path"] {
			return fmt.Errorf("legacy stream backlink drift for %q", id)
		}
	}
	actual := 0
	for _, operation := range locked {
		if vercelIsETLCandidate(operation) {
			actual++
		}
	}
	if actual != 22 || actual-len(seen) != 18 {
		return fmt.Errorf("legacy ETL source candidate accounting=%d legacy=%d", actual, len(seen))
	}
	return nil
}

func validateVercelLaneCell(id, lane string, cell, operation map[string]any) error {
	applicability, disposition := vercelExpectedLane(operation, lane)
	if stringAt(cell, "applicability") != applicability || stringAt(cell, "disposition") != disposition {
		return fmt.Errorf("lane %s %s applicability=%q disposition=%q, want applicability=%q disposition=%q", lane, id, stringAt(cell, "applicability"), stringAt(cell, "disposition"), applicability, disposition)
	}
	if strings.TrimSpace(stringAt(cell, "reason")) == "" {
		return fmt.Errorf("lane %s %s lacks a reason", lane, id)
	}
	mapping, hasMapping := cell["mapping"]
	if applicability == "not_applicable" {
		if hasMapping || disposition != "not_applicable" {
			return fmt.Errorf("not-applicable lane promoted or mapped: %s %s", lane, id)
		}
		return nil
	}
	if disposition == "implemented" || (disposition != "mapped_unproven" && disposition != "missing_foundation") || !hasMapping {
		return fmt.Errorf("invalid applicable lane state: %s %s", lane, id)
	}
	if err := validateVercelLaneMapping(lane, mustVercelObjectFrom(mapping), operation); err != nil {
		return fmt.Errorf("mapping evidence %s %s: %w", lane, id, err)
	}
	return nil
}

func validateVercelLaneMapping(lane string, mapping, operation map[string]any) error {
	id := stringAt(operation, "id")
	switch lane {
	case "direct_read":
		sourceFact := objectAt(mapping, "source_fact")
		backlink := objectAt(mapping, "definition_backlink")
		if stringAt(sourceFact, "method") != stringAt(operation, "method") || stringAt(sourceFact, "classification") != vercelReadSemanticState(operation) || stringAt(backlink, "kind") != "api_surface_crosswalk" || stringAt(backlink, "path") != vercelCrosswalkPath || stringAt(backlink, "source_id") != id || strings.TrimSpace(stringAt(mapping, "runtime_claim")) == "" {
			return fmt.Errorf("direct-read source mapping drift")
		}
	case "direct_write", "reverse_etl":
		sourceFact := objectAt(mapping, "source_fact")
		backlink := objectAt(mapping, "definition_backlink")
		if !vercelIsMutationCandidate(operation) || stringAt(sourceFact, "method") != stringAt(operation, "method") || stringAt(sourceFact, "classification") != "mutation_verb_candidate" || stringAt(backlink, "kind") != "api_surface_crosswalk" || stringAt(backlink, "path") != vercelCrosswalkPath || stringAt(backlink, "source_id") != id || strings.TrimSpace(stringAt(mapping, "runtime_claim")) == "" {
			return fmt.Errorf("mutation source mapping drift")
		}
		if lane == "reverse_etl" && strings.TrimSpace(stringAt(mapping, "required_flow")) == "" {
			return fmt.Errorf("reverse-ETL required flow is absent")
		}
	case "binary_download":
		backlink := objectAt(mapping, "definition_backlink")
		if !slices.Equal(stringSlice(mapping["source_binary_signals"]), vercelBinaryDownloadSignals(operation)) || stringAt(backlink, "kind") != "future_binary_download_projection_required" || stringAt(backlink, "path") != "operations.json" || stringAt(backlink, "source_id") != id || strings.TrimSpace(stringAt(mapping, "runtime_claim")) == "" {
			return fmt.Errorf("binary-download source mapping drift")
		}
	case "binary_upload":
		backlink := objectAt(mapping, "definition_backlink")
		if !slices.Equal(stringSlice(mapping["source_binary_signals"]), vercelBinaryUploadSignals(operation)) || stringAt(backlink, "kind") != "future_binary_upload_projection_required" || stringAt(backlink, "path") != "writes.json" || stringAt(backlink, "source_id") != id || strings.TrimSpace(stringAt(mapping, "runtime_claim")) == "" {
			return fmt.Errorf("binary source mapping drift")
		}
	case "etl":
		sourceFact := objectAt(mapping, "source_fact")
		backlink := objectAt(mapping, "definition_backlink")
		if stringAt(sourceFact, "method") != "GET" || stringAt(sourceFact, "criterion") != "required pagination response wrapper and retained query controls" || !slices.Equal(stringSlice(sourceFact["parameters"]), vercelPagingQueryParameters(operation)) || strings.TrimSpace(stringAt(mapping, "runtime_claim")) == "" {
			return fmt.Errorf("ETL source mapping drift")
		}
		if legacy, ok := vercelLegacyETLStreams[id]; ok {
			if stringAt(backlink, "kind") != "legacy_stream_backlink" || stringAt(backlink, "path") != vercelStreamsPath || stringAt(backlink, "stream") != legacy["stream"] || stringAt(backlink, "stream_path") != legacy["path"] || stringAt(backlink, "source_id") != id {
				return fmt.Errorf("legacy stream backlink drift")
			}
		} else if stringAt(backlink, "kind") != "future_stream_projection_required" || stringAt(backlink, "path") != vercelStreamsPath || stringAt(backlink, "stream") != "" || stringAt(backlink, "stream_path") != "" || stringAt(backlink, "source_id") != id {
			return fmt.Errorf("future stream projection backlink drift")
		}
	case "sync_transport":
		lookup := objectAt(mapping, "atlas_lookup")
		evidence := objectAt(mapping, "source_event_evidence")
		if stringAt(mapping, "foundation_id") != "cli-webhook-event-surface-foundation-r1" || stringAt(lookup, "consulted_id") != "transport.sync-contract.v1" || stringAt(lookup, "classification") != "actual_gap" || !slices.Equal(stringSlice(lookup["owner_symbols"]), []string{"internal/connectors/sync_transport.go#SyncTransportDescriptor", "internal/synctransport/orchestrator.go#(*Orchestrator).Run"}) || strings.TrimSpace(stringAt(lookup, "insufficiency")) == "" || stringAt(evidence, "url_field") != "url" || stringAt(evidence, "events_field") != "events" || stringAt(evidence, "request_media") != "application/json" || strings.TrimSpace(stringAt(mapping, "runtime_claim")) == "" {
			return fmt.Errorf("missing-foundation mapping drift")
		}
	default:
		return fmt.Errorf("unknown lane %q", lane)
	}
	return nil
}

func vercelExpectedLane(operation map[string]any, lane string) (string, string) {
	applicable := false
	switch lane {
	case "direct_read":
		applicable = vercelReadSemanticState(operation) != ""
	case "direct_write", "reverse_etl":
		applicable = vercelIsMutationCandidate(operation)
	case "binary_download":
		applicable = len(vercelBinaryDownloadSignals(operation)) > 0
	case "binary_upload":
		applicable = len(vercelBinaryUploadSignals(operation)) > 0
	case "etl":
		applicable = vercelIsETLCandidate(operation)
	case "sync_transport":
		applicable = stringAt(operation, "id") == "vercel.rest.createWebhook"
	}
	if !applicable {
		return "not_applicable", "not_applicable"
	}
	if lane == "sync_transport" {
		return "applicable", "missing_foundation"
	}
	return "applicable", "mapped_unproven"
}

func vercelExpectedSourceFacts(operation map[string]any) map[string]any {
	facts := map[string]any{
		"protocol":     stringAt(operation, "protocol"),
		"method":       stringAt(operation, "method"),
		"path":         stringAt(operation, "path"),
		"operation_id": stringAt(operation, "operation_id"),
		"deprecated":   operation["deprecated"],
		"citation": map[string]any{
			"source_location": stringAt(operation, "source_location"),
		},
		"scope_and_fanout": map[string]any{
			"path_variables":   vercelPathVariables(operation),
			"query_parameters": vercelQueryParameters(operation),
			"fanout": map[string]any{
				"state":  "not_declared",
				"reason": "The retained source operation declares no standalone fanout contract.",
			},
		},
		"media": map[string]any{
			"request_media_types":          vercelRequestMedia(operation),
			"success_response_media_types": vercelSuccessMedia(operation),
			"binary_signals":               vercelBinarySignals(operation),
		},
		"pagination": map[string]any{
			"state":                   vercelPaginationState(operation),
			"paging_query_parameters": vercelPagingQueryParameters(operation),
		},
		"event_cursor": map[string]any{
			"state": vercelEventCursorState(operation),
		},
		"operation_semantics": map[string]any{
			"state": vercelOperationSemantics(operation),
		},
	}
	return facts
}

func vercelIsMutationVerb(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
}

func vercelIsMutationCandidate(operation map[string]any) bool {
	return vercelIsMutationVerb(stringAt(operation, "method")) && vercelReadSemanticState(operation) == ""
}

func vercelIsETLCandidate(operation map[string]any) bool {
	if stringAt(operation, "method") != "GET" || !vercelHasRequiredPaginationResponseWrapper(operation) {
		return false
	}
	params := vercelQueryParameters(operation)
	return slices.Contains(params, "limit") && (slices.Contains(params, "since") || slices.Contains(params, "until") || slices.Contains(params, "cursor") || slices.Contains(params, "from")) || slices.Contains(params, "page") && slices.Contains(params, "per_page")
}

func vercelHasRequiredPaginationResponseWrapper(operation map[string]any) bool {
	source := objectAt(operation, "source_operation")
	for _, response := range objectAt(source, "responses") {
		if vercelJSONHasRequiredPagination(response) {
			return true
		}
	}
	return false
}

func vercelJSONHasRequiredPagination(value any) bool {
	switch typed := value.(type) {
	case []any:
		for _, child := range typed {
			if vercelJSONHasRequiredPagination(child) {
				return true
			}
		}
	case map[string]any:
		properties, propertiesOK := typed["properties"].(map[string]any)
		required, requiredOK := typed["required"].([]any)
		if propertiesOK && requiredOK {
			if _, hasPagination := properties["pagination"]; hasPagination && slices.Contains(stringSlice(required), "pagination") {
				return true
			}
		}
		for _, child := range typed {
			if vercelJSONHasRequiredPagination(child) {
				return true
			}
		}
	}
	return false
}

func vercelPathVariables(operation map[string]any) []string {
	return vercelParameterNames(operation, "path")
}

func vercelQueryParameters(operation map[string]any) []string {
	return vercelParameterNames(operation, "query")
}

func vercelParameterNames(operation map[string]any, location string) []string {
	source := objectAt(operation, "source_operation")
	values := make([]string, 0)
	parameters, ok := source["parameters"].([]any)
	if !ok {
		return []string{}
	}
	for _, parameter := range mustMapSlice(parameters) {
		if stringAt(parameter, "in") == location && stringAt(parameter, "name") != "" {
			values = append(values, stringAt(parameter, "name"))
		}
	}
	return vercelSortedUnique(values)
}

func vercelRequestMedia(operation map[string]any) []string {
	source := objectAt(operation, "source_operation")
	values := make([]string, 0)
	requestBody, ok := source["requestBody"].(map[string]any)
	if ok {
		if content, contentOK := requestBody["content"].(map[string]any); contentOK {
			values = append(values, vercelSortedKeys(content)...)
		}
	}
	values = append(values, vercelDescriptionContentTypeMedia(stringAt(source, "description"))...)
	return vercelSortedUnique(values)
}

func vercelSuccessMedia(operation map[string]any) []string {
	source := objectAt(operation, "source_operation")
	values := make([]string, 0)
	for status, response := range objectAt(source, "responses") {
		if strings.HasPrefix(status, "2") {
			content, ok := mustVercelObjectFrom(response)["content"].(map[string]any)
			if ok {
				values = append(values, vercelSortedKeys(content)...)
			}
		}
	}
	return vercelSortedUnique(values)
}

func vercelBinarySignals(operation map[string]any) []string {
	values := append([]string{}, vercelBinaryDownloadSignals(operation)...)
	values = append(values, vercelBinaryUploadSignals(operation)...)
	return vercelSortedUnique(values)
}

func vercelBinaryDownloadSignals(operation map[string]any) []string {
	source := objectAt(operation, "source_operation")
	values := make([]string, 0)
	for status, response := range objectAt(source, "responses") {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		content, ok := mustVercelObjectFrom(response)["content"].(map[string]any)
		if !ok {
			continue
		}
		for media, mediaContract := range content {
			if vercelMediaSupportsBinary(media) && vercelJSONHasBinaryFormat(mediaContract) {
				values = append(values, "success_response_media:"+media)
			}
		}
	}
	return vercelSortedUnique(values)
}

func vercelBinaryUploadSignals(operation map[string]any) []string {
	source := objectAt(operation, "source_operation")
	values := make([]string, 0)
	if requestBody, ok := source["requestBody"].(map[string]any); ok {
		if content, ok := requestBody["content"].(map[string]any); ok {
			for media, mediaContract := range content {
				if vercelMediaSupportsBinary(media) && vercelJSONHasBinaryFormat(mediaContract) {
					values = append(values, "request_body_media:"+media)
				}
			}
		}
	}
	description := stringAt(source, "description")
	if vercelDescriptionDeclaresBinaryPayload(description) {
		for _, media := range vercelDescriptionContentTypeMedia(description) {
			if vercelMediaSupportsBinary(media) {
				values = append(values, "description_content_type:"+media)
			}
		}
	}
	return vercelSortedUnique(values)
}

func vercelDescriptionContentTypeMedia(description string) []string {
	description = strings.ToLower(description)
	values := make([]string, 0)
	for start := 0; start < len(description); {
		markerOffset := strings.Index(description[start:], "content-type")
		if markerOffset < 0 {
			break
		}
		markerEnd := start + markerOffset + len("content-type")
		mediaOffset := strings.Index(description[markerEnd:], "application/")
		if mediaOffset < 0 {
			start = markerEnd
			continue
		}
		mediaStart := markerEnd + mediaOffset
		mediaEnd := mediaStart
		for mediaEnd < len(description) && vercelMediaTypeTokenByte(description[mediaEnd]) {
			mediaEnd++
		}
		media, _, err := mime.ParseMediaType(description[mediaStart:mediaEnd])
		if err == nil {
			values = append(values, media)
		}
		start = mediaEnd
	}
	return vercelSortedUnique(values)
}

func vercelMediaTypeTokenByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '/' || value == '.' || value == '+' || value == '-'
}

func vercelMediaSupportsBinary(media string) bool {
	media, _, err := mime.ParseMediaType(media)
	if err != nil {
		return false
	}
	media = strings.ToLower(media)
	return !strings.HasSuffix(media, "/json") && !strings.Contains(media, "+json") && !strings.HasSuffix(media, "/xml") && !strings.Contains(media, "+xml") && media != "application/x-www-form-urlencoded"
}

func vercelDescriptionDeclaresBinaryPayload(description string) bool {
	description = strings.ToLower(description)
	return strings.Contains(description, "binary") || strings.Contains(description, "tarball")
}

func vercelJSONHasBinaryFormat(value any) bool {
	switch typed := value.(type) {
	case []any:
		for _, child := range typed {
			if vercelJSONHasBinaryFormat(child) {
				return true
			}
		}
	case map[string]any:
		if strings.EqualFold(stringAt(typed, "format"), "binary") {
			return true
		}
		for _, child := range typed {
			if vercelJSONHasBinaryFormat(child) {
				return true
			}
		}
	}
	return false
}

func vercelPagingQueryParameters(operation map[string]any) []string {
	known := map[string]struct{}{"cursor": {}, "from": {}, "limit": {}, "next": {}, "page": {}, "per_page": {}, "since": {}, "until": {}}
	values := make([]string, 0)
	for _, parameter := range vercelQueryParameters(operation) {
		if _, ok := known[parameter]; ok {
			values = append(values, parameter)
		}
	}
	return values
}

func vercelPaginationState(operation map[string]any) string {
	if vercelIsETLCandidate(operation) {
		return "selected_paging_candidate"
	}
	if vercelHasRequiredPaginationResponseWrapper(operation) {
		return "pagination_response_unselected"
	}
	if len(vercelPagingQueryParameters(operation)) > 0 {
		return "paging_parameters_unselected"
	}
	return "not_declared"
}

func vercelEventCursorState(operation map[string]any) string {
	switch stringAt(operation, "id") {
	case "vercel.rest.createWebhook":
		return "webhook_registration_url_and_events"
	case "vercel.rest.getDeploymentEvents", "vercel.rest.listUserEvents":
		return "event_collection_read"
	case "vercel.rest.listEventTypes":
		return "event_type_catalog"
	default:
		return "not_declared"
	}
}

func vercelReadSemanticState(operation map[string]any) string {
	if stringAt(operation, "method") == "GET" {
		return "read_verb_candidate"
	}
	if !vercelHasSuccessfulResponse(operation) {
		return ""
	}
	source := objectAt(operation, "source_operation")
	description := stringAt(source, "description")
	if vercelDescriptionDeclaresGETEquivalence(description) {
		return "source_get_equivalent_read_candidate"
	}
	if vercelSummaryDeclaresRead(stringAt(source, "summary")) && !vercelDescriptionDeclaresMutation(description) {
		return "source_declared_read_candidate"
	}
	return ""
}

func vercelHasSuccessfulResponse(operation map[string]any) bool {
	for status := range objectAt(objectAt(operation, "source_operation"), "responses") {
		if strings.HasPrefix(status, "2") {
			return true
		}
	}
	return false
}

func vercelDescriptionDeclaresGETEquivalence(description string) bool {
	description = strings.ReplaceAll(strings.ToLower(description), "`", "")
	return strings.Contains(description, "equivalent to a get request")
}

func vercelSummaryDeclaresRead(summary string) bool {
	words := strings.Fields(strings.ToLower(summary))
	if len(words) == 0 {
		return false
	}
	return words[0] == "read" || words[0] == "query"
}

func vercelDescriptionDeclaresMutation(description string) bool {
	description = strings.ToLower(description)
	for _, term := range []string{"create", "update", "delete", "write", "upload", "remove", "execute", "purchase"} {
		if strings.Contains(description, term) {
			return true
		}
	}
	return false
}

func vercelOperationSemantics(operation map[string]any) string {
	if state := vercelReadSemanticState(operation); state != "" {
		return state
	}
	method := stringAt(operation, "method")
	switch {
	case vercelIsMutationVerb(method):
		return "mutation_verb_candidate"
	case method == "HEAD":
		return "head_verb_candidate"
	default:
		return "other_verb_candidate"
	}
}

func vercelMethodPathKey(method, path string) string {
	return method + " " + path
}

func vercelSortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func vercelSortedUnique(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return slices.Compact(copyValues)
}

func equalVercelCounts(got, want map[string]int) bool {
	if len(got) != len(want) {
		return false
	}
	for key, count := range want {
		if got[key] != count {
			return false
		}
	}
	return true
}

func sameVercelJSON(left, right any) bool {
	leftRaw, leftErr := json.Marshal(left)
	rightRaw, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftRaw) == string(rightRaw)
}

func loadVercelObject(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return value
}

func cloneVercelObject(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal clone: %v", err)
	}
	var clone map[string]any
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatalf("decode clone: %v", err)
	}
	return clone
}

func vercelMatrixRow(t *testing.T, matrix map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, row := range mustMapSlice(matrix["source_operations"]) {
		if stringAt(row, "source_id") == sourceID {
			return row
		}
	}
	t.Fatalf("matrix row %q not found", sourceID)
	return nil
}

func mustVercelObject(t *testing.T, value any) map[string]any {
	t.Helper()
	return mustVercelObjectFrom(value)
}

func mustVercelObjectFrom(value any) map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("wanted JSON object, got %T", value))
	}
	return object
}

func mustVercelArray(t *testing.T, value any) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("wanted JSON array, got %T", value)
	}
	return array
}

func mustMapSlice(value any) []map[string]any {
	array, ok := value.([]any)
	if !ok {
		panic(fmt.Sprintf("wanted JSON object array, got %T", value))
	}
	result := make([]map[string]any, 0, len(array))
	for _, item := range array {
		result = append(result, mustVercelObjectFrom(item))
	}
	return result
}

func objectAt(value map[string]any, key string) map[string]any {
	return mustVercelObjectFrom(value[key])
}

func stringAt(value map[string]any, key string) string {
	result, _ := value[key].(string)
	return result
}

func numberAt(value map[string]any, key string) int {
	number, _ := value[key].(float64)
	return int(number)
}

func stringSlice(value any) []string {
	array, ok := value.([]any)
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(array))
	for _, item := range array {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func numberMap(value map[string]any) map[string]int {
	result := make(map[string]int, len(value))
	for key, raw := range value {
		if number, ok := raw.(float64); ok {
			result[key] = int(number)
		}
	}
	return result
}
