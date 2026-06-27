package temporalprobe

import (
	"context"
	"testing"
	"time"
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
