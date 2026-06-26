package smoketest_test

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	smoketest "polymetrics.ai/internal/connectors/smoke-test"
)

func TestSmokeTestIsDeterministicReadOnly(t *testing.T) {
	c := smoketest.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || cat.Connector != "smoke-test" || len(cat.Streams) < 2 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}

	var first []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users"}, func(rec connectors.Record) error { first = append(first, rec); return nil }); err != nil {
		t.Fatalf("Read first: %v", err)
	}
	var second []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users"}, func(rec connectors.Record) error { second = append(second, rec); return nil }); err != nil {
		t.Fatalf("Read second: %v", err)
	}
	if len(first) != len(second) || len(first) == 0 || first[0]["id"] != second[0]["id"] {
		t.Fatalf("smoke-test records are not deterministic: first=%+v second=%+v", first, second)
	}
	if err := c.Check(context.Background(), connectors.RuntimeConfig{}); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if _, ok := connectors.NewRegistry().Get("smoke-test"); !ok {
		t.Fatal("registry did not resolve smoke-test")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
