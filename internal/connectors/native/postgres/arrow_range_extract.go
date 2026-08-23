package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// ExtractArrowRanges is PostgreSQL's bounded range-extraction primitive for
// the generic transformed fast path. It is deliberately a source adapter: it
// owns PostgreSQL's snapshot transaction, identifiers and pgx row decoding,
// while synctransport owns transform, Parquet, credits and checkpoints.
func (s *PollingTransportSource) ExtractArrowRanges(ctx context.Context, request synctransport.ArrowExtractRequest, emit func(synctransport.ArrowSourceBatch) error) error {
	if s == nil || ctx == nil || emit == nil || request.TransformHash == "" || request.TransformPlanJSON == "" || request.BatchSize <= 0 || request.UnitDeadline <= 0 {
		return synctransport.ErrArrowFastPathInvalid
	}
	if err := s.validateRequestBeforeIO(ctx, synctransport.SourceRequest{
		Connector: request.Connector, Runtime: request.Runtime, Stream: request.Stream, CursorField: request.CursorField,
		Mode: synccontract.ModeFullOverwrite, BatchSize: request.BatchSize, PrimaryKey: request.PrimaryKey,
		Resume: request.Resume, Checkpoint: request.Checkpoint, UnitDeadline: request.UnitDeadline,
	}, func(synctransport.SourcePage) error { return nil }); err != nil {
		return err
	}
	transform, err := database.ParseTransformPlanV1([]byte(request.TransformPlanJSON))
	if err != nil || transform.Hash() != request.TransformHash {
		return synctransport.ErrArrowFastPathInvalid
	}
	return executeWithAuthenticationAdmission(ctx, request.Runtime, func(admitted context.Context) error {
		return s.extractArrowRanges(admitted, request, transform, emit)
	})
}

func (s *PollingTransportSource) extractArrowRanges(ctx context.Context, request synctransport.ArrowExtractRequest, transform database.TransformPlanV1, emit func(synctransport.ArrowSourceBatch) error) error {
	if fixtureMode(request.Runtime) {
		return errors.New("PostgreSQL Arrow range extraction requires a live typed catalog")
	}
	legacyRequest := synctransport.SourceRequest{
		Connector: request.Connector, Runtime: request.Runtime, Stream: request.Stream, CursorField: request.CursorField,
		Mode: synccontract.ModeFullOverwrite, BatchSize: request.BatchSize, PrimaryKey: request.PrimaryKey,
		Resume: request.Resume, Checkpoint: request.Checkpoint, UnitDeadline: request.UnitDeadline,
	}
	runnerValue, declaration, _, closeRunner, err := s.preparePollingRunner(ctx, legacyRequest)
	if err != nil {
		return err
	}
	defer closeRunner()
	runner, ok := runnerValue.(*postgresPollingSourceRunner)
	if !ok || runner == nil {
		return synctransport.ErrArrowFastPathInvalid
	}
	plan, err := newPostgresArrowRangePlan(runner.plan, transform)
	if err != nil {
		return err
	}
	config, err := resolveConfig(request.Runtime)
	if err != nil {
		return err
	}
	pgxConfig, err := config.dataConfig()
	if err != nil {
		return err
	}
	if err := checkPostgresRequestAdmission(ctx); err != nil {
		return err
	}
	connection, err := pgx.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return fmt.Errorf("postgres Arrow range extraction: connect source: %w", err)
	}
	defer func() { _ = connection.Close(context.Background()) }()
	tx, err := connection.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("postgres Arrow range extraction: begin repeatable-read snapshot: %w", err)
	}
	tx = admitPostgresTx(tx)
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	var after *synccontract.CheckpointPosition
	if request.Checkpoint != nil {
		position := request.Checkpoint.Position.Clone()
		after = &position
	}
	for page := 0; page < declaration.Source.Read.MaxPages; page++ {
		batch, next, more, err := plan.readArrowRange(ctx, tx, after, request.UnitDeadline, request.Resume.Source, runner.state)
		if err != nil {
			return err
		}
		if err := emit(batch); err != nil {
			batch.Record.Release()
			return err
		}
		batch.Record.Release()
		if !more {
			return nil
		}
		position := next.Clone()
		after = &position
	}
	return fmt.Errorf("PostgreSQL Arrow range extraction reached declared maximum of %d pages", declaration.Source.Read.MaxPages)
}

type postgresArrowRangePlan struct {
	polling            postgresPollingReadPlan
	schema             *arrow.Schema
	recordQueryIndexes []int
}

func newPostgresArrowRangePlan(base postgresPollingReadPlan, transform database.TransformPlanV1) (postgresArrowRangePlan, error) {
	fields, err := transform.SourceFields()
	if err != nil {
		return postgresArrowRangePlan{}, err
	}
	recordFields := make(map[string]struct{}, len(fields))
	required := make(map[string]struct{})
	for _, field := range fields {
		recordFields[field] = struct{}{}
		required[field] = struct{}{}
	}
	required[base.cursor] = struct{}{}
	required[base.tieBreaker] = struct{}{}
	columns := make([]database.Column, 0, len(required))
	for _, column := range base.columns {
		if _, needed := required[column.Ref.Name]; needed {
			if _, err := postgresArrowType(column.Type); err != nil {
				return postgresArrowRangePlan{}, synctransport.ErrArrowFastPathInvalid
			}
			columns = append(columns, column)
		}
	}
	if len(columns) != len(required) {
		return postgresArrowRangePlan{}, synctransport.ErrArrowFastPathInvalid
	}
	sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
	arrowFields := make([]arrow.Field, 0, len(recordFields))
	recordQueryIndexes := make([]int, 0, len(recordFields))
	for index, column := range columns {
		if _, include := recordFields[column.Ref.Name]; !include {
			continue
		}
		dataType, err := postgresArrowType(column.Type)
		if err != nil {
			return postgresArrowRangePlan{}, synctransport.ErrArrowFastPathInvalid
		}
		arrowFields = append(arrowFields, arrow.Field{Name: column.Ref.Name, Type: dataType, Nullable: column.Nullable})
		recordQueryIndexes = append(recordQueryIndexes, index)
	}
	if len(recordQueryIndexes) != len(recordFields) {
		return postgresArrowRangePlan{}, synctransport.ErrArrowFastPathInvalid
	}
	base.columns = columns
	return postgresArrowRangePlan{polling: base, schema: arrow.NewSchema(arrowFields, nil), recordQueryIndexes: recordQueryIndexes}, nil
}

func (p postgresArrowRangePlan) readArrowRange(ctx context.Context, tx pgx.Tx, after *synccontract.CheckpointPosition, deadline time.Duration, source synccontract.SourceIdentity, state engine.PollingSourceRuntimeState) (synctransport.ArrowSourceBatch, synccontract.CheckpointPosition, bool, error) {
	arguments, err := p.polling.afterValues(after)
	if err != nil {
		return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, err
	}
	unitCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	started := time.Now()
	rows, err := tx.Query(unitCtx, p.polling.query(arguments, p.polling.pageSize+1), p.polling.queryArguments(arguments, p.polling.pageSize+1)...)
	if err != nil {
		return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, fmt.Errorf("postgres Arrow range query: %w", err)
	}
	defer rows.Close()
	builder := array.NewRecordBuilder(memory.DefaultAllocator, p.schema)
	defer builder.Release()
	var last synccontract.CheckpointPosition
	count := 0
	more := false
	for rows.Next() {
		if count == p.polling.pageSize {
			more = true
			break
		}
		values, err := rows.Values()
		if err != nil {
			return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, fmt.Errorf("postgres Arrow range values: %w", err)
		}
		raw := rows.RawValues()
		if len(values) != len(p.polling.columns) || len(raw) != len(values) {
			return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, synctransport.ErrArrowFastPathInvalid
		}
		for outputIndex, queryIndex := range p.recordQueryIndexes {
			column := p.polling.columns[queryIndex]
			if err := appendPostgresArrowValue(builder.Field(outputIndex), column.Type, values[queryIndex]); err != nil {
				return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, err
			}
		}
		last, err = p.position(values, raw)
		if err != nil {
			return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, fmt.Errorf("postgres Arrow range iterate: %w", err)
	}
	record := builder.NewRecordBatch()
	candidate, err := postgresArrowRangeCheckpoint(source, state, after, last, count, time.Now().UTC())
	if err != nil {
		record.Release()
		return synctransport.ArrowSourceBatch{}, synccontract.CheckpointPosition{}, false, err
	}
	return synctransport.ArrowSourceBatch{Record: record, SourceLogicalBytes: postgresArrowRecordBytes(record), SourceRows: int64(count), ExtractElapsed: time.Since(started), CandidateCheckpoint: candidate}, last, more, nil
}

func (p postgresArrowRangePlan) position(values []any, raw [][]byte) (synccontract.CheckpointPosition, error) {
	cursorIndex, tieIndex := -1, -1
	for index, column := range p.polling.columns {
		switch column.Ref.Name {
		case p.polling.cursor:
			cursorIndex = index
		case p.polling.tieBreaker:
			tieIndex = index
		}
	}
	if cursorIndex < 0 || tieIndex < 0 {
		return synccontract.CheckpointPosition{}, synctransport.ErrArrowFastPathInvalid
	}
	cursor, _ := postgresPollingColumn(p.polling.columns, p.polling.cursor)
	tie, _ := postgresPollingColumn(p.polling.columns, p.polling.tieBreaker)
	primary, err := postgresPollingEncodeToken(cursor, values[cursorIndex], raw[cursorIndex])
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	tieToken, err := postgresPollingEncodeToken(tie, values[tieIndex], raw[tieIndex])
	if err != nil {
		return synccontract.CheckpointPosition{}, err
	}
	return synccontract.CheckpointPosition{Primary: primary, TieBreaker: tieToken}, nil
}

func postgresArrowRangeCheckpoint(source synccontract.SourceIdentity, state engine.PollingSourceRuntimeState, after *synccontract.CheckpointPosition, last synccontract.CheckpointPosition, count int, observedAt time.Time) (synccontract.CheckpointEnvelope, error) {
	positionObserved := count > 0 || after != nil
	position := last.Clone()
	if count == 0 && after != nil {
		position = after.Clone()
	}
	barrier := synccontract.SnapshotBarrier{Kind: state.SnapshotBarrier.Kind, Token: append(synccontract.OpaqueToken(nil), state.SnapshotBarrier.Token...)}
	candidate := synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source:       source,
		Mechanism:    engine.PollingSourceCheckpointMechanism, SnapshotBarrier: &barrier,
		Position: position, PositionObserved: &positionObserved, Partitions: clonePostgresArrowPartitions(state.Partitions),
		SourceGeneration: append(synccontract.OpaqueToken(nil), state.SourceGeneration...), SchemaVersion: state.SchemaVersion,
		ProtocolVersion: engine.PollingSourceProtocolVersion, Dedupe: state.Dedupe.Clone(), DedupeWindow: state.DedupeWindow.Clone(), ObservedAt: observedAt,
	}
	// The source identity was preflighted from the connector's closed
	// declaration. Copy it through exact values rather than deriving it from a
	// PostgreSQL connection string or query.
	if candidate.Source != source || candidate.Source.Engine != postgresSnapshotSourceEngine {
		return synccontract.CheckpointEnvelope{}, synctransport.ErrArrowFastPathInvalid
	}
	if err := candidate.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	return candidate, nil
}

func clonePostgresArrowPartitions(values []synccontract.PartitionState) []synccontract.PartitionState {
	if values == nil {
		return nil
	}
	result := make([]synccontract.PartitionState, len(values))
	for index := range values {
		result[index] = values[index].Clone()
	}
	return result
}

func postgresArrowType(logical database.LogicalType) (arrow.DataType, error) {
	switch logical.Kind() {
	case database.LogicalSignedInteger:
		return arrow.PrimitiveTypes.Int64, nil
	case database.LogicalString:
		return arrow.BinaryTypes.String, nil
	case database.LogicalTimestamp:
		return arrow.FixedWidthTypes.Timestamp_ns, nil
	case database.LogicalDate:
		return arrow.PrimitiveTypes.Date32, nil
	default:
		return nil, synctransport.ErrArrowFastPathInvalid
	}
}

func appendPostgresArrowValue(builder array.Builder, logical database.LogicalType, value any) error {
	if value == nil {
		builder.AppendNull()
		return nil
	}
	switch typed := builder.(type) {
	case *array.Int64Builder:
		var integer int64
		switch value := value.(type) {
		case int64:
			integer = value
		case int32:
			integer = int64(value)
		case int16:
			integer = int64(value)
		case int8:
			integer = int64(value)
		default:
			return synctransport.ErrArrowFastPathInvalid
		}
		typed.Append(integer)
		return nil
	case *array.StringBuilder:
		text, ok := value.(string)
		if !ok {
			return synctransport.ErrArrowFastPathInvalid
		}
		typed.Append(text)
		return nil
	case *array.TimestampBuilder:
		instant, ok := value.(time.Time)
		if !ok {
			return synctransport.ErrArrowFastPathInvalid
		}
		typed.Append(arrow.Timestamp(instant.UnixNano()))
		return nil
	case *array.Date32Builder:
		instant, ok := value.(time.Time)
		if !ok {
			return synctransport.ErrArrowFastPathInvalid
		}
		days := int(instant.UTC().Sub(time.Unix(0, 0).UTC()) / (24 * time.Hour))
		typed.Append(arrow.Date32(days))
		return nil
	default:
		return synctransport.ErrArrowFastPathInvalid
	}
}

func postgresArrowRecordBytes(record arrow.RecordBatch) int64 {
	var total uint64
	for index := 0; index < int(record.NumCols()); index++ {
		total += record.Column(index).Data().SizeInBytes()
	}
	if total > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1)
	}
	return int64(total)
}

var _ synctransport.ArrowRangeExtractor = (*PollingTransportSource)(nil)
