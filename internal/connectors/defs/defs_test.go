package defs

import (
	"context"
	"errors"
	"io/fs"
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

func TestIntercomVariantWriteSchemasValidateTargets(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	var intercom engine.Bundle
	for _, bundle := range bundles {
		if bundle.Name == "intercom" {
			intercom = bundle
			break
		}
	}
	if intercom.Name == "" {
		t.Fatal("LoadAll(FS) missing intercom bundle")
	}

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
