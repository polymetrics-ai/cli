package connectors

import (
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
