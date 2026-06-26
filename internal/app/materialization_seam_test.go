package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

type aliasWarehouse struct{}

func (aliasWarehouse) Name() string { return "warehouse-alias" }
func (aliasWarehouse) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "warehouse-alias", DisplayName: "Warehouse Alias", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: true}}
}
func (aliasWarehouse) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return ctx.Err()
}
func (aliasWarehouse) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Warehouse{}.Catalog(ctx, cfg)
}
func (aliasWarehouse) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return connectors.Warehouse{}.Read(ctx, req, emit)
}
func (aliasWarehouse) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, errors.New("connector Write should not be used for materializing destinations")
}
func (aliasWarehouse) MaterializesLocalWarehouse() bool { return true }

func TestRunETLUsesMaterializationInterfaceInsteadOfWarehouseName(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	a.Registry().Register(aliasWarehouse{})
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "sample-local", Connector: "sample", Secrets: map[string]string{"token": "sample-token"}}); err != nil {
		t.Fatalf("AddCredential(sample) error = %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{Name: "warehouse-alias-local", Connector: "warehouse-alias", Config: map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")}}); err != nil {
		t.Fatalf("AddCredential(alias) error = %v", err)
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "sample_to_alias",
		Source:      app.EndpointConfig{Connector: "sample", Credential: "sample-local"},
		Destination: app.EndpointConfig{Connector: "warehouse-alias", Credential: "warehouse-alias-local"},
		Streams:     map[string]app.StreamConfig{"customers": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "sample_customers"}},
	}); err != nil {
		t.Fatalf("CreateConnection() error = %v", err)
	}
	run, err := a.RunETL(ctx, app.RunETLRequest{Connection: "sample_to_alias", Stream: "customers"})
	if err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	if run.RecordsLoaded != 3 {
		t.Fatalf("RecordsLoaded = %d, want 3", run.RecordsLoaded)
	}
}
