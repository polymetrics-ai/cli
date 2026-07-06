package faker_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/faker"
)

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "faker" {
		t.Fatalf("Name() = %q, want faker", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("faker source connector must be read-only, got Write=true")
	}
}

// TestNoInitRegistration is the required grep-guard (mirrors
// native/postgres's TestNoInitRegistration): the native package must NOT
// call RegisterFactory from anywhere in its own source,
// nor declare an init() function. The registration flip (wiring
// native/faker into the production registry) is a wave6 change; this wave
// only builds and tests the package standalone. This is a structural guard,
// so it inspects the actual .go source files rather than runtime registry
// state.
func TestNoInitRegistration(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed; cannot locate package directory")
	}
	dir := filepath.Dir(thisFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}

	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			// The grep-guard covers the package's own production source, not
			// its tests (this very test file legitimately mentions the
			// forbidden identifiers in prose/identifiers above).
			continue
		}
		found = true
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", e.Name(), err)
		}
		src := string(raw)
		if strings.Contains(src, "RegisterFactory(") {
			t.Fatalf("%s calls RegisterFactory — native/faker must NOT self-register (registration flip is wave6)", e.Name())
		}
		if strings.Contains(src, "func init()") {
			t.Fatalf("%s declares an init() function — native/faker must perform no registration side effects", e.Name())
		}
	}
	if !found {
		t.Fatal("no non-test .go source files found in native/faker; grep-guard did not actually scan anything")
	}
}

// TestConnectorSatisfiesCoreInterfaces compile/runtime-asserts the shape
// required by API-CONTRACT.md / design §B.7 Tier-3: Connector and
// DefinitionProvider. StatefulReader/CDCReader are deliberately NOT
// asserted: legacy faker implements neither (it is a stateless, full-refresh
// generator on every Read call), so this native port carries the identical
// interface surface forward.
func TestConnectorSatisfiesCoreInterfaces(t *testing.T) {
	c := native.New()
	var _ connectors.Connector = c
	if _, ok := any(c).(connectors.DefinitionProvider); !ok {
		t.Fatal("native faker connector must implement connectors.DefinitionProvider (engine.Base)")
	}
}

func TestCheckAlwaysSucceeds(t *testing.T) {
	c := native.New()
	if err := c.Check(context.Background(), connectors.RuntimeConfig{}); err != nil {
		t.Fatalf("Check: %v", err)
	}
}

func TestCheckRespectsContextCancellation(t *testing.T) {
	c := native.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Check(ctx, connectors.RuntimeConfig{}); err == nil {
		t.Fatal("Check with a cancelled context: want error, got nil")
	}
}

func TestCatalogHasThreeStreams(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) != 3 {
		t.Fatalf("Catalog Streams = %d, want 3: %+v", len(cat.Streams), cat.Streams)
	}
	names := map[string]bool{}
	for _, s := range cat.Streams {
		names[s.Name] = true
	}
	for _, want := range []string{"users", "purchases", "products"} {
		if !names[want] {
			t.Fatalf("Catalog missing stream %q: %+v", want, cat.Streams)
		}
	}
}

func TestReadUsersDefaultCount(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: connectors.RuntimeConfig{}}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1000 {
		t.Fatalf("len(got) = %d, want 1000 (default count)", len(got))
	}
	if got[0]["id"] != "user_001" || got[0]["email"] != "user001@example.com" {
		t.Fatalf("got[0] = %+v, unexpected shape", got[0])
	}
}

func TestReadDefaultsToUsersStream(t *testing.T) {
	c := native.New()
	var n int
	err := c.Read(context.Background(), connectors.ReadRequest{Config: connectors.RuntimeConfig{Config: map[string]string{"count": "2"}}}, func(connectors.Record) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n != 2 {
		t.Fatalf("n = %d, want 2 (empty stream defaults to users)", n)
	}
}

func TestReadPurchasesTiesUserAndProduct(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "3"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "purchases", Config: cfg}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
	if got[0]["id"] != "purchase_001" || got[0]["user_id"] != "user_001" || got[0]["product_id"] != "product_002" {
		t.Fatalf("got[0] = %+v, unexpected shape", got[0])
	}
	if amt, ok := got[0]["amount"].(float64); !ok || amt <= 0 {
		t.Fatalf("got[0][amount] = %v, want a positive float64", got[0]["amount"])
	}
}

func TestReadProductsAlwaysTen(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "500"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "products", Config: cfg}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 10 {
		t.Fatalf("len(got) = %d, want 10 regardless of count", len(got))
	}
	if got[0]["id"] != "product_001" || got[0]["sku"] != "SKU-001" {
		t.Fatalf("got[0] = %+v, unexpected shape", got[0])
	}
}

func TestReadSeedOffsetsIds(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "1", "seed": "41"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "user_042" {
		t.Fatalf("got = %+v, want id user_042 (seed 41 + i 1)", got)
	}
}

func TestReadNegativeSeedClampsToZero(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "1", "seed": "-5"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(r connectors.Record) error {
		got = append(got, r)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 1 || got[0]["id"] != "user_001" {
		t.Fatalf("got = %+v, want id user_001 (negative seed clamped to 0)", got)
	}
}

func TestReadInvalidCountErrors(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"count": "0"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with count=0: want error, got nil")
	}
}

func TestReadInvalidSeedErrors(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"seed": "not-a-number"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "users", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with non-numeric seed: want error, got nil")
	}
}

func TestReadUnknownStreamErrors(t *testing.T) {
	c := native.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "bogus", Config: connectors.RuntimeConfig{}}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read with unknown stream: want error, got nil")
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write err = %v, want ErrUnsupportedOperation", err)
	}
}
