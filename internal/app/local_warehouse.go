package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/durability"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

type etlExecutionResult struct {
	RecordsRead               int
	RecordsTransformed        int
	RecordsLoaded             int
	RecordsFailed             int
	BatchCount                int
	Checkpoint                map[string]string
	TransportPhaseMeasurement *TransportPhaseMeasurement
	PendingStreamState        *pendingStreamState
}

func cloneTransportPhaseMeasurement(measurement *TransportPhaseMeasurement) *TransportPhaseMeasurement {
	if measurement == nil {
		return nil
	}
	clone := *measurement
	return &clone
}

type pendingStreamState struct {
	Key   string
	State StreamState
}

type localRawRecord struct {
	RawID        string               `json:"_polymetrics_raw_id"`
	RunID        string               `json:"_polymetrics_run_id"`
	SyncID       string               `json:"_polymetrics_sync_id"`
	GenerationID int64                `json:"_polymetrics_generation_id"`
	ExtractedAt  string               `json:"_polymetrics_extracted_at"`
	LoadedAt     string               `json:"_polymetrics_loaded_at"`
	Cursor       string               `json:"_polymetrics_cursor,omitempty"`
	CursorState  *localRawCursorState `json:"_polymetrics_cursor_state,omitempty"`
	PrimaryKey   string               `json:"_polymetrics_primary_key,omitempty"`
	Deleted      bool                 `json:"_polymetrics_deleted"`
	Record       connectors.Record    `json:"record"`
}

type localRawCursorState struct {
	Token []byte `json:"token"`
}

// syncLocalWarehouseDirectoryCommit is a test seam for the platform directory
// sync primitive; production defaults to durability.SyncDirectory.
var syncLocalWarehouseDirectoryCommit = durability.SyncDirectory

func (a *App) runWarehouseETL(ctx context.Context, runID string, conn Connection, source connectors.Connector, sourceRuntime connectors.RuntimeConfig, destination connectors.Connector, destRuntime connectors.RuntimeConfig, sourceExpectation synccontract.ResumeExpectation, streamName string, stream StreamConfig, mode SyncMode, batchSize int) (etlExecutionResult, error) {
	stateKey := streamStateKey(conn.Name, streamName)
	prior := a.state.StreamStates[stateKey]
	if prior.Checkpoint != nil {
		if err := validateStreamStateResume(prior, sourceExpectation); err != nil {
			return etlExecutionResult{}, err
		}
	}
	generationID := prior.GenerationID
	if generationID == 0 || mode.IsOverwrite() {
		generationID++
	}
	cursorTracker, err := newStreamCursorTracker(prior, source, sourceRuntime, stream.CursorField, mode.Source)
	if err != nil {
		return etlExecutionResult{}, err
	}

	dir := localWarehouseDir(destRuntime)
	if err := warehouse.CheckLegacyLayout(dir); err != nil {
		return etlExecutionResult{}, err
	}
	// pm does not write into a warehouse it will not read. Without this, a sync
	// into a pre-Parquet warehouse reported records loaded and exit status 0
	// while every read of that warehouse was refused — an operator told at once
	// that the sync worked and that the table cannot be read.
	if err := warehouse.CheckLegacyTableFormat(dir); err != nil {
		return etlExecutionResult{}, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return etlExecutionResult{}, fmt.Errorf("create warehouse directory: %w", err)
	}
	table := stream.DestinationTable
	if table == "" {
		table = streamName
	}
	location, err := a.warehouseLocation(dir, conn)
	if err != nil {
		return etlExecutionResult{}, err
	}
	if err := location.EnsureOwnership(); err != nil {
		return etlExecutionResult{}, err
	}
	finalPath, err := location.TablePath(table)
	if err != nil {
		return etlExecutionResult{}, err
	}
	rawPath, err := location.WALPath(streamName)
	if err != nil {
		return etlExecutionResult{}, err
	}
	if mode.IsDeduped() && !mode.IsOverwrite() && cursorTracker.sourceOrdered != nil {
		if err := validateSourceOrderedLocalRawRecords(ctx, rawPath); err != nil {
			return etlExecutionResult{}, err
		}
	}
	// Directory nesting already makes a shared table path unrepresentable.
	// Re-deriving ownership from the table path is the independent check that
	// fails loudly if a future change ever reintroduces a shared path.
	if err := location.AssertOwnedTable(finalPath, table); err != nil {
		return etlExecutionResult{}, err
	}
	tmpRawPath := rawPath + "." + runID + ".tmp"
	if err := os.MkdirAll(filepath.Dir(rawPath), 0o700); err != nil {
		return etlExecutionResult{}, fmt.Errorf("create warehouse wal directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o700); err != nil {
		return etlExecutionResult{}, fmt.Errorf("create warehouse tables directory: %w", err)
	}

	rawTarget := rawPath
	rawFlags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if mode.IsOverwrite() {
		rawTarget = tmpRawPath
		rawFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	rawFile, err := os.OpenFile(rawTarget, rawFlags, 0o600)
	if err != nil {
		return etlExecutionResult{}, fmt.Errorf("open raw table: %w", err)
	}
	rawEncoder := json.NewEncoder(rawFile)

	success := false
	defer func() {
		_ = rawFile.Close()
		if !success && mode.IsOverwrite() {
			_ = os.Remove(tmpRawPath)
		}
	}()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	result := etlExecutionResult{}
	rawBatch := make([]localRawRecord, 0, batchSize)
	rawSeq := 0
	observedAt := time.Time{}

	flush := func() error {
		if len(rawBatch) == 0 {
			return nil
		}
		for _, raw := range rawBatch {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := rawEncoder.Encode(raw); err != nil {
				return fmt.Errorf("write raw record: %w", err)
			}
		}
		result.BatchCount++
		rawBatch = rawBatch[:0]
		return nil
	}

	err = source.Read(ctx, cursorTracker.readRequest(streamName, sourceRuntime, prior, generationID, mode.Source), func(record connectors.Record) error {
		result.RecordsRead++
		cursor := ""
		cursorState := connectors.OpaqueCursorState{}
		if stream.CursorField != "" {
			var include bool
			var err error
			cursor, cursorState, include, err = cursorTracker.observe(record, stream.CursorField, mode.Source)
			if err != nil {
				return err
			}
			if !include {
				return nil
			}
		}
		deleted := isDeletedRecord(record)
		enriched := cloneRecord(record)
		enriched["_polymetrics_run_id"] = runID
		enriched["_polymetrics_synced_at"] = now
		enriched["_polymetrics_deleted"] = deleted
		if stream.CursorField != "" {
			enriched["_polymetrics_cursor"] = cursor
		}

		pk := ""
		if len(stream.PrimaryKey) > 0 {
			var err error
			pk, err = primaryKeyTuple(enriched, stream.PrimaryKey)
			if err != nil {
				return err
			}
		}
		rawSeq++
		raw := localRawRecord{
			RawID:        fmt.Sprintf("%s:%012d", runID, rawSeq),
			RunID:        runID,
			SyncID:       runID,
			GenerationID: generationID,
			ExtractedAt:  time.Now().UTC().Format(time.RFC3339Nano),
			LoadedAt:     now,
			Cursor:       cursor,
			CursorState:  localRawCursorStateFor(cursorState),
			PrimaryKey:   pk,
			Deleted:      deleted,
			Record:       enriched,
		}
		rawBatch = append(rawBatch, raw)
		result.RecordsTransformed++
		observedAt = time.Now().UTC()
		result.RecordsLoaded++
		if len(rawBatch) >= batchSize {
			return flush()
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if err := flush(); err != nil {
		return result, err
	}
	if err := rawFile.Sync(); err != nil {
		return result, fmt.Errorf("sync raw table: %w", err)
	}
	if err := rawFile.Close(); err != nil {
		return result, fmt.Errorf("close raw table: %w", err)
	}

	// The write-ahead log is durable and complete, so it becomes the log of
	// record before the table is derived from it. If materialization then
	// fails, the next run rebuilds the table from this same log; no record is
	// lost by the table being stale.
	if mode.IsOverwrite() {
		if err := os.Rename(tmpRawPath, rawPath); err != nil {
			return result, fmt.Errorf("replace raw table: %w", err)
		}
	}

	// Every mode now materializes the table wholesale from the write-ahead
	// log. Deduped modes already did. Append modes used to stream into the
	// table O_APPEND, which Parquet cannot support: a Parquet file cannot be
	// appended to once closed. Rebuilding from the log is what makes the table
	// format a derived detail rather than a constraint on the write pattern.
	finalCount, err := materializeFinalTable(ctx, rawPath, finalPath, mode.IsDeduped(), cursorTracker.sourceOrdered)
	if err != nil {
		return result, err
	}
	if mode.IsDeduped() {
		// A deduped table's row count is the folded count, not the number of
		// records this run read. An append table's stays per-run, because the
		// table it rebuilds spans every run.
		result.RecordsLoaded = finalCount
	}

	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(finalPath)); err != nil {
		return result, err
	}
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(rawPath)); err != nil {
		return result, err
	}

	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(destination.Name(), time.Now().UTC())
	if err != nil {
		return result, err
	}
	nextCursor, nextCursorObserved := cursorTracker.checkpoint()
	updated, err := committedLegacyStreamState(conn, sourceExpectation, streamName, stream, runID, nextCursor, nextCursorObserved, generationID, result.RecordsLoaded, observedAt, acknowledgement)
	if err != nil {
		return result, err
	}
	reportedCursor, reportedCursorObserved := cursorTracker.reportedCursor()
	result.Checkpoint = checkpointForResult(result, mode, stateKey, updated, reportedCursor, reportedCursorObserved)
	result.PendingStreamState = &pendingStreamState{Key: stateKey, State: updated}
	success = true
	return result, nil
}

func syncLocalWarehouseDirectory(dir string) error {
	if err := syncLocalWarehouseDirectoryCommit(dir); err != nil {
		return fmt.Errorf("sync warehouse directory: %w", err)
	}
	return nil
}

// syncLocalWarehouseDirectoryChain synchronizes dir and every ancestor through
// the filesystem root. This establishes a known durable parent boundary after
// MkdirAll without inferring which components were new, and it must complete
// before a downstream checkpoint is acknowledged.
func syncLocalWarehouseDirectoryChain(dir string) error {
	path, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve warehouse directory: %w", err)
	}
	for {
		if err := syncLocalWarehouseDirectory(path); err != nil {
			return err
		}
		parent := filepath.Dir(path)
		if parent == path {
			return nil
		}
		path = parent
	}
}

func checkpointForResult(result etlExecutionResult, mode SyncMode, stateKey string, state StreamState, cursor string, cursorObserved bool) map[string]string {
	// This map is a backward-compatible run report, not resumable sync state.
	// StreamState.Checkpoint is the sole durable resume record.
	checkpoint := map[string]string{
		"records_read":        strconv.Itoa(result.RecordsRead),
		"records_transformed": strconv.Itoa(result.RecordsTransformed),
		"records_loaded":      strconv.Itoa(result.RecordsLoaded),
		"records_failed":      strconv.Itoa(result.RecordsFailed),
		"batches":             strconv.Itoa(result.BatchCount),
		"sync_mode":           mode.Name,
		"state_key":           stateKey,
		"generation_id":       strconv.FormatInt(state.GenerationID, 10),
	}
	if cursorObserved {
		checkpoint["cursor"] = cursor
	}
	return checkpoint
}

// materializeFinalTable rebuilds a table wholesale from its write-ahead log and
// swaps it into place as one Parquet file.
//
// A deduped table folds the log by primary key, drops deletions, and emits in
// key order. An append table replays the log in the order it was written, which
// is the order the streaming append path produced — including across earlier
// runs, because the log accumulates the same way the table used to.
func materializeFinalTable(ctx context.Context, rawPath, finalPath string, deduped bool, sourceOrdered connectors.SourceOrderedCursorReader) (int, error) {
	writer, err := warehouse.NewTableWriter(finalPath)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			writer.Abort()
		}
	}()

	if deduped {
		best, err := readBestLocalRawRecords(ctx, rawPath, sourceOrdered)
		if err != nil {
			return 0, err
		}
		keys := make([]string, 0, len(best))
		for key, raw := range best {
			if raw.Deleted {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			if err := writer.Write(best[key].Record); err != nil {
				return 0, err
			}
		}
	} else if err := forEachLocalRawRecord(ctx, rawPath, func(raw localRawRecord) error {
		return writer.Write(raw.Record)
	}); err != nil {
		return 0, err
	}

	count := writer.Rows()
	if err := writer.Commit(ctx); err != nil {
		return 0, err
	}
	committed = true
	return count, nil
}

func localRawCursorStateFor(state connectors.OpaqueCursorState) *localRawCursorState {
	if !state.Present {
		return nil
	}
	return &localRawCursorState{Token: append([]byte(nil), state.Token...)}
}

func (r localRawRecord) opaqueCursorState() (connectors.OpaqueCursorState, error) {
	if r.CursorState == nil {
		return connectors.OpaqueCursorState{}, fmt.Errorf("raw record %s is missing source cursor state", r.RawID)
	}
	return connectors.OpaqueCursorState{Token: append([]byte(nil), r.CursorState.Token...), Present: true}, nil
}

func rawRecordNewer(candidate, current localRawRecord, candidateSequence, currentSequence uint64, sourceOrdered connectors.SourceOrderedCursorReader) (bool, error) {
	if sourceOrdered != nil {
		candidateState, err := candidate.opaqueCursorState()
		if err != nil {
			return false, err
		}
		currentState, err := current.opaqueCursorState()
		if err != nil {
			return false, err
		}
		cmp, err := sourceOrdered.CompareCursorStates(candidateState, currentState)
		if err == nil && cmp != 0 {
			return cmp > 0, nil
		}
		if err != nil && !errors.Is(err, connectors.ErrOpaqueCursorOrderUnavailable) {
			return false, err
		}
		return candidateSequence > currentSequence, nil
	}
	if cmp := compareCursor(candidate.Cursor, current.Cursor); cmp != 0 {
		return cmp > 0, nil
	}
	if cmp := compareCursor(candidate.ExtractedAt, current.ExtractedAt); cmp != 0 {
		return cmp > 0, nil
	}
	return candidate.RawID > current.RawID, nil
}

type indexedLocalRawRecord struct {
	record   localRawRecord
	sequence uint64
}

func readBestLocalRawRecords(ctx context.Context, path string, sourceOrdered connectors.SourceOrderedCursorReader) (map[string]localRawRecord, error) {
	best := map[string]indexedLocalRawRecord{}
	var sequence uint64
	err := forEachLocalRawRecord(ctx, path, func(record localRawRecord) error {
		sequence++
		if record.PrimaryKey == "" {
			return fmt.Errorf("raw record %s is missing primary key metadata", record.RawID)
		}
		current, ok := best[record.PrimaryKey]
		if !ok {
			if sourceOrdered != nil {
				if _, err := record.opaqueCursorState(); err != nil {
					return err
				}
			}
			best[record.PrimaryKey] = indexedLocalRawRecord{record: record, sequence: sequence}
			return nil
		}
		newer, err := rawRecordNewer(record, current.record, sequence, current.sequence, sourceOrdered)
		if err != nil {
			return err
		}
		if newer {
			best[record.PrimaryKey] = indexedLocalRawRecord{record: record, sequence: sequence}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resolved := make(map[string]localRawRecord, len(best))
	for key, record := range best {
		resolved[key] = record.record
	}
	return resolved, nil
}

func validateSourceOrderedLocalRawRecords(ctx context.Context, path string) error {
	return forEachLocalRawRecord(ctx, path, func(record localRawRecord) error {
		_, err := record.opaqueCursorState()
		return err
	})
}

// forEachLocalRawRecord streams a write-ahead log in the order it was written.
// A log that does not exist yet is an empty log, not an error: a connection can
// be materialized before its first record ever arrives.
func forEachLocalRawRecord(ctx context.Context, path string, visit func(localRawRecord) error) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open raw table: %w", err)
	}
	defer func() { _ = file.Close() }()
	reader := bufio.NewScanner(file)
	reader.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for reader.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}
		var record localRawRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("decode raw record: %w", err)
		}
		if err := visit(record); err != nil {
			return err
		}
	}
	if err := reader.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("scan raw records: %w", err)
	}
	return nil
}

func isDeletedRecord(record connectors.Record) bool {
	for _, key := range []string{"_polymetrics_deleted", "_ab_cdc_deleted_at", "_upstream_deleted", "_deleted"} {
		value, ok := record[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case bool:
			if v {
				return true
			}
		case string:
			if strings.TrimSpace(v) != "" && strings.TrimSpace(v) != "false" {
				return true
			}
		default:
			return true
		}
	}
	return false
}

func localWarehouseDir(cfg connectors.RuntimeConfig) string {
	if cfg.Config["path"] != "" {
		return cfg.Config["path"]
	}
	return filepath.Join(cfg.ProjectDir, "warehouse")
}

// warehouseLocation resolves the directory conn owns inside a warehouse root.
// The connection's opaque ID is the path component, never its display name,
// and never anything derived from a credential.
func (a *App) warehouseLocation(dir string, conn Connection) (warehouse.Location, error) {
	return warehouse.LocationFor(dir, a.state.WorkspaceID, conn.Source.Connector, conn.ID, conn.Name)
}
