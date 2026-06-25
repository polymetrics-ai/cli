package connectors

import (
	"context"
	"testing"
)

// stubConnector is a minimal Connector used to exercise factory registration.
type stubConnector struct{ name string }

func (c stubConnector) Name() string { return c.name }
func (c stubConnector) Metadata() Metadata {
	return Metadata{Name: c.name, Capabilities: Capabilities{Check: true}}
}
func (c stubConnector) Check(ctx context.Context, cfg RuntimeConfig) error { return ctx.Err() }
func (c stubConnector) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	return Catalog{Connector: c.name}, nil
}
func (c stubConnector) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	return nil
}
func (c stubConnector) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	return WriteResult{}, nil
}

// TestRegisterFactoryIsResolvedByRegistry is the red-first test for the
// self-registration mechanism: a factory registered via RegisterFactory must be
// constructed and resolvable from NewRegistry().
func TestRegisterFactoryIsResolvedByRegistry(t *testing.T) {
	const name = "zz_factory_test_conn"
	RegisterFactory(name, func() Connector { return stubConnector{name: name} })
	t.Cleanup(func() { unregisterFactory(name) })

	r := NewRegistry()
	got, ok := r.Get(name)
	if !ok {
		t.Fatalf("registry did not resolve factory-registered connector %q", name)
	}
	if got.Name() != name {
		t.Fatalf("resolved connector name = %q, want %q", got.Name(), name)
	}
}
