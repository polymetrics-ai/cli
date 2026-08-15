package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	postgresSnapshotTransportID      = "postgres_bounded_snapshot"
	postgresSnapshotTransportStream  = "snapshot"
	postgresSnapshotCheckpointKind   = "postgres_repeatable_read_snapshot"
	postgresSnapshotMechanism        = "postgres_bounded_full_snapshot"
	postgresSnapshotProtocolVersion  = "postgres_snapshot_v1"
	postgresSnapshotDedupeKind       = "postgres_snapshot_page"
	postgresSnapshotDedupeWindowKind = "postgres_repeatable_read_window"
	postgresSnapshotSourceEngine     = "postgres"
	postgresSnapshotConformanceSuite = "postgres_snapshot"
	postgresSnapshotConformanceRunID = "bounded_full_and_bootstrap_v1"
)

var postgresSnapshotTransportReference = connectors.TransportExecutorReference{
	Family: connectors.TransportExecutorFamilyNativeDatabase,
	ID:     postgresSnapshotTransportID,
}

// SnapshotTransportSource is the PostgreSQL-specific source side of the
// closed transport contract. It has no generic SQL input: both relation and
// ordering come from the typed catalog and the checkpoint source identity.
type SnapshotTransportSource struct {
	connector Connector
}

// NewSnapshotTransportSource creates a bounded PostgreSQL full-snapshot
// source. Callers register it explicitly with RegisterSnapshotTransportSource.
func NewSnapshotTransportSource(connector Connector) *SnapshotTransportSource {
	return &SnapshotTransportSource{connector: connector}
}

// SnapshotTransportDefinitionFactory supplies the PostgreSQL-local adapter for
// the exact executor declared by defs/postgres/sync_transport.json. The
// conformance reference is an external composition allow-list; constructing a
// descriptor does not certify itself or register the adapter.
func SnapshotTransportDefinitionFactory() synctransport.DefinitionFactory {
	return synctransport.DefinitionFactory{
		Reference: postgresSnapshotTransportReference,
		SourceEvidence: connectors.ConformanceEvidenceReference{
			Suite: postgresSnapshotConformanceSuite,
			RunID: postgresSnapshotConformanceRunID,
		},
		BuildSource: func(connector connectors.Connector) (synctransport.SourceExecutor, error) {
			switch typed := connector.(type) {
			case Connector:
				return NewSnapshotTransportSource(typed), nil
			case *Connector:
				if typed == nil {
					return nil, errors.New("postgres snapshot transport connector is required")
				}
				return NewSnapshotTransportSource(*typed), nil
			default:
				return nil, errors.New("postgres snapshot transport requires the native postgres connector")
			}
		},
	}
}

// SyncTransportDefinitionFactories exposes the PostgreSQL-local adapter to
// generic production composition. The bundle still decides whether the source
// role is declared and therefore whether this factory is used.
func (Connector) SyncTransportDefinitionFactories() []synctransport.DefinitionFactory {
	return []synctransport.DefinitionFactory{SnapshotTransportDefinitionFactory(), ManagedTargetTransportDefinitionFactory()}
}

// RegisterSnapshotTransportSource registers the one executor selected by the
// PostgreSQL Definition. Keeping the adapter in this native package prevents
// App composition from assembling a connector-specific transport pair.
func RegisterSnapshotTransportSource(registry *synctransport.Registry, connector Connector) error {
	if registry == nil {
		return errors.New("postgres snapshot transport registry is required")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(connector)
	if !ok || descriptor.Executor != postgresSnapshotTransportReference {
		return errors.New("postgres snapshot transport definition is unavailable")
	}
	return registry.RegisterSource(NewSnapshotTransportSource(connector))
}

func (*SnapshotTransportSource) TransportExecutorReference() connectors.TransportExecutorReference {
	return postgresSnapshotTransportReference
}

// AllowEmptySourceResult admits an empty relation or an incremental poll with
// no rows after the acknowledged tuple. Neither case has a new source
// position, so manufacturing a checkpoint would make resume unsafe.
func (*SnapshotTransportSource) AllowEmptySourceResult() {}

// ReadTransport executes one exact, bounded full snapshot in a read-only
// repeatable-read transaction. Catalog discovery and record queries share that
// transaction, so the emitted catalog fingerprint, stable key order, rows, and
// PostgreSQL snapshot token describe the same source observation.
func (s *SnapshotTransportSource) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) (err error) {
	return executeWithAuthenticationAdmission(ctx, request.Runtime, func(admitted context.Context) error {
		return s.readTransport(admitted, request, emit)
	})
}

func (s *SnapshotTransportSource) readTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) (err error) {
	if ctx == nil {
		return errors.New("postgres snapshot transport context is required")
	}
	if emit == nil {
		return errors.New("postgres snapshot transport emit function is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if request.Mode == synccontract.ModeIncrementalUpsert && strings.EqualFold(strings.TrimSpace(request.Runtime.Config["transport_bootstrap"]), "true") {
		return s.readBootstrapTransport(ctx, request, emit)
	}
	if request.Mode == synccontract.ModeIncrementalUpsert && strings.TrimSpace(request.CursorField) != "" {
		return s.readPollingTransport(ctx, request, emit)
	}

	conn, relationRef, resources, pageSize, err := s.validateReadRequest(request)
	if err != nil {
		return err
	}
	operationCtx, cancel, err := resources.operationContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()

	pool, err := conn.openTypedCatalogPool(operationCtx, resources)
	if err != nil {
		return fmt.Errorf("postgres snapshot transport: open pool: %w", err)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(operationCtx, typedCatalogTransactionOptions())
	if err != nil {
		return fmt.Errorf("postgres snapshot transport: begin snapshot: %w", err)
	}
	defer func() {
		rollbackErr := rollbackTypedCatalogSnapshot(operationCtx, tx, resources)
		if rollbackErr == nil {
			return
		}
		if err == nil {
			err = rollbackErr
			return
		}
		err = errors.Join(err, rollbackErr)
	}()

	var snapshotToken string
	if err := tx.QueryRow(operationCtx, "SELECT txid_current_snapshot()").Scan(&snapshotToken); err != nil {
		return fmt.Errorf("postgres snapshot transport: observe snapshot: %w", err)
	}
	if strings.TrimSpace(snapshotToken) == "" {
		return errors.New("postgres snapshot transport: PostgreSQL returned an empty snapshot token")
	}

	catalog, err := discoverTypedCatalogSnapshot(operationCtx, tx, conn.database, conn.schema, s.connector.databaseDefinition)
	if err != nil {
		return fmt.Errorf("postgres snapshot transport: discover typed catalog: %w", err)
	}
	plan, err := newPostgresSnapshotReadPlan(catalog, relationRef, pageSize)
	if err != nil {
		return err
	}

	var after []any
	for page := 0; ; page++ {
		records, nextAfter, err := plan.readPage(operationCtx, tx, after)
		if err != nil {
			return err
		}
		candidate, err := postgresSnapshotCheckpoint(request.Resume, snapshotToken, plan.fingerprint, page)
		if err != nil {
			return err
		}
		if err := emit(synctransport.SourcePage{Records: records, Tombstones: []synccontract.Tombstone{}, CandidateCheckpoint: candidate}); err != nil {
			return err
		}
		if len(records) < plan.pageSize {
			return nil
		}
		after = nextAfter
	}
}

type bootstrapTransportCommitter struct {
	resume              synccontract.ResumeExpectation
	primaryKey          []string
	emit                func(synctransport.SourcePage) error
	finalSnapshot       *synccontract.CheckpointEnvelope
	initialCommitted    bool
	pendingRecords      []connectors.Record
	pendingTombstones   []synccontract.Tombstone
	pendingEventOrdinal uint64
}

func (c *bootstrapTransportCommitter) CommitDurableChangefeedCheckpoint(ctx context.Context, candidate synccontract.CheckpointEnvelope) error {
	translated := bootstrapTransportCheckpoint(candidate, c.resume)
	if err := translated.Validate(); err != nil {
		return err
	}
	if translated.Source != c.resume.Source || string(translated.SourceGeneration) != string(c.resume.SourceGeneration) {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeSourceGenerationChanged, "PostgreSQL bootstrap identity changed before its warehouse checkpoint commit")
	}
	if !c.initialCommitted {
		if c.finalSnapshot == nil || c.finalSnapshot.SnapshotBarrier == nil || translated.SnapshotBarrier == nil || string(c.finalSnapshot.SnapshotBarrier.Token) != string(translated.SnapshotBarrier.Token) {
			return errors.New("postgres bootstrap final snapshot page was not durably committed")
		}
		c.initialCommitted = true
		return nil
	}
	if c.emit == nil {
		return errors.New("postgres bootstrap transport page receiver is unavailable")
	}
	page := synctransport.SourcePage{
		Records:             append([]connectors.Record(nil), c.pendingRecords...),
		Tombstones:          append([]synccontract.Tombstone(nil), c.pendingTombstones...),
		CandidateCheckpoint: translated,
	}
	if err := c.emit(page); err != nil {
		return err
	}
	c.pendingRecords = nil
	c.pendingTombstones = nil
	c.pendingEventOrdinal = 0
	return nil
}

func (s *SnapshotTransportSource) readBootstrapTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	committer := &bootstrapTransportCommitter{
		resume: request.Resume, primaryKey: append([]string(nil), request.PrimaryKey...), emit: emit,
	}
	if request.Checkpoint != nil {
		nativeCheckpoint, err := nativeBootstrapTransportCheckpoint(*request.Checkpoint, request.Runtime)
		if err != nil {
			return err
		}
		committer.initialCommitted = true
		return s.connector.ReadCDC(ctx, connectors.CDCReadRequest{
			Stream: request.Stream, Config: request.Runtime, Checkpoint: &nativeCheckpoint,
			DurableCheckpointCommitter: committer,
		}, committer.emitChange)
	}
	return s.connector.BootstrapCDC(ctx, BootstrapCDCRequest{
		Stream: request.Stream, Config: request.Runtime, BatchSize: request.BatchSize,
		DurableCheckpointCommitter: committer,
		Snapshot: func(_ context.Context, page BootstrapSnapshotPage) error {
			candidate := bootstrapTransportCheckpoint(page.CandidateCheckpoint, request.Resume)
			if page.Final {
				copy := candidate.Clone()
				committer.finalSnapshot = &copy
			}
			return committer.emit(synctransport.SourcePage{
				Records: page.Records, Tombstones: []synccontract.Tombstone{}, CandidateCheckpoint: candidate,
				DeferCheckpoint: !page.Final,
			})
		},
	}, committer.emitChange)
}

func (c *bootstrapTransportCommitter) emitChange(event connectors.CDCEvent) error {
	c.pendingEventOrdinal++
	switch event.Operation {
	case "insert", "update":
		c.pendingRecords = append(c.pendingRecords, cloneBootstrapRecords([]connectors.Record{event.Record})[0])
		return nil
	case "delete":
		tombstone, err := CDCDeleteTombstone(event, c.primaryKey, c.pendingEventOrdinal)
		if err != nil {
			return err
		}
		c.pendingTombstones = append(c.pendingTombstones, tombstone)
		return nil
	default:
		return fmt.Errorf("postgres bootstrap transport received unsupported CDC operation %q", event.Operation)
	}
}

func nativeBootstrapTransportCheckpoint(checkpoint synccontract.CheckpointEnvelope, runtime connectors.RuntimeConfig) (synccontract.CheckpointEnvelope, error) {
	barrier, ok, err := postgresBootstrapBarrierFromCheckpoint(&checkpoint)
	if err != nil || !ok {
		return synccontract.CheckpointEnvelope{}, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL transport checkpoint has no valid bootstrap barrier")
	}
	resolved, err := resolveConfig(runtime)
	if err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	native := checkpoint.Clone()
	native.Source = synccontract.SourceIdentity{
		Engine: "postgres", AccountOrCluster: barrier.SystemID + ":" + resolved.database, ObjectScope: barrier.Relation,
	}
	native.SourceGeneration = synccontract.OpaqueToken([]byte(strconv.FormatInt(int64(barrier.Timeline), 10) + "\n" + barrier.Publication))
	return native, nil
}

func bootstrapTransportCheckpoint(candidate synccontract.CheckpointEnvelope, resume synccontract.ResumeExpectation) synccontract.CheckpointEnvelope {
	translated := candidate.Clone()
	translated.Source = resume.Source
	translated.SourceGeneration = append(synccontract.OpaqueToken(nil), resume.SourceGeneration...)
	return translated
}

func (s *SnapshotTransportSource) validateReadRequest(request synctransport.SourceRequest) (connConfig, database.RelationRef, typedCatalogResources, int, error) {
	if s == nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport source is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != s.connector.Name() || strings.TrimSpace(request.Stream) == "" {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport requires a declared dynamic stream")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(request.Connector)
	if !ok || descriptor.Executor != postgresSnapshotTransportReference {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport requires its declared source descriptor")
	}
	if request.Mode != synccontract.ModeFullAppend && request.Mode != synccontract.ModeIncrementalUpsert {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres transport accepts only full_append or incremental_upsert")
	}
	if request.BatchSize <= 0 {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport requires a positive batch size")
	}
	if request.Checkpoint != nil && request.Mode != synccontract.ModeIncrementalUpsert {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport does not resume a full snapshot checkpoint")
	}
	if request.Checkpoint != nil {
		if err := request.Checkpoint.ValidateResume(request.Resume); err != nil {
			return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, err
		}
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport requires a complete resume identity")
	}
	if request.Resume.Source.Engine != postgresSnapshotSourceEngine {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport source identity engine must be postgres")
	}
	if fixtureMode(request.Runtime) {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errTypedCatalogFixtureMode
	}
	conn, err := resolveConfig(request.Runtime)
	if err != nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, err
	}
	if err := validateIdentifier(conn.schema); err != nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, ErrUnsupportedCatalogShape
	}
	if isSystemCatalogSchema(conn.schema) {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, ErrSystemCatalogSchema
	}
	relationRef, err := postgresSnapshotRelationRef(request.Resume.Source.ObjectScope, conn.database, conn.schema)
	if err != nil || relationRef.Schema.Catalog.Name != conn.database || relationRef.Schema.Name != conn.schema {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport source identity does not match configured relation")
	}
	if err := s.connector.databaseDefinition.Validate(); err != nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport definition is unavailable")
	}
	resources, err := newTypedCatalogResources(s.connector.databaseDefinition.Resources())
	if err != nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, err
	}
	pageSize, err := s.connector.databaseDefinition.Resources().EffectivePageSize(request.BatchSize)
	if err != nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport batch size is outside the declared resource bound")
	}
	return conn, relationRef, resources, pageSize, nil
}

func postgresSnapshotRelationRef(scope, configuredDatabase, configuredSchema string) (database.RelationRef, error) {
	parts := strings.Split(scope, ".")
	switch len(parts) {
	case 1:
		parts = []string{configuredDatabase, configuredSchema, parts[0]}
	case 2:
		parts = []string{configuredDatabase, parts[0], parts[1]}
	case 3:
	default:
		return database.RelationRef{}, errors.New("postgres snapshot transport source scope must be relation, schema.relation, or database.schema.relation")
	}
	if strings.TrimSpace(parts[0]) == "" {
		return database.RelationRef{}, errors.New("postgres snapshot transport source scope requires a database")
	}
	// The configured database/catalog name is compared as identity only; it is
	// never rendered into the PostgreSQL query. Schema and relation are the SQL
	// identifiers and retain the catalog discovery validator's strict grammar.
	for _, part := range parts[1:] {
		if err := validateIdentifier(part); err != nil {
			return database.RelationRef{}, err
		}
	}
	return database.RelationRef{
		Schema: database.SchemaRef{Catalog: database.CatalogRef{Name: parts[0]}, Name: parts[1]},
		Name:   parts[2],
	}, nil
}

// postgresSnapshotReadPlan is the PostgreSQL rendering of the existing typed
// catalog read-plan invariants: selected catalog columns, a non-null declared
// unique key, and a finite resource-policy page size. It deliberately has no
// inbound warehouse reference because synctransport owns the neutral warehouse
// mediator separately from a source executor invocation.
type postgresSnapshotReadPlan struct {
	relation    database.Relation
	columns     []database.Column
	order       []database.ColumnRef
	pageSize    int
	fingerprint database.SchemaFingerprint
}

func newPostgresSnapshotReadPlan(catalog database.Catalog, relationRef database.RelationRef, pageSize int) (postgresSnapshotReadPlan, error) {
	if pageSize <= 0 {
		return postgresSnapshotReadPlan{}, errors.New("postgres snapshot transport requires a finite page size")
	}
	for _, relation := range catalog.Relations() {
		if relation.Ref.Schema.Catalog.Name != relationRef.Schema.Catalog.Name || relation.Ref.Schema.Name != relationRef.Schema.Name || relation.Ref.Name != relationRef.Name {
			continue
		}
		columns := append([]database.Column(nil), relation.Columns...)
		sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
		order, err := postgresSnapshotStableOrder(relation)
		if err != nil {
			return postgresSnapshotReadPlan{}, err
		}
		if err := postgresSnapshotValidatePaginationOrder(order, columns); err != nil {
			return postgresSnapshotReadPlan{}, err
		}
		return postgresSnapshotReadPlan{relation: relation, columns: columns, order: order, pageSize: pageSize, fingerprint: catalog.Fingerprint()}, nil
	}
	return postgresSnapshotReadPlan{}, errors.New("postgres snapshot transport source relation is absent from the typed catalog")
}

func postgresSnapshotStableOrder(relation database.Relation) ([]database.ColumnRef, error) {
	for _, keyKind := range []database.KeyKind{database.KeyPrimary, database.KeyUnique} {
		for _, key := range relation.Keys {
			if key.Kind != keyKind || len(key.Columns) == 0 || !postgresSnapshotKeyColumnsNonNullable(key, relation.Columns) {
				continue
			}
			return append([]database.ColumnRef(nil), key.Columns...), nil
		}
	}
	return nil, errors.New("postgres snapshot transport source relation requires a non-null primary or unique key")
}

func postgresSnapshotKeyColumnsNonNullable(key database.Key, columns []database.Column) bool {
	for _, keyColumn := range key.Columns {
		found := false
		for _, column := range columns {
			if column.Ref.Name != keyColumn.Name {
				continue
			}
			found = true
			if column.Nullable {
				return false
			}
			break
		}
		if !found {
			return false
		}
	}
	return true
}

func postgresSnapshotValidatePaginationOrder(order []database.ColumnRef, columns []database.Column) error {
	for _, keyColumn := range order {
		for _, column := range columns {
			if column.Ref != keyColumn {
				continue
			}
			if err := postgresSnapshotPaginationLogicalType(column.Type); err != nil {
				return fmt.Errorf("postgres snapshot transport: stable key %q: %w", keyColumn.Name, err)
			}
			break
		}
	}
	return nil
}

func (p postgresSnapshotReadPlan) readPage(ctx context.Context, tx pgx.Tx, after []any) ([]connectors.Record, []any, error) {
	rows, err := tx.Query(ctx, p.query(after), p.queryArguments(after)...)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres snapshot transport: query bounded page: %w", err)
	}
	defer rows.Close()

	records := make([]connectors.Record, 0, p.pageSize)
	var nextAfter []any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, nil, fmt.Errorf("postgres snapshot transport: read bounded row: %w", err)
		}
		rawValues := rows.RawValues()
		record, err := p.recordForValues(values, rawValues)
		if err != nil {
			return nil, nil, err
		}
		nextAfter, err = p.orderValues(values, rawValues)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("postgres snapshot transport: iterate bounded page: %w", err)
	}
	return records, nextAfter, nil
}

func (p postgresSnapshotReadPlan) recordForValues(values []any, rawValues [][]byte) (connectors.Record, error) {
	if len(values) != len(p.columns) {
		return nil, errors.New("postgres snapshot transport: bounded row does not match catalog projection")
	}
	if len(rawValues) != len(values) {
		return nil, errors.New("postgres snapshot transport: bounded raw row does not match catalog projection")
	}
	record := make(connectors.Record, len(p.columns))
	for index, column := range p.columns {
		var (
			value any
			err   error
		)
		if column.Type.Kind() == database.LogicalJSON {
			value, err = postgresSnapshotRawJSONValue(rawValues[index])
		} else {
			value, err = postgresSnapshotRecordValue(column.Type, values[index])
		}
		if err != nil {
			return nil, fmt.Errorf("postgres snapshot transport: normalize bounded column %q: %w", column.Ref.Name, err)
		}
		record[column.Ref.Name] = value
	}
	return record, nil
}

func postgresSnapshotRecordValue(logical database.LogicalType, value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch logical.Kind() {
	case database.LogicalSignedInteger:
		switch value.(type) {
		case int16, int32, int64:
			return value, nil
		}
	case database.LogicalUnsignedInteger:
		switch value.(type) {
		case uint8, uint16, uint32, uint64:
			return value, nil
		}
	case database.LogicalFloat:
		switch value.(type) {
		case float32, float64:
			return value, nil
		}
	case database.LogicalBoolean:
		if _, ok := value.(bool); ok {
			return value, nil
		}
	case database.LogicalString:
		if _, ok := value.(string); ok {
			return value, nil
		}
	case database.LogicalBinary:
		if bytes, ok := value.([]byte); ok {
			return append([]byte(nil), bytes...), nil
		}
	case database.LogicalDecimal:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case pgtype.Numeric:
			return postgresSnapshotValuerString(typed)
		}
	case database.LogicalDate:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case time.Time:
			return typed.Format(time.DateOnly), nil
		case pgtype.InfinityModifier:
			return typed.String(), nil
		}
	case database.LogicalTime:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case []byte:
			return string(typed), nil
		case pgtype.Time:
			return postgresSnapshotValuerString(typed)
		}
	case database.LogicalTimestamp:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case time.Time:
			if !logical.WithTimezone() {
				return typed.Format("2006-01-02T15:04:05.999999999"), nil
			}
			return typed.Format(time.RFC3339Nano), nil
		case pgtype.InfinityModifier:
			return typed.String(), nil
		}
	case database.LogicalUUID:
		switch typed := value.(type) {
		case string:
			return typed, nil
		case [16]byte:
			return pgtype.UUID{Bytes: typed, Valid: true}.String(), nil
		case pgtype.UUID:
			if !typed.Valid {
				return nil, nil
			}
			return typed.String(), nil
		}
	default:
		return nil, fmt.Errorf("unsupported typed catalog logical kind %q", logical.Kind())
	}
	return nil, fmt.Errorf("unexpected %T for typed catalog logical kind %q", value, logical.Kind())
}

func postgresSnapshotValuerString(value interface{ Value() (driver.Value, error) }) (any, error) {
	text, err := value.Value()
	if err != nil {
		return nil, err
	}
	if text == nil {
		return nil, nil
	}
	stringValue, ok := text.(string)
	if !ok {
		return nil, fmt.Errorf("expected string value, got %T", text)
	}
	return stringValue, nil
}

func postgresSnapshotRawJSONValue(raw []byte) (any, error) {
	if raw == nil {
		return nil, nil
	}
	if !json.Valid(raw) {
		return nil, errors.New("invalid JSON value")
	}
	return append(json.RawMessage(nil), raw...), nil
}

func (p postgresSnapshotReadPlan) query(after []any) string {
	columns := make([]string, 0, len(p.columns))
	for _, column := range p.columns {
		columns = append(columns, quoteIdentifier(column.Ref.Name))
	}
	order := make([]string, 0, len(p.order))
	for _, column := range p.order {
		order = append(order, quoteIdentifier(column.Name)+" ASC")
	}
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(quoteIdentifier(p.relation.Ref.Schema.Name))
	builder.WriteString(".")
	builder.WriteString(quoteIdentifier(p.relation.Ref.Name))
	if len(after) > 0 {
		terms := make([]string, 0, len(p.order))
		placeholders := make([]string, 0, len(p.order))
		for index, column := range p.order {
			terms = append(terms, quoteIdentifier(column.Name))
			placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		}
		builder.WriteString(" WHERE (")
		builder.WriteString(strings.Join(terms, ", "))
		builder.WriteString(") > (")
		builder.WriteString(strings.Join(placeholders, ", "))
		builder.WriteString(")")
	}
	builder.WriteString(" ORDER BY ")
	builder.WriteString(strings.Join(order, ", "))
	fmt.Fprintf(&builder, " LIMIT $%d", len(after)+1)
	return builder.String()
}

func (p postgresSnapshotReadPlan) queryArguments(after []any) []any {
	arguments := append([]any(nil), after...)
	return append(arguments, p.pageSize)
}

func (p postgresSnapshotReadPlan) orderValues(values []any, rawValues [][]byte) ([]any, error) {
	if len(values) != len(p.columns) {
		return nil, errors.New("postgres snapshot transport: bounded row does not match catalog projection")
	}
	if len(rawValues) != len(values) {
		return nil, errors.New("postgres snapshot transport: bounded raw row does not match catalog projection")
	}
	valuesByColumn := make(map[string]any, len(p.columns))
	rawValuesByColumn := make(map[string][]byte, len(p.columns))
	columnsByName := make(map[string]database.Column, len(p.columns))
	for index, column := range p.columns {
		valuesByColumn[column.Ref.Name] = values[index]
		rawValuesByColumn[column.Ref.Name] = rawValues[index]
		columnsByName[column.Ref.Name] = column
	}
	ordered := make([]any, 0, len(p.order))
	for _, orderColumn := range p.order {
		value, found := valuesByColumn[orderColumn.Name]
		column := columnsByName[orderColumn.Name]
		if column.Type.Kind() == database.LogicalJSON {
			var err error
			value, err = postgresSnapshotRawJSONValue(rawValuesByColumn[orderColumn.Name])
			if err != nil {
				return nil, fmt.Errorf("postgres snapshot transport: normalize stable key %q: %w", orderColumn.Name, err)
			}
		}
		if !found || value == nil {
			return nil, errors.New("postgres snapshot transport: stable key row value is absent")
		}
		paginationValue, err := postgresSnapshotPaginationValue(column.Type, value)
		if err != nil {
			return nil, fmt.Errorf("postgres snapshot transport: normalize stable key %q: %w", orderColumn.Name, err)
		}
		ordered = append(ordered, paginationValue)
	}
	return ordered, nil
}

// postgresSnapshotPaginationValue keeps cursor values in pgx's parameter
// vocabulary. Record values may deliberately be JSON-safe strings or raw JSON,
// but the next keyset query needs the original PostgreSQL type. Unknown or
// non-encodable catalog shapes fail here, before a partially-read page can ask
// pgx to guess at a driver-native value.
func postgresSnapshotPaginationValue(logical database.LogicalType, value any) (any, error) {
	if value == nil {
		return nil, fmt.Errorf("typed catalog logical type %q has no stable key value", logical.Kind())
	}
	if err := postgresSnapshotPaginationLogicalType(logical); err != nil {
		return nil, err
	}
	switch logical.Kind() {
	case database.LogicalSignedInteger:
		switch typed := value.(type) {
		case int16, int32, int64:
			return typed, nil
		}
	case database.LogicalUnsignedInteger:
		switch typed := value.(type) {
		case uint8, uint16, uint32, uint64:
			return typed, nil
		}
	case database.LogicalDecimal:
		if typed, ok := value.(pgtype.Numeric); ok {
			return typed, nil
		}
	case database.LogicalFloat:
		switch typed := value.(type) {
		case float32, float64:
			return typed, nil
		}
	case database.LogicalBoolean:
		if typed, ok := value.(bool); ok {
			return typed, nil
		}
	case database.LogicalString:
		if typed, ok := value.(string); ok {
			return typed, nil
		}
	case database.LogicalBinary:
		if typed, ok := value.([]byte); ok {
			return append([]byte(nil), typed...), nil
		}
	case database.LogicalDate:
		switch typed := value.(type) {
		case time.Time, pgtype.Date:
			return typed, nil
		case pgtype.InfinityModifier:
			if typed == pgtype.Infinity || typed == pgtype.NegativeInfinity {
				return pgtype.Date{InfinityModifier: typed, Valid: true}, nil
			}
		}
	case database.LogicalTime:
		switch typed := value.(type) {
		case pgtype.Time, string:
			return typed, nil
		case []byte:
			return append([]byte(nil), typed...), nil
		}
	case database.LogicalTimestamp:
		switch typed := value.(type) {
		case time.Time:
			return typed, nil
		case pgtype.InfinityModifier:
			if typed != pgtype.Infinity && typed != pgtype.NegativeInfinity {
				break
			}
			if logical.WithTimezone() {
				return pgtype.Timestamptz{InfinityModifier: typed, Valid: true}, nil
			}
			return pgtype.Timestamp{InfinityModifier: typed, Valid: true}, nil
		case pgtype.Timestamp:
			if !logical.WithTimezone() {
				return typed, nil
			}
		case pgtype.Timestamptz:
			if logical.WithTimezone() {
				return typed, nil
			}
		}
	case database.LogicalUUID:
		switch typed := value.(type) {
		case [16]byte:
			return pgtype.UUID{Bytes: typed, Valid: true}, nil
		case pgtype.UUID:
			if typed.Valid {
				return typed, nil
			}
		}
	default:
		return nil, fmt.Errorf("typed catalog logical type %q cannot normalize PostgreSQL stable key value %T", logical.Kind(), value)
	}
	return nil, fmt.Errorf("typed catalog logical type %q cannot normalize PostgreSQL stable key value %T", logical.Kind(), value)
}

func postgresSnapshotPaginationLogicalType(logical database.LogicalType) error {
	switch logical.Kind() {
	case database.LogicalSignedInteger,
		database.LogicalUnsignedInteger,
		database.LogicalDecimal,
		database.LogicalFloat,
		database.LogicalBoolean,
		database.LogicalString,
		database.LogicalBinary,
		database.LogicalDate,
		database.LogicalTime,
		database.LogicalTimestamp,
		database.LogicalUUID:
		return nil
	default:
		return fmt.Errorf("typed catalog logical type %q is not encodable for a PostgreSQL stable key", logical.Kind())
	}
}

func postgresSnapshotCheckpoint(resume synccontract.ResumeExpectation, snapshotToken string, fingerprint database.SchemaFingerprint, page int) (synccontract.CheckpointEnvelope, error) {
	if page < 0 || strings.TrimSpace(snapshotToken) == "" || fingerprint.IsZero() {
		return synccontract.CheckpointEnvelope{}, errors.New("postgres snapshot transport checkpoint inputs are invalid")
	}
	identity, err := postgresSnapshotPageIdentity(snapshotToken, fingerprint, resume.Source, page)
	if err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	positionObserved := false
	barrier := synccontract.OpaqueToken([]byte(snapshotToken))
	checkpoint := synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           resume.Source,
		Mechanism:        postgresSnapshotMechanism,
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: postgresSnapshotCheckpointKind, Token: append(synccontract.OpaqueToken(nil), barrier...)},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: append(synccontract.OpaqueToken(nil), resume.SourceGeneration...),
		SchemaVersion:    fingerprint.String(),
		ProtocolVersion:  postgresSnapshotProtocolVersion,
		Dedupe:           synccontract.DedupeIdentity{Kind: postgresSnapshotDedupeKind, Value: identity},
		DedupeWindow:     synccontract.DedupeWindow{Kind: postgresSnapshotDedupeWindowKind, Start: append(synccontract.OpaqueToken(nil), barrier...), End: append(synccontract.OpaqueToken(nil), barrier...)},
		ObservedAt:       time.Now().UTC(),
	}
	if err := checkpoint.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	return checkpoint, nil
}

func postgresSnapshotPageIdentity(snapshotToken string, fingerprint database.SchemaFingerprint, source synccontract.SourceIdentity, page int) (synccontract.OpaqueToken, error) {
	payload, err := json.Marshal(struct {
		Snapshot string                      `json:"snapshot"`
		Schema   string                      `json:"schema"`
		Source   synccontract.SourceIdentity `json:"source"`
		Page     int                         `json:"page"`
	}{
		Snapshot: snapshotToken,
		Schema:   fingerprint.String(),
		Source:   source,
		Page:     page,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres snapshot transport: encode checkpoint identity: %w", err)
	}
	sum := sha256.Sum256(payload)
	return append(synccontract.OpaqueToken(nil), sum[:]...), nil
}
