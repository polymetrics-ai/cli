package youtubeanalytics_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	youtubeanalytics "polymetrics/internal/connectors/youtube-analytics"
)

// liveConfig builds a RuntimeConfig pointed at an httptest server, with the
// OAuth token endpoint overridden so the refresh-token exchange stays local.
func liveConfig(baseURL, tokenURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  baseURL,
			"token_url": tokenURL,
		},
		Secrets: map[string]string{
			"client_id":     "client-123",
			"client_secret": "secret-xyz",
			"refresh_token": "refresh-abc",
		},
	}
}

// TestReadJobsAuthenticatesAndPaginates is the red-first test: it asserts the
// connector exchanges the refresh token for an access token, sends it as a
// Bearer header to the Reporting API, follows pageToken/nextPageToken cursor
// pagination across two pages of jobs[], and maps records.
func TestReadJobsAuthenticatesAndPaginates(t *testing.T) {
	var (
		sawAuth       string
		tokenForm     string
		tokenRequests int
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		_ = r.ParseForm()
		tokenForm = r.Form.Encode()
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_TOKEN_1","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		switch r.URL.Query().Get("pageToken") {
		case "":
			_, _ = w.Write([]byte(`{"jobs":[{"id":"job_1","reportTypeId":"channel_basic_a3","name":"Job One","createTime":"2026-01-01T00:00:00Z"},{"id":"job_2","reportTypeId":"channel_demographics_a1","name":"Job Two","createTime":"2026-01-02T00:00:00Z"}],"nextPageToken":"PAGE2"}`))
		case "PAGE2":
			_, _ = w.Write([]byte(`{"jobs":[{"id":"job_3","reportTypeId":"playlist_basic_a2","name":"Job Three","createTime":"2026-01-03T00:00:00Z"}]}`))
		default:
			t.Errorf("unexpected pageToken=%q", r.URL.Query().Get("pageToken"))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := youtubeanalytics.New()
	cfg := liveConfig(srv.URL+"/v1", srv.URL+"/oauth/token")

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ACCESS_TOKEN_1" {
		t.Fatalf("Authorization = %q, want Bearer ACCESS_TOKEN_1", sawAuth)
	}
	if tokenRequests == 0 {
		t.Fatal("expected a refresh-token exchange against the token endpoint")
	}
	if !strings.Contains(tokenForm, "grant_type=refresh_token") || !strings.Contains(tokenForm, "refresh_token=refresh-abc") {
		t.Fatalf("token form = %q, want refresh_token grant", tokenForm)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	if got[0]["id"] != "job_1" || got[0]["report_type_id"] != "channel_basic_a3" {
		t.Fatalf("record[0] = %+v, want mapped job_1/channel_basic_a3", got[0])
	}
	if got[2]["id"] != "job_3" {
		t.Fatalf("record[2] id = %v, want job_3", got[2]["id"])
	}
}

// TestReadContentOwnerScope asserts the optional content_owner_id config is sent
// as the onBehalfOfContentOwner query parameter the Reporting API expects.
func TestReadContentOwnerScope(t *testing.T) {
	var sawOwner string
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/reportTypes", func(w http.ResponseWriter, r *http.Request) {
		sawOwner = r.URL.Query().Get("onBehalfOfContentOwner")
		_, _ = w.Write([]byte(`{"reportTypes":[{"id":"channel_basic_a3","name":"User Activity"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := liveConfig(srv.URL+"/v1", srv.URL+"/oauth/token")
	cfg.Config["content_owner_id"] = "owner_42"

	c := youtubeanalytics.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "report_types", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawOwner != "owner_42" {
		t.Fatalf("onBehalfOfContentOwner = %q, want owner_42", sawOwner)
	}
	if len(got) != 1 || got[0]["id"] != "channel_basic_a3" {
		t.Fatalf("records = %+v, want one report_type channel_basic_a3", got)
	}
}

// TestFixtureModeNoNetwork confirms credential-free conformance: fixture mode
// emits deterministic records without any network call.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := youtubeanalytics.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "jobs", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("fixture record missing id: %+v", rec)
		}
	}
}

func TestCatalogStreams(t *testing.T) {
	c := youtubeanalytics.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want at least 3", len(cat.Streams))
	}
	want := map[string]bool{"jobs": false, "report_types": false, "reports": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

func TestRegistryResolution(t *testing.T) {
	_ = youtubeanalytics.New() // ensure init ran
	caps := youtubeanalytics.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities Write = true, want read-only connector")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("youtube-analytics"); !ok {
		t.Fatal("registry did not resolve youtube-analytics (self-registration)")
	}
}
