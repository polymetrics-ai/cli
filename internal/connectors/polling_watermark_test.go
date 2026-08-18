package connectors

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func TestLegacyPollingWatermarkModeUsesSharedFiveModeCompatibilityTable(t *testing.T) {
	for _, public := range synccontract.PublicModes() {
		got, ok := LegacyPollingWatermarkMode(public.Name)
		if !ok || got != public.ContractMode {
			t.Fatalf("LegacyPollingWatermarkMode(%q) = %q, %t; want #3810 mapping %q, true", public.Name, got, ok, public.ContractMode)
		}
	}
	for _, unsupported := range []string{"incremental_merge", "full_append", "incremental_append_dedup"} {
		if got, ok := LegacyPollingWatermarkMode(unsupported); ok || got != "" {
			t.Fatalf("LegacyPollingWatermarkMode(%q) = %q, %t; want no independent alias", unsupported, got, ok)
		}
	}
}

func TestPlannedPollingWatermarkJSONOmitsUndeclaredRuntimeContract(t *testing.T) {
	raw, err := json.Marshal(PollingWatermarkDescriptor{
		Status: PollingWatermarkStatusPlanned,
		Reason: "native polling binding is not registered",
	})
	if err != nil {
		t.Fatalf("Marshal(planned polling watermark): %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("Unmarshal(planned polling watermark): %v", err)
	}
	if got, want := string(fields["status"]), `"planned"`; got != want {
		t.Fatalf("planned polling watermark status = %s, want %s", got, want)
	}
	if _, found := fields["reason"]; !found {
		t.Fatalf("planned polling watermark omitted blocking reason: %s", raw)
	}
	for _, fabricated := range []string{"source", "target"} {
		if _, found := fields[fabricated]; found {
			t.Fatalf("planned polling watermark fabricated %s contract: %s", fabricated, raw)
		}
	}
}
