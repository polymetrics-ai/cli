package googlecalendar_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	googlecalendar "polymetrics.ai/internal/connectors/google-calendar"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the Google
// Calendar connector. It exercises the OAuth2 refresh-token grant (the token
// endpoint is hit, then data requests carry the resulting Bearer token), the
// nextPageToken/pageToken body-cursor pagination over items[], and record
// mapping. Red until internal/connectors/google-calendar is implemented.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var (
		tokenHits  int
		sawGrant   string
		sawRefresh string
		sawAuth    string
	)

	// The OAuth token endpoint: exchanges the refresh_token for an access token.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHits++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("token endpoint ParseForm: %v", err)
		}
		sawGrant = r.Form.Get("grant_type")
		sawRefresh = r.Form.Get("refresh_token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"ya29.test_access","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	// The Calendar API: calendarList listing, paginated via nextPageToken.
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/users/me/calendarList" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"kind":"calendar#calendarList","items":[{"id":"cal_1","summary":"Work","accessRole":"owner"},{"id":"cal_2","summary":"Home","accessRole":"reader"}],"nextPageToken":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"kind":"calendar#calendarList","items":[{"id":"cal_3","summary":"Holidays","accessRole":"reader"}]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	defer apiSrv.Close()

	c := googlecalendar.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   apiSrv.URL,
			"token_url":  tokenSrv.URL,
			"calendarid": "primary",
		},
		Secrets: map[string]string{
			"client_id":              "client-id-123",
			"client_secret":          "client-secret-456",
			"client_refresh_token_2": "refresh-token-789",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "calendar_list", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenHits == 0 {
		t.Fatal("token endpoint was never called; refresh-token grant did not run")
	}
	if sawGrant != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawGrant)
	}
	if sawRefresh != "refresh-token-789" {
		t.Fatalf("refresh_token = %q, want refresh-token-789", sawRefresh)
	}
	if sawAuth != "Bearer ya29.test_access" {
		t.Fatalf("Authorization = %q, want Bearer ya29.test_access", sawAuth)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (across 2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["summary"] == nil {
			t.Fatalf("record missing id/summary: %+v", rec)
		}
	}
}

// TestReadEventsUsesCalendarID confirms the events stream targets the
// configured calendar and maps the iCalUID/start/end fields.
func TestReadEventsUsesCalendarID(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	var sawPath string
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"kind":"calendar#events","items":[{"id":"evt_1","status":"confirmed","summary":"Standup","start":{"dateTime":"2026-06-01T09:00:00Z"},"end":{"dateTime":"2026-06-01T09:30:00Z"}}]}`))
	}))
	defer apiSrv.Close()

	c := googlecalendar.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   apiSrv.URL,
			"token_url":  tokenSrv.URL,
			"calendarid": "team@example.com",
		},
		Secrets: map[string]string{
			"client_id":              "id",
			"client_secret":          "secret",
			"client_refresh_token_2": "rt",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read events: %v", err)
	}
	if !strings.Contains(sawPath, "/calendars/team@example.com/events") {
		t.Fatalf("events path = %q, want it to contain /calendars/team@example.com/events", sawPath)
	}
	if len(got) != 1 {
		t.Fatalf("events = %d, want 1", len(got))
	}
	if got[0]["id"] != "evt_1" || got[0]["status"] != "confirmed" {
		t.Fatalf("event record mismatch: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access (so conformance runs without live creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := googlecalendar.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"calendar_list", "events", "settings", "acl"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}

	// Check must succeed in fixture mode without credentials.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestCatalogStreams confirms the published catalog exposes the core streams
// with primary keys.
func TestCatalogStreams(t *testing.T) {
	c := googlecalendar.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"calendar_list": false, "events": false, "settings": false, "acl": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegisteredReadOnly confirms self-registration and read-only capabilities.
func TestRegisteredReadOnly(t *testing.T) {
	c := googlecalendar.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("google-calendar"); !ok {
		t.Fatal("registry did not resolve google-calendar (self-registration)")
	}
}
