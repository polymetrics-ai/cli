package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestStructuredRESTBodyCommandHelpAndManualExposeOnlyDeclaredTypedFields(t *testing.T) {
	command := engine.CLICommand{
		Path:         "widgets configure",
		Summary:      "Configure one widget",
		Intent:       "direct_write",
		Availability: "implemented",
		Operation:    "acme.widgets.configure",
		APISurface:   []engine.CLISurfaceEndpointRef{{Method: http.MethodPost, Path: "/widgets/{id}/configure"}},
		OutputPolicy: "json",
		Flags: []engine.CLIFlag{
			{Name: "id", Type: "string", MapsTo: "path.id", Required: true},
			{Name: "settings", Type: "json", MapsTo: "body.settings", Required: true},
			{Name: "targets", Type: "json", MapsTo: "body.targets", Required: true},
		},
	}
	bundle := engine.Bundle{
		Name:     "acme",
		Metadata: engine.Metadata{Name: "acme", DisplayName: "Acme", Description: "Fixture connector"},
		Operations: []engine.OperationSpec{{
			ID:            command.Operation,
			Kind:          "rest_write",
			Summary:       command.Summary,
			Risk:          "high",
			Approval:      "plan-preview-confirm-execute",
			OutputPolicy:  command.OutputPolicy,
			MutationClass: "update",
			Confirmation:  &engine.ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive},
			REST: &engine.RESTOperationSpec{
				Method:      http.MethodPost,
				Path:        "/widgets/{id}/configure",
				ContentType: "application/json",
				MaxBytes:    1024,
				BodySchema:  json.RawMessage(`{"type":"object","additionalProperties":false,"required":["settings","targets"],"properties":{"settings":{"type":"object","additionalProperties":false,"properties":{}},"targets":{"type":"array","maxItems":2,"items":{"type":"object","additionalProperties":false,"properties":{}}}}}`),
			},
		}},
		CLISurface: &engine.CLISurface{Usage: "pm acme <command> [flags]", Commands: []engine.CLICommand{command}},
	}
	connector := engine.New(bundle, nil)
	surface := connector.CommandSurface()
	detail := renderConnectorCommandDetail("acme", connector, surface, surface.Commands[0])
	manual := connectors.RenderConnectorManual(connector)

	for _, want := range []string{
		"--settings (json) required maps_to=body.settings",
		"--targets (json) required maps_to=body.targets",
		"OPERATION\n  acme.widgets.configure",
	} {
		if !strings.Contains(detail, want) {
			t.Fatalf("command help missing %q:\n%s", want, detail)
		}
	}
	if strings.Contains(detail, "--body") || strings.Contains(manual, "--body") {
		t.Fatalf("structured REST body surface exposed opaque raw-body flag:\nhelp=%s\nmanual=%s", detail, manual)
	}
	for _, want := range []string{"widgets configure", "--settings", "--targets"} {
		if !strings.Contains(manual, want) {
			t.Fatalf("generated manual missing %q:\n%s", want, manual)
		}
	}
}
