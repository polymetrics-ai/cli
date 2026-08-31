package bitbucket

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

const bitbucketMissingFoundationPath = "missing-foundation.json"

type bitbucketMissingFoundationLedger struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	Purpose       string `json:"purpose"`
	SourceLock    struct {
		Path      string `json:"path"`
		SourceURL string `json:"source_url"`
		SHA256    string `json:"sha256"`
	} `json:"source_lock"`
	Foundations          []bitbucketMissingFoundation `json:"foundations"`
	MappingContractDebts []struct {
		ID                      string   `json:"id"`
		State                   string   `json:"state"`
		AffectedLanes           []string `json:"affected_lanes"`
		PrimarySourceLock       string   `json:"primary_source_lock"`
		RetainedMappingContract string   `json:"retained_mapping_contract"`
		SourceRows              int      `json:"source_rows"`
		CanonicalEvidence       bool     `json:"canonical_evidence"`
		RuntimeClaim            string   `json:"runtime_claim"`
	} `json:"mapping_contract_debts"`
}

type bitbucketMissingFoundation struct {
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

func TestBitbucketMissingFoundationLedgerBindsOnlyActualNonExecutingGaps(t *testing.T) {
	ledger := loadBitbucketJSON[bitbucketMissingFoundationLedger](t, bitbucketMissingFoundationPath)
	lock := loadBitbucketSourceLaneLock(t)
	matrix := loadBitbucketSourceLaneMatrix(t)

	if ledger.SchemaVersion != 1 || ledger.Connector != "bitbucket" || strings.TrimSpace(ledger.Purpose) == "" {
		t.Fatalf("Bitbucket missing-foundation identity = %#v", ledger)
	}
	if ledger.SourceLock.Path != bitbucketSourceLockPath ||
		ledger.SourceLock.SourceURL != lock.REST.SourceURL ||
		ledger.SourceLock.SHA256 != lock.REST.SHA256 {
		t.Fatalf("Bitbucket missing-foundation source lock = %#v, want current pinned lock", ledger.SourceLock)
	}
	if len(ledger.Foundations) != 1 {
		t.Fatalf("Bitbucket missing-foundation entries = %d, want one receiver gap", len(ledger.Foundations))
	}

	foundation := ledger.Foundations[0]
	if foundation.ID != "cli-webhook-event-surface-foundation-r1" ||
		foundation.State != "recorded_only_requires_captain_approval_before_implementation" ||
		foundation.AtlasCapability != "transport.sync-contract.v1" ||
		foundation.AtlasClassification != "actual_gap" ||
		foundation.AffectedLane != "sync_transport" {
		t.Fatalf("Bitbucket webhook foundation = %#v, want the exact recorded receiver gap", foundation)
	}

	wantWebhookIDs := []string{
		"bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/hooks",
		"bitbucket.rest.post_/workspaces/{workspace}/hooks",
		"bitbucket.rest.put_/repositories/{workspace}/{repo_slug}/hooks/{uid}",
		"bitbucket.rest.put_/workspaces/{workspace}/hooks/{uid}",
	}
	gotWebhookIDs := make([]string, 0, len(foundation.SourceIDs))
	for _, source := range foundation.SourceIDs {
		gotWebhookIDs = append(gotWebhookIDs, source.ID)
		operation := findBitbucketLockedOperation(t, lock, source.ID)
		if source.Method != operation.Method || source.Path != operation.Path || source.SourceLocation != operation.SourceLocation {
			t.Fatalf("Bitbucket webhook source %q = %#v, want exact locked identity %#v", source.ID, source, operation)
		}
		row := findBitbucketMatrixSourceRow(t, matrix, source.ID)
		cell := row.Lanes["sync_transport"]
		if cell.Applicability != "applicable" || cell.Disposition != "missing_foundation" {
			t.Fatalf("Bitbucket webhook matrix %q = %#v, want applicable missing-foundation", source.ID, cell)
		}
	}
	slices.Sort(gotWebhookIDs)
	slices.Sort(wantWebhookIDs)
	if !slices.Equal(gotWebhookIDs, wantWebhookIDs) {
		t.Fatalf("Bitbucket webhook source IDs = %v, want %v", gotWebhookIDs, wantWebhookIDs)
	}
	claim := strings.ToLower(foundation.RuntimeClaim)
	if !strings.Contains(claim, "no inbound bitbucket webhook receiver") ||
		!strings.Contains(claim, "no inbound bitbucket webhook receiver, selected source executor, or executable sync transport is claimed") {
		t.Fatalf("Bitbucket webhook runtime claim = %q, want explicit non-execution boundary", foundation.RuntimeClaim)
	}

	if len(ledger.MappingContractDebts) != 1 {
		t.Fatalf("Bitbucket mapping-contract debts = %d, want one descriptor debt", len(ledger.MappingContractDebts))
	}
	debt := ledger.MappingContractDebts[0]
	if debt.ID != "bitbucket-retained-source-descriptor-r1" ||
		debt.State != "recorded_mapping_contract_debt" ||
		!slices.Equal(debt.AffectedLanes, bitbucketSourceLaneNames) ||
		debt.PrimarySourceLock != bitbucketSourceLockPath ||
		debt.RetainedMappingContract != "sources/bitbucket-retained-mapping-contract.json" ||
		debt.SourceRows != lock.Counts.Total ||
		debt.CanonicalEvidence {
		t.Fatalf("Bitbucket descriptor debt = %#v, want exact descriptor-free retained-source boundary", debt)
	}
	debtClaim := strings.ToLower(debt.RuntimeClaim)
	if !strings.Contains(debtClaim, "no source-bound command") ||
		!strings.Contains(debtClaim, "no executor") ||
		!strings.Contains(debtClaim, "no executable declaration") {
		t.Fatalf("Bitbucket descriptor-debt runtime claim = %q, want explicit non-execution boundary", debt.RuntimeClaim)
	}
}

func findBitbucketMatrixSourceRow(t *testing.T, matrix bitbucketSourceLaneMatrix, sourceID string) bitbucketSourceLaneMatrixRow {
	t.Helper()
	for _, row := range matrix.SourceOperations {
		if row.SourceID == sourceID {
			return row
		}
	}
	t.Fatalf("Bitbucket source matrix row %q is absent", sourceID)
	return bitbucketSourceLaneMatrixRow{}
}

func TestBitbucketDescriptorDebtPreventsSourceBoundArtifactPromotion(t *testing.T) {
	ledger := loadBitbucketJSON[bitbucketMissingFoundationLedger](t, bitbucketMissingFoundationPath)
	if len(ledger.MappingContractDebts) != 1 || ledger.MappingContractDebts[0].CanonicalEvidence {
		t.Fatalf("Bitbucket descriptor debt = %#v, want descriptor-free retention boundary", ledger.MappingContractDebts)
	}

	for _, path := range []string{"operations.json", "streams.json", "writes.json", "cli_surface.json"} {
		artifact := loadBitbucketJSON[any](t, path)
		if location, found := bitbucketSourceOperationField(artifact, path); found {
			t.Fatalf("Bitbucket descriptor-free artifact %s unexpectedly promotes source_operation at %s", path, location)
		}
	}
}

func bitbucketSourceOperationField(value any, location string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			next := fmt.Sprintf("%s.%s", location, key)
			if key == "source_operation" {
				return next, true
			}
			if foundAt, found := bitbucketSourceOperationField(nested, next); found {
				return foundAt, true
			}
		}
	case []any:
		for index, nested := range typed {
			if foundAt, found := bitbucketSourceOperationField(nested, fmt.Sprintf("%s[%d]", location, index)); found {
				return foundAt, true
			}
		}
	}
	return "", false
}
