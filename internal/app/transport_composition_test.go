package app

import (
	"context"
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

func TestOpenRegistersDefinitionOwnedProductionTransports(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	postgres, ok := a.registry.Get("postgres")
	if !ok {
		t.Fatal("PostgreSQL connector is not registered")
	}
	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	warehouse, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("local warehouse connector is not registered")
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: github,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-GitHub preflight = %v", err)
	}
	if got, want := resolved.Source.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_polling_watermark"}); got != want {
		t.Fatalf("registered source reference = %+v, want %+v", got, want)
	}
	if got, want := resolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_destination"}); got != want {
		t.Fatalf("registered destination reference = %+v, want %+v", got, want)
	}
	warehouseResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: warehouse,
		Stream:      "snapshot",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-warehouse preflight = %v", err)
	}
	if got, want := warehouseResolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "local_parquet_warehouse"}); got != want {
		t.Fatalf("registered warehouse destination reference = %+v, want %+v", got, want)
	}
	githubResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: github,
		Stream:      "issues",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned GitHub-to-GitHub preflight = %v", err)
	}
	if got, want := githubResolved.Source.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyDeclarativeAPI, ID: "issue_label_source"}); got != want {
		t.Fatalf("registered GitHub source reference = %+v, want %+v", got, want)
	}
	postgresResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      postgres,
		Destination: postgres,
		Stream:      "snapshot",
		Mode:        synccontract.ModeIncrementalUpsert,
	})
	if err != nil {
		t.Fatalf("definition-owned PostgreSQL-to-PostgreSQL preflight = %v", err)
	}
	if got, want := postgresResolved.Destination.TransportExecutorReference(), (connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target"}); got != want {
		t.Fatalf("registered PostgreSQL destination reference = %+v, want %+v", got, want)
	}
	githubPostgresResolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: postgres,
		Stream:      "issues",
		Mode:        synccontract.ModeFullAppend,
	})
	if err != nil {
		t.Fatalf("definition-owned GitHub-to-PostgreSQL preflight = %v", err)
	}
	if got, want := githubPostgresResolved.Source.TransportExecutorReference(), issueLabelSourceReference; got != want {
		t.Fatalf("registered API source reference = %+v, want %+v", got, want)
	}
	if got, want := githubPostgresResolved.Destination.TransportExecutorReference(), postgresResolved.Destination.TransportExecutorReference(); got != want {
		t.Fatalf("registered API destination reference = %+v, want %+v", got, want)
	}
	if a.shouldRunTransport(Connection{}, "issues", SyncMode{ContractMode: synccontract.ModeFullAppend}, github, postgres) != true {
		t.Fatal("declared GitHub-to-PostgreSQL route was not selected for production dispatch")
	}
	if postgres.Metadata().Capabilities.Write {
		t.Fatal("PostgreSQL published generic write capability for its closed managed destination")
	}
	if err := validateClosedTransportBatchSize(github, github, 2); err == nil {
		t.Fatal("closed issue-label destination accepted a batch larger than its one-record contract")
	}
	if err := validateClosedTransportBatchSize(github, postgres, 50); err != nil {
		t.Fatalf("GitHub managed-target transport rejected its bounded collection batch: %v", err)
	}
	if err := validateClosedTransportBatchSize(github, postgres, issueCollectionTransportMaxRecords+1); err == nil {
		t.Fatal("GitHub managed-target transport accepted an allocation-sized batch above its fixed bound")
	}
	if err := validateClosedTransportBatchSize(postgres, postgres, 1000); err != nil {
		t.Fatalf("PostgreSQL managed transport rejected its bounded database batch: %v", err)
	}
}

func TestLocalWarehouseDestinationExecutorWritesAndReadBacksConnectionOwnedParquet(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	warehouseConnector, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("local warehouse connector is not registered")
	}
	conn := Connection{
		ID:          "connection_transport_warehouse",
		Name:        "transport-warehouse",
		Source:      EndpointConfig{Connector: "postgres"},
		Destination: EndpointConfig{Connector: "warehouse"},
		Streams: map[string]StreamConfig{
			"snapshot": {DestinationTable: "snapshot_rows"},
		},
	}
	a.state.Connections = append(a.state.Connections, conn)
	executor, err := newLocalWarehouseDestinationExecutor(a, warehouseConnector)
	if err != nil {
		t.Fatal(err)
	}
	runtime := connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"path": t.TempDir()}}
	strategy, err := localWarehouseApplyStrategy(synccontract.ModeFullAppend)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := executor.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
		Connector:     warehouseConnector,
		Runtime:       runtime,
		Stream:        "snapshot",
		Mode:          synccontract.ModeFullAppend,
		ApplyStrategy: strategy,
	})
	if err != nil {
		t.Fatalf("PlanDestination() error = %v", err)
	}
	receipt := synctransport.WarehouseReceipt{
		ID:               "stage_transport_warehouse",
		Owner:            conn.ID,
		Generation:       1,
		Stream:           "snapshot",
		Mode:             synccontract.ModeFullAppend,
		CheckpointSHA256: "checkpoint",
		TombstonesSHA256: "tombstones",
		ManifestSHA256:   "manifest",
		ContentSHA256:    "content",
		ParquetSHA256:    "parquet",
		Records:          1,
	}
	workset := synctransport.WarehouseWorkset{ID: receipt.ID, Records: []connectors.Record{{"id": "row-1", "name": "Ada"}}}
	ack, err := executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
		ConnectionID: conn.ID,
		Plan:         plan,
		Receipt:      receipt,
		Workset:      workset,
		Runtime:      runtime,
	})
	if err != nil {
		t.Fatalf("ApplyDestination() error = %v", err)
	}
	location, err := a.warehouseLocation(runtime.Config["path"], conn)
	if err != nil {
		t.Fatal(err)
	}
	tablePath, err := location.TablePath("snapshot_rows")
	if err != nil {
		t.Fatal(err)
	}
	var rows []warehouse.Row
	if err := warehouse.ReadTable(context.Background(), tablePath, func(row warehouse.Row) error {
		rows = append(rows, row)
		return nil
	}); err != nil {
		t.Fatalf("read connection-owned table: %v", err)
	}
	if len(rows) != 1 || rows[0]["id"] != "row-1" {
		t.Fatalf("connection-owned table rows = %#v, want the reopened workset row", rows)
	}
	if err := executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan:            plan,
		Workset:         workset,
		Acknowledgement: ack,
		Runtime:         runtime,
	}); err != nil {
		t.Fatalf("ReadBackDestination() error = %v", err)
	}
	if err := warehouse.WriteTable(context.Background(), tablePath, []warehouse.Row{{"id": "changed"}}); err != nil {
		t.Fatal(err)
	}
	if err := executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan:            plan,
		Workset:         workset,
		Acknowledgement: ack,
		Runtime:         runtime,
	}); err == nil {
		t.Fatal("ReadBackDestination() accepted a table changed after acknowledgement")
	}
}

func TestLocalWarehouseDestinationExecutorAppliesChangeCaptureTombstones(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	warehouseConnector, ok := a.registry.Get("warehouse")
	if !ok {
		t.Fatal("local warehouse connector is not registered")
	}
	conn := Connection{
		ID:          "connection_transport_change_capture",
		Name:        "transport-change-capture",
		Source:      EndpointConfig{Connector: "postgres"},
		Destination: EndpointConfig{Connector: "warehouse"},
		Streams: map[string]StreamConfig{
			"changes": {DestinationTable: "change_rows", PrimaryKey: []string{"id"}},
		},
	}
	a.state.Connections = append(a.state.Connections, conn)
	executor, err := newLocalWarehouseDestinationExecutor(a, warehouseConnector)
	if err != nil {
		t.Fatal(err)
	}
	runtime := connectors.RuntimeConfig{ProjectDir: root, Config: map[string]string{"path": t.TempDir()}}
	strategy, err := localWarehouseApplyStrategy(synccontract.ModeChangeCapture)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := executor.PlanDestination(context.Background(), synctransport.DestinationPlanRequest{
		Connector:     warehouseConnector,
		Runtime:       runtime,
		Stream:        "changes",
		Mode:          synccontract.ModeChangeCapture,
		ApplyStrategy: strategy,
	})
	if err != nil {
		t.Fatalf("PlanDestination() error = %v", err)
	}
	receipt := synctransport.WarehouseReceipt{
		ID:               "stage_transport_change_capture",
		Owner:            conn.ID,
		Generation:       1,
		Stream:           "changes",
		Mode:             synccontract.ModeChangeCapture,
		CheckpointSHA256: "checkpoint",
		TombstonesSHA256: "tombstones",
		ManifestSHA256:   "manifest",
		ContentSHA256:    "content",
		ParquetSHA256:    "parquet",
		Records:          2,
		Tombstones:       1,
	}
	workset := synctransport.WarehouseWorkset{
		ID:      receipt.ID,
		Records: []connectors.Record{{"id": "kept", "name": "Ada"}, {"id": "removed", "name": "Alan"}},
		Tombstones: []synccontract.Tombstone{{
			Operation:   synccontract.OperationDelete,
			EventID:     synccontract.OpaqueToken("event-1"),
			Key:         json.RawMessage(`{"id":"removed"}`),
			DeleteImage: synccontract.DeleteImageKeyOnly,
			Position: synccontract.CheckpointPosition{
				Primary:    synccontract.OpaqueToken("2"),
				TieBreaker: synccontract.OpaqueToken("2"),
			},
		}},
	}
	ack, err := executor.ApplyDestination(context.Background(), synctransport.DestinationApplyRequest{
		ConnectionID: conn.ID,
		Plan:         plan,
		Receipt:      receipt,
		Workset:      workset,
		Runtime:      runtime,
	})
	if err != nil {
		t.Fatalf("ApplyDestination() error = %v", err)
	}
	location, err := a.warehouseLocation(runtime.Config["path"], conn)
	if err != nil {
		t.Fatal(err)
	}
	tablePath, err := location.TablePath("change_rows")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	if err := warehouse.ReadTable(context.Background(), tablePath, func(row warehouse.Row) error {
		id, _ := row["id"].(string)
		ids = append(ids, id)
		return nil
	}); err != nil {
		t.Fatalf("read change-capture table: %v", err)
	}
	if len(ids) != 1 || ids[0] != "kept" {
		t.Fatalf("change-capture table ids = %#v, want only the non-tombstoned record", ids)
	}
	if err := executor.ReadBackDestination(context.Background(), synctransport.DestinationReadBackRequest{
		Plan:            plan,
		Workset:         workset,
		Acknowledgement: ack,
		Runtime:         runtime,
	}); err != nil {
		t.Fatalf("ReadBackDestination() error = %v", err)
	}
}
