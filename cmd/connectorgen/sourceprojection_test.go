package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceProjection_AddChangeDeletePropagatesToEverySurface(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{
  "schema_version": 1,
  "actions": [{
    "name": "items", "kind": "custom", "method": "POST", "path": "/items/{{ record.owner }}",
    "path_fields": ["owner"], "body_type": "json", "body_fields": ["stale"],
    "record_schema": {"type":"object","additionalProperties":false,"properties":{"stale":{"type":"string"}}},
    "risk": "standard"
  }]
}`)
	writeProjectionFixture(t, cliPath, `{
  "schema_version": 1,
  "commands": [
    {"path":"items create","summary":"create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"stale","type":"string","maps_to":"record.stale"}]},
    {"path":"items add","summary":"alias","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"stale","type":"string","maps_to":"record.stale"}]}
  ]
}`)

	operation := sourceProjectionTestOperation()
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project added contract: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 2 || stats.Missing != 0 {
		t.Fatalf("add stats = %+v", stats)
	}
	writes := readProjectionFixture(t, writesPath)
	cli := readProjectionFixture(t, cliPath)
	for _, field := range []string{"owner", "mode", "name", "meta", "tags"} {
		if !strings.Contains(writes, `"`+field+`"`) {
			t.Fatalf("generated write schema omitted %q:\n%s", field, writes)
		}
	}
	if strings.Count(cli, `"maps_to": "record.name"`) != 2 || strings.Count(cli, `"maps_to": "record.mode"`) != 2 {
		t.Fatalf("semantic aliases did not receive the same generated fields:\n%s", cli)
	}
	if !strings.Contains(writes, `"additionalProperties": false`) || !strings.Contains(writes, `"type": [`) {
		t.Fatalf("nested/nullable contract was not retained as a closed schema:\n%s", writes)
	}

	operation.Request.Query = []sourceParameterDescriptor{{Name: "limit", Schema: map[string]any{"type": "integer"}}}
	operation.Request.Body.Schema = map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"title"},
		"properties":           map[string]any{"title": map[string]any{"type": "string", "maxLength": 32}},
	}
	stats, err = projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project changed contract: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 2 {
		t.Fatalf("change stats = %+v", stats)
	}
	cli = readProjectionFixture(t, cliPath)
	if strings.Contains(cli, `record.mode`) || strings.Contains(cli, `record.name`) || strings.Count(cli, `"maps_to": "record.title"`) != 2 {
		t.Fatalf("changed source fields did not replace stale alias fields:\n%s", cli)
	}

	operation.Request.Query = nil
	operation.Request.Body = nil
	operation.Request.MediaType = ""
	stats, err = projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project deleted contract: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 2 {
		t.Fatalf("delete stats = %+v", stats)
	}
	writes = readProjectionFixture(t, writesPath)
	cli = readProjectionFixture(t, cliPath)
	if strings.Contains(writes, `"body_fields"`) || !strings.Contains(writes, `"body_type": "none"`) || strings.Count(cli, `"maps_to": "record.owner"`) != 2 {
		t.Fatalf("deleted source fields survived projection:\nwrites=%s\ncli=%s", writes, cli)
	}
}

func TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly(t *testing.T) {
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{
  "schema_version": 1,
  "actions": [{
    "name": "create_webhook", "kind": "create", "method": "POST", "path": "/webhook",
    "record_schema": {"type":"object","additionalProperties":false,"properties":{}},
    "risk": "standard"
  }]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{
  "schema_version": 1,
  "commands": [{
    "path":"webhook create","summary":"create","intent":"reverse_etl","availability":"implemented","write":"create_webhook","flags":[]
  }]
}`)
	operation := sourceOperationDescriptor{
		Connector: "circleci", SourceID: "createWebhook", Method: "POST", Path: "/webhook",
		Request: sourceRequestDescriptor{
			Query: []sourceParameterDescriptor{{Name: "callback_token", Schema: map[string]any{"type": "string", "x-secret": true}}},
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
				"type": "object", "additionalProperties": false,
				"properties": map[string]any{
					"signing_secret": map[string]any{"type": "string", "x-secret": true},
					"secure_payload": map[string]any{"type": "object", "properties": map[string]any{
						"token": map[string]any{"type": "string", "x-secret": true},
					}},
					"name": map[string]any{"type": "string"},
				},
			}},
			MediaType: "application/json",
		},
	}
	if _, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false); err != nil {
		t.Fatalf("project CircleCI webhook: %v", err)
	}

	var surface engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"))), &surface); err != nil {
		t.Fatalf("decode projected CircleCI CLI surface: %v", err)
	}
	if len(surface.Commands) != 1 {
		t.Fatalf("CircleCI commands = %d, want 1", len(surface.Commands))
	}
	flags := map[string]engine.CLIFlag{}
	for _, flag := range surface.Commands[0].Flags {
		flags[flag.Name] = flag
	}
	for _, name := range []string{"signing-secret", "callback-token", "secure-payload"} {
		if !flags[name].EnvOnly {
			t.Fatalf("CircleCI %s env_only = false, want true", name)
		}
	}
	if flags["name"].EnvOnly {
		t.Fatal("non-secret CircleCI webhook name was projected env_only")
	}
	writesRaw := readProjectionFixture(t, filepath.Join(bundleDir, "writes.json"))
	if !strings.Contains(writesRaw, `"x-secret": true`) {
		t.Fatal("projected CircleCI write schema lost its declaration-owned x-secret marker")
	}
	var writes struct {
		Actions []engine.WriteAction `json:"actions"`
	}
	if err := json.Unmarshal([]byte(writesRaw), &writes); err != nil {
		t.Fatalf("decode projected CircleCI writes: %v", err)
	}
	if len(writes.Actions) != 1 {
		t.Fatalf("CircleCI write actions = %d, want 1", len(writes.Actions))
	}
	if findings := checkCLISurfaceEnvOnlyFlags(engine.Bundle{Name: "circleci"}, 0, surface.Commands[0], nil, map[string]engine.WriteAction{"create_webhook": writes.Actions[0]}); len(findings) != 0 {
		t.Fatalf("projected CircleCI webhook env_only validation findings = %+v, want none", findings)
	}
}

func TestSourceProjection_MissingOperationOrFieldFailsValidateAndSurfaceCheck(t *testing.T) {
	operation := sourceProjectionTestOperation()
	bundle := engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{}}
	if findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "no executable action") {
		t.Fatalf("missing operation findings = %+v", findings)
	}

	bundleDir := filepath.Join(t.TempDir(), "alpha")
	if err := os.MkdirAll(filepath.Join(bundleDir, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "alpha-operation-source-lock.json"), `{}`)
	descriptorRaw, err := json.Marshal(sourceImportDescriptorDocument{SchemaVersion: 2, Operations: []sourceOperationDescriptor{operation}})
	if err != nil {
		t.Fatal(err)
	}
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "alpha-operation-descriptor.json"), string(descriptorRaw))
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[{"name":"items","kind":"custom","method":"POST","path":"/items/{{ record.owner }}","path_fields":["owner"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"owner":{"type":"string"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[{"path":"items create","summary":"create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"owner","type":"string","maps_to":"record.owner","required":true}]}]}`)
	stats, err := syncCheckedInSourceProjection(bundleDir, "alpha", true)
	if err != nil {
		t.Fatalf("surface source projection check: %v", err)
	}
	if !stats.Changed() || stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("missing field was not reported as source projection drift: %+v", stats)
	}
}

func TestSourceProjectionRequiresReachableRESTReadOrConcreteGap(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/widgets",
	}
	empty := engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{}}
	if findings := validateSourceExecutableCoverage(empty, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "no reachable executable operation") {
		t.Fatalf("missing REST read findings = %+v", findings)
	}

	reachable := engine.Bundle{
		Name: "alpha",
		Operations: []engine.OperationSpec{{
			ID: "alpha.widgets.list", Kind: "rest_read", REST: &engine.RESTOperationSpec{Method: "GET", Path: "/widgets", MaxBytes: 1024},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widgets list", Availability: "implemented", Operation: "alpha.widgets.list",
		}}},
	}
	if findings := validateSourceExecutableCoverage(reachable, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 0 {
		t.Fatalf("reachable REST read findings = %+v", findings)
	}

	source.Runtime.Gaps = []sourceContractGap{{Foundation: "typed-read-foundation-r1", Location: "response", Reason: "provider response is not yet representable"}}
	if findings := validateSourceExecutableCoverage(empty, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 0 {
		t.Fatalf("concrete deferred REST read findings = %+v", findings)
	}
}

func TestSourceProjectionCountsDeclaredPaginationParametersAsReachableInputs(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/widgets",
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{{
			Name: "page", Schema: map[string]any{"type": "integer"},
		}}},
	}
	bundle := engine.Bundle{
		Name: "alpha",
		Operations: []engine.OperationSpec{{
			ID: "alpha.widgets.list", Kind: "rest_read",
			REST: &engine.RESTOperationSpec{
				Method: "GET", Path: "/widgets", MaxBytes: 1024,
				PaginationParameters: []engine.OperationParameter{{Name: "page", In: "query"}},
			},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widgets list", Availability: "implemented", Operation: "alpha.widgets.list",
		}}},
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 0 {
		t.Fatalf("declared pagination parameter did not close source reachability: %+v", findings)
	}
}

func TestSourceProjectionDowngradesUnboundImplementedAPICommandForSourceGap(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/widgets",
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
			Foundation: sourceOperationExecutionFoundation,
			Location:   "source operation alpha.widgets.list",
			Reason:     "locked provider operation has no declaration-owned executable stream, direct-read, binary, or status route",
		}}},
	}
	bundle := engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
		Path: "widgets list", Intent: "direct_read", Availability: "implemented",
		APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/widgets"}},
	}}}}
	if findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "retains an unresolved source-bound gap") {
		t.Fatalf("unbound implemented API command was accepted: %+v", findings)
	}

	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{
  "schema_version": 1,
  "commands": [{
    "path": "widgets list",
    "summary": "list widgets",
    "intent": "direct_read",
    "availability": "implemented",
    "api_surface": [{"method":"GET","path":"/widgets"}]
  }]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{
  "api": "alpha",
  "endpoints": [{
    "method": "GET",
    "path": "/widgets",
    "covered_by": {"direct_read":"widgets list"}
  }]
}`)
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{source}}, false)
	if err != nil {
		t.Fatalf("project source-bound read gap: %v", err)
	}
	if stats.CLI != 1 {
		t.Fatalf("projected CLI stats = %+v, want one downgraded command", stats)
	}
	projected := readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"))
	if !strings.Contains(projected, `"availability": "partial"`) || !strings.Contains(projected, source.SourceID) || !strings.Contains(projected, "declaration-owned executable") {
		t.Fatalf("unbound command did not become source-bound partial capability:\n%s", projected)
	}
	projectedSurface := readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))
	if strings.Contains(projectedSurface, `"covered_by"`) || !strings.Contains(projectedSurface, `"model": "direct_read"`) || !strings.Contains(projectedSurface, source.SourceID) || !strings.Contains(projectedSurface, "Named dependency:") {
		t.Fatalf("source-bound partial command retained executable API coverage:\n%s", projectedSurface)
	}
}

func TestSourceProjectionDoesNotBlockReadForUnusedOptionalAmbiguousParameter(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/orgs/{org}/widgets",
		Request: sourceRequestDescriptor{
			Path: []sourceParameterDescriptor{{Name: "org", Required: true, Schema: map[string]any{"type": "string"}}},
			Query: []sourceParameterDescriptor{{Name: "has", Required: false, Schema: map[string]any{"oneOf": []any{
				map[string]any{"type": "string"},
				map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			}}}},
		},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1",
			Location:   "parameter has",
			Reason:     "ambiguous request schema uses oneOf",
		}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{source}}
	if blocked := sourceProjectionBlockedReadSources(result); len(blocked) != 0 {
		t.Fatalf("unused optional ambiguous query parameter blocked executable read: %+v", blocked)
	}
	if reachable := sourceProjectionReachableReadSources(result); reachable[source.SourceID].SourceID != source.SourceID {
		t.Fatalf("read with only an unused optional ambiguous parameter was not reachable: %+v", reachable)
	}
}

func TestSourceProjectionDoesNotBlockReadForOmittedOptionalRequestBody(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.get", Method: "get", Path: "/widgets/{id}",
		Request: sourceRequestDescriptor{
			Path: []sourceParameterDescriptor{{Name: "id", Required: true, Schema: map[string]any{"type": "string"}}},
			Body: &sourceRequestBodyDescriptor{Required: false, Schema: true},
		},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1",
			Location:   "request body",
			Reason:     "unsupported openapi boolean schema",
		}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{source}}
	if blocked := sourceProjectionBlockedReadSources(result); len(blocked) != 0 {
		t.Fatalf("omitted optional request body blocked executable read: %+v", blocked)
	}
	if reachable := sourceProjectionReachableReadSources(result); reachable[source.SourceID].SourceID != source.SourceID {
		t.Fatalf("read with an omitted optional body was not reachable: %+v", reachable)
	}
}

func TestSourceProjectionNormalizesOnlyOptionalReadSchemaGaps(t *testing.T) {
	result := sourceImportResult{Operations: []sourceOperationDescriptor{
		{
			Connector: "alpha", SourceID: "alpha.widgets.list", Method: "GET", Path: "/widgets",
			Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{{Name: "has", Required: false}}},
			Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
				Foundation: "cli-request-schema-foundation-r1", Location: "parameter has", Reason: "ambiguous request schema uses oneOf",
			}}},
		},
		{
			Connector: "alpha", SourceID: "alpha.widgets.get", Method: "GET", Path: "/widgets/{id}",
			Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{Name: "id", Required: true}}},
			Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
				Foundation: "cli-request-schema-foundation-r1", Location: "parameter id", Reason: "ambiguous request schema uses oneOf",
			}}},
		},
	}}

	sourceProjectionNormalizeNonBlockingReadGaps(&result)
	if got := result.Operations[0].Runtime; got.MergeBlocked || len(got.Gaps) != 0 {
		t.Fatalf("optional omitted input runtime = %+v, want no availability gap", got)
	}
	if got := result.Operations[1].Runtime; !got.MergeBlocked || len(got.Gaps) != 1 {
		t.Fatalf("required input runtime = %+v, want retained gap", got)
	}
}

func TestSourceProjectionKeepsIndependentSurfaceCoverageWhenBlockingReadCommand(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.get", Method: "get", Path: "/widgets/{id}",
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
			Foundation: sourceOperationExecutionFoundation,
			Location:   "source operation alpha.widgets.get",
			Reason:     "locked provider operation has no field-complete declaration-owned executable route",
		}}},
	}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{
  "schema_version": 1,
  "commands": [{
    "path": "widgets get",
    "summary": "get widget",
    "intent": "direct_read",
    "availability": "implemented",
    "api_surface": [{"method":"GET","path":"/widgets/{id}"}]
  }]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{
  "api": "alpha",
  "endpoints": [{
    "method": "GET",
    "path": "/widgets/{id}",
    "covered_by": {"stream":"widgets", "direct_read":"widgets get"}
  }]
}`)
	if _, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{source}}, false); err != nil {
		t.Fatalf("project source-bound read gap with an independent stream: %v", err)
	}
	projectedSurface := readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))
	if !strings.Contains(projectedSurface, `"stream": "widgets"`) || strings.Contains(projectedSurface, `"direct_read"`) || strings.Contains(projectedSurface, `"operation"`) {
		t.Fatalf("blocked direct read changed independent stream coverage:\n%s", projectedSurface)
	}
}

func TestSourceProjectionRequiresReachableGraphQLRootOrConcreteGap(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.graphql.query.widgets", Protocol: "graphql",
		GraphQL: &sourceGraphQLOperationDescriptor{Root: "Query", Name: "widgets", Line: 1, Signature: "widgets: [Widget!]!"},
	}
	empty := engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{}}
	if findings := validateSourceExecutableCoverage(empty, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "no reachable executable operation") {
		t.Fatalf("missing GraphQL root findings = %+v", findings)
	}

	reachable := engine.Bundle{
		Name: "alpha",
		Operations: []engine.OperationSpec{{
			ID: "alpha.graphql.query.widgets", Kind: "graphql_query", OutputPolicy: "json", GraphQL: &engine.GraphQLOperationSpec{Document: "query Widgets { widgets { id } }", OperationName: "Widgets", Path: "/graphql", MaxBytes: 1024},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widgets list", Availability: "implemented", Operation: "alpha.graphql.query.widgets",
		}}},
	}
	if findings := validateSourceExecutableCoverage(reachable, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 0 {
		t.Fatalf("reachable GraphQL root findings = %+v", findings)
	}

	source.Runtime.Gaps = []sourceContractGap{{Foundation: "graphql-output-foundation-r1", Location: "selection", Reason: "source selection is not yet representable"}}
	if findings := validateSourceExecutableCoverage(empty, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 0 {
		t.Fatalf("concrete deferred GraphQL root findings = %+v", findings)
	}
}

func TestSourceProjectionAnnotatesUnreachableReadWithConcreteSourceGap(t *testing.T) {
	result := sourceImportResult{Operations: []sourceOperationDescriptor{{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/widgets",
	}}}
	sourceProjectionAnnotateUnreachableReadGaps(engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{}}, &result)
	if !result.Operations[0].Runtime.MergeBlocked || len(result.Operations[0].Runtime.Gaps) != 1 {
		t.Fatalf("unreachable read runtime = %+v, want one source-bound gap", result.Operations[0].Runtime)
	}
	gap := result.Operations[0].Runtime.Gaps[0]
	if gap.Foundation != sourceOperationExecutionFoundation || !strings.Contains(gap.Location, result.Operations[0].SourceID) || !strings.Contains(gap.Reason, "declaration-owned") {
		t.Fatalf("unreachable read gap = %+v, want exact source-bound execution gap", gap)
	}

	result = sourceImportResult{Operations: []sourceOperationDescriptor{{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/widgets",
	}}}
	reachable := engine.Bundle{
		Name: "alpha",
		Operations: []engine.OperationSpec{{
			ID: "alpha.widgets.list", Kind: "rest_read", REST: &engine.RESTOperationSpec{Method: "GET", Path: "/widgets", MaxBytes: 1024},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{Path: "widgets list", Availability: "implemented", Operation: "alpha.widgets.list"}}},
	}
	sourceProjectionAnnotateUnreachableReadGaps(reachable, &result)
	if result.Operations[0].Runtime.MergeBlocked || len(result.Operations[0].Runtime.Gaps) != 0 {
		t.Fatalf("reachable read was marked deferred: %+v", result.Operations[0].Runtime)
	}

	result = sourceImportResult{Operations: []sourceOperationDescriptor{{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "get", Path: "/widgets",
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{{Name: "page", Required: true, Schema: map[string]any{"type": "integer"}}}},
	}}}
	partial := engine.Bundle{
		Name: "alpha",
		Operations: []engine.OperationSpec{{
			ID: "alpha.widgets.list", Kind: "rest_read", REST: &engine.RESTOperationSpec{Method: "GET", Path: "/widgets", MaxBytes: 1024},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widgets list", Availability: "implemented", Intent: "direct_read", Operation: "alpha.widgets.list",
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/widgets"}},
		}}},
	}
	sourceProjectionAnnotateUnreachableReadGaps(partial, &result)
	if !result.Operations[0].Runtime.MergeBlocked || !sourceOperationHasFoundationGap(result.Operations[0], sourceOperationExecutionFoundation) {
		t.Fatalf("incomplete declared read was not marked source-bound partial: %+v", result.Operations[0].Runtime)
	}
}

func TestSourceProjectionRetainsRequiredFieldCompleteSourceBoundDirectRead(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "GET", Path: "/widgets/{owner}",
		Request: sourceRequestDescriptor{
			Path:  []sourceParameterDescriptor{{Name: "owner", Required: true, Schema: map[string]any{"type": "string"}}},
			Query: []sourceParameterDescriptor{{Name: "state", Required: false, Schema: map[string]any{"type": "string"}}},
		},
	}
	spec, err := engine.CompileSchema(json.RawMessage(`{
  "type":"object","additionalProperties":false,
  "required":["owner"],"properties":{"owner":{"type":"string"}}
}`))
	if err != nil {
		t.Fatal(err)
	}
	bundle := engine.Bundle{
		Name: "alpha", Spec: spec,
		Certification: &engine.CertificationSpec{DirectReadGeneration: &engine.CertificationReadCandidateGeneration{Cohorts: []engine.CertificationReadCandidateCohort{{
			Name: "fixture", CommandCount: 1, Commands: []string{"widgets list"},
		}}}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widgets list", Intent: "direct_read", Availability: "partial",
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/widgets/{owner}"}},
			Notes:      sourceProjectionBlockedReadCommandNote(source.SourceID),
		}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{source}}
	sourceProjectionAnnotateUnreachableReadGaps(bundle, &result)
	if sourceProjectionHasBlockingGap(result.Operations[0].Runtime.Gaps) {
		t.Fatalf("required-field-complete source-bound direct read was marked unreachable: %+v", result.Operations[0].Runtime.Gaps)
	}
}

func TestSourceProjectionRestoresRequiredPathFlagForSourceBoundDirectRead(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.get", Method: "GET", Path: "/accounts/{account}/widgets/{widget}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "account", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "widget", Required: true, Schema: map[string]any{"type": "string"}},
		}},
	}
	spec, err := engine.CompileSchema(json.RawMessage(`{
  "type":"object","additionalProperties":false,
  "required":["account"],"properties":{"account":{"type":"string"}}
}`))
	if err != nil {
		t.Fatal(err)
	}
	bundle := engine.Bundle{
		Name: "alpha", Spec: spec,
		Certification: &engine.CertificationSpec{DirectReadGeneration: &engine.CertificationReadCandidateGeneration{Cohorts: []engine.CertificationReadCandidateCohort{{
			Name: "fixture", CommandCount: 1, Commands: []string{"widgets get"},
		}}}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widgets get", Intent: "direct_read", Availability: "partial",
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/accounts/{account}/widgets/{widget}"}},
			Notes:      sourceProjectionBlockedReadCommandNote(source.SourceID),
		}}},
	}
	if changed := sourceProjectionRestoreSourceBoundDirectReadPathFlags(&bundle, sourceImportResult{Operations: []sourceOperationDescriptor{source}}); changed != 1 {
		t.Fatalf("restored source-bound direct-read path flags = %d, want 1", changed)
	}
	flags := bundle.CLISurface.Commands[0].Flags
	if len(flags) != 1 || flags[0].Name != "widget" || flags[0].MapsTo != "path.widget" || !flags[0].Required {
		t.Fatalf("restored source-bound direct-read path flags = %+v, want required path.widget", flags)
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{source}}
	sourceProjectionAnnotateUnreachableReadGaps(bundle, &result)
	if sourceProjectionHasBlockingGap(result.Operations[0].Runtime.Gaps) {
		t.Fatalf("restored required path flag left source-bound direct read unreachable: %+v", result.Operations[0].Runtime.Gaps)
	}
}

func TestSourceProjectionRequiresExistingPathFlagForRestoredNonCandidateDirectRead(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.get", Method: "GET", Path: "/accounts/{account}/widgets/{widget}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "account", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "widget", Required: true, Schema: map[string]any{"type": "string"}},
		}},
	}
	spec, err := engine.CompileSchema(json.RawMessage(`{
  "type":"object","additionalProperties":false,
  "required":["account"],"properties":{"account":{"type":"string"}}
}`))
	if err != nil {
		t.Fatal(err)
	}
	var cli orderedJSON
	if err := json.Unmarshal([]byte(`{
  "schema_version":1,
  "commands":[{
    "path":"widgets get","intent":"direct_read","availability":"partial",
    "api_surface":[{"method":"GET","path":"/accounts/{account}/widgets/{widget}"}],
    "flags":[{"name":"widget","type":"string","maps_to":"path.widget"}],
    "notes":"Blocked: locked source operation alpha.widgets.get has no declaration-owned executable stream, direct-read, binary, or status route."
  }]
}`), &cli); err != nil {
		t.Fatal(err)
	}
	if changed := sourceProjectionRestoreSourceBoundDirectReadPathFlagObjects(cli.root, spec, sourceImportResult{Operations: []sourceOperationDescriptor{source}}); changed != 1 {
		t.Fatalf("restored noncandidate direct-read path flags = %d, want 1", changed)
	}
	command := arrayField(cli.root, "commands")[0].(*orderedObject)
	flag := arrayField(command, "flags")[0].(*orderedObject)
	if required, _ := flag.get("required"); required != true {
		t.Fatalf("restored noncandidate path flag required = %#v, want true", required)
	}
}

func TestSourceProjection_DerivesHyphenatedPathFieldsFromExecutableTemplate(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{
  "schema_version": 1,
  "actions": [{
    "name": "team_membership", "kind": "update", "method": "PUT",
    "path": "/enterprises/{{ record.enterprise }}/teams/{{ record.enterprise-team }}/memberships/{{ record.username }}",
    "path_fields": ["enterprise", "username"], "body_type": "none",
    "record_schema": {"type":"object","additionalProperties":false,"properties":{}},
    "risk": "standard"
  }]
}`)
	writeProjectionFixture(t, cliPath, `{
  "schema_version": 1,
  "commands": [{"path":"team membership add","summary":"add","intent":"reverse_etl","availability":"implemented","write":"team_membership","flags":[]}]
}`)
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "team-memberships/add", Method: "put",
		Path: "/enterprises/{enterprise}/teams/{enterprise-team}/memberships/{username}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "enterprise", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "enterprise-team", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "username", Required: true, Schema: map[string]any{"type": "string"}},
		}},
	}
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project hyphenated path contract: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("projection stats = %+v, want one write and CLI repair", stats)
	}
	writes := readProjectionFixture(t, writesPath)
	cli := readProjectionFixture(t, cliPath)
	if !strings.Contains(writes, `"enterprise-team"`) || !strings.Contains(cli, `"maps_to": "record.enterprise-team"`) {
		t.Fatalf("hyphenated provider path field was not restored:\nwrites=%s\ncli=%s", writes, cli)
	}
}

func TestSourceProjectionSkipsUnprojectableGapAndRepairsCompleteAction(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"items","method":"POST","path":"/items/{{ record.owner }}","record_schema":{"type":"object","additionalProperties":false,"properties":{}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[]}]}`)

	complete := sourceProjectionTestOperation()
	blocked := sourceProjectionTestOperation()
	blocked.SourceID = "items/blocked"
	blocked.Path = "/blocked/{owner}"
	blocked.Request.Body.Schema = map[string]any{
		"type": "object", "additionalProperties": map[string]any{"type": "string"},
	}
	blocked.Runtime.Gaps = []sourceContractGap{{
		Foundation: "cli-request-schema-foundation-r1",
		Location:   "request body",
		Reason:     "unbounded request schema object has dynamic additionalProperties",
	}}

	if _, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{complete, blocked}}, false); err != nil {
		t.Fatalf("projection must skip an explicitly gap-bound operation while repairing a complete action: %v", err)
	}
	if got := readProjectionFixture(t, cliPath); !strings.Contains(got, `"maps_to": "record.name"`) {
		t.Fatalf("complete action did not receive its declaration-owned CLI field: %s", got)
	}
}

func TestSourceProjectionNewCommandMatchesDeclaredConfirmationLifecycle(t *testing.T) {
	operation := sourceOperationDescriptor{SourceID: "widgets/create", Method: "POST", Path: "/widgets"}

	tests := []struct {
		name             string
		method           string
		confirmation     bool
		legacyConfirm    bool
		wantApprovalText string
	}{
		{
			name:             "safe action keeps preview optional",
			method:           "POST",
			wantApprovalText: "Reverse ETL writes require plan, approval, execute; preview is optional.",
		},
		{
			name:             "confirmed action requires preview",
			method:           "POST",
			confirmation:     true,
			wantApprovalText: "Reverse ETL writes require plan, preview, approval, execute.",
		},
		{
			name:             "legacy confirmation requires preview",
			method:           "POST",
			legacyConfirm:    true,
			wantApprovalText: "Reverse ETL writes require plan, preview, approval, execute.",
		},
		{
			name:             "delete requires preview",
			method:           "DELETE",
			wantApprovalText: "Reverse ETL writes require plan, preview, approval, execute.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := newOrderedObject()
			action.set("name", "widgets")
			action.set("method", tt.method)
			action.set("risk", "standard")
			if tt.confirmation {
				confirmation := newOrderedObject()
				confirmation.set("kind", "destructive")
				action.set("confirmation", confirmation)
			}
			if tt.legacyConfirm {
				action.set("confirm", "destructive")
			}

			command := sourceProjectionNewCommand(operation, action)
			if got := stringField(command, "approval"); got != tt.wantApprovalText {
				t.Fatalf("approval = %q, want %q", got, tt.wantApprovalText)
			}
		})
	}
}

func TestSourceProjectionGapCompletesExistingClosedActionCLI(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	// Legacy named actions can have a closed record contract but body_type:none.
	// The source-declared JSON body must promote only the named non-path field;
	// it must not turn into a generic body escape hatch.
	writes := `{"schema_version":1,"actions":[{"name":"items","method":"POST","path":"/items/{{ record.owner }}","body_type":"none","record_schema":{"type":"object","additionalProperties":false,"properties":{"owner":{"type":"string"},"custom_properties":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["property_name"],"properties":{"property_name":{"type":"string"}}}}}},"risk":"standard"}]}`
	writeProjectionFixture(t, writesPath, writes)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"owner","type":"string","maps_to":"record.owner","required":true}]}]}`)

	operation := sourceProjectionTestOperation()
	operation.Request.Query = nil
	operation.Request.Body.Schema = map[string]any{"oneOf": []any{
		map[string]any{"type": "object", "properties": map[string]any{
			"custom_properties": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
		}},
		map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	}}
	operation.Runtime.Gaps = []sourceContractGap{{
		Foundation: "cli-request-schema-foundation-r1",
		Location:   "request body",
		Reason:     "ambiguous request schema uses oneOf",
	}}

	if _, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false); err != nil {
		t.Fatalf("gap-bound projection must repair only the existing closed action CLI: %v", err)
	}
	projectedWrites := readProjectionFixture(t, writesPath)
	if got := projectedWrites; !strings.Contains(got, `"additionalProperties": false`) || !strings.Contains(got, `"custom_properties"`) {
		t.Fatalf("gap-bound projection did not retain a closed declared custom_properties field:\n%s", got)
	}
	var projected struct {
		Actions []struct {
			BodyType   string   `json:"body_type"`
			BodyFields []string `json:"body_fields"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(projectedWrites), &projected); err != nil {
		t.Fatalf("decode projected writes: %v", err)
	}
	if len(projected.Actions) != 1 || projected.Actions[0].BodyType != "json" || !reflect.DeepEqual(projected.Actions[0].BodyFields, []string{"custom_properties"}) {
		t.Fatalf("gap-bound projection body = %#v, want named json custom_properties only", projected.Actions)
	}
	if got := strings.Join(sourceProjectionActionRequired(t, readProjectionFixture(t, writesPath), "items"), ","); got != "owner" {
		t.Fatalf("gap-bound projection required fields = %q, want source-owned path field owner", got)
	}
	if got := readProjectionFixture(t, cliPath); !strings.Contains(got, `"maps_to": "record.custom_properties"`) || !strings.Contains(got, `"type": "json"`) {
		t.Fatalf("existing declared custom_properties field has no typed JSON CLI flag:\n%s", got)
	}
}

func sourceProjectionActionRequired(t *testing.T, raw, name string) []string {
	t.Helper()
	var document struct {
		Actions []struct {
			Name         string `json:"name"`
			RecordSchema struct {
				Required []string `json:"required"`
			} `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(raw), &document); err != nil {
		t.Fatalf("decode projected writes: %v", err)
	}
	for _, action := range document.Actions {
		if action.Name == name {
			return action.RecordSchema.Required
		}
	}
	t.Fatalf("projected action %q is missing", name)
	return nil
}

func TestSourceProjectionPathRequirednessWinsOverOptionalSameNamedBodyField(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"items","method":"PATCH","path":"/items/{{ record.name }}","body_type":"json","body_fields":["name","value"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string"},"value":{"type":"string"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items update","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"name","type":"string","maps_to":"record.name"},{"name":"value","type":"string","maps_to":"record.value"}]}]}`)

	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/update", Method: "patch", Path: "/items/{name}",
		Request: sourceRequestDescriptor{
			Path:      []sourceParameterDescriptor{{Name: "name", Required: true, Schema: map[string]any{"type": "string"}}},
			MediaType: "application/json",
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
				"type": "object", "properties": map[string]any{
					"name":  map[string]any{"type": "string"},
					"value": map[string]any{"type": "string"},
				},
			}},
		},
	}
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project path/body contract: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 1 || stats.Missing != 0 {
		t.Fatalf("projection stats = %+v, want exactly one write and CLI correction", stats)
	}
	if got := strings.Join(sourceProjectionActionRequired(t, readProjectionFixture(t, writesPath), "items"), ","); got != "name" {
		t.Fatalf("required fields = %q, want required path field name", got)
	}
	if got := readProjectionFixture(t, cliPath); !strings.Contains(got, `"maps_to": "record.name"`) || !strings.Contains(got, `"required": true`) {
		t.Fatalf("path field did not remain required on the generated CLI: %s", got)
	}
}

func TestSourceProjectionGapSealsExistingActionRootBeforeAddingTypedJSONFlag(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"repos_create","method":"POST","path":"/orgs/{{ record.org }}/repos","body_type":"json","record_schema":{"type":"object","required":["org","name"],"properties":{"org":{"type":"string"},"name":{"type":"string"},"custom_properties":{"type":"object"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"repos create","intent":"reverse_etl","availability":"implemented","write":"repos_create","flags":[{"name":"org","type":"string","maps_to":"record.org","required":true},{"name":"name","type":"string","maps_to":"record.name","required":true}]}]}`)
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "repos/create", Method: "post", Path: "/orgs/{org}/repos",
		Request: sourceRequestDescriptor{
			MediaType: "application/json",
			Path:      []sourceParameterDescriptor{{Name: "org", Required: true, Schema: map[string]any{"type": "string"}}},
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
				"type": "object", "properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"custom_properties": map[string]any{
						"type": "object", "additionalProperties": map[string]any{"type": "string"},
					},
				},
			}},
		},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1",
			Location:   "request body",
			Reason:     "unbounded request schema object has dynamic additionalProperties",
		}}},
	}
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project gap-bound existing action: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("projection stats = %+v, want sealed write and typed CLI repair", stats)
	}
	writes := readProjectionFixture(t, writesPath)
	if !strings.Contains(writes, `"additionalProperties": false`) {
		t.Fatalf("gap-bound action root stayed open: %s", writes)
	}
	var projected struct {
		Actions []struct {
			RecordSchema struct {
				Properties map[string]map[string]any `json:"properties"`
			} `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(writes), &projected); err != nil {
		t.Fatalf("decode projected write: %v", err)
	}
	customProperties := projected.Actions[0].RecordSchema.Properties["custom_properties"]
	if customProperties["additionalProperties"] != true || customProperties["maxProperties"] != float64(sourceProjectionDefaultObjectProperties) || customProperties["type"] != "object" {
		t.Fatalf("gap-bound named object has no explicit bounded shape: %#v", customProperties)
	}
	cli := readProjectionFixture(t, cliPath)
	if !strings.Contains(cli, `"maps_to": "record.custom_properties"`) || !strings.Contains(cli, `"type": "json"`) || !strings.Contains(cli, `"max_bytes": 1048576`) {
		t.Fatalf("typed bounded custom_properties flag was not generated: %s", cli)
	}
}

func TestSourceProjectionGapKeepsNamedBoundedUnionFieldsReachable(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"items_update","method":"PATCH","path":"/items/{{ record.item_id }}","body_type":"json","record_schema":{"type":"object","additionalProperties":false,"required":["item_id"],"properties":{"item_id":{"type":"integer"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items update","intent":"reverse_etl","availability":"implemented","write":"items_update","flags":[{"name":"item-id","type":"integer","maps_to":"record.item_id","required":true}]}]}`)
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/update", Method: "patch", Path: "/items/{item_id}",
		Request: sourceRequestDescriptor{
			Path:      []sourceParameterDescriptor{{Name: "item_id", Required: true, Schema: map[string]any{"type": "integer"}}},
			MediaType: "application/json",
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
				"type": "object", "additionalProperties": true,
				"properties": map[string]any{
					"name":    map[string]any{"type": "string"},
					"payload": map[string]any{"oneOf": []any{map[string]any{"type": "string"}, map[string]any{"type": "integer"}}},
				},
			}},
		},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1", Location: "request body", Reason: "unbounded request schema object has dynamic additionalProperties",
		}}},
	}

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project bounded source gap: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("projection stats = %+v, want one closed action and one CLI repair", stats)
	}
	writes := readProjectionFixture(t, writesPath)
	if !strings.Contains(writes, `"additionalProperties": false`) || !strings.Contains(writes, `"payload"`) {
		t.Fatalf("gap-bound action did not retain its named union field in a closed root: %s", writes)
	}
	cli := readProjectionFixture(t, cliPath)
	if !strings.Contains(cli, `"maps_to": "record.payload"`) || !strings.Contains(cli, `"type": "json"`) || !strings.Contains(cli, `"max_bytes": 1048576`) {
		t.Fatalf("gap-bound union field did not receive a bounded declaration-owned JSON flag: %s", cli)
	}
	if strings.Contains(cli, `"maps_to": "body"`) || strings.Contains(cli, `"name": "body"`) {
		t.Fatalf("gap projection introduced a generic body flag: %s", cli)
	}
}

// TestSourceProjectionStringUnionKeepsTextCLIAndProviderArms protects the
// ordinary command spelling for source contracts such as GitHub's issue title.
// A source oneOf may retain several provider scalar arms, but one named,
// bounded JSON flag must still admit the explicitly declared string arm as
// regular CLI text rather than becoming a raw JSON/body escape hatch.
func TestSourceProjectionStringUnionKeepsTextCLIAndProviderArms(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"items","method":"POST","path":"/items","body_type":"json","body_fields":["title"],"record_schema":{"type":"object","additionalProperties":false,"required":["title"],"properties":{"title":{"type":"string","maxLength":42}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"title","type":"string","maps_to":"record.title","required":true,"max_bytes":168}]}]}`)
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/create", Method: "post", Path: "/items",
		Request: sourceRequestDescriptor{MediaType: "application/json", Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
			"type": "object", "additionalProperties": false, "required": []any{"title"},
			"properties": map[string]any{"title": map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "string"}}}},
		}}},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1", Location: "request body.properties.title", Reason: "provider scalar oneOf is projected as one bounded named field",
		}}},
	}

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project direct scalar union: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("projection stats = %+v, want schema and bounded bare-string flag repair", stats)
	}
	var projected struct {
		Actions []struct {
			RecordSchema struct {
				Properties map[string]map[string]any `json:"properties"`
			} `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, writesPath)), &projected); err != nil {
		t.Fatalf("decode projected write: %v", err)
	}
	if got := projected.Actions[0].RecordSchema.Properties["title"]["type"]; !reflect.DeepEqual(got, []any{"integer", "string"}) {
		t.Fatalf("provider scalar union = %#v, want integer|string retained", got)
	}
	var surface struct {
		Commands []struct {
			Flags []struct {
				Name            string `json:"name"`
				Type            string `json:"type"`
				MaxBytes        int    `json:"max_bytes"`
				AllowBareString bool   `json:"allow_bare_string"`
			} `json:"flags"`
		} `json:"commands"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, cliPath)), &surface); err != nil {
		t.Fatalf("decode projected CLI: %v", err)
	}
	if got := surface.Commands[0].Flags[0]; got.Name != "title" || got.Type != "json" || got.MaxBytes != sourceProjectionDefaultJSONBytes || !got.AllowBareString {
		t.Fatalf("source string union flag = %+v, want bounded named JSON with bare text", got)
	}
}

func TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"labels","method":"POST","path":"/items/{{ record.item_id }}/labels","body_type":"json","body_fields":["labels"],"record_schema":{"type":"object","additionalProperties":false,"required":["item_id"],"properties":{"item_id":{"type":"integer"},"labels":{"type":"array","items":{"type":"string"}}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[]}`)
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/add-labels", Method: "post", Path: "/items/{item_id}/labels",
		Request: sourceRequestDescriptor{
			Path:      []sourceParameterDescriptor{{Name: "item_id", Required: true, Schema: map[string]any{"type": "integer"}}},
			MediaType: "application/json",
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{"oneOf": []any{
				map[string]any{"type": "object", "properties": map[string]any{"labels": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}},
				map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			}}},
		},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1", Location: "request body", Reason: "ambiguous request schema uses oneOf",
		}}},
	}

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project existing closed action variant: %v", err)
	}
	if stats.CLI != 1 {
		t.Fatalf("projection stats = %+v, want exactly one generated action command", stats)
	}
	cli := readProjectionFixture(t, cliPath)
	if !strings.Contains(cli, `"write": "labels"`) || !strings.Contains(cli, `"maps_to": "record.labels"`) {
		t.Fatalf("existing closed action variant was left unreachable: %s", cli)
	}
	if strings.Contains(cli, `"maps_to": "body"`) || strings.Contains(cli, `"name": "body"`) {
		t.Fatalf("variant projection introduced a generic body flag: %s", cli)
	}
}

func TestSourceProjectionPreservesDeclaredHookFollowupFieldsOutsideProviderBody(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"pulls","method":"POST","path":"/pulls","body_type":"json","hook":"compound","hook_fields":["labels"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"title":{"type":"string"},"labels":{"type":"array","items":{"type":"string"}}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"pulls create","intent":"reverse_etl","availability":"implemented","write":"pulls","flags":[]}]}`)

	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "pulls/create", Method: "post", Path: "/pulls",
		Request: sourceRequestDescriptor{
			MediaType: "application/json",
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
				"type": "object", "properties": map[string]any{"title": map[string]any{"type": "string"}},
			}},
		},
	}
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project compound hook contract: %v", err)
	}
	if stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("projection stats = %+v, want exactly one write and command repair", stats)
	}
	writes := readProjectionFixture(t, writesPath)
	var projected struct {
		Actions []struct {
			BodyFields   []string `json:"body_fields"`
			RecordSchema struct {
				Properties map[string]any `json:"properties"`
			} `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(writes), &projected); err != nil {
		t.Fatalf("decode projected compound hook action: %v", err)
	}
	if len(projected.Actions) != 1 || projected.Actions[0].RecordSchema.Properties["labels"] == nil || !reflect.DeepEqual(projected.Actions[0].BodyFields, []string{"title"}) {
		t.Fatalf("declared hook follow-up field was dropped or leaked into provider body: %s", writes)
	}
	cli := readProjectionFixture(t, cliPath)
	if !strings.Contains(cli, `"maps_to": "record.labels"`) || strings.Contains(cli, `"maps_to": "body"`) {
		t.Fatalf("compound hook command did not retain its closed supplemental field: %s", cli)
	}
}

func TestInstalledReverseActions_CoverProviderRequestContract(t *testing.T) {
	bundle, descriptor := loadInstalledGitHubSourceProjection(t)
	if findings := validateSourceExecutableCoverage(bundle, "sources/github-operation-descriptor.json", descriptor); len(findings) != 0 {
		t.Fatalf("installed GitHub source coverage findings = %+v", findings)
	}
}

func TestInstalledReverseActions_RequiredFieldRemovalFailsBeforeIO(t *testing.T) {
	bundle, descriptor := loadInstalledGitHubSourceProjection(t)
	commands := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Write] = command
	}
	removed := false
	for _, operation := range descriptor.Operations {
		if operation.Protocol == "graphql" || !sourceProjectionMutationMethod(operation.Method) || sourceProjectionHasBlockingGap(operation.Runtime.Gaps) {
			continue
		}
		for index := range bundle.Writes {
			action := bundle.Writes[index]
			if sourceProjectionEndpointKey(action.Method, sourceProjectionPath(action.Path)) != sourceProjectionEndpointKey(operation.Method, operation.Path) || !sourceActionCoversOperation(action, commands[action.Name], operation) {
				continue
			}
			bundle.Writes[index].RecordSchema = json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{}}`)
			for commandIndex := range bundle.CLISurface.Commands {
				if bundle.CLISurface.Commands[commandIndex].Write == action.Name {
					bundle.CLISurface.Commands[commandIndex].Flags = nil
				}
			}
			removed = true
			break
		}
		if removed {
			break
		}
	}
	if !removed {
		t.Fatal("no installed complete mutation action found for removal regression")
	}
	findings := validateSourceExecutableCoverage(bundle, "sources/github-operation-descriptor.json", descriptor)
	if len(findings) == 0 || !strings.Contains(findings[0].Message, "request fields are missing") {
		t.Fatalf("required-field removal did not fail static validation before I/O: %+v", findings)
	}
}

func TestSourceProjectionGapOperationsCannotMasqueradeAsImplemented(t *testing.T) {
	operation := sourceProjectionTestOperation()
	operation.Runtime.Gaps = []sourceContractGap{{
		Foundation: "typed-request-contract-r1",
		Location:   "request.body",
		Reason:     "provider request schema is not yet representable",
	}}
	bundle := engine.Bundle{
		Name: "alpha",
		Writes: []engine.WriteAction{{
			Name:         "items",
			Method:       "POST",
			Path:         "/items/{{ record.owner }}",
			PathFields:   []string{"owner"},
			RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"owner":{"type":"string"}}}`),
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path:         "items create",
			Availability: "implemented",
			Write:        "items",
			Flags:        []engine.CLIFlag{{Name: "owner", Type: "string", MapsTo: "record.owner", Required: true}},
		}}},
	}

	findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "source-bound gap") {
		t.Fatalf("gap-tagged implemented operation findings = %+v, want source-bound gap refusal", findings)
	}
}

func TestSourceProjectionReportsEveryIncompleteGapBeforeIO(t *testing.T) {
	first := sourceProjectionTestOperation()
	first.SourceID = "items/first"
	second := sourceProjectionTestOperation()
	second.SourceID = "items/second"
	for _, operation := range []*sourceOperationDescriptor{&first, &second} {
		operation.Runtime.Gaps = []sourceContractGap{{
			Foundation: "typed-request-contract-r1",
			Location:   "request.body",
			Reason:     "provider request schema is not yet representable",
		}}
	}
	bundle := engine.Bundle{
		Name: "alpha",
		Writes: []engine.WriteAction{{
			Name: "items", Method: "POST", Path: "/items/{{ record.owner }}", PathFields: []string{"owner"},
			RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"owner":{"type":"string"}}}`),
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "items create", Availability: "implemented", Write: "items",
			Flags: []engine.CLIFlag{{Name: "owner", Type: "string", MapsTo: "record.owner", Required: true}},
		}}},
	}
	findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{first, second}})
	if len(findings) != 2 || !strings.Contains(findings[0].Message, "items/first") || !strings.Contains(findings[1].Message, "items/second") {
		t.Fatalf("incomplete source gap findings = %+v, want both identities before I/O", findings)
	}
}

func TestSourceProjectionGapCoverageHonorsDeclaredConfigPathBinding(t *testing.T) {
	spec, err := engine.CompileSchema(json.RawMessage(`{
  "type":"object","additionalProperties":false,
  "required":["owner"],"properties":{"owner":{"type":"string"}}
}`))
	if err != nil {
		t.Fatal(err)
	}
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/read", Method: "get", Path: "/items/{owner}/{item_id}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "owner", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "item_id", Required: true, Schema: map[string]any{"type": "string"}},
		}},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1",
			Location:   "parameter item_id",
			Reason:     "ambiguous request schema uses oneOf",
		}}},
	}
	bundle := engine.Bundle{
		Name: "alpha", Spec: spec,
		Operations: []engine.OperationSpec{{
			ID: "alpha.items.read", Kind: "rest_read",
			REST: &engine.RESTOperationSpec{Method: "GET", Path: "/items/{owner}/{item_id}", MaxBytes: 1024},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "items view", Availability: "implemented", Operation: "alpha.items.read",
			Flags: []engine.CLIFlag{{Name: "item-id", Type: "string", MapsTo: "path.item_id", Required: true}},
		}}},
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}}); len(findings) != 0 {
		t.Fatalf("declared config path binding was mistaken for an incomplete caller input: %+v", findings)
	}
}

func TestSourceProjectionExecutionSurfaceHonorsDeclaredConfigPathReachability(t *testing.T) {
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "spec.json"), `{
  "type": "object",
  "additionalProperties": false,
  "required": ["owner", "repo"],
  "properties": {
    "owner": {"type": "string"},
    "repo": {"type": "string"}
  }
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "operations.json"), `{
  "operations": [{
    "id": "github.actions_permissions_artifact_and_log_retention",
    "kind": "rest_read",
    "output_policy": "json",
    "rest": {
      "method": "GET",
      "path": "/repos/{owner}/{repo}/actions/permissions/artifact-and-log-retention",
      "max_bytes": 1024
    }
  }]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{
  "commands": [{
    "path": "actions artifact-and-log-retention view",
    "intent": "direct_read",
    "availability": "implemented",
    "operation": "github.actions_permissions_artifact_and_log_retention"
  }]
}`)

	surface, err := sourceProjectionExecutionSurface(bundleDir, "github")
	if err != nil {
		t.Fatalf("sourceProjectionExecutionSurface() error = %v", err)
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{{
		Connector: "github", SourceID: "actions/get-artifact-and-log-retention-settings-repository", Method: "GET",
		Path: "/repos/{owner}/{repo}/actions/permissions/artifact-and-log-retention",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "owner", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "repo", Required: true, Schema: map[string]any{"type": "string"}},
		}},
	}}}
	sourceProjectionAnnotateUnreachableReadGaps(surface, &result)
	if sourceProjectionHasBlockingGap(result.Operations[0].Runtime.Gaps) {
		t.Fatalf("config-owned GitHub path fields left source read unreachable: %+v", result.Operations[0].Runtime.Gaps)
	}
}

func TestSourceProjectionRestoresReachableSourceBoundRead(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.list", Method: "GET", Path: "/widgets",
	}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), fmt.Sprintf(`{
  "schema_version": 1,
  "commands": [{
    "path": "widgets list",
    "summary": "list widgets",
    "intent": "direct_read",
    "availability": "partial",
    "api_surface": [{"method":"GET","path":"/widgets"}],
    "notes": %q
  }]
}`, sourceProjectionBlockedReadCommandNote(source.SourceID)))
	writeProjectionFixture(t, filepath.Join(bundleDir, "certification.json"), `{
  "schema_version": 1,
  "direct_read_generation": {
    "cohorts": [{"name":"fixture","command_count":1,"commands":["widgets list"]}]
  }
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), fmt.Sprintf(`{
  "api": "alpha",
  "endpoints": [{
    "method": "GET",
    "path": "/widgets",
    "operation": {
      "model": "direct_read",
      "status": "blocked",
      "risk": "low",
      "blocked_by_default": true,
      "reason": %q,
      "notes": %q
    }
  }]
}`, sourceProjectionBlockedReadSurfaceReason(source.SourceID), sourceProjectionBlockedReadSurfaceNote(source.SourceID)))

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{source}}, false)
	if err != nil {
		t.Fatalf("project reachable source-bound read: %v", err)
	}
	if stats.CLI != 1 || stats.Surface != 1 {
		t.Fatalf("reachable source-bound read stats = %+v, want CLI and surface restoration", stats)
	}
	cli := readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"))
	if !strings.Contains(cli, `"availability": "implemented"`) || strings.Contains(cli, `"notes"`) {
		t.Fatalf("reachable source-bound CLI was not restored:\n%s", cli)
	}
	surface := readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))
	if strings.Contains(surface, `"operation"`) || !strings.Contains(surface, `"direct_read": "widgets list"`) {
		t.Fatalf("reachable source-bound API surface was not restored:\n%s", surface)
	}
}

func TestSourceProjectionRestoresFieldCompleteNonCandidateSourceBoundRead(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.widgets.get", Method: "GET", Path: "/widgets/{widget}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{
			Name: "widget", Required: true, Schema: map[string]any{"type": "string"},
		}}},
	}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), fmt.Sprintf(`{
  "schema_version": 1,
  "commands": [{
    "path": "widgets get",
    "summary": "get widget",
    "intent": "direct_read",
    "availability": "partial",
    "api_surface": [{"method":"GET","path":"/widgets/{widget}"}],
    "flags": [{"name":"widget","type":"string","maps_to":"path.widget","required":true}],
    "notes": %q
  }]
}`, sourceProjectionBlockedReadCommandNote(source.SourceID)))
	// This established route is deliberately absent from certification.json:
	// cohort selection is evidence scope, not execution authority.
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), fmt.Sprintf(`{
  "api": "alpha",
  "endpoints": [{
    "method": "GET",
    "path": "/widgets/{widget}",
    "operation": {
      "model": "direct_read",
      "status": "blocked",
      "risk": "low",
      "blocked_by_default": true,
      "reason": %q,
      "notes": %q
    }
  }]
}`, sourceProjectionBlockedReadSurfaceReason(source.SourceID), sourceProjectionBlockedReadSurfaceNote(source.SourceID)))

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{source}}, false)
	if err != nil {
		t.Fatalf("project field-complete noncandidate source-bound read: %v", err)
	}
	if stats.CLI != 1 || stats.Surface != 1 {
		t.Fatalf("field-complete noncandidate stats = %+v, want CLI and surface restoration", stats)
	}
	cli := readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"))
	if !strings.Contains(cli, `"availability": "implemented"`) || strings.Contains(cli, `"notes"`) {
		t.Fatalf("field-complete noncandidate CLI was not restored:\n%s", cli)
	}
	surface := readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))
	if strings.Contains(surface, `"operation"`) || !strings.Contains(surface, `"direct_read": "widgets get"`) {
		t.Fatalf("field-complete noncandidate API surface was not restored:\n%s", surface)
	}
}

func TestSourceProjectionRestoresRepositorySourceReadWithPluralCoverage(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.repositories.get", Method: "GET", Path: "/repos/{owner}/{repo}/widgets/{widget}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "owner", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "repo", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "widget", Required: true, Schema: map[string]any{"type": "string"}},
		}},
	}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), fmt.Sprintf(`{
  "schema_version": 1,
  "commands": [{
    "path": "widgets get",
    "summary": "get widget",
    "intent": "direct_read",
    "availability": "partial",
    "operation": "alpha.repositories.get",
    "api_surface": [{"method":"GET","path":"/repos/{owner}/{repo}/widgets/{widget}"}],
    "flags": [
      {"name":"owner","type":"string","maps_to":"path.owner","required":true},
      {"name":"repo","type":"string","maps_to":"path.repo","required":true},
      {"name":"widget","type":"string","maps_to":"path.widget","required":true}
    ],
    "notes": %q
  }]
}`, sourceProjectionBlockedReadCommandNote(source.SourceID)))
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), fmt.Sprintf(`{
  "api": "alpha",
  "endpoints": [{
    "method": "GET",
    "path": "/repos/{owner}/{repo}/widgets/{widget}",
    "operation": {
      "model": "direct_read",
      "status": "blocked",
      "risk": "low",
      "blocked_by_default": true,
      "reason": %q,
      "notes": %q
    }
  }]
}`, sourceProjectionBlockedReadSurfaceReason(source.SourceID), sourceProjectionBlockedReadSurfaceNote(source.SourceID)))

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{source}}, false)
	if err != nil {
		t.Fatalf("project repository source-bound read: %v", err)
	}
	if stats.CLI != 1 || stats.Surface != 1 {
		t.Fatalf("repository source-bound stats = %+v, want CLI and surface restoration", stats)
	}
	surface := readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))
	if strings.Contains(surface, `"operation"`) || !strings.Contains(surface, `"direct_reads": [`) || !strings.Contains(surface, `"widgets get"`) {
		t.Fatalf("repository source-bound API surface did not retain plural coverage:\n%s", surface)
	}
}

func TestSourceProjectionKeepsSourceOnlyRepositoryReadCoverageSingular(t *testing.T) {
	source := sourceOperationDescriptor{Path: "/repos/{owner}/{repo}/widgets/{widget}"}
	command := newOrderedObject()
	command.set("path", "widgets get")
	endpoint := newOrderedObject()
	endpoint.set("method", "GET")
	endpoint.set("path", source.Path)
	command.set("api_surface", []any{endpoint})
	cli := newOrderedObject()
	cli.set("commands", []any{command})

	coverage := sourceProjectionReadSurfaceCoverage(source, cli, sourceProjectionEndpointKey("GET", source.Path), []string{"widgets get"})
	if got := stringField(coverage, "direct_read"); got != "widgets get" {
		t.Fatalf("source-only repository read coverage = %#v, want singular direct_read", coverage)
	}
	if _, plural := coverage.get("direct_reads"); plural {
		t.Fatalf("source-only repository read coverage = %#v, want no plural direct_reads", coverage)
	}
}

func TestSourceProjectionRetainsFieldCompleteBinaryDownloadSourceRead(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "alpha.exports.download", Method: "GET", Path: "/exports/{export_id}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{
			Name: "export_id", Required: true, Schema: map[string]any{"type": "string"},
		}}},
	}
	bundle := engine.Bundle{
		Name: "alpha",
		Operations: []engine.OperationSpec{{
			ID: "alpha.exports.download", Kind: "binary_download",
			Binary: &engine.BinaryOperationSpec{
				Method: "GET", Path: "/exports/{export_id}", MaxBytes: 1024,
				Parameters: []engine.OperationParameter{{Name: "export_id", In: "path", Required: true}},
			},
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "exports download", Intent: "binary_download", Availability: "implemented", Operation: "alpha.exports.download",
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/exports/{export_id}"}},
			Flags:      []engine.CLIFlag{{Name: "export-id", MapsTo: "path.export_id", Required: true}},
		}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{source}}
	sourceProjectionAnnotateUnreachableReadGaps(bundle, &result)
	if sourceProjectionHasBlockingGap(result.Operations[0].Runtime.Gaps) {
		t.Fatalf("field-complete binary download was marked unreachable: %+v", result.Operations[0].Runtime.Gaps)
	}
}

func TestSourceProjectionGapCoverageHonorsDeclaredActionConfigPathBinding(t *testing.T) {
	spec, err := engine.CompileSchema(json.RawMessage(`{
  "type":"object","additionalProperties":false,
  "required":["owner"],"properties":{"owner":{"type":"string"}}
}`))
	if err != nil {
		t.Fatal(err)
	}
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/update", Method: "patch", Path: "/items/{owner}/{item_id}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "owner", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "item_id", Required: true, Schema: map[string]any{"type": "string"}},
		}},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1",
			Location:   "parameter item_id",
			Reason:     "ambiguous request schema uses oneOf",
		}}},
	}
	bundle := engine.Bundle{
		Name: "alpha", Spec: spec,
		Writes: []engine.WriteAction{{
			Name: "items_update", Method: "PATCH", Path: "/items/{{ config.owner }}/{{ record.item_id }}",
			PathFields:   []string{"item_id"},
			RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"required":["item_id"],"properties":{"item_id":{"type":"string"}}}`),
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "items update", Availability: "implemented", Write: "items_update",
			Flags: []engine.CLIFlag{{Name: "item-id", Type: "string", MapsTo: "record.item_id", Required: true}},
		}}},
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/alpha-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}}); len(findings) != 0 {
		t.Fatalf("declared action config path binding was mistaken for an incomplete caller input: %+v", findings)
	}
}

func TestGoogleAdsGeneratedPOSTReadsAcceptDeclaredNestedObjects(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "internal", "connectors", "defs", "google-ads", "operations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Operations []struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
			REST struct {
				Method     string         `json:"method"`
				BodySchema map[string]any `json:"body_schema"`
			} `json:"rest"`
		} `json:"operations"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	var empty []string
	for _, operation := range document.Operations {
		if operation.Kind != "rest_read" || operation.REST.Method != "POST" || operation.REST.BodySchema == nil {
			continue
		}
		walkClosedObjectSchemas(operation.REST.BodySchema, operation.ID+".body", &empty)
	}
	if len(empty) > 0 {
		t.Fatalf("implemented Google Ads POST reads contain reachable closed-empty object schemas: %s", strings.Join(empty, ", "))
	}

	cliRaw, err := os.ReadFile(filepath.Join(root, "internal", "connectors", "defs", "google-ads", "cli_surface.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cli engine.CLISurface
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatal(err)
	}
	var keywordIdeas *engine.CLICommand
	for i := range cli.Commands {
		if cli.Commands[i].Path == "customers generate-keyword-ideas" {
			keywordIdeas = &cli.Commands[i]
			break
		}
	}
	if keywordIdeas == nil {
		t.Fatal("Google Ads keyword-ideas command is not implemented")
	}
	seedTargets := []string{"body.keywordAndUrlSeed", "body.urlSeed", "body.keywordSeed", "body.siteSeed"}
	seedFlags := map[string]string{}
	for _, flag := range keywordIdeas.Flags {
		seedFlags[flag.MapsTo] = flag.Type
	}
	for _, target := range seedTargets {
		if seedFlags[target] != "json" {
			t.Errorf("keyword-ideas target %s flag type = %q, want json", target, seedFlags[target])
		}
	}
	if len(keywordIdeas.Constraints) != 1 || keywordIdeas.Constraints[0].Kind != "exactly_one" || !reflect.DeepEqual(keywordIdeas.Constraints[0].Fields, seedTargets) {
		t.Errorf("keyword-ideas constraints = %#v, want exact seed one-of %v", keywordIdeas.Constraints, seedTargets)
	}
}

func walkClosedObjectSchemas(value any, path string, empty *[]string) {
	object, ok := value.(map[string]any)
	if !ok {
		if values, ok := value.([]any); ok {
			for index, item := range values {
				walkClosedObjectSchemas(item, fmt.Sprintf("%s[%d]", path, index), empty)
			}
		}
		return
	}
	if object["type"] == "object" && object["additionalProperties"] == false {
		properties, _ := object["properties"].(map[string]any)
		if len(properties) == 0 {
			*empty = append(*empty, path)
		}
	}
	for key, child := range object {
		walkClosedObjectSchemas(child, path+"."+key, empty)
	}
}

func sourceProjectionTestOperation() sourceOperationDescriptor {
	return sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/create", Method: "post", Path: "/items/{owner}",
		Request: sourceRequestDescriptor{
			Path:  []sourceParameterDescriptor{{Name: "owner", Required: true, Schema: map[string]any{"type": "string"}}},
			Query: []sourceParameterDescriptor{{Name: "mode", Required: true, Schema: map[string]any{"type": "string", "enum": []any{"fast", "safe"}}}},
			Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
				"type": "object", "additionalProperties": false, "required": []any{"name"},
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"meta": map[string]any{"type": []any{"object", "null"}, "additionalProperties": false, "properties": map[string]any{"note": map[string]any{"type": "string"}}},
					"tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			}},
			MediaType: "application/json",
		},
	}
}

func loadInstalledGitHubSourceProjection(t *testing.T) (engine.Bundle, sourceImportDescriptorDocument) {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	defs := filepath.Join(root, "internal", "connectors", "defs")
	bundle, err := engine.Load(os.DirFS(defs), "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(defs, "github", "sources", "github-operation-descriptor.json"))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(raw, &descriptor); err != nil {
		t.Fatalf("decode GitHub source descriptor: %v", err)
	}
	return bundle, descriptor
}

func writeProjectionFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSourceProjectionAcceptsVersion3DocumentProvenance(t *testing.T) {
	artifact := []byte(`{"openapi":"3.0.3","info":{"title":"alpha","version":"1"},"paths":{"/alpha":{"get":{"operationId":"shared","responses":{"200":{"description":"ok"}}}}}}`)
	lock, err := parseSourceImportLock(sourceImportV3FixtureLock(t, "fixture", []sourceImportV3FixtureDocument{{ID: "alpha", Path: "/alpha", Artifact: artifact}}), "fixture")
	if err != nil {
		t.Fatalf("parse v3 fixture lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(context.Context, string) ([]byte, error) { return artifact, nil }), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v3 fixture lock: %v", err)
	}
	if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("v3 descriptor provenance findings = %+v", findings)
	}
	for _, change := range []struct {
		name   string
		mutate func(*sourceOperationDescriptor)
	}{
		{name: "provider operation id", mutate: func(operation *sourceOperationDescriptor) { operation.ProviderOperationID = "rewritten" }},
		{name: "artifact url", mutate: func(operation *sourceOperationDescriptor) {
			operation.Source.URL = "https://fixtures.polymetrics.invalid/reassigned.openapi.json"
		}},
		{name: "document id", mutate: func(operation *sourceOperationDescriptor) { operation.Source.DocumentID = "reassigned" }},
		{name: "published url", mutate: func(operation *sourceOperationDescriptor) {
			operation.Source.PublishedURL = "https://published.polymetrics.invalid/reassigned?slug=reassigned"
		}},
		{name: "published capture url", mutate: func(operation *sourceOperationDescriptor) {
			operation.Source.PublishedCaptureURL = "https://fixtures.polymetrics.invalid/reassigned.capture.json"
		}},
		{name: "published digest", mutate: func(operation *sourceOperationDescriptor) { operation.Source.PublishedSHA256 = strings.Repeat("0", 64) }},
		{name: "published bytes", mutate: func(operation *sourceOperationDescriptor) { operation.Source.PublishedBytes++ }},
		{name: "published adapter", mutate: func(operation *sourceOperationDescriptor) { operation.Source.PublishedAdapter = "other-adapter" }},
		{name: "source form", mutate: func(operation *sourceOperationDescriptor) { operation.Source.Form = "swagger" }},
		{name: "source version", mutate: func(operation *sourceOperationDescriptor) { operation.Source.Version = "2.0" }},
	} {
		change := change
		t.Run(change.name, func(t *testing.T) {
			operation := result.Operations[0]
			change.mutate(&operation)
			if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{operation}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "provenance drift") {
				t.Fatalf("%s findings = %+v", change.name, findings)
			}
		})
	}
}

func TestSurfaceSyncAcceptsSchema3SourceDescriptor(t *testing.T) {
	bundleDir := filepath.Join(t.TempDir(), "alpha")
	if err := os.MkdirAll(filepath.Join(bundleDir, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	operation := sourceProjectionTestOperation()
	descriptorRaw, err := json.Marshal(sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{operation}})
	if err != nil {
		t.Fatal(err)
	}
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "alpha-operation-source-lock.json"), `{}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "alpha-operation-descriptor.json"), string(descriptorRaw))
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[{"name":"items","kind":"custom","method":"POST","path":"/items/{{ record.owner }}","path_fields":["owner"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"owner":{"type":"string"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[{"path":"items create","summary":"create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"owner","type":"string","maps_to":"record.owner","required":true}]}]}`)
	stats, err := syncCheckedInSourceProjection(bundleDir, "alpha", true)
	if err != nil {
		t.Fatalf("schema-3 source projection check: %v", err)
	}
	if !stats.Changed() || stats.Writes != 1 || stats.CLI != 1 {
		t.Fatalf("schema-3 source projection drift = %+v", stats)
	}
}

func readProjectionFixture(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestSourceProjectionSourceCitedNonExecutableMutationDispositionsCoverAbsentAndIncompleteActions(t *testing.T) {
	tests := []struct {
		name       string
		connector  string
		operation  sourceOperationDescriptor
		writes     string
		cli        string
		bundle     engine.Bundle
		wantBase   string
		wantReason string
	}{
		{
			name:       "asana absent action",
			connector:  "asana",
			operation:  sourceCitedMutationTestOperation("asana", "asana.create_access_request", "POST", "/access_requests"),
			writes:     `{"schema_version":1,"actions":[]}`,
			cli:        `{"schema_version":1,"commands":[]}`,
			bundle:     engine.Bundle{Name: "asana", CLISurface: &engine.CLISurface{}},
			wantBase:   "no executable action",
			wantReason: "provider exposes the access-request mutation, but the connector has no declaration-owned action",
		},
		{
			name:      "jira incomplete contract",
			connector: "jira",
			operation: sourceCitedMutationTestOperation("jira", "jira.bulk_submit_bulk_edit", "POST", "/rest/api/3/bulk/issues/fields"),
			writes:    `{"schema_version":1,"actions":[{"name":"bulk_submit_bulk_edit","kind":"create","method":"POST","path":"/rest/api/3/bulk/issues/fields","body_type":"json","record_schema":{"type":"object","additionalProperties":false,"properties":{"selectedActions":{"type":"array","items":{"type":"string"}},"selectedIssueIdsOrKeys":{"type":"array","items":{"type":"string"}}}},"risk":"standard"}]}`,
			cli:       `{"schema_version":1,"commands":[{"path":"bulk submit-bulk-edit","summary":"bulk edit issues","intent":"reverse_etl","availability":"partial","write":"bulk_submit_bulk_edit","flags":[{"name":"selected-actions","type":"string_array","maps_to":"record.selectedActions","required":true},{"name":"selected-issue-ids-or-keys","type":"string_array","maps_to":"record.selectedIssueIdsOrKeys","required":true}]}]}`,
			bundle: engine.Bundle{Name: "jira", Writes: []engine.WriteAction{{
				Name: "bulk_submit_bulk_edit", Kind: "create", Method: "POST", Path: "/rest/api/3/bulk/issues/fields", BodyType: "json",
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"selectedActions":{"type":"array","items":{"type":"string"}},"selectedIssueIdsOrKeys":{"type":"array","items":{"type":"string"}}}}`),
			}}, CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
				Path: "bulk submit-bulk-edit", Intent: "reverse_etl", Availability: "partial", Write: "bulk_submit_bulk_edit",
				Flags: []engine.CLIFlag{{Name: "selected-actions", Type: "string_array", MapsTo: "record.selectedActions", Required: true}, {Name: "selected-issue-ids-or-keys", Type: "string_array", MapsTo: "record.selectedIssueIdsOrKeys", Required: true}},
			}}}},
			wantBase:   "source request fields are missing",
			wantReason: "provider request body uses an unbounded object that has no complete declared action contract",
		},
		{
			name:       "sentry scim patch absent action",
			connector:  "sentry",
			operation:  sourceCitedMutationTestOperation("sentry", "sentry.rest.updateOrganizationScimV2Group", "PATCH", "/api/0/organizations/{organization_id_or_slug}/scim/v2/Groups/{team_id_or_slug}"),
			writes:     `{"schema_version":1,"actions":[]}`,
			cli:        `{"schema_version":1,"commands":[]}`,
			bundle:     engine.Bundle{Name: "sentry", CLISurface: &engine.CLISurface{}},
			wantBase:   "no executable action",
			wantReason: "Sentry SCIM group update is provider-cited but has no declaration-owned action",
		},
		{
			name:       "sentry dashboard post absent action",
			connector:  "sentry",
			operation:  sourceCitedMutationTestOperation("sentry", "sentry.rest.createOrganizationDashboard", "POST", "/api/0/organizations/{organization_id_or_slug}/dashboards/"),
			writes:     `{"schema_version":1,"actions":[]}`,
			cli:        `{"schema_version":1,"commands":[]}`,
			bundle:     engine.Bundle{Name: "sentry", CLISurface: &engine.CLISurface{}},
			wantBase:   "no executable action",
			wantReason: "Sentry dashboard creation is provider-cited but has no declaration-owned action",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			operation := tc.operation
			if tc.connector == "asana" {
				operation.Source.URL = "https://developers.asana.com/reference/createaccessrequest"
				operation.Request.Body = &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
					"type": "object", "properties": map[string]any{"data": map[string]any{"type": "object", "additionalProperties": true}},
				}}
			} else if tc.connector == "jira" {
				operation.Source.URL = "https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/"
				operation.Request.Body = &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
					"type": "object", "properties": map[string]any{"editedFieldsInput": map[string]any{"type": "object"}, "selectedActions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "selectedIssueIdsOrKeys": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}},
				}}
			} else if operation.Method == "PATCH" {
				operation.Source.URL = "https://sentry.io/api/0/"
				operation.Request.Path = []sourceParameterDescriptor{{Name: "organization_id_or_slug", Required: true, Schema: map[string]any{"type": "string"}}, {Name: "team_id_or_slug", Required: true, Schema: map[string]any{"type": "string"}}}
				operation.Request.Body = &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
					"type": "object", "properties": map[string]any{"displayName": map[string]any{"type": "string"}},
				}}
			} else {
				operation.Source.URL = "https://sentry.io/api/0/"
				operation.Request.Path = []sourceParameterDescriptor{{Name: "organization_id_or_slug", Required: true, Schema: map[string]any{"type": "string"}}}
				operation.Request.Body = &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
					"type": "object", "properties": map[string]any{"title": map[string]any{"type": "string"}},
				}}
			}
			if findings := validateSourceExecutableCoverage(tc.bundle, "sources/"+tc.connector+"-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}}); len(findings) != 1 || !strings.Contains(findings[0].Message, tc.wantBase) {
				t.Fatalf("undisposed source finding = %+v, want %q", findings, tc.wantBase)
			}
			disposition := sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{
				SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path,
			}, Reason: tc.wantReason}
			bundleDir := filepath.Join(t.TempDir(), tc.connector)
			if err := os.MkdirAll(filepath.Join(bundleDir, "sources"), 0o755); err != nil {
				t.Fatal(err)
			}
			dispositionRaw, err := json.Marshal(sourceNonExecutableMutationDispositionDocument{
				SchemaVersion: 1,
				Dispositions:  []sourceNonExecutableMutationDisposition{disposition},
			})
			if err != nil {
				t.Fatal(err)
			}
			writeProjectionFixture(t, filepath.Join(bundleDir, "sources", tc.connector+"-mutation-dispositions.json"), string(dispositionRaw))
			dispositions, err := sourceProjectionReadNonExecutableMutationDispositions(bundleDir)
			if err != nil {
				t.Fatalf("read source-cited dispositions: %v", err)
			}
			result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
			if err := sourceProjectionApplyNonExecutableMutationDispositions(tc.bundle, &result, dispositions); err != nil {
				t.Fatalf("apply source-cited disposition: %v", err)
			}
			descriptorRaw, err := marshalSourceImportResult(result)
			if err != nil {
				t.Fatalf("marshal source descriptor: %v", err)
			}
			var descriptor sourceImportDescriptorDocument
			if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
				t.Fatalf("decode source descriptor: %v", err)
			}
			got := descriptor.Operations[0]
			if !got.Runtime.MergeBlocked || got.Runtime.NonExecutableMutation == nil || got.Runtime.NonExecutableMutation.Source != disposition.Source {
				t.Fatalf("runtime disposition = %#v, want source-cited non-executable mutation", got.Runtime)
			}
			if len(got.Runtime.Gaps) != 1 || got.Runtime.Gaps[0].Foundation != sourceNonExecutableMutationDispositionFoundation || !strings.Contains(got.Runtime.Gaps[0].Location, got.Source.URL) || !strings.Contains(got.Runtime.Gaps[0].Reason, tc.wantReason) {
				t.Fatalf("runtime gap = %#v, want provider-cited mutation gap", got.Runtime.Gaps)
			}

			writesPath := filepath.Join(bundleDir, "writes.json")
			cliPath := filepath.Join(bundleDir, "cli_surface.json")
			writeProjectionFixture(t, writesPath, tc.writes)
			writeProjectionFixture(t, cliPath, tc.cli)
			stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: descriptor.Operations}, false)
			if err != nil {
				t.Fatalf("source projection: %v", err)
			}
			if stats.Missing != 0 {
				t.Fatalf("source projection reported disposed mutation as missing: %+v", stats)
			}
			if gotWrites, gotCLI := readProjectionFixture(t, writesPath), readProjectionFixture(t, cliPath); gotWrites != tc.writes || gotCLI != tc.cli {
				t.Fatalf("source projection fabricated an action or command:\nwrites=%s\ncli=%s", gotWrites, gotCLI)
			}
			if findings := validateSourceExecutableCoverage(tc.bundle, "sources/"+tc.connector+"-operation-descriptor.json", descriptor); len(findings) != 0 {
				t.Fatalf("executable coverage findings = %+v", findings)
			}
		})
	}
}

func TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsCompleteAction(t *testing.T) {
	operation := sourceProjectionTestOperation()
	operation.Connector = "sentry"
	operation.SourceID = "sentry.items.create"
	operation.Source = sourceCitedMutationTestOperation("sentry", operation.SourceID, operation.Method, operation.Path).Source
	bundleDir := filepath.Join(t.TempDir(), "sentry")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"items","kind":"custom","method":"POST","path":"/items/{{ record.owner }}","path_fields":["owner"],"body_type":"json","body_fields":["stale"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"stale":{"type":"string"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items create","summary":"create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"stale","type":"string","maps_to":"record.stale"}]}]}`)
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("materialize complete source action: %v", err)
	}
	if stats.Missing != 0 {
		t.Fatalf("complete action source projection missing = %+v", stats)
	}
	var writes struct {
		Actions []engine.WriteAction `json:"actions"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, writesPath)), &writes); err != nil {
		t.Fatal(err)
	}
	var cli engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, cliPath)), &cli); err != nil {
		t.Fatal(err)
	}
	bundle := engine.Bundle{Name: "sentry", Writes: writes.Actions, CLISurface: &cli}
	if findings := validateSourceExecutableCoverage(bundle, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}}); len(findings) != 0 {
		t.Fatalf("complete action became unreachable without a disposition: %+v", findings)
	}
	disposition := sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path}, Reason: "must not hide a complete action"}
	if err := sourceProjectionApplyNonExecutableMutationDispositions(bundle, &sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, []sourceNonExecutableMutationDisposition{disposition}); err == nil || !strings.Contains(err.Error(), "complete executable action") {
		t.Fatalf("complete action disposition error = %v, want refusal", err)
	}
	operation.Runtime = sourceRuntimeReachability{
		MergeBlocked:          true,
		NonExecutableMutation: &disposition,
		Gaps:                  []sourceContractGap{sourceProjectionNonExecutableMutationRuntimeGap(operation, disposition)},
	}
	if _, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, true); err == nil || !strings.Contains(err.Error(), "claims a complete executable action") {
		t.Fatalf("source projection complete-action disposition error = %v, want refusal", err)
	}
	findings := validateSourceExecutableCoverage(bundle, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "claims a complete executable action") {
		t.Fatalf("executable coverage complete-action findings = %+v, want refusal", findings)
	}
	if bundle.CLISurface.Commands[0].Availability != "implemented" {
		t.Fatalf("complete action command availability = %q, want implemented", bundle.CLISurface.Commands[0].Availability)
	}
}

func TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsImplementedIncompleteActionClaim(t *testing.T) {
	operation := sourceCitedMutationTestOperation("sentry", "sentry.items.update", "PATCH", "/items/{item_id}")
	operation.Request.Path = []sourceParameterDescriptor{{Name: "item_id", Required: true, Schema: map[string]any{"type": "string"}}}
	operation.Request.Body = &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
		"type": "object", "properties": map[string]any{"payload": map[string]any{"type": "object"}},
	}}
	disposition := sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path}, Reason: "an incomplete action must not hide an implemented command claim"}
	bundle := engine.Bundle{Name: "sentry", Writes: []engine.WriteAction{{
		Name: "update-item", Kind: "update", Method: "PATCH", Path: "/items/{{ record.item_id }}", PathFields: []string{"item_id"}, BodyType: "json",
		RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"item_id":{"type":"string"}}}`),
	}}, CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
		Path: "items update", Intent: "reverse_etl", Availability: "implemented", Write: "update-item",
		Flags: []engine.CLIFlag{{Name: "item-id", Type: "string", MapsTo: "record.item_id", Required: true}},
	}}}}
	if err := sourceProjectionApplyNonExecutableMutationDispositions(bundle, &sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, []sourceNonExecutableMutationDisposition{disposition}); err == nil || !strings.Contains(err.Error(), "implemented executable action") {
		t.Fatalf("implemented incomplete action disposition error = %v, want refusal", err)
	}
	operation.Runtime = sourceRuntimeReachability{
		MergeBlocked:          true,
		NonExecutableMutation: &disposition,
		Gaps:                  []sourceContractGap{sourceProjectionNonExecutableMutationRuntimeGap(operation, disposition)},
	}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[{"name":"update-item","kind":"update","method":"PATCH","path":"/items/{{ record.item_id }}","path_fields":["item_id"],"body_type":"json","record_schema":{"type":"object","additionalProperties":false,"properties":{"item_id":{"type":"string"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[{"path":"items update","summary":"update","intent":"reverse_etl","availability":"implemented","write":"update-item","flags":[{"name":"item-id","type":"string","maps_to":"record.item_id","required":true}]}]}`)
	if _, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, true); err == nil || !strings.Contains(err.Error(), "claims an implemented executable action") {
		t.Fatalf("source projection implemented incomplete action error = %v, want refusal", err)
	}
	findings := validateSourceExecutableCoverage(bundle, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "claims an implemented executable action") {
		t.Fatalf("executable coverage implemented incomplete action findings = %+v, want refusal", findings)
	}
	if bundle.CLISurface.Commands[0].Availability != "implemented" {
		t.Fatalf("implemented command availability = %q, want preserved implementation claim", bundle.CLISurface.Commands[0].Availability)
	}
}

func TestSourceProjectionSourceCitedNonExecutableMutationDispositionScalesAcrossVercelMutationShapes(t *testing.T) {
	seeds := []sourceOperationDescriptor{
		sourceCitedMutationTestOperation("vercel", "vercel.rest.editRedirect", "PATCH", "/v1/bulk-redirects"),
		sourceCitedMutationTestOperation("vercel", "vercel.rest.restoreRedirects", "POST", "/v1/bulk-redirects/restore"),
		sourceCitedMutationTestOperation("vercel", "vercel.rest.writeSessionFiles", "POST", "/v2/sandboxes/sessions/{sessionId}/fs/write"),
		sourceCitedMutationTestOperation("vercel", "vercel.rest.dangerouslyDeleteByTags", "POST", "/v1/edge-cache/dangerously-delete-by-tags"),
	}
	operations := make([]sourceOperationDescriptor, 0, 159)
	for index := 0; index < cap(operations); index++ {
		operation := seeds[index%len(seeds)]
		operation.SourceID += ".handoff." + fmt.Sprintf("%03d", index)
		operation.Source.URL = "https://openapi.vercel.sh/"
		operations = append(operations, operation)
	}
	dispositions := make([]sourceNonExecutableMutationDisposition, 0, len(operations))
	for _, operation := range operations {
		dispositions = append(dispositions, sourceNonExecutableMutationDisposition{
			Source: sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path},
			Reason: "Vercel provider mutation is cited but has no declaration-owned executable action",
		})
	}
	bundleDir := filepath.Join(t.TempDir(), "vercel")
	if err := os.MkdirAll(filepath.Join(bundleDir, "sources"), 0o755); err != nil {
		t.Fatal(err)
	}
	dispositionRaw, err := json.Marshal(sourceNonExecutableMutationDispositionDocument{SchemaVersion: 1, Dispositions: dispositions})
	if err != nil {
		t.Fatal(err)
	}
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "vercel-mutation-dispositions.json"), string(dispositionRaw))
	dispositions, err = sourceProjectionReadNonExecutableMutationDispositions(bundleDir)
	if err != nil {
		t.Fatalf("read Vercel source-cited dispositions: %v", err)
	}
	result := sourceImportResult{Operations: operations}
	bundle := engine.Bundle{Name: "vercel", CLISurface: &engine.CLISurface{}}
	if err := sourceProjectionApplyNonExecutableMutationDispositions(bundle, &result, dispositions); err != nil {
		t.Fatalf("apply Vercel source-cited dispositions: %v", err)
	}
	for _, operation := range result.Operations {
		if !sourceProjectionHasNonExecutableMutationDisposition(operation) {
			t.Fatalf("Vercel source operation %q did not retain its cited non-executable mutation gap", operation.SourceID)
		}
	}
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	const emptyWrites = "{\"schema_version\":1,\"actions\":[]}"
	const emptyCLI = "{\"schema_version\":1,\"commands\":[]}"
	writeProjectionFixture(t, writesPath, emptyWrites)
	writeProjectionFixture(t, cliPath, emptyCLI)
	stats, err := projectSourceDescriptorToBundle(bundleDir, result, false)
	if err != nil || stats.Missing != 0 {
		t.Fatalf("Vercel source projection = stats:%+v err:%v, want every cited mutation retained without a generated action", stats, err)
	}
	if gotWrites, gotCLI := readProjectionFixture(t, writesPath), readProjectionFixture(t, cliPath); gotWrites != emptyWrites || gotCLI != emptyCLI {
		t.Fatalf("Vercel source projection fabricated an action or command:\nwrites=%s\ncli=%s", gotWrites, gotCLI)
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/vercel-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("Vercel executable coverage findings = %+v", findings)
	}
}

func TestSourceProjectionReadOnlyFoundationCannotSatisfyMutationCoverage(t *testing.T) {
	operation := sourceCitedMutationTestOperation("sentry", "sentry.issues.delete", "DELETE", "/issues/{issue_id}")
	operation.Runtime = sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
		Foundation: sourceReadOnlyOperationFoundation,
		Location:   "source operation sentry.issues.delete",
		Reason:     "provider source was incorrectly classified as read-only",
	}}}
	bundle := engine.Bundle{Name: "sentry", CLISurface: &engine.CLISurface{}}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[]}`)
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, true)
	if err == nil || stats.Missing != 1 {
		t.Fatalf("source projection read-only mutation result = stats:%+v err:%v, want visible missing mutation", stats, err)
	}
	findings := validateSourceExecutableCoverage(bundle, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "read-only disposition cannot cover a mutating source operation") {
		t.Fatalf("read-only mutation findings = %+v, want mutation-only refusal", findings)
	}
}

func TestSourceProjectionSourceCitedMutationDispositionRejectsPOSTGraphQLQuery(t *testing.T) {
	operation := sourceCitedMutationTestOperation("sentry", "sentry.graphql.query.issue", "POST", "/graphql")
	operation.Protocol = "graphql"
	operation.GraphQL = &sourceGraphQLOperationDescriptor{Root: "query", Name: "issue"}
	disposition := sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path}, Reason: "a query cannot be treated as a mutation gap"}
	err := sourceProjectionApplyNonExecutableMutationDispositions(engine.Bundle{Name: "sentry", CLISurface: &engine.CLISurface{}}, &sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, []sourceNonExecutableMutationDisposition{disposition})
	if err == nil || !strings.Contains(err.Error(), "not mutating") {
		t.Fatalf("GraphQL query mutation-disposition error = %v, want non-mutating refusal", err)
	}
}

func TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	_, descriptor := loadInstalledGitHubSourceProjection(t)
	bundleDir := filepath.Join(root, "internal", "connectors", "defs", "github")
	paths := []string{filepath.Join(bundleDir, "writes.json"), filepath.Join(bundleDir, "cli_surface.json"), filepath.Join(bundleDir, "api_surface.json")}
	before := make([]string, len(paths))
	for index, path := range paths {
		before[index] = readProjectionFixture(t, path)
	}
	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: descriptor.Operations}, true)
	if err != nil {
		t.Fatalf("existing GitHub source projection: %v", err)
	}
	if stats.Changed() {
		t.Fatalf("existing GitHub source projection changed without a mutation disposition: %+v", stats)
	}
	for index, path := range paths {
		if after := readProjectionFixture(t, path); after != before[index] {
			t.Fatalf("existing GitHub projection changed %s", filepath.Base(path))
		}
	}
}

func sourceCitedMutationTestOperation(connector, sourceID, method, path string) sourceOperationDescriptor {
	return sourceOperationDescriptor{
		Connector: connector, SourceID: sourceID, Method: method, Path: path,
		Source: sourceImportSource{
			URL: "https://provider.example.test/openapi.json", SHA256: strings.Repeat("a", 64), Bytes: 1024,
			Location: "#/paths/~1mutation/" + strings.ToLower(method),
		},
	}
}
