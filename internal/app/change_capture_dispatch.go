package app

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

type warehouseChangeCaptureReceiver struct {
	app             *App
	runID           string
	connection      Connection
	streamName      string
	stream          StreamConfig
	mode            SyncMode
	destination     connectors.Connector
	location        warehouse.Location
	rawPath         string
	finalPath       string
	stateKey        string
	generationID    int64
	workLease       *transportWorkLease
	expectedState   StreamState
	expectedPresent bool
	acknowledgement synccontract.DownstreamAcknowledgement
	receiptPending  bool
	transactions    int
	records         int
	lastCheckpoint  *synccontract.CheckpointEnvelope
}

var ErrChangeCaptureArtifactReconciliationRequired = errors.New("change capture warehouse artifact reconciliation is required")

// ChangeCaptureArtifactReconciliationError prevents a recovered CDC receipt
// from acknowledging source progress unless its private warehouse artifact
// manifest still proves the exact previously durable artifacts.
type ChangeCaptureArtifactReconciliationError struct{ Reason string }

func (e *ChangeCaptureArtifactReconciliationError) Error() string {
	if e == nil || strings.TrimSpace(e.Reason) == "" {
		return ErrChangeCaptureArtifactReconciliationRequired.Error()
	}
	return ErrChangeCaptureArtifactReconciliationRequired.Error() + ": " + e.Reason
}

func (e *ChangeCaptureArtifactReconciliationError) Unwrap() error {
	return ErrChangeCaptureArtifactReconciliationRequired
}

const changeCaptureWarehouseArtifactManifestDirectory = "cdc-receipts"

func changeCaptureWarehouseArtifactManifestPath(location warehouse.Location, stream string, generationID int64, transactionKey string) (string, error) {
	if generationID <= 0 {
		return "", errors.New("change capture artifact manifest generation is invalid")
	}
	streamPart, err := warehouse.PathComponent("stream", stream)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(transactionKey) == "" || len(transactionKey) > 1024 {
		return "", errors.New("change capture artifact manifest transaction identity is invalid")
	}
	transactionDigest := sha256.Sum256([]byte("polymetrics-change-capture-artifact-path-v1\x00" + transactionKey))
	return filepath.Join(
		location.ConnectionDir,
		changeCaptureWarehouseArtifactManifestDirectory,
		streamPart,
		fmt.Sprintf("generation-%020d", generationID),
		hex.EncodeToString(transactionDigest[:])+".json",
	), nil
}

func newChangeCaptureArtifactReconciliationError(reason string) error {
	return &ChangeCaptureArtifactReconciliationError{Reason: reason}
}

func writeChangeCaptureWarehouseArtifactManifest(path string, manifest connectors.CDCArtifactManifest) error {
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("validate change capture artifact manifest: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create change capture artifact manifest directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".cdc-artifact-manifest-*")
	if err != nil {
		return fmt.Errorf("create change capture artifact manifest: %w", err)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := json.NewEncoder(temporary).Encode(manifest); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("encode change capture artifact manifest: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync change capture artifact manifest: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close change capture artifact manifest: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish change capture artifact manifest: %w", err)
	}
	removeTemporary = false
	return nil
}

func readChangeCaptureWarehouseArtifactManifest(path string) (connectors.CDCArtifactManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return connectors.CDCArtifactManifest{}, err
	}
	defer func() { _ = file.Close() }()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var manifest connectors.CDCArtifactManifest
	if err := decoder.Decode(&manifest); err != nil {
		return connectors.CDCArtifactManifest{}, fmt.Errorf("decode change capture artifact manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return connectors.CDCArtifactManifest{}, errors.New("change capture artifact manifest has trailing values")
		}
		return connectors.CDCArtifactManifest{}, fmt.Errorf("decode change capture artifact manifest trailing value: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return connectors.CDCArtifactManifest{}, err
	}
	return manifest, nil
}

func newWarehouseChangeCaptureReceiver(app *App, lease *transportWorkLease, runID string, connection Connection, streamName string, stream StreamConfig, mode SyncMode, destination connectors.Connector, destinationRuntime connectors.RuntimeConfig) (*warehouseChangeCaptureReceiver, error) {
	if app == nil || lease == nil || strings.TrimSpace(runID) == "" {
		return nil, errors.New("change capture warehouse receiver is unavailable")
	}
	dir := localWarehouseDir(destinationRuntime)
	if err := warehouse.CheckLegacyLayout(dir); err != nil {
		return nil, err
	}
	if err := warehouse.CheckLegacyTableFormat(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create warehouse directory: %w", err)
	}
	location, err := app.warehouseLocation(dir, connection)
	if err != nil {
		return nil, err
	}
	if err := location.EnsureOwnership(); err != nil {
		return nil, err
	}
	table := stream.DestinationTable
	if table == "" {
		table = streamName
	}
	finalPath, err := location.TablePath(table)
	if err != nil {
		return nil, err
	}
	rawPath, err := location.WALPath(streamName)
	if err != nil {
		return nil, err
	}
	if err := location.AssertOwnedTable(finalPath, table); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(rawPath), 0o700); err != nil {
		return nil, fmt.Errorf("create warehouse wal directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o700); err != nil {
		return nil, fmt.Errorf("create warehouse tables directory: %w", err)
	}
	stateKey := streamStateKey(connection.Name, streamName)
	prior := lease.stateForTerminalRun()
	generationID := prior.GenerationID
	if generationID == 0 {
		generationID = 1
	}
	return &warehouseChangeCaptureReceiver{
		app:             app,
		runID:           runID,
		connection:      connection,
		streamName:      streamName,
		stream:          stream,
		mode:            mode,
		destination:     destination,
		location:        location,
		rawPath:         rawPath,
		finalPath:       finalPath,
		stateKey:        stateKey,
		generationID:    generationID,
		workLease:       lease,
		expectedState:   prior,
		expectedPresent: true,
	}, nil
}

func (r *warehouseChangeCaptureReceiver) ReceiveCDCTransaction(ctx context.Context, transaction connectors.CDCTransaction) (connectors.CDCTransactionReceipt, error) {
	if r == nil || r.app == nil {
		return connectors.CDCTransactionReceipt{}, errors.New("change capture warehouse receiver is unavailable")
	}
	if err := r.workLease.renew(ctx); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if r.receiptPending {
		return connectors.CDCTransactionReceipt{}, errors.New("change capture checkpoint is still pending for the prior warehouse transaction")
	}
	r.transactions++
	temporary := fmt.Sprintf("%s.%s.%06d.cdc.tmp", r.rawPath, r.runID, r.transactions)
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("open change capture warehouse transaction: %w", err)
	}
	removeTemporary := true
	defer func() {
		_ = file.Close()
		if removeTemporary {
			_ = os.Remove(temporary)
		}
	}()
	if err := copyChangeCaptureWAL(ctx, file, r.rawPath); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}

	encoder := json.NewEncoder(file)
	transactionRecords := int64(0)
	err = transaction.StreamEvents(ctx, func(event connectors.CDCEvent) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		operation := strings.ToLower(strings.TrimSpace(event.Operation))
		switch operation {
		case "insert", "update", "delete":
		default:
			return fmt.Errorf("change capture event has unsupported operation %q", event.Operation)
		}
		record := cloneRecord(event.Record)
		deleted := operation == "delete"
		record["_polymetrics_run_id"] = r.runID
		record["_polymetrics_synced_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		record["_polymetrics_deleted"] = deleted
		primaryKey, err := primaryKeyTuple(record, r.stream.PrimaryKey)
		if err != nil {
			return err
		}
		transactionRecords++
		if transactionRecords > int64(int(^uint(0)>>1)) {
			return errors.New("change capture transaction record count exceeds platform limits")
		}
		raw := localRawRecord{
			RawID:        fmt.Sprintf("%s:%06d:%012d", r.runID, r.transactions, transactionRecords),
			RunID:        r.runID,
			SyncID:       r.runID,
			GenerationID: r.generationID,
			ExtractedAt:  time.Now().UTC().Format(time.RFC3339Nano),
			LoadedAt:     time.Now().UTC().Format(time.RFC3339Nano),
			PrimaryKey:   primaryKey,
			Deleted:      deleted,
			Record:       record,
		}
		if err := encoder.Encode(raw); err != nil {
			return fmt.Errorf("write change capture warehouse transaction: %w", err)
		}
		return nil
	})
	if err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if transactionRecords != transaction.Records() {
		return connectors.CDCTransactionReceipt{}, errors.New("change capture transaction record count does not match its source receipt")
	}
	if err := file.Sync(); err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("sync change capture warehouse transaction: %w", err)
	}
	if err := file.Close(); err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("close change capture warehouse transaction: %w", err)
	}
	if err := os.Rename(temporary, r.rawPath); err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("publish change capture warehouse transaction: %w", err)
	}
	removeTemporary = false
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(r.rawPath)); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if _, err := materializeFinalTable(ctx, r.rawPath, r.finalPath, true, nil); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(r.finalPath)); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	rawWALSHA256, _, err := digestPayloadFile(r.rawPath)
	if err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("digest change capture warehouse raw WAL: %w", err)
	}
	finalTableSHA256, _, err := digestPayloadFile(r.finalPath)
	if err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("digest change capture warehouse final table: %w", err)
	}
	manifest, err := connectors.NewCDCArtifactManifest(
		r.connection.ID,
		r.streamName,
		r.generationID,
		transaction.ID(),
		transaction.Records(),
		transaction.ContentDigest(),
		rawWALSHA256,
		finalTableSHA256,
	)
	if err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("bind change capture warehouse artifacts: %w", err)
	}
	manifestPath, err := changeCaptureWarehouseArtifactManifestPath(r.location, r.streamName, r.generationID, transaction.ID())
	if err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if err := writeChangeCaptureWarehouseArtifactManifest(manifestPath, manifest); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(manifestPath)); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}

	durableAt := time.Now().UTC()
	receiptID := changeCaptureWarehouseReceiptID(r.connection.ID, transaction.ID())
	receipt, err := connectors.NewCDCTransactionReceiptWithArtifactManifest(receiptID, "warehouse:"+r.connection.ID, durableAt, manifest)
	if err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	acknowledgement, err := receipt.Acknowledgement()
	if err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	r.acknowledgement = acknowledgement
	r.receiptPending = true
	r.records += int(transactionRecords)
	return receipt, nil
}

func (r *warehouseChangeCaptureReceiver) CommitDurableChangefeedCheckpoint(ctx context.Context, candidate synccontract.CheckpointEnvelope) error {
	if r == nil || !r.receiptPending {
		return errors.New("change capture checkpoint has no durable warehouse transaction receipt")
	}
	var committed synccontract.CheckpointEnvelope
	err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, r.acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		updated, err := r.workLease.commit(ctx, checkpoint)
		if err != nil {
			return fmt.Errorf("persist change capture warehouse checkpoint: %w", err)
		}
		updated.Connection = r.connection.Name
		updated.Stream = r.streamName
		updated.GenerationID = r.generationID
		updated.RecordsLoaded = r.records
		updated.UpdatedAt = r.acknowledgement.AcknowledgedAt
		// The receipt's checkpoint must remain bound to the same fence while
		// metadata is refreshed. The second lease mutation is still pre-I/O and
		// rejects a stale owner rather than recreating the former late CAS.
		if _, err := r.workLease.mutate(ctx, func(StreamState) (StreamState, error) { return updated, nil }); err != nil {
			return fmt.Errorf("persist change capture stream metadata: %w", err)
		}
		r.expectedState = cloneStreamState(updated)
		r.expectedPresent = true
		committed = checkpoint.Clone()
		return nil
	})
	if err != nil {
		return err
	}
	r.lastCheckpoint = &committed
	r.receiptPending = false
	r.acknowledgement = synccontract.DownstreamAcknowledgement{}
	return nil
}

func (r *warehouseChangeCaptureReceiver) RestoreCDCTransactionReceipt(ctx context.Context, transactionID string, receipt connectors.CDCTransactionReceipt) error {
	if r == nil || r.app == nil {
		return errors.New("change capture warehouse receiver is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.workLease.renew(ctx); err != nil {
		return err
	}
	if r.receiptPending {
		return errors.New("change capture checkpoint is still pending for another warehouse transaction")
	}
	if strings.TrimSpace(transactionID) == "" || receipt.ID() != changeCaptureWarehouseReceiptID(r.connection.ID, transactionID) {
		return errors.New("change capture staged receipt does not match the connection-owned warehouse transaction")
	}
	acknowledgement, err := receipt.Acknowledgement()
	if err != nil {
		return err
	}
	if acknowledgement.Sink != "warehouse:"+r.connection.ID {
		return errors.New("change capture staged receipt names another downstream sink")
	}
	manifest, err := receipt.ArtifactManifest()
	if err != nil {
		return newChangeCaptureArtifactReconciliationError("staged receipt has no valid artifact manifest")
	}
	if manifest.ConnectionID != r.connection.ID || manifest.Stream != r.streamName || manifest.GenerationID != r.generationID || manifest.TransactionKey != transactionID {
		return newChangeCaptureArtifactReconciliationError("staged receipt manifest does not match this connection, stream, generation, and transaction")
	}
	manifestPath, err := changeCaptureWarehouseArtifactManifestPath(r.location, r.streamName, r.generationID, transactionID)
	if err != nil {
		return newChangeCaptureArtifactReconciliationError("staged receipt manifest path is invalid")
	}
	storedManifest, err := readChangeCaptureWarehouseArtifactManifest(manifestPath)
	if err != nil || storedManifest != manifest {
		return newChangeCaptureArtifactReconciliationError("durable artifact manifest does not match the staged receipt")
	}
	rawWALSHA256, _, err := digestPayloadFile(r.rawPath)
	if err != nil || rawWALSHA256 != manifest.RawWALSHA256 {
		return newChangeCaptureArtifactReconciliationError("raw WAL artifact does not match the staged receipt")
	}
	finalTableSHA256, _, err := digestPayloadFile(r.finalPath)
	if err != nil || finalTableSHA256 != manifest.FinalTableSHA256 {
		return newChangeCaptureArtifactReconciliationError("final table artifact does not match the staged receipt")
	}
	r.acknowledgement = acknowledgement
	r.receiptPending = true
	return nil
}

func (r *warehouseChangeCaptureReceiver) result(completed bool) etlExecutionResult {
	result := etlExecutionResult{
		RecordsRead:        r.records,
		RecordsTransformed: r.records,
		RecordsLoaded:      r.records,
		BatchCount:         r.transactions,
	}
	if r.lastCheckpoint == nil || !r.expectedPresent {
		return result
	}
	state := cloneStreamState(r.expectedState)
	if completed {
		state = r.workLease.stateForTerminalRun()
		state.LastSuccessfulRunID = r.runID
	}
	result.PendingStreamState = &pendingStreamState{Key: r.stateKey, State: state}
	result.Checkpoint = checkpointForResult(result, r.mode, r.stateKey, state, "", false)
	return result
}

func (a *App) runWarehouseChangeCapture(ctx context.Context, request etlModeDispatchRequest) (etlExecutionResult, error) {
	if !hasImplementedChangefeedExecutor(request.source) {
		return etlExecutionResult{}, &synccontract.ModeNotExecutableError{Mode: synccontract.ModeChangeCapture, Reason: "source has no matching implemented changefeed executor"}
	}
	changefeed := request.source.(connectors.ChangefeedExecutor)
	stateKey := streamStateKey(request.connection.Name, request.streamName)
	workLease, err := a.claimTransportWorkLease(ctx, stateKey, request.connection.Name, request.streamName, request.runID, request.sourceExpectation, false, request.transportAdmissionFence)
	if err != nil {
		return etlExecutionResult{}, err
	}
	receiver, err := newWarehouseChangeCaptureReceiver(a, workLease, request.runID, request.connection, request.streamName, request.stream, request.mode, request.destination, request.destinationRuntime)
	if err != nil {
		return etlExecutionResult{}, err
	}
	prior := workLease.stateForTerminalRun()
	err = changefeed.ReadCDC(ctx, connectors.CDCReadRequest{
		Stream:                     request.streamName,
		Config:                     request.sourceRuntime,
		Checkpoint:                 prior.Checkpoint,
		TransactionReceiver:        receiver,
		DurableCheckpointCommitter: receiver,
	}, func(connectors.CDCEvent) error {
		return errors.New("application change capture requires a committed transaction receiver")
	})
	result := receiver.result(err == nil)
	if err != nil {
		return result, err
	}
	if result.PendingStreamState == nil {
		return result, errors.New("change capture completed without a durable warehouse checkpoint")
	}
	return result, nil
}

func hasImplementedChangefeedExecutor(source connectors.Connector) bool {
	definition, ok := connectors.DefinitionOf(source)
	if !ok || !connectors.HasImplementedChangefeed(source, definition.Changefeed) {
		return false
	}
	_, ok = source.(connectors.ChangefeedExecutor)
	return ok
}

func copyChangeCaptureWAL(ctx context.Context, destination io.Writer, existing string) error {
	source, err := os.Open(existing)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open existing change capture warehouse wal: %w", err)
	}
	defer func() { _ = source.Close() }()
	reader := bufio.NewReaderSize(source, 32*1024)
	buffer := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		read, readErr := reader.Read(buffer)
		if read > 0 {
			if _, err := destination.Write(buffer[:read]); err != nil {
				return fmt.Errorf("copy existing change capture warehouse wal: %w", err)
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read existing change capture warehouse wal: %w", readErr)
		}
	}
}

func changeCaptureWarehouseReceiptID(connectionID, transactionID string) string {
	digest := sha256.Sum256([]byte("polymetrics-change-capture-warehouse-v1\x00" + connectionID + "\x00" + transactionID))
	return "warehouse-cdc-" + hex.EncodeToString(digest[:])
}
