package connectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const nativeFixtureTimestamp = "2026-01-01T00:00:00Z"

type NativeCatalogConnector struct {
	def ConnectorDefinition
}

func NewNativeCatalogConnector(def ConnectorDefinition) NativeCatalogConnector {
	return NativeCatalogConnector{def: normalizeConnectorDefinition(def)}
}

func (c NativeCatalogConnector) Name() string { return c.def.Slug }

func (c NativeCatalogConnector) Metadata() Metadata {
	return MetadataWithIcon(Metadata{
		Name:            c.def.Slug,
		DisplayName:     c.def.Name,
		IntegrationType: nativeIntegrationType(c.def),
		Description:     nativeDescription(c.def),
		Capabilities:    c.def.RuntimeCapabilities.toCapabilities(),
	})
}

func (c NativeCatalogConnector) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(c.def.Slug) == "" {
		return errors.New("native catalog connector missing slug")
	}
	if !nativeFixtureMode(cfg) && !c.def.RuntimeCapabilities.Check {
		return fmt.Errorf("%w: connector %s has implementation_status=%s", ErrUnsupportedOperation, c.def.Slug, c.def.ImplementationStatus)
	}
	return nil
}

func (c NativeCatalogConnector) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := c.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	if !nativeFixtureMode(cfg) && !c.def.RuntimeCapabilities.Catalog {
		return Catalog{}, fmt.Errorf("%w: connector %s catalog is not enabled", ErrUnsupportedOperation, c.def.Slug)
	}
	return Catalog{Connector: c.Name(), Streams: []Stream{c.fixtureStream()}}, nil
}

func (c NativeCatalogConnector) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !nativeFixtureMode(req.Config) && !c.def.RuntimeCapabilities.Read {
		return ErrUnsupportedOperation
	}
	stream := req.Stream
	if stream == "" {
		stream = c.fixtureStream().Name
	}
	record := c.fixtureRecord(stream)
	if cursor := req.State["cursor"]; cursor != "" {
		record["previous_cursor"] = cursor
	}
	return emit(record)
}

func (c NativeCatalogConnector) ValidateWrite(ctx context.Context, req WriteRequest, records []Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !nativeFixtureMode(req.Config) && !c.def.RuntimeCapabilities.Write {
		return ErrUnsupportedOperation
	}
	action := strings.TrimSpace(req.Action)
	switch action {
	case "", "append", "overwrite", "upsert", "dedup", "write", "fixture_write":
	default:
		return fmt.Errorf("native connector %s action %q is not in the approved write fixture action set", c.def.Slug, action)
	}
	return nil
}

func (c NativeCatalogConnector) DryRunWrite(ctx context.Context, req WriteRequest, records []Record) (WritePreview, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return WritePreview{}, err
	}
	action := req.Action
	if action == "" {
		action = "upsert"
	}
	return WritePreview{
		RecordsStaged: len(records),
		Action:        action,
		Warnings:      []string{"fixture-backed native write; no external mutation is executed without connector-specific action support"},
	}, nil
}

func (c NativeCatalogConnector) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return WriteResult{RecordsFailed: len(records)}, err
	}
	if err := c.writeReceiptFile(ctx, req, records); err != nil {
		return WriteResult{RecordsFailed: len(records)}, err
	}
	return WriteResult{RecordsWritten: len(records)}, nil
}

func (c NativeCatalogConnector) Query(ctx context.Context, req QueryRequest) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return QueryResult{}, err
	}
	if !nativeFixtureMode(req.Config) && !c.def.RuntimeCapabilities.Query {
		return QueryResult{}, ErrUnsupportedOperation
	}
	if err := validateNativeSelect(req.SQL); err != nil {
		return QueryResult{}, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 1
	}
	rows := make([]Record, 0, limit)
	for i := 0; i < limit; i++ {
		record := c.fixtureRecord("query")
		record["id"] = fmt.Sprintf("%s:query:%d", c.def.Slug, i+1)
		rows = append(rows, record)
	}
	return QueryResult{Rows: rows}, nil
}

func (c NativeCatalogConnector) ReadCDC(ctx context.Context, req CDCReadRequest, emit func(CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	cdc := CDCPlanForDefinition(c.def)
	if !cdc.Supported {
		return ErrUnsupportedOperation
	}
	state := Record{}
	for _, field := range cdc.StateFields {
		state[field] = c.def.Slug + ":fixture"
	}
	return emit(CDCEvent{Operation: "snapshot", Record: c.fixtureRecord(req.Stream), State: state})
}

func (c NativeCatalogConnector) InitialState(ctx context.Context, stream string, cfg RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]string{
		"cursor":             "",
		"snapshot_completed": "false",
		"stream":             stream,
	}, nil
}

func (c NativeCatalogConnector) MapSchema(ctx context.Context, stream Stream) (Stream, error) {
	if err := ctx.Err(); err != nil {
		return Stream{}, err
	}
	if stream.Name == "" {
		stream.Name = c.fixtureStream().Name
	}
	if len(stream.Fields) == 0 {
		stream.Fields = c.fixtureStream().Fields
	}
	return stream, nil
}

func (c NativeCatalogConnector) Manifest() Manifest {
	catalog, _ := c.Catalog(context.Background(), RuntimeConfig{})
	manifest := Manifest{
		Metadata:             c.Metadata(),
		ConfigFields:         nativeConfigFields(c.def),
		SecretFields:         nativeSecretFields(c.def),
		AuthModes:            nativeAuthModes(c.def),
		Streams:              catalog.Streams,
		WriteActions:         nativeWriteActions(c.def),
		SyncModes:            allSyncModes(),
		SourceSyncModes:      readSourceSyncModes(),
		DestinationSyncModes: warehouseDestinationSyncModes(),
		Pagination: PaginationSpec{
			Type:           nativePaginationType(c.def),
			PageSizeField:  "page_size",
			PageLimitField: "max_pages",
			DelayField:     "page_delay",
			DefaultLimit:   "1",
		},
		Risk: RiskSpec{
			Read:     "native fixture-backed connector read",
			Write:    "native receipt-backed connector write",
			Mutation: "generic native writes produce local receipts unless a connector-specific adapter implements external mutations",
			Approval: "reverse ETL plan approval required before writes",
		},
	}
	if c.def.Type != ConnectorTypeSource {
		manifest.SourceSyncModes = nil
	}
	if c.def.Type != ConnectorTypeDestination {
		manifest.DestinationSyncModes = nil
	}
	return manifest
}

func (c NativeCatalogConnector) Guide() ConnectorGuide {
	manifest := c.Manifest()
	sections := []GuideSection{
		capabilitySection(manifest),
		implementationSection(c.def),
		runtimeCapabilitiesSection(c.def),
		nativePortPlanSection(c.def),
		officialApplicationDocsSection(c.def),
		authSection(manifest),
		configSection(manifest),
		streamSection(manifest),
		syncModeSection(manifest),
		writeActionSection(manifest),
		paginationSection(manifest),
		securitySection(manifest),
	}
	return ConnectorGuide{
		Name:        c.def.Slug,
		DisplayName: c.def.Name,
		Summary:     nativeDescription(c.def),
		Sections:    compactSections(sections),
		Examples: []GuideExample{
			{Title: "Inspect connector", Command: "pm connectors inspect " + c.def.Slug},
			{Title: "Inspect connector as JSON", Command: "pm connectors inspect " + c.def.Slug + " --json"},
			{Title: "Add credential", Command: "pm credentials add " + c.def.Slug + "-local --connector " + c.def.Slug + " --config mode=fixture"},
		},
		Links: nativeGuideLinks(c.def),
		AgentNotes: []string{
			"Use --json for machine-readable discovery and run results.",
			"Never ask for secret values in chat; use --from-env or --value-stdin for secret fields.",
			"Generic native bindings are fixture-backed unless a connector-specific adapter documents live external mutations.",
			"Reverse ETL writes require plan, preview, approval, and receipts.",
		},
	}
}

func (c NativeCatalogConnector) fixtureStream() Stream {
	name := "records"
	if c.def.Type == ConnectorTypeSource {
		name = "records"
	}
	if c.def.SourceType == "database" {
		name = "tables"
	}
	if c.def.SourceType == "file" {
		name = "files"
	}
	return Stream{
		Name:         name,
		Description:  "Native fixture stream for " + c.def.Name + ".",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"emitted_at"},
		Fields: []Field{
			{Name: "id", Type: "string"},
			{Name: "connector_slug", Type: "string"},
			{Name: "connector_name", Type: "string"},
			{Name: "runtime_kind", Type: "string"},
			{Name: "emitted_at", Type: "timestamp"},
			{Name: "payload", Type: "object"},
		},
	}
}

func (c NativeCatalogConnector) fixtureRecord(stream string) Record {
	if stream == "" {
		stream = c.fixtureStream().Name
	}
	return Record{
		"id":              c.def.Slug + ":fixture:1",
		"connector_slug":  c.def.Slug,
		"connector_name":  c.def.Name,
		"connector_type":  string(c.def.Type),
		"runtime_kind":    string(c.def.RuntimeKind),
		"stream":          stream,
		"release_stage":   c.def.ReleaseStage,
		"support_level":   c.def.SupportLevel,
		"source_type":     c.def.SourceType,
		"emitted_at":      nativeFixtureTimestamp,
		"fixture_backed":  true,
		"payload":         Record{"documentation_url": c.def.DocumentationURL},
		"_polymetrics_id": c.def.Slug + ":fixture:1",
	}
}

func (c NativeCatalogConnector) writeReceiptFile(ctx context.Context, req WriteRequest, records []Record) error {
	if len(records) == 0 || req.Config.ProjectDir == "" {
		return nil
	}
	table := req.Table
	if table == "" {
		table = req.Stream
	}
	if table == "" {
		table = "records"
	}
	dir := filepath.Join(req.Config.ProjectDir, "native", c.def.Slug)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create native receipt directory: %w", err)
	}
	path := filepath.Join(dir, table+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open native receipt file: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		receipt := Record{
			"connector_slug": c.def.Slug,
			"action":         valueOrDefault(req.Action, "upsert"),
			"record":         copyRecord(record),
		}
		if err := encoder.Encode(receipt); err != nil {
			return fmt.Errorf("write native receipt: %w", err)
		}
	}
	return nil
}

func nativeIntegrationType(def ConnectorDefinition) string {
	switch def.RuntimeKind {
	case RuntimeDatabaseGo:
		return "database"
	case RuntimeFileGo:
		return "file"
	case RuntimeDestinationGo:
		return "destination"
	default:
		if def.Type == ConnectorTypeDestination {
			return "destination"
		}
		return "api"
	}
}

func nativeDescription(def ConnectorDefinition) string {
	return fmt.Sprintf("%s native Go %s connector. Runtime family: %s.", def.Name, def.Type, def.RuntimeKind)
}

func nativeConfigFields(def ConnectorDefinition) []ConfigField {
	fields := make([]ConfigField, 0, len(def.ConfigFields)+2)
	fields = append(fields,
		ConfigField{Name: "mode", Description: "Runtime mode: fixture or live.", Default: "fixture"},
		ConfigField{Name: "max_pages", Description: "Maximum pages for paginated native reads.", Default: "1"},
	)
	for _, field := range def.ConfigFields {
		if field.Secret {
			continue
		}
		fields = append(fields, ConfigField{Name: field.Name, Description: compactPublicDescription(field.Description), Required: field.Required})
	}
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

func nativeSecretFields(def ConnectorDefinition) []SecretField {
	fields := make([]SecretField, 0, len(def.SecretFields)+len(def.ConfigFields))
	seen := map[string]bool{}
	for _, name := range def.SecretFields {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		fields = append(fields, SecretField{Name: name, Description: "Secret field from connector schema."})
	}
	for _, field := range def.ConfigFields {
		if !field.Secret || field.Name == "" || seen[field.Name] {
			continue
		}
		seen[field.Name] = true
		fields = append(fields, SecretField{Name: field.Name, Description: compactPublicDescription(field.Description), Required: field.Required})
	}
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

func nativeAuthModes(def ConnectorDefinition) []AuthModeSpec {
	secrets := def.SecretFields
	if len(secrets) == 0 {
		return []AuthModeSpec{{Name: "fixture", Description: "Fixture-backed native conformance mode.", Read: true, Write: def.Type == ConnectorTypeDestination}}
	}
	return []AuthModeSpec{
		{Name: "fixture", Description: "Fixture-backed native conformance mode.", Read: true, Write: def.Type == ConnectorTypeDestination},
		{Name: "secret_fields", Description: "Live mode uses connector-specific secret fields through the pm vault.", SecretFields: append([]string(nil), secrets...), Read: true, Write: def.Type == ConnectorTypeDestination},
	}
}

func nativeWriteActions(def ConnectorDefinition) []WriteActionSpec {
	if !RuntimeCapabilitiesForDefinition(def).Write {
		return nil
	}
	return []WriteActionSpec{
		{Name: "append", Description: "Append records through the native destination writer.", RequiredFields: []string{"id"}, Risk: "local receipt or connector-specific external write"},
		{Name: "overwrite", Description: "Replace target records through the native destination writer.", RequiredFields: []string{"id"}, Risk: "destructive; approval required"},
		{Name: "upsert", Description: "Deduplicate by primary key and update existing records.", RequiredFields: []string{"id"}, Risk: "mutation; approval required"},
		{Name: "dedup", Description: "Write latest record per primary key.", RequiredFields: []string{"id"}, Risk: "mutation; approval required"},
	}
}

func nativeFixtureMode(cfg RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func nativePaginationType(def ConnectorDefinition) string {
	switch def.RuntimeKind {
	case RuntimeDeclarativeHTTPGo:
		return "manifest"
	case RuntimeDatabaseGo:
		return "cursor"
	case RuntimeFileGo:
		return "listing"
	default:
		return "runtime_family"
	}
}

func nativeGuideLinks(def ConnectorDefinition) []GuideLink {
	return ApplicationDocumentationLinks(def)
}

func validateNativeSelect(sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil
	}
	lower := strings.ToLower(sql)
	if !strings.HasPrefix(lower, "select ") {
		return errors.New("native query supports SELECT-only statements")
	}
	for _, forbidden := range []string{";", " insert ", " update ", " delete ", " drop ", " alter ", " truncate ", " create "} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("native query rejected unsafe token %q", strings.TrimSpace(forbidden))
		}
	}
	return nil
}
