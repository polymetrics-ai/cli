package temporalprobe

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
)

func TestProbe_EmptyAddr(t *testing.T) {
	if Probe(context.Background(), "") {
		t.Fatal("empty addr should be false")
	}
}

func TestProbe_UnreachableFailsFast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	if Probe(ctx, "127.0.0.1:1") { // nothing listening on port 1
		t.Fatal("unreachable addr should be false")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("probe took too long: %v", elapsed)
	}
}

func TestProbeTemporalDialUsesDialContext(t *testing.T) {
	src, err := os.ReadFile("temporalprobe.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "client.DialContext") {
		t.Fatal("temporal probe dial must use client.DialContext so caller cancellation bounds dial setup")
	}
}

func TestProbeUsesFiniteContextDial(t *testing.T) {
	oldDial := dialTemporal
	t.Cleanup(func() { dialTemporal = oldDial })
	called := make(chan struct{}, 1)
	dialTemporal = func(ctx context.Context, addr string, timeout time.Duration, logger tlog.Logger) (temporalHealthClient, error) {
		called <- struct{}{}
		if _, ok := ctx.Deadline(); !ok {
			t.Errorf("dial context has no deadline")
		}
		<-ctx.Done()
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if Probe(ctx, "temporal.example.test:7233") {
		t.Fatal("probe should fail when bounded dial context expires")
	}
	select {
	case <-called:
	default:
		t.Fatal("probe did not call bounded dial")
	}
}

func TestProbeUsesContextAwareHealthCheck(t *testing.T) {
	oldDial := dialTemporal
	t.Cleanup(func() { dialTemporal = oldDial })
	dialTemporal = func(ctx context.Context, addr string, timeout time.Duration, logger tlog.Logger) (temporalHealthClient, error) {
		return fakeTemporalClient{}, ctx.Err()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if !Probe(ctx, "temporal.example.test:7233") {
		t.Fatal("probe should report healthy fake Temporal client")
	}
}

type fakeTemporalClient struct{}

func (fakeTemporalClient) CheckHealth(ctx context.Context, _ *client.CheckHealthRequest) (*client.CheckHealthResponse, error) {
	if _, ok := ctx.Deadline(); !ok {
		return nil, context.Canceled
	}
	return &client.CheckHealthResponse{}, nil
}

func (fakeTemporalClient) Close() {}
