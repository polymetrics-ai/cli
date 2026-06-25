package breezometer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/breezometer"
)

// TestReadPaginatesAndAuthenticates is the red-first test for the BreezoMeter
// connector: API-key query-param auth, next_page_token pagination over the
// data[] array, and record mapping. Red until internal/connectors/breezometer
// exists.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawKey string
	var sawLat, sawLng string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.URL.Query().Get("key")
		sawLat = r.URL.Query().Get("lat")
		sawLng = r.URL.Query().Get("lon")
		if !strings.HasPrefix(r.URL.Path, "/air-quality/v2/forecast/hourly") {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("page_token") {
		case "":
			_, _ = w.Write([]byte(`{"metadata":{"x":1},"data":[` +
				`{"datetime":"2026-01-01T00:00:00Z","indexes":{"baqi":{"aqi":42}}},` +
				`{"datetime":"2026-01-01T01:00:00Z","indexes":{"baqi":{"aqi":43}}}` +
				`],"next_page_token":"pg2"}`))
		case "pg2":
			_, _ = w.Write([]byte(`{"metadata":{"x":2},"data":[` +
				`{"datetime":"2026-01-01T02:00:00Z","indexes":{"baqi":{"aqi":44}}}` +
				`]}`))
		default:
			t.Errorf("unexpected page_token=%q", r.URL.Query().Get("page_token"))
			_, _ = w.Write([]byte(`{"data":[]}`))
		}
	}))
	defer srv.Close()

	c := breezometer.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"latitude":  "54.675003",
			"longitude": "-113.550282",
		},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "air_quality_forecast", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawKey != "key_test_123" {
		t.Fatalf("key query param = %q, want key_test_123", sawKey)
	}
	if sawLat != "54.675003" || sawLng != "-113.550282" {
		t.Fatalf("lat/lon = %q/%q, want 54.675003/-113.550282", sawLat, sawLng)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["datetime"] == nil {
			t.Fatalf("record missing datetime: %+v", rec)
		}
		if rec["latitude"] != "54.675003" {
			t.Fatalf("record missing injected latitude: %+v", rec)
		}
	}
}

// TestCurrentConditionsSingleObject verifies the single-object (non-list)
// streams emit exactly one record mapped from the data object.
func TestCurrentConditionsSingleObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/air-quality/v2/current-conditions") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"datetime":"2026-01-01T00:00:00Z","indexes":{"baqi":{"aqi":50,"category":"Good air quality"}}}}`))
	}))
	defer srv.Close()

	c := breezometer.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"latitude":  "1.0",
			"longitude": "2.0",
		},
		Secrets: map[string]string{"api_key": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "air_quality_current", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
	if got[0]["datetime"] == nil {
		t.Fatalf("record missing datetime: %+v", got[0])
	}
}

// TestFixtureMode verifies the credential-free fixture path emits deterministic
// records without network access (required for conformance).
func TestFixtureMode(t *testing.T) {
	c := breezometer.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"air_quality_current", "air_quality_forecast", "pollen_forecast", "weather_current"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("Read fixture %s: %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture %s emitted no records", stream)
		}
	}
	// Check must also short-circuit in fixture mode with no creds.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestCheckRequiresCreds(t *testing.T) {
	c := breezometer.New()
	err := c.Check(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{"latitude": "1", "longitude": "2"},
	})
	if err == nil {
		t.Fatal("Check without api_key should fail")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := breezometer.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 3 {
		t.Fatalf("want at least 3 streams, got %d", len(cat.Streams))
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = breezometer.New() // ensure init ran
	caps := breezometer.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("breezometer is read-only, Write should be false")
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("breezometer"); !ok {
		t.Fatal("registry did not resolve breezometer (self-registration)")
	}
}
