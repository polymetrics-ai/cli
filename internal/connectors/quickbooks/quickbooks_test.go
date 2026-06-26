package quickbooks_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/quickbooks"
)

func TestReadCustomersAuthenticatesPaginatesAndMapsRecords(t *testing.T) {
	var sawAuth string
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v3/company/123/query" {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query().Get("query")
		queries = append(queries, query)
		switch {
		case strings.Contains(query, "STARTPOSITION 1"):
			_, _ = w.Write([]byte(`{"QueryResponse":{"Customer":[{"Id":"1","DisplayName":"Ada Lovelace","Active":true},{"Id":"2","DisplayName":"Grace Hopper","Active":true}]}}`))
		case strings.Contains(query, "STARTPOSITION 3"):
			_, _ = w.Write([]byte(`{"QueryResponse":{"Customer":[{"Id":"3","DisplayName":"Katherine Johnson","Active":true}]}}`))
		default:
			t.Fatalf("unexpected query %q", query)
		}
	}))
	defer srv.Close()

	c := quickbooks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "realm_id": "123", "page_size": "2"}, Secrets: map[string]string{"access_token": "qb_token"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer qb_token" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if len(queries) != 2 {
		t.Fatalf("queries = %v, want two pages", queries)
	}
	if len(got) != 3 || got[0]["id"] != "1" || got[0]["display_name"] != "Ada Lovelace" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := quickbooks.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "quickbooks" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v", cat)
	}
	for _, stream := range cat.Streams {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("Read fixture %s: %v", stream.Name, err)
		}
		if len(got) == 0 || got[0]["id"] == nil {
			t.Fatalf("fixture %s records = %+v", stream.Name, got)
		}
	}
	if _, ok := connectors.NewRegistry().Get("quickbooks"); !ok {
		t.Fatal("registry did not resolve quickbooks")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v", err)
	}
}
