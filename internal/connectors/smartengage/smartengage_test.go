package smartengage_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/smartengage"
)

func TestReadAvatarsUsesBearerAuth(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/avatars/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"avatar_id":"1","brand_name":"Primary"}]`))
	}))
	defer srv.Close()

	c := smartengage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_key": "fixture-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "avatars", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer fixture-token" {
		t.Fatalf("Authorization was not set from api_key")
	}
	if len(got) != 1 || got[0]["avatar_id"] == nil {
		t.Fatalf("records = %+v, want avatar record", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := smartengage.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "smartengage" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["avatar_id"] == nil {
		t.Fatalf("fixture records = %+v, want avatar_id", got)
	}
	if _, ok := connectors.NewRegistry().Get("smartengage"); !ok {
		t.Fatal("registry did not resolve smartengage")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
