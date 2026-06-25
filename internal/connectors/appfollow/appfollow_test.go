package appfollow_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/appfollow"
)

// TestReadAppCollectionsAuthAndRecordPath is the red-first test for the
// AppFollow connector: the X-AppFollow-API-Token header is sent, records are
// extracted from the nested "apps" array, and each app collection is mapped.
func TestReadAppCollectionsAuthAndRecordPath(t *testing.T) {
	var sawToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawToken = r.Header.Get("X-AppFollow-API-Token")
		if r.URL.Path != "/account/apps" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"apps":[{"id":11,"title":"Collection A","count_apps":3},{"id":22,"title":"Collection B","count_apps":1}]}`))
	}))
	defer srv.Close()

	c := appfollow.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_secret": "af_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "app_collections", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawToken != "af_secret_123" {
		t.Fatalf("X-AppFollow-API-Token = %q, want af_secret_123", sawToken)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["title"] == nil {
		t.Fatalf("record missing id/title: %+v", got[0])
	}
}

// TestReadRatingsFansOutOverExtIDs proves the connector issues one request per
// ext_id (multi-request traversal across two "pages") and forwards the ext_id as
// a query parameter, mapping the flattened ratings list rows.
func TestReadRatingsFansOutOverExtIDs(t *testing.T) {
	var seenExtIDs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta/ratings" {
			http.NotFound(w, r)
			return
		}
		ext := r.URL.Query().Get("ext_id")
		seenExtIDs = append(seenExtIDs, ext)
		switch ext {
		case "ios:123":
			_, _ = w.Write([]byte(`{"ratings":{"ext_id":"ios:123","store":"ios","list":[{"date":"2026-01-01","country":"US","rating":4.5,"stars_total":100}]}}`))
		case "android:456":
			_, _ = w.Write([]byte(`{"ratings":{"ext_id":"android:456","store":"android","list":[{"date":"2026-01-02","country":"GB","rating":4.1,"stars_total":50}]}}`))
		default:
			_, _ = w.Write([]byte(`{"ratings":{"list":[]}}`))
		}
	}))
	defer srv.Close()

	c := appfollow.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "ext_ids": "ios:123,android:456"},
		Secrets: map[string]string{"api_secret": "af_secret_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ratings", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(seenExtIDs) != 2 {
		t.Fatalf("requests = %d, want 2 (one per ext_id)", len(seenExtIDs))
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one row per ext_id)", len(got))
	}
	for _, rec := range got {
		if rec["ext_id"] == nil || rec["rating"] == nil || rec["date"] == nil {
			t.Fatalf("ratings record missing fields: %+v", rec)
		}
	}
}

// TestFixtureModeReadsWithoutNetwork ensures credential-free conformance can
// exercise the connector: fixture mode emits deterministic records over every
// core stream with no HTTP call.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := appfollow.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"users", "app_collections", "app_lists", "ratings"} {
		count := 0
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(connectors.Record) error {
			count++
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if count == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
	}
}

func TestCheckFixtureModeSucceeds(t *testing.T) {
	c := appfollow.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestRegisteredReadOnly(t *testing.T) {
	_ = appfollow.New() // ensure init ran
	caps := appfollow.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write=false (read-only API)", caps)
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("appfollow"); !ok {
		t.Fatal("registry did not resolve appfollow (self-registration)")
	}
}

func TestCatalogStreams(t *testing.T) {
	c := appfollow.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	want := map[string]bool{"users": false, "app_collections": false, "app_lists": false, "ratings": false}
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
