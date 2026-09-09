package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

var _ synctransport.FullOverwriteDestination = (*localWarehouseDestinationExecutor)(nil)
var _ synctransport.EmptyPublicationReadBackDestination = (*localWarehouseDestinationExecutor)(nil)

// Keep the existing WAL-before-single-Parquet publication order. This seam
// exercises actual rename failures without replacing the warehouse executor.
var localWarehouseOverwriteRename = os.Rename

type localWarehouseOverwriteOutput struct {
	Version        int       `json:"version"`
	ConnectionID   string    `json:"connection_id"`
	StreamID       string    `json:"stream_id"`
	Stream         string    `json:"stream"`
	OwnerSHA256    string    `json:"owner_sha256"`
	Generation     int64     `json:"generation"`
	Rows           int       `json:"rows"`
	TableSHA256    string    `json:"table_sha256"`
	WALSHA256      string    `json:"wal_sha256"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
}

type localWarehouseOverwriteRun struct {
	executor               *localWarehouseDestinationExecutor
	request                synctransport.FullOverwriteRunRequest
	location               warehouse.Location
	stream                 StreamConfig
	table, wal, privateDir string
	file                   *os.File
	mu                     sync.Mutex
	pages, records         int
	receipts               map[string]struct{}
	closed, published      bool
	output                 localWarehouseOverwriteOutput
	ack                    synccontract.DownstreamAcknowledgement
}

func (e *localWarehouseDestinationExecutor) overwriteLocation(binding synctransport.DestinationBinding, streamName string, runtime connectors.RuntimeConfig, source connectors.Connector) (warehouse.Location, StreamConfig, error) {
	if e == nil || e.app == nil || source == nil {
		return warehouse.Location{}, StreamConfig{}, fmt.Errorf("local warehouse overwrite is unavailable")
	}
	conn, ok := e.app.findConnectionByID(binding.ConnectionID)
	stream, streamOK := conn.Streams[streamName]
	if !ok || !streamOK || !e.connectionOwnsLocalWarehouse(conn) || binding.WorkspaceID != e.app.state.WorkspaceID || binding.SourceConnectorID != conn.Source.Connector || source.Name() != conn.Source.Connector || binding.StreamID == "" || binding.StreamID != stream.StreamID || !slices.Equal(binding.PrimaryKey, stream.PrimaryKey) {
		return warehouse.Location{}, StreamConfig{}, fmt.Errorf("local warehouse overwrite ownership binding is invalid")
	}
	location, err := e.app.warehouseLocation(localWarehouseDir(runtime), conn)
	return location, stream, err
}

func (e *localWarehouseDestinationExecutor) BeginFullOverwrite(ctx context.Context, request synctransport.FullOverwriteRunRequest) (synctransport.FullOverwriteRun, error) {
	if ctx == nil {
		return nil, fmt.Errorf("local warehouse overwrite requires context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	strategy, err := localWarehouseApplyStrategy(synccontract.ModeFullOverwrite)
	if err != nil {
		return nil, err
	}
	if request.Mode != synccontract.ModeFullOverwrite || request.Plan.ApplyStrategy != strategy || request.Generation <= 0 || request.BatchSize <= 0 || request.ConnectionID != request.Binding.ConnectionID || request.TransformPlanJSON != "" || request.TransformPlanHash != "" || request.Plan.TransformPlanHash != "" {
		return nil, fmt.Errorf("local warehouse overwrite run binding is invalid")
	}
	loc, stream, err := e.overwriteLocation(request.Binding, request.Stream, request.Runtime, request.Source)
	if err != nil {
		return nil, err
	}
	if err := warehouse.CheckLegacyLayout(loc.Root); err != nil {
		return nil, err
	}
	if err := warehouse.CheckLegacyTableFormat(loc.Root); err != nil {
		return nil, err
	}
	if err := loc.EnsureOwnership(); err != nil {
		return nil, err
	}
	tableName := stream.DestinationTable
	if tableName == "" {
		tableName = request.Stream
	}
	table, err := loc.TablePath(tableName)
	if err != nil {
		return nil, err
	}
	wal, err := loc.WALPath(request.Stream)
	if err != nil {
		return nil, err
	}
	if err := loc.AssertOwnedTable(table, tableName); err != nil {
		return nil, err
	}
	for _, dir := range []string{filepath.Dir(wal), filepath.Dir(table)} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("create overwrite directory: %w", err)
		}
	}
	privateDir, err := os.MkdirTemp(filepath.Dir(wal), ".overwrite-")
	if err != nil {
		return nil, fmt.Errorf("prepare private overwrite: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(privateDir, "snapshot.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, errors.Join(err, os.Remove(privateDir))
	}
	request.Binding.PrimaryKey = slices.Clone(request.Binding.PrimaryKey)
	stream.PrimaryKey = slices.Clone(stream.PrimaryKey)
	return &localWarehouseOverwriteRun{executor: e, request: request, location: loc, stream: stream, table: table, wal: wal, privateDir: privateDir, file: file, receipts: make(map[string]struct{})}, nil
}

func (s *localWarehouseOverwriteRun) ApplyFullOverwrite(ctx context.Context, request synctransport.DestinationApplyRequest) error {
	if s == nil || ctx == nil {
		return fmt.Errorf("local warehouse overwrite page is invalid")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r := request.Receipt
	if s.closed || s.published || r.Validate() != nil || request.ConnectionID != s.request.ConnectionID || r.Owner != s.request.ConnectionID || r.Generation != s.request.Generation || r.Stream != s.request.Stream || r.Mode != s.request.Mode || request.Mode != s.request.Mode || request.Stream != s.request.Stream || request.Plan.ApplyStrategy != s.request.Plan.ApplyStrategy || request.Plan.TransformPlanHash != "" || request.Workset.ID != r.ID || r.Records != len(request.Workset.Records) || len(request.Workset.Records) > s.request.BatchSize || r.Tombstones != 0 || len(request.Workset.Tombstones) != 0 || request.Binding.ConnectionID != s.request.Binding.ConnectionID || request.Binding.WorkspaceID != s.request.Binding.WorkspaceID || request.Binding.SourceConnectorID != s.request.Binding.SourceConnectorID || request.Binding.StreamID != s.request.Binding.StreamID || !slices.Equal(request.Binding.PrimaryKey, s.request.Binding.PrimaryKey) || localWarehouseDir(request.Runtime) != s.location.Root {
		return fmt.Errorf("local warehouse overwrite page does not match its run")
	}
	if _, exists := s.receipts[r.ID]; exists {
		return fmt.Errorf("local warehouse overwrite page was already applied")
	}
	raw, err := localWarehouseTransportRawRecords(r, request.Workset, s.stream, s.request.Plan.ApplyStrategy)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(s.file)
	for _, record := range raw {
		if err := ctx.Err(); err != nil {
			s.closed = true
			return err
		}
		if err := encoder.Encode(record); err != nil {
			s.closed = true
			return fmt.Errorf("write private overwrite wal: %w", err)
		}
	}
	if err := s.file.Sync(); err != nil {
		s.closed = true
		return fmt.Errorf("sync private overwrite wal: %w", err)
	}
	s.receipts[r.ID] = struct{}{}
	s.pages++
	s.records += len(raw)
	return nil
}

func overwriteOwnerDigest(loc warehouse.Location) string {
	sum := sha256.Sum256([]byte(loc.ConnectionDir))
	return hex.EncodeToString(sum[:])
}

func (s *localWarehouseOverwriteRun) PublishFullOverwrite(ctx context.Context, request synctransport.FullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	empty := synccontract.DownstreamAcknowledgement{}
	if s == nil || ctx == nil {
		return empty, fmt.Errorf("local warehouse overwrite publication is invalid")
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.published || request.Pages != s.pages || request.Records != s.records || request.Tombstones != 0 {
		return empty, fmt.Errorf("local warehouse overwrite publication totals do not match prepared pages")
	}
	// Any failure after preparation closes this run; callers must abort/recover,
	// never append another page to a partially written or published snapshot.
	s.closed = true
	if err := s.file.Sync(); err != nil {
		return empty, err
	}
	closeErr := s.file.Close()
	s.file = nil
	if closeErr != nil {
		return empty, closeErr
	}
	privateWAL := filepath.Join(s.privateDir, "snapshot.jsonl")
	privateTable := filepath.Join(s.privateDir, "snapshot.parquet")
	rows, err := materializeFinalTable(ctx, privateWAL, privateTable, false, nil)
	if err != nil {
		return empty, err
	}
	if rows != s.records {
		return empty, fmt.Errorf("local warehouse overwrite materialization count mismatch")
	}
	if err := syncLocalWarehouseDirectoryChain(s.privateDir); err != nil {
		return empty, err
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	loc, stream, err := s.executor.overwriteLocation(s.request.Binding, s.request.Stream, s.request.Runtime, s.request.Source)
	if err != nil {
		return empty, err
	}
	tableName := stream.DestinationTable
	if tableName == "" {
		tableName = s.request.Stream
	}
	if loc.ConnectionDir != s.location.ConnectionDir {
		return empty, fmt.Errorf("local warehouse overwrite owner changed")
	}
	if err := loc.AssertOwnedTable(s.table, tableName); err != nil {
		return empty, err
	}
	walHash, _, err := digestPayloadFile(privateWAL)
	if err != nil {
		return empty, err
	}
	tableHash, _, err := digestPayloadFile(privateTable)
	if err != nil {
		return empty, err
	}
	if err := localWarehouseOverwriteRename(privateWAL, s.wal); err != nil {
		return empty, fmt.Errorf("publish overwrite wal: %w", err)
	}
	// A later failure retains this complete WAL as the source of truth, matching
	// the existing warehouse rebuild contract. There is no atomic two-file claim.
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(s.wal)); err != nil {
		return empty, err
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	if err := localWarehouseOverwriteRename(privateTable, s.table); err != nil {
		return empty, fmt.Errorf("publish overwrite table: %w", err)
	}
	if err := syncLocalWarehouseDirectoryChain(filepath.Dir(s.table)); err != nil {
		return empty, err
	}
	now := time.Now().UTC()
	output := localWarehouseOverwriteOutput{Version: 1, ConnectionID: s.request.ConnectionID, StreamID: s.request.Binding.StreamID, Stream: s.request.Stream, OwnerSHA256: overwriteOwnerDigest(loc), Generation: s.request.Generation, Rows: rows, TableSHA256: tableHash, WALSHA256: walHash, AcknowledgedAt: now}
	payload, err := json.Marshal(output)
	if err != nil {
		return empty, err
	}
	ack, err := synccontract.NewDurableDownstreamAcknowledgement("warehouse", now)
	if err != nil {
		return empty, err
	}
	ack, err = ack.WithOutput(payload)
	if err != nil {
		return empty, err
	}
	s.output = output
	s.ack = ack
	s.ack.Output = append(json.RawMessage(nil), ack.Output...)
	s.published = true
	if err := os.Remove(s.privateDir); err != nil {
		return empty, fmt.Errorf("retire empty overwrite preparation: %w", err)
	}
	return ack, nil
}

func readBackLocalWarehouseOverwrite(ctx context.Context, table, wal string, output localWarehouseOverwriteOutput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, target := range []struct{ path, want string }{{table, output.TableSHA256}, {wal, output.WALSHA256}} {
		digest, _, err := digestPayloadFile(target.path)
		if err != nil {
			return err
		}
		if digest != target.want {
			return fmt.Errorf("local warehouse overwrite changed after publication")
		}
	}
	count := 0
	if err := warehouse.ReadTable(ctx, table, func(warehouse.Row) error { count++; return nil }); err != nil {
		return err
	}
	if count != output.Rows {
		return fmt.Errorf("local warehouse overwrite read-back count mismatch")
	}
	return nil
}

func (s *localWarehouseOverwriteRun) ReadBackFullOverwrite(ctx context.Context, ack synccontract.DownstreamAcknowledgement) error {
	if s == nil || ctx == nil {
		return fmt.Errorf("local warehouse overwrite read-back is invalid")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := ack.PublicationWitness(); err != nil {
		return err
	}
	if !s.published || ack.Sink != s.ack.Sink || !ack.AcknowledgedAt.Equal(s.ack.AcknowledgedAt) || !bytes.Equal(ack.Output, s.ack.Output) {
		return fmt.Errorf("local warehouse overwrite acknowledgement mismatch")
	}
	return readBackLocalWarehouseOverwrite(ctx, s.table, s.wal, s.output)
}

func (s *localWarehouseOverwriteRun) AbortFullOverwrite(_ context.Context) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.published {
		return nil
	}
	s.closed = true
	var err error
	if s.file != nil {
		err = s.file.Close()
		s.file = nil
	}
	return errors.Join(err, os.RemoveAll(s.privateDir))
}

func (e *localWarehouseDestinationExecutor) ReadBackEmptyFullOverwrite(ctx context.Context, request synctransport.EmptyPublicationReadBackRequest) error {
	if ctx == nil {
		return fmt.Errorf("local warehouse empty read-back requires context")
	}
	if err := request.Receipt.Validate(); err != nil {
		return err
	}
	if len(request.Receipt.Output) == 0 || len(request.Receipt.Output) > 16384 || request.TransformPlanJSON != "" || request.TransformPlanHash != "" {
		return fmt.Errorf("local warehouse empty receipt is invalid")
	}
	var output localWarehouseOverwriteOutput
	decoder := json.NewDecoder(bytes.NewReader(request.Receipt.Output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&output); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("local warehouse empty receipt has trailing data")
	}
	loc, stream, err := e.overwriteLocation(request.Binding, request.Stream, request.Runtime, request.Source)
	if err != nil {
		return err
	}
	if output.Version != 1 || output.Rows != 0 || output.Generation <= 0 || output.ConnectionID != request.Binding.ConnectionID || output.StreamID != request.Binding.StreamID || output.Stream != request.Stream || output.OwnerSHA256 != overwriteOwnerDigest(loc) || request.Receipt.Witness.Sink != "warehouse" || !output.AcknowledgedAt.Equal(request.Receipt.Witness.AcknowledgedAt) {
		return fmt.Errorf("local warehouse empty receipt binding mismatch")
	}
	for _, value := range []string{output.TableSHA256, output.WALSHA256} {
		if data, err := hex.DecodeString(value); err != nil || len(data) != sha256.Size {
			return fmt.Errorf("local warehouse empty receipt digest is invalid")
		}
	}
	tableName := stream.DestinationTable
	if tableName == "" {
		tableName = request.Stream
	}
	table, err := loc.TablePath(tableName)
	if err != nil {
		return err
	}
	if err := loc.AssertOwnedTable(table, tableName); err != nil {
		return err
	}
	wal, err := loc.WALPath(request.Stream)
	if err != nil {
		return err
	}
	return readBackLocalWarehouseOverwrite(ctx, table, wal, output)
}
