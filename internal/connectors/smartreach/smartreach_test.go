package smartreach_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/smartreach"
)

func TestReadCampaignsUsesAPIKeyAndTeam(t *testing.T) {
	var sawKey string
	var sawTeam string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("X-API-KEY")
		sawTeam = r.URL.Query().Get("team_id")
		if r.URL.Path != "/campaigns" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"campaigns":[{"id":"cmp_1","name":"Outbound"}],"links":{"next":""}}`))
	}))
	defer srv.Close()

	c := smartreach.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "teamid": "team_1"}, Secrets: map[string]string{"api_key": "fixture-token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "fixture-token" || sawTeam != "team_1" {
		t.Fatalf("auth/team query not set as expected")
	}
	if len(got) != 1 || got[0]["id"] == nil {
		t.Fatalf("records = %+v, want campaign record", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := smartreach.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "smartreach" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, ok := connectors.NewRegistry().Get("smartreach"); !ok {
		t.Fatal("registry did not resolve smartreach")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
