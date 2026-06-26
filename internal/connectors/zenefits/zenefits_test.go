package zenefits_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/zenefits"
)

func TestReadPeopleAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/core/people" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"person_1","first_name":"Ada","last_name":"Lovelace","status":"active"}]}`))
	}))
	defer srv.Close()

	c := zenefits.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/core"}, Secrets: map[string]string{"token": "zenefits_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer zenefits_token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "person_1" || got[0]["first_name"] != "Ada" {
		t.Fatalf("records = %+v, want mapped person", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zenefits.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "companies", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "zenefits" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zenefits streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zenefits"); !ok {
		t.Fatal("registry did not resolve zenefits")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
