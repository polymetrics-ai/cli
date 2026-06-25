// Package postgres implements the native pm PostgreSQL source connector. It is a
// database connector (family: db) built on github.com/jackc/pgx/v5 (already in
// go.mod — no new dependency). It mirrors the declarative shape of the stripe
// reference connector: a thin package that self-registers via RegisterFactory in
// init() and resolves the secret from cfg.Secrets without ever logging it.
//
// Capabilities:
//   - Check:   pgxpool connect + ping using host/port/database/username/sslmode
//     and the password secret.
//   - Catalog: discover tables and columns from information_schema and map each
//     PostgreSQL data_type to a coarse Field type.
//   - Read:    snapshot SELECT over a stream, with optional cursor-incremental
//     filtering on a configurable cursor column (see StatefulReader below).
//   - Write:   not implemented; this is a read-only source. Capabilities.Write is
//     false and Write returns ErrUnsupportedOperation.
//
// CDC (change data capture) is a documented STUB: ReadCDC returns
// ErrUnsupportedOperation because full logical-replication CDC requires the
// pglogrepl dependency, a gated add. See ReadCDC for the recorded CDC plan.
//
// A mode=fixture path (cfg.Config["mode"]=="fixture") short-circuits all network
// access so the conformance harness and unit tests can run with no live DB: in
// fixture mode Check succeeds, Catalog returns canned streams, and Read emits
// canned rows.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPort    = 5432
	defaultSSLMode = "disable"
	defaultSchema  = "public"
	// defaultReadLimit bounds a snapshot SELECT so a Read never streams an entire
	// large table unbounded; override with config read_limit.
	defaultReadLimit = 10000
)

// validSSLModes is the libpq sslmode allow-list pgx accepts.
var validSSLModes = map[string]bool{
	"disable":     true,
	"allow":       true,
	"prefer":      true,
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

func init() {
	connectors.RegisterFactory("postgres", New)
}

// New returns the PostgreSQL connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm PostgreSQL source connector.
type Connector struct{}

func (Connector) Name() string { return "postgres" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "postgres",
		DisplayName:     "PostgreSQL",
		IntegrationType: "database",
		Description:     "Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and supports cursor-incremental reads on a configurable cursor column. Read-only source; CDC is a documented stub pending the gated pglogrepl dependency.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// connConfig is the validated connection configuration. The password lives in a
// dedicated field and is never logged.
type connConfig struct {
	host     string
	port     int
	database string
	username string
	password string
	sslmode  string
	schema   string
}

// dsn builds a libpq keyword/value connection string. Values are quoted to
// tolerate spaces and special characters. The password is included for pgx to
// authenticate but the returned string is never logged by this package.
func (c connConfig) dsn() string {
	kv := func(k, v string) string {
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `'`, `\'`)
		return k + "='" + v + "'"
	}
	parts := []string{
		kv("host", c.host),
		kv("port", strconv.Itoa(c.port)),
		kv("dbname", c.database),
		kv("user", c.username),
		kv("password", c.password),
		kv("sslmode", c.sslmode),
	}
	return strings.Join(parts, " ")
}

// resolveConfig validates config + secrets into a connConfig. It enforces the
// required fields, a valid sslmode, a numeric port, and that host is a bare
// hostname (no scheme/path) to bound SSRF risk from a connection-string
// injection. It never logs the password.
func resolveConfig(cfg connectors.RuntimeConfig) (connConfig, error) {
	get := func(k string) string { return strings.TrimSpace(cfg.Config[k]) }

	host := get("host")
	if host == "" {
		return connConfig{}, errors.New("postgres connector requires config host")
	}
	if err := validateHost(host); err != nil {
		return connConfig{}, err
	}

	database := get("database")
	if database == "" {
		return connConfig{}, errors.New("postgres connector requires config database")
	}
	username := get("username")
	if username == "" {
		return connConfig{}, errors.New("postgres connector requires config username")
	}

	password := ""
	if cfg.Secrets != nil {
		password = cfg.Secrets["password"]
	}
	if strings.TrimSpace(password) == "" {
		return connConfig{}, errors.New("postgres connector requires secret password")
	}

	port := defaultPort
	if raw := get("port"); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil {
			return connConfig{}, fmt.Errorf("postgres config port must be an integer: %w", err)
		}
		if p < 1 || p > 65535 {
			return connConfig{}, fmt.Errorf("postgres config port must be between 1 and 65535, got %d", p)
		}
		port = p
	}

	sslmode := strings.ToLower(get("sslmode"))
	if sslmode == "" {
		sslmode = defaultSSLMode
	}
	if !validSSLModes[sslmode] {
		return connConfig{}, fmt.Errorf("postgres config sslmode %q is not one of disable/allow/prefer/require/verify-ca/verify-full", sslmode)
	}

	schema := get("schema")
	if schema == "" {
		schema = defaultSchema
	}

	return connConfig{
		host:     host,
		port:     port,
		database: database,
		username: username,
		password: password,
		sslmode:  sslmode,
		schema:   schema,
	}, nil
}

// validateHost rejects hosts that look like a URL or carry path/query/credential
// characters. A real host is a hostname or IP (optionally IPv6 in brackets). This
// bounds SSRF / connection-string-injection risk from operator-supplied config.
func validateHost(host string) error {
	if strings.ContainsAny(host, "/\\@?#'\" \t") {
		return fmt.Errorf("postgres config host %q must be a bare hostname or IP, not a URL", host)
	}
	if strings.Contains(host, "://") {
		return fmt.Errorf("postgres config host %q must not include a scheme", host)
	}
	// Bracketed IPv6 is allowed; otherwise reject stray brackets.
	if strings.HasPrefix(host, "[") {
		if !strings.HasSuffix(host, "]") || net.ParseIP(strings.Trim(host, "[]")) == nil {
			return fmt.Errorf("postgres config host %q is not a valid bracketed IPv6 address", host)
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Check verifies connection config and, outside fixture mode, opens a pgx pool
// and pings. Fixture mode validates config shape only (no network).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	pool, err := pgxpool.New(ctx, conn.dsn())
	if err != nil {
		return fmt.Errorf("check postgres: open pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("check postgres: ping: %w", err)
	}
	return nil
}

// Catalog discovers tables and columns. Fixture mode returns canned streams.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return connectors.Catalog{}, err
	}
	if fixtureMode(cfg) {
		return connectors.Catalog{Connector: c.Name(), Streams: fixtureStreams()}, nil
	}

	pool, err := pgxpool.New(ctx, conn.dsn())
	if err != nil {
		return connectors.Catalog{}, fmt.Errorf("catalog postgres: open pool: %w", err)
	}
	defer pool.Close()

	streams, err := discoverStreams(ctx, pool, conn.schema)
	if err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

// discoverStreams reads information_schema.tables/columns for the target schema
// and builds one Stream per base table, with Fields typed from each column's
// data_type and a primary key derived from table_constraints when present.
func discoverStreams(ctx context.Context, pool *pgxpool.Pool, schema string) ([]connectors.Stream, error) {
	const colsSQL = `
SELECT c.table_name, c.column_name, c.data_type, c.ordinal_position
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = $1 AND t.table_type = 'BASE TABLE'
ORDER BY c.table_name, c.ordinal_position`

	rows, err := pool.Query(ctx, colsSQL, schema)
	if err != nil {
		return nil, fmt.Errorf("catalog postgres: query columns: %w", err)
	}
	defer rows.Close()

	byTable := map[string][]connectors.Field{}
	var order []string
	for rows.Next() {
		var table, column, dataType string
		var pos int
		if err := rows.Scan(&table, &column, &dataType, &pos); err != nil {
			return nil, fmt.Errorf("catalog postgres: scan column: %w", err)
		}
		if _, seen := byTable[table]; !seen {
			order = append(order, table)
		}
		byTable[table] = append(byTable[table], connectors.Field{Name: column, Type: pgTypeToFieldType(dataType)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog postgres: iterate columns: %w", err)
	}

	pks, err := discoverPrimaryKeys(ctx, pool, schema)
	if err != nil {
		return nil, err
	}

	streams := make([]connectors.Stream, 0, len(order))
	for _, table := range order {
		qualified := schema + "." + table
		streams = append(streams, connectors.Stream{
			Name:        qualified,
			Description: "PostgreSQL table " + qualified,
			Fields:      byTable[table],
			PrimaryKey:  pks[table],
		})
	}
	sort.Slice(streams, func(i, j int) bool { return streams[i].Name < streams[j].Name })
	return streams, nil
}

// discoverPrimaryKeys returns table_name -> ordered primary key columns for the
// schema, used to populate Stream.PrimaryKey during discovery.
func discoverPrimaryKeys(ctx context.Context, pool *pgxpool.Pool, schema string) (map[string][]string, error) {
	const pkSQL = `
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
 AND tc.table_schema = kcu.table_schema
WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = $1
ORDER BY tc.table_name, kcu.ordinal_position`

	rows, err := pool.Query(ctx, pkSQL, schema)
	if err != nil {
		return nil, fmt.Errorf("catalog postgres: query primary keys: %w", err)
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return nil, fmt.Errorf("catalog postgres: scan primary key: %w", err)
		}
		out[table] = append(out[table], column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog postgres: iterate primary keys: %w", err)
	}
	return out, nil
}

// InitialState satisfies connectors.StatefulReader. A stream starts with an empty
// incremental cursor (full snapshot); subsequent reads advance the cursor stored
// under the conventional connsdk cursor key.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read performs a snapshot SELECT over the stream's table, optionally filtered to
// rows whose cursor column is greater than the incremental lower bound carried in
// req.State (or the start cursor in config). Fixture mode emits canned rows.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	if req.Stream == "" {
		return errors.New("postgres read requires a stream (schema.table)")
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, req, emit)
	}

	cursorColumn := strings.TrimSpace(req.Config.Config["cursor_field"])
	lowerBound := connsdk.Cursor(req.State)
	limit, err := readLimit(req.Config)
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, conn.dsn())
	if err != nil {
		return fmt.Errorf("read postgres: open pool: %w", err)
	}
	defer pool.Close()

	return c.snapshot(ctx, pool, conn.schema, req.Stream, cursorColumn, lowerBound, limit, emit)
}

// snapshot builds and runs the SELECT for a stream. The table and cursor column
// are validated as plain identifiers and quoted, so they are never concatenated
// raw; the lower bound is passed as a bound parameter.
func (c Connector) snapshot(ctx context.Context, pool *pgxpool.Pool, schema, stream, cursorColumn, lowerBound string, limit int, emit func(connectors.Record) error) error {
	table, err := qualifyStream(schema, stream)
	if err != nil {
		return err
	}

	sql := "SELECT * FROM " + table
	var args []any
	if cursorColumn != "" {
		if err := validateIdentifier(cursorColumn); err != nil {
			return fmt.Errorf("read postgres: cursor_field: %w", err)
		}
		if lowerBound != "" {
			sql += " WHERE " + quoteIdentifier(cursorColumn) + " > $1"
			args = append(args, lowerBound)
		}
		sql += " ORDER BY " + quoteIdentifier(cursorColumn) + " ASC"
	}
	if limit > 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
	}

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("read postgres %s: %w", stream, err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("read postgres %s: scan row: %w", stream, err)
		}
		record := make(connectors.Record, len(fieldDescs))
		for i, fd := range fieldDescs {
			record[string(fd.Name)] = values[i]
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read postgres %s: iterate: %w", stream, err)
	}
	return nil
}

// readFixture emits canned rows for a fixture stream without any network access.
// It applies the incremental cursor lower bound numerically so cursor behaviour
// is exercised credential-free.
func (c Connector) readFixture(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	rows, ok := fixtureRows(req.Stream)
	if !ok {
		return fmt.Errorf("postgres fixture stream %q not found", req.Stream)
	}
	var lower int64 = -1
	if cur := connsdk.Cursor(req.State); cur != "" {
		if n, err := strconv.ParseInt(cur, 10, 64); err == nil {
			lower = n
		}
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if lower >= 0 && row.cursor <= lower {
			continue
		}
		if err := emit(connsdkCopy(row.record)); err != nil {
			return err
		}
	}
	return nil
}

// ReadCDC is a DOCUMENTED STUB. Full PostgreSQL change data capture uses logical
// replication, which requires the pglogrepl dependency — a gated add not present
// in go.mod. Until that dependency is approved, CDC is unsupported.
//
// Recorded CDC plan (for the future pglogrepl implementation):
//   - Server prerequisite: wal_level=logical.
//   - Create a logical replication slot (pgoutput plugin), e.g.
//     SELECT pg_create_logical_replication_slot('pm_<connector>', 'pgoutput').
//   - Create a PUBLICATION for the target tables:
//     CREATE PUBLICATION pm_pub FOR TABLE schema.table[, ...].
//   - Start replication from the stored LSN (confirmed_flush_lsn) via the
//     START_REPLICATION protocol; decode pgoutput Insert/Update/Delete messages
//     into connectors.CDCEvent{Operation, Record}.
//   - Persist the last committed LSN in CDCEvent.State (e.g. {"lsn": "0/1A2B3C"})
//     so the next run resumes after the last flushed change.
func (c Connector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	return fmt.Errorf("postgres CDC requires the gated pglogrepl dependency (wal_level=logical, replication slot, publication, lsn state): %w", connectors.ErrUnsupportedOperation)
}

// Write is unsupported: this is a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// qualifyStream turns a stream name (either "table" or "schema.table") into a
// quoted, validated qualified identifier safe to embed in SQL.
func qualifyStream(defaultSchemaName, stream string) (string, error) {
	stream = strings.TrimSpace(stream)
	schemaName := defaultSchemaName
	table := stream
	if idx := strings.IndexByte(stream, '.'); idx >= 0 {
		schemaName = stream[:idx]
		table = stream[idx+1:]
	}
	if err := validateIdentifier(schemaName); err != nil {
		return "", fmt.Errorf("read postgres: schema: %w", err)
	}
	if err := validateIdentifier(table); err != nil {
		return "", fmt.Errorf("read postgres: table: %w", err)
	}
	return quoteIdentifier(schemaName) + "." + quoteIdentifier(table), nil
}

// validateIdentifier rejects identifiers that are not a plain
// [A-Za-z_][A-Za-z0-9_$]* token, preventing SQL injection through table/column
// names that cannot be passed as bound parameters.
func validateIdentifier(id string) error {
	if id == "" {
		return errors.New("identifier must not be empty")
	}
	if len(id) > 63 {
		return fmt.Errorf("identifier %q exceeds 63 characters", id)
	}
	for i, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case (r >= '0' && r <= '9' || r == '$') && i > 0:
		default:
			return fmt.Errorf("identifier %q contains an illegal character", id)
		}
	}
	return nil
}

// quoteIdentifier double-quotes an identifier, escaping embedded quotes. Callers
// must validate with validateIdentifier first; this is defence in depth.
func quoteIdentifier(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

func readLimit(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["read_limit"]))
	if raw == "" {
		return defaultReadLimit, nil
	}
	if raw == "0" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("postgres config read_limit must be an integer, 0, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("postgres config read_limit must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// connsdkCopy returns a shallow copy of a record so emitted fixture rows are not
// aliased between calls.
func connsdkCopy(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
