package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
	"strings"
	"testing"
	"testing/fstest"
)

type coordinateFS171 struct {
	fs.FS
	file string
	data []byte
}

func (f coordinateFS171) Open(name string) (fs.File, error) {
	if name == "github/"+f.file {
		return fstest.MapFS{name: &fstest.MapFile{Data: f.data}}.Open(name)
	}
	return f.FS.Open(name)
}
func fixture171(t *testing.T, file string) []byte {
	t.Helper()
	raw, err := os.ReadFile("../connectors/engine/testdata/diagnostic_coordinates_171.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string]json.RawMessage
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	return cases[strings.TrimSuffix(file, ".json")]
}

var coordinateCases171 = []struct{ file, field, code, reason string }{
	{"writes.json", "/actions/0/multipart/parts/0/allowed_media_types", "multipart_media_types_type_invalid", "allowed_media_types is only meaningful on a file part"},
	{"operations.json", "/operations/0/rest/multipart/parts/0/allowed_media_types", "multipart_media_types_type_invalid", "allowed_media_types is only meaningful on a file part"},
	{"sync_transport.json", "/destination_transport/apply_strategies/0/strategy", "transport_strategy_mode_conflict", "change_apply strategy is only valid for change_capture mode"},
}

func TestAppDiagnosticCoordinates171(t *testing.T) {
	for _, tc := range coordinateCases171 {
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
				fixture := fixture171(t, tc.file)
				loads := 0
				store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: 1, Bytes: selected.Bytes}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
					loads++
					_, failure := engine.Load(coordinateFS171{FS: defs.FS, file: tc.file, data: fixture}, entry.Connector)
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
