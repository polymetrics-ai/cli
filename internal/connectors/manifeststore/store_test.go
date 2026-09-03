package manifeststore

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/manifestindex"
)

func TestStoreBoundsEntryAndByteCache(t *testing.T) {
	index := testIndex(t, "alpha", "bravo", "charlie")
	var mu sync.Mutex
	calls := map[string]int{}
	store, err := New(index, Limits{Entries: 2, Bytes: 3}, func(_ context.Context, entry manifestindex.Entry) ([]byte, error) {
		mu.Lock()
		calls[entry.Connector]++
		mu.Unlock()
		switch entry.Connector {
		case "charlie":
			return []byte("four"), nil
		default:
			return []byte("ok"), nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, connector := range []string{"alpha", "alpha", "bravo", "alpha", "charlie", "charlie"} {
		if _, err := store.Load(context.Background(), connector); err != nil {
			t.Fatalf("Load(%q): %v", connector, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if calls["alpha"] != 2 {
		t.Fatalf("alpha loads = %d, want 2 after the byte budget evicts the least-recent result", calls["alpha"])
	}
	if calls["bravo"] != 1 {
		t.Fatalf("bravo loads = %d, want 1", calls["bravo"])
	}
	if calls["charlie"] != 2 {
		t.Fatalf("charlie loads = %d, want 2 because a four-byte result exceeds the three-byte cache", calls["charlie"])
	}
}

func TestStoreBoundsEntryCache(t *testing.T) {
	index := testIndex(t, "alpha", "bravo", "charlie")
	var mu sync.Mutex
	calls := map[string]int{}
	store, err := New(index, Limits{Entries: 2, Bytes: 16}, func(_ context.Context, entry manifestindex.Entry) ([]byte, error) {
		mu.Lock()
		calls[entry.Connector]++
		mu.Unlock()
		return []byte("ok"), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, connector := range []string{"alpha", "bravo", "charlie", "alpha"} {
		if _, err := store.Load(context.Background(), connector); err != nil {
			t.Fatalf("Load(%q): %v", connector, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if calls["alpha"] != 2 {
		t.Fatalf("alpha loads = %d, want 2 after the two-entry cache evicts the least-recent result", calls["alpha"])
	}
}

func TestStoreRejectsUnknownWithoutLoading(t *testing.T) {
	index := testIndex(t, "alpha")
	var calls atomic.Int32
	store, err := New(index, Limits{Entries: 1, Bytes: 16}, func(_ context.Context, _ manifestindex.Entry) ([]byte, error) {
		calls.Add(1)
		return []byte("manifest"), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Load(context.Background(), "unknown"); err == nil {
		t.Fatal("Load accepted an unindexed connector")
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("loader calls = %d, want 0 for an unindexed connector", got)
	}
}

func TestStoreReturnsPrivateCachedCopies(t *testing.T) {
	index := testIndex(t, "alpha")
	var calls atomic.Int32
	store, err := New(index, Limits{Entries: 1, Bytes: 16}, func(_ context.Context, _ manifestindex.Entry) ([]byte, error) {
		calls.Add(1)
		return []byte("manifest"), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := store.Load(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	first[0] = 'X'
	second, err := store.Load(context.Background(), "alpha")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(second); got != "manifest" {
		t.Fatalf("cached manifest = %q, want manifest after caller mutation", got)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("loader calls = %d, want 1 for cached copy", got)
	}
}

func TestStoreSharesLoadWhenFirstWaiterCancels(t *testing.T) {
	index := testIndex(t, "alpha")
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	store, err := New(index, Limits{Entries: 1, Bytes: 16}, func(ctx context.Context, _ manifestindex.Entry) ([]byte, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		select {
		case <-release:
			return []byte("manifest"), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	firstContext, cancelFirst := context.WithCancel(context.Background())
	defer cancelFirst()
	first := make(chan error, 1)
	go func() {
		_, err := store.Load(firstContext, "alpha")
		first <- err
	}()
	<-started

	second := make(chan struct {
		data []byte
		err  error
	}, 1)
	go func() {
		data, err := store.Load(context.Background(), "alpha")
		second <- struct {
			data []byte
			err  error
		}{data, err}
	}()
	waitForFlightWaiters(t, store, "alpha", 2)
	cancelFirst()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("first Load error = %v, want context cancellation", err)
	}
	close(release)

	result := <-second
	if result.err != nil {
		t.Fatalf("second Load error = %v", result.err)
	}
	if got := string(result.data); got != "manifest" {
		t.Fatalf("second Load data = %q, want manifest", got)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("loader calls = %d, want 1", got)
	}
	if _, err := store.Load(context.Background(), "alpha"); err != nil {
		t.Fatalf("cached Load: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("cached Load triggered %d loader calls, want 1", got)
	}
}

func TestStoreCancelsAbandonedLoadAndAllowsRetry(t *testing.T) {
	index := testIndex(t, "alpha")
	started := make(chan struct{})
	cancelled := make(chan struct{})
	var calls atomic.Int32
	store, err := New(index, Limits{Entries: 1, Bytes: 16}, func(ctx context.Context, _ manifestindex.Entry) ([]byte, error) {
		if calls.Add(1) == 1 {
			close(started)
			<-ctx.Done()
			close(cancelled)
			return nil, ctx.Err()
		}
		return []byte("retry"), nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() {
		_, err := store.Load(ctx, "alpha")
		first <- err
	}()
	<-started
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Load error = %v, want context cancellation", err)
	}
	<-cancelled

	data, err := store.Load(context.Background(), "alpha")
	if err != nil {
		t.Fatalf("retry Load: %v", err)
	}
	if got := string(data); got != "retry" {
		t.Fatalf("retry Load data = %q, want retry", got)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("loader calls = %d, want a fresh second load", got)
	}
}

func waitForFlightWaiters(t *testing.T, store *Store, connector string, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		store.mu.Lock()
		pending := store.flights[connector]
		got := 0
		if pending != nil {
			got = pending.waiters
		}
		store.mu.Unlock()
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("in-flight waiters = %d, want %d", got, want)
		}
		time.Sleep(time.Millisecond)
	}
}

func testIndex(t *testing.T, connectors ...string) manifestindex.Index {
	t.Helper()
	entries := make([]manifestindex.Entry, 0, len(connectors))
	for _, connector := range connectors {
		entries = append(entries, manifestindex.Entry{Connector: connector, Generation: "g", Digest: "d", Executor: "e"})
	}
	index, err := manifestindex.New(entries, len(entries))
	if err != nil {
		t.Fatal(err)
	}
	return index
}
