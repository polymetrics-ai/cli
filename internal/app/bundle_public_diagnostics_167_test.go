package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
)

type malformedRateFS167 struct{ fs.FS }

func (f malformedRateFS167) Open(name string) (fs.File, error) {
	if name == "github/rate_limits.json" {
		return fstest.MapFS{name: &fstest.MapFile{Data: []byte(`{"state":}`)}}.Open(name)
	}
	return f.FS.Open(name)
}
func TestAppPublicBundleDiagnostic167(t *testing.T) {
	for _, method := range []string{"plan", "connector"} {
		t.Run(method, func(t *testing.T) {
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
			loads := 0
			store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: 1, Bytes: selected.Bytes}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
				loads++
				_, failure := engine.Load(malformedRateFS167{defs.FS}, entry.Connector)
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
			var syntax *json.SyntaxError
			if !errors.As(err, &d) || !errors.As(err, &syntax) || !errors.Is(err, companion) || !errors.Is(err, observed) {
				t.Fatalf("selected cause graph erased: %v", err)
			}
			if loads != 1 || d.Connector != selected.Connector || d.Generation != selected.Generation || d.Digest != selected.Digest || d.File != "rate_limits.json" || d.Field != "/" || d.ReasonCode != "invalid_json" || d.Reason != "malformed JSON" {
				t.Fatalf("wrong selected App tuple/count: loads=%d diagnostic=%+v", loads, d)
			}
			if !strings.Contains(err.Error(), "malformed JSON") || strings.Contains(err.Error(), "private-") {
				t.Errorf("unsafe or uninformative App public reason: %v", err)
			}
		})
	}
}

func TestAppCredentialSelectionDiagnostic168(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(t.Context(), AddCredentialRequest{Name: "diagnostic-fixture", Connector: "github", Secrets: map[string]string{"token": "synthetic-fixture-token-168"}}); err != nil {
		t.Fatal(err)
	}
	loads := 0
	companion := errors.New("private-sibling-168")
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(context.Context, string) (connectors.Connector, error) {
		loads++
		_, e := engine.Load(malformedRateFS167{defs.FS}, "github")
		return nil, errors.Join(e, companion)
	})
	if err != nil {
		t.Fatal(err)
	}
	a.registry = registry
	// Credential validation still precedes bundle selection at this frontier.
	if _, _, err := a.ResolveConnectorCredential(t.Context(), "github", "", nil); err == nil || loads != 0 {
		t.Fatalf("missing credential ordering changed: loads=%d err=%v", loads, err)
	}
	_, _, err = a.ResolveConnectorCredential(t.Context(), "github", "diagnostic-fixture", nil)
	var d *engine.BundleDiagnosticError
	var syntax *json.SyntaxError
	if loads != 1 || !errors.As(err, &d) || !errors.As(err, &syntax) || !errors.Is(err, companion) {
		t.Fatalf("credential selector erased selected cause: loads=%d err=%v", loads, err)
	}
	if d.File != "rate_limits.json" || d.ReasonCode != "invalid_json" || strings.Contains(err.Error(), "private-") {
		t.Errorf("credential selector public view: %v", err)
	}
}
