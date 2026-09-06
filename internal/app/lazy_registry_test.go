package app

import (
	"context"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
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
