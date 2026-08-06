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
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/synccontract"
)

type etlExecutionResult struct {
	RecordsRead        int
	RecordsTransformed int
	RecordsLoaded      int
	RecordsFailed      int
	BatchCount         int
	Checkpoint         map[string]string
	PendingStreamState *pendingStreamState
}

type pendingStreamState struct {
	Key   string
	State StreamState
}

type localRawRecord struct {
	RawID        string            `json:"_polymetrics_raw_id"`
	RunID        string            `json:"_polymetrics_run_id"`
	SyncID       string            `json:"_polymetrics_sync_id"`
	GenerationID int64             `json:"_polymetrics_generation_id"`
	ExtractedAt  string            `json:"_polymetrics_extracted_at"`
	LoadedAt     string            `json:"_polymetrics_loaded_at"`
	Cursor       string            `json:"_polymetrics_cursor,omitempty"`
	PrimaryKey   string            `json:"_polymetrics_primary_key,omitempty"`
	Deleted      bool              `json:"_polymetrics_deleted"`
	Record       connectors.Record `json:"record"`
}

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

	dir := localWarehouseDir(destRuntime)
	table := stream.DestinationTable
	if table == "" {
		table = streamName
	}
	finalPath := localWarehouseTablePath(dir, table)
	tmpFinalPath := finalPath + "." + runID + ".tmp"
	rawPath := localRawPath(dir, conn.Name, streamName, table)
	tmpRawPath := rawPath + "." + runID + ".tmp"
	if err := a.prepareLocalWarehouseDirectories(dir, filepath.Dir(rawPath)); err != nil {
		return etlExecutionResult{}, err
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

	var finalFile *os.File
	var finalEncoder *json.Encoder
	if !mode.IsDeduped() {
		finalTarget := finalPath
		finalFlags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
		if mode.IsOverwrite() {
			finalTarget = tmpFinalPath
			finalFlags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		}
		finalFile, err = os.OpenFile(finalTarget, finalFlags, 0o600)
		if err != nil {
			_ = rawFile.Close()
			return etlExecutionResult{}, fmt.Errorf("open final table: %w", err)
		}
		finalEncoder = json.NewEncoder(finalFile)
	}

	success := false
	defer func() {
		_ = rawFile.Close()
		if finalFile != nil {
			_ = finalFile.Close()
		}
		if !success && mode.IsOverwrite() {
			_ = os.Remove(tmpRawPath)
			_ = os.Remove(tmpFinalPath)
		}
	}()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	result := etlExecutionResult{}
	recordBatch := make([]connectors.Record, 0, batchSize)
	rawBatch := make([]localRawRecord, 0, batchSize)
	priorCursor := streamStateCursor(prior)
	nextCursor := priorCursor
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
		if finalEncoder != nil {
			for _, record := range recordBatch {
				if err := ctx.Err(); err != nil {
					return err
				}
				if err := finalEncoder.Encode(record); err != nil {
					return fmt.Errorf("write final record: %w", err)
				}
			}
		}
		result.BatchCount++
		rawBatch = rawBatch[:0]
		recordBatch = recordBatch[:0]
		return nil
	}

	readConfig := sourceRuntime
	readConfig.Config = cloneStringMap(sourceRuntime.Config)
	if priorCursor != "" {
		readConfig.Config["since"] = priorCursor
	}
	err = source.Read(ctx, connectors.ReadRequest{
		Stream: streamName,
		Config: readConfig,
		State:  map[string]string{"cursor": priorCursor, "generation_id": strconv.FormatInt(generationID, 10)},
	}, func(record connectors.Record) error {
		result.RecordsRead++
		cursor := ""
		if stream.CursorField != "" {
			var err error
			cursor, err = recordCursor(record, stream.CursorField)
			if err != nil {
				return err
			}
			if mode.Source == SourceSyncIncremental && priorCursor != "" && compareCursor(cursor, priorCursor) < 0 {
				return nil
			}
			if nextCursor == "" || compareCursor(cursor, nextCursor) > 0 {
				nextCursor = cursor
			}
		}
		deleted := isDeletedRecord(record)
		enriched := cloneRecord(record)
		enriched["_polymetrics_run_id"] = runID
		enriched["_polymetrics_synced_at"] = now
		enriched["_polymetrics_deleted"] = deleted
		if cursor != "" {
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
			PrimaryKey:   pk,
			Deleted:      deleted,
			Record:       enriched,
		}
		rawBatch = append(rawBatch, raw)
		recordBatch = append(recordBatch, enriched)
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
	if finalFile != nil {
		if err := finalFile.Sync(); err != nil {
			return result, fmt.Errorf("sync final table: %w", err)
		}
		if err := finalFile.Close(); err != nil {
			return result, fmt.Errorf("close final table: %w", err)
		}
	}

	if mode.IsDeduped() {
		readRawPath := rawPath
		if mode.IsOverwrite() {
			readRawPath = tmpRawPath
		}
		finalCount, err := materializeDedupedFinal(ctx, readRawPath, tmpFinalPath)
		if err != nil {
			return result, err
		}
		result.RecordsLoaded = finalCount
	}

	if mode.IsOverwrite() {
		if err := os.Rename(tmpRawPath, rawPath); err != nil {
			return result, fmt.Errorf("replace raw table: %w", err)
		}
	}
	if mode.IsOverwrite() || mode.IsDeduped() {
		if err := os.Rename(tmpFinalPath, finalPath); err != nil {
			return result, fmt.Errorf("replace final table: %w", err)
		}
	}
	for _, outputDir := range []string{filepath.Dir(rawPath), filepath.Dir(finalPath)} {
		if err := syncLocalWarehouseDirectory(outputDir); err != nil {
			return result, err
		}
	}

	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(destination.Name(), time.Now().UTC())
	if err != nil {
		return result, err
	}
	updated, err := committedLegacyStreamState(conn, sourceExpectation, streamName, stream, runID, nextCursor, generationID, result.RecordsLoaded, observedAt, acknowledgement)
	if err != nil {
		return result, err
	}
	result.Checkpoint = checkpointForResult(result, mode, stateKey, updated)
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

func (a *App) prepareLocalWarehouseDirectories(warehouseDir, rawDir string) (err error) {
	lock := statestore.FileLock{Path: a.statePath + ".warehouse.lock"}
	unlock, err := lock.Lock()
	if err != nil {
		return fmt.Errorf("lock warehouse directory setup: %w", err)
	}
	defer func() {
		if unlockErr := unlock(); err == nil && unlockErr != nil {
			err = fmt.Errorf("unlock warehouse directory setup: %w", unlockErr)
		}
	}()

	warehouseCreated, err := mkdirAllTrackingCreatedDirectories(warehouseDir)
	if err != nil {
		return fmt.Errorf("create warehouse directory: %w", err)
	}
	rawCreated, err := mkdirAllTrackingCreatedDirectories(rawDir)
	if err != nil {
		return fmt.Errorf("create raw directory: %w", err)
	}
	return syncCreatedLocalWarehouseDirectoryParents(append(warehouseCreated, rawCreated...), make(map[string]struct{}))
}

// mkdirAllTrackingCreatedDirectories returns the newly created directories in
// leaf-to-ancestor order. The parent entry for each must be synced separately
// before a checkpoint can acknowledge the warehouse write.
func mkdirAllTrackingCreatedDirectories(dir string) ([]string, error) {
	if dir == "" {
		return nil, os.MkdirAll(dir, 0o700)
	}
	path := filepath.Clean(dir)
	paths := make([]string, 0)
	for parent := filepath.Dir(path); parent != path; parent = filepath.Dir(path) {
		paths = append(paths, path)
		path = parent
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect warehouse directory: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("warehouse directory is not a directory")
	}

	created := make([]string, 0, len(paths))
	for index := len(paths) - 1; index >= 0; index-- {
		path := paths[index]
		err := os.Mkdir(path, 0o700)
		if err == nil {
			created = append(created, path)
			continue
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("inspect warehouse directory: %w", err)
		}
		if !info.IsDir() {
			return nil, errors.New("warehouse directory is not a directory")
		}
	}
	for left, right := 0, len(created)-1; left < right; left, right = left+1, right-1 {
		created[left], created[right] = created[right], created[left]
	}
	return created, nil
}

// syncCreatedLocalWarehouseDirectoryParents makes every newly created
// directory entry durable through the first ancestor that existed before
// MkdirAll. Directories already synced after their writes are not repeated.
func syncCreatedLocalWarehouseDirectoryParents(created []string, synced map[string]struct{}) error {
	for _, dir := range created {
		parent := filepath.Clean(filepath.Dir(dir))
		if _, ok := synced[parent]; ok {
			continue
		}
		if err := syncLocalWarehouseDirectory(parent); err != nil {
			return fmt.Errorf("sync newly created warehouse directory parent: %w", err)
		}
		synced[parent] = struct{}{}
	}
	return nil
}

func checkpointForResult(result etlExecutionResult, mode SyncMode, stateKey string, state StreamState) map[string]string {
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
	if cursor := streamStateCursor(state); cursor != "" {
		checkpoint["cursor"] = cursor
	}
	return checkpoint
}

func materializeDedupedFinal(ctx context.Context, rawPath, finalPath string) (int, error) {
	best, err := readBestLocalRawRecords(ctx, rawPath)
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
	file, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return 0, fmt.Errorf("open deduped final table: %w", err)
	}
	encoder := json.NewEncoder(file)
	count := 0
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return count, err
		}
		if err := encoder.Encode(best[key].Record); err != nil {
			_ = file.Close()
			return count, fmt.Errorf("write deduped final record: %w", err)
		}
		count++
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return count, fmt.Errorf("sync deduped final table: %w", err)
	}
	if err := file.Close(); err != nil {
		return count, fmt.Errorf("close deduped final table: %w", err)
	}
	return count, nil
}

func rawRecordNewer(candidate, current localRawRecord) bool {
	if cmp := compareCursor(candidate.Cursor, current.Cursor); cmp != 0 {
		return cmp > 0
	}
	if cmp := compareCursor(candidate.ExtractedAt, current.ExtractedAt); cmp != 0 {
		return cmp > 0
	}
	return candidate.RawID > current.RawID
}

func readBestLocalRawRecords(ctx context.Context, path string) (map[string]localRawRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]localRawRecord{}, nil
		}
		return nil, fmt.Errorf("open raw table: %w", err)
	}
	defer func() { _ = file.Close() }()
	reader := bufio.NewScanner(file)
	reader.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	best := map[string]localRawRecord{}
	for reader.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}
		var record localRawRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("decode raw record: %w", err)
		}
		if record.PrimaryKey == "" {
			return nil, fmt.Errorf("raw record %s is missing primary key metadata", record.RawID)
		}
		current, ok := best[record.PrimaryKey]
		if !ok || rawRecordNewer(record, current) {
			best[record.PrimaryKey] = record
		}
	}
	if err := reader.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("scan raw records: %w", err)
	}
	return best, nil
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

func localWarehouseTablePath(dir, table string) string {
	return filepath.Join(dir, localSafeName(table)+".jsonl")
}

func localRawPath(dir, connection, stream, table string) string {
	name := localSafeName(connection + "__" + stream + "__" + table)
	return filepath.Join(dir, "_pm_raw", name+".jsonl")
}

func localSafeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		case r == '.' || r == '/' || r == ' ' || r == ':':
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "records"
	}
	return b.String()
}
