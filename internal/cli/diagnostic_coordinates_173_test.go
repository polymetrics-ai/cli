package cli

import (
	"bytes"
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

func TestCLIDiagnosticCoordinates173(t *testing.T) {
	for _, tc := range coordinateCases173 {
		for _, asJSON := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/json=%t", tc.file, asJSON), func(t *testing.T) {
				var selected manifestindex.Entry
				for _, entry := range manifestindex.GeneratedEntries() {
					if entry.Connector == "github" {
						selected = entry
						break
					}
				}
				if selected.Connector != "github" {
					t.Fatal("missing real generated GitHub entry")
				}
				selected.Generation = "selected-cli-generation-168"
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
				companion := errors.New("private-companion-168")
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
				var observed error
				registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(ctx context.Context, name string) (connectors.Connector, error) {
					handle, failure := store.Acquire(ctx, name)
					if failure != nil {
						observed = failure
						return nil, failure
					}
					defer handle.Release()
					return engine.New(*handle.Bundle(), nil), nil
				})
				if err != nil {
					t.Fatal(err)
				}
				var stdout, stderr bytes.Buffer
				args := []string{"connectors", "inspect", "github"}
				if asJSON {
					args = append(args, "--json")
				}
				code := runWithPreflightRegistryAndApprovalReader(args, &stdout, &stderr, registry, nil)
				var diagnostic *engine.BundleDiagnosticError

				if code != 1 || loads != 1 || !errors.As(observed, &diagnostic) || !errors.Is(observed, companion) {
					t.Fatalf("selected store graph/counter failed: code=%d loads=%d error=%v", code, loads, observed)
				}
				if diagnostic.Connector != selected.Connector || diagnostic.Generation != selected.Generation || diagnostic.Digest != selected.Digest || diagnostic.File != tc.file || diagnostic.Field != tc.field || diagnostic.ReasonCode != tc.code || diagnostic.Reason != tc.reason {
					t.Fatalf("store identity mismatch: %+v", diagnostic)
				}
				if stderr.String() != "error: "+diagnostic.Error()+"\n" || strings.Contains(stdout.String()+stderr.String(), "private-") {
					t.Fatalf("unsafe or misbound text: stdout=%q stderr=%q", stdout.String(), stderr.String())
				}
				if asJSON {
					var response struct {
						Error struct {
							Code   string
							Bundle map[string]any
						}
					}
					if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
						t.Fatal(err)
					}
					b := response.Error.Bundle
					if response.Error.Code != "internal_error" || b["connector"] != selected.Connector || b["generation"] != selected.Generation || b["digest"] != selected.Digest || b["file"] != tc.file || b["field"] != tc.field || b["reason_code"] != tc.code || b["reason"] != tc.reason || len(b) != 7 {
						t.Fatalf("wrong selected public tuple: %+v", response)
					}
				} else if stdout.Len() != 0 {
					t.Fatalf("plain failure emitted successful output: %q", stdout.String())
				}
			})
		}
	}

}
