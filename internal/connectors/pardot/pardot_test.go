package pardot_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/pardot"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawBusinessUnit string
	var srv *httptest.Server
	var sawOffset string
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawBusinessUnit = r.Header.Get("Pardot-Business-Unit-Id")
		if r.URL.Path != "/api/v5/objects/prospects" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"values":[{"id":1,"email":"a@example.com","firstName":"Ada"},{"id":2,"email":"b@example.com","firstName":"Ben"}],"nextPageUrl":"` + srv.URL + `/api/v5/objects/prospects?offset=2&limit=2"}`))
		case "2":
			sawOffset = "2"
			_, _ = w.Write([]byte(`{"values":[{"id":3,"email":"c@example.com","firstName":"Cid"}],"nextPageUrl":null}`))
		default:
			t.Fatalf("unexpected offset query %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := pardot.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "business_unit_id": "0Uvxx0000000001", "limit": "2"}, Secrets: map[string]string{"access_token": "pardot_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "prospects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer pardot_token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if sawBusinessUnit != "0Uvxx0000000001" {
		t.Fatalf("Pardot-Business-Unit-Id = %q", sawBusinessUnit)
	}
	if sawOffset != "2" {
		t.Fatalf("second page offset = %q, want 2", sawOffset)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[2]["email"] != "c@example.com" {
		t.Fatalf("mapped records = %+v, want three prospects", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := pardot.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "pardot" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want pardot streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("pardot"); !ok {
		t.Fatal("registry did not resolve pardot")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
