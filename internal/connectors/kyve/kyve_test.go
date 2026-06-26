package kyve_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/kyve"
)

func TestReadPaginatesWithoutAuthAndMapsPools(t *testing.T) {
	var sawAuth string
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/kyve/query/v1beta1/pools" {
			http.NotFound(w, r)
			return
		}
		keys = append(keys, r.URL.Query().Get("pagination.key"))
		switch r.URL.Query().Get("pagination.key") {
		case "":
			_, _ = w.Write([]byte(`{"pools":[{"id":"1","name":"Moonbeam","runtime":"@kyve/evm"},{"id":"2","name":"Cosmos","runtime":"@kyve/cosmos"}],"pagination":{"next_key":"NEXT"}}`))
		case "NEXT":
			_, _ = w.Write([]byte(`{"pools":[{"id":"3","name":"Arweave","runtime":"@kyve/arweave"}],"pagination":{"next_key":""}}`))
		default:
			t.Fatalf("unexpected pagination.key %q", r.URL.Query().Get("pagination.key"))
		}
	}))
	defer srv.Close()

	c := kyve.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "page_size": "2"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pools", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "" {
		t.Fatalf("Authorization = %q, want empty for public KYVE API", sawAuth)
	}
	if len(keys) != 2 || keys[0] != "" || keys[1] != "NEXT" {
		t.Fatalf("keys = %v", keys)
	}
	if len(got) != 3 || got[0]["id"] == nil || got[0]["runtime"] != "@kyve/evm" {
		t.Fatalf("records not mapped: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := kyve.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	var fixture []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pools", Config: cfg}, func(rec connectors.Record) error {
		fixture = append(fixture, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(fixture) == 0 || fixture[0]["fixture"] != true {
		t.Fatalf("fixture records = %+v", fixture)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "kyve" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v err=%v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write err = %v", err)
	}
	if got, ok := connectors.NewRegistry().Get("kyve"); !ok || got.Metadata().Capabilities.Write {
		t.Fatalf("registry/capabilities failed: ok=%v got=%T", ok, got)
	}
}
