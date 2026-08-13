package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// #4081 must not treat a typed writes.json action as self-approval. The
// pre-run App plan is closed over the connection's configured source predicate
// and fixed destination record; every apply below receives only a workset that
// was actually staged and independently reopened from the connection warehouse.
func TestIssueLabelDestinationRejectsUnapprovedOrMismatchedOrExpiredOrReplayedPlanBeforeProviderWrite(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, issueLabelTransportApprovalFixture)
	}{
		{
			name: "missing pre-run approval",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
				if err := fixture.apply(t, receipt, workset, synctransport.DestinationApproval{}); err == nil {
					t.Fatal("ApplyDestination() accepted a missing pre-run approval")
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "same target and configuration with different reopened source issue",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				_, approval := fixture.preRunApproval(t)
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue+1)
				if err := fixture.apply(t, receipt, workset, approval); err == nil {
					t.Fatal("ApplyDestination() accepted an approval for a different reopened source issue")
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "expired pre-run approval",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				plan, approval := fixture.preRunApproval(t)
				fixture.expirePlanSeal(t, plan.ID)
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
				err := fixture.apply(t, receipt, workset, approval)
				if err == nil || !strings.Contains(err.Error(), "expired") {
					t.Fatalf("ApplyDestination() expired approval error = %v, want expired rejection", err)
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "replayed pre-run approval",
			run: func(t *testing.T, fixture issueLabelTransportApprovalFixture) {
				_, approval := fixture.preRunApproval(t)
				receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
				if err := fixture.apply(t, receipt, workset, approval); err != nil {
					t.Fatalf("first ApplyDestination() = %v", err)
				}
				fixture.assertProviderWrites(t, 1)
				if err := fixture.apply(t, receipt, workset, approval); err == nil {
					t.Fatal("ApplyDestination() accepted a replayed approval")
				}
				fixture.assertProviderWrites(t, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, newIssueLabelTransportApprovalFixture(t))
		})
	}
}

func TestIssueLabelTransportCleanupUsesItsOwnDerivedApproval(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixture(t)
	forwardPlan, forwardApproval := fixture.preRunApproval(t)
	receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
	if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
		t.Fatalf("forward ApplyDestination() = %v", err)
	}
	fixture.assertProviderWrites(t, 1)

	cleanupPlan, cleanupApproval := fixture.preRunCleanupApproval(t, forwardPlan.ID)
	if cleanupPlan.TransportForwardPlanID != forwardPlan.ID || cleanupPlan.Action != issueLabelRemoveAction {
		t.Fatalf("cleanup plan = %+v, want inverse derived from forward plan %q", cleanupPlan, forwardPlan.ID)
	}
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, forwardApproval); err == nil {
		t.Fatal("cleanup accepted the forward approval token")
	}
	fixture.assertProviderDeletes(t, 0)
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err != nil {
		t.Fatalf("ApplyIssueLabelTransportCleanup() = %v", err)
	}
	fixture.assertProviderDeletes(t, 1)
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err == nil {
		t.Fatal("cleanup accepted a replayed cleanup approval")
	}
	fixture.assertProviderDeletes(t, 1)
}

func TestIssueLabelTransportCleanupRejectsUnauthenticatedForwardPlanBeforeProviderDelete(t *testing.T) {
	for _, tt := range issueLabelTransportForwardPlanTamperCases() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			forwardPlan, forwardApproval := fixture.preRunApproval(t)
			receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
			if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
				t.Fatalf("forward ApplyDestination() = %v", err)
			}
			fixture.assertProviderWrites(t, 1)
			tt.mutate(t, fixture, forwardPlan.ID)
			if _, err := fixture.app.PlanIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, forwardPlan.ID); err == nil {
				t.Fatal("PlanIssueLabelTransportCleanup() accepted a tampered forward plan")
			}
			fixture.assertProviderDeletes(t, 0)
		})
	}
}

func TestIssueLabelTransportCleanupApplyRejectsTamperedForwardPlanBeforeProviderDelete(t *testing.T) {
	for _, tt := range issueLabelTransportForwardPlanTamperCases() {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newIssueLabelTransportApprovalFixture(t)
			forwardPlan, forwardApproval := fixture.preRunApproval(t)
			receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
			if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
				t.Fatalf("forward ApplyDestination() = %v", err)
			}
			_, cleanupApproval := fixture.preRunCleanupApproval(t, forwardPlan.ID)
			tt.mutate(t, fixture, forwardPlan.ID)
			if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err == nil {
				t.Fatal("ApplyIssueLabelTransportCleanup() accepted a tampered forward plan")
			}
			fixture.assertProviderDeletes(t, 0)
		})
	}
}

func TestIssueLabelTransportCleanupTreatsMissingLabelAsSuccessfulInverse(t *testing.T) {
	fixture := newIssueLabelTransportApprovalFixtureWithMissingCleanupLabel(t)
	forwardPlan, forwardApproval := fixture.preRunApproval(t)
	receipt, workset := fixture.stageAndReopen(t, fixture.sourceIssue)
	if err := fixture.apply(t, receipt, workset, forwardApproval); err != nil {
		t.Fatalf("forward ApplyDestination() = %v", err)
	}
	cleanupPlan, cleanupApproval := fixture.preRunCleanupApproval(t, forwardPlan.ID)
	if _, err := fixture.app.ApplyIssueLabelTransportCleanup(context.Background(), fixture.connection.ID, cleanupApproval); err != nil {
		t.Fatalf("ApplyIssueLabelTransportCleanup() = %v, want missing-label success for %q", err, cleanupPlan.ID)
	}
	fixture.assertProviderDeletes(t, 1)
}

type issueLabelTransportApprovalFixture struct {
	app         *App
	connection  Connection
	executor    *issueLabelDestinationExecutor
	runtime     connectors.RuntimeConfig
	sourceIssue int
	targetIssue int
	label       string
	writes      *int
	deletes     *int
}

func newIssueLabelTransportApprovalFixture(t *testing.T) issueLabelTransportApprovalFixture {
	return newIssueLabelTransportApprovalFixtureWithCleanupStatus(t, http.StatusNoContent)
}

func newIssueLabelTransportApprovalFixtureWithMissingCleanupLabel(t *testing.T) issueLabelTransportApprovalFixture {
	return newIssueLabelTransportApprovalFixtureWithCleanupStatus(t, http.StatusNotFound)
}

func newIssueLabelTransportApprovalFixtureWithCleanupStatus(t *testing.T, cleanupStatus int) issueLabelTransportApprovalFixture {
	t.Helper()
	ctx := context.Background()
	writes := 0
	deletes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			if request.URL.Path != "/repos/acme/widgets/issues/200/labels" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			writes++
			_ = json.NewEncoder(w).Encode([]map[string]any{{"name": "transport-demo"}})
		case http.MethodDelete:
			if request.URL.Path != "/repos/acme/widgets/issues/200/labels/transport-demo" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			deletes++
			w.WriteHeader(cleanupStatus)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "github-transport-local",
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
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "other-github-credential",
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
	connection, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "github_transport_approval",
		Source: EndpointConfig{Connector: "github", Credential: "github-transport-local", Config: map[string]string{
			issueLabelTransportSourceIssueConfig: "100",
		}},
		Destination: EndpointConfig{Connector: "github", Credential: "github-transport-local", Config: map[string]string{
			issueLabelTransportTargetIssueConfig: "200",
			issueLabelTransportLabelConfig:       "transport-demo",
		}},
		Streams: map[string]StreamConfig{
			"issues": {SyncMode: string(synccontract.ModeFullAppend), DestinationTable: "issues"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, runtime, err := a.resolveEndpoint(ctx, connection.Destination)
	if err != nil {
		t.Fatal(err)
	}
	registered, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	github, ok := registered.(*issueLabelTransportConnector)
	if !ok || github.Connector == nil {
		t.Fatalf("GitHub transport connector = %T, want concrete engine wrapper", registered)
	}
	return issueLabelTransportApprovalFixture{
		app:         a,
		connection:  connection,
		executor:    &issueLabelDestinationExecutor{app: a, connector: github.Connector},
		runtime:     runtime,
		sourceIssue: 100,
		targetIssue: 200,
		label:       "transport-demo",
		writes:      &writes,
		deletes:     &deletes,
	}
}

func (f issueLabelTransportApprovalFixture) preRunApproval(t *testing.T) (ReversePlan, synctransport.DestinationApproval) {
	t.Helper()
	plan, err := f.app.PlanIssueLabelTransport(context.Background(), f.connection.ID)
	if err != nil {
		t.Fatalf("PlanIssueLabelTransport() = %v", err)
	}
	plan, preview, err := f.app.PreviewIssueLabelTransport(context.Background(), plan.ID)
	if err != nil {
		t.Fatalf("PreviewIssueLabelTransport() = %v", err)
	}
	if preview.Digest == "" || plan.ApprovalToken == "" || plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("pre-run GitHub transport approval = plan=%+v preview=%+v, want persisted destructive grant", plan, preview)
	}
	return plan, synctransport.DestinationApproval{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}

func (f issueLabelTransportApprovalFixture) preRunCleanupApproval(t *testing.T, forwardPlanID string) (ReversePlan, synctransport.DestinationApproval) {
	t.Helper()
	plan, err := f.app.PlanIssueLabelTransportCleanup(context.Background(), f.connection.ID, forwardPlanID)
	if err != nil {
		t.Fatalf("PlanIssueLabelTransportCleanup() = %v", err)
	}
	plan, preview, err := f.app.PreviewIssueLabelTransport(context.Background(), plan.ID)
	if err != nil {
		t.Fatalf("PreviewIssueLabelTransport(cleanup) = %v", err)
	}
	if preview.Digest == "" || plan.ApprovalToken == "" || plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive {
		t.Fatalf("pre-run GitHub transport cleanup approval = plan=%+v preview=%+v, want persisted destructive grant", plan, preview)
	}
	return plan, synctransport.DestinationApproval{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}

func (f issueLabelTransportApprovalFixture) stageAndReopen(t *testing.T, sourceIssue int) (synctransport.WarehouseReceipt, synctransport.WarehouseWorkset) {
	t.Helper()
	page := synctransport.SourcePage{Records: []connectors.Record{{
		"id":     "source-issue",
		"number": sourceIssue,
		"title":  "warehouse-owned transport source",
	}}}
	receipt, err := f.app.transportStage.Stage(context.Background(), synctransport.WarehouseStageRequest{
		ConnectionID:    f.connection.ID,
		Generation:      1,
		SourceName:      "github",
		DestinationName: "github",
		Stream:          "issues",
		Mode:            synccontract.ModeFullAppend,
		Page:            page,
	})
	if err != nil {
		t.Fatalf("Stage() = %v", err)
	}
	// Destroy the source-owned alias before Reopen; the apply receives only the
	// persisted WAL/DuckDB/Parquet representation and immutable receipt.
	page.Records[0]["number"] = -1
	page.Records = nil
	page = synctransport.SourcePage{}
	workset, err := f.app.transportStage.Reopen(context.Background(), receipt)
	if err != nil {
		t.Fatalf("Reopen() = %v", err)
	}
	return receipt, workset
}

func (f issueLabelTransportApprovalFixture) expirePlanSeal(t *testing.T, planID string) {
	t.Helper()
	found := false
	for i := range f.app.state.ReversePlans {
		plan := &f.app.state.ReversePlans[i]
		if plan.ID != planID {
			continue
		}
		if plan.PlanSeal == nil {
			t.Fatalf("plan %q has no seal", planID)
		}
		seal := *plan.PlanSeal
		seal.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		var err error
		seal.MAC, err = f.app.approval.planSealMAC(seal)
		if err != nil {
			t.Fatal(err)
		}
		plan.PlanSeal = &seal
		plan.ExpiresAt = seal.ExpiresAt
		found = true
		break
	}
	if !found {
		t.Fatalf("plan %q was not stored", planID)
	}
	if err := f.app.save(); err != nil {
		t.Fatal(err)
	}
}

func (f issueLabelTransportApprovalFixture) mutateForwardPlan(t *testing.T, planID string, mutate func(*ReversePlan)) {
	t.Helper()
	found := false
	for i := range f.app.state.ReversePlans {
		if f.app.state.ReversePlans[i].ID != planID {
			continue
		}
		mutate(&f.app.state.ReversePlans[i])
		found = true
		break
	}
	if !found {
		t.Fatalf("forward plan %q was not stored", planID)
	}
	if err := f.app.save(); err != nil {
		t.Fatal(err)
	}
}

func (f issueLabelTransportApprovalFixture) apply(t *testing.T, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset, approval synctransport.DestinationApproval) error {
	t.Helper()
	acknowledgement, err := f.executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
		ConnectionID: f.connection.ID,
		Plan: synctransport.DestinationPlan{ApplyStrategy: connectors.DestinationApplyStrategy{
			Mode:     synccontract.ModeFullAppend,
			Strategy: connectors.ApplyStrategyAppend,
			Action:   issueLabelAddAction,
		}},
		Receipt:  receipt,
		Workset:  workset,
		Runtime:  f.runtime,
		Approval: approval,
	})
	if err == nil && acknowledgement.Sink != "github" {
		t.Fatalf("ApplyDestination() acknowledgement sink = %q, want github", acknowledgement.Sink)
	}
	return err
}

func (f issueLabelTransportApprovalFixture) assertProviderWrites(t *testing.T, want int) {
	t.Helper()
	if got := *f.writes; got != want {
		t.Fatalf("GitHub label POST calls = %d, want %d", got, want)
	}
}

func (f issueLabelTransportApprovalFixture) assertProviderDeletes(t *testing.T, want int) {
	t.Helper()
	if got := *f.deletes; got != want {
		t.Fatalf("GitHub label DELETE calls = %d, want %d", got, want)
	}
}

type issueLabelTransportForwardPlanTamperCase struct {
	name   string
	mutate func(*testing.T, issueLabelTransportApprovalFixture, string)
}

func issueLabelTransportForwardPlanTamperCases() []issueLabelTransportForwardPlanTamperCase {
	return []issueLabelTransportForwardPlanTamperCase{
		{
			name: "destination connector",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.DestinationConnector = "warehouse" })
			},
		},
		{
			name: "destination credential",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.DestinationCredential = "other-github-credential" })
			},
		},
		{
			name: "destination config",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) {
					plan.DestinationConfig = cloneStringMap(plan.DestinationConfig)
					plan.DestinationConfig[issueLabelTransportTargetIssueConfig] = "201"
				})
			},
		},
		{
			name: "binding",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.TransportBindingSHA256 = "tampered-binding" })
			},
		},
		{
			name: "plan hash",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) { plan.PlanHash = "tampered-plan-hash" })
			},
		},
		{
			name: "plan seal",
			mutate: func(t *testing.T, fixture issueLabelTransportApprovalFixture, planID string) {
				fixture.mutateForwardPlan(t, planID, func(plan *ReversePlan) {
					if plan.PlanSeal == nil {
						t.Fatal("forward plan has no seal")
					}
					plan.PlanSeal.MAC = "tampered-seal"
				})
			},
		},
	}
}
