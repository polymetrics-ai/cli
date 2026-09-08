package manifeststore

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

type malformedRateFS165 struct{ fs.FS }

func (f malformedRateFS165) Open(name string) (fs.File, error) {
	if name == "github/rate_limits.json" {
		return fstest.MapFS{name: &fstest.MapFile{Data: []byte(`{"schema_version":}`)}}.Open(name)
	}
	return f.FS.Open(name)
}
func TestBundleDiagnosticSelectedGeneration165(t *testing.T) {
	var entry manifestindex.Entry
	for _, candidate := range manifestindex.GeneratedEntries() {
		if candidate.Connector == "github" {
			entry = candidate
			break
		}
	}
	if entry.Connector == "" {
		t.Fatal("actual generated GitHub identity missing")
	}
	entry.Generation = "selected-generation-165"
	index := bundleIndex(t, entry)
	var original *engine.BundleDiagnosticError
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: entry.Bytes}, func(context.Context, manifestindex.Entry) (LoadedBundle, error) {
		_, err := engine.Load(malformedRateFS165{defs.FS}, "github")
		if !errors.As(err, &original) {
			t.Fatalf("actual malformed loader diagnostic missing: %v", err)
		}
		return LoadedBundle{}, err
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Acquire(t.Context(), "github")
	var selected *engine.BundleDiagnosticError
	if !errors.As(err, &selected) {
		t.Fatalf("store lost typed diagnostic: %v", err)
	}
	if selected.Connector != entry.Connector || selected.Generation != entry.Generation || selected.Digest != entry.Digest || selected.File != "rate_limits.json" {
		t.Errorf("selected identity not bound: %+v", selected)
	}
	if selected == original || original.Generation != "embedded-v1" {
		t.Error("store mutated or reused loader's diagnostic")
	}
	if !errors.Is(err, original) {
		t.Error("store lost original diagnostic cause")
	}
}

func TestBundleDiagnosticJoinedSelected168(t *testing.T) {
	var entry manifestindex.Entry
	for _, candidate := range manifestindex.GeneratedEntries() {
		if candidate.Connector == "github" {
			entry = candidate
			break
		}
	}
	entry.Generation = "selected-generation-168"
	companion := errors.New("private-companion-168")
	var original *engine.BundleDiagnosticError
	store, err := NewBundleStore(bundleIndex(t, entry), Limits{Entries: 1, Bytes: entry.Bytes}, func(context.Context, manifestindex.Entry) (LoadedBundle, error) {
		_, e := engine.Load(malformedRateFS165{defs.FS}, "github")
		if !errors.As(e, &original) {
			t.Fatalf("missing actual loader error: %v", e)
		}
		return LoadedBundle{}, errors.Join(e, companion, fs.ErrNotExist)
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Acquire(t.Context(), "github")
	var selected *engine.BundleDiagnosticError
	if !errors.As(err, &selected) || selected.Generation != entry.Generation || selected.Digest != entry.Digest || selected.ReasonCode != "invalid_json" || selected.File != "rate_limits.json" {
		t.Fatalf("selected diagnostic binding lost: %v", err)
	}
	if !errors.Is(err, original) || !errors.Is(err, companion) || !errors.Is(err, fs.ErrNotExist) || original.Generation != "embedded-v1" {
		t.Fatal("selected wrapper lost a branch or mutated original")
	}
	if strings.Contains(selected.Error(), "private-companion-168") {
		t.Fatal("joined sibling leaked")
	}
}
