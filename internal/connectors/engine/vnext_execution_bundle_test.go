package engine

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/credential"
)

var vNextReferenceConnectors = []string{"asana", "github", "gitlab"}

var vNextExecutionFiles = map[string]bool{
	"changefeed.json":        true,
	"cli_surface.json":       true,
	"database.json":          true,
	"metadata.json":          true,
	"operations.json":        true,
	"polling_watermark.json": true,
	"rate_limits.json":       true,
	"spec.json":              true,
	"streams.json":           true,
	"sync_transport.json":    true,
	"writes.json":            true,
}

func TestVNextReferenceConnectorsDiscoverFromExecutionJSONOnly(t *testing.T) {
	executionFS := vNextReferenceExecutionFS(t)
	for _, name := range vNextReferenceConnectors {
		t.Run(name, func(t *testing.T) {
			bundle, err := Load(executionFS, name)
			if err != nil {
				t.Fatalf("Load(execution-only %s): %v", name, err)
			}
			connector := New(bundle, nil)
			err = connector.Check(context.Background(), connectors.RuntimeConfig{})
			if err == nil {
				t.Fatal("Check() unexpectedly passed without a credential; want credential-bound refusal before provider I/O")
			}
			var empty *credential.EmptySecretError
			if !errors.As(err, &empty) && !strings.Contains(err.Error(), "auth") && !strings.Contains(err.Error(), "credential") && !strings.Contains(err.Error(), " in secrets") {
				t.Fatalf("Check() error = %v, want credential/auth boundary refusal", err)
			}
		})
	}
}

func TestVNextMalformedExecutionJSONIsConnectorLocal(t *testing.T) {
	executionFS := vNextReferenceExecutionFS(t)
	executionFS["gitlab/operations.json"] = &fstest.MapFile{Data: []byte(`{"operations": [`)}

	bundles, err := LoadAll(executionFS)
	if err == nil || !strings.Contains(err.Error(), "gitlab") || !strings.Contains(err.Error(), "operations.json") {
		t.Fatalf("LoadAll() error = %v, want connector-local malformed GitLab operations diagnostic", err)
	}
	names := make(map[string]bool, len(bundles))
	for _, bundle := range bundles {
		names[bundle.Name] = true
	}
	for _, healthy := range []string{"asana", "github"} {
		if !names[healthy] {
			t.Errorf("healthy connector %q was suppressed by malformed GitLab execution JSON", healthy)
		}
	}
	if names["gitlab"] {
		t.Error("malformed GitLab connector was registered")
	}
}

func TestVNextReferenceBundlesSurfaceExecutionLanes(t *testing.T) {
	executionFS := vNextReferenceExecutionFS(t)
	wantIntents := map[string][]string{
		"asana":  {"binary_upload", "direct_read", "direct_write"},
		"github": {"binary_download", "binary_upload", "direct_read", "direct_write", "etl", "reverse_etl"},
		"gitlab": {"direct_read", "direct_write", "etl", "reverse_etl"},
	}
	for _, name := range vNextReferenceConnectors {
		bundle, err := Load(executionFS, name)
		if err != nil {
			t.Fatalf("Load(execution-only %s): %v", name, err)
		}
		if bundle.CLISurface == nil {
			t.Fatalf("%s execution bundle has no CLI surface", name)
		}
		got := make(map[string]bool)
		for _, command := range bundle.CLISurface.Commands {
			got[command.Intent] = true
		}
		for _, intent := range wantIntents[name] {
			if !got[intent] {
				t.Errorf("%s did not surface %s lane", name, intent)
			}
		}
		if bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil {
			t.Errorf("%s did not surface its configured sync source transport", name)
		}
		if name == "asana" && bundle.SyncTransport != nil && bundle.SyncTransport.Destination != nil {
			t.Error("asana surfaced destination sync despite its explicit unsupported/empty lane")
		}
	}
}

func vNextReferenceExecutionFS(t *testing.T) fstest.MapFS {
	t.Helper()
	source := os.DirFS("../defs")
	out := make(fstest.MapFS)
	for _, connector := range vNextReferenceConnectors {
		err := fs.WalkDir(source, connector, func(filePath string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			base := path.Base(filePath)
			if path.Dir(filePath) != connector+"/schemas" && !vNextExecutionFiles[base] {
				return nil
			}
			data, err := fs.ReadFile(source, filePath)
			if err != nil {
				return err
			}
			out[filePath] = &fstest.MapFile{Data: data}
			return nil
		})
		if err != nil {
			t.Fatalf("copy %s execution files: %v", connector, err)
		}
	}
	return out
}
