package e2etest_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	e2etest "polymetrics.ai/internal/connectors/e2e-test"
)

func TestReadDeterministicInProcessRecords(t *testing.T) {
	c := e2etest.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"max_records": "3", "seed": "7"}}
	read := func() []connectors.Record {
		var got []connectors.Record
		if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "data", Config: cfg}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		}); err != nil {
			t.Fatalf("Read: %v", err)
		}
		return got
	}
	first := read()
	second := read()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("records not deterministic: first=%+v second=%+v", first, second)
	}
	if len(first) != 3 || first[0]["id"] != "7-0" || first[0]["column1"] == nil {
		t.Fatalf("records mapped wrong: %+v", first)
	}
}

func TestFixtureCatalogRegistrationAndReadOnly(t *testing.T) {
	c := e2etest.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	count := 0
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "data", Config: cfg}, func(connectors.Record) error { count++; return nil }); err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture Read emitted no records")
	}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "e2e-test" || len(cat.Streams) != 1 {
		t.Fatalf("Catalog = %+v err=%v", cat, err)
	}
	if _, ok := connectors.NewRegistry().Get("e2e-test"); !ok {
		t.Fatal("registry did not resolve e2e-test")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
