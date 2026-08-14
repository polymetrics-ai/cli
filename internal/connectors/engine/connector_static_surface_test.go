package engine_test

import (
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func Test100msRoomsRetainsFullRefreshTypedCompatibilityMode(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "100ms")
	if err != nil {
		t.Fatalf("Load(100ms) error = %v", err)
	}

	definition := engine.New(bundle, nil).Definition()
	for _, stream := range definition.Streams {
		if stream.Name != "rooms" {
			continue
		}
		want := []string{"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped"}
		if !reflect.DeepEqual(stream.SyncModes, want) {
			t.Fatalf("rooms SyncModes = %v, want %v", stream.SyncModes, want)
		}
		return
	}
	t.Fatal("Definition().Streams is missing rooms")
}
