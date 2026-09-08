package engine

import (
	"context"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// readRequestBudget is one aggregate caller-owned send budget shared by every
// requester clone in a stream read. The mutex makes the admission safe for a
// hook or transport that issues concurrent requests without changing the
// declarative reader's established sequential behavior.
type readRequestBudget struct {
	mu    sync.Mutex
	limit int
	used  int
}

func newReadRequestBudget(limit int) *readRequestBudget {
	if limit <= 0 {
		return nil
	}
	return &readRequestBudget{limit: limit}
}

func (b *readRequestBudget) AdmitSend(ctx context.Context, _ connsdk.RateLimitRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.used >= b.limit {
		return &connectors.ReadRequestBudgetExceededError{Limit: b.limit, Used: b.used}
	}
	b.used++
	return nil
}

func attachReadRequestBudget(rt *Runtime, limit int) {
	budget := newReadRequestBudget(limit)
	if budget == nil || rt == nil {
		return
	}
	if rt.Requester != nil {
		rt.Requester.SendAdmission = budget
	}
	if rt.baseRequester != nil {
		rt.baseRequester.SendAdmission = budget
	}
}
