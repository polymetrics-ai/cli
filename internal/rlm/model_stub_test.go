package rlm

import (
	"context"
	"errors"
	"testing"
)

func TestModelStubReturnsNotImplemented(t *testing.T) {
	m := &ModelAnalyzer{}
	result, err := m.Run(context.Background(), RunRequest{})
	if err == nil {
		t.Fatal("expected ErrNotImplemented, got nil")
	}
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("expected ErrNotImplemented, got %v", err)
	}
	// Result should be zero value
	if result.RecordsScored != 0 || result.RecordsRead != 0 {
		t.Errorf("expected zero RunResult, got %+v", result)
	}
}

func TestModelModeString(t *testing.T) {
	m := &ModelAnalyzer{}
	if got := m.Mode(); got != "model" {
		t.Errorf("Mode() = %q, want %q", got, "model")
	}
}
