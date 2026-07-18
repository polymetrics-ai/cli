package connectors

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/safety"
)

var ErrUnsupportedOperation = errors.New("unsupported connector operation")

var defaultRegistryBuilder = struct {
	mu sync.RWMutex
	fn func() *Registry
}{}

// RegisterDefaultRegistryBuilder installs the process default registry builder.
// Wave 6 uses this to let the bundle-backed registry live in a package that can
// import engine/defs without creating a connectors<->engine cycle.
func RegisterDefaultRegistryBuilder(fn func() *Registry) {
	defaultRegistryBuilder.mu.Lock()
	defer defaultRegistryBuilder.mu.Unlock()
	defaultRegistryBuilder.fn = fn
}

func registeredDefaultRegistryBuilder() func() *Registry {
	defaultRegistryBuilder.mu.RLock()
	defer defaultRegistryBuilder.mu.RUnlock()
	return defaultRegistryBuilder.fn
}

type Record map[string]any

type Capabilities struct {
	Check   bool `json:"check"`
	Catalog bool `json:"catalog"`
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Query   bool `json:"query"`
}

type Metadata struct {
	Name            string         `json:"name"`
	DisplayName     string         `json:"display_name"`
	IntegrationType string         `json:"integration_type"`
	Description     string         `json:"description"`
	Capabilities    Capabilities   `json:"capabilities"`
	Icon            *ConnectorIcon `json:"icon,omitempty"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Stream struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Fields       []Field  `json:"fields"`
	PrimaryKey   []string `json:"primary_key"`
	CursorFields []string `json:"cursor_fields"`
}

type Catalog struct {
	Connector string   `json:"connector"`
	Streams   []Stream `json:"streams"`
}

type RuntimeConfig struct {
	ProjectDir       string            `json:"-"`
	Config           map[string]string `json:"config"`
	Secrets          map[string]string `json:"-"`
	LocalWritePolicy *LocalWritePolicy `json:"-"`
}

type LocalWritePolicy struct {
	ProjectRoot   string
	AllowExternal bool
}

type ReadRequest struct {
	Stream string
	Config RuntimeConfig
	State  map[string]string
	Query  map[string]string
	Limit  int
}

type DirectReadRequest struct {
	Method       string
	Path         string
	Config       RuntimeConfig
	PathParams   map[string]string
	Query        map[string]string
	MaxBytes     int
	OutputPolicy string
}

type DirectReadResult struct {
	Connector string `json:"connector"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Body      any    `json:"body"`
}

type DirectReader interface {
	DirectRead(context.Context, DirectReadRequest) (DirectReadResult, error)
}

var ErrReadLimitReached = errors.New("connector read limit reached")

func LimitEmitter(limit int, emit func(Record) error) func(Record) error {
	if limit <= 0 {
		return emit
	}
	count := 0
	return func(record Record) error {
		if count >= limit {
			return ErrReadLimitReached
		}
		if err := emit(record); err != nil {
			return err
		}
		count++
		if count >= limit {
			return ErrReadLimitReached
		}
		return nil
	}
}

func IgnoreReadLimit(err error) error {
	if errors.Is(err, ErrReadLimitReached) {
		return nil
	}
	return err
}

func RejectLegacyConnectorName(name string) error {
	if !IsLegacyConnectorName(name) {
		return nil
	}
	return fmt.Errorf("connector %q uses a legacy source-/destination- prefix; use bare connector name %q", name, legacyBareConnectorName(name))
}

func IsLegacyConnectorName(name string) bool {
	normalized := strings.TrimSpace(strings.ToLower(name))
	return strings.HasPrefix(normalized, "source-") || strings.HasPrefix(normalized, "destination-")
}

func legacyBareConnectorName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(name))
	normalized = strings.TrimPrefix(normalized, "source-")
	normalized = strings.TrimPrefix(normalized, "destination-")
	return normalized
}

type WriteRequest struct {
	Stream     string
	Table      string
	Action     string
	Overwrite  bool
	Config     RuntimeConfig
	PrimaryKey []string
}

type WriteResult struct {
	RecordsWritten int `json:"records_written"`
	RecordsFailed  int `json:"records_failed"`
}

type QueryRequest struct {
	SQL    string
	Stream string
	Limit  int
	Config RuntimeConfig
}

type QueryResult struct {
	Rows []Record `json:"rows"`
}

type WritePreview struct {
	RecordsStaged int      `json:"records_staged"`
	Action        string   `json:"action"`
	Warnings      []string `json:"warnings,omitempty"`
}

type CDCReadRequest struct {
	Stream string
	Config RuntimeConfig
	State  map[string]string
}

type CDCEvent struct {
	Operation string `json:"operation"`
	Record    Record `json:"record"`
	State     Record `json:"state,omitempty"`
}

type WriteValidator interface {
	ValidateWrite(ctx context.Context, req WriteRequest, records []Record) error
}

type DryRunWriter interface {
	DryRunWrite(ctx context.Context, req WriteRequest, records []Record) (WritePreview, error)
}

type Querier interface {
	Query(ctx context.Context, req QueryRequest) (QueryResult, error)
}

type CDCReader interface {
	ReadCDC(ctx context.Context, req CDCReadRequest, emit func(CDCEvent) error) error
}

type StatefulReader interface {
	InitialState(ctx context.Context, stream string, cfg RuntimeConfig) (map[string]string, error)
}

type SchemaMapper interface {
	MapSchema(ctx context.Context, stream Stream) (Stream, error)
}

type LiveConformanceProvider interface {
	LiveConformanceConfig(ctx context.Context) (RuntimeConfig, bool, error)
}

type Connector interface {
	Name() string
	Metadata() Metadata
	Check(ctx context.Context, cfg RuntimeConfig) error
	Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error)
	Read(ctx context.Context, req ReadRequest, emit func(Record) error) error
	Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error)
}

type LocalWarehouseMaterializer interface {
	MaterializesLocalWarehouse() bool
}

type Registry struct {
	connectors map[string]Connector
}

func NewEmptyRegistry() *Registry {
	return &Registry{connectors: make(map[string]Connector)}
}

func NewRegistry() *Registry {
	if builder := registeredDefaultRegistryBuilder(); builder != nil {
		return builder()
	}
	r := NewEmptyRegistry()
	r.RegisterBuiltins()
	return r
}

// RegisterBuiltins adds the primitive local connectors that are implemented in
// this package rather than in defs/. They are not legacy per-connector packages.
func (r *Registry) RegisterBuiltins() {
	r.Register(Sample{})
	r.Register(File{})
	r.Register(Warehouse{})
	r.Register(Outbox{})
}

func (r *Registry) Register(c Connector) {
	r.connectors[c.Name()] = c
}

func (r *Registry) Get(name string) (Connector, bool) {
	c, ok := r.connectors[name]
	return c, ok
}

func (r *Registry) List() []Metadata {
	out := make([]Metadata, 0, len(r.connectors))
	for _, connector := range r.connectors {
		out = append(out, MetadataWithIcon(connector.Metadata()))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) CatalogEntries() []Definition {
	list := r.List()
	out := make([]Definition, 0, len(list))
	for _, item := range list {
		connector, ok := r.Get(item.Name)
		if !ok {
			continue
		}
		def, ok := DefinitionOf(connector)
		if !ok {
			manifest := ManifestOf(connector)
			def = Definition{
				Name:            manifest.Metadata.Name,
				DisplayName:     manifest.Metadata.DisplayName,
				Description:     manifest.Metadata.Description,
				IntegrationType: manifest.Metadata.IntegrationType,
				Capabilities:    manifest.Metadata.Capabilities,
				Streams:         streamSummariesFromManifest(manifest),
				WriteActions:    writeActionInfosFromManifest(manifest),
				Risk:            manifest.Risk,
			}
		}
		def.Icon = MetadataWithIcon(connector.Metadata()).Icon
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type Sample struct{}

func (Sample) Name() string { return "sample" }

func (Sample) Metadata() Metadata {
	return Metadata{
		Name:            "sample",
		DisplayName:     "Sample",
		IntegrationType: "api",
		Description:     "Built-in deterministic source connector for local development and tests.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (Sample) Check(ctx context.Context, cfg RuntimeConfig) error {
	return ctx.Err()
}

func (s Sample) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := ctx.Err(); err != nil {
		return Catalog{}, err
	}
	return Catalog{Connector: s.Name(), Streams: []Stream{
		{
			Name:        "customers",
			Description: "Sample customer records.",
			PrimaryKey:  []string{"id"},
			CursorFields: []string{
				"updated_at",
			},
			Fields: []Field{
				{Name: "id", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "email", Type: "string"},
				{Name: "plan", Type: "string"},
				{Name: "updated_at", Type: "timestamp"},
			},
		},
		{
			Name:         "events",
			Description:  "Sample event records.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"occurred_at"},
			Fields: []Field{
				{Name: "id", Type: "string"},
				{Name: "customer_id", Type: "string"},
				{Name: "event", Type: "string"},
				{Name: "occurred_at", Type: "timestamp"},
			},
		},
	}}, nil
}

func (Sample) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	var records []Record
	switch req.Stream {
	case "customers", "":
		records = []Record{
			{"id": "cus_001", "name": "Ada Lovelace", "email": "ada@example.com", "plan": "enterprise", "updated_at": "2026-06-20T10:00:00Z"},
			{"id": "cus_002", "name": "Grace Hopper", "email": "grace@example.com", "plan": "team", "updated_at": "2026-06-21T12:30:00Z"},
			{"id": "cus_003", "name": "Katherine Johnson", "email": "katherine@example.com", "plan": "starter", "updated_at": "2026-06-22T09:15:00Z"},
		}
	case "events":
		records = []Record{
			{"id": "evt_001", "customer_id": "cus_001", "event": "signed_in", "occurred_at": "2026-06-22T10:00:00Z"},
			{"id": "evt_002", "customer_id": "cus_002", "event": "upgraded", "occurred_at": "2026-06-22T11:00:00Z"},
		}
	default:
		return fmt.Errorf("sample stream %q not found", req.Stream)
	}
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(copyRecord(record)); err != nil {
			return err
		}
	}
	return nil
}

func (Sample) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	return WriteResult{}, ErrUnsupportedOperation
}

type File struct{}

func (File) Name() string { return "file" }

func (File) Metadata() Metadata {
	return Metadata{
		Name:            "file",
		DisplayName:     "File",
		IntegrationType: "file",
		Description:     "Reads local JSONL or CSV files as source streams.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (File) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := cfg.Config["path"]
	if path == "" {
		return errors.New("file connector requires config path")
	}
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file source %s: %w", path, err)
	}
	return nil
}

func (f File) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := f.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	stream := cfg.Config["stream"]
	if stream == "" {
		stream = strings.TrimSuffix(filepath.Base(cfg.Config["path"]), filepath.Ext(cfg.Config["path"]))
	}
	return Catalog{Connector: f.Name(), Streams: []Stream{{
		Name:        stream,
		Description: "Local file stream.",
	}}}, nil
}

func (File) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	path := req.Config.Config["path"]
	if path == "" {
		return errors.New("file connector requires config path")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file source %s: %w", path, err)
	}
	defer file.Close()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv":
		return readCSV(ctx, file, emit)
	default:
		return readJSONL(ctx, file, emit)
	}
}

func (File) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	return WriteResult{}, ErrUnsupportedOperation
}

type Warehouse struct{}

func (Warehouse) Name() string { return "warehouse" }

func (Warehouse) MaterializesLocalWarehouse() bool { return true }

func (Warehouse) Metadata() Metadata {
	return Metadata{
		Name:            "warehouse",
		DisplayName:     "Local Warehouse",
		IntegrationType: "database",
		Description:     "Local JSONL warehouse destination used by the dependency-free MVP.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Read: true, Write: true, Query: true},
	}
}

func (Warehouse) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := warehousePath(cfg)
	if err := validateLocalWriteEffect(cfg, dir, "warehouse path"); err != nil {
		return err
	}
	effects, err := openLocalWriteFS(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = effects.Close() }()
	return effects.MkdirAll(dir, 0o700)
}

func (w Warehouse) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := w.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	entries, err := os.ReadDir(warehousePath(cfg))
	if err != nil {
		return Catalog{}, fmt.Errorf("read warehouse directory: %w", err)
	}
	streams := make([]Stream, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".jsonl")
		streams = append(streams, Stream{Name: name, Description: "Warehouse table " + name})
	}
	sort.Slice(streams, func(i, j int) bool { return streams[i].Name < streams[j].Name })
	return Catalog{Connector: w.Name(), Streams: streams}, nil
}

func (Warehouse) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	path := tablePath(warehousePath(req.Config), req.Stream)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open warehouse table %s: %w", req.Stream, err)
	}
	defer file.Close()
	return readJSONL(ctx, file, emit)
}

func (Warehouse) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return WriteResult{}, err
	}
	dir := warehousePath(req.Config)
	if err := validateLocalWriteEffect(req.Config, dir, "warehouse path"); err != nil {
		return WriteResult{}, err
	}
	effects, err := openLocalWriteFS(req.Config)
	if err != nil {
		return WriteResult{}, err
	}
	defer func() { _ = effects.Close() }()
	if err := effects.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{}, fmt.Errorf("create warehouse directory: %w", err)
	}
	table := req.Table
	if table == "" {
		table = req.Stream
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if req.Overwrite {
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	file, err := effects.OpenFile(tablePath(dir, table), flag, 0o600)
	if err != nil {
		return WriteResult{}, fmt.Errorf("open warehouse table %s: %w", table, err)
	}
	defer file.Close()
	n, err := writeJSONL(ctx, file, records)
	if err != nil {
		return WriteResult{}, err
	}
	return WriteResult{RecordsWritten: n}, nil
}

type Outbox struct{}

func (Outbox) Name() string { return "outbox" }

func (Outbox) Metadata() Metadata {
	return Metadata{
		Name:            "outbox",
		DisplayName:     "Local Outbox",
		IntegrationType: "api",
		Description:     "Local JSONL destination that records reverse ETL writes and receipts.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Write: true},
	}
}

func (Outbox) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := outboxPath(cfg)
	if err := validateLocalWriteEffect(cfg, dir, "outbox path"); err != nil {
		return err
	}
	effects, err := openLocalWriteFS(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = effects.Close() }()
	return effects.MkdirAll(dir, 0o700)
}

func (o Outbox) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := o.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	return Catalog{Connector: o.Name(), Streams: []Stream{{Name: "records", Description: "Reverse ETL outbox records."}}}, nil
}

func (Outbox) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	return ErrUnsupportedOperation
}

func (Outbox) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return WriteResult{}, err
	}
	dir := outboxPath(req.Config)
	if err := validateLocalWriteEffect(req.Config, dir, "outbox path"); err != nil {
		return WriteResult{}, err
	}
	effects, err := openLocalWriteFS(req.Config)
	if err != nil {
		return WriteResult{}, err
	}
	defer func() { _ = effects.Close() }()
	if err := effects.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{}, fmt.Errorf("create outbox directory: %w", err)
	}
	name := req.Table
	if name == "" {
		name = "records"
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	enriched := make([]Record, 0, len(records))
	for _, record := range records {
		r := copyRecord(record)
		r["_outbox_action"] = req.Action
		r["_outbox_written_at"] = now
		enriched = append(enriched, r)
	}
	file, err := effects.OpenFile(tablePath(dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return WriteResult{}, fmt.Errorf("open outbox %s: %w", name, err)
	}
	defer file.Close()
	n, err := writeJSONL(ctx, file, enriched)
	if err != nil {
		return WriteResult{}, err
	}
	return WriteResult{RecordsWritten: n}, nil
}

func readJSONL(ctx context.Context, r io.Reader, emit func(Record) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("decode jsonl record: %w", err)
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan jsonl: %w", err)
	}
	return nil
}

func readCSV(ctx context.Context, r io.Reader, emit func(Record) error) error {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read csv header: %w", err)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read csv row: %w", err)
		}
		record := make(Record, len(header))
		for i, name := range header {
			if i < len(row) {
				record[name] = row[i]
			}
		}
		if err := emit(record); err != nil {
			return err
		}
	}
}

func writeJSONL(ctx context.Context, w io.Writer, records []Record) (int, error) {
	enc := json.NewEncoder(w)
	for i, record := range records {
		if err := ctx.Err(); err != nil {
			return i, err
		}
		if err := enc.Encode(record); err != nil {
			return i, fmt.Errorf("encode jsonl record: %w", err)
		}
	}
	return len(records), nil
}

func validateLocalWriteEffect(cfg RuntimeConfig, path, field string) error {
	if cfg.LocalWritePolicy == nil {
		return nil
	}
	return safety.ValidateLocalWritePath(
		cfg.LocalWritePolicy.ProjectRoot,
		path,
		field,
		cfg.LocalWritePolicy.AllowExternal,
	)
}

func openLocalWriteFS(cfg RuntimeConfig) (*safety.LocalWriteFS, error) {
	if cfg.LocalWritePolicy == nil {
		return safety.OpenLocalWriteFS("", true)
	}
	return safety.OpenLocalWriteFS(
		cfg.LocalWritePolicy.ProjectRoot,
		cfg.LocalWritePolicy.AllowExternal,
	)
}

func warehousePath(cfg RuntimeConfig) string {
	if cfg.Config["path"] != "" {
		return cfg.Config["path"]
	}
	return filepath.Join(cfg.ProjectDir, "warehouse")
}

func outboxPath(cfg RuntimeConfig) string {
	if cfg.Config["path"] != "" {
		return cfg.Config["path"]
	}
	return filepath.Join(cfg.ProjectDir, "outbox")
}

func tablePath(dir, table string) string {
	return filepath.Join(dir, safeName(table)+".jsonl")
}

func safeName(name string) string {
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
		case r == '.' || r == '/' || r == ' ':
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "records"
	}
	return b.String()
}

func copyRecord(in Record) Record {
	out := make(Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
