package smaily_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/smaily"
)

func TestReadCampaignsUsesBasicAuth(t *testing.T) {
	var sawUser string
	var sawBasic bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		sawUser, sawBasic = user, ok && pass != ""
		if r.URL.Path != "/api/campaign.php" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":1,"name":"Welcome"},{"id":2,"name":"Followup"}]`))
	}))
	defer srv.Close()

	c := smaily.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "api_username": "api-user", "api_subdomain": "example"}, Secrets: map[string]string{"api_password": "fixture-password"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawBasic || sawUser != "api-user" {
		t.Fatalf("basic auth was not set from api_username/api_password")
	}
	if len(got) != 2 || got[0]["id"] == nil {
		t.Fatalf("records = %+v, want two campaign records", got)
	}
}

func TestFixtureRegistryCatalogAndWrite(t *testing.T) {
	c := smaily.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "smaily" || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Config: cfg}, func(rec connectors.Record) error { got = append(got, rec); return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records = %+v, want id", got)
	}
	if _, ok := connectors.NewRegistry().Get("smaily"); !ok {
		t.Fatal("registry did not resolve smaily")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
