package elasticsearch

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func baseSpec() engine.AuthSpec {
	return engine.AuthSpec{Mode: "custom", Hook: "elasticsearch", Value: "{{ config.api_key_id }}"}
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
	h := engine.HooksFor("elasticsearch")
	if h == nil {
		t.Fatal("engine.HooksFor(\"elasticsearch\") = nil, want a registered hook set (init() must call engine.RegisterHooks)")
	}
	if h.ConnectorName() != "elasticsearch" {
		t.Fatalf("ConnectorName() = %q, want elasticsearch", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("registered hooks do not implement AuthHook")
	}
	if _, ok := h.(engine.RecordHook); !ok {
		t.Fatal("registered hooks do not implement RecordHook")
	}
}

// --- Authenticator: composite API-key base64 encoding ---

func TestAuthenticator_BuildsApiKeyHeader(t *testing.T) {
	h := Hooks{}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"api_key_id": "id123"},
		Secrets: map[string]string{"api_key_secret": "secret456"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	want := "ApiKey " + base64.StdEncoding.EncodeToString([]byte("id123:secret456"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_MissingIDFallsBackToBasic(t *testing.T) {
	h := Hooks{}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"username": "elastic"},
		Secrets: map[string]string{"api_key_secret": "secret456", "password": "password"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("elastic:password"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_MissingSecretFallsBackToBasic(t *testing.T) {
	h := Hooks{}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"api_key_id": "id123", "username": "elastic"},
		Secrets: map[string]string{"password": "password"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("elastic:password"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestAuthenticator_IncompleteAPIKeyWithoutBasicIsNoAuth(t *testing.T) {
	h := Hooks{}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"api_key_id": "id123"}}
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization = %q, want empty header", got)
	}
}

func TestAuthenticator_AcceptsLegacyCamelCaseAPIKeyNames(t *testing.T) {
	h := Hooks{}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"apiKeyId": "id123"},
		Secrets: map[string]string{"apiKeySecret": "secret456"},
	}
	auth, err := h.Authenticator(context.Background(), cfg, baseSpec())
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := doAuthenticatedRequest(t, auth)
	want := "ApiKey " + base64.StdEncoding.EncodeToString([]byte("id123:secret456"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

// --- MapRecord: documents-stream _source flatten + id stamp ---

func TestMapRecord_DocumentsFlattensSourceAndStampsID(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{
		"_index":  "fixture_index_1",
		"_id":     "doc_fixture_1",
		"_score":  1.0,
		"_source": map[string]any{"order_number": "F-0001", "status": "open"},
	}
	projected := connsdk.Record{}
	out, keep, err := h.MapRecord("documents", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("keep = false, want true")
	}
	if out["id"] != "doc_fixture_1" {
		t.Fatalf("id = %v, want doc_fixture_1", out["id"])
	}
	if out["order_number"] != "F-0001" || out["status"] != "open" {
		t.Fatalf("out = %v, want _source fields flattened to top level", out)
	}
	if _, ok := out["_index"]; ok {
		t.Fatal("_index leaked into the flattened record, want dropped like legacy's mapHit")
	}
	if _, ok := out["_score"]; ok {
		t.Fatal("_score leaked into the flattened record, want dropped like legacy's mapHit")
	}
}

func TestMapRecord_DocumentsNilProjectedIsHandled(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"_id": "doc_fixture_2", "_source": map[string]any{"status": "closed"}}
	out, keep, err := h.MapRecord("documents", raw, nil)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("keep = false, want true")
	}
	if out["id"] != "doc_fixture_2" || out["status"] != "closed" {
		t.Fatalf("out = %v, want id+status populated", out)
	}
}

func TestMapRecord_NonDocumentsStreamIsUntouched(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"index": "fixture_index_1", "docs.count": "12"}
	projected := connsdk.Record{"index": "fixture_index_1", "docs.count": "12"}
	out, keep, err := h.MapRecord("indices", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("keep = false, want true")
	}
	if out["index"] != "fixture_index_1" || out["docs.count"] != "12" {
		t.Fatalf("out = %v, want the indices record unchanged", out)
	}
	if _, ok := out["id"]; ok {
		t.Fatal("id was stamped on a non-documents stream, want untouched")
	}
}

func TestMapRecord_MissingSourceIsHandledGracefully(t *testing.T) {
	h := Hooks{}
	raw := connsdk.Record{"_id": "doc_fixture_3"}
	projected := connsdk.Record{}
	out, keep, err := h.MapRecord("documents", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("keep = false, want true")
	}
	if out["id"] != "doc_fixture_3" {
		t.Fatalf("id = %v, want doc_fixture_3 even with no _source", out["id"])
	}
}
