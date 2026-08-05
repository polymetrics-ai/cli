package app

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
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
	"polymetrics.ai/internal/safety"
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/vault"
)

const reversePlanModeConnectorCommand = "connector_command"

var errStateRevisionConflict = errors.New("project state changed in another process")

type App struct {
	root       string
	projectDir string
	statePath  string
	store      statestore.JSONStore[state]
	state      state
	vault      *vault.Vault
	approval   *projectWriteApprovalAuthority
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
	Revision     uint64                       `json:"revision"`
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
	approval, err := newProjectWriteApprovalAuthority(projectDir)
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
		approval:   approval,
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
	expectedRevision := a.state.Revision
	next := a.state
	updated, err := a.store.Update(func(current state) (state, error) {
		if current.Revision != expectedRevision {
			return current, errStateRevisionConflict
		}
		next.Revision = current.Revision + 1
		return next, nil
	})
	if err == nil || errors.Is(err, errStateRevisionConflict) {
		a.state = updated
	}
	return err
}

func (a *App) updateState(update func(state) (state, error)) (state, error) {
	updated, err := a.store.Update(func(current state) (state, error) {
		next, updateErr := update(current)
		if updateErr != nil {
			return current, updateErr
		}
		next.Revision = current.Revision + 1
		return next, nil
	})
	if err == nil {
		a.state = updated
	}
	return updated, err
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
	connector, ok := a.registry.Get(req.Connector)
	if !ok {
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
	if err := connectors.ValidateConfiguration(connector, req.Config); err != nil {
		return CredentialMeta{}, fmt.Errorf("credential configuration: %w", err)
	}
	if err := a.vault.Put(ctx, id, req.Secrets); err != nil {
		return CredentialMeta{}, err
	}
	now := time.Now().UTC()
	fields := make([]string, 0, len(req.Secrets))
	for k := range req.Secrets {
		fields = append(fields, k)
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

func (a *App) RunETL(ctx context.Context, req RunETLRequest) (Run, error) {
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
	run := Run{ID: runID, Type: "etl", Connection: req.Connection, Stream: req.Stream, Status: "running", StartedAt: time.Now().UTC()}
	a.state.Runs = append(a.state.Runs, run)
	_ = a.save()

	source, sourceRuntime, err := a.resolveEndpoint(ctx, conn.Source)
	if err != nil {
		return a.failRun(runID, err)
	}
	destination, destRuntime, err := a.resolveEndpoint(ctx, conn.Destination)
	if err != nil {
		return a.failRun(runID, err)
	}
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}
	mode, err := ParseSyncMode(stream.SyncMode)
	if err != nil {
		return a.failRun(runID, err)
	}
	stream.SyncMode = mode.Name
	if err := ValidateStreamSyncConfig(stream); err != nil {
		return a.failRun(runID, err)
	}
	var result etlExecutionResult
	if materializer, ok := destination.(connectors.LocalWarehouseMaterializer); ok && materializer.MaterializesLocalWarehouse() {
		result, err = a.runWarehouseETL(ctx, runID, conn, source, sourceRuntime, destRuntime, req.Stream, stream, mode, batchSize)
	} else {
		result, err = a.runConnectorETL(ctx, runID, conn, source, sourceRuntime, destination, destRuntime, req.Stream, stream, mode, batchSize)
	}
	if err != nil {
		return a.failRun(runID, err)
	}
	return a.completeRun(runID, result)
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
		result.RecordsFailed += writeResult.RecordsFailed
		result.BatchCount++
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

func (a *App) completeRun(runID string, result etlExecutionResult) (Run, error) {
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
	// Refuse before any rows are read, any plan row is stored, or any approval
	// token is minted: a non-batchable action must leave nothing approvable
	// behind.
	if err := a.guardBatchableAction(req.DestinationConnector, req.Action, req.SourceTable); err != nil {
		return ReversePlan{}, err
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: req.SourceTable, Limit: req.Limit})
	if err != nil {
		return ReversePlan{}, err
	}
	mapped := mapReverseRecords(records, req.Mappings)
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
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, mapped)
	if err != nil {
		return ReversePlan{}, err
	}
	planHash, err := reversePlanHash(req.Name, req.SourceTable, req.DestinationConnector, req.DestinationCredential, req.Action, req.DestinationConfig, req.Mappings, mapped, payloadIdentity)
	if err != nil {
		return ReversePlan{}, err
	}
	sampleCount := min(3, len(mapped))
	redactFields := reversePlanRedactFields(destination, req.Action)
	confirmation := a.confirmationPolicyForAction(req.DestinationConnector, req.Action)
	challenge := string(confirmation.Kind)
	created := time.Now().UTC()
	expires := created.Add(24 * time.Hour)
	var planSeal *connectors.WritePlanSeal
	if confirmation.Kind != "" {
		seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
			PlanID: id, PlanHash: planHash, Connector: req.DestinationConnector, Operation: req.Action,
			CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
			Batchable: a.actionIsBatchable(req.DestinationConnector, req.Action), Scope: runtime.WriteApprovalScope,
			Confirmation: confirmation,
		})
		if err != nil {
			return ReversePlan{}, err
		}
		planSeal = &seal
		created = seal.IssuedAt
		expires = seal.ExpiresAt
	}
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
		PayloadIdentity:       payloadIdentity,
		ConfirmationChallenge: challenge,
		ConfirmationPolicy:    confirmationFromChallenge(challenge),
		RedactFields:          redactFields,
		RecordCount:           len(records),
		Sample:                RedactReversePlanRecords(mapped[:sampleCount], redactFields),
		PlanHash:              planHash,
		PlanSeal:              planSeal,
		CreatedAt:             created,
		ExpiresAt:             expires,
	}
	if challenge == "" {
		token, err := randomToken(18)
		if err != nil {
			return ReversePlan{}, err
		}
		plan.ApprovalTokenHash = hashString(token)
		plan.ApprovalToken = token
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
		Preview: false,
	})
	if err != nil {
		return ReversePlan{}, nil, err
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.ReplaceAll(writeCommand.Command, " ", "_")
	}
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, []connectors.Record{writeCommand.Record})
	if err != nil {
		return ReversePlan{}, nil, err
	}
	if req.Preview {
		runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(payloadIdentity)
		dryRunner, ok := connector.(connectors.DryRunWriter)
		if !ok {
			return ReversePlan{}, nil, fmt.Errorf("connector %q does not support reverse ETL previews", connector.Name())
		}
		preview, err := dryRunner.DryRunWrite(ctx, connectors.WriteRequest{Action: writeCommand.Write, Config: runtime}, []connectors.Record{writeCommand.Record})
		if err != nil {
			return ReversePlan{}, nil, err
		}
		writeCommand.Preview = &preview
	}
	planHash, err := connectorCommandPlanHash(name, req.Connector, req.Credential, req.Config, writeCommand.Command, req.Path, writeCommand.Write, writeCommand.Record, payloadIdentity)
	if err != nil {
		return ReversePlan{}, nil, err
	}
	confirmation := confirmationFromChallenge(writeCommand.ConfirmationChallenge)
	created := time.Now().UTC()
	expires := created.Add(24 * time.Hour)
	var planSeal *connectors.WritePlanSeal
	if confirmation.Kind != "" {
		seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
			PlanID: id, PlanHash: planHash, Mode: reversePlanModeConnectorCommand,
			Connector: req.Connector, Operation: writeCommand.Write,
			CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
			Batchable: a.actionIsBatchable(req.Connector, writeCommand.Write), Scope: runtime.WriteApprovalScope,
			Confirmation: confirmation,
		})
		if err != nil {
			return ReversePlan{}, nil, err
		}
		planSeal = &seal
		created = seal.IssuedAt
		expires = seal.ExpiresAt
	}
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
		PayloadIdentity:        payloadIdentity,
		ConfirmationChallenge:  writeCommand.ConfirmationChallenge,
		ConfirmationPolicy:     confirmationFromChallenge(writeCommand.ConfirmationChallenge),
		RecordCount:            1,
		Sample:                 []connectors.Record{cloneRecord(writeCommand.RedactedRecord)},
		PlanHash:               planHash,
		PlanSeal:               planSeal,
		CreatedAt:              created,
		ExpiresAt:              expires,
	}
	if strings.TrimSpace(writeCommand.ConfirmationChallenge) == "" {
		token, err := randomToken(18)
		if err != nil {
			return ReversePlan{}, nil, err
		}
		plan.ApprovalTokenHash = hashString(token)
		plan.ApprovalToken = token
	}
	stored := plan
	stored.ApprovalToken = ""
	a.state.ReversePlans = append(a.state.ReversePlans, stored)
	if err := a.save(); err != nil {
		return ReversePlan{}, nil, err
	}
	if writeCommand.Preview != nil && a.confirmationChallengeForPlan(plan) != "" {
		plan, err = a.persistDestructivePreview(plan, *writeCommand.Preview)
		if err != nil {
			return ReversePlan{}, nil, err
		}
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
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, []connectors.Record{plan.ConnectorCommandRecord})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	currentHash, err := connectorCommandPlanHash(
		plan.Name,
		plan.DestinationConnector,
		plan.DestinationCredential,
		plan.DestinationConfig,
		plan.ConnectorCommand,
		plan.ConnectorCommandPath,
		plan.Action,
		plan.ConnectorCommandRecord,
		payloadIdentity,
	)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if currentHash != plan.PlanHash {
		return ReversePlan{}, connectors.WritePreview{}, errors.New("reverse plan command payload changed before preview")
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(payloadIdentity)
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
	if a.confirmationChallengeForPlan(plan) != "" {
		plan, err = a.persistDestructivePreview(plan, preview)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	return plan, preview, nil
}

// PreviewReversePlan materializes the exact mapped write request without
// dispatching it. Destructive plans become approvable only after this preview
// identity is persisted; source-row or payload drift fails before a token is
// minted.
func (a *App) PreviewReversePlan(ctx context.Context, id string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(id)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if plan.Mode == reversePlanModeConnectorCommand {
		return a.PreviewConnectorCommandPlan(ctx, id)
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.guardBatchableAction(plan.DestinationConnector, plan.Action, plan.SourceTable); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Limit: max(1, plan.RecordCount+1)})
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	mapped := mapReverseRecords(records, plan.Mappings)
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, mapped)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(payloadIdentity)
	currentHash, err := reversePlanHash(plan.Name, plan.SourceTable, plan.DestinationConnector, plan.DestinationCredential, plan.Action, plan.DestinationConfig, plan.Mappings, mapped, payloadIdentity)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if currentHash != plan.PlanHash {
		return ReversePlan{}, connectors.WritePreview{}, errors.New("reverse plan source rows or payload files changed before preview")
	}
	request := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, request, mapped); err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("connector %q does not support reverse ETL previews", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, mapped)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if a.confirmationChallengeForPlan(plan) != "" {
		plan, err = a.persistDestructivePreview(plan, preview)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	return plan, preview, nil
}

func (a *App) persistDestructivePreview(plan ReversePlan, preview connectors.WritePreview) (ReversePlan, error) {
	if strings.TrimSpace(preview.Digest) == "" {
		return ReversePlan{}, fmt.Errorf("connector preview for destructive plan %q has no digest", plan.ID)
	}
	if strings.TrimSpace(preview.ApprovalTarget.Connector) == "" || preview.ApprovalTarget.Connector != plan.DestinationConnector || preview.ApprovalTarget.Operation != plan.Action {
		return ReversePlan{}, fmt.Errorf("connector preview for destructive plan %q has no matching approval target", plan.ID)
	}
	token, err := randomToken(18)
	if err != nil {
		return ReversePlan{}, err
	}
	now := time.Now().UTC()
	var issued ReversePlan
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			stored := current.ReversePlans[i]
			if stored.ID != plan.ID {
				continue
			}
			if err := a.previewabilityError(stored, now); err != nil {
				return current, err
			}
			if stored.PlanHash != plan.PlanHash || stored.DestinationConnector != plan.DestinationConnector || stored.DestinationCredential != plan.DestinationCredential || stored.Action != plan.Action {
				return current, fmt.Errorf("reverse plan %q changed while its preview was prepared", plan.ID)
			}
			if err := a.verifyPlanSealForTarget(stored, preview.ApprovalTarget); err != nil {
				return current, err
			}
			grant, err := a.approval.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
				PlanID:        stored.ID,
				PlanHash:      stored.PlanHash,
				Mode:          stored.Mode,
				PlanSeal:      stored.PlanSeal,
				PreviewDigest: preview.Digest,
				ApprovalToken: token,
				Target:        preview.ApprovalTarget,
				Confirmation:  a.confirmationPolicyForPlan(stored),
			})
			if err != nil {
				return current, err
			}
			stored.Status = "previewed"
			stored.PreviewDigest = preview.Digest
			stored.PreviewedAt = now
			stored.ApprovalTokenHash = hashString(token)
			stored.ApprovalGrant = &grant
			stored.ApprovalConsumedAt = time.Time{}
			current.ReversePlans[i] = stored
			issued = stored
			issued.ApprovalToken = token
			return current, nil
		}
		return current, fmt.Errorf("reverse plan %q not found", plan.ID)
	})
	if err != nil {
		return ReversePlan{}, err
	}
	a.state = updated
	return issued, nil
}

func (a *App) previewabilityError(plan ReversePlan, now time.Time) error {
	if plan.Status != "planned" && plan.Status != "previewed" {
		return fmt.Errorf("reverse plan %q was already %s", plan.ID, plan.Status)
	}
	if a.confirmationPolicyForPlan(plan).Kind == "" && (plan.CreatedAt.IsZero() || plan.ExpiresAt.IsZero() || !plan.ExpiresAt.After(plan.CreatedAt) || now.Before(plan.CreatedAt) || !now.Before(plan.ExpiresAt)) {
		return fmt.Errorf("reverse plan %q approval has expired or is not active", plan.ID)
	}
	return nil
}

func confirmationFromChallenge(challenge string) connectors.WriteConfirmation {
	confirmation, err := connectors.ParseWriteConfirmation(challenge)
	if err != nil {
		return connectors.WriteConfirmation{}
	}
	return confirmation
}

func (a *App) confirmationPolicyForAction(connectorName, actionName string) connectors.WriteConfirmation {
	connector, ok := a.registry.Get(connectorName)
	if !ok {
		return connectors.WriteConfirmation{}
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return connectors.ConfirmationForWriteAction(action)
		}
	}
	return connectors.WriteConfirmation{}
}

func (a *App) confirmationPolicyForPlan(plan ReversePlan) connectors.WriteConfirmation {
	// Prefer the current connector manifest so a local state edit cannot remove
	// a destructive-action confirmation gate from an already-created plan. The
	// stored plan challenge remains a compatibility fallback for older plans or
	// connectors that are temporarily unavailable.
	if confirmation := a.confirmationPolicyForAction(plan.DestinationConnector, plan.Action); confirmation.Kind != "" {
		return confirmation
	}
	if plan.ConfirmationPolicy.Kind != "" {
		return plan.ConfirmationPolicy
	}
	return confirmationFromChallenge(plan.ConfirmationChallenge)
}

func (a *App) actionIsBatchable(connectorName, actionName string) bool {
	connector, ok := a.registry.Get(connectorName)
	if !ok {
		return true
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name == actionName {
			return action.IsBatchable()
		}
	}
	return true
}

func (a *App) verifyPlanSealForRuntime(plan ReversePlan, runtime connectors.RuntimeConfig) error {
	confirmation := a.confirmationPolicyForPlan(plan)
	if confirmation.Kind == "" {
		return nil
	}
	if plan.PlanSeal == nil {
		return fmt.Errorf("reverse plan %q has no authenticated plan seal", plan.ID)
	}
	return a.approval.VerifyWritePlanSeal(*plan.PlanSeal, connectors.WritePlanSealExpectation{
		PlanID: plan.ID, PlanHash: plan.PlanHash, Mode: plan.Mode,
		Connector: plan.DestinationConnector, Operation: plan.Action,
		CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
		Batchable: a.actionIsBatchable(plan.DestinationConnector, plan.Action), Scope: runtime.WriteApprovalScope,
		Confirmation: confirmation,
	})
}

func (a *App) verifyPlanSealForTarget(plan ReversePlan, target connectors.WriteApprovalTarget) error {
	if plan.PlanSeal == nil {
		return fmt.Errorf("reverse plan %q has no authenticated plan seal", plan.ID)
	}
	return a.approval.VerifyWritePlanSeal(*plan.PlanSeal, connectors.WritePlanSealExpectation{
		PlanID: plan.ID, PlanHash: plan.PlanHash, Mode: plan.Mode,
		Connector: target.Connector, Operation: target.Operation,
		CredentialRevision: target.CredentialRevision, ConfigurationDigest: target.ConfigurationDigest,
		Batchable: target.Batchable, Scope: target.Scope, Confirmation: target.Confirmation,
	})
}

func (a *App) confirmationChallengeForPlan(plan ReversePlan) string {
	return string(a.confirmationPolicyForPlan(plan).Kind)
}

func (a *App) validatePlanConfirmation(plan ReversePlan, got connectors.WriteConfirmation) error {
	want := a.confirmationPolicyForPlan(plan)
	if want.Kind == "" {
		return nil
	}
	if got.Kind != want.Kind {
		return fmt.Errorf("reverse plan %q requires typed confirmation: pass --confirm %s", plan.ID, want.Kind)
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
	plan, err := a.loadReversePlan(req.PlanID)
	if err != nil {
		return ReverseRun{}, err
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReverseRun{}, err
	}
	if a.confirmationChallengeForPlan(plan) != "" && (plan.Status != "previewed" || plan.PreviewDigest == "" || plan.PreviewedAt.IsZero()) {
		return ReverseRun{}, fmt.Errorf("reverse plan %q must be previewed before approval", plan.ID)
	}
	if plan.ApprovalTokenHash == "" {
		return ReverseRun{}, errors.New("reverse plan approval has already been consumed")
	}
	if !constantTimeStringEqual(hashString(req.ApprovalToken), plan.ApprovalTokenHash) {
		return ReverseRun{}, errors.New("approval token is invalid")
	}
	if err := a.validatePlanConfirmation(plan, req.Confirmation); err != nil {
		return ReverseRun{}, err
	}
	if plan.Mode == reversePlanModeConnectorCommand {
		return a.runConnectorCommandPlan(ctx, plan, req)
	}
	return a.runBulkReversePlan(ctx, plan, req)
}

func (a *App) runBulkReversePlan(ctx context.Context, plan ReversePlan, req RunReverseETLRequest) (ReverseRun, error) {
	if err := a.guardBatchableAction(plan.DestinationConnector, plan.Action, plan.SourceTable); err != nil {
		return ReverseRun{}, err
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Limit: max(1, plan.RecordCount+1)})
	if err != nil {
		return ReverseRun{}, err
	}
	mappedForHash := mapReverseRecords(records, plan.Mappings)
	dest := EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: plan.DestinationConfig}
	writer, runtime, err := a.resolveEndpoint(ctx, dest)
	if err != nil {
		return ReverseRun{}, err
	}
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, mappedForHash)
	if err != nil {
		return ReverseRun{}, err
	}
	planHash, err := reversePlanHash(plan.Name, plan.SourceTable, plan.DestinationConnector, plan.DestinationCredential, plan.Action, plan.DestinationConfig, plan.Mappings, mappedForHash, payloadIdentity)
	if err != nil {
		return ReverseRun{}, err
	}
	if planHash != plan.PlanHash {
		a.invalidateReversePlan(plan.ID)
		return ReverseRun{}, errors.New("reverse plan source rows or payload files changed since approval")
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(plan.PayloadIdentity)
	mapped := mapReverseRecords(records, plan.Mappings)
	writeRequest := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	preview, err := a.validateDestructivePreview(ctx, writer, plan, writeRequest, mapped)
	if err != nil {
		return ReverseRun{}, err
	}
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(mapped), StartedAt: time.Now().UTC()}
	evidence, _, err := a.consumePlanApproval(plan, req, preview)
	if err != nil {
		return ReverseRun{}, err
	}
	writeRequest.Approval = evidence
	result, err := writer.Write(ctx, writeRequest, mapped)
	return a.finishReverseWrite(plan.ID, run, result, len(mapped), err)
}

func (a *App) validateDestructivePreview(ctx context.Context, writer connectors.Connector, plan ReversePlan, request connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if a.confirmationPolicyForPlan(plan).Kind == "" {
		return connectors.WritePreview{}, nil
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return connectors.WritePreview{}, fmt.Errorf("connector %q no longer supports the required destructive preview", writer.Name())
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, records)
	if err != nil {
		return connectors.WritePreview{}, fmt.Errorf("revalidate destructive preview: %w", err)
	}
	if strings.TrimSpace(plan.PreviewDigest) == "" || strings.TrimSpace(preview.Digest) == "" || subtle.ConstantTimeCompare([]byte(plan.PreviewDigest), []byte(preview.Digest)) != 1 {
		return connectors.WritePreview{}, fmt.Errorf("reverse plan %q no longer matches its approved preview", plan.ID)
	}
	return preview, nil
}

func (a *App) runConnectorCommandPlan(ctx context.Context, plan ReversePlan, req RunReverseETLRequest) (ReverseRun, error) {
	writer, runtime, err := a.resolveEndpoint(ctx, EndpointConfig{
		Connector:  plan.DestinationConnector,
		Credential: plan.DestinationCredential,
		Config:     plan.DestinationConfig,
	})
	if err != nil {
		return ReverseRun{}, err
	}
	payloadIdentity, err := payloadIdentitiesForRecords(runtime.ProjectDir, []connectors.Record{plan.ConnectorCommandRecord})
	if err != nil {
		return ReverseRun{}, err
	}
	planHash, err := connectorCommandPlanHash(
		plan.Name,
		plan.DestinationConnector,
		plan.DestinationCredential,
		plan.DestinationConfig,
		plan.ConnectorCommand,
		plan.ConnectorCommandPath,
		plan.Action,
		plan.ConnectorCommandRecord,
		payloadIdentity,
	)
	if err != nil {
		return ReverseRun{}, err
	}
	if planHash != plan.PlanHash {
		a.invalidateReversePlan(plan.ID)
		return ReverseRun{}, errors.New("reverse plan command payload changed since approval")
	}
	runtime.ApprovedPayloadSHA256 = approvedPayloadSHA256(plan.PayloadIdentity)
	runID, err := prefixedID("rrun")
	if err != nil {
		return ReverseRun{}, err
	}
	records := []connectors.Record{plan.ConnectorCommandRecord}
	writeRequest := connectors.WriteRequest{Stream: "records", Table: plan.Name, Action: plan.Action, Config: runtime}
	preview, err := a.validateDestructivePreview(ctx, writer, plan, writeRequest, records)
	if err != nil {
		return ReverseRun{}, err
	}
	run := ReverseRun{ID: runID, PlanID: plan.ID, Status: "running", RecordsStaged: len(records), StartedAt: time.Now().UTC()}
	evidence, _, err := a.consumePlanApproval(plan, req, preview)
	if err != nil {
		return ReverseRun{}, err
	}
	writeRequest.Approval = evidence
	result, err := writer.Write(ctx, writeRequest, records)
	return a.finishReverseWrite(plan.ID, run, result, len(records), err)
}

func (a *App) loadReversePlan(id string) (ReversePlan, error) {
	loaded, err := a.store.Load()
	if err != nil {
		return ReversePlan{}, err
	}
	a.state = loaded
	for _, plan := range loaded.ReversePlans {
		if plan.ID == id {
			return plan, nil
		}
	}
	return ReversePlan{}, fmt.Errorf("reverse plan %q not found", id)
}

func (a *App) consumePlanApproval(expected ReversePlan, req RunReverseETLRequest, preview connectors.WritePreview) (*connectors.WriteApprovalEvidence, ReversePlan, error) {
	now := time.Now().UTC()
	var consumed ReversePlan
	var evidence *connectors.WriteApprovalEvidence
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			stored := current.ReversePlans[i]
			if stored.ID != expected.ID {
				continue
			}
			if err := a.previewabilityError(stored, now); err != nil {
				return current, err
			}
			if stored.PlanHash != expected.PlanHash || stored.DestinationConnector != expected.DestinationConnector || stored.DestinationCredential != expected.DestinationCredential || stored.Action != expected.Action || stored.Mode != expected.Mode {
				return current, fmt.Errorf("reverse plan %q changed before approval consumption", stored.ID)
			}
			if err := a.validatePlanConfirmation(stored, req.Confirmation); err != nil {
				return current, err
			}
			if stored.ApprovalTokenHash == "" {
				return current, errors.New("reverse plan approval has already been consumed")
			}
			if !constantTimeStringEqual(stored.ApprovalTokenHash, hashString(req.ApprovalToken)) {
				return current, errors.New("approval token is invalid")
			}
			if a.confirmationPolicyForPlan(stored).Kind != "" {
				if stored.Status != "previewed" || stored.PreviewDigest == "" || stored.PreviewedAt.IsZero() || stored.ApprovalGrant == nil {
					return current, fmt.Errorf("reverse plan %q must be previewed before approval", stored.ID)
				}
				if !constantTimeStringEqual(stored.PreviewDigest, preview.Digest) {
					return current, fmt.Errorf("reverse plan %q no longer matches its approved preview", stored.ID)
				}
				verified, err := a.approval.VerifyWriteGrant(*stored.ApprovalGrant, connectors.WriteApprovalExpectation{
					PlanID:        expected.ID,
					PlanHash:      expected.PlanHash,
					Mode:          expected.Mode,
					PreviewDigest: preview.Digest,
					ApprovalToken: req.ApprovalToken,
					Target:        preview.ApprovalTarget,
					Confirmation:  req.Confirmation,
				}, stored.PlanSeal)
				if err != nil {
					return current, err
				}
				evidence = verified
			}
			stored.Status = "executing"
			stored.ApprovalTokenHash = ""
			stored.ApprovalGrant = nil
			stored.ApprovalConsumedAt = now
			current.ReversePlans[i] = stored
			consumed = stored
			return current, nil
		}
		return current, fmt.Errorf("reverse plan %q not found", expected.ID)
	})
	if err != nil {
		return nil, ReversePlan{}, err
	}
	a.state = updated
	return evidence, consumed, nil
}

func (a *App) finishReverseWrite(planID string, run ReverseRun, result connectors.WriteResult, staged int, writeErr error) (ReverseRun, error) {
	run.RecordsSucceeded = result.RecordsWritten
	run.RecordsFailed = result.RecordsFailed
	run.CompletedAt = time.Now().UTC()
	planStatus := "executed"
	if writeErr != nil {
		run.Status = "failed"
		planStatus = "failed"
		if run.RecordsFailed == 0 {
			run.RecordsFailed = staged - result.RecordsWritten
		}
		run.Error = safety.RedactErrorText(writeErr.Error())
	} else {
		run.Status = "completed"
	}
	updated, persistErr := a.updateState(func(current state) (state, error) {
		current.ReverseRuns = append(current.ReverseRuns, run)
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID == planID && current.ReversePlans[i].Status == "executing" {
				current.ReversePlans[i].Status = planStatus
				break
			}
		}
		return current, nil
	})
	if persistErr == nil {
		a.state = updated
	}
	if writeErr != nil {
		return run, writeErr
	}
	if persistErr != nil {
		return ReverseRun{}, persistErr
	}
	return run, nil
}

func (a *App) invalidateReversePlan(planID string) {
	now := time.Now().UTC()
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID != planID {
				continue
			}
			current.ReversePlans[i].Status = "invalidated"
			current.ReversePlans[i].ApprovalTokenHash = ""
			current.ReversePlans[i].ApprovalGrant = nil
			current.ReversePlans[i].ApprovalConsumedAt = now
			break
		}
		return current, nil
	})
	if err == nil {
		a.state = updated
	}
}

func constantTimeStringEqual(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
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
	secrets, err := a.vault.Get(ctx, cred.ID)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	config := cloneStringMap(cred.Config)
	for k, v := range overlay {
		config[k] = v
	}
	credentialRevision, err := a.approval.CredentialRevision(cred.ID, secrets)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	configurationDigest, err := a.approval.ConfigurationDigest(cred.ID, config)
	if err != nil {
		return CredentialMeta{}, connectors.RuntimeConfig{}, err
	}
	return cred, connectors.RuntimeConfig{
		ProjectDir:          a.projectDir,
		Config:              config,
		Secrets:             secrets,
		CredentialRevision:  credentialRevision,
		ConfigurationDigest: configurationDigest,
		WriteApprovalScope:  connectors.WriteApprovalScopeProject,
		// Scoped to this credential, so a provider-rotated secret (an OAuth2
		// refresh token) is written back to the same encrypted vault entry it
		// was read from, and to no other.
		SecretStore: a.credentialSecretStore(cred.ID),
	}, nil
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

func (a *App) failRun(runID string, err error) (Run, error) {
	for i := range a.state.Runs {
		if a.state.Runs[i].ID == runID {
			a.state.Runs[i].Status = "failed"
			a.state.Runs[i].Error = safety.RedactErrorText(err.Error())
			a.state.Runs[i].CompletedAt = time.Now().UTC()
			run := a.state.Runs[i]
			_ = a.save()
			return run, err
		}
	}
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
