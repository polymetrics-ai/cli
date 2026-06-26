package yandexmetrica_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	yandexmetrica "polymetrics.ai/internal/connectors/yandex-metrica"
)

func TestReadTrafficPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var offsets []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/stat/v1/data" {
			http.NotFound(w, r)
			return
		}
		offsets = append(offsets, r.URL.Query().Get("offset"))
		if r.URL.Query().Get("ids") != "123" {
			t.Fatalf("ids = %q, want 123", r.URL.Query().Get("ids"))
		}
		switch r.URL.Query().Get("offset") {
		case "", "1":
			_, _ = w.Write([]byte(`{"total_rows":2,"data":[{"dimensions":[{"name":"Google","id":"google"}],"metrics":[10]}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_rows":2,"data":[{"dimensions":[{"name":"Direct","id":"direct"}],"metrics":[3]}]}`))
		default:
			t.Fatalf("unexpected offset=%q", r.URL.Query().Get("offset"))
		}
	}))
	defer srv.Close()

	c := yandexmetrica.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "counter_id": "123", "start_date": "2026-01-01", "end_date": "2026-01-02", "limit": "1"}, Secrets: map[string]string{"auth_token": "test-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "traffic", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if len(offsets) != 2 || len(got) != 2 {
		t.Fatalf("offsets=%v records=%d, want two pages", offsets, len(got))
	}
	if got[0]["dimension_1_name"] != "Google" || got[1]["metric_1"] != float64(3) {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := yandexmetrica.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var n int
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "traffic", Config: cfg}, func(connectors.Record) error { n++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if n == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("yandex-metrica"); !ok {
		t.Fatal("registry did not resolve yandex-metrica")
	}
}
