package defs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProductionEmbedLoadsRuntimeBundles(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	if len(bundles) == 0 {
		t.Fatal("LoadAll(FS) returned zero bundles")
	}

	var github *engine.Bundle
	for i := range bundles {
		if bundles[i].Name == "github" {
			github = &bundles[i]
			break
		}
	}
	if github == nil {
		t.Fatal("LoadAll(FS) missing github bundle")
	}
	if github.Metadata.Name != "github" {
		t.Fatalf("github metadata name = %q", github.Metadata.Name)
	}
	if len(github.Streams) == 0 {
		t.Fatal("github bundle has zero streams")
	}
	if github.Docs == "" {
		t.Fatal("github bundle docs are empty")
	}
	if github.Surface != nil {
		t.Fatal("production embed should not include api_surface.json")
	}
	if github.Fixtures != nil {
		t.Fatal("production embed should not include fixtures")
	}
}

func TestProductionEmbedExcludesConformanceArtifacts(t *testing.T) {
	for _, path := range []string{"github/api_surface.json", "github/fixtures"} {
		if _, err := fs.Stat(FS, path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("fs.Stat(%q) err = %v, want fs.ErrNotExist", path, err)
		}
	}
}

func loadIntercomBundle(t *testing.T) engine.Bundle {
	t.Helper()

	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	for _, bundle := range bundles {
		if bundle.Name == "intercom" {
			return bundle
		}
	}
	t.Fatal("LoadAll(FS) missing intercom bundle")
	return engine.Bundle{}
}

func TestIntercomVariantWriteSchemasValidateTargets(t *testing.T) {
	intercom := loadIntercomBundle(t)

	tests := []struct {
		name    string
		action  string
		record  connectors.Record
		wantErr bool
	}{
		{
			name:    "redact conversation part rejects source-only target",
			action:  "redact_conversation",
			record:  connectors.Record{"type": "conversation_part", "conversation_id": "conversation", "source_id": "source"},
			wantErr: true,
		},
		{
			name:   "redact source accepts source target",
			action: "redact_conversation",
			record: connectors.Record{"type": "source", "conversation_id": "conversation", "source_id": "source"},
		},
		{
			name:    "manage assignment requires assignee",
			action:  "manage_conversation",
			record:  connectors.Record{"conversation_id": "conversation", "message_type": "assignment", "type": "admin", "admin_id": "admin"},
			wantErr: true,
		},
		{
			name:    "reply conversation user requires identity",
			action:  "reply_conversation",
			record:  connectors.Record{"conversation_id": "conversation", "message_type": "comment", "type": "user", "body": "hello"},
			wantErr: true,
		},
		{
			name:    "create options data attribute requires options",
			action:  "create_data_attribute",
			record:  connectors.Record{"name": "size", "model": "contact", "data_type": "options"},
			wantErr: true,
		},
		{
			name:    "create contact requires identity",
			action:  "create_contact",
			record:  connectors.Record{"name": "Ada"},
			wantErr: true,
		},
		{
			name:   "create message email rejects in-app body",
			action: "create_message",
			record: connectors.Record{
				"message_type": "email",
				"body":         "hello",
				"from":         map[string]any{"type": "admin", "id": 394051},
				"to":           map[string]any{"type": "user", "id": "contact_id_fixture"},
			},
			wantErr: true,
		},
		{
			name:   "create message accepts whatsapp components",
			action: "create_message",
			record: connectors.Record{
				"message_type": "whatsapp",
				"template":     "keep_live",
				"components": []any{
					map[string]any{
						"type": "BODY",
						"parameters": []any{
							map[string]any{"type": "text", "text": "Username 123"},
						},
					},
				},
				"from": map[string]any{"type": "admin", "id": 394051},
				"to":   map[string]any{"type": "user", "id": "contact_id_fixture"},
			},
		},
		{
			name:   "create tag accepts company variant",
			action: "create_tag",
			record: connectors.Record{
				"name":      "vip",
				"companies": []any{map[string]any{"id": "company_id_fixture"}},
			},
		},
		{
			name:   "create tag accepts company untag variant",
			action: "create_tag",
			record: connectors.Record{
				"name":      "vip",
				"companies": []any{map[string]any{"id": "company_id_fixture", "untag": true}},
			},
		},
		{
			name:   "create tag accepts user variant",
			action: "create_tag",
			record: connectors.Record{
				"name":  "vip",
				"users": []any{map[string]any{"id": "user_id_fixture"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateWrite(context.Background(), intercom, connectors.WriteRequest{Action: tt.action}, []connectors.Record{tt.record})
			if tt.wantErr && err == nil {
				t.Fatal("ValidateWrite() error = nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateWrite() error = %v", err)
			}
		})
	}
}

func TestIntercomCreateMessageWriteIncludesWhatsAppComponents(t *testing.T) {
	intercom := loadIntercomBundle(t)
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/messages" {
			t.Errorf("path = %s, want /messages", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll(body): %v", err)
			return
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Errorf("Unmarshal(body): %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	record := connectors.Record{
		"message_type": "whatsapp",
		"template":     "keep_live",
		"components": []any{
			map[string]any{
				"type": "BODY",
				"parameters": []any{
					map[string]any{"type": "text", "text": "Username 123"},
				},
			},
		},
		"from": map[string]any{"type": "admin", "id": 394051},
		"to":   map[string]any{"type": "user", "id": "contact_id_fixture"},
	}

	result, err := engine.Write(context.Background(), intercom, connectors.WriteRequest{
		Action: "create_message",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"access_token": "fixture_token_placeholder"},
		},
	}, []connectors.Record{record}, nil)
	if err != nil {
		t.Fatalf("Write(): %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("Write() result = %+v, want 1 written and 0 failed", result)
	}
	components, ok := gotBody["components"].([]any)
	if !ok || len(components) != 1 {
		t.Fatalf("components = %#v, want one component", gotBody["components"])
	}
	component, ok := components[0].(map[string]any)
	if !ok {
		t.Fatalf("components[0] = %#v, want object", components[0])
	}
	parameters, ok := component["parameters"].([]any)
	if !ok || len(parameters) != 1 {
		t.Fatalf("parameters = %#v, want one parameter", component["parameters"])
	}
}
