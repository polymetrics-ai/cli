package hookset

import (
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestGeneratedHooksetConstructsRepresentativeHooks(t *testing.T) {
	factories := make(map[string]Factory, len(Factories()))
	connectors := make(map[string]string, len(Factories()))
	for _, factory := range Factories() {
		if factory.ID == "" || factory.Connector == "" || factory.New == nil {
			t.Fatalf("incomplete generated factory %#v", factory)
		}
		if _, exists := factories[factory.ID]; exists {
			t.Fatalf("duplicate generated factory %q", factory.ID)
		}
		if _, exists := connectors[factory.Connector]; exists {
			t.Fatalf("duplicate generated connector %q", factory.Connector)
		}
		factories[factory.ID] = factory
		connectors[factory.Connector] = factory.ID
	}
	if len(factories) != 49 {
		t.Fatalf("generated factory count = %d, want 49", len(factories))
	}
	cases := []struct {
		connector  string
		wantAuth   bool
		wantStream bool
	}{
		{connector: "ebay-fulfillment", wantAuth: true, wantStream: true},
		{connector: "hoorayhr", wantAuth: true},
		{connector: "snapchat-marketing", wantAuth: true},
		{connector: "strava", wantAuth: true},
		{connector: "uptick", wantAuth: true},
	}

	for _, tc := range cases {
		t.Run(tc.connector, func(t *testing.T) {
			id := "hook/" + tc.connector + ".v1"
			factory, ok := factories[id]
			if !ok {
				t.Fatalf("factory %q is missing", id)
			}
			if factory.Connector != tc.connector {
				t.Fatalf("factory %q connector = %q, want %q", id, factory.Connector, tc.connector)
			}
			hooks := factory.New()
			if hooks == nil {
				t.Fatalf("factory for %q returned nil", tc.connector)
			}
			if hooks.ConnectorName() != tc.connector {
				t.Fatalf("ConnectorName() = %q, want %q", hooks.ConnectorName(), tc.connector)
			}
			if _, ok := hooks.(engine.AuthHook); ok != tc.wantAuth {
				t.Fatalf("AuthHook implemented = %v, want %v", ok, tc.wantAuth)
			}
			if _, ok := hooks.(engine.StreamHook); ok != tc.wantStream {
				t.Fatalf("StreamHook implemented = %v, want %v", ok, tc.wantStream)
			}
		})
	}
}
