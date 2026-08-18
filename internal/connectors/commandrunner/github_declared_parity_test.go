package commandrunner

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
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
	}
	retained := map[string]string{
		"pr diff":            "media type",
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
		"run download":       "multiple artifacts",
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
