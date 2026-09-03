package manifeststore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

func TestBundleStoreRejectsOversizeBeforeLoading(t *testing.T) {
	index := bundleIndex(t, manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 3})
	var calls atomic.Int32
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 2}, func(context.Context, manifestindex.Entry) (*engine.Bundle, error) {
		calls.Add(1)
		return &engine.Bundle{Name: "alpha"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Acquire(context.Background(), "alpha"); !errors.Is(err, ErrBundleOversize) {
		t.Fatalf("Acquire() error = %v, want ErrBundleOversize", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("loader calls = %d, want 0", got)
	}
}

func TestBundleStoreDoesNotEvictHeldBundle(t *testing.T) {
	index := bundleIndex(t,
		manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "a", Executor: "api_engine.v1", Bytes: 1},
		manifestindex.Entry{Connector: "bravo", Generation: "g", Digest: "b", Executor: "api_engine.v1", Bytes: 1},
	)
	var calls atomic.Int32
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(_ context.Context, entry manifestindex.Entry) (*engine.Bundle, error) {
		calls.Add(1)
		return &engine.Bundle{Name: entry.Connector}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Acquire(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Acquire(context.Background(), "bravo"); !errors.Is(err, ErrBundleCapacity) {
		t.Fatalf("Acquire(bravo) error = %v, want ErrBundleCapacity while alpha is held", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("loader calls = %d, want 1 before held-capacity rejection", got)
	}
	first.Release()
	second, err := store.Acquire(context.Background(), "bravo")
	if err != nil {
		t.Fatal(err)
	}
	defer second.Release()
	if second.Bundle().Name != "bravo" {
		t.Fatalf("acquired bundle = %q, want bravo", second.Bundle().Name)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("loader calls = %d, want 2", got)
	}
}

func TestBundleStoreSharesConcurrentAcquire(t *testing.T) {
	index := bundleIndex(t, manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1})
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(ctx context.Context, entry manifestindex.Entry) (*engine.Bundle, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		select {
		case <-release:
			return &engine.Bundle{Name: entry.Connector}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	first := make(chan *BundleHandle, 1)
	errs := make(chan error, 2)
	go func() {
		handle, err := store.Acquire(context.Background(), "alpha")
		if err != nil {
			errs <- err
			return
		}
		first <- handle
	}()
	<-started
	second := make(chan *BundleHandle, 1)
	go func() {
		handle, err := store.Acquire(context.Background(), "alpha")
		if err != nil {
			errs <- err
			return
		}
		second <- handle
	}()
	close(release)
	left := <-first
	right := <-second
	defer left.Release()
	defer right.Release()
	select {
	case err := <-errs:
		t.Fatal(err)
	default:
	}
	if left.Bundle() != right.Bundle() {
		t.Fatal("concurrent acquires did not receive the same immutable bundle")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("loader calls = %d, want 1", got)
	}
}

func TestBundleStoreCancelsAbandonedLoadAndAllowsRetry(t *testing.T) {
	entry := manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1}
	index := bundleIndex(t, entry)
	started := make(chan struct{})
	var calls atomic.Int32
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(ctx context.Context, entry manifestindex.Entry) (*engine.Bundle, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-ctx.Done()
			return nil, ctx.Err()
		}
		return &engine.Bundle{Name: entry.Connector}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		handle, err := store.Acquire(ctx, "alpha")
		if handle != nil {
			handle.Release()
		}
		result <- err
	}()
	<-started
	store.mu.Lock()
	pending := store.flights[identityFor(entry)]
	store.mu.Unlock()
	if pending == nil {
		t.Fatal("abandoned load has no tracked flight")
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("abandoned Acquire() error = %v, want context.Canceled", err)
	}
	<-pending.done

	handle, err := store.Acquire(context.Background(), "alpha")
	if err != nil {
		t.Fatalf("retry Acquire(): %v", err)
	}
	defer handle.Release()
	if got := calls.Load(); got != 2 {
		t.Fatalf("loader calls = %d, want 2 after an abandoned flight", got)
	}
}

func TestBundleStoreReturnsCanonicalEntryIdentity(t *testing.T) {
	want := manifestindex.Entry{
		Connector:  "alpha",
		Generation: "generation-7",
		Digest:     "sha256:alpha",
		Executor:   "api_engine.v1",
		Extension:  "hook/alpha.v1",
		Metadata:   connectors.Metadata{Name: "alpha", DisplayName: "Alpha", IntegrationType: "api"},
		Bytes:      1,
	}
	index := bundleIndex(t, want)
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(_ context.Context, entry manifestindex.Entry) (*engine.Bundle, error) {
		if entry.Generation != want.Generation || entry.Digest != want.Digest || entry.Extension != want.Extension {
			t.Fatalf("loader entry = %#v, want canonical identity %#v", entry, want)
		}
		return &engine.Bundle{Name: entry.Connector}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	handle, err := store.Acquire(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()
	if got := handle.Entry(); got.Generation != want.Generation || got.Digest != want.Digest || got.Extension != want.Extension {
		t.Fatalf("handle entry = %#v, want canonical identity %#v", got, want)
	}
}

func TestBundleStoreSeparatesConnectorGenerations(t *testing.T) {
	first := manifestindex.Entry{Connector: "alpha", Generation: "g1", Digest: "sha256:first", Executor: "api_engine.v1", Bytes: 1}
	second := manifestindex.Entry{Connector: "alpha", Generation: "g2", Digest: "sha256:second", Executor: "api_engine.v1", Bytes: 1}
	var calls atomic.Int32
	store, err := NewBundleStore(bundleIndex(t, first), Limits{Entries: 2, Bytes: 2}, func(_ context.Context, entry manifestindex.Entry) (*engine.Bundle, error) {
		calls.Add(1)
		return &engine.Bundle{Name: entry.Connector}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	firstHandle, err := store.Acquire(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	firstHandle.Release()

	secondIndex := bundleIndex(t, second)
	store.mu.Lock()
	store.index = secondIndex
	store.mu.Unlock()
	secondHandle, err := store.Acquire(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	defer secondHandle.Release()
	if got := calls.Load(); got != 2 {
		t.Fatalf("loader calls = %d, want 2 for distinct connector generations", got)
	}
	if got := secondHandle.Entry(); got.Generation != second.Generation || got.Digest != second.Digest {
		t.Fatalf("second handle entry = %#v, want %#v", got, second)
	}
}

func TestBundleStoreRejectsGenerationOrDigestMismatch(t *testing.T) {
	canonical := manifestindex.Entry{Connector: "alpha", Generation: "g1", Digest: "sha256:one", Executor: "api_engine.v1", Bytes: 1}
	var calls atomic.Int32
	store, err := NewBundleStore(bundleIndex(t, canonical), Limits{Entries: 1, Bytes: 1}, func(_ context.Context, entry manifestindex.Entry) (*engine.Bundle, error) {
		calls.Add(1)
		return &engine.Bundle{Name: entry.Connector}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []manifestindex.Entry{
		{Connector: "alpha", Generation: "g2", Digest: canonical.Digest, Executor: canonical.Executor, Bytes: canonical.Bytes},
		{Connector: "alpha", Generation: canonical.Generation, Digest: "sha256:two", Executor: canonical.Executor, Bytes: canonical.Bytes},
	} {
		if _, err := store.AcquireEntry(context.Background(), entry); !errors.Is(err, ErrBundleIdentityMismatch) {
			t.Fatalf("AcquireEntry(%#v) error = %v, want ErrBundleIdentityMismatch", entry, err)
		}
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("loader calls = %d, want 0 for generation/digest mismatch", got)
	}
}

func TestBundleHandlePinsGenerationAfterCacheRelease(t *testing.T) {
	entry := manifestindex.Entry{Connector: "alpha", Generation: "g1", Digest: "sha256:one", Executor: "api_engine.v1", Bytes: 1}
	store, err := NewBundleStore(bundleIndex(t, entry), Limits{Entries: 1, Bytes: 1}, func(_ context.Context, selected manifestindex.Entry) (*engine.Bundle, error) {
		return &engine.Bundle{Name: selected.Connector}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	handle, err := store.Acquire(context.Background(), entry.Connector)
	if err != nil {
		t.Fatal(err)
	}
	lease := handle.HoldGeneration()
	if lease == nil {
		t.Fatal("HoldGeneration() returned nil")
	}
	handle.Release()
	if !store.GenerationHeld(entry) {
		t.Fatal("generation hold disappeared when the cache handle was released")
	}
	if got := lease.Entry(); got.Generation != entry.Generation || got.Digest != entry.Digest {
		t.Fatalf("lease entry = %#v, want %#v", got, entry)
	}
	lease.Release()
	if store.GenerationHeld(entry) {
		t.Fatal("generation hold remained after lease release")
	}
}

func bundleIndex(t *testing.T, entries ...manifestindex.Entry) manifestindex.Index {
	t.Helper()
	index, err := manifestindex.New(entries, len(entries))
	if err != nil {
		t.Fatal(err)
	}
	return index
}
