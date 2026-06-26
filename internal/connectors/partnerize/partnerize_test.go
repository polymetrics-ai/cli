package partnerize_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/partnerize"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var sawOffset string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/conversions" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"data":[{"id":"conv_1","status":"approved","value":1200},{"id":"conv_2","status":"pending","value":3400}],"meta":{"total_count":3}}`))
		case "2":
			sawOffset = "2"
			_, _ = w.Write([]byte(`{"data":[{"id":"conv_3","status":"rejected","value":100}],"meta":{"total_count":3}}`))
		default:
			t.Fatalf("unexpected offset query %q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := partnerize.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "limit": "2"}, Secrets: map[string]string{"application_key": "app_key", "user_api_key": "user_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "conversions", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("app_key:user_key"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if sawOffset != "2" {
		t.Fatalf("second page offset = %q, want 2", sawOffset)
	}
	if len(got) != 3 || got[0]["id"] != "conv_1" || got[2]["status"] != "rejected" {
		t.Fatalf("mapped records = %+v, want three conversions", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := partnerize.New()
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
	if cat.Connector != "partnerize" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want partnerize streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("partnerize"); !ok {
		t.Fatal("registry did not resolve partnerize")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
