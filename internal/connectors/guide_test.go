package connectors

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type guideTestConnector struct {
	manifest Manifest
	surface  *CommandSurface
}

func (c guideTestConnector) Name() string { return c.manifest.Metadata.Name }
func (c guideTestConnector) Metadata() Metadata {
	return c.manifest.Metadata
}
func (c guideTestConnector) Check(context.Context, RuntimeConfig) error { return nil }
func (c guideTestConnector) Catalog(context.Context, RuntimeConfig) (Catalog, error) {
	return Catalog{}, nil
}
func (c guideTestConnector) Read(context.Context, ReadRequest, func(Record) error) error { return nil }
func (c guideTestConnector) Write(context.Context, WriteRequest, []Record) (WriteResult, error) {
	return WriteResult{}, ErrUnsupportedOperation
}
func (c guideTestConnector) Manifest() Manifest              { return c.manifest }
func (c guideTestConnector) CommandSurface() *CommandSurface { return c.surface }

func TestGuideSeparatesDirectReadsProviderOperationsAndWarehouseQuery(t *testing.T) {
	manual := RenderConnectorManual(guideTestConnector{
		manifest: Manifest{
			Metadata: Metadata{
				Name:            "acme",
				DisplayName:     "Acme",
				Description:     "Acme connector.",
				IntegrationType: "api",
				Capabilities:    Capabilities{Check: true, Read: true, Query: false, ProviderSearch: true, ProviderQuery: true},
			},
			Streams: []Stream{{Name: "widgets", Description: "Widget stream."}},
			ProviderOperations: []ProviderOperationInfo{
				{
					ID:             "acme.widgets.search",
					Kind:           "provider_search",
					Summary:        "Search widgets.",
					OutputPolicy:   "json_redacted",
					RequestSchema:  json.RawMessage(`{"type":"object"}`),
					ResponseSchema: json.RawMessage(`{"type":"object"}`),
					Bounds:         ProviderOperationBounds{DefaultLimit: 25, MaxLimit: 50, MaxPages: 2, MaxBytes: 65536},
				},
				{
					ID:             "acme.widgets.query",
					Kind:           "provider_query",
					Summary:        "Query widgets with typed filters.",
					OutputPolicy:   "json_redacted",
					RequestSchema:  json.RawMessage(`{"type":"object"}`),
					ResponseSchema: json.RawMessage(`{"type":"object"}`),
					Bounds:         ProviderOperationBounds{DefaultLimit: 10, MaxLimit: 10, MaxPages: 1, MaxBytes: 32768},
				},
			},
			Risk: RiskSpec{Read: "low"},
		},
		surface: &CommandSurface{Usage: "pm acme <command> [flags]", Commands: []CommandSurfaceCommand{
			{Path: "widget read", Summary: "Read one widget.", Intent: "direct_read", Availability: "implemented", OutputPolicy: "json_redacted"},
			{Path: "widget search", Summary: "Search widgets.", Intent: "provider_search", Availability: "implemented", Operation: "acme.widgets.search", OutputPolicy: "json_redacted"},
		}},
	})

	for _, want := range []string{
		"CAPABILITIES",
		"warehouse query: false",
		"provider_search=true provider_query=true",
		"ETL STREAMS",
		"DIRECT READ COMMANDS",
		"PROVIDER SEARCH/QUERY OPERATIONS",
		"acme.widgets.search",
		"max_limit=50 max_pages=2 max_bytes=65536",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("manual missing %q:\n%s", want, manual)
		}
	}
}

func TestEveryRegisteredConnectorHasGuideManualAndSkill(t *testing.T) {
	registry := NewRegistry()
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			t.Fatalf("connector %s not found", meta.Name)
		}
		if err := ValidateConnectorGuide(connector); err != nil {
			t.Fatalf("ValidateConnectorGuide(%s) error = %v", meta.Name, err)
		}
		manual := RenderConnectorManual(connector)
		skill := RenderConnectorSkill(connector)
		if strings.Contains(manual, "{\n") {
			t.Fatalf("manual for %s should be human-readable, not raw JSON:\n%s", meta.Name, manual)
		}
		if strings.Contains(skill, "ghp_") || strings.Contains(skill, "secret-token") {
			t.Fatalf("skill for %s contains secret-like text:\n%s", meta.Name, skill)
		}
	}
}
