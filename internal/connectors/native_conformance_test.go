package connectors

import (
	"context"
	"errors"
	"testing"
)

func TestNativeRegistryBindsOnlyEnabledCatalogConnectors(t *testing.T) {
	registry := NewRegistry()
	enabled := 0
	for _, def := range ConnectorCatalog() {
		if def.ImplementationStatus != ImplementationEnabled {
			if _, ok := registry.Get(def.Slug); ok {
				t.Fatalf("registry exposes planned connector %s as executable", def.Slug)
			}
			continue
		}
		enabled++
		connector, ok := registry.Get(def.Slug)
		if !ok {
			t.Fatalf("registry missing native binding for %s", def.Slug)
		}
		if connector.Name() != def.Slug {
			t.Fatalf("native binding name = %q, want %q", connector.Name(), def.Slug)
		}
		if connector.Metadata().Capabilities != def.RuntimeCapabilities.toCapabilities() {
			t.Fatalf("%s metadata capabilities = %+v, want %+v", def.Slug, connector.Metadata().Capabilities, def.RuntimeCapabilities.toCapabilities())
		}
	}
	if enabled != 2 {
		t.Fatalf("enabled catalog connector count = %d, want 2", enabled)
	}
	if _, ok := registry.Get("source-github"); !ok {
		t.Fatal("registry missing source-github catalog alias")
	}
	if _, ok := registry.Get("source-stripe"); !ok {
		t.Fatal("registry missing source-stripe catalog alias after native port is enabled")
	}
}

func TestNativeFixtureConformancePassesForEveryCatalogConnector(t *testing.T) {
	reports := NativeConformanceReports(context.Background(), ConnectorCatalog())
	if got, want := len(reports), 646; got != want {
		t.Fatalf("NativeConformanceReports len = %d, want %d", got, want)
	}
	for _, report := range reports {
		if !report.Passed {
			t.Fatalf("native conformance failed for %s: %+v", report.Slug, report)
		}
		if report.RuntimeKind == "" || len(report.Tests) == 0 || len(report.BenchmarkHooks) == 0 {
			t.Fatalf("native conformance report incomplete for %s: %+v", report.Slug, report)
		}
	}
}

func TestNativeCatalogFixtureOperationsDoNotEnablePlannedConnectors(t *testing.T) {
	ctx := context.Background()
	sourceDef, ok := ConnectorDefinitionBySlug("source-strava")
	if !ok {
		t.Fatal("source-strava not found")
	}
	source := NewNativeCatalogConnector(sourceDef)
	if err := source.Check(ctx, RuntimeConfig{}); !errors.Is(err, ErrUnsupportedOperation) {
		t.Fatalf("source check error = %v, want ErrUnsupportedOperation", err)
	}
	fixtureConfig := RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	catalog, err := source.Catalog(ctx, fixtureConfig)
	if err != nil {
		t.Fatalf("source catalog: %v", err)
	}
	if len(catalog.Streams) == 0 {
		t.Fatalf("source catalog has no streams: %+v", catalog)
	}
	var records []Record
	if err := source.Read(ctx, ReadRequest{Stream: catalog.Streams[0].Name, Config: fixtureConfig}, func(record Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("source read: %v", err)
	}
	if len(records) != 1 || records[0]["connector_slug"] != "source-strava" {
		t.Fatalf("source records = %+v", records)
	}

	destDef, ok := ConnectorDefinitionBySlug("destination-postgres")
	if !ok {
		t.Fatal("destination-postgres not found")
	}
	var dest Connector = NewNativeCatalogConnector(destDef)
	if err := dest.Check(ctx, RuntimeConfig{}); !errors.Is(err, ErrUnsupportedOperation) {
		t.Fatalf("destination check error = %v, want ErrUnsupportedOperation", err)
	}
	if validator, ok := dest.(WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, WriteRequest{Action: "upsert", Config: fixtureConfig}, records); err != nil {
			t.Fatalf("destination validate write: %v", err)
		}
	} else {
		t.Fatal("destination does not implement WriteValidator")
	}
	result, err := dest.Write(ctx, WriteRequest{Action: "upsert", Config: fixtureConfig}, records)
	if err != nil {
		t.Fatalf("destination write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("destination write result = %+v", result)
	}
	querier, ok := dest.(Querier)
	if !ok {
		t.Fatal("destination does not implement Querier")
	}
	queryResult, err := querier.Query(ctx, QueryRequest{SQL: "select * from records", Limit: 1, Config: fixtureConfig})
	if err != nil {
		t.Fatalf("destination query: %v", err)
	}
	if len(queryResult.Rows) != 1 {
		t.Fatalf("destination query rows = %+v", queryResult.Rows)
	}
}
