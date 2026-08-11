package postgres

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/sqltls"
)

const (
	defaultPort    = 5432
	defaultSSLMode = "disable"
	defaultSchema  = "public"
	// defaultReadLimit bounds a snapshot SELECT so a Read never streams an
	// entire large table unbounded; override with config read_limit.
	defaultReadLimit = 10000
)

// validSSLModes is the libpq sslmode allow-list pgx accepts verbatim. A value
// outside this set is still accepted when sqltls recognises it, so the
// canonical vocabulary (disabled/preferred/required/verify-ca/verify-identity)
// means the same thing here as on the MySQL connector.
var validSSLModes = map[string]bool{
	"disable":     true,
	"allow":       true,
	"prefer":      true,
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

var (
	errPostgresPasswordAuthenticationRequired      = errors.New("postgres password authentication requires secret password")
	errPostgresAuthenticationModeUnsupported       = errors.New("postgres authentication mode is unsupported")
	errPostgresAmbientClientCertificateUnsupported = fmt.Errorf(
		"%w: ambient client-certificate", errPostgresAuthenticationModeUnsupported,
	)
)

// libpqSSLMode normalises one accepted spelling to what pgx expects. A libpq
// name passes through unchanged so existing configuration keeps its exact
// behaviour, including "allow", which sqltls folds into "preferred" and which
// must not be rewritten here.
func libpqSSLMode(raw string) (string, error) {
	if validSSLModes[raw] {
		return raw, nil
	}
	mode, err := sqltls.ParseMode(raw)
	if err != nil {
		return "", fmt.Errorf("postgres config sslmode is not one of disable/allow/prefer/require/verify-ca/verify-full or %s", sqltls.AcceptedModes)
	}
	return mode.LibpqSSLMode(), nil
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
	tls      sqltls.Options
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
	if c.tls.RootCAPath != "" {
		parts = append(parts, kv("sslrootcert", c.tls.RootCAPath))
	}
	return strings.Join(parts, " ")
}

func (c connConfig) connectionString() (string, error) {
	if c.tls.Encrypted() && postgresAmbientClientCertificateConfigured() {
		return "", errPostgresAmbientClientCertificateUnsupported
	}
	return c.dsn(), nil
}

func postgresAmbientClientCertificateConfigured() bool {
	if os.Getenv("PGSSLCERT") != "" || os.Getenv("PGSSLKEY") != "" {
		return true
	}
	currentUser, err := user.Current()
	return err == nil && postgresDefaultClientCertificateConfigured(currentUser.HomeDir)
}

func postgresDefaultClientCertificateConfigured(home string) bool {
	if strings.TrimSpace(home) == "" {
		return false
	}
	clientCertificate := filepath.Join(home, ".postgresql", "postgresql.crt")
	clientKey := filepath.Join(home, ".postgresql", "postgresql.key")
	if _, err := os.Stat(clientCertificate); err != nil {
		return false
	}
	if _, err := os.Stat(clientKey); err != nil {
		return false
	}
	return true
}

// applyTLSServerName applies the shared verify-identity override after pgx
// has parsed the libpq-compatible connection fields. It is deliberately used
// by normal SQL, replication, and slot-inspection connections so CDC cannot
// silently use a different transport policy from the rest of the connector.
func (c connConfig) applyTLSServerName(config *pgconn.Config) error {
	clientCertificate := clearClientCertificates(config.TLSConfig)
	for _, fallback := range config.Fallbacks {
		if fallback != nil && clearClientCertificates(fallback.TLSConfig) {
			clientCertificate = true
		}
	}
	if clientCertificate {
		return errPostgresAmbientClientCertificateUnsupported
	}
	if c.tls.Mode != sqltls.ModeVerifyIdentity || c.tls.ServerName == "" {
		return nil
	}
	if config.TLSConfig == nil {
		return errors.New("postgres verify-identity did not configure TLS")
	}
	config.TLSConfig.ServerName = c.tls.ServerName
	for _, fallback := range config.Fallbacks {
		if fallback != nil && fallback.TLSConfig != nil {
			fallback.TLSConfig.ServerName = c.tls.ServerName
		}
	}
	return nil
}

func clearClientCertificates(config *tls.Config) bool {
	if config == nil {
		return false
	}
	configured := len(config.Certificates) > 0 || config.GetClientCertificate != nil
	config.Certificates = nil
	config.GetClientCertificate = nil
	return configured
}

// replicationConfig returns a parsed protocol connection configuration without
// ever returning a parser error that could embed a connection string.
func (c connConfig) replicationConfig() (*pgconn.Config, error) {
	connectionString, err := c.connectionString()
	if err != nil {
		return nil, err
	}
	config, err := pgconn.ParseConfig(connectionString)
	if err != nil {
		return nil, errors.New("postgres replication configuration is invalid")
	}
	if err := c.applyTLSServerName(config); err != nil {
		return nil, err
	}
	return config, nil
}

// dataConfig returns the ordinary pgx configuration used to inspect a slot.
// It shares the same TLS and secret-safe parse path as replicationConfig.
func (c connConfig) dataConfig() (*pgx.ConnConfig, error) {
	connectionString, err := c.connectionString()
	if err != nil {
		return nil, err
	}
	config, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, errors.New("postgres connection configuration is invalid")
	}
	if err := c.applyTLSServerName(&config.Config); err != nil {
		return nil, err
	}
	return config, nil
}

// poolConfig converts the validated connector config into pgx's pool
// configuration. sslrootcert is rendered through libpq's supported keyword;
// sslservername is deliberately applied after parsing because pgx uses the
// TLS config's ServerName rather than accepting a separate libpq option for
// it. Strict modes have no plaintext fallback, and preferred/allow retain the
// driver's documented fallback behavior.
func (c connConfig) poolConfig() (*pgxpool.Config, error) {
	connectionString, err := c.connectionString()
	if err != nil {
		return nil, err
	}
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		// pgx can include caller-provided config in parse errors. Never return
		// it to a command surface because it may include authentication data.
		return nil, errors.New("postgres configuration could not create a connection pool")
	}
	if err := c.applyTLSServerName(&poolConfig.ConnConfig.Config); err != nil {
		return nil, err
	}
	return poolConfig, nil
}

// openPool makes every PostgreSQL operation use the same validated transport
// security configuration. In particular, it prevents Catalog and Read from
// silently ignoring sslservername while Check honours it.
func (c connConfig) openPool(ctx context.Context) (*pgxpool.Pool, error) {
	poolConfig, err := c.poolConfig()
	if err != nil {
		return nil, err
	}
	return openPostgresPool(ctx, poolConfig)
}

func (c connConfig) typedCatalogPoolConfig(resources typedCatalogResources) (*pgxpool.Config, error) {
	poolConfig, err := c.poolConfig()
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = resources.poolSize
	poolConfig.ConnConfig.ConnectTimeout = resources.policy.ConnectTimeout
	return poolConfig, nil
}

func (c connConfig) openTypedCatalogPool(ctx context.Context, resources typedCatalogResources) (*pgxpool.Pool, error) {
	poolConfig, err := c.typedCatalogPoolConfig(resources)
	if err != nil {
		return nil, err
	}
	return openPostgresPool(ctx, poolConfig)
}

func openPostgresPool(ctx context.Context, poolConfig *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		// Keep endpoint and credentials out of user-visible errors.
		return nil, errors.New("connect postgres failed")
	}
	return pool, nil
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
	if strings.HasPrefix(host, "/") {
		return connConfig{}, fmt.Errorf("%w: peer/socket", errPostgresAuthenticationModeUnsupported)
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
	if err := validateAuthentication(cfg); err != nil {
		return connConfig{}, err
	}

	password := ""
	if cfg.Secrets != nil {
		password = cfg.Secrets["password"]
	}
	if strings.TrimSpace(password) == "" && !fixtureMode(cfg) {
		return connConfig{}, errPostgresPasswordAuthenticationRequired
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
	sslmode, err := libpqSSLMode(sslmode)
	if err != nil {
		return connConfig{}, err
	}
	transport, err := sqltls.Resolve(get, sqltls.ModeDisabled)
	if err != nil {
		return connConfig{}, fmt.Errorf("postgres config: %w", err)
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
		tls:      transport,
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

func validateAuthentication(cfg connectors.RuntimeConfig) error {
	switch strings.ToLower(strings.TrimSpace(cfg.Config["auth_mode"])) {
	case "", "password":
	case "peer", "socket", "peer/socket":
		return fmt.Errorf("%w: peer/socket", errPostgresAuthenticationModeUnsupported)
	case "client-certificate", "client_certificate", "certificate", "cert":
		return fmt.Errorf("%w: client-certificate", errPostgresAuthenticationModeUnsupported)
	default:
		return errPostgresAuthenticationModeUnsupported
	}
	for _, key := range []string{"sslcert", "sslkey", "sslpassword", "client_certificate", "client_key"} {
		if strings.TrimSpace(cfg.Config[key]) != "" || strings.TrimSpace(cfg.Secrets[key]) != "" {
			return fmt.Errorf("%w: client-certificate", errPostgresAuthenticationModeUnsupported)
		}
	}
	return nil
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
	pool, err := conn.openPool(ctx)
	if err != nil {
		return fmt.Errorf("check postgres: open pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("check postgres: ping: %w", err)
	}
	return nil
}
