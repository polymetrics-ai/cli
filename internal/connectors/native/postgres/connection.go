package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
)

const (
	defaultPort    = 5432
	defaultSSLMode = "disable"
	defaultSchema  = "public"
	// defaultReadLimit bounds a snapshot SELECT so a Read never streams an
	// entire large table unbounded; override with config read_limit.
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

// connConfig is the validated connection configuration. The password lives
// in a dedicated field and is never logged.
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
// tolerate spaces and special characters. The password is included for pgx
// to authenticate but the returned string is never logged by this package.
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

// resolveConfig validates config + secrets into a connConfig. It enforces
// the required fields, a valid sslmode, a numeric port, and that host is a
// bare hostname (no scheme/path) to bound SSRF risk from a
// connection-string injection. It never logs the password. Ported verbatim
// (rule-for-rule) from the legacy internal/connectors/postgres/postgres.go
// resolveConfig (postgres.go:119); error wording is this package's own (see
// ledger "parity choices" #1 — classification parity, not string parity).
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

// validateHost rejects hosts that look like a URL or carry path/query/
// credential characters. A real host is a hostname or IP (optionally IPv6
// in brackets). This bounds SSRF / connection-string-injection risk from
// operator-supplied config.
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

// validateIdentifier rejects identifiers that are not a plain
// [A-Za-z_][A-Za-z0-9_$]* token, preventing SQL injection through
// table/column names that cannot be passed as bound parameters.
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

// quoteIdentifier double-quotes an identifier, escaping embedded quotes.
// Callers must validate with validateIdentifier first; this is defence in
// depth.
func quoteIdentifier(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

// readLimit parses config read_limit (default defaultReadLimit; 0/"all"/
// "unlimited" disables the bound).
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

// Check verifies connection config and, outside fixture mode, opens a pgx
// pool and pings. Fixture mode validates config shape only (no network).
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
