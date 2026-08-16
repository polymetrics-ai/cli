package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

// BeginFullOverwrite implements the connector-neutral run-scoped replacement
// port. The PostgreSQL-specific shadow relation and publish transaction remain
// entirely here; synctransport sees only staged worksets and a durable receipt.
func (d *ManagedTargetTransportDestination) BeginFullOverwrite(ctx context.Context, request synctransport.FullOverwriteRunRequest) (synctransport.FullOverwriteRun, error) {
	if ctx == nil || d == nil || request.Mode != synccontract.ModeFullOverwrite || request.Plan.ApplyStrategy.Strategy != connectors.ApplyStrategyReplace {
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
	if (strings.TrimSpace(request.TransformPlanJSON) == "") != (strings.TrimSpace(request.TransformPlanHash) == "") || request.TransformPlanHash != request.Plan.TransformPlanHash {
		return nil, ErrManagedTargetTransportBindingInvalid
	}
	// The legacy map-backed stage is intentionally not a transformed fast path.
	// Reject a configured transform here until the Arrow segment route owns this
	// port, rather than silently applying source rows under output-column names.
	if request.TransformPlanHash != "" {
		return nil, ErrManagedTargetTransportBindingInvalid
	}
	resolved, err := d.resolveManagedTarget(ctx, request.Source, request.SourceRuntime, request.Runtime, request.Binding, request.Stream)
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
	return &managedTargetFullOverwriteRun{destination: d, resolved: resolved, control: control, request: request}, nil
}

type managedTargetFullOverwriteRun struct {
	destination *ManagedTargetTransportDestination
	resolved    managedTargetTransportResolution
	control     database.ManagedTargetControlRecord
	request     synctransport.FullOverwriteRunRequest

	mu        sync.Mutex
	pages     []managedTargetFullOverwritePage
	published bool
	closed    bool
	receipt   database.FullOverwriteReceiptV1
}

type managedTargetFullOverwritePage struct {
	receipt       synctransport.WarehouseReceipt
	sourceParquet string
	records       int
}

func (s *managedTargetFullOverwriteRun) ApplyFullOverwrite(ctx context.Context, request synctransport.DestinationApplyRequest) error {
	if ctx == nil || s == nil || ctx.Err() != nil || request.Mode != synccontract.ModeFullOverwrite || request.Plan.ApplyStrategy != s.request.Plan.ApplyStrategy || request.Plan.TransformPlanHash != s.request.Plan.TransformPlanHash || request.ConnectionID != s.request.ConnectionID || request.Workset.ID != request.Receipt.ID || request.Receipt.Validate() != nil || request.Workset.SourceParquet == "" || len(request.Workset.Tombstones) != 0 {
		return ErrManagedTargetTransportBindingInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.published || request.Receipt.Owner != s.request.Binding.ConnectionID || request.Receipt.Records != len(request.Workset.Records) {
		return ErrManagedTargetTransportBindingInvalid
	}
	s.pages = append(s.pages, managedTargetFullOverwritePage{receipt: request.Receipt, sourceParquet: request.Workset.SourceParquet, records: len(request.Workset.Records)})
	return nil
}

func (s *managedTargetFullOverwriteRun) PublishFullOverwrite(ctx context.Context, request synctransport.FullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error) {
	if ctx == nil || s == nil || ctx.Err() != nil || request.Tombstones != 0 || request.Records < 0 || request.Pages < 0 {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.published || request.Pages != len(s.pages) {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	records := 0
	for _, page := range s.pages {
		records += page.records
	}
	if records != request.Records {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportBindingInvalid
	}
	receipt, err := s.newReceipt(request)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if acknowledgedAt, found, err := s.lookupReceipt(ctx, receipt); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	} else if found {
		s.receipt, s.published = receipt, true
		return synccontract.NewDurableDownstreamAcknowledgement(s.destination.connector.Name(), acknowledgedAt)
	}
	if err := s.publishShadow(ctx, receipt); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	s.receipt, s.published = receipt, true
	return synccontract.NewDurableDownstreamAcknowledgement(s.destination.connector.Name(), receipt.PublishedAt)
}

func (s *managedTargetFullOverwriteRun) ReadBackFullOverwrite(ctx context.Context, acknowledgement synccontract.DownstreamAcknowledgement) error {
	if ctx == nil || s == nil || ctx.Err() != nil || acknowledgement.Sink != s.destination.connector.Name() || acknowledgement.AcknowledgedAt.IsZero() {
		return ErrManagedTargetTransportReadBackFailed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || !s.published || s.receipt.Validate() != nil {
		return ErrManagedTargetTransportReadBackFailed
	}
	_, found, err := s.lookupReceipt(ctx, s.receipt)
	s.closeLocked()
	if err != nil || !found {
		return ErrManagedTargetTransportReadBackFailed
	}
	return nil
}

func (s *managedTargetFullOverwriteRun) AbortFullOverwrite(_ context.Context) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked()
	return nil
}

func (s *managedTargetFullOverwriteRun) closeLocked() {
	if s.closed {
		return
	}
	s.closed = true
	s.resolved.close()
}

func (s *managedTargetFullOverwriteRun) newReceipt(request synctransport.FullOverwritePublicationRequest) (database.FullOverwriteReceiptV1, error) {
	checkpoint := ""
	if request.LastCheckpoint != nil {
		encoded, err := json.Marshal(request.LastCheckpoint.Clone())
		if err != nil {
			return database.FullOverwriteReceiptV1{}, ErrManagedTargetTransportBindingInvalid
		}
		checkpoint = hashFullOverwriteBytes(encoded)
	}
	if checkpoint == "" {
		checkpoint = hashFullOverwriteBytes([]byte("empty-full-overwrite-checkpoint-v1"))
	}
	contents := make([]string, len(s.pages))
	for index, page := range s.pages {
		contents[index] = page.receipt.ContentSHA256
	}
	sort.Strings(contents)
	content := hashFullOverwriteBytes([]byte(strings.Join(contents, "\x00")))
	plan := postgresFullOverwritePlanHash(s.control, s.resolved.mapping, s.request.Plan.TransformPlanHash)
	return database.NewFullOverwriteReceiptV1(plan, checkpoint, content, int64(request.Records), s.destination.currentTime())
}

func (s *managedTargetFullOverwriteRun) lookupReceipt(ctx context.Context, receipt database.FullOverwriteReceiptV1) (time.Time, bool, error) {
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

func (s *managedTargetFullOverwriteRun) publishShadow(ctx context.Context, receipt database.FullOverwriteReceiptV1) error {
	if s.resolved.conn == nil || s.resolved.driver == nil {
		return ErrManagedTargetTransportUnavailable
	}
	columns, err := postgresManagedTargetColumns(s.resolved.mapping)
	if err != nil {
		return ErrManagedTargetTransportBindingInvalid
	}
	shadow := postgresFullOverwriteShadowName(s.control, receipt.ID)
	qualified := postgresManagedTargetQualifiedRelation(s.control)
	qualifiedShadow := quoteIdentifier(s.control.Target().Namespace()) + "." + quoteIdentifier(shadow)
	s.resolved.driver.connMu.Lock()
	defer s.resolved.driver.connMu.Unlock()
	if err := postgresPreflightDurability(ctx, s.resolved.conn); err != nil {
		return err
	}
	tx, err := s.resolved.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return postgresFullOverwriteUnavailable("begin transaction", err)
	}
	rollback := func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }
	if _, err := tx.Exec(ctx, "SET LOCAL synchronous_commit = 'on'"); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("enable synchronous commit", err)
	}
	if _, err := tx.Exec(ctx, "CREATE TABLE "+qualifiedShadow+" (LIKE "+qualified+" INCLUDING DEFAULTS INCLUDING IDENTITY INCLUDING GENERATED)"); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("create shadow relation", err)
	}
	for _, page := range s.pages {
		rows := 0
		err := warehouse.ReadTable(ctx, page.sourceParquet, func(row warehouse.Row) error {
			mapped, err := s.resolved.mapping.MapRecord(row)
			if err != nil {
				return err
			}
			args, err := postgresMappedArguments(columns, mapped)
			if err != nil {
				return err
			}
			if err := postgresInsertMappedRow(ctx, tx, qualifiedShadow, columns, args); err != nil {
				return err
			}
			rows++
			return nil
		})
		if err != nil || rows != page.records {
			rollback()
			if err != nil {
				return postgresFullOverwriteUnavailable("stage workset page", err)
			}
			return fmt.Errorf("%w: staged workset rows=%d want=%d", ErrManagedTargetTransportUnavailable, rows, page.records)
		}
	}
	if _, err := tx.Exec(ctx, "LOCK TABLE "+qualified+" IN ACCESS EXCLUSIVE MODE"); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("lock target relation", err)
	}
	if _, err := tx.Exec(ctx, "TRUNCATE TABLE "+qualified); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("truncate target relation", err)
	}
	names, _ := postgresColumnNamesAndPlaceholders(columns, 1)
	if _, err := tx.Exec(ctx, "INSERT INTO "+qualified+" ("+strings.Join(names, ", ")+") SELECT "+strings.Join(names, ", ")+" FROM "+qualifiedShadow); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("replace target relation", err)
	}
	qualifiedFence := postgresQualifiedControlTable(s.control.Target().Namespace(), postgresOrderFenceTable)
	if _, err := tx.Exec(ctx, "DELETE FROM "+qualifiedFence+" WHERE relation_name = $1", s.control.Target().Relation()); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("clear order fence", err)
	}
	receiptTable := postgresQualifiedControlTable(s.control.Target().Namespace(), postgresFullOverwriteReceiptTable)
	if _, err := tx.Exec(ctx, `INSERT INTO `+receiptTable+` (receipt_id, relation_name, plan_hash, checkpoint_hash, content_hash, records, published_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`, receipt.ID, s.control.Target().Relation(), receipt.PlanHash, receipt.CheckpointHash, receipt.ContentHash, receipt.Records, receipt.PublishedAt); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("record receipt", err)
	}
	if _, err := tx.Exec(ctx, "DROP TABLE "+qualifiedShadow); err != nil {
		rollback()
		return postgresFullOverwriteUnavailable("drop shadow relation", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return postgresFullOverwriteUnavailable("commit replacement", err)
	}
	return nil
}

// postgresFullOverwriteUnavailable preserves the closed public failure class
// while carrying the safe database operation that failed. Parameters are
// always bound, so no record values or credentials enter this error path.
func postgresFullOverwriteUnavailable(operation string, err error) error {
	return fmt.Errorf("%w: %s: %v", ErrManagedTargetTransportUnavailable, operation, err)
}

func postgresFullOverwritePlanHash(control database.ManagedTargetControlRecord, mapping database.MappingContractV1, transformHash string) string {
	columns := mapping.Columns()
	builder := strings.Builder{}
	builder.WriteString("postgres-full-overwrite-plan-v1\x00")
	builder.WriteString(control.TargetDatabase().Kind())
	builder.WriteString("\x00")
	builder.WriteString(control.TargetDatabase().Value())
	builder.WriteString("\x00")
	builder.WriteString(control.Target().Namespace())
	builder.WriteString("\x00")
	builder.WriteString(control.Target().Relation())
	builder.WriteString("\x00")
	builder.WriteString(transformHash)
	for _, column := range columns {
		builder.WriteString("\x00")
		builder.WriteString(column.Source)
		builder.WriteString("\x00")
		builder.WriteString(column.Target)
	}
	return hashFullOverwriteBytes([]byte(builder.String()))
}

func postgresFullOverwriteShadowName(control database.ManagedTargetControlRecord, receiptID string) string {
	digest := sha256.Sum256([]byte("postgres-shadow-v1\x00" + control.Target().Namespace() + "\x00" + control.Target().Relation() + "\x00" + receiptID))
	return "pmsh_" + hex.EncodeToString(digest[:])[:48]
}

func hashFullOverwriteBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func postgresEnsureFullOverwriteReceiptTable(ctx context.Context, driver *DatabaseDriver, control database.ManagedTargetControlRecord) error {
	if ctx == nil || driver == nil || driver.conn == nil || driver.connMu == nil {
		return ErrManagedTargetTransportUnavailable
	}
	driver.connMu.Lock()
	defer driver.connMu.Unlock()
	table := postgresQualifiedControlTable(control.Target().Namespace(), postgresFullOverwriteReceiptTable)
	_, err := driver.conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS `+table+` (
		receipt_id TEXT PRIMARY KEY,
		relation_name TEXT NOT NULL,
		plan_hash TEXT NOT NULL,
		checkpoint_hash TEXT NOT NULL,
		content_hash TEXT NOT NULL,
		records BIGINT NOT NULL CHECK (records >= 0),
		published_at TIMESTAMPTZ NOT NULL
	)`)
	if err != nil {
		return ErrManagedTargetTransportUnavailable
	}
	return nil
}

var _ synctransport.FullOverwriteDestination = (*ManagedTargetTransportDestination)(nil)
var _ synctransport.FullOverwriteRun = (*managedTargetFullOverwriteRun)(nil)
