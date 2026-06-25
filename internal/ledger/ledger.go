package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RunRecord struct {
	ID             string    `json:"id"`
	Mode           string    `json:"mode"`
	Operation      string    `json:"operation"`
	Status         string    `json:"status"`
	RecordsRead    int       `json:"records_read"`
	RecordsWritten int       `json:"records_written"`
	Duration       int64     `json:"duration_ns"`
	CreatedAt      time.Time `json:"created_at"`
}

type JSONLedger struct {
	Path string
}

func (l JSONLedger) Append(ctx context.Context, record RunRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o700); err != nil {
		return fmt.Errorf("create ledger directory: %w", err)
	}
	file, err := os.OpenFile(l.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open json ledger: %w", err)
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	if err := enc.Encode(record); err != nil {
		return fmt.Errorf("write json ledger: %w", err)
	}
	return nil
}

type PostgresLedger struct {
	Pool *pgxpool.Pool
}

func OpenPostgres(ctx context.Context, postgresURL string) (*PostgresLedger, error) {
	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresLedger{Pool: pool}, nil
}

func (l *PostgresLedger) Close() {
	if l != nil && l.Pool != nil {
		l.Pool.Close()
	}
}

func (l *PostgresLedger) Migrate(ctx context.Context) error {
	_, err := l.Pool.Exec(ctx, `
create table if not exists polymetrics_run_ledger (
  id text primary key,
  mode text not null,
  operation text not null,
  status text not null,
  records_read integer not null default 0,
  records_written integer not null default 0,
  duration_ns bigint not null default 0,
  created_at timestamptz not null
)`)
	if err != nil {
		return fmt.Errorf("migrate postgres ledger: %w", err)
	}
	return nil
}

func (l *PostgresLedger) Append(ctx context.Context, record RunRecord) error {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	_, err := l.Pool.Exec(ctx, `
insert into polymetrics_run_ledger
  (id, mode, operation, status, records_read, records_written, duration_ns, created_at)
values
  ($1, $2, $3, $4, $5, $6, $7, $8)
on conflict (id) do update set
  status = excluded.status,
  records_read = excluded.records_read,
  records_written = excluded.records_written,
  duration_ns = excluded.duration_ns`,
		record.ID,
		record.Mode,
		record.Operation,
		record.Status,
		record.RecordsRead,
		record.RecordsWritten,
		record.Duration,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("append postgres ledger: %w", err)
	}
	return nil
}
