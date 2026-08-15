package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const postgresPollingSourceID = "postgres_polling_source_v1"

var postgresPollingSourceReference = connectors.TransportExecutorReference{
	Family: connectors.TransportExecutorFamilyNativeDatabase,
	ID:     postgresPollingSourceID,
}

// SharedPollingSourceExecutorReference exposes the exact inner shared-polling
// source selected by the PostgreSQL definition. The outward transport role
// remains postgres_bounded_snapshot so the closed synctransport registry is
// unchanged; incremental requests with a declared cursor delegate here.
func (*SnapshotTransportSource) SharedPollingSourceExecutorReference() connectors.TransportExecutorReference {
	return postgresPollingSourceReference
}

func (s *SnapshotTransportSource) readPollingTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	// SnapshotTransportSource already entered the production authentication
	// cohort before calling this adapter. Prevent catalog discovery and the
	// shared executor from recursively entering the same admission boundary.
	request.Runtime.AuthenticationAdmission = nil
	runner, declaration, object, err := s.bindPollingSource(ctx, request)
	if err != nil {
		return err
	}
	registry := engine.NewPollingPreflightRegistry()
	if err := registry.RegisterSource(runner); err != nil {
		return err
	}
	apply := &ManagedTargetTransportDestination{connector: s.connector}
	if err := registry.RegisterApply(apply); err != nil {
		return err
	}
	resolved, err := engine.PollingPreflight(ctx, registry, declaration, object, request.Mode)
	if err != nil {
		return fmt.Errorf("postgres polling transport preflight: %w", err)
	}
	shared, err := engine.NewPollingSourceExecutor(resolved)
	if err != nil {
		return fmt.Errorf("postgres polling transport shared executor: %w", err)
	}
	return shared.ReadTransport(ctx, request, emit)
}

func (s *SnapshotTransportSource) bindPollingSource(ctx context.Context, request synctransport.SourceRequest) (*postgresPollingSourceRunner, *connectors.PollingWatermarkDescriptor, connectors.PollingCatalogObject, error) {
	conn, relationRef, resources, pageSize, err := s.validateReadRequest(request)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, err
	}
	definition := s.connector.Definition()
	if definition.PollingWatermark == nil || definition.PollingWatermark.Status != connectors.PollingWatermarkStatusImplemented {
		return nil, nil, connectors.PollingCatalogObject{}, errors.New("postgres polling watermark definition is unavailable")
	}
	cursorField := strings.TrimSpace(request.CursorField)
	if cursorField == "" {
		return nil, nil, connectors.PollingCatalogObject{}, errors.New("postgres polling transport requires cursor_field")
	}
	catalog, err := s.connector.typedCatalog(ctx, request.Runtime)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, fmt.Errorf("postgres polling transport: discover typed catalog: %w", err)
	}
	plan, err := newPostgresPollingReadPlan(catalog, relationRef, cursorField, pageSize)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, err
	}
	if err := plan.validateRequestedPrimaryKey(request.PrimaryKey); err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, err
	}
	declaration, err := bindPostgresPollingDeclaration(definition.PollingWatermark, request.Resume, plan)
	if err != nil {
		return nil, nil, connectors.PollingCatalogObject{}, err
	}
	object := plan.pollingObject()
	runner := &postgresPollingSourceRunner{
		request:   request,
		conn:      conn,
		resources: resources,
		plan:      plan,
	}
	return runner, declaration, object, nil
}

func bindPostgresPollingDeclaration(template *connectors.PollingWatermarkDescriptor, resume synccontract.ResumeExpectation, plan postgresPollingReadPlan) (*connectors.PollingWatermarkDescriptor, error) {
	if template == nil {
		return nil, errors.New("postgres polling declaration template is required")
	}
	bound := template.Clone()
	bound.Source.Executor = postgresPollingSourceReference
	bound.Source.Ordering.Watermark.CatalogField = plan.cursor.Ref.Name
	bound.Source.Ordering.Watermark.CatalogFields = nil
	bound.Source.Ordering.TieBreaker.CatalogField = ""
	bound.Source.Ordering.TieBreaker.CatalogFields = plan.descriptorTieFields()
	bound.Source.Identity = connectors.PollingSourceIdentity{
		Engine:       resume.Source.Engine,
		AccountScope: resume.Source.AccountOrCluster,
		ObjectScope:  resume.Source.ObjectScope,
	}
	bound.Source.Cursor = plan.pollingCursor()
	bound.Target.Executor = postgresManagedTargetTransportReference
	bound.Target.StableKeyMapping = plan.primaryKeyNames()
	if err := bound.Validate(); err != nil {
		return nil, fmt.Errorf("postgres bound polling declaration: %w", err)
	}
	return bound, nil
}

func (d *ManagedTargetTransportDestination) PollingApplyExecutorReference() connectors.TransportExecutorReference {
	return postgresManagedTargetTransportReference
}

func (d *ManagedTargetTransportDestination) PollingApplyConformanceEvidence() engine.PollingWatermarkConformanceEvidence {
	return engine.RequiredPollingWatermarkConformanceEvidence()
}

type postgresPollingSourceRunner struct {
	request   synctransport.SourceRequest
	conn      connConfig
	resources typedCatalogResources
	plan      postgresPollingReadPlan
}

var _ engine.PollingSourceRunner = (*postgresPollingSourceRunner)(nil)

func (*postgresPollingSourceRunner) PollingSourceExecutorReference() connectors.TransportExecutorReference {
	return postgresPollingSourceReference
}

func (*postgresPollingSourceRunner) PollingSourceConformanceEvidence() engine.PollingWatermarkConformanceEvidence {
	return engine.RequiredPollingWatermarkConformanceEvidence()
}

func (r *postgresPollingSourceRunner) PollingSourceRuntimeState(ctx context.Context, object connectors.PollingCatalogObject) (engine.PollingSourceRuntimeState, error) {
	if r == nil || objectKey(object) != objectKey(r.plan.pollingObject()) {
		return engine.PollingSourceRuntimeState{}, errors.New("postgres polling runtime object does not match the bound catalog relation")
	}
	if err := ctx.Err(); err != nil {
		return engine.PollingSourceRuntimeState{}, err
	}
	fingerprint := r.plan.fingerprint.Bytes()
	barrierInput := append(append([]byte(nil), fingerprint...), r.request.Resume.SourceGeneration...)
	barrier := sha256.Sum256(barrierInput)
	return engine.PollingSourceRuntimeState{
		SourceGeneration: append(synccontract.OpaqueToken(nil), r.request.Resume.SourceGeneration...),
		SchemaVersion:    r.plan.fingerprint.String(),
		SnapshotBarrier:  synccontract.SnapshotBarrier{Kind: string(connectors.PollingSnapshotBarrierNone), Token: append(synccontract.OpaqueToken(nil), barrier[:]...)},
		Partitions:       []synccontract.PartitionState{},
		Dedupe:           synccontract.DedupeIdentity{Kind: "postgres_polling_relation", Value: append(synccontract.OpaqueToken(nil), fingerprint...)},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "postgres_polling_schema", Start: append(synccontract.OpaqueToken(nil), fingerprint...), End: append(synccontract.OpaqueToken(nil), fingerprint...)},
	}, nil
}

func (r *postgresPollingSourceRunner) FetchPollingSourcePage(ctx context.Context, request engine.PollingSourcePageRequest) (engine.PollingSourcePage, error) {
	if r == nil || objectKey(request.Object) != objectKey(r.plan.pollingObject()) {
		return engine.PollingSourcePage{}, errors.New("postgres polling fetch object does not match the bound catalog relation")
	}
	if err := request.RequestBudget.Consume(ctx); err != nil {
		return engine.PollingSourcePage{}, err
	}
	operationCtx, cancel, err := r.resources.operationContext(ctx)
	if err != nil {
		return engine.PollingSourcePage{}, err
	}
	defer cancel()
	pool, err := r.conn.openTypedCatalogPool(operationCtx, r.resources)
	if err != nil {
		return engine.PollingSourcePage{}, fmt.Errorf("postgres polling source: open pool: %w", err)
	}
	defer pool.Close()
	return r.plan.readPage(operationCtx, pool, request.After, request.PageSize)
}

func (r *postgresPollingSourceRunner) ValidatePollingSourcePageTraversal(ctx context.Context, after *synccontract.CheckpointPosition, page engine.PollingSourcePage) error {
	if r == nil {
		return errors.New("postgres polling traversal validator is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	previous := after
	for index := range page.Items {
		if previous != nil {
			comparison, err := r.plan.comparePositions(*previous, page.Items[index].Position)
			if err != nil {
				return err
			}
			if comparison >= 0 {
				return &engine.PollingWatermarkNonAdvancingError{
					Reason: engine.PollingWatermarkNonAdvancingReasonCursor,
					Page:   1,
					Source: "postgres",
				}
			}
		}
		position := page.Items[index].Position.Clone()
		previous = &position
	}
	return nil
}

type postgresPollingReadPlan struct {
	relation    database.Relation
	columns     []database.Column
	cursor      database.Column
	primaryKey  []database.Column
	order       []database.Column
	pageSize    int
	fingerprint database.SchemaFingerprint
}

func newPostgresPollingReadPlan(catalog database.Catalog, relationRef database.RelationRef, cursorField string, pageSize int) (postgresPollingReadPlan, error) {
	if pageSize <= 0 {
		return postgresPollingReadPlan{}, errors.New("postgres polling transport requires a finite page size")
	}
	for _, relation := range catalog.Relations() {
		if relation.Ref != relationRef {
			continue
		}
		columns := append([]database.Column(nil), relation.Columns...)
		sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
		cursor, found := postgresPollingColumn(columns, cursorField)
		if !found || cursor.Nullable {
			return postgresPollingReadPlan{}, errors.New("postgres polling cursor_field must name a non-null typed catalog column")
		}
		if err := postgresPollingCursorLogicalType(cursor.Type); err != nil {
			return postgresPollingReadPlan{}, err
		}
		key, err := postgresPollingPrimaryKey(relation, columns)
		if err != nil {
			return postgresPollingReadPlan{}, err
		}
		order := []database.Column{cursor}
		for _, column := range key {
			if column.Ref.Name != cursor.Ref.Name {
				order = append(order, column)
			}
		}
		if len(order) == 1 {
			return postgresPollingReadPlan{}, errors.New("postgres polling transport requires a primary-key tie field distinct from cursor_field")
		}
		return postgresPollingReadPlan{
			relation: relation, columns: columns, cursor: cursor, primaryKey: key,
			order: order, pageSize: pageSize, fingerprint: catalog.Fingerprint(),
		}, nil
	}
	return postgresPollingReadPlan{}, errors.New("postgres polling source relation is absent from the typed catalog")
}

func postgresPollingPrimaryKey(relation database.Relation, columns []database.Column) ([]database.Column, error) {
	for _, key := range relation.Keys {
		if key.Kind != database.KeyPrimary || len(key.Columns) == 0 {
			continue
		}
		resolved := make([]database.Column, 0, len(key.Columns))
		for _, ref := range key.Columns {
			column, found := postgresPollingColumn(columns, ref.Name)
			if !found || column.Nullable {
				return nil, errors.New("postgres polling primary key must be complete and non-null")
			}
			if err := postgresPollingKeyLogicalType(column.Type); err != nil {
				return nil, fmt.Errorf("postgres polling primary key %q: %w", column.Ref.Name, err)
			}
			resolved = append(resolved, column)
		}
		return resolved, nil
	}
	return nil, errors.New("postgres polling transport requires a declared primary key")
}

func postgresPollingColumn(columns []database.Column, name string) (database.Column, bool) {
	for _, column := range columns {
		if column.Ref.Name == name {
			return column, true
		}
	}
	return database.Column{}, false
}

func postgresPollingCursorLogicalType(logical database.LogicalType) error {
	switch logical.Kind() {
	case database.LogicalSignedInteger, database.LogicalUnsignedInteger, database.LogicalTimestamp:
		return nil
	default:
		return fmt.Errorf("postgres polling cursor_field type %q is not a lossless integer or timestamp cursor", logical.Kind())
	}
}

func postgresPollingKeyLogicalType(logical database.LogicalType) error {
	switch logical.Kind() {
	case database.LogicalSignedInteger, database.LogicalUnsignedInteger, database.LogicalBoolean,
		database.LogicalString, database.LogicalBinary, database.LogicalDate, database.LogicalTime,
		database.LogicalTimestamp, database.LogicalUUID:
		return nil
	default:
		return fmt.Errorf("typed catalog logical type %q is not a lossless polling key", logical.Kind())
	}
}

func (p postgresPollingReadPlan) pollingCursor() connectors.PollingCursor {
	if p.cursor.Type.Kind() == database.LogicalTimestamp {
		return connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, Precision: "nanosecond"}
	}
	return connectors.PollingCursor{Codec: connectors.PollingCursorCodecDecimal, Type: connectors.PollingCursorTypeInteger, Precision: "exact"}
}

func (p postgresPollingReadPlan) descriptorTieFields() []string {
	columns := p.tieColumns()
	fields := make([]string, 0, len(columns))
	for _, column := range columns {
		fields = append(fields, column.Ref.Name)
	}
	return fields
}

func (p postgresPollingReadPlan) tieColumns() []database.Column {
	return append([]database.Column(nil), p.order[1:]...)
}

func (p postgresPollingReadPlan) primaryKeyNames() []string {
	names := make([]string, 0, len(p.primaryKey))
	for _, column := range p.primaryKey {
		names = append(names, column.Ref.Name)
	}
	return names
}

func (p postgresPollingReadPlan) validateRequestedPrimaryKey(requested []string) error {
	want := p.primaryKeyNames()
	if len(requested) != len(want) {
		return errors.New("postgres polling destination primary key does not match the typed catalog")
	}
	for index := range want {
		if requested[index] != want[index] {
			return errors.New("postgres polling destination primary key does not match the typed catalog")
		}
	}
	return nil
}

func (p postgresPollingReadPlan) pollingObject() connectors.PollingCatalogObject {
	columns := make([]string, 0, len(p.columns))
	for _, column := range p.columns {
		columns = append(columns, column.Ref.Name)
	}
	return connectors.PollingCatalogObject{
		Kind:      connectors.PollingCatalogObjectRelation,
		NameParts: []string{p.relation.Ref.Schema.Catalog.Name, p.relation.Ref.Schema.Name, p.relation.Ref.Name},
		Columns:   columns,
	}
}

func objectKey(object connectors.PollingCatalogObject) string {
	return string(object.Kind) + "\x00" + strings.Join(object.NameParts, "\x00") + "\x00" + strings.Join(object.Columns, "\x00")
}

type postgresPollingQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (p postgresPollingReadPlan) readPage(ctx context.Context, queryer postgresPollingQuerier, after *synccontract.CheckpointPosition, pageSize int) (engine.PollingSourcePage, error) {
	if pageSize <= 0 || pageSize > p.pageSize {
		return engine.PollingSourcePage{}, errors.New("postgres polling page size exceeds the bound read plan")
	}
	arguments, err := p.afterArguments(after)
	if err != nil {
		return engine.PollingSourcePage{}, err
	}
	rows, err := queryer.Query(ctx, p.query(after != nil, len(arguments)), append(arguments, pageSize+1)...)
	if err != nil {
		return engine.PollingSourcePage{}, fmt.Errorf("postgres polling source: query keyset page: %w", err)
	}
	defer rows.Close()
	items := make([]engine.PollingSourceItem, 0, pageSize)
	more := false
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return engine.PollingSourcePage{}, fmt.Errorf("postgres polling source: read row: %w", err)
		}
		rawValues := rows.RawValues()
		if len(items) == pageSize {
			more = true
			continue
		}
		record, err := p.snapshotProjection().recordForValues(values, rawValues)
		if err != nil {
			return engine.PollingSourcePage{}, err
		}
		position, err := p.positionForValues(values, rawValues)
		if err != nil {
			return engine.PollingSourcePage{}, err
		}
		items = append(items, engine.PollingSourceItem{Record: record, Position: position})
	}
	if err := rows.Err(); err != nil {
		return engine.PollingSourcePage{}, fmt.Errorf("postgres polling source: iterate page: %w", err)
	}
	return engine.PollingSourcePage{Items: items, More: more, ObservedAt: time.Now().UTC()}, nil
}

func (p postgresPollingReadPlan) snapshotProjection() postgresSnapshotReadPlan {
	return postgresSnapshotReadPlan{relation: p.relation, columns: p.columns, pageSize: p.pageSize, fingerprint: p.fingerprint}
}

func (p postgresPollingReadPlan) query(hasAfter bool, argumentCount int) string {
	selected := make([]string, 0, len(p.columns))
	for _, column := range p.columns {
		selected = append(selected, quoteIdentifier(column.Ref.Name))
	}
	ordered := make([]string, 0, len(p.order))
	for _, column := range p.order {
		ordered = append(ordered, quoteIdentifier(column.Ref.Name))
	}
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(strings.Join(selected, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(quoteIdentifier(p.relation.Ref.Schema.Name))
	builder.WriteString(".")
	builder.WriteString(quoteIdentifier(p.relation.Ref.Name))
	if hasAfter {
		placeholders := make([]string, argumentCount)
		for index := range argumentCount {
			placeholders[index] = fmt.Sprintf("$%d", index+1)
		}
		builder.WriteString(" WHERE (")
		builder.WriteString(strings.Join(ordered, ", "))
		builder.WriteString(") > (")
		builder.WriteString(strings.Join(placeholders, ", "))
		builder.WriteString(")")
	}
	builder.WriteString(" ORDER BY ")
	for index, field := range ordered {
		ordered[index] = field + " ASC"
	}
	builder.WriteString(strings.Join(ordered, ", "))
	fmt.Fprintf(&builder, " LIMIT $%d", argumentCount+1)
	return builder.String()
}

func (p postgresPollingReadPlan) afterArguments(after *synccontract.CheckpointPosition) ([]any, error) {
	if after == nil {
		return nil, nil
	}
	cursor, err := decodePostgresPollingToken([]database.Column{p.cursor}, after.Primary)
	if err != nil {
		return nil, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL polling cursor token is invalid")
	}
	tieColumns := p.tieColumns()
	keys, err := decodePostgresPollingToken(tieColumns, after.TieBreaker)
	if err != nil {
		return nil, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL polling primary-key tuple is invalid")
	}
	byName := make(map[string]any, len(tieColumns))
	for index, column := range tieColumns {
		byName[column.Ref.Name] = keys[index]
	}
	arguments := []any{cursor[0]}
	for _, column := range p.order[1:] {
		arguments = append(arguments, byName[column.Ref.Name])
	}
	return arguments, nil
}

func (p postgresPollingReadPlan) positionForValues(values []any, raw [][]byte) (synccontract.CheckpointPosition, error) {
	cursorValues, err := p.valuesForColumns(values, raw, []database.Column{p.cursor})
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	tieColumns := p.tieColumns()
	keyValues, err := p.valuesForColumns(values, raw, tieColumns)
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	primary, err := encodePostgresPollingToken([]database.Column{p.cursor}, cursorValues)
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	tie, err := encodePostgresPollingToken(tieColumns, keyValues)
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	return synccontract.CheckpointPosition{Primary: primary, TieBreaker: tie}, nil
}

func (p postgresPollingReadPlan) valuesForColumns(values []any, raw [][]byte, wanted []database.Column) ([]any, error) {
	if len(values) != len(p.columns) || len(raw) != len(values) {
		return nil, errors.New("postgres polling row does not match the typed catalog projection")
	}
	byName := make(map[string]any, len(p.columns))
	for index, column := range p.columns {
		byName[column.Ref.Name] = values[index]
	}
	result := make([]any, 0, len(wanted))
	for _, column := range wanted {
		value, found := byName[column.Ref.Name]
		if !found || value == nil {
			return nil, errors.New("postgres polling cursor or primary-key row value is null")
		}
		result = append(result, value)
	}
	return result, nil
}

func encodePostgresPollingToken(columns []database.Column, values []any) (synccontract.OpaqueToken, error) {
	if len(columns) == 0 || len(columns) != len(values) {
		return nil, errors.New("postgres polling token shape is invalid")
	}
	text := make([]string, len(values))
	for index := range values {
		encoded, err := postgresPollingValueText(columns[index].Type, values[index])
		if err != nil {
			return nil, err
		}
		text[index] = encoded
	}
	payload, err := json.Marshal(text)
	if err != nil {
		return nil, err
	}
	return synccontract.OpaqueToken(payload), nil
}

func decodePostgresPollingToken(columns []database.Column, token synccontract.OpaqueToken) ([]any, error) {
	var text []string
	if len(token) == 0 || json.Unmarshal(token, &text) != nil || len(text) != len(columns) {
		return nil, errors.New("postgres polling token shape is invalid")
	}
	values := make([]any, len(text))
	for index := range text {
		value, err := postgresPollingParameter(columns[index].Type, text[index])
		if err != nil {
			return nil, err
		}
		values[index] = value
	}
	return values, nil
}

func postgresPollingValueText(logical database.LogicalType, value any) (string, error) {
	normalized, err := postgresSnapshotRecordValue(logical, value)
	if err != nil || normalized == nil {
		return "", errors.New("postgres polling value is not losslessly encodable")
	}
	switch typed := normalized.(type) {
	case string:
		return typed, nil
	case int16:
		return strconv.FormatInt(int64(typed), 10), nil
	case int32:
		return strconv.FormatInt(int64(typed), 10), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case uint8:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint64:
		return strconv.FormatUint(typed, 10), nil
	case bool:
		return strconv.FormatBool(typed), nil
	case []byte:
		return base64.RawStdEncoding.EncodeToString(typed), nil
	default:
		return "", fmt.Errorf("postgres polling value type %T is not losslessly encodable", normalized)
	}
}

func postgresPollingParameter(logical database.LogicalType, text string) (any, error) {
	switch logical.Kind() {
	case database.LogicalSignedInteger:
		return strconv.ParseInt(text, 10, 64)
	case database.LogicalUnsignedInteger:
		return strconv.ParseUint(text, 10, 64)
	case database.LogicalBoolean:
		return strconv.ParseBool(text)
	case database.LogicalString, database.LogicalDate, database.LogicalTime, database.LogicalUUID:
		return text, nil
	case database.LogicalBinary:
		return base64.RawStdEncoding.DecodeString(text)
	case database.LogicalTimestamp:
		if logical.WithTimezone() {
			return time.Parse(time.RFC3339Nano, text)
		}
		return time.Parse("2006-01-02T15:04:05.999999999", text)
	default:
		return nil, fmt.Errorf("postgres polling token type %q is unsupported", logical.Kind())
	}
}

func (p postgresPollingReadPlan) comparePositions(left, right synccontract.CheckpointPosition) (int, error) {
	leftCursor, err := decodePostgresPollingToken([]database.Column{p.cursor}, left.Primary)
	if err != nil {
		return 0, err
	}
	rightCursor, err := decodePostgresPollingToken([]database.Column{p.cursor}, right.Primary)
	if err != nil {
		return 0, err
	}
	if compared := comparePostgresPollingValue(p.cursor.Type, leftCursor[0], rightCursor[0]); compared != 0 {
		return compared, nil
	}
	tieColumns := p.tieColumns()
	leftKeys, err := decodePostgresPollingToken(tieColumns, left.TieBreaker)
	if err != nil {
		return 0, err
	}
	rightKeys, err := decodePostgresPollingToken(tieColumns, right.TieBreaker)
	if err != nil {
		return 0, err
	}
	for index, column := range tieColumns {
		if compared := comparePostgresPollingValue(column.Type, leftKeys[index], rightKeys[index]); compared != 0 {
			return compared, nil
		}
	}
	return 0, nil
}

func comparePostgresPollingValue(logical database.LogicalType, left, right any) int {
	switch logical.Kind() {
	case database.LogicalSignedInteger:
		return compareOrdered(left.(int64), right.(int64))
	case database.LogicalUnsignedInteger:
		return compareOrdered(left.(uint64), right.(uint64))
	case database.LogicalBoolean:
		return bytes.Compare([]byte(strconv.FormatBool(left.(bool))), []byte(strconv.FormatBool(right.(bool))))
	case database.LogicalBinary:
		return bytes.Compare(left.([]byte), right.([]byte))
	case database.LogicalTimestamp:
		return left.(time.Time).Compare(right.(time.Time))
	default:
		return strings.Compare(left.(string), right.(string))
	}
}

func compareOrdered[T ~int64 | ~uint64](left, right T) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
