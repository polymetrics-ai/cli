package postgres

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/database"
)

var (
	errPostgresDatabaseDriverConnectionRequired = errors.New("postgres managed target driver requires a pinned connection")
	errPostgresDurabilityUnsafe                 = errors.New("postgres managed target durability settings are unsafe")
	errPostgresTargetLayoutUnavailable          = errors.New("postgres managed target layout contract is unavailable")
)

// DatabaseDriver is PostgreSQL's compile-time reference seam for the shared
// typed database foundation. A constructed driver pins one PostgreSQL
// connection for ownership observation and PostgreSQL advisory locks. It is
// deliberately not registered and remains unable to begin a mapped write
// session until #3973 provides the shared MappingContractV1.
type DatabaseDriver struct {
	conn   *pgx.Conn
	connMu *sync.Mutex
}

// NewDatabaseDriver binds the native managed-target adapter to one already
// validated PostgreSQL connection. It neither parses a DSN nor retains a
// password, so the source connection boundary remains responsible for auth and
// transport policy.
func NewDatabaseDriver(conn *pgx.Conn) (*DatabaseDriver, error) {
	if conn == nil {
		return nil, errPostgresDatabaseDriverConnectionRequired
	}
	return &DatabaseDriver{conn: conn, connMu: &sync.Mutex{}}, nil
}

// DatabaseDriverDescriptor returns the closed PostgreSQL wire-driver identity
// expected by defs/postgres/database.json.
func (DatabaseDriver) DatabaseDriverDescriptor() database.DriverDescriptor {
	return database.DriverDescriptor{
		ID:         "postgres",
		Protocol:   "postgres-wire",
		APIVersion: 1,
	}
}

// PreflightDurability proves the PostgreSQL settings that make a successful
// future write-session COMMIT durable to the server's local WAL. A driver never
// relaxes either setting; session configuration is established by the mapped
// future write transaction and must still be rechecked there.
func (d *DatabaseDriver) PreflightDurability(ctx context.Context) error {
	if ctx == nil {
		return errPostgresDurabilityUnsafe
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	d.connMu.Lock()
	defer d.connMu.Unlock()
	var fsync, synchronousCommit string
	if err := d.conn.QueryRow(ctx, "SELECT current_setting('fsync'), current_setting('synchronous_commit')").Scan(&fsync, &synchronousCommit); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		return errPostgresDurabilityUnsafe
	}
	if strings.TrimSpace(strings.ToLower(fsync)) != "on" || strings.TrimSpace(strings.ToLower(synchronousCommit)) != "on" {
		return errPostgresDurabilityUnsafe
	}
	return nil
}

var _ database.Driver = DatabaseDriver{}
