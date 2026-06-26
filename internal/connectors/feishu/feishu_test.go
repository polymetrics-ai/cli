package feishu_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/feishu"
)

// TestReadAuthenticatesAndPaginates is the red-first test: the connector must
// first exchange app_id/app_secret for a tenant_access_token, then use that
// token as a Bearer credential against the Bitable records endpoint, following
// Feishu's page_token/has_more cursor pagination across two pages and mapping
// each record's fields. Red until internal/connectors/feishu exists.
func TestReadAuthenticatesAndPaginates(t *testing.T) {
	var (
		sawTokenBody  map[string]any
		sawRecordAuth string
		recordCalls   int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &sawTokenBody)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":0,"msg":"ok","tenant_access_token":"t-abc123","expire":7200}`))
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/bitable/v1/apps/app_tok/tables/tbl_1/records":
			sawRecordAuth = r.Header.Get("Authorization")
			recordCalls++
			switch r.URL.Query().Get("page_token") {
			case "":
				_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"items":[{"record_id":"rec1","fields":{"Name":"Ada"}},{"record_id":"rec2","fields":{"Name":"Grace"}}],"has_more":true,"page_token":"pg2","total":3}}`))
			case "pg2":
				_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"items":[{"record_id":"rec3","fields":{"Name":"Katherine"}}],"has_more":false,"total":3}}`))
			default:
				t.Errorf("unexpected page_token=%q", r.URL.Query().Get("page_token"))
				_, _ = w.Write([]byte(`{"code":0,"data":{"items":[],"has_more":false}}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := feishu.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL, "table_id": "tbl_1"},
		Secrets: map[string]string{
			"app_id":     "cli_app",
			"app_secret": "secret_xyz",
			"app_token":  "app_tok",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawTokenBody["app_id"] != "cli_app" || sawTokenBody["app_secret"] != "secret_xyz" {
		t.Fatalf("token request body = %+v, want app_id/app_secret", sawTokenBody)
	}
	if sawRecordAuth != "Bearer t-abc123" {
		t.Fatalf("record Authorization = %q, want Bearer t-abc123", sawRecordAuth)
	}
	if recordCalls != 2 {
		t.Fatalf("record endpoint called %d times, want 2 (paginated)", recordCalls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["record_id"] != "rec1" {
		t.Fatalf("first record_id = %v, want rec1", got[0]["record_id"])
	}
	if got[0]["Name"] != "Ada" {
		t.Fatalf("first record Name = %v, want Ada (fields flattened)", got[0]["Name"])
	}
}

// TestReadTablesStream exercises a second stream (tables) to confirm the routing
// table and that the app-level endpoint (no table_id) is used.
func TestReadTablesStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal":
			_, _ = w.Write([]byte(`{"code":0,"tenant_access_token":"t-1","expire":7200}`))
		case r.URL.Path == "/open-apis/bitable/v1/apps/app_tok/tables":
			_, _ = w.Write([]byte(`{"code":0,"data":{"items":[{"table_id":"tbl_1","name":"Sheet1","revision":3}],"has_more":false}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := feishu.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"app_id": "a", "app_secret": "s", "app_token": "app_tok"},
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "tables", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read tables: %v", err)
	}
	if len(got) != 1 || got[0]["table_id"] != "tbl_1" || got[0]["name"] != "Sheet1" {
		t.Fatalf("tables = %+v, want one table tbl_1/Sheet1", got)
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records
// with no HTTP server, so conformance runs without credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := feishu.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "table_id": "tbl_1"}}
	for _, stream := range []string{"records", "tables", "fields"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
	}
	// Check must succeed in fixture mode without creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) and hostless base_url overrides.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := feishu.New()
	for _, bad := range []string{"file:///etc/passwd", "ftp://example.com", "://nohost"} {
		cfg := connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": bad, "table_id": "t"},
			Secrets: map[string]string{"app_id": "a", "app_secret": "s", "app_token": "x"},
		}
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: cfg}, func(connectors.Record) error { return nil })
		if err == nil {
			t.Fatalf("base_url %q should be rejected", bad)
		}
	}
}

// TestRegistryResolvesFeishu confirms self-registration and read-only caps.
func TestRegistryResolvesFeishu(t *testing.T) {
	_ = feishu.New() // ensure init ran
	r := connectors.NewRegistry()
	c, ok := r.Get("feishu")
	if !ok {
		t.Fatal("registry did not resolve feishu (self-registration)")
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Catalog && !Write", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("catalog streams = %d, want >= 3", len(cat.Streams))
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for _, want := range []string{"records", "tables", "fields"} {
		if !names[want] {
			t.Fatalf("catalog missing stream %q", want)
		}
	}
}

func TestMissingSecretsRejected(t *testing.T) {
	c := feishu.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"table_id": "t"}}
	if err := c.Check(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "feishu") {
		t.Fatalf("Check without secrets = %v, want feishu error", err)
	}
}
