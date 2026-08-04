package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/defs"
)

func validMetadata(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false }
	}`
}

func dynamicSchemaMetadata(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "database",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": true }
	}`
}

const validSpec = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["base_url"],
	"properties": {
		"base_url": { "type": "string" },
		"token": { "type": "string", "x-secret": true }
	}
}`

const validStreams = `{
	"base": {
		"url": "{{ config.base_url }}",
		"user_agent": "test-agent",
		"headers": {},
		"auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ cursor }}" } ],
		"pagination": { "type": "none" },
		"check": { "method": "GET", "path": "/ping" },
		"error_map": []
	},
	"streams": [
		{
			"name": "widgets",
			"path": "/widgets",
			"records": { "path": "data" },
			"schema": "schemas/widgets.json"
		}
	]
}`

const validWidgetsSchema = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"x-primary-key": ["id"],
	"x-cursor-field": "updated_at",
	"properties": {
		"id": { "type": "integer" },
		"updated_at": { "type": "string" }
	}
}`

const validAPISurface = `{
	"api": "test API v1",
	"endpoints": [
		{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } }
	]
}`

const validDocs = `# Overview

test

## Auth setup

none

## Streams notes

none

## Write actions & risks

none

## Known limits

none
`

func fullValidBundleFS(name string) fstest.MapFS {
	return fstest.MapFS{
		name + "/metadata.json":                        &fstest.MapFile{Data: []byte(validMetadata(name))},
		name + "/spec.json":                            &fstest.MapFile{Data: []byte(validSpec)},
		name + "/streams.json":                         &fstest.MapFile{Data: []byte(validStreams)},
		name + "/api_surface.json":                     &fstest.MapFile{Data: []byte(validAPISurface)},
		name + "/schemas/widgets.json":                 &fstest.MapFile{Data: []byte(validWidgetsSchema)},
		name + "/docs.md":                              &fstest.MapFile{Data: []byte(validDocs)},
		name + "/fixtures/streams/widgets/page_1.json": &fstest.MapFile{Data: []byte(`{"request":{"method":"GET","path":"/widgets","query":{}},"response":{"status":200,"body":{"data":[]}}}`)},
	}
}

func TestBundleLoadHappyPathFullBundle(t *testing.T) {
	fsys := fullValidBundleFS("acme")

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Name != "acme" {
		t.Fatalf("Name = %q", b.Name)
	}
	if b.Metadata.Name != "acme" {
		t.Fatalf("Metadata.Name = %q", b.Metadata.Name)
	}
	if b.Spec == nil {
		t.Fatalf("Spec not compiled")
	}
	if b.HTTP.URL != "{{ config.base_url }}" {
		t.Fatalf("HTTP.URL = %q", b.HTTP.URL)
	}
	if len(b.Streams) != 1 || b.Streams[0].Name != "widgets" {
		t.Fatalf("Streams = %+v", b.Streams)
	}
	if b.Writes != nil {
		t.Fatalf("Writes should be nil when writes.json absent, got %+v", b.Writes)
	}
	sch, ok := b.Schemas["widgets"]
	if !ok {
		t.Fatalf("Schemas missing widgets entry")
	}
	if len(sch.PrimaryKey) != 1 || sch.PrimaryKey[0] != "id" {
		t.Fatalf("PrimaryKey = %v", sch.PrimaryKey)
	}
	if sch.CursorField != "updated_at" {
		t.Fatalf("CursorField = %q", sch.CursorField)
	}
	if b.Surface == nil {
		t.Fatalf("Surface not parsed")
	}
	if b.Docs == "" {
		t.Fatalf("Docs not loaded")
	}
	if b.Fixtures == nil {
		t.Fatalf("Fixtures should be non-nil when fixtures/ present")
	}
}

func TestBundleLoadOptionalFilesAbsent(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/fixtures/streams/widgets/page_1.json")

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Writes != nil {
		t.Fatalf("Writes should be nil when writes.json absent")
	}
	if b.Fixtures != nil {
		t.Fatalf("Fixtures should be nil when fixtures/ absent")
	}
	if b.CLISurface != nil {
		t.Fatalf("CLISurface should be nil when cli_surface.json is absent")
	}
	if b.Certification != nil {
		t.Fatalf("Certification should be nil when certification.json is absent")
	}
}

func TestBundleLoadParsesCertification(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/certification.json"] = &fstest.MapFile{Data: []byte(`{
		"schema_version": 1,
		"source": {
			"default_stream": "widgets",
			"source_credential_defaults": {"base_url": "https://api.example.test"},
			"live_unavailable": [{"kind": "Error", "contains": ["status 403"]}]
		},
		"direct_read_candidates": [{
			"stage_name": "direct_read_sweep_widget",
			"command": "widget get",
			"args": [
				{"connector": true},
				{"literal": "widget"},
				{"config_key": "widget_id", "default": "fixture-widget"},
				{"source_credential": true}
			]
		}]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Certification == nil {
		t.Fatalf("Certification is nil")
	}
	if b.Certification.Source.DefaultStream != "widgets" {
		t.Fatalf("default_stream = %q", b.Certification.Source.DefaultStream)
	}
	if got := b.Certification.Source.SourceCredentialDefaults["base_url"]; got != "https://api.example.test" {
		t.Fatalf("source_credential_defaults.base_url = %q", got)
	}
	if len(b.Certification.DirectReadCandidates) != 1 {
		t.Fatalf("DirectReadCandidates = %+v", b.Certification.DirectReadCandidates)
	}
}

func TestBundleLoadRejectsUnknownCertificationKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/certification.json"] = &fstest.MapFile{Data: []byte(`{"schema_version":1,"surprise":true}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected unknown certification key to fail")
	}
	if !strings.Contains(err.Error(), "certification.json") || !strings.Contains(err.Error(), "surprise") {
		t.Fatalf("Load error = %q, want certification.json surprise rejection", err.Error())
	}
}

func TestBundleLoadRejectsCertificationUnknownStream(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/certification.json"] = &fstest.MapFile{Data: []byte(`{"schema_version":1,"source":{"default_stream":"missing"}}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected unknown default stream to fail")
	}
	if !strings.Contains(err.Error(), "certification.json") || !strings.Contains(err.Error(), "default_stream") {
		t.Fatalf("Load error = %q, want default_stream rejection", err.Error())
	}
}

func TestBundleLoadEmbeddedGitHubCertification(t *testing.T) {
	b, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	if b.Certification == nil {
		t.Fatalf("GitHub Certification is nil; defs.FS must embed certification.json")
	}
	if b.Certification.Source.DefaultStream != "issues" {
		t.Fatalf("GitHub certification default stream = %q", b.Certification.Source.DefaultStream)
	}
	if len(b.Certification.WritePairings) != 3 {
		t.Fatalf("GitHub certification write pairings = %d, want 3", len(b.Certification.WritePairings))
	}
}

func TestBundleLoadParsesGraphQLStreamAndWriteAction(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"pagination": { "type": "none" }
		},
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query ListWidgets($first: Int!) { widgets(first: $first) { nodes { id } } }",
					"operation_name": "ListWidgets",
					"variables": { "first": { "template": "{{ config.page_size }}", "type": "integer" } }
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [
			{
				"name": "delete_widget",
				"kind": "delete",
				"method": "POST",
				"path": "/graphql",
				"body_type": "graphql",
				"graphql": {
					"document": "mutation DeleteWidget($id: ID!) { deleteWidget(input: {id: $id}) { clientMutationId } }",
					"operation_name": "DeleteWidget",
					"variables": { "id": "{{ record.id }}" }
				},
				"record_schema": {
					"type": "object",
					"required": ["id"],
					"properties": { "id": { "type": "string" } }
				},
				"risk": "delete"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Streams[0].GraphQL == nil || b.Streams[0].GraphQL.OperationName != "ListWidgets" {
		t.Fatalf("stream GraphQL = %+v, want ListWidgets", b.Streams[0].GraphQL)
	}
	if len(b.Writes) != 1 || b.Writes[0].GraphQL == nil || b.Writes[0].GraphQL.OperationName != "DeleteWidget" {
		t.Fatalf("write GraphQL = %+v, want DeleteWidget", b.Writes)
	}
}

func TestBundleLoadRejectsGraphQLWriteWithoutGraphQLBlock(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [
			{
				"name": "delete_widget",
				"kind": "delete",
				"method": "POST",
				"path": "/graphql",
				"body_type": "graphql",
				"record_schema": { "type": "object" },
				"risk": "delete"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected body_type graphql without graphql block to fail")
	}
	if !strings.Contains(err.Error(), "writes.json") || !strings.Contains(err.Error(), "body_type graphql requires graphql") {
		t.Fatalf("Load error = %q, want graphql block requirement", err.Error())
	}
}

func TestBundleLoadRejectsTemplatedGraphQLDocument(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": { "url": "{{ config.base_url }}" },
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query {{ config.operation }} { widgets { nodes { id } } }",
					"operation_name": "ListWidgets"
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected templated GraphQL document to fail")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "fixed bundle metadata") {
		t.Fatalf("Load error = %q, want fixed document rejection", err.Error())
	}
}

func TestBundleLoadRejectsGraphQLWriteQueryDocument(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [
			{
				"name": "delete_widget",
				"kind": "delete",
				"method": "POST",
				"path": "/graphql",
				"body_type": "graphql",
				"graphql": {
					"document": "query DeleteWidget($id: ID!) { node(id: $id) { id } }",
					"operation_name": "DeleteWidget",
					"variables": { "id": "{{ record.id }}" }
				},
				"record_schema": { "type": "object" },
				"risk": "delete"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected query document in write action to fail")
	}
	if !strings.Contains(err.Error(), "writes.json") || !strings.Contains(err.Error(), "must start with mutation") {
		t.Fatalf("Load error = %q, want mutation document rejection", err.Error())
	}
}

func TestBundleLoadRejectsGraphQLVariableUnsupportedType(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": { "url": "{{ config.base_url }}" },
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query ListWidgets($first: Int!) { widgets(first: $first) { nodes { id } } }",
					"operation_name": "ListWidgets",
					"variables": {
						"first": { "template": "{{ config.page_size }}", "type": "int" }
					}
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected unsupported GraphQL variable type to fail")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("Load error = %q, want unsupported type rejection", err.Error())
	}
}

func TestBundleLoadParsesGraphQLVariableOmitWhenEmpty(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": { "url": "{{ config.base_url }}" },
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query ListWidgets($after: String) { widgets(after: $after) { nodes { id } } }",
					"operation_name": "ListWidgets",
					"variables": {
						"after": { "template": "{{ cursor }}", "omit_when_empty": true },
						"owner": { "template": "{{ query.owner }}", "default": "octocat" }
					}
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	if _, err := Load(fsys, "acme"); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestBundleLoadRejectsGraphQLVariableDefaultNonString(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": { "url": "{{ config.base_url }}" },
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query ListWidgets($owner: String!) { widgets(owner: $owner) { nodes { id } } }",
					"operation_name": "ListWidgets",
					"variables": {
						"owner": { "template": "{{ query.owner }}", "default": 42 }
					}
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected non-string default to fail")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "default must be a string") {
		t.Fatalf("Load error = %q, want default string rejection", err.Error())
	}
}

func TestBundleLoadRejectsGraphQLVariableOmitWhenEmptyNonBoolean(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": { "url": "{{ config.base_url }}" },
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query ListWidgets($after: String) { widgets(after: $after) { nodes { id } } }",
					"operation_name": "ListWidgets",
					"variables": {
						"after": { "template": "{{ cursor }}", "omit_when_empty": "yes" }
					}
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected non-boolean omit_when_empty to fail")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "omit_when_empty must be a boolean") {
		t.Fatalf("Load error = %q, want omit_when_empty boolean rejection", err.Error())
	}
}

func TestBundleLoadRejectsGraphQLVariableDefaultTypeMismatch(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": { "url": "{{ config.base_url }}" },
		"streams": [
			{
				"name": "widgets",
				"method": "POST",
				"path": "/graphql",
				"graphql": {
					"document": "query ListWidgets($count: Int!) { widgets(count: $count) { nodes { id } } }",
					"operation_name": "ListWidgets",
					"variables": {
						"count": { "template": "{{ query.count }}", "type": "integer", "default": "not-a-number" }
					}
				},
				"records": { "path": "data.widgets.nodes" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected default/type mismatch to fail")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "default") {
		t.Fatalf("Load error = %q, want default/type mismatch rejection", err.Error())
	}
}

func TestGitHubProjectsDiscussionsCommandsMapToGraphQLStreams(t *testing.T) {
	b, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load github: %v", err)
	}

	streams := map[string]StreamSpec{}
	for _, stream := range b.Streams {
		streams[stream.Name] = stream
	}
	for _, name := range []string{"projects", "project_items", "discussions", "discussion"} {
		stream, ok := streams[name]
		if !ok {
			t.Fatalf("github stream %q missing", name)
		}
		if stream.GraphQL == nil {
			t.Fatalf("github stream %q GraphQL = nil, want fixed GraphQL document", name)
		}
		if stream.Method != "POST" || stream.Path != "/graphql" {
			t.Fatalf("github stream %q method/path = %s %s, want POST /graphql", name, stream.Method, stream.Path)
		}
		if stream.SchemaRef == "" {
			t.Fatalf("github stream %q missing schema ref", name)
		}
	}

	if b.CLISurface == nil {
		t.Fatalf("github cli surface missing")
	}
	want := map[string]string{
		"project list":      "projects",
		"project item-list": "project_items",
		"discussion list":   "discussions",
		"discussion view":   "discussion",
	}
	for _, cmd := range b.CLISurface.Commands {
		stream, ok := want[cmd.Path]
		if !ok {
			continue
		}
		if cmd.Intent != "etl" || cmd.Availability != "implemented" || cmd.Stream != stream || cmd.Operation != "" {
			t.Fatalf("command %q = intent=%q availability=%q stream=%q operation=%q, want implemented etl stream %q with no operation",
				cmd.Path, cmd.Intent, cmd.Availability, cmd.Stream, cmd.Operation, stream)
		}
		delete(want, cmd.Path)
	}
	if len(want) > 0 {
		t.Fatalf("missing GitHub CLI commands: %v", want)
	}
}

func TestBundleLoadParsesCLISurface(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/cli_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"tagline": "Work with Acme from the command line.",
		"usage": "pm acme <command> [flags]",
		"source_cli": {
			"name": "acmectl",
			"docs": "https://example.com/acmectl",
			"reference": "https://example.com/acmectl/reference"
		},
		"groups": [
			{ "id": "core", "title": "Core Commands", "commands": ["widget"] }
		],
		"global_flags": [
			{ "name": "json", "type": "boolean", "summary": "Write machine-readable JSON output." }
		],
		"commands": [
			{
				"path": "widget list",
				"summary": "List widgets",
				"intent": "etl",
				"availability": "implemented",
				"stream": "widgets",
				"source_cli_path": "acmectl widget list",
				"flags": [
					{ "name": "state", "type": "string", "summary": "Filter by state.", "maps_to": "query.state" }
				],
				"examples": ["pm acme widget list --json"],
				"api_surface": [
					{ "method": "GET", "path": "/widgets" }
				]
			}
		],
		"help_topics": [
			{ "name": "authentication", "summary": "Credential setup and supported auth modes." }
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.CLISurface == nil {
		t.Fatalf("CLISurface is nil")
	}
	if b.CLISurface.Tagline != "Work with Acme from the command line." {
		t.Fatalf("Tagline = %q", b.CLISurface.Tagline)
	}
	if len(b.CLISurface.Commands) != 1 || b.CLISurface.Commands[0].Path != "widget list" {
		t.Fatalf("Commands = %+v", b.CLISurface.Commands)
	}
	if b.CLISurface.Commands[0].Stream != "widgets" {
		t.Fatalf("Command stream = %q", b.CLISurface.Commands[0].Stream)
	}
	if len(b.RawCLISurface) == 0 || !strings.Contains(string(b.RawCLISurface), `"widget list"`) {
		t.Fatalf("RawCLISurface = %q, want verbatim cli_surface.json bytes", string(b.RawCLISurface))
	}
}

func TestBundleLoadParsesOperations(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.widgets.get",
				"kind": "rest_read",
				"summary": "Read one widget",
				"risk": "low",
				"approval": "none",
				"output_policy": "json",
				"request_contract": {
					"source_tier": 3,
					"source_url": "https://example.invalid/widgets#get",
					"source_location": "Get widget request",
					"fields": [{"path":"path.id","source_url":"https://example.invalid/widgets#get","source_location":"path parameter id"}]
				},
				"rest": {
					"method": "GET",
					"path": "/widgets/{id}"
				}
			}
		]
	}`)}
	fsys["acme/cli_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"tagline": "Work with Acme from the command line.",
		"usage": "pm acme <command> [flags]",
		"commands": [
			{
				"path": "widget view",
				"summary": "View a widget",
				"intent": "direct_read",
				"availability": "implemented",
				"operation": "acme.widgets.get"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Operations) != 1 {
		t.Fatalf("Operations = %d, want 1", len(b.Operations))
	}
	op := b.Operations[0]
	if op.ID != "acme.widgets.get" || op.Kind != "rest_read" || op.REST == nil || op.REST.Path != "/widgets/{id}" {
		t.Fatalf("operation = %+v, want parsed rest_read operation", op)
	}
	if b.CLISurface.Commands[0].Operation != "acme.widgets.get" {
		t.Fatalf("command operation = %q, want acme.widgets.get", b.CLISurface.Commands[0].Operation)
	}
}

func TestBundleLoadParsesRequestContractAndNoBody(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.widgets.refresh",
				"kind": "rest_write",
				"summary": "Refresh one widget",
				"risk": "medium",
				"approval": "plan, preview, approval, execute",
				"output_policy": "json_redacted",
				"mutation_class": "update",
				"request_contract": {
					"source_tier": 3,
					"source_url": "https://example.invalid/widgets#refresh",
					"source_location": "Refresh widget request",
					"fields": [
						{
							"path": "path.id",
							"source_url": "https://example.invalid/widgets#refresh",
							"source_location": "Refresh widget path parameter id"
						}
					]
				},
				"rest": {
					"method": "POST",
					"path": "/widgets/{id}:refresh",
					"body": "none"
				}
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	op := b.Operations[0]
	if op.RequestContract == nil || op.RequestContract.SourceTier != 3 {
		t.Fatalf("request contract = %+v, want source tier 3", op.RequestContract)
	}
	if op.REST == nil || op.REST.Body == nil || !op.REST.Body.None {
		t.Fatalf("rest body = %+v, want explicit none", op.REST)
	}
	raw, err := json.Marshal(op.REST.Body)
	if err != nil {
		t.Fatalf("Marshal body: %v", err)
	}
	if string(raw) != `"none"` {
		t.Fatalf("marshaled body = %s, want %q", raw, "none")
	}
}

func TestBundleLoadRequiresRequestContractForRESTOperations(t *testing.T) {
	_, err := Load(operationsBundleFS(t, `{
		"operations": [{
			"id": "acme.widgets.list",
			"kind": "rest_read",
			"summary": "List widgets",
			"risk": "low",
			"approval": "none",
			"output_policy": "json_redacted",
			"rest": {"method": "GET", "path": "/widgets", "max_bytes": 1024, "body": "none"}
		}]
	}`), "acme")
	if err == nil || !strings.Contains(err.Error(), "REST operation must declare request_contract evidence") {
		t.Fatalf("Load error = %v, want mandatory request_contract error", err)
	}
}

func TestBundleLoadRequiresExplicitNoBodyDeclaration(t *testing.T) {
	for _, body := range []string{"", `,"body":{}`} {
		t.Run(fmt.Sprintf("body=%q", body), func(t *testing.T) {
			operations := fmt.Sprintf(`{
				"operations": [{
					"id":"acme.widgets.get",
					"kind":"rest_read",
					"summary":"Get widget",
					"risk":"low",
					"approval":"none",
					"output_policy":"json_redacted",
					"request_contract":{"source_tier":3,"source_url":"https://example.invalid/widgets#get","source_location":"Get widget","fields":[]},
					"rest":{"method":"GET","path":"/widgets","max_bytes":1024%s}
				}]
			}`, body)
			_, err := Load(operationsBundleFS(t, operations), "acme")
			if err == nil || !strings.Contains(err.Error(), `must declare body "none"`) {
				t.Fatalf("Load error = %v, want explicit body none error", err)
			}
		})
	}
}

func TestBundleLoadRequiresRequiredQueryCitations(t *testing.T) {
	_, err := Load(operationsBundleFS(t, `{
		"operations": [{
			"id": "acme.widgets.list",
			"kind": "rest_read",
			"summary": "List widgets",
			"risk": "low",
			"approval": "none",
			"output_policy": "json_redacted",
			"request_contract": {
				"source_tier": 3,
				"source_url": "https://example.invalid/widgets#list",
				"source_location": "List widgets request",
				"fields": []
			},
			"rest": {
				"method": "GET",
				"path": "/widgets",
				"max_bytes": 1024,
				"body": "none",
				"required_query": [{"any_of": ["email", "id"]}]
			}
		}]
	}`), "acme")
	if err == nil || !strings.Contains(err.Error(), `missing citation for "query.email"`) {
		t.Fatalf("Load error = %v, want required_query citation error", err)
	}
}

func TestBundleLoadValidatesWriteActionClaims(t *testing.T) {
	const validFieldMap = `"path.id":"path.widget-id","query.notify":"query.notify","body.name":"body.name"`
	contract := func(id, writeAction, fieldMap string) string {
		return fmt.Sprintf(`{
			"id": %q,
			"kind": "rest_write",
			"summary": "Create widget",
			"risk": "medium",
			"approval": "required",
			"output_policy": "json_redacted",
			"mutation_class": "create",
			"request_contract": {
				"source_tier": 3,
				"source_url": "https://example.invalid/widgets#create",
				"source_location": "Create widget request",
				"write_action": %q,
				"write_field_map": {%s},
				"fields": [
					{"path":"path.widget-id","source_url":"https://example.invalid/widgets#create","source_location":"widget-id path parameter"},
					{"path":"query.notify","source_url":"https://example.invalid/widgets#create","source_location":"notify query parameter"},
					{"path":"body.name","source_url":"https://example.invalid/widgets#create","source_location":"name body property"}
				]
			},
			"rest": {
				"method": "POST",
				"path": "/widgets/{widget-id}",
				"query": {"notify":"true"},
				"body_schema": {"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"}}}
			}
		}`, id, writeAction, fieldMap)
	}
	tests := []struct {
		name       string
		operations string
		want       string
		pathFields string
	}{
		{name: "valid", operations: contract("acme.widgets.create", "create_widget", validFieldMap)},
		{name: "unclaimed", operations: "", want: `writes.json action "create_widget" must be claimed`},
		{name: "dangling", operations: contract("acme.widgets.create", "missing_widget", validFieldMap), want: `does not name a writes.json action`},
		{name: "duplicate", operations: contract("acme.widgets.create", "create_widget", validFieldMap) + "," + contract("acme.widgets.create_again", "create_widget", validFieldMap), want: `both claim writes.json action "create_widget"`},
		{name: "missing path mapping", operations: contract("acme.widgets.create", "create_widget", `"query.notify":"query.notify","body.name":"body.name"`), want: `missing write_field_map entry for "path.id"`},
		{name: "uncited mapping target", operations: contract("acme.widgets.create", "create_widget", `"path.id":"path.widget-id","query.notify":"query.notify","body.name":"body.unlisted"`), want: `maps "body.name" to uncited request field "body.unlisted"`},
		{name: "cross namespace mapping", operations: contract("acme.widgets.create", "create_widget", `"path.id":"body.name","query.notify":"query.notify","body.name":"body.name"`), want: `maps "path.id" across namespaces to "body.name"`},
		{name: "stale mapping", operations: contract("acme.widgets.create", "create_widget", validFieldMap+`,"body.extra":"body.name"`), want: `write_field_map entry "body.extra" does not match a write input`},
		{name: "undeclared path input", operations: contract("acme.widgets.create", "create_widget", validFieldMap), want: `path template record field "id" is missing from path_fields`, pathFields: `[]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathFields := tt.pathFields
			if pathFields == "" {
				pathFields = `["id"]`
			}
			fsys := fullValidBundleFS("acme")
			fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{
				"actions": [{
					"name": "create_widget",
					"kind": "create",
					"method": "POST",
					"path": "/widgets/{{ record.id }}",
					"path_fields": %s,
					"query": {"notify":"{{ record.notify }}"},
					"body_fields": ["name"],
					"record_schema": {"type":"object","required":["id","notify","name"],"additionalProperties":false,"properties":{"id":{"type":"string"},"notify":{"type":"string"},"name":{"type":"string"}}},
					"risk": "creates a widget"
				}]
			}`, pathFields))}
			if tt.operations != "" {
				fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{"operations":[` + tt.operations + `]}`)}
			}
			bundle, err := Load(fsys, "acme")
			if tt.want == "" {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if got := bundle.Operations[0].RequestContract.WriteAction; got != "create_widget" {
					t.Fatalf("write_action = %q, want create_widget", got)
				}
				if got := bundle.Operations[0].RequestContract.WriteFieldMap["path.id"]; got != "path.widget-id" {
					t.Fatalf("write_field_map[path.id] = %q, want path.widget-id", got)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestBundleLoadResolvesRequestBodySchemaReferencesAndComposition(t *testing.T) {
	fsys := operationsBundleFS(t, `{
		"operations": [{
			"id": "acme.widgets.create",
			"kind": "rest_write",
			"summary": "Create widget",
			"risk": "medium",
			"approval": "required",
			"output_policy": "json_redacted",
			"mutation_class": "create",
			"request_contract": {
				"source_tier": 1,
				"source_url": "https://example.invalid/openapi.json",
				"source_location": "paths./widgets.post.requestBody",
				"fields": [
					{"path":"body.name","source_url":"https://example.invalid/openapi.json","source_location":"Widget.properties.name"},
					{"path":"body.metadata","source_url":"https://example.invalid/openapi.json","source_location":"Widget.properties.metadata"},
					{"path":"body.metadata.label","source_url":"https://example.invalid/openapi.json","source_location":"Metadata.properties.label"}
				]
			},
			"rest": {
				"method": "POST",
				"path": "/widgets",
				"body_schema": {
					"$defs": {
						"Widget": {
							"type":"object",
							"additionalProperties":false,
							"properties":{"name":{"type":"string"},"metadata":{"$ref":"#/$defs/Metadata"}},
							"allOf":[{"required":["name"]}]
						},
						"Metadata": {"type":"object","additionalProperties":false,"properties":{"label":{"type":"string"}}}
					},
					"$ref": "#/$defs/Widget"
				}
			}
		}]
	}`)
	bundle, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if strings.Contains(string(bundle.Operations[0].REST.BodySchema), `$ref`) {
		t.Fatalf("inlined body_schema still contains $ref: %s", bundle.Operations[0].REST.BodySchema)
	}
	schema, err := CompileSchema(bundle.Operations[0].REST.BodySchema)
	if err != nil {
		t.Fatalf("CompileSchema(inlined body_schema): %v", err)
	}
	if err := schema.Validate(map[string]any{"name": "widget", "metadata": map[string]any{"label": "primary"}}); err != nil {
		t.Fatalf("Validate(inlined body_schema): %v", err)
	}
	if err := schema.Validate(map[string]any{"name": 42}); err == nil {
		t.Fatal("Validate(inlined body_schema) error = nil, want allOf property type error")
	}
}

func TestBundleLoadRejectsUnresolvedOrUnenumerableRequestBodySchema(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		want   string
	}{
		{name: "unresolved local reference", schema: `{"$ref":"#/components/schemas/Missing"}`, want: `cannot resolve local reference`},
		{name: "external reference", schema: `{"$ref":"https://example.invalid/openapi.json#/components/schemas/Widget"}`, want: `external reference`},
		{name: "open root object", schema: `{"type":"object","properties":{"name":{"type":"string"}}}`, want: `must declare additionalProperties false`},
		{name: "open nested object", schema: `{"type":"object","additionalProperties":false,"properties":{"metadata":{"type":"object","properties":{"label":{"type":"string"}}}}}`, want: `must declare additionalProperties false`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := fmt.Sprintf(`{
				"operations": [{
					"id":"acme.widgets.create",
					"kind":"rest_write",
					"summary":"Create widget",
					"risk":"medium",
					"approval":"required",
					"output_policy":"json_redacted",
					"mutation_class":"create",
					"request_contract":{"source_tier":1,"source_url":"https://example.invalid/openapi.json","source_location":"requestBody","fields":[]},
					"rest":{"method":"POST","path":"/widgets","body_schema":%s}
				}]
			}`, tt.schema)
			_, err := Load(operationsBundleFS(t, operations), "acme")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestBundleLoadValidatesFullRESTPathParameterGrammar(t *testing.T) {
	valid := `{
		"operations": [{
			"id":"acme.subscriptions.get",
			"kind":"rest_read",
			"summary":"Get subscription",
			"risk":"low",
			"approval":"none",
			"output_policy":"json_redacted",
			"request_contract":{
				"source_tier":3,
				"source_url":"https://example.invalid/subscriptions#get",
				"source_location":"Get subscription",
				"fields":[{"path":"path.subscription-id","source_url":"https://example.invalid/subscriptions#get","source_location":"subscription-id path parameter"}]
			},
			"rest":{"method":"GET","path":"/subscriptions/{subscription-id}","max_bytes":1024,"body":"none"}
		}]
	}`
	if _, err := Load(operationsBundleFS(t, valid), "acme"); err != nil {
		t.Fatalf("Load hyphenated path variable: %v", err)
	}

	invalid := strings.Replace(valid, `{subscription-id}`, `{subscription?id}`, 1)
	_, err := Load(operationsBundleFS(t, invalid), "acme")
	if err == nil || !strings.Contains(err.Error(), "invalid path variable") {
		t.Fatalf("Load malformed path variable error = %v, want invalid path variable", err)
	}
}

func TestRequestContractFieldsModelRootArrays(t *testing.T) {
	fields := map[string]bool{}
	err := collectRequestSchemaFields(json.RawMessage(`{
		"type":"array",
		"items":{
			"type":"object",
			"additionalProperties":false,
			"properties":{"name":{"type":"string"}}
		}
	}`), "body", fields)
	if err != nil {
		t.Fatalf("collectRequestSchemaFields: %v", err)
	}
	want := map[string]bool{"body[]": true, "body[].name": true}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %v, want %v", fields, want)
	}

	writeFields, err := declaredWriteRequestFields(WriteAction{
		BodyType:   "json_array",
		BodyField:  "items",
		BodySchema: json.RawMessage(`{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"}}}}`),
	})
	if err != nil {
		t.Fatalf("declaredWriteRequestFields: %v", err)
	}
	if !reflect.DeepEqual(writeFields, []string{"body[]", "body[].name"}) {
		t.Fatalf("write fields = %v, want root array fields", writeFields)
	}
}

func TestDeclaredWriteRequestFieldsCoverWireMechanisms(t *testing.T) {
	tests := []struct {
		name    string
		action  WriteAction
		want    []string
		wantErr string
	}{
		{
			name:   "graphql variables",
			action: WriteAction{BodyType: "graphql", GraphQL: &GraphQLRequestSpec{Variables: map[string]any{"widget": map[string]any{"name": "{{ record.name }}"}}}},
			want:   []string{"body.variables.widget", "body.variables.widget.name"},
		},
		{
			name:   "multipart parts",
			action: WriteAction{BodyType: "multipart", Multipart: &MultipartSpec{Parts: []MultipartPartSpec{{Name: "metadata", Field: "metadata", Type: "field"}, {Name: "file", Field: "path", Type: "file"}}}},
			want:   []string{"body.file", "body.metadata"},
		},
		{
			name: "base64 upload omits source",
			action: WriteAction{
				BodyType:     "base64_upload",
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"path":{"type":"string"},"label":{"type":"string"}}}`),
				Base64Upload: &Base64UploadSpec{SourceField: "path", ContentField: "content"},
			},
			want: []string{"body.content", "body.label"},
		},
		{
			name: "dynamic fields rejected",
			action: WriteAction{
				DynamicFields: &DynamicFieldsSpec{Field: "custom_fields"},
			},
			wantErr: "cannot be enumerated",
		},
		{
			name: "base64 query rejected",
			action: WriteAction{
				BodyType: "base64_upload",
				Query:    map[string]QueryParam{"notify": {Template: "true"}},
			},
			wantErr: "not transmitted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := declaredWriteRequestFields(tt.action)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("declaredWriteRequestFields error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("declaredWriteRequestFields: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("fields = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBundleLoadRejectsWriteHookRequestContractClaims(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions":[{
			"name":"create_widget",
			"kind":"create",
			"method":"POST",
			"path":"/widgets",
			"body_type":"none",
			"record_schema":{"type":"object","additionalProperties":false,"properties":{}},
			"risk":"creates a widget",
			"hook":"acme"
		}]
	}`)}
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations":[{
			"id":"acme.widgets.create",
			"kind":"rest_write",
			"summary":"Create widget",
			"risk":"medium",
			"approval":"required",
			"output_policy":"json_redacted",
			"mutation_class":"create",
			"request_contract":{"source_tier":3,"source_url":"https://example.invalid/widgets#create","source_location":"Create widget","write_action":"create_widget","fields":[]},
			"rest":{"method":"POST","path":"/widgets","body":"none"}
		}]
	}`)}
	_, err := Load(fsys, "acme")
	if err == nil || !strings.Contains(err.Error(), "uses write hook") {
		t.Fatalf("Load error = %v, want write hook claim rejection", err)
	}
}

func TestBundleLoadRejectsInvalidRequestContracts(t *testing.T) {
	tests := []struct {
		name       string
		operations string
		want       string
	}{
		{
			name:       "tier four without sibling",
			operations: `{"operations":[{"id":"acme.widgets.update","kind":"rest_write","summary":"Update","risk":"medium","approval":"required","output_policy":"json_redacted","mutation_class":"update","request_contract":{"source_tier":4,"source_url":"https://example.invalid/widgets#update","source_location":"Update request","fields":[]},"rest":{"method":"PATCH","path":"/widgets"}}]}`,
			want:       "requires sibling_operation",
		},
		{
			name:       "field missing citation",
			operations: `{"operations":[{"id":"acme.widgets.update","kind":"rest_write","summary":"Update","risk":"medium","approval":"required","output_policy":"json_redacted","mutation_class":"update","request_contract":{"source_tier":2,"source_url":"https://example.invalid/openapi.json","source_location":"update description","fields":[{"path":"body.name","source_url":"https://example.invalid/openapi.json","source_location":""}]},"rest":{"method":"PATCH","path":"/widgets"}}]}`,
			want:       "source_location is required",
		},
		{
			name:       "uncited schema field",
			operations: `{"operations":[{"id":"acme.widgets.update","kind":"rest_write","summary":"Update","risk":"medium","approval":"required","output_policy":"json_redacted","mutation_class":"update","request_contract":{"source_tier":2,"source_url":"https://example.invalid/openapi.json","source_location":"update description","fields":[]},"rest":{"method":"PATCH","path":"/widgets","body_schema":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"}}}}}]}`,
			want:       `missing citation for "body.name"`,
		},
		{
			name:       "no body with body field",
			operations: `{"operations":[{"id":"acme.widgets.refresh","kind":"rest_write","summary":"Refresh","risk":"medium","approval":"required","output_policy":"json_redacted","mutation_class":"update","request_contract":{"source_tier":3,"source_url":"https://example.invalid/widgets#refresh","source_location":"Refresh request","fields":[{"path":"body.payload","source_url":"https://example.invalid/widgets#refresh","source_location":"Refresh payload"}]},"rest":{"method":"POST","path":"/widgets:refresh","body":"none"}}]}`,
			want:       "conflicts with rest body none",
		},
		{
			name:       "no body without evidence",
			operations: `{"operations":[{"id":"acme.widgets.refresh","kind":"rest_write","summary":"Refresh","risk":"medium","approval":"required","output_policy":"json_redacted","mutation_class":"update","rest":{"method":"POST","path":"/widgets:refresh","body":"none"}}]}`,
			want:       "must declare request_contract evidence",
		},
		{
			name:       "unsupported body mode",
			operations: `{"operations":[{"id":"acme.widgets.refresh","kind":"rest_write","summary":"Refresh","risk":"medium","approval":"required","output_policy":"json_redacted","mutation_class":"update","rest":{"method":"POST","path":"/widgets:refresh","body":"empty"}}]}`,
			want:       `rest body string must be "none"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fullValidBundleFS("acme")
			fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(tt.operations)}
			_, err := Load(fsys, "acme")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestBundleLoadRejectsUnsafeOperationKind(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.raw.shell",
				"kind": "shell",
				"summary": "Run shell",
				"risk": "critical",
				"approval": "blocked",
				"output_policy": "json"
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected unsafe operation kind to be rejected")
	}
	if !strings.Contains(err.Error(), "operations.json") ||
		!strings.Contains(err.Error(), "/operations/0/kind") ||
		!strings.Contains(err.Error(), "not in enum") {
		t.Fatalf("Load error = %q, want operations.json kind enum rejection", err.Error())
	}
}

func TestBundleLoadRejectsOperationWithoutMatchingBlock(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.projects.list",
				"kind": "graphql_query",
				"summary": "List projects",
				"risk": "low",
				"approval": "none",
				"output_policy": "json",
				"rest": {
					"method": "GET",
					"path": "/projects"
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected graphql_query without graphql block to be rejected")
	}
	if !strings.Contains(err.Error(), "operations.json") ||
		!strings.Contains(err.Error(), "graphql_query") ||
		!strings.Contains(err.Error(), "graphql") {
		t.Fatalf("Load error = %q, want operations.json matching-block rejection", err.Error())
	}
}

func TestBundleLoadRejectsOperationWithMultipleExecutionBlocks(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.widgets.get",
				"kind": "rest_read",
				"summary": "Read one widget",
				"risk": "low",
				"approval": "none",
				"output_policy": "json",
				"rest": {
					"method": "GET",
					"path": "/widgets/{id}"
				},
				"graphql": {
					"operation_name": "Widget",
					"document": "query Widget($id: ID!) { node(id: $id) { id } }"
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected operation with multiple execution blocks to be rejected")
	}
	if !strings.Contains(err.Error(), "operations.json") ||
		!strings.Contains(err.Error(), "exactly one execution block") {
		t.Fatalf("Load error = %q, want operations.json single-block rejection", err.Error())
	}
}

const secretWriteOp = `{
		"id": "acme.secrets.put",
		"kind": "rest_write",
		"summary": "Create or update a repo secret",
		"risk": "high",
		"approval": "plan, preview, approval, execute",
		"output_policy": "json",
		"mutation_class": "secret",
		"secret_sensitive": true,
		"rest": {
			"method": "PUT",
			"path": "/repos/{owner}/{repo}/actions/secrets/{secret_name}"
		}%s
	}`

func TestBundleLoadRejectsSecretOperationWithoutPolicy(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"operations":[%s]}`, fmt.Sprintf(secretWriteOp, "")))}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected secret-sensitive operation without sensitive_policy to be rejected")
	}
	if !strings.Contains(err.Error(), "sensitive_policy") {
		t.Fatalf("Load error = %q, want sensitive_policy rejection", err.Error())
	}
}

func TestBundleLoadRejectsInlineInputModeForSecretOperation(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	policy := `, "sensitive_policy": {"input_mode": "inline", "redact_fields": ["value"], "transform": "none", "approval_mode": "typed_confirmation"}`
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"operations":[%s]}`, fmt.Sprintf(secretWriteOp, policy)))}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected inline input_mode for a secret operation to be rejected")
	}
	if !strings.Contains(err.Error(), "inline") || !strings.Contains(err.Error(), "input_mode") {
		t.Fatalf("Load error = %q, want inline input_mode rejection", err.Error())
	}
}

func TestBundleLoadRejectsSecretOperationWithoutTypedConfirmation(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	policy := `, "sensitive_policy": {"input_mode": "env", "redact_fields": ["value"], "transform": "github_secret_encryption", "approval_mode": "none"}`
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"operations":[%s]}`, fmt.Sprintf(secretWriteOp, policy)))}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected secret operation without typed_confirmation to be rejected")
	}
	if !strings.Contains(err.Error(), "typed_confirmation") {
		t.Fatalf("Load error = %q, want typed_confirmation rejection", err.Error())
	}
}

func TestBundleLoadAcceptsSecretOperationWithFullPolicy(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	policy := `, "sensitive_policy": {"input_mode": "env", "redact_fields": ["value"], "transform": "github_secret_encryption", "approval_mode": "typed_confirmation", "preflight": "scope_check"}`
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`{"operations":[%s]}`, fmt.Sprintf(secretWriteOp, policy)))}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: secret operation with full policy should be accepted: %v", err)
	}
	if len(b.Operations) != 1 || b.Operations[0].SensitivePolicy == nil {
		t.Fatalf("loaded operation missing sensitive_policy: %+v", b.Operations)
	}
	if got := b.Operations[0].SensitivePolicy.ApprovalMode; got != "typed_confirmation" {
		t.Fatalf("approval_mode = %q, want typed_confirmation", got)
	}
}

func TestBundleLoadRejectsDuplicateOperationIDs(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.widgets.get",
				"kind": "rest_read",
				"summary": "Read one widget",
				"risk": "low",
				"approval": "none",
				"output_policy": "json",
				"rest": {
					"method": "GET",
					"path": "/widgets/{id}"
				}
			},
			{
				"id": "acme.widgets.get",
				"kind": "rest_read",
				"summary": "Read one widget again",
				"risk": "low",
				"approval": "none",
				"output_policy": "json",
				"rest": {
					"method": "GET",
					"path": "/widgets/{id}"
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected duplicate operation IDs to be rejected")
	}
	if !strings.Contains(err.Error(), "operations.json") ||
		!strings.Contains(err.Error(), "duplicate operation id") {
		t.Fatalf("Load error = %q, want duplicate operation id rejection", err.Error())
	}
}

func TestBundleLoadRejectsRestWriteWithReadMethod(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.widgets.update",
				"kind": "rest_write",
				"summary": "Update one widget",
				"risk": "medium",
				"approval": "reverse ETL writes require plan, preview, approval, execute",
				"output_policy": "json",
				"mutation_class": "update",
				"rest": {
					"method": "GET",
					"path": "/widgets/{id}"
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected rest_write with GET to be rejected")
	}
	if !strings.Contains(err.Error(), "rest_write method must be mutating") {
		t.Fatalf("Load error = %q, want rest_write method rejection", err.Error())
	}
}

func TestBundleLoadRejectsBinaryDownloadWithoutPositiveLimit(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(`{
		"operations": [
			{
				"id": "acme.assets.download",
				"kind": "binary_download",
				"summary": "Download one asset",
				"risk": "medium",
				"approval": "filesystem writes require explicit destination approval",
				"output_policy": "file_manifest",
				"binary": {
					"method": "GET",
					"path": "/assets/{id}"
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected binary_download without max_bytes to be rejected")
	}
	if !strings.Contains(err.Error(), "binary_download must declare positive max_bytes") {
		t.Fatalf("Load error = %q, want binary max_bytes rejection", err.Error())
	}
}

func TestBundleLoadEmbeddedGitHubOperations(t *testing.T) {
	b, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	if len(b.Operations) == 0 {
		t.Fatalf("GitHub Operations is empty; defs.FS must embed operations.json")
	}
	found := false
	for _, op := range b.Operations {
		if op.ID == "github.projects.list" && op.Kind == "graphql_query" {
			found = true
		}
	}
	if !found {
		t.Fatalf("GitHub operations missing github.projects.list graphql_query example: %+v", b.Operations)
	}
}

func TestBundleLoadEmbeddedGitHubCLISurface(t *testing.T) {
	b, err := Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github): %v", err)
	}
	if b.CLISurface == nil {
		t.Fatalf("GitHub CLISurface is nil; defs.FS must embed cli_surface.json")
	}
	if b.CLISurface.Usage != "pm github <command> <subcommand> [flags]" {
		t.Fatalf("GitHub CLISurface usage = %q", b.CLISurface.Usage)
	}
	if len(b.CLISurface.Commands) == 0 {
		t.Fatalf("GitHub CLISurface has no commands")
	}
}

func TestBundleLoadRejectsUnknownCLISurfaceCommandKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/cli_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"tagline": "Work with Acme from the command line.",
		"usage": "pm acme <command> [flags]",
		"commands": [
			{
				"path": "widget list",
				"summary": "List widgets",
				"intent": "etl",
				"availability": "implemented",
				"stream": "widgets",
				"surprise": true
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown cli_surface command key")
	}
	if !strings.Contains(err.Error(), "cli_surface.json") || !strings.Contains(err.Error(), "surprise") {
		t.Fatalf("Load error = %q, want it to name cli_surface.json and surprise", err.Error())
	}
}

func TestBundleLoadStreamsOptionalIffDynamicSchema(t *testing.T) {
	fsys := fstest.MapFS{
		"pg/metadata.json":    &fstest.MapFile{Data: []byte(dynamicSchemaMetadata("pg"))},
		"pg/spec.json":        &fstest.MapFile{Data: []byte(validSpec)},
		"pg/api_surface.json": &fstest.MapFile{Data: []byte(`{"api":"pg","endpoints":[]}`)},
		"pg/docs.md":          &fstest.MapFile{Data: []byte(validDocs)},
	}

	b, err := Load(fsys, "pg")
	if err != nil {
		t.Fatalf("Load should succeed without streams.json when dynamic_schema=true: %v", err)
	}
	if len(b.Streams) != 0 {
		t.Fatalf("Streams = %+v, want empty", b.Streams)
	}
}

func TestBundleLoadStreamsRequiredWithoutDynamicSchema(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/streams.json")

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("expected error: streams.json required when dynamic_schema=false")
	}
	if !strings.Contains(err.Error(), "streams.json") {
		t.Fatalf("error %q does not name streams.json", err.Error())
	}
}

func TestBundleLoadDirNameMismatch(t *testing.T) {
	fsys := fullValidBundleFS("actual-dir")
	fsys["actual-dir/metadata.json"] = &fstest.MapFile{Data: []byte(validMetadata("declared-name"))}

	_, err := Load(fsys, "actual-dir")
	if err == nil {
		t.Fatalf("expected dir-name/metadata.name mismatch error")
	}
	if !strings.Contains(err.Error(), "actual-dir") || !strings.Contains(err.Error(), "declared-name") {
		t.Fatalf("error %q does not name both dir and metadata name", err.Error())
	}
}

func TestBundleLoadBadNameRegex(t *testing.T) {
	fsys := fullValidBundleFS("Source-GitHub")
	fsys["Source-GitHub/metadata.json"] = &fstest.MapFile{Data: []byte(validMetadata("Source-GitHub"))}

	_, err := Load(fsys, "Source-GitHub")
	if err == nil {
		t.Fatalf("expected bad name regex error")
	}
	if !strings.Contains(err.Error(), "Source-GitHub") {
		t.Fatalf("error %q does not name the offending value", err.Error())
	}
}

func TestBundleLoadMissingRequiredFile(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/metadata.json")

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("expected missing required file error")
	}
	if !strings.Contains(err.Error(), "metadata.json") {
		t.Fatalf("error %q does not name the missing file", err.Error())
	}
}

func TestBundleLoadAPISurfaceOptionalForRuntime(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	delete(fsys, "acme/api_surface.json")

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load without api_surface.json: %v", err)
	}
	if b.Surface != nil {
		t.Fatalf("Surface = %+v, want nil when api_surface.json is absent", b.Surface)
	}
}

func TestBundleLoadMetaSchemaViolation(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	// metadata.json missing the required "capabilities" field -> meta-schema violation.
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(`{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga"
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("expected meta-schema violation error for metadata.json missing capabilities")
	}
}

func TestBundleLoadAllIteratesBundles(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	for k, v := range fullValidBundleFS("beta") {
		fsys[k] = v
	}

	bundles, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(bundles) != 2 {
		t.Fatalf("LoadAll returned %d bundles, want 2", len(bundles))
	}
	names := map[string]bool{}
	for _, b := range bundles {
		names[b.Name] = true
	}
	if !names["acme"] || !names["beta"] {
		t.Fatalf("LoadAll bundles = %v", names)
	}
}

func TestBundleLoadAllEmptyTreeIsFine(t *testing.T) {
	bundles, err := LoadAll(fstest.MapFS{})
	if err != nil {
		t.Fatalf("LoadAll on empty tree: %v", err)
	}
	if len(bundles) != 0 {
		t.Fatalf("LoadAll on empty tree returned %d bundles", len(bundles))
	}
}

// TestBundleLoadAllOneBadBundleDoesNotHideTheRest is an ENGINE HARDENING
// regression (hardening-ledger.md): LoadAll previously aborted the ENTIRE
// batch (returned nil bundles, a single-bundle error) the instant ANY ONE
// directory failed to load. With ~400 independently-authored bundles in
// defs/, and the newly-added strict-decode/meta-schema unknown-key checks
// now correctly failing a real (if large) subset of them, that all-or-
// nothing contract meant a single malformed bundle anywhere in the fleet
// silently hid every other (compliant) bundle from LoadAll's caller — the
// exact "one bad apple spoils fleet-wide discoverability" failure mode
// cmd/connectorgen's own validateBundleDir already avoids by design (it
// isolates one bundle's load error into a Finding and keeps validating the
// rest). LoadAll now mirrors that same resilience: it still returns every
// bundle that DID load cleanly, and a non-nil error whenever at least one
// did not — the error names every failing bundle (not just the first) so a
// caller that treats err!=nil as fatal still learns the full failing set
// from the error text, and a caller that wants the good subset (this
// package's own defs.FS-wide golden/parity tests, conformance) can keep
// going against bundles.
func TestBundleLoadAllOneBadBundleDoesNotHideTheRest(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	for k, v := range fullValidBundleFS("beta") {
		fsys[k] = v
	}
	// "broken" has an unknown base-level key (streams.json) and must fail
	// to load, but must not prevent acme/beta from coming back.
	for k, v := range fullValidBundleFS("broken") {
		fsys[k] = v
	}
	fsys["broken/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"query": { "limit": "1" },
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	bundles, err := LoadAll(fsys)
	if err == nil {
		t.Fatalf("LoadAll: expected a non-nil error naming the broken bundle, got nil (bundles: %+v)", bundles)
	}
	if !strings.Contains(err.Error(), "broken") {
		t.Fatalf("LoadAll error = %q, want it to name the failing bundle %q", err.Error(), "broken")
	}
	var loadErr *LoadAllError
	if !errors.As(err, &loadErr) {
		t.Fatalf("LoadAll error = %v (%T), want it to be (or wrap) a *LoadAllError", err, err)
	}
	if len(loadErr.Failures) != 1 || loadErr.Failures[0].Name != "broken" {
		t.Fatalf("LoadAllError.Failures = %+v, want exactly one entry named %q", loadErr.Failures, "broken")
	}

	names := map[string]bool{}
	for _, b := range bundles {
		names[b.Name] = true
	}
	if !names["acme"] || !names["beta"] {
		t.Fatalf("LoadAll bundles = %v, want acme and beta still returned despite broken's failure", names)
	}
	if names["broken"] {
		t.Fatalf("LoadAll bundles = %v, want broken itself excluded (it never loaded)", names)
	}
}

// TestBundleLoadAllDefsFS exercises the real embedded defs.FS scaffold
// (the `all:*` embed directive in defs.go) end-to-end: the stray embedded
// defs.go file is ignored, and the Wave F goldens must load cleanly.
//
// ENGINE HARDENING (hardening-ledger.md): this no longer asserts err == nil.
// The newly-added streams.json/writes.json/metadata.json unknown-key checks
// (meta-schema additionalProperties:false + loader strict-decode) correctly
// fail a real, pre-existing subset of defs/ bundles that declared fields
// the engine silently ignored (rentcast's "base.check.query" and ~150
// siblings' identical shape — RequestSpec only carries Method/Path, so that
// JSON never did anything at runtime). Repairing those bundles' own
// streams.json files is explicitly out of scope for this dispatch (listed
// in the hardening ledger for a follow-up instead), so LoadAll(defs.FS) is
// now expected to return a non-nil error naming them. What this test still
// pins, unweakened: LoadAll's resilience contract (TestBundleLoadAll
// OneBadBundleDoesNotHideTheRest) means the golden bundles must STILL come
// back in the returned slice even though err is non-nil, and any bundle
// that legitimately fails must be a KNOWN, currently-tracked case, not a
// silent new regression — this test fails loudly if an UNEXPECTED bundle
// starts failing to load.
func TestBundleLoadAllDefsFS(t *testing.T) {
	bundles, err := LoadAll(defs.FS)
	byName := map[string]bool{}
	for _, b := range bundles {
		byName[b.Name] = true
	}
	for _, golden := range []string{"stripe", "postgres"} {
		if !byName[golden] {
			t.Fatalf("LoadAll(defs.FS) missing golden bundle %q (got %v); err=%v", golden, byName, err)
		}
	}
	var loadErr *LoadAllError
	if err != nil && !errors.As(err, &loadErr) {
		t.Fatalf("LoadAll(defs.FS) returned an error NOT shaped like the known per-bundle unknown-key failures (hardening-ledger.md): %v", err)
	}
}

// TestBundleLoadFromOnDiskTestdata exercises the loader against a real
// os.DirFS-backed fixture bundle (testdata/bundles/widget-demo), rather than
// only the in-memory fstest.MapFS cases above.
func TestBundleLoadFromOnDiskTestdata(t *testing.T) {
	fsys := os.DirFS("testdata/bundles")

	b, err := Load(fsys, "widget-demo")
	if err != nil {
		t.Fatalf("Load(testdata/bundles, widget-demo): %v", err)
	}
	if b.Name != "widget-demo" {
		t.Fatalf("Name = %q", b.Name)
	}
	if len(b.Streams) != 1 || b.Streams[0].Name != "widgets" {
		t.Fatalf("Streams = %+v", b.Streams)
	}
	if b.Fixtures == nil {
		t.Fatalf("Fixtures should be non-nil")
	}

	bundles, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll(testdata/bundles): %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("LoadAll(testdata/bundles) returned %d bundles, want 1", len(bundles))
	}
}

// --- optional conformance skip markers (R3: hook-aware dynamic conformance) --
//
// A bundle may declare an OPTIONAL, explicit "conformance" marker at either
// stream level (streams.json's per-stream {"conformance": {"skip_dynamic":
// true, "reason": "..."}}) or bundle level (metadata.json's top-level
// equivalent), for connectors whose dynamic (fixture-replay) checks cannot
// meaningfully run because the bundle's real behavior lives entirely behind
// a Tier-2 hook that conformance's declarative-only replay harness cannot
// exercise. This is parsed by the loader (no behavior beyond struct
// population); dynamic.go interprets the marker, connectorgen validate
// requires a non-empty reason.

const streamsWithStreamConformanceMarker = `{
	"base": {
		"url": "{{ config.base_url }}",
		"user_agent": "test-agent",
		"headers": {},
		"auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ cursor }}" } ],
		"pagination": { "type": "none" },
		"check": { "method": "GET", "path": "/ping" },
		"error_map": []
	},
	"streams": [
		{
			"name": "widgets",
			"path": "/widgets",
			"records": { "path": "data" },
			"schema": "schemas/widgets.json",
			"conformance": { "skip_dynamic": true, "reason": "hook-covered; proven live by archived parity evidence for acme" }
		}
	]
}`

func TestBundleLoadParsesStreamConformanceMarker(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(streamsWithStreamConformanceMarker)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
	s := b.Streams[0]
	if s.Conformance == nil {
		t.Fatalf("stream %q Conformance marker not parsed (got nil)", s.Name)
	}
	if !s.Conformance.SkipDynamic {
		t.Fatalf("stream %q Conformance.SkipDynamic = false, want true", s.Name)
	}
	if s.Conformance.Reason == "" {
		t.Fatalf("stream %q Conformance.Reason is empty", s.Name)
	}
}

// TestBundleLoadStreamWithNoConformanceMarkerIsNil locks in that an ordinary
// stream (no "conformance" key at all) parses with a nil marker, not a
// zero-value non-nil struct — dynamic.go's marker-presence check must be
// able to distinguish "no marker" from "marker present but false".
func TestBundleLoadStreamWithNoConformanceMarkerIsNil(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Streams[0].Conformance != nil {
		t.Fatalf("Conformance = %+v, want nil for a stream with no conformance block", b.Streams[0].Conformance)
	}
}

func metadataWithBundleConformanceMarker(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"conformance": { "skip_dynamic": true, "reason": "custom-auth-only; hook not registered in conformance's replay harness" }
	}`
}

func TestBundleLoadParsesBundleLevelConformanceMarker(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(metadataWithBundleConformanceMarker("acme"))}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Conformance == nil {
		t.Fatalf("Metadata.Conformance marker not parsed (got nil)")
	}
	if !b.Metadata.Conformance.SkipDynamic {
		t.Fatalf("Metadata.Conformance.SkipDynamic = false, want true")
	}
	if b.Metadata.Conformance.Reason == "" {
		t.Fatalf("Metadata.Conformance.Reason is empty")
	}
}

func TestBundleLoadMetadataWithNoConformanceMarkerIsNil(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Conformance != nil {
		t.Fatalf("Metadata.Conformance = %+v, want nil for metadata with no conformance block", b.Metadata.Conformance)
	}
}

// streamsWithOptionalQueryDialect exercises the gap-loop item-3 optional-query
// dialect (REVIEW-B.md cross-cutting adjudication 2): a stream.Query entry
// may be either a plain string (today's exact hard-error semantics,
// "page[size]") or an object {template, omit_when_absent, default}.
const streamsWithOptionalQueryDialect = `{
	"base": { "url": "{{ config.base_url }}" },
	"streams": [
		{
			"name": "widgets",
			"path": "/widgets",
			"records": { "path": "data" },
			"schema": "schemas/widgets.json",
			"query": {
				"page[size]": "100",
				"status": { "template": "{{ config.status }}", "omit_when_absent": true },
				"count": { "template": "{{ config.page_size }}", "default": "100" }
			}
		}
	]
}`

// TestBundleLoadParsesOptionalQueryDialect proves streams.json's per-entry
// query dialect round-trips through the loader: a plain string entry stays a
// hard-required template; an object entry carries its template/
// omit_when_absent/default fields distinctly.
func TestBundleLoadParsesOptionalQueryDialect(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(streamsWithOptionalQueryDialect)}
	fsys["acme/spec.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["base_url"],
		"properties": {
			"base_url": { "type": "string" },
			"status": { "type": "string" },
			"page_size": { "type": "string" }
		}
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
	q := b.Streams[0].Query
	staticEntry, ok := q["page[size]"]
	if !ok || staticEntry.Template != "100" || staticEntry.OmitWhenAbsent {
		t.Fatalf("query[page[size]] = %+v, want plain string entry Template=100 OmitWhenAbsent=false", staticEntry)
	}
	statusEntry, ok := q["status"]
	if !ok || statusEntry.Template != "{{ config.status }}" || !statusEntry.OmitWhenAbsent {
		t.Fatalf("query[status] = %+v, want Template={{ config.status }} OmitWhenAbsent=true", statusEntry)
	}
	countEntry, ok := q["count"]
	if !ok || countEntry.Template != "{{ config.page_size }}" || countEntry.Default != "100" {
		t.Fatalf("query[count] = %+v, want Template={{ config.page_size }} Default=100", countEntry)
	}
}

// --- ENGINE HARDENING: unknown-key strict decode ---------------------------
//
// The re-review (hardening-ledger.md) found internal/connectors/defs/rentcast
// declaring "base.check.query" (and several other bundles declaring a bare
// "base.query"), a field HTTPBase/RequestSpec do not have at all. Because
// json.Unmarshal silently drops unknown object keys and the meta-schemas
// previously left every nested sub-object as a bare {"type":"object"} with no
// additionalProperties:false, that invented mechanism passed every gate
// (meta-schema validate, connectorgen validate, go build) while doing
// nothing at runtime — Check() never sends a query at all. These tests pin
// TWO independent layers of defense: (1) the meta-schemas
// (streams.schema.json/writes.schema.json/metadata.schema.json) now declare
// explicit property allowlists with additionalProperties:false on every
// structured sub-object (free-form maps like headers/query/body/
// computed_fields/record_schema and user JSON-Schema documents like
// spec.json's "properties" are deliberately left open); (2) the loader
// itself strict-decodes streams.json/writes.json/metadata.json (independent
// of the meta-schema, so a future meta-schema regression/relaxation cannot
// silently reopen this hole) and names the offending file+key in the error.

func TestBundleLoadRejectsUnknownBaseLevelKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"query": { "limit": "1" },
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown base-level key %q, got nil", "query")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "query") {
		t.Fatalf("Load error = %q, want it to name streams.json and the unknown key %q", err.Error(), "query")
	}
}

// TestBundleLoadAcceptsBaseCheckQueryKey supersedes the former
// TestBundleLoadRejectsUnknownBaseCheckQueryKey (checkquery-ledger.md):
// base.check.query (the exact rentcast shape the hardening ledger's trigger
// named, and 148 siblings' identical shape) is no longer an unknown key —
// RequestSpec now has a Query map[string]QueryParam field mirroring
// StreamSpec.Query's existing string-or-object dialect verbatim, per the
// hardening ledger's own suggested follow-up shape. Loading must now succeed
// AND the query must round-trip into RequestSpec.Query exactly like
// stream.Query does.
func TestBundleLoadAcceptsBaseCheckQueryKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping", "query": { "limit": "1", "offset": "0" } }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v, want base.check.query to load cleanly (RequestSpec.Query dialect addition)", err)
	}
	if b.HTTP.Check == nil {
		t.Fatalf("HTTP.Check is nil")
	}
	limit, ok := b.HTTP.Check.Query["limit"]
	if !ok || limit.Template != "1" {
		t.Fatalf("Check.Query[limit] = %+v, want plain string entry Template=1", limit)
	}
	offset, ok := b.HTTP.Check.Query["offset"]
	if !ok || offset.Template != "0" {
		t.Fatalf("Check.Query[offset] = %+v, want plain string entry Template=0", offset)
	}
}

// TestBundleLoadParsesCheckQueryOptionalDialect proves check.query accepts
// the SAME object-form (omit_when_absent/default) dialect as stream.Query,
// not just plain strings — since RequestSpec.Query reuses the identical
// QueryParam type.
func TestBundleLoadParsesCheckQueryOptionalDialect(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": {
				"method": "GET",
				"path": "/ping",
				"query": {
					"limit": "1",
					"status": { "template": "{{ config.status }}", "omit_when_absent": true },
					"count": { "template": "{{ config.page_size }}", "default": "100" }
				}
			}
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}
	fsys["acme/spec.json"] = &fstest.MapFile{Data: []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["base_url"],
		"properties": {
			"base_url": { "type": "string" },
			"status": { "type": "string" },
			"page_size": { "type": "string" }
		}
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	q := b.HTTP.Check.Query
	staticEntry, ok := q["limit"]
	if !ok || staticEntry.Template != "1" || staticEntry.OmitWhenAbsent {
		t.Fatalf("Check.Query[limit] = %+v, want plain string entry Template=1 OmitWhenAbsent=false", staticEntry)
	}
	statusEntry, ok := q["status"]
	if !ok || statusEntry.Template != "{{ config.status }}" || !statusEntry.OmitWhenAbsent {
		t.Fatalf("Check.Query[status] = %+v, want Template={{ config.status }} OmitWhenAbsent=true", statusEntry)
	}
	countEntry, ok := q["count"]
	if !ok || countEntry.Template != "{{ config.page_size }}" || countEntry.Default != "100" {
		t.Fatalf("Check.Query[count] = %+v, want Template={{ config.page_size }} Default=100", countEntry)
	}
}

func TestBundleLoadRejectsUnknownStreamLevelKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json", "not_a_real_field": true }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown stream-level key %q, got nil", "not_a_real_field")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "not_a_real_field") {
		t.Fatalf("Load error = %q, want it to name streams.json and the unknown key %q", err.Error(), "not_a_real_field")
	}
}

func TestBundleLoadRejectsUnknownAuthCandidateKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "scope": "read" } ],
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown auth-candidate key %q (note: valid key is \"scopes\", not \"scope\"), got nil", "scope")
	}
	if !strings.Contains(err.Error(), "streams.json") || !strings.Contains(err.Error(), "scope") {
		t.Fatalf("Load error = %q, want it to name streams.json and the unknown key %q", err.Error(), "scope")
	}
}

func TestBundleLoadRejectsUnknownWritesActionKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(`{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": true, "query": false, "cdc": false, "dynamic_schema": false }
	}`)}
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{
		"actions": [
			{
				"name": "create_widget",
				"kind": "create",
				"method": "POST",
				"path": "/widgets",
				"record_schema": { "type": "object", "properties": {} },
				"risk": "low",
				"retries": 3
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown writes-action key %q, got nil", "retries")
	}
	if !strings.Contains(err.Error(), "writes.json") || !strings.Contains(err.Error(), "retries") {
		t.Fatalf("Load error = %q, want it to name writes.json and the unknown key %q", err.Error(), "retries")
	}
}

func TestBundleLoadRejectsUnknownAPISurfaceEndpointKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" }, "deprecated": true }
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown api_surface.json endpoint key %q, got nil", "deprecated")
	}
	if !strings.Contains(err.Error(), "api_surface.json") || !strings.Contains(err.Error(), "deprecated") {
		t.Fatalf("Load error = %q, want it to name api_surface.json and the unknown key %q", err.Error(), "deprecated")
	}
}

func TestBundleLoadAPISurfaceOperationLedger(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{
				"method": "GET",
				"path": "/widgets/{id}",
				"operation": {
					"model": "direct_read",
					"status": "blocked",
					"risk": "low",
					"blocked_by_default": true,
					"reason": "point lookup candidate, not yet modeled as a stream",
					"source_url": "https://example.invalid/rest/widgets"
				}
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Surface.OperationLedgerVersion != 1 {
		t.Fatalf("OperationLedgerVersion = %d, want 1", b.Surface.OperationLedgerVersion)
	}
	if len(b.Surface.Endpoints) != 2 {
		t.Fatalf("endpoints = %d, want 2", len(b.Surface.Endpoints))
	}
	op := b.Surface.Endpoints[1].Operation
	if op == nil {
		t.Fatalf("Operation = nil, want operation metadata")
	}
	if op.Model != "direct_read" || op.Status != "blocked" || op.Risk != "low" {
		t.Fatalf("Operation = %+v, want direct_read/blocked/low", op)
	}
	if !op.BlockedByDefault {
		t.Fatalf("BlockedByDefault = false, want true")
	}
}

func TestBundleLoadAPISurfaceOperationRejectsUnblockedDefault(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"operation_ledger_version": 1,
		"endpoints": [
			{
				"method": "GET",
				"path": "/widgets/{id}",
				"operation": {
					"model": "direct_read",
					"status": "blocked",
					"risk": "low",
					"blocked_by_default": false,
					"reason": "point lookup candidate, not yet modeled as a stream"
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected api_surface.json schema error for blocked_by_default=false, got nil")
	}
	if !strings.Contains(err.Error(), "api_surface.json") || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("Load error = %q, want api_surface.json enum error", err.Error())
	}
}

func TestBundleLoadRejectsUnknownMetadataTopLevelKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/metadata.json"] = &fstest.MapFile{Data: []byte(`{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"maintainer": "nobody"
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown metadata.json top-level key %q, got nil", "maintainer")
	}
	if !strings.Contains(err.Error(), "metadata.json") || !strings.Contains(err.Error(), "maintainer") {
		t.Fatalf("Load error = %q, want it to name metadata.json and the unknown key %q", err.Error(), "maintainer")
	}
}

// TestBundleLoadStillAcceptsFreeFormMapKeys pins the deliberate scope
// boundary: headers, stream.query (string-or-object dialect), body, and
// computed_fields are genuinely free-form maps (arbitrary caller-defined
// keys), and must NOT be rejected by the strict-decode/meta-schema
// tightening above.
func TestBundleLoadStillAcceptsFreeFormMapKeys(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"headers": { "X-Anything-Custom": "v1", "X-Another-One": "v2" },
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"query": { "arbitrary_param_name": "{{ config.base_url }}" },
				"body": { "any_shape_here": { "nested": true } },
				"records": { "path": "data" },
				"computed_fields": { "whatever_field": "{{ record.id }}" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: unexpected error for free-form map keys: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
}

// TestLoadStreamsPageNumberStartPageZeroRoundTrips (S4 engine mini-wave item
// 1): streams.json's stream-level "start_page": 0 must decode into a non-nil
// *int pointing at 0 — not a nil pointer (which newPaginator/legacy would
// read as "absent, default to 1"). This is what makes an explicit 0 start
// distinguishable from an omitted start_page at every layer between the JSON
// file and the paginator.
func TestLoadStreamsPageNumberStartPageZeroRoundTrips(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"pagination": { "type": "page_number", "page_param": "page", "start_page": 0, "page_size": 10 },
				"records": { "path": "data" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(b.Streams) != 1 {
		t.Fatalf("Streams = %+v, want 1", b.Streams)
	}
	pag := b.Streams[0].Pagination
	if pag == nil {
		t.Fatalf("Streams[0].Pagination is nil, want a decoded pagination block")
	}
	if pag.StartPage == nil {
		t.Fatalf("Streams[0].Pagination.StartPage is nil, want a pointer to 0 (explicit start_page:0 must not decode as absent)")
	}
	if *pag.StartPage != 0 {
		t.Fatalf("*Streams[0].Pagination.StartPage = %d, want 0", *pag.StartPage)
	}
}

// TestLoadStreamsPageNumberStartPageAbsentIsNilPointer pins the companion
// case: a pagination block that never mentions start_page at all must decode
// to a nil pointer (not a pointer to the JSON zero value), preserving the
// "absent -> default to 1" behavior for every bundle that predates this
// change.
func TestLoadStreamsPageNumberStartPageAbsentIsNilPointer(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"pagination": { "type": "page_number", "page_param": "page", "page_size": 10 },
				"records": { "path": "data" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Streams[0].Pagination.StartPage != nil {
		t.Fatalf("Streams[0].Pagination.StartPage = %v, want nil (start_page never declared)", *b.Streams[0].Pagination.StartPage)
	}
}

// --- S4 engine mini-wave item 2: sub-resource fan-out -----------------------

// TestLoadStreamsFanOutConfigKeyRoundTrips pins the config_key + query_param
// shape (appfollow's app_collection_ids -> apps_id).
func TestLoadStreamsFanOutConfigKeyRoundTrips(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "app_lists",
				"path": "/account/apps/app",
				"records": { "path": "data" },
				"schema": "schemas/widgets.json",
				"fan_out": {
					"ids_from": { "config_key": "app_collection_ids" },
					"into": { "query_param": "apps_id" },
					"stamp_field": "app_id"
				}
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	fo := b.Streams[0].FanOut
	if fo == nil {
		t.Fatalf("Streams[0].FanOut is nil, want a decoded fan_out block")
	}
	if fo.IDsFrom.ConfigKey != "app_collection_ids" {
		t.Fatalf("FanOut.IDsFrom.ConfigKey = %q, want %q", fo.IDsFrom.ConfigKey, "app_collection_ids")
	}
	if fo.IDsFrom.Request != nil {
		t.Fatalf("FanOut.IDsFrom.Request = %+v, want nil (config_key form)", fo.IDsFrom.Request)
	}
	if fo.Into.QueryParam != "apps_id" {
		t.Fatalf("FanOut.Into.QueryParam = %q, want %q", fo.Into.QueryParam, "apps_id")
	}
	if fo.StampField != "app_id" {
		t.Fatalf("FanOut.StampField = %q, want %q", fo.StampField, "app_id")
	}
}

// TestLoadStreamsFanOutRequestFormRoundTrips pins the preliminary-request +
// path_var shape.
func TestLoadStreamsFanOutRequestFormRoundTrips(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "tasks",
				"path": "/projects/{{ fanout.id }}/tasks",
				"records": { "path": "data" },
				"schema": "schemas/widgets.json",
				"fan_out": {
					"ids_from": { "request": { "path": "/projects", "records_path": "data", "id_field": "id" } },
					"into": { "path_var": "parent_id" },
					"stamp_field": "project_id"
				}
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	fo := b.Streams[0].FanOut
	if fo == nil {
		t.Fatalf("Streams[0].FanOut is nil, want a decoded fan_out block")
	}
	if fo.IDsFrom.ConfigKey != "" {
		t.Fatalf("FanOut.IDsFrom.ConfigKey = %q, want empty (request form)", fo.IDsFrom.ConfigKey)
	}
	if fo.IDsFrom.Request == nil {
		t.Fatalf("FanOut.IDsFrom.Request is nil, want a decoded request block")
	}
	if fo.IDsFrom.Request.Path != "/projects" || fo.IDsFrom.Request.RecordsPath != "data" || fo.IDsFrom.Request.IDField != "id" {
		t.Fatalf("FanOut.IDsFrom.Request = %+v, want Path=/projects RecordsPath=data IDField=id", fo.IDsFrom.Request)
	}
	if fo.Into.PathVar != "parent_id" {
		t.Fatalf("FanOut.Into.PathVar = %q, want %q", fo.Into.PathVar, "parent_id")
	}
}

// TestLoadStreamsWithoutFanOutIsNilPointer pins the zero-impact case: an
// ordinary stream declaring no fan_out block at all decodes to a nil
// *FanOutSpec, not a zero-valued non-nil struct.
func TestLoadStreamsWithoutFanOutIsNilPointer(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Streams[0].FanOut != nil {
		t.Fatalf("Streams[0].FanOut = %+v, want nil", b.Streams[0].FanOut)
	}
}

// TestLoadStreamsFanOutRejectsUnknownKey proves the meta-schema's
// additionalProperties:false on fan_out/ids_from/into rejects a typo'd key
// rather than silently dropping it (the exact hardening-ledger.md class of
// defect this repo's meta-schemas are disciplined about).
func TestLoadStreamsFanOutRejectsUnknownKey(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "app_lists",
				"path": "/account/apps/app",
				"records": { "path": "data" },
				"schema": "schemas/widgets.json",
				"fan_out": {
					"ids_from": { "config_key": "app_collection_ids" },
					"into": { "query_param": "apps_id" },
					"stamp_field": "app_id",
					"unexpected_key": true
				}
			}
		]
	}`)}

	_, err := Load(fsys, "acme")
	if err == nil {
		t.Fatalf("Load: expected an error for unknown fan_out key %q, got nil", "unexpected_key")
	}
}

// --- S4 engine mini-wave item 3: keyed-object flatten -----------------------

// TestLoadStreamsRecordsKeyedObjectRoundTrips proves records.keyed_object and
// records.key_field decode onto RecordsSpec.
func TestLoadStreamsRecordsKeyedObjectRoundTrips(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	fsys["acme/streams.json"] = &fstest.MapFile{Data: []byte(`{
		"base": {
			"url": "{{ config.base_url }}",
			"check": { "method": "GET", "path": "/ping" }
		},
		"streams": [
			{
				"name": "widgets",
				"path": "/widgets",
				"records": { "path": "products", "keyed_object": true, "key_field": "product_id" },
				"schema": "schemas/widgets.json"
			}
		]
	}`)}

	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	rec := b.Streams[0].Records
	if !rec.KeyedObject {
		t.Fatalf("Records.KeyedObject = false, want true")
	}
	if rec.KeyField != "product_id" {
		t.Fatalf("Records.KeyField = %q, want %q", rec.KeyField, "product_id")
	}
}

// TestLoadStreamsRecordsWithoutKeyedObjectDefaultsFalse pins the zero-impact
// case: a records block that never mentions keyed_object decodes to false.
func TestLoadStreamsRecordsWithoutKeyedObjectDefaultsFalse(t *testing.T) {
	fsys := fullValidBundleFS("acme")
	b, err := Load(fsys, "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Streams[0].Records.KeyedObject {
		t.Fatalf("Records.KeyedObject = true, want false (never declared)")
	}
}

func operationsBundleFS(t *testing.T, operationsJSON string) fstest.MapFS {
	t.Helper()
	fsys := fullValidBundleFS("acme")
	fsys["acme/operations.json"] = &fstest.MapFile{Data: []byte(operationsJSON)}
	return fsys
}

func TestBundleLoadAcceptsRequiredQueryGroups(t *testing.T) {
	b, err := Load(operationsBundleFS(t, `{
		"operations": [{
			"id": "acme.users.list",
			"kind": "rest_read",
			"summary": "List users",
			"risk": "medium",
			"approval": "none",
			"output_policy": "json_redacted",
			"request_contract": {
				"source_tier": 3,
				"source_url": "https://example.invalid/users#list",
				"source_location": "List users request",
				"fields": [
					{"path":"query.email","source_url":"https://example.invalid/users#list","source_location":"query parameter email"},
					{"path":"query.id","source_url":"https://example.invalid/users#list","source_location":"query parameter id"}
				]
			},
			"rest": {
				"method": "GET",
				"path": "/users",
				"max_bytes": 1024,
				"required_query": [{"any_of": ["email", "id"]}]
			}
		}]
	}`), "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	groups := b.Operations[0].REST.RequiredQuery
	if len(groups) != 1 || len(groups[0].AnyOf) != 2 {
		t.Fatalf("required_query = %+v, want one group of two", groups)
	}
}

func TestBundleLoadRejectsUnenforceableRequiredQuery(t *testing.T) {
	tests := []struct {
		name  string
		group string
	}{
		{name: "empty any_of", group: `{"any_of": []}`},
		{name: "missing any_of", group: `{}`},
		{name: "blank parameter name", group: `{"any_of": ["email", "  "]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(operationsBundleFS(t, `{
				"operations": [{
					"id": "acme.users.list",
					"kind": "rest_read",
					"summary": "List users",
					"risk": "medium",
					"approval": "none",
					"output_policy": "json_redacted",
					"rest": {
						"method": "GET",
						"path": "/users",
						"max_bytes": 1024,
						"required_query": [`+tt.group+`]
					}
				}]
			}`), "acme")
			if err == nil {
				t.Fatal("want load error: a group that can never be satisfied is unenforceable and must fail loudly")
			}
			if !strings.Contains(err.Error(), "required_query") {
				t.Fatalf("error should name required_query, got %v", err)
			}
		})
	}
}

func base64UploadBundleFS(t *testing.T, action string) fstest.MapFS {
	t.Helper()
	fsys := fullValidBundleFS("acme")
	fsys["acme/writes.json"] = &fstest.MapFile{Data: []byte(`{"actions": [` + action + `]}`)}
	return fsys
}

const validBase64UploadAction = `{
	"name": "upload_attachment",
	"kind": "create",
	"method": "POST",
	"path": "/v0/{base_id}/{record_id}/{field}/uploadAttachment",
	"risk": "medium",
	"body_type": "base64_upload",
	"base64_upload": {
		"source_field": "file_path",
		"content_field": "file",
		"max_decoded_bytes": 3932160,
		"max_encoded_bytes": 5242880
	},
	"record_schema": {"type": "object", "properties": {"file_path": {"type": "string"}}}
}`

func TestBundleLoadAcceptsBase64UploadAction(t *testing.T) {
	b, err := Load(base64UploadBundleFS(t, validBase64UploadAction), "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	spec := b.Writes[0].Base64Upload
	if spec == nil {
		t.Fatal("base64_upload spec not parsed")
	}
	if spec.SourceField != "file_path" || spec.ContentField != "file" || spec.MaxDecodedBytes != 3932160 {
		t.Fatalf("base64_upload = %+v", spec)
	}
}

func TestBundleLoadRejectsInvalidBase64UploadAction(t *testing.T) {
	tests := []struct {
		name   string
		action string
		want   string
	}{
		{
			name: "body_type without spec",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low",
				"body_type":"base64_upload","record_schema":{"type":"object"}}`,
			want: "base64_upload",
		},
		{
			name: "spec without body_type",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low",
				"base64_upload":{"source_field":"p","content_field":"f","max_decoded_bytes":10},
				"record_schema":{"type":"object"}}`,
			want: "base64_upload",
		},
		{
			name: "missing content_field",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low","body_type":"base64_upload",
				"base64_upload":{"source_field":"p","max_decoded_bytes":10},
				"record_schema":{"type":"object"}}`,
			want: "content_field",
		},
		{
			name: "non-positive max_decoded_bytes",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low","body_type":"base64_upload",
				"base64_upload":{"source_field":"p","content_field":"f","max_decoded_bytes":0},
				"record_schema":{"type":"object"}}`,
			want: "max_decoded_bytes",
		},
		{
			name: "unknown source mode",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low","body_type":"base64_upload",
				"base64_upload":{"source":"ftp","source_field":"p","content_field":"f","max_decoded_bytes":10},
				"record_schema":{"type":"object"}}`,
			want: "source",
		},
		{
			name: "source_field equals content_field in path mode",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low","body_type":"base64_upload",
				"base64_upload":{"source":"path","source_field":"f","content_field":"f","max_decoded_bytes":10},
				"record_schema":{"type":"object"}}`,
			want: "source_field",
		},
		{
			name: "unsatisfiable encoded bound",
			action: `{"name":"a","kind":"create","method":"POST","path":"/x","risk":"low","body_type":"base64_upload",
				"base64_upload":{"source_field":"p","content_field":"f","max_decoded_bytes":1000,"max_encoded_bytes":4},
				"record_schema":{"type":"object"}}`,
			want: "max_encoded_bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(base64UploadBundleFS(t, tt.action), "acme")
			if err == nil {
				t.Fatal("want load error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error should mention %q, got %v", tt.want, err)
			}
		})
	}
}
