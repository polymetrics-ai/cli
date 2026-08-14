package perf

import (
	"context"
	"reflect"
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

func TestCompareSyncModesSkipsTypedOnlyCompatibilityNames(t *testing.T) {
	benchmark, err := CompareSyncModes(context.Background(), SyncModeBenchmarkRequest{Records: 3})
	if err != nil {
		t.Fatalf("CompareSyncModes() error = %v", err)
	}

	wantModes := []string{"full_refresh_append", "full_refresh_overwrite", "incremental_append"}
	gotModes := make([]string, 0, len(benchmark.Results))
	for _, result := range benchmark.Results {
		gotModes = append(gotModes, result.Mode)
		if result.Error != "" {
			t.Fatalf("CompareSyncModes() result for %q failed: %s", result.Mode, result.Error)
		}
		if result.Records != 3 {
			t.Fatalf("CompareSyncModes() records for %q = %d, want 3", result.Mode, result.Records)
		}
	}
	if !reflect.DeepEqual(gotModes, wantModes) {
		t.Fatalf("CompareSyncModes() modes = %v, want %v", gotModes, wantModes)
	}
}
