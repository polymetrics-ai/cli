package zohoanalyticsmetadataapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohoanalyticsmetadataapi "polymetrics.ai/internal/connectors/zoho-analytics-metadata-api"
)

func TestReadWorkspacesRefreshesTokenAuthenticatesAndMaps(t *testing.T) {
	var sawAuth, tokenForm string
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/v2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		tokenForm = r.Form.Encode()
		_, _ = w.Write([]byte(`{"access_token":"ZOHO_AT","expires_in":3600}`))
	})
	mux.HandleFunc("/restapi/v2/workspaces", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":[{"workspaceId":"ws_1","workspaceName":"Finance","createdTime":"2026-01-01T00:00:00Z"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := zohoanalyticsmetadataapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/restapi/v2", "token_url": srv.URL + "/oauth/v2/token", "data_center": "com", "org_id": "12345"}, Secrets: map[string]string{"client_id": "client", "client_secret": "secret", "refresh_token": "refresh"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "workspaces", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ZOHO_AT" {
		t.Fatalf("Authorization = %q, want refreshed bearer token", sawAuth)
	}
	if !strings.Contains(tokenForm, "grant_type=refresh_token") || !strings.Contains(tokenForm, "refresh_token=refresh") {
		t.Fatalf("token form = %q, want refresh_token grant", tokenForm)
	}
	if len(got) != 1 || got[0]["id"] != "ws_1" || got[0]["name"] != "Finance" {
		t.Fatalf("records = %+v, want mapped workspace", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohoanalyticsmetadataapi.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "views", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zoho-analytics-metadata-api" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho analytics metadata streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-analytics-metadata-api"); !ok {
		t.Fatal("registry did not resolve zoho-analytics-metadata-api")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
