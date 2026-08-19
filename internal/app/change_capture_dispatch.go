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
	rawPath         string
	finalPath       string
	stateKey        string
	generationID    int64
	expectedState   StreamState
	expectedPresent bool
	acknowledgement synccontract.DownstreamAcknowledgement
	receiptPending  bool
	transactions    int
	records         int
	lastCheckpoint  *synccontract.CheckpointEnvelope
}

func newWarehouseChangeCaptureReceiver(app *App, runID string, connection Connection, streamName string, stream StreamConfig, mode SyncMode, destination connectors.Connector, destinationRuntime connectors.RuntimeConfig) (*warehouseChangeCaptureReceiver, error) {
	if app == nil || strings.TrimSpace(runID) == "" {
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
	prior, present := app.state.StreamStates[stateKey]
	prior = cloneStreamState(prior)
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
		rawPath:         rawPath,
		finalPath:       finalPath,
		stateKey:        stateKey,
		generationID:    generationID,
		expectedState:   prior,
		expectedPresent: present,
	}, nil
}

func (r *warehouseChangeCaptureReceiver) ReceiveCDCTransaction(ctx context.Context, transaction connectors.CDCTransaction) (connectors.CDCTransactionReceipt, error) {
	if r == nil || r.app == nil {
		return connectors.CDCTransactionReceipt{}, errors.New("change capture warehouse receiver is unavailable")
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

	durableAt := time.Now().UTC()
	receiptID := changeCaptureWarehouseReceiptID(r.connection.ID, transaction.ID())
	receipt, err := connectors.NewCDCTransactionReceipt(receiptID, "warehouse:"+r.connection.ID, durableAt)
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

func (r *warehouseChangeCaptureReceiver) CommitDurableChangefeedCheckpoint(_ context.Context, candidate synccontract.CheckpointEnvelope) error {
	if r == nil || !r.receiptPending {
		return errors.New("change capture checkpoint has no durable warehouse transaction receipt")
	}
	var committed synccontract.CheckpointEnvelope
	err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, r.acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		updated := cloneStreamState(r.expectedState)
		updated.Connection = r.connection.Name
		updated.Stream = r.streamName
		updated.Checkpoint = &checkpoint
		updated.GenerationID = r.generationID
		updated.RecordsLoaded = r.records
		updated.UpdatedAt = r.acknowledgement.AcknowledgedAt
		if _, err := r.app.updateState(func(current state) (state, error) {
			currentState, present := current.StreamStates[r.stateKey]
			if present != r.expectedPresent || (present && !transportStreamStateEqual(currentState, r.expectedState)) {
				return current, errTransportStreamStateConflict
			}
			if current.StreamStates == nil {
				current.StreamStates = map[string]StreamState{}
			}
			current.StreamStates[r.stateKey] = cloneStreamState(updated)
			return current, nil
		}); err != nil {
			return fmt.Errorf("persist change capture warehouse checkpoint: %w", err)
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
	for _, path := range []string{r.rawPath, r.finalPath} {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("verify recovered change capture warehouse artifact: %w", err)
		}
		if !info.Mode().IsRegular() {
			return errors.New("recovered change capture warehouse artifact is not a regular file")
		}
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
	receiver, err := newWarehouseChangeCaptureReceiver(a, request.runID, request.connection, request.streamName, request.stream, request.mode, request.destination, request.destinationRuntime)
	if err != nil {
		return etlExecutionResult{}, err
	}
	prior := a.state.StreamStates[streamStateKey(request.connection.Name, request.streamName)]
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
