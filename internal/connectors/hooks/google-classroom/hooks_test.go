package googleclassroom

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/google-classroom"
)

func TestHooksRegistered(t *testing.T) {
	h := engine.HooksFor("google-classroom")
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "google-classroom" {
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
	handled, err = h.ReadStream(context.Background(), engine.StreamSpec{Name: "courses"}, connectors.ReadRequest{Stream: "courses", Config: cfg}, nil, func(connectors.Record) error {
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

func TestReadStreamUnknownFallsBack(t *testing.T) {
	h := Hooks{Connector: native.New()}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "not_a_stream"}, connectors.ReadRequest{Stream: "not_a_stream"}, nil, func(connectors.Record) error {
		t.Fatal("emit should not be called for an unknown stream")
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("handled = true for unknown stream, want false")
	}
}

func TestReadStreamEmptyDefaultsToCourses(t *testing.T) {
	h := Hooks{Connector: native.New()}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	count := 0
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{}, connectors.ReadRequest{Config: cfg}, nil, func(connectors.Record) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("handled = false, want true")
	}
	if count == 0 {
		t.Fatal("ReadStream emitted zero fixture records")
	}
}
