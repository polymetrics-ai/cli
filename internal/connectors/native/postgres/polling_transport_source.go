package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	postgresPollingTransportID      = "postgres_polling_watermark"
	postgresPollingConformanceSuite = "postgres_polling_watermark"
	postgresPollingConformanceRunID = "shared_source_v1"
	// postgresPollingProtocolVersion is shared-engine-owned so PostgreSQL's
	// pre-I/O admission cannot reject a checkpoint the shared executor emits.
	postgresPollingProtocolVersion = engine.PollingSourceProtocolVersion
	postgresPollingBarrierToken    = "postgres-polling-none-v1"
)

var (
	postgresPollingTransportReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyNativeDatabase,
		ID:     postgresPollingTransportID,
	}

	// ErrPollingCursorFieldRequired is returned before configuration parsing,
	// pool creation, catalog discovery, or page delivery. A stored checkpoint
	// without the stream's cursor must never degrade into an unfiltered scan.
	ErrPollingCursorFieldRequired = errors.New("postgres polling source requires a per-stream cursor field")
	// ErrPollingCursorFieldNotFound is a live-catalog refusal; no provider page
	// query is issued when a configured stream cursor is absent.
	ErrPollingCursorFieldNotFound = errors.New("postgres polling cursor field is absent from the selected relation")
	// ErrPollingCursorNullable prevents PostgreSQL three-valued comparison from
	// silently omitting NULL cursor rows.
	ErrPollingCursorNullable     = errors.New("postgres polling cursor field must be non-null")
	ErrPollingCursorUnsupported  = errors.New("postgres polling cursor field has an unsupported resumable type")
	ErrPollingTieBreakerRequired = errors.New("postgres polling source requires one distinct non-null unique tie-breaker")
)

// PollingTransportSource owns PostgreSQL's catalog-bound, one-page runner.
// It deliberately delegates paging, candidate construction, checkpoint
// validation, and durable-resume sequencing to engine.PollingSourceExecutor.
// No local page loop is permitted here.
type PollingTransportSource struct{ connector Connector }

var _ synctransport.SourceExecutor = (*PollingTransportSource)(nil)

func NewPollingTransportSource(connector Connector) *PollingTransportSource {
	return &PollingTransportSource{connector: connector}
}

func PollingTransportDefinitionFactory() synctransport.DefinitionFactory {
	return synctransport.DefinitionFactory{
		Reference: postgresPollingTransportReference,
		SourceEvidence: connectors.ConformanceEvidenceReference{
			Suite: postgresPollingConformanceSuite,
			RunID: postgresPollingConformanceRunID,
		},
		BuildSource: func(connector connectors.Connector) (synctransport.SourceExecutor, error) {
			switch typed := connector.(type) {
			case Connector:
				return NewPollingTransportSource(typed), nil
			case *Connector:
				if typed == nil {
					return nil, errors.New("postgres polling transport connector is required")
				}
				return NewPollingTransportSource(*typed), nil
			default:
				return nil, errors.New("postgres polling transport requires the native postgres connector")
			}
		},
	}
}

// RegisterPollingTransportSource is retained for focused registry integration
// tests. Production composition uses PollingTransportDefinitionFactory via
// Connector.SyncTransportDefinitionFactories.
func RegisterPollingTransportSource(registry *synctransport.Registry, connector Connector) error {
	if registry == nil {
		return errors.New("postgres polling transport registry is required")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(connector)
	if !ok || descriptor.Executor != postgresPollingTransportReference {
		return errors.New("postgres polling transport definition is unavailable")
	}
	return registry.RegisterSource(NewPollingTransportSource(connector))
}

func (*PollingTransportSource) TransportExecutorReference() connectors.TransportExecutorReference {
	return postgresPollingTransportReference
}

func (s *PollingTransportSource) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if err := s.validateRequestBeforeIO(ctx, request, emit); err != nil {
		return err
	}
	return executeWithAuthenticationAdmission(ctx, request.Runtime, func(admitted context.Context) error {
		return s.readPollingTransport(admitted, request, emit)
	})
}

func (s *PollingTransportSource) validateRequestBeforeIO(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if s == nil {
		return errors.New("postgres polling transport source is unavailable")
	}
	if ctx == nil {
		return errors.New("postgres polling transport context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if emit == nil {
		return errors.New("postgres polling transport emit function is required")
	}
	if request.Connector == nil || request.Connector.Name() != s.connector.Name() || strings.TrimSpace(request.Stream) == "" {
		return errors.New("postgres polling transport requires a declared dynamic stream")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(request.Connector)
	if !ok || descriptor.Executor != postgresPollingTransportReference {
		return errors.New("postgres polling transport requires its declared source descriptor")
	}
	if err := request.Mode.Validate(); err != nil {
		return err
	}
	if request.BatchSize <= 0 {
		return errors.New("postgres polling transport requires a positive batch size")
	}
	if strings.TrimSpace(request.CursorField) == "" {
		return ErrPollingCursorFieldRequired
	}
	if err := validateIdentifier(request.CursorField); err != nil {
		return fmt.Errorf("postgres polling cursor field: %w", err)
	}
	if len(request.PrimaryKey) != 1 || strings.TrimSpace(request.PrimaryKey[0]) == "" || request.PrimaryKey[0] == request.CursorField {
		return ErrPollingTieBreakerRequired
	}
	if err := validateIdentifier(request.PrimaryKey[0]); err != nil {
		return fmt.Errorf("postgres polling tie-breaker: %w", err)
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return errors.New("postgres polling transport requires a complete resume identity")
	}
	if request.Resume.Source.Engine != postgresSnapshotSourceEngine {
		return errors.New("postgres polling transport source identity engine must be postgres")
	}
	if request.Checkpoint != nil {
		if err := request.Checkpoint.ValidateResume(request.Resume); err != nil {
			return err
		}
		if postgresBootstrapTransportRequested(request) {
			if err := validateCDCCheckpointProtocol(request.Checkpoint); err != nil {
				return err
			}
			_, err := nativeBootstrapTransportCheckpoint(*request.Checkpoint, request.Runtime)
			return err
		}
		if request.Checkpoint.Mechanism != engine.PollingSourceCheckpointMechanism {
			return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL polling checkpoint mechanism is not resumable")
		}
		if request.Checkpoint.ProtocolVersion != postgresPollingProtocolVersion {
			return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL polling checkpoint protocol is not resumable")
		}
		if request.Checkpoint.SnapshotBarrier == nil || request.Checkpoint.SnapshotBarrier.Kind != string(connectors.PollingSnapshotBarrierNone) || string(request.Checkpoint.SnapshotBarrier.Token) != postgresPollingBarrierToken {
			return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL polling checkpoint barrier is not resumable")
		}
	}
	return nil
}

func (s *PollingTransportSource) readPollingTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	// Logical replication bootstrap is a separate changefeed handover with a
	// slot barrier, not a polling source. Preserve that sealed CDC path while
	// ensuring every ordinary PostgreSQL source mode reaches the shared poller.
	if postgresBootstrapTransportRequested(request) {
		return (&SnapshotTransportSource{connector: s.connector}).readBootstrapTransport(ctx, request, emit)
	}
	runner, declaration, object, closeRunner, err := s.preparePollingRunner(ctx, request)
	if err != nil {
		return err
	}
	defer closeRunner()

	registry := engine.NewPollingPreflightRegistry()
	if err := registry.RegisterSource(runner); err != nil {
		return fmt.Errorf("register postgres polling source: %w", err)
	}
	if err := registry.RegisterApply(postgresPollingPreflightApply{}); err != nil {
		return fmt.Errorf("register postgres polling apply: %w", err)
	}
	resolved, err := engine.PollingPreflight(ctx, registry, declaration, object, request.Mode)
	if err != nil {
		return fmt.Errorf("preflight postgres polling source: %w", err)
	}
	shared, err := engine.NewPollingSourceExecutor(resolved)
	if err != nil {
		return fmt.Errorf("construct shared postgres polling source: %w", err)
	}
	return shared.ReadTransport(ctx, request, emit)
}

func postgresBootstrapTransportRequested(request synctransport.SourceRequest) bool {
	return request.Mode == synccontract.ModeIncrementalUpsert && strings.EqualFold(strings.TrimSpace(request.Runtime.Config["transport_bootstrap"]), "true")
}

func (s *PollingTransportSource) preparePollingRunner(ctx context.Context, request synctransport.SourceRequest) (engine.PollingSourceRunner, *connectors.PollingWatermarkDescriptor, connectors.PollingCatalogObject, func(), error) {
	if fixtureMode(request.Runtime) {
		return s.prepareFixturePollingRunner(request)
	}
	conn, err := resolveConfig(request.Runtime)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	if err := validateIdentifier(conn.schema); err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, nil, ErrUnsupportedCatalogShape
	}
	if isSystemCatalogSchema(conn.schema) {
		return nil, nil, connectors.PollingCatalogObject{}, nil, ErrSystemCatalogSchema
	}
	if err := s.connector.databaseDefinition.Validate(); err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, nil, errors.New("postgres polling transport definition is unavailable")
	}
	resources, err := newTypedCatalogResources(s.connector.databaseDefinition.Resources())
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	operationCtx, cancel, err := resources.operationContext(ctx)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	pool, err := conn.openTypedCatalogPool(operationCtx, resources)
	if err != nil {
		cancel()
		return nil, nil, connectors.PollingCatalogObject{}, nil, fmt.Errorf("postgres polling transport: open pool: %w", err)
	}
	closeRunner := func() { pool.Close(); cancel() }
	catalog, err := discoverTypedCatalog(operationCtx, pool, conn.database, conn.schema, s.connector.databaseDefinition, resources)
	if err != nil {
		closeRunner()
		return nil, nil, connectors.PollingCatalogObject{}, nil, fmt.Errorf("postgres polling transport: discover live catalog: %w", err)
	}
	relation, err := postgresSnapshotRelationRef(request.Stream, conn.database, conn.schema)
	if err != nil {
		closeRunner()
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	pageSize, err := s.connector.databaseDefinition.Resources().EffectivePageSize(request.BatchSize)
	if err != nil {
		closeRunner()
		return nil, nil, connectors.PollingCatalogObject{}, nil, errors.New("postgres polling transport batch size is outside the declared resource bound")
	}
	plan, err := newPostgresPollingReadPlan(catalog, relation, request.CursorField, request.PrimaryKey, pageSize)
	if err != nil {
		closeRunner()
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	object := pollingCatalogObject(plan.relation)
	declaration, err := postgresPollingDeclaration(request, plan)
	if err != nil {
		closeRunner()
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	return &postgresPollingSourceRunner{pool: pool, plan: plan, state: postgresPollingRuntimeState(request, plan), reference: postgresPollingTransportReference, definition: s.connector.databaseDefinition}, declaration, object, closeRunner, nil
}

func (s *PollingTransportSource) prepareFixturePollingRunner(request synctransport.SourceRequest) (engine.PollingSourceRunner, *connectors.PollingWatermarkDescriptor, connectors.PollingCatalogObject, func(), error) {
	rows, ok := fixtureRows(request.Stream)
	if !ok {
		return nil, nil, connectors.PollingCatalogObject{}, nil, fmt.Errorf("postgres fixture stream %q not found", request.Stream)
	}
	stream, found := fixturePollingStream(request.Stream)
	if !found {
		return nil, nil, connectors.PollingCatalogObject{}, nil, fmt.Errorf("postgres fixture stream %q is unavailable for polling", request.Stream)
	}
	if !containsString(stream.CursorFields, request.CursorField) {
		return nil, nil, connectors.PollingCatalogObject{}, nil, ErrPollingCursorFieldNotFound
	}
	if len(request.PrimaryKey) != 1 || request.PrimaryKey[0] != stream.PrimaryKey[0] {
		return nil, nil, connectors.PollingCatalogObject{}, nil, ErrPollingTieBreakerRequired
	}
	plan := postgresPollingReadPlan{cursor: request.CursorField, tieBreaker: request.PrimaryKey[0], pageSize: request.BatchSize, cursorType: connectors.PollingCursor{Codec: connectors.PollingCursorCodecDecimal, Type: connectors.PollingCursorTypeInteger, Precision: "exact"}}
	for _, field := range stream.Fields {
		plan.columns = append(plan.columns, database.Column{Ref: database.ColumnRef{Name: field.Name}, Ordinal: len(plan.columns) + 1})
	}
	plan.relation = database.Relation{Ref: database.RelationRef{Name: request.Stream}}
	declaration, err := postgresPollingDeclaration(request, plan)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, nil, err
	}
	object := connectors.PollingCatalogObject{Kind: connectors.PollingCatalogObjectRelation, NameParts: []string{"fixture", "public", strings.TrimPrefix(request.Stream, "public.")}}
	for _, field := range stream.Fields {
		object.Columns = append(object.Columns, field.Name)
	}
	runner := &postgresFixturePollingRunner{rows: rows, cursor: request.CursorField, tieBreaker: request.PrimaryKey[0], state: postgresPollingRuntimeState(request, plan), reference: postgresPollingTransportReference, definition: s.connector.databaseDefinition}
	return runner, declaration, object, func() {}, nil
}

func fixturePollingStream(name string) (connectors.Stream, bool) {
	for _, stream := range fixtureStreams() {
		if stream.Name == name {
			return stream, true
		}
	}
	return connectors.Stream{}, false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type postgresPollingPreflightApply struct{}

func (postgresPollingPreflightApply) PollingApplyExecutorReference() connectors.TransportExecutorReference {
	return postgresManagedTargetTransportReference
}

func (postgresPollingPreflightApply) PollingApplyConformanceEvidence() engine.PollingWatermarkConformanceEvidence {
	return engine.RequiredPollingWatermarkConformanceEvidence()
}

type postgresPollingReadPlan struct {
	relation      database.Relation
	columns       []database.Column
	cursor        string
	tieBreaker    string
	pageSize      int
	cursorType    connectors.PollingCursor
	schemaVersion string
}

func newPostgresPollingReadPlan(catalog database.Catalog, relationRef database.RelationRef, cursor string, primaryKey []string, pageSize int) (postgresPollingReadPlan, error) {
	if pageSize <= 0 {
		return postgresPollingReadPlan{}, errors.New("postgres polling source requires a finite page size")
	}
	if len(primaryKey) != 1 || strings.TrimSpace(primaryKey[0]) == "" || primaryKey[0] == cursor {
		return postgresPollingReadPlan{}, ErrPollingTieBreakerRequired
	}
	for _, relation := range catalog.Relations() {
		if relation.Ref.Schema.Catalog.Name != relationRef.Schema.Catalog.Name || relation.Ref.Schema.Name != relationRef.Schema.Name || relation.Ref.Name != relationRef.Name {
			continue
		}
		columns := append([]database.Column(nil), relation.Columns...)
		sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
		cursorColumn, cursorFound := postgresPollingColumn(columns, cursor)
		if !cursorFound {
			return postgresPollingReadPlan{}, ErrPollingCursorFieldNotFound
		}
		if cursorColumn.Nullable {
			return postgresPollingReadPlan{}, ErrPollingCursorNullable
		}
		cursorType, err := postgresPollingCursorType(cursorColumn)
		if err != nil {
			return postgresPollingReadPlan{}, err
		}
		tieColumn, tieFound := postgresPollingColumn(columns, primaryKey[0])
		if !tieFound || tieColumn.Nullable || !postgresPollingHasSingleUniqueKey(relation, primaryKey[0]) || !postgresPollingTieTypeSupported(tieColumn) {
			return postgresPollingReadPlan{}, ErrPollingTieBreakerRequired
		}
		return postgresPollingReadPlan{relation: relation, columns: columns, cursor: cursor, tieBreaker: primaryKey[0], pageSize: pageSize, cursorType: cursorType, schemaVersion: catalog.Fingerprint().String()}, nil
	}
	return postgresPollingReadPlan{}, errors.New("postgres polling source relation is absent from the live catalog")
}

func postgresPollingColumn(columns []database.Column, name string) (database.Column, bool) {
	for _, column := range columns {
		if column.Ref.Name == name {
			return column, true
		}
	}
	return database.Column{}, false
}

func postgresPollingHasSingleUniqueKey(relation database.Relation, name string) bool {
	for _, key := range relation.Keys {
		if (key.Kind == database.KeyPrimary || key.Kind == database.KeyUnique) && len(key.Columns) == 1 && key.Columns[0].Name == name {
			return true
		}
	}
	return false
}

func postgresPollingCursorType(column database.Column) (connectors.PollingCursor, error) {
	switch column.Type.Kind() {
	case database.LogicalSignedInteger, database.LogicalUnsignedInteger:
		return connectors.PollingCursor{Codec: connectors.PollingCursorCodecDecimal, Type: connectors.PollingCursorTypeInteger, Precision: "exact"}, nil
	case database.LogicalTimestamp:
		return connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, Precision: "nanosecond"}, nil
	default:
		return connectors.PollingCursor{}, ErrPollingCursorUnsupported
	}
}

func postgresPollingTieTypeSupported(column database.Column) bool {
	switch column.Type.Kind() {
	case database.LogicalSignedInteger, database.LogicalUnsignedInteger, database.LogicalUUID, database.LogicalString:
		return true
	default:
		return false
	}
}

func pollingCatalogObject(relation database.Relation) connectors.PollingCatalogObject {
	object := connectors.PollingCatalogObject{Kind: connectors.PollingCatalogObjectRelation, NameParts: []string{relation.Ref.Schema.Catalog.Name, relation.Ref.Schema.Name, relation.Ref.Name}}
	for _, column := range relation.Columns {
		object.Columns = append(object.Columns, column.Ref.Name)
	}
	return object
}

func postgresPollingDeclaration(request synctransport.SourceRequest, plan postgresPollingReadPlan) (*connectors.PollingWatermarkDescriptor, error) {
	declaration := &connectors.PollingWatermarkDescriptor{
		Status: connectors.PollingWatermarkStatusImplemented,
		Source: connectors.PollingWatermarkSourceDescriptor{
			Executor:         postgresPollingTransportReference,
			Object:           connectors.PollingCatalogObjectSelector{Kind: connectors.PollingCatalogObjectRelation},
			Read:             connectors.PollingReadProtocol{Kind: connectors.PollingReadProtocolKeyset, MaxPageSize: 10000, MaxPages: 10000, MaxRequests: 10000, StableTraversal: true, Predicate: connectors.PollingKeysetPredicateLexicographic},
			Snapshot:         connectors.PollingSnapshotBarrier{Kind: connectors.PollingSnapshotBarrierNone},
			Cursor:           plan.cursorType,
			Ordering:         connectors.PollingOrderingTuple{Watermark: connectors.PollingOrderingField{CatalogField: plan.cursor, Ascending: true}, TieBreaker: connectors.PollingOrderingField{CatalogField: plan.tieBreaker, Ascending: true, Unique: true}},
			Mutation:         connectors.PollingMutationPolicy{Mutable: false},
			Identity:         connectors.PollingSourceIdentity{Engine: request.Resume.Source.Engine, AccountScope: request.Resume.Source.AccountOrCluster, ObjectScope: request.Resume.Source.ObjectScope},
			Schema:           connectors.PollingSchemaCompatibilityExactFingerprint,
			DeleteVisibility: connectors.PollingDeleteVisibilityHardDeleteInvisible,
			Modes: []synccontract.Mode{
				synccontract.ModeFullOverwrite, synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend,
				synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupe, synccontract.ModeIncrementalDedupeHistory,
			},
		},
		Target: connectors.PollingApplyDescriptor{
			Executor: postgresManagedTargetTransportReference, MaxBatchRecords: 10000, MaxBatchBytes: 1 << 30,
			Staging: connectors.PollingStagingReplaceSupported, StableKeyMapping: []string{plan.tieBreaker}, ConditionalOrderFence: true,
			Transaction: connectors.PollingTransactionRequired, PartialResult: connectors.PollingPartialResultRollback,
			RetrySafeCloseAndInsert: true, ValidityWindow: connectors.PollingValidityWindowSupported,
			Strategies: []connectors.PollingApplyStrategy{connectors.PollingApplyStrategyReplace, connectors.PollingApplyStrategyAppend, connectors.PollingApplyStrategyMerge, connectors.PollingApplyStrategyDedupe, connectors.PollingApplyStrategyDedupeHistory},
		},
	}
	if err := declaration.Validate(); err != nil {
		return nil, fmt.Errorf("postgres polling declaration: %w", err)
	}
	return declaration, nil
}

func postgresPollingRuntimeState(request synctransport.SourceRequest, plan postgresPollingReadPlan) engine.PollingSourceRuntimeState {
	digest := sha256.Sum256([]byte(request.Resume.Source.Engine + "\n" + request.Resume.Source.AccountOrCluster + "\n" + request.Resume.Source.ObjectScope + "\n" + plan.cursor + "\n" + plan.tieBreaker))
	return engine.PollingSourceRuntimeState{
		SourceGeneration: append(synccontract.OpaqueToken(nil), request.Resume.SourceGeneration...),
		SchemaVersion:    postgresPollingSchemaVersion(plan),
		SnapshotBarrier:  synccontract.SnapshotBarrier{Kind: string(connectors.PollingSnapshotBarrierNone), Token: synccontract.OpaqueToken(postgresPollingBarrierToken)},
		Partitions:       []synccontract.PartitionState{},
		Dedupe:           synccontract.DedupeIdentity{Kind: "postgres_polling_tuple", Value: append(synccontract.OpaqueToken(nil), digest[:]...)},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "postgres_polling_overlap", Start: synccontract.OpaqueToken(postgresPollingBarrierToken), End: synccontract.OpaqueToken(postgresPollingBarrierToken)},
	}
}

func postgresPollingSchemaVersion(plan postgresPollingReadPlan) string {
	if plan.schemaVersion != "" {
		return plan.schemaVersion
	}
	if plan.relation.Ref.Schema.Catalog.Name == "" {
		return "postgres-fixture-" + plan.cursor + "-" + plan.tieBreaker + "-v1"
	}
	var source strings.Builder
	source.WriteString(plan.relation.Ref.Schema.Catalog.Name)
	source.WriteByte('\n')
	source.WriteString(plan.relation.Ref.Schema.Name)
	source.WriteByte('\n')
	source.WriteString(plan.relation.Ref.Name)
	for _, column := range plan.columns {
		source.WriteByte('\n')
		source.WriteString(column.Ref.Name)
		source.WriteByte(':')
		source.WriteString(string(column.Type.Kind()))
		source.WriteByte(':')
		source.WriteString(strconv.FormatBool(column.Nullable))
	}
	sum := sha256.Sum256([]byte(source.String()))
	return fmt.Sprintf("%x", sum[:])
}

type postgresPollingSourceRunner struct {
	pool       *pgxpool.Pool
	plan       postgresPollingReadPlan
	state      engine.PollingSourceRuntimeState
	reference  connectors.TransportExecutorReference
	definition database.Definition
}

func (r *postgresPollingSourceRunner) PollingSourceExecutorReference() connectors.TransportExecutorReference {
	return r.reference
}
func (*postgresPollingSourceRunner) PollingSourceConformanceEvidence() engine.PollingWatermarkConformanceEvidence {
	return engine.RequiredPollingWatermarkConformanceEvidence()
}
func (r *postgresPollingSourceRunner) PollingSourceDatabaseDefinition() database.Definition {
	if r == nil {
		return database.Definition{}
	}
	return r.definition
}
func (r *postgresPollingSourceRunner) PollingSourceRuntimeState(context.Context, connectors.PollingCatalogObject) (engine.PollingSourceRuntimeState, error) {
	return r.state.Clone(), nil
}

func (r *postgresPollingSourceRunner) FetchPollingSourcePage(ctx context.Context, request engine.PollingSourcePageRequest) (engine.PollingSourcePage, error) {
	if r == nil || r.pool == nil {
		return engine.PollingSourcePage{}, errors.New("postgres polling source runner is unavailable")
	}
	if err := request.RequestBudget.Consume(ctx); err != nil {
		return engine.PollingSourcePage{}, err
	}
	after, err := r.plan.afterValues(request.After)
	if err != nil {
		return engine.PollingSourcePage{}, err
	}
	rows, err := r.pool.Query(ctx, r.plan.query(after, request.PageSize+1), r.plan.queryArguments(after, request.PageSize+1)...)
	if err != nil {
		return engine.PollingSourcePage{}, fmt.Errorf("postgres polling query bounded page: %w", err)
	}
	defer rows.Close()
	items, err := r.plan.items(rows)
	if err != nil {
		return engine.PollingSourcePage{}, err
	}
	more := len(items) > request.PageSize
	if more {
		items = items[:request.PageSize]
	}
	return engine.PollingSourcePage{Items: items, More: more, ObservedAt: time.Now().UTC()}, nil
}

func (r *postgresPollingSourceRunner) ValidatePollingSourcePageTraversal(_ context.Context, after *synccontract.CheckpointPosition, page engine.PollingSourcePage) error {
	return r.plan.validateTraversal(after, page)
}

type postgresFixturePollingRunner struct {
	rows       []fixtureRow
	cursor     string
	tieBreaker string
	state      engine.PollingSourceRuntimeState
	reference  connectors.TransportExecutorReference
	definition database.Definition
}

func (r *postgresFixturePollingRunner) PollingSourceExecutorReference() connectors.TransportExecutorReference {
	return r.reference
}
func (*postgresFixturePollingRunner) PollingSourceConformanceEvidence() engine.PollingWatermarkConformanceEvidence {
	return engine.RequiredPollingWatermarkConformanceEvidence()
}
func (r *postgresFixturePollingRunner) PollingSourceDatabaseDefinition() database.Definition {
	if r == nil {
		return database.Definition{}
	}
	return r.definition
}
func (r *postgresFixturePollingRunner) PollingSourceRuntimeState(context.Context, connectors.PollingCatalogObject) (engine.PollingSourceRuntimeState, error) {
	return r.state.Clone(), nil
}

func (r *postgresFixturePollingRunner) FetchPollingSourcePage(ctx context.Context, request engine.PollingSourcePageRequest) (engine.PollingSourcePage, error) {
	if err := request.RequestBudget.Consume(ctx); err != nil {
		return engine.PollingSourcePage{}, err
	}
	var afterCursor int64 = -1
	var afterTie int64 = -1
	if request.After != nil {
		var err error
		afterCursor, err = strconv.ParseInt(string(request.After.Primary), 10, 64)
		if err != nil {
			return engine.PollingSourcePage{}, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL fixture cursor tuple is invalid")
		}
		afterTie, err = strconv.ParseInt(string(request.After.TieBreaker), 10, 64)
		if err != nil {
			return engine.PollingSourcePage{}, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL fixture tie-breaker tuple is invalid")
		}
	}
	items := make([]engine.PollingSourceItem, 0, request.PageSize+1)
	for _, row := range r.rows {
		id, ok := row.record[r.tieBreaker].(int)
		if !ok || row.record[r.cursor] == nil {
			return engine.PollingSourcePage{}, errors.New("postgres fixture polling row is invalid")
		}
		if row.cursor < afterCursor || (row.cursor == afterCursor && int64(id) <= afterTie) {
			continue
		}
		items = append(items, engine.PollingSourceItem{Record: copyRecord(row.record), Position: synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken(strconv.FormatInt(row.cursor, 10)), TieBreaker: synccontract.OpaqueToken(strconv.Itoa(id))}})
		if len(items) > request.PageSize {
			break
		}
	}
	more := len(items) > request.PageSize
	if more {
		items = items[:request.PageSize]
	}
	return engine.PollingSourcePage{Items: items, More: more, ObservedAt: time.Now().UTC()}, nil
}

func (*postgresFixturePollingRunner) ValidatePollingSourcePageTraversal(context.Context, *synccontract.CheckpointPosition, engine.PollingSourcePage) error {
	return nil
}

func (p postgresPollingReadPlan) query(after []any, limit int) string {
	columns := make([]string, 0, len(p.columns))
	for _, column := range p.columns {
		columns = append(columns, quoteIdentifier(column.Ref.Name))
	}
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(quoteIdentifier(p.relation.Ref.Schema.Name))
	builder.WriteString(".")
	builder.WriteString(quoteIdentifier(p.relation.Ref.Name))
	if len(after) == 2 {
		builder.WriteString(" WHERE (")
		builder.WriteString(quoteIdentifier(p.cursor))
		builder.WriteString(", ")
		builder.WriteString(quoteIdentifier(p.tieBreaker))
		builder.WriteString(") > ($1, $2)")
	}
	builder.WriteString(" ORDER BY ")
	builder.WriteString(quoteIdentifier(p.cursor))
	builder.WriteString(" ASC, ")
	builder.WriteString(quoteIdentifier(p.tieBreaker))
	builder.WriteString(" ASC")
	fmt.Fprintf(&builder, " LIMIT $%d", len(after)+1)
	return builder.String()
}

func (p postgresPollingReadPlan) queryArguments(after []any, limit int) []any {
	arguments := append([]any(nil), after...)
	return append(arguments, limit)
}

func (p postgresPollingReadPlan) items(rows pgx.Rows) ([]engine.PollingSourceItem, error) {
	items := make([]engine.PollingSourceItem, 0, p.pageSize+1)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("postgres polling read row: %w", err)
		}
		raw := rows.RawValues()
		record, err := (postgresSnapshotReadPlan{columns: p.columns}).recordForValues(values, raw)
		if err != nil {
			return nil, err
		}
		position, err := p.position(values, raw)
		if err != nil {
			return nil, err
		}
		items = append(items, engine.PollingSourceItem{Record: record, Position: position})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres polling iterate bounded page: %w", err)
	}
	return items, nil
}

func (p postgresPollingReadPlan) position(values []any, raw [][]byte) (synccontract.CheckpointPosition, error) {
	valuesByColumn := make(map[string]any, len(p.columns))
	rawByColumn := make(map[string][]byte, len(p.columns))
	for index, column := range p.columns {
		valuesByColumn[column.Ref.Name] = values[index]
		rawByColumn[column.Ref.Name] = raw[index]
	}
	cursor, found := postgresPollingColumn(p.columns, p.cursor)
	if !found {
		return synccontract.CheckpointPosition{}, ErrPollingCursorFieldNotFound
	}
	tie, found := postgresPollingColumn(p.columns, p.tieBreaker)
	if !found {
		return synccontract.CheckpointPosition{}, ErrPollingTieBreakerRequired
	}
	primary, err := postgresPollingEncodeToken(cursor, valuesByColumn[p.cursor], rawByColumn[p.cursor])
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	tieToken, err := postgresPollingEncodeToken(tie, valuesByColumn[p.tieBreaker], rawByColumn[p.tieBreaker])
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	return synccontract.CheckpointPosition{Primary: primary, TieBreaker: tieToken}, nil
}

func (p postgresPollingReadPlan) afterValues(position *synccontract.CheckpointPosition) ([]any, error) {
	if position == nil {
		return nil, nil
	}
	cursor, _ := postgresPollingColumn(p.columns, p.cursor)
	tie, _ := postgresPollingColumn(p.columns, p.tieBreaker)
	primary, err := postgresPollingDecodeToken(cursor, position.Primary)
	if err != nil {
		return nil, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL cursor tuple cannot be decoded")
	}
	tieValue, err := postgresPollingDecodeToken(tie, position.TieBreaker)
	if err != nil {
		return nil, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL tie-breaker tuple cannot be decoded")
	}
	return []any{primary, tieValue}, nil
}

func (p postgresPollingReadPlan) validateTraversal(after *synccontract.CheckpointPosition, page engine.PollingSourcePage) error {
	var previous *synccontract.CheckpointPosition
	if after != nil {
		copy := after.Clone()
		previous = &copy
	}
	for _, item := range page.Items {
		if previous != nil {
			comparison, err := p.comparePositions(*previous, item.Position)
			if err != nil || comparison >= 0 {
				return errors.New("postgres polling source did not advance its complete cursor tuple")
			}
		}
		copy := item.Position.Clone()
		previous = &copy
	}
	return nil
}

func (p postgresPollingReadPlan) comparePositions(left, right synccontract.CheckpointPosition) (int, error) {
	cursor, _ := postgresPollingColumn(p.columns, p.cursor)
	tie, _ := postgresPollingColumn(p.columns, p.tieBreaker)
	leftCursor, err := postgresPollingDecodeToken(cursor, left.Primary)
	if err != nil {
		return 0, err
	}
	rightCursor, err := postgresPollingDecodeToken(cursor, right.Primary)
	if err != nil {
		return 0, err
	}
	if comparison := postgresPollingCompareValue(leftCursor, rightCursor); comparison != 0 {
		return comparison, nil
	}
	leftTie, err := postgresPollingDecodeToken(tie, left.TieBreaker)
	if err != nil {
		return 0, err
	}
	rightTie, err := postgresPollingDecodeToken(tie, right.TieBreaker)
	if err != nil {
		return 0, err
	}
	return postgresPollingCompareValue(leftTie, rightTie), nil
}

func postgresPollingCompareValue(left, right any) int {
	switch a := left.(type) {
	case int64:
		b := right.(int64)
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	case uint64:
		b := right.(uint64)
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	case time.Time:
		b := right.(time.Time)
		if a.Before(b) {
			return -1
		}
		if a.After(b) {
			return 1
		}
	case string:
		b := right.(string)
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	}
	return 0
}

func postgresPollingEncodeToken(column database.Column, value any, raw []byte) (synccontract.OpaqueToken, error) {
	if value == nil {
		if column.Ref.Name == "" {
			return nil, ErrPollingCursorNullable
		}
		return nil, errors.New("postgres polling cursor tuple contains null")
	}
	switch column.Type.Kind() {
	case database.LogicalSignedInteger:
		switch typed := value.(type) {
		case int16:
			return synccontract.OpaqueToken(strconv.FormatInt(int64(typed), 10)), nil
		case int32:
			return synccontract.OpaqueToken(strconv.FormatInt(int64(typed), 10)), nil
		case int64:
			return synccontract.OpaqueToken(strconv.FormatInt(typed, 10)), nil
		}
	case database.LogicalUnsignedInteger:
		switch typed := value.(type) {
		case uint8:
			return synccontract.OpaqueToken(strconv.FormatUint(uint64(typed), 10)), nil
		case uint16:
			return synccontract.OpaqueToken(strconv.FormatUint(uint64(typed), 10)), nil
		case uint32:
			return synccontract.OpaqueToken(strconv.FormatUint(uint64(typed), 10)), nil
		case uint64:
			return synccontract.OpaqueToken(strconv.FormatUint(typed, 10)), nil
		}
	case database.LogicalTimestamp:
		if typed, ok := value.(time.Time); ok {
			return synccontract.OpaqueToken(typed.UTC().Format(time.RFC3339Nano)), nil
		}
	case database.LogicalUUID:
		switch typed := value.(type) {
		case string:
			return synccontract.OpaqueToken(typed), nil
		case [16]byte:
			return synccontract.OpaqueToken(pgtype.UUID{Bytes: typed, Valid: true}.String()), nil
		case pgtype.UUID:
			if typed.Valid {
				return synccontract.OpaqueToken(typed.String()), nil
			}
		}
	case database.LogicalString:
		if typed, ok := value.(string); ok {
			return synccontract.OpaqueToken(typed), nil
		}
	}
	if len(raw) > 0 && column.Type.Kind() == database.LogicalString {
		return append(synccontract.OpaqueToken(nil), raw...), nil
	}
	return nil, fmt.Errorf("postgres polling cursor tuple cannot encode %q", column.Ref.Name)
}

func postgresPollingDecodeToken(column database.Column, token synccontract.OpaqueToken) (any, error) {
	if len(token) == 0 {
		return nil, errors.New("empty tuple token")
	}
	value := string(token)
	switch column.Type.Kind() {
	case database.LogicalSignedInteger:
		return strconv.ParseInt(value, 10, 64)
	case database.LogicalUnsignedInteger:
		return strconv.ParseUint(value, 10, 64)
	case database.LogicalTimestamp:
		return time.Parse(time.RFC3339Nano, value)
	case database.LogicalUUID, database.LogicalString:
		return value, nil
	default:
		return nil, ErrPollingTieBreakerRequired
	}
}
