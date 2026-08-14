package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

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
	postgresSnapshotConformanceRunID = "bounded_full_v1"
)

var postgresSnapshotTransportReference = connectors.TransportExecutorReference{
	Family: connectors.TransportExecutorFamilyNativeDatabase,
	ID:     postgresSnapshotTransportID,
}

// postgresSnapshotTransportDescriptor is intentionally constructed next to
// its executor. The PostgreSQL bundle supplies the connector definition, while
// this native connector supplies the one dynamic catalog-backed source role.
func postgresSnapshotTransportDescriptor() *connectors.SourceTransportDescriptor {
	return &connectors.SourceTransportDescriptor{
		Executor:        postgresSnapshotTransportReference,
		EligibleStreams: []string{postgresSnapshotTransportStream},
		Modes:           []synccontract.Mode{synccontract.ModeFullAppend, synccontract.ModeFullOverwrite},
		Delivery: connectors.DeliveryGuarantees{
			Idempotency: connectors.DeliveryIdempotencyAtLeastOnce,
			Ordering:    connectors.DeliveryOrderingSource,
			Deletes:     connectors.DeliveryDeletesUnavailable,
		},
		Conformance: connectors.ConformanceEvidenceReference{
			Suite: postgresSnapshotConformanceSuite,
			RunID: postgresSnapshotConformanceRunID,
		},
	}
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

// ReadTransport executes one exact, bounded full snapshot in a read-only
// repeatable-read transaction. Catalog discovery and record queries share that
// transaction, so the emitted catalog fingerprint, stable key order, rows, and
// PostgreSQL snapshot token describe the same source observation.
func (s *SnapshotTransportSource) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) (err error) {
	if ctx == nil {
		return errors.New("postgres snapshot transport context is required")
	}
	if emit == nil {
		return errors.New("postgres snapshot transport emit function is required")
	}
	if err := ctx.Err(); err != nil {
		return err
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
		candidate, err := postgresSnapshotCheckpoint(request.Resume, snapshotToken, plan.fingerprint, page, nextAfter)
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

func (s *SnapshotTransportSource) validateReadRequest(request synctransport.SourceRequest) (connConfig, database.RelationRef, typedCatalogResources, int, error) {
	if s == nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport source is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != s.connector.Name() || request.Stream != postgresSnapshotTransportStream {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport accepts only its snapshot stream")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(request.Connector)
	if !ok || descriptor.Executor != postgresSnapshotTransportReference {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport requires its declared source descriptor")
	}
	if request.Mode != synccontract.ModeFullAppend && request.Mode != synccontract.ModeFullOverwrite {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport accepts only full_append or full_overwrite")
	}
	if request.BatchSize <= 0 {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport requires a positive batch size")
	}
	if request.Checkpoint != nil {
		return connConfig{}, database.RelationRef{}, typedCatalogResources{}, 0, errors.New("postgres snapshot transport does not resume a full snapshot checkpoint")
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
	relationRef, err := postgresSnapshotRelationRef(request.Resume.Source.ObjectScope)
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

func postgresSnapshotRelationRef(scope string) (database.RelationRef, error) {
	parts := strings.Split(scope, ".")
	if len(parts) != 3 {
		return database.RelationRef{}, errors.New("postgres snapshot transport source scope must be database.schema.relation")
	}
	for _, part := range parts {
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
		if len(values) != len(p.columns) {
			return nil, nil, errors.New("postgres snapshot transport: bounded row does not match catalog projection")
		}
		record := make(connectors.Record, len(p.columns))
		for index, column := range p.columns {
			record[column.Ref.Name] = values[index]
		}
		nextAfter, err = p.orderValues(values)
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

func (p postgresSnapshotReadPlan) orderValues(values []any) ([]any, error) {
	valuesByColumn := make(map[string]any, len(p.columns))
	for index, column := range p.columns {
		valuesByColumn[column.Ref.Name] = values[index]
	}
	ordered := make([]any, 0, len(p.order))
	for _, orderColumn := range p.order {
		value, found := valuesByColumn[orderColumn.Name]
		if !found || value == nil {
			return nil, errors.New("postgres snapshot transport: stable key row value is absent")
		}
		ordered = append(ordered, value)
	}
	return ordered, nil
}

func postgresSnapshotCheckpoint(resume synccontract.ResumeExpectation, snapshotToken string, fingerprint database.SchemaFingerprint, page int, position []any) (synccontract.CheckpointEnvelope, error) {
	if page < 0 || strings.TrimSpace(snapshotToken) == "" || fingerprint.IsZero() {
		return synccontract.CheckpointEnvelope{}, errors.New("postgres snapshot transport checkpoint inputs are invalid")
	}
	identity, err := postgresSnapshotPageIdentity(snapshotToken, fingerprint, resume.Source, page, position)
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

func postgresSnapshotPageIdentity(snapshotToken string, fingerprint database.SchemaFingerprint, source synccontract.SourceIdentity, page int, position []any) (synccontract.OpaqueToken, error) {
	payload, err := json.Marshal(struct {
		Snapshot string                      `json:"snapshot"`
		Schema   string                      `json:"schema"`
		Source   synccontract.SourceIdentity `json:"source"`
		Page     int                         `json:"page"`
		Position []any                       `json:"position"`
	}{
		Snapshot: snapshotToken,
		Schema:   fingerprint.String(),
		Source:   source,
		Page:     page,
		Position: position,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres snapshot transport: encode checkpoint identity: %w", err)
	}
	sum := sha256.Sum256(payload)
	return append(synccontract.OpaqueToken(nil), sum[:]...), nil
}
