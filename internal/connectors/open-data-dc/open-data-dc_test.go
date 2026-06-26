package opendatadc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	opendatadc "polymetrics.ai/internal/connectors/open-data-dc"
)

// TestReadLocationsAuthAndMapping exercises the locations stream end to end: the
// api key flows into the apikey query parameter, the request path embeds the
// configured location, records are extracted from Result.addresses, and the
// nested address.properties block is flattened into the emitted record.
func TestReadLocationsAuthAndMapping(t *testing.T) {
	var sawAPIKey, sawPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAPIKey = r.URL.Query().Get("apikey")
		sawPath = r.URL.Path
		if !strings.HasPrefix(r.URL.Path, "/locations/") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Result":{"addresses":[
			{"address":{"properties":{"MarId":"1001","FullAddress":"1350 Pennsylvania Ave NW","SSL":"0223    0801","Ward":"Ward 2","Zipcode":"20004","Latitude":38.9,"Longitude":-77.03}}},
			{"address":{"properties":{"MarId":"1002","FullAddress":"1600 Pennsylvania Ave NW","SSL":"0163    0800","Ward":"Ward 2","Zipcode":"20500","Latitude":38.9,"Longitude":-77.04}}}
		]}}`))
	}))
	defer srv.Close()

	c := opendatadc.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "location": "pennsylvania ave"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "locations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_test_123" {
		t.Fatalf("apikey query = %q, want key_test_123", sawAPIKey)
	}
	if sawPath != "/locations/pennsylvania ave" && sawPath != "/locations/pennsylvania%20ave" {
		t.Fatalf("path = %q, want /locations/<location>", sawPath)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["MarId"] != "1001" || got[0]["FullAddress"] != "1350 Pennsylvania Ave NW" {
		t.Fatalf("first record not flattened from address.properties: %+v", got[0])
	}
	if got[1]["MarId"] != "1002" {
		t.Fatalf("second record MarId = %v, want 1002", got[1]["MarId"])
	}
}

// TestReadSslsWithMaridParam confirms the ssls stream sends marid as a query
// parameter alongside the api key and extracts records from Result.ssls. This
// covers a second stream with a distinct extraction path and request shape.
func TestReadSslsWithMaridParam(t *testing.T) {
	var sawMarid, sawAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawMarid = r.URL.Query().Get("marid")
		sawAPIKey = r.URL.Query().Get("apikey")
		if r.URL.Path != "/ssls" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Result":{"ssls":[
			{"SSL":"0223    0801","MarId":"1001","FullAddress":"1350 Pennsylvania Ave NW","Square":"0223","Lot":"0801"},
			{"SSL":"0163    0800","MarId":"1002","FullAddress":"1600 Pennsylvania Ave NW","Square":"0163","Lot":"0800"}
		]}}`))
	}))
	defer srv.Close()

	c := opendatadc.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "marid": "1001"},
		Secrets: map[string]string{"api_key": "key_test_123"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ssls", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAPIKey != "key_test_123" {
		t.Fatalf("apikey query = %q, want key_test_123", sawAPIKey)
	}
	if sawMarid != "1001" {
		t.Fatalf("marid query = %q, want 1001", sawMarid)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["SSL"] != "0223    0801" || got[0]["Square"] != "0223" {
		t.Fatalf("ssl record not mapped: %+v", got[0])
	}
}

// TestFixtureModeNoNetwork verifies fixture mode emits deterministic records with
// no network access so conformance can run without live credentials.
func TestFixtureModeNoNetwork(t *testing.T) {
	c := opendatadc.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	for _, stream := range []string{"locations", "units", "ssls"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) != 2 {
			t.Fatalf("fixture %s records = %d, want 2", stream, len(got))
		}
		if got[0]["MarId"] == nil {
			t.Fatalf("fixture %s record missing MarId: %+v", stream, got[0])
		}
	}
}

// TestCheckFixtureMode confirms Check short-circuits in fixture mode.
func TestCheckFixtureMode(t *testing.T) {
	c := opendatadc.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

// TestCatalogStreams confirms the published catalog covers the three core MAR
// streams with their declared primary keys.
func TestCatalogStreams(t *testing.T) {
	c := opendatadc.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	wantPK := map[string]string{"locations": "MarId", "units": "UnitNum", "ssls": "SSL"}
	seen := map[string]bool{}
	for _, s := range cat.Streams {
		seen[s.Name] = true
		if pk, ok := wantPK[s.Name]; ok {
			if len(s.PrimaryKey) == 0 || s.PrimaryKey[0] != pk {
				t.Fatalf("stream %s primary key = %v, want [%s]", s.Name, s.PrimaryKey, pk)
			}
		}
	}
	for name := range wantPK {
		if !seen[name] {
			t.Fatalf("catalog missing stream %q", name)
		}
	}
}

// TestRegistryResolves confirms the connector self-registers and is read-only.
func TestRegistryResolves(t *testing.T) {
	r := connectors.NewRegistry()
	got, ok := r.Get("open-data-dc")
	if !ok {
		t.Fatal("registry did not resolve open-data-dc (self-registration)")
	}
	caps := got.Metadata().Capabilities
	if !caps.Read || !caps.Catalog {
		t.Fatalf("capabilities = %+v, want Read && Catalog", caps)
	}
	if caps.Write {
		t.Fatalf("capabilities = %+v, want Write false (read-only lookup API)", caps)
	}
}
