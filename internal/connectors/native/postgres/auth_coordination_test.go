package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors"
)

func TestPostgresAuthenticationClassificationUsesProtocolCodeOnly(t *testing.T) {
	invalid := markPostgresAuthenticationFailure(&pgconn.PgError{Code: "28P01", Message: "password authentication failed"})
	if !connectors.IsVerifiedAuthenticationFailure(invalid) {
		t.Fatalf("28P01 classification = %T %v, want verified authentication failure", invalid, invalid)
	}
	permission := &pgconn.PgError{Code: "42501", Message: "permission denied"}
	if got := markPostgresAuthenticationFailure(permission); got != permission || connectors.IsVerifiedAuthenticationFailure(got) {
		t.Fatalf("permission classification = %T %v, want original non-auth refusal", got, got)
	}
	transport := errors.New("connection reset")
	if got := markPostgresAuthenticationFailure(transport); got != transport || connectors.IsVerifiedAuthenticationFailure(got) {
		t.Fatalf("transport classification = %T %v, want original non-auth failure", got, got)
	}
}
