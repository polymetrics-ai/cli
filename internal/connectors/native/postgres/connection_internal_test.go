package postgres

import (
	"strings"
	"testing"

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
}
