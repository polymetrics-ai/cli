package asana

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

const asanaSourceLaneMatrixPath = "sources/asana-source-lane-matrix.json"
const asanaEnabledConnectorContractPath = "enabled_connector_contract.json"
const asanaMissingFoundationPath = "missing-foundation.json"

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

const asanaSourceBackedFullRefreshETLStreamCount = 64

type asanaFullRefreshETLMapping struct {
	SourceID   string `json:"source_id"`
	Stream     string `json:"stream"`
	Schema     string `json:"schema"`
	Mode       string `json:"mode"`
	APISurface struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"api_surface"`
}

type asanaDirectReadMapping struct {
	Stream *string `json:"stream"`
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
	Operations []asanaSourceLaneDescriptorOperation `json:"operations"`
}

type asanaSourceLaneDescriptorOperation struct {
	SourceID   string                 `json:"source_id"`
	Pagination json.RawMessage        `json:"pagination"`
	Request    asanaSourceLaneRequest `json:"request"`
}

type asanaSourceLaneRequest struct {
	Path  []asanaSourceLaneRequestParameter `json:"path"`
	Query []asanaSourceLaneRequestParameter `json:"query"`
}

type asanaSourceLaneRequestParameter struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type asanaEnabledConnectorContract struct {
	Lanes []asanaEnabledConnectorLane `json:"lanes"`
}

type asanaMissingFoundationLedger struct {
	LaneMapping struct {
		ETL struct {
			FullRefreshStreamCount int      `json:"full_refresh_stream_count"`
			FullRefreshStreams     []string `json:"full_refresh_streams"`
			IncrementalStreamCount int      `json:"incremental_stream_count"`
			IncrementalStreams     []string `json:"incremental_streams"`
		} `json:"etl"`
	} `json:"lane_mapping"`
}

type asanaConnectionSpec struct {
	Properties map[string]json.RawMessage `json:"properties"`
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

func TestAsanaSourceBackedFullRefreshETLDeclarations(t *testing.T) {
	matrix := loadAsanaSourceLaneMatrix(t)
	descriptor := loadAsanaSourceLaneDescriptor(t)
	if err := validateAsanaSourceBackedFullRefreshETLDeclarations(matrix, descriptor); err != nil {
		t.Fatalf("validate Asana source-backed full-refresh ETL declarations: %v", err)
	}

	// Red counterpart: a stream schema is part of the source-to-definition
	// mapping, so a plausible-looking but uncited schema cannot be admitted.
	brokenSchema := cloneAsanaSourceLaneMatrix(matrix)
	rowIndex := firstAsanaApplicableETLRow(t, brokenSchema)
	brokenSchema.SourceOperations[rowIndex].Lanes = mapsClone(brokenSchema.SourceOperations[rowIndex].Lanes)
	brokenCell := brokenSchema.SourceOperations[rowIndex].Lanes["etl"]
	var brokenMapping map[string]any
	if err := json.Unmarshal(brokenCell.Mapping, &brokenMapping); err != nil {
		t.Fatalf("decode mutable ETL mapping: %v", err)
	}
	brokenMapping["schema"] = "schemas/not-source-backed.json"
	brokenCell.Mapping, _ = json.Marshal(brokenMapping)
	brokenSchema.SourceOperations[rowIndex].Lanes["etl"] = brokenCell
	if err := validateAsanaSourceBackedFullRefreshETLDeclarations(brokenSchema, descriptor); err == nil || !strings.Contains(err.Error(), "schema") {
		t.Fatalf("schema-drift ETL declaration error = %v, want source-bound schema rejection", err)
	}

	// A stream cannot borrow another operation's source citation even when its
	// HTTP path and response shape happen to look compatible.
	brokenSourceID := cloneAsanaSourceLaneMatrix(matrix)
	rowIndex = firstAsanaApplicableETLRow(t, brokenSourceID)
	brokenSourceID.SourceOperations[rowIndex].Lanes = mapsClone(brokenSourceID.SourceOperations[rowIndex].Lanes)
	brokenCell = brokenSourceID.SourceOperations[rowIndex].Lanes["etl"]
	if err := json.Unmarshal(brokenCell.Mapping, &brokenMapping); err != nil {
		t.Fatalf("decode mutable ETL source mapping: %v", err)
	}
	brokenMapping["source_id"] = "asana.rest.not_a_retained_operation"
	brokenCell.Mapping, _ = json.Marshal(brokenMapping)
	brokenSourceID.SourceOperations[rowIndex].Lanes["etl"] = brokenCell
	if err := validateAsanaSourceBackedFullRefreshETLDeclarations(brokenSourceID, descriptor); err == nil || !strings.Contains(err.Error(), "source ID") {
		t.Fatalf("source-ID-drift ETL declaration error = %v, want exact source ID rejection", err)
	}
}

func TestAsanaFullRefreshETLLedgerReconcilesDeclaredStreams(t *testing.T) {
	ledger := loadAsanaMissingFoundationLedger(t)
	bundle, err := engine.Load(os.DirFS(".."), "asana")
	if err != nil {
		t.Fatalf("load Asana declared stream bundle: %v", err)
	}
	wantStreams := make([]string, 0, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		wantStreams = append(wantStreams, stream.Name)
	}
	sort.Strings(wantStreams)
	gotStreams := append([]string(nil), ledger.LaneMapping.ETL.FullRefreshStreams...)
	sort.Strings(gotStreams)
	if ledger.LaneMapping.ETL.FullRefreshStreamCount != asanaSourceBackedFullRefreshETLStreamCount || !slices.Equal(gotStreams, wantStreams) {
		t.Fatalf("full-refresh ledger count/streams = %d/%v, want %d/%v", ledger.LaneMapping.ETL.FullRefreshStreamCount, gotStreams, asanaSourceBackedFullRefreshETLStreamCount, wantStreams)
	}
	if ledger.LaneMapping.ETL.IncrementalStreamCount != 1 || !slices.Equal(ledger.LaneMapping.ETL.IncrementalStreams, []string{"tasks"}) {
		t.Fatalf("incremental ledger = count:%d streams:%v, want only source-backed tasks event transport", ledger.LaneMapping.ETL.IncrementalStreamCount, ledger.LaneMapping.ETL.IncrementalStreams)
	}
}

func TestAsanaDeclaredStreamConfigReferencesExistInConnectionSpec(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), "asana")
	if err != nil {
		t.Fatalf("load Asana declared stream bundle: %v", err)
	}
	spec := loadAsanaConnectionSpec(t)
	if err := validateAsanaDeclaredStreamConfigReferences(bundle.Streams, spec); err != nil {
		t.Fatalf("validate Asana stream config references: %v", err)
	}

	// Red counterpart: an ETL stream cannot silently interpolate an undeclared
	// source scope just because the generic config object accepts extra keys.
	broken := spec
	broken.Properties = make(map[string]json.RawMessage, len(spec.Properties))
	for key, value := range spec.Properties {
		broken.Properties[key] = value
	}
	delete(broken.Properties, "goal_id")
	if err := validateAsanaDeclaredStreamConfigReferences(bundle.Streams, broken); err == nil || !strings.Contains(err.Error(), "goal_id") {
		t.Fatalf("missing stream config declaration error = %v, want goal_id rejection", err)
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

func loadAsanaMissingFoundationLedger(t *testing.T) asanaMissingFoundationLedger {
	t.Helper()
	raw, err := os.ReadFile(asanaMissingFoundationPath)
	if err != nil {
		t.Fatalf("read Asana missing-foundation ledger: %v", err)
	}
	var ledger asanaMissingFoundationLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode Asana missing-foundation ledger: %v", err)
	}
	return ledger
}

func loadAsanaConnectionSpec(t *testing.T) asanaConnectionSpec {
	t.Helper()
	raw, err := os.ReadFile("spec.json")
	if err != nil {
		t.Fatalf("read Asana connection spec: %v", err)
	}
	var spec asanaConnectionSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("decode Asana connection spec: %v", err)
	}
	return spec
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
		"etl":             {"implemented": 64, "not_applicable": 185},
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

func validateAsanaSourceBackedFullRefreshETLDeclarations(matrix asanaSourceLaneMatrix, descriptor asanaSourceLaneDescriptor) error {
	bundle, err := engine.Load(os.DirFS(".."), "asana")
	if err != nil {
		return fmt.Errorf("load Asana definition bundle: %w", err)
	}
	streams := make(map[string]engine.StreamSpec, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		if _, duplicate := streams[stream.Name]; duplicate {
			return fmt.Errorf("duplicate declared stream %q", stream.Name)
		}
		streams[stream.Name] = stream
	}
	descriptors := make(map[string]asanaSourceLaneDescriptorOperation, len(descriptor.Operations))
	for _, operation := range descriptor.Operations {
		descriptors[operation.SourceID] = operation
	}

	seenStreams := map[string]string{}
	applicableETL := 0
	boundedDirectOperations := 0
	boundedDirectStreams := 0
	for _, row := range matrix.SourceOperations {
		etl := row.Lanes["etl"]
		if etl.Applicability != "applicable" {
			continue
		}
		applicableETL++
		if etl.Disposition != "implemented" {
			return fmt.Errorf("source-backed ETL %q disposition = %q, want implemented", row.SourceID, etl.Disposition)
		}
		var mapping asanaFullRefreshETLMapping
		if err := json.Unmarshal(etl.Mapping, &mapping); err != nil {
			return fmt.Errorf("decode ETL mapping for %q: %w", row.SourceID, err)
		}
		if mapping.SourceID != row.SourceID {
			return fmt.Errorf("ETL mapping source ID for %q = %q, want exact retained source ID", row.SourceID, mapping.SourceID)
		}
		descriptorOperation, ok := descriptors[row.SourceID]
		if !ok || len(descriptorOperation.Pagination) == 0 || string(descriptorOperation.Pagination) == "null" {
			return fmt.Errorf("ETL mapping source ID %q lacks retained continuation facts", row.SourceID)
		}
		if mapping.Mode != "full_refresh" {
			return fmt.Errorf("ETL mapping mode for %q = %q, want full_refresh", row.SourceID, mapping.Mode)
		}
		if mapping.APISurface.Method != row.SourceFacts.Method || mapping.APISurface.Path != row.SourceFacts.Path {
			return fmt.Errorf("ETL mapping API surface for %q = %s %s, want %s %s", row.SourceID, mapping.APISurface.Method, mapping.APISurface.Path, row.SourceFacts.Method, row.SourceFacts.Path)
		}
		stream, ok := streams[mapping.Stream]
		if !ok {
			return fmt.Errorf("ETL mapping source ID %q names absent stream %q", row.SourceID, mapping.Stream)
		}
		if stream.SchemaRef != mapping.Schema {
			return fmt.Errorf("ETL mapping schema for %q = %q, stream %q declares %q", row.SourceID, mapping.Schema, mapping.Stream, stream.SchemaRef)
		}
		if stream.Records.Path != "data" || stream.Incremental != nil {
			return fmt.Errorf("ETL stream %q for %q must use records.data with no incremental checkpoint", mapping.Stream, row.SourceID)
		}
		if !asanaSourcePathMatchesStreamTemplate(row.SourceFacts.Path, stream.Path) {
			return fmt.Errorf("ETL stream %q path %q does not match retained source path %q", mapping.Stream, stream.Path, row.SourceFacts.Path)
		}
		for _, query := range descriptorOperation.Request.Query {
			if !query.Required {
				continue
			}
			if _, present := stream.Query[query.Name]; !present {
				return fmt.Errorf("ETL stream %q for %q omits required retained query %q", mapping.Stream, row.SourceID, query.Name)
			}
		}
		if prior, duplicate := seenStreams[mapping.Stream]; duplicate {
			return fmt.Errorf("ETL stream %q is mapped by both %q and %q", mapping.Stream, prior, row.SourceID)
		}
		seenStreams[mapping.Stream] = row.SourceID

		var directMapping asanaDirectReadMapping
		if err := json.Unmarshal(row.Lanes["direct_read"].Mapping, &directMapping); err != nil {
			return fmt.Errorf("decode direct-read mapping for %q: %w", row.SourceID, err)
		}
		if row.Lanes["direct_read"].Disposition != "implemented" {
			return fmt.Errorf("direct-read lane for %q = %q, want separately implemented bounded read", row.SourceID, row.Lanes["direct_read"].Disposition)
		}
		if directMapping.Stream == nil {
			boundedDirectOperations++
		} else {
			boundedDirectStreams++
		}
	}
	if applicableETL != asanaSourceBackedFullRefreshETLStreamCount || len(seenStreams) != asanaSourceBackedFullRefreshETLStreamCount || len(bundle.Streams) != asanaSourceBackedFullRefreshETLStreamCount {
		return fmt.Errorf("full-refresh source/stream count = %d/%d/%d, want %d", applicableETL, len(seenStreams), len(bundle.Streams), asanaSourceBackedFullRefreshETLStreamCount)
	}
	if boundedDirectOperations != 52 || boundedDirectStreams != 12 {
		return fmt.Errorf("ETL counterparts direct-read partition = operations:%d streams:%d, want 52 bounded operations and 12 bounded streams", boundedDirectOperations, boundedDirectStreams)
	}
	return nil
}

func validateAsanaDeclaredStreamConfigReferences(streams []engine.StreamSpec, spec asanaConnectionSpec) error {
	for _, stream := range streams {
		for _, key := range asanaConfigTemplateKeys(stream.Path) {
			if _, declared := spec.Properties[key]; !declared {
				return fmt.Errorf("stream %q path references undeclared config key %q", stream.Name, key)
			}
		}
		for queryName, query := range stream.Query {
			for _, key := range asanaConfigTemplateKeys(query.Template) {
				if _, declared := spec.Properties[key]; !declared {
					return fmt.Errorf("stream %q query %q references undeclared config key %q", stream.Name, queryName, key)
				}
			}
		}
	}
	return nil
}

func asanaConfigTemplateKeys(value string) []string {
	const prefix = "{{ config."
	const suffix = " }}"
	var keys []string
	for {
		start := strings.Index(value, prefix)
		if start < 0 {
			return keys
		}
		value = value[start+len(prefix):]
		end := strings.Index(value, suffix)
		if end < 0 {
			return keys
		}
		keys = append(keys, value[:end])
		value = value[end+len(suffix):]
	}
}

func asanaSourcePathMatchesStreamTemplate(sourcePath, streamPath string) bool {
	sourceSegments := strings.Split(strings.Trim(sourcePath, "/"), "/")
	streamSegments := strings.Split(strings.Trim(streamPath, "/"), "/")
	if len(sourceSegments) != len(streamSegments) {
		return false
	}
	for index, sourceSegment := range sourceSegments {
		streamSegment := streamSegments[index]
		if strings.HasPrefix(sourceSegment, "{") && strings.HasSuffix(sourceSegment, "}") {
			if strings.HasPrefix(streamSegment, "{{ config.") && strings.HasSuffix(streamSegment, " }}") || streamSegment == "{{ fanout.id }}" {
				continue
			}
			return false
		}
		if sourceSegment != streamSegment {
			return false
		}
	}
	return true
}

func cloneAsanaSourceLaneMatrix(matrix asanaSourceLaneMatrix) asanaSourceLaneMatrix {
	clone := matrix
	clone.SourceOperations = append([]asanaSourceLaneMatrixRow(nil), matrix.SourceOperations...)
	return clone
}

func firstAsanaApplicableETLRow(t *testing.T, matrix asanaSourceLaneMatrix) int {
	t.Helper()
	for index, row := range matrix.SourceOperations {
		if row.Lanes["etl"].Applicability == "applicable" {
			return index
		}
	}
	t.Fatal("Asana matrix has no applicable ETL row")
	return -1
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
