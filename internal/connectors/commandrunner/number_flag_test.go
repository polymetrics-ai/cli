package commandrunner

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestCoerceFlagValueNumber(t *testing.T) {
	flag := connectors.CommandSurfaceFlag{Name: "unit-amount", Type: "number"}

	got, err := coerceFlagValue(flag, []string{"12.34"})
	if err != nil {
		t.Fatalf("coerceFlagValue: %v", err)
	}
	if got != 12.34 {
		t.Fatalf("coerceFlagValue = %#v, want 12.34", got)
	}

	for _, value := range []string{"NaN", "+Inf", "twelve"} {
		if _, err := coerceFlagValue(flag, []string{value}); err == nil {
			t.Errorf("coerceFlagValue(%q) accepted a non-finite or invalid number", value)
		}
	}
}
