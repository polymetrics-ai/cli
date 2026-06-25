package box_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/box"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Box connector.
// It exercises the OAuth2 client-credentials token exchange, the Bearer auth
// header carried on the data request, Box offset/total_count pagination across
// two pages of /users, and record mapping. Red until internal/connectors/box
// implements Read.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		sawDataAuth   string
		sawGrantType  string
		sawSubjectTyp string
		tokenCalls    int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls++
			_ = r.ParseForm()
			sawGrantType = r.Form.Get("grant_type")
			sawSubjectTyp = r.Form.Get("box_subject_type")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok_box_123","token_type":"bearer","expires_in":3600}`))
		case "/2.0/users":
			sawDataAuth = r.Header.Get("Authorization")
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			switch offset {
			case 0:
				_, _ = w.Write([]byte(`{"total_count":3,"offset":0,"limit":2,"entries":[{"id":"11","type":"user","name":"Ada","login":"ada@example.com","status":"active"},{"id":"22","type":"user","name":"Grace","login":"grace@example.com","status":"active"}]}`))
			case 2:
				_, _ = w.Write([]byte(`{"total_count":3,"offset":2,"limit":2,"entries":[{"id":"33","type":"user","name":"Kat","login":"kat@example.com","status":"inactive"}]}`))
			default:
				t.Errorf("unexpected offset=%d", offset)
				_, _ = w.Write([]byte(`{"total_count":3,"offset":4,"limit":2,"entries":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := box.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/2.0",
			"token_url": srv.URL + "/oauth2/token",
			"page_size": "2",
		},
		Secrets: map[string]string{
			"client_id":     "cid",
			"client_secret": "csecret",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalls == 0 {
		t.Fatal("expected an OAuth2 token exchange call")
	}
	if sawGrantType != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", sawGrantType)
	}
	if sawSubjectTyp != "enterprise" {
		t.Fatalf("box_subject_type = %q, want enterprise (default)", sawSubjectTyp)
	}
	if sawDataAuth != "Bearer tok_box_123" {
		t.Fatalf("data Authorization = %q, want Bearer tok_box_123", sawDataAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	if got[0]["id"] != "11" || got[0]["login"] != "ada@example.com" {
		t.Fatalf("first record mapped wrong: %+v", got[0])
	}
	if got[2]["id"] != "33" || got[2]["status"] != "inactive" {
		t.Fatalf("last record mapped wrong: %+v", got[2])
	}
}

// TestSubjectTypeUser confirms box_subject_type/box_subject_id are wired from
// config (user mode) onto the token request.
func TestSubjectTypeUser(t *testing.T) {
	var sawType, sawID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_ = r.ParseForm()
			sawType = r.Form.Get("box_subject_type")
			sawID = r.Form.Get("box_subject_id")
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		case "/2.0/groups":
			_, _ = w.Write([]byte(`{"total_count":0,"offset":0,"limit":100,"entries":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := box.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":         srv.URL + "/2.0",
			"token_url":        srv.URL + "/oauth2/token",
			"box_subject_type": "user",
			"user":             "987654",
		},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "groups", Config: cfg}, func(connectors.Record) error { return nil }); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawType != "user" {
		t.Fatalf("box_subject_type = %q, want user", sawType)
	}
	if sawID != "987654" {
		t.Fatalf("box_subject_id = %q, want 987654", sawID)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance runs credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := box.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "groups", "folder_items", "collections"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("Read(%s) fixture records = %d, want 2", stream, len(got))
		}
		if got[0]["id"] == nil {
			t.Fatalf("Read(%s) fixture record missing id: %+v", stream, got[0])
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := box.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := box.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "groups": false, "folder_items": false, "collections": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestBadBaseURLRejected(t *testing.T) {
	c := box.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"client_id": "cid", "client_secret": "csecret"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url")
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	c := box.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Check && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("box"); !ok {
		t.Fatal("registry did not resolve box (self-registration)")
	}
}
