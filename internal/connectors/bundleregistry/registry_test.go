package bundleregistry

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
)

func TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	if len(bundles) != 548 {
		t.Fatalf("bundle count = %d, want 548", len(bundles))
	}

	registry := New()

	for _, b := range bundles {
		if _, ok := registry.Get(b.Name); !ok {
			t.Fatalf("registry missing bundle connector %q", b.Name)
		}
	}
	for _, legacySlug := range []string{"source-github", "destination-postgres"} {
		if _, ok := registry.Get(legacySlug); ok {
			t.Fatalf("registry contains legacy slug %q; want bare names only", legacySlug)
		}
	}

	akeneo, ok := registry.Get("akeneo")
	if !ok {
		t.Fatal("registry missing akeneo")
	}
	if _, ok := akeneo.(*engine.Connector); !ok {
		t.Fatalf("akeneo registry type = %T, want engine-backed connector", akeneo)
	}
	if engine.HooksFor("github") == nil {
		t.Fatal("hookset side effects were not loaded; github hook is missing")
	}

	postgresConnector, ok := registry.Get("postgres")
	if !ok {
		t.Fatal("registry missing postgres")
	}
	if _, ok := postgresConnector.(nativepostgres.Connector); !ok {
		t.Fatalf("postgres registry type = %T, want Tier-3 native override", postgresConnector)
	}
}

func TestRegistryCatalogEntriesComeFromDefinitions(t *testing.T) {
	registry := New()
	entries := registry.CatalogEntries()
	if len(entries) < 548 {
		t.Fatalf("CatalogEntries() count = %d, want at least 548 bundle/native definitions", len(entries))
	}

	var github connectors.Definition
	foundGithub := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name, "source-") || strings.HasPrefix(entry.Name, "destination-") {
			t.Fatalf("CatalogEntries() contains legacy slug %q", entry.Name)
		}
		if entry.Name == "github" {
			github = entry
			foundGithub = true
		}
	}
	if !foundGithub {
		t.Fatal("CatalogEntries() missing github")
	}
	if !github.Capabilities.Read || len(github.Streams) == 0 {
		t.Fatalf("github definition not sourced from bundle metadata/schemas: %+v", github)
	}
	if len(github.WriteActions) == 0 || !github.Capabilities.Write {
		t.Fatalf("github definition missing bundle write capability/actions: %+v", github)
	}
}

func TestWhatsAppSendAndMediaWriteSchemasRequireProviderFields(t *testing.T) {
	registry := New()
	connector, ok := registry.Get("whatsapp")
	if !ok {
		t.Fatal("registry missing whatsapp")
	}
	validator, ok := connector.(connectors.WriteValidator)
	if !ok {
		t.Fatalf("whatsapp connector type = %T, want WriteValidator", connector)
	}
	bundle, err := engine.Load(defs.FS, "whatsapp")
	if err != nil {
		t.Fatalf("Load(whatsapp): %v", err)
	}

	tests := []struct {
		action  string
		typ     string
		payload string
		value   any
	}{
		{"send_text_message", "text", "text", map[string]any{"body": "appointment reminder"}},
		{"send_image_message", "image", "image", map[string]any{"link": "https://example.test/image.jpg"}},
		{"send_audio_message", "audio", "audio", map[string]any{"id": "media_audio"}},
		{"send_video_message", "video", "video", map[string]any{"id": "media_video"}},
		{"send_document_message", "document", "document", map[string]any{"id": "media_document"}},
		{"send_sticker_message", "sticker", "sticker", map[string]any{"id": "media_sticker"}},
		{"send_location_message", "location", "location", map[string]any{"latitude": 12.34, "longitude": 56.78}},
		{"send_contacts_message", "contacts", "contacts", []any{map[string]any{"name": map[string]any{"formatted_name": "Clinic"}}}},
		{"send_interactive_message", "interactive", "interactive", map[string]any{"type": "button", "action": map[string]any{"buttons": []any{map[string]any{"type": "reply"}}}}},
		{"send_template_message", "template", "template", map[string]any{"name": "appointment_reminder", "language": map[string]any{"code": "en_US"}}},
		{"send_reaction_message", "reaction", "reaction", map[string]any{"message_id": "wamid.TEST", "emoji": "ok"}},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: tt.action}, []connectors.Record{{"to": "+15551234567"}})
			if err == nil {
				t.Fatalf("ValidateWrite(%s) accepted record with only to", tt.action)
			}
			record := connectors.Record{
				"messaging_product": "whatsapp",
				"to":                "+15551234567",
				"type":              tt.typ,
				tt.payload:          tt.value,
			}
			if err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: tt.action}, []connectors.Record{record}); err != nil {
				t.Fatalf("ValidateWrite(%s) valid record: %v", tt.action, err)
			}
			record["raw_graph_body"] = map[string]any{"unsafe": true}
			if err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: tt.action}, []connectors.Record{record}); err == nil {
				t.Fatalf("ValidateWrite(%s) accepted an undeclared top-level field", tt.action)
			}
			action, ok := findBundleWriteAction(bundle, tt.action)
			if !ok {
				t.Fatalf("bundle missing write action %s", tt.action)
			}
			wantBodyFields := []string{"messaging_product", "to", "type", tt.payload}
			if !reflect.DeepEqual(action.BodyFields, wantBodyFields) {
				t.Fatalf("%s body_fields = %v, want %v", tt.action, action.BodyFields, wantBodyFields)
			}
		})
	}

	err = validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "upload_media"}, []connectors.Record{{"media_file": "media/patient.pdf", "type": "application/pdf"}})
	if err == nil {
		t.Fatal("ValidateWrite(upload_media) accepted missing messaging_product")
	}
	if err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "upload_media"}, []connectors.Record{{
		"messaging_product": "whatsapp",
		"media_file":        "media/patient.pdf",
		"type":              "application/pdf",
	}}); err != nil {
		t.Fatalf("ValidateWrite(upload_media) valid record: %v", err)
	}
}

func findBundleWriteAction(bundle engine.Bundle, name string) (engine.WriteAction, bool) {
	for _, action := range bundle.Writes {
		if action.Name == name {
			return action, true
		}
	}
	return engine.WriteAction{}, false
}

func TestGitHubGuideIncludesCLISurfaceHelp(t *testing.T) {
	registry := New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatalf("github connector not found")
	}

	manual := connectors.RenderConnectorManual(connector)
	for _, want := range []string{
		"COMMAND SURFACE",
		"Usage: pm github <command> <subcommand> [flags]",
		"Core Commands",
		"issue list - List issues",
		"intent=etl availability=implemented stream=issues",
		"issue create - Create an issue",
		"intent=reverse_etl availability=implemented write=create_issue",
		"approval: reverse ETL writes require plan, preview, approval, execute",
		"unsupported local workflow",
		"--json (boolean): Write machine-readable JSON output.",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("GitHub manual missing %q:\n%s", want, manual)
		}
	}
}
