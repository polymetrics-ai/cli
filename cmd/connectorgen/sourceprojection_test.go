package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceProjectionSingleDataEnvelopeUsesRequestBodyRequiredness(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name         string
		bodyRequired bool
		wantRequired bool
	}{
		{name: "required provider body", bodyRequired: true, wantRequired: true},
		{name: "optional provider body", bodyRequired: false, wantRequired: false},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var action orderedJSON
			if err := json.Unmarshal([]byte(`{
  "name":"create_item","method":"POST","path":"/items","body_type":"json",
  "record_schema":{"type":"object","additionalProperties":false,"properties":{}},
  "risk":"standard"
}`), &action); err != nil {
				t.Fatal(err)
			}
			operation := sourceOperationDescriptor{
				Method: "POST", Path: "/items",
				Request: sourceRequestDescriptor{
					MediaType: "application/json",
					Body: &sourceRequestBodyDescriptor{
						Required: testCase.bodyRequired,
						Schema: map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"properties": map[string]any{
								"data": map[string]any{
									"type":                 "object",
									"additionalProperties": false,
									"properties": map[string]any{
										"name": map[string]any{"type": "string", "minLength": json.Number("1"), "maxLength": json.Number("64")},
									},
								},
							},
						},
					},
				},
			}
			contract, err := sourceContractForAction(operation, action.root)
			if err != nil {
				t.Fatalf("project single data envelope: %v", err)
			}
			if got := contract.Required["data"]; got != testCase.wantRequired {
				t.Fatalf("projected data required = %t, want provider requestBody.required=%t", got, testCase.wantRequired)
			}
		})
	}

	t.Run("non GitLab PATCH mutation remains a mutation", func(t *testing.T) {
		mutation := sourceCitedMutationTestOperation("stripe", "stripe.items.update", http.MethodPatch, "/items/{item_id}")
		bundle := engine.Bundle{Name: "stripe", CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "items update", Intent: "direct_write", Availability: "implemented",
			APISurface: []engine.CLISurfaceEndpointRef{{Method: http.MethodPatch, Path: "/items/{item_id}"}},
		}}}}
		if !sourceProjectionMutationClaimsImplementedAction(bundle, mutation) {
			t.Fatal("non-GitLab PATCH mutation was incorrectly admitted as a semantic direct read")
		}
	})
}

func TestSourceProjectionQueryArrayUsesSourceFormEncoding(t *testing.T) {
	t.Parallel()
	explode := false
	var action orderedJSON
	if err := json.Unmarshal([]byte(`{
  "name":"create_item","method":"POST","path":"/items","body_type":"none",
  "record_schema":{"type":"object","additionalProperties":false,"properties":{}},
  "risk":"standard"
}`), &action); err != nil {
		t.Fatal(err)
	}
	operation := sourceOperationDescriptor{
		Method: "POST", Path: "/items",
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{{
			Name: "fields", Required: false,
			Schema: map[string]any{
				"type": "array", "minItems": json.Number("1"), "maxItems": json.Number("8"),
				"items": map[string]any{"type": "string", "minLength": json.Number("1"), "maxLength": json.Number("64")},
			},
			Wire: sourceParameterWireDescriptor{Style: "form", Explode: &explode},
		}}},
	}
	contract, err := sourceContractForAction(operation, action.root)
	if err != nil {
		t.Fatalf("project query array contract: %v", err)
	}
	sourceProjectAction(action.root, contract)
	rawQuery, _ := action.root.get("query")
	query, _ := rawQuery.(*orderedObject)
	rawFields, _ := query.get("fields")
	fields, _ := rawFields.(*orderedObject)
	if got := stringField(fields, "template"); got != "{{ record.fields | join:, }}" {
		t.Fatalf("form/explode=false query template = %q, want comma join", got)
	}
}

func TestSourceProjectCommandPreservesDeclaredFlagOrderAndNames(t *testing.T) {
	t.Parallel()
	var command orderedJSON
	if err := json.Unmarshal([]byte(`{
  "path":"items update","flags":[
    {"name":"second-json","summary":"second","type":"json","maps_to":"record.second"},
    {"name":"first-value","summary":"first","type":"string","maps_to":"record.first"}
  ]
}`), &command); err != nil {
		t.Fatal(err)
	}
	contract := sourceActionContract{
		Fields: map[string]any{
			"first":  map[string]any{"type": "string", "maxLength": json.Number("32")},
			"second": map[string]any{"type": "object", "additionalProperties": true, "maxProperties": json.Number("8")},
			"third":  map[string]any{"type": "boolean"},
		},
		BareStringFields: map[string]bool{}, SecretFields: map[string]bool{},
		Required: map[string]bool{"first": true, "second": false, "third": false},
	}
	sourceProjectCommand(command.root, contract)
	flags := arrayField(command.root, "flags")
	if len(flags) != 3 {
		t.Fatalf("flag count = %d, want 3", len(flags))
	}
	wantNames := []string{"second-json", "first-value", "third"}
	wantTargets := []string{"record.second", "record.first", "record.third"}
	for index := range wantNames {
		flag, _ := flags[index].(*orderedObject)
		if got := stringField(flag, "name"); got != wantNames[index] {
			t.Errorf("flag %d name = %q, want %q", index, got, wantNames[index])
		}
		if got := stringField(flag, "maps_to"); got != wantTargets[index] {
			t.Errorf("flag %d target = %q, want %q", index, got, wantTargets[index])
		}
	}
}

func TestSourceProjectCommandSerializesNumericEnumsAsCLIValues(t *testing.T) {
	t.Parallel()
	command := newOrderedObject()
	command.set("path", "items update")
	contract := sourceActionContract{
		Fields: map[string]any{
			"access_level": map[string]any{"type": "integer", "enum": []any{json.Number("-1"), json.Number("0"), json.Number("30")}},
		},
		BareStringFields: map[string]bool{},
		SecretFields:     map[string]bool{},
		Required:         map[string]bool{"access_level": true},
	}
	if !sourceProjectCommand(command, contract) {
		t.Fatal("source command should gain the declared numeric enum field")
	}
	raw, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	var loaded engine.CLICommand
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("numeric enum CLI command must remain loadable: %v", err)
	}
	if got := loaded.Flags[0].Type; got != "integer" {
		t.Fatalf("numeric enum CLI type = %q, want integer", got)
	}
	if got, want := loaded.Flags[0].Values, []string{"-1", "0", "30"}; !slices.Equal(got, want) {
		t.Fatalf("numeric enum CLI values = %#v, want %#v", got, want)
	}
}

func TestSourceProjectionPreservesOnlySourceBackedDeclaredBatch(t *testing.T) {
	t.Parallel()
	var action orderedJSON
	if err := json.Unmarshal([]byte(`{
  "name":"submit_batch","method":"POST","path":"/batch","body_type":"declared_batch",
  "declared_batch":{
    "max_actions":10,"allowed_actions":["create_item"],"allowed_methods":["POST"],
    "provider_envelope_field":"data","provider_actions_field":"actions","provider_method_field":"method",
    "provider_path_field":"relative_path","provider_data_field":"data","inner_body_field":"data",
    "response_envelope_field":"data","response_status_field":"status_code"
  },
  "record_schema":{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","additionalProperties":false,"required":["actions"],"properties":{"actions":{"type":"array","minItems":1,"maxItems":10,"items":{"type":"object","additionalProperties":false,"required":["action","record"],"properties":{"action":{"type":"string","maxLength":128},"record":{"type":"object","additionalProperties":true,"maxProperties":256}}}}}},
  "risk":"high"
}`), &action); err != nil {
		t.Fatal(err)
	}
	inventory := &sourceImportBatchActionInventory{
		SourceDocument: "provider", SourceOperation: "alpha.rest.submitBatch", MaxActions: 10,
		ProviderMethods:       []string{"get", "post", "put", "delete", "patch", "head"},
		RequestEnvelopeField:  "data",
		RequestActionsField:   "actions",
		ActionMethodField:     "method",
		ActionPathField:       "relative_path",
		ActionDataField:       "data",
		ResponseEnvelopeField: "data",
		ResponseStatusField:   "status_code",
	}
	operation := sourceOperationDescriptor{SourceID: inventory.SourceOperation, Method: "POST", Path: "/batch", BatchAction: inventory}
	contract, err := sourceContractForAction(operation, action.root)
	if err != nil {
		t.Fatalf("project source-backed declared batch: %v", err)
	}
	if changed := sourceProjectAction(action.root, contract); changed {
		t.Fatalf("source-backed declared batch action drifted: %#v", action.root)
	}
	var command orderedJSON
	if err := json.Unmarshal([]byte(`{"path":"batch submit","flags":[{"name":"actions-json","summary":"typed actions","type":"json","maps_to":"record.actions","max_bytes":1048576,"required":true}]}`), &command); err != nil {
		t.Fatal(err)
	}
	if changed := sourceProjectCommand(command.root, contract); changed {
		t.Fatalf("source-backed declared batch command drifted: %#v", command.root)
	}
	encodedAction, err := json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	var loadedAction engine.WriteAction
	if err := json.Unmarshal(encodedAction, &loadedAction); err != nil {
		t.Fatal(err)
	}
	encodedCommand, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	var loadedCommand engine.CLICommand
	if err := json.Unmarshal(encodedCommand, &loadedCommand); err != nil {
		t.Fatal(err)
	}
	if !sourceActionCoversOperation(loadedAction, loadedCommand, operation) {
		t.Fatal("source-backed declared batch lost its closed spec or record schema during executable-coverage validation")
	}

	explode := false
	operation.Request.Query = []sourceParameterDescriptor{
		{Name: "opt_fields", Schema: map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, Wire: sourceParameterWireDescriptor{Style: "form", Explode: &explode}},
		{Name: "opt_pretty", Schema: map[string]any{"type": "boolean"}},
	}
	contract, err = sourceContractForAction(operation, action.root)
	if err != nil {
		t.Fatalf("project source-backed declared batch query: %v", err)
	}
	if !sourceProjectAction(action.root, contract) || !sourceProjectCommand(command.root, contract) {
		t.Fatal("source-backed declared batch query fields were not projected")
	}
	encodedAction, err = json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encodedAction, &loadedAction); err != nil {
		t.Fatal(err)
	}
	encodedCommand, err = json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encodedCommand, &loadedCommand); err != nil {
		t.Fatal(err)
	}
	if got := loadedAction.Query["opt_fields"].Template; got != "{{ record.opt_fields | join:, }}" {
		t.Fatalf("declared batch opt_fields query template = %q, want source form/explode=false encoding", got)
	}
	if !sourceActionCoversOperation(loadedAction, loadedCommand, operation) {
		t.Fatalf("source-backed declared batch with projected outer query fields was not executable-complete: action=%+v command=%+v", loadedAction, loadedCommand)
	}

	wrongQuery := loadedAction
	wrongQuery.Query = make(map[string]engine.QueryParam, len(loadedAction.Query))
	for name, query := range loadedAction.Query {
		wrongQuery.Query[name] = query
	}
	query := wrongQuery.Query["opt_fields"]
	query.Template = "{{ record.opt_fields }}"
	wrongQuery.Query["opt_fields"] = query
	if sourceActionCoversOperation(wrongQuery, loadedCommand, operation) {
		t.Fatal("declared batch with a non-source query encoding passed executable coverage")
	}
	wrongSchema := loadedAction
	wrongSchema.RecordSchema = json.RawMessage(`{"type":"object","additionalProperties":false,"required":["actions"],"properties":{"actions":{"type":"array","minItems":1,"maxItems":10}}}`)
	if sourceActionCoversOperation(wrongSchema, loadedCommand, operation) {
		t.Fatal("declared batch with an incomplete source-projected record schema passed executable coverage")
	}
	wrongBatch := loadedAction
	wrongBatch.DeclaredBatch = new(engine.DeclaredBatchSpec)
	*wrongBatch.DeclaredBatch = *loadedAction.DeclaredBatch
	wrongBatch.DeclaredBatch.MaxActions = 9
	if sourceActionCoversOperation(wrongBatch, loadedCommand, operation) {
		t.Fatal("declared batch with a source-inconsistent batch contract passed executable coverage")
	}

	withoutInventory := operation
	withoutInventory.BatchAction = nil
	if _, err := sourceContractForAction(withoutInventory, action.root); err == nil || !strings.Contains(err.Error(), "no source-backed batch action inventory") {
		t.Fatalf("unbound declared batch error = %v", err)
	}
	wrongMax := *inventory
	wrongMax.MaxActions = 9
	operation.BatchAction = &wrongMax
	if _, err := sourceContractForAction(operation, action.root); err == nil || !strings.Contains(err.Error(), "max_actions") {
		t.Fatalf("mismatched declared batch error = %v", err)
	}
}

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

func TestSourceProjectionGeneratedParameterizedCommandIsRuntimeValidAndStable(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{
  "schema_version": 1,
  "actions": [{
    "name": "delete_branch_restriction", "kind": "delete", "method": "DELETE",
    "path": "/repositories/{{ record.workspace }}/{{ record.repo_slug }}/branch-restrictions/{{ record.id }}",
    "path_fields": ["workspace", "repo_slug", "id"],
    "record_schema": {"type":"object","additionalProperties":false,"properties":{}},
    "risk": "high", "confirm": "destructive"
  }]
}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[]}`)

	operation := sourceOperationDescriptor{
		Connector: "bitbucket",
		// This is intentionally the source form that used to reach the command
		// path verbatim. The command identity must instead come from Method/Path.
		SourceID: "bitbucket.rest.delete-/repositories/{workspace}/{repo-slug}/branch-restrictions/{id}",
		Method:   "DELETE",
		Path:     "/repositories/{workspace}/{repo_slug}/branch-restrictions/{id}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{
			{Name: "workspace", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "repo_slug", Required: true, Schema: map[string]any{"type": "string"}},
			{Name: "id", Required: true, Schema: map[string]any{"type": "string"}},
		}},
	}

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project parameterized source operation: %v", err)
	}
	if stats.CLI != 1 {
		t.Fatalf("projected CLI updates = %d, want 1", stats.CLI)
	}

	var surface engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, cliPath)), &surface); err != nil {
		t.Fatalf("decode projected CLI surface: %v", err)
	}
	if len(surface.Commands) != 1 {
		t.Fatalf("projected commands = %d, want 1", len(surface.Commands))
	}
	command := surface.Commands[0]
	if strings.ContainsAny(command.Path, "{}") {
		t.Fatalf("generated command path retained a raw source parameter: %q", command.Path)
	}
	if command.Path != sourceProjectionGeneratedCommandPath(operation) {
		t.Fatalf("generated command path = %q, want operation-derived %q", command.Path, sourceProjectionGeneratedCommandPath(operation))
	}
	for _, field := range []string{"workspace", "repo_slug", "id"} {
		found := false
		for _, flag := range command.Flags {
			if flag.MapsTo == "record."+field {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("generated command lost path binding record.%s: %+v", field, command.Flags)
		}
	}

	var writes struct {
		Actions []engine.WriteAction `json:"actions"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, writesPath)), &writes); err != nil {
		t.Fatalf("decode projected writes: %v", err)
	}
	connector := engine.New(engine.Bundle{
		Name:       "bitbucket",
		Metadata:   engine.Metadata{Name: "bitbucket"},
		Writes:     writes.Actions,
		CLISurface: &surface,
	}, nil)
	if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
		t.Fatalf("real commandrunner rejected generated path %q before the credential boundary: %v", command.Path, err)
	}
	planned, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
		Path: strings.Fields(command.Path),
		Flags: map[string][]string{
			"workspace": {"acme"},
			"repo-slug": {"docs"},
			"id":        {"42"},
		},
	})
	if err != nil {
		t.Fatalf("real commandrunner lost path parameter flags: %v", err)
	}
	if got, want := map[string]any(planned.Record), map[string]any{"workspace": "acme", "repo_slug": "docs", "id": "42"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("planned path parameter bindings = %#v, want %#v", got, want)
	}

	before := readProjectionFixture(t, cliPath)
	stats, err = projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("repeat projection: %v", err)
	}
	if stats.Changed() || readProjectionFixture(t, cliPath) != before {
		t.Fatalf("repeat projection changed stable generated command: stats=%+v", stats)
	}

	parameterA := operation
	parameterA.SourceID = "same-source-id"
	parameterA.Path = "/repositories/{workspace}/{repo_slug}/branch-restrictions/{a}"
	parameterB := parameterA
	parameterB.Path = "/repositories/{workspace}/{repo_slug}/branch-restrictions/{b}"
	literal := parameterA
	literal.Path = "/repositories/{workspace}/{repo_slug}/branch-restrictions/a"
	paths := map[string]struct{}{
		sourceProjectionGeneratedCommandPath(parameterA): {},
		sourceProjectionGeneratedCommandPath(parameterB): {},
		sourceProjectionGeneratedCommandPath(literal):    {},
	}
	if len(paths) != 3 {
		t.Fatalf("operation-derived command identities collided: a=%q b=%q literal=%q",
			sourceProjectionGeneratedCommandPath(parameterA),
			sourceProjectionGeneratedCommandPath(parameterB),
			sourceProjectionGeneratedCommandPath(literal))
	}

	validLegacyOperation := sourceOperationDescriptor{SourceID: "github.repos.create", Method: "POST", Path: "/repos/{owner}/{repo}"}
	validLegacy := newOrderedObject()
	validLegacy.set("path", sourceProjectionTestLegacyGeneratedCommandPath(validLegacyOperation))
	validLegacy.set("approval", sourceProjectionApproval(newOrderedObject()))
	if sourceProjectionRefreshGeneratedCommandMetadata(validLegacy, validLegacyOperation, newOrderedObject()) {
		t.Fatalf("reachable legacy command was renamed or modified: %q", stringField(validLegacy, "path"))
	}
	invalidLegacy := newOrderedObject()
	invalidLegacy.set("path", sourceProjectionTestLegacyGeneratedCommandPath(operation))
	if !sourceProjectionRefreshGeneratedCommandMetadata(invalidLegacy, operation, newOrderedObject()) {
		t.Fatal("invalid legacy generated command was not migrated")
	}
	if got, want := stringField(invalidLegacy, "path"), sourceProjectionGeneratedCommandPath(operation); got != want {
		t.Fatalf("invalid legacy command path = %q, want %q", got, want)
	}
	if got, want := stringField(invalidLegacy, "approval"), sourceProjectionApproval(newOrderedObject()); got != want {
		t.Fatalf("invalid legacy command approval = %q, want refreshed %q", got, want)
	}
}

func TestSourceProjectionMaterializesEligibleSourceLockLanesWithMappedRoute(t *testing.T) {
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "metadata.json"), `{"name":"alpha","capabilities":{"read":false,"write":false}}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "streams.json"), `{"schema_version":1,"base":{"pagination":{"type":"none"}},"streams":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{"api":"alpha","endpoints":[{"method":"GET","path":"/widgets/{id}"},{"method":"DELETE","path":"/widgets/{id}"}]}`)

	read := sourceOperationDescriptor{
		Connector: "alpha", Protocol: "rest", SourceID: "getSourceWidget", ProviderOperationID: "getSourceWidget",
		Method: "GET", Path: "/api/v1/widgets/{id}", MappingPath: "/widgets/{id}",
		Source:  sourceImportSource{URL: "https://provider.invalid/openapi.json", Location: `paths["/api/v1/widgets/{id}"].get`},
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{Name: "id", Required: true, Schema: map[string]any{"type": "string"}}}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: sourceOperationExecutionFoundation, Location: "source operation getSourceWidget"}}},
	}
	write := sourceOperationDescriptor{
		Connector: "alpha", Protocol: "rest", SourceID: "deleteSourceWidget", ProviderOperationID: "deleteSourceWidget",
		Method: "DELETE", Path: "/api/v1/widgets/{id}", MappingPath: "/widgets/{id}",
		Source:  sourceImportSource{URL: "https://provider.invalid/openapi.json", Location: `paths["/api/v1/widgets/{id}"].delete`},
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{Name: "id", Required: true, Schema: map[string]any{"type": "string"}}}},
		Output:  sourceOutputDescriptor{Class: sourceOutputStatus, Success: []sourceOutputVariant{{Status: "204", Class: sourceOutputStatus}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: sourceNonExecutableMutationDispositionFoundation, Location: "source operation deleteSourceWidget"}}, NonExecutableMutation: &sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{SourceID: "deleteSourceWidget", Method: "DELETE", Path: "/api/v1/widgets/{id}"}, Reason: "no declaration"}},
	}
	bodyWrite := sourceOperationDescriptor{
		Connector: "alpha", Protocol: "rest", SourceID: "createSourceWidget", ProviderOperationID: "createSourceWidget",
		Method: "POST", Path: "/api/v1/widgets", MappingPath: "/widgets",
		Source: sourceImportSource{URL: "https://provider.invalid/openapi.json", Location: `paths["/api/v1/widgets"].post`},
		Request: sourceRequestDescriptor{Body: &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
			"type": "object", "properties": map[string]any{
				"title":    map[string]any{"type": "string"},
				"metadata": map[string]any{"type": "object", "additionalProperties": true},
			}, "required": []any{"title"},
		}}, MediaType: "application/json"},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "201", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: "cli-request-schema-foundation-r1", Location: "request body"}, {Foundation: sourceNonExecutableMutationDispositionFoundation, Location: "source operation createSourceWidget"}}, NonExecutableMutation: &sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{SourceID: "createSourceWidget", Method: "POST", Path: "/api/v1/widgets"}, Reason: "no declaration"}},
	}
	blocked := sourceOperationDescriptor{
		Connector: "alpha", Protocol: "rest", SourceID: "getOpenWidget", ProviderOperationID: "getOpenWidget",
		Method: "GET", Path: "/api/v1/open-widgets", MappingPath: "/open-widgets",
		Source:  sourceImportSource{URL: "https://provider.invalid/openapi.json", Location: `paths["/api/v1/open-widgets"].get`},
		Request: sourceRequestDescriptor{Body: &sourceRequestBodyDescriptor{Required: false, Schema: map[string]any{"type": "object", "additionalProperties": true}}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: "cli-request-schema-foundation-r1", Location: "request body"}}},
	}

	stats, err := projectEligibleSourceLockLanesToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{read, write, bodyWrite, blocked}}, false)
	if err != nil {
		t.Fatalf("materialize eligible source-lock lanes: %v", err)
	}
	if stats.Operations != 1 || stats.Writes != 2 || stats.CLI != 5 || stats.Surface != 3 {
		t.Fatalf("materialization stats = %+v, want one read operation, two writes, five commands, and three surface bindings", stats)
	}

	var operations struct {
		Operations []struct {
			ID              string `json:"id"`
			SourceOperation struct {
				ID     string `json:"id"`
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"source_operation"`
			REST struct {
				Path string `json:"path"`
			} `json:"rest"`
		} `json:"operations"`
	}
	operationsRaw := readProjectionFixture(t, filepath.Join(bundleDir, "operations.json"))
	if err := json.Unmarshal([]byte(operationsRaw), &operations); err != nil {
		t.Fatalf("decode generated operations: %v", err)
	}
	if len(operations.Operations) != 1 || operations.Operations[0].REST.Path != "/widgets/{id}" || operations.Operations[0].SourceOperation.Path != "/widgets/{id}" || operations.Operations[0].SourceOperation.ID != read.SourceID {
		t.Fatalf("generated read lost mapped source binding: %+v", operations.Operations)
	}

	var cli engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"))), &cli); err != nil {
		t.Fatalf("decode generated CLI: %v", err)
	}
	bySourceIntent := map[string]map[string]engine.CLICommand{}
	for _, command := range cli.Commands {
		if bySourceIntent[command.SourceOperation] == nil {
			bySourceIntent[command.SourceOperation] = map[string]engine.CLICommand{}
		}
		bySourceIntent[command.SourceOperation][command.Intent] = command
	}
	if command, found := bySourceIntent[read.SourceID]["direct_read"]; !found || command.Availability != "implemented" || command.Operation == "" || len(command.APISurface) != 1 || command.APISurface[0].Path != "/widgets/{id}" {
		t.Fatalf("generated direct read = %+v, want mapped implemented source-bound command", command)
	}
	if command, found := bySourceIntent[write.SourceID]["reverse_etl"]; !found || command.Availability != "implemented" || command.Write == "" || len(command.APISurface) != 1 || command.APISurface[0].Path != "/widgets/{id}" {
		t.Fatalf("generated reverse-ETL command = %+v, want mapped implemented source-bound command", command)
	}
	if command, found := bySourceIntent[bodyWrite.SourceID]["reverse_etl"]; !found || command.Availability != "implemented" || command.Write == "" || len(command.APISurface) != 1 || command.APISurface[0].Path != "/widgets" {
		t.Fatalf("generated JSON-body reverse-ETL command = %+v, want mapped implemented source-bound command", command)
	}
	for _, source := range []sourceOperationDescriptor{write, bodyWrite} {
		command, found := bySourceIntent[source.SourceID]["direct_write"]
		if !found || command.Availability != "implemented" || command.Write == "" || len(command.APISurface) != 1 || command.APISurface[0].Path != source.MappingPath {
			t.Fatalf("generated direct-write command for %s = %+v, want same bounded action", source.SourceID, command)
		}
	}
	if _, found := bySourceIntent[blocked.SourceID]; found {
		t.Fatalf("blocked source contract unexpectedly received a command: %+v", bySourceIntent[blocked.SourceID])
	}
	var writes struct {
		Actions []engine.WriteAction `json:"actions"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "writes.json"))), &writes); err != nil {
		t.Fatalf("decode generated writes: %v", err)
	}
	var declaredOperations struct {
		Operations []engine.OperationSpec `json:"operations"`
	}
	if err := json.Unmarshal([]byte(operationsRaw), &declaredOperations); err != nil {
		t.Fatalf("decode generated operation contracts: %v", err)
	}
	connector := engine.New(engine.Bundle{Name: "alpha", Metadata: engine.Metadata{Name: "alpha"}, Operations: declaredOperations.Operations, Writes: writes.Actions, CLISurface: &cli}, nil)
	if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: strings.Fields(bySourceIntent[write.SourceID]["reverse_etl"].Path), Flags: map[string][]string{"id": {"42"}}}); err != nil {
		t.Fatalf("generated reverse-ETL command lost source path flag: %v", err)
	}
	bodyCommand, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: strings.Fields(bySourceIntent[bodyWrite.SourceID]["direct_write"].Path), Flags: map[string][]string{"title": {"closed named body"}, "metadata": {`{"retained":"bounded"}`}}})
	if err != nil {
		t.Fatalf("generated JSON-body reverse-ETL command rejected named fields: %v", err)
	}
	if bodyCommand.Record["title"] != "closed named body" {
		t.Fatalf("generated JSON-body record = %#v, want named title", bodyCommand.Record)
	}
	if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{Path: strings.Fields(bySourceIntent[bodyWrite.SourceID]["direct_write"].Path), Flags: map[string][]string{"title": {"closed named body"}, "body": {`{"undeclared":true}`}}}); err == nil || !strings.Contains(err.Error(), "unknown flag --body") {
		t.Fatalf("source-open JSON body raw escape error = %v, want unknown --body refusal", err)
	}
	var bodyAction *engine.WriteAction
	for index := range writes.Actions {
		if writes.Actions[index].Name == bySourceIntent[bodyWrite.SourceID]["reverse_etl"].Write {
			bodyAction = &writes.Actions[index]
			break
		}
	}
	if bodyAction == nil || bodyAction.BodyType != "json" || !slices.Equal(bodyAction.SuccessStatuses, []int{201}) {
		t.Fatalf("generated JSON-body action = %+v, want exact JSON body and 201 response", bodyAction)
	}
	var bodySchema struct {
		AdditionalProperties bool                       `json:"additionalProperties"`
		Properties           map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(bodyAction.RecordSchema, &bodySchema); err != nil || bodySchema.AdditionalProperties || len(bodySchema.Properties) != 2 {
		t.Fatalf("generated JSON-body record schema = %s err=%v, want closed named root", bodyAction.RecordSchema, err)
	}
	completeWrite := write
	completeWrite.Runtime = sourceRuntimeReachability{}
	if findings := validateSourceExecutableCoverage(engine.Bundle{Name: "alpha", Writes: writes.Actions, CLISurface: &cli}, "fixture", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{completeWrite}}); len(findings) != 0 {
		t.Fatalf("mapped source mutation was not covered by generated action/command: %+v", findings)
	}
	staleCLI := strings.Replace(readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json")), `"max_bytes": 32768`, `"max_bytes": 0`, 1)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), staleCLI)
	staleWrite := write
	staleWrite.Runtime = sourceRuntimeReachability{
		MergeBlocked:          true,
		Gaps:                  []sourceContractGap{{Foundation: sourceNonExecutableMutationDispositionFoundation, Location: "source operation deleteSourceWidget"}},
		NonExecutableMutation: &sourceNonExecutableMutationDisposition{Source: sourceOperationCitation{SourceID: write.SourceID, Method: write.Method, Path: write.Path}, Reason: "stale generated command flag"},
	}
	var generatedWrites orderedJSON
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "writes.json"))), &generatedWrites); err != nil {
		t.Fatalf("decode stale generated writes: %v", err)
	}
	if action := sourceProjectionActionForEndpoint(generatedWrites.root, staleWrite.Method, staleWrite.MappingPath); action == nil {
		t.Fatalf("stale generated action not found at mapped endpoint %s", staleWrite.MappingPath)
	}
	if !sourceProjectionGeneratedNoBodyMutationExists(generatedWrites.root, staleWrite) {
		t.Fatal("stale generated mutation was not recognized for refresh")
	}
	stats, err = projectEligibleSourceLockLanesToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{read, staleWrite, blocked}}, false)
	if err != nil {
		t.Fatalf("refresh stale generated mutation: %v", err)
	}
	if stats.CLI != 3 || stats.Operations != 0 || stats.Writes != 0 {
		t.Fatalf("stale generated mutation refresh = %+v, want reverse-ETL and direct-write CLI refresh", stats)
	}
	if !strings.Contains(readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json")), `"max_bytes": 32768`) {
		t.Fatal("stale generated mutation did not restore source-derived CLI byte cap")
	}

	// A second run with the exact same source result must be a true no-op. This
	// deliberately differs from the stale-flag reconciliation above: no source
	// identity or pre-existing action is omitted from this idempotence fixture.
	fullResult := sourceImportResult{Operations: []sourceOperationDescriptor{read, write, bodyWrite, blocked}}
	if _, err := projectEligibleSourceLockLanesToBundle(bundleDir, fullResult, false); err != nil {
		t.Fatalf("restore complete source result before idempotence check: %v", err)
	}
	stats, err = projectEligibleSourceLockLanesToBundle(bundleDir, fullResult, true)
	if err != nil {
		t.Fatalf("repeat materialization: %v", err)
	}
	if stats.Changed() {
		t.Fatalf("exact repeat materialization drifted: %+v", stats)
	}
}

func TestSourceProjectionGeneratedDirectReadProjectsGitLabBracketedQueryAlias(t *testing.T) {
	operation := sourceOperationDescriptor{
		Connector: "gitlab", Protocol: "rest", SourceID: "getApiV4Projects", ProviderOperationID: "getApiV4Projects",
		Method: "GET", Path: "/api/v4/projects", MappingPath: "/api/v4/projects",
		Source: sourceImportSource{
			URL:      "https://docs.gitlab.com/api/projects/",
			Location: `GET /api/v4/projects query filter[state]`,
		},
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{
			{Name: "filter[state]", Schema: map[string]any{"type": "string"}},
		}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: sourceOperationExecutionFoundation, Location: "source operation getApiV4Projects"}}},
	}
	if parameter, unsafe := sourceProjectionUnsafeProviderQueryParameter(operation); unsafe {
		t.Fatalf("source-backed GitLab key %q remained unsafe after alias foundation", parameter.Name)
	}
	if !sourceProjectionGeneratedDirectReadEligible(operation) {
		t.Fatal("source-backed bracketed GitLab query key was not eligible for typed direct-read projection")
	}
	rest := newOrderedObject()
	if !sourceProjectionSyncReadParameters(rest, operation) {
		t.Fatal("source projection did not retain source-backed query parameter")
	}
	command := newOrderedObject()
	if got := deriveCommandParameterFlags(command, rest); got != 1 {
		t.Fatalf("derived flag changes = %d, want 1", got)
	}
	flags := arrayField(command, "flags")
	if len(flags) != 1 {
		t.Fatalf("derived flags = %#v, want one", flags)
	}
	flag, ok := flags[0].(*orderedObject)
	if !ok || stringField(flag, "name") != "filter-state" || stringField(flag, "maps_to") != "query.filter[state]" {
		t.Fatalf("derived GitLab alias flag = %#v, want --filter-state -> query.filter[state]", flag)
	}
}

func TestGitLabSourceDescriptorBracketedQueryKeyHasClosedAlias(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "gitlab", "sources", "gitlab-operation-descriptor.json"))
	if err != nil {
		t.Fatalf("read retained GitLab source descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := json.Unmarshal(raw, &descriptor); err != nil {
		t.Fatalf("decode retained GitLab source descriptor: %v", err)
	}
	for _, operation := range descriptor.Operations {
		if operation.SourceID != "getApiV4ProjectsIdVariablesKey" {
			continue
		}
		for _, parameter := range operation.Request.Query {
			if parameter.Name != "filter[environment_scope]" {
				continue
			}
			if got, ok := engine.ProviderQueryParameterCLIName(parameter.Name); !ok || got != "filter-environment-scope" {
				t.Fatalf("retained GitLab key alias = (%q, %t), want (filter-environment-scope, true)", got, ok)
			}
			if unsafeParameter, unsafe := sourceProjectionUnsafeProviderQueryParameter(operation); unsafe {
				t.Fatalf("retained GitLab source operation remained unsafe at %q", unsafeParameter.Name)
			}
			return
		}
		t.Fatal("retained GitLab read operation lost filter[environment_scope] source evidence")
	}
	t.Fatal("retained GitLab read operation getApiV4ProjectsIdVariablesKey not found")
}

func TestSourceProjectionGeneratedDirectReadRefusesUnsupportedProviderQueryParameter(t *testing.T) {
	operation := sourceOperationDescriptor{
		Connector: "alpha", Protocol: "rest", SourceID: "getFilteredWidgets", ProviderOperationID: "getFilteredWidgets",
		Method: "GET", Path: "/api/v1/widgets", MappingPath: "/widgets",
		Source: sourceImportSource{URL: "https://provider.invalid/openapi.json", Location: `paths["/api/v1/widgets"].get`},
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{
			{Name: "$filter", Schema: map[string]any{"type": "string"}},
		}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: sourceOperationExecutionFoundation, Location: "source operation getFilteredWidgets"}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if changed := sourceProjectionApplyUnsafeProviderQueryParameterGaps(engine.Bundle{}, &result); changed != 1 {
		t.Fatalf("unsafe provider query gap changes = %d, want 1", changed)
	}
	if changed := sourceProjectionApplyUnsafeProviderQueryParameterGaps(engine.Bundle{}, &result); changed != 0 {
		t.Fatalf("unsafe provider query gap was not idempotent: changes = %d", changed)
	}
	projected := result.Operations[0]
	if len(projected.Runtime.Gaps) != 2 {
		t.Fatalf("unsafe provider query gaps = %+v, want existing execution and alias gaps", projected.Runtime.Gaps)
	}
	aliasGap := projected.Runtime.Gaps[1]
	if aliasGap.Foundation != sourceProviderParameterAliasFoundation || aliasGap.Phase != "request" || aliasGap.Location != "query parameter $filter" {
		t.Fatalf("unsafe provider query gap = %+v, want exact Atlas candidate disposition", aliasGap)
	}
	if sourceProjectionGeneratedDirectReadEligible(projected) {
		t.Fatal("generated direct read admitted an unsupported provider query key outside the approved alias grammar")
	}
}

func TestSourceProjectionGeneratedDirectReadRejectsProviderAliasCollision(t *testing.T) {
	operation := sourceOperationDescriptor{
		Connector: "gitlab", Protocol: "rest", SourceID: "getApiV4Projects", ProviderOperationID: "getApiV4Projects",
		Method: "GET", Path: "/api/v4/projects", MappingPath: "/api/v4/projects",
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{
			{Name: "filter[state]", Schema: map[string]any{"type": "string"}},
			{Name: "filter-state", Schema: map[string]any{"type": "string"}},
		}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: sourceOperationExecutionFoundation, Location: "source operation getApiV4Projects"}}},
	}
	parameter, unsafe := sourceProjectionUnsafeProviderQueryParameter(operation)
	if !unsafe || parameter.Name != "filter-state" {
		t.Fatalf("provider alias collision = (%+v, %t), want filter-state collision", parameter, unsafe)
	}
}

func TestSourceProjectionPrunesUnsafeGeneratedDirectReadToCitedBlockedSurface(t *testing.T) {
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "metadata.json"), `{"name":"alpha","capabilities":{"read":false,"write":false}}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "streams.json"), `{"schema_version":1,"base":{"pagination":{"type":"none"}},"streams":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{"api":"alpha","endpoints":[{"method":"GET","path":"/widgets"}]}`)

	operation := sourceOperationDescriptor{
		Connector: "alpha", Protocol: "rest", SourceID: "getWidgets", ProviderOperationID: "getWidgets",
		Method: "GET", Path: "/api/v1/widgets", MappingPath: "/widgets",
		Source:  sourceImportSource{URL: "https://provider.invalid/openapi.json", Location: `paths["/api/v1/widgets"].get`},
		Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{{Name: "visibility", Schema: map[string]any{"type": "string"}}}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
		Runtime: sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{Foundation: sourceOperationExecutionFoundation, Location: "source operation getWidgets"}}},
	}
	if _, err := projectEligibleSourceLockLanesToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false); err != nil {
		t.Fatalf("materialize safe source operation: %v", err)
	}

	operation.Request.Query = []sourceParameterDescriptor{{Name: "$filter", Schema: map[string]any{"type": "string"}}}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if changed := sourceProjectionApplyUnsafeProviderQueryParameterGaps(engine.Bundle{}, &result); changed != 1 {
		t.Fatalf("unsafe provider query gap changes = %d, want 1", changed)
	}
	stats, err := projectEligibleSourceLockLanesToBundle(bundleDir, result, false)
	if err != nil {
		t.Fatalf("reconcile unsafe source operation: %v", err)
	}
	if stats.Operations != 1 || stats.CLI != 1 || stats.Surface != 1 {
		t.Fatalf("unsafe source projection stats = %+v, want generated operation/command retracted and source surface blocked", stats)
	}
	var operations struct {
		Operations []json.RawMessage `json:"operations"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "operations.json"))), &operations); err != nil {
		t.Fatalf("decode pruned operations: %v", err)
	}
	if len(operations.Operations) != 0 {
		t.Fatalf("unsafe provider query retained generated operation: %s", operations.Operations)
	}
	var cli struct {
		Commands []json.RawMessage `json:"commands"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"))), &cli); err != nil {
		t.Fatalf("decode pruned CLI surface: %v", err)
	}
	if len(cli.Commands) != 0 {
		t.Fatalf("unsafe provider query retained generated command: %s", cli.Commands)
	}
	var surface struct {
		Endpoints []struct {
			Operation struct {
				Status string `json:"status"`
				Reason string `json:"reason"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))), &surface); err != nil {
		t.Fatalf("decode pruned API surface: %v", err)
	}
	if len(surface.Endpoints) != 1 || surface.Endpoints[0].Operation.Status != "blocked" || !strings.Contains(surface.Endpoints[0].Operation.Reason, sourceProviderParameterAliasFoundation) {
		t.Fatalf("unsafe provider query lost cited blocked API surface: %+v", surface.Endpoints)
	}
}

func TestGitLabSourceLockSurfaceStatesCurrentEligibleLaneDisposition(t *testing.T) {
	metadataRaw, err := os.ReadFile("../../internal/connectors/defs/gitlab/metadata.json")
	if err != nil {
		t.Fatalf("read GitLab metadata: %v", err)
	}
	var metadata struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		t.Fatalf("decode GitLab metadata: %v", err)
	}
	if !strings.Contains(metadata.Description, "582 source-bound direct reads") || !strings.Contains(metadata.Description, "381 source-bound mutations through direct-write and approval-gated reverse-ETL commands") || strings.Contains(metadata.Description, "no write command") {
		t.Fatalf("GitLab metadata description = %q, want current source-lock lane disposition", metadata.Description)
	}

	cliRaw, err := os.ReadFile("../../internal/connectors/defs/gitlab/cli_surface.json")
	if err != nil {
		t.Fatalf("read GitLab CLI surface: %v", err)
	}
	var surface struct {
		Tagline    string `json:"tagline"`
		HelpTopics []struct {
			Name    string `json:"name"`
			Summary string `json:"summary"`
		} `json:"help_topics"`
	}
	if err := json.Unmarshal(cliRaw, &surface); err != nil {
		t.Fatalf("decode GitLab CLI surface: %v", err)
	}
	if !strings.Contains(surface.Tagline, "604 source-bound direct reads") || !strings.Contains(surface.Tagline, "381 source-bound mutations through direct-write and approval-gated reverse-ETL commands") {
		t.Fatalf("GitLab CLI tagline = %q, want current source-lock lane disposition", surface.Tagline)
	}
	topics := strings.Join(func() []string {
		values := make([]string, 0, len(surface.HelpTopics))
		for _, topic := range surface.HelpTopics {
			values = append(values, topic.Name+": "+topic.Summary)
		}
		return values
	}(), "\n")
	if !strings.Contains(topics, "1,752") || !strings.Contains(topics, "785") || !strings.Contains(topics, "1,845") || strings.Contains(topics, "only the four existing stream reads") {
		t.Fatalf("GitLab CLI help topics = %q, want source-lock counts and no stale four-stream-only claim", topics)
	}

	docsRaw, err := os.ReadFile("../../internal/connectors/defs/gitlab/docs.md")
	if err != nil {
		t.Fatalf("read GitLab connector docs: %v", err)
	}
	docs := string(docsRaw)
	if !strings.Contains(docs, "582 source-bound direct reads") || !strings.Contains(docs, "381 source-bound mutations") || !strings.Contains(docs, "direct-write and approval-gated reverse-ETL") || !strings.Contains(docs, "785") || !strings.Contains(docs, "1,845") || strings.Contains(docs, "only the four existing stream reads executable") {
		t.Fatalf("GitLab connector docs have stale source-lock lane disposition")
	}
}

func sourceProjectionTestLegacyGeneratedCommandPath(operation sourceOperationDescriptor) string {
	path := strings.NewReplacer("/", " ", "_", "-").Replace(operation.SourceID)
	return "api " + path
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
	if _, err := parseSourceImportLock([]byte(`{}`), "alpha"); err == nil || !strings.Contains(err.Error(), "unsupported schema version 0") {
		t.Fatalf("invalid source lock parse error = %v, want unsupported schema version", err)
	}
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "alpha-operation-source-lock.json"), sourceProjectionMinimalLegacyLock(t, &operation))
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

func TestSourceProjectionRequiresExplicitReadOnlyNonMutationDeclaration(t *testing.T) {
	source := sourceOperationDescriptor{
		Connector: "alpha",
		SourceID:  "alpha.widgets.get",
		Method:    "GET",
		Path:      "/widgets",
	}
	file := "sources/alpha-operation-descriptor.json"
	undeclared := engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{}}
	if findings := validateSourceExecutableCoverage(undeclared, file, sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "no reachable executable operation") {
		t.Fatalf("undeclared read-only operation findings = %+v", findings)
	}

	declared := engine.Bundle{
		Name: "alpha",
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
			Method: "GET",
			Path:   "/widgets",
			Operation: &engine.SurfaceOperation{
				Model:            "read_only",
				Status:           "blocked",
				Risk:             "low",
				BlockedByDefault: true,
				Reason:           "The connector intentionally does not implement this source-cited read.",
				Notes:            "Named policy: source-cited-read-only-operations-r1",
			},
		}}},
		CLISurface: &engine.CLISurface{},
	}
	if findings := validateSourceExecutableCoverage(declared, file, sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 0 {
		t.Fatalf("declared read-only operation findings = %+v", findings)
	}

	declared.Operations = []engine.OperationSpec{{
		ID:   "alpha.widgets.get",
		Kind: "rest_read",
		REST: &engine.RESTOperationSpec{Method: "GET", Path: "/widgets", MaxBytes: 1024},
	}}
	declared.CLISurface.Commands = []engine.CLICommand{{
		Path:         "widgets get",
		Summary:      "get widget",
		Intent:       "direct_read",
		Availability: "implemented",
		Operation:    "alpha.widgets.get",
	}}
	if findings := validateSourceExecutableCoverage(declared, file, sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{source}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "read-only declaration conflicts with executable operation") {
		t.Fatalf("read-only executable contradiction findings = %+v", findings)
	}

	mutation := source
	mutation.SourceID = "alpha.widgets.create"
	mutation.Method = "POST"
	mutationSurface := *declared.Surface
	mutationSurface.Endpoints = append([]engine.SurfaceEndpoint(nil), declared.Surface.Endpoints...)
	mutationSurface.Endpoints[0].Method = "POST"
	if findings := validateSourceExecutableCoverage(engine.Bundle{Name: "alpha", Surface: &mutationSurface, CLISurface: &engine.CLISurface{}}, file, sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{mutation}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "read-only declaration cannot cover a mutating source operation") {
		t.Fatalf("mutating read-only declaration findings = %+v", findings)
	}
}

func TestSourceProjectionMappingIgnoresRetentionAndEmbeddedSourceOperation(t *testing.T) {
	lockRaw := []byte(`{
  "schema_version": 2,
  "connector": "alpha",
  "source_contract": {"kind": "provider-evidence-only"},
  "rest": {
    "source_url": "https://provider.example.test/openapi.json",
    "sha256": {"malformed": true},
    "bytes": "not-a-byte-count",
    "openapi": false,
    "operations": [{
      "id": "alpha.rest.listWidgets",
      "protocol": "rest",
      "method": "GET",
      "path": "/widgets",
      "operation_id": "listWidgets",
      "deprecated": false,
      "source_location": "paths[\"/widgets\"].get",
      "source_operation": {"summary": "List widgets", "responses": {"200": {"description": "ok"}}}
    }]
  },
  "counts": {"rest": 1, "graphql_query": 0, "graphql_mutation": 0, "total": 1}
}`)
	if _, err := parseSourceImportLock(lockRaw, "alpha"); err == nil {
		t.Fatal("strict source-import parser accepted malformed retention or enriched mapping representation")
	}
	if _, err := parseDeclarationAdmissionSourceLock(lockRaw, "alpha"); err != nil {
		t.Fatalf("mapping-only source-lock parser rejected source evidence: %v", err)
	}

	descriptor := sourceImportDescriptorDocument{
		SchemaVersion: 3,
		MergeBlocked:  true,
		Operations: []sourceOperationDescriptor{{
			Connector:           "alpha",
			Protocol:            "rest",
			SourceID:            "alpha.rest.listWidgets",
			ProviderOperationID: "listWidgets",
			Source: sourceImportSource{
				URL:      "https://provider.example.test/openapi.json",
				Location: `paths["/widgets"].get`,
			},
			Method: "get",
			Path:   "/widgets",
			Runtime: sourceRuntimeReachability{
				MergeBlocked: true,
				Gaps: []sourceContractGap{{
					Foundation: "fixture-execution-foundation-r1",
					Location:   "paths[/widgets].get",
					Reason:     "fixture keeps execution outside this mapping-only proof",
				}},
			},
		}},
	}
	descriptorRaw, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatalf("encode source descriptor: %v", err)
	}
	fsys := fstest.MapFS{
		"alpha/sources/alpha-operation-source-lock.json": &fstest.MapFile{Data: lockRaw},
		"alpha/sources/alpha-operation-descriptor.json":  &fstest.MapFile{Data: descriptorRaw},
	}
	bundle := engine.Bundle{Name: "alpha", CLISurface: &engine.CLISurface{}}
	if findings := checkSourceProjection(fsys, bundle); len(findings) != 0 {
		t.Fatalf("mapping-only projection findings = %+v, want none", findings)
	}

	descriptor.Operations[0].Path = "/invented"
	descriptorRaw, err = json.Marshal(descriptor)
	if err != nil {
		t.Fatalf("encode drifted source descriptor: %v", err)
	}
	fsys["alpha/sources/alpha-operation-descriptor.json"] = &fstest.MapFile{Data: descriptorRaw}
	findings := checkSourceProjection(fsys, bundle)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "provider contract drift") {
		t.Fatalf("drifted mapping findings = %+v, want provider contract drift", findings)
	}
}

func TestSourceProjectionSourceReferenceIgnoresRetentionButPreservesClosedGap(t *testing.T) {
	const connector = "alpha"
	raw := sourceImportV3SourceReferenceLock(
		t,
		connector,
		"widgets",
		"https://provider.example.test/openapi.json",
		strings.Repeat("a", 64),
		512,
		"GET",
		"/widgets",
	)
	strictLock, err := parseSourceImportLock(raw, connector)
	if err != nil {
		t.Fatalf("parse valid source-reference lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), strictLock, nil, defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import valid source-reference lock: %v", err)
	}
	descriptorRaw, err := marshalSourceImportResult(result)
	if err != nil {
		t.Fatalf("marshal source-reference descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
		t.Fatalf("decode source-reference descriptor: %v", err)
	}

	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("decode source-reference fixture: %v", err)
	}
	rest, ok := wire["rest"].(map[string]any)
	if !ok {
		t.Fatalf("source-reference fixture REST value = %T, want object", wire["rest"])
	}
	documents, ok := rest["source_documents"].([]any)
	if !ok || len(documents) != 1 {
		t.Fatalf("source-reference fixture documents = %#v, want one", rest["source_documents"])
	}
	document, ok := documents[0].(map[string]any)
	if !ok {
		t.Fatalf("source-reference fixture document = %T, want object", documents[0])
	}
	reference, ok := document["source_reference"].(map[string]any)
	if !ok {
		t.Fatalf("source-reference fixture reference = %T, want object", document["source_reference"])
	}
	reference["sha256"] = map[string]any{"retention": "malformed"}
	raw, err = json.Marshal(wire)
	if err != nil {
		t.Fatalf("encode malformed-retention source reference: %v", err)
	}
	if _, err := parseSourceImportLock(raw, connector); err == nil {
		t.Fatal("strict source-import parser accepted malformed source-reference retention")
	}
	mappingLock, err := parseDeclarationAdmissionSourceLock(raw, connector)
	if err != nil {
		t.Fatalf("mapping-only source-lock parser rejected source-reference identity: %v", err)
	}
	if findings := validateSourceDescriptorAgainstMappingLock(connector, "sources/alpha-operation-descriptor.json", mappingLock, descriptor); len(findings) != 0 {
		t.Fatalf("mapping-only source-reference findings = %+v, want none", findings)
	}

	descriptor.Operations[0].Runtime.Gaps[0].Foundation = "invented-executor"
	findings := validateSourceDescriptorAgainstMappingLock(connector, "sources/alpha-operation-descriptor.json", mappingLock, descriptor)
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "reference contract drift") {
		t.Fatalf("tampered source-reference findings = %+v, want closed-gap refusal", findings)
	}
}

// The three controls below are copied from Asana's locked OpenAPI inventory:
// getAccessRequests (paths["/access_requests"].get), getAgent
// (paths["/agents/{agent_gid}"].get), and getWorkspaces
// (paths["/workspaces"].get).  They exercise a zero-input bounded read, a
// source-owned path input, and an already-proven paginated stream. A stream
// binding adds ETL evidence, but it does not replace a command's user-facing
// direct-read intent.
func TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL(t *testing.T) {
	bundleDir := t.TempDir()
	operationsPath := filepath.Join(bundleDir, "operations.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, operationsPath, `{
  "schema_version": 1,
  "operations": [
    {"id":"get_access_requests","kind":"rest_read","summary":"Get access requests","risk":"none","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/access_requests","max_bytes":1024,"parameters":[]}},
    {"id":"get_agent","kind":"rest_read","summary":"Get an agent","risk":"none","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/agents/{agent_gid}","max_bytes":1024,"parameters":[{"name":"agent_gid","in":"path","type":"string","required":true}]}},
    {"id":"get_workspaces","kind":"stream_etl","summary":"Get workspaces","risk":"none","approval":"none","output_policy":"json_redacted","composite":{"steps":["stream:workspaces"]}},
    {"id":"get_pending","kind":"rest_read","summary":"Get pending","risk":"none","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/pending","max_bytes":1024,"parameters":[]}}
  ]
}`)
	writeProjectionFixture(t, cliPath, `{
  "schema_version": 1,
  "commands": [
    {"path":"access-requests get-access-requests","summary":"Get access requests","intent":"direct_read","availability":"implemented","operation":"get_access_requests","notes":"Blocked until a historical certification lane completes.","api_surface":[{"method":"GET","path":"/access_requests"}]},
    {"path":"agents get-agent","summary":"Get an agent","intent":"direct_read","availability":"implemented","operation":"get_agent","flags":[{"name":"agent-gid","type":"string","maps_to":"query.agent_gid"}],"api_surface":[{"method":"GET","path":"/agents/{agent_gid}"}]},
    {"path":"workspaces get-workspaces","summary":"Get workspaces","intent":"direct_read","availability":"implemented","stream":"workspaces","api_surface":[{"method":"GET","path":"/workspaces"}]},
    {"path":"pending get-pending","summary":"Planned fixed-target Alpha read: Get pending.","intent":"etl","availability":"planned","operation":"get_pending","api_surface":[{"method":"GET","path":"/pending"}]}
  ]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "streams.json"), `{
  "schema_version": 1,
  "base": {"pagination":{"type":"next_url","next_url_path":"next_page.uri"}},
  "streams": [{"name":"workspaces","path":"/workspaces","records":{"path":"data"},"schema":"schemas/workspaces.json"}]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{
  "api":"asana",
  "endpoints":[
    {"method":"GET","path":"/access_requests"},
    {"method":"GET","path":"/agents/{agent_gid}"},
    {"method":"GET","path":"/workspaces","covered_by":{"stream":"workspaces"}},
    {"method":"GET","path":"/pending"}
  ]
}`)

	result := sourceImportResult{Operations: []sourceOperationDescriptor{
		{Connector: "asana", SourceID: "asana.rest.getAccessRequests", ProviderOperationID: "getAccessRequests", Method: "GET", Path: "/access_requests", Request: sourceRequestDescriptor{Query: []sourceParameterDescriptor{{Name: "target", Required: true, Schema: map[string]any{"type": "string"}}, {Name: "limit", Required: false, Schema: map[string]any{"type": "integer"}}, {Name: "offset", Required: false, Schema: map[string]any{"type": "integer"}}}}, Pagination: map[string]any{"type": "next_url", "next_url_path": "next_page.uri", "size_param": "limit", "limit_param": "limit", "offset_param": "offset", "page_size": 100}, Output: sourceOutputDescriptor{Class: sourceOutputJSON}},
		{Connector: "asana", SourceID: "asana.rest.getAgent", ProviderOperationID: "getAgent", Method: "GET", Path: "/agents/{agent_gid}", Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{Name: "agent_gid", Required: true, Schema: map[string]any{"type": "string"}}}}, Output: sourceOutputDescriptor{Class: sourceOutputJSON}},
		{Connector: "asana", SourceID: "asana.rest.getWorkspaces", ProviderOperationID: "getWorkspaces", Method: "GET", Path: "/workspaces", Pagination: map[string]any{"type": "next_url", "next_url_path": "next_page.uri"}, Output: sourceOutputDescriptor{Class: sourceOutputJSON}},
		{Connector: "asana", SourceID: "asana.rest.getPending", ProviderOperationID: "getPending", Method: "GET", Path: "/pending", Output: sourceOutputDescriptor{Class: sourceOutputJSON}},
	}}
	if _, err := projectSourceBoundReadDescriptorToBundle(bundleDir, result, false); err != nil {
		t.Fatalf("project source-bound reads: %v", err)
	}

	var operations struct {
		Operations []struct {
			ID   string `json:"id"`
			REST struct {
				Pagination           map[string]any `json:"pagination"`
				PaginationParameters []struct {
					Name string `json:"name"`
					In   string `json:"in"`
				} `json:"pagination_parameters"`
			} `json:"rest"`
			SourceOperation *struct {
				ID     string `json:"id"`
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"source_operation"`
		} `json:"operations"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, operationsPath)), &operations); err != nil {
		t.Fatalf("decode projected operations: %v", err)
	}
	bound := map[string]string{}
	for _, operation := range operations.Operations {
		if operation.SourceOperation != nil {
			bound[operation.ID] = operation.SourceOperation.ID + " " + operation.SourceOperation.Method + " " + operation.SourceOperation.Path
		}
	}
	for id, want := range map[string]string{
		"get_access_requests": "asana.rest.getAccessRequests GET /access_requests",
		"get_agent":           "asana.rest.getAgent GET /agents/{agent_gid}",
		"get_workspaces":      "asana.rest.getWorkspaces GET /workspaces",
	} {
		if got := bound[id]; got != want {
			t.Fatalf("operation %s source binding = %q, want %q", id, got, want)
		}
	}
	for _, operation := range operations.Operations {
		if operation.ID != "get_access_requests" {
			continue
		}
		wantPagination := map[string]any{"type": "next_url", "next_url_path": "next_page.uri", "size_param": "limit", "limit_param": "limit", "offset_param": "offset", "page_size": float64(100)}
		if !reflect.DeepEqual(operation.REST.Pagination, wantPagination) {
			t.Fatalf("source-bound direct-read pagination = %#v, want %#v", operation.REST.Pagination, wantPagination)
		}
		gotParameters := make([]string, 0, len(operation.REST.PaginationParameters))
		for _, parameter := range operation.REST.PaginationParameters {
			gotParameters = append(gotParameters, parameter.In+"."+parameter.Name)
		}
		if !reflect.DeepEqual(gotParameters, []string{"query.limit", "query.offset"}) {
			t.Fatalf("source-bound direct-read pagination parameters = %#v, want limit/offset only", gotParameters)
		}
	}
	if got := bound["get_pending"]; got != "asana.rest.getPending GET /pending" {
		t.Fatalf("complete planned read operation binding = %q, want exact source identity", got)
	}

	var cli engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, cliPath)), &cli); err != nil {
		t.Fatalf("decode projected CLI surface: %v", err)
	}
	commands := map[string]engine.CLICommand{}
	for _, command := range cli.Commands {
		commands[command.Path] = command
	}
	if command := commands["access-requests get-access-requests"]; command.Intent != "direct_read" || command.Availability != "implemented" || command.SourceOperation != "asana.rest.getAccessRequests" {
		t.Fatalf("access-requests command = %+v, want bounded implemented direct_read", command)
	} else if len(command.Flags) != 1 || command.Flags[0].MapsTo != "query.target" || !command.Flags[0].Required || command.Flags[0].Type != "string" {
		t.Fatalf("access-requests typed query contract = %+v, want only required query.target; raw provider paging stays behind --page/--page-cursor", command.Flags)
	} else if command.Notes != "" {
		t.Fatalf("promoted source-bound command note = %q, want historical blocker cleared", command.Notes)
	}
	agent := commands["agents get-agent"]
	if agent.Intent != "direct_read" || agent.Availability != "implemented" || agent.SourceOperation != "asana.rest.getAgent" {
		t.Fatalf("agent command = %+v, want bounded implemented direct_read", agent)
	}
	if len(agent.Flags) != 1 || agent.Flags[0].MapsTo != "path.agent_gid" || !agent.Flags[0].Required {
		t.Fatalf("agent path contract = %+v, want required path.agent_gid", agent.Flags)
	}
	if command := commands["workspaces get-workspaces"]; command.Intent != "direct_read" || command.Availability != "implemented" || command.Stream != "workspaces" || command.Operation != "" || command.SourceOperation != "asana.rest.getWorkspaces" {
		t.Fatalf("workspace command = %+v, want preserved source-backed direct read with ETL stream binding", command)
	}
	var surface struct {
		Endpoints []struct {
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"))), &surface); err != nil {
		t.Fatalf("decode projected API surface: %v", err)
	}
	foundWorkspaceEndpoint := false
	for _, endpoint := range surface.Endpoints {
		if endpoint.Path != "/workspaces" {
			continue
		}
		foundWorkspaceEndpoint = true
		if want := map[string]any{"stream": "workspaces", "direct_read": "workspaces get-workspaces"}; !reflect.DeepEqual(endpoint.CoveredBy, want) {
			t.Fatalf("workspace endpoint coverage = %#v, want both stream and direct-read lanes", endpoint.CoveredBy)
		}
		break
	}
	if !foundWorkspaceEndpoint {
		t.Fatal("workspace endpoint is absent from projected API surface")
	}
	if command := commands["pending get-pending"]; command.Intent != "direct_read" || command.Availability != "implemented" || command.Operation != "get_pending" || command.SourceOperation != "asana.rest.getPending" || command.Summary != "Get pending." {
		t.Fatalf("complete planned operation = %+v, want materialized bounded direct read", command)
	}
}

func TestSourceProjectionMergeDirectReadSurfaceCoveragePreservesOnlyValidStream(t *testing.T) {
	var streams orderedJSON
	if err := json.Unmarshal([]byte(`{"streams":[{"name":"widgets","path":"/widgets"}]}`), &streams); err != nil {
		t.Fatalf("decode streams fixture: %v", err)
	}
	source := sourceOperationDescriptor{Method: "GET", Path: "/widgets"}
	for _, tc := range []struct {
		name        string
		current     string
		want        string
		wantChanged bool
	}{
		{
			name:        "valid stream becomes dual lane coverage",
			current:     `{"stream":"widgets"}`,
			want:        `{"stream":"widgets","direct_read":"widgets list"}`,
			wantChanged: true,
		},
		{
			name:        "already dual lane coverage is stable",
			current:     `{"stream":"widgets","direct_read":"widgets list"}`,
			want:        `{"stream":"widgets","direct_read":"widgets list"}`,
			wantChanged: false,
		},
		{
			name:        "missing stream coverage restores exact declared stream",
			current:     `{"direct_read":"stale widgets command"}`,
			want:        `{"stream":"widgets","direct_read":"widgets list"}`,
			wantChanged: true,
		},
		{
			name:        "malformed stream is replaced with exact declared stream",
			current:     "{\"stream\":\"widgets\\n\",\"direct_read\":\"stale widgets command\"}",
			want:        `{"stream":"widgets","direct_read":"widgets list"}`,
			wantChanged: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var current, want orderedJSON
			if err := json.Unmarshal([]byte(tc.current), &current); err != nil {
				t.Fatalf("decode current coverage: %v", err)
			}
			if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatalf("decode wanted coverage: %v", err)
			}
			got, changed := sourceProjectionMergeDirectReadSurfaceCoverage(current.root, streams.root, source, []string{"widgets list"})
			if changed != tc.wantChanged {
				t.Fatalf("coverage changed = %t, want %t", changed, tc.wantChanged)
			}
			if !orderedSemanticEqual(got, want.root) {
				t.Fatalf("coverage = %#v, want %s", got, tc.want)
			}
		})
	}
}

func TestSourceProjectionExistingGeneratedReadRefreshesComplementaryStreamCoverage(t *testing.T) {
	var surface, streams, wantCoverage, wantUnrelated orderedJSON
	if err := json.Unmarshal([]byte(`{
  "endpoints":[
    {"method":"GET","path":"/widgets","covered_by":{"direct_read":"widgets list"}},
    {"method":"GET","path":"/unrelated","covered_by":{"direct_read":"unrelated list"}}
  ]
}`), &surface); err != nil {
		t.Fatalf("decode API surface fixture: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"streams":[{"name":"widgets","method":"GET","path":"/widgets"}]}`), &streams); err != nil {
		t.Fatalf("decode streams fixture: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"stream":"widgets","direct_read":"widgets list"}`), &wantCoverage); err != nil {
		t.Fatalf("decode expected coverage: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"method":"GET","path":"/unrelated","covered_by":{"direct_read":"unrelated list"}}`), &wantUnrelated); err != nil {
		t.Fatalf("decode expected unrelated endpoint: %v", err)
	}

	source := sourceOperationDescriptor{Method: "GET", Path: "/widgets"}
	if changed := sourceProjectionSetEndpointCoverage(surface.root, source, "direct_read", "widgets list", streams.root); !changed {
		t.Fatal("existing generated read coverage was not refreshed")
	}
	endpoints := arrayField(surface.root, "endpoints")
	if len(endpoints) != 2 {
		t.Fatalf("endpoint count = %d, want 2", len(endpoints))
	}
	widgets, ok := endpoints[0].(*orderedObject)
	if !ok {
		t.Fatal("widgets endpoint is not an object")
	}
	gotCoverage, _ := widgets.get("covered_by")
	if !orderedSemanticEqual(gotCoverage, wantCoverage.root) {
		t.Fatalf("widgets coverage = %#v, want stream plus generated direct read", gotCoverage)
	}
	if !orderedSemanticEqual(endpoints[1], wantUnrelated.root) {
		t.Fatalf("unrelated endpoint changed = %#v, want %#v", endpoints[1], wantUnrelated.root)
	}
}

// TestRetainedAsanaSourceImportRejectsReadProjectionDrift reads the actual
// connector-owned Asana lock and retained OpenAPI capture.  It does not create
// a descriptor in Go: source-import must reconstruct the descriptor from the
// pinned bytes and reject every altered bundle contract in check mode.
func TestRetainedAsanaSourceImportRejectsReadProjectionDrift(t *testing.T) {
	defsDir, bundleDir := copyInstalledAsanaProjectionBundle(t)
	writesPath := filepath.Join(bundleDir, "writes.json")
	writesBefore, err := os.ReadFile(writesPath)
	if err != nil {
		t.Fatal(err)
	}
	reverseETLBefore, err := asanaReverseETLCommandArtifacts(filepath.Join(bundleDir, "cli_surface.json"))
	if err != nil {
		t.Fatal(err)
	}
	if stdout, stderr, exit := runAsanaSourceImport(t, defsDir); exit != 0 {
		t.Fatalf("generate retained Asana source descriptor exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
	if writesAfter, err := os.ReadFile(writesPath); err != nil || !bytes.Equal(writesBefore, writesAfter) {
		t.Fatalf("read-only retained source import rewrote writes.json: read error=%v", err)
	}
	if reverseETLAfter, err := asanaReverseETLCommandArtifacts(filepath.Join(bundleDir, "cli_surface.json")); err != nil || !bytes.Equal(reverseETLBefore, reverseETLAfter) {
		t.Fatalf("read-only retained source import rewrote reverse-ETL or delete commands: read error=%v", err)
	}
	if stdout, stderr, exit := runAsanaSourceImport(t, defsDir, "--read-projection-only", "--check"); exit != 0 {
		t.Fatalf("check retained Asana source projection exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
	assertAsanaRetainedReadSourceContracts(t, filepath.Join(bundleDir, "sources", "asana-operation-descriptor.json"))

	for _, change := range []struct {
		name             string
		wantInputClosure bool
		mutate           func(t *testing.T, bundleDir string)
	}{
		{
			name: "invented source identity",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_custom_fields_for_workspace", func(operation map[string]any) {
					operation["source_operation"].(map[string]any)["id"] = "asana.rest.inventedAccessRequests"
				})
			},
		},
		{
			name: "source method substitution",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_custom_fields_for_workspace", func(operation map[string]any) {
					operation["source_operation"].(map[string]any)["method"] = "POST"
				})
			},
		},
		{
			name: "source route substitution",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_custom_fields_for_workspace", func(operation map[string]any) {
					operation["source_operation"].(map[string]any)["path"] = "/workspaces/{workspace_gid}/custom_fields/replaced"
				})
			},
		},
		{
			name: "typed workspace path contract",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaJSON(t, filepath.Join(bundleDir, "streams.json"), func(document map[string]any) {
					for _, raw := range document["streams"].([]any) {
						stream := raw.(map[string]any)
						if stream["name"] == "custom_fields" {
							stream["path"] = "/workspaces/{{ config.workspace_id }}/custom_fields/replaced"
							return
						}
					}
					t.Fatal("custom_fields stream is missing")
				})
			},
		},
		{
			name: "workspace pagination semantics",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaJSON(t, filepath.Join(bundleDir, "streams.json"), func(document map[string]any) {
					document["base"].(map[string]any)["pagination"].(map[string]any)["next_url_path"] = "next_page.token"
				})
			},
		},
		{
			name:             "operation-only undeclared query input",
			wantInputClosure: true,
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_access_requests", func(operation map[string]any) {
					rest := operation["rest"].(map[string]any)
					rest["parameters"] = append(rest["parameters"].([]any), map[string]any{"name": "rogue", "in": "query", "type": "string"})
				})
			},
		},
		{
			name:             "operation and CLI undeclared query input",
			wantInputClosure: true,
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_access_requests", func(operation map[string]any) {
					rest := operation["rest"].(map[string]any)
					rest["parameters"] = append(rest["parameters"].([]any), map[string]any{"name": "rogue", "in": "query", "type": "string"})
				})
				mutateAsanaCommand(t, filepath.Join(bundleDir, "cli_surface.json"), "access-requests get-access-requests", func(command map[string]any) {
					command["flags"] = append(command["flags"].([]any), map[string]any{"name": "rogue", "type": "string", "maps_to": "query.rogue"})
				})
			},
		},
		{
			name:             "operation-only undeclared header input",
			wantInputClosure: true,
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_access_requests", func(operation map[string]any) {
					rest := operation["rest"].(map[string]any)
					rest["parameters"] = append(rest["parameters"].([]any), map[string]any{"name": "X-Rogue", "in": "header", "type": "string", "schema": map[string]any{"type": "string", "maxLength": 32}, "max_bytes": 32})
				})
			},
		},
		{
			name:             "operation-only undeclared body input",
			wantInputClosure: true,
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaOperation(t, filepath.Join(bundleDir, "operations.json"), "get_access_requests", func(operation map[string]any) {
					operation["rest"].(map[string]any)["body"] = map[string]any{"rogue": "fixture"}
				})
			},
		},
		{
			name: "complete direct read marked planned",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaCommand(t, filepath.Join(bundleDir, "cli_surface.json"), "access-requests get-access-requests", func(command map[string]any) {
					command["availability"] = "planned"
				})
			},
		},
		{
			name: "complete fan-out ETL marked planned",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaCommand(t, filepath.Join(bundleDir, "cli_surface.json"), "sections list", func(command map[string]any) {
					command["availability"] = "planned"
				})
			},
		},
		{
			name: "response-only-gap direct read marked planned",
			mutate: func(t *testing.T, bundleDir string) {
				mutateAsanaCommand(t, filepath.Join(bundleDir, "cli_surface.json"), "memberships get-membership", func(command map[string]any) {
					command["availability"] = "planned"
				})
			},
		},
	} {
		change := change
		t.Run(change.name, func(t *testing.T) {
			defsDir, bundleDir := copyInstalledAsanaProjectionBundle(t)
			change.mutate(t, bundleDir)
			stdout, stderr, exit := runAsanaSourceImport(t, defsDir, "--read-projection-only", "--check")
			if exit != 1 || !strings.Contains(stderr, "derived bundle projection has drifted") {
				t.Fatalf("retained source check exit=%d stdout=%q stderr=%q, want projection drift refusal", exit, stdout, stderr)
			}
			if change.wantInputClosure {
				report, err := validateDir(os.DirFS(defsDir))
				if err != nil {
					t.Fatalf("validate retained altered Asana bundle: %v", err)
				}
				if !sourceProjectionFindingContains(report, "asana.rest.getAccessRequests", "request input absent from locked source contract") {
					t.Fatalf("retained altered Asana validation findings = %+v, want source-bound request-input closure refusal", report.Findings)
				}
			}
		})
	}
}

func sourceProjectionFindingContains(report Report, sourceOperation, message string) bool {
	for _, finding := range report.Findings {
		if finding.Rule == ruleSourceProjection && strings.Contains(finding.Message, sourceOperation) && strings.Contains(finding.Message, message) {
			return true
		}
	}
	return false
}

func TestSourceProjectionReadInputClosureAdmitsOnlyRetainedRequestClasses(t *testing.T) {
	source := sourceOperationDescriptor{
		SourceID: "asana.rest.getAccessRequests",
		Request: sourceRequestDescriptor{
			Header: []sourceParameterDescriptor{{Name: "X-Request-ID", Schema: map[string]any{"type": "string"}}},
			Body:   &sourceRequestBodyDescriptor{Schema: map[string]any{"type": "object", "properties": map[string]any{"filter": map[string]any{"type": "string"}}}},
		},
	}
	newBundle := func(parameters []engine.OperationParameter, body map[string]any, flags []engine.CLIFlag) engine.Bundle {
		return engine.Bundle{
			Operations: []engine.OperationSpec{{
				ID:              "get_access_requests",
				SourceOperation: &engine.SourceOperationBinding{ID: source.SourceID, Method: "GET", Path: "/access_requests"},
				REST:            &engine.RESTOperationSpec{Parameters: parameters, Body: body},
			}},
			CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{Path: "access-requests get-access-requests", SourceOperation: source.SourceID, Flags: flags}}},
		}
	}
	for _, testCase := range []struct {
		name   string
		bundle engine.Bundle
		want   string
	}{
		{
			name: "retained header and body fields are admitted",
			bundle: newBundle(
				[]engine.OperationParameter{{Name: "X-Request-ID", In: "header"}},
				map[string]any{"filter": "fixture"},
				[]engine.CLIFlag{{Name: "header-x-request-id", MapsTo: "header.X-Request-ID"}, {Name: "filter", MapsTo: "body.filter"}},
			),
		},
		{
			name:   "undeclared header is refused",
			bundle: newBundle([]engine.OperationParameter{{Name: "X-Rogue", In: "header"}}, nil, nil),
			want:   "parameter header.X-Rogue",
		},
		{
			name:   "undeclared body field is refused",
			bundle: newBundle(nil, map[string]any{"rogue": "fixture"}, nil),
			want:   "body.rogue",
		},
		{
			name:   "CLI cannot manufacture an undeclared header",
			bundle: newBundle(nil, nil, []engine.CLIFlag{{Name: "rogue", MapsTo: "header.X-Rogue"}}),
			want:   "flag --rogue mapping header.X-Rogue",
		},
		{
			name:   "CLI cannot manufacture an undeclared body field",
			bundle: newBundle(nil, nil, []engine.CLIFlag{{Name: "rogue", MapsTo: "body.rogue"}}),
			want:   "flag --rogue mapping body.rogue",
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if got := sourceProjectionReadInputClosure(testCase.bundle, source); !strings.Contains(got, testCase.want) {
				t.Fatalf("source-bound request-input closure = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams(t *testing.T) {
	defsDir, bundleDir := copyInstalledAsanaProjectionBundle(t)
	lock, err := loadConnectorSourceImportLock(defsDir, "asana")
	if err != nil {
		t.Fatal(err)
	}
	fetcher, err := newConnectorSourceImportRetainedArtifactFetcher(defsDir, "asana", defaultSourceImportLimits())
	if err != nil {
		t.Fatal(err)
	}
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		t.Fatal(err)
	}
	sourceProjectionNormalizeNonBlockingReadGaps(&result)
	selected, err := sourceProjectionReadOnlyResult(bundleDir, result)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projectSourceBoundReadDescriptorToBundle(bundleDir, selected, false); err != nil {
		t.Fatalf("materialize retained source-backed reads before idempotence check: %v", err)
	}
	var operationsDocument, cliDocument, streamsDocument orderedJSON
	for _, document := range []struct {
		path string
		into *orderedJSON
	}{
		{path: filepath.Join(bundleDir, "operations.json"), into: &operationsDocument},
		{path: filepath.Join(bundleDir, "cli_surface.json"), into: &cliDocument},
		{path: filepath.Join(bundleDir, "streams.json"), into: &streamsDocument},
	} {
		raw, err := os.ReadFile(document.path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, document.into); err != nil {
			t.Fatal(err)
		}
	}
	if operation := sourceProjectionStreamOperationForEndpoint(operationsDocument.root, cliDocument.root, streamsDocument.root, "GET", "/projects/{project_gid}/sections"); operation == nil || stringField(operation, "id") != "get_sections_for_project" {
		t.Fatalf("retained fan-out stream selection = %+v, want get_sections_for_project", operation)
	}
	var sectionsSource sourceOperationDescriptor
	for _, operation := range result.Operations {
		if operation.SourceID == "asana.rest.getSectionsForProject" {
			sectionsSource = operation
			break
		}
	}
	if sectionsSource.SourceID == "" {
		t.Fatal("retained source import omitted getSectionsForProject")
	}
	if len(sectionsSource.Runtime.Gaps) != 0 {
		t.Fatalf("retained fan-out source contract has unexpected gap: %+v", sectionsSource.Runtime.Gaps)
	}
	if operation := sourceProjectionStreamOperationForEndpoint(operationsDocument.root, cliDocument.root, streamsDocument.root, sectionsSource.Method, sectionsSource.Path); operation == nil || stringField(operation, "id") != "get_sections_for_project" {
		t.Fatalf("retained imported fan-out source = %+v selects operation %+v, want get_sections_for_project", sectionsSource, operation)
	}
	sectionsOperation := sourceProjectionStreamOperationForEndpoint(operationsDocument.root, cliDocument.root, streamsDocument.root, sectionsSource.Method, sectionsSource.Path)
	if !sourceProjectionStreamMatchesReadOperation(sectionsOperation, streamsDocument.root, "sections", sectionsSource) {
		t.Fatalf("retained fan-out stream does not prove source records/pagination semantics: source=%+v", sectionsSource)
	}
	found := map[string]bool{}
	for _, operation := range selected.Operations {
		found[operation.SourceID] = true
	}
	for _, sourceID := range []string{
		"asana.rest.getProjectStatusesForProject",
		"asana.rest.getSectionsForProject",
		"asana.rest.getStoriesForTask",
	} {
		if !found[sourceID] {
			t.Fatalf("source-backed fan-out ETL operation %q was omitted from projection: %+v", sourceID, found)
		}
	}
	if count := sourceProjectionSourceEndpointCount(selected.Operations, sectionsSource.Method, sectionsSource.Path); count != 1 {
		t.Fatalf("retained fan-out endpoint source count = %d, want 1", count)
	}
	bindingRaw, bound := sectionsOperation.get("source_operation")
	binding, bindingOK := bindingRaw.(*orderedObject)
	if !bound || !bindingOK || stringField(binding, "id") != sectionsSource.SourceID || stringField(binding, "method") != "GET" || stringField(binding, "path") != sectionsSource.Path {
		t.Fatalf("retained fan-out source binding = %+v, want %s GET %s", bindingRaw, sectionsSource.SourceID, sectionsSource.Path)
	}
	directStats, err := sourceProjectionMaterializeReadOperation(operationsDocument.root, cliDocument.root, streamsDocument.root, sectionsSource, selected.Operations)
	if err != nil {
		t.Fatal(err)
	}
	if directStats.Operations != 0 || directStats.Streams != 0 || directStats.CLI != 0 {
		bound, hasBinding := sectionsOperation.get("source_operation")
		t.Fatalf("retained fan-out direct materializer = %+v, binding=%+v present=%t, want idempotent exact source binding", directStats, bound, hasBinding)
	}
	fanOut := sourceImportResult{}
	for _, operation := range selected.Operations {
		if operation.SourceID == "asana.rest.getProjectStatusesForProject" || operation.SourceID == "asana.rest.getSectionsForProject" || operation.SourceID == "asana.rest.getStoriesForTask" {
			fanOut.Operations = append(fanOut.Operations, operation)
		}
	}
	projection, err := projectSourceBoundReadDescriptorToBundle(bundleDir, fanOut, false)
	if err != nil {
		t.Fatal(err)
	}
	if projection.Operations != 0 || projection.CLI != 0 {
		t.Fatalf("fan-out ETL projection = %+v, want idempotent canonical bindings", projection)
	}
}

func TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation(t *testing.T) {
	defsDir, bundleDir := copyInstalledAsanaProjectionBundle(t)
	absent, err := sourceProjectionReadNonExecutableMutationDispositions(bundleDir)
	if err != nil {
		t.Fatalf("read Asana absent-action dispositions: %v", err)
	}
	if len(absent) != 0 {
		t.Fatalf("Asana absent-action dispositions = %d, want none: source-complete no-body mutations use the existing reverse-ETL/delete action lane", len(absent))
	}
	partial, err := sourceProjectionReadPartialMutationCoverageDispositions(bundleDir)
	if err != nil {
		t.Fatalf("read Asana partial-coverage dispositions: %v", err)
	}
	if len(partial) != 69 {
		t.Fatalf("Asana partial-coverage dispositions = %d, want 65 typed-contract plus 4 path-alias mappings", len(partial))
	}
	foundations := map[string]int{}
	for _, disposition := range partial {
		foundations[disposition.Foundation]++
	}
	if foundations["cli-request-schema-foundation-r1"] != 65 || foundations["source-path-parameter-alias-foundation-r1"] != 4 || len(foundations) != 2 {
		t.Fatalf("Asana partial-coverage foundations = %+v, want 65 typed-contract and 4 legacy path-alias mappings", foundations)
	}
	if stdout, stderr, exit := runAsanaSourceImport(t, defsDir); exit != 0 {
		t.Fatalf("generate retained Asana mutation dispositions exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}

	descriptorRaw, err := os.ReadFile(filepath.Join(bundleDir, "sources", "asana-operation-descriptor.json"))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
		t.Fatalf("decode retained Asana descriptor: %v", err)
	}
	eligible := 0
	for _, operation := range descriptor.Operations {
		if sourceProjectionNoBodyMutationActionEligible(operation) {
			eligible++
		}
	}
	if eligible != 21 {
		t.Fatalf("source-complete no-body mutation candidates = %d, want 21", eligible)
	}
	surface, err := sourceProjectionExecutionSurface(bundleDir, "asana")
	if err != nil {
		t.Fatalf("load Asana execution surface: %v", err)
	}
	if findings := validateSourceExecutableCoverage(surface, "sources/asana-operation-descriptor.json", descriptor); len(findings) != 0 {
		t.Fatalf("retained Asana source dispositions leave executable-coverage findings: %+v", findings)
	}
	endpoints := map[string]engine.SurfaceEndpoint{}
	for _, endpoint := range surface.Surface.Endpoints {
		endpoints[sourceProjectionEndpointKey(endpoint.Method, endpoint.Path)] = endpoint
	}
	actions := map[string]engine.WriteAction{}
	for _, action := range surface.Writes {
		actions[action.Name] = action
	}
	wantPromoted := map[string]bool{
		"asana.rest.deleteAllocation": true, "asana.rest.deleteAttachment": true, "asana.rest.deleteBudget": true, "asana.rest.deleteCustomField": true, "asana.rest.deleteGoal": true, "asana.rest.deleteMembership": true, "asana.rest.deleteOooEntry": true, "asana.rest.deletePortfolio": true, "asana.rest.deleteProjectBrief": true, "asana.rest.deleteProjectStatus": true, "asana.rest.deleteProjectTemplate": true, "asana.rest.deleteRate": true, "asana.rest.deleteRole": true, "asana.rest.deleteStatus": true, "asana.rest.deleteStory": true, "asana.rest.deleteTaskTemplate": true, "asana.rest.deleteTimeTrackingCategory": true, "asana.rest.deleteTimeTrackingEntry": true, "asana.rest.deleteWebhook": true, "asana.rest.approveAccessRequest": true, "asana.rest.rejectAccessRequest": true,
	}
	deletePromotions, postPromotions := 0, 0
	for _, operation := range descriptor.Operations {
		if !wantPromoted[operation.SourceID] {
			continue
		}
		if !sourceProjectionMutationActionIsComplete(surface, operation) {
			t.Fatalf("source-complete mutation %q did not reach the reverse-ETL/delete action lane", operation.SourceID)
		}
		endpoint, found := endpoints[sourceProjectionEndpointKey(operation.Method, operation.Path)]
		if !found || endpoint.CoveredBy == nil || endpoint.CoveredBy.Write == "" {
			t.Fatalf("source-complete mutation %q is not bound to an api_surface write endpoint", operation.SourceID)
		}
		action := actions[endpoint.CoveredBy.Write]
		if strings.EqualFold(operation.Method, "DELETE") {
			deletePromotions++
			if action.Kind != "delete" || action.Confirm != "destructive" {
				t.Fatalf("source DELETE %q uses action %+v, want destructive delete action", operation.SourceID, action)
			}
		} else if strings.EqualFold(operation.Method, "POST") {
			postPromotions++
			if action.Kind != "update" {
				t.Fatalf("source no-body POST %q uses action %+v, want update action", operation.SourceID, action)
			}
		} else {
			t.Fatalf("source-complete mutation %q has method %s, want DELETE or POST", operation.SourceID, operation.Method)
		}
		delete(wantPromoted, operation.SourceID)
	}
	if len(wantPromoted) != 0 {
		t.Fatalf("source-complete mutation promotions missing: %v", maps.Keys(wantPromoted))
	}
	if deletePromotions != 19 || postPromotions != 2 {
		t.Fatalf("source-complete action lanes = DELETE %d, POST %d; want DELETE 19, POST 2", deletePromotions, postPromotions)
	}
	if surface.CLISurface == nil {
		t.Fatal("Asana execution surface has no CLI declaration")
	}
	for _, command := range surface.CLISurface.Commands {
		if strings.HasPrefix(command.Summary, "Planned fixed-target Asana ") {
			t.Fatalf("source-projected Asana command %q retains a historical planned label", command.Path)
		}
	}
}

func TestRetainedAsanaMultipartActionsCoverLockedAttachmentOperation(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	bundleDir := filepath.Join(root, "internal", "connectors", "defs", "asana")
	bundle, err := sourceProjectionExecutionSurface(bundleDir, "asana")
	if err != nil {
		t.Fatalf("load Asana execution surface: %v", err)
	}
	descriptorRaw, err := os.ReadFile(filepath.Join(bundleDir, "sources", "asana-operation-descriptor.json"))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
		t.Fatalf("decode Asana descriptor: %v", err)
	}
	var source sourceOperationDescriptor
	for _, operation := range descriptor.Operations {
		if operation.SourceID == "asana.rest.createAttachmentForObject" {
			source = operation
			break
		}
	}
	if source.SourceID == "" {
		t.Fatal("retained Asana attachment source operation is absent")
	}
	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}
	commands := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		if command.Write != "" && command.Availability == "implemented" {
			commands[command.Write] = command
		}
	}

	for _, name := range []string{"upload_attachment_file", "create_external_attachment"} {
		action, actionFound := actions[name]
		command, commandFound := commands[name]
		if !actionFound || !commandFound {
			t.Fatalf("retained Asana attachment action/command %q found=%t/%t", name, actionFound, commandFound)
		}
		if !sourceActionCoversOperation(action, command, source) {
			t.Fatalf("retained Asana multipart action %q does not cover its locked provider operation", name)
		}
	}
	if !sourceProjectionMutationActionIsComplete(bundle, source) {
		t.Fatal("retained Asana attachment source operation has no complete declared multipart action")
	}

	fileAction := actions["upload_attachment_file"]
	fileCommand := commands["upload_attachment_file"]
	for _, tc := range []struct {
		name    string
		action  engine.WriteAction
		command engine.CLICommand
		source  sourceOperationDescriptor
	}{
		{
			name:    "non-multipart provider media is not accepted",
			action:  fileAction,
			command: fileCommand,
			source: func() sourceOperationDescriptor {
				wrong := source
				wrong.Request.MediaType = "application/json"
				return wrong
			}(),
		},
		{
			name: "missing provider-required part is not accepted",
			action: func() engine.WriteAction {
				wrong := fileAction
				multipart := *fileAction.Multipart
				multipart.Parts = append([]engine.MultipartPartSpec(nil), fileAction.Multipart.Parts[1:]...)
				wrong.Multipart = &multipart
				return wrong
			}(),
			command: fileCommand,
			source:  source,
		},
		{
			name: "provider-required part must be required in the action and record",
			action: func() engine.WriteAction {
				wrong := fileAction
				multipart := *fileAction.Multipart
				multipart.Parts = append([]engine.MultipartPartSpec(nil), fileAction.Multipart.Parts...)
				multipart.Parts[0].Required = false
				wrong.Multipart = &multipart
				return wrong
			}(),
			command: fileCommand,
			source:  source,
		},
		{
			name: "multipart part must bind a closed record property",
			action: func() engine.WriteAction {
				wrong := fileAction
				wrong.RecordSchema = bytes.Replace(wrong.RecordSchema, []byte(`"file_path"`), []byte(`"unmapped_file_path"`), 1)
				return wrong
			}(),
			command: fileCommand,
			source:  source,
		},
		{
			name: "unmapped multipart part is not accepted",
			action: func() engine.WriteAction {
				wrong := fileAction
				multipart := *fileAction.Multipart
				multipart.Parts = append([]engine.MultipartPartSpec(nil), fileAction.Multipart.Parts...)
				multipart.Parts[0].Name = "not_a_provider_part"
				wrong.Multipart = &multipart
				return wrong
			}(),
			command: fileCommand,
			source:  source,
		},
		{
			name:   "missing required multipart flag is not accepted",
			action: fileAction,
			command: func() engine.CLICommand {
				wrong := fileCommand
				wrong.Flags = append([]engine.CLIFlag(nil), fileCommand.Flags[1:]...)
				return wrong
			}(),
			source: source,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if sourceActionCoversOperation(tc.action, tc.command, tc.source) {
				t.Fatal("multipart coverage accepted a malformed provider/action binding")
			}
		})
	}
}

func copyInstalledAsanaProjectionBundle(t *testing.T) (string, string) {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	defsDir := filepath.Join(t.TempDir(), "defs")
	bundleDir := filepath.Join(defsDir, "asana")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(bundleDir, os.DirFS(filepath.Join(root, "internal", "connectors", "defs", "asana"))); err != nil {
		t.Fatalf("copy installed Asana bundle: %v", err)
	}
	return defsDir, bundleDir
}

func runAsanaSourceImport(t *testing.T, defsDir string, extraArgs ...string) (string, string, int) {
	t.Helper()
	args := append([]string{"source-import", "asana", "--defs", defsDir}, extraArgs...)
	var stdout, stderr strings.Builder
	exit := runSourceImport(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), exit
}

func asanaReverseETLCommandArtifacts(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var surface struct {
		Commands []json.RawMessage `json:"commands"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		return nil, err
	}
	artifacts := make([][]byte, 0, len(surface.Commands))
	for _, rawCommand := range surface.Commands {
		var command struct {
			Intent string `json:"intent"`
		}
		if err := json.Unmarshal(rawCommand, &command); err != nil {
			return nil, err
		}
		if command.Intent == "reverse_etl" {
			artifacts = append(artifacts, bytes.TrimSpace(rawCommand))
		}
	}
	return bytes.Join(artifacts, []byte{'\n'}), nil
}

func assertAsanaRetainedReadSourceContracts(t *testing.T, descriptorPath string) {
	t.Helper()
	raw, err := os.ReadFile(descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(raw, &descriptor); err != nil {
		t.Fatalf("decode retained Asana descriptor: %v", err)
	}
	byID := map[string]sourceOperationDescriptor{}
	for _, operation := range descriptor.Operations {
		byID[operation.SourceID] = operation
	}
	for sourceID, want := range map[string]struct {
		endpoint string
		location string
	}{
		"asana.rest.getCustomFieldsForWorkspace":         {endpoint: "GET /workspaces/{workspace_gid}/custom_fields", location: `paths["/workspaces/{workspace_gid}/custom_fields"].get`},
		"asana.rest.getProjectStatusesForProject":        {endpoint: "GET /projects/{project_gid}/project_statuses", location: `paths["/projects/{project_gid}/project_statuses"].get`},
		"asana.rest.getProjects":                         {endpoint: "GET /projects", location: `paths["/projects"].get`},
		"asana.rest.getSectionsForProject":               {endpoint: "GET /projects/{project_gid}/sections", location: `paths["/projects/{project_gid}/sections"].get`},
		"asana.rest.getStoriesForTask":                   {endpoint: "GET /tasks/{task_gid}/stories", location: `paths["/tasks/{task_gid}/stories"].get`},
		"asana.rest.getTags":                             {endpoint: "GET /tags", location: `paths["/tags"].get`},
		"asana.rest.getTasks":                            {endpoint: "GET /tasks", location: `paths["/tasks"].get`},
		"asana.rest.getTeamMemberships":                  {endpoint: "GET /team_memberships", location: `paths["/team_memberships"].get`},
		"asana.rest.getTeamsForWorkspace":                {endpoint: "GET /workspaces/{workspace_gid}/teams", location: `paths["/workspaces/{workspace_gid}/teams"].get`},
		"asana.rest.getUsers":                            {endpoint: "GET /users", location: `paths["/users"].get`},
		"asana.rest.getWorkspaceMembershipsForWorkspace": {endpoint: "GET /workspaces/{workspace_gid}/workspace_memberships", location: `paths["/workspaces/{workspace_gid}/workspace_memberships"].get`},
		"asana.rest.getWorkspaces":                       {endpoint: "GET /workspaces", location: `paths["/workspaces"].get`},
	} {
		operation, found := byID[sourceID]
		if !found || strings.ToUpper(operation.Method)+" "+operation.Path != want.endpoint || operation.Source.Location != want.location {
			t.Fatalf("retained source operation %s = %+v, want endpoint %s at %s", sourceID, operation, want.endpoint, want.location)
		}
	}
	customFields := byID["asana.rest.getCustomFieldsForWorkspace"]
	if !sourceProjectionHasRequiredStringParameter(customFields.Request.Path, "workspace_gid") {
		t.Fatalf("retained custom-fields source contract = %+v, want required string workspace_gid path input", customFields.Request.Path)
	}
	workspaces := byID["asana.rest.getWorkspaces"]
	wantPagination := map[string]any{
		"type":          "next_url",
		"next_url_path": "next_page.uri",
		"size_param":    "limit",
		"limit_param":   "limit",
		"offset_param":  "offset",
		"page_size":     json.Number("100"),
	}
	if !reflect.DeepEqual(workspaces.Pagination, wantPagination) {
		t.Fatalf("retained workspace pagination = %+v, want closed limit/offset next_url contract %+v", workspaces.Pagination, wantPagination)
	}
}

func sourceProjectionHasRequiredStringParameter(parameters []sourceParameterDescriptor, name string) bool {
	for _, parameter := range parameters {
		schema, _ := parameter.Schema.(map[string]any)
		if parameter.Name == name && parameter.Required && schema["type"] == "string" {
			return true
		}
	}
	return false
}

func mutateAsanaOperation(t *testing.T, path, id string, mutate func(map[string]any)) {
	t.Helper()
	mutateAsanaJSON(t, path, func(document map[string]any) {
		for _, raw := range document["operations"].([]any) {
			operation := raw.(map[string]any)
			if operation["id"] == id {
				mutate(operation)
				return
			}
		}
		t.Fatalf("installed Asana operation %q is missing", id)
	})
}

func mutateAsanaCommand(t *testing.T, path, commandPath string, mutate func(map[string]any)) {
	t.Helper()
	mutateAsanaJSON(t, path, func(document map[string]any) {
		for _, raw := range document["commands"].([]any) {
			command := raw.(map[string]any)
			if command["path"] == commandPath {
				mutate(command)
				return
			}
		}
		t.Fatalf("installed Asana command %q is missing", commandPath)
	})
}

func mutateAsanaJSON(t *testing.T, path string, mutate func(map[string]any)) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode %s: %v", filepath.Base(path), err)
	}
	mutate(document)
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", filepath.Base(path), err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", filepath.Base(path), err)
	}
}

func TestSourceProjectionImportsRequiredSourceBoundReadParameters(t *testing.T) {
	bundleDir := t.TempDir()
	operationsPath := filepath.Join(bundleDir, "operations.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	writeProjectionFixture(t, operationsPath, `{
  "schema_version": 1,
  "operations": [
    {"id":"get_agent","kind":"rest_read","summary":"Get an agent","risk":"none","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/agents/{agent_gid}","max_bytes":1024,"parameters":[]}}
  ]
}`)
	writeProjectionFixture(t, cliPath, `{
  "schema_version": 1,
  "commands": [
    {"path":"agents get-agent","summary":"Get an agent","intent":"direct_read","availability":"implemented","operation":"get_agent","api_surface":[{"method":"GET","path":"/agents/{agent_gid}"}]}
  ]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{"api":"asana","endpoints":[{"method":"GET","path":"/agents/{agent_gid}"}]}`)

	result := sourceImportResult{Operations: []sourceOperationDescriptor{{
		Connector: "asana", SourceID: "asana.rest.getAgent", Method: "GET", Path: "/agents/{agent_gid}",
		Request: sourceRequestDescriptor{Path: []sourceParameterDescriptor{{Name: "agent_gid", Required: true, Schema: map[string]any{"type": "string"}}}},
		Output:  sourceOutputDescriptor{Class: sourceOutputJSON},
	}}}
	if _, err := projectSourceBoundReadDescriptorToBundle(bundleDir, result, false); err != nil {
		t.Fatalf("project incomplete source-bound read: %v", err)
	}

	var cli engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, cliPath)), &cli); err != nil {
		t.Fatalf("decode projected CLI surface: %v", err)
	}
	command := cli.Commands[0]
	if command.Availability != "implemented" || command.Intent != "direct_read" || command.SourceOperation != "asana.rest.getAgent" {
		t.Fatalf("source-backed command = %+v, want implemented direct read", command)
	}
	if len(command.Flags) != 1 || command.Flags[0].MapsTo != "path.agent_gid" || command.Flags[0].Type != "string" || !command.Flags[0].Required {
		t.Fatalf("source-backed path contract = %+v, want required path.agent_gid string", command.Flags)
	}
	if command.Notes != "" {
		t.Fatalf("source-backed command note = %q, want no missing-foundation note", command.Notes)
	}
}

func TestSourceProjectionStreamPathMatchesOnlyDeclaredWholeVariableSegments(t *testing.T) {
	for _, tc := range []struct {
		name       string
		streamPath string
		sourcePath string
		want       bool
	}{
		{name: "fixed config path", streamPath: "/workspaces/{{ config.workspace_id }}/teams", sourcePath: "/workspaces/{workspace_gid}/teams", want: true},
		{name: "declared fan-out path", streamPath: "/projects/{{ fanout.id }}/sections", sourcePath: "/projects/{project_gid}/sections", want: true},
		{name: "fan-out cannot change literal route", streamPath: "/projects/{{ fanout.id }}/stories", sourcePath: "/projects/{project_gid}/sections", want: false},
		{name: "arbitrary interpolation is not source bound", streamPath: "/projects/{{ record.path }}/sections", sourcePath: "/projects/{project_gid}/sections", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sourceProjectionStreamPathMatchesSourcePath(tc.streamPath, tc.sourcePath); got != tc.want {
				t.Fatalf("sourceProjectionStreamPathMatchesSourcePath(%q, %q) = %t, want %t", tc.streamPath, tc.sourcePath, got, tc.want)
			}
		})
	}
}

func TestSourceProjectionMaterializesDeclaredFanOutStreamWithoutInventingDirectRead(t *testing.T) {
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), `{"schema_version":1,"actions":[]}`)
	operationsPath := filepath.Join(bundleDir, "operations.json")
	writeProjectionFixture(t, operationsPath, `{
  "schema_version":1,
  "operations":[{
    "id":"get_sections_for_project","kind":"stream_etl","summary":"Sections","risk":"none","approval":"none","output_policy":"json_redacted",
    "composite":{"steps":["stream:sections"]}
  }]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), `{"schema_version":1,"commands":[]}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "streams.json"), `{
  "base":{"pagination":{"type":"next_url","next_url_path":"next_page.uri"}},
  "streams":[{
    "name":"sections","path":"/projects/{{ fanout.id }}/sections","records":{"path":"data"},"schema":"schemas/sections.json",
    "fan_out":{"ids_from":{"request":{"path":"/projects","records_path":"data","id_field":"gid"}},"into":{"path_var":"project_gid"}}
  }]
}`)
	writeProjectionFixture(t, filepath.Join(bundleDir, "api_surface.json"), `{"api":"asana","endpoints":[]}`)

	source := sourceOperationDescriptor{
		Connector: "asana", SourceID: "asana.rest.getSectionsForProject", Method: "GET", Path: "/projects/{project_gid}/sections",
		Pagination: map[string]any{"type": "next_url", "next_url_path": "next_page.uri"}, Output: sourceOutputDescriptor{Class: sourceOutputJSON},
	}
	selected, err := sourceProjectionReadOnlyResult(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{source}})
	if err != nil {
		t.Fatalf("select fan-out ETL source operation: %v", err)
	}
	if len(selected.Operations) != 1 || selected.Operations[0].SourceID != source.SourceID {
		t.Fatalf("selected source operations = %+v, want the declared fan-out stream", selected.Operations)
	}
	var operationsDocument, cliDocument, streamsDocument orderedJSON
	for _, document := range []struct {
		path string
		into *orderedJSON
	}{
		{path: operationsPath, into: &operationsDocument},
		{path: filepath.Join(bundleDir, "cli_surface.json"), into: &cliDocument},
		{path: filepath.Join(bundleDir, "streams.json"), into: &streamsDocument},
	} {
		raw, err := os.ReadFile(document.path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, document.into); err != nil {
			t.Fatal(err)
		}
	}
	if operation := sourceProjectionStreamOperationForEndpoint(operationsDocument.root, cliDocument.root, streamsDocument.root, source.Method, source.Path); operation == nil || stringField(operation, "id") != "get_sections_for_project" {
		t.Fatalf("fan-out stream operation selection = %+v, want get_sections_for_project", operation)
	}
	directStats, err := sourceProjectionMaterializeReadOperation(operationsDocument.root, cliDocument.root, streamsDocument.root, source, selected.Operations)
	if err != nil {
		t.Fatalf("materialize fan-out ETL source operation directly: %v", err)
	}
	if directStats.Operations != 1 || directStats.CLI != 0 {
		t.Fatalf("direct fan-out ETL projection stats = %+v, want one operation and no invented CLI command", directStats)
	}
	stats, err := projectSourceBoundReadDescriptorToBundle(bundleDir, selected, false)
	if err != nil {
		t.Fatalf("materialize fan-out ETL source operation: %v", err)
	}
	if stats.Operations != 1 || stats.CLI != 0 {
		t.Fatalf("fan-out ETL projection stats = %+v, want one operation and no invented CLI command", stats)
	}
	var projected struct {
		Operations []struct {
			ID              string `json:"id"`
			SourceOperation *struct {
				ID     string `json:"id"`
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"source_operation"`
		} `json:"operations"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, operationsPath)), &projected); err != nil {
		t.Fatalf("decode projected operations: %v", err)
	}
	if got := projected.Operations[0].SourceOperation; got == nil || got.ID != source.SourceID || got.Method != "GET" || got.Path != source.Path {
		t.Fatalf("fan-out ETL source binding = %+v, want exact source operation", got)
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
	if strings.Contains(projectedSurface, `"covered_by"`) || !strings.Contains(projectedSurface, `"model": "direct_read"`) || !strings.Contains(projectedSurface, source.SourceID) || !strings.Contains(projectedSurface, sourceOperationExecutionFoundation) {
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

func TestSourceProjectionDoesNotBlockReadForUnusedOptionalNonScalarParameter(t *testing.T) {
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
			Reason:     "query parameter requires non-scalar serialization support",
		}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{source}}
	if blocked := sourceProjectionBlockedReadSources(result); len(blocked) != 0 {
		t.Fatalf("unused optional non-scalar query parameter blocked executable read: %+v", blocked)
	}
	if reachable := sourceProjectionReachableReadSources(result); reachable[source.SourceID].SourceID != source.SourceID {
		t.Fatalf("read with only an unused optional non-scalar parameter was not reachable: %+v", reachable)
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

func TestSourceProjectionRequestSchemaExcludesReadOnlyFieldsAcrossAllOf(t *testing.T) {
	bundleDir := t.TempDir()
	writesPath := filepath.Join(bundleDir, "writes.json")
	cliPath := filepath.Join(bundleDir, "cli_surface.json")
	writeProjectionFixture(t, writesPath, `{"schema_version":1,"actions":[{"name":"items_create","method":"POST","path":"/items","body_type":"json","body_fields":["data"],"record_schema":{"type":"object","additionalProperties":false,"required":["data"],"properties":{"data":{"type":"object","maxProperties":256,"additionalProperties":true}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, cliPath, `{"schema_version":1,"commands":[{"path":"items create","intent":"reverse_etl","availability":"implemented","write":"items_create","flags":[{"name":"data","type":"json","maps_to":"record.data","required":true,"max_bytes":1048576}]}]}`)
	operation := sourceOperationDescriptor{
		Connector: "alpha", SourceID: "items/create", Method: "post", Path: "/items",
		Request: sourceRequestDescriptor{MediaType: "application/json", Body: &sourceRequestBodyDescriptor{Schema: map[string]any{
			"type": "object", "required": []any{"data", "server_echo"},
			"properties": map[string]any{
				"server_echo": map[string]any{"type": "string", "readOnly": true},
				"data": map[string]any{"allOf": []any{
					map[string]any{
						"type": "object", "required": []any{"gid", "name"},
						"properties": map[string]any{
							"gid":  map[string]any{"type": "string", "readOnly": true},
							"name": map[string]any{"type": "string", "maxLength": json.Number("32")},
						},
					},
					map[string]any{
						"type": "object", "required": []any{"workspace", "modified_at"},
						"properties": map[string]any{
							"workspace":   map[string]any{"type": "string"},
							"modified_at": map[string]any{"type": "string", "readOnly": true},
						},
					},
				}},
			},
		}}},
		Runtime: sourceRuntimeReachability{Gaps: []sourceContractGap{{
			Foundation: "cli-request-schema-foundation-r1", Location: "request body", Reason: "source-owned allOf request shape",
		}}},
	}

	stats, err := projectSourceDescriptorToBundle(bundleDir, sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, false)
	if err != nil {
		t.Fatalf("project allOf request contract: %v", err)
	}
	if stats.Writes != 1 || stats.Missing != 0 {
		t.Fatalf("projection stats = %+v, want one direction-safe action", stats)
	}
	var projected struct {
		Actions []struct {
			RecordSchema json.RawMessage `json:"record_schema"`
		} `json:"actions"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, writesPath)), &projected); err != nil {
		t.Fatalf("decode projected writes: %v", err)
	}
	if len(projected.Actions) != 1 {
		t.Fatalf("projected actions = %d, want 1", len(projected.Actions))
	}
	var record map[string]any
	if err := json.Unmarshal(projected.Actions[0].RecordSchema, &record); err != nil {
		t.Fatalf("decode projected record schema: %v", err)
	}
	properties := record["properties"].(map[string]any)
	if _, exists := properties["server_echo"]; exists {
		t.Fatalf("top-level readOnly field remained executable: %#v", properties)
	}
	data := properties["data"].(map[string]any)
	if data["additionalProperties"] != false {
		t.Fatalf("allOf data projection stayed open: %#v", data)
	}
	dataProperties := data["properties"].(map[string]any)
	if len(dataProperties) != 2 || dataProperties["name"] == nil || dataProperties["workspace"] == nil {
		t.Fatalf("allOf writable properties = %#v, want only name/workspace", dataProperties)
	}
	if got := data["required"]; !reflect.DeepEqual(got, []any{"name", "workspace"}) {
		t.Fatalf("allOf required = %#v, want readOnly names removed", got)
	}
	schema, err := engine.CompileSchema(projected.Actions[0].RecordSchema)
	if err != nil {
		t.Fatalf("compile projected request schema: %v", err)
	}
	if err := schema.Validate(map[string]any{"data": map[string]any{"name": "n", "workspace": "w"}}); err != nil {
		t.Fatalf("writable request rejected: %v", err)
	}
	for _, record := range []map[string]any{
		{"server_echo": "forged", "data": map[string]any{"name": "n", "workspace": "w"}},
		{"data": map[string]any{"gid": "forged", "name": "n", "workspace": "w"}},
		{"data": map[string]any{"modified_at": "forged", "name": "n", "workspace": "w"}},
	} {
		if err := schema.Validate(record); err == nil {
			t.Fatalf("readOnly request unexpectedly validated: %#v", record)
		}
	}
	cli := readProjectionFixture(t, cliPath)
	if strings.Contains(cli, "server-echo") || strings.Contains(cli, "record.server_echo") {
		t.Fatalf("readOnly field leaked into command flags: %s", cli)
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
	writeProjectionFixture(t, filepath.Join(bundleDir, "certification.json"), `{"schema_version":"invalid"}`)
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
		Connector: "alpha", Protocol: "rest", SourceID: "items/create", Method: "post", Path: "/items/{owner}",
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

// sourceProjectionMinimalLegacyLock makes the synthetic surface-sync fixture
// exercise the real checked-in lock admission path. The descriptor remains
// deliberately richer than the lock so projection can still report its field
// drift; its source identity is nevertheless bound to the one declared lock
// operation.
func sourceProjectionMinimalLegacyLock(t *testing.T, operation *sourceOperationDescriptor) string {
	t.Helper()
	artifact := []byte(`{"openapi":"3.0.3"}`)
	lock := sourceImportFixtureLock("alpha", "https://fixtures.polymetrics.invalid/alpha-openapi.json", artifact)
	location := `paths["/items/{owner}"].post`
	lock.Rest.Operations = []sourceImportRESTOperation{{
		ID:             operation.SourceID,
		Protocol:       "rest",
		Method:         "POST",
		Path:           operation.Path,
		SourceLocation: location,
	}}
	lock.Counts = sourceImportCounts{REST: 1, Total: 1}
	operation.Source = sourceImportSource{URL: lock.Rest.SourceURL, Location: location}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal minimal legacy source lock: %v", err)
	}
	return string(raw)
}

// sourceProjectionMinimalV3Lock uses the ordinary v3 fixture constructor so
// surface-sync's schema-3 test has a fully valid checked-in lock, including
// document provenance and inventory counts, before it projects any fields.
func sourceProjectionMinimalV3Lock(t *testing.T, operation *sourceOperationDescriptor) string {
	t.Helper()
	raw := sourceImportV3FixtureLock(t, "alpha", []sourceImportV3FixtureDocument{{
		ID:          "items",
		Path:        operation.Path,
		Method:      "POST",
		OperationID: "items_create",
		Artifact:    []byte(`{"openapi":"3.0.3"}`),
	}})
	lock, err := parseSourceImportLock(raw, "alpha")
	if err != nil {
		t.Fatalf("parse minimal v3 source lock: %v", err)
	}
	document := lock.Rest.SourceDocuments[0]
	sourceOperation := document.Operations[0]
	operation.SourceID = sourceOperation.ID
	operation.ProviderOperationID = sourceOperation.OperationID
	operation.Method = sourceOperation.Method
	operation.Path = sourceOperation.Path
	operation.Source = sourceImportExpectedV3DescriptorProvenance(document, sourceOperation)
	return string(raw)
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
		{name: "source location", mutate: func(operation *sourceOperationDescriptor) { operation.Source.Location = `paths["/reassigned"].get` }},
		{name: "source method", mutate: func(operation *sourceOperationDescriptor) { operation.Method = "POST" }},
		{name: "source path", mutate: func(operation *sourceOperationDescriptor) { operation.Path = "/reassigned" }},
	} {
		change := change
		t.Run(change.name, func(t *testing.T) {
			operation := result.Operations[0]
			change.mutate(&operation)
			if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{operation}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "provider contract drift") {
				t.Fatalf("%s findings = %+v", change.name, findings)
			}
		})
	}

	for _, change := range []struct {
		name   string
		mutate func(*sourceOperationDescriptor)
	}{
		{name: "artifact digest", mutate: func(operation *sourceOperationDescriptor) { operation.Source.SHA256 = strings.Repeat("0", 64) }},
		{name: "artifact bytes", mutate: func(operation *sourceOperationDescriptor) { operation.Source.Bytes++ }},
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
			if findings := validateSourceDescriptorAgainstLock("fixture", "sources/fixture-operation-descriptor.json", lock, sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{operation}}); len(findings) != 0 {
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
	lockRaw := sourceProjectionMinimalV3Lock(t, &operation)
	descriptorRaw, err := json.Marshal(sourceImportDescriptorDocument{SchemaVersion: 3, Operations: []sourceOperationDescriptor{operation}})
	if err != nil {
		t.Fatal(err)
	}
	writeProjectionFixture(t, filepath.Join(bundleDir, "sources", "alpha-operation-source-lock.json"), lockRaw)
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

func TestSourceProjectionMutationDispositionCitationIgnoresRetainedArtifactMetadata(t *testing.T) {
	operation := sourceCitedMutationTestOperation("asana", "asana.rest.createAccessRequest", "POST", "/access_requests")
	operation.Source.SHA256 = ""
	operation.Source.Bytes = 0
	operation.Source.PublishedURL = "https://capture.polymetrics.invalid/provider-page"
	operation.Source.PublishedCaptureURL = "https://capture.polymetrics.invalid/capture"
	operation.Source.PublishedSHA256 = strings.Repeat("0", 64)
	operation.Source.PublishedBytes = 1
	operation.Source.PublishedAdapter = "fixture"
	operation.Source.ContentType = "application/json"
	citation := sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path}
	if err := sourceProjectionValidateMutationDispositionCitation(operation, citation, "mutation disposition"); err != nil {
		t.Fatalf("stable source identity/method/path citation rejected because capture metadata is absent: %v", err)
	}

	operation.Source.URL = ""
	if err := sourceProjectionValidateMutationDispositionCitation(operation, citation, "mutation disposition"); err == nil || !strings.Contains(err.Error(), "lacks a provider source citation") {
		t.Fatalf("missing provider URL error = %v, want stable-source citation refusal", err)
	}
}

func TestSourceProjectionWriteDisabledLockedSourcesRetainMutationArtifacts(t *testing.T) {
	pathFlags := func(path string) []engine.CLIFlag {
		flags := make([]engine.CLIFlag, 0)
		for _, match := range sourceProjectionPathVariableRE.FindAllStringSubmatch(path, -1) {
			flags = append(flags, engine.CLIFlag{
				Name:     strings.ReplaceAll(match[1], "_", "-"),
				Type:     "string",
				MapsTo:   "path." + match[1],
				Required: true,
			})
		}
		return flags
	}
	tests := []struct {
		name             string
		connector        string
		sourceURL        string
		sourceSHA        string
		sourceBytes      int64
		readID           string
		readPath         string
		readLocation     string
		mutationID       string
		mutationPath     string
		mutationLocation string
		method           string
	}{
		{
			name:             "sentry source lock",
			connector:        "sentry",
			sourceURL:        "https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json",
			sourceSHA:        "b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435",
			sourceBytes:      3868570,
			readID:           "listOrganizationProjects",
			readPath:         "/api/0/organizations/{organization_id_or_slug}/projects/",
			readLocation:     `paths["/api/0/organizations/{organization_id_or_slug}/projects/"].get`,
			mutationID:       "createOrganizationDashboard",
			mutationPath:     "/api/0/organizations/{organization_id_or_slug}/dashboards/",
			mutationLocation: `paths["/api/0/organizations/{organization_id_or_slug}/dashboards/"].post`,
			method:           "post",
		},
		{
			name:             "vercel source lock",
			connector:        "vercel",
			sourceURL:        "https://openapi.vercel.sh/",
			sourceSHA:        "74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28",
			sourceBytes:      10463249,
			readID:           "getProjects",
			readPath:         "/v10/projects",
			readLocation:     `paths["/v10/projects"].get`,
			mutationID:       "deleteStorageStoresBlobById",
			mutationPath:     "/storage/stores/blob/{id}",
			mutationLocation: `paths["/storage/stores/blob/{id}"].delete`,
			method:           "delete",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// These citations are copied from the preserved Sentry and Vercel
			// source-locked descriptors. Keep the acceptance vectors source-backed
			// without copying or rewriting either provider document into this branch.
			result := sourceImportResult{Operations: []sourceOperationDescriptor{
				{Connector: tc.connector, SourceID: tc.readID, Method: "get", Path: tc.readPath, Source: sourceImportSource{URL: tc.sourceURL, SHA256: tc.sourceSHA, Bytes: tc.sourceBytes, Location: tc.readLocation, Form: "openapi", Version: "3.0.3"}},
				{Connector: tc.connector, SourceID: tc.mutationID, Method: tc.method, Path: tc.mutationPath, Source: sourceImportSource{URL: tc.sourceURL, SHA256: tc.sourceSHA, Bytes: tc.sourceBytes, Location: tc.mutationLocation, Form: "openapi", Version: "3.0.3"}},
			}}

			bundle := engine.Bundle{
				Name: tc.connector,
				Metadata: engine.Metadata{
					Name:         tc.connector,
					Capabilities: engine.Capabilities{Read: true, Write: false, WriteDeclared: true},
				},
				CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
					Path:         "projects list",
					Intent:       "direct_read",
					Availability: "implemented",
					APISurface:   []engine.CLISurfaceEndpointRef{{Method: "GET", Path: tc.readPath}},
					Flags:        pathFlags(tc.readPath),
				}}},
			}
			if got := sourceProjectionApplyWriteDisabledMutationArtifacts(bundle, &result); got != 1 {
				t.Fatalf("automatic mutation artifact count = %d, want 1", got)
			}

			var read, mutation sourceOperationDescriptor
			for _, operation := range result.Operations {
				switch operation.SourceID {
				case tc.readID:
					read = operation
				case tc.mutationID:
					mutation = operation
				}
			}
			if read.SourceID == "" || read.Runtime.NonExecutableMutation != nil {
				t.Fatalf("read operation = %#v, want retained executable read without mutation artifact", read)
			}
			if !sourceProjectionHasNonExecutableMutationDisposition(mutation) {
				t.Fatalf("mutation operation = %#v, want cited non-executable artifact", mutation)
			}
			if mutation.Source.URL != tc.sourceURL || mutation.Source.SHA256 != tc.sourceSHA || mutation.Source.Bytes != tc.sourceBytes || mutation.Source.Location != tc.mutationLocation {
				t.Fatalf("mutation provider source = %#v, want exact retained source-lock citation", mutation.Source)
			}
			if mutation.Runtime.NonExecutableMutation.Source.SourceID != tc.mutationID ||
				!strings.EqualFold(mutation.Runtime.NonExecutableMutation.Source.Method, tc.method) ||
				mutation.Runtime.NonExecutableMutation.Source.Path != tc.mutationPath {
				t.Fatalf("mutation citation = %#v, want exact locked source identity", mutation.Runtime.NonExecutableMutation.Source)
			}
			if !sourceOperationHasFoundationGap(mutation, sourceNonExecutableMutationDispositionFoundation) || !strings.Contains(sourceProjectionNonExecutableMutationRuntimeGap(mutation, *mutation.Runtime.NonExecutableMutation).Location, tc.sourceURL) {
				t.Fatalf("mutation gap = %#v, want named source-cited foundation", mutation.Runtime.Gaps)
			}
			if findings := validateSourceExecutableCoverage(bundle, "sources/"+tc.connector+"-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 0 {
				t.Fatalf("read-only source executable coverage findings = %+v", findings)
			}

			bundleDir := t.TempDir()
			writesPath := filepath.Join(bundleDir, "writes.json")
			cliPath := filepath.Join(bundleDir, "cli_surface.json")
			writeProjectionFixture(t, filepath.Join(bundleDir, "metadata.json"), `{"name":"`+tc.connector+`","capabilities":{"write":false}}`)
			const emptyWrites = `{"schema_version":1,"actions":[]}`
			cliRaw, err := json.Marshal(engine.CLISurface{Commands: []engine.CLICommand{{
				Path: "projects list", Summary: "list", Intent: "direct_read", Availability: "implemented",
				APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: tc.readPath}}, Flags: pathFlags(tc.readPath),
			}}})
			if err != nil {
				t.Fatal(err)
			}
			cli := string(cliRaw)
			writeProjectionFixture(t, writesPath, emptyWrites)
			writeProjectionFixture(t, cliPath, cli)
			stats, err := projectSourceDescriptorToBundle(bundleDir, result, false)
			if err != nil || stats.Changed() {
				t.Fatalf("source projection = stats:%+v err:%v, want no fabricated write or command", stats, err)
			}
			if gotWrites, gotCLI := readProjectionFixture(t, writesPath), readProjectionFixture(t, cliPath); gotWrites != emptyWrites || gotCLI != cli {
				t.Fatalf("source projection fabricated a write or command:\nwrites=%s\ncli=%s", gotWrites, gotCLI)
			}
		})
	}
}

func TestSourceProjectionIssue4329SourceLocksRetainMutationInventoryAndReadSurface(t *testing.T) {
	pathFlags := func(path string) []engine.CLIFlag {
		flags := make([]engine.CLIFlag, 0)
		for _, match := range sourceProjectionPathVariableRE.FindAllStringSubmatch(path, -1) {
			flags = append(flags, engine.CLIFlag{Name: strings.ReplaceAll(match[1], "_", "-"), Type: "string", MapsTo: "path." + match[1], Required: true})
		}
		return flags
	}
	tests := []struct {
		connector       string
		lockFile        string
		expectedOps     int
		expectedMutates int
		readOperationID string
		mutationID      string
	}{
		{connector: "sentry", lockFile: "sentry-operation-source-lock.json", expectedOps: 223, expectedMutates: 103, readOperationID: "listOrganizationProjects", mutationID: "createOrganizationDashboard"},
		{connector: "vercel", lockFile: "vercel-operation-source-lock.json", expectedOps: 400, expectedMutates: 237, readOperationID: "getProjects", mutationID: "deleteStorageStoresBlobById"},
	}

	for _, tc := range tests {
		t.Run(tc.connector, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", "issue4329", tc.lockFile))
			if err != nil {
				t.Fatal(err)
			}
			lock, err := parseSourceImportLock(raw, tc.connector)
			if err != nil {
				t.Fatalf("parse preserved %s source lock: %v", tc.connector, err)
			}
			if got := len(lock.Rest.Operations); got != tc.expectedOps {
				t.Fatalf("locked operations = %d, want %d", got, tc.expectedOps)
			}
			toDescriptor := func(operation sourceImportRESTOperation) sourceOperationDescriptor {
				return sourceOperationDescriptor{
					Connector: tc.connector, SourceID: operation.OperationID, Method: operation.Method, Path: operation.Path,
					ProviderOperationID: operation.OperationID,
					Source:              sourceImportSource{URL: lock.Rest.SourceURL, SHA256: lock.Rest.SHA256, Bytes: lock.Rest.Bytes, Location: operation.SourceLocation, Form: "openapi", Version: lock.Rest.OpenAPI},
				}
			}

			allMutations := sourceImportResult{}
			var read, selectedMutation sourceOperationDescriptor
			for _, operation := range lock.Rest.Operations {
				descriptor := toDescriptor(operation)
				if sourceProjectionOperationMutates(descriptor) {
					allMutations.Operations = append(allMutations.Operations, descriptor)
				}
				if operation.OperationID == tc.readOperationID {
					read = descriptor
				}
				if operation.OperationID == tc.mutationID {
					selectedMutation = descriptor
				}
			}
			if len(allMutations.Operations) != tc.expectedMutates || read.SourceID == "" || selectedMutation.SourceID == "" {
				t.Fatalf("source-lock selection = mutations:%d read:%#v mutation:%#v", len(allMutations.Operations), read, selectedMutation)
			}

			writeDisabled := engine.Bundle{Name: tc.connector, Metadata: engine.Metadata{Name: tc.connector, Capabilities: engine.Capabilities{Read: true, Write: false, WriteDeclared: true}}}
			if got := sourceProjectionApplyWriteDisabledMutationArtifacts(writeDisabled, &allMutations); got != tc.expectedMutates {
				t.Fatalf("retained mutation artifacts = %d, want %d", got, tc.expectedMutates)
			}
			if findings := validateSourceExecutableCoverage(writeDisabled, "sources/"+tc.connector+"-operation-descriptor.json", sourceImportDescriptorDocument{Operations: allMutations.Operations}); len(findings) != 0 {
				t.Fatalf("full retained mutation inventory findings = %+v", findings)
			}

			readSurface := writeDisabled
			readSurface.CLISurface = &engine.CLISurface{Commands: []engine.CLICommand{{
				Path: "projects list", Intent: "direct_read", Availability: "implemented",
				APISurface: []engine.CLISurfaceEndpointRef{{Method: read.Method, Path: read.Path}}, Flags: pathFlags(read.Path),
			}}}
			selected := sourceImportResult{Operations: []sourceOperationDescriptor{read, selectedMutation}}
			if got := sourceProjectionApplyWriteDisabledMutationArtifacts(readSurface, &selected); got != 1 {
				t.Fatalf("selected mutation artifacts = %d, want 1", got)
			}
			if findings := validateSourceExecutableCoverage(readSurface, "sources/"+tc.connector+"-operation-descriptor.json", sourceImportDescriptorDocument{Operations: selected.Operations}); len(findings) != 0 {
				t.Fatalf("source-locked read surface findings = %+v", findings)
			}
			if !sourceProjectionHasNonExecutableMutationDisposition(selected.Operations[1]) || selected.Operations[1].Runtime.NonExecutableMutation.Source.SourceID != tc.mutationID {
				t.Fatalf("selected provider mutation = %#v, want exact source-cited artifact", selected.Operations[1])
			}
		})
	}
}

func TestSourceProjectionWriteDisabledMutationArtifactsPreserveExecutableDeletes(t *testing.T) {
	operation := sourceCitedMutationTestOperation("sentry", "deleteOrganizationDashboard", "DELETE", "/api/0/dashboards/current/")
	bundle := engine.Bundle{
		Name:     "sentry",
		Metadata: engine.Metadata{Name: "sentry", Capabilities: engine.Capabilities{Read: true, Write: false, WriteDeclared: true}},
		Writes: []engine.WriteAction{{
			Name: "delete_dashboard", Kind: "delete", Method: "DELETE",
			Path:         "/api/0/dashboards/current/",
			RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{}}`),
			Risk:         "destructive",
		}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "dashboards delete", Intent: "reverse_etl", Availability: "implemented", Write: "delete_dashboard",
		}}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if got := sourceProjectionApplyWriteDisabledMutationArtifacts(bundle, &result); got != 0 {
		t.Fatalf("automatic mutation artifact count = %d, want 0 for executable delete", got)
	}
	if result.Operations[0].Runtime.NonExecutableMutation != nil {
		t.Fatalf("executable delete received a non-executable mutation artifact: %#v", result.Operations[0].Runtime)
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("executable delete coverage findings = %+v", findings)
	}
}

// Regression acceptance for #4329: an exact, source-locked provider delete
// with a real reverse-ETL action remains executable. capabilities.write=false
// only retains mutations that lack this declaration-owned foundation; it must
// not suppress a usable delete route as a safety policy.
func TestSourceProjectionWriteDisabledMutationArtifactsPreserveSourceLockedExecutableDeletes(t *testing.T) {
	tests := []struct {
		connector   string
		lockFile    string
		operationID string
		action      engine.WriteAction
		command     engine.CLICommand
	}{
		{
			connector:   "sentry",
			lockFile:    "sentry-operation-source-lock.json",
			operationID: "deleteOrganizationDashboard",
			action: engine.WriteAction{
				Name:         "delete_organization_dashboard",
				Kind:         "delete",
				Method:       "DELETE",
				Path:         "/api/0/organizations/{{ record.organization_id_or_slug }}/dashboards/{{ record.dashboard_id }}/",
				PathFields:   []string{"organization_id_or_slug", "dashboard_id"},
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"organization_id_or_slug":{"type":"string","minLength":1},"dashboard_id":{"type":"string","minLength":1}},"required":["organization_id_or_slug","dashboard_id"]}`),
				Risk:         "destructive",
			},
			command: engine.CLICommand{
				Path: "dashboards delete", Intent: "reverse_etl", Availability: "implemented", Write: "delete_organization_dashboard",
				Flags: []engine.CLIFlag{
					{Name: "organization-id-or-slug", Type: "string", MapsTo: "record.organization_id_or_slug", Required: true},
					{Name: "dashboard-id", Type: "string", MapsTo: "record.dashboard_id", Required: true},
				},
			},
		},
		{
			connector:   "vercel",
			lockFile:    "vercel-operation-source-lock.json",
			operationID: "deleteStorageStoresBlobById",
			action: engine.WriteAction{
				Name:         "delete_storage_blob",
				Kind:         "delete",
				Method:       "DELETE",
				Path:         "/storage/stores/blob/{{ record.id }}",
				PathFields:   []string{"id"},
				RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"id":{"type":"string","minLength":1}},"required":["id"]}`),
				Risk:         "destructive",
			},
			command: engine.CLICommand{
				Path: "storage blobs delete", Intent: "reverse_etl", Availability: "implemented", Write: "delete_storage_blob",
				Flags: []engine.CLIFlag{
					{Name: "id", Type: "string", MapsTo: "record.id", Required: true},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.connector, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", "issue4329", tc.lockFile))
			if err != nil {
				t.Fatal(err)
			}
			lock, err := parseSourceImportLock(raw, tc.connector)
			if err != nil {
				t.Fatalf("parse preserved %s source lock: %v", tc.connector, err)
			}
			var operation sourceOperationDescriptor
			for _, candidate := range lock.Rest.Operations {
				if candidate.OperationID != tc.operationID {
					continue
				}
				operation = sourceOperationDescriptor{
					Connector: tc.connector, SourceID: candidate.OperationID, Method: candidate.Method, Path: candidate.Path,
					ProviderOperationID: candidate.OperationID,
					Source:              sourceImportSource{URL: lock.Rest.SourceURL, SHA256: lock.Rest.SHA256, Bytes: lock.Rest.Bytes, Location: candidate.SourceLocation, Form: "openapi", Version: lock.Rest.OpenAPI},
				}
				break
			}
			if operation.SourceID == "" || !sourceProjectionOperationMutates(operation) {
				t.Fatalf("source-locked executable delete %q = %#v", tc.operationID, operation)
			}

			tc.command.APISurface = []engine.CLISurfaceEndpointRef{{Method: operation.Method, Path: operation.Path}}
			bundle := engine.Bundle{
				Name:       tc.connector,
				Metadata:   engine.Metadata{Name: tc.connector, Capabilities: engine.Capabilities{Read: true, Write: false, WriteDeclared: true}},
				Writes:     []engine.WriteAction{tc.action},
				CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{tc.command}},
			}
			if !sourceProjectionMutationActionIsComplete(bundle, operation) {
				t.Fatalf("source-locked %s delete action is not complete", tc.connector)
			}
			if !sourceProjectionMutationClaimsImplementedAction(bundle, operation) {
				t.Fatalf("source-locked %s reverse-ETL command does not claim its exact provider operation", tc.connector)
			}

			result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
			if got := sourceProjectionApplyWriteDisabledMutationArtifacts(bundle, &result); got != 0 {
				t.Fatalf("automatic mutation artifact count = %d, want 0 for executable %s delete", got, tc.connector)
			}
			if result.Operations[0].Runtime.NonExecutableMutation != nil {
				t.Fatalf("executable %s delete received a non-executable mutation artifact: %#v", tc.connector, result.Operations[0].Runtime)
			}
			if findings := validateSourceExecutableCoverage(bundle, "sources/"+tc.connector+"-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 0 {
				t.Fatalf("source-locked %s executable delete coverage findings = %+v", tc.connector, findings)
			}
		})
	}
}

func TestSourceProjectionWriteCapableBundlesDoNotAutoDeferMutations(t *testing.T) {
	operation := sourceCitedMutationTestOperation("vercel", "deleteStorageStoresBlobById", "DELETE", "/storage/stores/blob/{id}")
	bundle := engine.Bundle{
		Name:       "vercel",
		Metadata:   engine.Metadata{Name: "vercel", Capabilities: engine.Capabilities{Read: true, Write: true}},
		CLISurface: &engine.CLISurface{},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if got := sourceProjectionApplyWriteDisabledMutationArtifacts(bundle, &result); got != 0 {
		t.Fatalf("automatic mutation artifact count = %d, want 0 for write-capable bundle", got)
	}
	if result.Operations[0].Runtime.NonExecutableMutation != nil {
		t.Fatalf("write-capable mutation received a non-executable artifact: %#v", result.Operations[0].Runtime)
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/vercel-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 1 || !strings.Contains(findings[0].Message, "no executable action") {
		t.Fatalf("write-capable mutation coverage findings = %+v, want missing-action refusal", findings)
	}

	writeDisabled := bundle
	writeDisabled.Metadata.Capabilities.Write = false
	writeDisabled.Metadata.Capabilities.WriteDeclared = true
	if got := sourceProjectionApplyWriteDisabledMutationArtifacts(writeDisabled, &result); got != 1 {
		t.Fatalf("automatic mutation artifact count = %d, want 1 for explicitly write-disabled bundle", got)
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/vercel-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 1 || !strings.Contains(findings[0].Message, "automatic write-disabled mutation artifact requires connector metadata capabilities.write=false") {
		t.Fatalf("write-capable automatic artifact coverage findings = %+v, want policy refusal", findings)
	}
}

func TestSourceProjectionAutomaticMutationArtifactRejectsOmittedWriteDeclaration(t *testing.T) {
	metadata, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "sentry", "metadata.json"))
	if err != nil {
		t.Fatalf("read current Sentry metadata: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(metadata, &document); err != nil {
		t.Fatalf("decode current Sentry metadata: %v", err)
	}
	capabilities, ok := document["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("current Sentry capabilities = %#v, want object", document["capabilities"])
	}
	delete(capabilities, "write")
	metadata, err = json.Marshal(document)
	if err != nil {
		t.Fatalf("encode adversarial Sentry metadata: %v", err)
	}
	bundleDir := t.TempDir()
	writeProjectionFixture(t, filepath.Join(bundleDir, "metadata.json"), string(metadata))
	_, err = sourceProjectionExecutionSurface(bundleDir, "sentry")
	if err == nil || !strings.Contains(err.Error(), "capabilities.write must be explicitly declared") {
		t.Fatalf("load adversarial Sentry metadata error = %v, want explicit write declaration refusal", err)
	}
	surface := engine.Bundle{Name: "sentry", Metadata: engine.Metadata{Name: "sentry", Capabilities: engine.Capabilities{Read: true}}}

	lockRaw, err := os.ReadFile(filepath.Join("testdata", "issue4329", "sentry-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read preserved Sentry source lock: %v", err)
	}
	lock, err := parseSourceImportLock(lockRaw, "sentry")
	if err != nil {
		t.Fatalf("parse preserved Sentry source lock: %v", err)
	}
	var operation sourceOperationDescriptor
	for _, candidate := range lock.Rest.Operations {
		if candidate.OperationID != "createOrganizationDashboard" {
			continue
		}
		operation = sourceOperationDescriptor{
			Connector: "sentry", SourceID: candidate.OperationID, Method: candidate.Method, Path: candidate.Path,
			ProviderOperationID: candidate.OperationID,
			Source:              sourceImportSource{URL: lock.Rest.SourceURL, SHA256: lock.Rest.SHA256, Bytes: lock.Rest.Bytes, Location: candidate.SourceLocation, Form: "openapi", Version: lock.Rest.OpenAPI},
		}
		break
	}
	if operation.SourceID == "" || !sourceProjectionOperationMutates(operation) {
		t.Fatalf("source-locked Sentry mutation = %#v", operation)
	}

	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if got := sourceProjectionApplyWriteDisabledMutationArtifacts(surface, &result); got != 0 {
		t.Fatalf("automatic mutation artifact count = %d, want 0 without explicit capabilities.write=false", got)
	}
	if result.Operations[0].Runtime.NonExecutableMutation != nil {
		t.Fatalf("omitted capabilities.write emitted automatic artifact: %#v", result.Operations[0].Runtime)
	}

	disposition := sourceNonExecutableMutationDisposition{
		Source: sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path},
		Reason: sourceWriteDisabledMutationArtifactReason,
	}
	result.Operations[0].Runtime.NonExecutableMutation = &disposition
	result.Operations[0].Runtime.Gaps = []sourceContractGap{sourceProjectionNonExecutableMutationRuntimeGap(result.Operations[0], disposition)}
	result.Operations[0].Runtime.MergeBlocked = true
	findings := validateSourceExecutableCoverage(surface, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations})
	if len(findings) != 1 || !strings.Contains(findings[0].Message, "requires connector metadata capabilities.write=false") {
		t.Fatalf("omitted capabilities.write coverage findings = %+v, want explicit write declaration refusal", findings)
	}
}

func TestSourceProjectionWriteDisabledMutationArtifactsRequireProviderCitation(t *testing.T) {
	operation := sourceCitedMutationTestOperation("sentry", "deleteOrganizationDashboard", "DELETE", "/api/0/dashboards/current/")
	operation.Source = sourceImportSource{}
	bundle := engine.Bundle{
		Name:     "sentry",
		Metadata: engine.Metadata{Name: "sentry", Capabilities: engine.Capabilities{Read: true, Write: false, WriteDeclared: true}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if got := sourceProjectionApplyWriteDisabledMutationArtifacts(bundle, &result); got != 0 {
		t.Fatalf("automatic mutation artifact count = %d, want 0 without a provider citation", got)
	}
	if result.Operations[0].Runtime.NonExecutableMutation != nil {
		t.Fatalf("uncited mutation received a non-executable artifact: %#v", result.Operations[0].Runtime)
	}
}

func TestSourceProjectionWriteDisabledMutationArtifactsRetainGraphQLMutations(t *testing.T) {
	operation := sourceCitedMutationTestOperation("sentry", "sentry.graphql.mutation.resolve", "POST", "/graphql")
	operation.Protocol = "graphql"
	operation.GraphQL = &sourceGraphQLOperationDescriptor{Root: "Mutation", Name: "resolveIssue"}
	bundle := engine.Bundle{
		Name:     "sentry",
		Metadata: engine.Metadata{Name: "sentry", Capabilities: engine.Capabilities{Read: true, Write: false, WriteDeclared: true}},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if got := sourceProjectionApplyWriteDisabledMutationArtifacts(bundle, &result); got != 1 {
		t.Fatalf("automatic mutation artifact count = %d, want 1 for a GraphQL mutation", got)
	}
	if !sourceProjectionHasNonExecutableMutationDisposition(result.Operations[0]) {
		t.Fatalf("GraphQL mutation = %#v, want cited non-executable artifact", result.Operations[0].Runtime)
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/sentry-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("GraphQL mutation coverage findings = %+v", findings)
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

func TestSourceProjectionNonExecutableMutationDispositionAllowsOnlyClosedBodylessPOSTRead(t *testing.T) {
	const (
		sourceID = "postSourceLookup"
		path     = "/api/v4/lookups"
		mapped   = "/lookups"
	)

	newSource := func(outputClass sourceOutputClass) sourceOperationDescriptor {
		mediaType := ""
		if outputClass == sourceOutputJSON {
			mediaType = "application/json"
		}
		return sourceOperationDescriptor{
			Connector: "gitlab", Protocol: "rest", SourceID: sourceID, ProviderOperationID: sourceID,
			Method: http.MethodPost, Path: path, MappingPath: mapped,
			Source: sourceImportSource{URL: "https://provider.example.test/openapi.json", Location: "#/paths/~1api~1v4~1lookups/post"},
			Output: sourceOutputDescriptor{Class: outputClass, Success: []sourceOutputVariant{{Status: "200", MediaType: mediaType, Class: outputClass}}},
		}
	}
	newBundle := func(noRequestBody bool, intent string, bodySchema json.RawMessage, contentType string, flags []engine.CLIFlag) engine.Bundle {
		return engine.Bundle{
			Name: "gitlab",
			Operations: []engine.OperationSpec{{
				ID: "source_read_post_lookup", Kind: "rest_read", Risk: "low", Approval: "none", OutputPolicy: "json_redacted",
				SourceOperation: &engine.SourceOperationBinding{ID: sourceID, Method: http.MethodPost, Path: mapped},
				REST:            &engine.RESTOperationSpec{Method: http.MethodPost, Path: mapped, NoRequestBody: noRequestBody, ContentType: contentType, BodySchema: bodySchema, MaxBytes: 1024},
			}},
			CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
				Path: "api source-read-post-lookup", Intent: intent, Availability: "implemented", Operation: "source_read_post_lookup", SourceOperation: sourceID,
				APISurface: []engine.CLISurfaceEndpointRef{{Method: http.MethodPost, Path: mapped}}, Flags: flags,
			}}},
		}
	}

	disposition := sourceNonExecutableMutationDisposition{
		Source: sourceOperationCitation{SourceID: sourceID, Method: http.MethodPost, Path: path},
		Reason: "legacy method-only mutation classification",
	}
	t.Run("closed bodyless semantic direct read", func(t *testing.T) {
		// A provider can document a bodyless semantic lookup as a status-only
		// response. It remains a closed direct read when its retained request
		// contract has no body and the declared operation opts into the exact
		// no_request_body form.
		source := newSource(sourceOutputStatus)
		bundle := newBundle(true, "direct_read", nil, "", nil)
		if sourceProjectionMutationClaimsImplementedAction(bundle, source) {
			t.Fatal("closed bodyless POST direct read was treated as an implemented mutation")
		}
		if err := sourceProjectionApplyNonExecutableMutationDispositions(bundle, &sourceImportResult{Operations: []sourceOperationDescriptor{source}}, []sourceNonExecutableMutationDisposition{disposition}); err != nil {
			t.Fatalf("closed bodyless POST direct read admission: %v", err)
		}
	})

	for _, testCase := range []struct {
		name          string
		noRequestBody bool
		intent        string
		bodySchema    json.RawMessage
		contentType   string
		flags         []engine.CLIFlag
		body          *sourceRequestBodyDescriptor
	}{
		{
			name:        "JSON body POST remains a mutation",
			intent:      "direct_read",
			bodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false}`),
			contentType: "application/json",
			body:        &sourceRequestBodyDescriptor{Schema: map[string]any{"type": "object", "additionalProperties": false}},
		},
		{
			name:   "missing no-request-body marker remains a mutation",
			intent: "direct_read",
		},
		{
			name:          "body flag defeats no-request-body marker",
			noRequestBody: true,
			intent:        "direct_read",
			flags:         []engine.CLIFlag{{Name: "payload", MapsTo: "body.payload"}},
		},
		{
			name:   "non GitLab mutation remains a mutation",
			intent: "direct_write",
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			source := newSource(sourceOutputJSON)
			source.Request.Body = testCase.body
			bundle := newBundle(testCase.noRequestBody, testCase.intent, testCase.bodySchema, testCase.contentType, testCase.flags)
			if !sourceProjectionMutationClaimsImplementedAction(bundle, source) {
				t.Fatal("ordinary or malformed POST was incorrectly admitted as a semantic direct read")
			}
			if err := sourceProjectionApplyNonExecutableMutationDispositions(bundle, &sourceImportResult{Operations: []sourceOperationDescriptor{source}}, []sourceNonExecutableMutationDisposition{disposition}); err == nil || !strings.Contains(err.Error(), "implemented executable action") {
				t.Fatalf("ordinary or malformed POST mutation disposition error = %v, want implemented-action rejection", err)
			}
		})
	}
}

// TestGitLabSourceProjectionAdmitsOnlyRetainedBodylessSemanticPOSTReads
// proves that the narrow exception reaches the real retained GitLab source
// rows consumed by the canonical surface-sync path. The table is intentionally
// source-ID based: a generic POST method exception would make unrelated
// mutations look executable.
func TestGitLabSourceProjectionAdmitsOnlyRetainedBodylessSemanticPOSTReads(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	bundle, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab bundle: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources", "gitlab-operation-descriptor.json"))
	if err != nil {
		t.Fatalf("read GitLab source descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(raw, &descriptor); err != nil {
		t.Fatalf("decode GitLab source descriptor: %v", err)
	}
	sources := make(map[string]sourceOperationDescriptor, len(descriptor.Operations))
	for _, source := range descriptor.Operations {
		sources[source.SourceID] = source
	}

	want := []string{
		"postApiV4AiThirdPartyAgentsDirectAccess",
		"postApiV4CodeSuggestionsConnectionDetails",
		"postApiV4GeoNodeProxyIdGraphql",
		"postApiV4IntegrationsSlackOptions",
		"postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls",
		"postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls",
		"postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls",
		"postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls",
	}
	// These legacy commands share a route with the semantic read. Their
	// continued reachability must not make the exact no-body read source row
	// method-classify as a mutation during surface-sync admission.
	legacyMutationViews := map[string]bool{
		"postApiV4GeoNodeProxyIdGraphql": true,
		"postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls": true,
		"postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls":                              true,
	}
	for _, sourceID := range want {
		source, found := sources[sourceID]
		if !found {
			t.Fatalf("GitLab source descriptor omits %q", sourceID)
		}
		if source.Request.Body != nil || strings.TrimSpace(source.Request.MediaType) != "" || len(source.Request.Media) != 0 {
			t.Fatalf("GitLab source %q is not a bodyless source contract: %+v", sourceID, source.Request)
		}
		if sourceProjectionMutationClaimsImplementedAction(bundle, source) {
			t.Fatalf("GitLab closed bodyless semantic POST read %q remains classified as an implemented mutation", sourceID)
		}
		intents := map[string]bool{}
		for _, command := range bundle.CLISurface.Commands {
			if command.SourceOperation == sourceID && command.Availability == "implemented" {
				intents[command.Intent] = true
			}
		}
		if !intents["direct_read"] {
			t.Fatalf("GitLab closed bodyless semantic POST read %q lost its direct_read command", sourceID)
		}
		if legacyMutationViews[sourceID] && (!intents["direct_write"] || !intents["reverse_etl"]) {
			t.Fatalf("GitLab semantic POST read %q legacy write/reverse views = %v, want both preserved", sourceID, intents)
		}
	}

	const genuineMutationID = "postApiV4AdminCiVariables"
	genuineMutation, found := sources[genuineMutationID]
	if !found || genuineMutation.Request.Body == nil {
		t.Fatalf("GitLab genuine POST mutation %q lost its retained request body", genuineMutationID)
	}
	if !sourceProjectionMutationClaimsImplementedAction(bundle, genuineMutation) {
		t.Fatalf("GitLab genuine POST mutation %q was incorrectly admitted as a semantic direct read", genuineMutationID)
	}
}

// TestGitLabClosedBodylessPOSTReadsReachSurfaceSync proves that the same
// retained semantic POST-read cohort can reach the canonical source-projection
// step that precedes endpoint-ledger reconciliation. This exercises the
// projection bundle, not merely an engine-loaded fixture.
func TestGitLabClosedBodylessPOSTReadsReachSurfaceSync(t *testing.T) {
	stats, err := syncCheckedInSourceProjection(filepath.Join("..", "..", "internal", "connectors", "defs", "gitlab"), "gitlab", true)
	if err != nil {
		t.Fatalf("GitLab source projection rejects closed bodyless semantic POST reads: %v", err)
	}
	// This is an admission boundary, not a broad surface-sync drift test. A
	// legacy write projection must not replace any of the exact bodyless POST
	// direct-read API bindings while the canonical ledger is being reconciled.
	if stats.Surface != 0 {
		t.Fatalf("GitLab source projection rewrote semantic POST direct-read API bindings: %+v", stats)
	}
}

func TestSourceProjectionSourceCitedPartialMutationCoveragePreservesImplementedIncompleteAction(t *testing.T) {
	operation := sourceCitedMutationTestOperation("asana", "asana.items.update", "PATCH", "/items/{item_id}")
	operation.Request.Path = []sourceParameterDescriptor{{Name: "item_id", Required: true, Schema: map[string]any{"type": "string"}}}
	operation.Request.Body = &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
		"type": "object", "properties": map[string]any{"payload": map[string]any{"type": "object", "additionalProperties": true}},
	}}
	operation.Runtime = sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{{
		Foundation: "cli-request-schema-foundation-r1",
		Location:   "request body",
		Reason:     "unbounded request schema object has dynamic additionalProperties",
	}}}
	bundle := engine.Bundle{Name: "asana", Writes: []engine.WriteAction{{
		Name: "update-item", Kind: "update", Method: "PATCH", Path: "/items/{{ record.item_id }}", PathFields: []string{"item_id"}, BodyType: "json",
		RecordSchema: json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"item_id":{"type":"string"}}}`),
	}}, CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
		Path: "items update", Intent: "reverse_etl", Availability: "implemented", Write: "update-item",
		Flags: []engine.CLIFlag{{Name: "item-id", Type: "string", MapsTo: "record.item_id", Required: true}},
	}}}}
	if findings := validateSourceExecutableCoverage(bundle, "sources/asana-operation-descriptor.json", sourceImportDescriptorDocument{Operations: []sourceOperationDescriptor{operation}}); len(findings) != 1 || !strings.Contains(findings[0].Message, "retains an unresolved source-bound gap") {
		t.Fatalf("undisposed implemented incomplete action findings = %+v", findings)
	}
	disposition := sourcePartialMutationCoverageDisposition{
		Source:     sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path},
		Foundation: "cli-request-schema-foundation-r1",
		Reason:     "provider request contract needs typed dynamic-object support",
	}
	wrongFoundation := disposition
	wrongFoundation.Foundation = "source-path-parameter-alias-foundation-r1"
	if err := sourceProjectionApplyPartialMutationCoverageDispositions(bundle, &sourceImportResult{Operations: []sourceOperationDescriptor{operation}}, []sourcePartialMutationCoverageDisposition{wrongFoundation}); err == nil || !strings.Contains(err.Error(), "no matching missing foundation") {
		t.Fatalf("non-alias partial-coverage disposition error = %v, want named foundation refusal", err)
	}
	nonMutation := operation
	nonMutation.SourceID = "asana.items.inspect"
	nonMutation.Method = "GET"
	nonMutation.Runtime = sourceRuntimeReachability{}
	nonMutationDisposition := disposition
	nonMutationDisposition.Source = sourceOperationCitation{SourceID: nonMutation.SourceID, Method: nonMutation.Method, Path: nonMutation.Path}
	if err := sourceProjectionApplyPartialMutationCoverageDispositions(bundle, &sourceImportResult{Operations: []sourceOperationDescriptor{nonMutation}}, []sourcePartialMutationCoverageDisposition{nonMutationDisposition}); err == nil || !strings.Contains(err.Error(), "is not mutating") {
		t.Fatalf("non-mutating partial-coverage disposition error = %v, want source-operation refusal", err)
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	if err := sourceProjectionApplyPartialMutationCoverageDispositions(bundle, &result, []sourcePartialMutationCoverageDisposition{disposition}); err != nil {
		t.Fatalf("apply partial source coverage disposition: %v", err)
	}
	got := result.Operations[0]
	if got.Runtime.PartialCoverageMutation == nil || got.Runtime.PartialCoverageMutation.Foundation != disposition.Foundation || got.Runtime.NonExecutableMutation != nil || !sourceProjectionHasPartialMutationCoverageDisposition(got) {
		t.Fatalf("partial source coverage runtime = %+v, want exact partial disposition", got.Runtime)
	}
	if findings := validateSourceExecutableCoverage(bundle, "sources/asana-operation-descriptor.json", sourceImportDescriptorDocument{Operations: result.Operations}); len(findings) != 0 {
		t.Fatalf("partial source coverage findings = %+v", findings)
	}
	bundleDir := t.TempDir()
	const writes = `{"schema_version":1,"actions":[{"name":"update-item","kind":"update","method":"PATCH","path":"/items/{{ record.item_id }}","path_fields":["item_id"],"body_type":"json","record_schema":{"type":"object","additionalProperties":false,"properties":{"item_id":{"type":"string"}}},"risk":"standard"}]}`
	const cli = `{"schema_version":1,"commands":[{"path":"items update","summary":"update","intent":"reverse_etl","availability":"implemented","write":"update-item","flags":[{"name":"item-id","type":"string","maps_to":"record.item_id","required":true}]}]}`
	writeProjectionFixture(t, filepath.Join(bundleDir, "writes.json"), writes)
	writeProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json"), cli)
	stats, err := projectSourceDescriptorToBundle(bundleDir, result, false)
	if err != nil || stats.Changed() {
		t.Fatalf("partial source coverage projection = stats:%+v err:%v, want byte-stable declared action", stats, err)
	}
	if gotWrites, gotCLI := readProjectionFixture(t, filepath.Join(bundleDir, "writes.json")), readProjectionFixture(t, filepath.Join(bundleDir, "cli_surface.json")); gotWrites != writes || gotCLI != cli {
		t.Fatalf("partial source coverage changed a working action or command:\nwrites=%s\ncli=%s", gotWrites, gotCLI)
	}
	if bundle.CLISurface.Commands[0].Availability != "implemented" {
		t.Fatalf("partial source coverage command availability = %q, want implemented", bundle.CLISurface.Commands[0].Availability)
	}

	complete := sourceProjectionTestOperation()
	complete.Connector = "asana"
	complete.SourceID = "asana.items.create"
	complete.Source = sourceCitedMutationTestOperation("asana", complete.SourceID, complete.Method, complete.Path).Source
	completeDisposition := sourcePartialMutationCoverageDisposition{
		Source:     sourceOperationCitation{SourceID: complete.SourceID, Method: complete.Method, Path: complete.Path},
		Foundation: "cli-request-schema-foundation-r1",
		Reason:     "must not hide a complete action",
	}
	completeDir := t.TempDir()
	completeWritesPath := filepath.Join(completeDir, "writes.json")
	completeCLIPath := filepath.Join(completeDir, "cli_surface.json")
	writeProjectionFixture(t, completeWritesPath, `{"schema_version":1,"actions":[{"name":"items","kind":"custom","method":"POST","path":"/items/{{ record.owner }}","path_fields":["owner"],"body_type":"json","body_fields":["stale"],"record_schema":{"type":"object","additionalProperties":false,"properties":{"stale":{"type":"string"}}},"risk":"standard"}]}`)
	writeProjectionFixture(t, completeCLIPath, `{"schema_version":1,"commands":[{"path":"items create","summary":"create","intent":"reverse_etl","availability":"implemented","write":"items","flags":[{"name":"stale","type":"string","maps_to":"record.stale"}]}]}`)
	if _, err := projectSourceDescriptorToBundle(completeDir, sourceImportResult{Operations: []sourceOperationDescriptor{complete}}, false); err != nil {
		t.Fatalf("materialize complete action before partial-coverage counterfactual: %v", err)
	}
	var completeWrites struct {
		Actions []engine.WriteAction `json:"actions"`
	}
	if err := json.Unmarshal([]byte(readProjectionFixture(t, completeWritesPath)), &completeWrites); err != nil {
		t.Fatal(err)
	}
	var completeCLI engine.CLISurface
	if err := json.Unmarshal([]byte(readProjectionFixture(t, completeCLIPath)), &completeCLI); err != nil {
		t.Fatal(err)
	}
	completeBundle := engine.Bundle{Name: "asana", Writes: completeWrites.Actions, CLISurface: &completeCLI}
	if err := sourceProjectionApplyPartialMutationCoverageDispositions(completeBundle, &sourceImportResult{Operations: []sourceOperationDescriptor{complete}}, []sourcePartialMutationCoverageDisposition{completeDisposition}); err == nil || !strings.Contains(err.Error(), "complete executable action") {
		t.Fatalf("complete action partial-coverage disposition error = %v, want refusal", err)
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

func TestSourceProjectionGeneratedNoBodyMutationClearsOnlySyntheticDisposition(t *testing.T) {
	operation := sourceCitedMutationTestOperation("bitbucket", "updateEnvironmentForRepository", "POST", "/repositories/{workspace}/{repo_slug}/environments/{environment_uuid}/changes")
	disposition := sourceNonExecutableMutationDisposition{
		Source: sourceOperationCitation{SourceID: operation.SourceID, Method: operation.Method, Path: operation.Path},
		Reason: "provider-cited mutation has no complete declaration-owned executable action for its full retained request contract",
	}
	operation.Runtime = sourceRuntimeReachability{
		MergeBlocked:          true,
		NonExecutableMutation: &disposition,
		Gaps:                  []sourceContractGap{sourceProjectionNonExecutableMutationRuntimeGap(operation, disposition)},
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{operation}}
	sourceProjectionClearGeneratedNoBodyMutationDisposition(&result, operation)
	if got := result.Operations[0].Runtime; got.MergeBlocked || got.NonExecutableMutation != nil || len(got.Gaps) != 0 {
		t.Fatalf("generated no-body mutation disposition = %#v, want cleared exact synthetic gap", got)
	}

	blocked := operation
	blocked.Runtime.Gaps = append(blocked.Runtime.Gaps, sourceContractGap{Foundation: "other-foundation", Location: "request.body"})
	result = sourceImportResult{Operations: []sourceOperationDescriptor{blocked}}
	sourceProjectionClearGeneratedNoBodyMutationDisposition(&result, blocked)
	if got := result.Operations[0].Runtime; !got.MergeBlocked || got.NonExecutableMutation == nil || len(got.Gaps) != 2 {
		t.Fatalf("non-sole mutation disposition = %#v, want retained source gaps", got)
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
