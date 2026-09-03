// Package manifeststore lazily loads immutable execution manifests from a bounded index.
package manifeststore

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"polymetrics.ai/internal/connectors/manifestindex"
)

// Limits bounds data retained after successful loads.
type Limits struct {
	Entries int
	Bytes   int
}

// Loader transfers ownership of a successful immutable execution manifest to Store.
type Loader func(context.Context, manifestindex.Entry) ([]byte, error)

// Store shares indexed manifest loads and retains least-recently-used results within Limits.
type Store struct {
	index  manifestindex.Index
	limits Limits
	loader Loader

	mu      sync.Mutex
	cache   map[string]*cacheEntry
	lru     list.List
	bytes   int
	flights map[string]*flight
}

type cacheEntry struct {
	data []byte
	elem *list.Element
}

type flight struct {
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	waiters int
	data    []byte
	err     error
}

// New creates a lazy store over index. Neither retained manifest count nor bytes may be unbounded.
func New(index manifestindex.Index, limits Limits, loader Loader) (*Store, error) {
	if limits.Entries <= 0 {
		return nil, fmt.Errorf("manifest store entry limit must be positive")
	}
	if limits.Bytes <= 0 {
		return nil, fmt.Errorf("manifest store byte limit must be positive")
	}
	if loader == nil {
		return nil, fmt.Errorf("manifest store loader is required")
	}
	return &Store{
		index:   index,
		limits:  limits,
		loader:  loader,
		cache:   make(map[string]*cacheEntry),
		flights: make(map[string]*flight),
	}, nil
}

// Load returns a private copy of an indexed execution manifest.
func (s *Store) Load(ctx context.Context, connector string) ([]byte, error) {
	if ctx == nil {
		return nil, fmt.Errorf("manifest load context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entry, ok := s.index.Lookup(connector)
	if !ok {
		return nil, fmt.Errorf("unknown manifest connector %q", connector)
	}

	s.mu.Lock()
	if cached, ok := s.cache[connector]; ok {
		s.lru.MoveToFront(cached.elem)
		data := clone(cached.data)
		s.mu.Unlock()
		return data, nil
	}
	if pending, ok := s.flights[connector]; ok {
		pending.waiters++
		s.mu.Unlock()
		return s.wait(ctx, connector, pending)
	}
	loadCtx, cancel := context.WithCancel(context.Background())
	pending := &flight{ctx: loadCtx, cancel: cancel, done: make(chan struct{}), waiters: 1}
	s.flights[connector] = pending
	s.mu.Unlock()

	go s.load(connector, entry, pending)
	return s.wait(ctx, connector, pending)
}

func (s *Store) wait(ctx context.Context, connector string, pending *flight) ([]byte, error) {
	select {
	case <-pending.done:
		if pending.err != nil {
			return nil, pending.err
		}
		return clone(pending.data), nil
	case <-ctx.Done():
		s.release(connector, pending)
		return nil, ctx.Err()
	}
}

func (s *Store) release(connector string, pending *flight) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.flights[connector] != pending {
		return
	}
	pending.waiters--
	if pending.waiters != 0 {
		return
	}
	delete(s.flights, connector)
	pending.cancel()
}

func (s *Store) load(connector string, entry manifestindex.Entry, pending *flight) {
	data, err := s.loader(pending.ctx, entry)

	s.mu.Lock()
	if s.flights[connector] == pending {
		delete(s.flights, connector)
		if err == nil {
			s.admit(connector, data)
		}
	}
	pending.data = data
	pending.err = err
	close(pending.done)
	s.mu.Unlock()
}

func (s *Store) admit(connector string, data []byte) {
	if len(data) > s.limits.Bytes {
		return
	}
	for len(s.cache) >= s.limits.Entries || s.bytes > s.limits.Bytes-len(data) {
		oldest := s.lru.Back()
		oldConnector := oldest.Value.(string)
		old := s.cache[oldConnector]
		delete(s.cache, oldConnector)
		s.lru.Remove(old.elem)
		s.bytes -= len(old.data)
	}
	elem := s.lru.PushFront(connector)
	s.cache[connector] = &cacheEntry{data: data, elem: elem}
	s.bytes += len(data)
}

func clone(data []byte) []byte { return append([]byte(nil), data...) }
