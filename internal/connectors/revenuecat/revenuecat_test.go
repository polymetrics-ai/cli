package revenuecat

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorContract(t *testing.T) {
	c := New()
	if c.Name() != "revenuecat" {
		t.Fatalf("Name() = %q", c.Name())
	}
	meta := c.Metadata()
	if !meta.Capabilities.Check || !meta.Capabilities.Catalog || !meta.Capabilities.Read || meta.Capabilities.Write {
		t.Fatalf("unexpected capabilities: %+v", meta.Capabilities)
	}
	reg := connectors.NewRegistry()
	if _, ok := reg.Get("revenuecat"); !ok {
		t.Fatal("connector did not self-register")
	}
}

func TestFixtureModeCredentialFree(t *testing.T) {
	ctx := context.Background()
	c := New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(ctx, cfg); err != nil {
		t.Fatalf("fixture Check returned error: %v", err)
	}
	catalog, err := c.Catalog(ctx, cfg)
	if err != nil {
		t.Fatalf("Catalog returned error: %v", err)
	}
	if catalog.Connector != "revenuecat" || len(catalog.Streams) == 0 {
		t.Fatalf("unexpected catalog: %+v", catalog)
	}
	for _, stream := range catalog.Streams {
		var records []connectors.Record
		if err := c.Read(ctx, connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(record connectors.Record) error {
			records = append(records, record)
			return nil
		}); err != nil {
			t.Fatalf("fixture Read(%s) returned error: %v", stream.Name, err)
		}
		if len(records) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream.Name)
		}
		if records[0]["id"] == nil {
			t.Fatalf("fixture Read(%s) record missing id: %+v", stream.Name, records[0])
		}
	}
}

func TestReadOnlyWrite(t *testing.T) {
	_, err := New().Write(context.Background(), connectors.WriteRequest{}, []connectors.Record{{"id": "1"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
