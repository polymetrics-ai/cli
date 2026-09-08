package manifeststore

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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
	var loaderErr error
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: entry.Bytes}, func(context.Context, manifestindex.Entry) (LoadedBundle, error) {
		_, err := engine.Load(malformedRateFS165{defs.FS}, "github")
		loaderErr = err
		return LoadedBundle{}, err
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err = store.Acquire(ctx, "github")
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("loader did not complete within owned bound")
	}
	var original *engine.BundleDiagnosticError
	if !errors.As(loaderErr, &original) {
		t.Fatalf("actual loader diagnostic missing: %v", loaderErr)
	}
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
	var loaderErr error
	store, err := NewBundleStore(bundleIndex(t, entry), Limits{Entries: 1, Bytes: entry.Bytes}, func(context.Context, manifestindex.Entry) (LoadedBundle, error) {
		_, e := engine.Load(malformedRateFS165{defs.FS}, "github")
		loaderErr = e
		return LoadedBundle{}, errors.Join(e, companion, fs.ErrNotExist)
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err = store.Acquire(ctx, "github")
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("loader did not complete within owned bound")
	}
	var original *engine.BundleDiagnosticError
	if !errors.As(loaderErr, &original) {
		t.Fatalf("actual loader diagnostic missing: %v", loaderErr)
	}
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

// Unexpected loader outcomes must reach the test goroutine without Goexit
// stranding a pending acquisition. These controls do not claim a product RED.
func TestDiagnosticLoaderUnexpectedOutcomes171(t *testing.T) {
	for _, success := range []bool{false, true} {
		name := "plain_error"
		if success {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			entry := manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1}
			cause := errors.New("unexpected non-diagnostic loader failure")
			store, err := NewBundleStore(bundleIndex(t, entry), Limits{Entries: 1, Bytes: 1}, func(_ context.Context, selected manifestindex.Entry) (LoadedBundle, error) {
				if success {
					return loadedFor(selected), nil
				}
				return LoadedBundle{}, cause
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			handle, err := store.Acquire(ctx, "alpha")
			if errors.Is(err, context.DeadlineExceeded) {
				t.Fatal("unexpected loader outcome stranded acquisition")
			}
			var d *engine.BundleDiagnosticError
			if errors.As(err, &d) {
				t.Fatal("unexpected loader outcome fabricated diagnostic")
			}
			if success {
				if err != nil || handle == nil {
					t.Fatalf("successful loader: %v", err)
				}
				handle.Release()
			} else if !errors.Is(err, cause) {
				t.Fatalf("original plain error lost: %v", err)
			}
		})
	}
}
