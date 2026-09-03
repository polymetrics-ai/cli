package hubspot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/discovery"
	"polymetrics.ai/internal/connectors/engine"
)

func TestCatalogAndReadUnknownCustomObject(t *testing.T) {
	server := newHubSpotFixtureServer(t)
	defer server.Close()

	connector := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}}
	catalog, err := connector.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	stream := catalogStream(t, catalog, "2-8675309")
	if _, err := engine.CompileSchema(stream.Schema); err != nil {
		t.Fatalf("discovered schema is not accepted by the same static schema compiler: %v", err)
	}
	if stream.PrimaryKey[0] != "hs_object_id" || stream.CursorFields[0] != "hs_lastmodifieddate" {
		t.Fatalf("custom stream sync fields = %#v", stream)
	}
	var schema map[string]json.RawMessage
	if err := json.Unmarshal(stream.Schema, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	var properties map[string]json.RawMessage
	if err := json.Unmarshal(schema["properties"], &properties); err != nil {
		t.Fatalf("unmarshal properties: %v", err)
	}
	if _, ok := properties["tenant_only_flag"]; !ok {
		t.Fatalf("custom field missing from schema: %s", stream.Schema)
	}
	if future := properties["provider_future_type"]; len(future) == 0 || strings.Contains(string(future), `"type"`) {
		t.Fatalf("unknown provider type was estimated instead of preserved: %s", future)
	}
	if !strings.Contains(string(properties["tenant_only_flag"]), `"type":"boolean"`) || !strings.Contains(string(properties["plan"]), `"enum":["basic","enterprise"]`) || !strings.Contains(string(properties["hs_lastmodifieddate"]), `"format":"date-time"`) {
		t.Fatalf("provider field type mapping is incomplete: %s", stream.Schema)
	}
	if !strings.Contains(string(properties["related_record"]), `"x-references":["2-42"]`) {
		t.Fatalf("reference field schema = %s", properties["related_record"])
	}

	var records []connectors.Record
	err = connector.Read(context.Background(), connectors.ReadRequest{Stream: "2-8675309", Config: cfg, Limit: 1}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v, want one", records)
	}
	if _, ok := records[0]["not_in_discovery_schema"]; ok {
		t.Fatalf("read emitted field absent from discovered schema: %#v", records[0])
	}
	if got := records[0]["tenant_only_flag"]; got != true {
		t.Fatalf("custom field value = %#v, want true", got)
	}
}

func TestCatalogRetriesHubSpotRateLimitThroughSharedDriver(t *testing.T) {
	base := hubSpotFixtureHandler(t)
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/crm/v3/properties/2-8675309" && attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"status":"error"}`))
			return
		}
		base.ServeHTTP(w, request)
	}))
	defer server.Close()

	connector := New()
	connector.driver = testDriver(t)
	_, err := connector.Catalog(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("rate-limited property attempts = %d, want 2", attempts.Load())
	}
}

func TestCatalogUsesDeclaredFallbackAndMarksItPartial(t *testing.T) {
	base := hubSpotFixtureHandler(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == schemasPath {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"error"}`))
			return
		}
		base.ServeHTTP(w, request)
	}))
	defer server.Close()

	connector := New()
	connector.driver = testDriver(t)
	catalog, err := connector.Catalog(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"base_url": server.URL}})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if catalog.Discovery == nil || catalog.Discovery.Complete || !catalog.Discovery.UsedFallback {
		t.Fatalf("fallback discovery status = %#v", catalog.Discovery)
	}
	if _, found := discoveredStream(catalog, "contacts"); !found {
		t.Fatalf("fallback catalog omitted declared standard stream: %#v", catalog.Streams)
	}
}

func catalogStream(t *testing.T, catalog connectors.Catalog, name string) connectors.Stream {
	t.Helper()
	for _, stream := range catalog.Streams {
		if stream.Name == name {
			return stream
		}
	}
	t.Fatalf("catalog has no %q stream: %#v", name, catalog.Streams)
	return connectors.Stream{}
}

func newHubSpotFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(hubSpotFixtureHandler(t))
}

func hubSpotFixtureHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/crm-object-schemas/v3/schemas":
			_, _ = w.Write([]byte(`{"results":[{"objectTypeId":"2-8675309","labels":{"singular":"Tenant Subscription","plural":"Tenant Subscriptions"}}]}`))
		case "/crm/v3/properties/2-8675309":
			_, _ = w.Write([]byte(`{"results":[
              {"name":"hs_object_id","label":"Object ID","type":"number","hasUniqueValue":true},
              {"name":"hs_lastmodifieddate","label":"Last modified","type":"datetime"},
              {"name":"tenant_only_flag","label":"Tenant-only flag","type":"bool"},
              {"name":"provider_future_type","label":"Future provider type","type":"future_type"},
              {"name":"plan","label":"Plan","type":"enumeration","options":[{"value":"basic"},{"value":"enterprise"}]},
              {"name":"related_record","label":"Related record","type":"string","referencedObjectType":"2-42"}
            ]}`))
		case "/crm/v3/objects/2-8675309":
			if got := request.URL.Query().Get("properties"); !strings.Contains(got, "tenant_only_flag") {
				t.Errorf("read properties = %q, missing discovered custom field", got)
			}
			_, _ = w.Write([]byte(`{"results":[{"id":"123","properties":{"hs_object_id":"123","hs_lastmodifieddate":"2026-08-06T00:00:00Z","tenant_only_flag":true,"plan":"enterprise","related_record":"456","not_in_discovery_schema":"drop-me"}}]}`))
		default:
			if strings.HasPrefix(request.URL.Path, "/crm/v3/properties/") {
				_, _ = w.Write([]byte(`{"results":[{"name":"hs_object_id","type":"number","hasUniqueValue":true},{"name":"hs_lastmodifieddate","type":"datetime"}]}`))
				return
			}
			http.NotFound(w, request)
		}
	})
}

func testDriver(t *testing.T) *discovery.Driver {
	t.Helper()
	driver, err := discovery.New(discovery.Spec{
		Connector:          connectorName,
		Fallback:           standardObjects(),
		FallbackPrimaryKey: []string{"hs_object_id"},
		FallbackCursor:     "hs_lastmodifieddate",
		Converter:          hubspotFieldSchema,
		MaxAttempts:        2,
		Sleep:              func(context.Context, time.Duration) error { return nil },
		Jitter:             func(time.Duration) time.Duration { return 0 },
	})
	if err != nil {
		t.Fatalf("New discovery driver: %v", err)
	}
	return driver
}
