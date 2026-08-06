package nativeset

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
)

func TestFactoriesExposeDefinitions(t *testing.T) {
	want := map[string]bool{
		"alpha-vantage":             false,
		"amazon-sqs":                false,
		"apify-dataset":             false,
		"ashby":                     false,
		"aws-cloudtrail":            false,
		"babelforce":                false,
		"basecamp":                  false,
		"bing-ads":                  false,
		"bunny-inc":                 false,
		"canny":                     false,
		"copper":                    false,
		"dixa":                      false,
		"dynamodb":                  false,
		"faker":                     false,
		"fastbill":                  false,
		"feishu":                    false,
		"free-agent-connector":      false,
		"freightview":               false,
		"google-analytics-data-api": false,
		"google-calendar":           false,
		"google-classroom":          false,
		"google-pagespeed-insights": false,
		"less-annoying-crm":         false,
		"lokalise":                  false,
		"mendeley":                  false,
		"mercado-ads":               false,
		"metabase":                  false,
		"mode":                      false,
		"my-hours":                  false,
		"pocket":                    false,
		"postgres":                  false,
		"prestashop":                false,
		"rootly":                    false,
		"safetyculture":             false,
		"tally-prime":               false,
		"yahoo-finance-price":       false,
	}

	for _, factory := range Factories() {
		if factory.New == nil {
			t.Fatalf("factory %q New = nil", factory.Name)
		}
		c := factory.New()
		if c.Name() != factory.Name {
			t.Fatalf("factory %q New().Name() = %q", factory.Name, c.Name())
		}
		def, ok := connectors.DefinitionOf(c)
		if !ok {
			t.Fatalf("factory %q connector does not implement DefinitionProvider", factory.Name)
		}
		if def.Name != factory.Name {
			t.Fatalf("factory %q Definition().Name = %q", factory.Name, def.Name)
		}
		if _, tracked := want[factory.Name]; tracked {
			want[factory.Name] = true
		}
	}

	for name, seen := range want {
		if !seen {
			t.Fatalf("Factories() missing %q", name)
		}
	}
}

func TestBundleDefinitionForwardsNativeCommandSurface(t *testing.T) {
	var cloudTrail connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "aws-cloudtrail" {
			cloudTrail = factory.New()
			break
		}
	}
	if cloudTrail == nil {
		t.Fatal("aws-cloudtrail factory not found")
	}
	provider, ok := cloudTrail.(connectors.CommandSurfaceProvider)
	if !ok {
		t.Fatal("aws-cloudtrail factory does not implement connectors.CommandSurfaceProvider")
	}
	surface := provider.CommandSurface()
	if surface == nil {
		t.Fatal("aws-cloudtrail CommandSurface() = nil")
	}
	if got, want := len(surface.Commands), 60; got != want {
		t.Fatalf("aws-cloudtrail command rows = %d, want %d", got, want)
	}
}

func TestCloudTrailBundleDefinitionPreservesActualRuntimeCapabilities(t *testing.T) {
	var cloudTrail connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "aws-cloudtrail" {
			cloudTrail = factory.New()
			break
		}
	}
	if cloudTrail == nil {
		t.Fatal("aws-cloudtrail factory not found")
	}
	if _, ok := cloudTrail.(connectors.CDCReader); ok {
		t.Fatal("aws-cloudtrail factory unexpectedly implements connectors.CDCReader")
	}
	for _, capability := range []struct {
		name string
		ok   bool
	}{
		{name: "CommandSurfaceProvider", ok: implementsCommandSurface(cloudTrail)},
		{name: "OperationDirectReader", ok: implementsOperationDirectRead(cloudTrail)},
		{name: "WriteValidator", ok: implementsWriteValidator(cloudTrail)},
		{name: "DryRunWriter", ok: implementsDryRunWriter(cloudTrail)},
		{name: "StatefulReader", ok: implementsStatefulReader(cloudTrail)},
	} {
		if !capability.ok {
			t.Fatalf("aws-cloudtrail factory does not implement connectors.%s", capability.name)
		}
	}
	manifest, ok := cloudTrail.(connectors.ManifestProvider)
	if !ok {
		t.Fatal("aws-cloudtrail factory does not implement connectors.ManifestProvider")
	}
	if got, want := len(manifest.Manifest().WriteActions), 30; got != want {
		t.Fatalf("aws-cloudtrail manifest write actions = %d, want %d", got, want)
	}
	if _, err := commandrunner.BuildWriteCommand(t.Context(), cloudTrail, commandrunner.Request{
		Path:  []string{"query", "cancel"},
		Flags: map[string][]string{"query-id": {"11111111-1111-1111-1111-111111111111"}},
	}); err == nil || !strings.Contains(err.Error(), "missing required flag --event-data-store") {
		t.Fatalf("BuildWriteCommand(query cancel without event data store) error = %v, want required flag error", err)
	}
	plan, err := commandrunner.BuildWriteCommand(t.Context(), cloudTrail, commandrunner.Request{
		Path:    []string{"query", "cancel"},
		Flags:   map[string][]string{"event-data-store": {"arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example"}, "query-id": {"11111111-1111-1111-1111-111111111111"}},
		Preview: true,
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand(query cancel): %v", err)
	}
	if !plan.ApprovalRequired || plan.Preview == nil {
		t.Fatalf("query cancel plan = %+v, want approval-gated preview", plan)
	}
	_, err = commandrunner.BuildWriteCommand(t.Context(), cloudTrail, commandrunner.Request{
		Path: []string{"import", "start"},
		Flags: map[string][]string{
			"import-id":        {"11111111-1111-1111-1111-111111111111"},
			"start-event-time": {"2026-01-02T00:00:00Z"},
			"end-event-time":   {"2026-01-01T00:00:00Z"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "invalid import time range") {
		t.Fatalf("BuildWriteCommand(import start) error = %v, want time-range validation", err)
	}
	_, err = commandrunner.BuildWriteCommand(t.Context(), cloudTrail, commandrunner.Request{
		Path: []string{"import", "start"},
		Flags: map[string][]string{
			"import-id":    {"11111111-1111-1111-1111-111111111111"},
			"destinations": {"arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/import-destination"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "cannot combine ImportId with Destinations") {
		t.Fatalf("BuildWriteCommand(import start mixed modes) error = %v, want mode validation", err)
	}
}

func TestCloudTrailWriteExamplesBuildApprovalGatedPreviews(t *testing.T) {
	var cloudTrail connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "aws-cloudtrail" {
			cloudTrail = factory.New()
			break
		}
	}
	if cloudTrail == nil {
		t.Fatal("aws-cloudtrail factory not found")
	}
	provider, ok := cloudTrail.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("aws-cloudtrail factory does not expose a command surface")
	}
	for _, command := range provider.CommandSurface().Commands {
		if command.Intent != "reverse_etl" || command.Availability != "implemented" {
			continue
		}
		for _, example := range command.Examples {
			t.Run(command.Path, func(t *testing.T) {
				req := writeRequestFromExample(t, command, example)
				plan, err := commandrunner.BuildWriteCommand(t.Context(), cloudTrail, req)
				if err != nil {
					t.Fatalf("BuildWriteCommand(%q): %v", example, err)
				}
				if !plan.ApprovalRequired || plan.Preview == nil {
					t.Fatalf("plan = %+v, want approval-gated preview", plan)
				}
			})
		}
	}
}

func TestCloudTrailNestedWriteFlagsBuildClosedRecords(t *testing.T) {
	var cloudTrail connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "aws-cloudtrail" {
			cloudTrail = factory.New()
			break
		}
	}
	if cloudTrail == nil {
		t.Fatal("aws-cloudtrail factory not found")
	}
	for _, test := range []struct {
		name  string
		path  []string
		flags map[string][]string
	}{
		{
			name: "advanced event selector",
			path: []string{"event-selectors", "set"},
			flags: map[string][]string{
				"trail-name":            {"example-trail"},
				"advanced-event-field":  {"eventCategory", "resources.type"},
				"advanced-event-equals": {"Data", "AWS::S3::Object"},
			},
		},
		{
			name: "context key selector",
			path: []string{"event-configuration", "set"},
			flags: map[string][]string{
				"event-data-store":   {"arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example"},
				"max-event-size":     {"Large"},
				"context-key-type":   {"RequestContext", "TagContext"},
				"context-key-equals": {"aws:PrincipalArn", "Environment"},
			},
		},
		{
			name: "disable insights",
			path: []string{"insight-selectors", "set"},
			flags: map[string][]string{
				"trail-name":       {"example-trail"},
				"disable-insights": {"true"},
			},
		},
		{
			name: "dashboard query parameters",
			path: []string{"dashboard", "refresh"},
			flags: map[string][]string{
				"dashboard-id":        {"AWSCloudTrail-Overview"},
				"query-start-time":    {"2024-11-13T08:00:00Z"},
				"query-end-time":      {"2024-11-13T12:00:00Z"},
				"query-period":        {"minute"},
				"event-data-store-id": {"example-event-store"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan, err := commandrunner.BuildWriteCommand(t.Context(), cloudTrail, commandrunner.Request{Path: test.path, Flags: test.flags, Preview: true})
			if err != nil {
				t.Fatalf("BuildWriteCommand(%q): %v", test.name, err)
			}
			if !plan.ApprovalRequired || plan.Preview == nil {
				t.Fatalf("plan = %+v, want approval-gated preview", plan)
			}
		})
	}
	_, err := commandrunner.BuildWriteCommand(t.Context(), cloudTrail, commandrunner.Request{
		Path:    []string{"tags", "add"},
		Flags:   map[string][]string{"resource-id": {"arn:aws:cloudtrail:us-east-1:123456789012:trail/example"}, "tag-key": {"Environment"}, "tags-list": {"[]"}},
		Preview: true,
	})
	if err == nil || !strings.Contains(err.Error(), "unknown flag --tags-list") {
		t.Fatalf("BuildWriteCommand(tags add raw tags) error = %v, want closed flag error", err)
	}
}

func writeRequestFromExample(t *testing.T, command connectors.CommandSurfaceCommand, example string) commandrunner.Request {
	t.Helper()
	tokens := strings.Fields(example)
	path := strings.Fields(command.Path)
	if len(tokens) < len(path)+2 || tokens[0] != "pm" || tokens[1] != "aws-cloudtrail" {
		t.Fatalf("invalid example prefix %q", example)
	}
	for index, part := range path {
		if tokens[index+2] != part {
			t.Fatalf("example %q path = %q, want %q", example, tokens[2:2+len(path)], path)
		}
	}
	req := commandrunner.Request{Path: path, Flags: map[string][]string{}}
	for index := len(path) + 2; index < len(tokens); {
		token := tokens[index]
		if !strings.HasPrefix(token, "--") {
			t.Fatalf("example %q has unexpected token %q", example, token)
		}
		name := strings.TrimPrefix(token, "--")
		if name == "preview" {
			req.Preview = true
			index++
			continue
		}
		if index+1 >= len(tokens) {
			t.Fatalf("example %q omits a value for --%s", example, name)
		}
		if name != "credential" {
			req.Flags[name] = append(req.Flags[name], strings.Trim(tokens[index+1], "'"))
		}
		index += 2
	}
	if !req.Preview {
		t.Fatalf("example %q omits --preview", example)
	}
	return req
}

func implementsCommandSurface(c connectors.Connector) bool {
	_, ok := c.(connectors.CommandSurfaceProvider)
	return ok
}

func implementsOperationDirectRead(c connectors.Connector) bool {
	_, ok := c.(connectors.OperationDirectReader)
	return ok
}

func implementsWriteValidator(c connectors.Connector) bool {
	_, ok := c.(connectors.WriteValidator)
	return ok
}

func implementsDryRunWriter(c connectors.Connector) bool {
	_, ok := c.(connectors.DryRunWriter)
	return ok
}

func implementsStatefulReader(c connectors.Connector) bool {
	_, ok := c.(connectors.StatefulReader)
	return ok
}
