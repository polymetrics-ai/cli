package okta_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/okta"
)

func TestReadPaginatesAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	var calls int
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/users" {
			http.NotFound(w, r)
			return
		}
		calls++
		switch r.URL.Query().Get("after") {
		case "":
			w.Header().Set("Link", fmt.Sprintf(`<%s/api/v1/users?after=u2>; rel="next"`, srv.URL))
			_, _ = w.Write([]byte(`[{"id":"u1","status":"ACTIVE","profile":{"email":"ada@example.com","login":"ada@example.com"}},{"id":"u2","status":"ACTIVE","profile":{"email":"grace@example.com","login":"grace@example.com"}}]`))
		case "u2":
			_, _ = w.Write([]byte(`[{"id":"u3","status":"DEPROVISIONED","profile":{"email":"kay@example.com","login":"kay@example.com"}}]`))
		default:
			t.Fatalf("unexpected after %q", r.URL.Query().Get("after"))
		}
	}))
	defer srv.Close()

	c := okta.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL}, Secrets: map[string]string{"api_token": "okta_tok"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "SSWS okta_tok" {
		t.Fatalf("Authorization = %q, want SSWS token", sawAuth)
	}
	if calls != 2 || len(got) != 3 || got[0]["email"] != "ada@example.com" || got[2]["status"] != "DEPROVISIONED" {
		t.Fatalf("pagination/mapping wrong: calls=%d records=%+v", calls, got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := okta.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "groups", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 || got[0]["id"] == nil {
		t.Fatalf("fixture records missing id: %+v", got)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("okta"); !ok {
		t.Fatal("registry did not resolve okta")
	}
}
