package commercetools_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commercetools"
)

func TestReadCustomersPaginatesAuthenticatesAndMapsRecords(t *testing.T) {
	var tokenCalls int
	var sawBearer string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			tokenCalls++
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "client_credentials" {
				t.Fatalf("grant_type = %q", r.Form.Get("grant_type"))
			}
			_, _ = w.Write([]byte(`{"access_token":"ct_tok","expires_in":3600}`))
		case "/proj/customers":
			sawBearer = r.Header.Get("Authorization")
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			switch offset {
			case 0:
				_, _ = w.Write([]byte(`{"limit":2,"offset":0,"total":3,"results":[{"id":"cust_1","email":"a@example.com"},{"id":"cust_2","email":"b@example.com"}]}`))
			case 2:
				_, _ = w.Write([]byte(`{"limit":2,"offset":2,"total":3,"results":[{"id":"cust_3","email":"c@example.com"}]}`))
			default:
				t.Fatalf("unexpected offset %d", offset)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := commercetools.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "token_url": srv.URL + "/oauth/token", "project_key": "proj", "page_size": "2"}, Secrets: map[string]string{"client_id": "cid", "client_secret": "secret"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalls == 0 || sawBearer != "Bearer ct_tok" {
		t.Fatalf("auth wrong tokenCalls=%d bearer=%q", tokenCalls, sawBearer)
	}
	if len(got) != 3 || got[0]["id"] != "cust_1" || got[0]["email"] != "a@example.com" {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := commercetools.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"customers", "orders", "products"} {
		count := 0
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if count == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "commercetools" || len(cat.Streams) != 3 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("commercetools"); !ok {
		t.Fatal("registry did not resolve commercetools")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
