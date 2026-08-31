package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// TestSourceProjectionDescriptorFreeRetentionFromFrozenSourceEvidence proves
// that an exact, source-only seven-lane contract can preserve mapping
// accounting without importing a historical canonical descriptor. The helper
// reads the actual checked-in source locks and matrices; it does not invent a
// request, response, command, stream, or runtime artifact.
func TestSourceProjectionDescriptorFreeRetentionFromFrozenSourceEvidence(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	fsys := os.DirFS(defsRoot)

	for _, connector := range []string{"jira", "sentry", "vercel"} {
		t.Run(connector, func(t *testing.T) {
			bundle, err := engine.Load(fsys, connector)
			if err != nil {
				t.Fatalf("load %s bundle: %v", connector, err)
			}
			contract := descriptorFreeRetentionContractFromFrozenMatrix(t, fsys, connector)
			bundle.EnabledContract = &contract

			if findings := checkEnabledConnectorContract(fsys, bundle); len(findings) != 0 {
				t.Fatalf("%s exact retained-source contract findings = %+v", connector, findings)
			}
			if findings := checkSourceProjection(fsys, bundle); len(findings) != 0 {
				t.Fatalf("%s descriptor-free retention findings = %+v", connector, findings)
			}
		})
	}
}

func TestSourceProjectionDescriptorFreeRetentionStillRequiresDescriptorForExecutableClaims(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	fsys := os.DirFS(defsRoot)
	bundle, err := engine.Load(fsys, "jira")
	if err != nil {
		t.Fatalf("load Jira bundle: %v", err)
	}
	contract := descriptorFreeRetentionContractFromFrozenMatrix(t, fsys, "jira")
	directRead := descriptorFreeRetentionLane(t, &contract, "direct_read")
	directRead.State = connectors.EnabledLaneImplemented
	directRead.Source.Implemented = directRead.Source.Expected
	directRead.Source.MappedUnproven = 0
	directRead.Source.DeferredFoundation = 0
	directRead.Source.Unsupported = 0
	directRead.Source.Coverage = connectors.EnabledCoverageComplete
	directRead.Artifacts = []string{"operations.json"}
	bundle.EnabledContract = &contract

	findings := checkSourceProjection(fsys, bundle)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "canonical source descriptor is missing") {
		t.Fatalf("descriptor-less executable retention findings = %+v, want canonical descriptor failure", findings)
	}
}

func TestSourceProjectionDescriptorFreeRetentionRejectsIncompleteExactSourceClaims(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	fsys := os.DirFS(defsRoot)
	bundle, err := engine.Load(fsys, "vercel")
	if err != nil {
		t.Fatalf("load Vercel bundle: %v", err)
	}
	contract := descriptorFreeRetentionContractFromFrozenMatrix(t, fsys, "vercel")
	directRead := descriptorFreeRetentionLane(t, &contract, "direct_read")
	directRead.Source.OperationIDs = directRead.Source.OperationIDs[1:]
	directRead.Source.Expected--
	directRead.Source.MappedUnproven--
	bundle.EnabledContract = &contract

	findings := checkSourceProjection(fsys, bundle)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "canonical source descriptor is missing") {
		t.Fatalf("incomplete descriptor-free retention findings = %+v, want canonical descriptor failure", findings)
	}
}

func TestSourceProjectionDescriptorFreeRetentionRejectsDuplicateExactSourceClaims(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	fsys := os.DirFS(defsRoot)
	bundle, err := engine.Load(fsys, "sentry")
	if err != nil {
		t.Fatalf("load Sentry bundle: %v", err)
	}
	contract := descriptorFreeRetentionContractFromFrozenMatrix(t, fsys, "sentry")
	directRead := descriptorFreeRetentionLane(t, &contract, "direct_read")
	if len(directRead.Source.OperationIDs) < 2 {
		t.Fatalf("Sentry direct-read retention fixture has %d IDs, want at least two", len(directRead.Source.OperationIDs))
	}
	directRead.Source.OperationIDs[1] = directRead.Source.OperationIDs[0]
	bundle.EnabledContract = &contract

	findings := checkSourceProjection(fsys, bundle)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "canonical source descriptor is missing") {
		t.Fatalf("duplicate descriptor-free retention findings = %+v, want canonical descriptor failure", findings)
	}
}

type descriptorFreeRetentionMatrix struct {
	SourceOperations []descriptorFreeRetentionMatrixOperation `json:"source_operations"`
	Operations       []descriptorFreeRetentionMatrixOperation `json:"operations"`
}

type descriptorFreeRetentionMatrixOperation struct {
	SourceID string                                       `json:"source_id"`
	Lanes    map[string]descriptorFreeRetentionMatrixCell `json:"lanes"`
	Cells    []descriptorFreeRetentionNamedMatrixCell     `json:"cells"`
}

type descriptorFreeRetentionMatrixCell struct {
	Applicability string `json:"applicability"`
	Disposition   string `json:"disposition"`
}

type descriptorFreeRetentionNamedMatrixCell struct {
	Lane  string `json:"lane"`
	State string `json:"state"`
}

// descriptorFreeRetentionContractFromFrozenMatrix builds a test-only contract
// from retained connector evidence. The primary partition chooses one actual
// non-N/A lane per source operation; all other cited lanes remain exact-ID
// overlays. A source row with no applicable lane is retained as a cited
// provider-N/A direct-read partition member rather than being dropped.
func descriptorFreeRetentionContractFromFrozenMatrix(t *testing.T, fsys fs.FS, connector string) connectors.EnabledConnectorContract {
	t.Helper()
	lockPath := path.Join("sources", connector+"-operation-source-lock.json")
	lock, err := loadEnabledContractSourceLock(fsys, connector, lockPath)
	if err != nil {
		t.Fatalf("load %s source lock: %v", connector, err)
	}
	operations := enabledContractSourceOperations(lock)
	if len(operations) != lock.Counts.REST {
		t.Fatalf("%s retained source operation count = %d, want lock REST count %d", connector, len(operations), lock.Counts.REST)
	}

	raw, err := fs.ReadFile(fsys, path.Join(connector, "sources", connector+"-source-lane-matrix.json"))
	if err != nil {
		t.Fatalf("read %s source-lane matrix: %v", connector, err)
	}
	var matrix descriptorFreeRetentionMatrix
	if err := json.Unmarshal(raw, &matrix); err != nil {
		t.Fatalf("decode %s source-lane matrix: %v", connector, err)
	}
	rows := matrix.SourceOperations
	if len(rows) == 0 {
		rows = matrix.Operations
	}
	if len(rows) != len(operations) {
		t.Fatalf("%s matrix source rows = %d, want retained operation count %d", connector, len(rows), len(operations))
	}

	byID := make(map[string]map[string]string, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.SourceID) == "" {
			t.Fatalf("%s matrix contains an empty source ID", connector)
		}
		if _, exists := byID[row.SourceID]; exists {
			t.Fatalf("%s matrix duplicates source ID %q", connector, row.SourceID)
		}
		byID[row.SourceID] = descriptorFreeRetentionCellStates(row)
	}

	lanes := descriptorFreeRetentionEmptyLanes(connector, lockPath, lock.Rest.SourceURL)
	for _, operation := range operations {
		states, exists := byID[operation.ID]
		if !exists {
			t.Fatalf("%s retained source ID %q is absent from source-lane matrix", connector, operation.ID)
		}
		primaryLane, primaryState := descriptorFreeRetentionPrimaryCell(states)
		descriptorFreeRetentionAddCell(t, lanes[primaryLane], operation.ID, primaryState, true)
		for laneName, state := range states {
			if laneName == primaryLane {
				continue
			}
			descriptorFreeRetentionAddCell(t, lanes[laneName], operation.ID, state, false)
		}
	}
	for _, lane := range lanes {
		descriptorFreeRetentionFinalizeLane(lane)
	}

	result := connectors.EnabledConnectorContract{
		SchemaVersion: 1,
		Connector:     connector,
		RetentionOnly: true,
		SourceLock: connectors.EnabledContractSourceLock{
			Path:              lockPath,
			SHA256:            lock.Rest.SHA256,
			Bytes:             lock.Rest.Bytes,
			CanonicalEvidence: lock.Rest.CanonicalEvidence,
		},
		Lanes: make([]connectors.EnabledConnectorLane, 0, len(lanes)),
	}
	for _, name := range descriptorFreeRetentionLaneOrder {
		result.Lanes = append(result.Lanes, *lanes[name])
	}
	return result
}

var descriptorFreeRetentionLaneOrder = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

func descriptorFreeRetentionEmptyLanes(connector, artifact, sourceURL string) map[string]*connectors.EnabledConnectorLane {
	if strings.TrimSpace(sourceURL) == "" {
		sourceURL = "https://example.invalid/retained-source"
	}
	lanes := make(map[string]*connectors.EnabledConnectorLane, len(descriptorFreeRetentionLaneOrder))
	for _, name := range descriptorFreeRetentionLaneOrder {
		lanes[name] = &connectors.EnabledConnectorLane{
			Name:      name,
			State:     connectors.EnabledLaneUnsupported,
			Reason:    "The frozen source-lane matrix records this provider lane as not applicable.",
			Citations: []connectors.EnabledContractCitation{{URL: sourceURL, Location: "source-lane-matrix"}},
			Artifacts: []string{artifact},
			Source: connectors.EnabledContractSourceCoverage{
				Coverage: connectors.EnabledCoverageNotApplicable,
			},
		}
	}
	return lanes
}

func descriptorFreeRetentionCellStates(row descriptorFreeRetentionMatrixOperation) map[string]string {
	states := make(map[string]string, len(descriptorFreeRetentionLaneOrder))
	for laneName, cell := range row.Lanes {
		if cell.Applicability != "applicable" || cell.Disposition == "not_applicable" {
			continue
		}
		states[laneName] = cell.Disposition
	}
	for _, cell := range row.Cells {
		if cell.State == "not_applicable" {
			continue
		}
		states[cell.Lane] = cell.State
	}
	return states
}

func descriptorFreeRetentionPrimaryCell(states map[string]string) (string, string) {
	for _, laneName := range []string{"direct_read", "reverse_etl", "direct_write", "binary_download", "binary_upload", "etl", "sync_transport"} {
		if state, exists := states[laneName]; exists {
			return laneName, state
		}
	}
	return "direct_read", "unsupported_with_provider_evidence"
}

func descriptorFreeRetentionAddCell(t *testing.T, lane *connectors.EnabledConnectorLane, sourceID, disposition string, partition bool) {
	t.Helper()
	if lane == nil {
		t.Fatalf("source ID %q maps to an unknown lane", sourceID)
	}
	if partition {
		lane.Source.Partition = true
	}
	lane.Source.OperationIDs = append(lane.Source.OperationIDs, sourceID)
	lane.Source.Expected++
	switch disposition {
	case "mapped_unproven":
		lane.Source.MappedUnproven++
	case "missing_foundation", "deferred_foundation":
		lane.Source.DeferredFoundation++
	case "unsupported_with_provider_evidence":
		lane.Source.Unsupported++
	default:
		t.Fatalf("source ID %q lane %q has non-retainable disposition %q", sourceID, lane.Name, disposition)
	}
}

func descriptorFreeRetentionFinalizeLane(lane *connectors.EnabledConnectorLane) {
	sort.Strings(lane.Source.OperationIDs)
	if lane.Source.Expected == 0 {
		return
	}
	lane.Source.Coverage = connectors.EnabledCoveragePartial
	switch {
	case lane.Source.MappedUnproven > 0:
		lane.State = connectors.EnabledLaneMappedUnproven
		lane.Reason = "The frozen source-lane matrix retains mapped source semantics without an executable declaration."
	case lane.Source.DeferredFoundation > 0:
		lane.State = connectors.EnabledLaneDeferred
		lane.Reason = "The frozen source-lane matrix retains a named foundation gap without an executable declaration."
	default:
		lane.State = connectors.EnabledLaneUnsupported
		lane.Reason = "The frozen source-lane matrix retains provider-evidenced non-applicability without an executable declaration."
	}
}

func descriptorFreeRetentionLane(t *testing.T, contract *connectors.EnabledConnectorContract, name string) *connectors.EnabledConnectorLane {
	t.Helper()
	for index := range contract.Lanes {
		if contract.Lanes[index].Name == name {
			return &contract.Lanes[index]
		}
	}
	t.Fatalf("contract omits lane %q", name)
	return nil
}

func TestDescriptorFreeRetentionSourceIDsStaySourceOnly(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	contract := descriptorFreeRetentionContractFromFrozenMatrix(t, os.DirFS(defsRoot), "sentry")
	directRead := descriptorFreeRetentionLane(t, &contract, "direct_read")
	if !slices.Contains(directRead.Source.OperationIDs, "sentry.rest.List a Project's Issues") {
		t.Fatalf("Sentry ordinary-space source ID was not retained: %v", directRead.Source.OperationIDs)
	}
	for _, lane := range contract.Lanes {
		if lane.State == connectors.EnabledLaneImplemented || lane.Source.Implemented != 0 || lane.Source.UnmappedMapping != 0 {
			t.Fatalf("%s is not source-only: %+v", lane.Name, lane)
		}
		if len(lane.Artifacts) != 1 || lane.Artifacts[0] != contract.SourceLock.Path || lane.Transport != nil || len(lane.Warehouse) != 0 {
			t.Fatalf("%s retention-only evidence has runtime binding: %+v", lane.Name, lane)
		}
	}
	if err := contract.ValidateRetentionOnly(); err != nil {
		t.Fatalf("Sentry source-only contract must validate: %v", err)
	}
	if got := fmt.Sprintf("%d/%d", directRead.Source.MappedUnproven, directRead.Source.Expected); got == "0/0" {
		t.Fatal("Sentry source-only contract unexpectedly has no direct-read evidence")
	}
}
