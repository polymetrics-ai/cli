package app_test

import (
	"context"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestLocalETLAndReverseETLWorkflow(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}

	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "sample-local",
		Connector: "sample",
		Secrets:   map[string]string{"token": "sample-token"},
		Config:    map[string]string{"workspace": "local"},
	}); err != nil {
		t.Fatalf("AddCredential(sample) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential(warehouse) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "outbox-local",
		Connector: "outbox",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "outbox")},
	}); err != nil {
		t.Fatalf("AddCredential(outbox) error = %v", err)
	}

	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name: "sample_to_warehouse",
		Source: app.EndpointConfig{
			Connector:  "sample",
			Credential: "sample-local",
		},
		Destination: app.EndpointConfig{
			Connector:  "warehouse",
			Credential: "warehouse-local",
		},
		Streams: map[string]app.StreamConfig{
			"customers": {
				SyncMode:         "full_refresh_overwrite",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: "sample_customers",
			},
		},
	}); err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}

	run, err := a.RunETL(ctx, app.RunETLRequest{Connection: "sample_to_warehouse", Stream: "customers"})
	if err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	if run.Status != "completed" || run.RecordsRead == 0 || run.RecordsLoaded != run.RecordsRead {
		t.Fatalf("unexpected ETL run: %+v", run)
	}

	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "sample_customers", Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("QueryTable() returned %d rows, want 3", len(rows))
	}

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "customers_to_outbox",
		SourceTable:           "sample_customers",
		DestinationConnector:  "outbox",
		DestinationCredential: "outbox-local",
		Action:                "upsert",
		Mappings: map[string]string{
			"id":    "external_id",
			"name":  "full_name",
			"email": "email",
		},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	if plan.RecordCount != 3 || plan.ApprovalToken == "" {
		t.Fatalf("unexpected reverse plan: %+v", plan)
	}

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() after reverse plan error = %v", err)
	}

	reverseRun, err := reopened.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err != nil {
		t.Fatalf("RunReverseETL() error = %v", err)
	}
	if reverseRun.Status != "completed" || reverseRun.RecordsSucceeded != 3 {
		t.Fatalf("unexpected reverse run: %+v", reverseRun)
	}
}
