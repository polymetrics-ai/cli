package strava_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/strava"
)

// TestReadActivitiesPaginatesAndAuthenticates is the red-first test for the
// Strava connector: it exchanges the refresh token for a bearer access token at
// the OAuth token endpoint, sends that bearer on the data request, paginates the
// top-level activities array across two pages via page/per_page, and maps
// records. Red until internal/connectors/strava exists.
func TestReadActivitiesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawTokenForm string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			_ = r.ParseForm()
			sawTokenForm = r.Form.Encode()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token_type":"Bearer","access_token":"access_abc","expires_in":21600}`))
		case "/api/v3/athlete/activities":
			sawAuth = r.Header.Get("Authorization")
			switch r.URL.Query().Get("page") {
			case "1":
				_, _ = w.Write([]byte(`[{"id":1001,"name":"Morning Run","start_date":"2026-01-01T07:00:00Z","distance":5000},{"id":1002,"name":"Evening Ride","start_date":"2026-01-02T18:00:00Z","distance":20000}]`))
			case "2":
				_, _ = w.Write([]byte(`[{"id":1003,"name":"Long Hike","start_date":"2026-01-03T09:00:00Z","distance":12000}]`))
			case "3":
				_, _ = w.Write([]byte(`[]`))
			default:
				t.Errorf("unexpected page=%q", r.URL.Query().Get("page"))
				_, _ = w.Write([]byte(`[]`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := strava.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL + "/api/v3",
			"token_url":  srv.URL + "/oauth/token",
			"client_id":  "12345",
			"athlete_id": "17831421",
			"page_size":  "2",
		},
		Secrets: map[string]string{
			"client_secret": "secret_fff",
			"refresh_token": "refresh_aaa",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "activities", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer access_abc" {
		t.Fatalf("Authorization = %q, want Bearer access_abc", sawAuth)
	}
	if !strings.Contains(sawTokenForm, "grant_type=refresh_token") {
		t.Fatalf("token form = %q, want grant_type=refresh_token", sawTokenForm)
	}
	if !strings.Contains(sawTokenForm, "refresh_token=refresh_aaa") {
		t.Fatalf("token form = %q, want refresh_token=refresh_aaa", sawTokenForm)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["start_date"] == nil {
			t.Fatalf("record missing id/start_date: %+v", rec)
		}
	}
}

// TestReadSingletonAthlete confirms that single-object streams (athlete) yield
// exactly one record without pagination.
func TestReadSingletonAthlete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			_, _ = w.Write([]byte(`{"token_type":"Bearer","access_token":"tok","expires_in":21600}`))
		case "/api/v3/athlete":
			if r.URL.Query().Get("page") != "" {
				t.Errorf("singleton stream should not paginate, got page=%q", r.URL.Query().Get("page"))
			}
			_, _ = w.Write([]byte(`{"id":17831421,"username":"runner","firstname":"Ada","lastname":"Lovelace"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := strava.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srv.URL + "/api/v3",
			"token_url":  srv.URL + "/oauth/token",
			"client_id":  "12345",
			"athlete_id": "17831421",
		},
		Secrets: map[string]string{"client_secret": "s", "refresh_token": "r"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "athlete", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["id"] == nil {
		t.Fatalf("athlete record missing id: %+v", got[0])
	}
}

// TestFixtureModeNeedsNoNetwork confirms fixture mode emits deterministic
// records without any credentials or network access (conformance support).
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := strava.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"activities", "athlete", "athlete_stats", "clubs"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture stream %s emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureModeNoNetwork(t *testing.T) {
	c := strava.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := strava.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "strava" {
		t.Fatalf("catalog connector = %q, want strava", cat.Connector)
	}
	want := map[string]bool{"activities": false, "athlete": false, "athlete_stats": false, "clubs": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
		if len(s.Fields) == 0 {
			t.Fatalf("stream %q has no fields", s.Name)
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestReadOnlyCapabilities(t *testing.T) {
	caps := strava.New().Metadata().Capabilities
	if !caps.Read || !caps.Check || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read+Check+Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("strava is read-only; Write should be false")
	}
}

func TestRegistryResolvesStrava(t *testing.T) {
	_ = strava.New() // ensure init() ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("strava"); !ok {
		t.Fatal("registry did not resolve strava (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := strava.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"client_secret": "s", "refresh_token": "r"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "athlete", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for non-http(s) base_url scheme")
	}
}
