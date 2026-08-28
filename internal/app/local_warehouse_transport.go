package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

// localWarehouseTransportDefinitionFactories supplies the one concrete
// materializer owned by the local warehouse primitive. RegisterDeclaredTransports
// still selects it from the descriptor's exact reference and evidence; this
// composition root never picks a connector by name or capability.
func localWarehouseTransportDefinitionFactories(a *App) []synctransport.DefinitionFactory {
	return []synctransport.DefinitionFactory{{
		Reference:           connectors.LocalWarehouseDestinationTransportReference,
		DestinationEvidence: connectors.LocalWarehouseDestinationTransportConformance,
		BuildDestination: func(connector connectors.Connector) (synctransport.DestinationExecutor, error) {
			return newLocalWarehouseDestinationExecutor(a, connector)
		},
	}}
}

type localWarehouseDestinationExecutor struct {
	app *App

	mu      sync.Mutex
	applied map[string]localWarehouseAppliedWorkset
}

type localWarehouseAppliedWorkset struct {
	finalPath string
	sha256    string
	rows      int
	strategy  connectors.DestinationApplyStrategy
}

func newLocalWarehouseDestinationExecutor(app *App, connector connectors.Connector) (*localWarehouseDestinationExecutor, error) {
	if app == nil {
		return nil, fmt.Errorf("local warehouse transport requires an app")
	}
	materializer, ok := connector.(connectors.LocalWarehouseMaterializer)
	if !ok || !materializer.MaterializesLocalWarehouse() || !isLocalWarehouseDestination(connector) {
		return nil, fmt.Errorf("local warehouse transport factory received an incompatible declaration")
	}
	return &localWarehouseDestinationExecutor{app: app, applied: make(map[string]localWarehouseAppliedWorkset)}, nil
}

func (*localWarehouseDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return connectors.LocalWarehouseDestinationTransportReference
}

func (e *localWarehouseDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if e == nil || e.app == nil || request.Connector == nil || !isLocalWarehouseDestination(request.Connector) {
		return synctransport.DestinationPlan{}, fmt.Errorf("local warehouse transport received an incompatible declaration")
	}
	strategy, err := localWarehouseApplyStrategy(request.Mode)
	if err != nil || strategy != request.ApplyStrategy {
		return synctransport.DestinationPlan{}, fmt.Errorf("local warehouse transport received an undeclared apply strategy")
	}
	return synctransport.DestinationPlan{ApplyStrategy: strategy}, nil
}

func (e *localWarehouseDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if e == nil || e.app == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport is unavailable")
	}
	strategy, err := localWarehouseApplyStrategy(request.Plan.ApplyStrategy.Mode)
	if err != nil || strategy != request.Plan.ApplyStrategy {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport received an undeclared apply strategy")
	}
	if err := request.Receipt.Validate(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport receipt: %w", err)
	}
	if request.ConnectionID == "" || request.ConnectionID != request.Receipt.Owner || request.Workset.ID != request.Receipt.ID || len(request.Workset.Records) != request.Receipt.Records || len(request.Workset.Tombstones) != request.Receipt.Tombstones {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport received a workset that does not match its receipt")
	}
	if request.Plan.ApplyStrategy.Mode != request.Receipt.Mode {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport receipt mode does not match its declared apply strategy")
	}
	conn, ok := e.app.findConnectionByID(request.ConnectionID)
	if !ok {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport connection %q was not found", request.ConnectionID)
	}
	if !e.connectionOwnsLocalWarehouse(conn) {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport connection %q does not declare the local warehouse destination", request.ConnectionID)
	}
	stream, ok := conn.Streams[request.Receipt.Stream]
	if !ok {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("local warehouse transport connection %q does not declare stream %q", request.ConnectionID, request.Receipt.Stream)
	}
	rawRecords, err := localWarehouseTransportRawRecords(request.Receipt, request.Workset, stream, strategy)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}

	dir := localWarehouseDir(request.Runtime)
	if err := warehouse.CheckLegacyLayout(dir); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := warehouse.CheckLegacyTableFormat(dir); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("create warehouse directory: %w", err)
	}
	location, err := e.app.warehouseLocation(dir, conn)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := location.EnsureOwnership(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	table := stream.DestinationTable
	if table == "" {
		table = request.Receipt.Stream
	}
	finalPath, err := location.TablePath(table)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	rawPath, err := location.WALPath(request.Receipt.Stream)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := location.AssertOwnedTable(finalPath, table); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := os.MkdirAll(filepath.Dir(rawPath), 0o700); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("create warehouse wal directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o700); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("create warehouse tables directory: %w", err)
	}

	if err := writeLocalWarehouseTransportWAL(ctx, rawPath, request.Receipt.ID, rawRecords, strategy.Strategy == connectors.ApplyStrategyReplace); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	rows, err := materializeLocalWarehouseTransportTable(ctx, rawPath, finalPath, strategy.Strategy)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("materialize local warehouse transport table: %w", err)
	}
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(finalPath)); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(rawPath)); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	sha256, _, err := digestPayloadFile(finalPath)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("digest local warehouse transport table: %w", err)
	}
	e.mu.Lock()
	e.applied[request.Workset.ID] = localWarehouseAppliedWorkset{finalPath: finalPath, sha256: sha256, rows: rows, strategy: strategy}
	e.mu.Unlock()
	return synccontract.NewDurableDownstreamAcknowledgement(connectors.Warehouse{}.Name(), time.Now().UTC())
}

func (e *localWarehouseDestinationExecutor) ReadBackDestination(ctx context.Context, request synctransport.DestinationReadBackRequest) error {
	if e == nil || e.app == nil || request.Workset.ID == "" {
		return fmt.Errorf("local warehouse transport read-back received an invalid workset")
	}
	strategy, err := localWarehouseApplyStrategy(request.Plan.ApplyStrategy.Mode)
	if err != nil || strategy != request.Plan.ApplyStrategy {
		return fmt.Errorf("local warehouse transport read-back received an undeclared apply strategy")
	}
	if request.Acknowledgement.Sink != (connectors.Warehouse{}).Name() || request.Acknowledgement.AcknowledgedAt.IsZero() {
		return fmt.Errorf("local warehouse transport read-back received an invalid acknowledgement")
	}
	e.mu.Lock()
	applied, ok := e.applied[request.Workset.ID]
	e.mu.Unlock()
	if !ok || applied.strategy != strategy {
		return fmt.Errorf("local warehouse transport has no durable apply record for workset %q", request.Workset.ID)
	}
	sha256, _, err := digestPayloadFile(applied.finalPath)
	if err != nil {
		return fmt.Errorf("read back local warehouse transport table: %w", err)
	}
	if sha256 != applied.sha256 {
		return fmt.Errorf("local warehouse transport table changed after acknowledgement")
	}
	rows := 0
	if err := warehouse.ReadTable(ctx, applied.finalPath, func(warehouse.Row) error {
		rows++
		return nil
	}); err != nil {
		return fmt.Errorf("read back local warehouse transport rows: %w", err)
	}
	if rows != applied.rows {
		return fmt.Errorf("local warehouse transport table rows = %d, want %d", rows, applied.rows)
	}
	return nil
}

func (e *localWarehouseDestinationExecutor) connectionOwnsLocalWarehouse(conn Connection) bool {
	if e.app.registry == nil {
		return false
	}
	destination, ok := e.app.registry.Get(conn.Destination.Connector)
	return ok && isLocalWarehouseDestination(destination)
}

func isLocalWarehouseDestination(connector connectors.Connector) bool {
	descriptor, ok := connectors.DestinationTransportDescriptorOf(connector)
	return ok && descriptor.Executor == connectors.LocalWarehouseDestinationTransportReference &&
		descriptor.Conformance == connectors.LocalWarehouseDestinationTransportConformance &&
		descriptor.Acknowledgement == connectors.TransportAcknowledgementDurableWarehouse
}

func localWarehouseApplyStrategy(mode synccontract.Mode) (connectors.DestinationApplyStrategy, error) {
	descriptor := connectors.Warehouse{}.SyncTransportDescriptor()
	if descriptor == nil || descriptor.Destination == nil {
		return connectors.DestinationApplyStrategy{}, fmt.Errorf("local warehouse destination declaration is unavailable")
	}
	return descriptor.Destination.ApplyStrategyFor(mode)
}

func localWarehouseDedupes(strategy connectors.ApplyStrategy) bool {
	switch strategy {
	case connectors.ApplyStrategyMerge, connectors.ApplyStrategyDedupe, connectors.ApplyStrategyDedupeHistory:
		return true
	default:
		return false
	}
}

func materializeLocalWarehouseTransportTable(ctx context.Context, rawPath, finalPath string, strategy connectors.ApplyStrategy) (int, error) {
	if strategy == connectors.ApplyStrategyDedupeHistory {
		return materializeHistoryFinalTable(ctx, rawPath, finalPath)
	}
	return materializeFinalTable(ctx, rawPath, finalPath, localWarehouseDedupes(strategy), nil)
}

func localWarehouseTransportRawRecords(receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset, stream StreamConfig, strategy connectors.DestinationApplyStrategy) ([]localRawRecord, error) {
	deduped := localWarehouseDedupes(strategy.Strategy)
	if deduped && len(stream.PrimaryKey) == 0 {
		return nil, fmt.Errorf("local warehouse transport %q requires primary key fields", strategy.Mode)
	}
	if !deduped && strategy.Mode != synccontract.ModeIncrementalAppend && len(workset.Tombstones) > 0 {
		return nil, fmt.Errorf("local warehouse transport %q cannot apply tombstones", strategy.Mode)
	}
	if strategy.Strategy == connectors.ApplyStrategyDedupeHistory && stream.CursorField == "" {
		return nil, fmt.Errorf("local warehouse transport %q requires a cursor field", strategy.Mode)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	raw := make([]localRawRecord, 0, len(workset.Records)+len(workset.Tombstones))
	for index, record := range workset.Records {
		cloned := cloneRecord(record)
		primaryKey := ""
		if deduped {
			var err error
			primaryKey, err = primaryKeyTuple(cloned, stream.PrimaryKey)
			if err != nil {
				return nil, fmt.Errorf("local warehouse transport record %d: %w", index, err)
			}
		}
		cursor := ""
		if strategy.Mode == synccontract.ModeIncrementalAppend {
			// Incremental append is checkpointed by the provider token carried by
			// the acknowledged workset. A row field is not a valid substitute for
			// that source position.
			cursor = string(workset.CandidateCheckpoint.Position.Primary)
		} else {
			var err error
			cursor, err = recordCursor(cloned, stream.CursorField)
			if err != nil {
				return nil, fmt.Errorf("local warehouse transport record %d: %w", index, err)
			}
		}
		raw = append(raw, localRawRecord{
			RawID:        fmt.Sprintf("%s:%012d", receipt.ID, index+1),
			RunID:        receipt.ID,
			SyncID:       receipt.ID,
			GenerationID: receipt.Generation,
			ExtractedAt:  now,
			LoadedAt:     now,
			Cursor:       cursor,
			PrimaryKey:   primaryKey,
			Record:       cloned,
		})
	}
	for index, tombstone := range workset.Tombstones {
		if err := tombstone.Validate(); err != nil {
			return nil, fmt.Errorf("local warehouse transport tombstone %d: %w", index, err)
		}
		if tombstone.Operation != synccontract.OperationDelete {
			return nil, fmt.Errorf("local warehouse transport does not support %s tombstones", tombstone.Operation)
		}
		var key connectors.Record
		if err := json.Unmarshal(tombstone.Key, &key); err != nil || key == nil {
			return nil, fmt.Errorf("local warehouse transport tombstone %d has an invalid key", index)
		}
		if strategy.Mode == synccontract.ModeIncrementalAppend {
			// Append preserves deletions as changelog rows rather than folding
			// them away. The marker is added only by the warehouse materializer;
			// the provider record emitted by the source remains untouched.
			key["_polymetrics_deleted"] = true
		}
		primaryKey, err := primaryKeyTuple(key, stream.PrimaryKey)
		if err != nil {
			return nil, fmt.Errorf("local warehouse transport tombstone %d: %w", index, err)
		}
		if strategy.Strategy == connectors.ApplyStrategyDedupeHistory && len(tombstone.Position.Primary) == 0 {
			return nil, fmt.Errorf("local warehouse transport tombstone %d requires a source cursor", index)
		}
		raw = append(raw, localRawRecord{
			RawID:        fmt.Sprintf("%s:tombstone:%012d", receipt.ID, index+1),
			RunID:        receipt.ID,
			SyncID:       receipt.ID,
			GenerationID: receipt.Generation,
			ExtractedAt:  now,
			LoadedAt:     now,
			Cursor:       string(tombstone.Position.Primary),
			PrimaryKey:   primaryKey,
			Deleted:      true,
			Record:       key,
		})
	}
	return raw, nil
}

func writeLocalWarehouseTransportWAL(ctx context.Context, rawPath, receiptID string, records []localRawRecord, replace bool) error {
	rawTarget := rawPath
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if replace {
		receiptPart, err := warehouse.PathComponent("transport receipt", receiptID)
		if err != nil {
			return err
		}
		rawTarget = rawPath + "." + receiptPart + ".tmp"
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		defer func() { _ = os.Remove(rawTarget) }()
	}
	file, err := os.OpenFile(rawTarget, flags, 0o600)
	if err != nil {
		return fmt.Errorf("open local warehouse transport wal: %w", err)
	}
	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return err
		}
		if err := encoder.Encode(record); err != nil {
			_ = file.Close()
			return fmt.Errorf("write local warehouse transport wal: %w", err)
		}
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync local warehouse transport wal: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close local warehouse transport wal: %w", err)
	}
	if replace {
		if err := os.Rename(rawTarget, rawPath); err != nil {
			return fmt.Errorf("replace local warehouse transport wal: %w", err)
		}
	}
	return nil
}
