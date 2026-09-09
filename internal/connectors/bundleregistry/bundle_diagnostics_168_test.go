package bundleregistry

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
)

type malformedRateFS168 struct{ fs.FS }

func (f malformedRateFS168) Open(name string) (fs.File, error) {
	if name == "github/rate_limits.json" {
		return fstest.MapFS{name: &fstest.MapFile{Data: []byte(`{"state":}`)}}.Open(name)
	}
	return f.FS.Open(name)
}
func TestConstructionSelectedDiagnostic168(t *testing.T) {
	c, err := NewConstruction()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var entries []manifestindex.Entry
	totalBytes := 0
	for _, entry := range manifestindex.GeneratedEntries() {
		if entry.Connector == "github" || entry.Connector == "gitlab" {
			if entry.Connector == "github" {
				entry.Generation = "selected-generation-168"
			}
			entries = append(entries, entry)
			totalBytes += entry.Bytes
		}
	}
	if len(entries) != 2 {
		t.Fatal("expected actual two manifest entries")
	}
	index, err := manifestindex.New(entries, 2)
	if err != nil {
		t.Fatal(err)
	}
	loads := map[string]int{}
	companion := errors.New("private-companion-168")
	store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: 2, Bytes: totalBytes}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		loads[entry.Connector]++
		bundle, e := engine.Load(malformedRateFS168{defs.FS}, entry.Connector)
		if e != nil {
			return manifeststore.LoadedBundle{}, errors.Join(e, companion)
		}
		return manifeststore.LoadedBundle{Bundle: &bundle, Identity: bundle.Identity}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	baseFactories := c.factories
	factoryCalls := 0
	factories, err := NewExecutorFactories(ExecutorFactory{ID: "api_engine.v1", Construct: func(b engine.Bundle) (connectors.Connector, error) {
		factoryCalls++
		entry, ok := index.Lookup(b.Name)
		if !ok {
			t.Fatal("unexpected factory input")
		}
		return baseFactories.Construct(entry, b)
	}})
	if err != nil {
		t.Fatal(err)
	}
	c.index = index
	c.store = store
	c.factories = factories
	registry, err := c.BuildRegistry()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, m := range registry.List() {
		names = append(names, m.Name)
	}
	if !reflect.DeepEqual(names, []string{"file", "github", "gitlab", "outbox", "sample", "warehouse"}) || len(loads) != 0 || factoryCalls != 0 {
		t.Fatalf("lazy metadata membership/load boundary: %v %v %d", names, loads, factoryCalls)
	}
	_, err = registry.Resolve(t.Context(), "github")
	var d *engine.BundleDiagnosticError
	var syntax *json.SyntaxError
	selected, _ := index.Lookup("github")
	if !errors.As(err, &d) || !errors.As(err, &syntax) || !errors.Is(err, companion) {
		t.Fatalf("selected construction erased error graph: %v", err)
	}
	if d.Generation != selected.Generation || d.Digest != selected.Digest || d.File != "rate_limits.json" || d.Field != "/" || d.ReasonCode != "invalid_json" || factoryCalls != 0 || loads["github"] != 1 {
		t.Fatalf("wrong selected boundary: %v loads=%v factories=%d", err, loads, factoryCalls)
	}
	if strings.Contains(err.Error(), "private-companion-168") {
		t.Fatal("joined companion exposed")
	}
	healthy, err := registry.Resolve(t.Context(), "gitlab")
	if err != nil || healthy == nil || healthy.Name() != "gitlab" || factoryCalls != 1 || loads["gitlab"] != 1 {
		t.Fatalf("healthy sibling construction failed: %v loads=%v factories=%d", err, loads, factoryCalls)
	}
}
