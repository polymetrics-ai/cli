package awscloudtrail

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "aws-cloudtrail")
}

func assertConnectorContract(t *testing.T, c connectors.Connector, wantName string) {
	t.Helper()
	if c == nil {
		t.Fatal("New() = nil")
	}
	if got := c.Name(); got != wantName {
		t.Fatalf("Name() = %q, want %q", got, wantName)
	}
	meta := c.Metadata()
	if meta.Name != wantName {
		t.Fatalf("Metadata().Name = %q, want %q", meta.Name, wantName)
	}
	caps := meta.Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check, Catalog, and Read", caps)
	}
	if caps.Write {
		t.Fatalf("%s scope-corrected connector must not expose write capability until shared promoted-native forwarding is available", wantName)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != wantName {
		t.Fatalf("Catalog().Connector = %q, want %q", cat.Connector, wantName)
	}
	if got, want := len(cat.Streams), 19; got != want {
		t.Fatalf("Catalog streams = %d, want %d", got, want)
	}
}
