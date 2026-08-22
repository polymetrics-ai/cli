package commandrunner

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestGitHubDeclaredParityVerdicts pins the 50-row issue #4015 inventory to
// executable commands or an exact, durable capability-boundary explanation.
// It uses commandrunner.Preflight for promotions so a populated api_surface
// string cannot masquerade as a command the binary actually dispatches.
func TestGitHubDeclaredParityVerdicts(t *testing.T) {
	t.Parallel()

	implemented := []string{
		"issue pin",
		"issue unpin",
		"pr diff",
		"pr ready",
		"repo list",
		"repo autolink create",
		"repo autolink delete",
		"workflow enable",
		"workflow disable",
		"variable get",
		"variable delete",
		"codespace create",
		"release download",
		"run download",
	}
	partial := []string{
		"repo license list",
		"repo gitignore list",
		"cache list",
		"secret list",
		"variable list",
		"org list",
		"gist list",
		"codespace list",
		"gpg-key list",
		"ssh-key list",
		"agent-task list",
	}
	retained := map[string]string{
		"label clone":        "composite",
		"variable set":       "conditional composite",
		"gist create":        "filename-keyed",
		"skill list":         "local filesystem",
		"issue develop":      "git",
		"pr checkout":        "git",
		"repo clone":         "git",
		"repo sync":          "git",
		"repo set-default":   "config",
		"release upload":     "binary-upload",
		"release verify":     "cryptographic verification",
		"run watch":          "polling",
		"codespace ssh":      "SSH",
		"auth login":         "credential bootstrap",
		"auth status":        "credential store",
		"auth token":         "prints a credential",
		"config get":         "config executor",
		"config set":         "config executor",
		"browse":             "browser-launch",
		"alias list":         "local aliases",
		"extension list":     "local extensions",
		"completion":         "shell completion",
		"api":                "generic authenticated",
		"attestation verify": "signature",
		"copilot":            "external interactive extension",
	}
	if got := len(implemented) + len(partial) + len(retained); got != 50 {
		t.Fatalf("verdict inventory = %d commands, want 50", got)
	}

	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}
	commands := make(map[string]connectors.CommandSurfaceCommand, len(provider.CommandSurface().Commands))
	for _, command := range provider.CommandSurface().Commands {
		commands[command.Path] = command
	}

	seen := make(map[string]struct{}, 50)
	for _, path := range implemented {
		if _, duplicate := seen[path]; duplicate {
			t.Fatalf("duplicate verdict for %q", path)
		}
		seen[path] = struct{}{}
		command, ok := commands[path]
		if !ok {
			t.Fatalf("implemented verdict command %q is not declared", path)
		}
		if command.Availability != "implemented" {
			t.Errorf("github %q availability = %q, want implemented", path, command.Availability)
			continue
		}
		if len(command.APISurface) != 1 {
			t.Errorf("github %q api_surface endpoints = %d, want exactly one fixed endpoint", path, len(command.APISurface))
			continue
		}
		if err := Preflight(connector, strings.Fields(path)); err != nil {
			t.Errorf("github %q does not pass runtime preflight: %v", path, err)
		}
	}

	for _, path := range partial {
		if _, duplicate := seen[path]; duplicate {
			t.Fatalf("duplicate verdict for %q", path)
		}
		seen[path] = struct{}{}
		command, ok := commands[path]
		if !ok {
			t.Fatalf("partial verdict command %q is not declared", path)
		}
		if command.Availability != "partial" {
			t.Errorf("github %q availability = %q, want partial without a declaration-owned operation", path, command.Availability)
			continue
		}
		if err := Preflight(connector, strings.Fields(path)); err == nil {
			t.Errorf("github %q passes runtime preflight despite having no declaration-owned operation", path)
		}
	}

	for path, evidence := range retained {
		if _, duplicate := seen[path]; duplicate {
			t.Fatalf("duplicate verdict for %q", path)
		}
		seen[path] = struct{}{}
		command, ok := commands[path]
		if !ok {
			t.Fatalf("retained verdict command %q is not declared", path)
		}
		if command.Availability != "unsupported_api" && command.Availability != "unsupported_local" {
			t.Errorf("github %q availability = %q, want retained unsupported_api/unsupported_local", path, command.Availability)
		}
		if !strings.Contains(strings.ToLower(command.Notes), strings.ToLower(evidence)) {
			t.Errorf("github %q notes = %q, want concrete evidence containing %q", path, command.Notes, evidence)
		}
		if err := Preflight(connector, strings.Fields(path)); err == nil {
			t.Errorf("github %q passes runtime preflight despite retained verdict", path)
		}
	}
	if len(seen) != 50 {
		t.Fatalf("unique verdict inventory = %d commands, want 50", len(seen))
	}
}

// TestGitHubDeclaredParityProviderContracts protects the provider shapes that
// live certification exposed while promoting issue #4015 commands. Provider
// identifiers remain strings across plan persistence (JSON numbers lose exact
// integer spelling), variable writes accept only the declared body fields, and
// release downloads request bytes rather than GitHub's JSON asset metadata.
func TestGitHubDeclaredParityProviderContracts(t *testing.T) {
	t.Parallel()

	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load github bundle: %v", err)
	}
	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}
	releaseDownload, ok := operations["github.release.download_assets"]
	if !ok {
		t.Fatal("github release download operation is not declared")
	}
	if releaseDownload.Binary == nil {
		t.Fatal("github release download has no binary declaration")
	}
	if got := releaseDownload.Binary.Accept; got != "application/octet-stream" {
		t.Errorf("github release download Accept = %q, want application/octet-stream", got)
	}
	if releaseDownload.Binary.Redirect == nil {
		t.Fatal("github release download has no redirect declaration")
	}
	if releaseDownload.Binary.Redirect.MaxHops != 1 || !releaseDownload.Binary.Redirect.AllowSameOrigin {
		t.Errorf("github release download redirect = %+v, want one same-origin hop", releaseDownload.Binary.Redirect)
	}
	if got := strings.Join(releaseDownload.Binary.Redirect.AllowedHosts, ","); got != "release-assets.githubusercontent.com" {
		t.Errorf("github release download allowed_hosts = %q, want release-assets.githubusercontent.com", got)
	}
	artifactDownload, ok := operations["github.actions_artifacts_artifact_id_archive_format"]
	if !ok {
		t.Fatal("github Actions artifact download operation is not declared")
	}
	if artifactDownload.Binary == nil || artifactDownload.Binary.Redirect == nil {
		t.Fatal("github Actions artifact download must declare a redirect policy")
	}
	if artifactDownload.Binary.Redirect.MaxHops != 1 || !artifactDownload.Binary.Redirect.AllowSameOrigin {
		t.Errorf("github Actions artifact download redirect = %+v, want one same-origin hop", artifactDownload.Binary.Redirect)
	}
	if got := strings.Join(artifactDownload.Binary.Redirect.AllowedHosts, ","); got != "pipelines.actions.githubusercontent.com" {
		t.Errorf("github Actions artifact download allowed_hosts = %q, want pipelines.actions.githubusercontent.com", got)
	}
	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}

	for _, tt := range []struct {
		name  string
		field string
		types []string
	}{
		{name: "autolinks_autolink_id", field: "autolink_id", types: []string{"integer"}},
		{name: "actions_workflows_workflow_id_disable", field: "workflow_id", types: []string{"integer", "string"}},
		{name: "actions_workflows_workflow_id_enable", field: "workflow_id", types: []string{"integer", "string"}},
	} {
		action, ok := actions[tt.name]
		if !ok {
			t.Fatalf("github write action %q is not declared", tt.name)
		}
		var schema struct {
			Properties map[string]struct {
				Type json.RawMessage `json:"type"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("decode github write action %q schema: %v", tt.name, err)
		}
		property, ok := schema.Properties[tt.field]
		if !ok {
			t.Errorf("github write action %q has no %q property", tt.name, tt.field)
			continue
		}
		gotTypes, err := githubDeclaredSchemaTypes(property.Type)
		if err != nil {
			t.Errorf("github write action %q field %q has invalid type %s: %v", tt.name, tt.field, property.Type, err)
			continue
		}
		if !reflect.DeepEqual(gotTypes, tt.types) {
			t.Errorf("github write action %q field %q types = %#v, want %#v", tt.name, tt.field, gotTypes, tt.types)
		}
	}

	for _, tt := range []struct {
		name     string
		required string
	}{
		{name: "actions_variables2", required: "name,value"},
		// The provider requires `name` structurally in the PATCH path while
		// both body members are optional. Keep those declaration-owned
		// requirements distinct instead of promoting the body contract.
		{name: "actions_variables_name3", required: "name"},
	} {
		action, ok := actions[tt.name]
		if !ok {
			t.Fatalf("github write action %q is not declared", tt.name)
		}
		var schema struct {
			AdditionalProperties *bool                      `json:"additionalProperties"`
			Required             []string                   `json:"required"`
			Properties           map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("decode github write action %q schema: %v", tt.name, err)
		}
		if schema.AdditionalProperties == nil || *schema.AdditionalProperties {
			t.Errorf("github write action %q schema must reject undeclared fields", tt.name)
		}
		if strings.Join(schema.Required, ",") != tt.required {
			t.Errorf("github write action %q required = %v, want %q", tt.name, schema.Required, tt.required)
		}
		if len(schema.Properties) != 2 || schema.Properties["name"] == nil || schema.Properties["value"] == nil {
			t.Errorf("github write action %q properties = %v, want only name and value", tt.name, schema.Properties)
		}
	}
}

func githubDeclaredSchemaTypes(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, nil
	}
	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err != nil {
		return nil, err
	}
	return multiple, nil
}
