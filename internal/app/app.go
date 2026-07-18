package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/events"
	pmlogging "polymetrics.ai/internal/logging"
	"polymetrics.ai/internal/safety"
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/telemetry"
	"polymetrics.ai/internal/vault"
)

const reversePlanModeConnectorCommand = "connector_command"

type App struct {
	root       string
	projectDir string
	statePath  string
	store      statestore.JSONStore[state]
	state      state
	vault      *vault.Vault
	registry   *connectors.Registry
	sqlEngine  sqlQueryEngine
}

// sqlQueryEngine is the pluggable backend for App.QuerySQL. The default build
// uses a JSONL engine that reproduces the historical SELECT * behavior; the
// duckdb-tagged build swaps in an analytical DuckDB engine.
type sqlQueryEngine interface {
	QuerySQL(ctx context.Context, sql string, limit int) ([]connectors.Record, error)
	Name() string
}

type state struct {
	Credentials  []CredentialMeta             `json:"credentials"`
	Connections  []Connection                 `json:"connections"`
	Catalogs     []CatalogSnapshot            `json:"catalogs"`
	Runs         []Run                        `json:"runs"`
	ReversePlans []ReversePlan                `json:"reverse_plans"`
	ReverseRuns  []ReverseRun                 `json:"reverse_runs"`
	Checkpoints  map[string]map[string]string `json:"checkpoints,omitempty"`
	StreamStates map[string]StreamState       `json:"stream_states,omitempty"`
}

func InitProject(root string) error {
	if root == "" {
		root = "."
	}
	projectDir := filepath.Join(root, ".polymetrics")
	for _, dir := range []string{
		projectDir,
		filepath.Join(projectDir, "state"),
		filepath.Join(projectDir, "warehouse"),
		filepath.Join(projectDir, "outbox"),
		filepath.Join(projectDir, "logs"),
	} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	if _, err := vault.Init(projectDir); err != nil {
		return err
	}
	configPath := filepath.Join(projectDir, "config.yaml")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		config := "version: 1\nproject: polymetrics-local\nwarehouse:\n  connector: warehouse\n  path: .polymetrics/warehouse\n"
		if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}
	statePath := filepath.Join(projectDir, "state", "state.json")
	if _, err := os.Stat(statePath); errors.Is(err, os.ErrNotExist) {
		initial := state{Checkpoints: map[string]map[string]string{}, StreamStates: map[string]StreamState{}}
		if err := writeJSONAtomic(statePath, initial); err != nil {
			return err
		}
	}
	return nil
}

func Open(root string) (*App, error) {
	if root == "" {
		root = "."
	}
	projectDir := filepath.Join(root, ".polymetrics")
	info, err := os.Stat(projectDir)
	if err != nil {
		return nil, fmt.Errorf("open project at %s: %w", projectDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", projectDir)
	}
	v, err := vault.Open(projectDir)
	if err != nil {
		return nil, err
	}
	statePath := filepath.Join(projectDir, "state", "state.json")
	a := &App{
		root:       root,
		projectDir: projectDir,
		statePath:  statePath,
		store:      newStateStore(statePath),
		vault:      v,
		registry:   bundleregistry.New(),
	}
	a.sqlEngine = newSQLEngine(a)
	if err := a.load(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) ProjectDir() string { return a.projectDir }

func (a *App) projectRoot() string { return filepath.Dir(a.projectDir) }

func (a *App) Registry() *connectors.Registry { return a.registry }

func (a *App) Connectors() []connectors.Metadata {
	return a.registry.List()
}

func (a *App) Connector(name string) (connectors.Metadata, error) {
	if err := connectors.RejectLegacyConnectorName(name); err != nil {
		return connectors.Metadata{}, err
	}
	c, ok := a.registry.Get(name)
	if !ok {
		return connectors.Metadata{}, fmt.Errorf("connector %q not found", name)
	}
	return c.Metadata(), nil
}

func (a *App) load() error {
	loaded, err := a.store.Load()
	if err != nil {
		return err
	}
	a.state = loaded
	if a.state.Checkpoints == nil {
		a.state.Checkpoints = map[string]map[string]string{}
	}
	if a.state.StreamStates == nil {
		a.state.StreamStates = map[string]StreamState{}
	}
	return nil
}

func (a *App) save() error {
	return a.store.Save(a.state)
}

func newStateStore(path string) statestore.JSONStore[state] {
	return statestore.JSONStore[state]{
		Path: path,
		Initial: func() state {
			return state{Checkpoints: map[string]map[string]string{}, StreamStates: map[string]StreamState{}}
		},
		Locker: statestore.FileLock{Path: path + ".lock"},
	}
}

func (a *App) AddCredential(ctx context.Context, req AddCredentialRequest) (CredentialMeta, error) {
	if strings.TrimSpace(req.Name) == "" {
		return CredentialMeta{}, errors.New("credential name is required")
	}
	if err := connectors.RejectLegacyConnectorName(req.Connector); err != nil {
		return CredentialMeta{}, err
	}
	if _, ok := a.registry.Get(req.Connector); !ok {
		return CredentialMeta{}, fmt.Errorf("connector %q not found", req.Connector)
	}
	if _, ok := a.findCredential(req.Name); ok {
		return CredentialMeta{}, fmt.Errorf("credential %q already exists", req.Name)
	}
	id, err := prefixedID("cred")
	if err != nil {
		return CredentialMeta{}, err
	}
	if req.Config == nil {
		req.Config = map[string]string{}
	}
	if req.Secrets == nil {
		req.Secrets = map[string]string{}
	}
	if err := a.validateCredentialConfig(req.Connector, req.Config); err != nil {
		return CredentialMeta{}, err
	}
	if err := a.vault.Put(ctx, id, req.Secrets); err != nil {
		return CredentialMeta{}, err
	}
	now := time.Now().UTC()
	fields := make([]string, 0, len(req.Secrets))
	for k := range req.Secrets {
		fields = append(fields, k)
		pmlogging.RegisterSensitiveKey(k)
	}
	sort.Strings(fields)
	meta := CredentialMeta{
		ID:           id,
		Name:         req.Name,
		Connector:    req.Connector,
		Config:       cloneStringMap(req.Config),
		SecretFields: fields,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	a.state.Credentials = append(a.state.Credentials, meta)
	if err := a.save(); err != nil {
		return CredentialMeta{}, err
	}
	return meta, nil
}

func (a *App) ListCredentials() []CredentialMeta {
	out := append([]CredentialMeta(nil), a.state.Credentials...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (a *App) validateCredentialConfig(connector string, config map[string]string) error {
	path := config["path"]
	if path == "" {
		return nil
	}
	switch connector {
	case "warehouse", "outbox":
		allowExternal := strings.EqualFold(config["allow_external_path"], "true")
		return safety.ValidateLocalWritePath(a.projectRoot(), path, connector+" path", allowExternal)
	default:
		return safety.RejectDangerousChars(path, connector+" path")
	}
}

func (a *App) normalizeLocalWriteRuntime(connector string, config map[string]string) (*connectors.LocalWritePolicy, error) {
	if connector != "warehouse" && connector != "outbox" {
		return nil, nil
	}
	allowExternal := strings.EqualFold(config["allow_external_path"], "true")
	projectRoot, err := filepath.Abs(a.projectRoot())
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	if path := config["path"]; path != "" {
		normalized, err := safety.ResolveLocalWritePath(projectRoot, path, connector+" path", allowExternal)
		if err != nil {
			return nil, err
		}
		config["path"] = normalized
	}
	return &connectors.LocalWritePolicy{
		ProjectRoot:   projectRoot,
		AllowExternal: allowExternal,
	}, nil
}

func validateLocalWriteRuntimeEffect(runtime connectors.RuntimeConfig, path, field string) error {
	if runtime.LocalWritePolicy == nil {
		return nil
	}
	return safety.ValidateLocalWritePath(
		runtime.LocalWritePolicy.ProjectRoot,
		path,
		field,
		runtime.LocalWritePolicy.AllowExternal,
	)
}

func (a *App) InspectCredential(name string) (CredentialMeta, error) {
	cred, ok := a.findCredential(name)
	if !ok {
		return CredentialMeta{}, fmt.Errorf("credential %q not found", name)
	}
	return cred, nil
}

func (a *App) TestCredential(ctx context.Context, name string) (CredentialMeta, error) {
	cred, runtime, err := a.resolveCredential(ctx, name, nil)
	if err != nil {
		return CredentialMeta{}, err
	}
	connector, ok := a.registry.Get(cred.Connector)
	if !ok {
		return CredentialMeta{}, fmt.Errorf("connector %q not found", cred.Connector)
	}
	if err := connector.Check(ctx, runtime); err != nil {
		return CredentialMeta{}, err
	}
	for i := range a.state.Credentials {
		if a.state.Credentials[i].Name == name {
			a.state.Credentials[i].LastValidatedAt = time.Now().UTC()
			cred = a.state.Credentials[i]
			break
		}
	}
	if err := a.save(); err != nil {
		return CredentialMeta{}, err
	}
	return cred, nil
}

func (a *App) RemoveCredential(ctx context.Context, name string) error {
	for i, cred := range a.state.Credentials {
		if cred.Name == name {
			if err := a.vault.Delete(ctx, cred.ID); err != nil {
				return err
			}
			a.state.Credentials = append(a.state.Credentials[:i], a.state.Credentials[i+1:]...)
			return a.save()
		}
	}
	return fmt.Errorf("credential %q not found", name)
}

func (a *App) CreateConnection(ctx context.Context, req CreateConnectionRequest) (Connection, error) {
	if strings.TrimSpace(req.Name) == "" {
		return Connection{}, errors.New("connection name is required")
	}
	if _, ok := a.findConnection(req.Name); ok {
		return Connection{}, fmt.Errorf("connection %q already exists", req.Name)
	}
	if req.Streams == nil || len(req.Streams) == 0 {
		return Connection{}, errors.New("at least one stream is required")
	}
	source, sourceRuntime, err := a.resolveEndpoint(ctx, req.Source)
	if err != nil {
		return Connection{}, fmt.Errorf("resolve source: %w", err)
	}
	if _, _, err := a.resolveEndpoint(ctx, req.Destination); err != nil {
		return Connection{}, fmt.Errorf("resolve destination: %w", err)
	}
	catalog, catalogErr := source.Catalog(ctx, sourceRuntime)
	for name, stream := range req.Streams {
		if stream.SyncMode == "" {
			stream.SyncMode = DefaultUserFacingSyncMode
		}
		mode, err := ParseSyncMode(stream.SyncMode)
		if err != nil {
			return Connection{}, err
		}
		stream.SyncMode = mode.Name
		if catalogErr == nil {
			if sourceStream, ok := findCatalogStream(catalog, name); ok {
				if stream.CursorField == "" && len(sourceStream.CursorFields) > 0 {
					stream.CursorField = sourceStream.CursorFields[0]
				}
				if len(stream.PrimaryKey) == 0 && len(sourceStream.PrimaryKey) > 0 {
					stream.PrimaryKey = append([]string(nil), sourceStream.PrimaryKey...)
				}
			}
		}
		if stream.DestinationTable == "" {
			stream.DestinationTable = name
		}
		if err := ValidateStreamSyncConfig(stream); err != nil {
			return Connection{}, fmt.Errorf("validate stream %q: %w", name, err)
		}
		req.Streams[name] = stream
	}
	now := time.Now().UTC()
	conn := Connection{
		Name:        req.Name,
		Source:      req.Source,
		Destination: req.Destination,
		Streams:     req.Streams,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	a.state.Connections = append(a.state.Connections, conn)
	if err := a.save(); err != nil {
		return Connection{}, err
	}
	return conn, nil
}

func (a *App) ListConnections() []Connection {
	out := append([]Connection(nil), a.state.Connections...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (a *App) RefreshCatalog(ctx context.Context, connectionName string) (CatalogSnapshot, error) {
	conn, ok := a.findConnection(connectionName)
	if !ok {
		return CatalogSnapshot{}, fmt.Errorf("connection %q not found", connectionName)
	}
	source, runtime, err := a.resolveEndpoint(ctx, conn.Source)
	if err != nil {
		return CatalogSnapshot{}, err
	}
	catalog, err := source.Catalog(ctx, runtime)
	if err != nil {
		return CatalogSnapshot{}, err
	}
	snapshot := CatalogSnapshot{Connection: conn.Name, Catalog: catalog, UpdatedAt: time.Now().UTC()}
	replaced := false
	for i := range a.state.Catalogs {
		if a.state.Catalogs[i].Connection == conn.Name {
			a.state.Catalogs[i] = snapshot
			replaced = true
			break
		}
	}
	if !replaced {
		a.state.Catalogs = append(a.state.Catalogs, snapshot)
	}
	if err := a.save(); err != nil {
		return CatalogSnapshot{}, err
	}
	return snapshot, nil
}

func (a *App) ShowCatalog(ctx context.Context, connectionName string) (CatalogSnapshot, error) {
	for _, snapshot := range a.state.Catalogs {
		if snapshot.Connection == connectionName {
			return snapshot, nil
		}
	}
	return a.RefreshCatalog(ctx, connectionName)
}

func (a *App) RunETL(ctx context.Context, req RunETLRequest) (run Run, err error) {
	started := time.Now()
	ctx, span := telemetry.StartSpan(ctx, "pm.etl.run",
		telemetry.StringAttr("pm.etl.connection", req.Connection),
		telemetry.StringAttr("pm.etl.stream", req.Stream),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(telemetry.StringAttr("pm.etl.status", "failed"))
		} else {
			span.SetAttributes(
				telemetry.StringAttr("pm.etl.status", "ok"),
				telemetry.IntAttr("pm.etl.records_read", run.RecordsRead),
				telemetry.IntAttr("pm.etl.records_loaded", run.RecordsLoaded),
			)
		}
		span.End()
		telemetry.RecordStageDuration(ctx, "etl", time.Since(started))
	}()

	conn, ok := a.findConnection(req.Connection)
	if !ok {
		return Run{}, fmt.Errorf("connection %q not found", req.Connection)
	}
	stream, ok := conn.Streams[req.Stream]
	if !ok {
		return Run{}, fmt.Errorf("stream %q not configured on connection %q", req.Stream, req.Connection)
	}
	runID, err := prefixedID("run")
	if err != nil {
		return Run{}, err
	}
	ctx = pmlogging.WithRunID(ctx, runID)
	run = Run{ID: runID, Type: "etl", Connection: req.Connection, Stream: req.Stream, Status: "running", StartedAt: time.Now().UTC()}
	a.state.Runs = append(a.state.Runs, run)
	_ = a.save()
	pmlogging.FromContext(ctx).InfoContext(ctx, "etl run started", "run_id", runID, "connection", req.Connection, "stream", req.Stream)
	a.emitETLEvent(ctx, events.KindStarted, runID, req.Stream, "running", etlExecutionResult{}, "")

	source, sourceRuntime, err := a.resolveEndpoint(ctx, conn.Source)
	if err != nil {
		return a.failRun(ctx, runID, err)
	}
	destination, destRuntime, err := a.resolveEndpoint(ctx, conn.Destination)
	if err != nil {
		return a.failRun(ctx, runID, err)
	}
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	mode, err := ParseSyncMode(stream.SyncMode)
	if err != nil {
		return a.failRun(ctx, runID, err)
	}
	stream.SyncMode = mode.Name
	if err := ValidateStreamSyncConfig(stream); err != nil {
		return a.failRun(ctx, runID, err)
	}
	var result etlExecutionResult
	if materializer, ok := destination.(connectors.LocalWarehouseMaterializer); ok && materializer.MaterializesLocalWarehouse() {
		result, err = a.runWarehouseETL(ctx, runID, conn, source, sourceRuntime, destRuntime, req.Stream, stream, mode, batchSize)
	} else {
		result, err = a.runConnectorETL(ctx, runID, conn, source, sourceRuntime, destination, destRuntime, req.Stream, stream, mode, batchSize)
	}
	if err != nil {
		return a.failRun(ctx, runID, err)
	}
	return a.completeRun(ctx, runID, result)
}

func (a *App) runConnectorETL(ctx context.Context, runID string, conn Connection, source connectors.Connector, sourceRuntime connectors.RuntimeConfig, destination connectors.Connector, destRuntime connectors.RuntimeConfig, streamName string, stream StreamConfig, mode SyncMode, batchSize int) (etlExecutionResult, error) {
	if mode.IsDeduped() {
		return etlExecutionResult{}, fmt.Errorf("sync mode %s requires the local warehouse destination in this dependency-free implementation", mode.Name)
	}
	if a.state.StreamStates == nil {
		a.state.StreamStates = map[string]StreamState{}
	}
	stateKey := streamStateKey(conn.Name, streamName)
	prior := a.state.StreamStates[stateKey]
	generationID := prior.GenerationID
	if generationID == 0 || mode.IsOverwrite() {
		generationID++
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result := etlExecutionResult{}
	metrics := telemetry.NewRunCounters(ctx)
	batch := make([]connectors.Record, 0, batchSize)
	firstWrite := true
	nextCursor := prior.Cursor

	flush := func(force bool) error {
		if len(batch) == 0 {
			if force && mode.IsOverwrite() && firstWrite {
				_, err := destination.Write(ctx, connectors.WriteRequest{
					Stream:     streamName,
					Table:      stream.DestinationTable,
					Action:     "upsert",
					Overwrite:  true,
					Config:     destRuntime,
					PrimaryKey: stream.PrimaryKey,
				}, nil)
				firstWrite = false
				return err
			}
			return nil
		}
		metrics.RecordBatchCreated()
		metrics.Flush(ctx)
		writeResult, err := destination.Write(ctx, connectors.WriteRequest{
			Stream:     streamName,
			Table:      stream.DestinationTable,
			Action:     "upsert",
			Overwrite:  mode.IsOverwrite() && firstWrite,
			Config:     destRuntime,
			PrimaryKey: stream.PrimaryKey,
		}, batch)
		firstWrite = false
		if err != nil {
			return err
		}
		result.RecordsLoaded += writeResult.RecordsWritten
		metrics.RecordLoaded(writeResult.RecordsWritten)
		result.RecordsFailed += writeResult.RecordsFailed
		metrics.RecordFailed(writeResult.RecordsFailed)
		result.BatchCount++
		metrics.RecordBatchFlushed()
		metrics.Flush(ctx)
		a.emitETLEvent(ctx, events.KindProgress, runID, streamName, "batch", result, "")
		batch = batch[:0]
		return nil
	}

	readConfig := sourceRuntime
	readConfig.Config = cloneStringMap(sourceRuntime.Config)
	if prior.Cursor != "" {
		readConfig.Config["since"] = prior.Cursor
	}
	err := source.Read(ctx, connectors.ReadRequest{
		Stream: streamName,
		Config: readConfig,
		State:  map[string]string{"cursor": prior.Cursor, "generation_id": strconv.FormatInt(generationID, 10)},
	}, func(record connectors.Record) error {
		result.RecordsRead++
		metrics.RecordRead()
		cursor := ""
		if stream.CursorField != "" {
			var err error
			cursor, err = recordCursor(record, stream.CursorField)
			if err != nil {
				return err
			}
			if mode.Source == SourceSyncIncremental && prior.Cursor != "" && compareCursor(cursor, prior.Cursor) < 0 {
				return nil
			}
			if nextCursor == "" || compareCursor(cursor, nextCursor) > 0 {
				nextCursor = cursor
			}
		}
		r := cloneRecord(record)
		r["_polymetrics_run_id"] = runID
		r["_polymetrics_synced_at"] = now
		r["_polymetrics_deleted"] = isDeletedRecord(record)
		if cursor != "" {
			r["_polymetrics_cursor"] = cursor
		}
		result.RecordsTransformed++
		metrics.RecordTransformed()
		batch = append(batch, r)
		if len(batch) >= batchSize {
			return flush(false)
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if err := flush(true); err != nil {
		return result, err
	}
	updated := StreamState{
		Connection:          conn.Name,
		Stream:              streamName,
		Cursor:              nextCursor,
		GenerationID:        generationID,
		LastSuccessfulRunID: runID,
		RecordsLoaded:       result.RecordsLoaded,
		UpdatedAt:           time.Now().UTC(),
	}
	a.state.StreamStates[stateKey] = updated
	result.Checkpoint = checkpointForResult(result, mode, stateKey, updated)
	return result, nil
}

func (a *App) completeRun(ctx context.Context, runID string, result etlExecutionResult) (Run, error) {
	run := Run{}
	for i := range a.state.Runs {
		if a.state.Runs[i].ID == runID {
			a.state.Runs[i].Status = "completed"
			a.state.Runs[i].RecordsRead = result.RecordsRead
			a.state.Runs[i].RecordsTransformed = result.RecordsTransformed
			a.state.Runs[i].RecordsLoaded = result.RecordsLoaded
			a.state.Runs[i].RecordsFailed = result.RecordsFailed
			a.state.Runs[i].BatchCount = result.BatchCount
			a.state.Runs[i].Checkpoint = result.Checkpoint
			a.state.Runs[i].CompletedAt = time.Now().UTC()
			run = a.state.Runs[i]
			break
		}
	}
	if a.state.Checkpoints == nil {
		a.state.Checkpoints = map[string]map[string]string{}
	}
	a.state.Checkpoints[runID] = cloneStringMap(result.Checkpoint)
	if err := a.save(); err != nil {
		return Run{}, err
	}
	pmlogging.FromContext(ctx).InfoContext(ctx, "etl run completed", "run_id", runID, "stream", run.Stream, "records_read", result.RecordsRead, "records_loaded", result.RecordsLoaded, "records_failed", result.RecordsFailed, "batches", result.BatchCount)
	a.emitETLEvent(ctx, events.KindCompleted, runID, run.Stream, "success", result, "")
	return run, nil
}

func (a *App) GetRun(id string) (Run, error) {
	for _, run := range a.state.Runs {
		if run.ID == id {
			return run, nil
		}
	}
	return Run{}, fmt.Errorf("run %q not found", id)
}

func (a *App) QueryTable(ctx context.Context, req QueryTableRequest) ([]connectors.Record, error) {
	if req.Table == "" {
		return nil, errors.New("table is required")
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	cfg := connectors.RuntimeConfig{
		ProjectDir: a.projectDir,
		Config: map[string]string{
			"path": filepath.Join(a.projectDir, "warehouse"),
		},
	}
	warehouse, ok := a.registry.Get("warehouse")
	if !ok {
		return nil, errors.New("warehouse connector not registered")
	}
	rows := make([]connectors.Record, 0)
	err := warehouse.Read(ctx, connectors.ReadRequest{Stream: req.Table, Config: cfg, Limit: req.Limit}, connectors.LimitEmitter(req.Limit, func(record connectors.Record) error {
		rows = append(rows, record)
		return nil
	}))
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return nil, err
	}
	return rows, nil
}

func (a *App) QuerySQL(ctx context.Context, sql string, limit int) ([]connectors.Record, error) {
	return a.sqlEngine.QuerySQL(ctx, sql, limit)
}

// QueryEngineName reports which SQL engine backs QuerySQL ("jsonl" by default,
// "duckdb" when built with -tags duckdb).
func (a *App) QueryEngineName() string {
	return a.sqlEngine.Name()
}

func (a *App) PlanReverseETL(ctx context.Context, req PlanReverseETLRequest) (ReversePlan, error) {
	if req.Name == "" {
		return ReversePlan{}, errors.New("reverse plan name is required")
	}
	if req.Action == "" {
		req.Action = "upsert"
	}
	if len(req.Mappings) == 0 {
		return ReversePlan{}, errors.New("at least one field mapping is required")
	}
	if req.Limit <= 0 {
		req.Limit = 100000
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: req.SourceTable, Limit: req.Limit})
	if err != nil {
		return ReversePlan{}, err
	}
	mapped := mapReverseRecords(records, req.Mappings, "")
	dest := EndpointConfig{Connector: req.DestinationConnector, Credential: req.DestinationCredential, Config: req.DestinationConfig}
	destination, runtime, err := a.resolveEndpoint(ctx, dest)
	if err != nil {
		return ReversePlan{}, fmt.Errorf("resolve reverse destination: %w", err)
	}
	if !destination.Metadata().Capabilities.Write {
		return ReversePlan{}, fmt.Errorf("connector %q does not support reverse ETL writes", destination.Name())
	}
	if validator, ok := destination.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, connectors.WriteRequest{
			Stream: "records",
			Table:  req.Name,
			Action: req.Action,
			Config: runtime,
		}, mapped); err != nil {
			return ReversePlan{}, fmt.Errorf("validate reverse destination: %w", err)
		}
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, err
	}
	token, err := randomToken(18)
	if err != nil {
		return ReversePlan{}, err
	}
	created := time.Now().UTC()
	planHash, err := reversePlanHash(req.Name, req.SourceTable, req.DestinationConnector, req.DestinationCredential, req.Action, req.DestinationConfig, req.Mappings, mapped)
	if err != nil {
		return ReversePlan{}, err
	}
	sampleCount := min(3, len(mapped))
	plan := ReversePlan{
		ID:                    id,
		Name:                  req.Name,
		Status:                "planned",
		SourceTable:           req.SourceTable,
		DestinationConnector:  req.DestinationConnector,
		DestinationCredential: req.DestinationCredential,
		DestinationConfig:     cloneStringMap(req.DestinationConfig),
		Action:                req.Action,
		Mappings:              cloneStringMap(req.Mappings),
		ConfirmationChallenge: a.confirmationChallengeForAction(req.DestinationConnector, req.Action),
		RecordCount:           len(records),
		Sample:                cloneRecords(mapped[:sampleCount]),
		PlanHash:              planHash,
		ApprovalTokenHash:     hashString(token),
		ApprovalToken:         token,
		CreatedAt:             created,
		ExpiresAt:             created.Add(24 * time.Hour),
	}
	stored := plan
	stored.ApprovalToken = ""
	a.state.ReversePlans = append(a.state.ReversePlans, stored)
	if err := a.save(); err != nil {
		return ReversePlan{}, err
	}
	return plan, nil
}

func (a *App) PlanConnectorCommand(ctx context.Context, req PlanConnectorCommandRequest) (ReversePlan, *connectors.WritePreview, error) {
	if err := connectors.RejectLegacyConnectorName(req.Connector); err != nil {
		return ReversePlan{}, nil, err
	}
	connector, runtime, err := a.ResolveConnectorCredential(ctx, req.Connector, req.Credential, req.Config)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	writeCommand, err := commandrunner.BuildWriteCommand(ctx, connector, commandrunner.Request{
		Path:    req.Path,
		Flags:   req.Flags,
		Config:  runtime,
		Preview: req.Preview,
	})
	if err != nil {
		return ReversePlan{}, nil, err
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, nil, err
	}
	token, err := randomToken(18)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.ReplaceAll(writeCommand.Command, " ", "_")
	}
	planHash, err := connectorCommandPlanHash(name, req.Connector, req.Credential, req.Config, writeCommand.Command, req.Path, writeCommand.Write, writeCommand.Record)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	created := time.Now().UTC()
	plan := ReversePlan{
		ID:                     id,
		Name:                   name,
		Status:                 "planned",
		Mode:                   reversePlanModeConnectorCommand,
		DestinationConnector:   req.Connector,
		DestinationCredential:  req.Credential,
		DestinationConfig:      cloneStringMap(req.Config),
		Action:                 writeCommand.Write,
		Mappings:               map[string]string{},
		ConnectorCommand:       writeCommand.Command,
		ConnectorCommandPath:   append([]string(nil), req.Path...),
		ConnectorCommandRecord: cloneRecord(writeCommand.Record),
		ConfirmationChallenge:  writeCommand.ConfirmationChallenge,
		RecordCount:            1,
		Sample:                 []connectors.Record{cloneRecord(writeCommand.RedactedRecord)},
		PlanHash:               planHash,
		ApprovalTokenHash:      hashString(token),
		ApprovalToken:          token,
		CreatedAt:              created,
		ExpiresAt:              created.Add(24 * time.Hour),
	}
	stored := plan
	stored.ApprovalToken = ""
	a.state.ReversePlans = append(a.state.ReversePlans, stored)
	if err := a.save(); err != nil {
		return ReversePlan{}, nil, err
	}
	return plan, writeCommand.Preview, nil
}

func (a *App) PreviewConnectorCommandPlan(ctx context.Context, id string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(id)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if plan.Mode != reversePlanModeConnectorCommand {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("reverse plan %q is not a connector command plan", id)
	}
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, connectors.WriteRequest{Action: plan.Action, Config: runtime}, []connectors.Record{plan.ConnectorCommandRecord}); err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("connector %q does not support reverse ETL previews", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, connectors.WriteRequest{Action: plan.Action, Config: runtime}, []connectors.Record{plan.ConnectorCommandRecord})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	return plan, preview, nil
}

func (a *App) confirmationChallengeForAction(connectorName, actionName string) string {
	connector, ok := a.registry.Get(connectorName)
	if !ok {
		return ""
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return strings.TrimSpace(action.Confirm)
		}
	}
	return ""
}

func (a *App) confirmationChallengeForPlan(plan ReversePlan) string {
	// Prefer the current connector manifest so a local state edit cannot remove
	// a destructive-action confirmation gate from an already-created plan. The
	// stored plan challenge remains a compatibility fallback for older plans or
	// connectors that are temporarily unavailable.
	if challenge := a.confirmationChallengeForAction(plan.DestinationConnector, plan.Action); challenge != "" {
		return challenge
	}
	return strings.TrimSpace(plan.ConfirmationChallenge)
}

func (a *App) validatePlanConfirmation(plan ReversePlan, got string) error {
	want := a.confirmationChallengeForPlan(plan)
	if want == "" {
		return nil
	}
	if strings.TrimSpace(got) != want {
		return fmt.Errorf("reverse plan %q requires typed confirmation: pass --confirm %s", plan.ID, want)
	}
	return nil
}

func (a *App) GetReversePlan(id string) (ReversePlan, error) {
	for _, plan := range a.state.ReversePlans {
		if plan.ID == id {
			return plan, nil
		}
	}
	return ReversePlan{}, fmt.Errorf("reverse plan %q not found", id)
}

func (a *App) ListReversePlans() []ReversePlan {
	out := append([]ReversePlan(nil), a.state.ReversePlans...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (a *App) GetReverseRun(id string) (ReverseRun, error) {
	for _, run := range a.state.ReverseRuns {
		if run.ID == id {
			return run, nil
		}
	}
	return ReverseRun{}, fmt.Errorf("reverse run %q not found", id)
}

func (a *App) ListReverseRuns() []ReverseRun {
	out := append([]ReverseRun(nil), a.state.ReverseRuns...)
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out
}

func (a *App) RunReverseETL(ctx context.Context, req RunReverseETLRequest) (ReverseRun, error) {
	planIndex := -1
	for i := range a.state.ReversePlans {
		if a.state.ReversePlans[i].ID == req.PlanID {
			planIndex = i
			break
		}
	}
	if planIndex < 0 {
		return ReverseRun{}, fmt.Errorf("reverse plan %q not found", req.PlanID)
	}
	plan := a.state.ReversePlans[planIndex]
	if plan.Status != "planned" {
		return ReverseRun{}, fmt.Errorf("reverse plan %q was already %s", req.PlanID, plan.Status)
	}
	if time.Now().UTC().After(plan.ExpiresAt) {
		return ReverseRun{}, errors.New("reverse plan approval has expired")
	}
	if plan.ApprovalTokenHash == "" {
		return ReverseRun{}, errors.New("reverse plan approval has already been consumed")
	}
	if hashString(req.ApprovalToken) != plan.ApprovalTokenHash {
		return ReverseRun{}, errors.New("approval token is invalid")
	}
	if err := a.validatePlanConfirmation(plan, req.Confirmation); err != nil {
		return ReverseRun{}, err
	}
	if plan.Mode == reversePlanModeConnectorCommand {
		return a.runConnectorCommandPlan(ctx, planIndex, plan)
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Limit: max(1, plan.RecordCount+1)})
	if err != nil {
		return ReverseRun{}, err
	}
	mappedForHash := mapReverseRecords(records, plan.Mappings, "")
	planHash, err := reversePlanHash(plan.Name, plan.SourceTable, plan.DestinationConnector, plan.DestinationCredential, plan.Action, plan.DestinationConfig, plan.Mappings, mappedForHash)
	if err != nil {
		return ReverseRun{}, err
	}
	if planHash != plan.PlanHash {
		a.state.ReversePlans[planIndex].Status = "invalidated"
		a.state.ReversePlans[planIndex].ApprovalTokenHash = ""
		a.state.ReversePlans[planIndex].ApprovalConsumedAt = time.Now().UTC()
		_ = a.save()
		return ReverseRun{}, errors.New("reverse plan source rows changed since approval")
	}
	mapped := mapReverseRecords(records, plan.Mappings, plan.ID)
	dest := EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: plan.DestinationConfig}
	writer, runtime, err := a.resolveEndpoint(ctx, dest)
	if err != nil {
		return ReverseRun{}, err
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(mapped), StartedAt: time.Now().UTC()}
	a.state.ReversePlans[planIndex].Status = "executing"
	a.state.ReversePlans[planIndex].ApprovalTokenHash = ""
	a.state.ReversePlans[planIndex].ApprovalConsumedAt = time.Now().UTC()
	if err := a.save(); err != nil {
		return ReverseRun{}, err
	}
	result, err := writer.Write(ctx, connectors.WriteRequest{
		Stream: "records",
		Table:  plan.Name,
		Action: plan.Action,
		Config: runtime,
	}, mapped)
	if err != nil {
		run.Status = "failed"
		run.RecordsSucceeded = result.RecordsWritten
		run.RecordsFailed = result.RecordsFailed
		if run.RecordsFailed == 0 {
			run.RecordsFailed = len(mapped) - result.RecordsWritten
		}
		run.Error = safety.RedactErrorText(err.Error())
		run.CompletedAt = time.Now().UTC()
		a.state.ReverseRuns = append(a.state.ReverseRuns, run)
		a.state.ReversePlans[planIndex].Status = "failed"
		_ = a.save()
		return run, err
	}
	run.Status = "completed"
	run.RecordsSucceeded = result.RecordsWritten
	run.RecordsFailed = result.RecordsFailed
	run.CompletedAt = time.Now().UTC()
	a.state.ReverseRuns = append(a.state.ReverseRuns, run)
	a.state.ReversePlans[planIndex].Status = "executed"
	if err := a.save(); err != nil {
		return ReverseRun{}, err
	}
	return run, nil
}

func (a *App) runConnectorCommandPlan(ctx context.Context, planIndex int, plan ReversePlan) (ReverseRun, error) {
	planHash, err := connectorCommandPlanHash(
		plan.Name,
		plan.DestinationConnector,
		plan.DestinationCredential,
		plan.DestinationConfig,
		plan.ConnectorCommand,
		plan.ConnectorCommandPath,
		plan.Action,
		plan.ConnectorCommandRecord,
	)
	if err != nil {
		return ReverseRun{}, err
	}
	if planHash != plan.PlanHash {
		a.state.ReversePlans[planIndex].Status = "invalidated"
		a.state.ReversePlans[planIndex].ApprovalTokenHash = ""
		a.state.ReversePlans[planIndex].ApprovalConsumedAt = time.Now().UTC()
		_ = a.save()
		return ReverseRun{}, errors.New("reverse plan command payload changed since approval")
	}
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReverseRun{}, err
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	records := []connectors.Record{cloneRecord(plan.ConnectorCommandRecord)}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(records), StartedAt: time.Now().UTC()}
	a.state.ReversePlans[planIndex].Status = "executing"
	a.state.ReversePlans[planIndex].ApprovalTokenHash = ""
	a.state.ReversePlans[planIndex].ApprovalConsumedAt = time.Now().UTC()
	if err := a.save(); err != nil {
		return ReverseRun{}, err
	}
	result, err := writer.Write(ctx, connectors.WriteRequest{
		Stream: "records",
		Table:  plan.Name,
		Action: plan.Action,
		Config: runtime,
	}, records)
	if err != nil {
		run.Status = "failed"
		run.RecordsSucceeded = result.RecordsWritten
		run.RecordsFailed = result.RecordsFailed
		if run.RecordsFailed == 0 {
			run.RecordsFailed = len(records) - result.RecordsWritten
		}
		run.Error = safety.RedactErrorText(err.Error())
		run.CompletedAt = time.Now().UTC()
		a.state.ReverseRuns = append(a.state.ReverseRuns, run)
		a.state.ReversePlans[planIndex].Status = "failed"
		_ = a.save()
		return run, err
	}
	run.Status = "completed"
	run.RecordsSucceeded = result.RecordsWritten
	run.RecordsFailed = result.RecordsFailed
	run.CompletedAt = time.Now().UTC()
	a.state.ReverseRuns = append(a.state.ReverseRuns, run)
	a.state.ReversePlans[planIndex].Status = "executed"
	if err := a.save(); err != nil {
		return ReverseRun{}, err
	}
	return run, nil
}

func (a *App) resolveEndpoint(ctx context.Context, endpoint EndpointConfig) (connectors.Connector, connectors.RuntimeConfig, error) {
	if err := connectors.RejectLegacyConnectorName(endpoint.Connector); err != nil {
		return nil, connectors.RuntimeConfig{}, err
	}
	cred, runtime, err := a.resolveCredential(ctx, endpoint.Credential, endpoint.Config)
	if err != nil {
		return nil, connectors.RuntimeConfig{}, err
	}
	if err := connectors.RejectLegacyConnectorName(cred.Connector); err != nil {
		return nil, connectors.RuntimeConfig{}, err
	}
	if endpoint.Connector != "" && endpoint.Connector != cred.Connector {
		return nil, connectors.RuntimeConfig{}, fmt.Errorf("credential %q is for connector %q, not %q", endpoint.Credential, cred.Connector, endpoint.Connector)
	}
	connector, ok := a.registry.Get(cred.Connector)
	if !ok {
		return nil, connectors.RuntimeConfig{}, fmt.Errorf("connector %q not found", cred.Connector)
	}
	registerConnectorSecretFields(connector)
	return connector, runtime, nil
}

func (a *App) ResolveConnectorCredential(ctx context.Context, connectorName, credentialName string, overlay map[string]string) (connectors.Connector, connectors.RuntimeConfig, error) {
	if strings.TrimSpace(credentialName) == "" {
		return nil, connectors.RuntimeConfig{}, errors.New("missing --credential")
	}
	return a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  connectorName,
		Credential: credentialName,
		Config:     overlay,
	})
}

func (a *App) resolveCredential(ctx context.Context, name string, overlay map[string]string) (CredentialMeta, connectors.RuntimeConfig, error) {
	cred, ok := a.findCredential(name)
	if !ok {
		return CredentialMeta{}, connectors.RuntimeConfig{}, fmt.Errorf("credential %q not found", name)
	}
	config := cloneStringMap(cred.Config)
	for k, v := range overlay {
		config[k] = v
	}
	if err := a.validateCredentialConfig(cred.Connector, config); err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	localWritePolicy, err := a.normalizeLocalWriteRuntime(cred.Connector, config)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	registerCredentialSecretFields(cred.SecretFields)
	secrets, err := a.vault.Get(ctx, cred.ID)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	runtimeProjectDir := a.projectDir
	if localWritePolicy != nil {
		runtimeProjectDir, err = filepath.Abs(a.projectDir)
		if err != nil {
			return CredentialMeta{}, connectors.RuntimeConfig{}, fmt.Errorf("resolve project directory: %w", err)
		}
	}
	return cred, connectors.RuntimeConfig{
		ProjectDir:       runtimeProjectDir,
		Config:           config,
		Secrets:          secrets,
		LocalWritePolicy: localWritePolicy,
	}, nil
}

func registerCredentialSecretFields(fields []string) {
	for _, field := range fields {
		pmlogging.RegisterSensitiveKey(field)
	}
}

func registerConnectorSecretFields(connector connectors.Connector) {
	manifest := connectors.ManifestOf(connector)
	for _, field := range manifest.SecretFields {
		pmlogging.RegisterSensitiveKey(field.Name)
	}
	for _, mode := range manifest.AuthModes {
		for _, field := range mode.SecretFields {
			pmlogging.RegisterSensitiveKey(field)
		}
	}
}

func (a *App) emitETLEvent(ctx context.Context, kind events.Kind, runID, streamName, status string, result etlExecutionResult, message string) {
	events.Emit(ctx, events.Event{
		Kind:    kind,
		Scope:   events.ScopeETL,
		RunID:   runID,
		StepID:  streamName,
		Status:  status,
		Message: message,
		Counters: events.Counters{
			RecordsRead:        int64(result.RecordsRead),
			RecordsTransformed: int64(result.RecordsTransformed),
			RecordsWritten:     int64(result.RecordsLoaded),
			RecordsFailed:      int64(result.RecordsFailed),
			Batches:            int64(result.BatchCount),
		},
	})
}

func (a *App) findCredential(name string) (CredentialMeta, bool) {
	for _, cred := range a.state.Credentials {
		if cred.Name == name || cred.ID == name {
			return cred, true
		}
	}
	return CredentialMeta{}, false
}

func (a *App) findConnection(name string) (Connection, bool) {
	for _, conn := range a.state.Connections {
		if conn.Name == name {
			return conn, true
		}
	}
	return Connection{}, false
}

func (a *App) failRun(ctx context.Context, runID string, err error) (Run, error) {
	for i := range a.state.Runs {
		if a.state.Runs[i].ID == runID {
			a.state.Runs[i].Status = "failed"
			a.state.Runs[i].Error = pmlogging.RedactText(ctx, err.Error())
			a.state.Runs[i].CompletedAt = time.Now().UTC()
			run := a.state.Runs[i]
			_ = a.save()
			pmlogging.FromContext(ctx).InfoContext(ctx, "etl run failed", "run_id", runID, "stream", run.Stream, "error", err)
			a.emitETLEvent(ctx, events.KindFailed, runID, run.Stream, "failed", etlExecutionResult{}, run.Error)
			return run, err
		}
	}
	pmlogging.FromContext(ctx).InfoContext(ctx, "etl run failed", "run_id", runID, "error", err)
	a.emitETLEvent(ctx, events.KindFailed, runID, "", "failed", etlExecutionResult{}, pmlogging.RedactText(ctx, err.Error()))
	return Run{}, err
}

func writeJSONAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write temp json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename temp json: %w", err)
	}
	return nil
}

func prefixedID(prefix string) (string, error) {
	token, err := randomToken(8)
	if err != nil {
		return "", err
	}
	return prefix + "_" + token, nil
}

func randomToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
