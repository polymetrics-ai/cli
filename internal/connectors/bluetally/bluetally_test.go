package bluetally_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bluetally"
)

// TestReadPaginatesAndAuthenticates is the red-first test: it asserts Bearer auth
// from the api_key secret, offset pagination across two pages (limit=50), top-level
// array record extraction, and id-based record mapping.
func TestReadPaginatesAndAuthenticates(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/assets" {
			http.NotFound(w, r)
			return
		}
		offset := r.URL.Query().Get("offset")
		if got := r.URL.Query().Get("limit"); got != "50" {
			t.Errorf("limit = %q, want 50", got)
		}
		switch offset {
		case "", "0":
			// First full page of 50 records forces a second request.
			w.Write([]byte(buildAssetPage(1, 50)))
		case "50":
			// Short page ends pagination.
			w.Write([]byte(buildAssetPage(51, 2)))
		default:
			t.Errorf("unexpected offset=%q", offset)
			w.Write([]byte("[]"))
		}
	}))
	defer srv.Close()

	c := bluetally.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL},
		Secrets: map[string]string{"api_key": "bt_test_key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "assets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer bt_test_key" {
		t.Fatalf("Authorization = %q, want Bearer bt_test_key", sawAuth)
	}
	if len(got) != 52 {
		t.Fatalf("records = %d, want 52 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil {
			t.Fatalf("record missing id: %+v", rec)
		}
	}
	if got[0]["asset_name"] == nil {
		t.Fatalf("first asset record not mapped: %+v", got[0])
	}
}

// buildAssetPage renders a JSON top-level array of `count` asset objects starting
// at the given id.
func buildAssetPage(startID, count int) string {
	out := "["
	for i := 0; i < count; i++ {
		if i > 0 {
			out += ","
		}
		id := startID + i
		out += `{"id":` + strconv.Itoa(id) +
			`,"asset_name":"Asset ` + strconv.Itoa(id) +
			`","asset_serial":"SN` + strconv.Itoa(id) +
			`","status_id":1,"updated_at":"2026-01-0` + strconv.Itoa((id%9)+1) + `T00:00:00Z"}`
	}
	out += "]"
	return out
}

func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := bluetally.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	for _, stream := range []string{"assets", "employees", "licenses", "maintenances", "accessories"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) produced no records", stream)
		}
		for _, rec := range got {
			if rec["id"] == nil {
				t.Fatalf("fixture %s record missing id: %+v", stream, rec)
			}
		}
	}
}

func TestCheckFixtureMode(t *testing.T) {
	c := bluetally.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCatalogStreams(t *testing.T) {
	c := bluetally.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "bluetally" {
		t.Fatalf("catalog connector = %q, want bluetally", cat.Connector)
	}
	want := map[string]bool{"assets": true, "employees": true, "licenses": true, "maintenances": true, "accessories": true}
	for _, s := range cat.Streams {
		delete(want, s.Name)
		if len(s.PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", s.Name)
		}
	}
	if len(want) != 0 {
		t.Fatalf("catalog missing streams: %v", want)
	}
}

func TestMetadataReadOnly(t *testing.T) {
	c := bluetally.New()
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("bluetally is read-only; Write should be false")
	}
}

func TestRegisteredInRegistry(t *testing.T) {
	_ = bluetally.New() // ensure init ran
	r := connectors.NewRegistry()
	if _, ok := r.Get("bluetally"); !ok {
		t.Fatal("registry did not resolve bluetally (self-registration)")
	}
}

func TestBaseURLRejectsBadScheme(t *testing.T) {
	c := bluetally.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"api_key": "k"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "assets", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with ftp base_url should be rejected (SSRF guard)")
	}
}
