package awscloudtrail

import (
	"context"
	"os"
	"slices"
	"strings"
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
	cancel := commands["query cancel"]
	if cancel.Intent != "reverse_etl" || cancel.Write != "cancel_query" || strings.TrimSpace(cancel.Approval) == "" || !commandFlagRequired(cancel, "event-data-store") {
		t.Fatalf("query cancel command = %+v, want approval-gated cancel_query write", cancel)
	}
	resourcePolicy := commands["resource-policy set"]
	if len(resourcePolicy.Examples) != 1 || !strings.Contains(resourcePolicy.Examples[0], "\"Action\":\"cloudtrail:StartQuery\"") {
		t.Fatalf("resource-policy set examples = %q, want event-data-store StartQuery policy", resourcePolicy.Examples)
	}
	channelCreate := commands["channel create"]
	if !commandFlagRequired(channelCreate, "destination-location") || !commandFlagRequired(channelCreate, "destination-type") {
		t.Fatalf("channel create command = %+v, want required typed destination flags", channelCreate)
	}
	assertCommandFlagTargets(t, commands["events lookup"], map[string]string{
		"lookup-attribute-key":   "body.LookupAttributes.0.AttributeKey",
		"lookup-attribute-value": "body.LookupAttributes.0.AttributeValue",
	})
	assertCommandFlagTargets(t, channelCreate, map[string]string{
		"destination-location": "record.Destinations.[].Location",
		"destination-type":     "record.Destinations.[].Type",
	})
	assertCommandFlagTargets(t, commands["event-selectors set"], map[string]string{
		"event-selector-data-resource-type":              "record.EventSelectors.[].DataResources.[].Type",
		"event-selector-data-resource-value":             "record.EventSelectors.[].DataResources.[].Values",
		"event-selector-exclude-management-event-source": "record.EventSelectors.[].ExcludeManagementEventSources.[]",
	})
	assertCommandFlagTargets(t, commands["insight-selectors set"], map[string]string{
		"insight-event-category": "record.InsightSelectors.[].EventCategories.[]",
	})
	assertCommandFlagTargets(t, commands["import start"], map[string]string{
		"import-source-s3-bucket-access-role-arn": "record.ImportSource.S3.S3BucketAccessRoleArn",
		"import-source-s3-bucket-region":          "record.ImportSource.S3.S3BucketRegion",
		"import-source-s3-location-uri":           "record.ImportSource.S3.S3LocationUri",
	})
	for _, command := range surface.Commands {
		if command.Availability != "implemented" {
			continue
		}
		for _, example := range command.Examples {
			if strings.Contains(example, " value") {
				t.Fatalf("command %q example contains placeholder input %q", command.Path, example)
			}
			if command.Intent == "reverse_etl" && !strings.Contains(example, "--preview") {
				t.Fatalf("reverse ETL command %q example omits --preview: %q", command.Path, example)
			}
		}
	}
}

func commandFlagRequired(command connectors.CommandSurfaceCommand, name string) bool {
	for _, flag := range command.Flags {
		if flag.Name == name {
			return flag.Required
		}
	}
	return false
}

func assertCommandFlagTargets(t *testing.T, command connectors.CommandSurfaceCommand, want map[string]string) {
	t.Helper()
	got := make(map[string]string, len(command.Flags))
	for _, flag := range command.Flags {
		got[flag.Name] = flag.MapsTo
	}
	for name, target := range want {
		if got[name] != target {
			t.Fatalf("command %q flag --%s maps to %q, want %q", command.Path, name, got[name], target)
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
