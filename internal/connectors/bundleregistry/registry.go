package bundleregistry

import (
	"io/fs"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/hookset"
	"polymetrics.ai/internal/connectors/native/nativeset"
)

func init() {
	connectors.RegisterDefaultRegistryBuilder(New)
}

// bundleCache keeps the parsed embedded definition corpus process-local. defs.FS is immutable at
// runtime, so reparsing its full fleet for every CLI invocation adds work without refreshing any
// state. New still constructs an independent Registry and connector wrappers for every caller.
type bundleCache struct {
	once    sync.Once
	loader  func(fs.FS) ([]engine.Bundle, error)
	bundles []engine.Bundle
	err     error
}

func (c *bundleCache) load(fsys fs.FS) ([]engine.Bundle, error) {
	c.once.Do(func() {
		c.bundles, c.err = c.loader(fsys)
	})
	// Protect the cached slice header from callers. Bundle contents are immutable projections of
	// the embedded definitions and engine connectors treat them as read-only.
	return append([]engine.Bundle(nil), c.bundles...), c.err
}

var embeddedBundles = bundleCache{loader: engine.LoadAll}

func New() *connectors.Registry {
	bundles, err := embeddedBundles.load(defs.FS)
	if err != nil {
		panic("load connector definition bundles: " + err.Error())
	}

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range bundles {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	registry.MustValidateIconCoverage()
	return registry
}
