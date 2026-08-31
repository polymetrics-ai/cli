package notion

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
	notionDeclarationDispositionPath = "sources/notion-declaration-disposition.json"
	notionRetentionContractPath      = "sources/notion-retained-mapping-contract.json"
)

type notionDispositionArtifactRecord struct {
	Present bool
	Count   int
}

func TestNotionDeclarationDisposition(t *testing.T) {
	disposition := loadNotionJSON(t, notionDeclarationDispositionPath)
	lock := loadNotionJSON(t, notionSourceLockPath)
	matrix := loadNotionJSON(t, notionSourceLaneMatrixPath)
	retentionContract := loadNotionJSON(t, notionRetentionContractPath)
	artifacts := notionDispositionArtifactRecords(t)
	if err := validateNotionDeclarationDisposition(disposition, lock, matrix, retentionContract, artifacts); err != nil {
		t.Fatal(err)
	}

	t.Run("rejects canonical descriptor claim", func(t *testing.T) {
		broken := cloneNotionJSON(t, disposition)
		object(broken["retention_boundary"])["canonical_evidence"] = true
		if err := validateNotionDeclarationDisposition(broken, lock, matrix, retentionContract, artifacts); err == nil || !strings.Contains(err.Error(), "canonical_evidence") {
			t.Fatalf("canonical descriptor claim error = %v, want canonical_evidence refusal", err)
		}
	})

	t.Run("rejects source-bound artifact claim without a descriptor", func(t *testing.T) {
		broken := cloneNotionJSON(t, disposition)
		object(broken["retention_boundary"])["source_bound_artifact_claims"] = float64(1)
		if err := validateNotionDeclarationDisposition(broken, lock, matrix, retentionContract, artifacts); err == nil || !strings.Contains(err.Error(), "source-bound artifact claims") {
			t.Fatalf("source-bound artifact claim error = %v, want refusal", err)
		}
	})

	t.Run("rejects missing existing mapping restriction", func(t *testing.T) {
		broken := cloneNotionJSON(t, disposition)
		object(broken["retention_boundary"])["mapping_restriction_ids"] = []any{
			"notion.source_projection.canonical_descriptor_absent",
		}
		if err := validateNotionDeclarationDisposition(broken, lock, matrix, retentionContract, artifacts); err == nil || !strings.Contains(err.Error(), "mapping restrictions") {
			t.Fatalf("mapping restriction error = %v, want refusal", err)
		}
	})

	t.Run("rejects reconstructing the absent historical crosswalk", func(t *testing.T) {
		broken := cloneNotionJSON(t, disposition)
		object(broken["historical_crosswalk_metadata"])["state"] = "reconstructed"
		if err := validateNotionDeclarationDisposition(broken, lock, matrix, retentionContract, artifacts); err == nil || !strings.Contains(err.Error(), "historical crosswalk state") {
			t.Fatalf("historical crosswalk state error = %v, want refusal", err)
		}
	})
}

func TestNotionSourceBoundArtifactClaimCounter(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{
			name:  "happy legacy artifact without a source binding",
			value: map[string]any{"stream": "pages"},
			want:  0,
		},
		{
			name:  "empty source operation is not a claim",
			value: map[string]any{"source_operation": ""},
			want:  0,
		},
		{
			name:  "bad source operation is a claim",
			value: map[string]any{"source_operation": "provider.rest.listRecords"},
			want:  1,
		},
		{
			name: "edge nested malformed source operation remains a claim",
			value: map[string]any{"commands": []any{
				map[string]any{"source_operation": map[string]any{"id": "provider.rest.listRecords"}},
			}},
			want: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := notionCountSourceOperationClaims(test.value); got != test.want {
				t.Fatalf("notionCountSourceOperationClaims() = %d, want %d", got, test.want)
			}
		})
	}
}

func validateNotionDeclarationDisposition(disposition, lock, matrix, retentionContract map[string]any, artifacts map[string]notionDispositionArtifactRecord) error {
	if intValue(disposition["schema_version"]) != 1 || stringValue(disposition["connector"]) != "notion" {
		return fmt.Errorf("declaration disposition identity mismatch")
	}
	if err := validateNotionDispositionSourceBasis(object(disposition["source_basis"]), lock, matrix, retentionContract); err != nil {
		return err
	}
	if err := validateNotionDispositionMapping(object(disposition["mapping"]), matrix, artifacts); err != nil {
		return err
	}
	if err := validateNotionDispositionRetentionBoundary(object(disposition["retention_boundary"]), retentionContract, matrix); err != nil {
		return err
	}
	if err := validateNotionHistoricalCrosswalkMetadata(object(disposition["historical_crosswalk_metadata"]), matrix); err != nil {
		return err
	}
	notes := stringsValue(disposition["notes"])
	if len(notes) != 4 {
		return fmt.Errorf("declaration disposition notes = %d, want four explicit boundaries", len(notes))
	}
	for _, note := range notes {
		if strings.TrimSpace(note) == "" {
			return fmt.Errorf("declaration disposition contains an empty note")
		}
	}
	return nil
}

func validateNotionDispositionSourceBasis(basis, lock, matrix, retentionContract map[string]any) error {
	rest := object(lock["rest"])
	lockBasis := object(basis["source_lock"])
	if stringValue(lockBasis["path"]) != notionSourceLockPath ||
		stringValue(lockBasis["source_url"]) != stringValue(rest["source_url"]) ||
		stringValue(lockBasis["sha256"]) != stringValue(rest["sha256"]) ||
		intValue(lockBasis["bytes"]) != intValue(rest["bytes"]) ||
		intValue(lockBasis["source_operation_count"]) != len(objectArray(rest["operations"])) {
		return fmt.Errorf("declaration disposition source-lock basis does not bind the exact retained lock")
	}
	if stringValue(basis["source_lane_matrix"]) != notionSourceLaneMatrixPath ||
		stringValue(basis["retention_contract"]) != notionRetentionContractPath {
		return fmt.Errorf("declaration disposition source basis paths mismatch")
	}
	if stringValue(retentionContract["connector"]) != "notion" || retentionContract["retention_only"] != true {
		return fmt.Errorf("retention contract identity is not notion/retention_only")
	}
	contractLock := object(retentionContract["source_lock"])
	if stringValue(contractLock["path"]) != notionSourceLockPath ||
		stringValue(contractLock["sha256"]) != stringValue(rest["sha256"]) ||
		intValue(contractLock["bytes"]) != intValue(rest["bytes"]) ||
		contractLock["canonical_evidence"] != false {
		return fmt.Errorf("retention contract does not bind the descriptor-absent Notion source lock")
	}
	if len(objectArray(matrix["source_operations"])) != len(objectArray(rest["operations"])) {
		return fmt.Errorf("matrix source rows do not reconcile with the retained lock")
	}
	return nil
}

func validateNotionDispositionMapping(mapping, matrix map[string]any, artifacts map[string]notionDispositionArtifactRecord) error {
	if stringValue(mapping["state"]) != "source_mapped_runtime_unasserted" {
		return fmt.Errorf("declaration disposition mapping state = %q", stringValue(mapping["state"]))
	}
	actualLaneCounts := notionDispositionLaneCounts(matrix)
	ledgerLaneCounts := object(mapping["lane_cells"])
	if len(ledgerLaneCounts) != len(notionLaneNames) {
		return fmt.Errorf("declaration disposition lane count = %d, want %d", len(ledgerLaneCounts), len(notionLaneNames))
	}
	for _, lane := range notionLaneNames {
		counts := object(ledgerLaneCounts[lane])
		actual := actualLaneCounts[lane]
		if intValue(counts["mapped_unproven"]) != actual["mapped_unproven"] ||
			intValue(counts["not_applicable"]) != actual["not_applicable"] {
			return fmt.Errorf("declaration disposition %s lane counts do not reconcile", lane)
		}
	}
	roles := map[string]string{
		"api_surface":    "legacy_inventory_not_source_bound_execution_proof",
		"streams":        "legacy_stream_declarations_not_source_bound_execution_proof",
		"operations":     "legacy_operation_definitions_not_source_bound_execution_proof",
		"writes":         "legacy_write_definitions_not_source_bound_execution_proof",
		"cli_surface":    "legacy_cli_surface_not_source_bound_execution_proof",
		"sync_transport": "no_declared_sync_transport",
	}
	ledgerArtifacts := object(mapping["legacy_definition_artifacts"])
	if len(ledgerArtifacts) != len(roles) {
		return fmt.Errorf("declaration disposition artifact count = %d, want %d", len(ledgerArtifacts), len(roles))
	}
	for name, role := range roles {
		ledger := object(ledgerArtifacts[name])
		actual, exists := artifacts[name]
		if !exists || ledger["present"] != actual.Present || intValue(ledger["record_count"]) != actual.Count || stringValue(ledger["role"]) != role {
			return fmt.Errorf("declaration disposition artifact %s does not match current definition evidence", name)
		}
	}
	return nil
}

func validateNotionDispositionRetentionBoundary(boundary, retentionContract, matrix map[string]any) error {
	if stringValue(boundary["mode"]) != "retention_only_descriptor_absent" || boundary["canonical_evidence"] != false {
		return fmt.Errorf("retention boundary canonical_evidence must remain false for a descriptor-absent lock")
	}
	if object(retentionContract["source_lock"])["canonical_evidence"] != false {
		return fmt.Errorf("retention contract canonical_evidence must remain false for a descriptor-absent lock")
	}
	actualClaims := notionDispositionSourceBoundArtifactClaims()
	if intValue(boundary["source_bound_artifact_claims"]) != actualClaims || actualClaims != 0 {
		return fmt.Errorf("retention boundary source-bound artifact claims = ledger:%d actual:%d, want zero", intValue(boundary["source_bound_artifact_claims"]), actualClaims)
	}
	if !sameStrings(sortedStrings(stringsValue(boundary["mapping_restriction_ids"])), notionDispositionRestrictionIDs(matrix)) {
		return fmt.Errorf("retention boundary mapping restrictions do not preserve the current matrix evidence")
	}
	return nil
}

func validateNotionHistoricalCrosswalkMetadata(metadata, matrix map[string]any) error {
	if stringValue(metadata["path"]) != notionCrosswalkPath ||
		stringValue(metadata["state"]) != "absent_historical_metadata_not_reconstructed" ||
		intValue(metadata["crosswalk_only_rows"]) != 2 ||
		strings.TrimSpace(stringValue(metadata["reason"])) == "" {
		return fmt.Errorf("historical crosswalk state is incomplete or claims reconstruction")
	}
	if _, err := os.Stat(notionCrosswalkPath); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("historical crosswalk state expects %s to be absent, stat error=%v", notionCrosswalkPath, err)
	}
	snapshot := object(matrix["source_snapshot"])
	if stringValue(metadata["source_snapshot_ref"]) != stringValue(snapshot["source_snapshot_ref"]) ||
		stringValue(metadata["source_snapshot_commit"]) != stringValue(snapshot["source_snapshot_commit"]) {
		return fmt.Errorf("historical crosswalk snapshot identity mismatch")
	}
	for _, retained := range objectArray(snapshot["retained_files"]) {
		if stringValue(retained["path"]) != notionCrosswalkPath {
			continue
		}
		if stringValue(metadata["git_blob_sha1"]) != stringValue(retained["git_blob_sha1"]) ||
			intValue(metadata["bytes"]) != intValue(retained["bytes"]) {
			return fmt.Errorf("historical crosswalk metadata does not preserve the matrix snapshot identity")
		}
		boundary := object(matrix["source_boundary_reconciliation"])
		if intValue(boundary["crosswalk_only_rows"]) != intValue(metadata["crosswalk_only_rows"]) {
			return fmt.Errorf("historical crosswalk-only row count does not reconcile")
		}
		return nil
	}
	return fmt.Errorf("source matrix has no historical crosswalk snapshot metadata")
}

func notionDispositionLaneCounts(matrix map[string]any) map[string]map[string]int {
	counts := make(map[string]map[string]int, len(notionLaneNames))
	for _, lane := range notionLaneNames {
		counts[lane] = map[string]int{}
	}
	for _, row := range objectArray(matrix["source_operations"]) {
		for lane, cell := range object(row["lanes"]) {
			counts[lane][stringValue(object(cell)["disposition"])]++
		}
	}
	return counts
}

func notionDispositionArtifactRecords(t *testing.T) map[string]notionDispositionArtifactRecord {
	t.Helper()
	artifacts := map[string]struct {
		path string
		key  string
	}{
		"api_surface": {path: "api_surface.json", key: "endpoints"},
		"streams":     {path: "streams.json", key: "streams"},
		"operations":  {path: "operations.json", key: "operations"},
		"writes":      {path: "writes.json", key: "actions"},
		"cli_surface": {path: "cli_surface.json", key: "commands"},
	}
	records := make(map[string]notionDispositionArtifactRecord, len(artifacts)+1)
	for name, artifact := range artifacts {
		document := loadNotionJSON(t, artifact.path)
		records[name] = notionDispositionArtifactRecord{Present: true, Count: len(objectArray(document[artifact.key]))}
	}
	_, err := os.Stat("sync_transport.json")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat sync_transport.json: %v", err)
	}
	records["sync_transport"] = notionDispositionArtifactRecord{Present: err == nil}
	return records
}

func notionDispositionSourceBoundArtifactClaims() int {
	paths := []string{
		"api_surface.json",
		"streams.json",
		"operations.json",
		"writes.json",
		"cli_surface.json",
	}
	claims := 0
	for _, path := range paths {
		var document any
		contents, err := os.ReadFile(path)
		if err != nil {
			return -1
		}
		if err := json.Unmarshal(contents, &document); err != nil {
			return -1
		}
		claims += notionCountSourceOperationClaims(document)
	}
	return claims
}

func notionCountSourceOperationClaims(value any) int {
	switch value := value.(type) {
	case map[string]any:
		claims := 0
		if sourceOperation, exists := value["source_operation"]; exists {
			switch sourceOperation := sourceOperation.(type) {
			case string:
				if strings.TrimSpace(sourceOperation) != "" {
					claims++
				}
			case nil:
				// A null field does not assert a source-bound execution path.
			default:
				claims++
			}
		}
		for _, child := range value {
			claims += notionCountSourceOperationClaims(child)
		}
		return claims
	case []any:
		claims := 0
		for _, child := range value {
			claims += notionCountSourceOperationClaims(child)
		}
		return claims
	default:
		return 0
	}
}

func notionDispositionRestrictionIDs(matrix map[string]any) []string {
	ids := make([]string, 0, len(objectArray(matrix["mapping_control_restrictions"])))
	for _, restriction := range objectArray(matrix["mapping_control_restrictions"]) {
		ids = append(ids, stringValue(restriction["id"]))
	}
	return sortedStrings(ids)
}

func sortedStrings(values []string) []string {
	sorted := append([]string{}, values...)
	sort.Strings(sorted)
	return sorted
}
