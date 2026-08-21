package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func readProjectionFixture(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
