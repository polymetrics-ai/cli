package zohocampaign_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohocampaign "polymetrics.ai/internal/connectors/zoho-campaign"
)

func TestReadCampaignsAuthenticatesAndExtractsWrappedRecords(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/campaigns/campaigns" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"campaigns":[{"campaign_key":"camp_1","campaign_name":"Launch","status":"sent","modified_time":"2026-01-02T00:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := zohocampaign.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/campaigns"}, Secrets: map[string]string{"access_token": "test_access_token"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Zoho-oauthtoken test_access_token" {
		t.Fatalf("Authorization = %q, want Zoho-oauthtoken", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "camp_1" || got[0]["name"] != "Launch" {
		t.Fatalf("records = %+v, want mapped campaign", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohocampaign.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "lists", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zoho-campaign" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho-campaign streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-campaign"); !ok {
		t.Fatal("registry did not resolve zoho-campaign")
	}
	if c.Metadata().Capabilities.Write {
		t.Fatal("zoho-campaign should be read-only")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
