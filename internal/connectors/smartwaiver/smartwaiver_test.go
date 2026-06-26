package smartwaiver_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/smartwaiver"
)

func TestReadWaiversUsesBearerAndOffset(t *testing.T) {
	var sawAuth string
	var sawLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawLimit = r.URL.Query().Get("limit")
		if r.URL.Path != "/v4/waivers" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"waivers":{"waivers":[{"waiverId":"w_1","templateId":"t_1"}]}}`))
	}))
	defer srv.Close()

	c := smartwaiver.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "1"}, Secrets: map[string]string{"api_key": "fixture-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "waivers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fixture-token" || sawLimit != "1" {
		t.Fatalf("auth/page query not set as expected")
	}
	if len(got) != 1 || got[0]["waiverId"] == nil {
		t.Fatalf("records = %+v, want waiver record", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := smartwaiver.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "smartwaiver" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["waiverId"] == nil {
		t.Fatalf("fixture records = %+v, want waiverId", got)
	}
	if _, ok := connectors.NewRegistry().Get("smartwaiver"); !ok {
		t.Fatal("registry did not resolve smartwaiver")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
