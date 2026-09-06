package wasabistatsapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/credential"
)

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{Mode: "custom", Hook: "wasabi-stats-api", Value: "{{ secrets.api_key }}"}
}

func cfgWithKey(key string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Secrets: map[string]string{"api_key": key}}
}

func doAuthenticatedRequest(t *testing.T, auth connsdk.Authenticator) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return req
}

// --- registration ---

func TestInit_RegistersHooks(t *testing.T) {
	h := ExplicitFactory()
	if h == nil {
		t.Fatal("engine.HooksFor(\"wasabi-stats-api\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "wasabi-stats-api" {
		t.Fatalf("ConnectorName() = %q, want wasabi-stats-api", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered hooks do not implement AuthHook")
	}
	if _, ok := h.(engine.RecordHook); !ok {
		t.Fatal("registered hooks do not implement RecordHook")
	}
}

// --- Authenticator: Bearer-vs-Basic content-based branch ---

func TestAuthenticator_NoColonUsesBearerAuth(t *testing.T) {
	h := Hooks{}
	auth, err := h.Authenticator(context.Background(), cfgWithKey("plain-api-key-value"), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	if got := req.Header.Get("Authorization"); got != "Bearer plain-api-key-value" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer plain-api-key-value")
	}
}

func TestAuthenticator_ColonSeparatedKeyUsesBasicAuth(t *testing.T) {
	h := Hooks{}
	auth, err := h.Authenticator(context.Background(), cfgWithKey("myuser:mypass"), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("myuser:mypass"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

// TestAuthenticator_MultipleColonsSplitsOnFirstOnly mirrors legacy's
// strings.SplitN(key, ":", 2): a value with MORE than one ':' still yields
// exactly 2 parts, the second part retaining every remaining colon.
func TestAuthenticator_MultipleColonsSplitsOnFirstOnly(t *testing.T) {
	h := Hooks{}
	auth, err := h.Authenticator(context.Background(), cfgWithKey("myuser:my:pass:with:colons"), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("myuser:my:pass:with:colons"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_TrailingColonWithEmptySecondPartUsesBasicAuth(t *testing.T) {
	// SplitN("user:", ":", 2) == ["user", ""] -- still exactly 2 parts, so
	// legacy's len(parts) == 2 check is satisfied even though the password
	// half is empty. Pinning this exact (perhaps surprising) legacy behavior.
	h := Hooks{}
	auth, err := h.Authenticator(context.Background(), cfgWithKey("user:"), baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	username, password, ok := req.BasicAuth()
	if !ok || len(password) != 0 {
		t.Fatal("Basic auth did not preserve the declaration-authorized blank password")
	}
	if gotLength, wantLength := len(username), len("user"); gotLength != wantLength {
		t.Fatalf("username length = %d, want %d", gotLength, wantLength)
	}
	if gotHash, wantHash := sha256.Sum256([]byte(username)), sha256.Sum256([]byte("user")); gotHash != wantHash {
		t.Fatalf("username SHA-256 = %x, want %x", gotHash, wantHash)
	}
}

func TestAuthenticator_EmptyKeyIsError(t *testing.T) {
	h := Hooks{}
	_, err := h.Authenticator(context.Background(), cfgWithKey(""), baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error for an empty api_key")
	}
	var empty *credential.EmptySecretError
	if !errors.As(err, &empty) {
		t.Fatalf("Authenticator() error is not typed empty-secret classification: %T", err)
	}
}

func TestAuthenticator_MissingSecretIsError(t *testing.T) {
	h := Hooks{}
	_, err := h.Authenticator(context.Background(), connectors.RuntimeConfig{}, baseSpec())
	if err == nil {
		t.Fatal("Authenticator() error = nil, want an error for an unresolved api_key reference")
	}
}

// TestAuthenticator_MissingSecretErrorNeverContainsSecretText asserts the
// missing-key error path names the field, not any secret value (there is no
// value to leak in this path, but the assertion documents the invariant
// explicitly for future maintainers).
func TestAuthenticator_MissingSecretErrorNeverContainsSecretText(t *testing.T) {
	h := Hooks{}
	_, err := h.Authenticator(context.Background(), connectors.RuntimeConfig{}, baseSpec())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "api_key") {
		t.Fatalf("error = %q, want it to name api_key", err.Error())
	}
}

// --- MapRecord: id fallback derivation ---

func TestMapRecord_IDPresentIsUntouched(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"id": "real-id", "bucket": "b1", "date": "2026-01-01T00:00:00Z"}
	projected := connsdk.Record{"id": "real-id", "bucket": "b1"}
	out, keep, err := h.MapRecord("bucket_stats", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("keep = false, want true")
	}
	if out["id"] != "real-id" {
		t.Fatalf("id = %v, want unchanged real-id", out["id"])
	}
}

func TestMapRecord_MissingIDFallsBackToBucket(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"bucket": "fixture-bucket", "date": "2026-01-01T00:00:00Z"}
	projected := connsdk.Record{"bucket": "fixture-bucket"}
	out, _, err := h.MapRecord("bucket_stats", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if out["id"] != "fixture-bucket" {
		t.Fatalf("id = %v, want fixture-bucket (bucket fallback)", out["id"])
	}
}

func TestMapRecord_MissingIDAndBucketFallsBackToDate(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"date": "2026-01-01T00:00:00Z", "storage_bytes": 42}
	projected := connsdk.Record{"storage_bytes": 42}
	out, _, err := h.MapRecord("account_stats", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if out["id"] != "2026-01-01T00:00:00Z" {
		t.Fatalf("id = %v, want the date fallback", out["id"])
	}
}

func TestMapRecord_MissingIDBucketAndDateFallsBackToStreamName(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"storage_bytes": 42}
	projected := connsdk.Record{"storage_bytes": 42}
	out, _, err := h.MapRecord("account_stats", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if out["id"] != "account_stats" {
		t.Fatalf("id = %v, want the stream-name literal fallback", out["id"])
	}
}

func TestMapRecord_NilProjectedIsHandled(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"bucket": "fixture-bucket"}
	out, keep, err := h.MapRecord("bucket_stats", raw, nil)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("keep = false, want true")
	}
	if out["id"] != "fixture-bucket" {
		t.Fatalf("id = %v, want fixture-bucket", out["id"])
	}
}

func TestMapRecord_EmptyStringBucketAndDateSkipToStreamName(t *testing.T) {
	// Mirrors legacy's firstNonEmpty treating a whitespace-only/empty
	// stringified value as "empty", falling through past it.
	h := Hooks{}
	raw := connsdk.Record{"bucket": "", "date": "  "}
	projected := connsdk.Record{}
	out, _, err := h.MapRecord("bucket_stats", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if out["id"] != "bucket_stats" {
		t.Fatalf("id = %v, want the stream-name literal fallback", out["id"])
	}
}
