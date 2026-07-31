package googleanalyticsdataapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "google-analytics-data-api")
}

func TestReadFixtureCoversEveryStream(t *testing.T) {
	ctx := context.Background()
	c := New()
	cat, err := c.Catalog(ctx, connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}

	for _, stream := range cat.Streams {
		stream := stream
		t.Run(stream.Name, func(t *testing.T) {
			var records []connectors.Record
			err := c.Read(ctx, connectors.ReadRequest{
				Stream: stream.Name,
				Config: connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}},
				State:  map[string]string{"cursor": "20251231"},
			}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Read(%s fixture): %v", stream.Name, err)
			}
			if len(records) != gaFixtureRecordCount {
				t.Fatalf("Read(%s) emitted %d records, want %d", stream.Name, len(records), gaFixtureRecordCount)
			}
			first := records[0]
			if first["property_id"] != gaFixturePropertyID {
				t.Fatalf("property_id = %v, want %s", first["property_id"], gaFixturePropertyID)
			}
			if first["previous_cursor"] != "20251231" {
				t.Fatalf("previous_cursor = %v, want 20251231", first["previous_cursor"])
			}
			for _, field := range stream.Fields {
				if _, ok := first[field.Name]; !ok {
					t.Fatalf("fixture record for %s missing field %s: %#v", stream.Name, field.Name, first)
				}
			}
		})
	}
}

func TestReadRunReportUsesFixedPostBodiesForEveryStream(t *testing.T) {
	ctx := context.Background()
	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1beta/properties/123456:runReport" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer fixture-access-token" {
			t.Fatalf("Authorization = %q", got)
		}
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Fatalf("close request body: %v", err)
			}
		}()
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, body)
		assertNumber(t, body["offset"], 0)
		assertNumber(t, body["limit"], 2)
		if body["keepEmptyRows"] != false {
			t.Fatalf("keepEmptyRows = %v, want false", body["keepEmptyRows"])
		}
		dateRanges, ok := body["dateRanges"].([]any)
		if !ok || len(dateRanges) != 1 {
			t.Fatalf("dateRanges = %#v", body["dateRanges"])
		}
		dateRange, ok := dateRanges[0].(map[string]any)
		if !ok || dateRange["startDate"] != "2026-01-01" || dateRange["endDate"] != "yesterday" {
			t.Fatalf("dateRange = %#v", dateRanges[0])
		}

		dims := namesFromBody(t, body, "dimensions")
		metrics := namesFromBody(t, body, "metrics")
		writeReportResponse(t, w, dims, metrics)
	}))
	defer server.Close()

	c := New()
	c.Client = server.Client()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":               server.URL,
			"property_ids":           "properties/123456,999999",
			"date_ranges_start_date": "2026-01-01",
			"date_ranges_end_date":   "yesterday",
			"page_size":              "2",
			"max_pages":              "1",
		},
		Secrets: map[string]string{"access_token": "fixture-access-token"},
	}
	for _, stream := range gaStreamOrder {
		stream := stream
		t.Run(stream, func(t *testing.T) {
			var records []connectors.Record
			err := c.Read(ctx, connectors.ReadRequest{Stream: stream, Config: cfg}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Read(%s): %v", stream, err)
			}
			if len(records) != 1 {
				t.Fatalf("Read(%s) emitted %d records, want 1", stream, len(records))
			}
			for _, dim := range gaReports[stream].dimensions {
				if _, ok := records[0][dim]; !ok {
					t.Fatalf("record missing dimension %s: %#v", dim, records[0])
				}
			}
			for _, metric := range gaReports[stream].metrics {
				if records[0][metric] != "1" {
					t.Fatalf("record[%s] = %v, want 1", metric, records[0][metric])
				}
			}
		})
	}
	if len(requests) != len(gaStreamOrder) {
		t.Fatalf("server saw %d runReport requests, want %d", len(requests), len(gaStreamOrder))
	}
}

func TestOperationDirectReadFixtureCoversImplementedOperations(t *testing.T) {
	ctx := context.Background()
	reader, ok := any(New()).(connectors.OperationDirectReader)
	if !ok {
		t.Fatal("New() does not expose OperationDirectReader")
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{
		"mode":               "fixture",
		"property_ids":       "properties/123456 999999",
		"audience_export_id": "audience_export_fixture_1",
	}}
	for _, operation := range []string{
		"google-analytics-data-api.get_metadata",
		"google-analytics-data-api.list_audience_exports",
		"google-analytics-data-api.get_audience_export",
	} {
		operation := operation
		t.Run(operation, func(t *testing.T) {
			result, err := reader.OperationDirectRead(ctx, connectors.OperationDirectReadRequest{
				Operation:    operation,
				Config:       cfg,
				OutputPolicy: "json_redacted",
			})
			if err != nil {
				t.Fatalf("OperationDirectRead(%s fixture): %v", operation, err)
			}
			if result.Connector != connectorName || result.Method != http.MethodGet || result.Status != http.StatusOK {
				t.Fatalf("result metadata = %#v", result)
			}
			if !strings.Contains(result.Path, "123456") {
				t.Fatalf("result path = %q, want normalized property id", result.Path)
			}
			if result.Body == nil {
				t.Fatalf("result body is nil")
			}
		})
	}
	_, err := reader.OperationDirectRead(ctx, connectors.OperationDirectReadRequest{
		Operation: "google-analytics-data-api.run_realtime_report",
		Config:    cfg,
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported connector operation") {
		t.Fatalf("unsupported planned operation error = %v", err)
	}
}

func TestOperationDirectReadLiveUsesFixedGETEndpoints(t *testing.T) {
	ctx := context.Background()
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer fixture-access-token" {
			t.Fatalf("Authorization = %q", got)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		seen[r.URL.RequestURI()] = true
		switch r.URL.Path {
		case "/v1beta/properties/123456/metadata":
			writeJSON(t, w, map[string]any{"name": "properties/123456/metadata", "dimensions": []any{map[string]any{"apiName": "date"}}})
		case "/v1beta/properties/123456/audienceExports":
			if r.URL.Query().Get("pageSize") != "2" || r.URL.Query().Get("pageToken") != "next" {
				t.Fatalf("query = %s", r.URL.RawQuery)
			}
			writeJSON(t, w, map[string]any{"audienceExports": []any{map[string]any{"name": "properties/123456/audienceExports/audience_export_fixture_1"}}})
		case "/v1beta/properties/123456/audienceExports/audience_export_fixture_1":
			writeJSON(t, w, map[string]any{"name": "properties/123456/audienceExports/audience_export_fixture_1", "state": "ACTIVE"})
		default:
			t.Fatalf("unexpected path %s", r.URL.RequestURI())
		}
	}))
	defer server.Close()

	c := New()
	c.Client = server.Client()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL, "property_ids": "123456", "audience_export_id": "audience_export_fixture_1"},
		Secrets: map[string]string{"access_token": "fixture-access-token"},
	}
	for _, req := range []connectors.OperationDirectReadRequest{
		{Operation: "google-analytics-data-api.get_metadata", Config: cfg, OutputPolicy: "json_redacted"},
		{Operation: "google-analytics-data-api.list_audience_exports", Config: cfg, Query: map[string]string{"pageSize": "2", "pageToken": "next"}, OutputPolicy: "json_redacted"},
		{Operation: "google-analytics-data-api.get_audience_export", Config: cfg, OutputPolicy: "json_redacted"},
	} {
		result, err := c.OperationDirectRead(ctx, req)
		if err != nil {
			t.Fatalf("OperationDirectRead(%s): %v", req.Operation, err)
		}
		if result.Connector != connectorName || result.Method != http.MethodGet || result.Status != http.StatusOK {
			t.Fatalf("result metadata = %#v", result)
		}
	}
	for _, uri := range []string{
		"/v1beta/properties/123456/metadata",
		"/v1beta/properties/123456/audienceExports?pageSize=2&pageToken=next",
		"/v1beta/properties/123456/audienceExports/audience_export_fixture_1",
	} {
		if !seen[uri] {
			t.Fatalf("server did not see %s; saw %#v", uri, seen)
		}
	}
}

func assertNumber(t *testing.T, value any, want float64) {
	t.Helper()
	got, ok := value.(float64)
	if !ok || got != want {
		t.Fatalf("number = %#v, want %v", value, want)
	}
}

func namesFromBody(t *testing.T, body map[string]any, key string) []string {
	t.Helper()
	items, ok := body[key].([]any)
	if !ok {
		t.Fatalf("%s = %#v", key, body[key])
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("%s item = %#v", key, item)
		}
		name, _ := m["name"].(string)
		if name == "" {
			t.Fatalf("%s item missing name: %#v", key, item)
		}
		names = append(names, name)
	}
	return names
}

func writeReportResponse(t *testing.T, w http.ResponseWriter, dims, metrics []string) {
	t.Helper()
	dimensionHeaders := make([]any, 0, len(dims))
	dimensionValues := make([]any, 0, len(dims))
	for _, dim := range dims {
		dimensionHeaders = append(dimensionHeaders, map[string]any{"name": dim})
		value := dim + "_value"
		if dim == "date" {
			value = "20260101"
		}
		dimensionValues = append(dimensionValues, map[string]any{"value": value})
	}
	metricHeaders := make([]any, 0, len(metrics))
	metricValues := make([]any, 0, len(metrics))
	for _, metric := range metrics {
		metricHeaders = append(metricHeaders, map[string]any{"name": metric})
		metricValues = append(metricValues, map[string]any{"value": "1"})
	}
	writeJSON(t, w, map[string]any{
		"dimensionHeaders": dimensionHeaders,
		"metricHeaders":    metricHeaders,
		"rows": []any{map[string]any{
			"dimensionValues": dimensionValues,
			"metricValues":    metricValues,
		}},
		"rowCount": 1,
	})
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func assertConnectorContract(t *testing.T, c connectors.Connector, wantName string) {
	t.Helper()
	if c == nil {
		t.Fatal("New() = nil")
	}
	if got := c.Name(); got != wantName {
		t.Fatalf("Name() = %q, want %q", got, wantName)
	}
	meta := c.Metadata()
	if meta.Name != wantName {
		t.Fatalf("Metadata().Name = %q, want %q", meta.Name, wantName)
	}
	caps := meta.Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check, Catalog, and Read", caps)
	}
	if caps.Write {
		t.Fatalf("%s is read-only; Write capability must be false", wantName)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != wantName {
		t.Fatalf("Catalog().Connector = %q, want %q", cat.Connector, wantName)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("Catalog returned zero streams")
	}
	manifest := connectors.ManifestOf(c)
	if len(manifest.Streams) != len(cat.Streams) {
		t.Fatalf("Manifest streams = %d, want catalog streams %d", len(manifest.Streams), len(cat.Streams))
	}
	if len(manifest.ConfigFields) == 0 || len(manifest.SecretFields) == 0 || len(manifest.AuthModes) == 0 {
		t.Fatalf("Manifest missing config/secret/auth guidance: %+v", manifest)
	}
}
