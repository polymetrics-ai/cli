package bundleregistry

import (
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

var bundleCache struct {
	once    sync.Once
	bundles []engine.Bundle
	err     error
}

func New() *connectors.Registry {
	bundleCache.once.Do(func() {
		bundleCache.bundles, bundleCache.err = engine.LoadAll(defs.FS)
	})
	if bundleCache.err != nil {
		panic("load connector definition bundles: " + bundleCache.err.Error())
	}

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range bundleCache.bundles {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	return registry
}
