package monday

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestMondayWriteHookBlocksModeledMutations(t *testing.T) {
	h := Hooks{}
	wh, ok := any(h).(engine.WriteHook)
	if !ok {
		t.Fatal("monday Hooks must implement WriteHook to block modeled mutations until typed execution is hardened")
	}
	handled, err := wh.ExecuteWrite(context.Background(), engine.WriteAction{Name: "create_item", Kind: "create"}, connectors.Record{}, &engine.Runtime{})
	if !handled {
		t.Fatal("handled = false, want Monday WriteHook to own mutation dispatch")
	}
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("err = %v, want ErrUnsupportedOperation", err)
	}
}
