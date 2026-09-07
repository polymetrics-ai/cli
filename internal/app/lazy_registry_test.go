package app

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestOpenWithLazyRegistryDoesNotResolveMetadataEntries(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}

	var loads atomic.Int32
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{
		{Name: "github", DisplayName: "GitHub", IntegrationType: "api"},
		{Name: "gitlab", DisplayName: "GitLab", IntegrationType: "api"},
	}, func(_ context.Context, name string) (connectors.Connector, error) {
		loads.Add(1)
		return engine.New(engine.Bundle{
			Name: name,
			Metadata: engine.Metadata{
				Name:            name,
				DisplayName:     name,
				IntegrationType: "api",
			},
		}, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	instance, err := openWithRegistry(root, false, func() (*connectors.Registry, error) {
		return registry, nil
	})
	if err != nil {
		t.Fatalf("openWithRegistry(): %v", err)
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("openWithRegistry() resolved %d metadata entries, want 0", got)
	}
	if got := instance.Connectors(); len(got) != 2 {
		t.Fatalf("Connectors() returned %d entries, want 2", len(got))
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("Connectors() resolved %d metadata entries, want 0", got)
	}
}

// Override one actual embedded artifact; all other bundle files remain real.
type malformedSelectedBundleFS160 struct{ fs.FS }

func (f malformedSelectedBundleFS160) Open(name string) (fs.File, error) {
	if name == "github/operations.json" {
		return fstest.MapFS{name: &fstest.MapFile{Data: []byte(`{"operations":[`)}}.Open(name)
	}
	return f.FS.Open(name)
}

func TestPlanConnectorCommandPreservesSelectedDataError160(t *testing.T) {
	valid, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatal(err)
	}
	good := connectors.NewEmptyRegistry()
	if err := good.Register(engine.New(valid, nil)); err != nil {
		t.Fatal(err)
	}
	control := &App{registry: good}
	_, _, err = control.PlanConnectorCommand(t.Context(), PlanConnectorCommandRequest{Connector: "github", Path: []string{"label", "delete"}})
	if err == nil || !strings.Contains(err.Error(), "missing required flag --name") {
		t.Fatalf("valid selected bundle did not reach required-input preflight: %v", err)
	}
	sentinel := errors.New("selected-bundle-sentinel")
	var observed error
	loads := 0
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(ctx context.Context, name string) (connectors.Connector, error) {
		loads++
		_, failure := engine.Load(malformedSelectedBundleFS160{defs.FS}, name)
		if failure == nil || !strings.Contains(failure.Error(), "operations.json") {
			t.Fatalf("malformed actual loader control: %v", failure)
		}
		observed = errors.Join(sentinel, failure)
		return nil, observed
	})
	if err != nil {
		t.Fatal(err)
	}
	instance := &App{registry: registry} // nil vault/store: selection must precede them.
	_, _, err = instance.PlanConnectorCommand(t.Context(), PlanConnectorCommandRequest{Connector: "github", Path: []string{"label", "delete"}})
	if loads != 1 || !errors.Is(err, sentinel) || !errors.Is(err, observed) {
		t.Fatalf("selected real bundle error erased before credential boundary: loads=%d got=%v want=%v", loads, err, observed)
	}
}

func TestPlanConnectorCommandSelectionBoundaries160(t *testing.T) {
	valid, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatal(err)
	}
	loads := 0
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(context.Context, string) (connectors.Connector, error) {
		loads++
		return engine.New(valid, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	instance := &App{registry: registry}
	_, _, err = instance.PlanConnectorCommand(t.Context(), PlanConnectorCommandRequest{Connector: "absent-connector", Path: []string{"label", "delete"}})
	if err == nil || err.Error() != `connector "absent-connector" not found` || loads != 0 {
		t.Fatalf("unknown selection changed: %v loads=%d", err, loads)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, err = instance.PlanConnectorCommand(ctx, PlanConnectorCommandRequest{Connector: "github", Path: []string{"label", "delete"}})
	if !errors.Is(err, context.Canceled) || loads != 0 {
		t.Fatalf("cancellation crossed selected construction: %v loads=%d", err, loads)
	}
	_, _, err = instance.PlanConnectorCommand(t.Context(), PlanConnectorCommandRequest{Connector: "github", Path: []string{"label", "delete"}})
	if err == nil || !strings.Contains(err.Error(), "missing required flag --name") || loads != 1 {
		t.Fatalf("valid selection did not reach preflight: %v loads=%d", err, loads)
	}
}
