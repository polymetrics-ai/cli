package postgres

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/sqltls"
)

// The captain's ruling is that MySQL and PostgreSQL must not drift into two
// spellings of the same transport-security choice. Existing libpq values keep
// their exact meaning; the canonical vocabulary is additionally accepted.
func TestSSLModeAcceptsTheSharedCanonicalVocabulary(t *testing.T) {
	for raw, want := range map[string]string{
		"disable":         "disable",
		"allow":           "allow",
		"prefer":          "prefer",
		"require":         "require",
		"verify-ca":       "verify-ca",
		"verify-full":     "verify-full",
		"disabled":        "disable",
		"preferred":       "prefer",
		"required":        "require",
		"verify-identity": "verify-full",
	} {
		got, err := libpqSSLMode(raw)
		if err != nil || got != want {
			t.Fatalf("libpqSSLMode(%q) = %q, %v, want %q", raw, got, err, want)
		}
	}
	if _, err := libpqSSLMode("not-a-mode"); err == nil {
		t.Fatal("libpqSSLMode() accepted an unknown sslmode")
	}
}

// The definition is the validation boundary used before a native connector is
// invoked. It must accept every spelling resolveConfig accepts, otherwise a
// real shared TLS option could be rejected before the connector can enforce
// it.
func TestDefinitionAcceptsSharedTransportSecurityVocabulary(t *testing.T) {
	connector := New()
	for _, mode := range []string{
		"disabled", "preferred", "required", "verify-ca", "verify-identity",
		"disable", "allow", "prefer", "require", "verify-full",
		"verify_ca", "verify_identity",
	} {
		if err := connectors.ValidateConfiguration(connector, map[string]string{"sslmode": mode}); err != nil {
			t.Fatalf("definition rejected accepted sslmode %q: %v", mode, err)
		}
	}
}

func TestResolveConfigAuthenticationRequirements(t *testing.T) {
	base := func() connectors.RuntimeConfig {
		return connectors.RuntimeConfig{
			Config: map[string]string{
				"host":     "postgres.example",
				"database": "analytics",
				"username": "reader",
			},
			Secrets: map[string]string{"password": t.Name()},
		}
	}

	tests := []struct {
		name    string
		cfg     connectors.RuntimeConfig
		wantErr error
	}{
		{
			name: "fixture permits no password",
			cfg: connectors.RuntimeConfig{Config: map[string]string{
				"mode":     "fixture",
				"host":     "postgres.example",
				"database": "analytics",
				"username": "reader",
			}},
		},
		{
			name: "live password mode requires password",
			cfg: connectors.RuntimeConfig{Config: map[string]string{
				"host":     "postgres.example",
				"database": "analytics",
				"username": "reader",
			}},
			wantErr: errPostgresPasswordAuthenticationRequired,
		},
		{
			name: "peer mode is rejected",
			cfg: func() connectors.RuntimeConfig {
				cfg := base()
				cfg.Config["auth_mode"] = "peer"
				return cfg
			}(),
			wantErr: errPostgresAuthenticationModeUnsupported,
		},
		{
			name: "socket host is rejected",
			cfg: func() connectors.RuntimeConfig {
				cfg := base()
				cfg.Config["host"] = "/var/run/postgresql"
				return cfg
			}(),
			wantErr: errPostgresAuthenticationModeUnsupported,
		},
		{
			name: "client certificate is rejected",
			cfg: func() connectors.RuntimeConfig {
				cfg := base()
				cfg.Config["sslcert"] = "/tmp/client.pem"
				return cfg
			}(),
			wantErr: errPostgresAuthenticationModeUnsupported,
		},
		{
			name: "unknown mode is rejected",
			cfg: func() connectors.RuntimeConfig {
				cfg := base()
				cfg.Config["auth_mode"] = "gssapi"
				return cfg
			}(),
			wantErr: errPostgresAuthenticationModeUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveConfig(tt.cfg)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("resolveConfig() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("resolveConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresPoolConfigUsesSharedTransportSecurityOptions(t *testing.T) {
	conn, err := resolveConfig(connectors.RuntimeConfig{
		Config: map[string]string{
			"host":          "postgres.example",
			"database":      "analytics",
			"username":      "reader",
			"sslmode":       "verify-identity",
			"sslservername": "database.example",
			"sslrootcert":   "/tmp/pm-postgres-root.pem",
		},
		// The resolver only needs a non-empty value. Derive it at runtime so
		// this test fixture never stores a credential-shaped string.
		Secrets: map[string]string{"password": t.Name()},
	})
	if err != nil {
		t.Fatalf("resolveConfig() error = %v", err)
	}
	if !strings.Contains(conn.dsn(), "sslrootcert='/tmp/pm-postgres-root.pem'") {
		t.Fatal("dsn() did not include the shared sslrootcert option")
	}

	// Parsing an explicit root certificate makes pgx read that file. The
	// option itself is asserted above; clear the synthetic path before asking
	// pgx to construct a TLS config for this no-network unit test.
	conn.tls.RootCAPath = ""
	poolConfig, err := conn.poolConfig()
	if err != nil {
		t.Fatalf("poolConfig() error = %v", err)
	}
	if poolConfig.ConnConfig.TLSConfig == nil {
		t.Fatal("poolConfig() did not configure TLS for verify-identity")
	}
	if got := poolConfig.ConnConfig.TLSConfig.ServerName; got != "database.example" {
		t.Fatalf("TLS server name = %q, want configured shared sslservername", got)
	}
	if got := len(poolConfig.ConnConfig.Fallbacks); got != 0 {
		t.Fatalf("strict TLS pool fallbacks = %d, want no plaintext fallback", got)
	}
	if got := conn.tls.Mode; got != sqltls.ModeVerifyIdentity {
		t.Fatalf("resolved TLS mode = %q, want verify-identity", got)
	}

	replicationConfig, err := conn.replicationConfig()
	if err != nil {
		t.Fatalf("replicationConfig() error = %v", err)
	}
	if replicationConfig.TLSConfig == nil || replicationConfig.TLSConfig.ServerName != "database.example" {
		t.Fatalf("replication TLS server name = %#v, want configured shared sslservername", replicationConfig.TLSConfig)
	}
	if got := len(replicationConfig.Fallbacks); got != 0 {
		t.Fatalf("strict TLS replication fallbacks = %d, want no plaintext fallback", got)
	}

	dataConfig, err := conn.dataConfig()
	if err != nil {
		t.Fatalf("dataConfig() error = %v", err)
	}
	if dataConfig.TLSConfig == nil || dataConfig.TLSConfig.ServerName != "database.example" {
		t.Fatalf("slot-inspection TLS server name = %#v, want configured shared sslservername", dataConfig.TLSConfig)
	}
}

func TestPostgresConnectionConfigsRejectAmbientClientCertificate(t *testing.T) {
	certificate, key := postgresClientCertificateFiles(t)
	t.Setenv("PGSSLCERT", certificate)
	t.Setenv("PGSSLKEY", key)
	t.Setenv("PGSERVICE", "")
	t.Setenv("PGSERVICEFILE", "")

	conn, err := resolveConfig(connectors.RuntimeConfig{
		Config: map[string]string{
			"host":     "postgres.example",
			"database": "analytics",
			"username": "reader",
			"sslmode":  "required",
		},
		Secrets: map[string]string{"password": t.Name()},
	})
	if err != nil {
		t.Fatalf("resolveConfig() error = %v", err)
	}

	tests := []struct {
		name  string
		parse func() error
	}{
		{name: "pool", parse: func() error {
			_, err := conn.poolConfig()
			return err
		}},
		{name: "data", parse: func() error {
			_, err := conn.dataConfig()
			return err
		}},
		{name: "replication", parse: func() error {
			_, err := conn.replicationConfig()
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.parse(); !errors.Is(err, errPostgresAmbientClientCertificateUnsupported) {
				t.Fatalf("connection configuration error = %v, want %v", err, errPostgresAmbientClientCertificateUnsupported)
			}
		})
	}
}

func TestPostgresParsedConfigRejectsAndClearsAmbientClientCertificate(t *testing.T) {
	certificate, key := postgresClientCertificateFiles(t)
	t.Setenv("PGSSLCERT", certificate)
	t.Setenv("PGSSLKEY", key)
	t.Setenv("PGSERVICE", "")
	t.Setenv("PGSERVICEFILE", "")

	conn := connConfig{
		host:     "postgres.example",
		port:     defaultPort,
		database: "analytics",
		username: "reader",
		password: t.Name(),
		sslmode:  "require",
	}
	config, err := pgconn.ParseConfig(conn.dsn())
	if err != nil {
		t.Fatalf("pgconn.ParseConfig() error = %v", err)
	}
	if config.TLSConfig == nil || len(config.TLSConfig.Certificates) == 0 {
		t.Fatal("pgconn.ParseConfig() did not configure the ambient client certificate")
	}
	if err := conn.applyTLSServerName(config); !errors.Is(err, errPostgresAmbientClientCertificateUnsupported) {
		t.Fatalf("applyTLSServerName() error = %v, want %v", err, errPostgresAmbientClientCertificateUnsupported)
	}
	if len(config.TLSConfig.Certificates) != 0 || config.TLSConfig.GetClientCertificate != nil {
		t.Fatal("applyTLSServerName() retained an ambient client certificate")
	}
}

func TestPostgresDefaultClientCertificateDetection(t *testing.T) {
	home := t.TempDir()
	clientDir := filepath.Join(home, ".postgresql")
	if err := os.Mkdir(clientDir, 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(clientDir, "postgresql.crt"), []byte("certificate"), 0o600); err != nil {
		t.Fatalf("WriteFile(certificate) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(clientDir, "postgresql.key"), []byte("key"), 0o600); err != nil {
		t.Fatalf("WriteFile(key) error = %v", err)
	}
	if !postgresDefaultClientCertificateConfigured(home) {
		t.Fatal("postgresDefaultClientCertificateConfigured() = false, want true")
	}
}

func postgresClientCertificateFiles(t *testing.T) (string, string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() error = %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certificate, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() error = %v", err)
	}
	privateKeyPEM, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("x509.MarshalPKCS8PrivateKey() error = %v", err)
	}

	dir := t.TempDir()
	certificatePath := filepath.Join(dir, "client.crt")
	keyPath := filepath.Join(dir, "client.key")
	if err := os.WriteFile(certificatePath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}), 0o600); err != nil {
		t.Fatalf("WriteFile(certificate) error = %v", err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyPEM}), 0o600); err != nil {
		t.Fatalf("WriteFile(key) error = %v", err)
	}
	return certificatePath, keyPath
}
