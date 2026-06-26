package salesforce_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/salesforce"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/services/data/v60.0/query":
			if !strings.Contains(r.URL.Query().Get("q"), "FROM Account") {
				t.Fatalf("SOQL query = %q, want Account", r.URL.Query().Get("q"))
			}
			_, _ = w.Write([]byte(`{"records":[{"Id":"001","Name":"Acme","LastModifiedDate":"2026-01-01T00:00:00Z"}],"nextRecordsUrl":"/services/data/v60.0/query/01g-next"}`))
		case "/services/data/v60.0/query/01g-next":
			_, _ = w.Write([]byte(`{"records":[{"Id":"002","Name":"Globex","LastModifiedDate":"2026-01-02T00:00:00Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := salesforce.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"instance_url": srv.URL, "api_version": "v60.0"}, Secrets: map[string]string{"access_token": "test-access-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-access-token" {
		t.Fatalf("Authorization = %q, want bearer test token", sawAuth)
	}
	if len(got) != 2 || got[0]["id"] != "001" || got[0]["name"] != "Acme" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := salesforce.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "salesforce" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want salesforce streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("salesforce"); !ok {
		t.Fatal("registry did not resolve salesforce")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
