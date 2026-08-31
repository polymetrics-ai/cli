package asana

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

const asanaSourceLaneMatrixPath = "sources/asana-source-lane-matrix.json"
const asanaEnabledConnectorContractPath = "enabled_connector_contract.json"

type asanaSourceLaneMatrix struct {
	Lanes            []string                   `json:"lanes"`
	SourceOperations []asanaSourceLaneMatrixRow `json:"source_operations"`
}

type asanaSourceLaneMatrixRow struct {
	SourceID    string                           `json:"source_id"`
	SourceFacts asanaSourceLaneMatrixSourceFacts `json:"source_facts"`
	Lanes       map[string]asanaSourceLaneCell   `json:"lanes"`
}

type asanaSourceLaneMatrixSourceFacts struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	OperationID string `json:"operation_id"`
	Citation    struct {
		SourceLocation string `json:"source_location"`
	} `json:"citation"`
	Pagination struct {
		State string `json:"state"`
	} `json:"pagination"`
}

type asanaSourceLaneCell struct {
	Applicability string          `json:"applicability"`
	Disposition   string          `json:"disposition"`
	Mapping       json.RawMessage `json:"mapping"`
}

type asanaSourceLaneLock struct {
	REST struct {
		SourceDocuments []struct {
			Operations []struct {
				ID             string `json:"id"`
				Method         string `json:"method"`
				Path           string `json:"path"`
				OperationID    string `json:"operation_id"`
				SourceLocation string `json:"source_location"`
			} `json:"operations"`
		} `json:"source_documents"`
	} `json:"rest"`
}

type asanaSourceLaneDescriptor struct {
	Operations []struct {
		SourceID   string          `json:"source_id"`
		Pagination json.RawMessage `json:"pagination"`
	} `json:"operations"`
}

type asanaEnabledConnectorContract struct {
	Lanes []asanaEnabledConnectorLane `json:"lanes"`
}

type asanaEnabledConnectorLane struct {
	Name      string                             `json:"name"`
	State     string                             `json:"state"`
	Artifacts []string                           `json:"artifacts"`
	Source    asanaEnabledContractSourceCoverage `json:"source"`
}

type asanaEnabledContractSourceCoverage struct {
	Partition          bool     `json:"partition"`
	OperationIDs       []string `json:"operation_ids"`
	Coverage           string   `json:"coverage"`
	Expected           int      `json:"expected"`
	Implemented        int      `json:"implemented"`
	MappedUnproven     int      `json:"mapped_unproven"`
	UnmappedMapping    int      `json:"unmapped_mapping"`
	DeferredFoundation int      `json:"deferred_foundation"`
	Unsupported        int      `json:"unsupported_with_provider_evidence"`
}

func TestAsanaSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadAsanaSourceLaneMatrix(t)
	lock := loadAsanaSourceLaneLock(t)
	descriptor := loadAsanaSourceLaneDescriptor(t)
	if err := validateAsanaSourceLaneMatrix(matrix, lock, descriptor); err != nil {
		t.Fatalf("validate Asana source lane matrix: %v", err)
	}

	// Red counterpart: a missing cell is a hidden-row regression and must never
	// be accepted just because the remaining source rows still parse.
	broken := matrix
	broken.SourceOperations = append([]asanaSourceLaneMatrixRow(nil), matrix.SourceOperations...)
	broken.SourceOperations[0].Lanes = mapsClone(matrix.SourceOperations[0].Lanes)
	delete(broken.SourceOperations[0].Lanes, "sync_transport")
	if err := validateAsanaSourceLaneMatrix(broken, lock, descriptor); err == nil || !strings.Contains(err.Error(), "missing lane cell") {
		t.Fatalf("missing-cell matrix validation error = %v, want missing lane cell", err)
	}
}

func TestAsanaEnabledContractReconcilesSourceLaneMatrix(t *testing.T) {
	matrix := loadAsanaSourceLaneMatrix(t)
	contract := loadAsanaEnabledConnectorContract(t)
	if err := validateAsanaEnabledContractSourceLaneMatrix(contract, matrix); err != nil {
		t.Fatalf("validate Asana enabled contract against source lane matrix: %v", err)
	}

	// Red counterpart: an overlay cannot hide a locked source ID by shrinking
	// its selector while leaving the source-coverage count unchanged.
	broken := contract
	broken.Lanes = append([]asanaEnabledConnectorLane(nil), contract.Lanes...)
	reverseETL := asanaEnabledContractLane(&broken, "reverse_etl")
	reverseETL.Source.OperationIDs = append([]string(nil), reverseETL.Source.OperationIDs[:len(reverseETL.Source.OperationIDs)-1]...)
	if err := validateAsanaEnabledContractSourceLaneMatrix(broken, matrix); err == nil || !strings.Contains(err.Error(), "operation IDs") {
		t.Fatalf("short reverse_etl source selector validation error = %v, want exact operation IDs failure", err)
	}
}

func loadAsanaSourceLaneMatrix(t *testing.T) asanaSourceLaneMatrix {
	t.Helper()
	raw, err := os.ReadFile(asanaSourceLaneMatrixPath)
	if err != nil {
		t.Fatalf("read source lane matrix: %v", err)
	}
	var matrix asanaSourceLaneMatrix
	if err := json.Unmarshal(raw, &matrix); err != nil {
		t.Fatalf("decode source lane matrix: %v", err)
	}
	return matrix
}

func loadAsanaSourceLaneLock(t *testing.T) asanaSourceLaneLock {
	t.Helper()
	raw, err := os.ReadFile("sources/asana-operation-source-lock.json")
	if err != nil {
		t.Fatalf("read source lock: %v", err)
	}
	var lock asanaSourceLaneLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode source lock: %v", err)
	}
	return lock
}

func loadAsanaSourceLaneDescriptor(t *testing.T) asanaSourceLaneDescriptor {
	t.Helper()
	raw, err := os.ReadFile("sources/asana-operation-descriptor.json")
	if err != nil {
		t.Fatalf("read source descriptor: %v", err)
	}
	var descriptor asanaSourceLaneDescriptor
	if err := json.Unmarshal(raw, &descriptor); err != nil {
		t.Fatalf("decode source descriptor: %v", err)
	}
	return descriptor
}

func loadAsanaEnabledConnectorContract(t *testing.T) asanaEnabledConnectorContract {
	t.Helper()
	raw, err := os.ReadFile(asanaEnabledConnectorContractPath)
	if err != nil {
		t.Fatalf("read Asana enabled connector contract: %v", err)
	}
	var contract asanaEnabledConnectorContract
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("decode Asana enabled connector contract: %v", err)
	}
	return contract
}

func validateAsanaSourceLaneMatrix(matrix asanaSourceLaneMatrix, lock asanaSourceLaneLock, descriptor asanaSourceLaneDescriptor) error {
	wantLanes := []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"}
	if !slices.Equal(matrix.Lanes, wantLanes) {
		return fmt.Errorf("lane order = %v, want %v", matrix.Lanes, wantLanes)
	}
	locked := map[string]struct {
		method, path, operationID, sourceLocation string
	}{}
	for _, document := range lock.REST.SourceDocuments {
		for _, operation := range document.Operations {
			locked[operation.ID] = struct {
				method, path, operationID, sourceLocation string
			}{operation.Method, operation.Path, operation.OperationID, operation.SourceLocation}
		}
	}
	pagination := map[string]bool{}
	for _, operation := range descriptor.Operations {
		pagination[operation.SourceID] = string(operation.Pagination) != "" && string(operation.Pagination) != "null"
	}
	if len(locked) != 249 || len(matrix.SourceOperations) != len(locked) {
		return fmt.Errorf("source rows matrix=%d lock=%d, want 249", len(matrix.SourceOperations), len(locked))
	}
	counts := map[string]map[string]int{}
	seen := map[string]bool{}
	for _, row := range matrix.SourceOperations {
		if seen[row.SourceID] {
			return fmt.Errorf("duplicate source ID %q", row.SourceID)
		}
		seen[row.SourceID] = true
		source, ok := locked[row.SourceID]
		if !ok {
			return fmt.Errorf("matrix source ID %q is absent from the lock", row.SourceID)
		}
		if row.SourceFacts.Method != source.method || row.SourceFacts.Path != source.path || row.SourceFacts.OperationID != source.operationID || row.SourceFacts.Citation.SourceLocation != source.sourceLocation {
			return fmt.Errorf("source facts drift for %q", row.SourceID)
		}
		for _, lane := range wantLanes {
			cell, ok := row.Lanes[lane]
			if !ok {
				return fmt.Errorf("missing lane cell: %s %s", row.SourceID, lane)
			}
			if cell.Applicability != "applicable" && cell.Applicability != "not_applicable" {
				return fmt.Errorf("invalid applicability for %s %s: %q", row.SourceID, lane, cell.Applicability)
			}
			if !slices.Contains([]string{"implemented", "mapped_unproven", "missing_foundation", "not_applicable"}, cell.Disposition) {
				return fmt.Errorf("invalid disposition for %s %s: %q", row.SourceID, lane, cell.Disposition)
			}
			if cell.Applicability == "not_applicable" && cell.Disposition != "not_applicable" {
				return fmt.Errorf("not-applicable cell promoted for %s %s", row.SourceID, lane)
			}
			if cell.Applicability == "applicable" && len(cell.Mapping) == 0 && cell.Disposition != "missing_foundation" {
				return fmt.Errorf("applicable cell lacks mapping evidence for %s %s", row.SourceID, lane)
			}
			if counts[lane] == nil {
				counts[lane] = map[string]int{}
			}
			counts[lane][cell.Disposition]++
		}
		if pagination[row.SourceID] != (row.SourceFacts.Pagination.State == "declared") {
			return fmt.Errorf("descriptor pagination drift for %q", row.SourceID)
		}
	}
	if len(seen) != len(locked) {
		return fmt.Errorf("matrix does not retain every locked source ID: matrix=%d lock=%d", len(seen), len(locked))
	}
	wantCounts := map[string]map[string]int{
		"direct_read":     {"implemented": 119, "not_applicable": 130},
		"direct_write":    {"implemented": 130, "not_applicable": 119},
		"binary_download": {"not_applicable": 249},
		"binary_upload":   {"implemented": 1, "not_applicable": 248},
		"etl":             {"implemented": 12, "mapped_unproven": 52, "not_applicable": 185},
		"reverse_etl":     {"implemented": 130, "not_applicable": 119},
		"sync_transport":  {"implemented": 3, "not_applicable": 246},
	}
	for lane, want := range wantCounts {
		if !equalAsanaLaneCounts(counts[lane], want) {
			return fmt.Errorf("%s disposition counts = %v, want %v", lane, counts[lane], want)
		}
	}
	return nil
}

func validateAsanaEnabledContractSourceLaneMatrix(contract asanaEnabledConnectorContract, matrix asanaSourceLaneMatrix) error {
	if len(contract.Lanes) != len(matrix.Lanes) {
		return fmt.Errorf("enabled contract lanes = %d, want %d", len(contract.Lanes), len(matrix.Lanes))
	}
	for _, laneName := range matrix.Lanes {
		lane := asanaEnabledContractLane(&contract, laneName)
		if lane == nil {
			return fmt.Errorf("enabled contract omits lane %q", laneName)
		}
		want := asanaSourceLaneMatrixCoverage(matrix, laneName)
		if lane.Source.Expected != want.expected || lane.Source.Implemented != want.implemented || lane.Source.MappedUnproven != want.mappedUnproven || lane.Source.UnmappedMapping != want.unmappedMapping || lane.Source.DeferredFoundation != want.deferredFoundation || lane.Source.Unsupported != 0 {
			return fmt.Errorf("%s source coverage = %+v, want expected=%d implemented=%d mapped_unproven=%d unmapped=%d deferred=%d unsupported=0", laneName, lane.Source, want.expected, want.implemented, want.mappedUnproven, want.unmappedMapping, want.deferredFoundation)
		}
		if lane.Source.Coverage != want.coverage {
			return fmt.Errorf("%s source coverage = %q, want %q", laneName, lane.Source.Coverage, want.coverage)
		}
		if lane.Source.Partition {
			if len(lane.Source.OperationIDs) != 0 {
				return fmt.Errorf("%s partition must not retain overlay operation IDs", laneName)
			}
			continue
		}
		if !slices.IsSorted(lane.Source.OperationIDs) || !slices.Equal(lane.Source.OperationIDs, want.operationIDs) {
			return fmt.Errorf("%s source operation IDs = %v, want exact sorted matrix IDs %v", laneName, lane.Source.OperationIDs, want.operationIDs)
		}
		if (laneName == "etl" || laneName == "reverse_etl") && (!slices.Contains(lane.Artifacts, asanaSourceLaneMatrixPath) || !slices.Contains(lane.Artifacts, "api_surface.json")) {
			return fmt.Errorf("%s artifacts must retain the source lane matrix and api surface", laneName)
		}
		if (laneName == "etl" || laneName == "reverse_etl") && lane.State != "implemented" {
			return fmt.Errorf("%s state = %q, want implemented existing runtime lane", laneName, lane.State)
		}
	}
	return nil
}

type asanaMatrixLaneCoverage struct {
	expected           int
	implemented        int
	mappedUnproven     int
	unmappedMapping    int
	deferredFoundation int
	coverage           string
	operationIDs       []string
}

func asanaSourceLaneMatrixCoverage(matrix asanaSourceLaneMatrix, laneName string) asanaMatrixLaneCoverage {
	coverage := asanaMatrixLaneCoverage{}
	for _, row := range matrix.SourceOperations {
		cell := row.Lanes[laneName]
		if cell.Applicability != "applicable" {
			continue
		}
		coverage.expected++
		coverage.operationIDs = append(coverage.operationIDs, row.SourceID)
		switch cell.Disposition {
		case "implemented":
			coverage.implemented++
		case "mapped_unproven":
			coverage.mappedUnproven++
		case "missing_foundation":
			coverage.deferredFoundation++
		default:
			coverage.unmappedMapping++
		}
	}
	sort.Strings(coverage.operationIDs)
	coverage.coverage = "partial"
	if coverage.expected == 0 {
		coverage.coverage = "not_applicable"
	} else if coverage.implemented == coverage.expected {
		coverage.coverage = "complete"
	}
	return coverage
}

func asanaEnabledContractLane(contract *asanaEnabledConnectorContract, laneName string) *asanaEnabledConnectorLane {
	for index := range contract.Lanes {
		if contract.Lanes[index].Name == laneName {
			return &contract.Lanes[index]
		}
	}
	return nil
}

func mapsClone(in map[string]asanaSourceLaneCell) map[string]asanaSourceLaneCell {
	out := make(map[string]asanaSourceLaneCell, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func equalAsanaLaneCounts(got, want map[string]int) bool {
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
