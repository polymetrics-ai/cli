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
	"polymetrics.ai/internal/connectors/registryset"
	"polymetrics.ai/internal/vault"
)

type App struct {
	root       string
	projectDir string
	statePath  string
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
	a := &App{
		root:       root,
		projectDir: projectDir,
		statePath:  filepath.Join(projectDir, "state", "state.json"),
		vault:      v,
		registry:   registryset.New(),
	}
	a.sqlEngine = newSQLEngine(a)
	if err := a.load(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) ProjectDir() string { return a.projectDir }

func (a *App) Registry() *connectors.Registry { return a.registry }

func (a *App) Connectors() []connectors.Metadata {
	return a.registry.List()
}

func (a *App) Connector(name string) (connectors.Metadata, error) {
	c, ok := a.registry.Get(name)
	if !ok {
		return connectors.Metadata{}, fmt.Errorf("connector %q not found", name)
	}
	return c.Metadata(), nil
}

func (a *App) load() error {
	b, err := os.ReadFile(a.statePath)
	if errors.Is(err, os.ErrNotExist) {
		a.state = state{Checkpoints: map[string]map[string]string{}, StreamStates: map[string]StreamState{}}
		return a.save()
	}
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}
	if len(b) == 0 {
		a.state = state{Checkpoints: map[string]map[string]string{}, StreamStates: map[string]StreamState{}}
		return nil
	}
	if err := json.Unmarshal(b, &a.state); err != nil {
		return fmt.Errorf("decode state: %w", err)
	}
	if a.state.Checkpoints == nil {
		a.state.Checkpoints = map[string]map[string]string{}
	}
	if a.state.StreamStates == nil {
		a.state.StreamStates = map[string]StreamState{}
	}
	return nil
}

func (a *App) save() error {
	return writeJSONAtomic(a.statePath, a.state)
}

func (a *App) AddCredential(ctx context.Context, req AddCredentialRequest) (CredentialMeta, error) {
	if strings.TrimSpace(req.Name) == "" {
		return CredentialMeta{}, errors.New("credential name is required")
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
	if destination.Name() == "warehouse" {
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
	err := warehouse.Read(ctx, connectors.ReadRequest{Stream: req.Table, Config: cfg}, func(record connectors.Record) error {
		if len(rows) >= req.Limit {
			return nil
		}
		rows = append(rows, record)
		return nil
	})
	if err != nil {
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
	planForHash := map[string]any{
		"name": req.Name, "source_table": req.SourceTable, "destination_connector": req.DestinationConnector,
		"destination_credential": req.DestinationCredential, "action": req.Action, "mappings": req.Mappings,
		"record_count": len(records),
	}
	planHash, err := hashJSON(planForHash)
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
	if time.Now().UTC().After(plan.ExpiresAt) {
		return ReverseRun{}, errors.New("reverse plan approval has expired")
	}
	if hashString(req.ApprovalToken) != plan.ApprovalTokenHash {
		return ReverseRun{}, errors.New("approval token is invalid")
	}
	records, err := a.QueryTable(ctx, QueryTableRequest{Table: plan.SourceTable, Limit: max(1, plan.RecordCount)})
	if err != nil {
		return ReverseRun{}, err
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
		run.Error = err.Error()
		run.CompletedAt = time.Now().UTC()
		a.state.ReverseRuns = append(a.state.ReverseRuns, run)
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
	cred, runtime, err := a.resolveCredential(ctx, endpoint.Credential, endpoint.Config)
	if err != nil {
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
	return cred, connectors.RuntimeConfig{ProjectDir: a.projectDir, Config: config, Secrets: secrets}, nil
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
			a.state.Runs[i].Error = err.Error()
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
