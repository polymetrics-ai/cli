package openweather_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/openweather"
)

// onecallBody is a trimmed One Call API 3.0 payload: one current object, two
// hourly entries, two daily entries, and one alert. dt fields are the cursor.
func onecallBody(curDT, h1, h2 int64) string {
	return `{
		"lat": 33.44, "lon": -94.04, "timezone": "America/Chicago", "timezone_offset": -18000,
		"current": {"dt": ` + itoa(curDT) + `, "temp": 292.55, "feels_like": 292.87, "pressure": 1014, "humidity": 89, "weather": [{"id": 803, "main": "Clouds", "description": "broken clouds", "icon": "04d"}]},
		"hourly": [
			{"dt": ` + itoa(h1) + `, "temp": 292.01, "feels_like": 292.33, "pressure": 1014, "humidity": 91, "pop": 0.15, "weather": [{"id": 500, "main": "Rain", "description": "light rain", "icon": "10d"}]},
			{"dt": ` + itoa(h2) + `, "temp": 292.84, "feels_like": 293.04, "pressure": 1014, "humidity": 88, "pop": 0.2, "weather": [{"id": 500, "main": "Rain", "description": "light rain", "icon": "10d"}]}
		],
		"daily": [
			{"dt": ` + itoa(h1) + `, "summary": "Expect rain", "pressure": 1017, "humidity": 69, "temp": {"day": 299.03, "min": 290.69, "max": 300.35}, "weather": [{"id": 500, "main": "Rain", "description": "light rain", "icon": "10d"}]},
			{"dt": ` + itoa(h2) + `, "summary": "Cloudy", "pressure": 1015, "humidity": 71, "temp": {"day": 298.0, "min": 289.0, "max": 299.0}, "weather": [{"id": 803, "main": "Clouds", "description": "broken clouds", "icon": "04d"}]}
		],
		"alerts": [
			{"sender_name": "NWS", "event": "Heat Advisory", "start": ` + itoa(h1) + `, "end": ` + itoa(h2) + `, "description": "stay cool", "tags": ["Extreme temperature value"]}
		]
	}`
}

func itoa(v int64) string {
	// small local helper to keep the literal readable
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v%10]
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// TestReadAcrossLocationsAuthenticatesAndMaps drives the openweather connector
// against a fake One Call API. It asserts: appid is sent as a query parameter
// (not a header), lat/lon vary per configured location (two "pages" -> two
// requests), and hourly records map dt+temp. Red until the package exists.
func TestReadAcrossLocationsAuthenticatesAndMaps(t *testing.T) {
	var calls []struct{ appid, lat, lon string }
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/onecall" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		calls = append(calls, struct{ appid, lat, lon string }{q.Get("appid"), q.Get("lat"), q.Get("lon")})
		_, _ = w.Write([]byte(onecallBody(1700000000, 1700003600, 1700007200)))
	}))
	defer srv.Close()

	c := openweather.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL,
			"locations": "33.44,-94.04;40.71,-74.01",
			"units":     "metric",
		},
		Secrets: map[string]string{"appid": "secret_key_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "hourly", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	// Two locations -> two requests (the connector's pagination loop).
	if len(calls) != 2 {
		t.Fatalf("requests = %d, want 2 (one per location)", len(calls))
	}
	for _, call := range calls {
		if call.appid != "secret_key_123" {
			t.Fatalf("appid query = %q, want secret_key_123", call.appid)
		}
	}
	if calls[0].lat != "33.44" || calls[0].lon != "-94.04" {
		t.Fatalf("first call lat/lon = %q/%q, want 33.44/-94.04", calls[0].lat, calls[0].lon)
	}
	if calls[1].lat != "40.71" || calls[1].lon != "-74.01" {
		t.Fatalf("second call lat/lon = %q/%q, want 40.71/-74.01", calls[1].lat, calls[1].lon)
	}

	// Each location returns 2 hourly entries -> 4 records total.
	if len(got) != 4 {
		t.Fatalf("hourly records = %d, want 4 (2 per location)", len(got))
	}
	for _, rec := range got {
		if rec["dt"] == nil || rec["temp"] == nil {
			t.Fatalf("record missing dt/temp: %+v", rec)
		}
		if rec["lat"] == nil || rec["lon"] == nil {
			t.Fatalf("record missing injected lat/lon: %+v", rec)
		}
	}
}

// TestReadCurrentSingleRecord verifies the single-object "current" stream maps
// into exactly one record per location.
func TestReadCurrentSingleRecord(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(onecallBody(1700000000, 1700003600, 1700007200)))
	}))
	defer srv.Close()

	c := openweather.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "lat": "33.44", "lon": "-94.04"},
		Secrets: map[string]string{"appid": "k"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "current", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("current records = %d, want 1", len(got))
	}
	if got[0]["temp"] == nil || got[0]["humidity"] == nil {
		t.Fatalf("current record missing fields: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records with
// no HTTP server configured (conformance without creds).
func TestFixtureModeNoNetwork(t *testing.T) {
	c := openweather.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"current", "hourly", "daily", "alerts"} {
		var n int
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			n++
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if n == 0 {
			t.Fatalf("fixture Read(%s) emitted 0 records", stream)
		}
	}
}

// TestCheckFixtureMode short-circuits without network.
func TestCheckFixtureMode(t *testing.T) {
	c := openweather.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestMetadataReadOnly asserts the connector is read-only.
func TestMetadataReadOnly(t *testing.T) {
	caps := openweather.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false", caps)
	}
}

// TestRegistryResolves confirms self-registration via init().
func TestRegistryResolves(t *testing.T) {
	_ = openweather.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("openweather"); !ok {
		t.Fatal("registry did not resolve openweather (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards SSRF validation.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := openweather.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "file:///etc/passwd", "lat": "1", "lon": "2"},
		Secrets: map[string]string{"appid": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "current", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with file:// base_url should fail SSRF validation")
	}
}
