package calcom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	calcom "polymetrics.ai/internal/connectors/cal-com"
)

// TestReadPaginatesAndAuthenticates is the red-first test: Bearer auth from the
// api_key secret, Cal.com offset (skip/take) pagination over data[], and record
// mapping for the bookings stream.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	var sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawVersion = r.Header.Get("cal-api-version")
		if r.URL.Path != "/v2/bookings" {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Query().Get("skip") {
		case "", "0":
			// A full page (take=2) signals there may be more.
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":1,"uid":"bk_1","title":"Intro"},{"id":2,"uid":"bk_2","title":"Demo"}]}`))
		case "2":
			// A short page ends pagination.
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":3,"uid":"bk_3","title":"Review"}]}`))
		default:
			t.Errorf("unexpected skip=%q", r.URL.Query().Get("skip"))
			_, _ = w.Write([]byte(`{"status":"success","data":[]}`))
		}
	}))
	defer srv.Close()

	c := calcom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "page_size": "2"},
		Secrets: map[string]string{"api_key": "cal_live_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer cal_live_123" {
		t.Fatalf("Authorization = %q, want Bearer cal_live_123", sawAuth)
	}
	if sawVersion == "" {
		t.Fatalf("expected cal-api-version header to be set")
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["uid"] == nil {
			t.Fatalf("record missing id/uid: %+v", rec)
		}
	}
}

// TestReadEventTypesFlattensNestedGroups verifies the nested event-types
// selector (data.eventTypeGroups[].eventTypes[]) is flattened correctly.
func TestReadEventTypesFlattensNestedGroups(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/event-types" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"eventTypeGroups":[{"eventTypes":[{"id":10,"slug":"15min","title":"15 Minute"},{"id":11,"slug":"30min","title":"30 Minute"}]},{"eventTypes":[{"id":12,"slug":"60min","title":"60 Minute"}]}]}}`))
	}))
	defer srv.Close()

	c := calcom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "cal_live_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "event_types", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("event_types records = %d, want 3", len(got))
	}
	if got[0]["slug"] != "15min" {
		t.Fatalf("first event_type slug = %v, want 15min", got[0]["slug"])
	}
}

// TestReadSingleObjectStream covers my_profile (records at data as a single
// object, no pagination).
func TestReadSingleObjectStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/me" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"id":99,"username":"ada","email":"ada@example.com"}}`))
	}))
	defer srv.Close()

	c := calcom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "cal_live_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "my_profile", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("my_profile records = %d, want 1", len(got))
	}
	if got[0]["username"] != "ada" {
		t.Fatalf("username = %v, want ada", got[0]["username"])
	}
}

// TestFixtureModeNoNetwork ensures fixture mode emits deterministic records
// without any network access, so conformance can run credential-free.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := calcom.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"bookings", "event_types", "schedules", "my_profile"} {
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
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := calcom.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams asserts the published catalog has the core streams.
func TestCatalogStreams(t *testing.T) {
	c := calcom.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"bookings": false, "event_types": false, "schedules": false, "my_profile": false}
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

// TestRegisteredReadOnly verifies self-registration and read-only capability.
func TestRegisteredReadOnly(t *testing.T) {
	_ = calcom.New() // ensure init ran
	c := calcom.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || caps.Write {
		t.Fatalf("capabilities = %+v, want Read && !Write", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("cal-com"); !ok {
		t.Fatal("registry did not resolve cal-com (self-registration)")
	}
}

// TestBaseURLRejectsBadScheme guards the SSRF validation on base_url override.
func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := calcom.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "cal_live_123"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bookings", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected base_url with bad scheme to be rejected")
	}
}
