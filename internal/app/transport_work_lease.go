package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"polymetrics.ai/internal/synccontract"
)

// transportWorkLeaseDuration is deliberately long enough for one bounded
// provider or warehouse unit, but never permanent. A successor may take over
// only after this durable lease expires; every old effect boundary renews and
// verifies the same fence, so an expired owner cannot continue after a
// takeover.
const transportWorkLeaseDuration = 2 * time.Minute

type transportWorkLease struct {
	app    *App
	key    string
	workID string
	fence  int64

	mu    sync.Mutex
	state StreamState
}

func (a *App) claimTransportWorkLease(ctx context.Context, key, connection, stream, workID string, source synccontract.ResumeExpectation, overwrite bool) (*transportWorkLease, error) {
	if a == nil || key == "" || connection == "" || stream == "" || workID == "" {
		return nil, errors.New("transport stream work lease is invalid")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	until := now.Add(transportWorkLeaseDuration)
	var claimed StreamState
	if _, err := a.updateState(func(current state) (state, error) {
		currentState, present := current.StreamStates[key]
		currentState = cloneStreamState(currentState)
		if currentState.Checkpoint != nil {
			if err := validateStreamStateResume(currentState, source); err != nil {
				return current, err
			}
		}
		if currentState.ActiveWorkID != "" {
			// A terminal owner is a crashed/failed handoff, not a live worker.
			// Its successor receives a higher fence and must still reconcile any
			// durable receipt before the source can replay; a running owner (or an
			// unknown historical owner) remains fail-closed until expiry.
			if !transportWorkOwnerTerminal(current.Runs, currentState.ActiveWorkID) && (currentState.ActiveWorkLeaseUntil == nil || currentState.ActiveWorkLeaseUntil.After(now)) {
				return current, errTransportStreamWorkInProgress
			}
		}
		if currentState.ActiveWorkFence == int64(^uint64(0)>>1) {
			return current, errors.New("transport stream work fences are exhausted")
		}
		currentState.Connection = connection
		currentState.Stream = stream
		if currentState.GenerationID == 0 || overwrite {
			currentState.GenerationID++
		}
		currentState.ActiveWorkID = workID
		currentState.ActiveWorkFence++
		currentState.ActiveWorkLeaseUntil = &until
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[key] = cloneStreamState(currentState)
		claimed = cloneStreamState(currentState)
		_ = present // an absent state is intentionally claimed as a new stream.
		return current, nil
	}); err != nil {
		return nil, fmt.Errorf("claim transport stream work before I/O: %w", err)
	}
	return &transportWorkLease{app: a, key: key, workID: workID, fence: claimed.ActiveWorkFence, state: claimed}, nil
}

func transportWorkOwnerTerminal(runs []Run, workID string) bool {
	for _, run := range runs {
		if run.ID == workID {
			return run.Status == "failed" || run.Status == "completed"
		}
	}
	return false
}

// renew verifies that this exact durable work identity still owns the stream,
// then atomically extends it. It is called immediately before source, stage,
// destination, and checkpoint effects. A stale process cannot renew after a
// successor claimed the next monotonic fence.
func (l *transportWorkLease) renew(ctx context.Context) error {
	_, err := l.mutate(ctx, func(current StreamState) (StreamState, error) { return current, nil })
	return err
}

func (l *transportWorkLease) commit(ctx context.Context, checkpoint synccontract.CheckpointEnvelope) (StreamState, error) {
	return l.mutate(ctx, func(current StreamState) (StreamState, error) {
		updated := cloneStreamState(current)
		copy := checkpoint.Clone()
		updated.Checkpoint = &copy
		if checkpoint.CommittedAt != nil {
			updated.UpdatedAt = checkpoint.CommittedAt.UTC()
		}
		return updated, nil
	})
}

func (l *transportWorkLease) mutate(ctx context.Context, change func(StreamState) (StreamState, error)) (StreamState, error) {
	if l == nil || l.app == nil || change == nil {
		return StreamState{}, errTransportStreamWorkFenceLost
	}
	if err := ctx.Err(); err != nil {
		return StreamState{}, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().UTC()
	until := now.Add(transportWorkLeaseDuration)
	var updated StreamState
	if _, err := l.app.updateState(func(current state) (state, error) {
		actual, present := current.StreamStates[l.key]
		if !present || actual.ActiveWorkID != l.workID || actual.ActiveWorkFence != l.fence {
			return current, errTransportStreamWorkFenceLost
		}
		if actual.ActiveWorkLeaseUntil == nil || !actual.ActiveWorkLeaseUntil.After(now) {
			return current, errTransportStreamWorkFenceLost
		}
		next, err := change(cloneStreamState(actual))
		if err != nil {
			return current, err
		}
		next.ActiveWorkID = l.workID
		next.ActiveWorkFence = l.fence
		next.ActiveWorkLeaseUntil = &until
		current.StreamStates[l.key] = cloneStreamState(next)
		updated = cloneStreamState(next)
		return current, nil
	}); err != nil {
		return StreamState{}, fmt.Errorf("renew transport stream work fence: %w", err)
	}
	l.state = cloneStreamState(updated)
	return cloneStreamState(updated), nil
}

func (l *transportWorkLease) stateForTerminalRun() StreamState {
	if l == nil {
		return StreamState{}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	terminal := cloneStreamState(l.state)
	terminal.ActiveWorkID = ""
	terminal.ActiveWorkLeaseUntil = nil
	return terminal
}
