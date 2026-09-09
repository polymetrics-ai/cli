package coordination

import (
	"context"
	"errors"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"sync"
)

// OwnedSharedRateLimitClient belongs to one scope. Runtime consumers receive
// only connectors.SharedRateLimitCoordinator, which deliberately has no Close.
type OwnedSharedRateLimitClient interface {
	connectors.SharedRateLimitCoordinator
	Close() error
}

// SharedRateLimitScopeOptions supplies an instance-owned constructor. Nil uses
// the real Redis registry constructor. It must construct without network I/O.
type SharedRateLimitScopeOptions struct {
	Open func(string) (OwnedSharedRateLimitClient, error)
}

// SharedRateLimitScope snapshots one invocation's configuration and opens at
// most one optional client, on first use. Owners must quiesce borrowers before
// Close. No mutex is held across availability/admission or borrower callbacks.
type SharedRateLimitScope struct {
	mu          sync.Mutex
	addr        string
	open        func(string) (OwnedSharedRateLimitClient, error)
	initialized bool
	client      OwnedSharedRateLimitClient
	initErr     error
	closed      bool
	closeDone   chan struct{}
	closeErr    error
}

func NewSharedRateLimitScope(addr string, options SharedRateLimitScopeOptions) *SharedRateLimitScope {
	open := options.Open
	if open == nil {
		open = func(addr string) (OwnedSharedRateLimitClient, error) { return OpenSharedRateLimitRegistry(addr), nil }
	}
	return &SharedRateLimitScope{addr: addr, open: open, closeDone: make(chan struct{})}
}

func (s *SharedRateLimitScope) acquire() (OwnedSharedRateLimitClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, sharedRateLimitUnavailable(SharedRateLimitCoordinatorClosed)
	}
	if !s.initialized {
		s.initialized = true
		s.client, s.initErr = s.open(s.addr)
		if s.initErr != nil || s.client == nil {
			if s.client != nil {
				s.initErr = errors.Join(s.initErr, s.client.Close())
				s.client = nil
			}
			if s.initErr == nil {
				s.initErr = sharedRateLimitUnavailable(SharedRateLimitCoordinatorNotConfigured)
			}
		}
	}
	if s.initErr != nil {
		return nil, sharedRateLimitUnavailable(SharedRateLimitCoordinatorUnreachable)
	}
	return s.client, nil
}

func (s *SharedRateLimitScope) ResolveRateLimit(ctx context.Context, connector, policyID string, scope connectors.RateLimitScopeKey, budgets []connsdk.RateLimitBudget) (connsdk.RateLimitAdmission, connsdk.RateLimitObserver, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	client, err := s.acquire()
	if err != nil {
		return nil, nil, err
	}
	return client.ResolveRateLimit(ctx, connector, policyID, scope, budgets)
}

// Close retains cleanup errors for direct callers. Repeated Close returns the
// same result. Closing an unused scope never calls its constructor.
func (s *SharedRateLimitScope) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.closed {
		done := s.closeDone
		s.mu.Unlock()
		<-done
		s.mu.Lock()
		err := s.closeErr
		s.mu.Unlock()
		return err
	}
	s.closed = true
	client := s.client
	initErr := s.initErr
	s.mu.Unlock()
	err := initErr
	if client != nil {
		err = errors.Join(err, client.Close())
	}
	s.mu.Lock()
	s.closeErr = err
	close(s.closeDone)
	s.mu.Unlock()
	return err
}
