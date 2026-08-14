package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/warehouse"
)

func TestRunReverseETLRecoversCommittedApprovalConsumptionUnlockFailure(t *testing.T) {
	testRunReverseETLRecoversUncertainApprovalConsumption(t, func(a *App) {
		// RunReverseETL first reads the plan, then UpdateAfterPreflight reads it
		// again before acquiring the lock that commits the consumed approval.
		a.store.Locker = &postCommitUnlockFailureLocker{failAt: 1}
	}, statestore.CommitOutcomeCommitted)
}

func TestRunReverseETLRecoversIndeterminateApprovalConsumptionDirectorySyncFailure(t *testing.T) {
	testRunReverseETLRecoversUncertainApprovalConsumption(t, func(a *App) {
		a.store.SyncDirectory = func(string) error {
			return errors.New("directory sync failed")
		}
	}, statestore.CommitOutcomeIndeterminate)
}

func testRunReverseETLRecoversUncertainApprovalConsumption(t *testing.T, configure func(*App), wantOutcome statestore.CommitOutcome) {
	t.Helper()
	ctx := context.Background()
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writes++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "github",
		Connector: "github",
		Config: map[string]string{
			"owner":         "acme",
			"repo":          "widgets",
			"public_access": "true",
			"base_url":      server.URL,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := warehouse.WriteTable(context.Background(), filepath.Join(a.projectDir, "warehouse", "repo_deletes"+warehouse.TableFileExt), []warehouse.Row{{"id": "row-1"}}); err != nil {
		t.Fatal(err)
	}

	planRequest := PlanReverseETLRequest{
		Name:                  "delete_repo",
		SourceTable:           "repo_deletes",
		DestinationConnector:  "github",
		DestinationCredential: "github",
		Action:                "repo",
		Mappings:              map[string]string{"id": "id"},
	}
	plan, err := a.PlanReverseETL(ctx, planRequest)
	if err != nil {
		t.Fatal(err)
	}
	plan, _, err = a.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	writes = 0
	request := RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
	configure(a)

	_, err = a.RunReverseETL(ctx, request)
	var recovery *ApprovalConsumptionUncertainError
	if !errors.As(err, &recovery) {
		t.Fatalf("RunReverseETL() error = %T %v, want ApprovalConsumptionUncertainError", err, err)
	}
	if recovery.PlanID != plan.ID || recovery.ConsumedAt.IsZero() {
		t.Fatalf("recovery = %#v, want plan %q and consumption timestamp", recovery, plan.ID)
	}
	var outcome *statestore.CommitOutcomeError
	if !errors.As(err, &outcome) || outcome.Outcome != wantOutcome {
		t.Fatalf("RunReverseETL() state outcome = %#v, want %s", outcome, wantOutcome)
	}
	if writes != 0 {
		t.Fatalf("authorized write calls = %d, want 0 after uncertain consumption", writes)
	}

	stored, err := a.GetReversePlan(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != reversePlanStatusApprovalConsumptionUncertain || stored.ApprovalTokenHash != "" || stored.ApprovalGrant != nil || stored.ApprovalConsumedAt.IsZero() || stored.ApprovalUncertainAt.IsZero() {
		t.Fatalf("recovered plan = %#v, want persisted uncertain consumed approval state", stored)
	}
	if stored.PreviewDigest != plan.PreviewDigest || stored.PreviewedAt.IsZero() {
		t.Fatalf("recovered plan lost exact preview evidence: %#v", stored)
	}

	reopened, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reopened.GetReversePlan(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Status != reversePlanStatusApprovalConsumptionUncertain || persisted.ApprovalConsumedAt.IsZero() || persisted.ApprovalUncertainAt.IsZero() || persisted.PreviewDigest != plan.PreviewDigest {
		t.Fatalf("persisted recovery state = %#v, want uncertain approval audit evidence", persisted)
	}

	if _, err := a.RunReverseETL(ctx, request); !errors.As(err, &recovery) {
		t.Fatalf("RunReverseETL() retry error = %T %v, want ApprovalConsumptionUncertainError", err, err)
	}
	if _, _, err := a.PreviewReversePlan(ctx, plan.ID, nil); !errors.As(err, &recovery) {
		t.Fatalf("PreviewReversePlan() error = %T %v, want ApprovalConsumptionUncertainError", err, err)
	}
	if writes != 0 {
		t.Fatalf("write calls after rejected retry = %d, want 0", writes)
	}

	a.store.SyncDirectory = nil
	freshPlan, err := a.PlanReverseETL(ctx, planRequest)
	if err != nil {
		t.Fatal(err)
	}
	freshPlan, _, err = a.PreviewReversePlan(ctx, freshPlan.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	writes = 0
	if _, err := a.RunReverseETL(ctx, RunReverseETLRequest{
		PlanID:        freshPlan.ID,
		ApprovalToken: freshPlan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}); err != nil {
		t.Fatalf("RunReverseETL() with fresh preview and approval error = %v", err)
	}
	if writes != 1 {
		t.Fatalf("freshly approved write calls = %d, want 1", writes)
	}
}
