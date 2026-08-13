package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

// #4081 must not treat a typed writes.json action as its own approval. These
// cases pin the normal PM plan -> preview -> approval -> execute route: no
// absent, target-mismatched, or replayed approval may dispatch the GitHub POST.
func TestGitHubIssueLabelDestinationRejectsUnapprovedOrMismatchedOrReplayedPlanBeforeProviderWrite(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, githubTransportApprovalFixture)
	}{
		{
			name: "missing approval",
			run: func(t *testing.T, fixture githubTransportApprovalFixture) {
				err := fixture.apply(t, synctransport.DestinationApproval{})
				if err == nil {
					t.Fatal("ApplyDestination() accepted a missing transport approval")
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "mismatched approved target",
			run: func(t *testing.T, fixture githubTransportApprovalFixture) {
				plan := fixture.previewPlan(t, fixture.targetIssue+1)
				err := fixture.apply(t, synctransport.DestinationApproval{
					PlanID:        plan.ID,
					ApprovalToken: plan.ApprovalToken,
					Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
				})
				if err == nil {
					t.Fatal("ApplyDestination() accepted a plan for a different GitHub issue")
				}
				fixture.assertProviderWrites(t, 0)
			},
		},
		{
			name: "replayed approval",
			run: func(t *testing.T, fixture githubTransportApprovalFixture) {
				plan := fixture.previewPlan(t, fixture.targetIssue)
				approval := synctransport.DestinationApproval{
					PlanID:        plan.ID,
					ApprovalToken: plan.ApprovalToken,
					Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
				}
				if err := fixture.apply(t, approval); err != nil {
					t.Fatalf("first ApplyDestination() = %v", err)
				}
				fixture.assertProviderWrites(t, 1)
				if err := fixture.apply(t, approval); err == nil {
					t.Fatal("ApplyDestination() accepted a replayed transport approval")
				}
				fixture.assertProviderWrites(t, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, newGitHubTransportApprovalFixture(t))
		})
	}
}

type githubTransportApprovalFixture struct {
	app         *App
	connection  Connection
	executor    *githubIssueLabelDestinationExecutor
	runtime     connectors.RuntimeConfig
	targetIssue int
	label       string
	writes      *int
}

func newGitHubTransportApprovalFixture(t *testing.T) githubTransportApprovalFixture {
	t.Helper()
	ctx := context.Background()
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			writes++
			if want := "/repos/acme/widgets/issues/200/labels"; request.URL.Path != want {
				t.Fatalf("GitHub write path = %q, want %q", request.URL.Path, want)
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode GitHub label body: %v", err)
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{{"name": "transport-demo"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
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
	connection, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name: "github_transport_approval",
		Source: EndpointConfig{Connector: "github", Credential: "github-transport-local", Config: map[string]string{
			githubTransportSourceIssueConfig: "100",
		}},
		Destination: EndpointConfig{Connector: "github", Credential: "github-transport-local", Config: map[string]string{
			githubTransportTargetIssueConfig: "200",
			githubTransportLabelConfig:       "transport-demo",
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
	github, ok := registered.(*githubTransportConnector)
	if !ok || github.Connector == nil {
		t.Fatalf("GitHub transport connector = %T, want concrete engine wrapper", registered)
	}
	return githubTransportApprovalFixture{
		app:         a,
		connection:  connection,
		executor:    &githubIssueLabelDestinationExecutor{app: a, connector: github.Connector},
		runtime:     runtime,
		targetIssue: 200,
		label:       "transport-demo",
		writes:      &writes,
	}
}

func (f githubTransportApprovalFixture) previewPlan(t *testing.T, issueNumber int) ReversePlan {
	t.Helper()
	ctx := context.Background()
	table := "github_transport_approval"
	location, err := f.app.warehouseLocation(f.app.warehouseRoot(), f.connection)
	if err != nil {
		t.Fatal(err)
	}
	if err := location.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	path, err := location.TablePath(table)
	if err != nil {
		t.Fatal(err)
	}
	if err := location.AssertOwnedTable(path, table); err != nil {
		t.Fatal(err)
	}
	if err := warehouse.WriteTable(ctx, path, []warehouse.Row{{
		"issue_number": issueNumber,
		"labels":       []string{f.label},
	}}); err != nil {
		t.Fatal(err)
	}
	plan, err := f.app.PlanReverseETL(ctx, PlanReverseETLRequest{
		Name:                  "github_transport_label_approval",
		SourceTable:           table,
		SourceConnection:      f.connection.Name,
		DestinationConnector:  "github",
		DestinationCredential: f.connection.Destination.Credential,
		DestinationConfig:     f.connection.Destination.Config,
		Action:                githubIssueAddLabelAction,
		Mappings:              map[string]string{"issue_number": "issue_number", "labels": "labels"},
		Limit:                 1,
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() = %v", err)
	}
	plan, _, err = f.app.PreviewReversePlan(ctx, plan.ID, nil)
	if err != nil {
		t.Fatalf("PreviewReversePlan() = %v", err)
	}
	return plan
}

func (f githubTransportApprovalFixture) apply(t *testing.T, approval synctransport.DestinationApproval) error {
	t.Helper()
	acknowledgement, err := f.executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
		ConnectionID: f.connection.ID,
		Plan: synctransport.DestinationPlan{ApplyStrategy: connectors.DestinationApplyStrategy{
			Mode:     synccontract.ModeFullAppend,
			Strategy: connectors.ApplyStrategyAppend,
			Action:   githubIssueAddLabelAction,
		}},
		Workset: synctransport.WarehouseWorkset{
			ID:      "reopened-transport-workset",
			Records: []connectors.Record{{"id": "source-issue-100"}},
		},
		Runtime:  f.runtime,
		Approval: approval,
	})
	if err == nil {
		if acknowledgement.Sink != "github" {
			t.Fatalf("ApplyDestination() acknowledgement sink = %q, want github", acknowledgement.Sink)
		}
	}
	return err
}

func (f githubTransportApprovalFixture) assertProviderWrites(t *testing.T, want int) {
	t.Helper()
	if got := *f.writes; got != want {
		t.Fatalf("GitHub label POST calls = %d, want %d", got, want)
	}
}
