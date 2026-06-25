package perf

import (
	"context"
	"testing"
)

func TestCompareDependencyFree(t *testing.T) {
	comparison, err := Compare(context.Background(), CompareRequest{Iterations: 2})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if comparison.DependencyFree.Mode != "dependency-free" {
		t.Fatalf("mode = %q", comparison.DependencyFree.Mode)
	}
	if comparison.DependencyFree.Records != 6 {
		t.Fatalf("records = %d, want 6", comparison.DependencyFree.Records)
	}
	if comparison.Explanation["dependency_free"] == "" {
		t.Fatalf("missing dependency-free explanation")
	}
}
