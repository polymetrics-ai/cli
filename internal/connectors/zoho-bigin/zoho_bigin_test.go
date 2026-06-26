package zohobigin_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	zohobigin "polymetrics.ai/internal/connectors/zoho-bigin"
)

func TestReadPipelinesRefreshesTokenAuthenticatesAndMaps(t *testing.T) {
	var sawAuth, tokenForm string
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/v2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		tokenForm = r.Form.Encode()
		_, _ = w.Write([]byte(`{"access_token":"BIGIN_AT","expires_in":3600}`))
	})
	mux.HandleFunc("/bigin/v2/Pipelines", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":[{"id":"pipe_1","name":"Sales","display_value":"Sales"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := zohobigin.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/bigin/v2", "token_url": srv.URL + "/oauth/v2/token", "data_center": "com", "module_name": "Deals"}, Secrets: map[string]string{"client_id": "client", "client_secret": "secret", "client_refresh_token": "refresh"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "pipelines", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer BIGIN_AT" {
		t.Fatalf("Authorization = %q, want refreshed bearer token", sawAuth)
	}
	if !strings.Contains(tokenForm, "grant_type=refresh_token") || !strings.Contains(tokenForm, "refresh_token=refresh") {
		t.Fatalf("token form = %q, want refresh_token grant", tokenForm)
	}
	if len(got) != 1 || got[0]["id"] != "pipe_1" || got[0]["name"] != "Sales" {
		t.Fatalf("records = %+v, want mapped pipeline", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := zohobigin.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: fixture}, func(rec connectors.Record) error {
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
	if cat.Connector != "zoho-bigin" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want zoho bigin streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("zoho-bigin"); !ok {
		t.Fatal("registry did not resolve zoho-bigin")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
