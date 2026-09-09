package engine

import (
	"context"
	"net/http"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestReadRecordsWrapFieldPreservesRawProviderRecord(t *testing.T) {
	server := jsonServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"item-1","nested":{"value":42}}]}`))
	})
	bundle := newTestBundle(t, server, StreamSpec{
		Records:    RecordsSpec{Path: "data", WrapField: "data"},
		Projection: "passthrough",
	})
	records, err := readAll(t, context.Background(), bundle, connectors.ReadRequest{}, nil)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v, want one wrapped record", records)
	}
	wrapped, ok := records[0]["data"].(map[string]any)
	if !ok || wrapped["id"] != "item-1" {
		t.Fatalf("wrapped record = %#v, want provider record under data", records[0]["data"])
	}
}
