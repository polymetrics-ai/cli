package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

const (
	connectionWarehouseStageManifestVersion = 1
	connectionWarehouseStageManifestDir     = "transport"
	connectionWarehouseStagePrefix          = "transport-"
	connectionWarehouseStageReconcileLimit  = 64
)

// connectionWarehouseStage owns temporary transport worksets inside the
// connection-owned warehouse layout. It deliberately receives an opaque
// connection ID, never a caller-selected directory or credential-derived
// identity.
type connectionWarehouseStage struct{ app *App }

func newConnectionWarehouseStage(app *App) synctransport.WarehouseStage {
	return &connectionWarehouseStage{app: app}
}

type connectionWarehouseStageManifest struct {
	Version             int                             `json:"version"`
	ID                  string                          `json:"id"`
	Owner               string                          `json:"owner"`
	Generation          int64                           `json:"generation"`
	SourceName          string                          `json:"source_name"`
	DestinationName     string                          `json:"destination_name"`
	Stream              string                          `json:"stream"`
	Mode                synccontract.Mode               `json:"mode"`
	Records             int                             `json:"records"`
	Tombstones          []synccontract.Tombstone        `json:"tombstones"`
	CandidateCheckpoint synccontract.CheckpointEnvelope `json:"candidate_checkpoint"`
	CheckpointSHA256    string                          `json:"checkpoint_sha256"`
	TombstonesSHA256    string                          `json:"tombstones_sha256"`
	WALSHA256           string                          `json:"wal_sha256"`
	ParquetSHA256       string                          `json:"parquet_sha256"`
	ContentSHA256       string                          `json:"content_sha256"`
}

type connectionWarehouseStageContent struct {
	ID               string            `json:"id"`
	Owner            string            `json:"owner"`
	Generation       int64             `json:"generation"`
	SourceName       string            `json:"source_name"`
	DestinationName  string            `json:"destination_name"`
	Stream           string            `json:"stream"`
	Mode             synccontract.Mode `json:"mode"`
	Records          int               `json:"records"`
	Tombstones       int               `json:"tombstones"`
	CheckpointSHA256 string            `json:"checkpoint_sha256"`
	TombstonesSHA256 string            `json:"tombstones_sha256"`
	WALSHA256        string            `json:"wal_sha256"`
	ParquetSHA256    string            `json:"parquet_sha256"`
}

type connectionWarehouseStageArtifact struct {
	location     warehouse.Location
	tableName    string
	walPath      string
	parquetPath  string
	manifestPath string
}

func (s *connectionWarehouseStage) Stage(ctx context.Context, request synctransport.WarehouseStageRequest) (synctransport.WarehouseReceipt, error) {
	if s == nil || s.app == nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("connection-owned warehouse stage is unavailable")
	}
	conn, err := s.connectionForStageRequest(request)
	if err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	if receipt, found, err := s.recoverUncommittedReceipt(ctx, conn, request); err != nil {
		return synctransport.WarehouseReceipt{}, err
	} else if found {
		return receipt, nil
	}
	stageID, err := prefixedID("stage")
	if err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	artifact, err := s.artifactFor(conn, stageID)
	if err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	if err := artifact.location.EnsureOwnership(); err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	if err := artifact.location.AssertOwnedTable(artifact.parquetPath, artifact.tableName); err != nil {
		return synctransport.WarehouseReceipt{}, err
	}

	// The WAL is durable before DuckDB materializes its derived Parquet table.
	// A source page's maps are serialized now and never retained in the receipt.
	if err := writeConnectionStageWAL(ctx, artifact.walPath, stageID, request.Generation, request.Page.Records); err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	if _, err := materializeFinalTable(ctx, artifact.walPath, artifact.parquetPath, false, nil); err != nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("materialize staged parquet: %w", err)
	}

	walSHA256, _, err := digestPayloadFile(artifact.walPath)
	if err != nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("digest staged wal: %w", err)
	}
	parquetSHA256, _, err := digestPayloadFile(artifact.parquetPath)
	if err != nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("digest staged parquet: %w", err)
	}
	checkpoint := request.Page.CandidateCheckpoint.Clone()
	tombstones := cloneConnectionStageTombstones(request.Page.Tombstones)
	checkpointSHA256, err := hashJSON(checkpoint)
	if err != nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("hash staged checkpoint: %w", err)
	}
	tombstonesSHA256, err := hashJSON(tombstones)
	if err != nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("hash staged tombstones: %w", err)
	}
	manifest := connectionWarehouseStageManifest{
		Version:             connectionWarehouseStageManifestVersion,
		ID:                  stageID,
		Owner:               conn.ID,
		Generation:          request.Generation,
		SourceName:          request.SourceName,
		DestinationName:     request.DestinationName,
		Stream:              request.Stream,
		Mode:                request.Mode,
		Records:             len(request.Page.Records),
		Tombstones:          tombstones,
		CandidateCheckpoint: checkpoint,
		CheckpointSHA256:    checkpointSHA256,
		TombstonesSHA256:    tombstonesSHA256,
		WALSHA256:           walSHA256,
		ParquetSHA256:       parquetSHA256,
	}
	manifest.ContentSHA256, err = manifest.contentSHA256()
	if err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	if err := writeConnectionStageManifest(artifact.manifestPath, manifest); err != nil {
		return synctransport.WarehouseReceipt{}, err
	}
	manifestSHA256, _, err := digestPayloadFile(artifact.manifestPath)
	if err != nil {
		return synctransport.WarehouseReceipt{}, fmt.Errorf("digest staged manifest: %w", err)
	}
	// The receipt becomes visible only after all three published directory
	// entries have crossed the durable directory boundary.
	for _, dir := range []string{filepath.Dir(artifact.walPath), filepath.Dir(artifact.parquetPath), filepath.Dir(artifact.manifestPath)} {
		if err := syncLocalWarehouseDirectoryChain(dir); err != nil {
			return synctransport.WarehouseReceipt{}, err
		}
	}
	return manifest.receipt(manifestSHA256), nil
}

func (s *connectionWarehouseStage) recoverUncommittedReceipt(ctx context.Context, conn Connection, request synctransport.WarehouseStageRequest) (synctransport.WarehouseReceipt, bool, error) {
	if err := ctx.Err(); err != nil {
		return synctransport.WarehouseReceipt{}, false, err
	}
	checkpointSHA256, err := connectionWarehouseStageRetryCheckpointSHA256(request.Page.CandidateCheckpoint)
	if err != nil {
		return synctransport.WarehouseReceipt{}, false, fmt.Errorf("hash staged retry checkpoint: %w", err)
	}
	tombstones := cloneConnectionStageTombstones(request.Page.Tombstones)
	tombstonesSHA256, err := hashJSON(tombstones)
	if err != nil {
		return synctransport.WarehouseReceipt{}, false, fmt.Errorf("hash staged retry tombstones: %w", err)
	}
	location, err := s.app.warehouseLocation(filepath.Join(s.app.projectDir, "warehouse"), conn)
	if err != nil {
		return synctransport.WarehouseReceipt{}, false, err
	}
	entries, err := os.ReadDir(filepath.Join(location.ConnectionDir, connectionWarehouseStageManifestDir))
	if errors.Is(err, os.ErrNotExist) {
		return synctransport.WarehouseReceipt{}, false, nil
	}
	if err != nil {
		return synctransport.WarehouseReceipt{}, false, fmt.Errorf("read staged receipt directory: %w", err)
	}
	var recovered synctransport.WarehouseReceipt
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return synctransport.WarehouseReceipt{}, false, err
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		stageID := strings.TrimSuffix(entry.Name(), ".json")
		artifact, err := s.artifactFor(conn, stageID)
		if err != nil {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("resolve staged receipt %q for retry: %w", stageID, err)
		}
		manifest, err := readConnectionStageManifest(artifact.manifestPath)
		if err != nil {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("read staged receipt %q for retry: %w", stageID, err)
		}
		if manifest.Version != connectionWarehouseStageManifestVersion ||
			manifest.Owner != conn.ID ||
			manifest.Generation != request.Generation ||
			manifest.SourceName != request.SourceName ||
			manifest.DestinationName != request.DestinationName ||
			manifest.Stream != request.Stream ||
			manifest.Mode != request.Mode ||
			manifest.Records != len(request.Page.Records) ||
			len(manifest.Tombstones) != len(tombstones) ||
			manifest.TombstonesSHA256 != tombstonesSHA256 {
			continue
		}
		if streamState, present := s.app.state.StreamStates[streamStateKey(conn.Name, manifest.Stream)]; present {
			alreadyCommitted, err := committedStageCheckpointMatches(manifest.CandidateCheckpoint, streamState.Checkpoint)
			if err != nil {
				return synctransport.WarehouseReceipt{}, false, fmt.Errorf("compare staged receipt %q with committed checkpoint: %w", manifest.ID, err)
			}
			if alreadyCommitted {
				continue
			}
		}
		manifestCheckpointSHA256, err := connectionWarehouseStageRetryCheckpointSHA256(manifest.CandidateCheckpoint)
		if err != nil {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("hash staged receipt %q retry checkpoint: %w", manifest.ID, err)
		}
		if manifestCheckpointSHA256 != checkpointSHA256 {
			continue
		}
		manifestSHA256, _, err := digestPayloadFile(artifact.manifestPath)
		if err != nil {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("digest staged receipt %q manifest: %w", manifest.ID, err)
		}
		receipt := manifest.receipt(manifestSHA256)
		workset, err := s.Reopen(ctx, receipt)
		if err != nil {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("reopen staged receipt %q for retry: %w", receipt.ID, err)
		}
		equal, err := connectionWarehouseStageRecordsEqual(workset.Records, request.Page.Records)
		if err != nil {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("compare staged receipt %q records for retry: %w", receipt.ID, err)
		}
		if !equal {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("staged receipt %q matches retry source identity but has different records", receipt.ID)
		}
		if recovered.ID != "" {
			return synctransport.WarehouseReceipt{}, false, fmt.Errorf("multiple staged receipts match one retry source workset")
		}
		recovered = receipt
	}
	if recovered.ID == "" {
		return synctransport.WarehouseReceipt{}, false, nil
	}
	return recovered, true, nil
}

func connectionWarehouseStageRetryCheckpointSHA256(checkpoint synccontract.CheckpointEnvelope) (string, error) {
	checkpoint = checkpoint.Clone()
	checkpoint.ObservedAt = time.Time{}
	checkpoint.CommittedAt = nil
	return hashJSON(checkpoint)
}

func connectionWarehouseStageRecordsEqual(left, right []connectors.Record) (bool, error) {
	if len(left) != len(right) {
		return false, nil
	}
	for index := range left {
		equal, err := declarativeReadBackValuesEqual(left[index], right[index])
		if err != nil {
			return false, err
		}
		if !equal {
			return false, nil
		}
	}
	return true, nil
}

func (s *connectionWarehouseStage) Reopen(ctx context.Context, receipt synctransport.WarehouseReceipt) (synctransport.WarehouseWorkset, error) {
	if s == nil || s.app == nil {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("connection-owned warehouse stage is unavailable")
	}
	if err := receipt.Validate(); err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	conn, ok := s.connectionByID(receipt.Owner)
	if !ok {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q has an unknown connection owner", receipt.ID)
	}
	artifact, err := s.artifactFor(conn, receipt.ID)
	if err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	// Reopen never calls EnsureOwnership: a missing or altered owner is a
	// tamper signal, not an invitation to recreate ownership metadata.
	if err := artifact.location.AssertOwnedTable(artifact.parquetPath, artifact.tableName); err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	manifest, err := readConnectionStageManifest(artifact.manifestPath)
	if err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	manifestSHA256, _, err := digestPayloadFile(artifact.manifestPath)
	if err != nil {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("digest staged manifest: %w", err)
	}
	if manifestSHA256 != receipt.ManifestSHA256 {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q manifest identity does not match", receipt.ID)
	}
	if err := manifest.matchesReceipt(receipt); err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	walSHA256, _, err := digestPayloadFile(artifact.walPath)
	if err != nil {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("digest staged wal: %w", err)
	}
	if walSHA256 != manifest.WALSHA256 {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q wal content does not match manifest", receipt.ID)
	}
	parquetSHA256, _, err := digestPayloadFile(artifact.parquetPath)
	if err != nil {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("digest staged parquet: %w", err)
	}
	if parquetSHA256 != manifest.ParquetSHA256 {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q parquet content does not match manifest", receipt.ID)
	}
	checkpointSHA256, err := hashJSON(manifest.CandidateCheckpoint)
	if err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	tombstonesSHA256, err := hashJSON(manifest.Tombstones)
	if err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	if checkpointSHA256 != manifest.CheckpointSHA256 || tombstonesSHA256 != manifest.TombstonesSHA256 {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q manifest payload identities do not match", receipt.ID)
	}
	contentSHA256, err := manifest.contentSHA256()
	if err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	if contentSHA256 != manifest.ContentSHA256 || contentSHA256 != receipt.ContentSHA256 {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q content identity does not match", receipt.ID)
	}

	records := make([]connectors.Record, 0, manifest.Records)
	err = warehouse.ReadTable(ctx, artifact.parquetPath, func(row warehouse.Row) error {
		if len(records) >= manifest.Records {
			return fmt.Errorf("warehouse stage receipt %q parquet has more rows than its manifest", receipt.ID)
		}
		records = append(records, connectors.Record(row))
		return nil
	})
	if err != nil {
		return synctransport.WarehouseWorkset{}, err
	}
	if len(records) != manifest.Records {
		return synctransport.WarehouseWorkset{}, fmt.Errorf("warehouse stage receipt %q parquet rows = %d, want %d", receipt.ID, len(records), manifest.Records)
	}
	return synctransport.WarehouseWorkset{
		ID:                  manifest.ID,
		SourceParquet:       artifact.parquetPath,
		Records:             records,
		Tombstones:          cloneConnectionStageTombstones(manifest.Tombstones),
		CandidateCheckpoint: manifest.CandidateCheckpoint.Clone(),
	}, nil
}

// retire removes exactly one committed, connection-owned transient receipt.
// The receipt controls neither its directory nor its file names: artifactFor
// derives all three paths from the validated owner and opaque stage ID.
//
// Connection-owned receipts intentionally do not implement the optional eager
// RetirableWarehouseStage: their durable manifest and Parquet remain observable
// through ordinary Open for recovery and certification. The generic
// pre-execution reconciliation path invokes this private operation instead.
func (s *connectionWarehouseStage) retire(ctx context.Context, receipt synctransport.WarehouseReceipt) error {
	if s == nil || s.app == nil {
		return fmt.Errorf("connection-owned warehouse stage is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	conn, ok := s.connectionByID(receipt.Owner)
	if !ok {
		return fmt.Errorf("warehouse stage receipt %q has an unknown connection owner", receipt.ID)
	}
	artifact, err := s.artifactFor(conn, receipt.ID)
	if err != nil {
		return err
	}
	manifest, err := readConnectionStageManifest(artifact.manifestPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := manifest.matchesReceipt(receipt); err != nil {
		return err
	}
	return removeConnectionStageArtifact(ctx, artifact)
}

// ReconcileCommitted retires only receipts whose candidate checkpoint is
// already durably committed in this project. It is intentionally bounded per
// open and leaves malformed, foreign, active, or newer receipts untouched.
// Thus a process killed after checkpoint persistence cannot repeat a
// destination effect, while a process killed before persistence retains the
// workset for later investigation/recovery.
func (s *connectionWarehouseStage) ReconcileCommitted(ctx context.Context) error {
	if s == nil || s.app == nil {
		return fmt.Errorf("connection-owned warehouse stage is unavailable")
	}
	inspected := 0
	for _, conn := range s.app.state.Connections {
		if inspected >= connectionWarehouseStageReconcileLimit {
			return nil
		}
		location, err := s.app.warehouseLocation(filepath.Join(s.app.projectDir, "warehouse"), conn)
		if err != nil {
			return err
		}
		manifestDir := filepath.Join(location.ConnectionDir, connectionWarehouseStageManifestDir)
		entries, err := os.ReadDir(manifestDir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read staged receipt directory: %w", err)
		}
		for _, entry := range entries {
			if inspected >= connectionWarehouseStageReconcileLimit {
				return nil
			}
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			inspected++
			stageID := strings.TrimSuffix(entry.Name(), ".json")
			artifact, err := s.artifactFor(conn, stageID)
			if err != nil {
				continue
			}
			manifest, err := readConnectionStageManifest(artifact.manifestPath)
			if err != nil || !manifest.belongsToConnection(conn) {
				continue
			}
			streamState, present := s.app.state.StreamStates[streamStateKey(conn.Name, manifest.Stream)]
			if !present {
				continue
			}
			matches, err := committedStageCheckpointMatches(manifest.CandidateCheckpoint, streamState.Checkpoint)
			if err != nil {
				return fmt.Errorf("compare staged receipt %q with committed checkpoint: %w", manifest.ID, err)
			}
			if !matches {
				continue
			}
			manifestSHA256, _, err := digestPayloadFile(artifact.manifestPath)
			if err != nil {
				continue
			}
			if err := s.retire(ctx, manifest.receipt(manifestSHA256)); err != nil {
				return fmt.Errorf("retire reconciled warehouse stage receipt %q: %w", manifest.ID, err)
			}
		}
	}
	return nil
}

func (s *connectionWarehouseStage) connectionForStageRequest(request synctransport.WarehouseStageRequest) (Connection, error) {
	if strings.TrimSpace(request.ConnectionID) == "" {
		return Connection{}, fmt.Errorf("warehouse stage connection ID is required")
	}
	if request.Generation <= 0 {
		return Connection{}, fmt.Errorf("warehouse stage generation must be positive")
	}
	if err := request.Mode.Validate(); err != nil {
		return Connection{}, err
	}
	if _, err := warehouse.PathComponent("stream", request.Stream); err != nil {
		return Connection{}, err
	}
	conn, ok := s.connectionByID(request.ConnectionID)
	if !ok {
		return Connection{}, fmt.Errorf("warehouse stage connection %q was not found", request.ConnectionID)
	}
	if request.SourceName != conn.Source.Connector || request.DestinationName != conn.Destination.Connector {
		return Connection{}, fmt.Errorf("warehouse stage connection %q does not own %q to %q", conn.ID, request.SourceName, request.DestinationName)
	}
	if _, ok := conn.Streams[request.Stream]; !ok {
		return Connection{}, fmt.Errorf("warehouse stage connection %q does not declare stream %q", conn.ID, request.Stream)
	}
	return conn, nil
}

func (s *connectionWarehouseStage) connectionByID(id string) (Connection, bool) {
	for _, conn := range s.app.state.Connections {
		if conn.ID == id {
			return conn, true
		}
	}
	return Connection{}, false
}

func (s *connectionWarehouseStage) artifactFor(conn Connection, stageID string) (connectionWarehouseStageArtifact, error) {
	stagePart, err := warehouse.PathComponent("transport receipt", stageID)
	if err != nil {
		return connectionWarehouseStageArtifact{}, err
	}
	location, err := s.app.warehouseLocation(filepath.Join(s.app.projectDir, "warehouse"), conn)
	if err != nil {
		return connectionWarehouseStageArtifact{}, err
	}
	tableName := connectionWarehouseStagePrefix + stagePart
	parquetPath, err := location.TablePath(tableName)
	if err != nil {
		return connectionWarehouseStageArtifact{}, err
	}
	walPath, err := location.WALPath(tableName)
	if err != nil {
		return connectionWarehouseStageArtifact{}, err
	}
	return connectionWarehouseStageArtifact{
		location:     location,
		tableName:    tableName,
		walPath:      walPath,
		parquetPath:  parquetPath,
		manifestPath: filepath.Join(location.ConnectionDir, connectionWarehouseStageManifestDir, stagePart+".json"),
	}, nil
}

func (m connectionWarehouseStageManifest) belongsToConnection(conn Connection) bool {
	if m.Owner != conn.ID || m.SourceName != conn.Source.Connector || m.DestinationName != conn.Destination.Connector {
		return false
	}
	if _, ok := conn.Streams[m.Stream]; !ok {
		return false
	}
	return m.Generation > 0 && m.Records >= 0 && m.Mode.Validate() == nil
}

func committedStageCheckpointMatches(candidate synccontract.CheckpointEnvelope, committed *synccontract.CheckpointEnvelope) (bool, error) {
	if committed == nil || committed.CommittedAt == nil {
		return false, nil
	}
	candidateSHA256, err := connectionWarehouseStageRetryCheckpointSHA256(candidate)
	if err != nil {
		return false, err
	}
	committedSHA256, err := connectionWarehouseStageRetryCheckpointSHA256(*committed)
	if err != nil {
		return false, err
	}
	return candidateSHA256 == committedSHA256, nil
}

func removeConnectionStageArtifact(ctx context.Context, artifact connectionWarehouseStageArtifact) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Delete data before the manifest, so an interrupted cleanup leaves the
	// ownership proof available for the next bounded reconciliation attempt.
	for _, path := range []string{artifact.parquetPath, artifact.walPath, artifact.manifestPath} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove staged artifact: %w", err)
		}
	}
	seen := make(map[string]struct{}, 3)
	for _, dir := range []string{filepath.Dir(artifact.parquetPath), filepath.Dir(artifact.walPath), filepath.Dir(artifact.manifestPath)} {
		if _, duplicate := seen[dir]; duplicate {
			continue
		}
		seen[dir] = struct{}{}
		if err := syncLocalWarehouseDirectoryChain(dir); err != nil {
			return err
		}
	}
	return nil
}

func writeConnectionStageWAL(ctx context.Context, path, stageID string, generation int64, records []connectors.Record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create staged wal directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open staged wal: %w", err)
	}
	encoder := json.NewEncoder(file)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for index, record := range records {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return err
		}
		raw := localRawRecord{
			RawID:        fmt.Sprintf("%s:%012d", stageID, index+1),
			RunID:        stageID,
			SyncID:       stageID,
			GenerationID: generation,
			ExtractedAt:  now,
			LoadedAt:     now,
			Record:       record,
		}
		if err := encoder.Encode(raw); err != nil {
			_ = file.Close()
			return fmt.Errorf("write staged wal record: %w", err)
		}
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync staged wal: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close staged wal: %w", err)
	}
	return nil
}

func (m connectionWarehouseStageManifest) contentSHA256() (string, error) {
	return hashJSON(connectionWarehouseStageContent{
		ID:               m.ID,
		Owner:            m.Owner,
		Generation:       m.Generation,
		SourceName:       m.SourceName,
		DestinationName:  m.DestinationName,
		Stream:           m.Stream,
		Mode:             m.Mode,
		Records:          m.Records,
		Tombstones:       len(m.Tombstones),
		CheckpointSHA256: m.CheckpointSHA256,
		TombstonesSHA256: m.TombstonesSHA256,
		WALSHA256:        m.WALSHA256,
		ParquetSHA256:    m.ParquetSHA256,
	})
}

func (m connectionWarehouseStageManifest) receipt(manifestSHA256 string) synctransport.WarehouseReceipt {
	return synctransport.WarehouseReceipt{
		ID:               m.ID,
		Owner:            m.Owner,
		Generation:       m.Generation,
		Stream:           m.Stream,
		Mode:             m.Mode,
		CheckpointSHA256: m.CheckpointSHA256,
		TombstonesSHA256: m.TombstonesSHA256,
		ManifestSHA256:   manifestSHA256,
		ContentSHA256:    m.ContentSHA256,
		ParquetSHA256:    m.ParquetSHA256,
		Records:          m.Records,
		Tombstones:       len(m.Tombstones),
	}
}

func (m connectionWarehouseStageManifest) matchesReceipt(receipt synctransport.WarehouseReceipt) error {
	if m.Version != connectionWarehouseStageManifestVersion ||
		m.ID != receipt.ID ||
		m.Owner != receipt.Owner ||
		m.Generation != receipt.Generation ||
		m.Stream != receipt.Stream ||
		m.Mode != receipt.Mode ||
		m.CheckpointSHA256 != receipt.CheckpointSHA256 ||
		m.TombstonesSHA256 != receipt.TombstonesSHA256 ||
		m.ContentSHA256 != receipt.ContentSHA256 ||
		m.ParquetSHA256 != receipt.ParquetSHA256 ||
		m.Records != receipt.Records ||
		len(m.Tombstones) != receipt.Tombstones {
		return fmt.Errorf("warehouse stage receipt %q does not match manifest", receipt.ID)
	}
	if m.Generation <= 0 || m.Records < 0 || strings.TrimSpace(m.SourceName) == "" || strings.TrimSpace(m.DestinationName) == "" {
		return fmt.Errorf("warehouse stage receipt %q manifest is invalid", receipt.ID)
	}
	return nil
}

func writeConnectionStageManifest(path string, manifest connectionWarehouseStageManifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create staged manifest directory: %w", err)
	}
	temporary := path + ".tmp"
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open staged manifest: %w", err)
	}
	encoded, err := json.Marshal(manifest)
	if err == nil {
		_, err = file.Write(append(encoded, '\n'))
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("write staged manifest: %w", err)
	}
	if closeErr != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("close staged manifest: %w", closeErr)
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("publish staged manifest: %w", err)
	}
	return nil
}

func readConnectionStageManifest(path string) (connectionWarehouseStageManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return connectionWarehouseStageManifest{}, fmt.Errorf("open staged manifest: %w", err)
	}
	defer func() { _ = file.Close() }()
	decoder := json.NewDecoder(file)
	var manifest connectionWarehouseStageManifest
	if err := decoder.Decode(&manifest); err != nil {
		return connectionWarehouseStageManifest{}, fmt.Errorf("decode staged manifest: %w", err)
	}
	if decoder.More() {
		return connectionWarehouseStageManifest{}, fmt.Errorf("decode staged manifest: trailing values")
	}
	return manifest, nil
}

func cloneConnectionStageTombstones(values []synccontract.Tombstone) []synccontract.Tombstone {
	if values == nil {
		return nil
	}
	clone := make([]synccontract.Tombstone, len(values))
	for index, value := range values {
		clone[index] = value.Clone()
	}
	return clone
}
