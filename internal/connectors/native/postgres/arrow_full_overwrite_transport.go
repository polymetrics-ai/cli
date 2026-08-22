package postgres

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// BeginArrowFullOverwrite is PostgreSQL's native binary-COPY implementation
// of the connector-neutral transformed full-overwrite port. PostgreSQL types,
// pgx and SQL remain in this adapter; synctransport sees only Arrow records,
// segment receipts and one durable final acknowledgement.
func (d *ManagedTargetTransportDestination) BeginArrowFullOverwrite(ctx context.Context, request synctransport.ArrowFullOverwriteRunRequest) (synctransport.ArrowFullOverwriteRun, error) {
	if ctx == nil || d == nil {
		return nil, ErrManagedTargetTransportBindingInvalid
	}
	return executeManagedTargetWithAuthenticationAdmission(ctx, request.Runtime, func(admitted context.Context) (synctransport.ArrowFullOverwriteRun, error) {
		return d.beginArrowFullOverwrite(admitted, request)
	})
}

func (d *ManagedTargetTransportDestination) beginArrowFullOverwrite(ctx context.Context, request synctransport.ArrowFullOverwriteRunRequest) (synctransport.ArrowFullOverwriteRun, error) {
	if ctx == nil || d == nil || request.Plan.ApplyStrategy.Strategy != connectors.ApplyStrategyReplace || strings.TrimSpace(request.TransformPlanJSON) == "" || request.TransformPlanHash != request.Plan.TransformPlanHash {
		return nil, ErrManagedTargetTransportBindingInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateManagedTargetTransportApproval(request.Approval); err != nil {
		return nil, err
	}
	if err := validateManagedTargetTransportApprovalAt(request.Approval, d.currentTime()); err != nil {
		return nil, err
	}
	if err := validateManagedTargetTransportBinding(request.Binding); err != nil || request.ConnectionID != request.Binding.ConnectionID || request.BatchSize <= 0 || strings.TrimSpace(request.Stream) == "" {
		return nil, ErrManagedTargetTransportBindingInvalid
	}
	transform, err := database.ParseTransformPlanV1([]byte(request.TransformPlanJSON))
	if err != nil || transform.Hash() != request.TransformPlanHash {
		return nil, ErrManagedTargetTransportBindingInvalid
	}
	resolved, err := d.resolveManagedTargetWithTransform(ctx, request.Source, request.SourceRuntime, request.Runtime, request.Binding, request.Stream, &transform)
	if err != nil {
		return nil, err
	}
	control, err := resolved.provision(ctx)
	if err != nil {
		resolved.close()
		return nil, err
	}
	if err := postgresEnsureFullOverwriteReceiptTable(ctx, resolved.driver, control); err != nil {
		resolved.close()
		return nil, err
	}
	shadow, err := postgresArrowFullOverwriteShadowName(control)
	if err != nil {
		resolved.close()
		return nil, ErrManagedTargetTransportUnavailable
	}
	return &managedTargetArrowFullOverwriteRun{destination: d, resolved: resolved, control: control, request: request, shadow: shadow, segments: make(map[string]synctransport.FastSegmentReceipt)}, nil
}

type managedTargetArrowFullOverwriteRun struct {
	destination *ManagedTargetTransportDestination
	resolved    managedTargetTransportResolution
	control     database.ManagedTargetControlRecord
	request     synctransport.ArrowFullOverwriteRunRequest
	shadow      string

	mu                     sync.Mutex
	shadowPrepared         bool
	indexConstraintElapsed time.Duration
	published              bool
	closed                 bool
	segments               map[string]synctransport.FastSegmentReceipt
	receipt                database.FullOverwriteReceiptV1
}

func (s *managedTargetArrowFullOverwriteRun) ApplyArrowSegment(ctx context.Context, request synctransport.ArrowBulkApplyRequest) error {
	if ctx == nil || s == nil {
		return ErrManagedTargetTransportBindingInvalid
	}
	_, err := executeManagedTargetWithAuthenticationAdmission(ctx, s.request.Runtime, func(admitted context.Context) (struct{}, error) {
		return struct{}{}, s.applyArrowSegment(admitted, request)
	})
	return err
}

func (s *managedTargetArrowFullOverwriteRun) applyArrowSegment(ctx context.Context, request synctransport.ArrowBulkApplyRequest) error {
	if ctx == nil || s == nil || ctx.Err() != nil || request.Record == nil || request.ConnectionID != s.request.ConnectionID || request.Plan != s.request.Plan || request.Segment.TransformPlanHash != s.request.TransformPlanHash || request.Segment.TransformPlanHash == "" || request.Segment.TransformedRows != request.Record.NumRows() || request.Segment.ParquetBytes < 1 {
		return ErrManagedTargetTransportBindingInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.published {
		return ErrManagedTargetTransportBindingInvalid
	}
	if _, exists := s.segments[request.Segment.ID]; exists {
		return ErrManagedTargetTransportBindingInvalid
	}
	if err := s.ensureShadowLocked(ctx); err != nil {
		return err
	}
	if request.Record.NumRows() > 0 {
		columns, err := postgresManagedTargetColumns(s.resolved.mapping)
		if err != nil {
			return ErrManagedTargetTransportBindingInvalid
		}
		names := make([]string, len(columns))
		for index, column := range columns {
			names[index] = column.name
		}
		s.resolved.driver.connMu.Lock()
		copied, err := s.resolved.conn.CopyFrom(ctx, pgx.Identifier{s.control.Target().Namespace(), s.shadow}, names, newPostgresArrowCopyFromSource(request.Record))
		s.resolved.driver.connMu.Unlock()
		if err != nil || copied != request.Record.NumRows() {
			return ErrManagedTargetTransportUnavailable
		}
	}
	s.segments[request.Segment.ID] = request.Segment
	return nil
}

func (s *managedTargetArrowFullOverwriteRun) ensureShadowLocked(ctx context.Context) error {
	if s.shadowPrepared {
		return nil
	}
	if s.resolved.conn == nil || s.resolved.driver == nil {
		return ErrManagedTargetTransportUnavailable
	}
	qualified := postgresManagedTargetQualifiedRelation(s.control)
	qualifiedShadow := quoteIdentifier(s.control.Target().Namespace()) + "." + quoteIdentifier(s.shadow)
	s.resolved.driver.connMu.Lock()
	defer s.resolved.driver.connMu.Unlock()
	started := time.Now()
	if _, err := s.resolved.conn.Exec(ctx, "CREATE TABLE "+qualifiedShadow+" (LIKE "+qualified+" INCLUDING ALL)"); err != nil {
		return ErrManagedTargetTransportUnavailable
	}
	s.indexConstraintElapsed += time.Since(started)
	s.shadowPrepared = true
	return nil
}

func (s *managedTargetArrowFullOverwriteRun) ArrowBulkPhaseMeasurement() synctransport.ArrowBulkPhaseMeasurement {
	if s == nil {
		return synctransport.ArrowBulkPhaseMeasurement{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return synctransport.ArrowBulkPhaseMeasurement{IndexConstraintBuildElapsed: s.indexConstraintElapsed}
}

func (s *managedTargetArrowFullOverwriteRun) PublishArrowFullOverwrite(ctx context.Context, request synctransport.ArrowFullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	if ctx == nil || s == nil {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	return executeManagedTargetWithAuthenticationAdmission(ctx, s.request.Runtime, func(admitted context.Context) (synccontract.DownstreamAcknowledgement, error) {
		return s.publishArrowFullOverwrite(admitted, request)
	})
}

func (s *managedTargetArrowFullOverwriteRun) publishArrowFullOverwrite(ctx context.Context, request synctransport.ArrowFullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	if ctx == nil || s == nil || ctx.Err() != nil || request.SourceLogicalBytes < 0 || request.SourceRows < 0 || request.TransformedRows < 0 || request.TransformedBytes < 0 {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.published || len(request.Segments) != len(s.segments) {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	for _, segment := range request.Segments {
		stored, ok := s.segments[segment.ID]
		if !ok || stored != segment {
			return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
		}
	}
	var transformedRows, transformedBytes int64
	for _, segment := range request.Segments {
		transformedRows += segment.TransformedRows
		transformedBytes += segment.TransformedBytes
	}
	if transformedRows != request.TransformedRows || transformedBytes != request.TransformedBytes || request.SourceRows < request.TransformedRows {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	if !s.shadowPrepared {
		if err := s.ensureShadowLocked(ctx); err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	receipt, err := s.newArrowReceipt(request)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if acknowledgedAt, found, err := s.lookupArrowReceipt(ctx, receipt); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	} else if found {
		s.receipt, s.published = receipt, true
		return synccontract.NewDurableDownstreamAcknowledgement(s.destination.connector.Name(), acknowledgedAt)
	}
	if err := s.publishArrowShadow(ctx, receipt); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	s.receipt, s.published = receipt, true
	return synccontract.NewDurableDownstreamAcknowledgement(s.destination.connector.Name(), receipt.PublishedAt)
}

func (s *managedTargetArrowFullOverwriteRun) ReadBackArrowFullOverwrite(ctx context.Context, acknowledgement synccontract.DownstreamAcknowledgement) error {
	if ctx == nil || s == nil {
		return ErrManagedTargetTransportReadBackFailed
	}
	_, err := executeManagedTargetWithAuthenticationAdmission(ctx, s.request.Runtime, func(admitted context.Context) (struct{}, error) {
		return struct{}{}, s.readBackArrowFullOverwrite(admitted, acknowledgement)
	})
	return err
}

func (s *managedTargetArrowFullOverwriteRun) readBackArrowFullOverwrite(ctx context.Context, acknowledgement synccontract.DownstreamAcknowledgement) error {
	if ctx == nil || s == nil || ctx.Err() != nil || acknowledgement.Sink != s.destination.connector.Name() || acknowledgement.AcknowledgedAt.IsZero() {
		return ErrManagedTargetTransportReadBackFailed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || !s.published || s.receipt.Validate() != nil {
		return ErrManagedTargetTransportReadBackFailed
	}
	_, found, err := s.lookupArrowReceipt(ctx, s.receipt)
	s.closeLocked()
	if err != nil || !found {
		return ErrManagedTargetTransportReadBackFailed
	}
	return nil
}

func (s *managedTargetArrowFullOverwriteRun) AbortArrowFullOverwrite(_ context.Context) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked()
	return nil
}

func (s *managedTargetArrowFullOverwriteRun) closeLocked() {
	if s.closed {
		return
	}
	s.closed = true
	s.resolved.close()
}

func (s *managedTargetArrowFullOverwriteRun) newArrowReceipt(request synctransport.ArrowFullOverwritePublicationRequest) (database.FullOverwriteReceiptV1, error) {
	checkpoint := ""
	if request.LastCheckpoint != nil {
		encoded, err := json.Marshal(request.LastCheckpoint.Clone())
		if err != nil {
			return database.FullOverwriteReceiptV1{}, ErrManagedTargetTransportBindingInvalid
		}
		checkpoint = hashFullOverwriteBytes(encoded)
	}
	if checkpoint == "" {
		checkpoint = hashFullOverwriteBytes([]byte("empty-arrow-full-overwrite-checkpoint-v1"))
	}
	contents := make([]string, len(request.Segments))
	for index, segment := range request.Segments {
		contents[index] = segment.ContentSHA256
	}
	sort.Strings(contents)
	content := hashFullOverwriteBytes([]byte(strings.Join(contents, "\x00")))
	plan := postgresFullOverwritePlanHash(s.control, s.resolved.mapping, s.request.TransformPlanHash)
	return database.NewFullOverwriteReceiptV1(plan, checkpoint, content, request.TransformedRows, s.destination.currentTime())
}

func (s *managedTargetArrowFullOverwriteRun) lookupArrowReceipt(ctx context.Context, receipt database.FullOverwriteReceiptV1) (time.Time, bool, error) {
	if receipt.Validate() != nil || s.resolved.conn == nil {
		return time.Time{}, false, ErrManagedTargetTransportReadBackFailed
	}
	query := `SELECT plan_hash, checkpoint_hash, content_hash, records, published_at FROM ` + postgresQualifiedControlTable(s.control.Target().Namespace(), postgresFullOverwriteReceiptTable) + ` WHERE receipt_id = $1`
	var planHash, checkpointHash, contentHash string
	var records int64
	var publishedAt time.Time
	err := s.resolved.conn.QueryRow(ctx, query, receipt.ID).Scan(&planHash, &checkpointHash, &contentHash, &records, &publishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil || planHash != receipt.PlanHash || checkpointHash != receipt.CheckpointHash || contentHash != receipt.ContentHash || records != receipt.Records {
		return time.Time{}, false, ErrManagedTargetTransportReadBackFailed
	}
	return publishedAt.UTC(), true, nil
}

func (s *managedTargetArrowFullOverwriteRun) publishArrowShadow(ctx context.Context, receipt database.FullOverwriteReceiptV1) error {
	if s.resolved.conn == nil || s.resolved.driver == nil || !s.shadowPrepared {
		return ErrManagedTargetTransportUnavailable
	}
	qualified := postgresManagedTargetQualifiedRelation(s.control)
	qualifiedShadow := quoteIdentifier(s.control.Target().Namespace()) + "." + quoteIdentifier(s.shadow)
	old, err := postgresArrowOldRelationName(s.control, receipt.ID)
	if err != nil {
		return ErrManagedTargetTransportUnavailable
	}
	qualifiedOld := quoteIdentifier(s.control.Target().Namespace()) + "." + quoteIdentifier(old)
	s.resolved.driver.connMu.Lock()
	defer s.resolved.driver.connMu.Unlock()
	if err := postgresPreflightDurability(ctx, s.resolved.conn); err != nil {
		return err
	}
	if err := checkPostgresRequestAdmission(ctx); err != nil {
		return err
	}
	tx, err := s.resolved.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrManagedTargetTransportUnavailable
	}
	tx = admitPostgresTx(tx)
	rollback := func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }
	if _, err := tx.Exec(ctx, "SET LOCAL synchronous_commit = 'on'"); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	if _, err := tx.Exec(ctx, "LOCK TABLE "+qualified+" IN ACCESS EXCLUSIVE MODE"); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	if _, err := tx.Exec(ctx, "ALTER TABLE "+qualified+" RENAME TO "+quoteIdentifier(old)); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	if _, err := tx.Exec(ctx, "ALTER TABLE "+qualifiedShadow+" RENAME TO "+quoteIdentifier(s.control.Target().Relation())); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	if err := postgresUpdateManagedTargetRelationOID(ctx, tx, s.control); err != nil {
		rollback()
		return err
	}
	qualifiedFence := postgresQualifiedControlTable(s.control.Target().Namespace(), postgresOrderFenceTable)
	if _, err := tx.Exec(ctx, "DELETE FROM "+qualifiedFence+" WHERE relation_name = $1", s.control.Target().Relation()); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	receiptTable := postgresQualifiedControlTable(s.control.Target().Namespace(), postgresFullOverwriteReceiptTable)
	if _, err := tx.Exec(ctx, `INSERT INTO `+receiptTable+` (receipt_id, relation_name, plan_hash, checkpoint_hash, content_hash, records, published_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`, receipt.ID, s.control.Target().Relation(), receipt.PlanHash, receipt.CheckpointHash, receipt.ContentHash, receipt.Records, receipt.PublishedAt); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	if _, err := tx.Exec(ctx, "DROP TABLE "+qualifiedOld); err != nil {
		rollback()
		return ErrManagedTargetTransportUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return ErrManagedTargetTransportUnavailable
	}
	return nil
}

func postgresUpdateManagedTargetRelationOID(ctx context.Context, tx pgx.Tx, control database.ManagedTargetControlRecord) error {
	table := postgresQualifiedControlTable(control.Target().Namespace(), postgresTargetControlTable)
	query := `UPDATE ` + table + ` AS control
	SET relation_oid = relation.oid::text
	FROM pg_catalog.pg_class AS relation
	JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
	WHERE control.relation_name = $1
	  AND relation.relname = $1
	  AND namespace.nspname = $2
	  AND relation.relkind IN ('r', 'p')`
	result, err := tx.Exec(ctx, query, control.Target().Relation(), control.Target().Namespace())
	if err != nil || result.RowsAffected() != 1 {
		return ErrManagedTargetTransportUnavailable
	}
	return nil
}

func postgresArrowFullOverwriteShadowName(control database.ManagedTargetControlRecord) (string, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	digest := hashFullOverwriteBytes(append([]byte("postgres-arrow-shadow-v1\x00"+control.Target().Namespace()+"\x00"+control.Target().Relation()+"\x00"), nonce[:]...))
	return "pmsh_" + digest[:48], nil
}

func postgresArrowOldRelationName(control database.ManagedTargetControlRecord, receiptID string) (string, error) {
	if len(receiptID) < 24 {
		return "", ErrManagedTargetTransportBindingInvalid
	}
	digest := hashFullOverwriteBytes([]byte("postgres-arrow-old-v1\x00" + control.Target().Namespace() + "\x00" + control.Target().Relation() + "\x00" + receiptID))
	return "pmold_" + digest[:48], nil
}

var _ synctransport.ArrowBulkDestination = (*ManagedTargetTransportDestination)(nil)
var _ synctransport.ArrowFullOverwriteRun = (*managedTargetArrowFullOverwriteRun)(nil)
var _ synctransport.ArrowBulkPhaseReporter = (*managedTargetArrowFullOverwriteRun)(nil)
