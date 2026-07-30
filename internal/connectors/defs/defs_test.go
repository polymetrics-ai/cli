package defs

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"

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

func TestLinearWriteRequiredFieldsAreNonNullable(t *testing.T) {
	linear := mustProductionBundle(t, "linear")

	for _, action := range linear.Writes {
		if action.GraphQL == nil || len(action.RecordSchema) == 0 {
			continue
		}
		var schema struct {
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type any `json:"type"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("%s record_schema: %v", action.Name, err)
		}
		for _, field := range schema.Required {
			property, ok := schema.Properties[field]
			if !ok {
				t.Fatalf("%s required field %q missing from properties", action.Name, field)
			}
			switch typ := property.Type.(type) {
			case string:
				if typ == "null" {
					t.Fatalf("%s required field %q is nullable", action.Name, field)
				}
			case []any:
				for _, entry := range typ {
					if entry == "null" {
						t.Fatalf("%s required field %q is nullable", action.Name, field)
					}
				}
			default:
				t.Fatalf("%s required field %q has unsupported type shape %T", action.Name, field, property.Type)
			}
		}
	}
}

func TestLinearStateReducingWritesRequireDestructiveConfirmation(t *testing.T) {
	linear := mustProductionBundle(t, "linear")

	for _, action := range linear.Writes {
		if !linearStateReducingWrite(action.Name) {
			continue
		}
		if action.Confirm != "destructive" {
			t.Fatalf("%s confirm = %q, want destructive", action.Name, action.Confirm)
		}
		if !strings.Contains(action.Risk, "destructive") {
			t.Fatalf("%s risk = %q, want destructive risk text", action.Name, action.Risk)
		}
	}
}

func mustProductionBundle(t *testing.T, name string) *engine.Bundle {
	t.Helper()

	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	for i := range bundles {
		if bundles[i].Name == name {
			return &bundles[i]
		}
	}
	t.Fatalf("LoadAll(FS) missing %s bundle", name)
	return nil
}

func linearStateReducingWrite(name string) bool {
	for _, token := range []string{
		"retire",
		"disable",
		"unlink",
		"revoke",
		"delete",
		"archive",
		"remove",
		"rotate",
		"cancel",
		"disconnect",
		"unsync",
		"logout",
		"leave",
	} {
		if name == token || strings.HasPrefix(name, token+"_") || strings.HasSuffix(name, "_"+token) || strings.Contains(name, "_"+token+"_") {
			return true
		}
	}
	return false
}
