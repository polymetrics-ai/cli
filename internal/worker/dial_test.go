package worker

import (
	"context"
	"errors"
	"testing"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
)

func TestSubmitterForActivitiesContextUsesBoundedContextDial(t *testing.T) {
	oldDial := temporalClientDial
	t.Cleanup(func() { temporalClientDial = oldDial })

	called := false
	temporalClientDial = func(ctx context.Context, _ string, _ tlog.Logger) (client.Client, error) {
		called = true
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("dial context missing deadline")
		}
		return nil, context.DeadlineExceeded
	}

	_, _, err := SubmitterForActivitiesContext(context.Background(), "temporal.invalid:7233", false, nil)
	if !called {
		t.Fatal("temporal dial seam was not called")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("SubmitterForActivitiesContext error = %v, want deadline exceeded", err)
	}
}

func TestServeWithActivitiesUsesCancelableDialContext(t *testing.T) {
	oldDial := temporalClientDial
	t.Cleanup(func() { temporalClientDial = oldDial })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	temporalClientDial = func(ctx context.Context, _ string, _ tlog.Logger) (client.Client, error) {
		if err := ctx.Err(); err == nil {
			t.Fatal("dial context was not canceled")
		}
		return nil, ctx.Err()
	}

	err := ServeWithActivities(ctx, "temporal.invalid:7233", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ServeWithActivities error = %v, want context canceled", err)
	}
}
