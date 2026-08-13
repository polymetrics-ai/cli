package app

import (
	"context"
	"runtime"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// These tests deliberately name the durable-stage contract that #4081 needs.
// They are committed RED before its production composition root and adapter.
func TestOpenInstallsGitHubWarehouseMediatedTransport(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	if a.transportStage == nil {
		t.Fatal("Open() left the GitHub warehouse-mediated transport stage nil")
	}
	if _, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: github,
		Stream:      "issues",
		Mode:        synccontract.ModeFullAppend,
	}); err != nil {
		t.Fatalf("GitHub transport preflight = %v, want explicit accepted construction", err)
	}
}

func TestGitHubWarehouseStageReopensDurableReceiptAfterSourceReferencesAreDiscarded(t *testing.T) {
	ctx := context.Background()
	fixture := newGitHubWarehouseStageFixture(t)
	page := synctransport.SourcePage{Records: []connectors.Record{{
		"id":    "issue-4081-source",
		"title": "durable GitHub issue fixture",
	}}}
	receipt := stageGitHubWarehousePage(t, ctx, fixture, page)

	// No source-owned page or record remains available to the reopen call.
	page.Records[0]["title"] = "mutated source alias"
	page.Records = nil
	page = synctransport.SourcePage{}
	runtime.GC()

	first, err := fixture.app.transportStage.Reopen(ctx, receipt)
	if err != nil {
		t.Fatalf("Reopen() = %v", err)
	}
	if got := first.Records; len(got) != 1 || got[0]["title"] != "durable GitHub issue fixture" {
		t.Fatalf("first reopened records = %#v, want the durable source value", got)
	}
	first.Records[0]["title"] = "caller-mutated reopen alias"

	second, err := fixture.app.transportStage.Reopen(ctx, receipt)
	if err != nil {
		t.Fatalf("second Reopen() = %v", err)
	}
	if got := second.Records; len(got) != 1 || got[0]["title"] != "durable GitHub issue fixture" {
		t.Fatalf("second reopened records = %#v, want immutable durable source value", got)
	}
}

func TestGitHubWarehouseStageRejectsTamperedReceipt(t *testing.T) {
	ctx := context.Background()
	fixture := newGitHubWarehouseStageFixture(t)
	receipt := stageGitHubWarehousePage(t, ctx, fixture, synctransport.SourcePage{Records: []connectors.Record{{
		"id": "issue-4081-tamper",
	}}})

	for _, tt := range []struct {
		name   string
		mutate func(synctransport.WarehouseReceipt) synctransport.WarehouseReceipt
	}{
		{
			name: "owner",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.Owner = "another-connection-owner"
				return receipt
			},
		},
		{
			name: "generation",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.Generation++
				return receipt
			},
		},
		{
			name: "manifest hash",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.ManifestSHA256 = "00"
				return receipt
			},
		},
		{
			name: "content hash",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.ContentSHA256 = "00"
				return receipt
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			workset, err := fixture.app.transportStage.Reopen(ctx, tt.mutate(receipt))
			if err == nil {
				t.Fatalf("Reopen() workset = %#v, want tampered receipt rejection", workset)
			}
			if len(workset.Records) != 0 {
				t.Fatalf("Reopen() leaked records after %s tampering: %#v", tt.name, workset.Records)
			}
		})
	}
}

type githubWarehouseStageFixture struct {
	app          *App
	connectionID string
}

func newGitHubWarehouseStageFixture(t *testing.T) githubWarehouseStageFixture {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "github-source", Connector: "github"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "github-destination", Connector: "github"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "github-issues-label-demo",
		Source:      EndpointConfig{Connector: "github", Credential: "github-source"},
		Destination: EndpointConfig{Connector: "github", Credential: "github-destination"},
		Streams: map[string]StreamConfig{
			"issues": {SyncMode: string(synccontract.ModeFullAppend), DestinationTable: "issues"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(a.state.Connections) != 1 || a.state.Connections[0].ID == "" {
		t.Fatalf("created GitHub connection = %#v, want one connection with an opaque ID", a.state.Connections)
	}
	return githubWarehouseStageFixture{app: a, connectionID: a.state.Connections[0].ID}
}

func stageGitHubWarehousePage(t *testing.T, ctx context.Context, fixture githubWarehouseStageFixture, page synctransport.SourcePage) synctransport.WarehouseReceipt {
	t.Helper()
	receipt, err := fixture.app.transportStage.Stage(ctx, synctransport.WarehouseStageRequest{
		ConnectionID:    fixture.connectionID,
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
	return receipt
}
