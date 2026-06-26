package klausapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	klausapi "polymetrics.ai/internal/connectors/klaus-api"
)

// TestReadUsersAuthenticates exercises Bearer auth, the account-scoped users
// path, and record mapping from the "users" envelope.
func TestReadUsersAuthenticates(t *testing.T) {
	var sawAuth, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"users":[{"id":"u_1","name":"Ada","email":"ada@example.com"},{"id":"u_2","name":"Grace","email":"grace@example.com"}]}`))
	}))
	defer srv.Close()

	c := klausapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account": "42", "workspace": "7"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer key_abc" {
		t.Fatalf("Authorization = %q, want Bearer key_abc", sawAuth)
	}
	if sawPath != "/account/42/users" {
		t.Fatalf("path = %q, want /account/42/users", sawPath)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "u_1" || got[0]["email"] != "ada@example.com" {
		t.Fatalf("unexpected first record: %+v", got[0])
	}
}

// TestReadReviewsPaginatesByDateWindow asserts the reviews stream walks forward
// in date windows (fromDate/toDate) across more than one request, extracting the
// "conversations" array each time.
func TestReadReviewsPaginatesByDateWindow(t *testing.T) {
	var windows []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/account/42/workspace/7/reviews" {
			http.NotFound(w, r)
			return
		}
		from := r.URL.Query().Get("fromDate")
		windows = append(windows, from)
		switch from {
		case "2026-01-01T00:00:00Z":
			_, _ = w.Write([]byte(`{"conversations":[{"id":"c_1","lastUpdatedISO":"2026-01-02T00:00:00Z"}]}`))
		case "2026-01-08T00:00:00Z":
			_, _ = w.Write([]byte(`{"conversations":[{"id":"c_2","lastUpdatedISO":"2026-01-09T00:00:00Z"}]}`))
		default:
			_, _ = w.Write([]byte(`{"conversations":[]}`))
		}
	}))
	defer srv.Close()

	c := klausapi.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL,
			"account":    "42",
			"workspace":  "7",
			"start_date": "2026-01-01T00:00:00Z",
			"end_date":   "2026-01-15T00:00:00Z",
		},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "reviews", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (two date windows)", len(got))
	}
	if len(windows) < 2 {
		t.Fatalf("expected at least 2 windowed requests, got %v", windows)
	}
	if windows[0] != "2026-01-01T00:00:00Z" || windows[1] != "2026-01-08T00:00:00Z" {
		t.Fatalf("unexpected window progression: %v", windows)
	}
}

// TestReadCategories exercises the workspace-scoped categories path.
func TestReadCategories(t *testing.T) {
	var sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		_, _ = w.Write([]byte(`{"categories":[{"id":"cat_1","name":"Tone","weight":1.5,"critical":false}]}`))
	}))
	defer srv.Close()

	c := klausapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "account": "42", "workspace": "7"},
		Secrets: map[string]string{"api_key": "key_abc"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "categories", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawPath != "/account/42/workspace/7/categories" {
		t.Fatalf("path = %q, want /account/42/workspace/7/categories", sawPath)
	}
	if len(got) != 1 || got[0]["name"] != "Tone" {
		t.Fatalf("unexpected categories result: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access or credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := klausapi.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "reviews", "categories"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			n++
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture %s emitted no records", stream)
		}
	}
}

// TestCheckFixtureNoCreds verifies Check short-circuits in fixture mode.
func TestCheckFixtureNoCreds(t *testing.T) {
	c := klausapi.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestRegisteredReadOnly asserts self-registration and the read-only capability
// set, and that the registry resolves the connector by its bare system name.
func TestRegisteredReadOnly(t *testing.T) {
	_ = klausapi.New() // ensure init ran
	c := klausapi.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("klaus-api"); !ok {
		t.Fatal("registry did not resolve klaus-api (self-registration)")
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) base_url overrides.
func TestBaseURLSSRFValidation(t *testing.T) {
	c := klausapi.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "account": "1", "workspace": "1"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http base_url, got nil")
	}
}
