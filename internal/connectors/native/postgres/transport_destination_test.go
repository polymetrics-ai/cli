package postgres

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestManagedTargetTransportDestinationRefusesBeforeSideEffects(t *testing.T) {
	root := t.TempDir()
	destination := &ManagedTargetTransportDestination{connector: New()}
	checkpoint := synccontract.CheckpointEnvelope{ObservedAt: time.Now().UTC()}
	request := synctransport.DestinationApplyRequest{
		Runtime: connectors.RuntimeConfig{ProjectDir: root},
		Workset: synctransport.WarehouseWorkset{
			ID: "workset-refusal", Records: []connectors.Record{{"id": int64(1)}}, CandidateCheckpoint: checkpoint,
		},
	}

	if _, err := destination.ApplyDestination(context.Background(), request); !errors.Is(err, ErrManagedTargetTransportApprovalRequired) {
		t.Fatalf("missing approval error = %v, want ErrManagedTargetTransportApprovalRequired", err)
	}
	assertManagedTargetTransportDirectoryEmpty(t, root)

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := destination.ApplyDestination(cancelled, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled apply error = %v, want context.Canceled", err)
	}
	assertManagedTargetTransportDirectoryEmpty(t, root)

	if err := destination.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{}); !errors.Is(err, ErrManagedTargetTransportReadBackFailed) {
		t.Fatalf("invalid read-back error = %v, want ErrManagedTargetTransportReadBackFailed", err)
	}
	assertManagedTargetTransportDirectoryEmpty(t, root)

	expiredApproval, expiredAt := managedTargetFixtureApproval(t)
	destination.now = func() time.Time { return expiredAt }
	request.Approval = expiredApproval
	if _, err := destination.ApplyDestination(context.Background(), request); !errors.Is(err, ErrManagedTargetTransportApprovalRequired) || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expired approval error = %v, want typed expiry refusal", err)
	}
	assertManagedTargetTransportDirectoryEmpty(t, root)
}

func TestManagedTargetHistoryRouteUsesTypedPostgresDefinitions(t *testing.T) {
	destinationConnector := New()
	destination := &ManagedTargetTransportDestination{connector: destinationConnector}

	route, err := destination.historyRoute(synccontract.ModeIncrementalDedupeHistory, New())
	if err != nil {
		t.Fatalf("historyRoute() error = %v", err)
	}
	want := destinationConnector.databaseDefinition.Driver()
	if route.Source != want || route.Destination != want {
		t.Fatalf("historyRoute() = %#v, want source and destination %v", route, want)
	}
}

func TestManagedTargetHistoryRouteRefusesSourceWithoutTypedDefinition(t *testing.T) {
	destination := &ManagedTargetTransportDestination{connector: New()}

	_, err := destination.historyRoute(synccontract.ModeIncrementalDedupeHistory, nil)
	var routeErr *database.DatabaseWriteHistoryRouteError
	if !errors.As(err, &routeErr) || routeErr.Reason != database.DatabaseWriteHistoryRouteSourceUnsupported {
		t.Fatalf("historyRoute(nil) error = %T %v, want typed source-route refusal", err, err)
	}
}

func TestManagedTargetHistoryWriteInputUsesSourceOwnedCandidatePosition(t *testing.T) {
	position := synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("20"), TieBreaker: synccontract.OpaqueToken("1")}
	input, err := managedTargetDatabaseWriteInput(synctransport.DestinationApplyRequest{
		Mode: synccontract.ModeIncrementalDedupeHistory,
		Workset: synctransport.WarehouseWorkset{
			Records:             []connectors.Record{{"id": int64(1)}, {"id": int64(2)}},
			CandidateCheckpoint: synccontract.CheckpointEnvelope{Position: position},
		},
	}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatalf("managedTargetDatabaseWriteInput() error = %v", err)
	}
	for index, record := range input.OrderedRecords() {
		if !reflect.DeepEqual(record.Position, position) {
			t.Fatalf("ordered record %d position = %#v, want source candidate %#v", index, record.Position, position)
		}
	}
}

func TestManagedTargetHistoryWriteUsesDeclaredPrimaryKey(t *testing.T) {
	keys := []string{"tenant", "id"}
	if got := managedTargetWriteKeys(synccontract.ModeIncrementalDedupeHistory, keys); !reflect.DeepEqual(got, keys) {
		t.Fatalf("managedTargetWriteKeys(history) = %#v, want %#v", got, keys)
	}
}

func TestManagedTargetBaselineWindowIsStableAcrossPageReplay(t *testing.T) {
	first := synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": int64(2), "value": "old"}, {"id": int64(1), "value": "old"}}}
	replayed := synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": int64(1), "value": "new"}, {"id": int64(2), "value": "new"}}}
	firstRoot, err := managedTargetBaselineWindowRoot("baseline", []string{"id"}, first)
	if err != nil {
		t.Fatal(err)
	}
	replayRoot, err := managedTargetBaselineWindowRoot("baseline", []string{"id"}, replayed)
	if err != nil {
		t.Fatal(err)
	}
	if firstRoot != replayRoot {
		t.Fatalf("same page key window changed across replay: first=%q replay=%q", firstRoot, replayRoot)
	}
	differentRoot, err := managedTargetBaselineWindowRoot("baseline", []string{"id"}, synctransport.WarehouseWorkset{Records: []connectors.Record{{"id": int64(3)}}})
	if err != nil {
		t.Fatal(err)
	}
	if differentRoot == firstRoot {
		t.Fatal("different page key window reused the prior baseline")
	}
}

func assertManagedTargetTransportDirectoryEmpty(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read refusal root: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("refusal created side effects: %v", entries)
	}
}

func managedTargetFixtureApproval(t *testing.T) (synctransport.DestinationApproval, time.Time) {
	t.Helper()
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatal(err)
	}
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	target := connectors.WriteApprovalTarget{
		Connector: "postgres", Operation: "managed_incremental_upsert", Method: "POSTGRESQL", MutationClass: "incremental_upsert",
		TargetDigest: strings.Repeat("a", 64), CredentialRevision: strings.Repeat("b", 64),
		ConfigurationDigest: strings.Repeat("c", 64), Batchable: true, Scope: connectors.WriteApprovalScopeFixture, Confirmation: confirmation,
	}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: "rplan_expired_transport", PlanHash: strings.Repeat("d", 64), PreviewDigest: strings.Repeat("e", 64),
		ApprovalToken: "fixture-token", Target: target, Confirmation: confirmation,
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, Mode: grant.Mode, PreviewDigest: grant.PreviewDigest,
		ApprovalToken: "fixture-token", Target: target, Confirmation: confirmation,
	})
	if err != nil {
		t.Fatal(err)
	}
	return synctransport.DestinationApproval{
		PlanID: grant.PlanID, Confirmation: confirmation, Evidence: evidence, Target: target, PreviewDigest: grant.PreviewDigest,
	}, grant.ExpiresAt
}
