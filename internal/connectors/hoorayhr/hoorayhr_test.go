package hoorayhr_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/hoorayhr"
)

// TestReadAuthenticatesAndMapsRecords is the red-first test for the HoorayHR
// connector. HoorayHR uses a SessionTokenAuthenticator: it POSTs credentials to
// /authentication, reads accessToken from the JSON response, and injects it into
// the Authorization header (raw token, no Bearer prefix) on subsequent requests.
// Stream responses are top-level JSON arrays with no pagination.
func TestReadAuthenticatesAndMapsRecords(t *testing.T) {
	var (
		loginBody map[string]any
		sawAuth   string
		loginHits int
		dataHits  int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/authentication":
			loginHits++
			if r.Method != http.MethodPost {
				t.Errorf("login method = %q, want POST", r.Method)
			}
			_ = json.NewDecoder(r.Body).Decode(&loginBody)
			_, _ = w.Write([]byte(`{"accessToken":"tok_abc","user":{"id":1}}`))
		case "/sick-leave":
			dataHits++
			sawAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`[{"id":11,"userId":7,"status":"open","notes":"flu"},{"id":12,"userId":8,"status":"closed","notes":"cold"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := hoorayhr.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":         srv.URL,
			"hoorayhrusername": "agent@example.com",
		},
		Secrets: map[string]string{"hoorayhrpassword": "s3cret"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "sick-leaves", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if loginHits != 1 {
		t.Fatalf("login hits = %d, want 1", loginHits)
	}
	if dataHits != 1 {
		t.Fatalf("data hits = %d, want 1", dataHits)
	}
	if loginBody["email"] != "agent@example.com" || loginBody["password"] != "s3cret" || loginBody["strategy"] != "local" {
		t.Fatalf("login body = %+v, want email/password/strategy=local", loginBody)
	}
	if sawAuth != "tok_abc" {
		t.Fatalf("Authorization = %q, want raw token tok_abc", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["status"] != "open" || got[0]["notes"] != "flu" {
		t.Fatalf("record[0] mapped wrong: %+v", got[0])
	}
}

// TestReadMultipleStreamsHitCorrectEndpoints confirms the stream->endpoint
// routing table maps the catalog stream names to the API resource paths
// (sick-leaves -> /sick-leave, time-off -> /time-off, etc.).
func TestReadMultipleStreamsHitCorrectEndpoints(t *testing.T) {
	cases := map[string]string{
		"users":       "/users",
		"time-off":    "/time-off",
		"leave-types": "/leave-types",
		"sick-leaves": "/sick-leave",
	}
	for stream, wantPath := range cases {
		stream, wantPath := stream, wantPath
		t.Run(stream, func(t *testing.T) {
			var hitPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/authentication" {
					_, _ = w.Write([]byte(`{"accessToken":"tok"}`))
					return
				}
				hitPath = r.URL.Path
				_, _ = w.Write([]byte(`[{"id":1}]`))
			}))
			defer srv.Close()

			c := hoorayhr.New()
			cfg := connectors.RuntimeConfig{
				Config:  map[string]string{"base_url": srv.URL, "hoorayhrusername": "u"},
				Secrets: map[string]string{"hoorayhrpassword": "p"},
			}
			var n int
			if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error {
				n++
				return nil
			}); err != nil {
				t.Fatalf("Read(%s): %v", stream, err)
			}
			if hitPath != wantPath {
				t.Fatalf("stream %s hit %q, want %q", stream, hitPath, wantPath)
			}
			if n != 1 {
				t.Fatalf("stream %s records = %d, want 1", stream, n)
			}
		})
	}
}

// TestFixtureModeNeedsNoNetwork ensures conformance can run credential-free.
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := hoorayhr.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	for _, stream := range []string{"users", "time-off", "leave-types", "sick-leaves"} {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("Read(fixture %s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCatalogListsCoreStreams(t *testing.T) {
	c := hoorayhr.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "time-off": false, "leave-types": false, "sick-leaves": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != "id" {
			t.Fatalf("stream %s primary key = %v, want [id]", s.Name, s.PrimaryKey)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := hoorayhr.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com", "hoorayhrusername": "u"},
		Secrets: map[string]string{"hoorayhrpassword": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Read with bad scheme err = %v, want base_url error", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = hoorayhr.New() // ensure init ran
	caps := hoorayhr.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only source)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("hoorayhr"); !ok {
		t.Fatal("registry did not resolve hoorayhr (self-registration)")
	}
}
