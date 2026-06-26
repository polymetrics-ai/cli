package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestRunReverseETLRejectsApprovalTokenReplay(t *testing.T) {
	ctx := context.Background()
	a, plan := setupApprovedReversePlan(t, ctx)

	if _, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken}); err != nil {
		t.Fatalf("first RunReverseETL() error = %v", err)
	}

	_, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err == nil {
		t.Fatal("second RunReverseETL() succeeded with replayed approval token")
	}
	if !strings.Contains(err.Error(), "already") {
		t.Fatalf("replay error = %v, want already-executed rejection", err)
	}
}

func TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange(t *testing.T) {
	ctx := context.Background()
	a, plan := setupApprovedReversePlan(t, ctx)

	rows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "sample_customers", Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected sample rows")
	}
	rows[0]["email"] = "changed@example.test"
	if err := writeWarehouseRows(t, a, "sample_customers", rows); err != nil {
		t.Fatalf("mutate warehouse rows: %v", err)
	}

	_, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken})
	if err == nil {
		t.Fatal("RunReverseETL() succeeded after source rows changed")
	}
	if !strings.Contains(err.Error(), "changed since approval") {
		t.Fatalf("mutation error = %v, want changed-since-approval rejection", err)
	}
}

func setupApprovedReversePlan(t *testing.T, ctx context.Context) (*app.App, app.ReversePlan) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "sample-local", Connector: "sample", Secrets: map[string]string{"token": "sample-token"}}); err != nil {
		t.Fatalf("AddCredential(sample) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "warehouse-local", Connector: "warehouse", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")}}); err != nil {
		t.Fatalf("AddCredential(warehouse) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "outbox-local", Connector: "outbox", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "outbox")}}); err != nil {
		t.Fatalf("AddCredential(outbox) error = %v", err)
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "sample_to_warehouse",
		Source:      app.EndpointConfig{Connector: "sample", Credential: "sample-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{
			"customers": {SyncMode: "full_refresh_overwrite", CursorField: "updated_at", PrimaryKey: []string{"id"}, DestinationTable: "sample_customers"},
		},
	}); err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}
	if _, err := a.RunETL(ctx, app.RunETLRequest{Connection: "sample_to_warehouse", Stream: "customers"}); err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "customers_to_outbox",
		SourceTable:           "sample_customers",
		DestinationConnector:  "outbox",
		DestinationCredential: "outbox-local",
		Action:                "upsert",
		Mappings:              map[string]string{"id": "external_id", "name": "full_name", "email": "email"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL() error = %v", err)
	}
	return a, plan
}

func writeWarehouseRows(t *testing.T, a *app.App, table string, rows []connectors.Record) error {
	t.Helper()
	path := filepath.Join(a.ProjectDir(), "warehouse", table+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, row := range rows {
		if err := encoder.Encode(row); err != nil {
			return err
		}
	}
	return nil
}
