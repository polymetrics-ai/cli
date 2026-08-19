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
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
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

// TestDeclaredChangeCaptureRoutesRefuseBeforeIO enumerates the current four
// production route cells rather than treating the dispatcher guard as a proxy
// for their declaration-level truth. R3 deliberately remains a refusal: the
// separately proven R4 route is PostgreSQL CDC -> connection warehouse ->
// PostgreSQL history target, never a direct CDC destination mode.
func TestDeclaredChangeCaptureRoutesRefuseBeforeIO(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		destination string
		reason      string
	}{
		{
			name:        "R1 GitHub CDC to GitHub",
			source:      "github",
			destination: "github",
			reason:      "change capture requires a connection-owned local warehouse destination",
		},
		{
			name:        "R2 GitHub CDC to PostgreSQL",
			source:      "github",
			destination: "postgres",
			reason:      "change capture requires a connection-owned local warehouse destination",
		},
		{
			name:        "R3 PostgreSQL CDC to GitHub",
			source:      "postgres",
			destination: "github",
			reason:      "change capture requires a connection-owned local warehouse destination",
		},
		{
			name:        "R4 PostgreSQL CDC to PostgreSQL",
			source:      "postgres",
			destination: "postgres",
			reason:      "change capture requires a connection-owned local warehouse destination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			if err := InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := Open(root)
			if err != nil {
				t.Fatal(err)
			}

			source := newDeclaredChangeCaptureIOProbe(t, tt.source)
			destination := source
			if tt.destination != tt.source {
				destination = newDeclaredChangeCaptureIOProbe(t, tt.destination)
			}
			registry := connectors.NewEmptyRegistry()
			registry.Register(source)
			if destination != source {
				registry.Register(destination)
			}
			a.registry = registry

			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
				t.Fatal(err)
			}
			if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "destination", Connector: destination.Name()}); err != nil {
				t.Fatal(err)
			}
			if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
				Name:        "declared_change_capture_refusal",
				Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
				Destination: EndpointConfig{Connector: destination.Name(), Credential: "destination"},
				Streams: map[string]StreamConfig{
					"records": {
						SyncMode:         string(synccontract.ModeChangeCapture),
						PrimaryKey:       []string{"id"},
						DestinationTable: "records",
					},
				},
			}); err != nil {
				t.Fatal(err)
			}
			// Connection creation discovers a source catalog. The assertion is
			// specifically that executing the refused route adds no I/O.
			source.resetIO()
			if destination != source {
				destination.resetIO()
			}

			_, err = a.RunETL(ctx, RunETLRequest{Connection: "declared_change_capture_refusal", Stream: "records", BatchSize: 1})
			var modeErr *synccontract.ModeNotExecutableError
			if !errors.As(err, &modeErr) || modeErr.Mode != synccontract.ModeChangeCapture || modeErr.Reason != tt.reason {
				t.Fatalf("RunETL() error = %T %v, want change_capture ModeNotExecutableError reason %q", err, err, tt.reason)
			}
			source.assertNoIO(t)
			if destination != source {
				destination.assertNoIO(t)
			}
		})
	}
}

// declaredChangeCaptureIOProbe preserves the production bundle definition
// while making every runtime connector operation observable. It lets this
// matrix prove the pre-I/O boundary without calling an API or database.
type declaredChangeCaptureIOProbe struct {
	name       string
	metadata   connectors.Metadata
	definition connectors.Definition
	checks     int
	catalogs   int
	reads      int
	cdcReads   int
	writes     int
}

func newDeclaredChangeCaptureIOProbe(t *testing.T, name string) *declaredChangeCaptureIOProbe {
	t.Helper()
	bundle, err := engine.Load(defs.FS, name)
	if err != nil {
		t.Fatalf("load declared %s bundle: %v", name, err)
	}
	declared := engine.New(bundle, nil)
	definition := declared.Definition()
	if definition.Name != name {
		t.Fatalf("declared connector identity = %q, want %q", definition.Name, name)
	}
	return &declaredChangeCaptureIOProbe{name: name, metadata: declared.Metadata(), definition: definition}
}

func (c *declaredChangeCaptureIOProbe) Name() string                  { return c.name }
func (c *declaredChangeCaptureIOProbe) Metadata() connectors.Metadata { return c.metadata }
func (c *declaredChangeCaptureIOProbe) Definition() connectors.Definition {
	return c.definition
}
func (c *declaredChangeCaptureIOProbe) Check(context.Context, connectors.RuntimeConfig) error {
	c.checks++
	return errors.New("declared change-capture refusal probe Check must not be called")
}
func (c *declaredChangeCaptureIOProbe) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	c.catalogs++
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "records", PrimaryKey: []string{"id"}}}}, nil
}
func (c *declaredChangeCaptureIOProbe) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	c.reads++
	return errors.New("declared change-capture refusal probe Read must not be called")
}
func (c *declaredChangeCaptureIOProbe) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	c.writes++
	return connectors.WriteResult{}, errors.New("declared change-capture refusal probe Write must not be called")
}
func (c *declaredChangeCaptureIOProbe) ChangefeedExecutorDescriptor() connectors.ChangefeedExecutorDescriptor {
	if c.definition.Changefeed == nil {
		return connectors.ChangefeedExecutorDescriptor{}
	}
	return connectors.ChangefeedExecutorDescriptor{
		Status:     c.definition.Changefeed.Status,
		Mechanism:  c.definition.Changefeed.Mechanism,
		Executor:   *c.definition.Changefeed.Executor,
		Checkpoint: *c.definition.Changefeed.Checkpoint,
	}
}
func (c *declaredChangeCaptureIOProbe) ReadCDC(context.Context, connectors.CDCReadRequest, func(connectors.CDCEvent) error) error {
	c.cdcReads++
	return errors.New("declared change-capture refusal probe ReadCDC must not be called")
}
func (c *declaredChangeCaptureIOProbe) resetIO() {
	c.checks = 0
	c.catalogs = 0
	c.reads = 0
	c.cdcReads = 0
	c.writes = 0
}
func (c *declaredChangeCaptureIOProbe) assertNoIO(t *testing.T) {
	t.Helper()
	if c.checks != 0 || c.catalogs != 0 || c.reads != 0 || c.cdcReads != 0 || c.writes != 0 {
		t.Fatalf("declared connector %q I/O calls check=%d catalog=%d read=%d read_cdc=%d write=%d, want zero", c.name, c.checks, c.catalogs, c.reads, c.cdcReads, c.writes)
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
