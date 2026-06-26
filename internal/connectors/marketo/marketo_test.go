package marketo_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/marketo"
)

func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/leads.json" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("nextPageToken") {
		case "":
			_, _ = w.Write([]byte(`{"success":true,"result":[{"id":101,"email":"one@example.com","updatedAt":"2026-01-01T00:00:00Z"}],"nextPageToken":"token-2","moreResult":true}`))
		case "token-2":
			_, _ = w.Write([]byte(`{"success":true,"result":[{"id":102,"email":"two@example.com","updatedAt":"2026-01-02T00:00:00Z"}],"moreResult":false}`))
		default:
			t.Fatalf("unexpected nextPageToken=%q", r.URL.Query().Get("nextPageToken"))
		}
	}))
	defer srv.Close()

	c := marketo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"access_token": "test-access-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-access-token" {
		t.Fatalf("Authorization = %q, want bearer test token", sawAuth)
	}
	if len(got) != 2 || got[0]["id"] == nil || got[0]["email"] != "one@example.com" {
		t.Fatalf("records not paginated/mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := marketo.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "marketo" || len(cat.Streams) == 0 {
		t.Fatalf("catalog = %+v, want marketo streams", cat)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "leads", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records not mapped: %+v", got)
	}
	if _, ok := connectors.NewRegistry().Get("marketo"); !ok {
		t.Fatal("registry did not resolve marketo")
	}
	_, err = c.Write(context.Background(), connectors.WriteRequest{Config: cfg}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
