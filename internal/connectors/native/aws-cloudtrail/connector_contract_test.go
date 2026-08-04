package awscloudtrail

import (
	"context"
	"os"
	"slices"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "aws-cloudtrail")
}

func TestCommandSurfaceExposesDocumentedOperations(t *testing.T) {
	provider, ok := New().(connectors.CommandSurfaceProvider)
	if !ok {
		t.Fatal("New() does not implement connectors.CommandSurfaceProvider")
	}
	surface := provider.CommandSurface()
	if surface == nil {
		t.Fatal("CommandSurface() = nil")
	}
	if got, want := len(surface.Commands), 60; got != want {
		t.Fatalf("CommandSurface commands = %d, want %d documented CloudTrail operations", got, want)
	}
	availability := map[string]int{}
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, command := range surface.Commands {
		availability[command.Availability]++
		commands[command.Path] = command
	}
	if got, want := availability["implemented"], 57; got != want {
		t.Fatalf("implemented command rows = %d, want %d", got, want)
	}
	if got, want := availability["unsafe_or_disallowed"], 3; got != want {
		t.Fatalf("policy-disallowed command rows = %d, want %d", got, want)
	}
	for path, wantAvailability := range map[string]string{
		"events lookup":    "implemented",
		"query cancel":     "implemented",
		"tags add":         "implemented",
		"query start":      "unsafe_or_disallowed",
		"dashboard create": "unsafe_or_disallowed",
	} {
		command, ok := commands[path]
		if !ok {
			t.Fatalf("CommandSurface missing %q", path)
		}
		if got := command.Availability; got != wantAvailability {
			t.Fatalf("CommandSurface %q availability = %q, want %q", path, got, wantAvailability)
		}
	}
}

func TestCatalogStreamsMatchBundleSchemas(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), connectorName)
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	catalog, err := New().Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	streamsByName := make(map[string]connectors.Stream, len(catalog.Streams))
	for _, stream := range catalog.Streams {
		streamsByName[stream.Name] = stream
	}
	for _, name := range cloudTrailPublishedStreams {
		stream, ok := streamsByName[name]
		if !ok {
			t.Fatalf("catalog missing stream %s", name)
		}
		schema := bundle.Schemas[name]
		if schema == nil {
			t.Fatalf("bundle missing schema for stream %s", name)
		}
		gotFields := make([]string, 0, len(stream.Fields))
		for _, field := range stream.Fields {
			gotFields = append(gotFields, field.Name)
		}
		if want := schema.Properties(); !slices.Equal(gotFields, want) {
			t.Fatalf("catalog fields for %s = %v, want schema properties %v", name, gotFields, want)
		}
		if !slices.Equal(stream.PrimaryKey, schema.PrimaryKey) {
			t.Fatalf("catalog primary key for %s = %v, want %v", name, stream.PrimaryKey, schema.PrimaryKey)
		}
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
	if !caps.Check || !caps.Catalog || !caps.Read || !caps.Write {
		t.Fatalf("capabilities = %+v, want Check, Catalog, Read, and Write", caps)
	}
	if caps.Query {
		t.Fatalf("%s must not expose raw query capability: this project disables unrestricted query-text execution for every connector", wantName)
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
	if got, want := len(cat.Streams), len(cloudTrailPublishedStreams); got != want {
		t.Fatalf("Catalog streams = %d, want %d", got, want)
	}
}
