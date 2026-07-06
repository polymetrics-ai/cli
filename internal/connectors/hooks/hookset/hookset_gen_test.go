package hookset

import (
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestGeneratedHooksetRegistersRepresentativeHooks(t *testing.T) {
	cases := []struct {
		name       string
		wantAuth   bool
		wantStream bool
	}{
		{name: "ebay-fulfillment", wantAuth: true, wantStream: true},
		{name: "hoorayhr", wantAuth: true},
		{name: "snapchat-marketing", wantAuth: true},
		{name: "strava", wantAuth: true},
		{name: "uptick", wantAuth: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := engine.HooksFor(tc.name)
			if h == nil {
				t.Fatalf("HooksFor(%q) = nil", tc.name)
			}
			if h.ConnectorName() != tc.name {
				t.Fatalf("ConnectorName() = %q, want %q", h.ConnectorName(), tc.name)
			}
			if _, ok := h.(engine.AuthHook); ok != tc.wantAuth {
				t.Fatalf("AuthHook implemented = %v, want %v", ok, tc.wantAuth)
			}
			if _, ok := h.(engine.StreamHook); ok != tc.wantStream {
				t.Fatalf("StreamHook implemented = %v, want %v", ok, tc.wantStream)
			}
		})
	}
}
