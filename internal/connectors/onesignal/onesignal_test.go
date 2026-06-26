package onesignal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/onesignal"
)

// TestReadDevicesPaginatesAndAuthenticates is the red-first test: the devices
// (players) stream authenticates with the per-app REST API key via
// Authorization: Basic, paginates over offset until total_count is exhausted,
// and maps records. Red until internal/connectors/onesignal exists.
func TestReadDevicesPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawAppID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawAppID = r.URL.Query().Get("app_id")
		if r.URL.Path != "/players" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("offset") {
		case "", "0":
			_, _ = w.Write([]byte(`{"total_count":3,"offset":0,"limit":2,"players":[{"id":"dev_1","identifier":"tok1","created_at":1700000000},{"id":"dev_2","identifier":"tok2","created_at":1700000100}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_count":3,"offset":2,"limit":2,"players":[{"id":"dev_3","identifier":"tok3","created_at":1700000200}]}`))
		default:
			t.Errorf("unexpected offset=%q", r.URL.Query().Get("offset"))
			_, _ = w.Write([]byte(`{"total_count":3,"offset":4,"limit":2,"players":[]}`))
		}
	}))
	defer srv.Close()

	c := onesignal.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"app_id":    "app_123",
			"page_size": "2",
		},
		Secrets: map[string]string{
			"app_api_key":   "rest_api_key_xyz",
			"user_auth_key": "user_auth_abc",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "devices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Basic rest_api_key_xyz" {
		t.Fatalf("Authorization = %q, want Basic rest_api_key_xyz", sawAuth)
	}
	if sawAppID != "app_123" {
		t.Fatalf("app_id = %q, want app_123", sawAppID)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
}

// TestReadAppsUsesUserAuthKey verifies the apps stream is an account-level
// endpoint authenticated with the user_auth_key (not the app key) and that the
// top-level JSON array is mapped.
func TestReadAppsUsesUserAuthKey(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/apps" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"app_1","name":"App One","players":10},{"id":"app_2","name":"App Two","players":20}]`))
	}))
	defer srv.Close()

	c := onesignal.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{
			"app_api_key":   "rest_api_key_xyz",
			"user_auth_key": "user_auth_abc",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "apps", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Basic user_auth_abc" {
		t.Fatalf("Authorization = %q, want Basic user_auth_abc", sawAuth)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] != "app_1" || got[1]["name"] != "App Two" {
		t.Fatalf("apps not mapped: %+v", got)
	}
}

// TestFixtureModeNoNetwork confirms fixture mode emits deterministic records
// without any network access, so conformance runs without live creds.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := onesignal.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"apps", "devices", "notifications", "outcomes"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read(%s) fixture: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("Read(%s) fixture emitted no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

// TestCatalogAndMetadata verifies the published catalog and read-only caps.
func TestCatalogAndMetadata(t *testing.T) {
	c := onesignal.New()
	meta := c.Metadata()
	if !meta.Capabilities.Read || !meta.Capabilities.Catalog || meta.Capabilities.Write {
		t.Fatalf("capabilities = %+v, want Read+Catalog read-only", meta.Capabilities)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("streams = %d, want >= 3", len(cat.Streams))
	}
	want := map[string]bool{"apps": false, "devices": false, "notifications": false}
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

// TestRegistryResolves verifies self-registration via init().
func TestRegistryResolves(t *testing.T) {
	_ = onesignal.New() // ensure package init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("onesignal"); !ok {
		t.Fatal("registry did not resolve onesignal (self-registration)")
	}
}
