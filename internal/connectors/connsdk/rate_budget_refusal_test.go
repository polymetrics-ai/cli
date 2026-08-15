package connsdk_test

import (
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestRateBudgetRefusalErrorContract(t *testing.T) {
	cause := errors.New("coordinator unavailable")
	err := &connsdk.RateBudgetRefusalError{
		Code:   connsdk.RateBudgetRefusalSharedCoordinatorUnavailable,
		Reason: "coordinator_not_configured",
		Err:    cause,
	}
	if got, want := err.Code, connsdk.RateBudgetRefusalCode("shared_coordinator_unavailable"); got != want {
		t.Fatalf("rate refusal code = %q, want %q", got, want)
	}
	if !errors.Is(err, cause) {
		t.Fatal("RateBudgetRefusalError does not unwrap its safe cause")
	}
	var typed *connsdk.RateBudgetRefusalError
	if !errors.As(err, &typed) || typed.Reason != "coordinator_not_configured" {
		t.Fatalf("typed rate refusal = %#v", typed)
	}
}
