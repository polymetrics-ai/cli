package circleci

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

const circleCIMissingFoundationPath = "missing-foundation.json"

type circleCIMissingFoundationLedger struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	Purpose       string `json:"purpose"`
	SourceLock    struct {
		Path      string `json:"path"`
		SourceURL string `json:"source_url"`
		SHA256    string `json:"sha256"`
	} `json:"source_lock"`
	Foundations          []circleCIMissingFoundation `json:"foundations"`
	MappingContractDebts []circleCIMappingDebt       `json:"mapping_contract_debts"`
}

type circleCIMissingFoundation struct {
	ID                  string `json:"id"`
	State               string `json:"state"`
	AtlasCapability     string `json:"atlas_capability"`
	AtlasClassification string `json:"atlas_classification"`
	AffectedLane        string `json:"affected_lane"`
	SourceIDs           []struct {
		ID             string `json:"id"`
		Method         string `json:"method"`
		Path           string `json:"path"`
		SourceLocation string `json:"source_location"`
	} `json:"source_ids"`
	Reason       string `json:"reason"`
	RuntimeClaim string `json:"runtime_claim"`
}

type circleCIMappingDebt struct {
	ID                          string                              `json:"id"`
	State                       string                              `json:"state"`
	AffectedLanes               []string                            `json:"affected_lanes"`
	PrimarySourceLock           string                              `json:"primary_source_lock"`
	RetainedMappingContract     string                              `json:"retained_mapping_contract"`
	SourceRows                  int                                 `json:"source_rows"`
	CanonicalEvidence           bool                                `json:"canonical_evidence"`
	InheritedArtifactMismatches []circleCIInheritedArtifactMismatch `json:"inherited_artifact_mismatches"`
	RuntimeClaim                string                              `json:"runtime_claim"`
}

type circleCIInheritedArtifactMismatch struct {
	ID                string `json:"id"`
	State             string `json:"state"`
	SourceLaneMatrix  string `json:"source_lane_matrix"`
	MissingArtifact   string `json:"missing_artifact"`
	LegacyDisposition string `json:"legacy_disposition"`
	ArtifactLinks     []struct {
		Artifact          string `json:"artifact"`
		Record            string `json:"record"`
		SourceOperationID string `json:"source_operation_id"`
		Lane              string `json:"lane"`
		MatrixState       string `json:"matrix_state"`
	} `json:"artifact_links"`
	RuntimeClaim string `json:"runtime_claim"`
}

func TestCircleCIMissingFoundationLedgerBindsOnlyActualNonExecutingGaps(t *testing.T) {
	ledger := loadCircleCIMissingFoundationLedger(t)
	lock := loadCircleCISourceLock(t)
	matrix := loadCircleCISourceLaneMatrix(t)

	if ledger.SchemaVersion != 1 || ledger.Connector != "circleci" || strings.TrimSpace(ledger.Purpose) == "" {
		t.Fatalf("CircleCI missing-foundation identity = %#v", ledger)
	}
	if ledger.SourceLock.Path != circleCISourceLockPath ||
		ledger.SourceLock.SourceURL != lock.REST.SourceURL ||
		ledger.SourceLock.SHA256 != lock.REST.SHA256 {
		t.Fatalf("CircleCI missing-foundation source lock = %#v, want current pinned lock", ledger.SourceLock)
	}
	if got, want := len(ledger.Foundations), 1; got != want {
		t.Fatalf("CircleCI missing-foundation entries = %d, want one receiver gap", got)
	}

	foundation := ledger.Foundations[0]
	if foundation.ID != "circleci-inbound-webhook-receiver-r1" ||
		foundation.State != "recorded_only_requires_captain_approval_before_implementation" ||
		foundation.AtlasCapability != "transport.sync-contract.v1" ||
		foundation.AtlasClassification != "actual_gap" ||
		foundation.AffectedLane != "sync_transport" {
		t.Fatalf("CircleCI webhook foundation = %#v, want exact recorded receiver gap", foundation)
	}

	wantWebhookIDs := []string{
		"circleci.rest.createWebhook",
		"circleci.rest.updateWebhook",
	}
	gotWebhookIDs := make([]string, 0, len(foundation.SourceIDs))
	contract, err := circleCIParameterContract(lock)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range foundation.SourceIDs {
		gotWebhookIDs = append(gotWebhookIDs, source.ID)
		operation := findCircleCISourceOperation(t, lock, source.ID)
		if source.Method != operation.Method || source.Path != operation.Path || source.SourceLocation != operation.SourceLocation {
			t.Fatalf("CircleCI webhook source %q = %#v, want exact locked identity %#v", source.ID, source, operation)
		}
		document, err := circleCIOperationDocumentFor(operation)
		if err != nil {
			t.Fatal(err)
		}
		if registered, err := circleCIWebhookRegistrationEvidence(contract, operation, document); err != nil || !registered {
			t.Fatalf("CircleCI webhook source %q registration evidence = %t, %v; want true, nil", source.ID, registered, err)
		}
		cell := findCircleCIMatrixCell(t, findCircleCIMatrixOperation(t, &matrix, source.ID), "sync_transport")
		if cell.State != "missing_foundation" || cell.SourceEvidence == nil || cell.SourceEvidence.Kind != circleCIWebhookRegistrationEvidenceKind {
			t.Fatalf("CircleCI webhook matrix %q = %+v, want source-backed missing-foundation", source.ID, cell)
		}
	}
	slices.Sort(gotWebhookIDs)
	slices.Sort(wantWebhookIDs)
	if !slices.Equal(gotWebhookIDs, wantWebhookIDs) {
		t.Fatalf("CircleCI webhook source IDs = %v, want %v", gotWebhookIDs, wantWebhookIDs)
	}
	claim := strings.ToLower(foundation.RuntimeClaim)
	if !strings.Contains(claim, "no inbound circleci webhook receiver") ||
		!strings.Contains(claim, "executable sync transport is claimed") {
		t.Fatalf("CircleCI webhook runtime claim = %q, want explicit non-execution boundary", foundation.RuntimeClaim)
	}

	if got, want := len(ledger.MappingContractDebts), 1; got != want {
		t.Fatalf("CircleCI mapping-contract debts = %d, want one descriptor debt", got)
	}
	debt := ledger.MappingContractDebts[0]
	if debt.ID != "circleci-retained-source-descriptor-r1" ||
		debt.State != "recorded_mapping_contract_debt" ||
		!slices.Equal(debt.AffectedLanes, circleCILanes) ||
		debt.PrimarySourceLock != circleCISourceLockPath ||
		debt.RetainedMappingContract != "sources/circleci-retained-mapping-contract.json" ||
		debt.SourceRows != lock.Counts.Total ||
		debt.CanonicalEvidence {
		t.Fatalf("CircleCI descriptor debt = %#v, want exact descriptor-free retained-source boundary", debt)
	}
	debtClaim := strings.ToLower(debt.RuntimeClaim)
	if !strings.Contains(debtClaim, "no source-bound command") ||
		!strings.Contains(debtClaim, "no executor") ||
		!strings.Contains(debtClaim, "no executable declaration") {
		t.Fatalf("CircleCI descriptor-debt runtime claim = %q, want explicit non-execution boundary", debt.RuntimeClaim)
	}
	assertCircleCIInheritedNonExecutionMismatch(t, debt.InheritedArtifactMismatches, matrix)
}

func TestCircleCIDescriptorDebtPreventsSourceBoundArtifactPromotion(t *testing.T) {
	ledger := loadCircleCIMissingFoundationLedger(t)
	if len(ledger.MappingContractDebts) != 1 || ledger.MappingContractDebts[0].CanonicalEvidence {
		t.Fatalf("CircleCI descriptor debt = %#v, want descriptor-free retention boundary", ledger.MappingContractDebts)
	}

	bundle, err := engine.Load(os.DirFS(".."), "circleci")
	if err != nil {
		t.Fatalf("load CircleCI declaration bundle: %v", err)
	}
	if bundle.CLISurface != nil || bundle.SyncTransport != nil || len(bundle.Operations) != 0 {
		t.Fatalf("CircleCI descriptor-free bundle unexpectedly exposes public/runtime declarations: cli=%t sync=%t operations=%d", bundle.CLISurface != nil, bundle.SyncTransport != nil, len(bundle.Operations))
	}
	for _, path := range []string{"operations.json", "cli_surface.json", "sync_transport.json", "enabled_connector_contract.json"} {
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("CircleCI descriptor-free artifact %s stat error = %v, want absent", path, err)
		}
	}

	for _, path := range []string{"api_surface.json", "streams.json", "writes.json"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read CircleCI retained artifact %s: %v", path, err)
		}
		var artifact any
		if err := json.Unmarshal(raw, &artifact); err != nil {
			t.Fatalf("decode CircleCI retained artifact %s: %v", path, err)
		}
		if location, found := circleCISourceOperationField(artifact, path); found {
			t.Fatalf("CircleCI descriptor-free artifact %s unexpectedly promotes source_operation at %s", path, location)
		}
	}
}

func loadCircleCIMissingFoundationLedger(t *testing.T) circleCIMissingFoundationLedger {
	t.Helper()
	raw, err := os.ReadFile(circleCIMissingFoundationPath)
	if err != nil {
		t.Fatalf("read CircleCI missing-foundation ledger: %v", err)
	}
	var ledger circleCIMissingFoundationLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode CircleCI missing-foundation ledger: %v", err)
	}
	return ledger
}

func assertCircleCIInheritedNonExecutionMismatch(t *testing.T, mismatches []circleCIInheritedArtifactMismatch, matrix circleCISourceLaneMatrix) {
	t.Helper()
	if got, want := len(mismatches), 1; got != want {
		t.Fatalf("CircleCI inherited artifact mismatches = %d, want one precise record", got)
	}
	mismatch := mismatches[0]
	if mismatch.ID != "circleci-legacy-stream-sync-backlinks-r1" ||
		mismatch.State != "recorded_nonexecuting_legacy_mismatch" ||
		mismatch.SourceLaneMatrix != circleCISourceMatrixPath ||
		mismatch.MissingArtifact != "sync_transport.json" ||
		mismatch.LegacyDisposition != "sources/circleci-declaration-disposition.json" {
		t.Fatalf("CircleCI inherited artifact mismatch = %#v, want exact retained non-execution record", mismatch)
	}
	wantLinks := map[string]string{
		"pipelines":                 "circleci.rest.listPipelinesForProject",
		"workflows":                 "circleci.rest.listWorkflowsByPipelineId",
		"jobs":                      "circleci.rest.listWorkflowJobs",
		"schedules":                 "circleci.rest.listSchedulesForProject",
		"insights_workflow_summary": "circleci.rest.getProjectWorkflowMetrics",
	}
	if got, want := len(mismatch.ArtifactLinks), len(wantLinks); got != want {
		t.Fatalf("CircleCI inherited sync links = %d, want %d", got, want)
	}
	for _, link := range mismatch.ArtifactLinks {
		if link.Artifact != "streams.json" || link.Lane != "sync_transport" || link.MatrixState != "not_applicable" || wantLinks[link.Record] != link.SourceOperationID {
			t.Fatalf("CircleCI inherited sync link = %#v, want exact non-executing legacy backlink", link)
		}
		delete(wantLinks, link.Record)
		cell := findCircleCIMatrixCell(t, findCircleCIMatrixOperation(t, &matrix, link.SourceOperationID), link.Lane)
		if cell.State != link.MatrixState {
			t.Fatalf("CircleCI inherited sync link %s/%s matrix state = %q, want %q", link.SourceOperationID, link.Lane, cell.State, link.MatrixState)
		}
	}
	if len(wantLinks) != 0 {
		t.Fatalf("CircleCI inherited sync links are missing records: %v", wantLinks)
	}
	claim := strings.ToLower(mismatch.RuntimeClaim)
	if !strings.Contains(claim, "does not declare or enable sync transport") {
		t.Fatalf("CircleCI inherited mismatch runtime claim = %q, want non-execution boundary", mismatch.RuntimeClaim)
	}
}

func circleCISourceOperationField(value any, location string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			next := fmt.Sprintf("%s.%s", location, key)
			if key == "source_operation" {
				return next, true
			}
			if foundAt, found := circleCISourceOperationField(nested, next); found {
				return foundAt, true
			}
		}
	case []any:
		for index, nested := range typed {
			if foundAt, found := circleCISourceOperationField(nested, fmt.Sprintf("%s[%d]", location, index)); found {
				return foundAt, true
			}
		}
	}
	return "", false
}
