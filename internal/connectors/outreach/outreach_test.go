package outreach_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/outreach"
)

func TestReadRefreshesTokenPaginatesAndMaps(t *testing.T) {
	var sawTokenRequest bool
	var sawAuth string
	var calls int
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		sawTokenRequest = true
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse token form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Fatalf("grant_type = %q, want refresh_token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"outreach_tok","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/prospects", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		calls++
		switch r.URL.Query().Get("page[after]") {
		case "":
			_, _ = w.Write([]byte(`{"data":[{"id":"p1","type":"prospect","attributes":{"email":"ada@example.com","updatedAt":"2026-01-01T00:00:00Z"}}],"links":{"next":"` + srv.URL + `/prospects?page%5Bafter%5D=p1"}}`))
		case "p1":
			_, _ = w.Write([]byte(`{"data":[{"id":"p2","type":"prospect","attributes":{"email":"grace@example.com","updatedAt":"2026-01-02T00:00:00Z"}}]}`))
		default:
			t.Fatalf("unexpected page[after] %q", r.URL.Query().Get("page[after]"))
		}
	})

	c := outreach.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "token_url": srv.URL + "/oauth/token", "client_id": "client", "redirect_uri": "https://example.com/callback"},
		Secrets: map[string]string{"client_secret": "secret", "refresh_token": "refresh"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "prospects", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !sawTokenRequest || sawAuth != "Bearer outreach_tok" || calls != 2 {
		t.Fatalf("token/auth/pagination wrong: token=%v auth=%q calls=%d", sawTokenRequest, sawAuth, calls)
	}
	if len(got) != 2 || got[0]["email"] != "ada@example.com" || got[1]["updated_at"] != "2026-01-02T00:00:00Z" {
		t.Fatalf("mapped records wrong: %+v", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := outreach.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
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
	if _, ok := connectors.NewRegistry().Get("outreach"); !ok {
		t.Fatal("registry did not resolve outreach")
	}
}
