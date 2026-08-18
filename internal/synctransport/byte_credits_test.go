package synctransport

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestByteCreditControllerBoundsConcurrentRetainedBytes(t *testing.T) {
	credits, err := NewByteCreditController(10)
	if err != nil {
		t.Fatal(err)
	}
	first, err := credits.Acquire(context.Background(), 8)
	if err != nil {
		t.Fatal(err)
	}
	acquired := make(chan *ByteCreditLease, 1)
	go func() {
		lease, acquireErr := credits.Acquire(context.Background(), 4)
		if acquireErr == nil {
			acquired <- lease
		}
	}()
	select {
	case <-acquired:
		t.Fatal("second acquisition bypassed retained-byte credit limit")
	case <-time.After(20 * time.Millisecond):
	}
	first.Release()
	select {
	case second := <-acquired:
		second.Release()
	case <-time.After(time.Second):
		t.Fatal("waiting credit acquisition did not resume after release")
	}
	snapshot := credits.Snapshot()
	if snapshot.InUse != 0 || snapshot.Peak != 8 || snapshot.WaitNanos <= 0 {
		t.Fatalf("credit snapshot = %#v, want released bounded peak and wait accounting", snapshot)
	}
}

func TestByteCreditControllerRefusesOversizedUnitBeforeWaiting(t *testing.T) {
	credits, err := NewByteCreditController(10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := credits.Acquire(context.Background(), 11); !errors.Is(err, ErrByteCreditsInvalid) {
		t.Fatalf("Acquire(oversized) error = %T %v, want ErrByteCreditsInvalid", err, err)
	}
	if got := credits.Snapshot(); got.InUse != 0 || got.Peak != 0 {
		t.Fatalf("oversized acquisition mutated credit state: %#v", got)
	}
}

func TestByteCreditControllerCancelledWaitDoesNotLeakCredit(t *testing.T) {
	credits, err := NewByteCreditController(1)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := credits.Acquire(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := credits.Acquire(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("Acquire(cancelled) error = %T %v, want context.Canceled", err, err)
	}
	lease.Release()
	lease.Release()
	if got := credits.Snapshot(); got.InUse != 0 {
		t.Fatalf("idempotent release leaked credit: %#v", got)
	}
}
