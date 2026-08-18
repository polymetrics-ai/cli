package commandrunner

import (
	"encoding/json"
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
		"repo license list",
		"repo gitignore list",
		"workflow enable",
		"workflow disable",
		"cache list",
		"secret list",
		"variable list",
		"variable get",
		"variable delete",
		"org list",
		"gist list",
		"codespace list",
		"codespace create",
		"gpg-key list",
		"ssh-key list",
		"agent-task list",
		"release download",
		"run download",
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
	if got := len(implemented) + len(retained); got != 50 {
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
	if got := releaseDownload.Binary.Accept; got != "application/octet-stream" {
		t.Errorf("github release download Accept = %q, want application/octet-stream", got)
	}
	if got := strings.Join(releaseDownload.Binary.AllowedHosts, ","); got != "release-assets.githubusercontent.com" {
		t.Errorf("github release download allowed_hosts = %q, want release-assets.githubusercontent.com", got)
	}
	artifactDownload, ok := operations["github.actions_artifacts_artifact_id_archive_format"]
	if !ok {
		t.Fatal("github Actions artifact download operation is not declared")
	}
	if !artifactDownload.Binary.AllowCrossHost {
		t.Error("github Actions artifact download must allow GitHub's signed cross-host storage redirect")
	}
	actions := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}

	for _, name := range []string{
		"autolinks_autolink_id",
		"actions_workflows_workflow_id_disable",
		"actions_workflows_workflow_id_enable",
	} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("github write action %q is not declared", name)
		}
		var schema struct {
			Properties map[string]struct {
				Type string `json:"type"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("decode github write action %q schema: %v", name, err)
		}
		field := "workflow_id"
		if name == "autolinks_autolink_id" {
			field = "autolink_id"
		}
		if got := schema.Properties[field].Type; got != "string" {
			t.Errorf("github write action %q field %q type = %q, want string", name, field, got)
		}
	}

	for _, name := range []string{"actions_variables2", "actions_variables_name3"} {
		action, ok := actions[name]
		if !ok {
			t.Fatalf("github write action %q is not declared", name)
		}
		var schema struct {
			AdditionalProperties *bool                      `json:"additionalProperties"`
			Required             []string                   `json:"required"`
			Properties           map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("decode github write action %q schema: %v", name, err)
		}
		if schema.AdditionalProperties == nil || *schema.AdditionalProperties {
			t.Errorf("github write action %q schema must reject undeclared fields", name)
		}
		if strings.Join(schema.Required, ",") != "name,value" {
			t.Errorf("github write action %q required = %v, want [name value]", name, schema.Required)
		}
		if len(schema.Properties) != 2 || schema.Properties["name"] == nil || schema.Properties["value"] == nil {
			t.Errorf("github write action %q properties = %v, want only name and value", name, schema.Properties)
		}
	}
}
