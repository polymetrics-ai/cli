package cimis_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/cimis"
)

// TestReadDataAuthenticatesAndFlattensProviders is the red-first test for the
// CIMIS connector: it asserts the appKey query-param auth, the required date
// range query params, that records are flattened out of Data.Providers[].Records[]
// across two providers, and that the data-item record mapping lifts nested
// {Value,Qc,Unit} objects. Red until internal/connectors/cimis exists.
func TestReadDataAuthenticatesAndFlattensProviders(t *testing.T) {
	var sawAppKey, sawTargets, sawStart, sawEnd, sawDataItems string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/data" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		sawAppKey = q.Get("appKey")
		sawTargets = q.Get("targets")
		sawStart = q.Get("startDate")
		sawEnd = q.Get("endDate")
		sawDataItems = q.Get("dataItems")
		_, _ = w.Write([]byte(`{"Data":{"Providers":[
			{"Name":"cimis","Type":"station","Owner":"a@b","Records":[
				{"Date":"2026-01-01","Julian":"1","Station":"2","Standard":"english","ZipCodes":["95823"],"Scope":"daily","DayAirTmpAvg":{"Value":"50.1","Qc":"","Unit":"(F)"}},
				{"Date":"2026-01-02","Julian":"2","Station":"2","Standard":"english","ZipCodes":["95823"],"Scope":"daily","DayAirTmpAvg":{"Value":"51.2","Qc":"","Unit":"(F)"}}
			]},
			{"Name":"scs","Type":"station","Owner":"c@d","Records":[
				{"Date":"2026-01-01","Julian":"1","Station":"5","Standard":"english","ZipCodes":["94203"],"Scope":"daily","DayAirTmpAvg":{"Value":"48.0","Qc":"","Unit":"(F)"}}
			]}
		]}}`))
	}))
	defer srv.Close()

	c := cimis.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":          srv.URL,
			"targets":           "2,5",
			"start_date":        "2026-01-01",
			"end_date":          "2026-01-02",
			"daily_data_items":  "day-air-tmp-avg",
			"hourly_data_items": "hly-air-tmp",
		},
		Secrets: map[string]string{"api_key": "test-app-key"},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAppKey != "test-app-key" {
		t.Fatalf("appKey = %q, want test-app-key", sawAppKey)
	}
	if sawTargets != "2,5" {
		t.Fatalf("targets = %q, want 2,5", sawTargets)
	}
	if sawStart != "2026-01-01" || sawEnd != "2026-01-02" {
		t.Fatalf("date range = (%q,%q), want (2026-01-01,2026-01-02)", sawStart, sawEnd)
	}
	if sawDataItems != "day-air-tmp-avg" {
		t.Fatalf("dataItems = %q, want day-air-tmp-avg (daily stream)", sawDataItems)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 providers flattened)", len(got))
	}
	for _, rec := range got {
		if rec["Station"] == nil || rec["Date"] == nil {
			t.Fatalf("record missing Station/Date: %+v", rec)
		}
		if rec["DayAirTmpAvg_Value"] != "50.1" && rec["DayAirTmpAvg_Value"] != "51.2" && rec["DayAirTmpAvg_Value"] != "48.0" {
			t.Fatalf("data item not flattened: %+v", rec)
		}
	}
}

// TestReadStationsNoAppKey verifies the stations stream reads the station
// metadata endpoint (which CIMIS serves without an appKey) and flattens the
// Stations array.
func TestReadStationsNoAppKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/station" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"Stations":[
			{"StationNbr":"2","Name":"FivePoints","City":"Five Points","County":"Fresno","IsActive":"True"},
			{"StationNbr":"5","Name":"Shafter","City":"Shafter","County":"Kern","IsActive":"True"}
		]}`))
	}))
	defer srv.Close()

	c := cimis.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": srv.URL, "targets": "2", "start_date": "2026-01-01", "end_date": "2026-01-02"},
		Secrets: map[string]string{"api_key": "test-app-key"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "stations", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read stations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("stations = %d, want 2", len(got))
	}
	if got[0]["StationNbr"] == nil {
		t.Fatalf("station missing StationNbr: %+v", got[0])
	}
}

// TestFixtureModeReadsWithoutNetwork confirms fixture mode emits deterministic
// records with no network call, so conformance works without live creds.
func TestFixtureModeReadsWithoutNetwork(t *testing.T) {
	c := cimis.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "daily", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture mode produced no records")
	}
	for _, rec := range got {
		if rec["Station"] == nil || rec["Date"] == nil {
			t.Fatalf("fixture record missing Station/Date: %+v", rec)
		}
	}
	// Check in fixture mode must not require creds or network.
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

// TestRegistryResolvesCimis confirms self-registration through init() and that
// the connector is read-only.
func TestRegistryResolvesCimis(t *testing.T) {
	_ = cimis.New()
	r := connectors.NewRegistry()
	if _, ok := r.Get("cimis"); !ok {
		t.Fatal("registry did not resolve cimis (self-registration)")
	}
	caps := cimis.New().Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read && Catalog && Check", caps)
	}
	if caps.Write {
		t.Fatalf("cimis should be read-only, got Write=true")
	}
}
