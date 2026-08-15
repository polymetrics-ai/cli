package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

var (
	errPostgresWriteTargetUnverified = errors.New("postgres managed target cannot be verified for write")
	errPostgresWriteSessionFailed    = errors.New("postgres managed target write session failed")
	errPostgresWriteValueInvalid     = errors.New("postgres managed target write value is invalid")
)

// DatabaseWriteCapabilities reports PostgreSQL's transactional full-overwrite
// guarantee. The public connector capability remains false until #3978
// certification; this private driver capability is consumed only by the shared
// write-session executor.
func (*DatabaseDriver) DatabaseWriteCapabilities() database.DatabaseWriteCapabilities {
	return database.DatabaseWriteCapabilities{AtomicFullOverwrite: true}
}

// PreviewDatabaseWrite proves that the sealed control record and exact mapped
// target relation still agree with PostgreSQL before an approval can be issued.
// Its opaque ID is a digest, never a relation or credential.
func (d *DatabaseDriver) PreviewDatabaseWrite(ctx context.Context, plan database.DatabaseWritePlan) (database.DatabaseWritePreview, error) {
	if ctx == nil {
		return database.DatabaseWritePreview{}, errPostgresWriteTargetUnverified
	}
	if err := ctx.Err(); err != nil {
		return database.DatabaseWritePreview{}, err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return database.DatabaseWritePreview{}, errPostgresDatabaseDriverConnectionRequired
	}
	d.connMu.Lock()
	defer d.connMu.Unlock()
	if err := postgresPreflightDurability(ctx, d.conn); err != nil {
		return database.DatabaseWritePreview{}, err
	}
	if err := d.assertManagedTargetForWrite(ctx, d.conn, plan); err != nil {
		return database.DatabaseWritePreview{}, err
	}
	return database.NewDatabaseWritePreview(plan, postgresWritePreviewID(plan))
}

// BeginDatabaseWrite opens one pinned PostgreSQL transaction. The driver's
// connection mutex deliberately remains held until commit or rollback: a pgx
// connection cannot safely interleave the transaction with an ownership query,
// ledger write, or another target session.
func (d *DatabaseDriver) BeginDatabaseWrite(ctx context.Context, plan database.DatabaseWritePlan) (database.WriteSession, error) {
	if ctx == nil {
		return nil, errPostgresWriteSessionFailed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return nil, errPostgresDatabaseDriverConnectionRequired
	}
	columns, err := postgresManagedTargetColumns(plan.Mapping())
	if err != nil {
		return nil, err
	}
	d.connMu.Lock()
	released := false
	handoff := false
	release := func() {
		if !released {
			released = true
			d.connMu.Unlock()
		}
	}
	defer func() {
		if !handoff {
			release()
		}
	}()
	if err := postgresPreflightDurability(ctx, d.conn); err != nil {
		return nil, err
	}
	if err := d.assertManagedTargetForWrite(ctx, d.conn, plan); err != nil {
		return nil, err
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errPostgresWriteSessionFailed
	}
	rollback := func() {
		_ = tx.Rollback(context.WithoutCancel(ctx))
	}
	if _, err := tx.Exec(ctx, "SET LOCAL synchronous_commit = 'on'"); err != nil {
		rollback()
		return nil, errPostgresWriteSessionFailed
	}
	if err := postgresPreflightDurability(ctx, tx); err != nil {
		rollback()
		return nil, err
	}
	qualified := postgresManagedTargetQualifiedRelation(plan.Control())
	historyAt := time.Time{}
	if plan.Mode() == synccontract.ModeIncrementalDedupeHistory {
		if err := postgresEnsureManagedTargetHistoryLayout(ctx, tx, plan); err != nil {
			rollback()
			return nil, errPostgresWriteSessionFailed
		}
		historyAt = time.Now().UTC()
	}
	qualifiedFence := postgresQualifiedControlTable(plan.Control().Target().Namespace(), postgresOrderFenceTable)
	if plan.Mode() == synccontract.ModeFullOverwrite {
		if _, err := tx.Exec(ctx, "TRUNCATE TABLE "+qualified); err != nil {
			rollback()
			return nil, errPostgresWriteSessionFailed
		}
		if _, err := tx.Exec(ctx, "DELETE FROM "+qualifiedFence+" WHERE relation_name = $1", plan.Control().Target().Relation()); err != nil {
			rollback()
			return nil, errPostgresWriteSessionFailed
		}
	}
	if postgresWriteModeRequiresTableLock(plan.Mode()) {
		if _, err := tx.Exec(ctx, "LOCK TABLE "+qualified+" IN SHARE ROW EXCLUSIVE MODE"); err != nil {
			rollback()
			return nil, errPostgresWriteSessionFailed
		}
	}
	session := &postgresWriteSession{
		tx:             tx,
		plan:           plan,
		columns:        columns,
		qualified:      qualified,
		qualifiedFence: qualifiedFence,
		release:        release,
		historyAt:      historyAt,
		nextBatch:      1,
	}
	handoff = true // session now owns the exact matching unlock.
	return session, nil
}

type postgresWriteSession struct {
	tx             pgx.Tx
	plan           database.DatabaseWritePlan
	columns        []postgresManagedTargetColumn
	qualified      string
	qualifiedFence string
	release        func()
	historyAt      time.Time

	mu                sync.Mutex
	closed            bool
	published         bool
	nextBatch         uint64
	recordsApplied    int
	tombstonesApplied int
}

func (s *postgresWriteSession) ApplyWriteBatch(ctx context.Context, batch database.WriteBatch) error {
	if ctx == nil {
		return errPostgresWriteSessionFailed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.tx == nil {
		return errPostgresWriteSessionFailed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || batch.Sequence() != s.nextBatch {
		return errPostgresWriteSessionFailed
	}
	records, tombstones := batch.OrderedRecords(), batch.Tombstones()
	if (len(records) == 0 && len(tombstones) == 0) || len(records)+len(tombstones) > s.plan.BatchSize() || s.recordsApplied+len(records) > s.plan.RecordCount() || s.tombstonesApplied+len(tombstones) > s.plan.TombstoneCount() {
		return errPostgresWriteSessionFailed
	}
	for _, record := range records {
		if err := s.applyRecord(ctx, record); err != nil {
			return err
		}
	}
	for _, tombstone := range tombstones {
		if err := s.applyTombstone(ctx, tombstone); err != nil {
			return err
		}
	}
	s.recordsApplied += len(records)
	s.tombstonesApplied += len(tombstones)
	s.nextBatch++
	return nil
}

func (s *postgresWriteSession) PublishFullOverwrite(ctx context.Context) error {
	if ctx == nil {
		return errPostgresWriteSessionFailed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return errPostgresWriteSessionFailed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.plan.Mode() != synccontract.ModeFullOverwrite || s.recordsApplied != s.plan.RecordCount() || s.tombstonesApplied != 0 {
		return errPostgresWriteSessionFailed
	}
	// The destructive TRUNCATE ran inside this same transaction at session
	// open. Marking publication only after every bounded batch makes commit the
	// single atomic visibility point; rollback restores the prior table.
	s.published = true
	return nil
}

func (s *postgresWriteSession) CommitWrite(ctx context.Context) (database.CommitOutcome, database.DeliveryReceiptV1, error) {
	if ctx == nil || s == nil || s.tx == nil {
		return database.CommitOutcomeUnknown, database.DeliveryReceiptV1{}, errPostgresWriteSessionFailed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return database.CommitOutcomeUnknown, database.DeliveryReceiptV1{}, errPostgresWriteSessionFailed
	}
	if s.recordsApplied != s.plan.RecordCount() || s.tombstonesApplied != s.plan.TombstoneCount() || (s.plan.Mode() == synccontract.ModeFullOverwrite && !s.published) {
		s.closed = true
		err := s.tx.Rollback(context.WithoutCancel(ctx))
		s.releaseConnection()
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			return database.CommitOutcomeUnknown, database.DeliveryReceiptV1{}, errPostgresWriteSessionFailed
		}
		return database.CommitOutcomeRolledBack, database.DeliveryReceiptV1{}, nil
	}
	deliveryID, err := postgresDeliveryID()
	if err != nil {
		s.closed = true
		_ = s.tx.Rollback(context.WithoutCancel(ctx))
		s.releaseConnection()
		return database.CommitOutcomeRolledBack, database.DeliveryReceiptV1{}, errPostgresWriteSessionFailed
	}
	err = s.tx.Commit(ctx)
	s.closed = true
	s.releaseConnection()
	if err != nil {
		// PostgreSQL may have committed after the client lost the response. The
		// shared executor must surface Unknown and never issue a blind retry.
		return database.CommitOutcomeUnknown, database.DeliveryReceiptV1{}, errPostgresWriteSessionFailed
	}
	receipt, err := database.NewDeliveryReceiptV1(s.plan, deliveryID, time.Now().UTC())
	if err != nil {
		return database.CommitOutcomeUnknown, database.DeliveryReceiptV1{}, errPostgresWriteSessionFailed
	}
	return database.CommitOutcomeCommitted, receipt, nil
}

func (s *postgresWriteSession) RollbackWrite(ctx context.Context) error {
	if ctx == nil || s == nil || s.tx == nil {
		return errPostgresWriteSessionFailed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errPostgresWriteSessionFailed
	}
	err := s.tx.Rollback(context.WithoutCancel(ctx))
	s.closed = true
	s.releaseConnection()
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return errPostgresWriteSessionFailed
	}
	return nil
}

func (s *postgresWriteSession) releaseConnection() {
	if s.release != nil {
		s.release()
		s.release = nil
	}
}

func (s *postgresWriteSession) applyRecord(ctx context.Context, record database.OrderedRecord) error {
	mapped, err := s.plan.Mapping().MapRecord(record.Record)
	if err != nil {
		return errPostgresWriteValueInvalid
	}
	args, err := postgresMappedArguments(s.columns, mapped)
	if err != nil {
		return errPostgresWriteValueInvalid
	}
	switch s.plan.Mode() {
	case synccontract.ModeFullOverwrite, synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend:
		return postgresInsertMappedRow(ctx, s.tx, s.qualified, s.columns, args)
	case synccontract.ModeIncrementalUpsert:
		if s.plan.ConditionalOrderFence() {
			return s.applyFencedRecord(ctx, mapped, args, record.Position)
		}
		if err := postgresDeleteMappedKeys(ctx, s.tx, s.qualified, s.plan.Keys(), mapped, s.columns); err != nil {
			return errPostgresWriteSessionFailed
		}
		if err := postgresInsertMappedRow(ctx, s.tx, s.qualified, s.columns, args); err != nil {
			return errPostgresWriteSessionFailed
		}
		return nil
	case synccontract.ModeIncrementalDedupe:
		if s.plan.ConditionalOrderFence() {
			return s.applyFencedRecord(ctx, mapped, args, record.Position)
		}
		if err := postgresInsertMappedRowIfAbsent(ctx, s.tx, s.qualified, s.plan.Keys(), s.columns, mapped, args); err != nil {
			return errPostgresWriteSessionFailed
		}
		return nil
	case synccontract.ModeIncrementalDedupeHistory:
		return s.applyFencedRecord(ctx, mapped, args, record.Position)
	default:
		return errPostgresWriteSessionFailed
	}
}

func (s *postgresWriteSession) applyFencedRecord(ctx context.Context, mapped connectors.Record, args []any, position synccontract.CheckpointPosition) error {
	keyDigest, err := postgresMappedKeyDigest(s.plan.Keys(), mapped, s.columns)
	if err != nil {
		return errPostgresWriteValueInvalid
	}
	if s.plan.Mode() == synccontract.ModeIncrementalDedupeHistory {
		return s.applyFencedHistoryRecord(ctx, mapped, args, keyDigest, position)
	}
	accepted, err := postgresOrderFenceAccepts(ctx, s.tx, s.qualifiedFence, s.plan.Control().Target().Relation(), keyDigest, position)
	if err != nil {
		return errPostgresWriteSessionFailed
	}
	if !accepted {
		return nil
	}
	switch s.plan.Mode() {
	case synccontract.ModeIncrementalUpsert:
		if err := postgresDeleteMappedKeys(ctx, s.tx, s.qualified, s.plan.Keys(), mapped, s.columns); err != nil {
			return errPostgresWriteSessionFailed
		}
		if err := postgresInsertMappedRow(ctx, s.tx, s.qualified, s.columns, args); err != nil {
			return errPostgresWriteSessionFailed
		}
	case synccontract.ModeIncrementalDedupe:
		if err := postgresInsertMappedRowIfAbsent(ctx, s.tx, s.qualified, s.plan.Keys(), s.columns, mapped, args); err != nil {
			return errPostgresWriteSessionFailed
		}
	default:
		return errPostgresWriteSessionFailed
	}
	if err := postgresStoreOrderFence(ctx, s.tx, s.qualifiedFence, s.plan.Control().Target().Relation(), keyDigest, position, false); err != nil {
		return errPostgresWriteSessionFailed
	}
	return nil
}

func (s *postgresWriteSession) applyFencedHistoryRecord(ctx context.Context, mapped connectors.Record, args []any, keyDigest []byte, position synccontract.CheckpointPosition) error {
	accepted, err := postgresOrderFenceAccepts(ctx, s.tx, s.qualifiedFence, s.plan.Control().Target().Relation(), keyDigest, position)
	if err != nil {
		return errPostgresWriteSessionFailed
	}
	if !accepted {
		return nil
	}
	if err := s.applyHistoryRecord(ctx, mapped, args); err != nil {
		return err
	}
	if err := postgresStoreOrderFence(ctx, s.tx, s.qualifiedFence, s.plan.Control().Target().Relation(), keyDigest, position, false); err != nil {
		return errPostgresWriteSessionFailed
	}
	return nil
}

func (s *postgresWriteSession) applyTombstone(ctx context.Context, tombstone synccontract.Tombstone) error {
	if s.plan.TombstoneCount() == 0 || !postgresWriteModeRequiresTableLock(s.plan.Mode()) || tombstone.Operation != synccontract.OperationDelete || tombstone.Validate() != nil {
		return errPostgresWriteSessionFailed
	}
	keys, err := postgresTombstoneKeyValues(tombstone, s.plan.Keys(), s.columns)
	if err != nil {
		return errPostgresWriteValueInvalid
	}
	if s.plan.ConditionalOrderFence() {
		keyDigest, err := postgresKeyValuesDigest(s.plan.Keys(), keys, s.columns)
		if err != nil {
			return errPostgresWriteValueInvalid
		}
		accepted, err := postgresOrderFenceAccepts(ctx, s.tx, s.qualifiedFence, s.plan.Control().Target().Relation(), keyDigest, tombstone.Position)
		if err != nil {
			return errPostgresWriteSessionFailed
		}
		if !accepted {
			return nil
		}
		if s.plan.Mode() == synccontract.ModeIncrementalDedupeHistory {
			if err := s.closeHistoryWindow(ctx, tombstone, keys); err != nil {
				return err
			}
		} else if err := postgresDeleteKeyValues(ctx, s.tx, s.qualified, s.plan.Keys(), keys); err != nil {
			return errPostgresWriteSessionFailed
		}
		if err := postgresStoreOrderFence(ctx, s.tx, s.qualifiedFence, s.plan.Control().Target().Relation(), keyDigest, tombstone.Position, true); err != nil {
			return errPostgresWriteSessionFailed
		}
		return nil
	}
	if s.plan.Mode() == synccontract.ModeIncrementalDedupeHistory {
		return errPostgresWriteSessionFailed
	}
	if err := postgresDeleteKeyValues(ctx, s.tx, s.qualified, s.plan.Keys(), keys); err != nil {
		return errPostgresWriteSessionFailed
	}
	return nil
}

// applyHistoryRecord preserves every observed keyed version. A duplicate or a
// late replay of an already-observed version is a no-op; it can never reopen a
// previously closed window. A new version closes the prior current window and
// inserts the new current row within this same pinned transaction.
func (s *postgresWriteSession) applyHistoryRecord(ctx context.Context, mapped connectors.Record, args []any) error {
	if s.historyAt.IsZero() {
		return errPostgresWriteSessionFailed
	}
	exists, err := postgresHistoryVersionExists(ctx, s.tx, s.qualified, s.plan.Keys(), s.columns, mapped, args)
	if err != nil {
		return errPostgresWriteSessionFailed
	}
	if exists {
		return nil
	}
	keyArgs, keyPredicate, err := postgresMappedKeyPredicate(s.plan.Keys(), mapped, s.columns, 1)
	if err != nil {
		return errPostgresWriteSessionFailed
	}
	valuePredicate := postgresMappedValuePredicate(s.columns, len(keyArgs)+1)
	closeArgs := append(append([]any{}, keyArgs...), args...)
	closeArgs = append(closeArgs, s.historyAt)
	if _, err := s.tx.Exec(ctx,
		"UPDATE "+s.qualified+
			" SET "+quoteIdentifier(synccontract.HistoryValidToColumn)+" = $"+strconv.Itoa(len(closeArgs))+", "+quoteIdentifier(synccontract.HistoryIsCurrentColumn)+" = FALSE"+
			" WHERE "+keyPredicate+" AND "+quoteIdentifier(synccontract.HistoryIsCurrentColumn)+" AND NOT ("+valuePredicate+")",
		closeArgs...,
	); err != nil {
		return errPostgresWriteSessionFailed
	}
	names, placeholders := postgresColumnNamesAndPlaceholders(s.columns, 1)
	names = append(names,
		quoteIdentifier(synccontract.HistoryValidFromColumn),
		quoteIdentifier(synccontract.HistoryValidToColumn),
		quoteIdentifier(synccontract.HistoryIsCurrentColumn),
	)
	placeholders = append(placeholders, "$"+strconv.Itoa(len(args)+1), "NULL", "TRUE")
	insertArgs := append(append([]any{}, args...), s.historyAt)
	if _, err := s.tx.Exec(ctx,
		"INSERT INTO "+s.qualified+" ("+strings.Join(names, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")",
		insertArgs...,
	); err != nil {
		return errPostgresWriteSessionFailed
	}
	return nil
}

func postgresHistoryVersionExists(ctx context.Context, tx pgx.Tx, qualified string, keys []string, columns []postgresManagedTargetColumn, mapped connectors.Record, args []any) (bool, error) {
	keyArgs, keyPredicate, err := postgresMappedKeyPredicate(keys, mapped, columns, 1)
	if err != nil {
		return false, err
	}
	valuePredicate := postgresMappedValuePredicate(columns, len(keyArgs)+1)
	queryArgs := append(append([]any{}, keyArgs...), args...)
	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM "+qualified+" WHERE "+keyPredicate+" AND "+valuePredicate+")", queryArgs...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *postgresWriteSession) closeHistoryWindow(ctx context.Context, tombstone synccontract.Tombstone, keys map[string]any) error {
	if s.historyAt.IsZero() {
		return errPostgresWriteSessionFailed
	}
	close, err := synccontract.CloseHistoryWindow(tombstone, s.historyAt)
	if err != nil || close.Action != synccontract.HistoryDeleteCloseValidityWindow || close.IsCurrent {
		return errPostgresWriteSessionFailed
	}
	args, keyPredicate, err := postgresKeyValuePredicate(s.plan.Keys(), keys, 1)
	if err != nil {
		return errPostgresWriteSessionFailed
	}
	args = append(args, close.ValidTo, close.IsCurrent)
	if _, err := s.tx.Exec(ctx,
		"UPDATE "+s.qualified+
			" SET "+quoteIdentifier(synccontract.HistoryValidToColumn)+" = $"+strconv.Itoa(len(args)-1)+", "+quoteIdentifier(synccontract.HistoryIsCurrentColumn)+" = $"+strconv.Itoa(len(args))+
			" WHERE "+keyPredicate+" AND "+quoteIdentifier(synccontract.HistoryIsCurrentColumn),
		args...,
	); err != nil {
		return errPostgresWriteSessionFailed
	}
	return nil
}

func postgresWritePreviewID(plan database.DatabaseWritePlan) string {
	control := plan.Control()
	digest := sha256.Sum256([]byte(control.TargetDatabase().Kind() + "\x00" + control.TargetDatabase().Value() + "\x00" + control.Target().Namespace() + "\x00" + control.Target().Relation() + "\x00" + control.NativeIdentity().Kind + "\x00" + control.NativeIdentity().Value + "\x00" + string(plan.Mode())))
	return "postgres-preview-" + hex.EncodeToString(digest[:])
}

func postgresDeliveryID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "postgres-delivery-" + hex.EncodeToString(bytes), nil
}

func postgresManagedTargetQualifiedRelation(control database.ManagedTargetControlRecord) string {
	return quoteIdentifier(control.Target().Namespace()) + "." + quoteIdentifier(control.Target().Relation())
}

func postgresWriteModeRequiresTableLock(mode synccontract.Mode) bool {
	return mode == synccontract.ModeIncrementalUpsert || mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory
}

var _ database.DatabaseWriteDriver = (*DatabaseDriver)(nil)
var _ database.WriteSession = (*postgresWriteSession)(nil)
