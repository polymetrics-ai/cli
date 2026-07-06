package safetyculture

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/safetyculture"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("safetyculture")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "safetyculture" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.CheckHook); !ok {
		t.Fatal("hooks do not implement CheckHook")
	}
	if _, ok := h.(engine.StreamHook); !ok {
		t.Fatal("hooks do not implement StreamHook")
	}
}

func TestHooksDelegateFixtureCheckAndRead(t *testing.T) {
	h := Hooks{Connector: native.New()}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	handled, err := h.Check(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !handled {
		t.Fatal("Check handled = false")
	}

	count := 0
	handled, err = h.ReadStream(context.Background(), engine.StreamSpec{Name: "audits"}, connectors.ReadRequest{Stream: "audits", Config: cfg}, nil, func(connectors.Record) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream handled = false")
	}
	if count == 0 {
		t.Fatal("ReadStream emitted zero fixture records")
	}
}

func TestHooksIgnoreUnknownStream(t *testing.T) {
	h := Hooks{Connector: native.New()}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "unknown"}, connectors.ReadRequest{Stream: "unknown"}, nil, func(connectors.Record) error {
		t.Fatal("unexpected emit")
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream handled unknown stream")
	}
}
