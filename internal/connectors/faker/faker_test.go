package faker_test

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/faker"
)

func TestReadUsersDeterministic(t *testing.T) {
	c := faker.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "3", "seed": "7"}}
	var first []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		first = append(first, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read first: %v", err)
	}
	var second []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(rec connectors.Record) error {
		second = append(second, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read second: %v", err)
	}
	if len(first) != 3 || len(second) != 3 {
		t.Fatalf("record counts = %d/%d, want 3/3", len(first), len(second))
	}
	if first[0]["id"] != second[0]["id"] || first[0]["email"] != second[0]["email"] {
		t.Fatalf("faker output not deterministic: %+v vs %+v", first[0], second[0])
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := faker.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) < 3 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("faker"); !ok {
		t.Fatal("registry did not resolve faker")
	}
}
