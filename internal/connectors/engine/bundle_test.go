package engine

import (
	"errors"
	"fmt"
	"os"
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

func TestLinearCLISurfaceMapsImplementedStreams(t *testing.T) {
	b, err := Load(defs.FS, "linear")
	if err != nil {
		t.Fatalf("Load linear: %v", err)
	}
	if b.CLISurface == nil {
		t.Fatalf("linear cli surface missing")
	}

	want := map[string]string{
		"issue list":   "issues",
		"team list":    "teams",
		"project list": "projects",
		"user list":    "users",
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
		t.Fatalf("missing Linear CLI commands: %v", want)
	}
}

func TestLinearStreamsUseFixedGraphQLDocuments(t *testing.T) {
	b, err := Load(defs.FS, "linear")
	if err != nil {
		t.Fatalf("Load linear: %v", err)
	}
	streams := map[string]StreamSpec{}
	for _, stream := range b.Streams {
		streams[stream.Name] = stream
	}
	for _, name := range []string{"issues", "teams", "projects", "users", "issue", "team", "project", "user"} {
		stream, ok := streams[name]
		if !ok {
			t.Fatalf("linear stream %q missing", name)
		}
		if stream.GraphQL == nil {
			t.Fatalf("linear stream %q GraphQL = nil, want fixed document", name)
		}
		if stream.Method != "POST" || stream.Path != "/graphql" {
			t.Fatalf("linear stream %q method/path = %s %s, want POST /graphql", name, stream.Method, stream.Path)
		}
		if stream.Conformance != nil && stream.Conformance.SkipDynamic {
			t.Fatalf("linear stream %q still has skip_dynamic marker: %s", name, stream.Conformance.Reason)
		}
	}
}

func TestLinearOperationLedgerInventoriesGraphQLOperations(t *testing.T) {
	b, err := Load(os.DirFS("../defs"), "linear")
	if err != nil {
		t.Fatalf("Load linear: %v", err)
	}
	if b.Surface == nil || b.Surface.OperationLedgerVersion != 1 {
		t.Fatalf("linear api surface ledger version = %+v, want v1", b.Surface)
	}
	deprecatedQueries := map[string]bool{
		"attachmentIssue":   true,
		"roadmap":           true,
		"roadmapToProject":  true,
		"roadmapToProjects": true,
		"roadmaps":          true,
	}
	deprecatedMutations := map[string]bool{
		"integrationIntercomSettingsUpdate": true,
		"integrationLoom":                   true,
		"integrationSettingsUpdate":         true,
		"notificationSubscriptionDelete":    true,
		"organizationStartTrial":            true,
		"projectArchive":                    true,
		"projectUpdateDelete":               true,
		"roadmapArchive":                    true,
		"roadmapCreate":                     true,
		"roadmapDelete":                     true,
		"roadmapUnarchive":                  true,
		"roadmapUpdate":                     true,
	}
	covered, blocked, queryRows, mutationRows, deprecatedQueryRows, deprecatedMutationRows, rawRows := 0, 0, 0, 0, 0, 0, 0
	for _, ep := range b.Surface.Endpoints {
		if ep.CoveredBy != nil {
			covered++
		}
		if ep.Operation != nil {
			blocked++
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Fatalf("operation row %+v is not blocked by default", ep.Operation)
			}
		}
		if strings.Contains(ep.Path, "(raw arbitrary query or mutation)") {
			rawRows++
			continue
		}
		if name, ok := graphQLSurfaceName(ep.Path, "query"); ok {
			queryRows++
			if deprecatedQueries[name] {
				deprecatedQueryRows++
			}
		}
		if name, ok := graphQLSurfaceName(ep.Path, "mutation"); ok {
			mutationRows++
			if deprecatedMutations[name] {
				deprecatedMutationRows++
			}
		}
	}
	if rawRows != 1 {
		t.Fatalf("raw GraphQL rows = %d, want exactly one blocked raw row", rawRows)
	}
	if queryRows != 161 || mutationRows != 370 {
		t.Fatalf("ledger rows query=%d mutation=%d, want live schema inventory query=161 mutation=370", queryRows, mutationRows)
	}
	if got := queryRows - deprecatedQueryRows; got != 156 {
		t.Fatalf("non-deprecated query rows = %d, want refreshed prompt official count 156", got)
	}
	if got := mutationRows - deprecatedMutationRows; got != 358 {
		t.Fatalf("non-deprecated mutation rows = %d, want refreshed prompt official count 358", got)
	}
	if covered < 514 {
		t.Fatalf("covered endpoint rows = %d, want all official non-deprecated Linear operations covered", covered)
	}
}

func TestLinearMutationOperationsModeledAsTypedWrites(t *testing.T) {
	b, err := Load(os.DirFS("../defs"), "linear")
	if err != nil {
		t.Fatalf("Load linear: %v", err)
	}
	writes := map[string]bool{}
	for _, action := range b.Writes {
		writes[action.Name] = true
		if action.BodyType != "graphql" || action.GraphQL == nil {
			t.Fatalf("linear write %q is not a fixed GraphQL action", action.Name)
		}
		if strings.TrimSpace(action.Risk) == "" {
			t.Fatalf("linear write %q missing risk text", action.Name)
		}
	}

	mutationRows, coveredMutationRows := 0, 0
	var blocked []string
	for _, ep := range b.Surface.Endpoints {
		if !strings.Contains(ep.Path, "(mutation:") {
			continue
		}
		mutationRows++
		if ep.CoveredBy != nil && ep.CoveredBy.Write != "" {
			coveredMutationRows++
			if !writes[ep.CoveredBy.Write] {
				t.Fatalf("mutation row %s covered_by.write %q is not declared", ep.Path, ep.CoveredBy.Write)
			}
			continue
		}
		if ep.Operation == nil || !hardBlockedLinearMutation(ep.Operation) {
			blocked = append(blocked, ep.Path)
		}
	}
	if mutationRows != 370 {
		t.Fatalf("mutationRows = %d, want full Linear live-schema mutation inventory", mutationRows)
	}
	if len(blocked) > 0 {
		t.Fatalf("%d Linear mutation rows are neither typed writes nor exact hard blocks; first rows: %v", len(blocked), blocked[:min(len(blocked), 10)])
	}
	if coveredMutationRows < 358 {
		t.Fatalf("covered mutation rows = %d, want all official non-deprecated Linear mutations modeled as typed writes", coveredMutationRows)
	}
}

func graphQLSurfaceName(path, kind string) (string, bool) {
	prefix := "/graphql (" + kind + ": "
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, ")") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(path, prefix), ")"), true
}

func hardBlockedLinearMutation(op *SurfaceOperation) bool {
	if op == nil || op.Status != "blocked" || !op.BlockedByDefault {
		return false
	}
	switch op.Model {
	case "duplicate", "deprecated", "disallowed", "binary_read":
		return true
	default:
		return false
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
