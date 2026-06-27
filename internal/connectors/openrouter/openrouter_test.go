package openrouter

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestMetadataCapabilities(t *testing.T) {
	m := Connector{}.Metadata()
	if m.Name != "openrouter" {
		t.Fatalf("name = %q", m.Name)
	}
	if !m.Capabilities.Check || !m.Capabilities.Catalog || !m.Capabilities.Read || m.Capabilities.Write {
		t.Fatalf("unexpected capabilities: %+v", m.Capabilities)
	}
}

func TestCheckFixtureMode(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := (Connector{}).Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check should pass without a key: %v", err)
	}
}

func TestCheckRequiresKey(t *testing.T) {
	if err := (Connector{}).Check(context.Background(), connectors.RuntimeConfig{}); err == nil {
		t.Fatal("Check without api_key should error")
	}
}

func TestReadFixtureModels(t *testing.T) {
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	var got []connectors.Record
	err := (Connector{}).Read(context.Background(), connectors.ReadRequest{Stream: "models", Config: cfg}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read returned no models")
	}
	if got[0]["id"] == nil {
		t.Fatalf("model record missing id: %+v", got[0])
	}
}

func TestCatalogHasModelsStream(t *testing.T) {
	cat, err := (Connector{}).Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 || cat.Streams[0].Name != "models" {
		t.Fatalf("expected a models stream, got %+v", cat.Streams)
	}
}
