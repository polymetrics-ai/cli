package manifeststore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestidentity"
	"polymetrics.ai/internal/connectors/manifestindex"
)

func TestBundleStoreRejectsOversizeBeforeLoading(t *testing.T) {
	index := bundleIndex(t, manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 3})
	var calls atomic.Int32
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 2}, func(_ context.Context, entry manifestindex.Entry) (LoadedBundle, error) {
		calls.Add(1)
		return loadedFor(entry), nil
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
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(_ context.Context, entry manifestindex.Entry) (LoadedBundle, error) {
		calls.Add(1)
		return loadedFor(entry), nil
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
	entry := manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "d", Executor: "api_engine.v1", Bytes: 1}
	index := bundleIndex(t, entry)
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(ctx context.Context, selected manifestindex.Entry) (LoadedBundle, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		select {
		case <-release:
			return loadedFor(selected), nil
		case <-ctx.Done():
			return LoadedBundle{}, ctx.Err()
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
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(ctx context.Context, selected manifestindex.Entry) (LoadedBundle, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-ctx.Done()
			return LoadedBundle{}, ctx.Err()
		}
		return loadedFor(selected), nil
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
	store, err := NewBundleStore(index, Limits{Entries: 1, Bytes: 1}, func(_ context.Context, entry manifestindex.Entry) (LoadedBundle, error) {
		if entry.Generation != want.Generation || entry.Digest != want.Digest || entry.Extension != want.Extension {
			t.Fatalf("loader entry = %#v, want canonical identity %#v", entry, want)
		}
		return loadedFor(entry), nil
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
	store, err := NewBundleStore(bundleIndex(t, first), Limits{Entries: 2, Bytes: 2}, func(_ context.Context, entry manifestindex.Entry) (LoadedBundle, error) {
		calls.Add(1)
		return loadedFor(entry), nil
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
	store, err := NewBundleStore(bundleIndex(t, canonical), Limits{Entries: 1, Bytes: 1}, func(_ context.Context, entry manifestindex.Entry) (LoadedBundle, error) {
		calls.Add(1)
		return loadedFor(entry), nil
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

func TestBundleStoreRejectsSameNameLoadedIdentityMismatch(t *testing.T) {
	indexed := manifestindex.Entry{Connector: "alpha", Generation: "g1", Digest: "sha256:index", Executor: "api_engine.v1", Bytes: 1}
	for _, test := range []struct {
		name string
		edit func(*manifestindex.Entry)
	}{
		{"generation", func(entry *manifestindex.Entry) { entry.Generation = "g2" }},
		{"digest", func(entry *manifestindex.Entry) { entry.Digest = "sha256:loaded" }},
		{"charge", func(entry *manifestindex.Entry) { entry.Bytes = 2 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			loaded := indexed
			test.edit(&loaded)
			var calls atomic.Int32
			store, err := NewBundleStore(bundleIndex(t, indexed), Limits{Entries: 1, Bytes: 2}, func(_ context.Context, _ manifestindex.Entry) (LoadedBundle, error) {
				calls.Add(1)
				return loadedFor(loaded), nil
			})
			if err != nil {
				t.Fatal(err)
			}
			handle, err := store.Acquire(context.Background(), indexed.Connector)
			if !errors.Is(err, ErrBundleIdentityMismatch) {
				t.Fatalf("Acquire() error = %v, want ErrBundleIdentityMismatch", err)
			}
			if handle != nil {
				t.Fatal("same-name loaded identity mismatch returned a handle")
			}
			if got := calls.Load(); got != 1 {
				t.Fatalf("loader calls = %d, want 1", got)
			}
		})
	}
}

func TestBundleHandlePinsGenerationAfterCacheRelease(t *testing.T) {
	entry := manifestindex.Entry{Connector: "alpha", Generation: "g1", Digest: "sha256:one", Executor: "api_engine.v1", Bytes: 1}
	store, err := NewBundleStore(bundleIndex(t, entry), Limits{Entries: 1, Bytes: 1}, func(_ context.Context, selected manifestindex.Entry) (LoadedBundle, error) {
		return loadedFor(selected), nil
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

func TestBundleStoreReservesEntryCapacityAcrossDistinctFlights(t *testing.T) {
	alpha := manifestindex.Entry{Connector: "alpha", Generation: "g", Digest: "a", Executor: "api_engine.v1", Bytes: 1}
	bravo := manifestindex.Entry{Connector: "bravo", Generation: "g", Digest: "b", Executor: "api_engine.v1", Bytes: 1}
	alphaStarted := make(chan struct{})
	alphaRelease := make(chan struct{})
	var alphaCalls atomic.Int32
	var bravoCalls atomic.Int32
	store, err := NewBundleStore(bundleIndex(t, alpha, bravo), Limits{Entries: 1, Bytes: 2}, func(ctx context.Context, selected manifestindex.Entry) (LoadedBundle, error) {
		switch selected.Connector {
		case alpha.Connector:
			if alphaCalls.Add(1) == 1 {
				close(alphaStarted)
				<-alphaRelease
				return LoadedBundle{}, ctx.Err()
			}
		case bravo.Connector:
			bravoCalls.Add(1)
		default:
			return LoadedBundle{}, errors.New("unexpected bundle selection")
		}
		return loadedFor(selected), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	type acquireResult struct {
		handle *BundleHandle
		err    error
	}
	alphaCtx, cancelAlpha := context.WithCancel(context.Background())
	defer cancelAlpha()
	alphaResult := make(chan acquireResult, 1)
	go func() {
		handle, err := store.Acquire(alphaCtx, alpha.Connector)
		alphaResult <- acquireResult{handle: handle, err: err}
	}()
	<-alphaStarted

	store.mu.Lock()
	pending := store.flights[identityFor(alpha)]
	store.mu.Unlock()
	if pending == nil {
		t.Fatal("alpha load was not tracked as a flight")
	}
	assertBound := func(stage string) {
		t.Helper()
		store.mu.Lock()
		cached := len(store.cache)
		reservedEntries := store.reservedEntries
		reservedBytes := store.reserved // Every fixture bundle has a one-byte charge.
		store.mu.Unlock()
		if reservedEntries != reservedBytes {
			t.Errorf("%s reserved entry slots = %d, want %d one-byte reservations", stage, reservedEntries, reservedBytes)
		}
		if got := cached + reservedEntries; got > 1 {
			t.Errorf("%s retained plus reserved entries = %d, want at most 1", stage, got)
		}
	}
	requireBravoCapacity := func(stage string) bool {
		handle, err := store.Acquire(context.Background(), bravo.Connector)
		if handle != nil {
			handle.Release()
		}
		if !errors.Is(err, ErrBundleCapacity) {
			t.Errorf("%s Acquire(bravo) error = %v, want ErrBundleCapacity", stage, err)
			assertBound(stage)
			return false
		}
		if got := bravoCalls.Load(); got != 0 {
			t.Errorf("%s bravo loader calls = %d, want 0", stage, got)
			return false
		}
		assertBound(stage)
		return true
	}

	assertBound("first flight")
	if !requireBravoCapacity("while alpha is loading") {
		cancelAlpha()
		result := <-alphaResult
		if result.handle != nil {
			result.handle.Release()
		}
		close(alphaRelease)
		<-pending.done
		return
	}

	cancelAlpha()
	result := <-alphaResult
	if result.handle != nil {
		result.handle.Release()
	}
	if !errors.Is(result.err, context.Canceled) {
		t.Errorf("canceled Acquire(alpha) error = %v, want context.Canceled", result.err)
	}
	if result.handle != nil {
		t.Error("canceled Acquire(alpha) returned a handle")
	}
	assertBound("after alpha cancellation")
	if !requireBravoCapacity("while canceled alpha loader remains live") {
		close(alphaRelease)
		<-pending.done
		return
	}

	close(alphaRelease)
	<-pending.done
	assertBound("after canceled alpha completion")

	bravoHandle, err := store.Acquire(context.Background(), bravo.Connector)
	if err != nil {
		t.Fatalf("retry Acquire(bravo): %v", err)
	}
	assertBound("after bravo retry")
	bravoHandle.Release()

	alphaHandle, err := store.Acquire(context.Background(), alpha.Connector)
	if err != nil {
		t.Fatalf("retry Acquire(alpha): %v", err)
	}
	alphaHandle.Release()
	assertBound("after alpha retry")

	if got := alphaCalls.Load(); got != 2 {
		t.Fatalf("alpha loader calls = %d, want 2 after cancellation and retry", got)
	}
	if got := bravoCalls.Load(); got != 1 {
		t.Fatalf("bravo loader calls = %d, want 1 after capacity becomes available", got)
	}
}

func loadedFor(entry manifestindex.Entry) LoadedBundle {
	identity := manifestidentity.Identity{Connector: entry.Connector, Generation: entry.Generation, Digest: entry.Digest, Bytes: entry.Bytes}
	bundle := &engine.Bundle{Name: entry.Connector, Identity: identity}
	return LoadedBundle{Bundle: bundle, Identity: identity}
}

func bundleIndex(t *testing.T, entries ...manifestindex.Entry) manifestindex.Index {
	t.Helper()
	index, err := manifestindex.New(entries, len(entries))
	if err != nil {
		t.Fatal(err)
	}
	return index
}
