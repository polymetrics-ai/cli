package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
	"strings"
	"testing"
	"testing/fstest"
)

type coordinateFS173 struct {
	fs.FS
	file string
	data []byte
}

func (f coordinateFS173) Open(name string) (fs.File, error) {
	if name == "github/"+f.file {
		return fstest.MapFS{name: &fstest.MapFile{Data: f.data}}.Open(name)
	}
	return f.FS.Open(name)
}
func fixture173(t *testing.T, owner string) []byte {
	t.Helper()
	disposition := map[string]any{"reason": "fixture unsupported operation", "target": map[string]any{"source_id": "source-173", "operation_id": "operation-173", "method": "GET", "path": "/../private-sentinel-173"}}
	availability := "unsupported_with_provider_evidence"
	if owner == "foundation_gap" {
		availability = "deferred"
		disposition["id"] = "fixture_gap"
		disposition["component"] = "runtime_executor"
		disposition["evidence"] = "runtime_executor_absent"
	}
	raw, err := json.Marshal(map[string]any{"tagline": "Fixture", "usage": "github", "commands": []any{map[string]any{"path": "widgets list", "summary": "Fixture", "intent": "direct_read", "availability": availability, owner: disposition}}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

var coordinateCases173 = []struct{ file, field, code, reason string }{
	{"cli_surface.json", "/commands/0/foundation_gap/target/path", "command_target_path_segment_invalid", "command target path contains a noncanonical segment"},
	{"cli_surface.json", "/commands/0/unsupported_disposition/target/path", "command_target_path_segment_invalid", "command target path contains a noncanonical segment"},
}

func TestAppDiagnosticCoordinates173(t *testing.T) {
	for _, tc := range coordinateCases173 {
		for _, method := range []string{"plan", "connector"} {
			t.Run(tc.file+"/"+method, func(t *testing.T) {
				companion := errors.New("private-sentinel-167")
				var observed error
				var selected manifestindex.Entry
				for _, entry := range manifestindex.GeneratedEntries() {
					if entry.Connector == "github" {
						selected = entry
						break
					}
				}
				if selected.Connector != "github" {
					t.Fatal("missing generated GitHub entry")
				}
				selected.Generation = "selected-app-generation-168"
				index, err := manifestindex.New([]manifestindex.Entry{selected}, 1)
				if err != nil {
					t.Fatal(err)
				}
				fixture := fixture173(t, strings.Split(tc.field, "/")[3])
				healthy := []byte(strings.ReplaceAll(string(fixture), "/../private-sentinel-173", "/widgets"))
				if _, err := engine.Load(coordinateFS173{FS: defs.FS, file: tc.file, data: healthy}, "github"); err != nil {
					t.Fatalf("healthy target setup: %v", err)
				}
				_, malformed := engine.Load(coordinateFS173{FS: defs.FS, file: tc.file, data: fixture}, "github")
				var reached *engine.BundleDiagnosticError
				if !errors.As(malformed, &reached) || reached.Cause == nil || !strings.Contains(reached.Cause.Error(), "noncanonical path segment") || !strings.Contains(reached.Cause.Error(), "private-sentinel-173") {
					t.Fatalf("semantic cause frontier not reached: %v", malformed)
				}
				loads := 0
				store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: 1, Bytes: selected.Bytes}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
					loads++
					_, failure := engine.Load(coordinateFS173{FS: defs.FS, file: tc.file, data: fixture}, entry.Connector)
					if failure == nil {
						return manifeststore.LoadedBundle{}, errors.New("malformed fixture unexpectedly loaded")
					}
					return manifeststore.LoadedBundle{}, errors.Join(failure, companion)
				})
				if err != nil {
					t.Fatal(err)
				}
				registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(ctx context.Context, name string) (connectors.Connector, error) {
					handle, failure := store.Acquire(ctx, name)
					if failure == nil {
						handle.Release()
						return nil, errors.New("malformed fixture acquired")
					}
					observed = fmt.Errorf("private-prefix-167: %w", failure)
					return nil, observed
				})
				if err != nil {
					t.Fatal(err)
				}
				a := &App{registry: registry}
				if method == "plan" {
					_, _, err = a.PlanConnectorCommand(t.Context(), PlanConnectorCommandRequest{Connector: "github", Path: []string{"label", "delete"}})
				} else {
					_, err = a.Connector("github")
				}
				var d *engine.BundleDiagnosticError

				if !errors.As(err, &d) || !errors.Is(err, companion) || !errors.Is(err, observed) {
					t.Fatalf("selected cause graph erased: %v", err)
				}
				if loads != 1 || d.Connector != selected.Connector || d.Generation != selected.Generation || d.Digest != selected.Digest || d.File != tc.file || d.Field != tc.field || d.ReasonCode != tc.code || d.Reason != tc.reason {
					t.Fatalf("wrong selected App tuple/count: loads=%d diagnostic=%+v", loads, d)
				}
				if !strings.Contains(err.Error(), tc.reason) || strings.Contains(err.Error(), "private-") {
					t.Errorf("unsafe or uninformative App public reason: %v", err)
				}
			})
		}
	}

}
