package oura_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/oura"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var tokens []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/daily_sleep" {
			http.NotFound(w, r)
			return
		}
		tokens = append(tokens, r.URL.Query().Get("next_token"))
		switch r.URL.Query().Get("next_token") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"sleep_1","day":"2026-01-01","score":88}],"next_token":"tok_2"}`))
		case "tok_2":
			_, _ = w.Write([]byte(`{"data":[{"id":"sleep_2","day":"2026-01-02","score":91}]}`))
		default:
			t.Fatalf("unexpected next_token %q", r.URL.Query().Get("next_token"))
		}
	}))
	defer srv.Close()

	c := oura.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "start_datetime": "2026-01-01T00:00:00Z"}, Secrets: map[string]string{"api_key": "oura_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily_sleep", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer oura_key" {
		t.Fatalf("Authorization = %q, want Bearer", sawAuth)
	}
	if len(tokens) != 2 || tokens[1] != "tok_2" || len(got) != 2 || got[0]["day"] != "2026-01-01" {
		t.Fatalf("pagination/mapping wrong: tokens=%v records=%+v", tokens, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := oura.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily_activity", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records missing id: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("oura"); !ok {
		t.Fatal("registry did not resolve oura")
	}
}
