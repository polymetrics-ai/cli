package onepagecrm_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/onepagecrm"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Basic auth (user ID +
// API key), page/per_page pagination across 2 pages, and the nested
// data.contacts[*].contact record unwrapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/contacts" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(`{"status":0,"data":{"page":1,"max_page":2,"contacts":[{"contact":{"id":"c1","first_name":"Ada","last_name":"Lovelace"}},{"contact":{"id":"c2","first_name":"Grace","last_name":"Hopper"}}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"status":0,"data":{"page":2,"max_page":2,"contacts":[{"contact":{"id":"c3","first_name":"Katherine","last_name":"Johnson"}}]}}`))
		default:
			t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
			_, _ = w.Write([]byte(`{"status":0,"data":{"page":3,"max_page":2,"contacts":[]}}`))
		}
	}))
	defer srv.Close()

	c := onepagecrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "user_42"},
		Secrets: map[string]string{"password": "apikey_secret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user_42:apikey_secret"))
	if sawAuth != wantAuth {
		t.Fatalf("Authorization = %q, want %q", sawAuth, wantAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing unwrapped id: %+v", rec)
		}
		if rec["first_name"] == nil {
			t.Fatalf("record missing first_name (unwrap failed): %+v", rec)
		}
	}
	if got[0]["id"] != "c1" {
		t.Fatalf("first record id = %v, want c1", got[0]["id"])
	}
}

// TestReadUsersTopLevelArray exercises the data[*].user selector shape (users
// list returns an array directly under data rather than a named sub-array).
func TestReadUsersTopLevelArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":0,"data":[{"user":{"id":"u1","email":"a@example.com"}},{"user":{"id":"u2","email":"b@example.com"}}]}`))
	}))
	defer srv.Close()

	c := onepagecrm.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "username": "user_42"},
		Secrets: map[string]string{"password": "apikey_secret"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read users: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("users = %d, want 2", len(got))
	}
	if got[0]["email"] != "a@example.com" {
		t.Fatalf("user email = %v, want a@example.com", got[0]["email"])
	}
}

func TestFixtureModeNoNetwork(t *testing.T) {
	c := onepagecrm.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"contacts", "deals", "actions", "users", "companies"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) returned no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := onepagecrm.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture mode: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := onepagecrm.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	wantPK := false
	for _, s := range cat.Streams {
		if s.Name == "contacts" {
			if len(s.PrimaryKey) == 1 && s.PrimaryKey[0] == "id" {
				wantPK = true
			}
		}
	}
	if !wantPK {
		t.Fatal("contacts stream must have primary key [id]")
	}
}

func TestRegistryResolves(t *testing.T) {
	_ = onepagecrm.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("onepagecrm"); !ok {
		t.Fatal("registry did not resolve onepagecrm (self-registration)")
	}
}

func TestMetadataReadOnly(t *testing.T) {
	caps := onepagecrm.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
}
