package mysql

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/client"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/sqltls"
)

const (
	defaultPort      = 3306
	defaultReadLimit = 10_000
	// defaultTLSMode encrypts opportunistically. Every stricter mode is
	// available and honestly enforced; see sqltls.
	defaultTLSMode = sqltls.ModePreferred
)

type connConfig struct {
	host     string
	port     int
	database string
	username string
	password string
	tls      sqltls.Options
}

func resolveConfig(cfg connectors.RuntimeConfig) (connConfig, error) {
	value := func(key string) string { return strings.TrimSpace(cfg.Config[key]) }
	host := value("host")
	if err := validateHost(host); err != nil {
		return connConfig{}, err
	}
	database := value("database")
	if err := validateIdentifier(database); err != nil {
		return connConfig{}, fmt.Errorf("mysql config database: %w", err)
	}
	username := value("username")
	if username == "" || strings.ContainsAny(username, "\r\n\x00") {
		return connConfig{}, errors.New("mysql connector requires a safe config username")
	}
	port, err := portConfig(value("port"))
	if err != nil {
		return connConfig{}, err
	}
	password := ""
	if cfg.Secrets != nil {
		password = cfg.Secrets["password"]
	}
	// defaultTLSMode is "preferred": a local container and a managed remote
	// server are both ordinary here, so the connector encrypts whenever the
	// server offers it and never refuses a server that does not. A user who
	// needs a guarantee states "required" or a verifying mode.
	transport, err := sqltls.Resolve(value, defaultTLSMode)
	if err != nil {
		return connConfig{}, fmt.Errorf("mysql config: %w", err)
	}
	return connConfig{
		host:     host,
		port:     port,
		database: database,
		username: username,
		password: password,
		tls:      transport,
	}, nil
}

func validateHost(host string) error {
	host = strings.TrimSpace(host)
	if host == "" {
		return errors.New("mysql connector requires config host")
	}
	if strings.ContainsAny(host, "/\\?#@\r\n\x00") || strings.Contains(host, "://") {
		return errors.New("mysql config host must be a bare hostname or IP")
	}
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}
	if strings.Contains(host, ":") || len(host) > 253 {
		return errors.New("mysql config host must be a bare hostname or IP")
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return errors.New("mysql config host must be a bare hostname or IP")
		}
		for _, r := range label {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' {
				return errors.New("mysql config host must be a bare hostname or IP")
			}
		}
	}
	return nil
}

func portConfig(raw string) (int, error) {
	if raw == "" {
		return defaultPort, nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("mysql config port must be an integer from 1 to 65535")
	}
	return port, nil
}

func readLimit(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["read_limit"])
	if raw == "" {
		return defaultReadLimit, nil
	}
	if raw == "0" || strings.EqualFold(raw, "all") || strings.EqualFold(raw, "unlimited") {
		return 0, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 {
		return 0, errors.New("mysql config read_limit must be a positive integer, 0, all, or unlimited")
	}
	return limit, nil
}

func replicationServerID(cfg connectors.RuntimeConfig) (uint32, error) {
	raw := strings.TrimSpace(cfg.Config["replication_server_id"])
	if raw == "" {
		return 0, errors.New("mysql CDC requires config replication_server_id")
	}
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || value == 0 {
		return 0, errors.New("mysql config replication_server_id must be a positive uint32")
	}
	return uint32(value), nil
}

func (c connConfig) address() string {
	return net.JoinHostPort(c.host, strconv.Itoa(c.port))
}

// serverRefusedTLS is the driver's wording when the server advertises no TLS
// capability. It is the only failure that may trigger the "preferred"
// plaintext retry. If a driver upgrade changes this wording the match stops
// firing and "preferred" fails closed instead of downgrading, which is the
// safe direction for the mismatch.
const serverRefusedTLS = "does not support TLS"

// open dials the server under the configured transport-security mode.
//
// A stricter-than-preferred mode is never downgraded: the tls.Config is
// non-nil, and the driver refuses a server that advertises no TLS. Only
// "preferred" retries in plaintext, and only for that one refusal.
func (c connConfig) open(ctx context.Context) (*client.Conn, error) {
	tlsConfig, err := c.tls.TLSConfig(c.host)
	if err != nil {
		return nil, err
	}
	conn, err := c.dial(ctx, tlsConfig)
	if err == nil {
		return conn, nil
	}
	if tlsConfig == nil || !c.tls.MayFallBackToPlaintext() || !strings.Contains(err.Error(), serverRefusedTLS) {
		return nil, dialError(ctx, err)
	}
	if conn, err = c.dial(ctx, nil); err != nil {
		return nil, dialError(ctx, err)
	}
	return conn, nil
}

// dialError keeps a cancellation or deadline distinguishable with errors.Is
// while never returning driver text, which can include the endpoint or
// authentication details. The sentinel itself is returned rather than the
// wrapped client error, so an operator's Ctrl-C during a sync propagates as a
// cancellation instead of reading as a connection failure, and no
// configuration value reaches a caller-visible error or log.
func dialError(ctx context.Context, err error) error {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, cause) || errors.Is(ctx.Err(), cause) {
			return cause
		}
	}
	return errors.New("connect mysql failed")
}

func (c connConfig) dial(ctx context.Context, tlsConfig *tls.Config) (*client.Conn, error) {
	var options []client.Option
	if tlsConfig != nil {
		options = append(options, func(conn *client.Conn) error {
			conn.SetTLSConfig(tlsConfig)
			return nil
		})
	}
	return client.ConnectWithContext(ctx, c.address(), c.username, c.password, c.database, 10*time.Second, options...)
}

// replicationTLS resolves the tls.Config the binary-log syncer must use, so
// change capture runs under the same transport-security choice as every other
// statement rather than quietly reverting to plaintext. Under "preferred" it
// probes once, because the syncer has no opportunistic mode of its own.
func (c connConfig) replicationTLS(ctx context.Context) (*tls.Config, error) {
	tlsConfig, err := c.tls.TLSConfig(c.host)
	if err != nil || tlsConfig == nil {
		return nil, err
	}
	if !c.tls.MayFallBackToPlaintext() {
		return tlsConfig, nil
	}
	probe, err := c.dial(ctx, tlsConfig)
	if err != nil {
		if strings.Contains(err.Error(), serverRefusedTLS) {
			return nil, nil
		}
		return nil, dialError(ctx, err)
	}
	_ = probe.Close()
	return tlsConfig, nil
}

// Check opens and pings the configured server without logging configuration or
// authentication material.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return err
	}
	db, err := conn.open(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}
	return nil
}

func validateIdentifier(value string) error {
	if value == "" {
		return errors.New("identifier is required")
	}
	for i, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9') {
			continue
		}
		// Do not echo a caller-supplied value: a mistaken configuration value can
		// itself contain sensitive connection material.
		return errors.New("identifier is unsafe")
	}
	return nil
}

func quoteIdentifier(value string) string {
	return "`" + value + "`"
}

func qualifyStream(defaultDatabase, stream string) (database, table string, err error) {
	parts := strings.Split(strings.TrimSpace(stream), ".")
	switch len(parts) {
	case 1:
		database, table = defaultDatabase, parts[0]
	case 2:
		database, table = parts[0], parts[1]
	default:
		return "", "", errors.New("mysql stream must be table or database.table")
	}
	if err := validateIdentifier(database); err != nil {
		return "", "", fmt.Errorf("mysql stream database: %w", err)
	}
	if err := validateIdentifier(table); err != nil {
		return "", "", fmt.Errorf("mysql stream table: %w", err)
	}
	return database, table, nil
}
