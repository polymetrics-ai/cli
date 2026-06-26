package hardcodedrecords_test

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	hardcodedrecords "polymetrics.ai/internal/connectors/hardcoded-records"
)

func TestReadRecordsDeterministic(t *testing.T) {
	c := hardcodedrecords.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "4"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("records = %d, want 4", len(got))
	}
	if got[0]["id"] != "record_001" || got[3]["status"] != "inactive" {
		t.Fatalf("unexpected hardcoded records: %+v", got)
	}
}

func TestFixtureCatalogRegistryAndWrite(t *testing.T) {
	c := hardcodedrecords.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil || len(cat.Streams) == 0 {
		t.Fatalf("Catalog = %+v, %v", cat, err)
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
	if _, ok := connectors.NewRegistry().Get("hardcoded-records"); !ok {
		t.Fatal("registry did not resolve hardcoded-records")
	}
}
