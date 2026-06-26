package googleforms_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	googleforms "polymetrics.ai/internal/connectors/google-forms"
)

// TestReadResponsesPaginatesAndAuthenticates is the red-first test: it asserts
// the OAuth2 refresh-token exchange happens, the Bearer access token is attached
// to the Forms API request, the responses[] stream is mapped, and pagination
// across two pages via nextPageToken/pageToken works.
func TestReadResponsesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawRefreshGrant string

	// OAuth2 token endpoint: exchanges the refresh token for an access token.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse token form: %v", err)
		}
		sawRefreshGrant = r.Form.Get("grant_type")
		if r.Form.Get("refresh_token") != "rt_test_123" {
			t.Errorf("refresh_token = %q, want rt_test_123", r.Form.Get("refresh_token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"ya29.access_test","token_type":"Bearer","expires_in":3599}`))
	}))
	defer tokenSrv.Close()

	// Forms API endpoint: serves two pages of responses for form abc123.
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/forms/abc123/responses" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"responses":[{"responseId":"r1","createTime":"2026-01-01T00:00:00Z","respondentEmail":"a@example.com"},{"responseId":"r2","createTime":"2026-01-02T00:00:00Z"}],"nextPageToken":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"responses":[{"responseId":"r3","createTime":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"responses":[]}`))
		}
	}))
	defer apiSrv.Close()

	c := googleforms.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  apiSrv.URL,
			"token_url": tokenSrv.URL,
			"form_id":   "abc123",
		},
		Secrets: map[string]string{
			"client_id":            "cid",
			"client_secret":        "csecret",
			"client_refresh_token": "rt_test_123",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "responses", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawRefreshGrant != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawRefreshGrant)
	}
	if sawAuth != "Bearer ya29.access_test" {
		t.Fatalf("Authorization = %q, want Bearer ya29.access_test", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["response_id"] == nil {
			t.Fatalf("record missing response_id: %+v", rec)
		}
		if rec["form_id"] != "abc123" {
			t.Fatalf("record form_id = %v, want abc123", rec["form_id"])
		}
	}
}

// TestReadFormsMapping verifies the forms stream fetches form metadata per
// configured form_id and maps it to a record keyed by form_id.
func TestReadFormsMapping(t *testing.T) {
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forms/abc123" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"formId":"abc123","revisionId":"00000001","responderUri":"https://docs.google.com/forms/d/e/abc/viewform","info":{"title":"My Form","documentTitle":"My Form Doc"},"items":[{"itemId":"i1","title":"Q1"}]}`))
	}))
	defer apiSrv.Close()
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3599}`))
	}))
	defer tokenSrv.Close()

	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": apiSrv.URL, "token_url": tokenSrv.URL, "form_id": "abc123"},
		Secrets: map[string]string{
			"client_id":            "cid",
			"client_secret":        "csecret",
			"client_refresh_token": "rt",
		},
	}
	var got []connectors.Record
	err := googleforms.New().Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read forms: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("forms records = %d, want 1", len(got))
	}
	if got[0]["form_id"] != "abc123" || got[0]["title"] != "My Form" {
		t.Fatalf("form record mapping wrong: %+v", got[0])
	}
}

// TestFixtureMode verifies fixture mode emits deterministic records with no
// network access, so conformance can run without credentials.
func TestFixtureMode(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"forms", "responses", "form_items"} {
		var got []connectors.Record
		err := googleforms.New().Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
}

// TestCheckFixtureNoNetwork verifies Check short-circuits in fixture mode.
func TestCheckFixtureNoNetwork(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := googleforms.New().Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams verifies the published catalog exposes the core streams.
func TestCatalogStreams(t *testing.T) {
	cat, err := googleforms.New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"forms": false, "responses": false, "form_items": false}
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
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestBaseURLSSRFValidation rejects non-http(s) base URLs.
func TestBaseURLSSRFValidation(t *testing.T) {
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": "file:///etc/passwd", "form_id": "x"},
		Secrets: map[string]string{
			"client_id":            "cid",
			"client_secret":        "csecret",
			"client_refresh_token": "rt",
		},
	}
	err := googleforms.New().Read(context.Background(), connectors.ReadRequest{Stream: "forms", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}

// TestRegistryResolves verifies self-registration via the connectors registry.
func TestRegistryResolves(t *testing.T) {
	_ = googleforms.New() // ensure init ran
	r := connectors.NewRegistry()
	got, ok := r.Get("google-forms")
	if !ok {
		t.Fatal("registry did not resolve google-forms (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("google-forms should be read-only, got Write=true")
	}
}

// TestTokenResponseShape is a guard that the token JSON we expect decodes.
func TestTokenResponseShape(t *testing.T) {
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal([]byte(`{"access_token":"x"}`), &out); err != nil {
		t.Fatal(err)
	}
	if out.AccessToken != "x" {
		t.Fatal("decode")
	}
}
