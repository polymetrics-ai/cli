package taboola_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/taboola"
)

func TestReadUsesOAuthAndAccountPath(t *testing.T) {
	var tokenRequests int
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/backstage/oauth/token":
			tokenRequests++
			if r.Method != http.MethodPost {
				t.Fatalf("token method = %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"access_token":"test_access_token","token_type":"bearer","expires_in":3600}`))
		case "/backstage/api/1.0/acct/campaigns":
			sawAuth = r.Header.Get("Authorization")
			if r.URL.Query().Get("page") == "2" {
				_, _ = w.Write([]byte(`{"results":[{"id":"camp_3","name":"Retargeting","created_at":"2026-01-03T00:00:00Z"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"results":[{"id":"camp_1","name":"Launch","created_at":"2026-01-01T00:00:00Z"},{"id":"camp_2","name":"Scale","created_at":"2026-01-02T00:00:00Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := taboola.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "account_id": "acct", "page_size": "2"}, Secrets: map[string]string{"client_id": "client_test", "client_secret": "client_test_value"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenRequests != 1 {
		t.Fatalf("token requests = %d, want 1", tokenRequests)
	}
	if sawAuth != "Bearer test_access_token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(got) != 3 || !strings.Contains(got[2]["name"].(string), "Retargeting") {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureModeNoCredentials(t *testing.T) {
	c := taboola.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var count int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		count++
		if rec["id"] == nil {
			t.Fatalf("fixture missing id: %+v", rec)
		}
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture emitted no records")
	}
}

func TestCatalogRegistrationAndReadOnly(t *testing.T) {
	c := taboola.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "taboola" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	if caps := c.Metadata().Capabilities; !caps.Check || !caps.Catalog || !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v", caps)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("taboola"); !ok {
		t.Fatal("registry did not resolve taboola")
	}
}
