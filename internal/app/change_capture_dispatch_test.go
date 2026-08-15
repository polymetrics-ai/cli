package app

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

type appChangeCaptureSource struct {
	*scriptedSyncSource
	readCalls            int
	receivedTransactions int
	restoredReceipts     int
	failAfterReceipt     bool
	receipt              connectors.CDCTransactionReceipt
}

func newAppChangeCaptureSource() *appChangeCaptureSource {
	return &appChangeCaptureSource{scriptedSyncSource: newScriptedSyncSource("app_change_capture", nil)}
}

func (s *appChangeCaptureSource) Metadata() connectors.Metadata {
	metadata := s.scriptedSyncSource.Metadata()
	metadata.Capabilities.CDC = true
	return metadata
}

func (s *appChangeCaptureSource) Definition() connectors.Definition {
	descriptor := appChangeCaptureDescriptor()
	metadata := s.Metadata()
	return connectors.Definition{
		Name:            metadata.Name,
		DisplayName:     metadata.DisplayName,
		IntegrationType: "database",
		ReleaseStage:    "test",
		Capabilities:    metadata.Capabilities,
		Changefeed:      &descriptor,
		// Mirror PostgreSQL's independently valid bounded-snapshot transport.
		// Descriptor presence must not divert an exact implemented changefeed
		// into generic transport preflight for an undeclared CDC mode.
		SyncTransport: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
			Executor:        connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "app_change_capture_snapshot"},
			EligibleStreams: []string{"records"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend, synccontract.ModeFullOverwrite},
			Delivery: connectors.DeliveryGuarantees{
				Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
				Ordering:    connectors.DeliveryOrderingSource,
				Deletes:     connectors.DeliveryDeletesUnavailable,
			},
			Conformance: connectors.ConformanceEvidenceReference{Suite: "app_change_capture_snapshot", RunID: "bounded_full_v1"},
		}},
	}
}

func (*appChangeCaptureSource) ChangefeedExecutorDescriptor() connectors.ChangefeedExecutorDescriptor {
	descriptor := appChangeCaptureDescriptor()
	return connectors.ChangefeedExecutorDescriptor{
		Status:     descriptor.Status,
		Mechanism:  descriptor.Mechanism,
		Executor:   *descriptor.Executor,
		Checkpoint: *descriptor.Checkpoint,
	}
}

func (s *appChangeCaptureSource) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	s.readCalls++
	events := []connectors.CDCEvent{
		{Operation: "insert", Record: connectors.Record{"id": "cdc-1", "name": "alpha"}},
		{Operation: "insert", Record: connectors.Record{"id": "cdc-2", "name": "beta"}},
	}
	if req.TransactionReceiver != nil && s.receipt.ID() != "" {
		restorer, ok := req.TransactionReceiver.(connectors.CDCTransactionReceiptRestorer)
		if !ok {
			return errors.New("application change-capture receiver cannot restore a durable receipt")
		}
		if err := restorer.RestoreCDCTransactionReceipt(ctx, "test-transaction", s.receipt); err != nil {
			return err
		}
		s.restoredReceipts++
	} else if req.TransactionReceiver != nil {
		transaction, err := connectors.NewCDCTransaction("test-transaction", int64(len(events)), func(_ context.Context, transactionEmit func(connectors.CDCEvent) error) error {
			for _, event := range events {
				if err := transactionEmit(event); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		receipt, err := req.TransactionReceiver.ReceiveCDCTransaction(ctx, transaction)
		if err != nil {
			return err
		}
		s.receipt = receipt
		s.receivedTransactions++
		if s.failAfterReceipt {
			s.failAfterReceipt = false
			return errors.New("simulated process loss after warehouse receipt")
		}
	} else {
		for _, event := range events {
			if err := emit(event); err != nil {
				return err
			}
		}
	}
	if req.DurableCheckpointCommitter == nil {
		return errors.New("application change-capture dispatch omitted its durable checkpoint committer")
	}
	observed := time.Now().UTC().Add(-time.Second)
	positionObserved := true
	return req.DurableCheckpointCommitter.CommitDurableChangefeedCheckpoint(ctx, synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           synccontract.SourceIdentity{Engine: s.Name(), AccountOrCluster: "cluster", ObjectScope: req.Stream},
		Mechanism:        "test_changefeed",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "test", Token: synccontract.OpaqueToken("barrier")},
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("lsn-1"), TieBreaker: synccontract.OpaqueToken("lsn-1")},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: synccontract.OpaqueToken("generation"),
		SchemaVersion:    "test-v1",
		ProtocolVersion:  "test-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "transaction", Value: synccontract.OpaqueToken("tx-1")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "transaction", Start: synccontract.OpaqueToken("lsn-1"), End: synccontract.OpaqueToken("lsn-1")},
		ObservedAt:       observed,
	})
}

func appChangeCaptureDescriptor() connectors.ChangefeedDescriptor {
	return connectors.ChangefeedDescriptor{
		Status:    connectors.ChangefeedStatusImplemented,
		Mechanism: connectors.ChangefeedMechanismLogicalReplication,
		Source: connectors.ChangefeedSource{
			ArtifactURL:     "https://example.com/changefeed",
			ArtifactVersion: "1",
			RetrievedAt:     "2026-08-15",
		},
		Executor:   &connectors.ChangefeedExecutorRef{Kind: "native", ID: "app_test_changefeed"},
		Checkpoint: &connectors.ChangefeedCheckpoint{Kind: "lsn", Keys: []string{"lsn"}, CommitAfter: "downstream_ack", OnInvalid: "resnapshot_required"},
		Delivery:   &connectors.ChangefeedDelivery{Ordering: "source_transaction_order", Duplicates: "at_least_once", Deletes: "tombstone", DedupeKey: []string{"lsn"}},
		Streams:    []string{"runtime_catalog"},
	}
}

func TestRunETLChangeCapturePublishesCommittedTransactionToConnectionWarehouse(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := newAppChangeCaptureSource()
	a.Registry().Register(source)
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": warehouseDir}}); err != nil {
		t.Fatal(err)
	}
	const connectionName = "cdc_to_warehouse"
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        connectionName,
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: string(synccontract.ModeChangeCapture), PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	run, err := a.RunETL(ctx, RunETLRequest{Connection: connectionName, Stream: "records", BatchSize: 1})
	if err != nil {
		t.Fatalf("RunETL(change_capture) error = %v", err)
	}
	if source.readCalls != 1 {
		t.Fatalf("changefeed reads = %d, want one production application dispatch", source.readCalls)
	}
	if run.Status != "completed" || run.RecordsRead != 2 || run.RecordsLoaded != 2 {
		t.Fatalf("change_capture run = %#v, want two durably materialized records", run)
	}

	connection, ok := a.findConnection(connectionName)
	if !ok {
		t.Fatal("change_capture connection disappeared")
	}
	location, err := a.warehouseLocation(warehouseDir, connection)
	if err != nil {
		t.Fatal(err)
	}
	tablePath, err := location.TablePath("records")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := parquetRowIDs(readParquetFile(t, tablePath)), []string{"cdc-1", "cdc-2"}; !slices.Equal(got, want) {
		t.Fatalf("change_capture warehouse rows = %v, want %v", got, want)
	}
	streamState := a.state.StreamStates[streamStateKey(connectionName, "records")]
	if streamState.Checkpoint == nil || streamState.Checkpoint.CommittedAt == nil || string(streamState.Checkpoint.Position.Primary) != "lsn-1" {
		t.Fatalf("change_capture stream state = %#v, want the committed full checkpoint", streamState)
	}
}

func TestRunETLChangeCaptureRestoresWarehouseReceiptBeforeCheckpointAfterRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := newAppChangeCaptureSource()
	source.failAfterReceipt = true
	a.Registry().Register(source)
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "warehouse", Connector: "warehouse", Config: map[string]string{"path": warehouseDir}}); err != nil {
		t.Fatal(err)
	}
	const connectionName = "cdc_receipt_restart"
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        connectionName,
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: string(synccontract.ModeChangeCapture), PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connectionName, Stream: "records", BatchSize: 1}); err == nil || !strings.Contains(err.Error(), "simulated process loss after warehouse receipt") {
		t.Fatalf("first RunETL(change_capture) error = %v, want simulated post-receipt failure", err)
	}
	if state := a.state.StreamStates[streamStateKey(connectionName, "records")]; state.Checkpoint != nil {
		t.Fatalf("failed change_capture state = %#v, want no checkpoint before receipt restoration", state)
	}

	run, err := a.RunETL(ctx, RunETLRequest{Connection: connectionName, Stream: "records", BatchSize: 1})
	if err != nil {
		t.Fatalf("restarted RunETL(change_capture) error = %v", err)
	}
	if run.RecordsLoaded != 0 || source.receivedTransactions != 1 || source.restoredReceipts != 1 {
		t.Fatalf("restarted change_capture = run=%#v received=%d restored=%d, want receipt-only checkpoint recovery", run, source.receivedTransactions, source.restoredReceipts)
	}
	state := a.state.StreamStates[streamStateKey(connectionName, "records")]
	if state.Checkpoint == nil || state.Checkpoint.CommittedAt == nil || string(state.Checkpoint.Position.Primary) != "lsn-1" {
		t.Fatalf("restarted change_capture state = %#v, want restored durable checkpoint", state)
	}
}
